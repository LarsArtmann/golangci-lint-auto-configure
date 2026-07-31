package config

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// mergeLintersConfig merges linter configurations.
func (cm *Merger) mergeLintersConfig(primary, secondary *types.LintersConfig) int {
	changes := mergeEnableDisable(&primary.Enable, &primary.Disable, secondary.Enable, secondary.Disable)

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
func (cm *Merger) mergeLintersExclusions(primary, secondary *types.LintersExclusionsConfig) int {
	changes := mergeCommonExclusionFields(
		primary, secondary,
		func(c *types.LintersExclusionsConfig) string { return c.Generated },
		func(c *types.LintersExclusionsConfig, v string) { c.Generated = v },
		func(c *types.LintersExclusionsConfig) bool { return c.WarnUnused },
		func(c *types.LintersExclusionsConfig, v bool) { c.WarnUnused = v },
	)

	changes += cm.mergeLintersExclusionPresets(primary, secondary)
	changes += cm.mergeLintersExclusionRules(primary, secondary)
	changes += cm.mergeLintersExclusionPaths(primary, secondary)

	return changes
}

func (cm *Merger) mergeLintersExclusionPresets(primary, secondary *types.LintersExclusionsConfig) int {
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

func (cm *Merger) mergeLintersExclusionRules(primary, secondary *types.LintersExclusionsConfig) int {
	if len(secondary.Rules) == 0 {
		return 0
	}

	if len(primary.Rules) == 0 {
		primary.Rules = secondary.Rules

		return len(secondary.Rules)
	}

	keyIndex := buildRuleKeyIndex(primary.Rules)

	return mergeSecondaryRules(&primary.Rules, secondary.Rules, keyIndex)
}

func buildRuleKeyIndex(rules []types.ExclusionRuleConfig) map[string]int {
	index := make(map[string]int, len(rules))
	for i := range rules {
		index[rules[i].RuleKey()] = i
	}

	return index
}

func mergeSecondaryRules(
	rules *[]types.ExclusionRuleConfig,
	secondary []types.ExclusionRuleConfig,
	keyIndex map[string]int,
) int {
	changes := 0

	for _, rule := range secondary {
		key := rule.RuleKey()
		if idx, exists := keyIndex[key]; exists {
			merged, added := mergeUniqueItems((*rules)[idx].Linters, rule.Linters)
			if added > 0 {
				(*rules)[idx].Linters = merged
				changes++
			}
		} else {
			*rules = append(*rules, rule)
			keyIndex[key] = len(*rules) - 1
			changes++
		}
	}

	return changes
}

func (cm *Merger) mergeLintersExclusionPaths(primary, secondary *types.LintersExclusionsConfig) int {
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
