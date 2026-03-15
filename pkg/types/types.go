package types

// TODO: Consider using generics for ConfigResult types to reduce boilerplate
// TODO: Add validation tags for struct fields using a validation library
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

// ConfigLoader defines the interface for loading and saving golangci-lint configurations.
type ConfigLoader interface {
	LoadConfig(path string) (*Config, error)
	SaveConfig(config *Config, path string) error
	FindConfigFile(startDir string) (string, error)
	FindOrGetDefaultConfigPath(startDir string) string
	EnsureGitRepo(startDir string) error
	ValidateConfig(config *Config) []error
	GetLintersEnabled(config *Config) []string
	GetLintersDisabled(config *Config) []string
	CreateDefaultConfig() *Config
	GetAllLinterNames() ([]string, error)
}

// Config represents a golangci-lint configuration file.
type Config struct {
	Version    string           `yaml:"version"`
	Run        RunConfig        `yaml:"run"`
	Output     OutputConfig     `yaml:"output"`
	Linters    LintersConfig    `yaml:"linters"`
	Formatters FormattersConfig `yaml:"formatters,omitempty"`
	Issues     IssuesConfig     `yaml:"issues"`
}

type RunConfig struct {
	Timeout              string   `yaml:"timeout"`
	Go                   string   `yaml:"go"`
	BuildTags            []string `yaml:"build-tags"`
	ModulesDownloadMode  string   `yaml:"modules-download-mode,omitempty"`
	AllowParallelRunners bool     `yaml:"allow-parallel-runners"`
	AllowSerialRunners   bool     `yaml:"allow-serial-runners"`
	IssuesExitCode       int      `yaml:"issues-exit-code,omitempty"`
	Tests                bool     `yaml:"tests,omitempty"`
	Concurrency          int      `yaml:"concurrency,omitempty"`
	RelativePathMode     string   `yaml:"relative-path-mode,omitempty"`
}

type OutputConfig struct {
	Formats    map[string]any `yaml:"formats"`
	PathPrefix string         `yaml:"path-prefix,omitempty"`
	PathMode   string         `yaml:"path-mode,omitempty"`
	SortOrder  []string       `yaml:"sort-order,omitempty"`
	ShowStats  bool           `yaml:"show-stats,omitempty"`
}

type LintersConfig struct {
	Enable     []string                `yaml:"enable,omitempty"`
	Disable    []string                `yaml:"disable,omitempty"`
	Default    string                  `yaml:"default,omitempty"`
	Settings   map[string]any          `yaml:"settings,omitempty"`
	Exclusions LintersExclusionsConfig `yaml:"exclusions,omitempty"`
}

type LintersExclusionsConfig struct {
	Generated   string                `yaml:"generated,omitempty"`
	WarnUnused  bool                  `yaml:"warn-unused,omitempty"`
	Presets     []string              `yaml:"presets,omitempty"`
	Rules       []ExclusionRuleConfig `yaml:"rules,omitempty"`
	Paths       []string              `yaml:"paths,omitempty"`
	PathsExcept []string              `yaml:"paths-except,omitempty"`
}

type ExclusionRuleConfig struct {
	Path       string   `yaml:"path,omitempty"`
	PathExcept string   `yaml:"path-except,omitempty"`
	Text       string   `yaml:"text,omitempty"`
	Source     string   `yaml:"source,omitempty"`
	Linters    []string `yaml:"linters,omitempty"`
}

type IssuesConfig struct {
	MaxIssuesPerLinter int    `yaml:"max-issues-per-linter,omitempty"`
	MaxSameIssues      int    `yaml:"max-same-issues,omitempty"`
	NewFromRev         string `yaml:"new-from-rev,omitempty"`
	NewFromPatch       string `yaml:"new-from-patch,omitempty"`
	New                bool   `yaml:"new,omitempty"`
	NewFromMergeBase   string `yaml:"new-from-merge-base,omitempty"`
	WholeFiles         bool   `yaml:"whole-files,omitempty"`
	Fix                bool   `yaml:"fix,omitempty"`
	UniqByLine         bool   `yaml:"uniq-by-line,omitempty"`
}

type FormattersConfig struct {
	Enable     []string                   `yaml:"enable,omitempty"`
	Disable    []string                   `yaml:"disable,omitempty"`
	Settings   map[string]any             `yaml:"settings,omitempty"`
	Exclusions FormattersExclusionsConfig `yaml:"exclusions,omitempty"`
}

type FormattersExclusionsConfig struct {
	Generated  string   `yaml:"generated,omitempty"`
	WarnUnused bool     `yaml:"warn-unused,omitempty"`
	Paths      []string `yaml:"paths,omitempty"`
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
