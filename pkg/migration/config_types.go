// Copyright (c) 2026 golangci-lint-auto-configure
// SPDX-License-Identifier: Apache-2.0

package migration

import "gopkg.in/yaml.v3"

// Config represents a golangci-lint configuration file.
// Supports both v1 and v2 schema versions for migration purposes.
type Config struct {
	ExcludeDirUseDefault   *bool          `yaml:"exclude-dir-use-default,omitempty"`
	ExcludeRulesUseDefault *bool          `yaml:"exclude-rules-use-default,omitempty"`
	ExcludeUseDefault      *bool          `yaml:"exclude-use-default,omitempty"`
	LintersSettingsV1      map[string]any `yaml:"linters-settings,omitempty"`
	Version                string         `yaml:"version"`
	ExcludeRules           []ExcludeRule  `yaml:"exclude-rules,omitempty"`
	ExcludeFiles           []string       `yaml:"exclude-files,omitempty"`
	ExcludeDirs            []string       `yaml:"exclude-dirs,omitempty"`
	Output                 Output         `yaml:"output,omitempty"`
	Linters                Linters        `yaml:"linters"`
	Run                    Run            `yaml:"run"`
	Formatters             Formatters     `yaml:"formatters,omitempty"`
	Issues                 Issues         `yaml:"issues,omitempty"`
}

// UnmarshalYAML custom unmarshaler to handle both v1 and v2 issues structures.
//
//nolint:gocognit,nestif,wrapcheck,wsl_v5 // Complex migration logic requires nested conditionals
func (c *Config) UnmarshalYAML(unmarshal func(any) error) error {
	temp := make(map[string]any)
	err := unmarshal(&temp)
	if err != nil {
		return err
	}

	issues, hasIssues := temp["issues"]
	if hasIssues {
		issuesMap, ok := issues.(map[string]any)
		if ok {
			if _, hasExcludeRules := issuesMap["exclude-rules"]; hasExcludeRules {
				if excludeRules, exists := issuesMap["exclude-rules"]; exists {
					temp["exclude-rules"] = excludeRules

					delete(issuesMap, "exclude-rules")
				}

				if excludeFiles, exists := issuesMap["exclude-files"]; exists {
					temp["exclude-files"] = excludeFiles

					delete(issuesMap, "exclude-files")
				}

				if excludeDirs, exists := issuesMap["exclude-dirs"]; exists {
					temp["exclude-dirs"] = excludeDirs

					delete(issuesMap, "exclude-dirs")
				}

				if excludeUseDefault, exists := issuesMap["exclude-use-default"]; exists {
					temp["exclude-use-default"] = excludeUseDefault

					delete(issuesMap, "exclude-use-default")
				}

				if excludeRulesUseDefault, exists := issuesMap["exclude-rules-use-default"]; exists {
					temp["exclude-rules-use-default"] = excludeRulesUseDefault

					delete(issuesMap, "exclude-rules-use-default")
				}

				if excludeDirUseDefault, exists := issuesMap["exclude-dir-use-default"]; exists {
					temp["exclude-dir-use-default"] = excludeDirUseDefault

					delete(issuesMap, "exclude-dir-use-default")
				}

				temp["issues"] = issuesMap
			}
		}
	}

	data, err := yaml.Marshal(temp)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, (*configWrapper)(c))
}

type configWrapper Config

// Run represents the run section of golangci-lint configuration.
type Run struct {
	SkipDirsUseDefault *bool     `yaml:"skip-dirs-use-default,omitempty"`
	Timeout            string    `yaml:"timeout,omitempty"`
	Issues             RunIssues `yaml:"issues,omitempty"`
	SkipDirs           []string  `yaml:"skip-dirs,omitempty"`
	Tests              bool      `yaml:"tests,omitempty"`
}

// RunIssues represents the issues subsection under run (deprecated v1 structure).
type RunIssues struct {
	ExcludeUseDefault      *bool         `yaml:"exclude-use-default,omitempty"`
	ExcludeRulesUseDefault *bool         `yaml:"exclude-rules-use-default,omitempty"`
	ExcludeDirUseDefault   *bool         `yaml:"exclude-dir-use-default,omitempty"`
	ExcludeRules           []ExcludeRule `yaml:"exclude-rules,omitempty"`
	ExcludeFiles           []string      `yaml:"exclude-files,omitempty"`
	ExcludeDirs            []string      `yaml:"exclude-dirs,omitempty"`
}

// Output represents the output section of golangci-lint configuration.
type Output struct {
	Format           string `yaml:"format,omitempty"`
	PrintIssuedLines bool   `yaml:"print-issued-lines,omitempty"`
	PrintLinterName  bool   `yaml:"print-linter-name,omitempty"`
	SortResults      bool   `yaml:"sort-results,omitempty"`
}

// Linters represents the linters section of golangci-lint configuration.
type Linters struct {
	Settings   map[string]any    `yaml:"settings,omitempty"`
	Default    string            `yaml:"default,omitempty"`
	Enable     []string          `yaml:"enable,omitempty"`
	Disable    []string          `yaml:"disable,omitempty"`
	Exclusions LintersExclusions `yaml:"exclusions,omitempty"`
	EnableAll  bool              `yaml:"enable-all,omitempty"`
	Fast       bool              `yaml:"fast,omitempty"`
}

// ExclusionRule represents an exclusion rule (v2.10+).
type ExclusionRule struct {
	Path       string   `yaml:"path,omitempty"`
	PathExcept string   `yaml:"path-except,omitempty"`
	Text       string   `yaml:"text,omitempty"`
	Source     string   `yaml:"source,omitempty"`
	Linters    []string `yaml:"linters,omitempty"`
}

// LintersExclusions represents exclusions for linters (v2.10+).
type LintersExclusions struct {
	Generated   string          `yaml:"generated,omitempty"`
	Presets     []string        `yaml:"presets,omitempty"`
	Rules       []ExclusionRule `yaml:"rules,omitempty"`
	Paths       []string        `yaml:"paths,omitempty"`
	PathsExcept []string        `yaml:"paths-except,omitempty"`
	WarnUnused  bool            `yaml:"warn-unused,omitempty"`
}

// Formatters represents the formatters section of golangci-lint v2 configuration.
type Formatters struct {
	Enable     []string             `yaml:"enable,omitempty"`
	Disable    []string             `yaml:"disable,omitempty"`
	Settings   map[string]any       `yaml:"settings,omitempty"`
	Exclusions FormattersExclusions `yaml:"exclusions,omitempty"`
}

// FormattersExclusions represents exclusions for formatters (v2.10+).
type FormattersExclusions struct {
	Generated  string   `yaml:"generated,omitempty"`
	Paths      []string `yaml:"paths,omitempty"`
	WarnUnused bool     `yaml:"warn-unused,omitempty"`
}

// Issues represents the issues section of golangci-lint configuration (v2 structure).
type Issues struct {
	DiffFromRevision   string   `yaml:"diff-from-revision,omitempty"`
	Include            []string `yaml:"include,omitempty"`
	MaxIssuesPerLinter int      `yaml:"max-issues-per-linter,omitempty"`
	MaxSameIssues      int      `yaml:"max-same-issues,omitempty"`
	New                bool     `yaml:"new,omitempty"`
}

// ExcludeRule represents an exclusion rule.
type ExcludeRule struct {
	Path    string   `yaml:"path,omitempty"`
	Text    string   `yaml:"text,omitempty"`
	Linters []string `yaml:"linters,omitempty"`
}
