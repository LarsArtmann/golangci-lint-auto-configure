package linter_test

import (
	"os"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAnalyzer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Analyzer Suite")
}

var _ = Describe("Analyzer", func() {
	var analyzer *linter.Analyzer

	BeforeEach(func() {
		logger := log.NewWithOptions(os.Stdout, log.Options{Level: log.ErrorLevel})
		analyzer = linter.NewAnalyzer(logger)
	})

	Context("Categorization", func() {
		It("should categorize gosec as critical", func() {
			recs := analyzer.GetLintersByPriority(
				[]types.LinterRecommendation{
					{Name: "gosec", Priority: types.LinterPriorityCritical},
				},
				types.LinterPriorityCritical,
			)

			Expect(recs).To(HaveLen(1))
			Expect(recs[0].Name).To(Equal(types.LinterName("gosec")))
			Expect(recs[0].Priority).To(Equal(types.LinterPriorityCritical))
		})

		It("should categorize wrapcheck as high priority", func() {
			recs := analyzer.GetLintersByPriority(
				[]types.LinterRecommendation{
					{Name: "wrapcheck", Priority: types.LinterPriorityHigh},
				},
				types.LinterPriorityHigh,
			)

			Expect(recs).To(HaveLen(1))
			Expect(recs[0].Name).To(Equal(types.LinterName("wrapcheck")))
			Expect(recs[0].Priority).To(Equal(types.LinterPriorityHigh))
		})

		It("should categorize misspell as medium priority", func() {
			recs := analyzer.GetLintersByPriority(
				[]types.LinterRecommendation{
					{Name: "misspell", Priority: types.LinterPriorityMedium},
				},
				types.LinterPriorityMedium,
			)

			Expect(recs).To(HaveLen(1))
			Expect(recs[0].Name).To(Equal(types.LinterName("misspell")))
			Expect(recs[0].Priority).To(Equal(types.LinterPriorityMedium))
		})
	})

	Context("Priority Filtering", func() {
		It("should filter linters by critical priority", func() {
			recs := []types.LinterRecommendation{
				{Name: "gosec", Priority: types.LinterPriorityCritical},
				{Name: "wrapcheck", Priority: types.LinterPriorityHigh},
				{Name: "misspell", Priority: types.LinterPriorityMedium},
			}

			filtered := analyzer.GetLintersByPriority(recs, types.LinterPriorityCritical)

			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(types.LinterName("gosec")))
		})

		It("should filter linters by high priority", func() {
			recs := []types.LinterRecommendation{
				{Name: "gosec", Priority: types.LinterPriorityCritical},
				{Name: "wrapcheck", Priority: types.LinterPriorityHigh},
				{Name: "misspell", Priority: types.LinterPriorityMedium},
			}

			filtered := analyzer.GetLintersByPriority(recs, types.LinterPriorityHigh)

			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].Name).To(Equal(types.LinterName("wrapcheck")))
		})

		It("should return empty list for non-existent priority", func() {
			recs := []types.LinterRecommendation{
				{Name: "gosec", Priority: types.LinterPriorityCritical},
			}

			filtered := analyzer.GetLintersByPriority(recs, types.LinterPriorityMedium)

			Expect(filtered).To(BeEmpty())
		})
	})

	Context("Recommendation Formatting", func() {
		It("should format recommendations with critical linters", func() {
			analysis := &types.ConfigAnalysis{
				CriticalCount:         2,
				LinterRecommendations: []types.LinterRecommendation{
					{Name: "gosec", Priority: types.LinterPriorityCritical, Reason: "Security"},
					{Name: "errcheck", Priority: types.LinterPriorityCritical, Reason: "Error checking"},
				},
			}

			formatted := analyzer.FormatRecommendations(analysis)

			Expect(formatted).To(ContainSubstring("🚨 2 CRITICAL"))
			Expect(formatted).To(ContainSubstring("gosec"))
			Expect(formatted).To(ContainSubstring("errcheck"))
		})

		It("should format recommendations with high value linters", func() {
			analysis := &types.ConfigAnalysis{
				CriticalCount:         0,
				HighValueCount:        1,
				LinterRecommendations: []types.LinterRecommendation{
					{Name: "wrapcheck", Priority: types.LinterPriorityHigh, Reason: "Error wrapping"},
				},
			}

			formatted := analyzer.FormatRecommendations(analysis)

			Expect(formatted).To(ContainSubstring("HIGH VALUE"))
			Expect(formatted).To(ContainSubstring("wrapcheck"))
		})

		It("should format recommendations with medium value linters", func() {
			analysis := &types.ConfigAnalysis{
				CriticalCount:         0,
				HighValueCount:        0,
				MediumValueCount:      1,
				LinterRecommendations: []types.LinterRecommendation{
					{Name: "misspell", Priority: types.LinterPriorityMedium, Reason: "Spelling"},
				},
			}

			formatted := analyzer.FormatRecommendations(analysis)

			Expect(formatted).To(ContainSubstring("MEDIUM VALUE"))
			Expect(formatted).To(ContainSubstring("misspell"))
		})

		It("should format recommendations with optional linters", func() {
			analysis := &types.ConfigAnalysis{
				CriticalCount:         0,
				HighValueCount:        0,
				MediumValueCount:      0,
				OptionalCount:         1,
				LinterRecommendations: []types.LinterRecommendation{
					{Name: "unknown", Priority: types.LinterPriorityOptional, Reason: "Optional"},
				},
			}

			formatted := analyzer.FormatRecommendations(analysis)

			Expect(formatted).To(ContainSubstring("💡 1 OPTIONAL"))
			Expect(formatted).To(ContainSubstring("unknown"))
		})
	})

	Context("Summary Generation", func() {
		It("should return summary for critical linters", func() {
			analysis := &types.ConfigAnalysis{
				CriticalCount:         3,
				LinterRecommendations: []types.LinterRecommendation{
					{Name: "gosec"},
					{Name: "errcheck"},
					{Name: "loggercheck"},
				},
			}

			summary := analyzer.GetSummary(analysis)

			Expect(summary).To(ContainSubstring("3 CRITICAL"))
			Expect(summary).To(ContainSubstring("3 disabled linters"))
		})

		It("should return summary for mixed priorities", func() {
			analysis := &types.ConfigAnalysis{
				CriticalCount:         1,
				HighValueCount:        2,
				LinterRecommendations: []types.LinterRecommendation{
					{Name: "gosec"},
					{Name: "wrapcheck"},
					{Name: "errorlint"},
				},
			}

			summary := analyzer.GetSummary(analysis)

			Expect(summary).To(ContainSubstring("1 CRITICAL"))
			Expect(summary).To(ContainSubstring("2 HIGH"))
		})

		It("should return message when all linters enabled", func() {
			analysis := &types.ConfigAnalysis{
				LinterRecommendations: []types.LinterRecommendation{},
			}

			summary := analyzer.GetSummary(analysis)

			Expect(summary).To(Equal("All linters enabled - no recommendations"))
		})
	})
})
