package linter_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Issue #15: the gohumanize recommendation gates on the dustin/go-humanize
// dependency, but the linter ships only as a golangci-lint module plugin — a
// stock binary rejects the enable entry and every subsequent lint run dies
// before analysis. The capability gate must withhold the recommendation, and
// a config already poisoned by an older tool version must be healed.
// issueConfig is the poisoned shape older tool versions wrote: a bare
// gohumanize enable entry a stock binary rejects.
const issueConfig = `version: "2"
linters:
  enable:
    - gohumanize
`

var _ = Describe("Fixer capability gate (issue #15)", func() {
	var (
		fixer      *linter.Fixer
		analyzer   *linter.Analyzer
		configPath string
	)

	BeforeEach(func() {
		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		analyzer = linter.NewAnalyzer(logger)

		known, err := analyzer.ProbeKnownLinters(context.Background())
		Expect(err).NotTo(HaveOccurred())

		if known["gohumanize"] {
			Skip("the environment's golangci-lint carries the gohumanize module plugin")
		}

		dir := GinkgoT().TempDir()

		goMod := "module example.com/humanized\n\ngo 1.25\n\n" +
			"require github.com/dustin/go-humanize v1.0.1\n"
		Expect(os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644)).To(Succeed())

		configPath = filepath.Join(dir, ".golangci.yml")
		fixer = linter.NewFixer(logger, analyzer, config.NewLoader(logger))
	})

	It("never re-adds gohumanize on a clean config (recommendation withheld)", func() {
		writeConfig(configPath, `version: "2"
linters:
  enable:
    - gosec
`)

		content, err := fixAndRead(fixer, configPath, `version: "2"
linters:
  enable:
    - gosec
`, types.LinterPriorityOptional, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).NotTo(ContainSubstring("gohumanize"),
			"the project depends on go-humanize, but a stock binary rejects the entry")
	})

	It("heals a poisoned config by removing the unknown entry (apply mode)", func() {
		content, err := fixAndRead(fixer, configPath, issueConfig, types.LinterPriorityOptional, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).NotTo(ContainSubstring("gohumanize"),
			"the stock binary cannot load the config while gohumanize is listed")

		// The fix must converge, not loop: a second run is a no-op.
		second, err := fixAndRead(fixer, configPath, content, types.LinterPriorityOptional, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(second).To(Equal(content))
	})

	It("reports the heal in check mode without touching the file", func() {
		writeConfig(configPath, issueConfig)

		result, err := fixer.FixConfig(context.Background(), configPath, types.LinterPriorityOptional, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.FixesApplied).To(BeNumerically(">", 0),
			"check mode must flag the poisoned config as drifted, not silently converged")

		saved, readErr := os.ReadFile(configPath)
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(saved)).To(ContainSubstring("gohumanize"),
			"a dry run leaves the config untouched")
	})

	It("loads the healed config with the pinned binary (post-apply tripwire stays silent)", func() {
		content, err := fixAndRead(fixer, configPath, issueConfig, types.LinterPriorityOptional, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Contains(content, "gohumanize")).To(BeFalse())

		Expect(analyzer.VerifyConfigLoads(context.Background(), configPath)).
			To(Succeed(), "the healed config must load with the binary that was probed")
	})
})
