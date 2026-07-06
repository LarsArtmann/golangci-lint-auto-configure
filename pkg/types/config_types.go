package types

// Config represents a golangci-lint configuration file.
// All field tags use kebab-case to match the golangci-lint schema for round-trip safety.
type Config struct {
	Version           string           `json:"version"    toml:"version"              yaml:"version"`
	Run               RunConfig        `json:"run"        toml:"run"                  yaml:"run"`
	Output            OutputConfig     `json:"output"     toml:"output"               yaml:"output"`
	Linters           LintersConfig    `json:"linters"    toml:"linters"              yaml:"linters"`
	Formatters        FormattersConfig `json:"formatters" toml:"formatters,omitempty" yaml:"formatters,omitempty"`
	Issues            IssuesConfig     `json:"issues"     toml:"issues"               yaml:"issues"`
	LintersSettingsV1 map[string]any   `json:"-"          toml:"-"                    yaml:"linters-settings,omitempty"`
}

type RunConfig struct {
	Timeout              string   `json:"timeout"                         toml:"timeout"                         yaml:"timeout"`
	Go                   string   `json:"go"                              toml:"go"                              yaml:"go"`
	BuildTags            []string `json:"build-tags"                      toml:"build-tags"                      yaml:"build-tags"`
	ModulesDownloadMode  string   `json:"modules-download-mode,omitempty" toml:"modules-download-mode,omitempty" yaml:"modules-download-mode,omitempty"`
	AllowParallelRunners bool     `json:"allow-parallel-runners"          toml:"allow-parallel-runners"          yaml:"allow-parallel-runners"`
	AllowSerialRunners   bool     `json:"allow-serial-runners"            toml:"allow-serial-runners"            yaml:"allow-serial-runners"`
	IssuesExitCode       int      `json:"issues-exit-code,omitempty"      toml:"issues-exit-code,omitempty"      yaml:"issues-exit-code,omitempty"`
	Tests                bool     `json:"tests,omitempty"                 toml:"tests,omitempty"                 yaml:"tests,omitempty"`
	Concurrency          int      `json:"concurrency,omitempty"           toml:"concurrency,omitempty"           yaml:"concurrency,omitempty"`
	RelativePathMode     string   `json:"relative-path-mode,omitempty"    toml:"relative-path-mode,omitempty"    yaml:"relative-path-mode,omitempty"`
}

type OutputConfig struct {
	Formats    map[string]any `json:"formats"               toml:"formats"               yaml:"formats"`
	PathPrefix string         `json:"path-prefix,omitempty" toml:"path-prefix,omitempty" yaml:"path-prefix,omitempty"`
	PathMode   string         `json:"path-mode,omitempty"   toml:"path-mode,omitempty"   yaml:"path-mode,omitempty"`
	SortOrder  []string       `json:"sort-order,omitempty"  toml:"sort-order,omitempty"  yaml:"sort-order,omitempty"`
	ShowStats  bool           `json:"show-stats,omitempty"  toml:"show-stats,omitempty"  yaml:"show-stats,omitempty"`
}

type LintersConfig struct {
	Enable     []string                `json:"enable,omitempty"   toml:"enable,omitempty"     yaml:"enable,omitempty"`
	Disable    []string                `json:"disable,omitempty"  toml:"disable,omitempty"    yaml:"disable,omitempty"`
	Default    string                  `json:"default,omitempty"  toml:"default,omitempty"    yaml:"default,omitempty"`
	Settings   map[string]any          `json:"settings,omitempty" toml:"settings,omitempty"   yaml:"settings,omitempty"`
	Exclusions LintersExclusionsConfig `json:"exclusions"         toml:"exclusions,omitempty" yaml:"exclusions,omitempty"`
}

