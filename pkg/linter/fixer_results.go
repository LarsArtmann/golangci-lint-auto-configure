package linter

import (
	"fmt"

	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// analysisError creates a standardized error for analysis failures.
func analysisError(
	action string,
	priority types.LinterPriority,
	dryRun bool,
	configPath string,
	version string,
	err error,
) error {
	return apperrors.NewAnalysisError(
		fmt.Sprintf("failed to %s (priority=%d, dryRun=%t, version=%s)", action, priority, dryRun, version),
		configPath,
		err,
	)
}

// migrationError creates a standardized migration error result.
func migrationError(
	operation string,
	priority types.LinterPriority,
	dryRun bool,
	configPath string,
	version string,
	err error,
) (*types.MigrationResult, error) {
	return nil, analysisError(operation, priority, dryRun, configPath, version, err)
}

// dryRunResult creates a result for dry-run mode with the number of fixes that would be applied.
func dryRunResult(counts fixCounts) (*types.MigrationResult, error) {
	return &types.MigrationResult{
		FixesApplied: counts.total(),
		Message:      fmt.Sprintf("[DRY-RUN] Would apply %d fixes", counts.total()),
		DryRun:       true,
		NextSteps: []string{
			"Run without --dry-run to apply these fixes",
			"Then run 'golangci-lint run --fix' to auto-fix code issues",
		},
	}, nil
}

// noFixesResult creates a result when no fixes are needed.
func noFixesResult() (*types.MigrationResult, error) {
	return &types.MigrationResult{
		FixesApplied: 0,
		Message:      "No fixes to apply",
		NextSteps: []string{
			"Your configuration is already up to date",
			"Run 'golangci-lint run' to check for code issues",
		},
	}, nil
}

// successResult creates a result after successfully applying fixes.
func successResult(counts fixCounts) (*types.MigrationResult, error) {
	return &types.MigrationResult{
		FixesApplied: counts.total(),
		Message: fmt.Sprintf(
			"Successfully applied %d fixes (%d linters, %d formatters, %d generated, %d deprecated, %d redundant, %d normalization)",
			counts.total(),
			counts.enable,
			counts.formatter,
			counts.generated,
			counts.deprecation,
			counts.redundant,
			counts.normalization,
		),
		NextSteps: []string{
			"Run 'golangci-lint run --fix' to auto-fix code issues found by the newly enabled linters",
			"Run 'golangci-lint run' to see remaining issues that require manual fixes",
		},
	}, nil
}
