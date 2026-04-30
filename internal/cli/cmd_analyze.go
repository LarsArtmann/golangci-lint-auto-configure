package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	appfinding "github.com/larsartmann/golangci-lint-auto-configure/pkg/finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/ui"
	"github.com/spf13/cobra"
)

const (
	formatJSON    = "json"
	formatSARIF   = "sarif"
	formatFinding = "finding"
)

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

func newAnalyzeCommand(builder *CommandBuilder) *cobra.Command {
	var format string

	cmd := builder.Build(
		"analyze",
		"Analyze golangci-lint configuration and show recommendations",
		func(cmd *cobra.Command, _ []string) error {
			return runAnalyze(
				cmd,
				builder.Logger(),
				builder.Analyzer(),
				builder.ConfigLoader(),
				format,
			)
		},
	)

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json, sarif, finding)")

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

	if err != nil {
		return nil, fmt.Errorf("analyze config: %w", err)
	}

	return analysis, nil
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

func outputAnalysis(analysis *types.ConfigAnalysis, format, configFile string) error {
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
	case formatSARIF:
		return outputSARIF(analysis)
	case formatFinding:
		return outputFindingJSON(analysis)
	default:
		fmt.Fprint(os.Stdout, ui.FormatConfigHeader(configFile))
		fmt.Fprint(os.Stdout, ui.FormatRecommendations(analysis))
		fmt.Fprint(os.Stdout, ui.FormatSummary(analysis))
	}

	return nil
}

func outputSARIF(analysis *types.ConfigAnalysis) error {
	sarif, err := appfinding.AnalysisToSARIF(analysis, Version)
	if err != nil {
		return fmt.Errorf("failed to generate SARIF: %w", err)
	}

	fmt.Fprintln(os.Stdout, string(sarif))

	return nil
}

func outputFindingJSON(analysis *types.ConfigAnalysis) error {
	r := appfinding.AnalysisToReport(analysis, Version)

	data, err := r.PrettyJSON()
	if err != nil {
		return fmt.Errorf("failed to generate finding JSON: %w", err)
	}

	fmt.Fprintln(os.Stdout, data)

	return nil
}
