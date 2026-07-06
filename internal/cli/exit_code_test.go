package cli_test

import (
	"errors"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("exit codes", func() {
	It("should exit 0 on successful configure", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		err := exec.Command(binaryPath, "configure", "--config", configPath, "--priority", "critical").Run()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should exit non-zero with invalid priority", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		err := exec.Command(binaryPath, "configure", "--config", configPath, "--priority", "invalid-priority").Run()
		Expect(err).To(HaveOccurred())

		var exitErr *exec.ExitError
		Expect(errors.As(err, &exitErr)).To(BeTrue())
		Expect(exitErr.ExitCode()).To(Equal(75))
	})

	It("should output JSON error with --json-errors flag", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		output, err := exec.Command(
			binaryPath, "configure", "--config", configPath, "--priority", "invalid-priority", "--json-errors",
		).CombinedOutput()

		Expect(err).To(HaveOccurred())

		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring(`"message"`))
		Expect(outputStr).To(ContainSubstring(`"family"`))
		Expect(outputStr).To(ContainSubstring(`"exit_code"`))
	})
})
