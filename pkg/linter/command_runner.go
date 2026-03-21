package linter

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	apperrors "github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
)

// runLintersCommand runs `golangci-lint linters` and returns JSON output.
func (a *Analyzer) runLintersCommand(ctx context.Context, configPath string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, a.golangciLintPath, "linters", "--config", configPath, "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		a.logger.Debugf("golangci-lint linters command failed: %v", err)
		a.logger.Debugf("Output: %s", outputStr)

		if outputStr != "" {
			return output, apperrors.NewAnalysisError(
				fmt.Sprintf("golangci-lint linters command failed: %s", outputStr),
				"",
				err,
			)
		}

		return output, apperrors.NewAnalysisError("golangci-lint linters command failed", "", err)
	}

	return output, nil
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
