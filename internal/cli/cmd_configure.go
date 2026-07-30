package cli

import (
	"context"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/spf13/cobra"
)

// presetConfigLoader defines the interface needed for applyPreset.
type presetConfigLoader interface {
	LoadConfig(path string) (*types.Config, error)
	SaveConfig(config *types.Config, path string) error
}

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
  - reference: All critical + high priority linters
  - format: Core formatters + essential linters
  - house: Validated winning formatter stack (4 formatters)

Or use --detect to automatically select a preset based on project type.
Or use --recommend to analyze your project and apply multiple presets at once.`

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
	var presets []string

	var detect, recommend bool

	cmd := builder.Build(
		"configure",
		"Auto-configure golangci-lint (default command)",
		func(cmd *cobra.Command, _ []string) error {
			builder.Analyzer().SetPragmatic(builder.Flags().Pragmatic)
			applyRecommendation(builder.Logger(), &presets, recommend, &detect)

			return runDetectOrConfigure(
				cmd,
				builder.Logger(),
				builder.Analyzer(),
				builder.ConfigLoader(),
				builder.Flags(),
				presets,
				detect,
			)
		},
		WithLong(configureLong),
	)

	addConfigureFlags(cmd, builder.Flags(), &presets, &detect, &recommend)

	return cmd
}

func applyRecommendation(logger *log.Logger, presets *[]string, recommend bool, detect *bool) {
	if !recommend {
		return
	}

	detector := detection.NewDetector(".")
	*presets = detector.RecommendPresets()
	*detect = true

	logger.Infof("🔍 Recommended presets: %v", *presets)
}

func addConfigureFlags(
	cmd *cobra.Command,
	flags *Flags,
	presets *[]string,
	detect, recommend *bool,
) {
	cmd.Flags().
		StringVar(&flags.Priority, "priority", "optional", "Minimum priority level to enable (critical, high, medium, optional)")
	cmd.Flags().
		StringArrayVar(presets, "preset", nil, "Use preset linter sets (can be repeated: --preset minimal --preset security)")
	cmd.Flags().
		BoolVar(detect, "detect", false, "Auto-detect project type and select appropriate preset")
	cmd.Flags().
		BoolVar(recommend, "recommend", false, "Analyze project and recommend multiple presets (implies --detect)")
	cmd.Flags().
		BoolVar(&flags.Check, "check", false, "Check mode: exit 0 if config is optimal, exit 1 if changes needed (no modifications)")
	cmd.Flags().
		BoolVar(&flags.NoAudit, "no-audit", false, "Skip writing to the audit ledger (also: "+auditEnvVar+" env var)")
	cmd.Flags().
		BoolVar(&flags.Pragmatic, "pragmatic", false, "Drop the 4 highest-noise linters (gochecknoglobals, wrapcheck, ireturn, funlen) from the enable set")
	cmd.Flags().
		BoolVar(&flags.ForceSettings, "force-settings", false, "Overwrite existing linter settings with curated defaults (useful for refreshing stale configs)")
}

func runDetectOrConfigure(
	cmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	flags *Flags,
	presets []string,
	detect bool,
) error {
	resolvedPresets, extraFormatters := resolvePresets(presets, detect, logger)

	return runConfigure(
		cmd.Context(),
		logger,
		analyzer,
		configLoader,
		flags,
		resolvedPresets,
		extraFormatters,
	)
}

func runConfigure(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	flags *Flags,
	presets []string,
	extraFormatters []types.FormatterName,
) error {
	isDryRun := flags.Check || flags.DryRun

	configFile, err := prepareConfigFile(ctx, flags.ConfigPath, configLoader, logger, isDryRun)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "configure.prepare_config",
			"prepare config failed (priority=%s, presets=%v, dryRun=%t)",
			flags.Priority, presets, isDryRun)
	}

	logger.Infof("Configuring golangci-lint with config: %s", configFile)

	if len(presets) > 0 {
		return handlePresetMode(ctx, logger, configLoader, analyzer,
			configFile, presets, flags, extraFormatters)
	}

	return runFixerMode(ctx, logger, analyzer, configLoader, configFile, flags)
}
