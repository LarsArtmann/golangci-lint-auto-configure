package linter

import (
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
func (f *Fixer) FixConfig(configPath string, priority types.LinterPriority, dryRun bool) (*types.MigrationResult, error) {
	f.logger.Infof("Loading configuration: %s", configPath)

	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to load config", configPath, err)
	}

	f.logger.Infof("Analyzing configuration...")

	analysis, err := f.analyzer.AnalyzeConfig(configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to analyze config", configPath, err)
	}

	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	disabledLinters := f.configLoader.GetLintersDisabled(cfg)

	// Handle deprecated linters first
	deprecationFixes := 0
	enableFixes := 0
	messages := []string{}

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
			if !linterSet[replacement.Replacement] {
				if dryRun {
					f.logger.Infof("[DRY-RUN] Would replace deprecated linter: %s -> %s (%s)", linter, replacement.Replacement, replacement.Reason)
				} else {
					f.logger.Infof("Replacing deprecated linter: %s -> %s (%s)", linter, replacement.Replacement, replacement.Reason)

					linterSet[replacement.Replacement] = true

					messages = append(messages, fmt.Sprintf("Replaced deprecated %s with %s: %s", linter, replacement.Replacement, replacement.Reason))
				}
			} else {
				if dryRun {
					f.logger.Infof("[DRY-RUN] Would remove deprecated %s (keeping existing %s)", linter, replacement.Replacement)
				} else {
					f.logger.Debugf("Removing deprecated %s (keeping existing %s)", linter, replacement.Replacement)
					messages = append(messages, fmt.Sprintf("Removed deprecated %s (replacement %s already enabled): %s", linter, replacement.Replacement, replacement.Reason))
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
			lintName = replacement.Replacement // Use the replacement name instead
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

				messages = append(messages, fmt.Sprintf("Enabled %s: %s", lintName, rec.Reason))
			}
		}
	}

	totalFixes := deprecationFixes + enableFixes

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would apply %d fixes", totalFixes)

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

	f.logger.Infof("Creating backup...")

	backupPath, err := f.configLoader.CreateBackup(configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to create backup", configPath, err)
	}

	// Convert final linter set to sorted slice for consistent output
	enabledLinters = make([]string, 0, len(linterSet))
	for linter := range linterSet {
		enabledLinters = append(enabledLinters, linter)
	}

	cfg.Linters.Enable = enabledLinters
	cfg.Linters.Disable = []string{}

	f.logger.Infof("Saving configuration...")

	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return nil, errors.NewAnalysisError("failed to save config", configPath, err)
	}

	result := &types.MigrationResult{
		Success:      true,
		FixesApplied: totalFixes,
		Message:      fmt.Sprintf("Successfully applied %d fixes", totalFixes),
		BackupPath:   backupPath,
	}

	return result, nil
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
