package config

import (
	"path/filepath"
	"slices"
	"sort"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// mergeSettingsMaps deep-merges secondary settings into primary, returning number of changes.
// Caller must ensure primary is non-nil when secondary has entries.
func mergeSettingsMaps(primary, secondary map[string]any) int {
	if len(secondary) == 0 {
		return 0
	}

	changes := 0

	for key, secondaryValue := range secondary {
		primaryValue, exists := primary[key]
		if !exists {
			primary[key] = secondaryValue
			changes++

			continue
		}

		primMap, primOK := primaryValue.(map[string]any)

		secMap, secOK := secondaryValue.(map[string]any)
		if primOK && secOK {
			changes += mergeSettingsMaps(primMap, secMap)

			continue
		}
	}

	return changes
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
func mergeSortedStringSlice[T ~string](primary, secondary []T) ([]T, int) {
	if len(primary) == 0 && len(secondary) > 0 {
		return secondary, 1
	}

	if len(secondary) == 0 {
		return primary, 0
	}

	primary, changes := mergeUniqueItems(primary, secondary)

	if changes > 0 {
		slices.Sort(primary)
	}

	return primary, changes
}

// mergeUniqueItems adds items from secondary to primary that don't already exist.
// Returns the updated slice and the number of changes made.
func mergeUniqueItems[T ~string](primary, secondary []T) ([]T, int) {
	primarySet := types.NewSet(primary...)
	changes := 0

	for _, item := range secondary {
		if primarySet.Contains(item) {
			continue
		}

		primary = append(primary, item)
		primarySet.Add(item)

		changes++
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

	var changes int

	*primary, changes = mergeUniqueItems(*primary, secondary)

	return changes
}

// sortByPriority sorts config paths by golangci-lint search order priority.
// Lower index = higher priority.
func sortByPriority(paths []string) []string {
	sorted := make([]string, 0, len(paths))
	sorted = append(sorted, paths...)

	sort.Slice(sorted, func(i, j int) bool {
		iPriority := configFilePriority[filepath.Base(sorted[i])]
		jPriority := configFilePriority[filepath.Base(sorted[j])]

		return iPriority < jPriority
	})

	return sorted
}

// configFilePriority defines the priority for each config filename.
// Built from constants.DefaultConfigFileNames to avoid duplicating the filenames.
var configFilePriority = buildConfigFilePriority()

func buildConfigFilePriority() map[string]int {
	priorities := []int{ConfigPriorityYML, ConfigPriorityYAML, ConfigPriorityTOML, ConfigPriorityJSON}

	m := make(map[string]int, len(constants.DefaultConfigFileNames))
	for i, name := range constants.DefaultConfigFileNames {
		m[name] = priorities[i]
	}

	return m
}

// createBackup creates a backup of the given config file.
func createBackup(fileSystem FS, path string) (string, error) {
	data, err := fileSystem.ReadFile(path)
	if err != nil {
		return "", apperrors.WrapClassifiedf(err, "config.backup_read",
			"failed to read config %s for backup", path)
	}

	backupPath := path + ".merge-backup"

	err = fileSystem.WriteFile(backupPath, data, backupFilePermission)
	if err != nil {
		return "", apperrors.WrapClassifiedf(err, "config.backup_write",
			"failed to write backup %s", backupPath)
	}

	return backupPath, nil
}
