package linter_test

import (
	"context"
	"os"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// brokenGoconstConfig is what tool versions before the 2026-09-11 schema fix
// emitted: goconst.min-length is rejected by golangci-lint <2.13 ("configuration
// contains invalid elements") and the idempotency guard previously prevented
// self-heal on re-runs.
const brokenGoconstConfig = `version: "2"
run:
  timeout: 5m
linters:
  default: none
  enable:
    - goconst
  settings:
    goconst:
      min-length: 4
      min-occurrences: 5
      ignore-tests: true
`

var _ = Describe("Known-bad settings key normalization", func() {
	var (
		fixer      *linter.Fixer
		testConfig string
	)

	BeforeEach(func() {
		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		fixer = linter.NewFixer(logger, linter.NewAnalyzer(logger), config.NewLoader(logger))
		testConfig = filepath.Join(GinkgoT().TempDir(), ".golangci.yml")
	})

	Context("schema-invalid key emitted by older tool versions", func() {
		It("rewrites goconst.min-length to min-len on configure", func() {
			content, err := fixAndRead(
				fixer, testConfig, brokenGoconstConfig, types.LinterPriorityHigh, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("min-len: 4"))
			Expect(content).NotTo(ContainSubstring("min-length"))
		})

		It("preserves unrelated goconst keys and their values", func() {
			content, err := fixAndRead(
				fixer, testConfig, brokenGoconstConfig, types.LinterPriorityHigh, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("min-occurrences: 5"))
			Expect(content).To(ContainSubstring("ignore-tests: true"))
		})

		It("keeps the valid key when both bad and good keys exist", func() {
			content, err := fixAndRead(fixer, testConfig, `version: "2"
run:
  timeout: 5m
linters:
  default: none
  enable:
    - goconst
  settings:
    goconst:
      min-length: 9
      min-len: 2
`, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("min-len: 2"))
			Expect(content).NotTo(ContainSubstring("min-length"))
		})
	})

	Context("idempotency guard stays intact for valid configs", func() {
		It("does not overwrite an existing valid min-len value", func() {
			content, err := fixAndRead(fixer, testConfig, `version: "2"
run:
  timeout: 5m
linters:
  default: none
  enable:
    - goconst
  settings:
    goconst:
      min-len: 3
`, types.LinterPriorityHigh, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("min-len: 3"))
			Expect(content).NotTo(ContainSubstring("min-len: 4"))
		})

		It("reports 0 fixes on the run after normalization (no re-fire)", func() {
			content, err := fixAndRead(
				fixer, testConfig, brokenGoconstConfig, types.LinterPriorityHigh, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("min-len: 4"))

			writeConfig(testConfig, content)
			result, err := fixer.FixConfig(
				context.Background(), testConfig, types.LinterPriorityHigh, false,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.FixesApplied).To(Equal(0))
		})
	})
})
