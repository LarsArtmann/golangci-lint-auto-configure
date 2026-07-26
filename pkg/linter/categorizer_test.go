package linter_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// disabledLinterEntry represents a linter entry for test setup.
type disabledLinterEntry struct {
	name       string
	deprecated bool
}

func newDisabledEntry(name string, deprecated bool) disabledLinterEntry {
	return disabledLinterEntry{name: name, deprecated: deprecated}
}

// shared linter sets for common test scenarios.
var (
	linterSetLllMisspell  = newDisabledEntries("lll", "misspell")
	linterSetLllOnly      = newDisabledEntries("lll")
	linterSetLllDeadcode  = append(newDisabledEntries("lll"), newDisabledEntry("deadcode", true))
	linterSetLllFuncorder = newDisabledEntries("lll", "funcorder")
	linterSetLllNoinline  = newDisabledEntries("lll", "noinlineerr")
)

func newDisabledEntries(names ...string) []disabledLinterEntry {
	entries := make([]disabledLinterEntry, 0, len(names))
	for _, name := range names {
		entries = append(entries, newDisabledEntry(name, false))
	}

	return entries
}

func extractLinterNames(recommendations []types.LinterRecommendation) []string {
	names := make([]string, 0, len(recommendations))
	for _, rec := range recommendations {
		names = append(names, rec.Name.String())
	}

	return names
}

func extractFormatterNames(recommendations []types.FormatterRecommendation) []string {
	names := make([]string, 0, len(recommendations))
	for _, rec := range recommendations {
		names = append(names, rec.Name.String())
	}

	return names
}

func disabledLintersWith(entries ...disabledLinterEntry) []types.LinterInfo {
	result := make([]types.LinterInfo, 0, len(entries))
	for _, e := range entries {
		result = append(result, types.LinterInfo{Name: types.LinterName(e.name), Deprecated: e.deprecated})
	}

	return result
}

var _ = Describe("CategorizeLinters", func() {
	var analyzer *linter.Analyzer

	BeforeEach(func() {
		analyzer = linter.NewAnalyzer(linter.NewTestLogger())
	})

	Context("Redundant Linter Detection", func() {
		It("should NOT recommend lll when golines formatter is enabled", func() {
			disabledLinters := disabledLintersWith(linterSetLllMisspell...)

			enabledFormatters := []types.FormatterInfo{
				{Name: "golines", AutoFix: true},
			}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElements("misspell"))
			Expect(names).NotTo(ContainElement("lll"))
		})

		It("should recommend lll when golines formatter is NOT enabled", func() {
			disabledLinters := disabledLintersWith(linterSetLllMisspell...)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).To(ContainElement("misspell"))
		})

		It("should recommend lll when only other formatters are enabled", func() {
			disabledLinters := disabledLintersWith(linterSetLllOnly...)

			enabledFormatters := []types.FormatterInfo{
				{Name: "gofmt", AutoFix: true},
			}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
		})
	})

	Context("Project-Specific Linter Detection", func() {
		It("should skip clickhouselint when projectRoot is empty (fail-open)", func() {
			disabledLinters := []types.LinterInfo{
				{Name: "clickhouselint"},
				{Name: "misspell"},
			}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, []types.FormatterInfo{}))

			Expect(names).To(ContainElement("clickhouselint"))
			Expect(names).To(ContainElement("misspell"))
		})
	})

	Context("Deprecated Linter Handling", func() {
		It("should skip deprecated linters", func() {
			disabledLinters := disabledLintersWith(linterSetLllDeadcode...)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).NotTo(ContainElement("deadcode"))
		})
	})

	Context("Disabled Linter Handling", func() {
		It("should skip explicitly disabled linters", func() {
			disabledLinters := disabledLintersWith(linterSetLllFuncorder...)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).NotTo(ContainElement("funcorder"))
		})

		It("should skip noinlineerr as conflicting with formatters", func() {
			disabledLinters := disabledLintersWith(linterSetLllNoinline...)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).NotTo(ContainElement("noinlineerr"))
		})
	})

	Context("Pragmatic Mode", func() {
		var noiseLinterEntries []disabledLinterEntry

		BeforeEach(func() {
			noiseLinterEntries = newDisabledEntries(
				"gochecknoglobals", "wrapcheck", "ireturn", "funlen", "misspell",
			)
		})

		It("should skip the 4 noise linters when pragmatic is enabled", func() {
			analyzer.SetPragmatic(true)

			disabledLinters := disabledLintersWith(noiseLinterEntries...)
			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, []types.FormatterInfo{}))

			Expect(names).To(ContainElement("misspell"))
			Expect(names).NotTo(ContainElement("gochecknoglobals"))
			Expect(names).NotTo(ContainElement("wrapcheck"))
			Expect(names).NotTo(ContainElement("ireturn"))
			Expect(names).NotTo(ContainElement("funlen"))
		})

		It("should keep all noise linters when pragmatic is disabled (default)", func() {
			disabledLinters := disabledLintersWith(noiseLinterEntries...)
			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, []types.FormatterInfo{}))

			Expect(names).To(ContainElement("gochecknoglobals"))
			Expect(names).To(ContainElement("wrapcheck"))
			Expect(names).To(ContainElement("ireturn"))
			Expect(names).To(ContainElement("funlen"))
			Expect(names).To(ContainElement("misspell"))
		})
	})

	Context("NeverAutoEnable Linters", func() {
		It("should never auto-enable exhaustruct even when pragmatic is disabled", func() {
			entries := newDisabledEntries("exhaustruct", "misspell")
			disabledLinters := disabledLintersWith(entries...)
			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, []types.FormatterInfo{}))

			Expect(names).NotTo(ContainElement("exhaustruct"))
			Expect(names).To(ContainElement("misspell"))
		})

		It("should never auto-enable exhaustruct even when pragmatic is enabled", func() {
			analyzer.SetPragmatic(true)

			entries := newDisabledEntries("exhaustruct", "misspell")
			disabledLinters := disabledLintersWith(entries...)
			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, []types.FormatterInfo{}))

			Expect(names).NotTo(ContainElement("exhaustruct"))
			Expect(names).To(ContainElement("misspell"))
		})
	})
})

