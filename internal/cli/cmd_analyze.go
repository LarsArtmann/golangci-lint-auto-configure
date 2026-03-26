package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/ui"
	"github.com/spf13/cobra"
)

const formatJSON = "json"

const (
	spinnerFrames      = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	spinnerDelayMillis = 50
)

// spinner prints a simple animated spinner.
func spinner(message string, done chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			for _, frame := range spinnerFrames {
				fmt.Fprintf(os.Stdout, "\r%c %s", frame, message)
				time.Sleep(time.Duration(spinnerDelayMillis) * time.Millisecond)
			}
		}
	}
}

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
					return fmt.Errorf("failed to find config file (format=%s): %w", format, err)
				}
			}

			configLoader.HasMultipleConfigFiles(".")

			logger.Infof("Analyzing configuration: %s", configFile)

			// Start spinner during analysis
			spinnerDone := make(chan bool, 1)
			go spinner("Analyzing configuration...", spinnerDone)

			// Perform analysis
			analysis, err := analyzer.AnalyzeConfig(cmd.Context(), configFile)

			// Stop spinner
			spinnerDone <- true

			fmt.Fprintf(os.Stdout, "\r\033[K") // Clear the line

			if err != nil {
				return fmt.Errorf("failed to analyze config (format=%s): %w", format, err)
			}

			// Output based on format
			switch format {
			case formatJSON:
				data, jsonErr := json.MarshalIndent(analysis, "", "  ")
				if jsonErr != nil {
					return fmt.Errorf("failed to marshal analysis to JSON (format=%s): %w", format, jsonErr)
				}

				fmt.Fprintln(os.Stdout, string(data))
			default:
				// Default styled text output
				fmt.Fprint(os.Stdout, ui.FormatConfigHeader(configFile))
				fmt.Fprint(os.Stdout, ui.FormatRecommendations(analysis))
				fmt.Fprint(os.Stdout, ui.FormatSummary(analysis))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json)")

	return cmd
}
