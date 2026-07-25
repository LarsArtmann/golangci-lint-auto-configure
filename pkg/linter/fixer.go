package linter

import (
	"context"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/policy"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// fixerConfigLoader is the narrowest interface the Fixer needs: load, save,
// and inspect linter state. This is a subset of types.ConfigLoader.
type fixerConfigLoader interface {
	types.ConfigReader
	types.ConfigWriter
	types.ConfigInspector
}

// Fixer provides functionality to fix golangci-lint configurations.
type Fixer struct {
	configLoader     fixerConfigLoader
	analyzer         types.LinterAnalyzer
	logger           *log.Logger
	formatterManager *FormatterManager
	ledger           audit.Recorder
	pol              *policy.Policy
}

// NewFixer creates a new fixer.
func NewFixer(logger *log.Logger, analyzer types.LinterAnalyzer, configLoader fixerConfigLoader) *Fixer {
	return &Fixer{
		configLoader:     configLoader,
		analyzer:         analyzer,
		logger:           logger,
		formatterManager: NewFormatterManager(logger),
		ledger:           audit.NoopRecorder{},
		pol:              nil,
	}
}

// SetLedger sets the audit recorder used to log every config change the fixer makes.
// If recorder is nil, the fixer falls back to a NoopRecorder (no recording).
func (f *Fixer) SetLedger(recorder audit.Recorder) {
	if recorder == nil {
		recorder = audit.NoopRecorder{}
	}

	f.ledger = recorder
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
	before := snapshotLinterState(cfg)

	f.loadPolicy(configPath)

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

	result, err := f.applyAndSave(ctx, cfg, linterSet, formatterSet, configPath, priority, dryRun, version, counts)
	if err != nil {
		return result, err
	}

	// Only record changes that were actually persisted to disk (skip dry-runs and no-op runs).
	if !dryRun && result.IsSuccess() {
		f.recordConfigChanges(before, cfg)
	}

	return result, nil
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
	var rec configChangeRecorder

	handler := newDeprecatedLinterHandler(f.logger, version)

	rec.deprecation(func() int {
		var count int

		linterSet, count = handler.replaceLinters(linterSet, originalEnabled, dryRun, cfg)

		return count
	})
	rec.formatter(func() int { return f.formatterManager.EnableCoreFormatters(formatterSet, dryRun) })
	rec.formatter(func() int { return f.formatterManager.EnableGolinesFormatter(formatterSet, analysis, dryRun) })
	rec.formatter(func() int { return f.formatterManager.EnableSwaggoFormatter(formatterSet, configPath, dryRun) })
	rec.redundant(func() int { return f.formatterManager.RemoveRedundantLinters(linterSet, formatterSet, dryRun) })
	rec.redundant(func() int { return f.formatterManager.RemoveRedundantGofmt(formatterSet, dryRun) })
	rec.enable(func() int {
		return f.enableRecommendedLinters(
			linterSet, f.configLoader.GetLintersDisabled(cfg), analysis, priority, dryRun,
		)
	})

	return rec.counts
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

	if f.pol != nil && !dryRun {
		rec.normalize(func() int { return f.enforceDisableReasons(cfg) })
	}

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
