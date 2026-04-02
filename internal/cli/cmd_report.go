package cli

import (
	"context"
	"fmt"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/report"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/spf13/cobra"
)

// newReportCommand creates the report command.
func newReportCommand(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate HTML report of configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReport(cmd, logger, analyzer, configLoader)
		},
	}

	return cmd
}

func runReport(
	cmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
) error {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	configFile, err := resolveConfigPath(configLoader, configPath)
	if err != nil {
		return fmt.Errorf("failed to find config file (configPath=%s): %w", configPath, err)
	}

	logger.Infof("Generating %s report for: %s", reportFormat, configFile)

	analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)
	if err != nil {
		return fmt.Errorf("failed to analyze config (configPath=%s): %w", configFile, err)
	}

	outputPath := outputReport
	if outputPath == "report.html" && reportFormat == "json" {
		outputPath = "report.json"
	}

	if reportFormat == "json" {
		return writeJSONReport(logger, analysis, outputPath, configFile)
	}

	return writeHTMLReport(cmd.Context(), logger, analysis, outputPath, configFile)
}

func writeJSONReport(
	logger *log.Logger,
	analysis *types.ConfigAnalysis,
	outputPath, configFile string,
) error {
	gen := report.NewJSONGenerator(logger)

	if err := gen.GenerateJSONReport(analysis, outputPath); err != nil {
		return fmt.Errorf(
			"failed to generate JSON report (configPath=%s, outputPath=%s): %w",
			configFile,
			outputPath,
			err,
		)
	}

	return nil
}

func writeHTMLReport(
	ctx context.Context,
	logger *log.Logger,
	analysis *types.ConfigAnalysis,
	outputPath, configFile string,
) error {
	gen := report.NewGenerator(logger)

	if err := gen.GenerateReport(ctx, analysis, outputPath); err != nil {
		return fmt.Errorf(
			"failed to generate HTML report (configPath=%s, outputPath=%s): %w",
			configFile,
			outputPath,
			err,
		)
	}

	return nil
}
