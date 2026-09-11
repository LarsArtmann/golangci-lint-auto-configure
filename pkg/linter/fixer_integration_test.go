package linter_test

import (
	"context"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs exercise the full public-API FixConfig flow with a REAL policy
// sidecar and a REAL audit ledger file on disk — the end-to-end story that
// unit-level doubles (fixer_enforce_test.go) cannot prove: that enforcement,
// never-enable suppression, and ledger persistence compose correctly.
var _ = Describe("FixConfig sidecar + ledger integration", func() {
	var (
		logger     *log.Logger
		fixer      *linter.Fixer
		testDir    string
		configPath string
		ledgerPath string
	)

	const unjustifiedDisables = `version: "2"
run:
  timeout: 5m
linters:
  disable:
    - godot
    - misspell
`

	BeforeEach(func() {
		logger = log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		analyzer := linter.NewAnalyzer(logger)
		configLoader := config.NewLoader(logger)
		fixer = linter.NewFixer(logger, analyzer, configLoader)

		testDir = GinkgoT().TempDir()
		configPath = filepath.Join(testDir, ".golangci.yml")
		ledgerPath = filepath.Join(testDir, "audit", "audit.jsonl")
	})

	writeLedger := func() {
		ledger := audit.NewLedger(logger, audit.RunContext{
			RunID:    "integration-test-run",
			RepoHash: "integrationtest",
			RepoPath: testDir,
		}, ledgerPath)

		fixer.SetLedger(ledger)
	}

	writeSidecar := func(content string) {
		Expect(os.WriteFile(filepath.Join(testDir, ".golangci-lint-auto-configure.yml"), []byte(content), 0o644)).To(Succeed())
	}

	It("re-enables unjustified disables when a sidecar exists, and records it", func() {
		Expect(os.WriteFile(configPath, []byte(unjustifiedDisables), 0o644)).To(Succeed())

		// An empty sidecar activates enforcement: every disable without a
		// justification entry becomes unjustified.
		writeSidecar("disabled: {}\n")

		writeLedger()

		result, err := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityHigh, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.IsSuccess()).To(BeTrue())

		fixed, readErr := os.ReadFile(configPath)
		Expect(readErr).NotTo(HaveOccurred())
		// Enforcement re-enabled both unjustified disables.
		Expect(string(fixed)).NotTo(ContainSubstring("misspell"))
		Expect(string(fixed)).NotTo(ContainSubstring("godot"))

		ledgerContents, readErr := os.ReadFile(ledgerPath)
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(ledgerContents)).To(ContainSubstring(`"action":"re-enabled"`))
		Expect(string(ledgerContents)).To(ContainSubstring(`"linter":"misspell"`))
	})

	It("never-enable blocks enforcement and the suppression lands in the ledger", func() {
		Expect(os.WriteFile(configPath, []byte(unjustifiedDisables), 0o644)).To(Succeed())

		writeSidecar(`never-enable:
  godot:
    reason: demands per-package godoc; this repo documents per-file
    category: convention
`)

		writeLedger()

		result, err := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityHigh, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.IsSuccess()).To(BeTrue())

		fixed, readErr := os.ReadFile(configPath)
		Expect(readErr).NotTo(HaveOccurred())
		// godot stays disabled (never-enable wins over enforcement)...
		Expect(string(fixed)).To(ContainSubstring("godot"))
		// ...while the plain unjustified disable is still enforced.
		Expect(string(fixed)).NotTo(ContainSubstring("misspell"))

		ledgerContents, readErr := os.ReadFile(ledgerPath)
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(ledgerContents)).To(ContainSubstring(`"action":"suppressed-re-enable"`))
		Expect(string(ledgerContents)).To(ContainSubstring(`"linter":"godot"`))
	})

	It("respects all disables when no sidecar exists (backward compat)", func() {
		Expect(os.WriteFile(configPath, []byte(unjustifiedDisables), 0o644)).To(Succeed())

		writeLedger()

		result, err := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityHigh, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.IsSuccess()).To(BeTrue())

		fixed, readErr := os.ReadFile(configPath)
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(fixed)).To(ContainSubstring("godot"))
		Expect(string(fixed)).To(ContainSubstring("misspell"))
	})
})
