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

const retriesExceededFmt = "%s failed after %d retries: %w"

func WithRetry(
	ctx context.Context, config Config, name string,
	shouldRetry ShouldRetry, op Operation,
) ([]byte, error) {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := range config.MaxRetries + 1 {
		output, err := op()
		if err == nil {
			return output, nil
		}

		lastErr = err

		if result, done := handleFailedAttempt(ctx, failedAttemptParams{
			name: name, maxRetries: config.MaxRetries,
			attempt: attempt, backoff: backoff,
			output: output, err: err, shouldRetry: shouldRetry, lastErr: lastErr,
		}); done {
			return result.val, result.err
		}

		backoff *= 2
	}

	return nil, fmt.Errorf(retriesExceededFmt, name, config.MaxRetries, lastErr)
}

type failedAttemptParams struct {
	name        string
	maxRetries  int
	attempt     int
	backoff     time.Duration
	output      []byte
	err         error
	shouldRetry ShouldRetry
	lastErr     error
}

type retryResult struct {
	val []byte
	err error
}

func handleFailedAttempt(ctx context.Context, p failedAttemptParams) (retryResult, bool) {
	if !p.shouldRetry(p.err, string(p.output)) {
		return retryResult{p.output, p.err}, true
	}

	if p.attempt >= p.maxRetries {
		return retryResult{p.output, fmt.Errorf(retriesExceededFmt, p.name, p.maxRetries, p.lastErr)}, true
	}

	select {
	case <-time.After(p.backoff):
		return retryResult{}, false
	case <-ctx.Done():
		return retryResult{nil, fmt.Errorf("%s retry interrupted (lastErr=%w): %w", p.name, p.lastErr, ctx.Err())}, true
	}
}

func IsContextCanceled(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
