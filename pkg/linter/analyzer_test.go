package linter_test

import (
	"context"
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAnalyzer(t *testing.T) {
	RegisterFailHandler(Fail) //art-dupl:accept Ginkgo per-package bootstrap
	RunSpecs(t, "Analyzer Suite")
}

var _ = Describe("Analyzer", func() {
	var (
		analyzer            *linter.Analyzer
		standardTestLinters []types.LinterRecommendation
	)

	BeforeEach(func() {
		analyzer = linter.NewAnalyzer(linter.NewTestLogger())
		standardTestLinters = []types.LinterRecommendation{
			{Name: "gosec", Priority: types.LinterPriorityCritical},
			{Name: "wrapcheck", Priority: types.LinterPriorityHigh},
			{Name: "misspell", Priority: types.LinterPriorityMedium},
		}
	})

	Context("Priority Filtering", func() {
		It("should filter linters by priority", func() {
			testCases := []struct {
				priority         types.LinterPriority
				expectedLinter   types.LinterName
				expectedPriority types.LinterPriority
			}{
				{types.LinterPriorityCritical, "gosec", types.LinterPriorityCritical},
				{types.LinterPriorityHigh, "wrapcheck", types.LinterPriorityHigh},
				{types.LinterPriorityMedium, "misspell", types.LinterPriorityMedium},
			}

			for _, tc := range testCases {
				filtered := analyzer.GetLintersByPriority(standardTestLinters, tc.priority)

				Expect(filtered).To(HaveLen(1))
				Expect(filtered[0].Name).To(Equal(tc.expectedLinter))
				Expect(filtered[0].Priority).To(Equal(tc.expectedPriority))
			}
		})

		It("should return empty list for non-existent priority", func() {
			recs := []types.LinterRecommendation{
				{Name: "gosec", Priority: types.LinterPriorityCritical},
			}

			filtered := analyzer.GetLintersByPriority(recs, types.LinterPriorityMedium)

			Expect(filtered).To(BeEmpty())
		})
	})

	Context("Summary Generation", func() {
		It("should return summary for critical linters", func() {
			analysis := &types.ConfigAnalysis{
				CriticalCount: 3,
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
				CriticalCount:  1,
				HighValueCount: 2,
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

	Context("Version Check", func() {
		var versionAnalyzer *linter.Analyzer

		BeforeEach(func() {
			versionAnalyzer = linter.NewAnalyzer(log.Default())
		})

		Context("When golangci-lint is installed", func() {
			It("should find the binary", func() {
				err := versionAnalyzer.FindBinary(context.Background())
				if err != nil {
					Skip("golangci-lint not found in PATH")
				}

				Expect(err).ToNot(HaveOccurred())
			})

			It("should pass version check with v2.10.1 or newer", func() {
				err := versionAnalyzer.FindBinary(context.Background())
				if err != nil {
					Skip("golangci-lint not found in PATH")
				}

				err = versionAnalyzer.CheckVersion(context.Background())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
