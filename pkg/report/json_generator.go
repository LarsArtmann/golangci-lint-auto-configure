package report

import (
	"encoding/json"
	"os"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// JSONGenerator generates JSON reports for golangci-lint configurations
type JSONGenerator struct {
	logger *log.Logger
}

// NewJSONGenerator creates a new JSON report generator
func NewJSONGenerator(logger *log.Logger) *JSONGenerator {
	return &JSONGenerator{
		logger: logger,
	}
}

// GenerateJSONReport generates a JSON report for the given analysis
func (g *JSONGenerator) GenerateJSONReport(analysis *types.ConfigAnalysis, outputPath string) error {
	g.logger.Infof("Generating JSON report: %s", outputPath)

	data := ReportData{
		Analysis: analysis,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return g.logger.Errorf("failed to marshal JSON report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return g.logger.Errorf("failed to write JSON report to %s: %w", outputPath, err)
	}

	g.logger.Infof("JSON report generated successfully: %s", outputPath)
	return nil
}
