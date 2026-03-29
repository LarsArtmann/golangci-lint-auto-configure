// Package utils provides shared utility functions for the golangci-lint-auto-configure tool.
package utils

import (
	"context"
	"fmt"
	"time"
)

// Config holds retry configuration parameters.
type Config struct {
	MaxRetries     int
	InitialBackoff time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxRetries:     3,
		InitialBackoff: 500 * time.Millisecond,
	}
}

// ShouldRetry is a function that determines if an error should trigger a retry.
type ShouldRetry func(error, string) bool

// Operation is the function to execute with retry logic.
type Operation func() ([]byte, error)

// WithRetry executes the given operation with retry logic.
// It retries on errors where shouldRetry returns true, up to MaxRetries times
// with exponential backoff.
func WithRetry(
	ctx context.Context,
	config Config,
	name string,
	shouldRetry ShouldRetry,
	op Operation,
) ([]byte, error) {
	var lastErr error

	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		output, err := op()
		if err == nil {
			return output, nil
		}

		lastErr = err

		// Check if this error should trigger a retry
		if shouldRetry(err, string(output)) && attempt < config.MaxRetries {
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, fmt.Errorf(
					"%s retry interrupted (lastErr=%w): %w",
					name,
					lastErr,
					ctx.Err(),
				)
			}

			backoff *= 2 // Exponential backoff

			continue
		}

		// Either not a retryable error or out of retries
		if attempt >= config.MaxRetries {
			return output, fmt.Errorf("%s failed after %d retries: %w", name, config.MaxRetries, lastErr)
		}

		return output, err
	}

	return nil, fmt.Errorf("%s failed after %d retries: %w", name, config.MaxRetries, lastErr)
}

// IsContextCanceled returns true if the error is context cancellation.
func IsContextCanceled(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
