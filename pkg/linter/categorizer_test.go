package linter_test

import (
	"os"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CategorizeLinters", func() {
	var analyzer *linter.Analyzer

	BeforeEach(func() {
		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		analyzer = linter.NewAnalyzer(logger)
	})

	Context("Redundant Linter Detection", func() {
		It("should NOT recommend lll when golines formatter is enabled", func() {
			disabledLinters := []types.LinterInfo{
				{Name: "lll", Deprecated: false},
				{Name: "misspell", Deprecated: false},
			}

			enabledFormatters := []types.FormatterInfo{
				{Name: "golines", AutoFix: true},
			}

			recommendations := analyzer.CategorizeLinters(disabledLinters, enabledFormatters)

			names := make([]string, len(recommendations))
			for i, rec := range recommendations {
				names[i] = rec.Name.String()
			}

			Expect(names).To(ContainElements("misspell"))
			Expect(names).NotTo(ContainElement("lll"))
		})

		It("should recommend lll when golines formatter is NOT enabled", func() {
			disabledLinters := []types.LinterInfo{
				{Name: "lll", Deprecated: false},
				{Name: "misspell", Deprecated: false},
			}

			enabledFormatters := []types.FormatterInfo{}

			recommendations := analyzer.CategorizeLinters(disabledLinters, enabledFormatters)

			names := make([]string, len(recommendations))
			for i, rec := range recommendations {
				names[i] = rec.Name.String()
			}

			Expect(names).To(ContainElement("lll"))
			Expect(names).To(ContainElement("misspell"))
		})

		It("should recommend lll when only other formatters are enabled", func() {
			disabledLinters := []types.LinterInfo{
				{Name: "lll", Deprecated: false},
			}

			enabledFormatters := []types.FormatterInfo{
				{Name: "gofmt", AutoFix: true},
			}

			recommendations := analyzer.CategorizeLinters(disabledLinters, enabledFormatters)

			names := make([]string, len(recommendations))
			for i, rec := range recommendations {
				names[i] = rec.Name.String()
			}

			Expect(names).To(ContainElement("lll"))
		})
	})

	Context("Deprecated Linter Handling", func() {
		It("should skip deprecated linters", func() {
			disabledLinters := []types.LinterInfo{
				{Name: "lll", Deprecated: false},
				{Name: "deadcode", Deprecated: true},
			}

			enabledFormatters := []types.FormatterInfo{}

			recommendations := analyzer.CategorizeLinters(disabledLinters, enabledFormatters)

			names := make([]string, len(recommendations))
			for i, rec := range recommendations {
				names[i] = rec.Name.String()
			}

			Expect(names).To(ContainElement("lll"))
			Expect(names).NotTo(ContainElement("deadcode"))
		})
	})

	Context("Disabled Linter Handling", func() {
		It("should skip explicitly disabled linters", func() {
			disabledLinters := []types.LinterInfo{
				{Name: "lll", Deprecated: false},
				{Name: "funcorder", Deprecated: false},
			}

			enabledFormatters := []types.FormatterInfo{}

			recommendations := analyzer.CategorizeLinters(disabledLinters, enabledFormatters)

			names := make([]string, len(recommendations))
			for i, rec := range recommendations {
				names[i] = rec.Name.String()
			}

			Expect(names).To(ContainElement("lll"))
			Expect(names).NotTo(ContainElement("funcorder"))
		})
	})
})
