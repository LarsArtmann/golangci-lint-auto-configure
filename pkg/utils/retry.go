package utils

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Config struct {
	MaxRetries     int
	InitialBackoff time.Duration
}

const (
	DefaultMaxRetries     = 3
	DefaultInitialBackoff = 500 // milliseconds
)

func DefaultConfig() Config {
	return Config{
		MaxRetries:     DefaultMaxRetries,
		InitialBackoff: DefaultInitialBackoff * time.Millisecond,
	}
}

type ShouldRetry func(error, string) bool

type Operation func() ([]byte, error)

//nolint:funlen // Retry logic with context cancellation and exponential backoff is inherently complex; extraction breaks error wrapping semantics.
func WithRetry(
	ctx context.Context,
	config Config,
	name string,
	shouldRetry ShouldRetry,
	executeOperation Operation,
) ([]byte, error) {
	var lastErr error

	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		output, err := executeOperation()
		if err == nil {
			return output, nil
		}

		lastErr = err

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

			backoff *= 2

			continue
		}

		if attempt >= config.MaxRetries {
			return output, fmt.Errorf("%s failed after %d retries: %w", name, config.MaxRetries, lastErr)
		}

		return output, err
	}

	return nil, fmt.Errorf("%s failed after %d retries: %w", name, config.MaxRetries, lastErr)
}

func IsContextCanceled(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
