package cli

import (
	"sort"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/spf13/cobra"
)

func newPresetsCommand(builder *CommandBuilder) *cobra.Command {
	return builder.Build(
		"presets",
		"List all available configuration presets",
		func(_ *cobra.Command, _ []string) error {
			return runListPresets(builder.Logger())
		},
	)
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
