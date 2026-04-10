package config

import "maps"

// mergeOutputConfig merges output configurations.
func (cm *Merger) mergeOutputConfig(primary, secondary *OutputConfig) int {
	changes := 0

	changes += mergeFormatMap(primary.Formats, secondary.Formats)

	if primary.PathPrefix == "" && secondary.PathPrefix != "" {
		primary.PathPrefix = secondary.PathPrefix
		changes++
	}

	if primary.PathMode == "" && secondary.PathMode != "" {
		primary.PathMode = secondary.PathMode
		changes++
	}

	if len(primary.SortOrder) == 0 && len(secondary.SortOrder) > 0 {
		primary.SortOrder = secondary.SortOrder
		changes++
	}

	if !primary.ShowStats && secondary.ShowStats {
		primary.ShowStats = secondary.ShowStats
		changes++
	}

	return changes
}

// mergeFormatMap merges secondary format map into primary.
func mergeFormatMap(primary, secondary map[string]any) int {
	if len(primary) == 0 && len(secondary) > 0 {
		maps.Copy(primary, secondary)

		return 1
	}

	if len(secondary) == 0 {
		return 0
	}

	return mergeMap(primary, secondary)
}
