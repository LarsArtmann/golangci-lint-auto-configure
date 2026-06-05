package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("analyze command with go-finding formats", func() {
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
