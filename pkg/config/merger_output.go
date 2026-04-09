package config

// mergeOutputConfig merges output configurations.
func (cm *Merger) mergeOutputConfig(primary, secondary *OutputConfig) int {
	changes := 0

	// Merge formats
	if len(primary.Formats) == 0 && len(secondary.Formats) > 0 {
		primary.Formats = secondary.Formats
		changes++
	} else if len(secondary.Formats) > 0 {
		for key, value := range secondary.Formats {
			if _, exists := primary.Formats[key]; !exists {
				primary.Formats[key] = value
				changes++
			}
		}
	}

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
