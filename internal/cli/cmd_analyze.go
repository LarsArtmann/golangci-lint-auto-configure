package cli

import (
	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/spf13/cobra"
)

// newAnalyzeCommand creates the analyze command.
func newAnalyzeCommand(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze golangci-lint configuration and show recommendations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Find config file if not specified
			configFile := configPath
			if configFile == "" {
				var err error

				configFile, err = configLoader.FindConfigFile(".")
				if err != nil {
					return err
				}
			}

			logger.Infof("Analyzing configuration: %s", configFile)

			// Perform analysis
			analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)
			if err != nil {
				return err
			}

			// Display recommendations
			recommendations := analyzer.FormatRecommendations(analysis)
			summary := analyzer.GetSummary(analysis)

			logger.Info("\n" + recommendations)
			logger.Infof("Summary: %s", summary)

			return nil
		},
	}

	return cmd
}
