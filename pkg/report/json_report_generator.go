package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// JSONGenerator generates JSON reports for configuration analysis
type JSONGenerator struct {
	logger *log.Logger
}

// NewJSONGenerator creates a new JSON report generator
func NewJSONGenerator(logger *log.Logger) *JSONGenerator {
	return &JSONGenerator{
		logger: logger,
	}
}

// JSONReport represents the structure of the JSON report output
type JSONReport struct {
	ConfigPath      string                       `json:"configPath"`
	Summary         JSONSummary                  `json:"summary"`
	Recommendations []types.LinterRecommendation `json:"recommendations"`
	EnabledLinters  []string                     `json:"enabledLinters"`
	DisabledLinters []string                     `json:"disabledLinters"`
}

// JSONSummary contains summary statistics
type JSONSummary struct {
	TotalLinters         int `json:"totalLinters"`
	EnabledCount         int `json:"enabledCount"`
	DisabledCount        int `json:"disabledCount"`
	RecommendationsCount int `json:"recommendationsCount"`
}

// GenerateJSONReport creates a JSON report from the analysis
func (g *JSONGenerator) GenerateJSONReport(analysis *types.ConfigAnalysis, outputPath string) error {
	g.logger.Infof("Generating JSON report: %s", outputPath)

	// Prepare JSON report structure
	enabledLinterNames := make([]string, len(analysis.EnabledLinters))
	for i, linter := range analysis.EnabledLinters {
		enabledLinterNames[i] = linter.Name
	}

	disabledLinterNames := make([]string, len(analysis.DisabledLinters))
	for i, linter := range analysis.DisabledLinters {
		disabledLinterNames[i] = linter.Name
	}

	jsonReport := JSONReport{
		ConfigPath: analysis.ConfigPath,
		Summary: JSONSummary{
			TotalLinters:         len(analysis.EnabledLinters) + len(analysis.DisabledLinters),
			EnabledCount:         len(analysis.EnabledLinters),
			DisabledCount:        len(analysis.DisabledLinters),
			RecommendationsCount: len(analysis.LinterRecommendations),
		},
		Recommendations: analysis.LinterRecommendations,
		EnabledLinters:  enabledLinterNames,
		DisabledLinters: disabledLinterNames,
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(jsonReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write JSON report to %s: %w", outputPath, err)
	}

	g.logger.Infof("JSON report generated successfully: %s", outputPath)
	return nil
}
