package linter

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/utils"
)

// isParallelRunningError checks if the error output indicates a parallel golangci-lint is running.
func isParallelRunningError(output string) bool {
	return strings.Contains(output, "parallel golangci-lint is running")
}

// executeCommand creates an exec.Cmd with the given arguments.
func (a *Analyzer) executeCommand(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, a.golangciLintPath, args...)
}

// runCommandWithRetry runs a command and retries if a parallel golangci-lint is running.
func (a *Analyzer) runCommandWithRetry(ctx context.Context, name string, args ...string) ([]byte, error) {
	config := utils.DefaultConfig()

	executeOperation := func() ([]byte, error) {
		return a.executeCommand(ctx, args...).CombinedOutput()
	}

	shouldRetry := func(_ error, output string) bool {
		return isParallelRunningError(strings.TrimSpace(output))
	}

	output, err := utils.WithRetry(ctx, config, name, shouldRetry, executeOperation)
	if err != nil {
		return output, a.formatCommandError(name, output, err)
	}

	return output, nil
}

func (a *Analyzer) formatCommandError(name string, output []byte, err error) error {
	outputStr := strings.TrimSpace(string(output))

	a.logger.Debugf("%s command failed: %v", name, err)
	a.logger.Debugf("Output: %s", outputStr)

	if outputStr != "" {
		return apperrors.NewAnalysisError(
			fmt.Sprintf("golangci-lint %s command failed: %s", name, outputStr),
			"", err,
		)
	}

	return apperrors.NewAnalysisError(
		fmt.Sprintf("golangci-lint %s command failed", name),
		"", err,
	)
}

// runLintersCommand runs `golangci-lint linters` and returns JSON output.
func (a *Analyzer) runLintersCommand(ctx context.Context, configPath string) ([]byte, error) {
	return a.runCommandWithRetry(ctx, "linters", "linters", "--config", configPath, "--json")
}

// runFormattersCommand runs `golangci-lint formatters` and returns JSON output.
func (a *Analyzer) runFormattersCommand(ctx context.Context, configPath string) ([]byte, error) {
	return a.runCommandWithRetry(ctx, "formatters", "formatters", "--config", configPath, "--json")
}

// RunFmtCommand runs `golangci-lint fmt` to format Go source files.
func (a *Analyzer) RunFmtCommand(ctx context.Context, configPath string) error {
	_, err := a.runCommandWithRetry(ctx, "fmt", "fmt", "--config", configPath)
	if err != nil {
		return fmt.Errorf("golangci-lint fmt failed: %w", err)
	}

	return nil
}
