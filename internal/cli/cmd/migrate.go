package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
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
2. Runs golangci-lint migrate to convert the schema
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
					return fmt.Errorf("no config file found: %w", err)
				}
			}

			// Read other flags
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			skipValidation, _ := cmd.Flags().GetBool("skip-validation")
			outputFormat, _ := cmd.Flags().GetString("format")

			logger.Infof("Migrating configuration: %s", configFile)

			// Load current config to check version
			oldConfig, err := configLoader.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("could not load config: %w", err)
			}

			// Check if already v2
			if oldConfig.Version == "2" {
				logger.Infof("Configuration is already version 2, no migration needed")

				return nil
			}

			// Ensure we're in a git repo (git provides version control, no backup needed)
			if !dryRun {
				err := configLoader.EnsureGitRepo(context.Background(), ".")
				if err != nil {
					return fmt.Errorf("failed to ensure git repo: %w", err)
				}
			} else {
				logger.Infof("[DRY-RUN] Would verify git repository")
			}

			// Build golangci-lint migrate command
			migrateArgs := []string{"migrate", "--config", configFile}
			if skipValidation {
				migrateArgs = append(migrateArgs, "--skip-validation")
			}

			if outputFormat != "" {
				migrateArgs = append(migrateArgs, "--format", outputFormat)
			}

			if dryRun {
				logger.Infof("[DRY-RUN] Would run: golangci-lint %s", strings.Join(migrateArgs, " "))
				logger.Infof("[DRY-RUN] Migration preview complete")

				return nil
			}

			// Run golangci-lint migrate
			logger.Infof("Running golangci-lint migrate...")

			migrateCmd := exec.Command("golangci-lint", migrateArgs...)

			output, err := migrateCmd.CombinedOutput()
			if err != nil {
				logger.Errorf("Migration failed: %v", err)
				logger.Infof("Output: %s", string(output))
				logger.Infof("Use git to restore if needed")

				return fmt.Errorf("migration failed: %w", err)
			}

			logger.Infof("Migration completed successfully")

			if len(output) > 0 {
				logger.Infof("Output: %s", string(output))
			}

			// Load new config to show changes
			newConfig, err := configLoader.LoadConfig(configFile)
			if err != nil {
				logger.Warnf("Could not load migrated config: %v", err)
			} else if oldConfig != nil {
				ShowMigrationChanges(logger, oldConfig, newConfig)
			}

			logger.Infof("Configuration migrated successfully!")

			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.SkipValidation, "skip-validation", false, "Skip validation of v1 configuration")
	cmd.Flags().StringVar(&flags.OutputFormat, "format", "", "Output format (yml, yaml, toml, json)")

	return cmd
}

// ShowMigrationChanges displays the differences between old and new config.
func ShowMigrationChanges(logger *log.Logger, old, new *config.Config) {
	oldLinters := len(old.Linters.Enable)
	newLinters := len(new.Linters.Enable)

	if old.Version != new.Version {
		logger.Infof("  Version: %s -> %s", old.Version, new.Version)
	}

	if oldLinters != newLinters {
		logger.Infof("  Linters: %d -> %d", oldLinters, newLinters)
	}

	// Check for renamed fields (basic check)
	if old.Run.Timeout != new.Run.Timeout {
		logger.Infof("  Timeout: %s -> %s", old.Run.Timeout, new.Run.Timeout)
	}
}
