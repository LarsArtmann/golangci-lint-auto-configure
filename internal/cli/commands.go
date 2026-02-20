package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	clicmd "github.com/larsartmann/golangcli-linter-auto-configure/internal/cli/cmd"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/report"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/workflow"
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
)

// runConfigure executes the configure command logic.
func runConfigure(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
	priorityParam, preset string,
	dryRun bool,
	configPath string,
) error {
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	// Find config file if not specified, or use default path
	configFile := configPath
	if configFile == "" {
		configFile = configLoader.FindOrGetDefaultConfigPath(".")
	}

	// Check if config file exists, create default if not
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		logger.Infof("No config file found, creating default: %s", configFile)

		defaultConfig := configLoader.CreateDefaultConfig()

		err := configLoader.SaveConfig(defaultConfig, configFile)
		if err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check config file: %w", err)
	}

	logger.Infof("Configuring golangci-lint with config: %s", configFile)

	// Handle preset mode
	if preset != "" {
		return applyPreset(logger, configLoader, configFile, preset, dryRun)
	}

	// Create fixer and apply fixes
	fixer := linter.NewFixer(logger, analyzer)

	var linterPriority types.LinterPriority

	switch priorityParam {
	case "critical":
		linterPriority = types.LinterPriorityCritical
	case "high":
		linterPriority = types.LinterPriorityHigh
	case "medium":
		linterPriority = types.LinterPriorityMedium
	case "optional":
		linterPriority = types.LinterPriorityOptional
	default:
		linterPriority = types.LinterPriorityHigh
	}

	result, err := fixer.FixConfig(configFile, linterPriority, dryRun)
	if err != nil {
		return fmt.Errorf("failed to fix configuration: %w", err)
	}

	logger.Infof("%s", result.Message)

	if result.BackupPath != "" {
		logger.Infof("Backup created: %s", result.BackupPath)
	}

	return nil
}

// NewRootCommand creates the root CLI command.
func NewRootCommand() *cobra.Command {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller: false,
		TimeFormat:   "15:04:05",
		Level:        log.InfoLevel,
	})

	rootCmd := &cobra.Command{
		Use:   "golangci-linter-auto-configure",
		Short: "Automatically configure and optimize golangci-lint",
		Long: `A tool that automatically analyzes golangci-lint configurations,
detects missing linters with smart categorization, and provides
actionable recommendations to improve your Go code quality.`,
		Version: Version,
	}

	// Create analyzer and workflow builder
	analyzer := linter.NewAnalyzer(logger)
	workflowBuilder := workflow.NewBuilder(logger, analyzer)

	// Configure root command
	configLoader := config.NewLoader(logger)

	migrateFlags := clicmd.MigrateFlags{
		ConfigPath: configPath,
		DryRun:     dryRun,
		Verbose:    verbose,
	}

	rootCmd.AddCommand(
		newConfigureCommand(logger, analyzer, workflowBuilder, configLoader),
		newAnalyzeCommand(logger, analyzer, configLoader),
		clicmd.NewMigrateCommand(logger, configLoader, migrateFlags),
		newValidateCommand(logger, configLoader),
		newReportCommand(logger, analyzer, configLoader),
		newRestoreCommand(logger, configLoader),
		clicmd.NewCompletionCommand(),
		clicmd.NewInstallHookCommand(logger),
	)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to golangci-lint config file")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&generateHTML, "html", false, "Generate HTML report")
	rootCmd.PersistentFlags().StringVar(&outputReport, "output", "report.html", "Output path for HTML report")
	rootCmd.PersistentFlags().
		StringVar(&priority, "priority", "high", "Minimum priority level to enable (critical, high, medium, optional)")
	rootCmd.PersistentFlags().StringVar(&reportFormat, "format", "html", "Output format (html, json)")

	return rootCmd
}

