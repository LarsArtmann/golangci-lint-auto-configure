package cli

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/spf13/cobra"
)

const formatJSON = "json"

// newAnalyzeCommand creates the analyze command.
func newAnalyzeCommand(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze golangci-lint configuration and show recommendations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Find config file if not specified
			configFile := configPath
			if configFile == "" {
				var err error

				configFile, err = configLoader.FindConfigFile(".")
				if err != nil {
					return fmt.Errorf("failed to find config file: %w", err)
				}
			}

			configLoader.HasMultipleConfigFiles(".")

			logger.Infof("Analyzing configuration: %s", configFile)

			// Perform analysis
			analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)
			if err != nil {
				return fmt.Errorf("failed to analyze config: %w", err)
			}

			// Output based on format
			switch format {
			case formatJSON:
				data, jsonErr := json.MarshalIndent(analysis, "", "  ")
				if jsonErr != nil {
					return fmt.Errorf("failed to marshal analysis to JSON: %w", jsonErr)
				}

				logger.Info(string(data))
			default:
				// Default text output
				recommendations := analyzer.FormatRecommendations(analysis)
				summary := analyzer.GetSummary(analysis)

				logger.Info("\n" + recommendations)
				logger.Infof("Summary: %s", summary)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json)")

	return cmd
}
