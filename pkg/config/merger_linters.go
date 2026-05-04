package config

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// mergeLintersConfig merges linter configurations.
func (cm *Merger) mergeLintersConfig(primary, secondary *LintersConfig) int {
	changes := 0

	updated, fieldChanges := mergeSortedStringSlice(primary.Enable, secondary.Enable)
	primary.Enable = updated
	changes += fieldChanges

	updated, fieldChanges = mergeSortedStringSlice(primary.Disable, secondary.Disable)
	primary.Disable = updated
	changes += fieldChanges

	if primary.Default == "" && secondary.Default != "" {
		primary.Default = secondary.Default
		changes++
	}

	if primary.Settings == nil && len(secondary.Settings) > 0 {
		types.InitLintersSettings(primary)
	}

	changes += mergeSettingsMaps(primary.Settings, secondary.Settings)

	changes += cm.mergeLintersExclusions(&primary.Exclusions, &secondary.Exclusions)

	return changes
}

// mergeCommonExclusionFields merges fields common to both linter and formatter exclusions.
func mergeCommonExclusionFields[T any](
	primary, secondary *T,
	getGenerated func(*T) string,
	setGenerated func(*T, string),
	getWarnUnused func(*T) bool,
	setWarnUnused func(*T, bool),
) int {
	changes := 0

	if getGenerated(primary) == "" && getGenerated(secondary) != "" {
		setGenerated(primary, getGenerated(secondary))

		changes++
	}

	if !getWarnUnused(primary) && getWarnUnused(secondary) {
		setWarnUnused(primary, getWarnUnused(secondary))

		changes++
	}

	return changes
}

// mergeLintersExclusions merges linter exclusion configurations.
func (cm *Merger) mergeLintersExclusions(primary, secondary *LintersExclusionsConfig) int {
	changes := mergeCommonExclusionFields(
		primary, secondary,
		func(c *LintersExclusionsConfig) string { return c.Generated },
		func(c *LintersExclusionsConfig, v string) { c.Generated = v },
		func(c *LintersExclusionsConfig) bool { return c.WarnUnused },
		func(c *LintersExclusionsConfig, v bool) { c.WarnUnused = v },
	)

	changes += cm.mergeLintersExclusionPresets(primary, secondary)
	changes += cm.mergeLintersExclusionRules(primary, secondary)
	changes += cm.mergeLintersExclusionPaths(primary, secondary)

	return changes
}

func (cm *Merger) mergeLintersExclusionPresets(primary, secondary *LintersExclusionsConfig) int {
	if len(primary.Presets) == 0 && len(secondary.Presets) > 0 {
		primary.Presets = secondary.Presets

		return 1
	}

	if len(secondary.Presets) == 0 {
		return 0
	}

	primarySet := types.NewSet(primary.Presets...)
	changes := 0

	for _, preset := range secondary.Presets {
		if !primarySet.Contains(preset) {
			primary.Presets = append(primary.Presets, preset)
			changes++
		}
	}

	return changes
}

func (cm *Merger) mergeLintersExclusionRules(primary, secondary *LintersExclusionsConfig) int {
	if len(primary.Rules) == 0 && len(secondary.Rules) > 0 {
		primary.Rules = secondary.Rules

		return len(secondary.Rules)
	}

	if len(secondary.Rules) > 0 {
		primary.Rules = append(primary.Rules, secondary.Rules...)

		return len(secondary.Rules)
	}

	return 0
}

func (cm *Merger) mergeLintersExclusionPaths(primary, secondary *LintersExclusionsConfig) int {
	changes := mergePaths(&primary.Paths, secondary.Paths)

	if len(primary.PathsExcept) == 0 && len(secondary.PathsExcept) > 0 {
		primary.PathsExcept = secondary.PathsExcept
		changes++
	} else if len(secondary.PathsExcept) > 0 {
		primary.PathsExcept = append(primary.PathsExcept, secondary.PathsExcept...)
		changes += len(secondary.PathsExcept)
	}

	return changes
}
