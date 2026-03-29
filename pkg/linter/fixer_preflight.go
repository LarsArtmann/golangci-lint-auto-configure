package linter

import (
	"fmt"
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
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
) (bool, error) {
	needsFix, reason := needsDurationFix(cfg.Run.Timeout)
	if !needsFix {
		return false, nil
	}

	f.logger.Infof("%s, would set to %q", reason, DefaultTimeout)

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would set run.timeout to %q", DefaultTimeout)

		return true, nil
	}

	cfg.Run.Timeout = DefaultTimeout

	err := f.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return false, apperrors.NewConfigError(
			fmt.Sprintf("failed to save pre-fixed config (dryRun=%t)", dryRun),
			configPath,
			err,
		)
	}

	return false, nil
}

// needsDurationFix checks if the timeout value needs to be fixed.
// Returns true and a reason if fixing is needed, false otherwise.
func needsDurationFix(timeout string) (bool, string) {
	if timeout == "" {
		return true, "run.timeout is empty"
	}

	_, err := time.ParseDuration(timeout)
	if err != nil {
		return true, fmt.Sprintf("run.timeout %q is invalid", timeout)
	}

	return false, ""
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

// preFixTypecheck removes typecheck from the config before analysis.
// This is necessary because golangci-lint linters command will fail if typecheck
// is in the enable list (typecheck is not a configurable linter in v2).
func (f *Fixer) preFixTypecheck(cfg *types.Config, configPath string, dryRun bool) error {
	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	disabledLinters := f.configLoader.GetLintersDisabled(cfg)

	var typecheckFound bool

	// Filter out typecheck from enabled linters
	fixedEnabled := make([]string, 0, len(enabledLinters))
	for _, linter := range enabledLinters {
		if linter == "typecheck" {
			typecheckFound = true

			continue
		}

		fixedEnabled = append(fixedEnabled, linter)
	}

	// Filter out typecheck from disabled linters
	fixedDisabled := make([]string, 0, len(disabledLinters))
	for _, linter := range disabledLinters {
		if linter == "typecheck" {
			typecheckFound = true

			continue
		}

		fixedDisabled = append(fixedDisabled, linter)
	}

	if !typecheckFound {
		return nil
	}

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would remove 'typecheck' from linters list (not configurable in v2)")

		return nil
	}

	f.logger.Infof("Removing 'typecheck' from linters list (not configurable in v2)")

	cfg.Linters.Enable = fixedEnabled
	cfg.Linters.Disable = fixedDisabled

	err := f.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return apperrors.NewConfigError(
			fmt.Sprintf("failed to save pre-fixed config after removing typecheck (dryRun=%t)", dryRun),
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
