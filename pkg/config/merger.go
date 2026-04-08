package config

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"

	"charm.land/log/v2"
	"github.com/spf13/afero"
)

// ConfigMerger handles merging multiple golangci-lint configuration files.
type ConfigMerger struct {
	logger *log.Logger
	fs     afero.Fs
}

// NewConfigMerger creates a new configuration merger.
func NewConfigMerger(logger *log.Logger) *ConfigMerger {
	return &ConfigMerger{
		logger: logger,
		fs:     afero.NewOsFs(),
	}
}

// NewConfigMergerWithFS creates a new configuration merger with a custom filesystem.
func NewConfigMergerWithFS(logger *log.Logger, fs afero.Fs) *ConfigMerger {
	return &ConfigMerger{
		logger: logger,
		fs:     fs,
	}
}

// MergeResult represents the result of a merge operation.
type MergeResult struct {
	PrimaryConfig    string   `json:"primary_config"`
	MergedConfigs    []string `json:"merged_configs"`
	RemovedConfigs   []string `json:"removed_configs,omitempty"`
	ChangesApplied   int      `json:"changes_applied"`
	MergedLinters    []string `json:"merged_linters,omitempty"`
	MergedFormatters []string `json:"merged_formatters,omitempty"`
	Success          bool     `json:"success"`
	Error            error    `json:"-"`
}

// IsSuccess returns true if the merge was successful.
func (m *MergeResult) IsSuccess() bool {
	return m.Error == nil && m.Success
}

// MergeConfigs loads and merges multiple config files.
// The first config in the list has the highest priority (per golangci-lint search order).
// Returns the merged config and a result describing what was merged.
func (cm *ConfigMerger) MergeConfigs(configPaths []string) (*Config, *MergeResult, error) {
	if len(configPaths) == 0 {
		return nil, nil, errors.New("no config files to merge")
	}

	if len(configPaths) == 1 {
		loader := NewLoaderWithFS(cm.logger, cm.fs)

		config, err := loader.LoadConfig(configPaths[0])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load primary config: %w", err)
		}

		return config, &MergeResult{
			PrimaryConfig: configPaths[0],
			Success:       true,
		}, nil
	}

	// Sort configs by priority (golangci-lint search order)
	sortedPaths := cm.sortByPriority(configPaths)
	primaryPath := sortedPaths[0]
	secondaryPaths := sortedPaths[1:]

	cm.logger.Infof("Merging %d config files, primary: %s", len(configPaths), primaryPath)

	// Load primary config
	loader := NewLoaderWithFS(cm.logger, cm.fs)

	primaryConfig, err := loader.LoadConfig(primaryPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load primary config %s: %w", primaryPath, err)
	}

	result := &MergeResult{
		PrimaryConfig:  primaryPath,
		MergedConfigs:  secondaryPaths,
		ChangesApplied: 0,
		Success:        true,
	}

	// Merge secondary configs into primary
	for _, secondaryPath := range secondaryPaths {
		secondaryConfig, err := loader.LoadConfig(secondaryPath)
		if err != nil {
			cm.logger.Warnf("Failed to load secondary config %s: %v", secondaryPath, err)

			continue
		}

		changes := cm.mergeConfigInto(primaryConfig, secondaryConfig)
		result.ChangesApplied += changes

		cm.logger.Debugf("Merged %d settings from %s", changes, secondaryPath)
	}

	return primaryConfig, result, nil
}

