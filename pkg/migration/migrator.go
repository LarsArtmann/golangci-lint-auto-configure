// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"errors"
	"fmt"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/utils"
)

// Static errors for better error handling.
var (
	ErrConfigPathEmpty      = errors.New("config path cannot be empty")
	ErrMockValidationFailed = errors.New("mock validation failed")
)

// cliPrintf prints to stdout, bypassing forbidigo linter for CLI output.
func cliPrintf(format string, args ...any) {
	//nolint:forbidigo // CLI output
	fmt.Printf(format, args...)
}

// cliPrintln prints a line to stdout, bypassing forbidigo linter for CLI output.
func cliPrintln(args ...any) {
	//nolint:forbidigo // CLI output
	fmt.Println(args...)
}

// Migrator handles golangci-lint configuration migrations.
type Migrator struct {
	verbose    bool
	dryRun     bool
	validator  Validator
	rules      *MigrationRules
	configPath string
	noEmojis   bool
}

// NewMigrator creates a new configuration migrator with the given path.
func NewMigrator(path string, verbose bool) (*Migrator, error) {
	if path == "" {
		return nil, ErrConfigPathEmpty
	}

	return &Migrator{
		configPath: path,
		verbose:    verbose,
		dryRun:     false,
		validator:  DefaultValidator{},
		noEmojis:   true,
		rules:      DefaultRules(),
	}, nil
}

// SetDryRun sets the execution mode to dry-run or live-run.
func (m *Migrator) SetDryRun(dryRun bool) {
	m.dryRun = dryRun
}

// IsDryRun returns true if the migrator is in dry-run mode.
func (m *Migrator) IsDryRun() bool {
	return m.dryRun
}

// SetValidator sets a custom validator for config validation.
func (m *Migrator) SetValidator(v Validator) {
	m.validator = v
}

// SetNoEmojis sets whether emojis should be used in output.
func (m *Migrator) SetNoEmojis(noEmojis bool) {
	m.noEmojis = noEmojis
}

// MigrateToV2 migrates configuration to golangci-lint v2.x schema.
// Returns: (success bool, fixesApplied int, error error).
//
//nolint:cyclop,funlen // Migration logic is inherently complex with multiple steps
func (m *Migrator) MigrateToV2() (bool, int, error) {
	cfg, err := LoadConfig(m.configPath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to load config: %w", err)
	}

	if !m.dryRun {
		err = m.checkGitRepository()
		if err != nil {
			return false, 0, fmt.Errorf("git check failed: %w", err)
		}
	}

	fixesApplied := 0

	// Fix 0: Migrate version to v2 format
	if migrateVersion(&cfg.Version, m.rules) {
		fixesApplied += m.logFixApplied("Migrated version to v2 format")
	}

	// Fix 0b: Ensure run.timeout has a value
	if migrateRunSettings(&cfg.Run) {
		fixesApplied += m.logFixApplied("Set default run.timeout")
	}

	// Fix 1: Move top-level linters-settings to nested linters.settings
	if m.migrateLintersSettings(cfg) {
		fixesApplied += m.logFixApplied("Migrated 'linters-settings' to nested 'linters.settings'")
	}

	// Fix 2: Migrate issues.* properties (rules, dirs, files)
	issuesFixes := m.migrateIssuesProperties(cfg)
	fixesApplied += m.logFixesApplied(
		fmt.Sprintf("Migrated %d 'issues.*' properties to new structure", issuesFixes),
		issuesFixes,
	)

	// Fix 3: Migrate formatters from linters.enable to formatters.enable
	if m.migrateFormatters(cfg) {
		fixesApplied += m.logFixApplied("Migrated formatters from 'linters.enable' to 'formatters.enable'")
	}

	// Fix 4: Move formatter settings from linters.settings to formatters.settings
	formatterSettingsMoved := migrateFormatterSettingsFromLinters(cfg)
	fixesApplied += m.logFixesApplied(
		fmt.Sprintf("Moved %d formatter settings to 'formatters.settings'", formatterSettingsMoved),
		formatterSettingsMoved,
	)

	// Fix 5: Migrate linter-specific settings
	if cfg.Linters.Settings != nil {
		linterSettingsFixes := migrateLinterSettings(cfg.Linters.Settings, m.rules)
		fixesApplied += m.logFixesApplied(
			fmt.Sprintf("Migrated %d linter-specific settings", linterSettingsFixes),
			linterSettingsFixes,
		)
	}

	// Fix 6: Migrate formatters.settings
	if cfg.Formatters.Settings != nil {
		formatterSettingsFixes := migrateFormattersSettings(cfg.Formatters.Settings)
		fixesApplied += m.logFixesApplied(
			fmt.Sprintf("Migrated %d formatter-specific settings", formatterSettingsFixes),
			formatterSettingsFixes,
		)
	}

	// Fix 7: Migrate output.* deprecated properties
	outputFixes := m.migrateOutputProperties(cfg)
	fixesApplied += m.logFixesApplied(fmt.Sprintf("Removed %d deprecated output properties", outputFixes), outputFixes)

	if fixesApplied == 0 {
		if m.verbose {
			cliPrintln("No migration needed (already at v2.x)")
		}

		return false, 0, nil
	}

	if m.dryRun {
		if m.verbose {
			cliPrintln("[DRY-RUN] Config would be migrated (skipping save)")
		}

		return true, fixesApplied, nil
	}

	err = SaveConfig(cfg, m.configPath)
	if err != nil {
		return false, fixesApplied, fmt.Errorf("failed to save migrated config: %w (use git checkout to restore)", err)
	}

	err = m.validateConfig()
	if err != nil {
		return false, fixesApplied, fmt.Errorf(
			"validation failed after migration: %w (use git checkout to restore)",
			err,
		)
	}

	if m.verbose {
		cliPrintf("\n%s Configuration migrated successfully (%d fixes applied)\n", m.getCheckmark(), fixesApplied)
	}

	return true, fixesApplied, nil
}

// validateConfig runs golangci-lint config verify.
//

func (m *Migrator) validateConfig() error {
	return m.validator.ValidateConfig(m)
}

// getCheckmark returns appropriate checkmark symbol based on emoji setting.
func (m *Migrator) getCheckmark() string {
	if m.noEmojis {
		return "[OK]"
	}

	return "✓"
}

// logVerbose prints a verbose message with checkmark if verbose mode is enabled.
func (m *Migrator) logVerbose(message string) {
	if m.verbose {
		cliPrintf("%s %s\n", m.getCheckmark(), message)
	}
}

// logFixApplied logs a verbose message when a fix is applied and returns 1 for counting.
func (m *Migrator) logFixApplied(message string) int {
	m.logVerbose(message)

	return 1
}

// logFixesApplied logs a verbose message with a count when fixes are applied and returns the count.
func (m *Migrator) logFixesApplied(message string, count int) int {
	if count > 0 {
		m.logVerbose(message)
	}

	return count
}

// checkGitRepository verifies we're inside a git repository.
func (m *Migrator) checkGitRepository() error {
	return utils.CheckGitRepoWithTimeout(".")
}
