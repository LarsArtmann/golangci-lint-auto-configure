package cli

import (
	"context"
	"slices"
	"strings"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/detection"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func resolvePresets(presets []string, detect bool, logger *log.Logger) []string {
	detectedExtraFormatters = nil

	if !detect {
		return presets
	}

	detector := detection.NewDetector(".")

	if slices.Contains(presets, "format") {
		if hasSwaggo, err := detector.HasSwaggo(); err == nil && hasSwaggo {
			detectedExtraFormatters = []types.FormatterName{"swaggo"}

			logger.Infof("🔍 Detected swaggo usage — adding swaggo formatter to format preset")
		}

		return presets
	}

	projectType := detector.Detect()

	logger.Infof("🔍 Detected project type: %s", projectType.String())
	logger.Infof("📋 Selected preset: %s", projectType.Preset())

	return []string{projectType.Preset()}
}

func handlePresetMode(
	ctx context.Context,
	logger *log.Logger,
	configLoader *config.Loader,
	analyzer *linter.Analyzer,
	configFile string,
	presets []string,
	dryRun bool,
	_ bool, // noAudit — presets don't record to the ledger (yet)
) error {
	if err := applyPreset(ctx, logger, configLoader, configFile, presets, dryRun); err != nil {
		return apperrors.WrapClassifiedf(err, "configure.preset",
			"apply preset failed (presets=%v, dryRun=%t)", presets, dryRun)
	}

	if !dryRun {
		runFmtCommand(ctx, logger, analyzer, configFile)
	}

	return nil
}

func loadPresetConfig(
	logger *log.Logger,
	configLoader presetConfigLoader,
	configFile string,
	presets []string,
	dryRun bool,
) (*types.Config, []types.LinterName, error) {
	logger.Infof("Applying presets: %s", strings.Join(presets, "+"))

	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return nil, nil, apperrors.WrapClassifiedf(err, "configure.load_preset_config",
			"failed to load config (presets=%v, dryRun=%t)", presets, dryRun)
	}

	linterSet := types.NewSet[types.LinterName]()

	for _, preset := range presets {
		linters, ok := constants.PresetLinters[preset]
		if !ok {
			return nil, nil, errorfamily.WrapRejectionf(apperrors.ErrUnknownPreset,
				"configure.unknown_preset", "%s (valid: %s, dryRun=%t)",
				preset, constants.ValidPresets, dryRun)
		}

		for _, l := range linters {
			linterSet.Add(l)
		}
	}

	return cfg, types.ToSortedSlice(linterSet), nil
}

func savePresetConfig(
	logger *log.Logger,
	configLoader presetConfigLoader,
	cfg *types.Config,
	configFile string,
	presets []string,
	linterNames []types.LinterName,
) error {
	cfg.Linters.Enable = linterNames
	cfg.Linters.Disable = []types.LinterName{}

	applyPresetFormatters(logger, cfg, presets)

	generatedCount := linter.ApplyGeneratedExclusions(logger, cfg, configFile)
	if generatedCount > 0 {
		logger.Infof("Added %d generated file exclusions", generatedCount)
	}

	if err := configLoader.SaveConfig(cfg, configFile); err != nil {
		return apperrors.WrapClassifiedf(err, "configure.save_preset",
			"failed to save config (presets=%v, linterCount=%d)",
			presets, len(linterNames))
	}

	logger.Infof(
		"✅ Applied presets %s with %d linters",
		strings.Join(presets, "+"),
		len(linterNames),
	)

	return nil
}

func applyPresetFormatters(logger *log.Logger, cfg *types.Config, presets []string) {
	formatterSet := types.NewSet[types.FormatterName]()

	for _, preset := range presets {
		formatters, ok := constants.PresetFormatters[preset]
		if !ok {
			continue
		}

		for _, f := range formatters {
			formatterSet.Add(f)
		}
	}

	if formatterSet.Len() == 0 {
		return
	}

	formatterNames := types.ToSortedSlice(formatterSet)
	formatterNames = append(formatterNames, detectedExtraFormatters...)
	cfg.Formatters.Enable = formatterNames

	logger.Infof("Enabling %d formatters from presets: %v", len(formatterNames), formatterNames)
}

// applyPreset applies a preset linter configuration.
func applyPreset(
	_ context.Context,
	logger *log.Logger,
	configLoader presetConfigLoader,
	configFile string,
	presets []string,
	dryRun bool,
) error {
	cfg, linterNames, err := loadPresetConfig(logger, configLoader, configFile, presets, dryRun)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "configure.load_preset",
			"load preset config failed (presets=%v, dryRun=%t)",
			presets, dryRun)
	}

	if dryRun {
		logDryRunPreset(logger, presets, linterNames)

		return nil
	}

	if err := backupConfigFile(logger, configFile); err != nil {
		return apperrors.WrapClassifiedf(err, "configure.backup",
			"failed to backup config before preset application (file=%s)", configFile)
	}

	return savePresetConfig(logger, configLoader, cfg, configFile, presets, linterNames)
}

func logDryRunPreset(logger *log.Logger, presets []string, linterNames []types.LinterName) {
	logger.Infof(
		"[DRY-RUN] Would apply presets %s with %d linters:",
		strings.Join(presets, "+"),
		len(linterNames),
	)

	for _, l := range linterNames {
		logger.Infof("  - %s", l)
	}

	formatterSet := make(map[string]struct{})

	for _, preset := range presets {
		if formatters, ok := constants.PresetFormatters[preset]; ok {
			for _, f := range formatters {
				formatterSet[string(f)] = struct{}{}
			}
		}
	}

	if len(formatterSet) > 0 {
		formatters := mapKeys(formatterSet)
		logger.Infof("[DRY-RUN] Would also enable %d formatters:", len(formatters))

		for _, f := range formatters {
			logger.Infof("  - %s", f)
		}
	}
}
