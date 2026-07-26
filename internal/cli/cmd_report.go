package cli

import (
	"context"
	"os"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
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
			return runReport(
				cmd,
				builder.Logger(),
				builder.Analyzer(),
				builder.ConfigLoader(),
				builder.Flags(),
			)
		},
	)
}

func runReport(
	cmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	flags *Flags,
) error {
	setLogLevel(logger, flags.Verbose)

	configFile, err := resolveReportConfig(cmd, configLoader, logger, flags)
	if err != nil {
		return apperrors.WrapClassified(err, "report.resolve_config", "resolve config")
	}

	analysis, err := analyzeConfig(logger, cmd, analyzer, configFile, flags.ReportFormat)
	if err != nil {
		return apperrors.WrapClassified(err, "report.analyze", "analyze config")
	}

	return writeReport(cmd.Context(), analysis, logger, configFile,
		determineOutputPath(flags.OutputReport, flags.ReportFormat), flags.ReportFormat)
}

func resolveReportConfig(
	cmd *cobra.Command,
	configLoader *config.Loader,
	logger *log.Logger,
	flags *Flags,
) (string, error) {
	return resolveConfig(cmd.Context(), configLoader, logger, flags, "find config")
}

func writeReport(
	ctx context.Context,
	analysis *types.ConfigAnalysis,
	logger *log.Logger,
	configFile string,
	outputPath, format string,
) error {
	switch format {
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

func setLogLevel(logger *log.Logger, verbose bool) {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}
}

func analyzeConfig(
	logger *log.Logger,
	cmd *cobra.Command,
	analyzer *linter.Analyzer,
	configFile string,
	reportFormat string,
) (*types.ConfigAnalysis, error) {
	logger.Infof("Generating %s report for: %s", reportFormat, configFile)

	analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)
	if err != nil {
		return nil, apperrors.WrapClassifiedf(err, "report.analyze_config",
			"failed to analyze config (configPath=%s, reportFormat=%s)",
			configFile, reportFormat)
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
		return apperrors.WrapClassifiedf(err, "report.json",
			"failed to generate JSON report (configPath=%s, outputPath=%s)",
			configFile, outputPath)
	}

	return nil
}

func writeReportFile(outputPath, configFile, label string, data []byte) error {
	writeErr := os.WriteFile(outputPath, data, filePermOwnerOnly)
	if writeErr != nil {
		return apperrors.WrapClassifiedf(writeErr, "report.write_file",
			"failed to write %s report (configPath=%s, outputPath=%s)",
			label, configFile, outputPath)
	}

	return nil
}

func writeSARIFReport(
	analysis *types.ConfigAnalysis,
	outputPath, configFile string,
) error {
	sarif, err := appfinding.AnalysisToSARIF(analysis, Version)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "report.sarif",
			"failed to generate SARIF report (configPath=%s, outputPath=%s)",
			configFile, outputPath)
	}

	return writeReportFile(outputPath, configFile, "SARIF", sarif)
}

func writeFindingJSONReport(
	analysis *types.ConfigAnalysis,
	outputPath, configFile string,
) error {
	report, err := appfinding.AnalysisToReport(analysis, Version)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "report.finding",
			"failed to generate finding report (configPath=%s)", configFile)
	}

	data, err := report.PrettyJSON()
	if err != nil {
		return apperrors.WrapClassifiedf(err, "report.finding_json",
			"failed to generate finding JSON report (configPath=%s)", configFile)
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
		return apperrors.WrapClassifiedf(err, "report.html",
			"failed to generate HTML report (configPath=%s, outputPath=%s)",
			configFile, outputPath)
	}

	return nil
}
