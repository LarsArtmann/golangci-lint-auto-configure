package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("install-hook command", func() {
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
