package cli_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("migrate command", func() {
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
		assertInvalidYAMLRejectedBy(buildBinary(), "migrate")
	})
})
