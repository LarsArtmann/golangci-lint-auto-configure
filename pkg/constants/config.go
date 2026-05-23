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
// be linted: templ generated files and vendored dependencies.
var DefaultLinterExclusionPaths = []string{
	`_templ\.go$`,
	"vendor/",
}

// DefaultFormatterExclusionPaths are exclusion paths always injected into formatters.exclusions.paths.
// Vendor is excluded because vendored code should never be reformatted.
var DefaultFormatterExclusionPaths = []string{
	`_templ\.go$`,
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
			"dupImport",
			"ifElseChain",
			"octalLiteral",
			"whyNoLint",
		},
	},
	"exhaustruct": map[string]any{
		"exclude": []string{
			"os/exec.Cmd",
		},
	},
}
