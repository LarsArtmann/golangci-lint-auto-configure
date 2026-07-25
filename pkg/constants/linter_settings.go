//go:generate go run ../../cmd/generate-settings -schema=./schema/golangci-lint.jsonschema.json -output=./linter_settings_generated.go

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
	_ SettingsConverter = FunlenSettings{}
	_ SettingsConverter = MndSettings{}
	_ SettingsConverter = GosecSettings{}
	_ SettingsConverter = ErrcheckSettings{}
	_ SettingsConverter = WrapcheckSettings{}
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

type FunlenSettings struct {
	Lines      int `yaml:"lines"`
	Statements int `yaml:"statements"`
}

func (s FunlenSettings) ToMap() map[string]any { return settingsToMap(s) }

type MndSettings struct {
	IgnoredNumbers []string `yaml:"ignored-numbers"`
}

func (s MndSettings) ToMap() map[string]any { return settingsToMap(s) }

type GosecSettings struct {
	Excludes []string `yaml:"excludes"`
}

func (s GosecSettings) ToMap() map[string]any { return settingsToMap(s) }

type ErrcheckSettings struct {
	ExcludeFunctions []string `yaml:"exclude-functions"`
}

func (s ErrcheckSettings) ToMap() map[string]any { return settingsToMap(s) }

type WrapcheckSettings struct {
	IgnoreSigs []string `yaml:"ignore-sigs"`
}

func (s WrapcheckSettings) ToMap() map[string]any { return settingsToMap(s) }

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
		// Stdlib structs that are routinely and safely partially-initialized.
		// Curated from a cross-project audit (954 exhaustruct nolints across 146
		// configs); these types dominate the noise. Project-specific types are
		// added per-project, not here.
		Exclude: []string{
			"net/http.Client",
			"net/http.Server",
			"net/http.Request",
			"net/http.Response",
			"net/http.Transport",
			"net/http.Cookie",
			"net.TCPAddr",
			"net.Dialer",
			"log/slog.HandlerOptions",
			"sync.WaitGroup",
			"bytes.Buffer",
			"time.Ticker",
			"time.Timer",
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
	"funlen": FunlenSettings{
		Lines:      200, //nolint:mnd // house style (dominant override across sibling projects); diverges from golangci-lint upstream default
		Statements: 100,
	},
	"mnd": MndSettings{
		IgnoredNumbers: []string{"0", "1", "2", "100"},
	},
	"gosec": GosecSettings{
		Excludes: []string{
			"G304", // file path via variable: near-universal false positive for config/file loaders
			"G115", // integer overflow conversion: noisy on legitimate casts
		},
	},
	"errcheck": ErrcheckSettings{
		ExcludeFunctions: []string{
			"(*os.File).Close",
			"(io.Closer).Close",
			"(*sql.DB).Close",
			"(*sql.Rows).Close",
			"(*sql.Stmt).Close",
			"(net.Conn).Close",
			"(*net.TCPConn).Close",
			"(*net.UDPConn).Close",
			"fmt.Fprint",
			"fmt.Fprintf",
			"fmt.Fprintln",
			"fmt.Print",
			"fmt.Printf",
			"fmt.Println",
			"(*strings.Builder).WriteString",
			"(*bytes.Buffer).WriteString",
		},
	},
	"wrapcheck": WrapcheckSettings{
		IgnoreSigs: []string{
			".Errorf(",
			"errors.New(",
			"errors.Unwrap(",
			"errors.Join(",
			".Wrap(",
			".Wrapf(",
			".WithMessage(",
			".WithMessagef(",
			".WithStack(",
		},
	},
}

// DefaultFormatterSettings provides compile-time-safe default settings for formatters
// that require configuration. Injected only when the formatter is enabled and no settings exist.
var DefaultFormatterSettings = map[types.FormatterName]SettingsConverter{
	"golines": GolinesFormatterSettings{
		MaxLen: 120, //nolint:mnd // intentional default line length
	},
}
