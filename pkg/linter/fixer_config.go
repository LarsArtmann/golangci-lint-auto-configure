package linter

import (
	"context"
	"os"
	"path/filepath"
	"slices"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/gogenfilter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// GoVersionProvider returns the locally installed Go version.
type GoVersionProvider func(ctx context.Context) string

// configUpdater handles updating config fields.
type configUpdater struct {
	logger          *log.Logger
	goVersionProvider GoVersionProvider
}

// newConfigUpdater creates a new config updater.
func newConfigUpdater(logger *log.Logger, goVersionProvider GoVersionProvider) *configUpdater {
	return &configUpdater{logger: logger, goVersionProvider: goVersionProvider}
}

// updateGoVersion sets the Go version in the config to the local version.
func (cu *configUpdater) updateGoVersion(ctx context.Context, cfg *types.Config) int {
	goVersion := cu.goVersionProvider(ctx)
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

// updateRunnerSettings enables parallel and serial runners if not already enabled,
// and ensures issues-exit-code is non-zero so lint failures break CI.
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

	if cfg.Run.IssuesExitCode == 0 {
		cu.logger.Infof("Setting run.issues-exit-code to 1 (was 0 — lint failures would not break CI)")

		cfg.Run.IssuesExitCode = 1
		added++
	}

	return added
}

// updateOutputFormats ensures output.formats is initialized to an empty map,
// matching the house standard. Without this, the YAML serialization omits the
// formats key entirely, which is valid but inconsistent across the ecosystem.
func (cu *configUpdater) updateOutputFormats(cfg *types.Config) int {
	if cfg.Output.Formats == nil {
		cu.logger.Infof("Setting output.formats to empty map")

		cfg.Output.Formats = map[string]any{}

		return 1
	}

	return 0
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
		cu.logger.Infof("Setting issues.max-issues-per-linter to %d", constants.DefaultMaxIssuesPerLinter)
		cfg.Issues.MaxIssuesPerLinter = constants.DefaultMaxIssuesPerLinter
		added++
	}

	if cfg.Issues.MaxSameIssues == 0 {
		cu.logger.Infof("Setting issues.max-same-issues to %d", constants.DefaultMaxSameIssues)
		cfg.Issues.MaxSameIssues = constants.DefaultMaxSameIssues
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
	linterSet types.Set[types.LinterName],
	formatterSet types.Set[types.FormatterName],
	formatterManager *FormatterManager,
	logger *log.Logger,
) int {
	enabledLinters := types.ToSortedSlice(linterSet)

	// Preserve the user's disable list. Previously this was rebuilt from scratch,
	// silently dropping every linter the user had committed to linters.disable and
	// making the disabledSet guard in enableRecommendedLinters useless on the next run.
	disabledSet := types.NewSet(cfg.Linters.Disable...)

	enabledLinters = slices.DeleteFunc(enabledLinters, func(linter types.LinterName) bool {
		if reason, ok := constants.DisabledLinters[linter]; ok {
			disabledSet.Add(linter)
			logger.Debugf("Moving disabled linter to disable list: %s (%s)", linter, reason)

			return true
		}

		return false
	})

	// Remove contradictions: a linter that ends up enabled must not also be disabled.
	for _, linter := range enabledLinters {
		disabledSet.Delete(linter)
	}

	disabledLintersList := types.ToSortedSlice(disabledSet)

	cfg.Linters.Enable = enabledLinters
	cfg.Linters.Disable = disabledLintersList

	settingsChanges := 0

	if formatterSet.Len() > 0 {
		cfg.Formatters.Enable = formatterManager.ToOrderedSlice(formatterSet)
	}

	settingsChanges += injectDefaultSettings(cfg, enabledLinters)
	settingsChanges += pruneDisabledLinterSettings(cfg, disabledLintersList, logger)

	if formatterSet.Len() > 0 {
		settingsChanges += injectDefaultFormatterSettings(cfg, cfg.Formatters.Enable)
	}

	return settingsChanges
}

// pruneDisabledLinterSettings removes settings blocks for linters that are in the
// disable list. These blocks are orphaned (the linter is disabled, so its settings
// have no effect) and previously caused confusion by implying the linter was active.
func pruneDisabledLinterSettings(cfg *types.Config, disabledLinters []types.LinterName, logger *log.Logger) int {
	if len(cfg.Linters.Settings) == 0 || len(disabledLinters) == 0 {
		return 0
	}

	pruned := 0

	for _, linter := range disabledLinters {
		key := string(linter)
		if _, exists := cfg.Linters.Settings[key]; exists {
			delete(cfg.Linters.Settings, key)
			logger.Debugf("Pruned orphaned settings for disabled linter: %s", linter)

			pruned++
		}
	}

	return pruned
}

// injectDefaultSettings injects safe default settings for linters that require
// configuration, but only if the config doesn't already have meaningful settings for them.
// Empty or nil values are treated as missing and will be overwritten with defaults.
// Returns the number of settings injected.
func injectDefaultSettings(cfg *types.Config, enabledLinters []types.LinterName) int {
	injected := 0

	for _, linterName := range enabledLinters {
		defaults, hasDefaults := constants.DefaultLinterSettings[linterName]
		if !hasDefaults {
			continue
		}

		types.InitLintersSettings(&cfg.Linters)

		key := string(linterName)
		if existing, exists := cfg.Linters.Settings[key]; exists && !isEmptySettingsValue(existing) {
			continue
		}

		cfg.Linters.Settings[key] = defaults.ToMap()
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
func injectDefaultFormatterSettings(cfg *types.Config, enabledFormatters []types.FormatterName) int {
	injected := 0

	for _, formatterName := range enabledFormatters {
		defaults, hasDefaults := constants.DefaultFormatterSettings[formatterName]
		if !hasDefaults {
			continue
		}

		if cfg.Formatters.Settings == nil {
			cfg.Formatters.Settings = make(map[string]any)
		}

		if existing, exists := cfg.Formatters.Settings[string(formatterName)]; exists && !isEmptySettingsValue(existing) {
			continue
		}

		cfg.Formatters.Settings[string(formatterName)] = defaults.ToMap()
		injected++
	}

	return injected
}
