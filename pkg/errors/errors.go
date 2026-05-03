package apperrors

import (
	stderrors "errors"
	"fmt"
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
)

// DomainError represents a domain-specific error with context about what went wrong
// and where. Each domain (config, analysis, report, migration) has its own type alias
// for type-safe error checking via errors.As.
type DomainError struct {
	Message string
	Path    string
	Cause   error
	domain  string
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s (%s: %s): %v", e.Message, e.domain, e.Path, e.Cause)
	}

	return fmt.Sprintf("%s (%s: %s)", e.Message, e.domain, e.Path)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// ConfigError is a DomainError in the config domain.
type ConfigError = DomainError

// AnalysisError is a DomainError in the analysis domain.
type AnalysisError = DomainError

// ReportError is a DomainError in the report domain.
type ReportError = DomainError

// MigrationError is a DomainError in the migration domain.
type MigrationError = DomainError

// NewConfigError creates a new configuration error.
func NewConfigError(msg, path string, err error) *ConfigError {
	return &DomainError{Message: msg, Path: path, Cause: err, domain: "path"}
}

// NewAnalysisError creates a new analysis error.
func NewAnalysisError(msg, file string, err error) *AnalysisError {
	return &DomainError{Message: msg, Path: file, Cause: err, domain: "file"}
}

// NewReportError creates a new report error.
func NewReportError(msg, path string, err error) *ReportError {
	return &DomainError{Message: msg, Path: path, Cause: err, domain: "path"}
}

// NewMigrationError creates a new migration error.
func NewMigrationError(msg, config string, err error) *MigrationError {
	return &DomainError{Message: msg, Path: config, Cause: err, domain: "config"}
}

// IsConfigError checks if an error is a ConfigError.
func IsConfigError(err error) bool {
	var cfgErr *ConfigError

	return stderrors.As(err, &cfgErr) && cfgErr.domain == "path"
}

// IsAnalysisError checks if an error is an AnalysisError.
func IsAnalysisError(err error) bool {
	var analysisErr *AnalysisError

	return stderrors.As(err, &analysisErr) && analysisErr.domain == "file"
}

// IsReportError checks if an error is a ReportError.
func IsReportError(err error) bool {
	var reportErr *ReportError

	return stderrors.As(err, &reportErr) && reportErr.domain == "path"
}

// IsMigrationError checks if an error is a MigrationError.
func IsMigrationError(err error) bool {
	var migrationErr *MigrationError

	return stderrors.As(err, &migrationErr) && migrationErr.domain == "config"
}
