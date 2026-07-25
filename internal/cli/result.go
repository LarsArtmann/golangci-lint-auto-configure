package cli

import (
	"errors"
	"fmt"
)

// CommandResult is an optional structured return type for CLI commands.
// It carries a user-facing message and an explicit exit code alongside the
// standard error. Commands that need to communicate success details or
// override the error-classification exit code can return a *CommandResult
// from their RunE handler.
//
// Usage:
//
//	return SuccessResult("configured 15 linters")
//	return ErrorResult(err, 1)
//
// Plain error returns continue to work — HandleError handles both.
//
//nolint:errname // intentionally not named *Error: represents both success and failure
type CommandResult struct {
	err      error
	message  string
	exitCode int
}

// Error implements the error interface. Returns the wrapped error's message
// when present, otherwise the success message.
func (r *CommandResult) Error() string {
	if r.err != nil {
		return r.err.Error()
	}

	return r.message
}

// Unwrap allows errors.Is/As to traverse the cause chain.
func (r *CommandResult) Unwrap() error {
	return r.err
}

// ExitCode returns the explicit exit code (0 for success).
func (r *CommandResult) ExitCode() int {
	return r.exitCode
}

// Message returns the user-facing message.
func (r *CommandResult) Message() string {
	return r.message
}

// IsSuccess returns true when the result represents a successful command.
func (r *CommandResult) IsSuccess() bool {
	return r.err == nil
}

// SuccessResult creates a CommandResult for a successful command with a
// user-facing message.
func SuccessResult(message string) *CommandResult {
	return &CommandResult{
		message:  message,
		exitCode: 0,
	}
}

// SuccessResultf is the formatted variant of SuccessResult.
func SuccessResultf(format string, args ...any) *CommandResult {
	return SuccessResult(fmt.Sprintf(format, args...))
}

// ErrorResult creates a CommandResult for a failed command with an explicit
// exit code. The error is preserved for classification and cause-chain
// inspection.
func ErrorResult(err error, exitCode int) *CommandResult {
	return &CommandResult{
		err:      err,
		exitCode: exitCode,
	}
}

// extractCommandResult attempts to find a *CommandResult in the error chain.
// Returns nil if the error is not a *CommandResult.
func extractCommandResult(err error) *CommandResult {
	if err == nil {
		return nil
	}

	if result, ok := errors.AsType[*CommandResult](err); ok {
		return result
	}

	return nil
}
