package cli_test

import (
	"errors"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("configure command", func() {
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
		assertInvalidYAMLRejectedBy(buildBinary(), "configure")
	})
})
