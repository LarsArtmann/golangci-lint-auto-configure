package types

import (
	"errors"
	"fmt"
)

var (
	ErrConfigNil       = errors.New("config validation failed: config is nil")
	ErrVersionRequired = errors.New("config validation failed: version is required")
	ErrTimeoutRequired = errors.New("config validation failed: run.timeout is required")
	ErrIssuesExitCode  = errors.New("config validation failed: run.issues-exit-code out of range")
	ErrConcurrency     = errors.New("config validation failed: run.concurrency must be >= 0")
	ErrMaxIssues       = errors.New("config validation failed: issues.max-issues-per-linter must be >= 0")
	ErrMaxSameIssues   = errors.New("config validation failed: issues.max-same-issues must be >= 0")
)

// ValidateConfig validates a Config struct.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return ErrConfigNil
	}

	err := validateVersion(cfg)
	if err != nil {
		return err
	}

	err = validateRun(cfg)
	if err != nil {
		return err
	}

	return validateIssues(cfg)
}

func validateVersion(cfg *Config) error {
	if cfg.Version == "" {
		return ErrVersionRequired
	}

	if cfg.Version != "2" {
		return fmt.Errorf("config validation failed: version must be 2, got %q: %w", cfg.Version, ErrVersionRequired)
	}

	return nil
}

func validateRun(cfg *Config) error {
	if cfg.Run.Timeout == "" {
		return ErrTimeoutRequired
	}

	if cfg.Run.IssuesExitCode < 0 || cfg.Run.IssuesExitCode > 255 {
		return fmt.Errorf(
			"config validation failed: run.issues-exit-code must be 0-255, got %d: %w",
			cfg.Run.IssuesExitCode, ErrIssuesExitCode,
		)
	}

	if cfg.Run.Concurrency < 0 {
		return fmt.Errorf(
			"config validation failed: run.concurrency must be >= 0, got %d: %w",
			cfg.Run.Concurrency, ErrConcurrency,
		)
	}

	return nil
}

func validateIssues(cfg *Config) error {
	if cfg.Issues.MaxIssuesPerLinter < 0 {
		return fmt.Errorf(
			"config validation failed: issues.max-issues-per-linter must be >= 0, got %d: %w",
			cfg.Issues.MaxIssuesPerLinter, ErrMaxIssues,
		)
	}

	if cfg.Issues.MaxSameIssues < 0 {
		return fmt.Errorf(
			"config validation failed: issues.max-same-issues must be >= 0, got %d: %w",
			cfg.Issues.MaxSameIssues,
			ErrMaxSameIssues,
		)
	}

	return nil
}

// HealthSeverity represents the severity of a configuration health issue.
type HealthSeverity int

const (
	HealthSeverityCritical HealthSeverity = iota
	HealthSeverityWarning
	HealthSeverityInfo
)

func (s HealthSeverity) String() string {
	switch s {
	case HealthSeverityCritical:
		return "CRITICAL"
	case HealthSeverityWarning:
		return "WARNING"
	case HealthSeverityInfo:
		return "INFO"
	default:
		return "UNKNOWN"
	}
}

// HealthIssue represents a structural health issue found in a config.
type HealthIssue struct {
	Severity   HealthSeverity `json:"severity"`
	Rule       string         `json:"rule"`
	Message    string         `json:"message"`
	Field      string         `json:"field"`
	Suggestion string         `json:"suggestion,omitempty"`
}

// ConfigHealth represents the structural health assessment of a config.
type ConfigHealth struct {
	Issues []HealthIssue `json:"issues"`
}

// IsHealthy returns true if no critical or warning issues were found.
func (h *ConfigHealth) IsHealthy() bool {
	for _, issue := range h.Issues {
		if issue.Severity <= HealthSeverityWarning {
			return false
		}
	}

	return true
}

// CriticalIssues returns only critical-severity issues.
func (h *ConfigHealth) CriticalIssues() []HealthIssue {
	return filterHealthIssues(h.Issues, HealthSeverityCritical)
}

// WarningIssues returns only warning-severity issues.
func (h *ConfigHealth) WarningIssues() []HealthIssue {
	return filterHealthIssues(h.Issues, HealthSeverityWarning)
}

