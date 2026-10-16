package constants

import (
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// DefaultTimeout is the default run timeout value for golangci-lint configurations.
const DefaultTimeout = "5m"

// DefaultMaxIssuesPerLinter is the enforced maximum issues per linter.
// 0 means unlimited: issue output must never be silently truncated, so the
// fixer converges every config to an explicit zero (an absent key would let
// golangci-lint fall back to its own truncating default of 50).
const DefaultMaxIssuesPerLinter = 0

// DefaultMaxSameIssues is the enforced maximum of identical issues.
// 0 means unlimited — same rationale as DefaultMaxIssuesPerLinter; golangci's
// absent-key default here is the even lower 3.
const DefaultMaxSameIssues = 0

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

// FormatterOrder defines the canonical ordering of formatters in config output.
var FormatterOrder = []types.FormatterName{"gci", "goimports", "gofumpt", "golines", "swaggo"}

// IsAutoEnableable reports whether the fixer may enable the formatter:
// the core stack unconditionally, project-specific formatters when their
// technology is detected. gci is the motivating non-case: it left the core
// stack on 2026-10-11 and is now manual-only.
func IsAutoEnableable(name types.FormatterName) bool {
	record, known := FormatterPolicyRecord(name)
	if !known {
		return false
	}

	return record.Core || record.Gate.Kind == types.GateRequiresTech
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

	record, known := FormatterPolicyRecord(name)

	return known && record.SupersededBy != "" && IsAutoEnableable(record.SupersededBy)
}

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
