package linter

import (
	"testing"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func BenchmarkCategorizeLinters(b *testing.B) {
	logger := log.NewWithOptions(ioDiscard{}, log.Options{Level: log.ErrorLevel})
	analyzer := NewAnalyzer(logger)

	disabled := make([]types.LinterInfo, 0, 100)
	for i := range 100 {
		disabled = append(disabled, types.LinterInfo{
			Name:        types.LinterName("linter-" + string(rune('a'+i%26))),
			Description: "Test linter",
		})
	}

	var enabledFormatters []types.FormatterInfo

	b.ResetTimer()

	for b.Loop() {
		analyzer.CategorizeLinters(disabled, enabledFormatters)
	}
}

func BenchmarkGetLintersByPriority(b *testing.B) {
	logger := log.NewWithOptions(ioDiscard{}, log.Options{Level: log.ErrorLevel})
	analyzer := NewAnalyzer(logger)

	recommendations := make([]types.LinterRecommendation, 0, 100)
	for i := range 100 {
		recommendations = append(recommendations, types.LinterRecommendation{
			Name:     types.LinterName("linter-" + string(rune('a'+i%26))),
			Priority: types.LinterPriority(i % 4),
			Reason:   "Test reason",
		})
	}

	b.ResetTimer()

	for b.Loop() {
		analyzer.GetLintersByPriority(recommendations, types.LinterPriorityHigh)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
