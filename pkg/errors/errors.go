package apperrors

import (
	stderrors "errors"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
)

// Static sentinel errors for use with stderrors.Is.
var (
	ErrNotGitRepository       = stderrors.New("not a git repository (no .git directory found)")
	ErrNotInGitWorkingTree    = stderrors.New("not inside git working tree")
	ErrHookAlreadyExists      = stderrors.New("hook already exists")
	ErrUnknownPreset          = stderrors.New("unknown preset")
	ErrInvalidActivityContext = stderrors.New("invalid activity context type")
	ErrVersionParse           = stderrors.New("could not parse version from output")
	ErrInvalidVersionFormat   = stderrors.New("invalid version format")
	ErrVersionTooOld          = stderrors.New("version is too old")
	ErrConfigValidationFailed = stderrors.New("configuration validation failed")
	ErrChangesNeeded          = stderrors.New("configuration changes needed")
	ErrNoConfigFiles          = stderrors.New("no config files to merge")
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
	var cfgErr *ConfigError

	return stderrors.As(err, &cfgErr)
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
	var analysisErr *AnalysisError

	return stderrors.As(err, &analysisErr)
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
	var reportErr *ReportError

	return stderrors.As(err, &reportErr)
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
	var migrationErr *MigrationError

	return stderrors.As(err, &migrationErr)
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
