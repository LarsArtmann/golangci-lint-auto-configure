package linter

import (
	"context"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Fixer provides functionality to fix golangci-lint configurations.
type Fixer struct {
	configLoader     types.ConfigLoader
	analyzer         types.LinterAnalyzer
	logger           *log.Logger
	formatterManager *FormatterManager
}

// NewFixer creates a new fixer.
func NewFixer(logger *log.Logger, analyzer types.LinterAnalyzer, configLoader types.ConfigLoader) *Fixer {
	return &Fixer{
		configLoader:     configLoader,
		analyzer:         analyzer,
		logger:           logger,
		formatterManager: NewFormatterManager(logger),
	}
}

// FixConfig fixes the golangci-lint configuration by enabling recommended linters.
func (f *Fixer) FixConfig(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) (*types.MigrationResult, error) {
	f.logger.Infof("Loading configuration: %s", configPath)

	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return migrationError("load config", priority, dryRun, configPath, "", err)
	}

	originalEnabled := f.configLoader.GetLintersEnabled(cfg)

	version := f.detectVersion(ctx)

	hasInvalid, err := f.runPreFlightChecks(cfg, configPath, priority, dryRun, version)
	if err != nil {
		return migrationError("pre-flight checks", priority, dryRun, configPath, version, err)
	}

	if dryRun {
		if result, shouldReturn := f.checkDryRunEarlyReturns(
			cfg, hasInvalid, hasDeprecatedLinters(originalEnabled, version),
		); shouldReturn {
			return result, nil
		}
	}

	return f.analyzeAndFix(ctx, cfg, configPath, priority, dryRun, originalEnabled, version)
}

// detectVersion tries to detect the golangci-lint version early for version-gated replacements.
// Returns empty string if detection fails (version-gating is skipped, all replacements applied).
func (f *Fixer) detectVersion(ctx context.Context) string {
	if f.analyzer == nil {
		return ""
	}

	version := f.analyzer.GetDetectedVersion()
	if version != "" {
		return version
	}

	if err := f.analyzer.FindBinary(ctx); err != nil {
		f.logger.Debugf("Version detection skipped: %v", err)

		return ""
	}

	if err := f.analyzer.CheckVersion(ctx); err != nil {
		f.logger.Debugf("Version check failed: %v", err)

		return ""
	}

	return f.analyzer.GetDetectedVersion()
}

func (f *Fixer) analyzeAndFix(
	ctx context.Context,
	cfg *types.Config,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	originalEnabled []string,
	version string,
) (*types.MigrationResult, error) {
	f.logger.Infof("Analyzing configuration...")

	analysis, err := f.analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return migrationError("analyze config", priority, dryRun, configPath, version, err)
	}

	return f.applyLintersFix(ctx, cfg, analysis, configPath, priority, dryRun, originalEnabled, version)
}

// checkDryRunEarlyReturns checks if we should early-return in dry-run mode.
func (f *Fixer) checkDryRunEarlyReturns(
	cfg *types.Config,
	hasInvalidDurations bool,
	deprecatedPresent bool,
) (*types.MigrationResult, bool) {
	if hasInvalidDurations {
		result := f.calculateDryRunResultWithInvalidDurations(cfg)

		return result, true
	}

	if deprecatedPresent {
		f.logger.Infof("Dry-run with deprecated linters - skipping analysis (run without --dry-run to fix)")

		result := f.calculateDryRunResultWithDeprecated(cfg)

		return result, true
	}

	return nil, false
}

// runPreFlightChecks runs all pre-flight fixes and returns whether invalid durations were found.
func (f *Fixer) runPreFlightChecks(
	cfg *types.Config,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	version string,
) (bool, error) {
	hasInvalid, err := f.preFixInvalidDurations(cfg, configPath, dryRun)
	if err != nil {
		return hasInvalid, analysisError("pre-fix invalid durations", priority, dryRun, configPath, version, err)
	}

	if err := f.preFixVersion(cfg, configPath, dryRun); err != nil {
		return hasInvalid, analysisError("pre-fix version field", priority, dryRun, configPath, version, err)
	}

	if err := f.preFixDeprecatedLinters(cfg, configPath, dryRun, version); err != nil {
		return hasInvalid, analysisError("pre-fix deprecated linters", priority, dryRun, configPath, version, err)
	}

	if _, err := f.preFixTypecheck(cfg, configPath, dryRun); err != nil {
		return hasInvalid, analysisError("pre-fix typecheck", priority, dryRun, configPath, version, err)
	}

	return hasInvalid, nil
}

// fixCounts tracks the number of fixes applied by category.
type fixCounts struct {
	deprecation   int
	enable        int
	formatter     int
	generated     int
	redundant     int
	normalization int // config-level normalizations: go version, runners, build tags, settings, issues
}

