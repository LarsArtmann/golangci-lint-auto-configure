package linter_test

import (
	"context"
	"testing"

	"charm.land/log/v2"
	linterpkg "github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
)

const defaultTestConfigPath = "../../.golangci.yml"

func newBenchLogger() *log.Logger {
	return log.NewWithOptions(nil, log.Options{ReportCaller: false, Level: log.ErrorLevel})
}

func BenchmarkAnalyzer_AnalyzeConfig(b *testing.B) {
	logger := newBenchLogger()
	analyzer := linterpkg.NewAnalyzer(logger)

	for b.Loop() {
		_, err := analyzer.AnalyzeConfig(context.Background(), defaultTestConfigPath)
		if err != nil {
			b.Fatalf("AnalyzeConfig failed: %v", err)
		}
	}
}

func BenchmarkAnalyzer_GetLintersByPriority(b *testing.B) {
	logger := newBenchLogger()
	analyzer := linterpkg.NewAnalyzer(logger)

	analysis, err := analyzer.AnalyzeConfig(context.Background(), defaultTestConfigPath)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	recommendations := analysis.LinterRecommendations

	for b.Loop() {
		_ = analyzer.GetLintersByPriority(recommendations, 0)
	}
}

func BenchmarkAnalyzer_FormatRecommendations(b *testing.B) {
	logger := newBenchLogger()
	analyzer := linterpkg.NewAnalyzer(logger)

	analysis, err := analyzer.AnalyzeConfig(context.Background(), defaultTestConfigPath)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	for b.Loop() {
		_ = analyzer.FormatRecommendations(analysis)
	}
}
