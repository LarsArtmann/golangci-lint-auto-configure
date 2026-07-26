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

// detectedExtraFormatters holds format-specific formatters detected by --detect
// (e.g., swaggo when Swagger annotations are found). Populated by resolvePresets.
//
//nolint:gochecknoglobals // populated by detection, read by savePresetConfig
var detectedExtraFormatters []types.FormatterName

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

	var detect, recommend, check bool

	cmd := builder.Build(
		"configure",
		"Auto-configure golangci-lint (default command)",
		func(cmd *cobra.Command, _ []string) error {
			builder.Analyzer().SetPragmatic(pragmatic)
			applyRecommendation(builder.Logger(), &presets, recommend, &detect)

			return runDetectOrConfigure(
				cmd,
				builder.Logger(),
				builder.Analyzer(),
				builder.ConfigLoader(),
				presets,
				detect,
				check,
				showDiff,
			)
		},
		WithLong(configureLong),
	)

	addConfigureFlags(cmd, &presets, &detect, &recommend, &check)

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

func addConfigureFlags(cmd *cobra.Command, presets *[]string, detect, recommend, check *bool) {
	cmd.Flags().
		StringVar(&priority, "priority", "optional", "Minimum priority level to enable (critical, high, medium, optional)")
	cmd.Flags().
		StringArrayVar(presets, "preset", nil, "Use preset linter sets (can be repeated: --preset minimal --preset security)")
	cmd.Flags().
		BoolVar(detect, "detect", false, "Auto-detect project type and select appropriate preset")
	cmd.Flags().
		BoolVar(recommend, "recommend", false, "Analyze project and recommend multiple presets (implies --detect)")
	cmd.Flags().
		BoolVar(check, "check", false, "Check mode: exit 0 if config is optimal, exit 1 if changes needed (no modifications)")
	cmd.Flags().
		BoolVar(&noAudit, "no-audit", false, "Skip writing to the audit ledger (also: "+auditEnvVar+" env var)")
	cmd.Flags().
		BoolVar(&pragmatic, "pragmatic", false, "Drop the 5 highest-noise linters (exhaustruct, gochecknoglobals, wrapcheck, ireturn, funlen) from the enable set")
}

func runDetectOrConfigure(
	cmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	presets []string,
	detect,
	check,
	showDiff bool,
) error {
	return runConfigure(
		cmd.Context(),
		logger,
		analyzer,
		configLoader,
		priority,
		resolvePresets(presets, detect, logger),
		check || dryRun,
		configPath,
		check,
		showDiff,
	)
}

// runConfigure executes the configure command logic.
func runConfigure(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	priorityParam string,
	presets []string,
	isDryRun bool,
	configPath string,
	check bool,
	showDiff bool,
) error {
	configFile, err := prepareConfigFile(ctx, configPath, configLoader, logger)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "configure.prepare_config",
			"prepare config failed (priority=%s, presets=%v, dryRun=%t)",
			priorityParam, presets, isDryRun)
	}

	logger.Infof("Configuring golangci-lint with config: %s", configFile)

	return runPresetOrFixer(
		ctx, logger, analyzer, configLoader,
		configFile, priorityParam, presets, isDryRun, check, showDiff,
	)
}

func runPresetOrFixer(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	configFile, priorityParam string,
	presets []string,
	isDryRun, check,
	showDiff bool,
) error {
	if len(presets) > 0 {
		return handlePresetMode(ctx, logger, configLoader, analyzer, configFile, presets, isDryRun)
	}

	return runFixerMode(ctx, logger, analyzer, configLoader, configFile, priorityParam, isDryRun, check, showDiff)
}
