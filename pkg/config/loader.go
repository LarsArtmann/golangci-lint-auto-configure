package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/log/v2"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/samber/mo"
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v3"
)

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
	default:
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(false)

		return decoder.Decode(config)
	}
}

func migrateLintersSettingsV1(config *Config, logger *log.Logger) {
	if len(config.LintersSettingsV1) == 0 {
		return
	}

	logger.Warnf("Migrating top-level linters-settings (v1) to linters.settings (v2)")

	if config.Linters.Settings == nil {
		config.Linters.Settings = make(map[string]any)
	}

	maps.Copy(config.Linters.Settings, config.LintersSettingsV1)
	config.LintersSettingsV1 = nil
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

	migrateLintersSettingsV1(&config, l.logger)

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
	var found []string

	for _, name := range constants.DefaultConfigFileNames {
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
	for _, name := range constants.DefaultConfigFileNames {
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
	ctx, cancel := context.WithTimeout(ctx, LintersTimeout)
	defer cancel()

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

// GetLocalGoVersion returns the locally installed Go version.
// Returns an empty string if the version cannot be determined.
func GetLocalGoVersion(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, GoVersionTimeout)
	defer cancel()

	output, err := exec.CommandContext(ctx, "go", "version").Output()
	if err != nil {
		return ""
	}

	// Parse output like: go version go1.26.1 darwin/arm64
	parts := strings.FieldsSeq(string(output))
	for part := range parts {
		if strings.HasPrefix(part, "go") && len(part) > 2 {
			// Extract version without "go" prefix: go1.26.1 -> 1.26.1
			return strings.TrimPrefix(part, "go")
		}
	}

	return ""
}

// GoVersionTimeout is the timeout for getting the local Go version.
const GoVersionTimeout = 5 * time.Second

// LintersTimeout is the timeout for fetching available linter names from golangci-lint.
const LintersTimeout = 30 * time.Second

// CreateDefaultConfig creates a default golangci-lint configuration with ALL linters enabled.
func (l *Loader) CreateDefaultConfig(ctx context.Context) *Config {
	allLinters := l.fetchLintersWithFallback(ctx)
	goVersion := l.detectGoVersion(ctx)

	return &Config{
		Version: "2",
		Run: RunConfig{
			Timeout:        "5m",
			Go:             goVersion,
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

func (l *Loader) fetchLintersWithFallback(ctx context.Context) []string {
	allLinters, err := l.GetAllLinterNames(ctx)
	if err != nil {
		l.logger.Warnf("Failed to fetch all linters, using critical set: %v", err)

		return []string{
			"gosec",
			"errcheck",
			"staticcheck",
			"govet",
			"ineffassign",
		}
	}

	l.logger.Infof("Enabled %d linters in default configuration", len(allLinters))

	return allLinters
}

func (l *Loader) detectGoVersion(ctx context.Context) string {
	goVersion := GetLocalGoVersion(ctx)
	if goVersion != "" {
		l.logger.Infof("Detected local Go version: %s", goVersion)
	}

	return goVersion
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

// SaveConfig saves a golangci-lint configuration to the given path.
func (l *Loader) SaveConfig(config *Config, path string) error {
	result := l.SaveConfigResult(config, path)
	_, err := result.Get()

	return err
}

// SaveConfigResult saves a config and returns a Result type for railway-oriented programming.
func (l *Loader) SaveConfigResult(config *Config, path string) mo.Result[Empty] {
	format := detectFormat(path)

	data, err := marshalConfig(config, format)
	if err != nil {
		return l.saveError("marshal config", path, err)
	}

	if err := afero.WriteFile(l.fs, path, data, defaultFilePermissions); err != nil {
		return l.saveError("write config file", path, err)
	}

	l.logger.Infof("Saved config to %s (format: %s)", path, format)

	return mo.Ok(Empty{})
}

// saveError creates a config error result for save operations.
func (l *Loader) saveError(operation, path string, err error) mo.Result[Empty] {
	return mo.Err[Empty](apperrors.NewConfigError("failed to "+operation, path, err))
}

// IsGitRepo checks if we're inside a git repository.
func (l *Loader) IsGitRepo(ctx context.Context, startDir string) bool {
	return utils.IsGitRepo(ctx, startDir)
}

// EnsureGitRepo checks if we're inside a git repository and returns an error if not.
// Since git provides version control, backup files are redundant.
func (l *Loader) EnsureGitRepo(ctx context.Context, startDir string) error {
	if err := utils.CheckGitRepo(ctx, startDir); err != nil {
		return apperrors.NewConfigError(
			"not in a git repository - git provides version control, so backup files are not created. "+
				"Please initialize a git repository first: git init",
			startDir,
			err,
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
