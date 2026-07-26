package cli

import (
	"context"
	"fmt"
	"os"
	"slices"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func prepareConfigFile(
	ctx context.Context,
	configPath string,
	configLoader *config.Loader,
	logger *log.Logger,
	isDryRun bool,
) (string, error) {
	configFile, err := resolveConfigPath(ctx, configLoader, logger, configPath, isDryRun, false)
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

func cloneConfig(configLoader *config.Loader, configFile string, logger *log.Logger) *types.Config {
	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		logger.Debugf("Failed to load config for diff: %v", err)

		return nil
	}

	return cfg.Clone()
}

func backupConfigFile(logger *log.Logger, configFile string) error { //nolint:erraudit // advisory: errors classified at command boundary via go-error-family, not per-function types
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

	logger.Infof("📦 Backed up config to %s", backupPath)

	return nil
}

func convertNames[T ~string](names []T) []string {
	result := make([]string, 0, len(names))

	for _, n := range names {
		result = append(result, string(n))
	}

	return result
}

// mapKeys extracts and sorts keys from a map[string]struct{} set.
func mapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}
