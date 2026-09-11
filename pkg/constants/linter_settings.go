//go:generate go run ../../cmd/generate-settings -schema=./schema/golangci-lint.jsonschema.json -output=./linter_settings_generated.go

package constants

import (
	"fmt"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"go.yaml.in/yaml/v3"
)

// SettingsConverter converts typed settings to map[string]any for config injection.
// Implemented by both linter and formatter settings structs to provide
// compile-time safety for default setting keys and value types.
type SettingsConverter interface {
	ToMap() map[string]any
}

// mustSettingsToMap converts a struct to map[string]any using YAML serialization.
// YAML tags on the struct control the key names (kebab-case).
// Panics if marshaling fails — this is safe because the input is always a statically-typed
// settings struct with known YAML tags. The panic is a defensive guard against a
// programming error (wrong struct type), not a runtime failure mode.
func mustSettingsToMap(v any) map[string]any {
	data, err := yaml.Marshal(v)
	mustSettingsAction("marshal", v, err)

	var m map[string]any

	err = yaml.Unmarshal(data, &m)
	mustSettingsAction("unmarshal", v, err)

	return m
}

// mustSettingsAction panics with a uniform message when a YAML action fails.
// The input is always a statically-typed settings struct, so a panic here
// signals a programming error, not a runtime failure.
func mustSettingsAction(action string, v any, err error) {
	if err == nil {
		return
	}

	panic(fmt.Sprintf("settingsToMap: failed to %s %T: %v", action, v, err))
}

// --- Compile-time interface compliance checks ---.
var (
	_ SettingsConverter = IreturnSettings{}
	_ SettingsConverter = GocriticSettings{}
	_ SettingsConverter = ExhaustructV5Settings{}
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
	_ SettingsConverter = GocognitSettings{}
	_ SettingsConverter = GocycloSettings{}
	_ SettingsConverter = NestifSettings{}
	_ SettingsConverter = GoconstSettings{}
	_ SettingsConverter = TagalignSettings{}
	_ SettingsConverter = WrapcheckSettings{}
	_ SettingsConverter = GolinesFormatterSettings{}
)

// --- Linter Settings Structs ---
// Each struct provides compile-time safety for the settings keys and value types.
// YAML tags use kebab-case to match golangci-lint's config schema.

type IreturnSettings struct {
	Allow []string `yaml:"allow"`
}

