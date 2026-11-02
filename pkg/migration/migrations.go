// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import (
	"maps"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// migrateLintersSettings migrates linters-settings to linters.settings.
func (m *Migrator) migrateLintersSettings(config *Config) bool {
	if len(config.LintersSettingsV1) == 0 {
		return false
	}

	if m.dryRun {
		return true
	}

	initSettings(&config.Linters.Settings)

	maps.Copy(config.Linters.Settings, config.LintersSettingsV1)
	config.LintersSettingsV1 = nil

	return true
}

func initSettings(settings *map[string]any) {
	if *settings == nil {
		*settings = make(map[string]any)
	}
}

// migrateIssuesProperties migrates all issues properties.
func (m *Migrator) migrateIssuesProperties(config *Config, wasV1 bool) int {
	fixes := 0

	if m.migrateIssuesExcludeRules(config) {
		fixes += m.logFixApplied("Migrated 'issues.exclude-rules' to 'linters.exclusions.rules'")
	}

	if m.migrateIssuesExclude(config) {
		fixes += m.logFixApplied("Migrated 'issues.exclude' to 'linters.exclusions.rules'")
	}

	if m.migrateIssuesExcludeDirs(config) {
		fixes += m.logFixApplied("Migrated 'issues.exclude-dirs' to exclusions.paths")
	}

	if m.migrateIssuesExcludeFiles(config) {
		fixes += m.logFixApplied("Migrated 'issues.exclude-files' to exclusions.paths")
	}

	if m.migrateIssuesExcludeUseDefault(config, wasV1) {
		fixes += m.logFixApplied("Migrated 'issues.exclude-use-default' to 'linters.exclusions.presets'")
	}

	if m.migrateIssuesExcludeDirUseDefault(config, wasV1) {
		fixes += m.logFixApplied("Migrated 'issues.exclude-dir-use-default' to exclusions.paths (default directories)")
	}

	flagFixes := m.migrateIssuesFlags(config)
	fixes += flagFixes

	return fixes
}

