package linter

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// Fixer provides functionality to fix golangci-lint configurations
type Fixer struct {
	configLoader *config.Loader
	analyzer     *Analyzer
	logger       *log.Logger
}

// NewFixer creates a new fixer
func NewFixer(logger *log.Logger, analyzer *Analyzer) *Fixer {
	return &Fixer{
		configLoader: config.NewLoader(logger),
		analyzer:     analyzer,
		logger:       logger,
	}
}

// FixConfig fixes the golangci-lint configuration by enabling recommended linters
func (f *Fixer) FixConfig(configPath string, priority types.LinterPriority, dryRun bool) (*types.MigrationResult, error) {
	f.logger.Infof("Loading configuration: %s", configPath)

	cfg, err := f.configLoader.LoadConfig(configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to load config", configPath, err)
	}

	f.logger.Infof("Analyzing configuration...")
	analysis, err := f.analyzer.AnalyzeConfig(configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to analyze config", configPath, err)
	}

	enabledLinters := f.configLoader.GetLintersEnabled(cfg)
	disabledLinters := f.configLoader.GetLintersDisabled(cfg)

	fixesApplied := 0
	messages := []string{}

	for _, rec := range analysis.LinterRecommendations {
		if rec.Priority < priority {
			continue
		}

		lintName := rec.Name.String()

		isEnabled := contains(enabledLinters, lintName)
		isDisabled := contains(disabledLinters, lintName)

		if !isEnabled && !isDisabled {
			if dryRun {
				f.logger.Infof("[DRY-RUN] Would enable: %s (%s)", lintName, rec.Reason)
			} else {
				f.logger.Infof("Enabling: %s (%s)", lintName, rec.Reason)
				enabledLinters = append(enabledLinters, lintName)
				fixesApplied++
				messages = append(messages, fmt.Sprintf("Enabled %s: %s", lintName, rec.Reason))
			}
		}
	}

	if dryRun {
		f.logger.Infof("[DRY-RUN] Would apply %d fixes", fixesApplied)
		return &types.MigrationResult{
			Success:      true,
			FixesApplied: fixesApplied,
			Message:      fmt.Sprintf("Would apply %d fixes (dry-run mode)", fixesApplied),
		}, nil
	}

	if fixesApplied == 0 {
		return &types.MigrationResult{
			Success:      true,
			FixesApplied: 0,
			Message:      "No linters to enable",
		}, nil
	}

	f.logger.Infof("Creating backup...")
	backupPath, err := f.configLoader.CreateBackup(configPath)
	if err != nil {
		return nil, errors.NewAnalysisError("failed to create backup", configPath, err)
	}

	cfg.Linters.Enable = enabledLinters
	cfg.Linters.Disable = []string{}

	f.logger.Infof("Saving configuration...")
	if err := f.configLoader.SaveConfig(cfg, configPath); err != nil {
		return nil, errors.NewAnalysisError("failed to save config", configPath, err)
	}

	result := &types.MigrationResult{
		Success:      true,
		FixesApplied: fixesApplied,
		Message:      fmt.Sprintf("Successfully enabled %d linters", fixesApplied),
		BackupPath:   backupPath,
	}

	return result, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
