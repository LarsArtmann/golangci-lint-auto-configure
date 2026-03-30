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

	// Check for deprecated linters before analysis
	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	hasDeprecatedLinters := false

	for _, linter := range enabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			hasDeprecatedLinters = true

			break
		}
	}

	// Pre-fix invalid duration fields before analysis to prevent golangci-lint from failing
	hasInvalidDurations, err := f.preFixInvalidDurations(cfg, configPath, dryRun)
	if err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to pre-fix invalid durations (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	// In dry-run mode with invalid durations, skip analysis (config is broken, can't run golangci-lint linters)
	if dryRun && hasInvalidDurations {
		f.logger.Infof("Dry-run with invalid durations - skipping analysis (run without --dry-run to fix)")

		return f.calculateDryRunResultWithInvalidDurations(cfg)
	}

	// Pre-fix version field before analysis to prevent golangci-lint linters command from failing
	if err := f.preFixVersion(cfg, configPath, dryRun); err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to pre-fix version field (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	// Pre-fix deprecated linters before analysis to prevent golangci-lint linters command from failing
	if err := f.preFixDeprecatedLinters(cfg, configPath, dryRun); err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to pre-fix deprecated linters (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	// Pre-fix typecheck before analysis to prevent golangci-lint linters command from failing
	// typecheck is not a configurable linter in v2, it cannot be enabled or disabled
	if _, err := f.preFixTypecheck(cfg, configPath, dryRun); err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to pre-fix typecheck (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	// In dry-run mode with deprecated linters, skip analysis (config is broken, can't run golangci-lint linters)
	if dryRun && hasDeprecatedLinters {
		f.logger.Infof("Dry-run with deprecated linters - skipping analysis (run without --dry-run to fix)")

		return f.calculateDryRunResultWithDeprecated(cfg)
	}

	f.logger.Infof("Analyzing configuration...")

	analysis, err := f.analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to analyze config (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	disabledLinters := f.configLoader.GetLintersDisabled(cfg)

	// Handle deprecated linters first
	deprecationFixes := 0
	enableFixes := 0
	formatterFixes := 0
	redundantFixes := 0

	// Track all linters to ensure uniqueness in the final list
	linterSet := make(map[string]bool)

	// Build set from current enabled linters (after pre-fixes)
	for _, linter := range cfg.Linters.Enable {
		linterSet[linter] = true
	}

	// Check for and replace deprecated linters
	for _, linter := range enabledLinters {
		if replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			// Always count this as a fix since we're removing the deprecated linter
			deprecationFixes++

			// Remove the deprecated linter from the set
			delete(linterSet, linter)

			// Add the replacement if not already present
			if !linterSet[string(replacement.Replacement)] {
				if dryRun {
					f.logger.Infof(
						"[DRY-RUN] Would replace deprecated linter: %s -> %s (%s)",
						linter,
						replacement.Replacement,
						replacement.Reason,
					)
				} else {
					f.logger.Infof(
						"Replacing deprecated linter: %s -> %s (%s)",
						linter,
						replacement.Replacement,
						replacement.Reason,
					)

					linterSet[string(replacement.Replacement)] = true
				}
			} else {
				if dryRun {
					f.logger.Infof(
						"[DRY-RUN] Would remove deprecated %s (keeping existing %s)",
						linter,
						replacement.Replacement,
					)
				} else {
					f.logger.Debugf("Removing deprecated %s (keeping existing %s)", linter, replacement.Replacement)
				}
			}
		}
	}

	// Track formatters to enable (based on high priority recommendations)
	formatterSet := make(map[string]bool)
	for _, formatter := range cfg.Formatters.Enable {
		formatterSet[formatter] = true
	}

	// Check if golines formatter should be enabled (high priority)
	shouldEnableGolines := false

	for _, rec := range analysis.FormatterRecommendations {
		if rec.Name == "golines" && rec.Priority == types.FormatterPriorityHigh {
			shouldEnableGolines = true

			break
		}
	}

	// Also enable golines if it's not already enabled and user wants high priority formatters
	if shouldEnableGolines && !formatterSet["golines"] {
		formatterFixes++

		if dryRun {
			f.logger.Debugf("[DRY-RUN] Would enable formatter: golines (formats code and fixes long lines)")
		} else {
			f.logger.Debugf("Enabling formatter: golines (formats code and fixes long lines)")

			formatterSet["golines"] = true
		}
	}

	// Check for redundant linters when formatters are enabled
	for linterName, mapping := range constants.RedundantLinters {
		if linterSet[string(linterName)] {
			// Check if the corresponding formatter is being enabled (or would be enabled in dry-run)
			formatterWillBeEnabled := formatterSet[string(mapping.Formatter)] ||
				(dryRun && string(mapping.Formatter) == "golines" && shouldEnableGolines)

			if formatterWillBeEnabled {
				redundantFixes++

				if dryRun {
					f.logger.Debugf("[DRY-RUN] Would remove redundant linter: %s (%s)", linterName, mapping.Reason)
				} else {
					f.logger.Debugf("Removing redundant linter: %s (%s)", linterName, mapping.Reason)

					delete(linterSet, string(linterName))
				}
			}
		}
	}

	for _, rec := range analysis.LinterRecommendations {
		if rec.Priority > priority {
			continue
		}

		lintName := rec.Name.String()

		// Check if this linter is deprecated and replace it with its successor
		if replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(lintName)]; isDeprecated {
			lintName = string(replacement.Replacement) // Use the replacement name instead
			// Continue to the checks below - the replacement might already be enabled
		}

		isEnabled := linterSet[lintName]
		isDisabled := contains(disabledLinters, lintName)

		if !isEnabled && !isDisabled {
			enableFixes++ // Count the fix regardless of dry-run mode

			if dryRun {
				f.logger.Debugf("[DRY-RUN] Would enable: %s (%s)", lintName, rec.Reason)
			} else {
				f.logger.Debugf("Enabling: %s (%s)", lintName, rec.Reason)

				linterSet[lintName] = true
			}
		}
	}

	totalFixes := deprecationFixes + enableFixes + formatterFixes + redundantFixes

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would apply %d fixes", totalFixes)

		return types.OkMigration(&types.MigrationResult{
			FixesApplied: totalFixes,
			Message:      fmt.Sprintf("Would apply %d fixes (dry-run mode)", totalFixes),
			NextSteps: []string{
				"Run without --dry-run to apply these fixes",
				"Then run 'golangci-lint run --fix' to auto-fix code issues",
			},
		})
	}

	if totalFixes == 0 {
		return types.OkMigration(&types.MigrationResult{
			FixesApplied: 0,
			Message:      "No fixes to apply",
			NextSteps: []string{
				"Your configuration is already up to date",
				"Run 'golangci-lint run' to check for code issues",
			},
		})
	}

	f.logger.Infof("Applying %d fixes...", totalFixes)

	// Convert final linter set to sorted slice for consistent output
	enabledLinters = make([]string, 0, len(linterSet))
	for linter := range linterSet {
		enabledLinters = append(enabledLinters, linter)
	}

	// Remove explicitly disabled linters from the enable list
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

	// Convert formatter set to slice and update config
	if len(formatterSet) > 0 {
		enabledFormatters := make([]string, 0, len(formatterSet))

		for formatter := range formatterSet {
			enabledFormatters = append(enabledFormatters, formatter)
		}

		cfg.Formatters.Enable = enabledFormatters
	}

	// Auto-detect and set local Go version
	if goVersion := config.GetLocalGoVersion(); goVersion != "" { //nolint:contextcheck // Creates its own context
		if cfg.Run.Go != goVersion {
			if dryRun {
				f.logger.Infof("[DRY-RUN] Would set run.go: %q -> %q", cfg.Run.Go, goVersion)
			} else {
				f.logger.Infof("Setting run.go to local version: %q -> %q", cfg.Run.Go, goVersion)
				cfg.Run.Go = goVersion
			}
		}
	}

	f.logger.Infof("Saving configuration...")

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError(
			fmt.Sprintf("failed to save config (priority=%d, dryRun=%t)", priority, dryRun),
			configPath, err,
		))
	}

	result := &types.MigrationResult{
		FixesApplied: totalFixes,
		Message: fmt.Sprintf(
			"Successfully applied %d fixes (%d linters, %d formatters, %d deprecated, %d redundant)",
			totalFixes,
			enableFixes,
			formatterFixes,
			deprecationFixes,
			redundantFixes,
		),
		NextSteps: []string{
			"Run 'golangci-lint run --fix' to auto-fix code issues found by the newly enabled linters",
			"Run 'golangci-lint run' to see remaining issues that require manual fixes",
		},
	}

	return types.OkMigration(result)
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
