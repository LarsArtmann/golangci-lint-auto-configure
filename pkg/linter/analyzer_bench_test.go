package linter

import (
	"testing"

	"github.com/charmbracelet/log"
)

func BenchmarkAnalyzer_AnalyzeConfig(b *testing.B) {
	logger := log.NewWithOptions(nil, log.Options{ReportCaller: false, Level: log.ErrorLevel})
	analyzer := NewAnalyzer(logger)

	// Find a real config to analyze
	configPath := "../../.golangci.yml"

	for b.Loop() {
		_, err := analyzer.AnalyzeConfig(configPath)
		if err != nil {
			b.Fatalf("AnalyzeConfig failed: %v", err)
		}
	}
}

func BenchmarkAnalyzer_GetLintersByPriority(b *testing.B) {
	logger := log.NewWithOptions(nil, log.Options{ReportCaller: false, Level: log.ErrorLevel})
	analyzer := NewAnalyzer(logger)

	// Get real recommendations
	configPath := "../../.golangci.yml"

	analysis, err := analyzer.AnalyzeConfig(configPath)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	recommendations := analysis.LinterRecommendations

	for b.Loop() {
		_ = analyzer.GetLintersByPriority(recommendations, 0)
	}
}

func BenchmarkAnalyzer_FormatRecommendations(b *testing.B) {
	logger := log.NewWithOptions(nil, log.Options{ReportCaller: false, Level: log.ErrorLevel})
	analyzer := NewAnalyzer(logger)

	configPath := "../../.golangci.yml"

	analysis, err := analyzer.AnalyzeConfig(configPath)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	for b.Loop() {
		_ = analyzer.FormatRecommendations(analysis)
	}
}

func BenchmarkAnalyzer_categorizeLinters(b *testing.B) {
	logger := log.NewWithOptions(nil, log.Options{ReportCaller: false, Level: log.ErrorLevel})
	analyzer := NewAnalyzer(logger)

	configPath := "../../.golangci.yml"

	analysis, err := analyzer.AnalyzeConfig(configPath)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	disabledLinters := analysis.DisabledLinters

	for b.Loop() {
		_ = analyzer.categorizeLinters(disabledLinters)
	}
}
