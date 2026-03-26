// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"errors"
	"fmt"
	"os/exec"
)

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
		return nil, errors.New("config path cannot be empty")
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

// getCheckmark returns appropriate checkmark symbol based on emoji setting.
func (m *Migrator) getCheckmark() string {
	if m.noEmojis {
		return "[OK]"
	}
	return "✓"
}

// getToolIcon returns appropriate tool icon based on emoji setting.
func (m *Migrator) getToolIcon() string {
	if m.noEmojis {
		return "[TOOL]"
	}
	return "🔧"
}

// checkGitRepository verifies we're inside a git repository.
func (m *Migrator) checkGitRepository() error {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	if err != nil {
		return errors.New("not in a git repository (use git init or clone a repo first)")
	}

	if string(output[:len(output)-1]) != "true" {
		return errors.New("not inside git working tree")
	}

	return nil
}

// MigrateToV2 migrates configuration to golangci-lint v2.x schema.
// Returns: (success bool, fixesApplied int, error error).
func (m *Migrator) MigrateToV2() (bool, int, error) {
	config, err := LoadConfig(m.configPath)
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
	if migrateVersion(&config.Version, m.rules) {
		fixesApplied++
		if m.verbose {
			fmt.Printf("%s Migrated version to v2 format\n", m.getCheckmark())
		}
	}

	// Fix 0b: Ensure run.timeout has a value
	if migrateRunSettings(&config.Run) {
		fixesApplied++
		if m.verbose {
			fmt.Printf("%s Set default run.timeout\n", m.getCheckmark())
		}
	}

	// Fix 1: Move top-level linters-settings to nested linters.settings
	if m.migrateLintersSettings(config) {
		fixesApplied++
		if m.verbose {
			fmt.Printf("%s Migrated 'linters-settings' to nested 'linters.settings'\n", m.getCheckmark())
		}
	}

	// Fix 2: Migrate issues.* properties (rules, dirs, files)
	issuesFixes := m.migrateIssuesProperties(config)
	if issuesFixes > 0 {
		fixesApplied += issuesFixes
		if m.verbose {
			fmt.Printf("%s Migrated %d 'issues.*' properties to new structure\n", m.getCheckmark(), issuesFixes)
		}
	}

	// Fix 3: Migrate formatters from linters.enable to formatters.enable
	if m.migrateFormatters(config) {
		fixesApplied++
		if m.verbose {
			fmt.Printf("%s Migrated formatters from 'linters.enable' to 'formatters.enable'\n", m.getCheckmark())
		}
	}

	// Fix 4: Move formatter settings from linters.settings to formatters.settings
	formatterSettingsMoved := migrateFormatterSettingsFromLinters(config)
	if formatterSettingsMoved > 0 {
		fixesApplied += formatterSettingsMoved
		if m.verbose {
			fmt.Printf("%s Moved %d formatter settings to 'formatters.settings'\n", m.getCheckmark(), formatterSettingsMoved)
		}
	}

	// Fix 5: Migrate linter-specific settings
	if config.Linters.Settings != nil {
		linterSettingsFixes := migrateLinterSettings(config.Linters.Settings, m.rules)
		if linterSettingsFixes > 0 {
			fixesApplied += linterSettingsFixes
			if m.verbose {
				fmt.Printf("%s Migrated %d linter-specific settings\n", m.getCheckmark(), linterSettingsFixes)
			}
		}
	}

	// Fix 6: Migrate formatters.settings
	if config.Formatters.Settings != nil {
		formatterSettingsFixes := migrateFormattersSettings(config.Formatters.Settings)
		if formatterSettingsFixes > 0 {
			fixesApplied += formatterSettingsFixes
			if m.verbose {
				fmt.Printf("%s Migrated %d formatter-specific settings\n", m.getCheckmark(), formatterSettingsFixes)
			}
		}
	}

	// Fix 7: Migrate output.* deprecated properties
	outputFixes := m.migrateOutputProperties(config)
	if outputFixes > 0 {
		fixesApplied += outputFixes
		if m.verbose {
			fmt.Printf("%s Removed %d deprecated output properties\n", m.getCheckmark(), outputFixes)
		}
	}

	if fixesApplied == 0 {
		if m.verbose {
			fmt.Println("No migration needed (already at v2.x)")
		}
		return false, 0, nil
	}

	if m.dryRun {
		if m.verbose {
			fmt.Println("[DRY-RUN] Config would be migrated (skipping save)")
		}
		return true, fixesApplied, nil
	}

	err = SaveConfig(config, m.configPath)
	if err != nil {
		return false, fixesApplied, fmt.Errorf("failed to save migrated config: %w (use git checkout to restore)", err)
	}

	err = m.validateConfig()
	if err != nil {
		return false, fixesApplied, fmt.Errorf("validation failed after migration: %w (use git checkout to restore)", err)
	}

	if m.verbose {
		fmt.Printf("\n%s Configuration migrated successfully (%d fixes applied)\n", m.getCheckmark(), fixesApplied)
	}

	return true, fixesApplied, nil
}

// validateConfig runs golangci-lint config verify.
func (m *Migrator) validateConfig() error {
	return m.validator.ValidateConfig(m)
}
