package cli

import (
	"context"
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/ui"
	"github.com/spf13/cobra"
)

// presetConfigLoader defines the interface needed for applyPreset.
type presetConfigLoader interface {
	LoadConfig(path string) (*types.Config, error)
	SaveConfig(config *types.Config, path string) error
}

// Priority string constants for command-line arguments.
const (
	priorityCritical = "critical"
	priorityHigh     = "high"
	priorityMedium   = "medium"
	priorityOptional = "optional"
)

const configureLong = `Automatically configures golangci-lint by enabling recommended linters.

Use --priority to filter which linters to enable (default: optional):
  - critical: Only enable critical linters (security, correctness)
  - high: Enable critical and high-value linters
  - medium: Enable all except optional linters
  - optional: Enable all linters (default)

Or use --preset for predefined linter sets:
  - minimal: Essential linters only (fastest)
  - standard: Recommended for most projects (default)
  - strict: Maximum linting (CI/CD, strict quality)
  - security: Security-focused only
  - performance: Performance optimization only

Or use --detect to automatically select a preset based on project type:`

// runFmtCommand runs golangci-lint fmt to format Go source files.
func runFmtCommand(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configFile string,
) {
	if err := analyzer.FindBinary(ctx); err != nil {
		logger.Warnf("⚠️  Could not find golangci-lint binary: %v", err)

		return
	}

	if err := analyzer.RunFmtCommand(ctx, configFile); err != nil {
		logger.Warnf("⚠️  golangci-lint fmt failed: %v", err)
	}
}

func newConfigureCommand(builder *CommandBuilder) *cobra.Command {
	var (
		preset string
		detect bool
	)

	cmd := builder.Build(
		"configure",
		"Auto-configure golangci-lint (default command)",
		func(cmd *cobra.Command, _ []string) error {
			return runDetectOrConfigure(
				cmd,
				builder.Logger(),
				builder.Analyzer(),
				builder.ConfigLoader(),
				preset,
				detect,
			)
		},
		WithLong(configureLong),
	)

	cmd.Flags().
		StringVar(&priority, "priority", "optional", "Minimum priority level to enable (critical, high, medium, optional)")
	cmd.Flags().
		StringVar(&preset, "preset", "", "Use a preset linter set (minimal, standard, strict, security, performance)")
	cmd.Flags().
		BoolVar(&detect, "detect", false, "Auto-detect project type and select appropriate preset")

	return cmd
}

func runDetectOrConfigure(
	cmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	preset string,
	detect bool,
) error {
	selectedPreset := preset

	if detect {
		detector := detection.NewDetector(".")
		projectType := detector.Detect()
		selectedPreset = projectType.Preset()

		logger.Infof("🔍 Detected project type: %s", projectType.String())
		logger.Infof("📋 Selected preset: %s", selectedPreset)
	}

	return runConfigure(
		cmd.Context(),
		logger,
		analyzer,
		configLoader,
		priority,
		selectedPreset,
		dryRun,
		configPath,
	)
}

func prepareConfigFile(
	ctx context.Context,
	configPath string,
	configLoader *config.Loader,
	logger *log.Logger,
) (string, error) {
	configFile, err := resolveConfigPath(ctx, configLoader, logger, configPath, dryRun)
	if err != nil {
		return "", err
	}

	inGitRepo := configLoader.IsGitRepo(ctx, ".")
	if !inGitRepo {
		logger.Warnf("⚠️  Not in a git repository - backup files won't be created")
		logger.Warnf("   (Initialize with: git init)")
	}

	if err := ensureConfigFile(ctx, configFile, inGitRepo, logger, configLoader); err != nil {
		return "", err
	}

	return configFile, nil
}

// runConfigure executes the configure command logic.
func runConfigure(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	priorityParam, preset string,
	dryRun bool,
	configPath string,
) error {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	configFile, err := prepareConfigFile(ctx, configPath, configLoader, logger)
	if err != nil {
		return fmt.Errorf(
			"prepare config failed (priority=%s, preset=%s, dryRun=%t): %w",
			priorityParam,
			preset,
			dryRun,
			err,
		)
	}

	logger.Infof("Configuring golangci-lint with config: %s", configFile)

	if preset != "" {
		return handlePresetMode(ctx, logger, configLoader, analyzer, configFile, preset, dryRun)
	}

	return runFixerMode(ctx, logger, analyzer, configLoader, configFile, priorityParam, dryRun)
}

func handlePresetMode(
	ctx context.Context,
	logger *log.Logger,
	configLoader *config.Loader,
	analyzer *linter.Analyzer,
	configFile, preset string,
	dryRun bool,
) error {
	if err := applyPreset(ctx, logger, configLoader, configFile, preset, dryRun); err != nil {
		return fmt.Errorf("apply preset failed (preset=%s, dryRun=%t): %w", preset, dryRun, err)
	}

	if !dryRun {
		runFmtCommand(ctx, logger, analyzer, configFile)
	}

	return nil
}

