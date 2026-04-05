package linter

import (
	"fmt"
	"time"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

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

	f.logger.Infof("%s, would set to %q", reason, constants.DefaultTimeout)

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would set run.timeout to %q", constants.DefaultTimeout)

		return true, nil
	}

	cfg.Run.Timeout = constants.DefaultTimeout

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
	f.logger.Infof("[DRY-RUN] Would fix invalid run.timeout: %q -> %q", cfg.Run.Timeout, constants.DefaultTimeout)

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

	linterSet := types.NewSet[string]()
	deprecatedFound = filterDeprecatedFrom(enabledLinters, "", deprecatedFound, linterSet)

	disabledSet := types.NewSet[string]()
	deprecatedFound = filterDeprecatedFrom(disabledLinters, " (disabled)", deprecatedFound, disabledSet)

	if len(deprecatedFound) == 0 {
		return nil
	}

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would pre-fix %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)

		return nil
	}

	f.logger.Infof("Pre-fixing %d deprecated linters: %v", len(deprecatedFound), deprecatedFound)

	cfg.Linters.Enable = setToSortedSlice(linterSet)
	cfg.Linters.Disable = setToSortedSlice(disabledSet)

	return f.savePrefixedConfig(cfg, configPath, fmt.Sprintf("deprecatedFound=%v", deprecatedFound))
}

func filterDeprecatedFrom(
	linters []string,
	suffix string,
	found []string,
	result types.Set[string],
) []string {
	for _, linter := range linters {
		if _, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]; isDeprecated {
			found = append(found, linter+suffix)

			continue
		}

		result.Add(linter)
	}

	return found
}

// preFixTypecheck removes typecheck from the config before analysis.
// This is necessary because golangci-lint linters command will fail if typecheck
// is in the enable list (typecheck is not a configurable linter in v2).
func (f *Fixer) preFixTypecheck(cfg *types.Config, configPath string, dryRun bool) (bool, error) {
	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	disabledLinters := f.configLoader.GetLintersDisabled(cfg)

	fixedEnabled, typecheckInEnabled := filterLinter(enabledLinters, "typecheck")
	fixedDisabled, typecheckInDisabled := filterLinter(disabledLinters, "typecheck")

	if !typecheckInEnabled && !typecheckInDisabled {
		return false, nil
	}

	cfg.Linters.Enable = fixedEnabled
	cfg.Linters.Disable = fixedDisabled

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would remove 'typecheck' from linters list (not configurable in v2)")
	}

	f.logger.Infof("Removing 'typecheck' from linters list (not configurable in v2)")

	return true, f.savePrefixedConfig(cfg, configPath, "after removing typecheck")
}

func filterLinter(linters []string, target string) ([]string, bool) {
	result := make([]string, 0, len(linters))
	found := false

	for _, linter := range linters {
		if linter == target {
			found = true

			continue
		}

		result = append(result, linter)
	}

	return result, found
}

// calculateDryRunResultWithDeprecated calculates the dry-run result when deprecated linters are present.
// Since the config has deprecated linters, we can't run golangci-lint linters for analysis,
// so we just report what would be fixed regarding deprecated linters.
func (f *Fixer) calculateDryRunResultWithDeprecated(cfg *types.Config) types.MigrationResultType {
	enabledLinters := f.configLoader.GetLintersEnabled(cfg)

	linterSet := types.NewSet[string]()
	deprecationFixes := f.applyDeprecatedReplacements(enabledLinters, linterSet)

	f.logger.Infof("[DRY-RUN] Would apply %d fixes", deprecationFixes)

	return types.OkMigration(&types.MigrationResult{
		FixesApplied: deprecationFixes,
		Message: fmt.Sprintf(
			"Would apply %d fixes (dry-run mode, skipped analysis due to deprecated linters)",
			deprecationFixes,
		),
	})
}

func (f *Fixer) applyDeprecatedReplacements(linters []string, linterSet types.Set[string]) int {
	deprecationFixes := 0

	for _, linter := range linters {
		replacement, isDeprecated := constants.DeprecatedLinters[types.LinterName(linter)]
		if !isDeprecated {
			linterSet.Add(linter)

			continue
		}

		deprecationFixes++

		if !linterSet.Contains(string(replacement.Replacement)) {
			f.logger.Infof("[DRY-RUN] Would replace deprecated linter: %s -> %s (%s)",
				linter, replacement.Replacement, replacement.Reason)
			linterSet.Add(string(replacement.Replacement))
		} else {
			f.logger.Infof("[DRY-RUN] Would remove deprecated %s (keeping existing %s)",
				linter, replacement.Replacement)
		}
	}

	return deprecationFixes
}

func (f *Fixer) savePrefixedConfig(cfg *types.Config, configPath, context string) error {
	err := f.configLoader.SaveConfig(cfg, configPath)
	if err != nil {
		return apperrors.NewConfigError(
			"failed to save pre-fixed config "+context,
			configPath,
			err,
		)
	}

	return nil
}
