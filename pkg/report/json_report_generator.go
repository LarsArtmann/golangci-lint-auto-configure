package report

import (
	"encoding/json"
	"os"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// JSONGenerator generates JSON reports for configuration analysis.
type JSONGenerator struct {
	logger *log.Logger
}

// jsonFilePerm restricts report output to owner-only read/write.
const jsonFilePerm = 0o600

// NewJSONGenerator creates a new JSON report generator.
func NewJSONGenerator(logger *log.Logger) *JSONGenerator {
	return &JSONGenerator{
		logger: logger,
	}
}

// JSONReport represents the structure of the JSON report output.
type JSONReport struct {
	ConfigPath      string
	Summary         JSONSummary
	Recommendations []types.LinterRecommendation
	EnabledLinters  []string
	DisabledLinters []string
}

// JSONSummary contains summary statistics.
type JSONSummary struct {
	TotalLinters         int
	EnabledCount         int
	DisabledCount        int
	RecommendationsCount int
}

// GenerateJSONReport creates a JSON report from the analysis.
func (g *JSONGenerator) GenerateJSONReport(analysis *types.ConfigAnalysis, outputPath string) error {
	g.logger.Infof("Generating JSON report: %s", outputPath)

	jsonReport := g.buildJSONReport(analysis)

	//nolint:musttag // intentionally tag-free: PascalCase via Go field names
	jsonData, err := json.MarshalIndent(jsonReport, "", "  ")
	if err != nil {
		return errorfamily.WrapCorruptionf(err, "report.json_marshal",
			"failed to marshal JSON report (outputPath=%s)", outputPath)
	}

	if writeErr := os.WriteFile(outputPath, jsonData, jsonFilePerm); writeErr != nil {
		return errorfamily.WrapRejectionf(writeErr, "report.json_write",
			"failed to write JSON report (outputPath=%s)", outputPath)
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