func (c fixCounts) total() int {
	return c.deprecation + c.enable + c.formatter + c.generated + c.redundant + c.normalization
}

// configChangeRecorder ensures every config mutation in applyAndSave is counted.
// Mutations are passed as closures so the recorder controls both execution and counting,
// preventing the footgun where a mutation runs but its count is not incremented
// (which would cause the total()==0 guard to silently discard all changes).
type configChangeRecorder struct {
	counts fixCounts
}

func (r *configChangeRecorder) normalize(fn func() int) {
	r.counts.normalization += fn()
}

func (r *configChangeRecorder) generated(fn func() int) {
	r.counts.generated += fn()
}

// applyLintersFix processes linter recommendations, applies fixes, and saves the config.
func (f *Fixer) applyLintersFix(
	ctx context.Context,
	cfg *types.Config,
	analysis *types.ConfigAnalysis,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	originalEnabled []string,
	version string,
) (*types.MigrationResult, error) {
	linterSet := types.NewSet(cfg.Linters.Enable...)
	formatterSet := types.NewSet(cfg.Formatters.Enable...)
	counts := f.applyAllFixes(
		linterSet,
		formatterSet,
		cfg,
		analysis,
		configPath,
		priority,
		dryRun,
		originalEnabled,
		version,
	)

	return f.applyAndSave(ctx, cfg, linterSet, formatterSet, configPath, priority, dryRun, version, counts)
}

func (f *Fixer) applyAllFixes(
	linterSet, formatterSet types.Set[string],
	cfg *types.Config,
	analysis *types.ConfigAnalysis,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	originalEnabled []string,
	version string,
) fixCounts {
	var counts fixCounts

	handler := newDeprecatedLinterHandler(f.logger, version)
	linterSet = handler.replaceLinters(linterSet, originalEnabled, dryRun, &counts, cfg)
	counts.formatter += f.formatterManager.EnableCoreFormatters(formatterSet, dryRun)
	counts.formatter += f.formatterManager.EnableGolinesFormatter(formatterSet, analysis, dryRun)
	counts.formatter += f.formatterManager.EnableSwaggoFormatter(formatterSet, configPath, dryRun)
	counts.redundant += f.formatterManager.RemoveRedundantLinters(linterSet, formatterSet, dryRun)
	counts.redundant += f.formatterManager.RemoveRedundantGofmt(formatterSet, dryRun)
	counts.enable = f.enableRecommendedLinters(
		linterSet, f.configLoader.GetLintersDisabled(cfg), analysis, priority, dryRun,
	)

	return counts
}

func (f *Fixer) applyAndSave(
	ctx context.Context,
	cfg *types.Config,
	linterSet, formatterSet types.Set[string],
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	version string,
	counts fixCounts,
) (*types.MigrationResult, error) {
	updater := newConfigUpdater(f.logger)

	rec := configChangeRecorder{counts: counts}
	rec.normalize(func() int { return updater.updateGoVersion(ctx, cfg) })
	rec.normalize(func() int { return updater.updateRunnerSettings(cfg) })
	rec.normalize(func() int { return updater.updateBuildTags(cfg) })
	rec.normalize(func() int { return updater.updateOutputFormats(cfg) })
	rec.normalize(func() int {
		return updateConfigFromSets(cfg, linterSet, formatterSet, f.formatterManager, f.logger)
	})

	rec.generated(func() int { return updater.updateGeneratedExclusions(cfg, configPath) })
	rec.generated(func() int { return updater.updateExclusionRules(cfg) })

	rec.normalize(func() int { return updater.updateIssuesSettings(cfg) })

	counts = rec.counts

	if counts.total() == 0 {
		return noFixesResult()
	}

	if dryRun {
		return dryRunResult(counts)
	}

	f.logger.Infof("Applying %d fixes...", counts.total())

	f.logger.Infof("Saving configuration...")

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return migrationError("save config", priority, dryRun, configPath, version, err)
	}

	return successResult(counts)
}

// enableRecommendedLinters enables recommended linters that aren't already enabled or explicitly disabled.
func (f *Fixer) enableRecommendedLinters(
	linterSet types.Set[string],
	disabledLinters []string,
	analysis *types.ConfigAnalysis,
	priority types.LinterPriority,
	dryRun bool,
) int {
	count := 0

	disabledSet := types.NewSet(disabledLinters...)

	for _, rec := range analysis.LinterRecommendations {
		if rec.Priority > priority {
			continue
		}

		lintName := resolveLinterName(rec.Name)

		if linterSet.Contains(lintName) || disabledSet.Contains(lintName) {
			continue
		}

		count++

		if dryRun {
			f.logger.Debugf("[DRY-RUN] Would enable: %s (%s)", lintName, rec.Reason)
		} else {
			f.logger.Debugf("Enabling: %s (%s)", lintName, rec.Reason)
			linterSet.Add(lintName)
		}
	}

	return count
}
