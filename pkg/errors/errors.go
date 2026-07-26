package apperrors

import (
	"errors"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
)

// Static sentinel errors for use with errors.Is.
var (
	ErrNotGitRepository       = errors.New("not a git repository (no .git directory found)")
	ErrNotInGitWorkingTree    = errors.New("not inside git working tree")
	ErrHookAlreadyExists      = errors.New("hook already exists")
	ErrUnknownPreset          = errors.New("unknown preset")
	ErrInvalidActivityContext = errors.New("invalid activity context type")
	ErrVersionParse           = errors.New("could not parse version from output")
	ErrInvalidVersionFormat   = errors.New("invalid version format")
	ErrVersionTooOld          = errors.New("version is too old")
	ErrConfigValidationFailed = errors.New("configuration validation failed")
	ErrChangesNeeded          = errors.New("configuration changes needed")
	ErrNoConfigFiles          = errors.New("no config files to merge")
)

// domainError provides shared Error() and Unwrap() for domain-specific error types.
type domainError struct {
	Message string
	Source  string
	Cause   error
}

func (e *domainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s (%s): %v", e.Message, e.Source, e.Cause)
	}

	return fmt.Sprintf("%s (%s)", e.Message, e.Source)
}

func (e *domainError) Unwrap() error {
	return e.Cause
}

// ConfigError represents an error during configuration loading, saving, or validation.
type ConfigError struct {
	Message string
	Path    string
	Cause   error
}

func (e *ConfigError) Error() string {
	return (&domainError{Message: e.Message, Source: "config: " + e.Path, Cause: e.Cause}).Error()
}

func (e *ConfigError) Unwrap() error { return e.Cause }

// NewConfigError creates a new configuration error.
func NewConfigError(msg, path string, err error) *ConfigError {
	return &ConfigError{Message: msg, Path: path, Cause: err}
}

// IsConfigError checks if an error is a ConfigError.
func IsConfigError(err error) bool {
	_, ok := errors.AsType[*ConfigError](err) //nolint:erraudit // AsType returns (value, bool); _ is the typed value, not an ignored error

	return ok
}

// AnalysisError represents an error during configuration analysis.
type AnalysisError struct {
	Message string
	File    string
	Cause   error
}

func (e *AnalysisError) Error() string {
	return (&domainError{Message: e.Message, Source: "analysis: " + e.File, Cause: e.Cause}).Error()
}

func (e *AnalysisError) Unwrap() error { return e.Cause }

// NewAnalysisError creates a new analysis error.
func NewAnalysisError(msg, file string, err error) *AnalysisError {
	return &AnalysisError{Message: msg, File: file, Cause: err}
}

// IsAnalysisError checks if an error is an AnalysisError.
func IsAnalysisError(err error) bool {
	_, ok := errors.AsType[*AnalysisError](err) //nolint:erraudit // AsType returns (value, bool); _ is the typed value, not an ignored error

	return ok
}

// ReportError represents an error during report generation.
type ReportError struct {
	Message string
	Path    string
	Cause   error
}

func (e *ReportError) Error() string {
	return (&domainError{Message: e.Message, Source: "report: " + e.Path, Cause: e.Cause}).Error()
}

func (e *ReportError) Unwrap() error { return e.Cause }

// NewReportError creates a new report error.
func NewReportError(msg, path string, err error) *ReportError {
	return &ReportError{Message: msg, Path: path, Cause: err}
}

// IsReportError checks if an error is a ReportError.
func IsReportError(err error) bool {
	_, ok := errors.AsType[*ReportError](err) //nolint:erraudit // AsType returns (value, bool); _ is the typed value, not an ignored error

	return ok
}

// MigrationError represents an error during configuration migration.
type MigrationError struct {
	Message string
	Config  string
	Cause   error
}

func (e *MigrationError) Error() string {
	return (&domainError{Message: e.Message, Source: "migration: " + e.Config, Cause: e.Cause}).Error()
}

func (e *MigrationError) Unwrap() error { return e.Cause }

// NewMigrationError creates a new migration error.
func NewMigrationError(msg, config string, err error) *MigrationError {
	return &MigrationError{Message: msg, Config: config, Cause: err}
}

// IsMigrationError checks if an error is a MigrationError.
func IsMigrationError(err error) bool {
	_, ok := errors.AsType[*MigrationError](err) //nolint:erraudit // AsType returns (value, bool); _ is the typed value, not an ignored error

	return ok
}

// WrapClassified wraps an error with a code and message, preserving the cause
// chain's behavioral family. This is the right choice when the wrapped error's
// family should be determined by its cause (e.g., exec.ErrNotFound → Infrastructure,
// validation sentinel → Rejection) rather than the wrap site itself.
//
// For errors where the family is always the same regardless of cause, prefer the
// explicit errorfamily.WrapRejection/WrapTransient/etc constructors instead.
//
// Note: Callers that may pass a nil error MUST check for nil before calling this
// function and return nil themselves. Returning WrapClassified(nilErr, ...) through
// an error-typed function triggers the typed-nil interface pitfall (the nil
// *errorfamily.Error becomes a non-nil error interface). The concrete return type
// is intentional so callers that need the classified methods can access them.
func WrapClassified(err error, code, message string) *errorfamily.Error {
	if err == nil {
		return nil
	}

	return errorfamily.Wrap(err, errorfamily.Classify(err), code, message)
}

// WrapClassifiedf is the formatted variant of WrapClassified.
func WrapClassifiedf(err error, code, format string, args ...any) *errorfamily.Error {
	if err == nil {
		return nil
	}

	return errorfamily.Wrap(err, errorfamily.Classify(err), code, fmt.Sprintf(format, args...))
}
