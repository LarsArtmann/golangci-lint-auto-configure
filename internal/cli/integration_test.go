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

// runCLI runs the CLI binary with the given arguments and returns output.
func runCLI(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CLITimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "../../../bin/golangci-lint-auto-configure", args...)
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

// testStandardConfigCommand runs a CLI command with a standard test config
func testStandardConfigCommand(tempDir, command string) {
	content := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - gosec
`
	output, err := testConfigCommand(tempDir, command, content)
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
})
