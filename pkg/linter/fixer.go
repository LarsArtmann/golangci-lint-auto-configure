package linter

import (
	"context"
	"fmt"
	"slices"

	"charm.land/log/v2"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// Fixer provides functionality to fix golangci-lint configurations.
type Fixer struct {
	configLoader types.ConfigLoader
	analyzer     *Analyzer
	logger       *log.Logger
}

// NewFixer creates a new fixer.
func NewFixer(logger *log.Logger, analyzer *Analyzer, configLoader types.ConfigLoader) *Fixer {
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
		return types.ErrMigration(apperrors.NewAnalysisError("failed to load config", configPath, err))
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

	// Pre-fix version field before analysis to prevent golangci-lint linters command from failing
	if err := f.preFixVersion(cfg, configPath, dryRun); err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError("failed to pre-fix version field", configPath, err))
	}

	// Pre-fix deprecated linters before analysis to prevent golangci-lint linters command from failing
	if err := f.preFixDeprecatedLinters(cfg, configPath, dryRun); err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError("failed to pre-fix deprecated linters", configPath, err))
	}

	// In dry-run mode with deprecated linters, skip analysis (config is broken, can't run golangci-lint linters)
	if dryRun && hasDeprecatedLinters {
		f.logger.Infof("Dry-run with deprecated linters - skipping analysis (run without --dry-run to fix)")

		return f.calculateDryRunResultWithDeprecated(cfg)
	}

	f.logger.Infof("Analyzing configuration...")

	analysis, err := f.analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError("failed to analyze config", configPath, err))
	}

	disabledLinters := f.configLoader.GetLintersDisabled(cfg)

	// Handle deprecated linters first
	deprecationFixes := 0
	enableFixes := 0
	formatterFixes := 0
	redundantFixes := 0

	// Track all linters to ensure uniqueness in the final list
	linterSet := make(map[string]bool)

	// Build set from existing enabled linters
	for _, linter := range enabledLinters {
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
	for linterName, reason := range constants.RedundantLinters {
		if linterSet[string(linterName)] {
			// Check if the corresponding formatter is being enabled (or would be enabled in dry-run)
			golinesWillBeEnabled := formatterSet["golines"] || (shouldEnableGolines && dryRun)

			if linterName == "lll" && golinesWillBeEnabled {
				redundantFixes++

				if dryRun {
					f.logger.Debugf("[DRY-RUN] Would remove redundant linter: %s (%s)", linterName, reason)
				} else {
					f.logger.Debugf("Removing redundant linter: %s (%s)", linterName, reason)

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

	f.logger.Infof("Saving configuration...")

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return types.ErrMigration(apperrors.NewAnalysisError("failed to save config", configPath, err))
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

// preFixDeprecatedLinters replaces deprecated linters in the config before analysis.
// This is necessary because golangci-lint linters command will fail if the config
// contains deprecated/removed linters (even in the disable list).
func (f *Fixer) preFixVersion(cfg *types.Config, configPath string, dryRun bool) error {
	if cfg.Version == "2" {
		return nil // Version is already correct
	}

	if cfg.Version == "" {
		f.logger.Infof("Version field is empty, setting to \"2\" for golangci-lint v2 compatibility")
	} else {
		f.logger.Infof("Version field is \"%s\", setting to \"2\" for golangci-lint v2 compatibility", cfg.Version)
	}

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would set version to \"2\"")

		return nil
	}

	cfg.Version = "2"

	// Save the fixed config so that golangci-lint linters command will work
	err := f.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return fmt.Errorf("failed to save pre-fixed config: %w", err)
	}

	return nil
}

func (f *Fixer) preFixDeprecatedLinters(cfg *types.Config, configPath string, dryRun bool) error {
	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	disabledLinters := f.configLoader.GetLintersDisabled(cfg)

	// Track if any deprecated linters were found
	var deprecatedFound []string

	// Build set from existing enabled linters, replacing deprecated ones
	linterSet := make(map[string]bool)

	for _, linter := range enabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			deprecatedFound = append(deprecatedFound, linter)

			continue
		}

		linterSet[linter] = true
	}

	// Also check disabled linters for deprecated ones
	disabledSet := make(map[string]bool)

	for _, linter := range disabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			deprecatedFound = append(deprecatedFound, linter+" (disabled)")

			continue
		}

		disabledSet[linter] = true
	}

	if len(deprecatedFound) == 0 {
		return nil // No deprecated linters to fix
	}

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would pre-fix %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)

		return nil
	}

	f.logger.Infof("Pre-fixing %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)

	// Convert sets to sorted slices
	fixedEnabled := make([]string, 0, len(linterSet))
	for linter := range linterSet {
		fixedEnabled = append(fixedEnabled, linter)
	}

	fixedDisabled := make([]string, 0, len(disabledSet))
	for linter := range disabledSet {
		fixedDisabled = append(fixedDisabled, linter)
	}

	cfg.Linters.Enable = fixedEnabled
	cfg.Linters.Disable = fixedDisabled

	// Save the fixed config so that golangci-lint linters command will work
	err := f.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return fmt.Errorf("failed to save pre-fixed config: %w", err)
	}

	return nil
}

// calculateDryRunResultWithDeprecated calculates the dry-run result when deprecated linters are present.
// Since the config has deprecated linters, we can't run golangci-lint linters for analysis,
// so we just report what would be fixed regarding deprecated linters.
func (f *Fixer) calculateDryRunResultWithDeprecated(cfg *types.Config) types.MigrationResultType {
	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	deprecationFixes := 0

	linterSet := make(map[string]bool)

	for _, linter := range enabledLinters {
		if replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			deprecationFixes++

			delete(linterSet, linter)

			if !linterSet[string(replacement.Replacement)] {
				f.logger.Infof(
					"[DRY-RUN] Would replace deprecated linter: %s -> %s (%s)",
					linter,
					replacement.Replacement,
					replacement.Reason,
				)
				linterSet[string(replacement.Replacement)] = true
			} else {
				f.logger.Infof(
					"[DRY-RUN] Would remove deprecated %s (keeping existing %s)",
					linter,
					replacement.Replacement,
				)
			}
		} else {
			linterSet[linter] = true
		}
	}

	f.logger.Infof("[DRY-RUN] Would apply %d fixes", deprecationFixes)

	return types.OkMigration(&types.MigrationResult{
		FixesApplied: deprecationFixes,
		Message: fmt.Sprintf(
			"Would apply %d fixes (dry-run mode, skipped analysis due to deprecated linters)",
			deprecationFixes,
		),
	})
}
