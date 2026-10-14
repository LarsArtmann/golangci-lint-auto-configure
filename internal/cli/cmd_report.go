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
				builder.Fixer(),
				builder.ConfigLoader(),
				builder.Flags(),
			)
		},
	)
}

func runReport(
	cmd *cobra.Command,
	logger *log.Logger,
	fixer *linter.Fixer,
	configLoader *config.Loader,
	flags *Flags,
) error {
	setLogLevel(logger, flags.Verbose)

	configFile, err := resolveReportConfig(cmd, configLoader, logger, flags)
	if err != nil {
		return apperrors.WrapClassified(err, "report.resolve_config", "resolve config")
	}

	logger.Infof("Generating %s report for: %s", flags.ReportFormat, configFile)

	evaluation, err := fixer.Evaluate(cmd.Context(), configFile, types.LinterPriorityOptional)
	if err != nil {
		return apperrors.WrapClassified(err, "report.evaluate", "evaluate config")
	}

	return writeReport(cmd.Context(), evaluation, logger, configFile,
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
	evaluation *linter.Evaluation,
	logger *log.Logger,
	configFile string,
	outputPath, format string,
) error {
	switch format {
	case "json":
		return writeJSONReport(logger, evaluation, outputPath, configFile)
	case formatSARIF:
		return writeSARIFReport(evaluation, outputPath, configFile)
	case formatFinding:
		return writeFindingJSONReport(evaluation, outputPath, configFile)
	default:
		return writeHTMLReport(ctx, logger, evaluation, outputPath, configFile)
	}
}

func setLogLevel(logger *log.Logger, verbose bool) {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}
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
	evaluation *linter.Evaluation,
	outputPath, configFile string,
) error {
	gen := report.NewJSONGenerator(logger)

	err := gen.GenerateJSONReport(evaluation, outputPath)
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
	evaluation *linter.Evaluation,
	outputPath, configFile string,
) error {
	sarif, err := appfinding.EvaluationToSARIF(evaluation, Version)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "report.sarif",
			"failed to generate SARIF report (configPath=%s, outputPath=%s)",
			configFile, outputPath)
	}

	return writeReportFile(outputPath, configFile, "SARIF", sarif)
}

func writeFindingJSONReport(
	evaluation *linter.Evaluation,
	outputPath, configFile string,
) error {
	report, err := appfinding.EvaluationToReport(evaluation, Version)
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
	evaluation *linter.Evaluation,
	outputPath, configFile string,
) error {
	gen := report.NewGenerator(logger)

	err := gen.GenerateReport(ctx, evaluation, outputPath)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "report.html",
			"failed to generate HTML report (configPath=%s, outputPath=%s)",
			configFile, outputPath)
	}

	return nil
}
