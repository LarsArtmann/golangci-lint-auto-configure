package errors

import "fmt"

// ConfigError represents a configuration-related error
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

// NewConfigError creates a new configuration error
func NewConfigError(msg, path string, err error) *ConfigError {
	return &ConfigError{
		Message: msg,
		Path:    path,
		Cause:   err,
	}
}

// AnalysisError represents an analysis-related error
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

// NewAnalysisError creates a new analysis error
func NewAnalysisError(msg, file string, err error) *AnalysisError {
	return &AnalysisError{
		Message: msg,
		File:    file,
		Cause:   err,
	}
}

// ReportError represents a report generation error
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

// NewReportError creates a new report error
func NewReportError(msg, path string, err error) *ReportError {
	return &ReportError{
		Message: msg,
		Path:    path,
		Cause:   err,
	}
}
