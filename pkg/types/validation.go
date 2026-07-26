package types

import (
	"errors"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
)

const (
	// ConfigVersionV2 is the golangci-lint v2 config schema version.
	ConfigVersionV2 Version = "2"

	RuleDuplicateLinter       = "duplicate-linter"
	RuleEnableDisableOverlap  = "enable-disable-overlap"
	RuleMissingCriticalLinter = "missing-critical-linter"
	RuleV1SyntaxInV2          = "v1-syntax-in-v2"
)

var (
	ErrConfigNil       = errors.New("config validation failed: config is nil")
	ErrVersionRequired = errors.New("config validation failed: version is required")
	ErrVersionInvalid  = errors.New("config validation failed: version is invalid")
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

	if cfg.Version != ConfigVersionV2 {
		return errorfamily.WrapRejectionf(ErrVersionInvalid, "config.validation.version",
			"version must be %s, got %q", ConfigVersionV2, cfg.Version)
	}

	return nil
}

func validateRun(cfg *Config) error {
	if cfg.Run.Timeout == "" {
		return ErrTimeoutRequired
	}

	if cfg.Run.IssuesExitCode < 0 || cfg.Run.IssuesExitCode > 255 {
		return errorfamily.WrapRejectionf(ErrIssuesExitCode, "config.validation.issues_exit_code",
			"run.issues-exit-code must be 0-255, got %d", cfg.Run.IssuesExitCode)
	}

	if cfg.Run.Concurrency < 0 {
		return errorfamily.WrapRejectionf(ErrConcurrency, "config.validation.concurrency",
			"run.concurrency must be >= 0, got %d", cfg.Run.Concurrency)
	}

	return nil
}

func validateIssues(cfg *Config) error {
	if cfg.Issues.MaxIssuesPerLinter < 0 {
		return errorfamily.WrapRejectionf(ErrMaxIssues, "config.validation.max_issues_per_linter",
			"issues.max-issues-per-linter must be >= 0, got %d", cfg.Issues.MaxIssuesPerLinter)
	}

	if cfg.Issues.MaxSameIssues < 0 {
		return errorfamily.WrapRejectionf(ErrMaxSameIssues, "config.validation.max_same_issues",
			"issues.max-same-issues must be >= 0, got %d", cfg.Issues.MaxSameIssues)
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
	Severity   HealthSeverity
	Rule       string
	Message    string
	Field      string
	Suggestion string `json:",omitempty"`
	Line       int    `json:",omitempty"`
}

// ConfigHealth represents the structural health assessment of a config.
type ConfigHealth struct {
	Issues []HealthIssue
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

// IssuesByRule returns all issues matching the given rule name.
func (h *ConfigHealth) IssuesByRule(rule string) []HealthIssue {
	var filtered []HealthIssue

	for _, issue := range h.Issues {
		if issue.Rule == rule {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

// HasRule returns true if any issue matches the given rule name.
func (h *ConfigHealth) HasRule(rule string) bool {
	for _, issue := range h.Issues {
		if issue.Rule == rule {
			return true
		}
	}

	return false
}

// CountBySeverity returns the number of issues with the given severity.
func (h *ConfigHealth) CountBySeverity(severity HealthSeverity) int {
	count := 0

	for _, issue := range h.Issues {
		if issue.Severity == severity {
			count++
		}
	}

	return count
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
				RuleDuplicateLinter,
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
				RuleDuplicateLinter,
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
				RuleEnableDisableOverlap,
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
				RuleMissingCriticalLinter,
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
			RuleV1SyntaxInV2,
			"Config declares version 2 but uses top-level linters-settings (v1 syntax)",
			"linters-settings",
			"Move settings into linters.settings block and remove top-level linters-settings",
		)
	}
}
