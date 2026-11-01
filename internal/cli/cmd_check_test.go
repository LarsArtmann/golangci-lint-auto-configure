package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("check mode combinations", func() {
	// configureThenCheck runs configure with given flags, then runs configure --check with the same flags.
	configureThenCheck := func(binaryPath, configPath, flagName, flagValue string) {
		_, err := exec.Command(binaryPath, "configure", "--config", configPath, "--"+flagName, flagValue).
			CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		_, err = exec.Command(binaryPath, "configure", "--config", configPath, "--"+flagName, flagValue, "--check").
			CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
	}

	It("should exit 0 with --check --priority critical after configure", func() {
		initGitRepo()
		configureThenCheck(
			buildBinary(),
			writeConfig(testConfigContentMinimal),
			"priority",
			"critical",
		)
	})

	It("should not modify file with --check --dry-run", func() {
		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		cmd := exec.Command(
			binaryPath,
			"configure",
			"--config",
			configPath,
			"--priority",
			"high",
			"--check",
			"--dry-run",
		)
		_, _ = cmd.CombinedOutput()

		content, _ := os.ReadFile(configPath)
		Expect(string(content)).To(Equal(testConfigContentMinimal))
	})

	It("should exit 0 with --check --preset after configure", func() {
		initGitRepo()
		configureThenCheck(
			buildBinary(),
			writeConfig(testConfigContentMinimal),
			"preset",
			"minimal",
		)
	})

	It("should restore config after --check --diff", func() {
		initGitRepo()

		binaryPath := buildBinary()
		configPath := writeConfig(testConfigContentMinimal)

		_, err := exec.Command(
			binaryPath,
			"configure",
			"--config", configPath,
			"--priority", "high",
			"--check",
			"--diff",
		).CombinedOutput()

		Expect(err).To(HaveOccurred(), "--check with fixes needed should exit non-zero")

		content, _ := os.ReadFile(configPath)
		Expect(string(content)).To(ContainSubstring("errcheck"),
			"config must retain original linters after --check --diff")
		Expect(string(content)).To(ContainSubstring("version:"),
			"config must still be valid YAML after --check --diff")
	})

	It("should not write audit ledger entries with --check --diff", func() {
		initGitRepo()
		binaryPath := buildBinary()

		runInCache := func(cacheDir, configContent string, args ...string) string {
			configPath := writeConfig(configContent)
			cmdArgs := append([]string{"configure", "--config", configPath}, args...)
			cmd := exec.Command(binaryPath, cmdArgs...)
			cmd.Env = append(os.Environ(), "XDG_CACHE_HOME="+cacheDir)
			_, _ = cmd.CombinedOutput()

			data, err := os.ReadFile(filepath.Join(
				cacheDir, "golangci-lint-auto-configure", "audit.jsonl"))
			if err != nil {
				return ""
			}

			return strings.TrimSpace(string(data))
		}

		control := runInCache(GinkgoT().TempDir(), testConfigContentMinimal, "--priority", "high")
		Expect(control).NotTo(BeEmpty(),
			"control: a real apply run must produce audit entries (isolates the ledger mechanism)")

		checkDiff := runInCache(GinkgoT().TempDir(), testConfigContentMinimal,
			"--priority", "high", "--check", "--diff")
		Expect(checkDiff).To(BeEmpty(),
			"--check --diff must not emit audit entries: its writes are rolled back, "+
				"phantom applied-changes would poison regression-loop detection")
	})
})
