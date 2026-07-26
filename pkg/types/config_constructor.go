package types

// ConfigOption is a functional option for configuring a Config via NewConfig.
type ConfigOption func(*Config)

// NewConfig creates a new Config with safe production defaults:
//   - Version set to ConfigVersionV2
//   - Run.Timeout set to 5m, Tests enabled, IssuesExitCode 1
//   - Output.Formats initialized to empty map
//   - Linters.Exclusions.Generated and Formatters.Exclusions.Generated set to "lax"
//   - Issues.MaxIssuesPerLinter 50, Issues.MaxSameIssues 10
//
// Options are applied in order after defaults are set.
func NewConfig(opts ...ConfigOption) *Config {
	cfg := &Config{
		Version: ConfigVersionV2,
		Run: RunConfig{
			Timeout:        "5m",
			IssuesExitCode: 1,
			Tests:          true,
		},
		Output: OutputConfig{
			Formats: map[string]any{},
		},
		Linters: LintersConfig{
			Exclusions: LintersExclusionsConfig{
				Generated: "lax",
			},
		},
		Formatters: FormattersConfig{
			Exclusions: FormattersExclusionsConfig{
				Generated: "lax",
			},
		},
		Issues: IssuesConfig{
			MaxIssuesPerLinter: 50,
			MaxSameIssues:      10,
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// WithTimeout sets the run.timeout value.
func WithTimeout(timeout string) ConfigOption {
	return func(c *Config) {
		c.Run.Timeout = timeout
	}
}

// WithGoVersion sets the run.go value.
func WithGoVersion(version string) ConfigOption {
	return func(c *Config) {
		c.Run.Go = version
	}
}

// WithLinters sets the linters.enable list.
func WithLinters(linters []LinterName) ConfigOption {
	return func(c *Config) {
		c.Linters.Enable = linters
	}
}

// WithFormatters sets the formatters.enable list.
func WithFormatters(formatters []FormatterName) ConfigOption {
	return func(c *Config) {
		c.Formatters.Enable = formatters
	}
}

// WithExclusionRules sets the linters.exclusions.rules list.
func WithExclusionRules(rules []ExclusionRuleConfig) ConfigOption {
	return func(c *Config) {
		c.Linters.Exclusions.Rules = rules
	}
}

// WithExclusionPaths sets the linters.exclusions.paths list.
func WithExclusionPaths(paths []string) ConfigOption {
	return func(c *Config) {
		c.Linters.Exclusions.Paths = paths
	}
}
