package linter

import (
	"fmt"
	"time"

	"charm.land/log/v2"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// DefaultTimeout is the default timeout value used when the config has an invalid duration.
const DefaultTimeout = "5m"

// preFixVersion ensures the config has the correct version field for golangci-lint v2.
// This is necessary because golangci-lint linters command will fail if the config
// has an incompatible version.
func (f *Fixer) preFixVersion(cfg *types.Config, configPath string, dryRun bool) error {
	if cfg.Version == "2" {
		return nil
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

	err := f.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return apperrors.NewConfigError(
			fmt.Sprintf("failed to save pre-fixed config (dryRun=%t)", dryRun),
			configPath,
			err,
		)
	}

	return nil
}

// preFixInvalidDurations fixes invalid duration fields in the config.
// This is necessary because golangci-lint will fail with "time: invalid duration" error
// if fields like run.timeout have invalid values (e.g., empty string).
// Returns true if the config has invalid durations that need fixing.
func (f *Fixer) preFixInvalidDurations(
	cfg *types.Config,
	configPath string,
	dryRun bool,
) (needsFixing bool, err error) {
	// Check if timeout is empty or invalid
	if cfg.Run.Timeout == "" {
		f.logger.Infof("run.timeout is empty, would set to %q", DefaultTimeout)

		if dryRun {
			f.logger.Infof("[DRY-RUN] Would set run.timeout to %q", DefaultTimeout)

			return true, nil // Needs fixing, but don't save in dry-run
		}

		cfg.Run.Timeout = DefaultTimeout
	} else if _, parseErr := time.ParseDuration(cfg.Run.Timeout); parseErr != nil {
		f.logger.Infof("run.timeout %q is invalid, would set to %q", cfg.Run.Timeout, DefaultTimeout)

		if dryRun {
			f.logger.Infof("[DRY-RUN] Would set run.timeout to %q", DefaultTimeout)

			return true, nil // Needs fixing, but don't save in dry-run
		}

		cfg.Run.Timeout = DefaultTimeout
	} else {
		return false, nil // No fix needed
	}

	// Save the fixed config (only in non-dry-run mode)
	err = f.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return false, apperrors.NewConfigError(
			fmt.Sprintf("failed to save pre-fixed config (dryRun=%t)", dryRun),
			configPath,
			err,
		)
	}

	return false, nil // Fixed, no longer needs fixing
}

// calculateDryRunResultWithInvalidDurations calculates the dry-run result when invalid durations are present.
// Since the config has invalid durations, we can't run golangci-lint linters for analysis,
// so we just report what would be fixed regarding durations.
func (f *Fixer) calculateDryRunResultWithInvalidDurations(cfg *types.Config) types.MigrationResultType {
	f.logger.Infof("[DRY-RUN] Would fix invalid run.timeout: %q -> %q", cfg.Run.Timeout, DefaultTimeout)

	return types.OkMigration(&types.MigrationResult{
		FixesApplied: 1,
		Message: fmt.Sprintf(
			"Would apply 1 fix (dry-run mode, skipped analysis due to invalid duration: run.timeout=%q)",
			cfg.Run.Timeout,
		),
		NextSteps: []string{
			"Run without --dry-run to fix the invalid duration",
			"Then run 'golangci-lint run --fix' to auto-fix code issues",
		},
	})
}

// preFixDeprecatedLinters replaces deprecated linters in the config before analysis.
// This is necessary because golangci-lint linters command will fail if the config
// contains deprecated/removed linters (even in the disable list).
func (f *Fixer) preFixDeprecatedLinters(cfg *types.Config, configPath string, dryRun bool) error {
	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	disabledLinters := f.configLoader.GetLintersDisabled(cfg)

	var deprecatedFound []string

	linterSet := make(map[string]bool)

	for _, linter := range enabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			deprecatedFound = append(deprecatedFound, linter)

			continue
		}

		linterSet[linter] = true
	}

	disabledSet := make(map[string]bool)

	for _, linter := range disabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			deprecatedFound = append(deprecatedFound, linter+" (disabled)")

			continue
		}

		disabledSet[linter] = true
	}

	if len(deprecatedFound) == 0 {
		return nil
	}

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would pre-fix %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)

		return nil
	}

	f.logger.Infof("Pre-fixing %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)

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

	err := f.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return apperrors.NewConfigError(
			fmt.Sprintf("failed to save pre-fixed config (dryRun=%t, deprecatedFound=%v)", dryRun, deprecatedFound),
			configPath,
			err,
		)
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

// FixerPreflight provides pre-flight fixing functionality.
type FixerPreflight struct {
	logger       *log.Logger
	configLoader types.ConfigLoader
}

// NewFixerPreflight creates a new pre-flight fixer.
func NewFixerPreflight(logger *log.Logger, configLoader types.ConfigLoader) *FixerPreflight {
	return &FixerPreflight{
		logger:       logger,
		configLoader: configLoader,
	}
}

// EnsureVersion ensures the config has the correct version field.
func (p *FixerPreflight) EnsureVersion(cfg *types.Config, configPath string, dryRun bool) error {
	if cfg.Version == "2" {
		return nil
	}

	if cfg.Version == "" {
		p.logger.Infof("Version field is empty, setting to \"2\" for golangci-lint v2 compatibility")
	} else {
		p.logger.Infof("Version field is \"%s\", setting to \"2\" for golangci-lint v2 compatibility", cfg.Version)
	}

	if dryRun {
		p.logger.Infof("[DRY-RUN] Would set version to \"2\"")

		return nil
	}

	cfg.Version = "2"

	err := p.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return apperrors.NewConfigError(
			fmt.Sprintf("failed to save pre-fixed config (dryRun=%t)", dryRun),
			configPath,
			err,
		)
	}

	return nil
}

// RemoveDeprecatedLinters removes deprecated linters from the config.
func (p *FixerPreflight) RemoveDeprecatedLinters(cfg *types.Config, configPath string, dryRun bool) ([]string, error) {
	enabledLinters := p.configLoader.GetLintersEnabled(cfg)
	disabledLinters := p.configLoader.GetLintersDisabled(cfg)

	var deprecatedFound []string

	linterSet := make(map[string]bool)

	for _, linter := range enabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			deprecatedFound = append(deprecatedFound, linter)

			continue
		}

		linterSet[linter] = true
	}

	disabledSet := make(map[string]bool)

	for _, linter := range disabledLinters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			deprecatedFound = append(deprecatedFound, linter+" (disabled)")

			continue
		}

		disabledSet[linter] = true
	}

	if len(deprecatedFound) == 0 {
		return nil, nil
	}

	if dryRun {
		p.logger.Infof("[DRY-RUN] Would pre-fix %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)

		return deprecatedFound, nil
	}

	p.logger.Infof("Pre-fixing %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)

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

	err := p.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return nil, apperrors.NewAnalysisError("failed to save pre-fixed config", configPath, err)
	}

	return deprecatedFound, nil
}