// migrateIssuesExcludeRules migrates issues.exclude-rules to linters.exclusions.rules.
//
//nolint:exhaustruct // Only partial fields needed for migration
func (m *Migrator) migrateIssuesExcludeRules(config *Config) bool {
	if len(config.ExcludeRules) == 0 {
		return false
	}

	if m.dryRun {
		return true
	}

	for _, rule := range config.ExcludeRules {
		config.Linters.Exclusions.Rules = append(config.Linters.Exclusions.Rules, ExclusionRule(rule))
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

	appendUnique(&config.Linters.Exclusions.Paths, config.ExcludeDirs)
	appendUnique(&config.Formatters.Exclusions.Paths, config.ExcludeDirs)
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

	appendUnique(&config.Linters.Exclusions.Paths, config.ExcludeFiles)
	appendUnique(&config.Formatters.Exclusions.Paths, config.ExcludeFiles)
	config.ExcludeFiles = nil

	return true
}

// appendUnique appends items to target, skipping entries already present.
func appendUnique(target *[]string, items []string) {
	seen := make(map[string]bool, len(*target)+len(items))
	for _, existing := range *target {
		seen[existing] = true
	}

	for _, item := range items {
		if seen[item] {
			continue
		}

		seen[item] = true
		*target = append(*target, item)
	}
}

// excludeRulePathAllGo is the path criterion the migration guide prescribes
// for exclusion rules converted from the plain 'issues.exclude' list.
const excludeRulePathAllGo = `(.+)\.go$`

// migrateIssuesExclude migrates the plain 'issues.exclude' regex list to
// 'linters.exclusions.rules' entries (path-anchored, per the upstream migration
// guide), folding v1's case-insensitive default into a (?i) prefix.
func (m *Migrator) migrateIssuesExclude(config *Config) bool {
	if len(config.Exclude) == 0 && config.ExcludeCaseSensitive == TriStateUnspecified {
		return false
	}

	if m.dryRun {
		return true
	}

	caseInsensitive := config.ExcludeCaseSensitive != TriStateEnabled

	config.Linters.Exclusions.Rules = appendExcludeRules(
		config.Linters.Exclusions.Rules, config.Exclude, caseInsensitive,
	)

	config.Exclude = nil
	config.ExcludeCaseSensitive = TriStateUnspecified

	return true
}

// appendExcludeRules converts the plain v1 'issues.exclude' regex list into
// path-anchored exclusion rules, skipping rules that already exist.
func appendExcludeRules(rules []ExclusionRule, patterns []string, caseInsensitive bool) []ExclusionRule {
	seen := make(map[string]bool, len(rules))

	for _, rule := range rules {
		seen[rule.Path+"\x00"+rule.Text] = true
	}

	for _, pattern := range patterns {
		text := pattern
		if caseInsensitive {
			text = "(?i)" + pattern
		}

		key := excludeRulePathAllGo + "\x00" + text
		if seen[key] {
			continue
		}

		seen[key] = true

		rules = append(rules, ExclusionRule{Path: excludeRulePathAllGo, Text: text})
	}

	return rules
}

// migrateIssuesExcludeUseDefault migrates 'issues.exclude-use-default' to the
// four v2 exclusion presets that supersede the v1 default exclusions. An
// explicit v1 `false` opted out of the defaults, so it converts to no presets.
// A v1 config without the key had the defaults active (v1 default true), so
// genuine v1 configs get the presets even when the key is absent.
func (m *Migrator) migrateIssuesExcludeUseDefault(config *Config, wasV1 bool) bool {
	if config.ExcludeUseDefault == TriStateUnspecified && !wasV1 {
		return false
	}

	defaultsActive := wasV1 && config.ExcludeUseDefault != TriStateDisabled
	if defaultsActive && !m.dryRun {
		mergeDefaultExclusionPresets(&config.Linters.Exclusions.Presets)
	}

	if m.dryRun {
		return true
	}

	config.ExcludeUseDefault = TriStateUnspecified

	return true
}

// migrateIssuesExcludeDirUseDefault migrates 'issues.exclude-dir-use-default'
// to the default directory exclusions the guide prescribes. Explicit v1
// `false` opted out; genuine v1 configs without the key had the defaults
// active, mirroring migrateIssuesExcludeUseDefault.
func (m *Migrator) migrateIssuesExcludeDirUseDefault(config *Config, wasV1 bool) bool {
	if config.ExcludeDirUseDefault == TriStateUnspecified && !wasV1 {
		return false
	}

	defaultsActive := wasV1 && config.ExcludeDirUseDefault != TriStateDisabled
	if defaultsActive && !m.dryRun {
		mergeDefaultExcludeDirs(config)
	}

	if m.dryRun {
		return true
	}

	config.ExcludeDirUseDefault = TriStateUnspecified

	return true
}

// migrateRunSkipDirsUseDefault migrates 'run.skip-dirs-use-default' (deprecated
// in v1.57.0 in favor of exclude-dir-use-default, itself removed in v2) to the
// same default directory exclusions.
func (m *Migrator) migrateRunSkipDirsUseDefault(config *Config, wasV1 bool) bool {
	if config.Run.SkipDirsUseDefault == TriStateUnspecified && !wasV1 {
		return false
	}

	defaultsActive := wasV1 && config.Run.SkipDirsUseDefault != TriStateDisabled
	if defaultsActive && !m.dryRun {
		mergeDefaultExcludeDirs(config)
	}

	if m.dryRun {
		return true
	}

	config.Run.SkipDirsUseDefault = TriStateUnspecified

	return true
}

// mergeDefaultExclusionPresets unions the four v2 exclusion presets (the v1
// default exclusions) into the existing preset list without duplicates.
func mergeDefaultExclusionPresets(presets *[]string) {
	defaults := []string{"comments", "common-false-positives", "legacy", "std-error-handling"}

	existing := make(map[string]bool, len(*presets)+len(defaults))
	for _, preset := range *presets {
		existing[preset] = true
	}

	for _, preset := range defaults {
		if existing[preset] {
			continue
		}

		existing[preset] = true
		*presets = append(*presets, preset)
	}
}

// mergeDefaultExcludeDirs unions the default directory exclusions prescribed by
// the migration guide into both linter and formatter exclusion paths.
func mergeDefaultExcludeDirs(config *Config) {
	defaults := []string{"third_party$", "builtin$", "examples$"}
	appendUnique(&config.Linters.Exclusions.Paths, defaults)
	appendUnique(&config.Formatters.Exclusions.Paths, defaults)
}

// migrateIssuesFlags migrates deprecated boolean flags.
func (m *Migrator) migrateIssuesFlags(config *Config) int {
	if m.dryRun {
		return m.countIssuesFlags(config)
	}

	fixes := m.clearIssuesFlags(config)

	return fixes
}

func (m *Migrator) countIssuesFlags(config *Config) int {
	if config.ExcludeRulesUseDefault != TriStateUnspecified {
		return 1
	}

	return 0
}

// migrateFormatters migrates formatters from linters.enable to formatters.enable.
func (m *Migrator) migrateFormatters(config *Config) bool {
	formattersToEnable, remainingLinters := m.partitionFormatters(config.Linters.Enable)
	if len(formattersToEnable) == 0 {
		return false
	}

	if m.dryRun {
		return true
	}

	config.Linters.Enable = remainingLinters
	config.Formatters.Enable = append(config.Formatters.Enable, formattersToEnable...)

	return true
}

func (m *Migrator) partitionFormatters(enabled []string) ([]string, []string) {
	formatterNames := types.NewSet("gofmt", "goimports", "gofumpt")

	var formatters, linters []string

	for _, name := range enabled {
		if formatterNames.Contains(name) {
			formatters = append(formatters, name)
		} else {
			linters = append(linters, name)
		}
	}

	return formatters, linters
}

// migrateFormatterSettingsFromLinters moves formatter settings from linters.settings to formatters.settings.
func migrateFormatterSettingsFromLinters(config *Config) int {
	formatterSettingNames := []string{"gofmt", "goimports", "gofumpt", "gci"}
	fixes := 0

	initSettings(&config.Formatters.Settings)

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
		return m.countOutputProperties(config)
	}

	return m.clearOutputProperties(config)
}

