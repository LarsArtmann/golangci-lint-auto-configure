package cli

import (
	"context"
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	appfinding "github.com/larsartmann/golangci-lint-auto-configure/pkg/finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/report"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/spf13/cobra"
)

const filePermOwnerOnly = 0o600

func newReportCommand(builder *CommandBuilder) *cobra.Command {
	return builder.Build(
		"report",
		"Generate HTML report of configuration",
		func(cmd *cobra.Command, _ []string) error {
			return runReport(cmd, builder.Logger(), builder.Analyzer(), builder.ConfigLoader())
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

	configFile, err := resolveReportConfig(cmd, configLoader, logger)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}

	analysis, err := analyzeConfig(logger, cmd, analyzer, configFile)
	if err != nil {
		return fmt.Errorf("analyze config: %w", err)
	}

	return writeReport(cmd.Context(), analysis, logger, configFile)
}

func resolveReportConfig(
	cmd *cobra.Command,
	configLoader *config.Loader,
	logger *log.Logger,
) (string, error) {
	return resolveConfig(cmd.Context(), configLoader, logger, "find config")
}

func writeReport(
	ctx context.Context,
	analysis *types.ConfigAnalysis,
	logger *log.Logger,
	configFile string,
) error {
	outputPath := determineOutputPath(outputReport, reportFormat)

	switch reportFormat {
	case "json":
		return writeJSONReport(logger, analysis, outputPath, configFile)
	case formatSARIF:
		return writeSARIFReport(analysis, outputPath, configFile)
	case formatFinding:
		return writeFindingJSONReport(analysis, outputPath, configFile)
	default:
		return writeHTMLReport(ctx, logger, analysis, outputPath, configFile)
	}
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
		return nil, fmt.Errorf(
			"failed to analyze config (configPath=%s, reportFormat=%s): %w",
			configFile,
			reportFormat,
			err,
		)
	}

	return analysis, nil
}

func determineOutputPath(outputPath, format string) string {
	if outputPath == "report.html" {
		switch format {
		case "json":
			return "report.json"
		case formatSARIF:
			return "report.sarif.json"
		case formatFinding:
			return "report.finding.json"
		}
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

func writeReportFile(outputPath, configFile, label string, data []byte) error {
	writeErr := os.WriteFile(outputPath, data, filePermOwnerOnly)
	if writeErr != nil {
		return fmt.Errorf(
			"failed to write %s report (configPath=%s, outputPath=%s): %w",
			label,
			configFile,
			outputPath,
			writeErr,
		)
	}

	return nil
}

func writeSARIFReport(
	analysis *types.ConfigAnalysis,
	outputPath, configFile string,
) error {
	sarif, err := appfinding.AnalysisToSARIF(analysis, Version)
	if err != nil {
		return fmt.Errorf(
			"failed to generate SARIF report (configPath=%s, outputPath=%s): %w",
			configFile,
			outputPath,
			err,
		)
	}

	return writeReportFile(outputPath, configFile, "SARIF", sarif)
}

func writeFindingJSONReport(
	analysis *types.ConfigAnalysis,
	outputPath, configFile string,
) error {
	r := appfinding.AnalysisToReport(analysis, Version)

	data, err := r.PrettyJSON()
	if err != nil {
		return fmt.Errorf(
			"failed to generate finding JSON report (configPath=%s): %w",
			configFile,
			err,
		)
	}

	return writeReportFile(outputPath, configFile, "finding JSON", []byte(data))
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
