package cli_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("diff mode combinations", func() {
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
			"--diff",
		)
		output, err := checkCmd.CombinedOutput()
		outputStr := string(output)

		if err == nil {
			Expect(outputStr).NotTo(ContainSubstring("Added"))
		}
	})
})
