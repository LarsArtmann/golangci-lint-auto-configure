package cli_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("completion command", func() {
	It("should generate bash completion script", func() {
		assertHelpContains(buildBinary(), "completion", "bash")
	})

	It("should generate zsh completion script", func() {
		binaryPath := buildBinary()
		cmd := exec.Command(binaryPath, "completion", "zsh")
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("zsh"))
	})
})
