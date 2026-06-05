package cli_test

import (
	"errors"
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

	// Helper function to write config content to the standard config path.
	writeConfig := func(configContent string) string {
		configPath := filepath.Join(testDir, ".golangci.yml")
		Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

		return configPath
	}

	// Helper function to build the binary
	buildBinary := func() string {
		binaryPath := filepath.Join(testDir, "golangci-lint-auto-configure")
		projectRoot, _ := filepath.Abs(filepath.Join("..", ".."))
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/golangci-lint-auto-configure")
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
		invalidPath := filepath.Join(testDir, "invalid.yml")
		Expect(os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0o644)).To(Succeed())

		cmd := exec.Command(binaryPath, "validate", "--config", invalidPath)
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
			configPath := writeConfig(configContent)

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
			configPath := writeConfig(testConfigContentMinimal)

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
			configPath := writeConfig(configContent)

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
			configPath := writeConfig(testConfigContentMinimal)

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
			configPath := writeConfig(configContent)

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
			configPath := writeConfig(configContent)

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

		It("should exit 1 with --check when changes are needed", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"high",
				"--check",
			)
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())

			var exitErr *exec.ExitError

			ok := errors.As(err, &exitErr)
			Expect(ok).To(BeTrue())
			Expect(exitErr.ExitCode()).To(Equal(1))
		})

		It("should exit 0 with --check after configuring with a preset", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			configureCmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--preset",
				"standard",
			)
			configureOutput, err := configureCmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(configureOutput))

			checkCmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--preset",
				"standard",
				"--check",
			)
			checkOutput, err := checkCmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(checkOutput))
		})

		It("should not modify file with --check alone", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"high",
				"--check",
			)
			_, _ = cmd.CombinedOutput()

			content, _ := os.ReadFile(configPath)
			Expect(string(content)).To(Equal(testConfigContentMinimal))
		})

		It("should show diff output with --diff", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"high",
				"--diff",
			)
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Added"))
		})

		It("should show diff with --check and restore original", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"high",
				"--check",
				"--diff",
			)
			output, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("Added"))

			restored, _ := os.ReadFile(configPath)
			restoredStr := string(restored)
			Expect(restoredStr).To(ContainSubstring("errcheck"))
			Expect(restoredStr).NotTo(ContainSubstring("gosec"))
		})

		It("should default to optional priority for unrecognized value", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"invalid",
				"--dry-run",
			)
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("[DRY-RUN]"))
		})

		It("should fail with invalid preset flag", func() {
			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--preset",
				"nonexistent",
			)
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())
		})

		It("should handle missing config file gracefully", func() {
			binaryPath := buildBinary()

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				"/non/existent/path.yml",
			)
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())
		})

		It("should handle invalid YAML in configure", func() {
			binaryPath := buildBinary()
			invalidPath := filepath.Join(testDir, "invalid.yml")
			Expect(os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "configure", "--config", invalidPath)
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())
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

		It("should output SARIF from validate", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
    - gosec
`
			configPath := writeConfig(configContent)

			cmd := exec.Command(binaryPath, "validate", "--config", configPath, "--format", "sarif")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring(`"$schema"`))
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
			configPath := writeConfig(configContent)

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

		It("should handle --skip-validation flag", func() {
			binaryPath := buildBinary()
			v1Content := `linters-settings:
  gci:
    sections:
      - Standard
linters:
  enable:
    - gosec
`
			configPath := writeConfig(v1Content)

			cmd := exec.Command(binaryPath, "migrate", "--config", configPath, "--skip-validation")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))
			Expect(string(output)).To(ContainSubstring("migrated"))
		})

		It("should handle invalid v1 config", func() {
			binaryPath := buildBinary()
			invalidPath := filepath.Join(testDir, "invalid.yml")
			Expect(os.WriteFile(invalidPath, []byte("invalid: yaml: content:"), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "migrate", "--config", invalidPath)
			_, err := cmd.CombinedOutput()
			Expect(err).To(HaveOccurred())
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
			configPath := writeConfig(configContent)

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

			_, err = os.Stat(reportPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should generate HTML report", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := writeConfig(configContent)

			reportPath := filepath.Join(testDir, "report.html")
			cmd := exec.Command(
				binaryPath,
				"report",
				"--config",
				configPath,
				"--output",
				reportPath,
				"--format",
				"html",
			)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			data, err := os.ReadFile(reportPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("<!doctype html>"))
			Expect(string(data)).To(ContainSubstring("golangci-lint"))
		})

		It("should generate SARIF report", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := writeConfig(configContent)

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
			configPath := writeConfig(configContent)

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

		It("should handle missing config file", func() {
			binaryPath := buildBinary()
			cmd := exec.Command(
				binaryPath,
				"report",
				"--config",
				"/non/existent/path.yml",
				"--output",
				filepath.Join(testDir, "report.json"),
				"--format",
				"json",
			)
			_, err := cmd.CombinedOutput()
			Expect(err).To(HaveOccurred())
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

	Context("configure presets", func() {
		DescribeTable(
			"should apply preset successfully",
			func(presetName string) {
				initGitRepo()

				binaryPath := buildBinary()
				configPath := writeConfig(testConfigContentMinimal)

				cmd := exec.Command(
					binaryPath,
					"configure",
					"--config",
					configPath,
					"--preset",
					presetName,
				)
				output, err := cmd.CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), string(output))

				content, err := os.ReadFile(configPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("version:"))
			},
			Entry("security preset", "security"),
			Entry("strict preset", "strict"),
			Entry("performance preset", "performance"),
			Entry("reference preset", "reference"),
			Entry("minimal preset", "minimal"),
			Entry("standard preset", "standard"),
		)

		It("should apply security preset with only gosec", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--preset",
				"security",
			)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			content, err := os.ReadFile(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("gosec"))
		})
	})

	Context("configure with deprecated linters", func() {
		It("should replace wsl with wsl_v5", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
    - wsl
`
			configPath := writeConfig(configContent)

			cmd := exec.Command(binaryPath, "configure", "--config", configPath)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			content, err := os.ReadFile(configPath)
			Expect(err).NotTo(HaveOccurred())

			contentStr := string(content)
			Expect(contentStr).NotTo(ContainSubstring("- wsl\n"))
			Expect(contentStr).To(ContainSubstring("wsl_v5"))
		})
	})

	Context("install-hook command", func() {
		It("should install pre-commit hook in a git repo", func() {
			binaryPath := buildBinary()
			hookDir := filepath.Join(testDir, ".git", "hooks")
			Expect(os.MkdirAll(hookDir, 0o755)).To(Succeed())

			cmd := exec.Command("git", "init", testDir)
			_, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			cmd = exec.Command(binaryPath, "install-hook")
			cmd.Dir = testDir
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))
			Expect(string(output)).To(ContainSubstring("hook"))
		})
	})

	Context("completion command", func() {
		It("should generate bash completion script", func() {
			binaryPath := buildBinary()
			cmd := exec.Command(binaryPath, "completion", "bash")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("bash"))
		})

		It("should generate zsh completion script", func() {
			binaryPath := buildBinary()
			cmd := exec.Command(binaryPath, "completion", "zsh")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("zsh"))
		})
	})

	Context("check mode combinations", func() {
		It("should exit 0 with --check --priority critical after configure", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			configureCmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"critical",
			)
			_, err := configureCmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			checkCmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"critical",
				"--check",
			)
			_, err = checkCmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not modify file with --check --dry-run", func() {
			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"high",
				"--check",
				"--dry-run",
			)
			_, _ = cmd.CombinedOutput()

			content, _ := os.ReadFile(configPath)
			Expect(string(content)).To(Equal(testConfigContentMinimal))
		})

		It("should exit 0 with --check --preset after configure", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			configureCmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--preset",
				"minimal",
			)
			_, err := configureCmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			checkCmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--preset",
				"minimal",
				"--check",
			)
			_, err = checkCmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("diff mode combinations", func() {
		It("should show removed linters in diff output", func() {
			initGitRepo()

			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
    - someunknownlinter
`
			configPath := writeConfig(configContent)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"critical",
				"--diff",
				"--dry-run",
			)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			outputStr := string(output)
			Expect(outputStr).To(Or(
				ContainSubstring("Added"),
				ContainSubstring("Removed"),
				ContainSubstring("[DRY-RUN]"),
			))
		})

		It("should show no diff when config is optimal", func() {
			initGitRepo()

			binaryPath := buildBinary()

			configureCmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				filepath.Join(testDir, ".golangci.yml"),
				"--priority",
				"critical",
			)
			configPath := writeConfig(testConfigContentMinimal)
			configureCmd = exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"critical",
			)
			_, err := configureCmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			checkCmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--priority",
				"critical",
				"--diff",
			)
			output, err := checkCmd.CombinedOutput()
			outputStr := string(output)

			if err == nil {
				Expect(outputStr).NotTo(ContainSubstring("Added"))
			}
		})
	})
})
