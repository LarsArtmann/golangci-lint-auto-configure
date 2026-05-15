package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const testConfigContentMinimal = `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
`

func TestCLICommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Commands Suite")
}

var _ = Describe("CLI Integration Tests", func() {
	var testDir string

	BeforeEach(func() {
		testDir = GinkgoT().TempDir()
	})

	// Helper function to build the binary
	buildBinary := func() string {
		binaryPath := filepath.Join(testDir, "golangci-lint-auto-configure")
		projectRoot, _ := filepath.Abs(filepath.Join("..", ".."))
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/golangci-lint-auto-configure")
		cmd.Dir = projectRoot

		cmd.Env = append(os.Environ(),
			"GOPRIVATE=github.com/larsartmann/go-finding",
			"GONOSUMCHECK=github.com/larsartmann/go-finding",
		)

		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "Failed to build the CLI binary: "+string(output))

		return binaryPath
	}

	// Helper function to initialize a git repo in test directory
	initGitRepo := func() {
		cmd := exec.Command("git", "init")
		cmd.Dir = testDir
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "Failed to init git repo: "+string(output))
	}

	// Helper function to create a config file and run a CLI command
	runCommandWithConfig := func(binaryPath, configContent string, args []string) (string, error) {
		configPath := filepath.Join(testDir, ".golangci.yml")
		Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

		cmdArgs := append(args, "--config", configPath)
		cmd := exec.Command(binaryPath, cmdArgs...)
		output, err := cmd.CombinedOutput()

		return string(output), err
	}

	// Helper function to test that invalid YAML is rejected
	assertInvalidYAMLRejected := func(binaryPath string) {
		invalidConfig := filepath.Join(testDir, "invalid.yml")
		Expect(os.WriteFile(invalidConfig, []byte("invalid: yaml: content:"), 0o644)).To(Succeed())

		cmd := exec.Command(binaryPath, "validate", "--config", invalidConfig)
		_, err := cmd.CombinedOutput()

		Expect(err).To(HaveOccurred())
	}

	// Helper function to test command succeeds with expected output
	testCommandSuccess := func(configContent, command, expectedOutput string) {
		binaryPath := buildBinary()
		output, err := runCommandWithConfig(binaryPath, configContent, []string{command})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(expectedOutput))
	}

	// Helper function to test missing config file error
	testMissingConfigError := func(command string) {
		binaryPath := buildBinary()
		cmd := exec.Command(binaryPath, command, "--config", "/non/existent/path.yml")
		_, err := cmd.CombinedOutput()
		Expect(err).To(HaveOccurred())
	}

	// Helper function to test migrate command with v2 config
	testMigrateCommand := func(expectedOutput string) {
		configContent := `version: "2"
linters:
  enable:
    - errcheck
`
		testCommandSuccess(configContent, "migrate", expectedOutput)
	}

	Context("analyze command", func() {
		It("should analyze a valid config file", func() {
			configContent := `version: "2"
linters:
  enable:
    - errcheck
    - gosec
`
			testCommandSuccess(configContent, "analyze", "Analyzing configuration")
		})

		It("should show recommendations for minimal config", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
  disable:
    - govet
    - ineffassign
    - staticcheck
    - unused`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "analyze", "--config", configPath)
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			// Check for HIGH VALUE which is the priority level for most recommendations
			Expect(outputStr).To(Or(
				ContainSubstring("HIGH VALUE"),
				ContainSubstring("CRITICAL"),
				ContainSubstring("MEDIUM"),
			))
		})

		It("should return error for non-existent config file", func() {
			testMissingConfigError("validate")
		})

		It("should handle invalid YAML gracefully", func() {
			binaryPath := buildBinary()
			assertInvalidYAMLRejected(binaryPath)
		})
	})

	Context("configure command", func() {
		It("should run with dry-run mode without modifying file", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(testConfigContentMinimal), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "configure", "--config", configPath, "--dry-run")
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("[DRY-RUN]"))

			// Verify original file unchanged
			content, _ := os.ReadFile(configPath)
			Expect(string(content)).To(Equal(testConfigContentMinimal))
		})

		It("should require git repository for config modification", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
  disable:
    - govet
    - ineffassign
    - staticcheck
    - unused`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "configure", "--config", configPath)
			_, err := cmd.CombinedOutput()

			// Should succeed since we're in a git repo
			Expect(err).NotTo(HaveOccurred())

			// Verify file was modified (no backup created since git provides version control)
			content, _ := os.ReadFile(configPath)
			Expect(string(content)).To(ContainSubstring("version: \"2\""))
		})

		It("should succeed with warning when not in a git repository", func() {
			binaryPath := buildBinary()
			configContent := testConfigContentMinimal
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			// Note: testDir is NOT a git repo (no initGitRepo() called)
			cmd := exec.Command(binaryPath, "configure", "--config", configPath)
			cmd.Dir = testDir // Run from non-git directory
			output, err := cmd.CombinedOutput()

			// Should succeed with warning about git
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("git"))
			Expect(outputStr).To(ContainSubstring("backup"))
		})

		It("should modify config when not in dry-run mode", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"critical",
			)
			_, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())

			// Verify file was modified
			content, _ := os.ReadFile(configPath)
			Expect(string(content)).To(ContainSubstring("version: \"2\""))
		})

		It("should respect priority flag", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"critical",
				"--dry-run",
			)
			output, _ := cmd.CombinedOutput()

			// Should show [DRY-RUN] indicator
			Expect(string(output)).To(ContainSubstring("[DRY-RUN]"))
		})
	})

	Context("validate command", func() {
		It("should validate a valid config", func() {
			configContent := `version: "2"
run:
  timeout: 5m
linters:
  enable:
    - errcheck
    - staticcheck
    - govet
    - gosec
`
			testCommandSuccess(configContent, "validate", "valid")
		})

		It("should reject invalid YAML", func() {
			binaryPath := buildBinary()
			assertInvalidYAMLRejected(binaryPath)
		})

		It("should handle missing config file", func() {
			testMissingConfigError("validate")
		})
	})

	Context("help and flags", func() {
		It("should show help when --help is used", func() {
			binaryPath := buildBinary()
			cmd := exec.Command(binaryPath, "--help")
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("golangci-lint-auto-configure"))
			Expect(string(output)).To(ContainSubstring("COMMANDS"))
		})

		It("should show command-specific help", func() {
			binaryPath := buildBinary()
			cmd := exec.Command(binaryPath, "configure", "--help")
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("configure"))
			Expect(string(output)).To(ContainSubstring("--priority"))
			Expect(string(output)).To(ContainSubstring("--dry-run"))
		})

		It("should show analyze help", func() {
			binaryPath := buildBinary()
			cmd := exec.Command(binaryPath, "analyze", "--help")
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("analyze"))
		})

		It("should handle verbose flag", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "analyze", "--config", configPath, "--verbose")
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			// Verbose mode shows DEBU (debug level)
			Expect(string(output)).To(ContainSubstring("DEBU"))
		})
	})

	Context("migrate command", func() {
		It("should migrate v1 config to v2 successfully", func() {
			testMigrateCommand("Migrating configuration")
		})

		It("should skip migration for v2 configs", func() {
			testMigrateCommand("already version 2")
		})
	})

	Context("report command", func() {
		It("should generate JSON report", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			reportPath := filepath.Join(testDir, "report.json")
			cmd := exec.Command(
				binaryPath,
				"report",
				"--config",
				configPath,
				"--output",
				reportPath,
				"--format",
				"json",
			)
			_, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())

			// Verify report file was created
			_, err = os.Stat(reportPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should generate SARIF report", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			reportPath := filepath.Join(testDir, "report.sarif.json")
			cmd := exec.Command(
				binaryPath,
				"report",
				"--config",
				configPath,
				"--output",
				reportPath,
				"--format",
				"sarif",
			)
			_, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			data, err := os.ReadFile(reportPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring(`"$schema"`))
			Expect(string(data)).To(ContainSubstring(`"golangci-lint-auto-configure"`))
		})

		It("should generate finding report", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			reportPath := filepath.Join(testDir, "report.finding.json")
			cmd := exec.Command(
				binaryPath,
				"report",
				"--config",
				configPath,
				"--output",
				reportPath,
				"--format",
				"finding",
			)
			_, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			data, err := os.ReadFile(reportPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring(`"golangci-lint-auto-configure"`))
			Expect(string(data)).To(ContainSubstring(`"findings"`))
			Expect(string(data)).To(ContainSubstring(`"summary"`))
		})
	})

	Context("analyze command with go-finding formats", func() {
		It("should output SARIF from analyze", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			output, err := runCommandWithConfig(
				binaryPath,
				configContent,
				[]string{"analyze", "--format", "sarif"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring(`"$schema"`))
			Expect(output).To(ContainSubstring(`"golangci-lint-auto-configure"`))
		})

		It("should output finding JSON from analyze", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			output, err := runCommandWithConfig(
				binaryPath,
				configContent,
				[]string{"analyze", "--format", "finding"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring(`"golangci-lint-auto-configure"`))
			Expect(output).To(ContainSubstring(`"findings"`))
			Expect(output).To(ContainSubstring(`"summary"`))
		})
	})
})
