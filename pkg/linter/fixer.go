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
	result := f.FixConfigResult(ctx, configPath, priority, dryRun)

	return result.Get()
}

// FixConfigResult fixes the config and returns a Result type for railway-oriented programming.
func (f *Fixer) FixConfigResult(
	ctx context.Context,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) types.MigrationResultType {
	f.logger.Infof("Loading configuration: %s", configPath)

	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return types.ErrMigration(analysisError("load config", priority, dryRun, configPath, err))
	}

	originalEnabled := f.configLoader.GetLintersEnabled(cfg)

	hasInvalid, err := f.runPreFlightChecks(cfg, configPath, priority, dryRun)
	if err != nil {
		return types.ErrMigration(err)
	}

	if dryRun {
		if result, shouldReturn := f.checkDryRunEarlyReturns(
			cfg, hasInvalid, hasDeprecatedLinters(originalEnabled),
		); shouldReturn {
			return result
		}
	}

	return f.analyzeAndFix(ctx, cfg, configPath, priority, dryRun, originalEnabled)
}

func (f *Fixer) analyzeAndFix(
	ctx context.Context,
	cfg *types.Config,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	originalEnabled []string,
) types.MigrationResultType {
	f.logger.Infof("Analyzing configuration...")

	analysis, err := f.analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return types.ErrMigration(analysisError("analyze config", priority, dryRun, configPath, err))
	}

	return f.applyLintersFix(ctx, cfg, analysis, configPath, priority, dryRun, originalEnabled)
}

// checkDryRunEarlyReturns checks if we should early-return in dry-run mode.
func (f *Fixer) checkDryRunEarlyReturns(
	cfg *types.Config,
	hasInvalidDurations bool,
	deprecatedPresent bool,
) (types.MigrationResultType, bool) {
	if hasInvalidDurations {
		return f.calculateDryRunResultWithInvalidDurations(cfg), true
	}

	if deprecatedPresent {
		f.logger.Infof("Dry-run with deprecated linters - skipping analysis (run without --dry-run to fix)")

		return f.calculateDryRunResultWithDeprecated(cfg), true
	}

	return types.OkMigration(nil), false
}

// analysisError is re-exported from fixer_results.go for backward compatibility.

// runPreFlightChecks runs all pre-flight fixes and returns whether invalid durations were found.
func (f *Fixer) runPreFlightChecks(
	cfg *types.Config,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) (bool, error) {
	hasInvalid, err := f.preFixInvalidDurations(cfg, configPath, dryRun)
	if err != nil {
		return hasInvalid, analysisError("pre-fix invalid durations", priority, dryRun, configPath, err)
	}

	if err := f.preFixVersion(cfg, configPath, dryRun); err != nil {
		return hasInvalid, analysisError("pre-fix version field", priority, dryRun, configPath, err)
	}

	if err := f.preFixDeprecatedLinters(cfg, configPath, dryRun); err != nil {
		return hasInvalid, analysisError("pre-fix deprecated linters", priority, dryRun, configPath, err)
	}

	if _, err := f.preFixTypecheck(cfg, configPath, dryRun); err != nil {
		return hasInvalid, analysisError("pre-fix typecheck", priority, dryRun, configPath, err)
	}

	return hasInvalid, nil
}

// fixCounts tracks the number of fixes applied by category.
type fixCounts struct {
	deprecation int
	enable      int
	formatter   int
	redundant   int
}

func (c fixCounts) total() int {
	return c.deprecation + c.enable + c.formatter + c.redundant
}

func newFixCounts() fixCounts {
	return fixCounts{
		deprecation: 0,
		enable:      0,
		formatter:   0,
		redundant:   0,
	}
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
) types.MigrationResultType {
	linterSet := types.NewSet(cfg.Linters.Enable...)
	formatterSet := types.NewSet(cfg.Formatters.Enable...)
	counts := f.applyAllFixes(linterSet, formatterSet, cfg, analysis, configPath, priority, dryRun, originalEnabled)

	if dryRun {
		return f.dryRunResult(counts)
	}

	if counts.total() == 0 {
		return noFixesResult()
	}

	return f.applyAndSave(ctx, cfg, linterSet, formatterSet, configPath, priority, dryRun, counts)
}

func (f *Fixer) applyAllFixes(
	linterSet, formatterSet types.Set[string],
	cfg *types.Config,
	analysis *types.ConfigAnalysis,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	originalEnabled []string,
) fixCounts {
	counts := newFixCounts()
	handler := newDeprecatedLinterHandler(f.logger)
	linterSet = handler.replaceLinters(linterSet, originalEnabled, dryRun, &counts)
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

// dryRunResult is re-exported from fixer_results.go.
func (f *Fixer) dryRunResult(counts fixCounts) types.MigrationResultType {
	return dryRunResult(counts)
}



func (f *Fixer) applyAndSave(
	ctx context.Context,
	cfg *types.Config,
	linterSet, formatterSet types.Set[string],
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	counts fixCounts,
) types.MigrationResultType {
	f.logger.Infof("Applying %d fixes...", counts.total())

	updater := newConfigUpdater(f.logger)
	updater.updateGoVersion(ctx, cfg)
	updater.updateRunnerSettings(cfg)
	updater.updateBuildTags(cfg)
	updateConfigFromSets(cfg, linterSet, formatterSet, f.formatterManager)

	f.logger.Infof("Saving configuration...")

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return types.ErrMigration(analysisError("save config", priority, dryRun, configPath, err))
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


