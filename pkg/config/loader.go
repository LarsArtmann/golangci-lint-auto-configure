package config

// TODO: Extract LinterList type into types package for consistency
// TODO: Consider using io.Reader/Writer interfaces instead of file paths for testability
// TODO: Add context.Context support for cancellation
// TODO: Extract default config values into constants

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	apperrors "github.com/larsartmann/golangcli-linter-auto-configure/pkg/errors"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
	"github.com/pelletier/go-toml/v2"
	"github.com/samber/mo"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// ConfigFormat represents the configuration file format.
// ConfigFormat represents the format of a golangci-lint configuration file.
//
//revive:disable:exported
type ConfigFormat string

//revive:enable:exported

const (
	// ConfigFormatYAML represents YAML configuration format.
	ConfigFormatYAML ConfigFormat = "yaml"
	// ConfigFormatTOML represents TOML configuration format.
	ConfigFormatTOML ConfigFormat = "toml"
	// ConfigFormatJSON represents JSON configuration format.
	ConfigFormatJSON ConfigFormat = "json"

	// Default file permissions for config files (read/write for owner only).
	defaultFilePermissions = 0o600

	// DefaultMaxIssuesPerLinter is the default maximum issues per linter.
	DefaultMaxIssuesPerLinter = 50
	// DefaultMaxSameIssues is the default maximum same issues.
	DefaultMaxSameIssues = 10
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
	fs     afero.Fs
}

// NewLoader creates a new configuration loader.
func NewLoader(logger *log.Logger) *Loader {
	return &Loader{
		logger: logger,
		fs:     afero.NewOsFs(),
	}
}

// NewLoaderWithFS creates a new configuration loader with a custom filesystem.
// Useful for testing with in-memory filesystems.
func NewLoaderWithFS(logger *log.Logger, fs afero.Fs) *Loader {
	return &Loader{
		logger: logger,
		fs:     fs,
	}
}

// LoadConfig loads a golangci-lint configuration from the given path.
func (l *Loader) LoadConfig(path string) (*Config, error) {
	result := l.LoadConfigResult(path)

	return result.Get()
}

// detectFormat determines the configuration format from the file extension.
func detectFormat(path string) ConfigFormat {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".toml":
		return ConfigFormatTOML
	case ".json":
		return ConfigFormatJSON
	case ".yml", ".yaml":
		return ConfigFormatYAML
	default:
		// Default to YAML for unknown extensions
		return ConfigFormatYAML
	}
}

// unmarshalConfig unmarshals data into a Config based on the format.
func unmarshalConfig(data []byte, format ConfigFormat, config *Config) error {
	switch format {
	case ConfigFormatTOML:
		return toml.Unmarshal(data, config)
	case ConfigFormatJSON:
		return json.Unmarshal(data, config)
	case ConfigFormatYAML:
		return yaml.Unmarshal(data, config)
	default:
		// Default to YAML
		return yaml.Unmarshal(data, config)
	}
}

// LoadConfigResult loads a config and returns a Result type for railway-oriented programming.
func (l *Loader) LoadConfigResult(path string) types.ConfigResult {
	data, err := afero.ReadFile(l.fs, path)
	if err != nil {
		return types.ErrConfig(apperrors.NewConfigError("failed to read config file", path, err))
	}

	format := detectFormat(path)

	var config Config

	if err := unmarshalConfig(data, format, &config); err != nil {
		return types.ErrConfig(apperrors.NewConfigError("failed to parse config file", path, err))
	}

	l.logger.Debugf("Loaded config from %s (format: %s)", path, format)

	return types.OkConfig(&config)
}

// FindConfigFile searches for a golangci-lint config file in the current directory and parent directories.
func (l *Loader) FindConfigFile(startDir string) (string, error) {
	result := l.FindConfigFileResult(startDir)

	return result.Get()
}

// FindAllConfigFiles returns all golangci-lint config files found in the directory.
func (l *Loader) FindAllConfigFiles(startDir string) []string {
	defaultNames := []string{
		".golangci.yml",
		".golangci.yaml",
		".golangci.toml",
		".golangci.json",
	}

	var found []string

	for _, name := range defaultNames {
		path := filepath.Join(startDir, name)
		if _, err := l.fs.Stat(path); err == nil {
			found = append(found, path)
		}
	}

	return found
}

