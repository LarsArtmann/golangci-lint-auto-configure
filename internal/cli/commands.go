package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	clicmd "github.com/larsartmann/golangcli-linter-auto-configure/internal/cli/cmd"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
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
