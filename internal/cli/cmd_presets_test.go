package cli_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("configure presets", func() {
	DescribeTable(
		"should apply preset successfully",
		func(presetName string) {
			initGitRepo()

			binaryPath := buildBinary()
			configPath := writeConfig(testConfigContentMinimal)

			cmd := exec.Command(
				binaryPath,
				"configure",
				"--config",
				configPath,
				"--preset",
				presetName,
			)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			content, err := os.ReadFile(configPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("version:"))
		},
		Entry("security preset", "security"),
		Entry("strict preset", "strict"),
		Entry("performance preset", "performance"),
		Entry("reference preset", "reference"),
		Entry("minimal preset", "minimal"),
		Entry("standard preset", "standard"),
		Entry("format preset", "format"),
	)

	It("should apply security preset with only gosec", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		cmd := exec.Command(
			binaryPath,
			"configure",
			"--config",
			configPath,
			"--preset",
			"security",
		)
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))

		content, err := os.ReadFile(configPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("gosec"))
	})
})
