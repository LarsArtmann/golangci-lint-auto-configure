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

	return &Client{
		configLoader: config.NewLoader(logger),
		analyzer:     linter.NewAnalyzer(logger),
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
	return c.analyzer.AnalyzeConfig(ctx, configPath)
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
	return c.configLoader.LoadConfig(configPath)
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
	return c.configLoader.SaveConfig(cfg, path)
}

// SimpleAnalyze is a one-line convenience function to analyze a config file
// and return a formatted summary of recommendations
//
// Example:
//
//	summary := client.SimpleAnalyze(context.Background(), client.Options{Verbose: true}, ".golangci.yml")
//	fmt.Println(summary)
func SimpleAnalyze(ctx context.Context, opts Options, configPath string) (string, error) {
	c := New(opts)

	if opts.Verbose {
		c.logger.Infof("Analyzing configuration: %s", configPath)
	}

	analysis, err := c.AnalyzeConfig(ctx, configPath)
	if err != nil {
		return "", fmt.Errorf("analysis failed: %w", err)
	}

	if opts.Verbose {
		c.logger.Infof("Analysis complete: %d enabled linters, %d disabled linters",
			len(analysis.EnabledLinters), len(analysis.DisabledLinters))
	}

	return c.GetSummary(analysis), nil
}