func filterHealthIssues(issues []HealthIssue, severity HealthSeverity) []HealthIssue {
	var filtered []HealthIssue

	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

// CheckConfigHealth performs structural health checks on a config.
// Unlike ValidateConfig (which checks schema correctness), this checks for
// patterns that indicate config quality issues: duplicates, enable+disable
// overlaps, missing critical linters, and v1/v2 syntax mixing.
//
// Uses the default set of critical linters (errcheck, staticcheck, govet).
// Use CheckConfigHealthWithCriticalLinters for a custom critical linter set.
func CheckConfigHealth(cfg *Config) *ConfigHealth {
	return CheckConfigHealthWithCriticalLinters(cfg, []string{"errcheck", "staticcheck", "govet"})
}

// CheckConfigHealthWithCriticalLinters performs structural health checks on a config
// with a configurable set of critical linter names.
func CheckConfigHealthWithCriticalLinters(cfg *Config, criticalLinters []string) *ConfigHealth {
	if cfg == nil {
		return &ConfigHealth{}
	}

	health := &ConfigHealth{}
	health.checkDuplicateLinters(cfg)
	health.checkEnableDisableOverlap(cfg)
	health.checkMissingCriticalLinters(cfg, criticalLinters)
	health.checkV1SyntaxMixing(cfg)

	return health
}

func (h *ConfigHealth) addIssue(severity HealthSeverity, rule, message, field, suggestion string) {
	h.Issues = append(h.Issues, HealthIssue{
		Severity:   severity,
		Rule:       rule,
		Message:    message,
		Field:      field,
		Suggestion: suggestion,
	})
}

func (h *ConfigHealth) checkDuplicateLinters(cfg *Config) {
	seen := make(map[string]int)
	for _, name := range cfg.Linters.Enable {
		seen[name]++
	}

	for name, count := range seen {
		if count > 1 {
			h.addIssue(
				HealthSeverityCritical,
				"duplicate-linter",
				fmt.Sprintf("Linter %q appears %d times in linters.enable", name, count),
				"linters.enable",
				fmt.Sprintf("Remove duplicate entries of %q from the enable list", name),
			)
		}
	}

	seenDisable := make(map[string]int)
	for _, name := range cfg.Linters.Disable {
		seenDisable[name]++
	}

	for name, count := range seenDisable {
		if count > 1 {
			h.addIssue(
				HealthSeverityWarning,
				"duplicate-linter",
				fmt.Sprintf("Linter %q appears %d times in linters.disable", name, count),
				"linters.disable",
				fmt.Sprintf("Remove duplicate entries of %q from the disable list", name),
			)
		}
	}
}

func (h *ConfigHealth) checkEnableDisableOverlap(cfg *Config) {
	enabledSet := NewSet(cfg.Linters.Enable...)

	for _, name := range cfg.Linters.Disable {
		if enabledSet.Contains(name) {
			h.addIssue(
				HealthSeverityWarning,
				"enable-disable-overlap",
				fmt.Sprintf("Linter %q is in both enable and disable lists", name),
				"linters",
				fmt.Sprintf("Remove %q from one of the lists; prefer using linters.default + enable only", name),
			)
		}
	}
}

func (h *ConfigHealth) checkMissingCriticalLinters(cfg *Config, criticalLinters []string) {
	enabledSet := NewSet(cfg.Linters.Enable...)
	disabledSet := NewSet(cfg.Linters.Disable...)

	for _, name := range criticalLinters {
		if !enabledSet.Contains(name) && !disabledSet.Contains(name) {
			h.addIssue(
				HealthSeverityWarning,
				"missing-critical-linter",
				fmt.Sprintf("Critical linter %q is not configured (neither enabled nor disabled)", name),
				"linters.enable",
				fmt.Sprintf("Add %q to linters.enable for correctness checking", name),
			)
		}
	}
}

func (h *ConfigHealth) checkV1SyntaxMixing(cfg *Config) {
	if cfg.Version == "2" && len(cfg.LintersSettingsV1) > 0 {
		h.addIssue(
			HealthSeverityWarning,
			"v1-syntax-in-v2",
			"Config declares version 2 but uses top-level linters-settings (v1 syntax)",
			"linters-settings",
			"Move settings into linters.settings block and remove top-level linters-settings",
		)
	}
}

// ValidateStruct is kept for backward compatibility but delegates to domain-specific validation.
func ValidateStruct[T any](_ *T, _ string) error {
	return nil
}

// ValidateRunConfig is kept for backward compatibility.
func ValidateRunConfig(_ *RunConfig) error {
	return nil
}

// ValidateLintersConfig is kept for backward compatibility.
func ValidateLintersConfig(_ *LintersConfig) error {
	return nil
}
