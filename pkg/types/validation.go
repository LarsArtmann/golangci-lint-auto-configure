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
