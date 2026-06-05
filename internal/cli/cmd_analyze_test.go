package cli_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("analyze command", func() {
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
		assertInvalidYAMLRejectedBy(buildBinary(), "validate")
	})
})
