package errors

import (
	"errors"
	"fmt"
)

// ConfigError represents a configuration-related error.
type ConfigError struct {
	Message string
	Path    string
	Cause   error
}

func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s (path: %s): %v", e.Message, e.Path, e.Cause)
	}

	return fmt.Sprintf("%s (path: %s)", e.Message, e.Path)
}

// Unwrap returns the underlying error for error chaining.
func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// NewConfigError creates a new configuration error.
func NewConfigError(msg, path string, err error) *ConfigError {
	return &ConfigError{
		Message: msg,
		Path:    path,
		Cause:   err,
	}
}

// AnalysisError represents an analysis-related error.
type AnalysisError struct {
	Message string
	File    string
	Cause   error
}

func (e *AnalysisError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s (file: %s): %v", e.Message, e.File, e.Cause)
	}

	return fmt.Sprintf("%s (file: %s)", e.Message, e.File)
}

// Unwrap returns the underlying error for error chaining.
func (e *AnalysisError) Unwrap() error {
	return e.Cause
}

// NewAnalysisError creates a new analysis error.
func NewAnalysisError(msg, file string, err error) *AnalysisError {
	return &AnalysisError{
		Message: msg,
		File:    file,
		Cause:   err,
	}
}

// ReportError represents a report generation error.
type ReportError struct {
	Message string
	Path    string
	Cause   error
}

func (e *ReportError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s (path: %s): %v", e.Message, e.Path, e.Cause)
	}

	return fmt.Sprintf("%s (path: %s)", e.Message, e.Path)
}

// Unwrap returns the underlying error for error chaining.
func (e *ReportError) Unwrap() error {
	return e.Cause
}

// NewReportError creates a new report error.
func NewReportError(msg, path string, err error) *ReportError {
	return &ReportError{
		Message: msg,
		Path:    path,
		Cause:   err,
	}
}

// --- Error Type Checking Helpers ---

// IsConfigError checks if an error is a ConfigError.
func IsConfigError(err error) bool {
	var cfgErr *ConfigError
	return errors.As(err, &cfgErr)
}

// IsAnalysisError checks if an error is an AnalysisError.
func IsAnalysisError(err error) bool {
	var analysisErr *AnalysisError
	return errors.As(err, &analysisErr)
}

// IsReportError checks if an error is a ReportError.
func IsReportError(err error) bool {
	var reportErr *ReportError
	return errors.As(err, &reportErr)
}
