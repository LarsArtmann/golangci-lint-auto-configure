package cmd

import (
	"context"
	"fmt"

	"charm.land/log/v2"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/audit"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/migration"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// tracerName names the OpenTelemetry tracer for this package. The tracer is
// the noop implementation unless the CLI registers a provider via
// pkg/telemetry, so the migrate span only exists for --otel runs.
const tracerName = "github.com/larsartmann/golangci-lint-auto-configure/internal/cli/cmd"

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
) *cobra.Command {
	var skipValidation bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate configuration from v1 to v2 schema",
		Long:  migrateLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMigrate(cmd, logger, configLoader, skipValidation)
		},
	}

	cmd.Flags().
		BoolVar(&skipValidation, "skip-validation", false, "Skip validation of v1 configuration")

	return cmd
}

func runMigrate(
	cmd *cobra.Command,
	logger *log.Logger,
	configLoader *config.Loader,
	skipValidation bool,
) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	configFile, err := resolveMigrateConfig(cmd, configLoader, verbose)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "migrate.resolve_config",
			"resolve migrate config failed (verbose=%t)", verbose)
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")

	return executeMigration(
		cmd.Context(),
		logger,
		configLoader,
		configFile,
		dryRun,
		skipValidation,
		verbose,
	)
}

func resolveMigrateConfig(
	cmd *cobra.Command,
	configLoader *config.Loader,
	verbose bool,
) (string, error) {
	configFile, _ := cmd.Flags().GetString("config")
	if configFile == "" {
		var err error

		configFile, err = configLoader.FindConfigFile(".")
		if err != nil {
			return "", apperrors.WrapClassifiedf(err, "migrate.find_config",
				"no config file found (verbose=%t)", verbose)
		}
	}

	configLoader.HasMultipleConfigFiles(".")

	return configFile, nil
}

func executeMigration(
	ctx context.Context,
	logger *log.Logger,
	configLoader *config.Loader,
	configFile string,
	dryRun, skipValidation, verbose bool,
) error {
	oldConfig, err := loadConfigForMigration(logger, configLoader, configFile)
	if err != nil {
		return apperrors.WrapClassifiedf(err, "migrate.load_config",
			"load config failed (configFile=%s, dryRun=%t, skipValidation=%t, verbose=%t)",
			configFile, dryRun, skipValidation, verbose)
	}

	if isAlreadyV2(oldConfig) {
		logger.Infof("Configuration is already version 2, no migration needed")

		return nil
	}

	return runMigrator(
		ctx,
		logger,
		configLoader,
		configFile,
		dryRun,
		skipValidation,
		verbose,
		oldConfig,
	)
}

func loadConfigForMigration(
	logger *log.Logger,
	configLoader *config.Loader,
	configFile string,
) (*types.Config, error) {
	logger.Infof("Migrating configuration: %s", configFile)

	oldConfig, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return nil, apperrors.WrapClassifiedf(err, "migrate.load_for_migration",
			"could not load config %s", configFile)
	}

	return oldConfig, nil
}

func isAlreadyV2(cfg *types.Config) bool {
	return cfg.Version == "2"
}

func runMigrator(
	ctx context.Context,
	logger *log.Logger,
	configLoader *config.Loader,
	configFile string,
	dryRun, skipValidation, verbose bool,
	oldConfig *types.Config,
) error {
	span := startMigrateSpan(ctx, configFile, dryRun, skipValidation)
	defer span.End()

	//nolint:contextcheck // MigrateToV2 predates context threading (pkg/migration)
	success, fixesApplied, err := executeMigrator(
		logger, configFile, dryRun, skipValidation, verbose,
	)
	if err != nil {
		return err
	}

	recordMigrationOutcome(ctx, logger, span, configFile, dryRun, success, fixesApplied)

	showMigrationResult(logger, configLoader, configFile, oldConfig, success, fixesApplied, dryRun)

	return nil
}

// startMigrateSpan starts the migrate engine span with its fixed attributes.
// The child span context is discarded: this span is a leaf around MigrateToV2.
//
//nolint:spancheck,ireturn // the span deliberately outlives this helper: runMigrator ends it via defer
func startMigrateSpan(
	ctx context.Context,
	configFile string,
	dryRun, skipValidation bool,
) trace.Span {
	_, span := otel.Tracer(tracerName).Start(ctx, "migrate")

	span.SetAttributes(
		attribute.String("config.path", configFile),
		attribute.Bool("dry_run", dryRun),
		attribute.Bool("skip_validation", skipValidation),
	)

	return span
}

// executeMigrator builds the migrator and runs the v1→v2 rewrite.
func executeMigrator(
	logger *log.Logger,
	configFile string,
	dryRun, skipValidation, verbose bool,
) (bool, int, error) {
	migrator, err := createMigrator(configFile, dryRun, skipValidation, verbose, logger)
	if err != nil {
		return false, 0, apperrors.WrapClassified(err, "migrate.create_migrator",
			"create migrator failed")
	}

	success, fixesApplied, err := migrator.MigrateToV2()
	if err != nil {
		return false, 0, apperrors.WrapClassified(err, "migrate.migrate", "migration failed")
	}

	return success, fixesApplied, nil
}

// recordMigrationOutcome stamps the migrate span's result attrs and writes
// the config-migrated ledger entry for non-dry successes. Dry runs record
// nothing, matching the configure contract. The
// GOLANGCI_LINT_AUTO_CONFIGURE_NO_AUDIT env var disables the entry (migrate
// binds no --no-audit flag of its own).
func recordMigrationOutcome(
	ctx context.Context,
	logger *log.Logger,
	span trace.Span,
	configFile string,
	dryRun, success bool,
	fixesApplied int,
) {
	if !success {
		return
	}

	span.SetAttributes(attribute.Int("fixes.applied", fixesApplied))

	if dryRun {
		return
	}

	audit.OpenRunLedger(ctx, logger, configFile, false).Record(
		audit.ActionConfigMigrated, "config",
		fmt.Sprintf("migrated v1 config to v2 schema (%d fixes applied)", fixesApplied),
	)
}

func createMigrator(
	configFile string,
	dryRun, skipValidation, verbose bool,
	logger *log.Logger,
) (*migration.Migrator, error) {
	migrator, err := migration.NewMigrator(configFile, verbose)
	if err != nil {
		return nil, apperrors.WrapClassifiedf(err, "migrate.create_migrator",
			"failed to create migrator (configFile=%s, dryRun=%t, skipValidation=%t, verbose=%t)",
			configFile, dryRun, skipValidation, verbose)
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
	oldConfig *types.Config,
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
func ShowMigrationChanges(logger *log.Logger, oldCfg, newCfg *types.Config) {
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
