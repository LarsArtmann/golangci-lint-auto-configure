package constants

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"go.yaml.in/yaml/v3"
)

// SettingsConverter converts typed settings to map[string]any for config injection.
// Implemented by both linter and formatter settings structs to provide
// compile-time safety for default setting keys and value types.
type SettingsConverter interface {
	ToMap() map[string]any
}

// settingsToMap converts a struct to map[string]any using YAML serialization.
// YAML tags on the struct control the key names (kebab-case).
func settingsToMap(v any) map[string]any {
	data, err := yaml.Marshal(v)
	if err != nil {
		panic(err)
	}

	var m map[string]any

	err = yaml.Unmarshal(data, &m)
	if err != nil {
		panic(err)
	}

	return m
}

// --- Compile-time interface compliance checks ---.
var (
	_ SettingsConverter = IreturnSettings{}
	_ SettingsConverter = GocriticSettings{}
	_ SettingsConverter = ExhaustructSettings{}
	_ SettingsConverter = ReviveSettings{}
	_ SettingsConverter = VarnamelenSettings{}
	_ SettingsConverter = GomoddirectivesSettings{}
	_ SettingsConverter = CyclopSettings{}
	_ SettingsConverter = GinkgolinterSettings{}
	_ SettingsConverter = TestifylintSettings{}
	_ SettingsConverter = MakezeroSettings{}
	_ SettingsConverter = GolinesFormatterSettings{}
)

// --- Linter Settings Structs ---
// Each struct provides compile-time safety for the settings keys and value types.
// YAML tags use kebab-case to match golangci-lint's config schema.

type IreturnSettings struct {
	Allow []string `yaml:"allow"`
}

func (s IreturnSettings) ToMap() map[string]any { return settingsToMap(s) }

type GocriticSettings struct {
	DisabledChecks []string `yaml:"disabled-checks"`
}

func (s GocriticSettings) ToMap() map[string]any { return settingsToMap(s) }

type ExhaustructSettings struct {
	Exclude []string `yaml:"exclude"`
}

func (s ExhaustructSettings) ToMap() map[string]any { return settingsToMap(s) }

type ReviveSettings struct {
	Rules []ReviveRule `yaml:"rules"`
}

type ReviveRule struct {
	Disabled bool   `yaml:"disabled"`
	Name     string `yaml:"name"`
}

func (s ReviveSettings) ToMap() map[string]any { return settingsToMap(s) }

type VarnamelenSettings struct {
	IgnoreMapIndexOk   bool     `yaml:"ignore-map-index-ok"`
	IgnoreNames        []string `yaml:"ignore-names"`
	IgnoreTypeAssertOk bool     `yaml:"ignore-type-assert-ok"`
}

func (s VarnamelenSettings) ToMap() map[string]any { return settingsToMap(s) }

type GomoddirectivesSettings struct {
	ReplaceLocal bool `yaml:"replace-local"`
}

func (s GomoddirectivesSettings) ToMap() map[string]any { return settingsToMap(s) }

type CyclopSettings struct {
	MaxComplexity int `yaml:"max-complexity"`
}

func (s CyclopSettings) ToMap() map[string]any { return settingsToMap(s) }

type GinkgolinterSettings struct {
	ForbidFocusContainer bool `yaml:"forbid-focus-container"`
	ForbidSpecPollution  bool `yaml:"forbid-spec-pollution"`
}

func (s GinkgolinterSettings) ToMap() map[string]any { return settingsToMap(s) }

type TestifylintSettings struct {
	EnableAll bool     `yaml:"enable-all"`
	Disable   []string `yaml:"disable"`
}

func (s TestifylintSettings) ToMap() map[string]any { return settingsToMap(s) }

type MakezeroSettings struct {
	Always bool `yaml:"always"`
}

func (s MakezeroSettings) ToMap() map[string]any { return settingsToMap(s) }

// --- Formatter Settings Structs ---

type GolinesFormatterSettings struct {
	MaxLen int `yaml:"max-len"`
}

func (s GolinesFormatterSettings) ToMap() map[string]any { return settingsToMap(s) }

// DefaultLinterSettings provides compile-time-safe default settings for linters
// that require configuration to work correctly when auto-enabled.
// Without these defaults, some linters break builds (e.g. ireturn denies common interfaces by default).
var DefaultLinterSettings = map[types.LinterName]SettingsConverter{
	"ireturn": IreturnSettings{
		Allow: []string{"error", "empty", "anon", "stdlib", "generic"},
	},
	"gocritic": GocriticSettings{
		DisabledChecks: []string{
			"ifElseChain",
		},
	},
	"exhaustruct": ExhaustructSettings{
		Exclude: []string{
			"os/exec.Cmd",
		},
	},
	"revive": ReviveSettings{
		Rules: []ReviveRule{
			{Disabled: true, Name: "exported"},
			{Disabled: true, Name: "package-comments"},
		},
	},
	"varnamelen": VarnamelenSettings{
		IgnoreMapIndexOk:   true,
		IgnoreNames:        []string{"err", "ok", "tt", "fn", "t", "i", "m", "g", "a", "b", "v"},
		IgnoreTypeAssertOk: true,
	},
	"gomoddirectives": GomoddirectivesSettings{
		ReplaceLocal: true,
	},
	"cyclop": CyclopSettings{
		MaxComplexity: 12, //nolint:mnd // intentional default complexity threshold
	},
	"ginkgolinter": GinkgolinterSettings{
		ForbidFocusContainer: true,
		ForbidSpecPollution:  true,
	},
	"testifylint": TestifylintSettings{
		EnableAll: true,
		Disable: []string{
			"go-require",
		},
	},
	"makezero": MakezeroSettings{
		Always: true,
	},
}

// DefaultFormatterSettings provides compile-time-safe default settings for formatters
// that require configuration. Injected only when the formatter is enabled and no settings exist.
var DefaultFormatterSettings = map[types.FormatterName]SettingsConverter{
	"golines": GolinesFormatterSettings{
		MaxLen: 120, //nolint:mnd // intentional default line length
	},
}
