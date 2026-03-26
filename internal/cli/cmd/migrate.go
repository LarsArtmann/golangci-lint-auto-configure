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
	OutputFormat   string
}

// NewMigrateCommand creates the migrate command.
func NewMigrateCommand(
	logger *log.Logger,
	configLoader *config.Loader,
	flags MigrateFlags,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate configuration from v1 to v2 schema",
		Long: `Migrates golangci-lint configuration from v1 to v2 schema.

This command:
1. Verifies you're in a git repository (for version control)
2. Migrates the configuration to v2 schema
3. Validates the migrated configuration
4. Shows what changed

Use --dry-run to preview changes without modifying files.
Use --skip-validation if the v1 config has known issues.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Read flags dynamically to get parsed values
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Find config file if not specified
			configFile, _ := cmd.Flags().GetString("config")
			if configFile == "" {
				var err error

				configFile, err = configLoader.FindConfigFile(".")
				if err != nil {
					return fmt.Errorf(
						"no config file found (configPath=%q, verbose=%t): %w",
						flags.ConfigPath, verbose, err,
					)
				}
			}

			configLoader.HasMultipleConfigFiles(".")

			// Read other flags
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			skipValidation, _ := cmd.Flags().GetBool("skip-validation")

			logger.Infof("Migrating configuration: %s", configFile)

			// Load current config to check version
			oldConfig, err := configLoader.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf(
					"could not load config (configFile=%s, dryRun=%t, skipValidation=%t): %w",
					configFile, dryRun, skipValidation, err,
				)
			}

			// Check if already v2
			if oldConfig.Version == "2" {
				logger.Infof("Configuration is already version 2, no migration needed")

				return nil
			}

			// Create migrator
			migrator, err := migration.NewMigrator(configFile, verbose)
			if err != nil {
				return fmt.Errorf("failed to create migrator: %w", err)
			}

			migrator.SetDryRun(dryRun)

			if skipValidation {
				migrator.SetValidator(migration.MockValidator{})
				logger.Infof("Skipping validation as requested")
			}

			if dryRun {
				logger.Infof("[DRY-RUN] Would migrate configuration from v1 to v2")
			}

			// Run migration
			success, fixesApplied, err := migrator.MigrateToV2()
			if err != nil {
				return fmt.Errorf(
					"migration failed (configFile=%s, dryRun=%t, skipValidation=%t): %w",
					configFile, dryRun, skipValidation, err,
				)
			}

			if !success {
				logger.Infof("No migration needed (already at v2.x)")

				return nil
			}

			if dryRun {
				logger.Infof("[DRY-RUN] Would apply %d fixes", fixesApplied)
				logger.Infof("[DRY-RUN] Migration preview complete")

				return nil
			}

			logger.Infof("Configuration migrated successfully (%d fixes applied)!", fixesApplied)

			// Load new config to show changes
			newConfig, err := configLoader.LoadConfig(configFile)
			if err != nil {
				logger.Warnf("Could not load migrated config: %v", err)
			} else if oldConfig != nil {
				ShowMigrationChanges(logger, oldConfig, newConfig)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.SkipValidation, "skip-validation", false, "Skip validation of v1 configuration")
	cmd.Flags().
		StringVar(&flags.OutputFormat, "format", "", "Output format (deprecated: format migration is no longer supported)")

	return cmd
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

	// Check for renamed fields (basic check)
	if oldCfg.Run.Timeout != newCfg.Run.Timeout {
		logger.Infof("  Timeout: %s -> %s", oldCfg.Run.Timeout, newCfg.Run.Timeout)
	}
}
