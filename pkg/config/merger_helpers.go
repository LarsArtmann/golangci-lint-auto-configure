package config

import (
	"fmt"
	"maps"
	"sort"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/spf13/afero"
)

// mergeSettingsMaps merges secondary settings into primary, returning number of changes.
func mergeSettingsMaps(primary, secondary map[string]any) int {
	changes := 0

	if len(primary) == 0 && len(secondary) > 0 {
		maps.Copy(primary, secondary)

		changes = len(secondary)
	} else if len(secondary) > 0 {
		for key, value := range secondary {
			if _, exists := primary[key]; !exists {
				primary[key] = value
				changes++
			}
		}
	}

	return changes
}

// mergeStringSlices merges secondary slice into primary slice, returning number of changes.
func mergeStringSlices(primary, secondary []string) int {
	if len(primary) == 0 && len(secondary) > 0 {
		return len(secondary)
	}

	if len(secondary) == 0 {
		return 0
	}

	primarySet := types.NewSet(primary...)
	changes := 0

	for _, p := range secondary {
		if !primarySet.Contains(p) {
			primary = append(primary, p)
			changes++
		}
	}

	return changes
}

// mergePaths merges secondary paths into primary paths, updating primary if empty.
// Returns the number of changes made.
func mergePaths(primary *[]string, secondary []string) int {
	if len(*primary) == 0 && len(secondary) > 0 {
		*primary = secondary

		return len(secondary)
	}

	if len(secondary) == 0 {
		return 0
	}

	primarySet := types.NewSet(*primary...)
	changes := 0

	for _, p := range secondary {
		if !primarySet.Contains(p) {
			*primary = append(*primary, p)
			changes++
		}
	}

	return changes
}

// sortByPriority sorts config paths by golangci-lint search order priority.
// Lower index = higher priority.
func sortByPriority(paths []string) []string {
	priorityMap := configPriorityMap()

	sorted := make([]string, len(paths))
	copy(sorted, paths)

	sort.Slice(sorted, func(i, j int) bool {
		iPriority := priorityMap[getFilename(sorted[i])]
		jPriority := priorityMap[getFilename(sorted[j])]

		return iPriority < jPriority
	})

	return sorted
}

func getFilename(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}

	return path
}

// configPriorityMap returns the priority for a config filename.
func configPriorityMap() map[string]int {
	return map[string]int{
		".golangci.yml":  ConfigPriorityYML,
		".golangci.yaml": ConfigPriorityYAML,
		".golangci.toml": ConfigPriorityTOML,
		".golangci.json": ConfigPriorityJSON,
	}
}

// createBackup creates a backup of the given config file.
func createBackup(fs afero.Fs, path string) (string, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return "", fmt.Errorf("failed to read config for backup: %w", err)
	}

	backupPath := path + ".merge-backup"

	err = afero.WriteFile(fs, backupPath, data, backupFilePermission)
	if err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}
