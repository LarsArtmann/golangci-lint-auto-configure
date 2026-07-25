package cli

import (
	"context"
	"fmt"
	"os"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	appdiff "github.com/larsartmann/golangci-lint-auto-configure/pkg/diff"
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
		check  bool
	)

	cmd := builder.Build(
		"configure",
		"Auto-configure golangci-lint (default command)",
		func(cmd *cobra.Command, _ []string) error {
			builder.Analyzer().SetPragmatic(pragmatic)

			return runDetectOrConfigure(
				cmd,
				builder.Logger(),
				builder.Analyzer(),
				builder.ConfigLoader(),
				preset,
				detect,
				check,
				showDiff,
			)
		},
		WithLong(configureLong),
	)

	addConfigureFlags(cmd, &preset, &detect, &check)

	return cmd
}

func addConfigureFlags(cmd *cobra.Command, preset *string, detect, check *bool) {
	cmd.Flags().
		StringVar(&priority, "priority", "optional", "Minimum priority level to enable (critical, high, medium, optional)")
	cmd.Flags().
		StringVar(preset, "preset", "", "Use a preset linter set (minimal, standard, strict, security, performance, reference, format)")
	cmd.Flags().
		BoolVar(detect, "detect", false, "Auto-detect project type and select appropriate preset")
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
	preset string,
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
		resolvePreset(preset, detect, logger),
		check || dryRun,
		configPath,
		check,
		showDiff,
	)
}

func resolvePreset(preset string, detect bool, logger *log.Logger) string {
	if !detect {
		return preset
	}

	detector := detection.NewDetector(".")
	projectType := detector.Detect()

	logger.Infof("🔍 Detected project type: %s", projectType.String())
	logger.Infof("📋 Selected preset: %s", projectType.Preset())

	return projectType.Preset()
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
	isDryRun bool,
	configPath string,
	check bool,
	showDiff bool,
) error {
	configFile, err := prepareConfigFile(ctx, configPath, configLoader, logger)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "configure.prepare_config",
			"prepare config failed (priority=%s, preset=%s, dryRun=%t)",
			priorityParam, preset, isDryRun)
	}

	logger.Infof("Configuring golangci-lint with config: %s", configFile)

	return runPresetOrFixer(
		ctx, logger, analyzer, configLoader,
		configFile, priorityParam, preset, isDryRun, check, showDiff,
	)
}

func runPresetOrFixer(
	ctx context.Context,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	configFile, priorityParam, preset string,
	isDryRun, check,
	showDiff bool,
) error {
	if preset != "" {
		return handlePresetMode(ctx, logger, configLoader, analyzer, configFile, preset, isDryRun)
	}

	return runFixerMode(ctx, logger, analyzer, configLoader, configFile, priorityParam, isDryRun, check, showDiff)
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
		return apperrors.WrapClassifiedf(err, "configure.preset",
			"apply preset failed (preset=%s, dryRun=%t)", preset, dryRun)
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
	isDryRun,
	check,
	showDiff bool,
) error {
	fixer := linter.NewFixer(logger, analyzer, configLoader)
	fixer.SetLedger(newRunLedger(ctx, logger, configFile))

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

func cloneConfig(configLoader *config.Loader, configFile string, logger *log.Logger) *types.Config {
	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		logger.Debugf("Failed to load config for diff: %v", err)

		return nil
	}

	return cfg.Clone()
}

func showConfigDiff(configLoader *config.Loader, oldCfg *types.Config, configFile string, logger *log.Logger) {
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
			return apperrors.WrapClassifiedf(err, "configure.create_default",
				"failed to create default config (inGitRepo=%t)", inGitRepo)
		}
	} else if err != nil {
		return apperrors.WrapClassifiedf(err, "configure.check_file",
			"failed to check config file (inGitRepo=%t)", inGitRepo)
	}

	return nil
}

// ParsePriorityParam converts a priority string to LinterPriority.
func ParsePriorityParam(priorityParam string) (types.LinterPriority, error) {
	priority, err := types.ParseLinterPriority(priorityParam)
	if err != nil {
		return 0, apperrors.WrapClassifiedf(err, "configure.parse_priority_param",
			"parsing linter priority %q", priorityParam)
	}

	return priority, nil
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
		return nil, nil, apperrors.WrapClassifiedf(err, "configure.load_preset_config",
			"failed to load config (preset=%s, dryRun=%t)", preset, dryRun)
	}

	linters, ok := constants.PresetLinters[preset]
	if !ok {
		return nil, nil, errorfamily.WrapRejectionf(apperrors.ErrUnknownPreset,
			"configure.unknown_preset", "%s (valid: %s, dryRun=%t)",
			preset, constants.ValidPresets, dryRun)
	}

	return cfg, convertNames(linters), nil
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

	if formatters, ok := constants.PresetFormatters[preset]; ok {
		cfg.Formatters.Enable = convertNames(formatters)
		logger.Infof("Enabling %d formatters from preset: %v", len(formatters), formatters)
	}

	generatedCount := linter.ApplyGeneratedExclusions(logger, cfg, configFile)
	if generatedCount > 0 {
		logger.Infof("Added %d generated file exclusions", generatedCount)
	}

	if err := configLoader.SaveConfig(cfg, configFile); err != nil {
		return apperrors.WrapClassifiedf(err, "configure.save_preset",
			"failed to save config (preset=%s, linterCount=%d)",
			preset, len(linterNames))
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
		return apperrors.WrapClassifiedf(err, "configure.load_preset",
			"load preset config failed (preset=%s, dryRun=%t)",
			preset, dryRun)
	}

	if dryRun {
		logDryRunPreset(logger, preset, linterNames)

		return nil
	}

	if err := backupConfigFile(logger, configFile); err != nil {
		return apperrors.WrapClassifiedf(err, "configure.backup",
			"failed to backup config before preset application (file=%s)", configFile)
	}

	return savePresetConfig(logger, configLoader, cfg, configFile, preset, linterNames)
}

func convertNames[T ~string](names []T) []string {
	result := make([]string, 0, len(names))

	for _, n := range names {
		result = append(result, string(n))
	}

	return result
}

func backupConfigFile(logger *log.Logger, configFile string) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("read config for backup: %w", err)
	}

	backupPath := configFile + ".bak"

	if err := os.WriteFile(backupPath, data, 0o600); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("write backup file: %w", err)
	}

	logger.Infof("Backed up %s to %s", configFile, backupPath)

	return nil
}

func logDryRunPreset(logger *log.Logger, preset string, linterNames []string) {
	logger.Infof("[DRY-RUN] Would apply preset %s with %d linters:", preset, len(linterNames))

	for _, l := range linterNames {
		logger.Infof("  - %s", l)
	}

	if formatters, ok := constants.PresetFormatters[preset]; ok {
		logger.Infof("[DRY-RUN] Would also enable %d formatters:", len(formatters))

		for _, f := range formatters {
			logger.Infof("  - %s", f)
		}
	}
}
