package types

// TODO: Consider using generics for ConfigResult types to reduce boilerplate
// TODO: Consider using time.Duration instead of string for timeout fields

import (
	"context"
	"fmt"
)

// LinterPriority represents the priority level for a linter.
type LinterPriority int

const (
	LinterPriorityCritical LinterPriority = iota
	LinterPriorityHigh
	LinterPriorityMedium
	LinterPriorityOptional
)

func (p LinterPriority) String() string {
	switch p {
	case LinterPriorityCritical:
		return "CRITICAL"
	case LinterPriorityHigh:
		return "HIGH"
	case LinterPriorityMedium:
		return "MEDIUM"
	case LinterPriorityOptional:
		return "OPTIONAL"
	default:
		return "UNKNOWN"
	}
}

// FormatterPriority represents the priority level for a formatter.
type FormatterPriority int

const (
	FormatterPriorityHigh FormatterPriority = iota
	FormatterPriorityMedium
	FormatterPriorityLow
)

func (p FormatterPriority) String() string {
	switch p {
	case FormatterPriorityHigh:
		return "HIGH"
	case FormatterPriorityMedium:
		return "MEDIUM"
	case FormatterPriorityLow:
		return "LOW"
	default:
		return "UNKNOWN"
	}
}

// LinterInfo contains information about a golangci-lint linter.
type LinterInfo struct {
	Name        LinterName `json:"name"`
	Description string     `json:"description"`
	Groups      []string   `json:"groups,omitempty"`
	Fast        bool       `json:"fast,omitempty"`
	AutoFix     bool       `json:"autoFix,omitempty"`
	Deprecated  bool       `json:"deprecated"`
	Since       string     `json:"since"`
	OriginalURL string     `json:"originalURL"`
}

// LinterRecommendation represents a linter with its priority and reason.
type LinterRecommendation struct {
	Name     LinterName     `json:"name"`
	Priority LinterPriority `json:"priority"`
	Reason   string         `json:"reason"`
}

// LinterName is a strongly-typed linter name to prevent typos.
type LinterName string

func (ln LinterName) String() string {
	return string(ln)
}

// LinterReplacement represents a replacement for a deprecated linter.
type LinterReplacement struct {
	Replacement LinterName `json:"replacement"`
	Reason      string     `json:"reason"`
}

// FormatterName is a strongly-typed formatter name to prevent typos.
type FormatterName string

func (fn FormatterName) String() string {
	return string(fn)
}

// FormatterInfo contains information about a golangci-lint formatter.
type FormatterInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AutoFix     bool   `json:"autoFix,omitempty"`
}

// FormatterRecommendation represents a formatter with its priority and reason.
type FormatterRecommendation struct {
	Name     FormatterName     `json:"name"`
	Priority FormatterPriority `json:"priority"`
	Reason   string            `json:"reason"`
}

// ConfigAnalysis represents the analysis results of a golangci-lint configuration.
type ConfigAnalysis struct {
	ConfigPath               string                    `json:"config_path"`
	EnabledLinters           []LinterInfo              `json:"enabled_linters"`
	DisabledLinters          []LinterInfo              `json:"disabled_linters"`
	EnabledFormatters        []FormatterInfo           `json:"enabled_formatters"`
	DisabledFormatters       []FormatterInfo           `json:"disabled_formatters"`
	LinterRecommendations    []LinterRecommendation    `json:"linter_recommendations"`
	FormatterRecommendations []FormatterRecommendation `json:"formatter_recommendations"`
	DeprecatedLinters        []LinterInfo              `json:"deprecated_linters"`
	CriticalCount            int                       `json:"critical_count"`
	HighValueCount           int                       `json:"high_value_count"`
	MediumValueCount         int                       `json:"medium_value_count"`
	OptionalCount            int                       `json:"optional_count"`
	DeprecatedCount          int                       `json:"deprecated_count"`
}

// MigrationResult represents the result of a configuration migration.
type MigrationResult struct {
	Success      bool   `json:"success"`
	FixesApplied int    `json:"fixes_applied"`
	Message      string `json:"message"`
}

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("validation error at line %d: %s (field: %s)", e.Line, e.Message, e.Field)
	}

	return fmt.Sprintf("validation error: %s (field: %s)", e.Message, e.Field)
}

// ValidationResult represents the result of configuration validation.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// --- Interfaces for Testability ---

// ConfigReader defines the interface for reading golangci-lint configurations.
type ConfigReader interface {
	LoadConfig(path string) (*Config, error)
	FindConfigFile(startDir string) (string, error)
}

// ConfigWriter defines the interface for writing golangci-lint configurations.
type ConfigWriter interface {
	SaveConfig(config *Config, path string) error
}

