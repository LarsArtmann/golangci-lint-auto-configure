package cli_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// readPresetLedger runs configure --preset with an isolated XDG_CACHE_HOME and
// returns the raw audit-ledger content ("" when no ledger file exists).
func readPresetLedger(extraArgs ...string) string {
	initGitRepo()

	cacheDir := GinkgoT().TempDir()
	configPath := writeConfig(testConfigContentMinimal)

	cmdArgs := append(
		[]string{"configure", "--config", configPath, "--preset", "minimal"},
		extraArgs...)
	cmd := exec.CommandContext(context.Background(), buildBinary(), cmdArgs...)

	cmd.Env = append(os.Environ(), "XDG_CACHE_HOME="+cacheDir)
	_, _ = cmd.CombinedOutput()

	data, err := os.ReadFile(filepath.Join(
		cacheDir, "golangci-lint-auto-configure", "audit.jsonl"))
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

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

	It("records a preset-applied ledger entry on a real preset run", func() {
		ledger := readPresetLedger()

		Expect(ledger).To(ContainSubstring(`"action":"preset-applied"`))
		Expect(ledger).To(ContainSubstring(`"linter":"minimal"`),
			"the preset names must ride in the Linter field")
	})

	It("records no preset-applied entry on a --check preset run", func() {
		ledger := readPresetLedger("--check")

		Expect(ledger).NotTo(ContainSubstring("preset-applied"),
			"check-mode preset runs mutate nothing and must record nothing")
	})
})
