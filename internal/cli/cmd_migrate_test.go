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

// v1ConfigForMigration carries v1-schema keys (linters-settings) so the
// migrate command performs a real v1→v2 rewrite.
const v1ConfigForMigration = `linters-settings:
  gci:
    sections:
      - Standard
linters:
  enable:
    - gosec
`

// runMigrateInCache executes migrate against a fresh v1 config with an
// isolated XDG_CACHE_HOME and returns the audit-ledger lines it produced
// (nil when no ledger file exists).
func runMigrateInCache(cacheDir string, args ...string) []string {
	configPath := writeConfig(v1ConfigForMigration)

	cmdArgs := append([]string{"migrate", "--config", configPath}, args...)
	cmd := exec.CommandContext(context.Background(), buildBinary(), cmdArgs...)

	cmd.Env = append(os.Environ(), "XDG_CACHE_HOME="+cacheDir)
	_, _ = cmd.CombinedOutput()

	data, err := os.ReadFile(filepath.Join(
		cacheDir, "golangci-lint-auto-configure", "audit.jsonl"))
	if err != nil {
		return nil
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil
	}

	return strings.Split(trimmed, "\n")
}

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

	It("records exactly one config-migrated ledger entry on a real migration", func() {
		initGitRepo()

		lines := runMigrateInCache(GinkgoT().TempDir())

		Expect(lines).To(HaveLen(1),
			"a non-dry migrate must write exactly one whole-file ledger entry")
		Expect(lines[0]).To(ContainSubstring(`"action":"config-migrated"`))
		Expect(lines[0]).To(ContainSubstring(`"linter":"config"`))
		Expect(lines[0]).To(ContainSubstring("fixes applied"))
	})

	It("records no ledger entry on a dry-run migration", func() {
		initGitRepo()

		lines := runMigrateInCache(GinkgoT().TempDir(), "--dry-run")

		Expect(lines).To(BeEmpty(),
			"dry runs must not emit audit entries — nothing was mutated")
	})
})
