package linter_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// disabledLinterEntry represents a linter entry for test setup
type disabledLinterEntry struct {
	name       string
	deprecated bool
}

func extractLinterNames(recommendations []types.LinterRecommendation) []string {
	names := make([]string, len(recommendations))
	for i, rec := range recommendations {
		names[i] = rec.Name.String()
	}
	return names
}

func disabledLintersWith(entries ...disabledLinterEntry) []types.LinterInfo {
	result := make([]types.LinterInfo, len(entries))
	for i, e := range entries {
		result[i] = types.LinterInfo{Name: types.LinterName(e.name), Deprecated: e.deprecated}
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
			disabledLinters := disabledLintersWith(
				disabledLinterEntry{name: "lll", deprecated: false},
				disabledLinterEntry{name: "misspell", deprecated: false},
			)

			enabledFormatters := []types.FormatterInfo{
				{Name: "golines", AutoFix: true},
			}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElements("misspell"))
			Expect(names).NotTo(ContainElement("lll"))
		})

		It("should recommend lll when golines formatter is NOT enabled", func() {
			disabledLinters := disabledLintersWith(
				disabledLinterEntry{name: "lll", deprecated: false},
				disabledLinterEntry{name: "misspell", deprecated: false},
			)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).To(ContainElement("misspell"))
		})

		It("should recommend lll when only other formatters are enabled", func() {
			disabledLinters := disabledLintersWith(
				disabledLinterEntry{name: "lll", deprecated: false},
			)

			enabledFormatters := []types.FormatterInfo{
				{Name: "gofmt", AutoFix: true},
			}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
		})
	})

	Context("Deprecated Linter Handling", func() {
		It("should skip deprecated linters", func() {
			disabledLinters := disabledLintersWith(
				disabledLinterEntry{name: "lll", deprecated: false},
				disabledLinterEntry{name: "deadcode", deprecated: true},
			)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).NotTo(ContainElement("deadcode"))
		})
	})

	Context("Disabled Linter Handling", func() {
		It("should skip explicitly disabled linters", func() {
			disabledLinters := disabledLintersWith(
				disabledLinterEntry{name: "lll", deprecated: false},
				disabledLinterEntry{name: "funcorder", deprecated: false},
			)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).NotTo(ContainElement("funcorder"))
		})
	})
})
