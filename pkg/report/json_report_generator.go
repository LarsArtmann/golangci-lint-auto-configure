package report

import (
	"encoding/json"
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// JSONGenerator generates JSON reports for configuration analysis.
type JSONGenerator struct {
	logger *log.Logger
}

// NewJSONGenerator creates a new JSON report generator.
func NewJSONGenerator(logger *log.Logger) *JSONGenerator {
	return &JSONGenerator{
		logger: logger,
	}
}

// JSONReport represents the structure of the JSON report output.
type JSONReport struct {
	ConfigPath      string                       `json:"configPath"`
	Summary         JSONSummary                  `json:"summary"`
	Recommendations []types.LinterRecommendation `json:"recommendations"`
	EnabledLinters  []string                     `json:"enabledLinters"`
	DisabledLinters []string                     `json:"disabledLinters"`
}

// JSONSummary contains summary statistics.
type JSONSummary struct {
	TotalLinters         int `json:"totalLinters"`
	EnabledCount         int `json:"enabledCount"`
	DisabledCount        int `json:"disabledCount"`
	RecommendationsCount int `json:"recommendationsCount"`
}

// GenerateJSONReport creates a JSON report from the analysis.
func (g *JSONGenerator) GenerateJSONReport(analysis *types.ConfigAnalysis, outputPath string) error {
	g.logger.Infof("Generating JSON report: %s", outputPath)

	jsonReport := g.buildJSONReport(analysis)

	jsonData, err := json.MarshalIndent(jsonReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report (outputPath=%s): %w", outputPath, err)
	}

	if writeErr := os.WriteFile(outputPath, jsonData, 0o644); writeErr != nil {
		return fmt.Errorf("failed to write JSON report (outputPath=%s): %w", outputPath, writeErr)
	}

	g.logger.Infof("JSON report generated successfully: %s", outputPath)

	return nil
}

func (g *JSONGenerator) buildJSONReport(analysis *types.ConfigAnalysis) JSONReport {
	enabledLinterNames := extractLinterNames(analysis.EnabledLinters)
	disabledLinterNames := extractLinterNames(analysis.DisabledLinters)

	return JSONReport{
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
}

func extractLinterNames(linters []types.LinterInfo) []string {
	names := make([]string, 0, len(linters))
	for _, linter := range linters {
		names = append(names, string(linter.Name))
	}

	return names
}
