package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config represents a golangci-lint configuration file
type Config struct {
	Version    string           `yaml:"version"`
	Run        RunConfig        `yaml:"run"`
	Output     OutputConfig     `yaml:"output"`
	Linters    LintersConfig    `yaml:"linters"`
	Formatters FormattersConfig `yaml:"formatters,omitempty"`
	Issues     IssuesConfig     `yaml:"issues"`
}

type RunConfig struct {
	Timeout              string   `yaml:"timeout"`
	Go                   string   `yaml:"go"`
	BuildTags            []string `yaml:"build-tags"`
	ModulesDownloadMode  string   `yaml:"modules-download-mode,omitempty"`
	AllowParallelRunners bool     `yaml:"allow-parallel-runners"`
	AllowSerialRunners   bool     `yaml:"allow-serial-runners"`
	IssuesExitCode       int      `yaml:"issues-exit-code,omitempty"`
	Tests                bool     `yaml:"tests,omitempty"`
	Concurrency          int      `yaml:"concurrency,omitempty"`
	RelativePathMode     string   `yaml:"relative-path-mode,omitempty"`
}

type OutputConfig struct {
	Formats    map[string]interface{} `yaml:"formats"`
	PathPrefix string                 `yaml:"path-prefix,omitempty"`
	PathMode   string                 `yaml:"path-mode,omitempty"`
	SortOrder  []string               `yaml:"sort-order,omitempty"`
	ShowStats  bool                   `yaml:"show-stats,omitempty"`
}

type LintersConfig struct {
	Enable     []string                `yaml:"enable,omitempty"`
	Disable    []string                `yaml:"disable,omitempty"`
	Default    string                  `yaml:"default,omitempty"`
	Settings   map[string]interface{}  `yaml:"settings,omitempty"`
	Exclusions LintersExclusionsConfig `yaml:"exclusions,omitempty"`
}

type LintersSettings map[string]interface{}

type LintersExclusionsConfig struct {
	Generated   string                `yaml:"generated,omitempty"`
	WarnUnused  bool                  `yaml:"warn-unused,omitempty"`
	Presets     []string              `yaml:"presets,omitempty"`
	Rules       []ExclusionRuleConfig `yaml:"rules,omitempty"`
	Paths       []string              `yaml:"paths,omitempty"`
	PathsExcept []string              `yaml:"paths-except,omitempty"`
}

type ExclusionRuleConfig struct {
	Path       []string `yaml:"path,omitempty"`
	PathExcept []string `yaml:"path-except,omitempty"`
	Text       []string `yaml:"text,omitempty"`
	Source     []string `yaml:"source,omitempty"`
	Linters    []string `yaml:"linters,omitempty"`
}

type IssuesConfig struct {
	MaxIssuesPerLinter int    `yaml:"max-issues-per-linter,omitempty"`
	MaxSameIssues      int    `yaml:"max-same-issues,omitempty"`
	NewFromRev         string `yaml:"new-from-rev,omitempty"`
	NewFromPatch       string `yaml:"new-from-patch,omitempty"`
	New                bool   `yaml:"new,omitempty"`
	NewFromMergeBase   string `yaml:"new-from-merge-base,omitempty"`
	WholeFiles         bool   `yaml:"whole-files,omitempty"`
	Fix                bool   `yaml:"fix,omitempty"`
	UniqByLine         bool   `yaml:"uniq-by-line,omitempty"`
}

type FormattersConfig struct {
	Enable     []string                   `yaml:"enable,omitempty"`
	Disable    []string                   `yaml:"disable,omitempty"`
	Settings   map[string]interface{}     `yaml:"settings,omitempty"`
	Exclusions FormattersExclusionsConfig `yaml:"exclusions,omitempty"`
}

type FormattersExclusionsConfig struct {
	Generated  string   `yaml:"generated,omitempty"`
	WarnUnused bool     `yaml:"warn-unused,omitempty"`
	Paths      []string `yaml:"paths,omitempty"`
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
		return nil, errors.NewConfigError(fmt.Sprintf("failed to read config file"), path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.NewConfigError(fmt.Sprintf("failed to parse config file"), path, err)
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

	return "", errors.NewConfigError(fmt.Sprintf("no golangci-lint config file found in %s", startDir), startDir, nil)
}

// SaveConfig saves a golangci-lint configuration to the given path
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

// CreateBackup creates a backup of the given file
func (l *Loader) CreateBackup(filePath string) (string, error) {
	backupPath := filePath + ".backup"

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", errors.NewConfigError("failed to read file for backup", filePath, err)
	}

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return "", errors.NewConfigError("failed to create backup", backupPath, err)
	}

	l.logger.Infof("Created backup: %s", backupPath)
	return backupPath, nil
}

// ValidateConfig performs basic validation on the configuration
func (l *Loader) ValidateConfig(config *Config) []error {
	var errs []error

	if config.Run.Timeout != "" && config.Run.Timeout == "" {
		errs = append(errs, errors.NewConfigError("run.timeout cannot be empty", "", nil))
	}

	if len(config.Linters.Enable) == 0 && len(config.Linters.Disable) == 0 && config.Linters.Default == "" {
		l.logger.Debugf("No linter configuration specified, using defaults")
	}

	return errs
}

// GetLintersEnabled returns the list of explicitly enabled linters
func (l *Loader) GetLintersEnabled(config *Config) []string {
	return config.Linters.Enable
}

// GetLintersDisabled returns the list of explicitly disabled linters
func (l *Loader) GetLintersDisabled(config *Config) []string {
	return config.Linters.Disable
}

// RestoreConfig restores a configuration from a backup file to target path
func (l *Loader) RestoreConfig(backupPath, targetPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return errors.NewConfigError("failed to read backup file", backupPath, err)
	}

	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		return errors.NewConfigError("failed to restore config", targetPath, err)
	}

	l.logger.Infof("Restored configuration from %s to %s", backupPath, targetPath)
	return nil
}
