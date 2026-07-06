//go:build integration

package cli_test

import (
	"context"
	"encoding/json"
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

// minimalConfigContent returns a minimal valid v2 config for testing.
const minimalConfigContent = `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`

// writeTestConfig writes a test config to the given temp dir and returns the path.
func writeTestConfig(tempDir, content string) string {
	configPath := filepath.Join(tempDir, ".golangci.yml")
	Expect(os.WriteFile(configPath, []byte(content), 0o644)).To(Succeed())

	return configPath
}

// writeMinimalTestConfig writes the standard minimal test config and returns the path.
func writeMinimalTestConfig(tempDir string) string {
	return writeTestConfig(tempDir, minimalConfigContent)
}

// testConfigCommand creates a test config file, runs a CLI command, and returns the output.
func testConfigCommand(tempDir, command, configContent string) ([]byte, error) {
	configPath := writeTestConfig(tempDir, configContent)
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
		t.Skip(
			"CLI binary not found, run 'go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure' first",
		)
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
			output, err := runCLI(
				"configure",
				"--config",
				writeMinimalTestConfig(tempDir),
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

	// analyzeWithFormat runs analyze with a format and verifies the output contains expected substrings.
	analyzeWithFormat := func(format string, expected ...string) {
		configPath := writeMinimalTestConfig(tempDir)
		output, err := runCLI("analyze", "--config", configPath, "--format", format)
		Expect(err).NotTo(HaveOccurred())
		for _, e := range expected {
			Expect(string(output)).To(ContainSubstring(e))
		}
	}

	testStandardCommandContext(tempDir, "analyze")
	testStandardCommandContext(tempDir, "validate")

	Context("Analyze command", func() {
		It("should analyze with JSON output", func() {
			analyzeWithFormat("json", "recommendations")
		})

		It("should emit PascalCase JSON keys", func() {
			configPath := writeMinimalTestConfig(tempDir)
			output, err := runCLI("analyze", "--config", configPath, "--format", "json")
			Expect(err).NotTo(HaveOccurred())

			var raw map[string]any
			Expect(json.Unmarshal(output, &raw)).To(Succeed())

			Expect(raw).To(HaveKey("ConfigPath"))
			Expect(raw).To(HaveKey("EnabledLinters"))
			Expect(raw).To(HaveKey("DisabledLinters"))
			Expect(raw).To(HaveKey("CriticalCount"))
			Expect(raw).To(HaveKey("HighValueCount"))

			Expect(raw).NotTo(HaveKey("config_path"))
			Expect(raw).NotTo(HaveKey("enabled_linters"))
			Expect(raw).NotTo(HaveKey("critical_count"))
		})

		It("should analyze with SARIF output", func() {
			analyzeWithFormat("sarif", "$schema")
		})
	})

	Context("Validate command", func() {
		It("should validate a valid config", func() {
			configPath := writeMinimalTestConfig(tempDir)

			output, err := runCLI("validate", "--config", configPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("valid"))
		})

		It("should reject an invalid config", func() {
			configPath := writeTestConfig(tempDir, "version: \"1\"\n")

			_, err := runCLI("validate", "--config", configPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Report command", func() {
		It("should generate JSON report", func() {
			configPath := writeMinimalTestConfig(tempDir)

			reportPath := filepath.Join(tempDir, "report.json")
			output, err := runCLI("report", "--config", configPath, "--format", "json", "--output", reportPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("Report generated"))

			_, err = os.Stat(reportPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should emit PascalCase JSON keys in report file", func() {
			configPath := writeMinimalTestConfig(tempDir)

			reportPath := filepath.Join(tempDir, "report.json")
			_, err := runCLI("report", "--config", configPath, "--format", "json", "--output", reportPath)
			Expect(err).NotTo(HaveOccurred())

			data, readErr := os.ReadFile(reportPath)
			Expect(readErr).NotTo(HaveOccurred())

			var raw map[string]any
			Expect(json.Unmarshal(data, &raw)).To(Succeed())

			Expect(raw).To(HaveKey("ConfigPath"))
			Expect(raw).To(HaveKey("Summary"))
			Expect(raw).To(HaveKey("EnabledLinters"))

			Expect(raw).NotTo(HaveKey("config_path"))
			Expect(raw).NotTo(HaveKey("enabled_linters"))
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
			configPath := writeMinimalTestConfig(tempDir)

			_, err := runCLI("configure", "--config", configPath, "--priority", "critical", "--check")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Diff mode", func() {
		It("should show diff for changes", func() {
			configPath := writeMinimalTestConfig(tempDir)

			output, err := runCLI("configure", "--config", configPath, "--priority", "critical", "--diff")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("Added"))
		})
	})
})