// HasMultipleConfigFiles checks if multiple golangci-lint config files exist and logs a warning.
func (l *Loader) HasMultipleConfigFiles(startDir string) bool {
	configs := l.FindAllConfigFiles(startDir)

	if len(configs) > 1 {
		l.logger.Warnf("⚠️  Multiple golangci-lint config files detected:")

		for _, cfg := range configs {
			l.logger.Warnf("   - %s", cfg)
		}

		l.logger.Warnf("   golangci-lint uses the first match in search order: %s", configs[0])

		return true
	}

	return false
}

// FindConfigFileResult searches for a config file and returns a Result type.
func (l *Loader) FindConfigFileResult(startDir string) types.StringResult {
	defaultNames := []string{
		".golangci.yml",
		".golangci.yaml",
		".golangci.toml",
		".golangci.json",
	}

	for _, name := range defaultNames {
		path := filepath.Join(startDir, name)
		if _, err := l.fs.Stat(path); err == nil {
			l.logger.Debugf("Found config file: %s", path)

			return types.OkString(path)
		}
	}

	return types.ErrString(apperrors.NewConfigError("no golangci-lint config file found in "+startDir, startDir, nil))
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
	} `json:"enabled"`
	Disabled []struct {
		Name string `json:"name"`
	} `json:"disabled"`
}

// GetAllLinterNames fetches all available linter names from golangci-lint.
func (l *Loader) GetAllLinterNames(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "golangci-lint", "linters", "--json")

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
func (l *Loader) CreateDefaultConfig(ctx context.Context) *Config {
	// Fetch all available linters dynamically
	allLinters, err := l.GetAllLinterNames(ctx)
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
			MaxIssuesPerLinter: DefaultMaxIssuesPerLinter,
			MaxSameIssues:      DefaultMaxSameIssues,
		},
	}
}

// SaveConfig saves a golangci-lint configuration to the given path.
func (l *Loader) SaveConfig(config *Config, path string) error {
	result := l.SaveConfigResult(config, path)
	_, err := result.Get()

	return err
}

// Empty is a type alias for an empty struct, used for operations that don't return a value.
type Empty = struct{}

// marshalConfig marshals a Config to bytes based on the format.
func marshalConfig(config *Config, format ConfigFormat) ([]byte, error) {
	switch format {
	case ConfigFormatTOML:
		return toml.Marshal(config)
	case ConfigFormatJSON:
		return json.MarshalIndent(config, "", "  ")
	case ConfigFormatYAML:
		return yaml.Marshal(config)
	default:
		// Default to YAML
		return yaml.Marshal(config)
	}
}

// SaveConfigResult saves a config and returns a Result type for railway-oriented programming.
func (l *Loader) SaveConfigResult(config *Config, path string) mo.Result[Empty] {
	format := detectFormat(path)

	data, err := marshalConfig(config, format)
	if err != nil {
		return mo.Err[Empty](apperrors.NewConfigError("failed to marshal config", path, err))
	}

	if err := afero.WriteFile(l.fs, path, data, defaultFilePermissions); err != nil {
		return mo.Err[Empty](apperrors.NewConfigError("failed to write config file", path, err))
	}

	l.logger.Infof("Saved config to %s (format: %s)", path, format)

	return mo.Ok(Empty{})
}

// IsGitRepo checks if we're inside a git repository.
func (l *Loader) IsGitRepo(ctx context.Context, startDir string) bool {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = startDir

	err := cmd.Run()

	return err == nil
}

// EnsureGitRepo checks if we're inside a git repository and returns an error if not.
// Since git provides version control, backup files are redundant.
func (l *Loader) EnsureGitRepo(ctx context.Context, startDir string) error {
	if !l.IsGitRepo(ctx, startDir) {
		return apperrors.NewConfigError(
			"not in a git repository - git provides version control, so backup files are not created. "+
				"Please initialize a git repository first: git init",
			startDir,
			nil,
		)
	}

	l.logger.Debugf("Git repository detected")

	return nil
}

// ValidateConfig performs validation on the configuration using struct validation.
func (l *Loader) ValidateConfig(config *Config) []error {
	var errs []error

	// Use struct validation from types package
	err := types.ValidateConfig(config)
	if err != nil {
		errs = append(errs, apperrors.NewConfigError("struct validation failed", "", err))
	}

	// Additional business logic validation
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