var _ = Describe("CategorizeFormatters", func() {
	var analyzer *linter.Analyzer

	BeforeEach(func() {
		analyzer = linter.NewAnalyzer(linter.NewTestLogger())
	})

	Context("Redundant Formatter Detection", func() {
		It("should NOT recommend gofmt when gofumpt is enabled", func() {
			disabledFormatters := []types.FormatterInfo{
				{Name: "gofmt"},
			}

			enabledFormatters := []types.FormatterInfo{
				{Name: "gofumpt"},
			}

			recs := analyzer.CategorizeFormatters(disabledFormatters, enabledFormatters)

			names := extractFormatterNames(recs)
			Expect(names).NotTo(ContainElement("gofmt"))
		})

		It("should recommend gofmt when gofumpt is NOT enabled", func() {
			disabledFormatters := []types.FormatterInfo{
				{Name: "gofmt"},
			}

			recs := analyzer.CategorizeFormatters(disabledFormatters, []types.FormatterInfo{})

			names := extractFormatterNames(recs)
			Expect(names).To(ContainElement("gofmt"))
		})

		It("should recommend non-redundant formatters regardless", func() {
			disabledFormatters := []types.FormatterInfo{
				{Name: "gci"},
				{Name: "goimports"},
				{Name: "gofmt"},
			}

			enabledFormatters := []types.FormatterInfo{
				{Name: "gofumpt"},
			}

			recs := analyzer.CategorizeFormatters(disabledFormatters, enabledFormatters)

			names := extractFormatterNames(recs)
			Expect(names).To(ContainElement("gci"))
			Expect(names).To(ContainElement("goimports"))
			Expect(names).NotTo(ContainElement("gofmt"))
		})
	})

	Context("Project-Specific Formatter Detection", func() {
		It("should skip swaggo when projectRoot is empty (fail-open)", func() {
			disabledFormatters := []types.FormatterInfo{
				{Name: "swaggo"},
				{Name: "gci"},
			}

			recs := analyzer.CategorizeFormatters(disabledFormatters, []types.FormatterInfo{})

			names := extractFormatterNames(recs)
			Expect(names).To(ContainElement("swaggo"))
			Expect(names).To(ContainElement("gci"))
		})
	})

	Context("Empty Input", func() {
		It("should return empty recommendations for no disabled formatters", func() {
			recs := analyzer.CategorizeFormatters([]types.FormatterInfo{}, []types.FormatterInfo{})

			Expect(recs).To(BeEmpty())
		})
	})
})
