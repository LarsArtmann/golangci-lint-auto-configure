package linter

import (
	"context"
	"os"
	"path/filepath"
	"slices"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/gogenfilter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// configUpdater handles updating config fields.
type configUpdater struct {
	logger *log.Logger
}

// newConfigUpdater creates a new config updater.
func newConfigUpdater(logger *log.Logger) *configUpdater {
	return &configUpdater{logger: logger}
}

// updateGoVersion sets the Go version in the config to the local version.
func (cu *configUpdater) updateGoVersion(ctx context.Context, cfg *types.Config) {
	goVersion := config.GetLocalGoVersion(ctx)
	if goVersion == "" {
		return
	}

	if cfg.Run.Go != goVersion {
		cu.logger.Infof("Setting run.go to local version: %q -> %q", cfg.Run.Go, goVersion)
		cfg.Run.Go = goVersion
	}
}

// updateRunnerSettings enables parallel and serial runners if not already enabled.
func (cu *configUpdater) updateRunnerSettings(cfg *types.Config) {
	if !cfg.Run.AllowParallelRunners {
		cu.logger.Infof("Enabling allow-parallel-runners: %v -> true", cfg.Run.AllowParallelRunners)
		cfg.Run.AllowParallelRunners = true
	}

	if !cfg.Run.AllowSerialRunners {
		cu.logger.Infof("Enabling allow-serial-runners: %v -> true", cfg.Run.AllowSerialRunners)
		cfg.Run.AllowSerialRunners = true
	}
}

// updateBuildTags adds Go experiment tags to the config.
func (cu *configUpdater) updateBuildTags(cfg *types.Config) {
	existingTags := types.NewSet(cfg.Run.BuildTags...)

	for _, tag := range constants.GoExperimentTags() {
		if !existingTags.Contains(tag) {
			cu.logger.Infof("Adding build tag: %s", tag)
			cfg.Run.BuildTags = append(cfg.Run.BuildTags, tag)
			existingTags.Add(tag)
		}
	}

	cfg.Run.BuildTags = sortAndDeduplicate(cfg.Run.BuildTags)
}

// sortAndDeduplicate sorts and deduplicates a slice of strings.
func sortAndDeduplicate(tags []string) []string {
	if len(tags) <= 1 {
		return tags
	}

	slices.Sort(tags)

	return slices.Compact(tags)
}

// updateGeneratedExclusions scans the project for auto-generated Go files
// and injects exclusion path patterns into both linters and formatters exclusions.
// Uses gogenfilter for two-phase detection (filename first, content second).
func (cu *configUpdater) updateGeneratedExclusions(cfg *types.Config, configPath string) int {
	projectDir := filepath.Dir(configPath)

	if cfg.Linters.Exclusions.Generated == "" {
		cu.logger.Infof("Setting linters.exclusions.generated to \"lax\"")

		cfg.Linters.Exclusions.Generated = "lax"
	}

	if cfg.Formatters.Exclusions.Generated == "" {
		cu.logger.Infof("Setting formatters.exclusions.generated to \"lax\"")

		cfg.Formatters.Exclusions.Generated = "lax"
	}

	result, err := gogenfilter.ScanProject(os.DirFS(projectDir), projectDir)
	if err != nil {
		cu.logger.Debugf("Generated file scan skipped: %v", err)

		return 0
	}

	if len(result.Exclusions) == 0 {
		return 0
	}

	cu.logger.Infof(
		"Found %d generated file types (%d files scanned, %d generated): %v",
		len(result.Generators), result.ScannedFiles, result.GeneratedFiles, result.Generators,
	)

	newPaths := gogenfilter.ExclusionPaths(result.Exclusions)

	linterPathsAdded := mergeExclusionPaths(&cfg.Linters.Exclusions.Paths, newPaths, cu.logger, "linters")
	formatterPathsAdded := mergeExclusionPaths(&cfg.Formatters.Exclusions.Paths, newPaths, cu.logger, "formatters")

	return linterPathsAdded + formatterPathsAdded
}

// ApplyGeneratedExclusions scans a project config for auto-generated Go files
// and injects exclusion paths into the config. This is the public entry point
// used by both the fixer flow and the preset flow.
func ApplyGeneratedExclusions(logger *log.Logger, cfg *types.Config, configPath string) int {
	updater := newConfigUpdater(logger)

	return updater.updateGeneratedExclusions(cfg, configPath)
}

func mergeExclusionPaths(existing *[]string, newPaths []string, logger *log.Logger, section string) int {
	if len(newPaths) == 0 {
		return 0
	}

	merged := gogenfilter.MergeExclusionPaths(*existing, newPaths)
	added := len(merged) - len(*existing)

	if added > 0 {
		logger.Infof("Adding %d generated file exclusions to %s: %v", added, section, newPaths)

		*existing = merged
	}

	return added
}

// updateConfigFromSets applies the linter and formatter sets back to the config struct.
func updateConfigFromSets(
	cfg *types.Config,
	linterSet types.Set[string],
	formatterSet types.Set[string],
	formatterManager *FormatterManager,
) {
	enabledLinters := types.ToSortedSlice(linterSet)

	disabledLintersList := make([]string, 0)
	enabledLinters = slices.DeleteFunc(enabledLinters, func(linter string) bool {
		if constants.DisabledLinters.Contains(types.LinterName(linter)) {
			disabledLintersList = append(disabledLintersList, linter)

			return true
		}

		return false
	})

	cfg.Linters.Enable = enabledLinters
	cfg.Linters.Disable = disabledLintersList

	if formatterSet.Len() > 0 {
		cfg.Formatters.Enable = formatterManager.ToOrderedSlice(formatterSet)
	}

	injectDefaultSettings(cfg, enabledLinters)
}

// injectDefaultSettings injects safe default settings for linters that require
// configuration, but only if the config doesn't already have settings for them.
func injectDefaultSettings(cfg *types.Config, enabledLinters []string) {
	for _, linterName := range enabledLinters {
		defaults, hasDefaults := constants.DefaultLinterSettings[types.LinterName(linterName)]
		if !hasDefaults {
			continue
		}

		types.InitLintersSettings(&cfg.Linters)

		if _, exists := cfg.Linters.Settings[linterName]; exists {
			continue
		}

		cfg.Linters.Settings[linterName] = defaults
	}
}
