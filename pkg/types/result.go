package types

// Result represents a value or an error, inspired by Rust's Result type.
type Result[T any] struct {
	value T
	err   error
}

// Ok creates a successful Result containing value.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err creates a failed Result containing err.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// Get returns the value and error.
// If the result is Ok, error is nil. If Err, value is zero-valued.
func (r Result[T]) Get() (T, error) {
	if r.err != nil {
		var zero T

		return zero, r.err
	}

	return r.value, nil
}

// IsOk returns true if the result contains a value.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// IsError returns true if the result contains an error.
func (r Result[T]) IsError() bool {
	return r.err != nil
}

// MustGet returns the value or panics if the result is an error.
func (r Result[T]) MustGet() T {
	if r.err != nil {
		panic(r.err)
	}

	return r.value
}

// Unwrap returns the value or panics.
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(r.err)
	}

	return r.value
}

// --- Domain-Specific Result Type Aliases ---

// ConfigResult is a Result type for Config operations.
type ConfigResult = Result[*Config]

// AnalysisResult is a Result type for ConfigAnalysis operations.
type AnalysisResult = Result[*ConfigAnalysis]

// MigrationResultType is a Result type for MigrationResult operations.
type MigrationResultType = Result[*MigrationResult]

// StringResult is a Result type for string operations.
type StringResult = Result[string]

// --- Helper Functions ---

// OkConfig wraps a Config in an Ok result.
func OkConfig(cfg *Config) ConfigResult {
	return Ok(cfg)
}

// ErrConfig creates an Err result for Config operations.
func ErrConfig(err error) ConfigResult {
	return Err[*Config](err)
}

// OkAnalysis wraps a ConfigAnalysis in an Ok result.
func OkAnalysis(analysis *ConfigAnalysis) AnalysisResult {
	return Ok(analysis)
}

// ErrAnalysis creates an Err result for analysis operations.
func ErrAnalysis(err error) AnalysisResult {
	return Err[*ConfigAnalysis](err)
}

// OkMigration wraps a MigrationResult in an Ok result.
func OkMigration(result *MigrationResult) MigrationResultType {
	return Ok(result)
}

// ErrMigration creates an Err result for migration operations.
func ErrMigration(err error) MigrationResultType {
	return Err[*MigrationResult](err)
}

// OkString wraps a string in an Ok result.
func OkString(value string) StringResult {
	return Ok(value)
}

// ErrString creates an Err result for string operations.
func ErrString(err error) StringResult {
	return Err[string](err)
}
