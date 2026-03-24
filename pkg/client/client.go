// Package client provides a high-level API for integrating golangci-linter-auto-configure into other Go applications.
// It offers a simplified interface that abstracts away the internal package structure and handles common use cases.
package client

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// Options configures the client behavior.
type Options struct {
	// Logger to use for output. If nil, a default logger is created.
	Logger *log.Logger

	// Verbose enables debug logging
	Verbose bool
}

// Client provides a simplified API for golangci-lint configuration analysis.
type Client struct {
	configLoader *config.Loader
	analyzer     *linter.Analyzer
	fixer        *linter.Fixer
	logger       *log.Logger
}

// New creates a new client with the given options.
func New(opts Options) *Client {
	logger := opts.Logger
	if logger == nil {
		logLevel := log.InfoLevel
		if opts.Verbose {
			logLevel = log.DebugLevel
		}

		logger = log.NewWithOptions(nil, log.Options{
			Level: logLevel,
		})
	}

	analyzer := linter.NewAnalyzer(logger)

	return &Client{
		configLoader: config.NewLoader(logger),
		analyzer:     analyzer,
		fixer:        linter.NewFixer(logger, analyzer),
		logger:       logger,
	}
}

// AnalyzeConfig analyzes a golangci-lint configuration file and returns recommendations
//
// Example:
//
//	client := client.New(client.Options{})
//	analysis, err := client.AnalyzeConfig(context.Background(), ".golangci.yml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d disabled linters", len(analysis.DisabledLinters))
func (c *Client) AnalyzeConfig(ctx context.Context, configPath string) (*types.ConfigAnalysis, error) {
	result, err := c.analyzer.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze config: %w", err)
	}

	return result, nil
}

// LoadConfig loads and parses a golangci-lint configuration file
//
// Example:
//
//	client := client.New(client.Options{})
//	cfg, err := client.LoadConfig(".golangci.yml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Timeout: %s\n", cfg.Run.Timeout)
func (c *Client) LoadConfig(configPath string) (*config.Config, error) {
	cfg, err := c.configLoader.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

// ValidateConfig validates a golangci-lint configuration
//
// Example:
//
//	client := client.New(client.Options{})
//	cfg, _ := client.LoadConfig(".golangci.yml")
//	errs := client.ValidateConfig(cfg)
//	if len(errs) > 0 {
//	    fmt.Printf("Config has %d validation errors\n", len(errs))
//	}
func (c *Client) ValidateConfig(cfg *config.Config) []error {
	return c.configLoader.ValidateConfig(cfg)
}

// GetSummary returns a human-readable summary of the analysis
//
// Example:
//
//	client := client.New(client.Options{})
//	analysis, _ := client.AnalyzeConfig(".golangci.yml")
//	summary := client.GetSummary(analysis)
//	fmt.Println(summary)
func (c *Client) GetSummary(analysis *types.ConfigAnalysis) string {
	return c.analyzer.GetSummary(analysis)
}

// SaveConfig saves a configuration to the specified path
//
// Example:
//
//	client := client.New(client.Options{})
//	cfg := &config.Config{Version: "2", Linters: config.LintersConfig{Enable: []string{"gofmt"}}}
//	err := client.SaveConfig(cfg, ".golangci.yml")
func (c *Client) SaveConfig(cfg *config.Config, path string) error {
	err := c.configLoader.SaveConfig(cfg, path)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// FixOptions configures the behavior of the FixConfig method.
type FixOptions struct {
	// Priority is the minimum priority level for linters to enable.
	// Options: LinterPriorityCritical, LinterPriorityHigh, LinterPriorityMedium, LinterPriorityOptional
	Priority types.LinterPriority

	// DryRun shows what would be changed without modifying files
	DryRun bool
}

// FixConfig analyzes and fixes a golangci-lint configuration file.
// It enables recommended linters, replaces deprecated linters, and removes redundant ones.
//
// Example:
//
//	client := client.New(client.Options{})
//	result, err := client.FixConfig(context.Background(), ".golangci.yml", client.FixOptions{
//	    Priority: types.LinterPriorityHigh,
//	    DryRun:   true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Applied %d fixes: %s\n", result.FixesApplied, result.Message)
func (c *Client) FixConfig(ctx context.Context, configPath string, opts FixOptions) (*types.MigrationResult, error) {
	if opts.Priority == 0 {
		opts.Priority = types.LinterPriorityHigh
	}

	result, err := c.fixer.FixConfig(ctx, configPath, opts.Priority, opts.DryRun)
	if err != nil {
		return nil, fmt.Errorf("failed to fix config: %w", err)
	}

	return result, nil
}

// SimpleFix is a one-line convenience function to analyze and fix a config file.
// Returns a summary of the changes made or would be made.
//
// Example:
//
//	result, err := client.SimpleFix(context.Background(), client.Options{Verbose: true}, ".golangci.yml", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Message)
func SimpleFix(ctx context.Context, opts Options, configPath string, dryRun bool) (*types.MigrationResult, error) {
	clientObj := New(opts)

	if opts.Verbose {
		clientObj.logger.Infof("Analyzing and fixing configuration: %s", configPath)
	}

	result, err := clientObj.FixConfig(ctx, configPath, FixOptions{
		Priority: types.LinterPriorityHigh,
		DryRun:   dryRun,
	})
	if err != nil {
		return nil, fmt.Errorf("fix failed: %w", err)
	}

	if opts.Verbose {
		clientObj.logger.Infof("Fix complete: %d fixes applied", result.FixesApplied)
	}

	return result, nil
}

// SimpleAnalyze is a one-line convenience function to analyze a config file
// and return a formatted summary of recommendations
//
// Example:
//
//	summary := client.SimpleAnalyze(context.Background(), client.Options{Verbose: true}, ".golangci.yml")
//	fmt.Println(summary)
func SimpleAnalyze(ctx context.Context, opts Options, configPath string) (string, error) {
	clientObj := New(opts)

	if opts.Verbose {
		clientObj.logger.Infof("Analyzing configuration: %s", configPath)
	}

	analysis, err := clientObj.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return "", fmt.Errorf("analysis failed: %w", err)
	}

	if opts.Verbose {
		clientObj.logger.Infof("Analysis complete: %d enabled linters, %d disabled linters",
			len(analysis.EnabledLinters), len(analysis.DisabledLinters))
	}

	return clientObj.GetSummary(analysis), nil
}
