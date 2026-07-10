package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// DefaultTimeout is the default run timeout value for golangci-lint configurations.
const DefaultTimeout = "5m"

// ToolName is the canonical name of this tool, used for reports, findings, and CLI identity.
const ToolName = "golangci-lint-auto-configure"

// GolangciLintBinaryName is the binary name of the external golangci-lint tool
// invoked for version checks, linter discovery, and formatting.
const GolangciLintBinaryName = "golangci-lint"

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

// DefaultFormatterSettings and DefaultLinterSettings are defined in linter_settings.go
// with typed structs for compile-time safety. See that file for the definitions.
