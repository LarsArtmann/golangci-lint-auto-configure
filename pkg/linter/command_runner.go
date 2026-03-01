package linter

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
)

// runLintersCommand runs `golangci-lint linters` and returns JSON output.
func (a *Analyzer) runLintersCommand(ctx context.Context, configPath string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, a.golangciLintPath, "linters", "--config", configPath, "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		a.logger.Debugf("golangci-lint linters command failed: %v", err)
		a.logger.Debugf("Output: %s", string(output))

		return output, errors.NewAnalysisError("golangci-lint linters command failed", "", err)
	}

	return output, nil
}

// runFormattersCommand runs `golangci-lint formatters` and returns JSON output.
func (a *Analyzer) runFormattersCommand(ctx context.Context, configPath string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, a.golangciLintPath, "formatters", "--config", configPath, "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Command may not exist in older golangci-lint versions
		return nil, fmt.Errorf("formatters command not available: %w", err)
	}

	return output, nil
}