// newConfigureCommand creates the configure command.
func newConfigureCommand(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	workflowBuilder *workflow.Builder,
	configLoader *config.Loader,
) *cobra.Command {
	var preset string

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Auto-configure golangci-lint (default command)",
		Long: `Automatically configures golangci-lint by enabling recommended linters.

Use --priority to filter which linters to enable:
  - critical: Only enable critical linters (security, correctness)
  - high: Enable critical and high-value linters (recommended)
  - medium: Enable all except optional linters
  - optional: Enable all linters (may be too strict)

Or use --preset for predefined linter sets:
  - minimal: Essential linters only (fastest)
  - standard: Recommended for most projects (default)
  - strict: Maximum linting (CI/CD, strict quality)
  - security: Security-focused only
  - performance: Performance optimization only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigure(logger, analyzer, configLoader, priority, preset, dryRun, configPath)
		},
	}

	cmd.Flags().
		StringVar(&priority, "priority", "high", "Minimum priority level to enable (critical, high, medium, optional)")
	cmd.Flags().
		StringVar(&preset, "preset", "", "Use a preset linter set (minimal, standard, strict, security, performance)")

	return cmd
}

// applyPreset applies a preset linter configuration.
func applyPreset(logger *log.Logger, configLoader *config.Loader, configFile, preset string, dryRun bool) error {
	logger.Infof("Applying preset: %s", preset)

	// Load current config
	cfg, err := configLoader.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get preset linters
	linters, ok := constants.PresetLinters[preset]
	if !ok {
		return fmt.Errorf("unknown preset: %s (valid: minimal, standard, strict, security, performance)", preset)
	}

	// Convert to strings
	var linterNames []string
	for _, l := range linters {
		linterNames = append(linterNames, string(l))
	}

	if dryRun {
		logger.Infof("[DRY-RUN] Would apply preset %s with %d linters:", preset, len(linterNames))

		for _, l := range linterNames {
			logger.Infof("  - %s", l)
		}

		return nil
	}

	// Create backup
	backupPath, err := configLoader.CreateBackup(configFile)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Update config
	cfg.Linters.Enable = linterNames
	cfg.Linters.Disable = []string{}

	// Save config
	if err := configLoader.SaveConfig(cfg, configFile); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	logger.Infof("✅ Applied preset %s with %d linters", preset, len(linterNames))
	logger.Infof("Backup created: %s", backupPath)

	return nil
}

// newAnalyzeCommand creates the analyze command.
func newAnalyzeCommand(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze golangci-lint configuration and show recommendations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Find config file if not specified
			configFile := configPath
			if configFile == "" {
				var err error

				configFile, err = configLoader.FindConfigFile(".")
				if err != nil {
					return fmt.Errorf("no config file found: %w", err)
				}
			}

			logger.Infof("Analyzing configuration: %s", configFile)

			// Perform analysis
			analysis, err := analyzer.AnalyzeConfig(configFile)
			if err != nil {
				return fmt.Errorf("analysis failed: %w", err)
			}

			// Display recommendations
			recommendations := analyzer.FormatRecommendations(analysis)
			summary := analyzer.GetSummary(analysis)

			logger.Info("\n" + recommendations)
			logger.Infof("Summary: %s", summary)

			return nil
		},
	}

	return cmd
}

// newValidateCommand creates the validate command.
func newValidateCommand(
	logger *log.Logger,
	configLoader *config.Loader,
) *cobra.Command {
	var skipGolangciLint bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate golangci-lint configuration",
		Long: `Validates the golangci-lint configuration file.

This command performs two levels of validation:
1. Basic YAML parsing and structure validation
2. Schema validation using golangci-lint config verify

Use --skip-golangci-lint to skip the schema validation (faster).
Use --verbose to see detailed validation output.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Find config file if not specified
			configFile := configPath
			if configFile == "" {
				var err error

				configFile, err = configLoader.FindConfigFile(".")
				if err != nil {
					return fmt.Errorf("no config file found: %w", err)
				}
			}

			logger.Infof("Validating configuration: %s", configFile)

			// Level 1: Basic load and validate
			cfg, err := configLoader.LoadConfig(configFile)
			if err != nil {
				logger.Errorf("❌ Failed to load configuration")

				return fmt.Errorf("failed to load config: %w", err)
			}

			logger.Infof("✓ Basic structure valid")

			// Run internal validation
			validationErrors := configLoader.ValidateConfig(cfg)
			if len(validationErrors) > 0 {
				logger.Errorf("❌ Internal validation failed:")

				for _, err := range validationErrors {
					logger.Errorf("  - %v", err)
				}

				return fmt.Errorf("configuration has %d validation errors", len(validationErrors))
			}

			logger.Infof("✓ Internal validation passed")

			// Level 2: Schema validation via golangci-lint
			if !skipGolangciLint {
				logger.Infof("Running golangci-lint schema validation...")

				verifyCmd := exec.Command("golangci-lint", "config", "verify", "--config", configFile)

				output, err := verifyCmd.CombinedOutput()
				if err != nil {
					logger.Errorf("❌ Schema validation failed:")
					logger.Errorf("%s", string(output))

					return fmt.Errorf("schema validation failed: %w", err)
				}

				if len(output) > 0 {
					logger.Infof("%s", string(output))
				}

				logger.Infof("✓ Schema validation passed")
			}

			logger.Infof("✅ Configuration is valid")

			return nil
		},
	}

	cmd.Flags().BoolVar(&skipGolangciLint, "skip-golangci-lint", false, "Skip golangci-lint schema validation")

	return cmd
}

