package linter

import (
	"context"
	"fmt"
	"slices"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Fixer provides functionality to fix golangci-lint configurations.
type Fixer struct {
	configLoader types.ConfigLoader
	analyzer     types.LinterAnalyzer
	logger       *log.Logger
}

// NewFixer creates a new fixer.
func NewFixer(logger *log.Logger, analyzer types.LinterAnalyzer, configLoader types.ConfigLoader) *Fixer {
	return &Fixer{
		configLoader: configLoader,
		analyzer:     analyzer,
		logger:       logger,
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
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to load config (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	// Capture original enabled linters before pre-flight modifications
	originalEnabled := f.configLoader.GetLintersEnabled(cfg)

	hasInvalid, err := f.runPreFlightChecks(cfg, configPath, priority, dryRun)
	if err != nil {
		return types.ErrMigration(err)
	}

	if dryRun {
		if result, shouldReturn := f.checkDryRunEarlyReturns(cfg, hasInvalid, hasDeprecatedLinters(originalEnabled)); shouldReturn {
			return result
		}
	}

	f.logger.Infof("Analyzing configuration...")

	analysis, err := f.analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to analyze config (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	return f.applyLintersFix(cfg, analysis, configPath, priority, dryRun, originalEnabled)
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

// runPreFlightChecks runs all pre-flight fixes and returns whether invalid durations were found.
func (f *Fixer) runPreFlightChecks(
	cfg *types.Config,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
) (bool, error) {
	hasInvalid, err := f.preFixInvalidDurations(cfg, configPath, dryRun)
	if err != nil {
		return false, apperrors.NewAnalysisError(
			fmt.Sprintf("failed to pre-fix invalid durations (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		)
	}

	if err := f.preFixVersion(cfg, configPath, dryRun); err != nil {
		return hasInvalid, apperrors.NewAnalysisError(
			fmt.Sprintf("failed to pre-fix version field (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		)
	}

	if err := f.preFixDeprecatedLinters(cfg, configPath, dryRun); err != nil {
		return hasInvalid, apperrors.NewAnalysisError(
			fmt.Sprintf("failed to pre-fix deprecated linters (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		)
	}

	if _, err := f.preFixTypecheck(cfg, configPath, dryRun); err != nil {
		return hasInvalid, apperrors.NewAnalysisError(
			fmt.Sprintf("failed to pre-fix typecheck (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		)
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

// applyLintersFix processes linter recommendations, applies fixes, and saves the config.
func (f *Fixer) applyLintersFix(
	cfg *types.Config,
	analysis *types.ConfigAnalysis,
	configPath string,
	priority types.LinterPriority,
	dryRun bool,
	originalEnabled []string,
) types.MigrationResultType {
	disabledLinters := f.configLoader.GetLintersDisabled(cfg)
	linterSet := buildLinterSet(cfg.Linters.Enable)

	counts := fixCounts{}

	linterSet = f.replaceDeprecatedLinters(linterSet, originalEnabled, dryRun, &counts)

	formatterSet := buildLinterSet(cfg.Formatters.Enable)
	counts.formatter = f.enableGolinesFormatter(formatterSet, analysis, dryRun)
	counts.redundant = f.removeRedundantLinters(linterSet, formatterSet, dryRun)

	counts.enable = f.enableRecommendedLinters(linterSet, disabledLinters, analysis, priority, dryRun)

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would apply %d fixes", counts.total())

		return types.OkMigration(&types.MigrationResult{
			FixesApplied: counts.total(),
			Message:      fmt.Sprintf("Would apply %d fixes (dry-run mode)", counts.total()),
			NextSteps: []string{
				"Run without --dry-run to apply these fixes",
				"Then run 'golangci-lint run --fix' to auto-fix code issues",
			},
		})
	}

	if counts.total() == 0 {
		return types.OkMigration(&types.MigrationResult{
			FixesApplied: 0,
			Message:      "No fixes to apply",
			NextSteps: []string{
				"Your configuration is already up to date",
				"Run 'golangci-lint run' to check for code issues",
			},
		})
	}

	f.logger.Infof("Applying %d fixes...", counts.total())

	f.updateConfigFromSets(cfg, linterSet, formatterSet)

	if goVersion := config.GetLocalGoVersion(); goVersion != "" { //nolint:contextcheck // Creates its own context
		if cfg.Run.Go != goVersion {
			f.logger.Infof("Setting run.go to local version: %q -> %q", cfg.Run.Go, goVersion)
			cfg.Run.Go = goVersion
		}
	}

	f.logger.Infof("Saving configuration...")

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to save config (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	return types.OkMigration(&types.MigrationResult{
		FixesApplied: counts.total(),
		Message: fmt.Sprintf(
			"Successfully applied %d fixes (%d linters, %d formatters, %d deprecated, %d redundant)",
			counts.total(),
			counts.enable,
			counts.formatter,
			counts.deprecation,
			counts.redundant,
		),
		NextSteps: []string{
			"Run 'golangci-lint run --fix' to auto-fix code issues found by the newly enabled linters",
			"Run 'golangci-lint run' to see remaining issues that require manual fixes",
		},
	})
}

// replaceDeprecatedLinters replaces deprecated linters with their successors in the linter set.
func (f *Fixer) replaceDeprecatedLinters(
	linterSet map[string]bool,
	enabledLinters []string,
	dryRun bool,
	counts *fixCounts,
) map[string]bool {
	for _, linter := range enabledLinters {
		replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]
		if !isDeprecated {
			continue
		}

		counts.deprecation++
		delete(linterSet, linter)

		if linterSet[string(replacement.Replacement)] {
			if dryRun {
				f.logger.Infof("[DRY-RUN] Would remove deprecated %s (keeping existing %s)", linter, replacement.Replacement)
			} else {
				f.logger.Debugf("Removing deprecated %s (keeping existing %s)", linter, replacement.Replacement)
			}

			continue
		}

		if dryRun {
			f.logger.Infof("[DRY-RUN] Would replace deprecated linter: %s -> %s (%s)", linter, replacement.Replacement, replacement.Reason)
		} else {
			f.logger.Infof("Replacing deprecated linter: %s -> %s (%s)", linter, replacement.Replacement, replacement.Reason)
			linterSet[string(replacement.Replacement)] = true
		}
	}

	return linterSet
}

// enableGolinesFormatter enables the golines formatter if recommended at high priority.
func (f *Fixer) enableGolinesFormatter(
	formatterSet map[string]bool,
	analysis *types.ConfigAnalysis,
	dryRun bool,
) int {
	shouldEnable := false

	for _, rec := range analysis.FormatterRecommendations {
		if rec.Name == "golines" && rec.Priority == types.FormatterPriorityHigh {
			shouldEnable = true

			break
		}
	}

	if !shouldEnable || formatterSet["golines"] {
		return 0
	}

	if dryRun {
		f.logger.Debugf("[DRY-RUN] Would enable formatter: golines (formats code and fixes long lines)")
	} else {
		f.logger.Debugf("Enabling formatter: golines (formats code and fixes long lines)")
		formatterSet["golines"] = true
	}

	return 1
}

// removeRedundantLinters removes linters that are superseded by enabled formatters.
func (f *Fixer) removeRedundantLinters(
	linterSet map[string]bool,
	formatterSet map[string]bool,
	dryRun bool,
) int {
	count := 0

	for linterName, mapping := range constants.RedundantLinters {
		if !linterSet[string(linterName)] {
			continue
		}

		if !formatterSet[string(mapping.Formatter)] {
			continue
		}

		count++

		if dryRun {
			f.logger.Debugf("[DRY-RUN] Would remove redundant linter: %s (%s)", linterName, mapping.Reason)
		} else {
			f.logger.Debugf("Removing redundant linter: %s (%s)", linterName, mapping.Reason)
			delete(linterSet, string(linterName))
		}
	}

	return count
}

// enableRecommendedLinters enables recommended linters that aren't already enabled or explicitly disabled.
func (f *Fixer) enableRecommendedLinters(
	linterSet map[string]bool,
	disabledLinters []string,
	analysis *types.ConfigAnalysis,
	priority types.LinterPriority,
	dryRun bool,
) int {
	count := 0

	for _, rec := range analysis.LinterRecommendations {
		if rec.Priority > priority {
			continue
		}

		lintName := rec.Name.String()

		if replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(lintName)]; isDeprecated {
			lintName = string(replacement.Replacement)
		}

		if linterSet[lintName] || contains(disabledLinters, lintName) {
			continue
		}

		count++

		if dryRun {
			f.logger.Debugf("[DRY-RUN] Would enable: %s (%s)", lintName, rec.Reason)
		} else {
			f.logger.Debugf("Enabling: %s (%s)", lintName, rec.Reason)
			linterSet[lintName] = true
		}
	}

	return count
}

// updateConfigFromSets applies the linter and formatter sets back to the config struct.
func (f *Fixer) updateConfigFromSets(
	cfg *types.Config,
	linterSet map[string]bool,
	formatterSet map[string]bool,
) {
	enabledLinters := setToSortedSlice(linterSet)

	disabledLintersList := make([]string, 0)
	enabledLinters = slices.DeleteFunc(enabledLinters, func(linter string) bool {
		if _, isDisabled := constants.DisabledLinters[types.LinterName(linter)]; isDisabled {
			disabledLintersList = append(disabledLintersList, linter)

			return true
		}

		return false
	})

	cfg.Linters.Enable = enabledLinters
	cfg.Linters.Disable = disabledLintersList

	if len(formatterSet) > 0 {
		cfg.Formatters.Enable = setToSortedSlice(formatterSet)
	}
}

func buildLinterSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))

	for _, item := range items {
		set[item] = true
	}

	return set
}

func setToSortedSlice(set map[string]bool) []string {
	result := make([]string, 0, len(set))

	for item := range set {
		result = append(result, item)
	}

	return result
}

func hasDeprecatedLinters(enabledLinters []string) bool {
	for _, linter := range enabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			return true
		}
	}

	return false
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
