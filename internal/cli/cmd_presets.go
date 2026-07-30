package cli

import (
	"encoding/json/v2"
	"sort"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
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

	cmd.Flags().
		BoolVar(&jsonOutput, "json", false, "Output presets as JSON for scripting")
		//art-dupl:accept Cobra flag registration boilerplate

	return cmd
}

type presetEntry struct {
	Name        string   `json:"Name"`
	Description string   `json:"Description"`
	Linters     []string `json:"Linters"`
	Formatters  []string `json:"Formatters"`
}

func outputPresetsJSON() error {
	presetNames := make([]string, 0, len(constants.PresetDescriptions))

	for name := range constants.PresetDescriptions {
		presetNames = append(presetNames, name)
	}

	sort.Strings(presetNames)

	entries := make([]presetEntry, 0, len(presetNames))

	for _, name := range presetNames {
		entries = append(entries, buildPresetEntry(name))
	}

	output := map[string]any{"Presets": entries}

	data, err := json.Marshal(output)
	if err != nil {
		return errorfamily.WrapCorruptionf(err, "presets.marshal_json", "marshal presets JSON")
	}

	printBytesToStdout(data)

	return nil
}

func buildPresetEntry(name string) presetEntry {
	linters := constants.PresetLinters[name]

	linterStrs := make([]string, 0, len(linters))
	for _, l := range linters {
		linterStrs = append(linterStrs, string(l))
	}

	formatterStrs := []string{}
	if formatters, ok := constants.PresetFormatters[name]; ok {
		formatterStrs = make([]string, 0, len(formatters))
		for _, f := range formatters {
			formatterStrs = append(formatterStrs, string(f))
		}
	}

	return presetEntry{
		Name:        name,
		Description: constants.PresetDescriptions[name],
		Linters:     linterStrs,
		Formatters:  formatterStrs,
	}
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
