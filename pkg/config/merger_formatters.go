package config

// mergeFormattersConfig merges formatter configurations.
func (cm *Merger) mergeFormattersConfig(primary, secondary *FormattersConfig) int {
	changes := 0

	updated, fieldChanges := mergeSortedStringSlice(primary.Enable, secondary.Enable)
	primary.Enable = updated
	changes += fieldChanges

	updated, fieldChanges = mergeSortedStringSlice(primary.Disable, secondary.Disable)
	primary.Disable = updated
	changes += fieldChanges

	changes += mergeSettingsMaps(primary.Settings, secondary.Settings)

	changes += cm.mergeFormattersExclusions(&primary.Exclusions, &secondary.Exclusions)

	return changes
}

// mergeFormattersExclusions merges formatter exclusion configurations.
func (cm *Merger) mergeFormattersExclusions(primary, secondary *FormattersExclusionsConfig) int {
	changes := mergeCommonExclusionFields(
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
