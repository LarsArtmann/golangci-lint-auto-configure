// Example of using golangci-linter-auto-configure as a library
package main

import (
	"log/slog"
	"os"

	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/client"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

func main() {
	slog.Info("golangci-linter-auto-configure API Usage Example")

	// Create client with verbose logging
	c := client.New(client.Options{
		Verbose: true,
	})

	// Analyze configuration
	configPath := ".golangci.yml"
	analysis, err := c.AnalyzeConfig(configPath)
	if err != nil {
		slog.Error("Analysis failed", "error", err)
		os.Exit(1)
	}

	// Print results
	slog.Info("Analysis complete", "path", configPath,
		"enabled_linters", len(analysis.EnabledLinters),
		"disabled_linters", len(analysis.DisabledLinters),
		"enabled_formatters", len(analysis.EnabledFormatters),
		"disabled_formatters", len(analysis.DisabledFormatters),
		"recommendations", len(analysis.LinterRecommendations))

	// Show critical recommendations
	var critical []types.LinterRecommendation
	var high []types.LinterRecommendation

	for _, rec := range analysis.LinterRecommendations {
		switch rec.Priority {
		case types.LinterPriorityCritical:
			critical = append(critical, rec)
		case types.LinterPriorityHigh:
			high = append(high, rec)
		}
	}

	if len(critical) > 0 {
		for _, rec := range critical {
			slog.Warn("Critical linter (should ALWAYS be enabled)", "name", rec.Name)
		}
	}

	if len(high) > 0 {
		for _, rec := range high {
			slog.Info("High priority linter", "name", rec.Name, "reason", rec.Reason)
		}
	}

	// Show summary
	slog.Info("Summary", "message", c.GetSummary(analysis))
}
