package config

import (
	"errors"
	"fmt"
	"os"

	"charm.land/log/v2"
)

// Backup file permission (read/write for owner only).
const backupFilePermission = os.FileMode(0o600)

// ErrNoConfigFiles is returned when no config files are provided to merge.
var ErrNoConfigFiles = errors.New("no config files to merge")

// Merger handles merging multiple golangci-lint configuration files.
type Merger struct {
	logger *log.Logger
	fs     FS
}

// NewMerger creates a new configuration merger.
func NewMerger(logger *log.Logger) *Merger {
	return &Merger{
		logger: logger,
		fs:     osFS{},
	}
}

// MergeResult represents the result of a merge operation.
type MergeResult struct {
	PrimaryConfig    string            `json:"primary_config"`
	MergedConfigs    []string          `json:"merged_configs"`
	RemovedConfigs   []string          `json:"removed_configs,omitempty"`
	BackedUpConfigs  map[string]string `json:"backed_up_configs,omitempty"`
	ChangesApplied   int               `json:"changes_applied"`
	MergedLinters    []string          `json:"merged_linters,omitempty"`
	MergedFormatters []string          `json:"merged_formatters,omitempty"`
	Success          bool              `json:"success"`
	Error            error             `json:"-"`
}

// IsSuccess returns true if the merge was successful.
func (m *MergeResult) IsSuccess() bool {
	return m.Error == nil && m.Success
}

// MergeConfigs loads and merges multiple config files.
// The first config in the list has the highest priority (per golangci-lint search order).
// Returns the merged config and a result describing what was merged.
func (cm *Merger) MergeConfigs(configPaths []string) (*Config, *MergeResult, error) {
	if len(configPaths) == 0 {
		return nil, nil, ErrNoConfigFiles
	}

	if len(configPaths) == 1 {
		loader := NewLoaderWithFS(cm.logger, cm.fs)

		config, err := loader.LoadConfig(configPaths[0])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load primary config: %w", err)
		}

		return config, &MergeResult{
			PrimaryConfig:    configPaths[0],
			MergedConfigs:    []string{},
			RemovedConfigs:   []string{},
			BackedUpConfigs:  nil,
			ChangesApplied:   0,
			MergedLinters:    []string{},
			MergedFormatters: []string{},
			Success:          true,
			Error:            nil,
		}, nil
	}

	// Sort configs by priority (golangci-lint search order)
	sortedPaths := sortByPriority(configPaths)
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
		PrimaryConfig:    primaryPath,
		MergedConfigs:    secondaryPaths,
		RemovedConfigs:   []string{},
		BackedUpConfigs:  nil,
		ChangesApplied:   0,
		MergedLinters:    []string{},
		MergedFormatters: []string{},
		Success:          true,
		Error:            nil,
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

// Config priority constants for sortByPriority.
const (
	// ConfigPriorityYML is the priority for .golangci.yml files.
	ConfigPriorityYML = iota
	// ConfigPriorityYAML is the priority for .golangci.yaml files.
	ConfigPriorityYAML
	// ConfigPriorityTOML is the priority for .golangci.toml files.
	ConfigPriorityTOML
	// ConfigPriorityJSON is the priority for .golangci.json files.
	ConfigPriorityJSON
)

// mergeConfigInto merges secondary config into primary.
// Primary values take precedence; secondary fills in gaps.
// Returns the number of changes applied.
func (cm *Merger) mergeConfigInto(primary, secondary *Config) int {
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

// SaveMergedConfig saves the merged config and optionally removes secondary configs.
// Creates backups of all modified configs before making changes.
func (cm *Merger) SaveMergedConfig(config *Config, result *MergeResult, removeSecondary bool) error {
	loader := NewLoaderWithFS(cm.logger, cm.fs)

	// Initialize backup tracking
	result.BackedUpConfigs = make(map[string]string)

	// Create backups of all configs before modifying
	allConfigs := append([]string{result.PrimaryConfig}, result.MergedConfigs...)
	for _, path := range allConfigs {
		backupPath, err := createBackup(cm.fs, path)
		if err != nil {
			cm.logger.Warnf("Failed to create backup for %s: %v", path, err)

			continue
		}

		result.BackedUpConfigs[path] = backupPath

		cm.logger.Debugf("Created backup: %s -> %s", path, backupPath)
	}

	// Save merged config to primary file
	err := loader.SaveConfig(config, result.PrimaryConfig)
	if err != nil {
		return fmt.Errorf("failed to save merged config to %s: %w", result.PrimaryConfig, err)
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