// newReportCommand creates the report command.
func newReportCommand(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate HTML report of configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Find config file if not specified
			configFile := configPath
			if configFile == "" {
				var err error

				configFile, err = configLoader.FindConfigFile(".")
				if err != nil {
					return fmt.Errorf("no config file found: %w", err)
				}
			}

			logger.Infof("Generating %s report for: %s", reportFormat, configFile)

			// Analyze configuration
			analysis, err := analyzer.AnalyzeConfig(configFile)
			if err != nil {
				return fmt.Errorf("analysis failed: %w", err)
			}

			// Determine output path
			outputPath := outputReport
			if outputPath == "report.html" {
				if reportFormat == "json" {
					outputPath = "report.json"
				}
			}

			// Generate report based on format
			if reportFormat == "json" {
				jsonGenerator := report.NewJSONGenerator(logger)

				err := jsonGenerator.GenerateJSONReport(analysis, outputPath)
				if err != nil {
					return fmt.Errorf("failed to generate JSON report: %w", err)
				}
			} else {
				htmlGenerator := report.NewGenerator(logger)

				err := htmlGenerator.GenerateReport(cmd.Context(), analysis, outputPath)
				if err != nil {
					return fmt.Errorf("failed to generate HTML report: %w", err)
				}
			}

			return nil
		},
	}

	return cmd
}

// newRestoreCommand creates the restore command.
func newRestoreCommand(
	logger *log.Logger,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore configuration from backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Get backup path from flag or argument
			backupPath, _ := cmd.Flags().GetString("backup-path")
			if backupPath == "" && len(args) > 0 {
				return errors.New("backup path required (use --backup-path flag or provide as argument)")
			}

			if backupPath == "" {
				backupPath = args[0]
			}

			logger.Infof("Restoring configuration from: %s", backupPath)

			// Check if backup file exists
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				return fmt.Errorf("backup file not found: %s", backupPath)
			}

			// Determine target config path
			targetPath := configPath
			if targetPath == "" {
				var err error

				targetPath, err = configLoader.FindConfigFile(".")
				if err != nil {
					return fmt.Errorf("no config file found to restore to: %w", err)
				}
			}

			logger.Infof("Restoring to: %s", targetPath)

			// Perform restore
			err := configLoader.RestoreConfig(backupPath, targetPath)
			if err != nil {
				return fmt.Errorf("failed to restore configuration: %w", err)
			}

			logger.Infof("Configuration restored successfully")

			return nil
		},
	}

	cmd.Flags().String("backup-path", "", "Path to backup file to restore from")

	return cmd
}

// Execute runs the CLI using fang for enhanced CLI features.
func Execute(ctx context.Context) error {
	rootCmd := NewRootCommand()

	return fang.Execute(ctx, rootCmd, fang.WithVersion(Version))
}

// Main is the entry point.
func Main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	err := Execute(context.Background())
	if err != nil {
		slog.Error("CLI execution failed", "error", err)
		os.Exit(1)
	}
}
