package linter_test

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func extractLinterNames(recommendations []types.LinterRecommendation) []string {
	names := make([]string, len(recommendations))
	for i, rec := range recommendations {
		names[i] = rec.Name.String()
	}
	return names
}

func disabledLintersWith(namesAndDeprecation ...struct{ name string; deprecated bool }) []types.LinterInfo {
	result := make([]types.LinterInfo, len(namesAndDeprecation))
	for i, nd := range namesAndDeprecation {
		result[i] = types.LinterInfo{Name: nd.name, Deprecated: nd.deprecated}
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
				{"lll", false},
				{"misspell", false},
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
				{"lll", false},
				{"misspell", false},
			)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).To(ContainElement("misspell"))
		})

		It("should recommend lll when only other formatters are enabled", func() {
			disabledLinters := disabledLintersWith(
				{"lll", false},
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
				{"lll", false},
				{"deadcode", true},
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
				{"lll", false},
				{"funcorder", false},
			)

			enabledFormatters := []types.FormatterInfo{}

			names := extractLinterNames(analyzer.CategorizeLinters(disabledLinters, enabledFormatters))

			Expect(names).To(ContainElement("lll"))
			Expect(names).NotTo(ContainElement("funcorder"))
		})
	})
})
