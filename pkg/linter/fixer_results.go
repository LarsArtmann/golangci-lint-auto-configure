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
