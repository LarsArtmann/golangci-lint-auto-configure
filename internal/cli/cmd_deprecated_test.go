package cli_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("configure with deprecated linters", func() {
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
