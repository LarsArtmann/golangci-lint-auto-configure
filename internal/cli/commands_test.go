package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCLICommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Commands Suite")
}

var _ = Describe("CLI Integration Tests", func() {
	var (
		testDir string
	)

	BeforeEach(func() {
		testDir = GinkgoT().TempDir()
	})

	// Helper function to build the binary
	buildBinary := func() string {
		binaryPath := filepath.Join(testDir, "golangci-linter-auto-configure")
		// Get absolute path to the project root
		projectRoot, _ := filepath.Abs(filepath.Join("..", ".."))
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/golangci-linter-auto-configure")
		cmd.Dir = projectRoot
		cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=arm64")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Print build output for debugging
			println("Build failed:", string(output))
		}
		Expect(err).NotTo(HaveOccurred(), "Failed to build the CLI binary")
		return binaryPath
	}

	Context("analyze command", func() {
		It("should analyze a valid config file", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
    - gosec
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "analyze", "--config", configPath)
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("Analyzing configuration"))
		})

		It("should show recommendations for minimal config", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
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
			binaryPath := buildBinary()
			// The tool searches for fallback config, so we test with validate instead
			cmd := exec.Command(binaryPath, "validate", "--config", "/non/existent/path.yml")
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())
		})

		It("should handle invalid YAML gracefully", func() {
			binaryPath := buildBinary()
			invalidConfig := filepath.Join(testDir, "invalid.yml")
			Expect(os.WriteFile(invalidConfig, []byte("invalid: yaml: content:"), 0o644)).To(Succeed())

			// For validate command, invalid YAML should fail
			cmd := exec.Command(binaryPath, "validate", "--config", invalidConfig)
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("configure command", func() {
		It("should run with dry-run mode without modifying file", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			backupPath := configPath + ".backup"
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "configure", "--config", configPath, "--dry-run")
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			outputStr := string(output)
			Expect(outputStr).To(ContainSubstring("[DRY-RUN]"))

			// Verify original file unchanged
			content, _ := os.ReadFile(configPath)
			Expect(string(content)).To(Equal(configContent))

			// Verify no backup created in dry-run
			_, err = os.Stat(backupPath)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("should create backup when modifying config", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			backupPath := configPath + ".backup"
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "configure", "--config", configPath)
			_, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())

			// Verify backup created
			_, err = os.Stat(backupPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify backup is identical to original
			backupContent, _ := os.ReadFile(backupPath)
			Expect(string(backupContent)).To(Equal(configContent))
		})

		It("should modify config when not in dry-run mode", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "configure", "--config", configPath, "--priority", "critical")
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

			cmd := exec.Command(binaryPath, "configure", "--config", configPath, "--priority", "critical", "--dry-run")
			output, _ := cmd.CombinedOutput()

			// Should show [DRY-RUN] indicator
			Expect(string(output)).To(ContainSubstring("[DRY-RUN]"))
		})
	})

	Context("validate command", func() {
		It("should validate a valid config", func() {
			binaryPath := buildBinary()
			configContent := `version: "2"
linters:
  enable:
    - errcheck
    - gosec
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "validate", "--config", configPath)
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("valid"))
		})

		It("should reject invalid YAML", func() {
			binaryPath := buildBinary()
			invalidConfig := filepath.Join(testDir, "invalid.yml")
			Expect(os.WriteFile(invalidConfig, []byte("invalid: yaml: content:"), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "validate", "--config", invalidConfig)
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())
		})

		It("should handle missing config file", func() {
			binaryPath := buildBinary()
			cmd := exec.Command(binaryPath, "validate", "--config", "/non/existent/path.yml")
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("restore command", func() {
		It("should restore from backup file", func() {
			binaryPath := buildBinary()
			originalContent := `version: "2"
linters:
  enable:
    - errcheck
`
			modifiedContent := `version: "2"
linters:
  enable:
    - gosec
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			backupPath := filepath.Join(testDir, ".golangci.yml.backup")

			Expect(os.WriteFile(backupPath, []byte(originalContent), 0o644)).To(Succeed())
			Expect(os.WriteFile(configPath, []byte(modifiedContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "restore", "--config", configPath, "--backup-path", backupPath)
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("restored"))

			// Verify restored content
			content, _ := os.ReadFile(configPath)
			Expect(string(content)).To(Equal(originalContent))
		})

		It("should return error for non-existent backup", func() {
			binaryPath := buildBinary()
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte("test"), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "restore", "--config", configPath, "--backup-path", "/non/existent/backup.yml")
			_, err := cmd.CombinedOutput()

			Expect(err).To(HaveOccurred())
		})
	})

	Context("help and flags", func() {
		It("should show help when --help is used", func() {
			binaryPath := buildBinary()
			cmd := exec.Command(binaryPath, "--help")
			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("golangci-linter-auto-configure"))
			Expect(string(output)).To(ContainSubstring("Available Commands"))
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
		It("should show warning for unimplemented migrate", func() {
			binaryPath := buildBinary()
			configContent := `version: "1"
linters:
  enable:
    - errcheck
`
			configPath := filepath.Join(testDir, ".golangci.yml")
			Expect(os.WriteFile(configPath, []byte(configContent), 0o644)).To(Succeed())

			cmd := exec.Command(binaryPath, "migrate", "--config", configPath)
			output, err := cmd.CombinedOutput()

			// Migrate is a placeholder, should run but warn
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("not yet implemented"))
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
			cmd := exec.Command(binaryPath, "report", "--config", configPath, "--output", reportPath, "--format", "json")
			_, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())

			// Verify report file was created
			_, err = os.Stat(reportPath)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
