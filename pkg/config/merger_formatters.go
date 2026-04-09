package config

import (
	"sort"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// mergeFormattersConfig merges formatter configurations.
func (cm *Merger) mergeFormattersConfig(primary, secondary *FormattersConfig) int {
	changes := 0

	// Merge enabled formatters
	if len(primary.Enable) == 0 && len(secondary.Enable) > 0 {
		primary.Enable = secondary.Enable
		changes++
	} else if len(secondary.Enable) > 0 {
		primarySet := types.NewSet(primary.Enable...)

		for _, f := range secondary.Enable {
			if !primarySet.Contains(f) {
				primary.Enable = append(primary.Enable, f)
				changes++
			}
		}

		sort.Strings(primary.Enable)
	}

	// Merge disabled formatters
	if len(primary.Disable) == 0 && len(secondary.Disable) > 0 {
		primary.Disable = secondary.Disable
		changes++
	} else if len(secondary.Disable) > 0 {
		primarySet := types.NewSet(primary.Disable...)

		for _, f := range secondary.Disable {
			if !primarySet.Contains(f) {
				primary.Disable = append(primary.Disable, f)
				changes++
			}
		}

		sort.Strings(primary.Disable)
	}

	// Merge settings
	changes += mergeSettingsMaps(primary.Settings, secondary.Settings)

	// Merge exclusions
	changes += cm.mergeFormattersExclusions(&primary.Exclusions, &secondary.Exclusions)

	return changes
}

// mergeFormattersExclusions merges formatter exclusion configurations.
func (cm *Merger) mergeFormattersExclusions(primary, secondary *FormattersExclusionsConfig) int {
	changes := MergeCommonExclusionFields(
		primary, secondary,
		func(c *FormattersExclusionsConfig) string { return c.Generated },
		func(c *FormattersExclusionsConfig, v string) { c.Generated = v },
		func(c *FormattersExclusionsConfig) bool { return c.WarnUnused },
		func(c *FormattersExclusionsConfig, v bool) { c.WarnUnused = v },
	)

	if len(primary.Paths) == 0 && len(secondary.Paths) > 0 {
		primary.Paths = secondary.Paths
		changes++
	} else if len(secondary.Paths) > 0 {
		changes += mergePaths(&primary.Paths, secondary.Paths)
	}

	return changes
}
