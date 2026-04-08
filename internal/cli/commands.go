package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"charm.land/log/v2"
	"github.com/charmbracelet/fang"
	clicmd "github.com/larsartmann/golangci-lint-auto-configure/internal/cli/cmd"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/spf13/cobra"
)

// Version is set by main package via ldflags.
var Version = "dev"

var (
	configPath   string
	dryRun       bool
	verbose      bool
	generateHTML bool
	outputReport string
	priority     string
	reportFormat string
	noAutoMerge  bool
)

// resolveConfigPath finds the config file if not specified, with multiple config warning.
func resolveConfigPath(
	_ context.Context,
	configLoader *config.Loader,
	logger *log.Logger,
	specifiedPath string,
	isDryRun bool,
) (string, error) {
	configFile := specifiedPath
	if configFile == "" {
		var err error

		configFile, err = configLoader.FindConfigFile(".")
		if err != nil {
			return "", fmt.Errorf("failed to find config file: %w", err)
		}
	}

	return resolveWithAutoMerge(configLoader, logger, configFile, isDryRun)
}

func resolveWithAutoMerge(
	configLoader *config.Loader,
	logger *log.Logger,
	configFile string,
	isDryRun bool,
) (string, error) {
	allConfigs := configLoader.FindAllConfigFiles(".")

	if len(allConfigs) <= 1 || noAutoMerge {
		if len(allConfigs) > 1 {
			logger.Warnf("⚠️  Multiple config files detected but auto-merge is disabled")
			logger.Warnf("   Using primary config: %s", allConfigs[0])
		}

		return configFile, nil
	}

	merger := config.NewMerger(logger)
	logger.Infof("🔄 Auto-merging %d config files...", len(allConfigs))

	return performAutoMerge(merger, logger, allConfigs, isDryRun, configFile)
}

func performAutoMerge(
	merger *config.Merger,
	logger *log.Logger,
	allConfigs []string,
	isDryRun bool,
	fallbackConfig string,
) (string, error) {
	mergedConfig, mergeResult, err := merger.MergeConfigs(allConfigs)
	if err != nil {
		logger.Warnf("⚠️  Failed to merge configs: %v", err)
		logger.Warnf("   Continuing with primary config: %s", allConfigs[0])

		return fallbackConfig, nil
	}

	logger.Infof("✅ Merged configs (primary: %s)", mergeResult.PrimaryConfig)

	if mergeResult.ChangesApplied > 0 {
		logger.Infof("   Applied %d configuration changes", mergeResult.ChangesApplied)
	}

	if isDryRun {
		logger.Infof("🔍 Dry-run mode: would save merged config to %s", mergeResult.PrimaryConfig)
		logger.Infof("   Secondary configs would be removed: %v", mergeResult.MergedConfigs)

		return mergeResult.PrimaryConfig, nil
	}

	return saveMergedConfigAndReturn(merger, mergedConfig, mergeResult, allConfigs, logger)
}

func saveMergedConfigAndReturn(
	merger *config.Merger,
	mergedConfig *config.Config,
	mergeResult *config.MergeResult,
	allConfigs []string,
	logger *log.Logger,
) (string, error) {
	err := merger.SaveMergedConfig(mergedConfig, mergeResult, true)
	if err != nil {
		logger.Warnf("⚠️  Failed to save merged config: %v", err)

		return allConfigs[0], nil
	}

	logger.Infof("💾 Saved merged config to: %s", mergeResult.PrimaryConfig)

	if len(mergeResult.RemovedConfigs) > 0 {
		logger.Infof("🗑️  Removed secondary configs: %v", mergeResult.RemovedConfigs)
	}

	return mergeResult.PrimaryConfig, nil
}

// NewRootCommand creates the root CLI command.
func NewRootCommand() *cobra.Command {
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller: false,
		TimeFormat:   "15:04:05",
		Level:        log.InfoLevel,
	})
	slog.SetDefault(slog.New(logger))

	rootCmd := &cobra.Command{
		Use:   "golangci-lint-auto-configure",
		Short: "Automatically configure and optimize golangci-lint",
		Long: `A tool that automatically analyzes golangci-lint configurations,
detects missing linters with smart categorization, and provides
actionable recommendations to improve your Go code quality.`,
		Version: Version,
	}

	analyzer := linter.NewAnalyzer(logger)
	configLoader := config.NewLoader(logger)

	migrateFlags := clicmd.MigrateFlags{
		ConfigPath: configPath,
		DryRun:     dryRun,
		Verbose:    verbose,
	}

	addSubCommands(rootCmd, logger, analyzer, configLoader, migrateFlags)
	registerGlobalFlags(rootCmd)

	return rootCmd
}

func addSubCommands(
	rootCmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	migrateFlags clicmd.MigrateFlags,
) {
	rootCmd.AddCommand(
		newConfigureCommand(logger, analyzer, configLoader),
		newAnalyzeCommand(logger, analyzer, configLoader),
		clicmd.NewMigrateCommand(logger, configLoader, migrateFlags),
		newValidateCommand(logger, configLoader),
		newReportCommand(logger, analyzer, configLoader),
		clicmd.NewCompletionCommand(),
		clicmd.NewInstallHookCommand(logger),
	)
}

func registerGlobalFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().
		StringVarP(&configPath, "config", "c", "", "Path to golangci-lint config file")
	rootCmd.PersistentFlags().
		BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&generateHTML, "html", false, "Generate HTML report")
	rootCmd.PersistentFlags().
		StringVar(&outputReport, "output", "report.html", "Output path for HTML report")
	rootCmd.PersistentFlags().
		StringVar(&priority, "priority", "high", "Minimum priority level to enable (critical, high, medium, optional)")
	rootCmd.PersistentFlags().
		StringVar(&reportFormat, "format", "html", "Output format (html, json)")
	rootCmd.PersistentFlags().
		BoolVar(&noAutoMerge, "no-auto-merge", false, "Disable automatic merging of multiple config files")
}

// Execute runs the CLI using fang for enhanced CLI features.
func Execute(ctx context.Context) error {
	//nolint:contextcheck // Context is passed through fang.Execute; linter doesn't trace third-party calls
	rootCmd := NewRootCommand()

	err := fang.Execute(ctx, rootCmd, fang.WithVersion(Version))
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

// Main is the entry point.
func Main() {
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller: false,
		TimeFormat:   "15:04:05",
		Level:        log.InfoLevel,
	})
	slog.SetDefault(slog.New(logger))

	err := Execute(context.Background())
	if err != nil {
		slog.Error("CLI execution failed", "error", err)
		os.Exit(1)
	}
}
