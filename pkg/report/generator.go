package report

import (
	"context"
	"os"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Generator generates HTML reports for golangci-lint configurations.
type Generator struct {
	logger *log.Logger
}

// NewGenerator creates a new report generator.
func NewGenerator(logger *log.Logger) *Generator {
	return &Generator{
		logger: logger,
	}
}

// GenerateReport generates an HTML report for given analysis.
func (g *Generator) GenerateReport(ctx context.Context, analysis *types.ConfigAnalysis, outputPath string) (err error) {
	g.logger.Infof("Generating HTML report: %s", outputPath)

	data := ReportData{
		Analysis: analysis,
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return errorfamily.WrapRejectionf(err, "report.create_output",
			"failed to create output file (outputPath=%s)", outputPath)
	}

	defer func() {
		if cerr := outputFile.Close(); cerr != nil && err == nil {
			err = errorfamily.WrapCorruptionf(cerr, "report.close_output",
				"failed to close output file (outputPath=%s)", outputPath)
		}
	}()

	//nolint:contextcheck // Context comes from caller; templ.Render receives it correctly
	err = Report(data).Render(ctx, outputFile)
	if err != nil {
		return errorfamily.WrapCorruptionf(err, "report.render",
			"failed to render report (outputPath=%s)", outputPath)
	}

	g.logger.Infof("Report generated successfully: %s", outputPath)

	return nil
}
