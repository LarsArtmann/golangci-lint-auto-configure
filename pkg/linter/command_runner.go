package linter

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
)

const (
	maxRetries     = 3
	initialBackoff = 500 * time.Millisecond
)

// isParallelRunningError checks if the error output indicates a parallel golangci-lint is running.
func isParallelRunningError(output string) bool {
	return strings.Contains(output, "parallel golangci-lint is running")
}

// runCommandWithRetry runs a command and retries if a parallel golangci-lint is running.
func (a *Analyzer) runCommandWithRetry(ctx context.Context, name string, args ...string) ([]byte, error) {
	var lastErr error

	backoff := initialBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		cmd := exec.CommandContext(ctx, a.golangciLintPath, args...)

		output, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(output))

		if err == nil {
			return output, nil
		}

		// Check if this is a parallel running error and we have retries left
		if isParallelRunningError(outputStr) && attempt < maxRetries {
			a.logger.Debugf(
				"parallel golangci-lint is running, retrying in %v (attempt %d/%d)",
				backoff,
				attempt+1,
				maxRetries,
			)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, fmt.Errorf("retry interrupted: %w", ctx.Err())
			}

			backoff *= 2 // Exponential backoff

			lastErr = err
			continue
		}

		// Either not a parallel error or out of retries
		a.logger.Debugf("%s command failed: %v", name, err)
		a.logger.Debugf("Output: %s", outputStr)

		if outputStr != "" {
			return output, apperrors.NewAnalysisError(
				fmt.Sprintf("golangci-lint %s command failed: %s", name, outputStr),
				"",
				err,
			)
		}

		return output, apperrors.NewAnalysisError(
			fmt.Sprintf("golangci-lint %s command failed", name),
			"",
			err,
		)
	}

	return nil, lastErr
}

// runLintersCommand runs `golangci-lint linters` and returns JSON output.
func (a *Analyzer) runLintersCommand(ctx context.Context, configPath string) ([]byte, error) {
	return a.runCommandWithRetry(ctx, "linters", "linters", "--config", configPath, "--json")
}

// runFormattersCommand runs `golangci-lint formatters` and returns JSON output.
func (a *Analyzer) runFormattersCommand(ctx context.Context, configPath string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, a.golangciLintPath, "formatters", "--config", configPath, "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "" {
			return nil, fmt.Errorf("formatters command not available: %s (cause: %w)", outputStr, err)
		}

		return nil, fmt.Errorf("formatters command not available: %w", err)
	}

	return output, nil
}

// RunFmtCommand runs `golangci-lint fmt` to format Go source files.
func (a *Analyzer) RunFmtCommand(ctx context.Context, configPath string) error {
	cmd := exec.CommandContext(ctx, a.golangciLintPath, "fmt", "--config", configPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		a.logger.Debugf("golangci-lint fmt command output: %s", outputStr)

		return fmt.Errorf("golangci-lint fmt failed: %w", err)
	}

	return nil
}
