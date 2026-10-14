package constants

import (
	"slices"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// DefaultTimeout is the default run timeout value for golangci-lint configurations.
const DefaultTimeout = "5m"

// DefaultMaxIssuesPerLinter is the default maximum issues per linter.
const DefaultMaxIssuesPerLinter = 50

// DefaultMaxSameIssues is the default maximum same issues.
const DefaultMaxSameIssues = 10

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
// gci was removed from the default stack (2026-10-11): too many projects
// rejected its reordering of import groups. It remains a valid formatter
// that users can enable manually (see FormatterOrder and FormatterPolicies).
// Derived from the formatter policy table.
var CoreFormatters = coreFormattersView()

// FormatterOrder defines the canonical ordering of formatters in config output.
var FormatterOrder = []types.FormatterName{"gci", "goimports", "gofumpt", "golines", "swaggo"}

// RedundantFormatters maps formatters that are superseded by another formatter.
// The key is the redundant formatter, the value is its superset.
// When the superset is enabled, the redundant formatter is not recommended.
// This mirrors the RedundantLinters pattern for linters.
// Derived from the formatter policy table.
var RedundantFormatters = redundantFormattersView()

// ProjectSpecificFormatters maps formatters that should only be recommended
// when the project actually uses the corresponding technology.
// The value is the technology key used by the analyzer's hasTechnology check.
// Derived from the formatter policy table.
var ProjectSpecificFormatters = projectSpecificFormattersView()

// IsAutoEnableable reports whether the fixer may enable the formatter:
// the core stack unconditionally, project-specific formatters when their
// technology is detected. gci is the motivating non-case: it left
// CoreFormatters on 2026-10-11 and is now manual-only.
func IsAutoEnableable(name types.FormatterName) bool {
	if slices.Contains(CoreFormatters, name) {
		return true
	}

	_, isProjectSpecific := ProjectSpecificFormatters[name]

	return isProjectSpecific
}

// IsRecommendable reports whether the detector may recommend the formatter:
// either the fixer can auto-enable it (the fix clears the finding), or it
// is redundant with an auto-enableable formatter (the fix enables the
// superset and the redundancy rule clears the finding on re-detect).
// Anything else is manual-only: recommending it would produce a finding
// the fixer can never clear.
func IsRecommendable(name types.FormatterName) bool {
	if IsAutoEnableable(name) {
		return true
	}

	if superset, isRedundant := RedundantFormatters[name]; isRedundant {
		return IsAutoEnableable(superset)
	}

	return false
}

// ProjectSpecificLinters maps linters that should only be recommended
// when the project actually uses the corresponding technology.
// The value is the technology key used by the analyzer's hasTechnology check.
// Derived from the policy table's Gate payloads.
var ProjectSpecificLinters = projectSpecificLintersView()

// KnownBadSettingsKeys maps linter names to settings keys that are schema-invalid
// for that linter and must be renamed to the valid key. Sources of bad keys:
// (a) keys this tool emitted under a wrong name (goconst.min-length, emitted
// before 2026-09-11, rejected by golangci-lint <2.13); (b) keys carried over
// verbatim when the fixer migrates a deprecated linter's settings block to its
// successor (exhaustruct_v5 dropped v4's exclude for ignore-patterns). The fixer
// renames such keys on every configure run so previously broken configs self-heal.
// Never add an entry for a key that is merely deprecated upstream — this map is
// strictly for keys that hard-fail golangci-lint config load.
var KnownBadSettingsKeys = map[string]map[string]string{
	"goconst": {
		"min-length": "min-len",
	},
	"exhaustruct_v5": {
		"exclude": "ignore-patterns",
	},
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
		Path: `_test\.go`,
		Linters: []string{
			"exhaustruct_v5",
			"testpackage",
			"gochecknoglobals",
			"funlen",
			"cyclop",
			"goconst",
			"forcetypeassert",
			"gosec",
			"errcheck",
			"wrapcheck",
			"ireturn",
			"recvcheck",
			"contextcheck",
			"exhaustive",
		},
	},
	{
		Path:    `_test\.go`,
		Text:    "unused",
		Linters: []string{"unused"},
	},
}

// DefaultFormatterSettings and DefaultLinterSettings are defined in linter_settings.go
// with typed structs for compile-time safety. See that file for the definitions.
