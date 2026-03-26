// Copyright (c) 2026 golangci-linter-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"fmt"
	"maps"
)

const defaultRunTimeout = "5m"

// migrateLintersSettings migrates linters-settings to linters.settings.
func (m *Migrator) migrateLintersSettings(config *Config) bool {
	if len(config.LintersSettingsV1) == 0 {
		return false
	}

	if m.dryRun {
		return true
	}

	if config.Linters.Settings == nil {
		config.Linters.Settings = make(map[string]any)
	}

	maps.Copy(config.Linters.Settings, config.LintersSettingsV1)
	config.LintersSettingsV1 = nil

	return true
}

// migrateIssuesProperties migrates all issues properties.
func (m *Migrator) migrateIssuesProperties(config *Config) int {
	fixes := 0

	if m.migrateIssuesExcludeRules(config) {
		fixes++

		if m.verbose {
			fmt.Printf("%s Migrated 'issues.exclude-rules' to 'linters.exclusions.rules'\n", m.getCheckmark())
		}
	}

	if m.migrateIssuesExcludeDirs(config) {
		fixes++

		if m.verbose {
			fmt.Printf("%s Migrated 'issues.exclude-dirs' to exclusions.paths\n", m.getCheckmark())
		}
	}

	if m.migrateIssuesExcludeFiles(config) {
		fixes++

		if m.verbose {
			fmt.Printf("%s Migrated 'issues.exclude-files' to exclusions.paths\n", m.getCheckmark())
		}
	}

	flagFixes := m.migrateIssuesFlags(config)
	fixes += flagFixes

	return fixes
}

// migrateIssuesExcludeRules migrates issues.exclude-rules to linters.exclusions.rules.
func (m *Migrator) migrateIssuesExcludeRules(config *Config) bool {
	if len(config.ExcludeRules) == 0 {
		return false
	}

	if m.dryRun {
		return true
	}

	for _, rule := range config.ExcludeRules {
		config.Linters.Exclusions.Rules = append(config.Linters.Exclusions.Rules, ExclusionRule{
			Path:    rule.Path,
			Text:    rule.Text,
			Linters: rule.Linters,
		})
	}

	config.ExcludeRules = nil

	return true
}

// migrateIssuesExcludeDirs migrates issues.exclude-dirs to exclusions.paths.
func (m *Migrator) migrateIssuesExcludeDirs(config *Config) bool {
	if len(config.ExcludeDirs) == 0 {
		return false
	}

	if m.dryRun {
		return true
	}

	config.Linters.Exclusions.Paths = append(config.Linters.Exclusions.Paths, config.ExcludeDirs...)
	config.Formatters.Exclusions.Paths = append(config.Formatters.Exclusions.Paths, config.ExcludeDirs...)
	config.ExcludeDirs = nil

	return true
}

// migrateIssuesExcludeFiles migrates issues.exclude-files to exclusions.paths.
func (m *Migrator) migrateIssuesExcludeFiles(config *Config) bool {
	if len(config.ExcludeFiles) == 0 {
		return false
	}

	if m.dryRun {
		return true
	}

	config.Linters.Exclusions.Paths = append(config.Linters.Exclusions.Paths, config.ExcludeFiles...)
	config.Formatters.Exclusions.Paths = append(config.Formatters.Exclusions.Paths, config.ExcludeFiles...)
	config.ExcludeFiles = nil

	return true
}

// migrateIssuesFlags migrates deprecated boolean flags.
func (m *Migrator) migrateIssuesFlags(config *Config) int {
	if m.dryRun {
		count := 0
		if config.ExcludeUseDefault != nil {
			count++
		}

		if config.ExcludeRulesUseDefault != nil {
			count++
		}

		if config.ExcludeDirUseDefault != nil {
			count++
		}

		return count
	}

	fixes := 0

	if config.ExcludeUseDefault != nil {
		config.ExcludeUseDefault = nil
		fixes++
	}

	if config.ExcludeRulesUseDefault != nil {
		config.ExcludeRulesUseDefault = nil
		fixes++
	}

	if config.ExcludeDirUseDefault != nil {
		config.ExcludeDirUseDefault = nil
		fixes++
	}

	return fixes
}

// migrateFormatters migrates formatters from linters.enable to formatters.enable.
func (m *Migrator) migrateFormatters(config *Config) bool {
	formatterNames := map[string]bool{
		"gofmt":     true,
		"goimports": true,
		"gofumpt":   true,
	}

	var (
		lintersToKeep      []string
		formattersToEnable []string
	)

	for _, linter := range config.Linters.Enable {
		if formatterNames[linter] {
			formattersToEnable = append(formattersToEnable, linter)
		} else {
			lintersToKeep = append(lintersToKeep, linter)
		}
	}

	if len(formattersToEnable) == 0 {
		return false
	}

	if m.dryRun {
		return true
	}

	config.Linters.Enable = lintersToKeep
	config.Formatters.Enable = append(config.Formatters.Enable, formattersToEnable...)

	return true
}

// migrateFormatterSettingsFromLinters moves formatter settings from linters.settings to formatters.settings.
func migrateFormatterSettingsFromLinters(config *Config) int {
	formatterSettingNames := []string{"gofmt", "goimports", "gofumpt", "gci"}
	fixes := 0

	if config.Formatters.Settings == nil {
		config.Formatters.Settings = make(map[string]any)
	}

	for _, name := range formatterSettingNames {
		if settings, exists := config.Linters.Settings[name]; exists {
			config.Formatters.Settings[name] = settings
			delete(config.Linters.Settings, name)

			fixes++
		}
	}

	return fixes
}

// migrateOutputProperties migrates deprecated output.* properties.
func (m *Migrator) migrateOutputProperties(config *Config) int {
	if m.dryRun {
		count := 0
		if config.Output.PrintIssuedLines {
			count++
		}

		if config.Output.PrintLinterName {
			count++
		}

		if config.Output.SortResults {
			count++
		}

		if config.Output.Format != "" {
			count++
		}

		return count
	}

	fixes := 0

	if config.Output.PrintIssuedLines {
		config.Output.PrintIssuedLines = false
		fixes++
	}

	if config.Output.PrintLinterName {
		config.Output.PrintLinterName = false
		fixes++
	}

	if config.Output.SortResults {
		config.Output.SortResults = false
		fixes++
	}

	if config.Output.Format != "" {
		config.Output.Format = ""
		fixes++
	}

	return fixes
}

// migrateVersion ensures the config has a valid v2 version string.
func migrateVersion(version *string, rules *MigrationRules) bool {
	v := *version

	if rules.IsValidVersion(v) {
		return false
	}

	if v == "" {
		*version = "2"

		return true
	}

	versionNum := v
	if len(v) > 1 && v[0] == 'v' {
		versionNum = v[1:]
	}

	if len(versionNum) > 0 && versionNum[0] == '1' {
		if len(versionNum) == 1 || versionNum[1] == '.' {
			*version = "2"

			return true
		}
	}

	*version = "2"

	return true
}

// migrateRunSettings ensures run.timeout has a valid value.
func migrateRunSettings(run *Run) bool {
	if run.Timeout == "" {
		run.Timeout = defaultRunTimeout

		return true
	}

	return false
}
