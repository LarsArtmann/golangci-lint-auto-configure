package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/spf13/cobra"
)

func newPresetsCommand(builder *CommandBuilder) *cobra.Command {
	var jsonOutput bool

	cmd := builder.Build(
		"presets",
		"List all available configuration presets",
		func(_ *cobra.Command, _ []string) error {
			if jsonOutput {
				return outputPresetsJSON()
			}

			return runListPresets(builder.Logger())
		},
	)

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output presets as JSON for scripting")

	return cmd
}

type presetEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Linters     []string `json:"linters"`
	Formatters  []string `json:"formatters"`
}

func outputPresetsJSON() error {
	presetNames := make([]string, 0, len(constants.PresetDescriptions))

	for name := range constants.PresetDescriptions {
		presetNames = append(presetNames, name)
	}

	sort.Strings(presetNames)

	entries := make([]presetEntry, 0, len(presetNames))

	for _, name := range presetNames {
		linters := constants.PresetLinters[name]
		linterStrs := make([]string, len(linters))
		for i, l := range linters {
			linterStrs[i] = string(l)
		}

		formatterStrs := []string{}
		if formatters, ok := constants.PresetFormatters[name]; ok {
			formatterStrs = make([]string, len(formatters))
			for i, f := range formatters {
				formatterStrs[i] = string(f)
			}
		}

		entries = append(entries, presetEntry{
			Name:        name,
			Description: constants.PresetDescriptions[name],
			Linters:     linterStrs,
			Formatters:  formatterStrs,
		})
	}

	output := map[string]any{"presets": entries}

	data, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshal presets JSON: %w", err)
	}

	fmt.Fprintln(os.Stdout, string(data))

	return nil
}

func runListPresets(logger *log.Logger) error {
	presetNames := make([]string, 0, len(constants.PresetDescriptions))

	for name := range constants.PresetDescriptions {
		presetNames = append(presetNames, name)
	}

	sort.Strings(presetNames)

	for _, name := range presetNames {
		desc := constants.PresetDescriptions[name]
		linterCount := len(constants.PresetLinters[name])

		formatterCount := 0
		if formatters, ok := constants.PresetFormatters[name]; ok {
			formatterCount = len(formatters)
		}

		logger.Infof("  %-15s %s", name, desc)

		if formatterCount > 0 {
			logger.Infof("  %-15s %d linters, %d formatters", "", linterCount, formatterCount)
		} else {
			logger.Infof("  %-15s %d linters", "", linterCount)
		}
	}

	return nil
}
