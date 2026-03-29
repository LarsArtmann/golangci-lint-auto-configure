package cli

import (
	"fmt"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/report"
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
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Find config file if not specified
			configFile, err := resolveConfigPath(configLoader, configPath)
			if err != nil {
				return fmt.Errorf("failed to find config file using analyzer (path=%s): %w", analyzer, err)
			}

			logger.Infof("Generating %s report for: %s", reportFormat, configFile)

			// Analyze configuration
			analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)
			if err != nil {
				return fmt.Errorf("failed to analyze config with analyzer (path=%s): %w", analyzer, err)
			}

			// Determine output path
			outputPath := outputReport
			if outputPath == "report.html" {
				if reportFormat == "json" {
					outputPath = "report.json"
				}
			}

			// Generate report based on format
			if reportFormat == "json" {
				jsonGenerator := report.NewJSONGenerator(logger)

				err := jsonGenerator.GenerateJSONReport(analysis, outputPath)
				if err != nil {
					return fmt.Errorf("failed to generate JSON report with analyzer (path=%s, outputPath=%s): %w", analyzer, outputPath, err)
				}
			} else {
				htmlGenerator := report.NewGenerator(logger)

				err := htmlGenerator.GenerateReport(cmd.Context(), analysis, outputPath)
				if err != nil {
					return fmt.Errorf("failed to generate HTML report with analyzer (path=%s, outputPath=%s): %w", analyzer, outputPath, err)
				}
			}

			return nil
		},
	}

	return cmd
}
