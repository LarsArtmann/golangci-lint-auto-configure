// Package types contains core type definitions and result types for railway-oriented programming.
package types

import (
	"github.com/samber/mo"
)

// --- Result Types for Railway-Oriented Programming ---

// ConfigResult is a Result type for Config operations.
type ConfigResult = mo.Result[*Config]

// AnalysisResult is a Result type for ConfigAnalysis operations.
type AnalysisResult = mo.Result[*ConfigAnalysis]

// MigrationResultType is a Result type for MigrationResult operations.
type MigrationResultType = mo.Result[*MigrationResult]

// StringResult is a Result type for string operations.
type StringResult = mo.Result[string]

// --- Helper Functions ---

// OkConfig wraps a Config in an Ok result.
func OkConfig(cfg *Config) ConfigResult {
	return mo.Ok(cfg)
}

// ErrConfig creates an Err result for Config operations.
func ErrConfig(err error) ConfigResult {
	return mo.Err[*Config](err)
}

// OkAnalysis wraps a ConfigAnalysis in an Ok result.
func OkAnalysis(analysis *ConfigAnalysis) AnalysisResult {
	return mo.Ok(analysis)
}

// ErrAnalysis creates an Err result for analysis operations.
func ErrAnalysis(err error) AnalysisResult {
	return mo.Err[*ConfigAnalysis](err)
}

// OkMigration wraps a MigrationResult in an Ok result.
func OkMigration(result *MigrationResult) MigrationResultType {
	return mo.Ok(result)
}

// ErrMigration creates an Err result for migration operations.
func ErrMigration(err error) MigrationResultType {
	return mo.Err[*MigrationResult](err)
}

// OkString wraps a string in an Ok result.
func OkString(value string) StringResult {
	return mo.Ok(value)
}

// ErrString creates an Err result for string operations.
func ErrString(err error) StringResult {
	return mo.Err[string](err)
}
