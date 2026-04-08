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

func newReportCommand(b *CommandBuilder) *cobra.Command {
	return b.Build(
		"report",
		"Generate HTML report of configuration",
		func(cmd *cobra.Command, _ []string) error {
			return runReport(cmd, b.Logger(), b.Analyzer(), b.ConfigLoader())
		},
	)
}

func runReport(
	cmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
) error {
	setLogLevel(logger)

	configFile, err := resolveConfigPath(cmd.Context(), configLoader, logger, configPath, dryRun)
	if err != nil {
		return fmt.Errorf("failed to find config file (configPath=%s): %w", configPath, err)
	}

	analysis, err := analyzeConfig(logger, cmd, analyzer, configFile)
	if err != nil {
		return err
	}

	outputPath := determineOutputPath(outputReport, reportFormat)

	if reportFormat == "json" {
		return writeJSONReport(logger, analysis, outputPath, configFile)
	}

	return writeHTMLReport(cmd.Context(), logger, analysis, outputPath, configFile)
}

func setLogLevel(logger *log.Logger) {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}
}

func analyzeConfig(
	logger *log.Logger,
	cmd *cobra.Command,
	analyzer *linter.Analyzer,
	configFile string,
) (*types.ConfigAnalysis, error) {
	logger.Infof("Generating %s report for: %s", reportFormat, configFile)

	analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze config (configPath=%s): %w", configFile, err)
	}

	return analysis, nil
}

func determineOutputPath(outputPath, format string) string {
	if outputPath == "report.html" && format == "json" {
		return "report.json"
	}

	return outputPath
}

func writeJSONReport(
	logger *log.Logger,
	analysis *types.ConfigAnalysis,
	outputPath, configFile string,
) error {
	gen := report.NewJSONGenerator(logger)

	err := gen.GenerateJSONReport(analysis, outputPath)
	if err != nil {
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

	err := gen.GenerateReport(ctx, analysis, outputPath)
	if err != nil {
		return fmt.Errorf(
			"failed to generate HTML report (configPath=%s, outputPath=%s): %w",
			configFile,
			outputPath,
			err,
		)
	}

	return nil
}