// ConfigDiscovery defines the interface for discovering configuration files.
type ConfigDiscovery interface {
	FindConfigFile(startDir string) (string, error)
	FindOrGetDefaultConfigPath(startDir string) string
}

// ConfigValidator defines the interface for validating configurations.
type ConfigValidator interface {
	ValidateConfig(config *Config) []error
}

// ConfigInspector defines the interface for inspecting configuration contents.
type ConfigInspector interface {
	GetLintersEnabled(config *Config) []string
	GetLintersDisabled(config *Config) []string
}

// ConfigCreator defines the interface for creating default configurations.
type ConfigCreator interface {
	CreateDefaultConfig(ctx context.Context) *Config
	GetAllLinterNames(ctx context.Context) ([]string, error)
}

// GitChecker defines the interface for checking git repository status.
type GitChecker interface {
	EnsureGitRepo(ctx context.Context, startDir string) error
}

// ConfigLoader defines the composite interface for loading and saving golangci-lint configurations.
// This interface combines all the smaller, focused interfaces above for backward compatibility.
type ConfigLoader interface {
	ConfigReader
	ConfigWriter
	ConfigDiscovery
	ConfigValidator
	ConfigInspector
	ConfigCreator
	GitChecker
}

// Config represents a golangci-lint configuration file.
type Config struct {
	Version    string           `validate:"required,oneof=2" yaml:"version"              json:"version"              toml:"version"`
	Run        RunConfig        `validate:"required"         yaml:"run"                  json:"run"                  toml:"run"`
	Output     OutputConfig     `                            yaml:"output"               json:"output"               toml:"output"`
	Linters    LintersConfig    `                            yaml:"linters"              json:"linters"              toml:"linters"`
	Formatters FormattersConfig `                            yaml:"formatters,omitempty" json:"formatters" toml:"formatters,omitempty"`
	Issues     IssuesConfig     `                            yaml:"issues"               json:"issues"               toml:"issues"`
}

type RunConfig struct {
	Timeout              string   `yaml:"timeout"                         json:"timeout"                         toml:"timeout"                         validate:"required"`
	Go                   string   `yaml:"go"                              json:"go"                              toml:"go"`
	BuildTags            []string `yaml:"build-tags"                      json:"build-tags"                      toml:"build-tags"`
	ModulesDownloadMode  string   `yaml:"modules-download-mode,omitempty" json:"modules-download-mode,omitempty" toml:"modules-download-mode,omitempty"`
	AllowParallelRunners bool     `yaml:"allow-parallel-runners"          json:"allow-parallel-runners"          toml:"allow-parallel-runners"`
	AllowSerialRunners   bool     `yaml:"allow-serial-runners"            json:"allow-serial-runners"            toml:"allow-serial-runners"`
	IssuesExitCode       int      `yaml:"issues-exit-code,omitempty"      json:"issues-exit-code,omitempty"      toml:"issues-exit-code,omitempty"      validate:"min=0,max=255"`
	Tests                bool     `yaml:"tests,omitempty"                 json:"tests,omitempty"                 toml:"tests,omitempty"`
	Concurrency          int      `yaml:"concurrency,omitempty"           json:"concurrency,omitempty"           toml:"concurrency,omitempty"           validate:"min=0"`
	RelativePathMode     string   `yaml:"relative-path-mode,omitempty"    json:"relative-path-mode,omitempty"    toml:"relative-path-mode,omitempty"`
}

type OutputConfig struct {
	Formats    map[string]any `yaml:"formats"               json:"formats"               toml:"formats"`
	PathPrefix string         `yaml:"path-prefix,omitempty" json:"path-prefix,omitempty" toml:"path-prefix,omitempty"`
	PathMode   string         `yaml:"path-mode,omitempty"   json:"path-mode,omitempty"   toml:"path-mode,omitempty"`
	SortOrder  []string       `yaml:"sort-order,omitempty"  json:"sort-order,omitempty"  toml:"sort-order,omitempty"`
	ShowStats  bool           `yaml:"show-stats,omitempty"  json:"show-stats,omitempty"  toml:"show-stats,omitempty"`
}

type LintersConfig struct {
	Enable     []string                `yaml:"enable,omitempty"     json:"enable,omitempty"     toml:"enable,omitempty"`
	Disable    []string                `yaml:"disable,omitempty"    json:"disable,omitempty"    toml:"disable,omitempty"`
	Default    string                  `yaml:"default,omitempty"    json:"default,omitempty"    toml:"default,omitempty"`
	Settings   map[string]any          `yaml:"settings,omitempty"   json:"settings,omitempty"   toml:"settings,omitempty"`
	Exclusions LintersExclusionsConfig `yaml:"exclusions,omitempty" json:"exclusions" toml:"exclusions,omitempty"`
}