func runFixerMode(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	configFile, priorityParam string,
	dryRun bool,
) error {
	fixer := linter.NewFixer(logger, analyzer, configLoader)

	linterPriority := ParsePriorityParam(priorityParam)

	result, err := fixer.FixConfig(ctx, configFile, linterPriority, dryRun)
	if err != nil {
		return fmt.Errorf(
			"failed to fix configuration (priority=%s, dryRun=%t): %w",
			priorityParam, dryRun, err,
		)
	}

	fmt.Fprintln(os.Stdout, "\n"+ui.FormatConfigHeader(configFile))
	fmt.Fprintln(os.Stdout, ui.FormatFixResult(result))

	if !dryRun {
		runFmtCommand(ctx, logger, analyzer, configFile)
	}

	logNextSteps(logger, result.NextSteps)

	return nil
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

// ensureConfigFile creates a default config file if it doesn't exist.
func ensureConfigFile(
	ctx context.Context,
	configFile string,
	inGitRepo bool,
	logger *log.Logger,
	configLoader *config.Loader,
) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if !inGitRepo {
			logger.Warnf(
				"⚠️  Creating config without git version control - changes cannot be easily reverted",
			)
		}

		logger.Infof("No config file found, creating default: %s", configFile)

		defaultConfig := configLoader.CreateDefaultConfig(ctx)

		if err := configLoader.SaveConfig(defaultConfig, configFile); err != nil {
			return fmt.Errorf("failed to create default config (inGitRepo=%t): %w", inGitRepo, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check config file (inGitRepo=%t): %w", inGitRepo, err)
	}

	return nil
}

// ParsePriorityParam converts a priority string to LinterPriority.
func ParsePriorityParam(priorityParam string) types.LinterPriority {
	switch priorityParam {
	case priorityCritical:
		return types.LinterPriorityCritical
	case priorityHigh:
		return types.LinterPriorityHigh
	case priorityMedium:
		return types.LinterPriorityMedium
	case priorityOptional:
		return types.LinterPriorityOptional
	default:
		return types.LinterPriorityOptional
	}
}

func loadPresetConfig(
	logger *log.Logger,
	configLoader presetConfigLoader,
	configFile, preset string,
	dryRun bool,
) (*types.Config, []string, error) {
	logger.Infof("Applying preset: %s", preset)

	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to load config (preset=%s, dryRun=%t): %w",
			preset, dryRun, err,
		)
	}

	linters, ok := constants.PresetLinters[preset]
	if !ok {
		return nil, nil, fmt.Errorf(
			"%w: %s (valid: %s, dryRun=%t)",
			apperrors.ErrUnknownPreset,
			preset,
			constants.ValidPresets,
			dryRun,
		)
	}

	return cfg, convertLinterNames(linters), nil
}

func savePresetConfig(
	logger *log.Logger,
	configLoader presetConfigLoader,
	cfg *types.Config,
	configFile, preset string,
	linterNames []string,
) error {
	cfg.Linters.Enable = linterNames
	cfg.Linters.Disable = []string{}

	generatedCount := linter.ApplyGeneratedExclusions(logger, cfg, configFile)
	if generatedCount > 0 {
		logger.Infof("Added %d generated file exclusions", generatedCount)
	}

	if err := configLoader.SaveConfig(cfg, configFile); err != nil {
		return fmt.Errorf(
			"failed to save config (preset=%s, linterCount=%d): %w",
			preset, len(linterNames), err,
		)
	}

	logger.Infof("✅ Applied preset %s with %d linters", preset, len(linterNames))

	return nil
}

// applyPreset applies a preset linter configuration.
func applyPreset(
	_ context.Context,
	logger *log.Logger,
	configLoader presetConfigLoader,
	configFile, preset string,
	dryRun bool,
) error {
	cfg, linterNames, err := loadPresetConfig(logger, configLoader, configFile, preset, dryRun)
	if err != nil {
		return fmt.Errorf(
			"load preset config failed (preset=%s, dryRun=%t): %w",
			preset,
			dryRun,
			err,
		)
	}

	if dryRun {
		logDryRunPreset(logger, preset, linterNames)

		return nil
	}

	return savePresetConfig(logger, configLoader, cfg, configFile, preset, linterNames)
}

func convertLinterNames(linters []types.LinterName) []string {
	names := make([]string, len(linters))

	for i, l := range linters {
		names[i] = string(l)
	}

	return names
}

func logDryRunPreset(logger *log.Logger, preset string, linterNames []string) {
	logger.Infof("[DRY-RUN] Would apply preset %s with %d linters:", preset, len(linterNames))

	for _, l := range linterNames {
		logger.Infof("  - %s", l)
	}
}
