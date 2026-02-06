package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/report"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/workflow"
	"github.com/spf13/cobra"
)

// Version is set by main package via ldflags
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

// NewRootCommand creates the root CLI command
func NewRootCommand() *cobra.Command {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

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
		Version: Version,
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
		newCompletionCommand(),
		newInstallHookCommand(logger),
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

			// Find config file if not specified, or use default path
			configFile := configPath
			if configFile == "" {
				configFile = configLoader.FindOrGetDefaultConfigPath(".")
			}

			// Check if config file exists, create default if not
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				logger.Infof("No config file found, creating default: %s", configFile)
				defaultConfig := configLoader.CreateDefaultConfig()
				if err := configLoader.SaveConfig(defaultConfig, configFile); err != nil {
					return fmt.Errorf("failed to create default config: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("failed to check config file: %w", err)
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
	var skipValidation bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate configuration from v1 to v2 schema",
		Long: `Migrates golangci-lint configuration from v1 to v2 schema.

This command:
1. Creates a backup of your current configuration
2. Runs golangci-lint migrate to convert the schema
3. Validates the migrated configuration
4. Shows what changed

Use --dry-run to preview changes without modifying files.
Use --skip-validation if the v1 config has known issues.`,
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

			// Create backup first
			backupPath := configFile + ".v1-backup"
			if !dryRun {
				logger.Infof("Creating backup: %s", backupPath)
				backupCreated, err := configLoader.CreateBackup(configFile)
				if err != nil {
					return fmt.Errorf("failed to create backup: %w", err)
				}
				// Rename to .v1-backup for clarity
				if err := os.Rename(backupCreated, backupPath); err != nil {
					logger.Warnf("Could not rename backup to %s: %v", backupPath, err)
					backupPath = backupCreated
				}
			} else {
				logger.Infof("[DRY-RUN] Would create backup: %s", backupPath)
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
				// Restore backup on failure
				logger.Errorf("Migration failed: %v", err)
				logger.Infof("Output: %s", string(output))
				logger.Infof("Restoring backup...")
				if restoreErr := configLoader.RestoreConfig(backupPath, configFile); restoreErr != nil {
					logger.Errorf("Failed to restore backup: %v", restoreErr)
					return fmt.Errorf("migration failed and restore failed: %w (restore error: %v)", err, restoreErr)
				}
				return fmt.Errorf("migration failed, backup restored: %w", err)
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
				showMigrationChanges(logger, oldConfig, newConfig)
			}

			logger.Infof("Configuration migrated successfully!")
			logger.Infof("Backup saved to: %s", backupPath)
			logger.Infof("If you need to restore: golangci-linter-auto-configure restore %s", backupPath)

			return nil
		},
	}

	cmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "Skip validation of v1 configuration")
	cmd.Flags().StringVar(&outputFormat, "format", "", "Output format (yml, yaml, toml, json)")

	return cmd
}

// showMigrationChanges displays the differences between old and new config
func showMigrationChanges(logger *log.Logger, old, new *config.Config) {
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

// newValidateCommand creates the validate command
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

// newCompletionCommand creates the completion command for shell autocompletion
func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for golangci-linter-auto-configure.

To load completions:

Bash:
  $ source <(golangci-linter-auto-configure completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ golangci-linter-auto-configure completion bash > /etc/bash_completion.d/golangci-linter-auto-configure
  # macOS:
  $ golangci-linter-auto-configure completion bash > $(brew --prefix)/etc/bash_completion.d/golangci-linter-auto-configure

Zsh:
  $ source <(golangci-linter-auto-configure completion zsh)
  # To load completions for each session, execute once:
  $ golangci-linter-auto-configure completion zsh > "${fpath[1]}/_golangci-linter-auto-configure"

Fish:
  $ golangci-linter-auto-configure completion fish | source
  # To load completions for each session, execute once:
  $ golangci-linter-auto-configure completion fish > ~/.config/fish/completions/golangci-linter-auto-configure.fish

PowerShell:
  PS> golangci-linter-auto-configure completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> golangci-linter-auto-configure completion powershell > golangci-linter-auto-configure.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
		},
	}
}

// newInstallHookCommand creates the install-hook command
func newInstallHookCommand(logger *log.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "install-hook",
		Short: "Install pre-commit hook for git",
		Long: `Installs a pre-commit hook that runs golangci-linter-auto-configure before each commit.

The hook will:
1. Check if a .golangci.yml config exists
2. Run 'analyze' to check for configuration issues
3. Warn if optimizations are available

The hook is installed at .git/hooks/pre-commit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if we're in a git repository
			if _, err := os.Stat(".git"); os.IsNotExist(err) {
				return fmt.Errorf("not a git repository (no .git directory found)")
			}

			hookDir := ".git/hooks"
			hookPath := filepath.Join(hookDir, "pre-commit")

			// Create hook content
			hookContent := `#!/bin/bash
# Pre-commit hook for golangci-linter-auto-configure
# Auto-generated by golangci-linter-auto-configure install-hook

set -e

echo "Running golangci-linter-auto-configure pre-commit hook..."

# Check if golangci-linter-auto-configure is installed
if ! command -v golangci-linter-auto-configure &> /dev/null; then
    echo "Error: golangci-linter-auto-configure is not installed"
    echo "Install from: https://github.com/LarsArtmann/golangcli-linter-auto-configure"
    exit 1
fi

# Find config file
CONFIG_FILE=""
for file in ".golangci.yml" ".golangci.yaml"; do
    if [ -f "$file" ]; then
        CONFIG_FILE="$file"
        break
    fi
done

if [ -z "$CONFIG_FILE" ]; then
    echo "No .golangci.yml config found, skipping..."
    exit 0
fi

echo "Found config: $CONFIG_FILE"

# Run analyze
if ! golangci-linter-auto-configure analyze --config "$CONFIG_FILE"; then
    echo ""
    echo "Warning: Your configuration could be optimized"
    echo "Run 'golangci-linter-auto-configure configure' to auto-fix"
    echo ""
    # Don't block commit, just warn
fi

echo "Pre-commit hook completed!"
exit 0
`

			// Check if hook already exists
			if _, err := os.Stat(hookPath); err == nil {
				logger.Warnf("Pre-commit hook already exists at %s", hookPath)
				logger.Infof("Use --force to overwrite (not implemented yet)")
				return fmt.Errorf("hook already exists")
			}

			// Write hook file
			if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
				return fmt.Errorf("failed to write hook: %w", err)
			}

			logger.Infof("✅ Pre-commit hook installed at %s", hookPath)
			logger.Infof("The hook will run on every commit to analyze your config")
			logger.Infof("To remove: rm %s", hookPath)

			return nil
		},
	}
}

// Execute runs the CLI using fang for enhanced CLI features
func Execute() error {
	cmd := NewRootCommand()
	return fang.Execute(context.Background(), cmd, fang.WithVersion(Version))
}

// Main is the entry point
func Main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	if err := Execute(); err != nil {
		slog.Error("CLI execution failed", "error", err)
		os.Exit(1)
	}
}