func (s IreturnSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type GocriticSettings struct {
	DisabledChecks []string `yaml:"disabled-checks"`
}

func (s GocriticSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type ExhaustructV5Settings struct {
	// IgnorePatterns are regexes for types (Type#Field paths) the analyzer skips.
	// Ported verbatim from the v4 exhaustruct `exclude` list so manually-enabled
	// configs keep the same curated stdlib coverage after migration.
	IgnorePatterns []string `yaml:"ignore-patterns"`
}

func (s ExhaustructV5Settings) ToMap() map[string]any { return mustSettingsToMap(s) }

type ReviveSettings struct {
	Rules []ReviveRule `yaml:"rules"`
}

type ReviveRule struct {
	Disabled bool   `yaml:"disabled"`
	Name     string `yaml:"name"`
}

func (s ReviveSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type VarnamelenSettings struct {
	IgnoreDecls        []string `yaml:"ignore-decls,omitempty"`
	IgnoreMapIndexOk   bool     `yaml:"ignore-map-index-ok"`
	IgnoreNames        []string `yaml:"ignore-names,omitempty"`
	IgnoreTypeAssertOk bool     `yaml:"ignore-type-assert-ok"`
	MaxDistance        int      `yaml:"max-distance,omitempty"`
	MinNameLength      int      `yaml:"min-name-length,omitempty"`
}

func (s VarnamelenSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type GomoddirectivesSettings struct {
	ReplaceLocal bool `yaml:"replace-local"`
}

func (s GomoddirectivesSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type CyclopSettings struct {
	MaxComplexity int `yaml:"max-complexity"`
}

func (s CyclopSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type GinkgolinterSettings struct {
	ForbidFocusContainer bool `yaml:"forbid-focus-container"`
	ForbidSpecPollution  bool `yaml:"forbid-spec-pollution"`
}

func (s GinkgolinterSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type TestifylintSettings struct {
	EnableAll bool     `yaml:"enable-all"`
	Disable   []string `yaml:"disable"`
}

func (s TestifylintSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type MakezeroSettings struct {
	Always bool `yaml:"always"`
}

func (s MakezeroSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type FunlenSettings struct {
	Lines      int `yaml:"lines"`
	Statements int `yaml:"statements"`
}

func (s FunlenSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type MndSettings struct {
	IgnoredFiles   []string `yaml:"ignored-files,omitempty"`
	IgnoredNumbers []string `yaml:"ignored-numbers,omitempty"`
}

func (s MndSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type GosecSettings struct {
	Excludes []string `yaml:"excludes"`
}

func (s GosecSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type ErrcheckSettings struct {
	CheckTypeAssertions bool     `yaml:"check-type-assertions"`
	ExcludeFunctions    []string `yaml:"exclude-functions,omitempty"`
}

func (s ErrcheckSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type GocognitSettings struct {
	MinComplexity int `yaml:"min-complexity"`
}

func (s GocognitSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type GocycloSettings struct {
	MinComplexity int `yaml:"min-complexity"`
}

func (s GocycloSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type NestifSettings struct {
	MinComplexity int `yaml:"min-complexity"`
}

func (s NestifSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type GoconstSettings struct {
	IgnoreTests    bool `yaml:"ignore-tests"`
	MinLength      int  `yaml:"min-len"`
	MinOccurrences int  `yaml:"min-occurrences"`
}

func (s GoconstSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type TagalignSettings struct {
	Align bool     `yaml:"align"`
	Order []string `yaml:"order,omitempty"`
	Sort  bool     `yaml:"sort"`
}

func (s TagalignSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

type WrapcheckSettings struct {
	IgnoreSigs       []string `yaml:"ignore-sigs,omitempty"`
	IgnoreSigRegexps []string `yaml:"ignore-sig-regexps,omitempty"`
}

func (s WrapcheckSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

// --- Formatter Settings Structs ---

type GolinesFormatterSettings struct {
	MaxLen int `yaml:"max-len"`
}

func (s GolinesFormatterSettings) ToMap() map[string]any { return mustSettingsToMap(s) }

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
	"exhaustruct_v5": ExhaustructV5Settings{
		// Stdlib structs that are routinely and safely partially-initialized.
		// Curated from a cross-project audit (954 exhaustruct nolints across 146
		// configs); these types dominate the noise. Project-specific types are
		// added per-project, not here.
		IgnorePatterns: []string{
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
		IgnoreDecls: []string{
			"err error",
			"wg sync.WaitGroup",
			"mu sync.Mutex",
			"c context.Context",
			"r *http.Request",
			"w http.ResponseWriter",
			"t *testing.T",
			"b *testing.B",
			"f *testing.F",
			"db *sql.DB",
			"tx *sql.Tx",
			"ok bool",
			"id string",
			"n int",
			"fn func()",
			"i int",
			"j int",
		},
		IgnoreMapIndexOk:   true,
		IgnoreNames:        []string{"err", "ok", "tt", "fn", "t", "i", "m", "g", "a", "b", "v"},
		IgnoreTypeAssertOk: true,
		MaxDistance:        15, //nolint:mnd // intentional default max distance for short variable names
		MinNameLength:      2,
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
		IgnoredFiles: []string{
			"_test\\.go",
		},
		// IgnoredNumbers intentionally empty: magic-number suppression is per-project,
		// not universal. Injecting a curated list creates a perpetual maintenance game
		// (CV expanded from 10 to 21 values and growing). Let each project decide.
	},
	"gosec": GosecSettings{
		Excludes: []string{
			"G304", // file path via variable: near-universal false positive for config/file loaders
			"G115", // integer overflow conversion: noisy on legitimate casts
		},
	},
	"errcheck": ErrcheckSettings{
		CheckTypeAssertions: true,
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
		IgnoreSigRegexps: []string{
			"encoding/json",
			"fmt\\.",
			"errors\\.",
			"slices\\.",
			"maps\\.",
			"sort\\.",
			"time\\.",
			"net/http",
			"os\\.",
			"io\\.",
			"strings\\.",
			"strconv\\.",
			"path\\.",
			"filepath\\.",
			"compress/gzip",
			"\\(context\\.Context\\)\\.Err",
		},
	},
	"gocognit": GocognitSettings{
		MinComplexity: 25, //nolint:mnd // intentional default cognitive complexity threshold
	},
	"gocyclo": GocycloSettings{
		MinComplexity: 20, //nolint:mnd // intentional default cyclomatic complexity threshold
	},
	"nestif": NestifSettings{
		MinComplexity: 6, //nolint:mnd // intentional default nesting complexity threshold
	},
	"goconst": GoconstSettings{
		IgnoreTests:    true,
		MinLength:      4, //nolint:mnd // intentional default minimum constant length
		MinOccurrences: 5, //nolint:mnd // intentional default minimum occurrences
	},
	"tagalign": TagalignSettings{
		Align: false,
		Order: []string{
			"binding",
			"json",
			"yaml",
			"xml",
			"toml",
			"validate",
			"mapstructure",
		},
		Sort: true,
	},
}

// DefaultFormatterSettings provides compile-time-safe default settings for formatters
// that require configuration. Injected only when the formatter is enabled and no settings exist.
var DefaultFormatterSettings = map[types.FormatterName]SettingsConverter{
	"golines": GolinesFormatterSettings{
		MaxLen: 120, //nolint:mnd // intentional default line length
	},
}
