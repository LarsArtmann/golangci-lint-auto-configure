package types

import (
	"fmt"
)

// LinterPriority represents the priority level for a linter
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

// LinterInfo contains information about a golangci-lint linter
type LinterInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Groups      []string `json:"groups,omitempty"`
	Fast        bool     `json:"fast,omitempty"`
	AutoFix     bool     `json:"autoFix,omitempty"`
	Deprecated  bool     `json:"deprecated"`
	Since       string   `json:"since"`
	OriginalURL string   `json:"originalURL"`
}

// LinterRecommendation represents a linter with its priority and reason
type LinterRecommendation struct {
	Name     LinterName     `json:"name"`
	Priority LinterPriority `json:"priority"`
	Reason   string         `json:"reason"`
}

// LinterName is a strongly-typed linter name to prevent typos
type LinterName string

func (ln LinterName) String() string {
	return string(ln)
}

// FormatterInfo contains information about a golangci-lint formatter
type FormatterInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AutoFix     bool   `json:"autoFix,omitempty"`
}

// ConfigAnalysis represents the analysis results of a golangci-lint configuration
type ConfigAnalysis struct {
	ConfigPath        string                    `json:"config_path"`
	EnabledLinters    []LinterInfo              `json:"enabled_linters"`
	DisabledLinters   []LinterInfo              `json:"disabled_linters"`
	EnabledFormatters  []FormatterInfo           `json:"enabled_formatters"`
	DisabledFormatters []FormatterInfo           `json:"disabled_formatters"`
	Recommendations   []LinterRecommendation    `json:"recommendations"`
	CriticalCount     int                       `json:"critical_count"`
	HighValueCount    int                       `json:"high_value_count"`
	MediumValueCount  int                       `json:"medium_value_count"`
	OptionalCount     int                       `json:"optional_count"`
}

// MigrationResult represents the result of a configuration migration
type MigrationResult struct {
	Success      bool   `json:"success"`
	FixesApplied int    `json:"fixes_applied"`
	Message      string `json:"message"`
	BackupPath   string `json:"backup_path,omitempty"`
}

// ValidationError represents a configuration validation error
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

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}
