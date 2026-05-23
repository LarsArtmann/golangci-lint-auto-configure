package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// DefaultTimeout is the default run timeout value for golangci-lint configurations.
const DefaultTimeout = "5m"

// DefaultConfigFileNames is a list of default golangci-lint config file names.
var DefaultConfigFileNames = []string{
	".golangci.yml",
	".golangci.yaml",
	".golangci.toml",
	".golangci.json",
}

// FormattersManagedByBuildFlow are formatters that should be run by buildflow, not golangci-lint.
var FormattersManagedByBuildFlow = []types.FormatterName{
	"goimports",
	"gofumpt",
}

// CoreFormatters are the formatters enabled by default for all projects.
var CoreFormatters = []string{"gci", "gofumpt", "goimports"}

// FormatterOrder defines the canonical ordering of formatters in config output.
var FormatterOrder = []string{"gci", "goimports", "gofumpt", "golines", "swaggo"}

// RedundantFormatters are formatters that are superseded by other formatters.
var RedundantFormatters = map[types.FormatterName]string{
	"gofmt": "redundant when gofumpt is enabled",
}

// DefaultLinterExclusionPaths are exclusion paths always injected into linters.exclusions.paths
// regardless of dynamic gogenfilter scan results. These match common patterns that should never
// be linted: templ generated files, generic generated files, and vendored dependencies.
var DefaultLinterExclusionPaths = []string{
	`_templ\.go$`,
	`\.gen\.go$`,
	"vendor/",
}

// DefaultFormatterExclusionPaths are exclusion paths always injected into formatters.exclusions.paths.
// Vendor is excluded because vendored code should never be reformatted.
var DefaultFormatterExclusionPaths = []string{
	`_templ\.go$`,
}

// DefaultExclusionRules are exclusion rules always injected into linters.exclusions.rules.
// These suppress linters that are noisy or inappropriate in test files.
var DefaultExclusionRules = []types.ExclusionRuleConfig{
	{
		Path:    `_test\.go`,
		Linters: []string{"exhaustruct", "testpackage", "gochecknoglobals", "funlen", "cyclop", "goconst"},
	},
	{
		Path:    `_test\.go`,
		Text:    "unused",
		Linters: []string{"unused"},
	},
}

// DefaultFormatterSettings provides safe default settings for formatters that require
// configuration. Injected only when the formatter is enabled and no settings exist.
var DefaultFormatterSettings = map[types.FormatterName]any{
	"golines": map[string]any{
		"max-len": 120, //nolint:mnd // intentional default line length
	},
}

// DefaultLinterSettings provides safe default settings for linters that require
// configuration to work correctly when auto-enabled. Without these defaults,
// some linters break builds (e.g. depguard denies everything by default).
var DefaultLinterSettings = map[types.LinterName]any{
	"depguard": map[string]any{
		"rules": map[string]any{
			"main": map[string]any{
				"allow": []string{"$gostd", "$module"},
			},
		},
	},
	"ireturn": map[string]any{
		"allow": []string{"error", "empty", "anon", "stdlib", "generic"},
	},
	"gocritic": map[string]any{
		"disabled-checks": []string{
			"ifElseChain",
		},
	},
	"exhaustruct": map[string]any{
		"exclude": []string{
			"os/exec.Cmd",
		},
	},
	"revive": map[string]any{
		"rules": []any{
			map[string]any{"disabled": true, "name": "exported"},
			map[string]any{"disabled": true, "name": "package-comments"},
		},
	},
	"varnamelen": map[string]any{
		"ignore-map-index-ok":   true,
		"ignore-names":          []string{"err", "ok", "tt", "fn", "t", "i", "m", "g", "a", "b", "v"},
		"ignore-type-assert-ok": true,
	},
	"gomoddirectives": map[string]any{
		"replace-local": true,
	},
	"cyclop": map[string]any{
		"max-complexity": 12, //nolint:mnd // intentional default complexity threshold
	},
}