type LintersExclusionsConfig struct {
	Generated   string                `yaml:"generated,omitempty"    json:"generated,omitempty"    toml:"generated,omitempty"`
	WarnUnused  bool                  `yaml:"warn-unused,omitempty"  json:"warn-unused,omitempty"  toml:"warn-unused,omitempty"`
	Presets     []string              `yaml:"presets,omitempty"      json:"presets,omitempty"      toml:"presets,omitempty"`
	Rules       []ExclusionRuleConfig `yaml:"rules,omitempty"        json:"rules,omitempty"        toml:"rules,omitempty"`
	Paths       []string              `yaml:"paths,omitempty"        json:"paths,omitempty"        toml:"paths,omitempty"`
	PathsExcept []string              `yaml:"paths-except,omitempty" json:"paths-except,omitempty" toml:"paths-except,omitempty"`
}

type ExclusionRuleConfig struct {
	Path       string   `yaml:"path,omitempty"        json:"path,omitempty"        toml:"path,omitempty"`
	PathExcept string   `yaml:"path-except,omitempty" json:"path-except,omitempty" toml:"path-except,omitempty"`
	Text       string   `yaml:"text,omitempty"        json:"text,omitempty"        toml:"text,omitempty"`
	Source     string   `yaml:"source,omitempty"      json:"source,omitempty"      toml:"source,omitempty"`
	Linters    []string `yaml:"linters,omitempty"     json:"linters,omitempty"     toml:"linters,omitempty"`
}

type IssuesConfig struct {
	MaxIssuesPerLinter int    `validate:"min=0" yaml:"max-issues-per-linter,omitempty" json:"max-issues-per-linter,omitempty" toml:"max-issues-per-linter,omitempty"`
	MaxSameIssues      int    `validate:"min=0" yaml:"max-same-issues,omitempty"       json:"max-same-issues,omitempty"       toml:"max-same-issues,omitempty"`
	NewFromRev         string `                 yaml:"new-from-rev,omitempty"          json:"new-from-rev,omitempty"          toml:"new-from-rev,omitempty"`
	NewFromPatch       string `                 yaml:"new-from-patch,omitempty"        json:"new-from-patch,omitempty"        toml:"new-from-patch,omitempty"`
	New                bool   `                 yaml:"new,omitempty"                   json:"new,omitempty"                   toml:"new,omitempty"`
	NewFromMergeBase   string `                 yaml:"new-from-merge-base,omitempty"   json:"new-from-merge-base,omitempty"   toml:"new-from-merge-base,omitempty"`
	WholeFiles         bool   `                 yaml:"whole-files,omitempty"           json:"whole-files,omitempty"           toml:"whole-files,omitempty"`
	Fix                bool   `                 yaml:"fix,omitempty"                   json:"fix,omitempty"                   toml:"fix,omitempty"`
	UniqByLine         bool   `                 yaml:"uniq-by-line,omitempty"          json:"uniq-by-line,omitempty"          toml:"uniq-by-line,omitempty"`
}

type FormattersConfig struct {
	Enable     []string                   `yaml:"enable,omitempty"     json:"enable,omitempty"     toml:"enable,omitempty"`
	Disable    []string                   `yaml:"disable,omitempty"    json:"disable,omitempty"    toml:"disable,omitempty"`
	Settings   map[string]any             `yaml:"settings,omitempty"   json:"settings,omitempty"   toml:"settings,omitempty"`
	Exclusions FormattersExclusionsConfig `yaml:"exclusions,omitempty" json:"exclusions" toml:"exclusions,omitempty"`
}

type FormattersExclusionsConfig struct {
	Generated  string   `yaml:"generated,omitempty"   json:"generated,omitempty"   toml:"generated,omitempty"`
	WarnUnused bool     `yaml:"warn-unused,omitempty" json:"warn-unused,omitempty" toml:"warn-unused,omitempty"`
	Paths      []string `yaml:"paths,omitempty"       json:"paths,omitempty"       toml:"paths,omitempty"`
}

// LinterAnalyzer defines the interface for analyzing golangci-lint configurations.
type LinterAnalyzer interface {
	AnalyzeConfig(ctx context.Context, configPath string) (*ConfigAnalysis, error)
	FindBinary(ctx context.Context) error
	CheckVersion(ctx context.Context) error
	FormatRecommendations(analysis *ConfigAnalysis) string
	GetSummary(analysis *ConfigAnalysis) string
	GetLintersByPriority(recommendations []LinterRecommendation, priority LinterPriority) []LinterRecommendation
}

// LinterFixer defines the interface for fixing golangci-lint configurations.
type LinterFixer interface {
	FixConfig(ctx context.Context, configPath string, priority LinterPriority, dryRun bool) (*MigrationResult, error)
}
