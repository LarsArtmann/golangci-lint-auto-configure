//go:build integration

package cli_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// CLITimeout is the timeout for CLI command execution.
const CLITimeout = 30 * time.Second

// findBinary locates the CLI binary.
func findBinary() string {
	candidates := []string{
		"../../../bin/golangci-lint-auto-configure", // from internal/cli/
		"../../bin/golangci-lint-auto-configure",    // from cli/ (if compiled elsewhere)
		"bin/golangci-lint-auto-configure",          // from project root
	}

	// Also try relative to the test binary location
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for dir != "/" {
			candidate := filepath.Join(dir, "bin", "golangci-lint-auto-configure")
			candidates = append(candidates, candidate)
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return "../../../bin/golangci-lint-auto-configure" // fallback
}

// runCLI runs the CLI binary with the given arguments and returns output.
func runCLI(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CLITimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, findBinary(), args...)
	return cmd.CombinedOutput()
}

// testConfigCommand creates a test config file, runs a CLI command, and returns the output.
func testConfigCommand(tempDir, command, configContent string) ([]byte, error) {
	configPath := filepath.Join(tempDir, ".golangci.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		return nil, err
	}
	return runCLI(command, "--config", configPath)
}

// validConfigContent returns a config that passes all validation checks.
func validConfigContent() string {
	return `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - errcheck
    - staticcheck
    - govet
`
}

// testStandardConfigCommand runs a CLI command with a standard test config
func testStandardConfigCommand(tempDir, command string) {
	output, err := testConfigCommand(tempDir, command, validConfigContent())
	// May fail if golangci-lint not installed, but should handle gracefully
	Expect(err).ToNot(HaveOccurred())
	Expect(output).ToNot(BeEmpty())
}

// testStandardCommandContext creates a Ginkgo Context and It block for testing a CLI command.
func testStandardCommandContext(tempDir, command string) {
	Context(command+" command", func() {
		It("should "+command+" config", func() {
			testStandardConfigCommand(tempDir, command)
		})
	})
}

// TestIntegration runs integration tests for the CLI.
func TestIntegration(t *testing.T) {
	// Skip if binary doesn't exist
	if _, err := os.Stat("../../../bin/golangci-lint-auto-configure"); os.IsNotExist(err) {
		t.Skip("CLI binary not found, run 'just build' first")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = Describe("CLI Integration", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "integration-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("Help command", func() {
		It("should show help text", func() {
			output, err := runCLI("--help")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("golangci-lint-auto-configure"))
			Expect(string(output)).To(ContainSubstring("configure"))
		})

		It("should show configure help", func() {
			output, err := runCLI("configure", "--help")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("--priority"))
			Expect(string(output)).To(ContainSubstring("--preset"))
			Expect(string(output)).To(ContainSubstring("--detect"))
		})
	})

	Context("Version command", func() {
		It("should show version", func() {
			output, err := runCLI("--version")
			Expect(err).NotTo(HaveOccurred())
			// Version output varies, just check it doesn't fail
			Expect(len(output)).To(BeNumerically(">", 0))
		})
	})

	Context("Configure command", func() {
		It("should create default config", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")

			output, err := runCLI("configure", "--config", configPath, "--priority", "critical")
			Expect(err).NotTo(HaveOccurred())

			// Check output
			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Configuring golangci-lint"))

			// Check file was created
			_, err = os.Stat(configPath)
			Expect(err).NotTo(HaveOccurred())

			// Check file content
			content, err := os.ReadFile(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("version:"))
		})

		It("should support dry-run mode", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")

			// Create initial config
			initialContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run dry-run
			output, err := runCLI(
				"configure",
				"--config",
				configPath,
				"--priority",
				"high",
				"--dry-run",
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("DRY-RUN"))
		})

		It("should support preset mode", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")

			output, err := runCLI("configure", "--config", configPath, "--preset", "minimal")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("Applying preset"))
			Expect(string(output)).To(ContainSubstring("minimal"))
		})

		It("should support detect mode", func() {
			// Create a simple Go module
			goModPath := filepath.Join(tempDir, "go.mod")
			err := os.WriteFile(goModPath, []byte("module test\n\ngo 1.21\n"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create a main package
			mainPath := filepath.Join(tempDir, "main.go")
			err = os.WriteFile(mainPath, []byte("package main\n\nfunc main() {}\n"), 0o644)
			Expect(err).NotTo(HaveOccurred())

			configPath := filepath.Join(tempDir, ".golangci.yml")

			// Change to temp dir for detection
			originalWd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			defer os.Chdir(originalWd)

			err = os.Chdir(tempDir)
			Expect(err).NotTo(HaveOccurred())

			output, err := runCLI("configure", "--config", configPath, "--detect")
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Detected project type"))
			Expect(outputStr).To(ContainSubstring("Selected preset"))
		})
	})

	testStandardCommandContext(tempDir, "analyze")
	testStandardCommandContext(tempDir, "validate")

	Context("Analyze command", func() {
		It("should analyze with JSON output", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			initialContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			output, err := runCLI("analyze", "--config", configPath, "--format", "json")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("recommendations"))
		})

		It("should analyze with SARIF output", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			initialContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			output, err := runCLI("analyze", "--config", configPath, "--format", "sarif")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("$schema"))
		})
	})

	Context("Validate command", func() {
		It("should validate a valid config", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			initialContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			output, err := runCLI("validate", "--config", configPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("valid"))
		})

		It("should reject an invalid config", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			initialContent := `version: "1"
` // v1 config without migration
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			_, err = runCLI("validate", "--config", configPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Report command", func() {
		It("should generate JSON report", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			initialContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			reportPath := filepath.Join(tempDir, "report.json")
			output, err := runCLI("report", "--config", configPath, "--format", "json", "--output", reportPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("Report generated"))

			_, err = os.Stat(reportPath)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Migrate command", func() {
		It("should migrate v1 config to v2", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			v1Content := `linters-settings:
  gci:
    sections:
      - Standard
linters:
  enable:
    - gosec
`
			err := os.WriteFile(configPath, []byte(v1Content), 0o644)
			Expect(err).NotTo(HaveOccurred())

			output, err := runCLI("migrate", "--config", configPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("Migrated"))

			// Verify v2 format
			content, err := os.ReadFile(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("version: \"2\""))
		})
	})

	Context("Check mode", func() {
		It("should exit 0 for optimal config", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			initialContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
    - errcheck
    - staticcheck
    - govet
`
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			_, err = runCLI("configure", "--config", configPath, "--priority", "critical", "--check")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should exit 1 for suboptimal config", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			initialContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			_, err = runCLI("configure", "--config", configPath, "--priority", "critical", "--check")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Diff mode", func() {
		It("should show diff for changes", func() {
			configPath := filepath.Join(tempDir, ".golangci.yml")
			initialContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
			err := os.WriteFile(configPath, []byte(initialContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			output, err := runCLI("configure", "--config", configPath, "--priority", "critical", "--diff")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("Added"))
		})
	})
})
