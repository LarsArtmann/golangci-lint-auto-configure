package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
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
			return runAnalyze(cmd, logger, analyzer, configLoader, format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json)")

	return cmd
}

func runAnalysisWithSpinner(
	ctx context.Context,
	analyzer *linter.Analyzer,
	configFile string,
) (*types.ConfigAnalysis, error) {
	spinnerDone := make(chan bool, 1)
	go spinner("Analyzing configuration...", spinnerDone)

	analysis, err := analyzer.AnalyzeConfig(ctx, configFile)

	spinnerDone <- true
	fmt.Fprintf(os.Stdout, "\r\033[K")

	return analysis, err
}

func runAnalyze(
	cmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	format string,
) error {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

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

	analysis, err := runAnalysisWithSpinner(cmd.Context(), analyzer, configFile)
	if err != nil {
		return fmt.Errorf("failed to analyze config (format=%s): %w", format, err)
	}

	return outputAnalysis(analysis, format, configFile)
}

func outputAnalysis(analysis *types.ConfigAnalysis, format string, configFile string) error {
	switch format {
	case formatJSON:
		data, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			return fmt.Errorf(
				"failed to marshal analysis to JSON (format=%s): %w",
				format,
				err,
			)
		}

		fmt.Fprintln(os.Stdout, string(data))
	default:
		fmt.Fprint(os.Stdout, ui.FormatConfigHeader(configFile))
		fmt.Fprint(os.Stdout, ui.FormatRecommendations(analysis))
		fmt.Fprint(os.Stdout, ui.FormatSummary(analysis))
	}

	return nil
}
