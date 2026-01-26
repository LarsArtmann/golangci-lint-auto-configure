package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/report"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/workflow"
	"github.com/spf13/cobra"
)

var (
	configPath   string
	dryRun       bool
	verbose      bool
	generateHTML bool
	outputReport string
	priority     string
	reportFormat string
)

// NewRootCommand creates the root CLI command
func NewRootCommand() *cobra.Command {
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller: false,
		TimeFormat:   "15:04:05",
		Level:        log.InfoLevel,
	})

	cmd := &cobra.Command{
		Use:   "golangci-linter-auto-configure",
		Short: "Automatically configure and optimize golangci-lint",
		Long: `A tool that automatically analyzes golangci-lint configurations,
detects missing linters with smart categorization, and provides
actionable recommendations to improve your Go code quality.`,
	}

	// Create analyzer and workflow builder
	analyzer := linter.NewAnalyzer(logger)
	workflowBuilder := workflow.NewBuilder(logger, analyzer)

	// Configure root command
	configLoader := config.NewLoader(logger)

	cmd.AddCommand(
		newConfigureCommand(logger, analyzer, workflowBuilder, configLoader),
		newAnalyzeCommand(logger, analyzer, configLoader),
		newMigrateCommand(logger, configLoader),
		newValidateCommand(logger, configLoader),
		newReportCommand(logger, analyzer, configLoader),
		newRestoreCommand(logger, configLoader),
	)

	// Global flags
	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to golangci-lint config file")
	cmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be done without making changes")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&generateHTML, "html", false, "Generate HTML report")
	cmd.PersistentFlags().StringVar(&outputReport, "output", "report.html", "Output path for HTML report")
	cmd.PersistentFlags().StringVar(&priority, "priority", "high", "Minimum priority level to enable (critical, high, medium, optional)")
	cmd.PersistentFlags().StringVar(&reportFormat, "format", "html", "Output format (html, json)")

	return cmd
}

// newConfigureCommand creates the configure command
func newConfigureCommand(
	logger *log.Logger,
	analyzer *linter.Analyzer,
	workflowBuilder *workflow.Builder,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Auto-configure golangci-lint (default command)",
		Long: `Automatically configures golangci-lint by enabling recommended linters.
Use --priority to filter which linters to enable:
  - critical: Only enable critical linters (security, correctness)
  - high: Enable critical and high-value linters (recommended)
  - medium: Enable all except optional linters
  - optional: Enable all linters (may be too strict)`,
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

			logger.Infof("Configuring golangci-lint with config: %s", configFile)

			// Create fixer and apply fixes
			fixer := linter.NewFixer(logger, analyzer)

			var linterPriority types.LinterPriority
			switch priority {
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
		},
	}

	cmd.Flags().StringVar(&priority, "priority", "high", "Minimum priority level to enable (critical, high, medium, optional)")

	return cmd
}

// newAnalyzeCommand creates the analyze command
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

// newMigrateCommand creates the migrate command
func newMigrateCommand(
	logger *log.Logger,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate configuration to golangci-lint v2.8+ schema",
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

			logger.Infof("Migrating configuration: %s", configFile)

			// Migration would be implemented here
			// For now, this is a placeholder
			logger.Warnf("Migration functionality not yet implemented")

			return nil
		},
	}

	return cmd
}

// newValidateCommand creates the validate command
func newValidateCommand(
	logger *log.Logger,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate golangci-lint configuration",
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

			// Load and validate config
			cfg, err := configLoader.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			errors := configLoader.ValidateConfig(cfg)
			if len(errors) > 0 {
				logger.Errorf("Configuration validation failed:")
				for _, err := range errors {
					logger.Errorf("  - %v", err)
				}
				return fmt.Errorf("configuration has %d validation errors", len(errors))
			}

			logger.Infof("Configuration is valid")

			return nil
		},
	}

	return cmd
}

// newReportCommand creates the report command
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
				if err := jsonGenerator.GenerateJSONReport(analysis, outputPath); err != nil {
					return fmt.Errorf("failed to generate JSON report: %w", err)
				}
			} else {
				htmlGenerator := report.NewGenerator(logger)
				if err := htmlGenerator.GenerateReport(analysis, outputPath); err != nil {
					return fmt.Errorf("failed to generate HTML report: %w", err)
				}
			}

			return nil
		},
	}

	return cmd
}

// newRestoreCommand creates the restore command
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
				return fmt.Errorf("backup path required (use --backup-path flag or provide as argument)")
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
			if err := configLoader.RestoreConfig(backupPath, targetPath); err != nil {
				return fmt.Errorf("failed to restore configuration: %w", err)
			}

			logger.Infof("Configuration restored successfully")

			return nil
		},
	}

	cmd.Flags().String("backup-path", "", "Path to backup file to restore from")

	return cmd
}

// Execute runs the CLI using fang for enhanced CLI features
func Execute() error {
	cmd := NewRootCommand()
	return fang.Execute(context.Background(), cmd)
}

// Main is the entry point
func Main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
