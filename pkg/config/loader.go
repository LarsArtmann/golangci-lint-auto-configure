package config

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"charm.land/log/v2"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	apperrors "github.com/larsartmann/golangci-lint-auto-configure/pkg/errors"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/utils"
	"github.com/pelletier/go-toml/v2"
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

// FS defines the filesystem operations needed by the config package.
type FS interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Remove(name string) error
}

type osFS struct{}

func (osFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
func (osFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (osFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (osFS) Remove(name string) error              { return os.Remove(name) }

// Loader handles loading golangci-lint configuration files.
type Loader struct {
	logger *log.Logger
	fs     FS
}

// NewLoader creates a new configuration loader.
func NewLoader(logger *log.Logger) *Loader {
	return &Loader{
		logger: logger,
		fs:     osFS{},
	}
}

// NewLoaderWithFS creates a new configuration loader with a custom filesystem.
func NewLoaderWithFS(logger *log.Logger, fs FS) *Loader {
	return &Loader{
		logger: logger,
		fs:     fs,
	}
}

// LoadConfig loads a golangci-lint configuration from the given path.
func (l *Loader) LoadConfig(path string) (*types.Config, error) {
	data, err := l.fs.ReadFile(path)
	if err != nil {
		return nil, apperrors.NewConfigError("failed to read config file", path, err)
	}

	format := detectFormat(path)

	var config types.Config

	if err := unmarshalConfig(data, format, &config); err != nil {
		return nil, apperrors.NewConfigError("failed to parse config file", path, err)
	}

	migrateLintersSettingsV1(&config, l.logger)

	l.logger.Debugf("Loaded config from %s (format: %s)", path, format)

	return &config, nil
}

// detectFormat determines the configuration format from the file extension.
func detectFormat(path string) ConfigFormat {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".toml":
		return ConfigFormatTOML
	case ".json":
		return ConfigFormatJSON
	case ".yml", ".yaml", "":
		return ConfigFormatYAML
	default:
		return ConfigFormatYAML
	}
}

// errUnsupportedConfigFormat is returned for unknown config formats.
var errUnsupportedConfigFormat = errors.New("unsupported config format")

// unmarshalConfig unmarshals data into a types.Config based on the format.
func unmarshalConfig(data []byte, format ConfigFormat, config *types.Config) error {
	switch format {
	case ConfigFormatTOML:
		return toml.Unmarshal(data, config)
	case ConfigFormatJSON:
		return json.Unmarshal(data, config)
	case ConfigFormatYAML:
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(false)

		return decoder.Decode(config)
	default:
		return errorfamily.WrapRejectionf(errUnsupportedConfigFormat, "config.unsupported_format",
			"format %s", format)
	}
}

func migrateLintersSettingsV1(config *types.Config, logger *log.Logger) {
	if len(config.LintersSettingsV1) == 0 {
		return
	}

	logger.Warnf("Migrating top-level linters-settings (v1) to linters.settings (v2)")

	types.InitLintersSettings(&config.Linters)

	maps.Copy(config.Linters.Settings, config.LintersSettingsV1)
	config.LintersSettingsV1 = nil
}

// FindConfigFile searches for a golangci-lint config file in the current directory and parent directories.
func (l *Loader) FindConfigFile(startDir string) (string, error) {
	for _, name := range constants.DefaultConfigFileNames {
		path := filepath.Join(startDir, name)
		if _, err := l.fs.Stat(path); err == nil {
			l.logger.Debugf("Found config file: %s", path)

			return path, nil
		}
	}

	return "", apperrors.NewConfigError("no golangci-lint config file found in "+startDir, startDir, nil)
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

// FindOrGetDefaultConfigPath searches for a config file and returns a default path if none exists.
func (l *Loader) FindOrGetDefaultConfigPath(startDir string) string {
	configFile, err := l.FindConfigFile(startDir)
	if err == nil {
		return configFile
	}

	// Return default path if no config found
	return filepath.Join(startDir, constants.DefaultConfigFileNames[0])
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

// getAllLinterNames fetches all available linter names from golangci-lint.
func (l *Loader) getAllLinterNames(ctx context.Context) ([]types.LinterName, error) {
	ctx, cancel := context.WithTimeout(ctx, LintersTimeout)
	defer cancel()

	//nolint:gosec // binary name is a constant, not user input
	cmd := exec.CommandContext(ctx, constants.GolangciLintBinaryName, "linters", "--json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, apperrors.WrapClassified(err, "config.linters_run",
			"failed to run golangci-lint linters")
	}

	var linterList LinterList
	if err := json.Unmarshal(output, &linterList); err != nil {
		return nil, errorfamily.WrapCorruptionf(err, "config.linters_parse",
			"failed to parse golangci-lint linters output (output=%s)", output)
	}

	// Extract all enabled linter names
	var linterNames []types.LinterName
	for _, linter := range linterList.Enabled {
		linterNames = append(linterNames, types.LinterName(linter.Name))
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
// The config includes production-ready defaults: exclusion paths, exclusion rules for test files,
// and the standard issues configuration.
func (l *Loader) CreateDefaultConfig(ctx context.Context) *types.Config {
	allLinters := l.fetchLintersWithFallback(ctx)
	goVersion := l.detectGoVersion(ctx)

	return newDefaultConfig(allLinters, goVersion)
}

func newDefaultConfig(allLinters []types.LinterName, goVersion string) *types.Config {
	return &types.Config{
		Version: types.ConfigVersionV2,
		Run: types.RunConfig{
			Timeout:        "5m",
			Go:             goVersion,
			IssuesExitCode: 1,
			Tests:          true,
		},
		Output: types.OutputConfig{
			Formats: map[string]any{},
		},
		Linters: types.LintersConfig{
			Enable: allLinters,
			Exclusions: types.LintersExclusionsConfig{
				Generated: "lax",
				Rules:     defaultExclusionRules(),
				Paths:     defaultExclusionPaths(),
			},
		},
		Formatters: types.FormattersConfig{
			Exclusions: types.FormattersExclusionsConfig{
				Generated: "lax",
				Paths:     defaultFormatterExclusionPaths(),
			},
		},
		Issues: types.IssuesConfig{
			MaxIssuesPerLinter: DefaultMaxIssuesPerLinter,
			MaxSameIssues:      DefaultMaxSameIssues,
		},
	}
}

func (l *Loader) fetchLintersWithFallback(ctx context.Context) []types.LinterName {
	allLinters, err := l.getAllLinterNames(ctx)
	if err != nil {
		l.logger.Warnf("Failed to fetch all linters, using critical set: %v", err)

		return []types.LinterName{
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

func defaultExclusionPaths() []string {
	paths := make([]string, 0, len(constants.DefaultLinterExclusionPaths))

	paths = append(paths, constants.DefaultLinterExclusionPaths...)

	return paths
}

func defaultFormatterExclusionPaths() []string {
	paths := make([]string, 0, len(constants.DefaultFormatterExclusionPaths))

	paths = append(paths, constants.DefaultFormatterExclusionPaths...)

	return paths
}

func defaultExclusionRules() []types.types.ExclusionRuleConfig {
	rules := make([]types.types.ExclusionRuleConfig, 0, len(constants.DefaultExclusionRules))

	rules = append(rules, constants.DefaultExclusionRules...)

	return rules
}

// marshalConfig marshals a types.Config to bytes based on the format.
func marshalConfig(config *types.Config, format ConfigFormat) ([]byte, error) {
	switch format {
	case ConfigFormatTOML:
		return toml.Marshal(config)
	case ConfigFormatJSON:
		return json.Marshal(config, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
	case ConfigFormatYAML:
		return yaml.Marshal(config)
	default:
		// Default to YAML
		return yaml.Marshal(config)
	}
}

// SaveConfig saves a golangci-lint configuration to the given path.
func (l *Loader) SaveConfig(config *types.Config, path string) error {
	format := detectFormat(path)

	data, err := marshalConfig(config, format)
	if err != nil {
		return l.saveError("marshal config", path, err)
	}

	if err := l.fs.WriteFile(path, data, defaultFilePermissions); err != nil {
		return l.saveError("write config file", path, err)
	}

	l.logger.Infof("Saved config to %s (format: %s)", path, format)

	return nil
}

// saveError creates a config error for save operations.
func (l *Loader) saveError(operation, path string, err error) error {
	return apperrors.NewConfigError("failed to "+operation, path, err)
}

// IsGitRepo checks if we're inside a git repository.
func (l *Loader) IsGitRepo(ctx context.Context, startDir string) bool {
	return utils.IsGitRepo(ctx, startDir)
}

// ValidateConfig performs validation on the configuration using struct validation.
func (l *Loader) ValidateConfig(config *types.Config) []error {
	var errs []error

	// Use struct validation from types package
	err := types.ValidateConfig(config)
	if err != nil {
		errs = append(errs, apperrors.NewConfigError("struct validation failed", "", err))
	}

	// Settings key validation — soft warnings for unknown keys
	for _, warning := range ValidateSettingsKeys(config) {
		l.logger.Warnf("⚠️  %s", warning)
	}

	// Additional business logic validation
	if len(config.Linters.Enable) == 0 && len(config.Linters.Disable) == 0 && config.Linters.Default == "" {
		l.logger.Debugf("No linter configuration specified, using defaults")
	}

	return errs
}

// GetLintersEnabled returns the list of explicitly enabled linters.
func (l *Loader) GetLintersEnabled(config *types.Config) []types.LinterName {
	return config.Linters.Enable
}

// GetLintersDisabled returns the list of explicitly disabled linters.
func (l *Loader) GetLintersDisabled(config *types.Config) []types.LinterName {
	return config.Linters.Disable
}
