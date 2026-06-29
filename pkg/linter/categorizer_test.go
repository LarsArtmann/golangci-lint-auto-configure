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
	})
})
