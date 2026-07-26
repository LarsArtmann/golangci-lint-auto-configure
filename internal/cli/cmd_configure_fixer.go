package cli

import (
	"context"
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	appdiff "github.com/larsartmann/golangci-lint-auto-configure/pkg/diff"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/ui"
)

func runFixerMode(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	configFile, priorityParam string,
	isDryRun,
	check,
	showDiff bool,
) error {
	fixer := newConfiguredFixer(ctx, logger, analyzer, configLoader, configFile)

	linterPriority, err := ParsePriorityParam(priorityParam)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "configure.parse_priority",
			"invalid priority %q", priorityParam)
	}

	originalCfg := captureOriginalConfig(showDiff, configLoader, configFile, logger)
	effectiveDryRun := effectiveDryRunForCheckDiff(isDryRun, check, showDiff, originalCfg)

	result, err := fixer.FixConfig(ctx, configFile, linterPriority, effectiveDryRun)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "configure.fixer",
			"failed to fix configuration (priority=%s, dryRun=%t)",
			priorityParam, isDryRun)
	}

	return finalizeFixerResult(ctx, logger, analyzer, configLoader,
		originalCfg, configFile, result, isDryRun, check, showDiff)
}

func finalizeFixerResult(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	originalCfg *types.Config,
	configFile string,
	result *types.MigrationResult,
	isDryRun, check,
	showDiff bool,
) error {
	applyCheckDiff(showDiff, check, configLoader, originalCfg, configFile, logger)
	displayFixResult(configFile, result)

	if !check {
		runFmtUnlessDry(ctx, logger, analyzer, configFile, isDryRun)
	}

	logNextSteps(logger, result.NextSteps)

	return handleCheckMode(check, result, logger)
}

func effectiveDryRunForCheckDiff(isDryRun, check, showDiff bool, originalCfg *types.Config) bool {
	if check && showDiff && originalCfg != nil {
		return false
	}

	return isDryRun
}

func applyCheckDiff(
	shouldDiff, check bool,
	configLoader *config.Loader,
	originalCfg *types.Config,
	configFile string,
	logger *log.Logger,
) {
	if !shouldDiff || originalCfg == nil {
		return
	}

	showConfigDiff(configLoader, originalCfg, configFile, logger)

	if check {
		restoreOriginalConfig(configLoader, originalCfg, configFile, logger)
	}
}

func captureOriginalConfig(
	shouldClone bool,
	configLoader *config.Loader,
	configFile string,
	logger *log.Logger,
) *types.Config {
	if !shouldClone {
		return nil
	}

	return cloneConfig(configLoader, configFile, logger)
}

func displayFixResult(configFile string, result *types.MigrationResult) {
	fmt.Fprintln(os.Stdout, "\n"+ui.FormatConfigHeader(configFile))
	fmt.Fprintln(os.Stdout, ui.FormatFixResult(result))
}

func runFmtUnlessDry(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configFile string,
	isDryRun bool,
) {
	if !isDryRun {
		runFmtCommand(ctx, logger, analyzer, configFile)
	}
}

func handleCheckMode(check bool, result *types.MigrationResult, logger *log.Logger) error {
	if check && result.FixesApplied > 0 {
		logger.Infof("Check mode: %d changes needed", result.FixesApplied)

		return apperrors.ErrChangesNeeded
	}

	return nil
}

func showConfigDiff(
	configLoader *config.Loader,
	oldCfg *types.Config,
	configFile string,
	logger *log.Logger,
) {
	newCfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		logger.Debugf("Failed to load modified config for diff: %v", err)

		return
	}

	differ := appdiff.NewDiffer()
	changes := differ.Compare(oldCfg, newCfg)

	if len(changes) == 0 {
		return
	}

	fmt.Fprintln(os.Stdout, differ.FormatChanges(changes))
}

func restoreOriginalConfig(
	configLoader *config.Loader,
	originalCfg *types.Config,
	configFile string,
	logger *log.Logger,
) {
	if err := configLoader.SaveConfig(originalCfg, configFile); err != nil {
		logger.Warnf("⚠️  Failed to restore original config after check+diff: %v", err)

		return
	}

	logger.Debugf("Restored original config after check+diff: %s", configFile)
}

func logNextSteps(logger *log.Logger, steps []string) {
	if len(steps) == 0 {
		return
	}

	logger.Infof("Next steps:")

	for _, step := range steps {
		logger.Infof("  → %s", step)
	}
}
