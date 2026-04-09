package config

import (
	"sort"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// mergeLintersConfig merges linter configurations.
func (cm *Merger) mergeLintersConfig(primary, secondary *LintersConfig) int {
	changes := 0

	// Merge enabled linters (union, but primary takes precedence for conflicts)
	if len(primary.Enable) == 0 && len(secondary.Enable) > 0 {
		primary.Enable = secondary.Enable
		changes++
	} else if len(secondary.Enable) > 0 {
		// Add linters from secondary that aren't in primary
		primarySet := types.NewSet(primary.Enable...)

		for _, l := range secondary.Enable {
			if !primarySet.Contains(l) {
				primary.Enable = append(primary.Enable, l)
				changes++
			}
		}

		// Sort for consistency
		sort.Strings(primary.Enable)
	}

	// Merge disabled linters (union)
	if len(primary.Disable) == 0 && len(secondary.Disable) > 0 {
		primary.Disable = secondary.Disable
		changes++
	} else if len(secondary.Disable) > 0 {
		primarySet := types.NewSet(primary.Disable...)

		for _, l := range secondary.Disable {
			if !primarySet.Contains(l) {
				primary.Disable = append(primary.Disable, l)
				changes++
			}
		}

		sort.Strings(primary.Disable)
	}

	// Merge default setting
	if primary.Default == "" && secondary.Default != "" {
		primary.Default = secondary.Default
		changes++
	}

	// Merge settings (deep merge for linter-specific settings)
	changes += mergeSettingsMaps(primary.Settings, secondary.Settings)

	// Merge exclusions (complex structure)
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

	if len(primary.Presets) == 0 && len(secondary.Presets) > 0 {
		primary.Presets = secondary.Presets
		changes++
	} else if len(secondary.Presets) > 0 {
		primarySet := types.NewSet(primary.Presets...)

		for _, p := range secondary.Presets {
			if !primarySet.Contains(p) {
				primary.Presets = append(primary.Presets, p)
				changes++
			}
		}
	}

	if len(primary.Rules) == 0 && len(secondary.Rules) > 0 {
		primary.Rules = secondary.Rules
		changes++
	} else if len(secondary.Rules) > 0 {
		primary.Rules = append(primary.Rules, secondary.Rules...)
		changes += len(secondary.Rules)
	}

	if len(primary.Paths) == 0 && len(secondary.Paths) > 0 {
		primary.Paths = secondary.Paths
		changes++
	} else if len(secondary.Paths) > 0 {
		changes += mergeStringSlices(primary.Paths, secondary.Paths)
	}

	if len(primary.PathsExcept) == 0 && len(secondary.PathsExcept) > 0 {
		primary.PathsExcept = secondary.PathsExcept
		changes++
	} else if len(secondary.PathsExcept) > 0 {
		primary.PathsExcept = append(primary.PathsExcept, secondary.PathsExcept...)
		changes += len(secondary.PathsExcept)
	}

	return changes
}
