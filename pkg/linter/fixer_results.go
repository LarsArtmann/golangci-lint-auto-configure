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
	err error,
) error {
	return apperrors.NewAnalysisError(
		fmt.Sprintf("failed to %s (priority=%d, dryRun=%t)", action, priority, dryRun),
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
	err error,
) types.MigrationResultType {
	return types.Err[*types.MigrationResult](analysisError(operation, priority, dryRun, configPath, err))
}

// dryRunResult creates a result for dry-run mode with the number of fixes that would be applied.
func dryRunResult(counts fixCounts) types.MigrationResultType {
	return types.Ok(&types.MigrationResult{
		FixesApplied: counts.total(),
		Message:      fmt.Sprintf("[DRY-RUN] Would apply %d fixes", counts.total()),
		NextSteps: []string{
			"Run without --dry-run to apply these fixes",
			"Then run 'golangci-lint run --fix' to auto-fix code issues",
		},
	})
}

// noFixesResult creates a result when no fixes are needed.
func noFixesResult() types.MigrationResultType {
	return types.Ok(&types.MigrationResult{
		FixesApplied: 0,
		Message:      "No fixes to apply",
		NextSteps: []string{
			"Your configuration is already up to date",
			"Run 'golangci-lint run' to check for code issues",
		},
	})
}

// successResult creates a result after successfully applying fixes.
func successResult(counts fixCounts) types.MigrationResultType {
	return types.Ok(&types.MigrationResult{
		FixesApplied: counts.total(),
		Message: fmt.Sprintf(
			"Successfully applied %d fixes (%d linters, %d formatters, %d generated, %d deprecated, %d redundant)",
			counts.total(), counts.enable, counts.formatter, counts.generated, counts.deprecation, counts.redundant,
		),
		NextSteps: []string{
			"Run 'golangci-lint run --fix' to auto-fix code issues found by the newly enabled linters",
			"Run 'golangci-lint run' to see remaining issues that require manual fixes",
		},
	})
}