func (m *Migrator) countOutputProperties(config *Config) int {
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

func (m *Migrator) clearOutputProperties(config *Config) int {
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
		mergeOutputFormatBlocks(&config.Output)
		config.Output.Format = ""
		fixes++
	}

	return fixes
}

// mergeOutputFormatBlocks folds the parsed v1 output.format blocks into
// output.formats without overwriting existing entries.
func mergeOutputFormatBlocks(out *Output) {
	blocks := outputFormatBlocks(out.Format)
	if len(blocks) == 0 {
		return
	}

	if out.Formats == nil {
		out.Formats = make(map[string]map[string]any, len(blocks))
	}

	for name, block := range blocks {
		if _, exists := out.Formats[name]; !exists {
			out.Formats[name] = block
		}
	}
}

func (m *Migrator) clearIssuesFlags(config *Config) int {
	if config.ExcludeRulesUseDefault != TriStateUnspecified {
		config.ExcludeRulesUseDefault = TriStateUnspecified

		return 1
	}

	return 0
}

// outputFormatPathKey is the v2 formats-block key naming the output target.
const outputFormatPathKey = "path"

// outputFormatBlocks parses a v1 'output.format' value (comma-separated
// format[:path] tokens) into v2 'output.formats' blocks per the migration
// guide. Unknown v2-dropped formats (e.g. github-actions) are skipped.
func outputFormatBlocks(format string) map[string]map[string]any {
	blocks := make(map[string]map[string]any)

	for token := range strings.SplitSeq(format, ",") {
		if name, block, ok := outputFormatBlock(token); ok {
			blocks[name] = block
		}
	}

	if len(blocks) == 0 {
		return nil
	}

	return blocks
}

// outputFormatBlock converts one v1 output.format token (format[:path]) into
// its v2 formats block; v2-dropped formats report ok=false and vanish.
func outputFormatBlock(token string) (string, map[string]any, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", nil, false
	}

	name, path, _ := strings.Cut(token, ":")
	if path == "" {
		path = "stdout"
	}

	switch name {
	case "colored-line-number":
		return "text", map[string]any{outputFormatPathKey: path, "colors": true}, true
	case "line-number":
		return "text", map[string]any{outputFormatPathKey: path}, true
	case "colored-tab":
		return "tab", map[string]any{outputFormatPathKey: path, "colors": true}, true
	case "tab":
		return "tab", map[string]any{outputFormatPathKey: path}, true
	case "junit-xml-extended":
		return "junit-xml", map[string]any{outputFormatPathKey: path, "extended": true}, true
	case "html", "checkstyle", "code-climate", "junit-xml", "json", "teamcity", "sarif":
		return name, map[string]any{outputFormatPathKey: path}, true
	}

	return "", nil, false
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
		run.Timeout = constants.DefaultTimeout

		return true
	}

	return false
}
