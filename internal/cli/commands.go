package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"charm.land/fang/v2"
	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	clicmd "github.com/larsartmann/golangci-lint-auto-configure/internal/cli/cmd"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/config"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/linter"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/version"
	"github.com/spf13/cobra"
)

// Version is the CLI version string, derived from pkg/version.
var Version = version.Get().Short()

var (
	configPath   string
	dryRun       bool
	verbose      bool
	quiet        bool
	outputReport string
	priority     string
	reportFormat string
	noAutoMerge  bool
	showDiff     bool
	jsonErrors   bool
	noColor      bool
	noAudit      bool
	pragmatic    bool
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
		configFile = configLoader.FindOrGetDefaultConfigPath(".")
	}

	return resolveWithAutoMerge(configLoader, logger, configFile, isDryRun)
}

// resolveConfig is a helper that resolves config path with a custom error prefix.
func resolveConfig(
	ctx context.Context,
	configLoader *config.Loader,
	logger *log.Logger,
	errorPrefix string,
) (string, error) {
	configFile, err := resolveConfigPath(ctx, configLoader, logger, configPath, dryRun)
	if err != nil {
		return "", apperrors.WrapClassifiedf(err, "cli.resolve_config_path",
			"%s", errorPrefix)
	}

	return configFile, nil
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
	mergedConfig *types.Config,
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

	if len(mergeResult.BackedUpConfigs) > 0 {
		logger.Infof("💾 Backups created:")

		for original, backup := range mergeResult.BackedUpConfigs {
			logger.Infof("   %s -> %s", original, backup)
		}
	}

	if len(mergeResult.RemovedConfigs) > 0 {
		logger.Infof("🗑️  Removed secondary configs: %v", mergeResult.RemovedConfigs)
	}

	return mergeResult.PrimaryConfig, nil
}

// newLogger creates a standard logger for CLI output.
func newLogger() *log.Logger {
	return log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller: false,
		TimeFormat:   "15:04:05",
		Level:        log.InfoLevel,
	})
}

// NewRootCommand creates the root CLI command.
func NewRootCommand() *cobra.Command {
	logger := newLogger()
	slog.SetDefault(slog.New(logger))

	rootCmd := &cobra.Command{
		Use:   constants.ToolName,
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

	rootCmd.PersistentPreRun = makeLogLevelConfigurer(logger)

	addSubCommands(rootCmd, logger, analyzer, configLoader, migrateFlags)
	registerGlobalFlags(rootCmd)

	rootCmd.SetVersionTemplate(version.Get().Full() + "\n")

	return rootCmd
}

// makeLogLevelConfigurer returns a PersistentPreRun that adjusts the logger level
// based on --quiet and --verbose flags.
func makeLogLevelConfigurer(logger *log.Logger) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		if noColor {
			_ = os.Setenv("NO_COLOR", "1")
		}

		switch {
		case quiet:
			logger.SetLevel(log.ErrorLevel)
		case verbose:
			logger.SetLevel(log.DebugLevel)
		}
	}
}

func addSubCommands(
	rootCmd *cobra.Command,
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	migrateFlags clicmd.MigrateFlags,
) {
	builder := NewCommandBuilder(logger, analyzer, configLoader)

	rootCmd.AddCommand(
		newConfigureCommand(builder),
		newAnalyzeCommand(builder),
		clicmd.NewMigrateCommand(logger, configLoader, migrateFlags),
		newValidateCommand(builder),
		newReportCommand(builder),
		newPresetsCommand(builder),
		newAuditCommand(builder),
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
	rootCmd.PersistentFlags().
		StringVar(&outputReport, "output", "report.html", "Output path for HTML report")
	rootCmd.PersistentFlags().
		StringVar(&priority, "priority", "high", "Minimum priority level to enable (critical, high, medium, optional)")
	rootCmd.PersistentFlags().
		StringVar(&reportFormat, "format", "html", "Output format (html, json, sarif, finding)")
	rootCmd.PersistentFlags().
		BoolVar(&noAutoMerge, "no-auto-merge", false, "Disable automatic merging of multiple config files")
	rootCmd.PersistentFlags().
		BoolVar(&quiet, "quiet", false, "Suppress all output except errors (useful for CI)")
	rootCmd.PersistentFlags().
		BoolVar(&showDiff, "diff", false, "Show diff of config changes before applying")
	rootCmd.PersistentFlags().
		BoolVar(&jsonErrors, "json-errors", false, "Output errors as JSON to stderr for programmatic consumption")
	rootCmd.PersistentFlags().
		BoolVar(&noColor, "no-color", false, "Disable colored output (also honored via NO_COLOR env var)")
}

// Execute runs the CLI using fang for enhanced CLI features.
func Execute(ctx context.Context) error {
	//nolint:contextcheck // Context is passed through fang.Execute; linter doesn't trace third-party calls
	rootCmd := NewRootCommand()

	err := fang.Execute(ctx, rootCmd, fang.WithVersion(Version))
	if err != nil {
		return apperrors.WrapClassified(err, "cli.execute", "failed to execute command")
	}

	return nil
}

// Main is the entry point.
func Main() {
	if err := Execute(context.Background()); err != nil {
		os.Exit(HandleError(err))
	}
}

// HandleError classifies an error, outputs it in the appropriate format
// (JSON via --json-errors, or a user-friendly message with structured slog
// fallback), and returns the BSD sysexits exit code. Called at the CLI
// boundary for all unhandled errors.
func HandleError(err error) int {
	// Check for explicit CommandResult with a user-provided exit code/message
	if result := extractCommandResult(err); result != nil {
		return handleCommandResult(result)
	}

	family := errorfamily.Classify(err)
	exitCode := errorfamily.ExitCode(err)

	if jsonErrors {
		outputJSONError(err, family, exitCode)
	} else if rendered := renderUserError(err); rendered != "" {
		fmt.Fprintln(os.Stderr, rendered)
	} else {
		slog.Error("CLI execution failed", "error", err, "family", family.String())
	}

	return exitCode
}

// handleCommandResult processes a *CommandResult, printing its message and
// returning its exit code. Success results print to stdout; error results
// delegate to the standard error handling path.
func handleCommandResult(result *CommandResult) int {
	if result.IsSuccess() {
		if result.Message() != "" {
			fmt.Fprintln(os.Stdout, result.Message())
		}

		return 0
	}

	return result.ExitCode()
}

// renderUserError resolves a domain MessageTemplate for the error and renders
// a user-friendly What/Fix message. Returns "" when no template matches,
// signaling the caller to fall back to slog.
func renderUserError(err error) string {
	code := errorfamily.Code(err)
	if code == "" {
		return ""
	}

	tmpl, ok := errorfamily.TemplateForCode(code)
	if !ok || tmpl.What == "" {
		return ""
	}

	msg := "❌ " + tmpl.What
	if tmpl.Fix != "" {
		msg += "\n   → " + tmpl.Fix
	}

	return msg
}

func outputJSONError(err error, family errorfamily.Family, exitCode int) {
	classified := errorfamily.Wrap(err, family, "cli.execution_failed", err.Error()).
		WithContext("exit_code", strconv.Itoa(exitCode))

	data, marshalErr := classified.JSON()
	if marshalErr != nil {
		slog.Error("failed to marshal JSON error", "error", marshalErr)

		return
	}

	fmt.Fprintln(os.Stderr, string(data))
}