type LintersExclusionsConfig struct {
	Generated   string                `json:"generated,omitempty"    toml:"generated,omitempty"    yaml:"generated,omitempty"`
	WarnUnused  bool                  `json:"warn-unused,omitempty"  toml:"warn-unused,omitempty"  yaml:"warn-unused,omitempty"`
	Presets     []string              `json:"presets,omitempty"      toml:"presets,omitempty"      yaml:"presets,omitempty"`
	Rules       []ExclusionRuleConfig `json:"rules,omitempty"        toml:"rules,omitempty"        yaml:"rules,omitempty"`
	Paths       []string              `json:"paths,omitempty"        toml:"paths,omitempty"        yaml:"paths,omitempty"`
	PathsExcept []string              `json:"paths-except,omitempty" toml:"paths-except,omitempty" yaml:"paths-except,omitempty"`
}

type ExclusionRuleConfig struct {
	Path       string   `json:"path,omitempty"        toml:"path,omitempty"        yaml:"path,omitempty"`
	PathExcept string   `json:"path-except,omitempty" toml:"path-except,omitempty" yaml:"path-except,omitempty"`
	Text       string   `json:"text,omitempty"        toml:"text,omitempty"        yaml:"text,omitempty"`
	Source     string   `json:"source,omitempty"      toml:"source,omitempty"      yaml:"source,omitempty"`
	Linters    []string `json:"linters,omitempty"     toml:"linters,omitempty"     yaml:"linters,omitempty"`
}

func (r ExclusionRuleConfig) RuleKey() string {
	return r.Path + "|" + r.Text + "|" + r.Source
}

type IssuesConfig struct {
	MaxIssuesPerLinter int    `json:"max-issues-per-linter,omitempty" toml:"max-issues-per-linter,omitempty" yaml:"max-issues-per-linter,omitempty"`
	MaxSameIssues      int    `json:"max-same-issues,omitempty"       toml:"max-same-issues,omitempty"       yaml:"max-same-issues,omitempty"`
	NewFromRev         string `json:"new-from-rev,omitempty"          toml:"new-from-rev,omitempty"          yaml:"new-from-rev,omitempty"`
	NewFromPatch       string `json:"new-from-patch,omitempty"        toml:"new-from-patch,omitempty"        yaml:"new-from-patch,omitempty"`
	New                bool   `json:"new,omitempty"                   toml:"new,omitempty"                   yaml:"new,omitempty"`
	NewFromMergeBase   string `json:"new-from-merge-base,omitempty"   toml:"new-from-merge-base,omitempty"   yaml:"new-from-merge-base,omitempty"`
	WholeFiles         bool   `json:"whole-files,omitempty"           toml:"whole-files,omitempty"           yaml:"whole-files,omitempty"`
	Fix                bool   `json:"fix,omitempty"                   toml:"fix,omitempty"                   yaml:"fix,omitempty"`
	UniqByLine         bool   `json:"uniq-by-line,omitempty"          toml:"uniq-by-line,omitempty"          yaml:"uniq-by-line,omitempty"`
}

type FormattersConfig struct {
	Enable     []string                   `json:"enable,omitempty"   toml:"enable,omitempty"     yaml:"enable,omitempty"`
	Disable    []string                   `json:"disable,omitempty"  toml:"disable,omitempty"    yaml:"disable,omitempty"`
	Settings   map[string]any             `json:"settings,omitempty" toml:"settings,omitempty"   yaml:"settings,omitempty"`
	Exclusions FormattersExclusionsConfig `json:"exclusions"         toml:"exclusions,omitempty" yaml:"exclusions,omitempty"`
}

type FormattersExclusionsConfig struct {
	Generated  string   `json:"generated,omitempty"   toml:"generated,omitempty"   yaml:"generated,omitempty"`
	WarnUnused bool     `json:"warn-unused,omitempty" toml:"warn-unused,omitempty" yaml:"warn-unused,omitempty"`
	Paths      []string `json:"paths,omitempty"       toml:"paths,omitempty"       yaml:"paths,omitempty"`
}

// InitLintersSettings initializes the Linters.Settings map if nil.
func InitLintersSettings(cfg *LintersConfig) {
	if cfg.Settings == nil {
		cfg.Settings = make(map[string]any)
	}
}
