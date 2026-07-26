package cli_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testDir string

var _ = BeforeEach(func() {
	testDir = GinkgoT().TempDir()
})

const testConfigContentMinimal = `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`

func writeConfig(configContent string) string {
	configPath := filepath.Join(testDir, ".golangci.yml")
	Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

	return configPath
}

// Helper function to build the binary.
func buildBinary() string {
	binaryPath := filepath.Join(testDir, "golangci-lint-auto-configure")
	projectRoot, _ := filepath.Abs(filepath.Join("..", ".."))
	cmd := exec.CommandContext(
		context.Background(),
		"go",
		"build",
		"-o",
		binaryPath,
		"./cmd/golangci-lint-auto-configure",
	)
	cmd.Dir = projectRoot

	cmd.Env = append(
		os.Environ(),
		"GOPRIVATE=github.com/larsartmann/go-finding",
		"GONOSUMCHECK=github.com/larsartmann/go-finding",
	)

	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "Failed to build the CLI binary: "+string(output))

	return binaryPath
}

// Helper function to initialize a git repo in test directory.
func initGitRepo() {
	cmd := exec.CommandContext(context.Background(), "git", "init")
	cmd.Dir = testDir
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "Failed to init git repo: "+string(output))
}

// Helper function to create a config file and run a CLI command.
func runCommandWithConfig(binaryPath, configContent string, args []string) (string, error) {
	configPath := filepath.Join(testDir, ".golangci.yml")
	Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

	cmdArgs := make([]string, 0, len(args)+2)
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--config", configPath)
	cmd := exec.CommandContext(context.Background(), binaryPath, cmdArgs...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

// Helper function to test that invalid YAML is rejected by a given command.
func assertInvalidYAMLRejectedBy(binaryPath, command string) {
	invalidPath := filepath.Join(testDir, "invalid.yml")
	Expect(os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0o644)).To(Succeed())

	cmd := exec.CommandContext(context.Background(), binaryPath, command, "--config", invalidPath)
	_, err := cmd.CombinedOutput()

	Expect(err).To(HaveOccurred())
}

// Helper function to test that a command help output contains expected text.
func assertHelpContains(binaryPath, subcommand, expected string) {
	cmd := exec.CommandContext(context.Background(), binaryPath, subcommand, "--help")
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred())
	Expect(string(output)).To(ContainSubstring(expected))
}

// Helper function to generate a report and return its file content.
func generateReport(binaryPath, configPath, reportExt, format string) string {
	reportPath := filepath.Join(testDir, "report."+reportExt)
	cmd := exec.CommandContext(
		context.Background(),
		binaryPath,
		"report",
		"--config",
		configPath,
		"--output",
		reportPath,
		"--format",
		format,
	)
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))

	data, err := os.ReadFile(reportPath)
	Expect(err).NotTo(HaveOccurred())

	return string(data)
}

// Helper function to test command succeeds with expected output.
func testCommandSuccess(configContent, command, expectedOutput string) {
	binaryPath := buildBinary()
	output, err := runCommandWithConfig(binaryPath, configContent, []string{command})
	Expect(err).NotTo(HaveOccurred())
	Expect(output).To(ContainSubstring(expectedOutput))
}

// Helper function to test missing config file error.
func testMissingConfigError(command string) {
	binaryPath := buildBinary()
	cmd := exec.CommandContext(
		context.Background(),
		binaryPath,
		command,
		"--config",
		"/non/existent/path.yml",
	)
	_, err := cmd.CombinedOutput()
	Expect(err).To(HaveOccurred())
}

// Helper function to test migrate command with v2 config.
func testMigrateCommand(expectedOutput string) {
	configContent := `version: "2"
linters:
  enable:
    - errcheck
`
	testCommandSuccess(configContent, "migrate", expectedOutput)
}
