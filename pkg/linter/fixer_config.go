package linter

import (
	"context"
	"slices"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
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
