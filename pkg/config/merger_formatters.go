package config

// mergeFormattersConfig merges formatter configurations.
func (cm *Merger) mergeFormattersConfig(primary, secondary *FormattersConfig) int {
	changes := mergeEnableDisable(&primary.Enable, &primary.Disable, secondary.Enable, secondary.Disable)

	changes += mergeSettingsMaps(primary.Settings, secondary.Settings)

	changes += cm.mergeFormattersExclusions(&primary.Exclusions, &secondary.Exclusions)

	return changes
}

// mergeEnableDisable merges enable and disable slices, returning the total change count.
func mergeEnableDisable(primaryEnable, primaryDisable *[]string, secondaryEnable, secondaryDisable []string) int {
	changes := 0

	updated, fieldChanges := mergeSortedStringSlice(*primaryEnable, secondaryEnable)
	*primaryEnable = updated
	changes += fieldChanges

	updated, fieldChanges = mergeSortedStringSlice(*primaryDisable, secondaryDisable)
	*primaryDisable = updated
	changes += fieldChanges

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

	changes += mergePaths(&primary.Paths, secondary.Paths)

	return changes
}
