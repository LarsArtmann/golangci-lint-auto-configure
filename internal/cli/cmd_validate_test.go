package cli_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("validate command", func() {
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
		assertInvalidYAMLRejectedBy(buildBinary(), "validate")
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

var _ = Context("help and flags", func() {
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
		assertHelpContains(buildBinary(), "analyze", "analyze")
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
