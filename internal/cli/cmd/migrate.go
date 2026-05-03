package cmd

import (
	"fmt"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/migration"
	"github.com/spf13/cobra"
)

// MigrateFlags holds the flags for the migrate command.
type MigrateFlags struct {
	ConfigPath     string
	DryRun         bool
	Verbose        bool
	SkipValidation bool
}

const migrateLong = `Migrates golangci-lint configuration from v1 to v2 schema.

This command:
1. Verifies you're in a git repository (for version control)
2. Migrates the configuration to v2 schema
3. Validates the migrated configuration
4. Shows what changed

Use --dry-run to preview changes without modifying files.
Use --skip-validation if the v1 config has known issues.`

// NewMigrateCommand creates the migrate command.
func NewMigrateCommand(
	logger *log.Logger,
	configLoader *config.Loader,
	flags MigrateFlags,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate configuration from v1 to v2 schema",
		Long:  migrateLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMigrate(cmd, logger, configLoader, flags)
		},
	}

	cmd.Flags().
		BoolVar(&flags.SkipValidation, "skip-validation", false, "Skip validation of v1 configuration")

	return cmd
}

func runMigrate(
	cmd *cobra.Command,
	logger *log.Logger,
	configLoader *config.Loader,
	flags MigrateFlags,
) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	configFile, err := resolveMigrateConfig(cmd, configLoader, flags, verbose)
	if err != nil {
		return err
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	skipValidation, _ := cmd.Flags().GetBool("skip-validation")

	return executeMigration(logger, configLoader, configFile, dryRun, skipValidation, verbose)
}

func resolveMigrateConfig(
	cmd *cobra.Command,
	configLoader *config.Loader,
	flags MigrateFlags,
	verbose bool,
) (string, error) {
	configFile, _ := cmd.Flags().GetString("config")
	if configFile == "" {
		var err error

		configFile, err = configLoader.FindConfigFile(".")
		if err != nil {
			return "", fmt.Errorf(
				"no config file found (configPath=%q, verbose=%t): %w",
				flags.ConfigPath, verbose, err,
			)
		}
	}

	configLoader.HasMultipleConfigFiles(".")

	return configFile, nil
}

func executeMigration(
	logger *log.Logger,
	configLoader *config.Loader,
	configFile string,
	dryRun, skipValidation, verbose bool,
) error {
	oldConfig, err := loadConfigForMigration(logger, configLoader, configFile)
	if err != nil {
		return err
	}

	if isAlreadyV2(oldConfig) {
		logger.Infof("Configuration is already version 2, no migration needed")

		return nil
	}

	return runMigrator(logger, configLoader, configFile, dryRun, skipValidation, verbose, oldConfig)
}

func loadConfigForMigration(
	logger *log.Logger,
	configLoader *config.Loader,
	configFile string,
) (*config.Config, error) {
	logger.Infof("Migrating configuration: %s", configFile)

	oldConfig, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not load config %s: %w", configFile, err)
	}

	return oldConfig, nil
}

func isAlreadyV2(cfg *config.Config) bool {
	return cfg.Version == "2"
}

func runMigrator(
	logger *log.Logger,
	configLoader *config.Loader,
	configFile string,
	dryRun, skipValidation, verbose bool,
	oldConfig *config.Config,
) error {
	migrator, err := createMigrator(configFile, dryRun, skipValidation, verbose, logger)
	if err != nil {
		return err
	}

	success, fixesApplied, err := migrator.MigrateToV2()
	if err != nil {
		return fmt.Errorf("migration failed for %s: %w", configFile, err)
	}

	showMigrationResult(logger, configLoader, configFile, oldConfig, success, fixesApplied, dryRun)

	return nil
}

func createMigrator(
	configFile string,
	dryRun, skipValidation, verbose bool,
	logger *log.Logger,
) (*migration.Migrator, error) {
	migrator, err := migration.NewMigrator(configFile, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	migrator.SetDryRun(dryRun)

	if skipValidation {
		migrator.SetValidator(migration.MockValidator{})
		logger.Infof("Skipping validation as requested")
	}

	if dryRun {
		logger.Infof("[DRY-RUN] Would migrate configuration from v1 to v2")
	}

	return migrator, nil
}

func showMigrationResult(
	logger *log.Logger,
	configLoader *config.Loader,
	configFile string,
	oldConfig *config.Config,
	success bool,
	fixesApplied int,
	dryRun bool,
) {
	if !success {
		logger.Infof("No migration needed (already at v2.x)")

		return
	}

	if dryRun {
		logger.Infof("[DRY-RUN] Would apply %d fixes", fixesApplied)
		logger.Infof("[DRY-RUN] Migration preview complete")

		return
	}

	logger.Infof("Configuration migrated successfully (%d fixes applied)!", fixesApplied)

	newConfig, err := configLoader.LoadConfig(configFile)
	if err != nil {
		logger.Warnf("Could not load migrated config: %v", err)
	} else if oldConfig != nil {
		ShowMigrationChanges(logger, oldConfig, newConfig)
	}
}

// ShowMigrationChanges displays the differences between old and new config.
func ShowMigrationChanges(logger *log.Logger, oldCfg, newCfg *config.Config) {
	oldLinters := len(oldCfg.Linters.Enable)
	newLinters := len(newCfg.Linters.Enable)

	if oldCfg.Version != newCfg.Version {
		logger.Infof("  Version: %s -> %s", oldCfg.Version, newCfg.Version)
	}

	if oldLinters != newLinters {
		logger.Infof("  Linters: %d -> %d", oldLinters, newLinters)
	}

	if oldCfg.Run.Timeout != newCfg.Run.Timeout {
		logger.Infof("  Timeout: %s -> %s", oldCfg.Run.Timeout, newCfg.Run.Timeout)
	}
}
