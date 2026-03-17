package linter

// TODO: Consider using a transaction pattern for config changes (all or nothing)
// TODO: Add dry-run mode that shows detailed diff instead of just counts
// TODO: Extract duplicate linter detection into a separate validation step
// TODO: Add rollback mechanism for failed config saves
// TODO: Consider using immutable config copies for safer modifications

import (
	"context"
	"fmt"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// Fixer provides functionality to fix golangci-lint configurations.
type Fixer struct {
	configLoader *config.Loader
	analyzer     *Analyzer
	logger       *log.Logger
}

// NewFixer creates a new fixer.
func NewFixer(logger *log.Logger, analyzer *Analyzer) *Fixer {
	return &Fixer{
		configLoader: config.NewLoader(logger),
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
	f.logger.Infof("Loading configuration: %s", configPath)

	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to load config", configPath, err)
	}

	f.logger.Infof("Analyzing configuration...")

	analysis, err := f.analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to analyze config", configPath, err)
	}

	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
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

					messages = append(
						messages,
						fmt.Sprintf(
							"Replaced deprecated %s with %s: %s",
							linter,
							replacement.Replacement,
							replacement.Reason,
						),
					)
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
					messages = append(
						messages,
						fmt.Sprintf(
							"Removed deprecated %s (replacement %s already enabled): %s",
							linter,
							replacement.Replacement,
							replacement.Reason,
						),
					)
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
			f.logger.Infof("[DRY-RUN] Would enable formatter: golines (formats code and fixes long lines)")
		} else {
			f.logger.Infof("Enabling formatter: golines (formats code and fixes long lines)")

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
					f.logger.Infof("[DRY-RUN] Would remove redundant linter: %s (%s)", linterName, reason)
				} else {
					f.logger.Infof("Removing redundant linter: %s (%s)", linterName, reason)

					delete(linterSet, string(linterName))

					messages = append(messages, fmt.Sprintf("Removed redundant %s: %s", linterName, reason))
				}
			}
		}
	}

	for _, rec := range analysis.LinterRecommendations {
		if rec.Priority < priority {
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
				f.logger.Infof("[DRY-RUN] Would enable: %s (%s)", lintName, rec.Reason)
			} else {
				f.logger.Infof("Enabling: %s (%s)", lintName, rec.Reason)

				linterSet[lintName] = true
			}
		}
	}

	totalFixes := deprecationFixes + enableFixes + formatterFixes + redundantFixes

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would apply %d fixes (%d linters, %d formatters, %d deprecated, %d redundant)",
			totalFixes, enableFixes, formatterFixes, deprecationFixes, redundantFixes)

		return &types.MigrationResult{
			Success:      true,
			FixesApplied: totalFixes,
			Message:      fmt.Sprintf("Would apply %d fixes (dry-run mode)", totalFixes),
		}, nil
	}

	if totalFixes == 0 {
		return &types.MigrationResult{
			Success:      true,
			FixesApplied: 0,
			Message:      "No fixes to apply",
		}, nil
	}

	// Ensure we're in a git repo (git provides version control, no backup needed)
	if err := f.configLoader.EnsureGitRepo("."); err != nil {
		return nil, err
	}

	// Convert final linter set to sorted slice for consistent output
	enabledLinters = make([]string, 0, len(linterSet))
	for linter := range linterSet {
		enabledLinters = append(enabledLinters, linter)
	}

	cfg.Linters.Enable = enabledLinters
	cfg.Linters.Disable = []string{}

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
		return nil, errors.NewAnalysisError("failed to save config", configPath, err)
	}

	result := &types.MigrationResult{
		Success:      true,
		FixesApplied: totalFixes,
		Message: fmt.Sprintf(
			"Successfully applied %d fixes (%d linters, %d formatters, %d deprecated, %d redundant)",
			totalFixes,
			enableFixes,
			formatterFixes,
			deprecationFixes,
			redundantFixes,
		),
	}

	return result, nil
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
