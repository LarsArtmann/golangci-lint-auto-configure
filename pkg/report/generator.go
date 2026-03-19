package report

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
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
func (g *Generator) GenerateReport(ctx context.Context, analysis *types.ConfigAnalysis, outputPath string) error {
	g.logger.Infof("Generating HTML report: %s", outputPath)

	data := ReportData{
		Analysis: analysis,
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	defer func() { _ = f.Close() }()

	//nolint:contextcheck // Context comes from caller; templ.Render receives it correctly
	err = Report(data).Render(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to render report: %w", err)
	}

	g.logger.Infof("Report generated successfully: %s", outputPath)

	return nil
}
