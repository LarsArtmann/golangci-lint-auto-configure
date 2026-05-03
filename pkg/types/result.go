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

// --- Domain-Specific Result Type Aliases ---

// ConfigResult is a Result type for Config operations.
type ConfigResult = Result[*Config]

// AnalysisResult is a Result type for ConfigAnalysis operations.
type AnalysisResult = Result[*ConfigAnalysis]

// MigrationResultType is a Result type for MigrationResult operations.
type MigrationResultType = Result[*MigrationResult]

// StringResult is a Result type for string operations.
type StringResult = Result[string]