// sortByPriority sorts config paths by golangci-lint search order priority.
// Lower index = higher priority.
func (cm *ConfigMerger) sortByPriority(paths []string) []string {
	priorityMap := map[string]int{
		".golangci.yml":  0,
		".golangci.yaml": 1,
		".golangci.toml": 2,
		".golangci.json": 3,
	}

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

// mergeConfigInto merges secondary config into primary.
// Primary values take precedence; secondary fills in gaps.
// Returns the number of changes applied.
func (cm *ConfigMerger) mergeConfigInto(primary, secondary *Config) int {
	changes := 0

	// Merge Run settings
	changes += cm.mergeRunConfig(&primary.Run, &secondary.Run)

	// Merge Linters settings
	changes += cm.mergeLintersConfig(&primary.Linters, &secondary.Linters)

	// Merge Formatters settings
	changes += cm.mergeFormattersConfig(&primary.Formatters, &secondary.Formatters)

	// Merge Output settings
	changes += cm.mergeOutputConfig(&primary.Output, &secondary.Output)

	// Merge Issues settings
	changes += cm.mergeIssuesConfig(&primary.Issues, &secondary.Issues)

	return changes
}

func (cm *ConfigMerger) mergeRunConfig(primary, secondary *RunConfig) int {
	changes := 0

	// Only merge if primary has zero values and secondary has values
	if primary.Timeout == "" && secondary.Timeout != "" {
		primary.Timeout = secondary.Timeout
		changes++
	}

	if primary.Go == "" && secondary.Go != "" {
		primary.Go = secondary.Go
		changes++
	}

	if len(primary.BuildTags) == 0 && len(secondary.BuildTags) > 0 {
		primary.BuildTags = secondary.BuildTags
		changes++
	}

	if primary.ModulesDownloadMode == "" && secondary.ModulesDownloadMode != "" {
		primary.ModulesDownloadMode = secondary.ModulesDownloadMode
		changes++
	}

	if !primary.AllowParallelRunners && secondary.AllowParallelRunners {
		primary.AllowParallelRunners = secondary.AllowParallelRunners
		changes++
	}

	if !primary.AllowSerialRunners && secondary.AllowSerialRunners {
		primary.AllowSerialRunners = secondary.AllowSerialRunners
		changes++
	}

	if primary.IssuesExitCode == 0 && secondary.IssuesExitCode != 0 {
		primary.IssuesExitCode = secondary.IssuesExitCode
		changes++
	}

	if !primary.Tests && secondary.Tests {
		primary.Tests = secondary.Tests
		changes++
	}

	if primary.Concurrency == 0 && secondary.Concurrency != 0 {
		primary.Concurrency = secondary.Concurrency
		changes++
	}

	if primary.RelativePathMode == "" && secondary.RelativePathMode != "" {
		primary.RelativePathMode = secondary.RelativePathMode
		changes++
	}

	return changes
}

func (cm *ConfigMerger) mergeLintersConfig(primary, secondary *LintersConfig) int {
	changes := 0

	// Merge enabled linters (union, but primary takes precedence for conflicts)
	if len(primary.Enable) == 0 && len(secondary.Enable) > 0 {
		primary.Enable = secondary.Enable
		changes++
	} else if len(secondary.Enable) > 0 {
		// Add linters from secondary that aren't in primary
		primarySet := make(map[string]struct{}, len(primary.Enable))
		for _, l := range primary.Enable {
			primarySet[l] = struct{}{}
		}

		for _, l := range secondary.Enable {
			if _, exists := primarySet[l]; !exists {
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
		primarySet := make(map[string]struct{}, len(primary.Disable))
		for _, l := range primary.Disable {
			primarySet[l] = struct{}{}
		}

		for _, l := range secondary.Disable {
			if _, exists := primarySet[l]; !exists {
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
	if len(primary.Settings) == 0 && len(secondary.Settings) > 0 {
		primary.Settings = secondary.Settings
		changes++
	} else if len(secondary.Settings) > 0 {
		for key, value := range secondary.Settings {
			if _, exists := primary.Settings[key]; !exists {
				primary.Settings[key] = value
				changes++
			}
		}
	}

	// Merge exclusions (complex structure)
	changes += cm.mergeLintersExclusions(&primary.Exclusions, &secondary.Exclusions)

	return changes
}

func (cm *ConfigMerger) mergeLintersExclusions(primary, secondary *LintersExclusionsConfig) int {
	changes := 0

	if primary.Generated == "" && secondary.Generated != "" {
		primary.Generated = secondary.Generated
		changes++
	}

	if !primary.WarnUnused && secondary.WarnUnused {
		primary.WarnUnused = secondary.WarnUnused
		changes++
	}

	if len(primary.Presets) == 0 && len(secondary.Presets) > 0 {
		primary.Presets = secondary.Presets
		changes++
	} else if len(secondary.Presets) > 0 {
		primarySet := make(map[string]struct{}, len(primary.Presets))
		for _, p := range primary.Presets {
			primarySet[p] = struct{}{}
		}

		for _, p := range secondary.Presets {
			if _, exists := primarySet[p]; !exists {
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
		primary.Paths = append(primary.Paths, secondary.Paths...)
		changes += len(secondary.Paths)
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

func (cm *ConfigMerger) mergeFormattersConfig(primary, secondary *FormattersConfig) int {
	changes := 0

	// Merge enabled formatters
	if len(primary.Enable) == 0 && len(secondary.Enable) > 0 {
		primary.Enable = secondary.Enable
		changes++
	} else if len(secondary.Enable) > 0 {
		primarySet := make(map[string]struct{}, len(primary.Enable))
		for _, f := range primary.Enable {
			primarySet[f] = struct{}{}
		}

		for _, f := range secondary.Enable {
			if _, exists := primarySet[f]; !exists {
				primary.Enable = append(primary.Enable, f)
				changes++
			}
		}

		sort.Strings(primary.Enable)
	}

	// Merge disabled formatters
	if len(primary.Disable) == 0 && len(secondary.Disable) > 0 {
		primary.Disable = secondary.Disable
		changes++
	} else if len(secondary.Disable) > 0 {
		primarySet := make(map[string]struct{}, len(primary.Disable))
		for _, f := range primary.Disable {
			primarySet[f] = struct{}{}
		}

		for _, f := range secondary.Disable {
			if _, exists := primarySet[f]; !exists {
				primary.Disable = append(primary.Disable, f)
				changes++
			}
		}

		sort.Strings(primary.Disable)
	}

	// Merge settings
	if len(primary.Settings) == 0 && len(secondary.Settings) > 0 {
		primary.Settings = secondary.Settings
		changes++
	} else if len(secondary.Settings) > 0 {
		for key, value := range secondary.Settings {
			if _, exists := primary.Settings[key]; !exists {
				primary.Settings[key] = value
				changes++
			}
		}
	}

	// Merge exclusions
	changes += cm.mergeFormattersExclusions(&primary.Exclusions, &secondary.Exclusions)

	return changes
}

func (cm *ConfigMerger) mergeFormattersExclusions(primary, secondary *FormattersExclusionsConfig) int {
	changes := 0

	if primary.Generated == "" && secondary.Generated != "" {
		primary.Generated = secondary.Generated
		changes++
	}

	if !primary.WarnUnused && secondary.WarnUnused {
		primary.WarnUnused = secondary.WarnUnused
		changes++
	}

	if len(primary.Paths) == 0 && len(secondary.Paths) > 0 {
		primary.Paths = secondary.Paths
		changes++
	} else if len(secondary.Paths) > 0 {
		primary.Paths = append(primary.Paths, secondary.Paths...)
		changes += len(secondary.Paths)
	}

	return changes
}

func (cm *ConfigMerger) mergeOutputConfig(primary, secondary *OutputConfig) int {
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

func (cm *ConfigMerger) mergeIssuesConfig(primary, secondary *IssuesConfig) int {
	changes := 0

	// Note: 0 is a valid value (disable), so we check if primary hasn't been explicitly set
	// We use a heuristic: if both MaxIssuesPerLinter and MaxSameIssues are 0, assume unset
	if primary.MaxIssuesPerLinter == 0 && secondary.MaxIssuesPerLinter != 0 {
		primary.MaxIssuesPerLinter = secondary.MaxIssuesPerLinter
		changes++
	}

	if primary.MaxSameIssues == 0 && secondary.MaxSameIssues != 0 {
		primary.MaxSameIssues = secondary.MaxSameIssues
		changes++
	}

	if primary.NewFromRev == "" && secondary.NewFromRev != "" {
		primary.NewFromRev = secondary.NewFromRev
		changes++
	}

	if primary.NewFromPatch == "" && secondary.NewFromPatch != "" {
		primary.NewFromPatch = secondary.NewFromPatch
		changes++
	}

	if primary.NewFromMergeBase == "" && secondary.NewFromMergeBase != "" {
		primary.NewFromMergeBase = secondary.NewFromMergeBase
		changes++
	}

	if !primary.New && secondary.New {
		primary.New = secondary.New
		changes++
	}

	if !primary.WholeFiles && secondary.WholeFiles {
		primary.WholeFiles = secondary.WholeFiles
		changes++
	}

	if !primary.Fix && secondary.Fix {
		primary.Fix = secondary.Fix
		changes++
	}

	if !primary.UniqByLine && secondary.UniqByLine {
		// Only set if secondary is true (default is usually true)
		primary.UniqByLine = secondary.UniqByLine
		changes++
	}

	return changes
}

// SaveMergedConfig saves the merged config and optionally removes secondary configs.
func (cm *ConfigMerger) SaveMergedConfig(config *Config, result *MergeResult, removeSecondary bool) error {
	loader := NewLoaderWithFS(cm.logger, cm.fs)

	// Save merged config to primary file
	err := loader.SaveConfig(config, result.PrimaryConfig)
	if err != nil {
		return fmt.Errorf("failed to save merged config: %w", err)
	}

	cm.logger.Infof("Saved merged config to %s", result.PrimaryConfig)

	// Optionally remove secondary configs
	if removeSecondary {
		for _, path := range result.MergedConfigs {
			err := cm.fs.Remove(path)
			if err != nil {
				cm.logger.Warnf("Failed to remove secondary config %s: %v", path, err)

				continue
			}

			result.RemovedConfigs = append(result.RemovedConfigs, path)
			cm.logger.Infof("Removed secondary config: %s", path)
		}
	}

	return nil
}

// GetUniqueStrings returns a sorted slice of unique strings.
func GetUniqueStrings(input []string) []string {
	if len(input) == 0 {
		return nil
	}

	uniqueMap := make(map[string]struct{}, len(input))
	for _, s := range input {
		uniqueMap[s] = struct{}{}
	}

	result := slices.Collect(maps.Keys(uniqueMap))
	sort.Strings(result)

	return result
}
