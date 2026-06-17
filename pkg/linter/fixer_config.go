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
func (cu *configUpdater) updateGoVersion(ctx context.Context, cfg *types.Config) int {
	goVersion := config.GetLocalGoVersion(ctx)
	if goVersion == "" {
		return 0
	}

	if cfg.Run.Go != goVersion {
		cu.logger.Infof("Setting run.go to local version: %q -> %q", cfg.Run.Go, goVersion)
		cfg.Run.Go = goVersion

		return 1
	}

	return 0
}

// updateRunnerSettings enables parallel and serial runners if not already enabled.
func (cu *configUpdater) updateRunnerSettings(cfg *types.Config) int {
	added := 0

	if !cfg.Run.AllowParallelRunners {
		cu.logger.Infof("Enabling allow-parallel-runners: %v -> true", cfg.Run.AllowParallelRunners)
		cfg.Run.AllowParallelRunners = true
		added++
	}

	if !cfg.Run.AllowSerialRunners {
		cu.logger.Infof("Enabling allow-serial-runners: %v -> true", cfg.Run.AllowSerialRunners)
		cfg.Run.AllowSerialRunners = true
		added++
	}

	return added
}

// updateBuildTags adds Go experiment tags to the config.
func (cu *configUpdater) updateBuildTags(cfg *types.Config) int {
	existingTags := types.NewSet(cfg.Run.BuildTags...)

	added := 0

	for _, tag := range constants.GoExperimentTags() {
		if !existingTags.Contains(tag) {
			cu.logger.Infof("Adding build tag: %s", tag)
			cfg.Run.BuildTags = append(cfg.Run.BuildTags, tag)
			existingTags.Add(tag)

			added++
		}
	}

	cfg.Run.BuildTags = sortAndDeduplicate(cfg.Run.BuildTags)

	return added
}

// sortAndDeduplicate sorts and deduplicates a slice of strings.
func sortAndDeduplicate(tags []string) []string {
	if len(tags) <= 1 {
		return tags
	}

	slices.Sort(tags)

	return slices.Compact(tags)
}

// updateGeneratedExclusions injects default exclusion paths (templ, vendor) and scans
// the project for auto-generated Go files to add additional exclusion patterns.
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

	totalAdded := 0

	linterDefaults := mergeExclusionPaths(
		&cfg.Linters.Exclusions.Paths, constants.DefaultLinterExclusionPaths, cu.logger, "linters",
	)
	formatterDefaults := mergeExclusionPaths(
		&cfg.Formatters.Exclusions.Paths, constants.DefaultFormatterExclusionPaths, cu.logger, "formatters",
	)
	totalAdded += linterDefaults + formatterDefaults

	result, err := gogenfilter.ScanProject(os.DirFS(projectDir), projectDir)
	if err != nil {
		cu.logger.Debugf("Generated file scan skipped: %v", err)

		return totalAdded
	}

	if len(result.Exclusions) > 0 {
		cu.logger.Infof(
			"Found %d generated file types (%d files scanned, %d generated): %v",
			len(result.Generators), result.ScannedFiles, result.GeneratedFiles, result.Generators,
		)

		newPaths := gogenfilter.ExclusionPaths(result.Exclusions)

		totalAdded += mergeExclusionPaths(&cfg.Linters.Exclusions.Paths, newPaths, cu.logger, "linters")
		totalAdded += mergeExclusionPaths(&cfg.Formatters.Exclusions.Paths, newPaths, cu.logger, "formatters")
	}

	return totalAdded
}

