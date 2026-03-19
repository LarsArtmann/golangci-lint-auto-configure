package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/workflow"
	"github.com/spf13/cobra"
)

// newConfigureCommand creates the configure command.
func newConfigureCommand(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	workflowBuilder *workflow.Builder,
	configLoader *config.Loader,
) *cobra.Command {
	var preset string

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Auto-configure golangci-lint (default command)",
		Long: `Automatically configures golangci-lint by enabling recommended linters.

Use --priority to filter which linters to enable:
  - critical: Only enable critical linters (security, correctness)
  - high: Enable critical and high-value linters (recommended)
  - medium: Enable all except optional linters
  - optional: Enable all linters (may be too strict)

Or use --preset for predefined linter sets:
  - minimal: Essential linters only (fastest)
  - standard: Recommended for most projects (default)
  - strict: Maximum linting (CI/CD, strict quality)
  - security: Security-focused only
  - performance: Performance optimization only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigure(cmd.Context(), logger, analyzer, configLoader, priority, preset, dryRun, configPath)
		},
	}

	cmd.Flags().
		StringVar(&priority, "priority", "high", "Minimum priority level to enable (critical, high, medium, optional)")
	cmd.Flags().
		StringVar(&preset, "preset", "", "Use a preset linter set (minimal, standard, strict, security, performance)")

	return cmd
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

	// Find config file if not specified, or use default path
	configFile := configPath
	if configFile == "" {
		configFile = configLoader.FindOrGetDefaultConfigPath(".")
	}

	// Check if config file exists, create default if not
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		logger.Infof("No config file found, creating default: %s", configFile)

		defaultConfig := configLoader.CreateDefaultConfig(ctx)

		err := configLoader.SaveConfig(defaultConfig, configFile)
		if err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check config file: %w", err)
	}

	logger.Infof("Configuring golangci-lint with config: %s", configFile)

	// Handle preset mode
	if preset != "" {
		return applyPreset(ctx, logger, configLoader, configFile, preset, dryRun)
	}

	// Create fixer and apply fixes
	fixer := linter.NewFixer(logger, analyzer)

	var linterPriority types.LinterPriority

	switch priorityParam {
	case "critical":
		linterPriority = types.LinterPriorityCritical
	case "high":
		linterPriority = types.LinterPriorityHigh
	case "medium":
		linterPriority = types.LinterPriorityMedium
	case "optional":
		linterPriority = types.LinterPriorityOptional
	default:
		linterPriority = types.LinterPriorityHigh
	}

	result, err := fixer.FixConfig(ctx, configFile, linterPriority, dryRun)
	if err != nil {
		return fmt.Errorf("failed to fix configuration: %w", err)
	}

	logger.Infof("%s", result.Message)

	return nil
}

// applyPreset applies a preset linter configuration.
func applyPreset(
	ctx context.Context,
	logger *log.Logger,
	configLoader *config.Loader,
	configFile, preset string,
	dryRun bool,
) error {
	logger.Infof("Applying preset: %s", preset)

	// Load current config
	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get preset linters
	linters, ok := constants.PresetLinters[preset]
	if !ok {
		return fmt.Errorf("%w: %s (valid: minimal, standard, strict, security, performance)", apperrors.ErrUnknownPreset, preset)
	}

	// Convert to strings
	var linterNames []string
	for _, l := range linters {
		linterNames = append(linterNames, string(l))
	}

	if dryRun {
		logger.Infof("[DRY-RUN] Would apply preset %s with %d linters:", preset, len(linterNames))

		for _, l := range linterNames {
			logger.Infof("  - %s", l)
		}

		return nil
	}

	// Ensure we're in a git repo (git provides version control, no backup needed)
	if err := configLoader.EnsureGitRepo(ctx, "."); err != nil {
		return err
	}

	// Update config
	cfg.Linters.Enable = linterNames
	cfg.Linters.Disable = []string{}

	// Save config
	if err := configLoader.SaveConfig(cfg, configFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	logger.Infof("✅ Applied preset %s with %d linters", preset, len(linterNames))

	return nil
}
