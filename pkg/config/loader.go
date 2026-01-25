package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

// Config represents a golangci-lint configuration file
type Config struct {
	Version     string                   `yaml:"version"`
	Run         RunConfig                `yaml:"run"`
	Output      OutputConfig             `yaml:"output"`
	Linters     LintersConfig            `yaml:"linters"`
	LintersSettings LintersSettings       `yaml:"linters-settings,omitempty"`
	Issues      IssuesConfig             `yaml:"issues"`
	Servers     ServersConfig            `yaml:"servers"`
}

type RunConfig struct {
	Timeout        string            `yaml:"timeout"`
	Go             string            `yaml:"go"`
	BuildTags       string            `yaml:"build-tags"`
	ModulesDownloadMode string         `yaml:"modules-download-mode"`
	AllowParallelRunners bool          `yaml:"allow-parallel-runners"`
	AllowSerialRunners   bool          `yaml:"allow-serial-runners"`
	GoVersion string            `yaml:"go"`
	Env      []string          `yaml:"env"`
}

type OutputConfig struct {
	Formats         []string `yaml:"formats"`
	PrintIssuedLines bool     `yaml:"print-issued-lines"`
	PrintLinterName  bool     `yaml:"print-linter-name"`
	SortResults     bool     `yaml:"sort-results"`
	PrintWelcomeMessage bool `yaml:"print-welcome-message"`
}

type LintersConfig struct {
	Enable     []string `yaml:"enable"`
	Disable    []string `yaml:"disable"`
	Fast       bool     `yaml:"fast"`
	Presets    []string `yaml:"presets"`
}

type LintersSettings map[string]interface{}

type IssuesConfig struct {
	Exclude           []string            `yaml:"exclude"`
	ExcludeRules      []string            `yaml:"exclude-rules,omitempty"`
	ExcludeGenerated  bool               `yaml:"exclude-generated"`
	ExcludeFiles      []string            `yaml:"exclude-files,omitempty"`
	ExcludeDirs       []string            `yaml:"exclude-dirs,omitempty"`
	MaxIssuesPerLinter int               `yaml:"max-issues-per-linter"`
	MaxSameIssues     int                `yaml:"max-same-issues"`
	NewFromRev       string              `yaml:"new-from-rev"`
	NewFromPatch     string              `yaml:"new-from-patch"`
	UseDefaultExcludes bool                `yaml:"use-default-excludes"`
}

type ServersConfig struct {
	HTTPHeaders map[string]string `yaml:"http-headers"`
}

// Loader handles loading golangci-lint configuration files
type Loader struct {
	logger *log.Logger
}

// NewLoader creates a new configuration loader
func NewLoader(logger *log.Logger) *Loader {
	return &Loader{
		logger: logger,
	}
}

// LoadConfig loads a golangci-lint configuration from the given path
func (l *Loader) LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	l.logger.Debugf("Loaded config from %s", path)
	return &config, nil
}

// FindConfigFile searches for a golangci-lint config file in the current directory and parent directories
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

	return "", fmt.Errorf("no golangci-lint config file found in %s", startDir)
}

// SaveConfig saves a golangci-lint configuration to the given path
func (l *Loader) SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	l.logger.Infof("Saved config to %s", path)
	return nil
}

// CreateBackup creates a backup of the given file
func (l *Loader) CreateBackup(filePath string) (string, error) {
	backupPath := filePath + ".backup"

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	l.logger.Infof("Created backup: %s", backupPath)
	return backupPath, nil
}

// ValidateConfig performs basic validation on the configuration
func (l *Loader) ValidateConfig(config *Config) []error {
	var errors []error

	if config.Run.Timeout != "" && config.Run.Timeout == "" {
		errors = append(errors, fmt.Errorf("run.timeout cannot be empty"))
	}

	if len(config.Linters.Enable) == 0 && len(config.Linters.Disable) == 0 && !config.Linters.Fast {
		l.logger.Debugf("No linter configuration specified, using defaults")
	}

	return errors
}

// GetLintersEnabled returns the list of explicitly enabled linters
func (l *Loader) GetLintersEnabled(config *Config) []string {
	return config.Linters.Enable
}

// GetLintersDisabled returns the list of explicitly disabled linters
func (l *Loader) GetLintersDisabled(config *Config) []string {
	return config.Linters.Disable
}
