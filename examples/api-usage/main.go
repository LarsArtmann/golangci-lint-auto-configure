// Example of using golangci-lint-auto-configure as a library
package main

import (
	"context"
	"log/slog"
	"os"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/client"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func main() {
	logger := setupLogger()
	slog.SetDefault(slog.New(logger))

	clientObj := client.New(client.Options{Verbose: true})

	analysis := analyzeConfig(clientObj)

	printAnalysisResults(analysis)

	critical, high := filterRecommendations(analysis)
	logCriticalLinters(critical)
	logHighPriorityLinters(high)

	showSummary(clientObj, analysis)
}

func setupLogger() *log.Logger {
	return log.NewWithOptions(os.Stdout, log.Options{
		ReportTimestamp: true,
		Level:           log.InfoLevel,
	})
}

func analyzeConfig(clientObj *client.Client) *types.ConfigAnalysis {
	configPath := ".golangci.yml"

	analysis, err := clientObj.AnalyzeConfig(context.Background(), configPath)
	if err != nil {
		slog.Error("Analysis failed", "error", err)
		os.Exit(1)
	}

	return analysis
}

func printAnalysisResults(analysis *types.ConfigAnalysis) {
	slog.Info("Analysis complete", "path", analysis.ConfigPath,
		"enabled_linters", len(analysis.EnabledLinters),
		"disabled_linters", len(analysis.DisabledLinters),
		"enabled_formatters", len(analysis.EnabledFormatters),
		"disabled_formatters", len(analysis.DisabledFormatters),
		"recommendations", len(analysis.LinterRecommendations))
}

func filterRecommendations(
	analysis *types.ConfigAnalysis,
) ([]types.LinterRecommendation, []types.LinterRecommendation) {
	var critical, high []types.LinterRecommendation

	for _, rec := range analysis.LinterRecommendations {
		switch rec.Priority {
		case types.LinterPriorityCritical:
			critical = append(critical, rec)
		case types.LinterPriorityHigh:
			high = append(high, rec)
		case types.LinterPriorityMedium, types.LinterPriorityOptional:
			// Skip lower priority recommendations
		}
	}

	return critical, high
}

func logCriticalLinters(critical []types.LinterRecommendation) {
	if len(critical) > 0 {
		for _, rec := range critical {
			slog.Warn("Critical linter (should ALWAYS be enabled)", "name", rec.Name)
		}
	}
}

func logHighPriorityLinters(high []types.LinterRecommendation) {
	if len(high) > 0 {
		for _, rec := range high {
			slog.Info("High priority linter", "name", rec.Name, "reason", rec.Reason)
		}
	}
}

func showSummary(client *client.Client, analysis *types.ConfigAnalysis) {
	slog.Info("Summary", "message", client.GetSummary(analysis))
}
