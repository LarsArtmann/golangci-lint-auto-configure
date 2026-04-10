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
	if len(primary) == 0 && len(secondary) > 0 {
		maps.Copy(primary, secondary)

		changes := len(secondary)

		return changes
	}

	if len(secondary) == 0 {
		return 0
	}

	return mergeMap(primary, secondary)
}

// mergeMap merges secondary map into primary, adding missing keys.
// Returns the number of new keys added.
func mergeMap[T any](primary, secondary map[string]T) int {
	changes := 0

	for key, value := range secondary {
		if _, exists := primary[key]; !exists {
			primary[key] = value
			changes++
		}
	}

	return changes
}

// mergeSortedStringSlice merges secondary into primary string slice with deduplication and sorting.
// Returns the number of changes made.
func mergeSortedStringSlice(primary, secondary []string) ([]string, int) {
	if len(primary) == 0 && len(secondary) > 0 {
		return secondary, 1
	}

	if len(secondary) == 0 {
		return primary, 0
	}

	primary, changes := mergeUniqueItems(primary, secondary)

	if changes > 0 {
		sort.Strings(primary)
	}

	return primary, changes
}

// mergeStringSetSlice merges secondary into primary string slice with deduplication (no sorting).
// Returns the number of changes made.
func mergeStringSetSlice(primary, secondary []string) (int, []string) {
	if len(primary) == 0 && len(secondary) > 0 {
		return len(secondary), secondary
	}

	if len(secondary) == 0 {
		return 0, primary
	}

	_, changes := mergeUniqueItems(primary, secondary)

	return changes, primary
}

// mergeUniqueItems adds items from secondary to primary that don't already exist.
// Returns the updated slice and the number of changes made.
func mergeUniqueItems(primary, secondary []string) ([]string, int) {
	primarySet := types.NewSet(primary...)
	changes := 0

	for _, item := range secondary {
		if !primarySet.Contains(item) {
			primary = append(primary, item)
			changes++
		}
	}

	return primary, changes
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

	*primary, _ = mergeUniqueItems(*primary, secondary)

	return len(secondary) - countDuplicates(*primary, secondary)
}

// countDuplicates returns the number of items in secondary that already exist in primary.
func countDuplicates(primary []string, secondary []string) int {
	primarySet := types.NewSet(primary...)
	duplicates := 0

	for _, item := range secondary {
		if primarySet.Contains(item) {
			duplicates++
		}
	}

	return duplicates
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
func createBackup(fileSystem afero.Fs, path string) (string, error) {
	data, err := afero.ReadFile(fileSystem, path)
	if err != nil {
		return "", fmt.Errorf("failed to read config for backup: %w", err)
	}

	backupPath := path + ".merge-backup"

	err = afero.WriteFile(fileSystem, backupPath, data, backupFilePermission)
	if err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}