// updateExclusionRules injects default exclusion rules for test files.
// These suppress linters that are noisy or inappropriate in test code.
func (cu *configUpdater) updateExclusionRules(cfg *types.Config) int {
	existingKeys := types.NewSetWithFunc(len(cfg.Linters.Exclusions.Rules), func(i int) string {
		return cfg.Linters.Exclusions.Rules[i].RuleKey()
	})

	added := 0

	for _, rule := range constants.DefaultExclusionRules {
		if !existingKeys.Contains(rule.RuleKey()) {
			cfg.Linters.Exclusions.Rules = append(cfg.Linters.Exclusions.Rules, rule)
			added++
		}
	}

	if added > 0 {
		cu.logger.Infof("Added %d default exclusion rules for test files", added)
	}

	return added
}

// updateIssuesSettings injects default issue limits when they are missing.
// Without these, golangci-lint defaults to max-same-issues:3, which hides
// duplicate problems in CI output.
func (cu *configUpdater) updateIssuesSettings(cfg *types.Config) int {
	added := 0

	if cfg.Issues.MaxIssuesPerLinter == 0 {
		cu.logger.Infof("Setting issues.max-issues-per-linter to %d", config.DefaultMaxIssuesPerLinter)
		cfg.Issues.MaxIssuesPerLinter = config.DefaultMaxIssuesPerLinter
		added++
	}

	if cfg.Issues.MaxSameIssues == 0 {
		cu.logger.Infof("Setting issues.max-same-issues to %d", config.DefaultMaxSameIssues)
		cfg.Issues.MaxSameIssues = config.DefaultMaxSameIssues
		added++
	}

	return added
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
// Returns the count of default settings injected.
func updateConfigFromSets(
	cfg *types.Config,
	linterSet types.Set[string],
	formatterSet types.Set[string],
	formatterManager *FormatterManager,
) int {
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

	settingsInjected := 0

	if formatterSet.Len() > 0 {
		cfg.Formatters.Enable = formatterManager.ToOrderedSlice(formatterSet)
	}

	settingsInjected += injectDefaultSettings(cfg, enabledLinters)

	if formatterSet.Len() > 0 {
		settingsInjected += injectDefaultFormatterSettings(cfg, cfg.Formatters.Enable)
	}

	return settingsInjected
}

// injectDefaultSettings injects safe default settings for linters that require
// configuration, but only if the config doesn't already have meaningful settings for them.
// Empty or nil values are treated as missing and will be overwritten with defaults.
// Returns the number of settings injected.
func injectDefaultSettings(cfg *types.Config, enabledLinters []string) int {
	injected := 0

	for _, linterName := range enabledLinters {
		defaults, hasDefaults := constants.DefaultLinterSettings[types.LinterName(linterName)]
		if !hasDefaults {
			continue
		}

		types.InitLintersSettings(&cfg.Linters)

		if existing, exists := cfg.Linters.Settings[linterName]; exists && !isEmptySettingsValue(existing) {
			continue
		}

		cfg.Linters.Settings[linterName] = defaults
		injected++
	}

	return injected
}

// isEmptySettingsValue reports whether a settings value is nil or an empty map,
// meaning it carries no actual configuration and should be treated as missing.
func isEmptySettingsValue(v any) bool {
	if v == nil {
		return true
	}

	m, ok := v.(map[string]any)

	return ok && len(m) == 0
}

// injectDefaultFormatterSettings injects safe default settings for formatters that require
// configuration, but only if the config doesn't already have meaningful settings for them.
// Empty or nil values are treated as missing and will be overwritten with defaults.
// Returns the number of settings injected.
func injectDefaultFormatterSettings(cfg *types.Config, enabledFormatters []string) int {
	injected := 0

	for _, formatterName := range enabledFormatters {
		defaults, hasDefaults := constants.DefaultFormatterSettings[types.FormatterName(formatterName)]
		if !hasDefaults {
			continue
		}

		if cfg.Formatters.Settings == nil {
			cfg.Formatters.Settings = make(map[string]any)
		}

		if existing, exists := cfg.Formatters.Settings[formatterName]; exists && !isEmptySettingsValue(existing) {
			continue
		}

		cfg.Formatters.Settings[formatterName] = defaults
		injected++
	}

	return injected
}
