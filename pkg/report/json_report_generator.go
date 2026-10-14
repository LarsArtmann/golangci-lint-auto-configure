package report

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
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

// JSONReport represents the structure of the JSON report output: a
// projection of one desired-state evaluation.
type JSONReport struct {
	ConfigPath      string
	Summary         JSONSummary
	Deviations      []DeviationRecord
	Suppressed      []string
	EnabledLinters  []types.LinterName
	DisabledLinters []types.LinterName
}

// DeviationRecord is one planned change in the report wire format.
type DeviationRecord struct {
	Path    string
	Element string
	Kind    string
	Reason  string
}

// JSONSummary contains summary statistics.
type JSONSummary struct {
	EnabledCount    int
	DisabledCount   int
	DeviationsCount int
	Converged       bool
}

// GenerateJSONReport creates a JSON report from one desired-state evaluation.
func (g *JSONGenerator) GenerateJSONReport(evaluation *linter.Evaluation, outputPath string) error {
	g.logger.Infof("Generating JSON report: %s", outputPath)

	jsonReport := g.buildJSONReport(evaluation)

	jsonData, err := json.Marshal(
		jsonReport,
		jsontext.WithIndentPrefix(""),
		jsontext.WithIndent("  "),
		json.Deterministic(true),
	)
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

func (g *JSONGenerator) buildJSONReport(evaluation *linter.Evaluation) JSONReport {
	deviations := make([]DeviationRecord, 0, len(evaluation.Deviations))
	for _, change := range evaluation.Deviations {
		deviations = append(deviations, DeviationRecord{
			Path:    change.Path,
			Element: change.Element,
			Kind:    string(change.Kind),
			Reason:  change.Reason,
		})
	}

	suppressed := make([]string, 0, len(evaluation.Suppressed))
	for _, name := range evaluation.Suppressed {
		suppressed = append(suppressed, name.String())
	}

	return JSONReport{
		ConfigPath: evaluation.ConfigPath,
		Summary: JSONSummary{
			EnabledCount:    len(evaluation.Observed.Linters.Enable),
			DisabledCount:   len(evaluation.Observed.Linters.Disable),
			DeviationsCount: len(evaluation.Deviations),
			Converged:       evaluation.Converged,
		},
		Deviations:      deviations,
		Suppressed:      suppressed,
		EnabledLinters:  evaluation.Observed.Linters.Enable,
		DisabledLinters: evaluation.Observed.Linters.Disable,
	}
}
