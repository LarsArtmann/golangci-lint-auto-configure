package types

import (
	"context"
	"errors"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
)

// ErrInvalidLinterPriority indicates an unrecognized linter priority value.
var ErrInvalidLinterPriority = errors.New("invalid linter priority: must be critical, high, medium, or optional")

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

// ParseLinterPriority parses a case-insensitive priority string.
// Returns an error for unrecognized values.
func ParseLinterPriority(input string) (LinterPriority, error) {
	switch input {
	case "critical":
		return LinterPriorityCritical, nil
	case "high":
		return LinterPriorityHigh, nil
	case "medium":
		return LinterPriorityMedium, nil
	case "optional":
		return LinterPriorityOptional, nil
	default:
		return LinterPriorityOptional, errorfamily.WrapRejectionf(
			ErrInvalidLinterPriority, "linter.priority.invalid", "%q", input,
		)
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

// GoExperiment represents a Go runtime experiment that exposes new standard library packages.
type GoExperiment struct {
	Tag         string
	Package     string
	Description string
}

// LinterInfo contains information about a golangci-lint linter.
type LinterInfo struct {
	Name        LinterName
	Description string
	Groups      []string `json:",omitempty"`
	Fast        bool     `json:",omitzero"`
	AutoFix     bool     `json:",omitzero"`
	Deprecated  bool
	Since       string
	OriginalURL string
}

// LinterRecommendation represents a linter with its priority and reason.
type LinterRecommendation struct {
	Name     LinterName
	Priority LinterPriority
	Reason   string
}

// LinterName is a strongly-typed linter name to prevent typos.
type LinterName string

func (ln LinterName) String() string {
	return string(ln)
}

// LinterReplacement represents a replacement for a deprecated linter.
type LinterReplacement struct {
	Replacement LinterName
	Reason      string
	MinVersion  string `json:",omitempty"` // Minimum golangci-lint version where the replacement exists
}

// LinterToFormatter represents a linter that is superseded by a formatter.
type LinterToFormatter struct {
	Formatter FormatterName
	Reason    string
}

// FormatterName is a strongly-typed formatter name to prevent typos.
type FormatterName string

func (fn FormatterName) String() string {
	return string(fn)
}

// FormatterInfo contains information about a golangci-lint formatter.
type FormatterInfo struct {
	Name        FormatterName
	Description string
	AutoFix     bool `json:",omitzero"`
}

// FormatterRecommendation represents a formatter with its priority and reason.
type FormatterRecommendation struct {
	Name     FormatterName
	Priority FormatterPriority
	Reason   string
}

// ConfigAnalysis represents the analysis results of a golangci-lint configuration.
type ConfigAnalysis struct {
	ConfigPath               string
	EnabledLinters           []LinterInfo
	DisabledLinters          []LinterInfo
	EnabledFormatters        []FormatterInfo
	DisabledFormatters       []FormatterInfo
	LinterRecommendations    []LinterRecommendation
	FormatterRecommendations []FormatterRecommendation
	DeprecatedLinters        []LinterInfo
	CriticalCount            int
	HighValueCount           int
	MediumValueCount         int
	OptionalCount            int
	DeprecatedCount          int
}

// TotalRecommendations returns the total number of linter + formatter recommendations.
func (a *ConfigAnalysis) TotalRecommendations() int {
	return len(a.LinterRecommendations) + len(a.FormatterRecommendations)
}

// EnabledLinterNames returns just the names of enabled linters.
func (a *ConfigAnalysis) EnabledLinterNames() []string {
	names := make([]string, 0, len(a.EnabledLinters))
	for _, l := range a.EnabledLinters {
		names = append(names, string(l.Name))
	}

	return names
}

// MigrationResult represents the result of a configuration migration.
type MigrationResult struct {
	FixesApplied int
	Message      string
	NextSteps    []string `json:",omitempty"`
	Error        error    `json:"-"` // Error is not serialized to JSON
	DryRun       bool
}

// IsSuccess returns true if the migration was successful.
func (m *MigrationResult) IsSuccess() bool {
	return m.Error == nil
}

// IsFailure returns true if the migration failed.
func (m *MigrationResult) IsFailure() bool {
	return m.Error != nil
}

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
	Line    int `json:",omitempty"`
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("validation error at line %d: %s (field: %s)", e.Line, e.Message, e.Field)
	}

	return fmt.Sprintf("validation error: %s (field: %s)", e.Message, e.Field)
}

// ToHealthIssue converts a ValidationError to a HealthIssue with Critical severity.
// This bridges the two representations for unified display and reporting.
func (e ValidationError) ToHealthIssue() HealthIssue {
	return HealthIssue{
		Severity: HealthSeverityCritical,
		Rule:     "validation",
		Message:  e.Message,
		Field:    e.Field,
		Line:     e.Line,
	}
}

// ValidationResult represents the result of configuration validation.
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError `json:",omitempty"`
}

// ToHealthIssues converts all ValidationErrors to HealthIssues for unified
// reporting alongside structural health checks.
func (r ValidationResult) ToHealthIssues() []HealthIssue {
	issues := make([]HealthIssue, 0, len(r.Errors))
	for _, e := range r.Errors {
		issues = append(issues, e.ToHealthIssue())
	}

	return issues
}

// --- Interfaces for Testability ---

// ConfigLoader defines the composite interface for loading, saving, discovering,
// validating, inspecting, and creating golangci-lint configurations.
type ConfigLoader interface {
	LoadConfig(path string) (*Config, error)
	FindConfigFile(startDir string) (string, error)
	FindOrGetDefaultConfigPath(startDir string) string
	SaveConfig(config *Config, path string) error
	ValidateConfig(config *Config) []error
	GetLintersEnabled(config *Config) []string
	GetLintersDisabled(config *Config) []string
	CreateDefaultConfig(ctx context.Context) *Config
}

// LinterAnalyzer defines the interface for analyzing golangci-lint configurations.
type LinterAnalyzer interface {
	AnalyzeConfig(ctx context.Context, configPath string) (*ConfigAnalysis, error)
	FindBinary(ctx context.Context) error
	CheckVersion(ctx context.Context) error
	GetDetectedVersion() string
	GetSummary(analysis *ConfigAnalysis) string
	GetLintersByPriority(recommendations []LinterRecommendation, priority LinterPriority) []LinterRecommendation
}
