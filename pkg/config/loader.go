package config

// TODO: Extract LinterList type into types package for consistency
// TODO: Add support for TOML and JSON config formats (currently only YAML)
// TODO: Consider using io.Reader/Writer interfaces instead of file paths for testability
// TODO: Add context.Context support for cancellation
// TODO: Extract default config values into constants

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	"gopkg.in/yaml.v3"
)

// Re-export types for backward compatibility.
type (
	Config                     = types.Config
	RunConfig                  = types.RunConfig
	OutputConfig               = types.OutputConfig
	LintersConfig              = types.LintersConfig
	LintersExclusionsConfig    = types.LintersExclusionsConfig
	ExclusionRuleConfig        = types.ExclusionRuleConfig
	IssuesConfig               = types.IssuesConfig
	FormattersConfig           = types.FormattersConfig
	FormattersExclusionsConfig = types.FormattersExclusionsConfig
)

// Loader handles loading golangci-lint configuration files.
type Loader struct {
	logger *log.Logger
}

// NewLoader creates a new configuration loader.
func NewLoader(logger *log.Logger) *Loader {
	return &Loader{
		logger: logger,
	}
}

// LoadConfig loads a golangci-lint configuration from the given path.
func (l *Loader) LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewConfigError("failed to read config file", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.NewConfigError("failed to parse config file", path, err)
	}

	l.logger.Debugf("Loaded config from %s", path)

	return &config, nil
}

// FindConfigFile searches for a golangci-lint config file in the current directory and parent directories.
func (l *Loader) FindConfigFile(startDir string) (string, error) {
	defaultNames := []string{
		".golangci.yml",
		".golangci.yaml",
		".golangci.toml",
		".golangci.json",
	}

	for _, name := range defaultNames {
		path := filepath.Join(startDir, name)
		if _, err := os.Stat(path); err == nil {
			l.logger.Debugf("Found config file: %s", path)

			return path, nil
		}
	}

	return "", errors.NewConfigError("no golangci-lint config file found in "+startDir, startDir, nil)
}

// FindOrGetDefaultConfigPath searches for a config file and returns a default path if none exists.
func (l *Loader) FindOrGetDefaultConfigPath(startDir string) string {
	configFile, err := l.FindConfigFile(startDir)
	if err == nil {
		return configFile
	}

	// Return default path if no config found
	return filepath.Join(startDir, ".golangci.yml")
}

// LinterList represents the JSON output from golangci-lint linters command.
type LinterList struct {
	Enabled []struct {
		Name string `json:"name"`
	} `json:"Enabled"`
	Disabled []struct {
		Name string `json:"name"`
	} `json:"Disabled"`
}

// GetAllLinterNames fetches all available linter names from golangci-lint.
func (l *Loader) GetAllLinterNames() ([]string, error) {
	cmd := exec.Command("golangci-lint", "linters", "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run golangci-lint linters: %w", err)
	}

	var linterList LinterList
	if err := json.Unmarshal(output, &linterList); err != nil {
		return nil, fmt.Errorf("failed to parse golangci-lint linters output: %w", err)
	}

	// Extract all enabled linter names
	var linterNames []string
	for _, linter := range linterList.Enabled {
		linterNames = append(linterNames, linter.Name)
	}

	return linterNames, nil
}

// CreateDefaultConfig creates a default golangci-lint configuration with ALL linters enabled.
func (l *Loader) CreateDefaultConfig() *Config {
	// Fetch all available linters dynamically
	allLinters, err := l.GetAllLinterNames()
	if err != nil {
		l.logger.Warnf("Failed to fetch all linters, using critical set: %v", err)
		// Fallback to critical linters if fetch fails
		allLinters = []string{
			"gosec",
			"errcheck",
			"staticcheck",
			"govet",
			"ineffassign",
		}
	} else {
		l.logger.Infof("Enabled %d linters in default configuration", len(allLinters))
	}

	return &Config{
		Version: "2",
		Run: RunConfig{
			Timeout:        "5m",
			IssuesExitCode: 1,
			Tests:          true,
		},
		Linters: LintersConfig{
			Enable: allLinters,
		},
		Issues: IssuesConfig{
			MaxIssuesPerLinter: 50,
			MaxSameIssues:      10,
		},
	}
}

// SaveConfig saves a golangci-lint configuration to the given path.
func (l *Loader) SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.NewConfigError("failed to marshal config", path, err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return errors.NewConfigError("failed to write config file", path, err)
	}

	l.logger.Infof("Saved config to %s", path)

	return nil
}

// EnsureGitRepo checks if we're inside a git repository.
// Since git provides version control, backup files are redundant.
func (l *Loader) EnsureGitRepo(startDir string) error {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = startDir

	if err := cmd.Run(); err != nil {
		return errors.NewConfigError(
			"not in a git repository - git provides version control, so backup files are not created. "+
				"Please initialize a git repository first: git init",
			startDir,
			nil,
		)
	}

	l.logger.Debugf("Git repository detected")

	return nil
}

// ValidateConfig performs basic validation on the configuration.
func (l *Loader) ValidateConfig(config *Config) []error {
	var errs []error

	// TODO: Implement proper timeout format validation (e.g., "5m", "10s")
	// Currently accepting any non-empty string as timeout
	if config.Run.Timeout == "" {
		errs = append(errs, errors.NewConfigError("run.timeout cannot be empty", "", nil))
	}

	if len(config.Linters.Enable) == 0 && len(config.Linters.Disable) == 0 && config.Linters.Default == "" {
		l.logger.Debugf("No linter configuration specified, using defaults")
	}

	return errs
}

// GetLintersEnabled returns the list of explicitly enabled linters.
func (l *Loader) GetLintersEnabled(config *Config) []string {
	return config.Linters.Enable
}

// GetLintersDisabled returns the list of explicitly disabled linters.
func (l *Loader) GetLintersDisabled(config *Config) []string {
	return config.Linters.Disable
}
