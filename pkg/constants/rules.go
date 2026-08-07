package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// DeprecatedLinters maps deprecated linter names to their recommended replacements.
var DeprecatedLinters = map[types.LinterName]types.LinterReplacement{
	"wsl": {
		Replacement: "wsl_v5",
		Reason:      "wsl is deprecated since golangci-lint v2.2.0, use wsl_v5 instead",
	},
	"deadcode": {
		Replacement: "staticcheck",
		Reason:      "deadcode was removed in golangci-lint v2, use staticcheck instead",
	},
	"varcheck": {
		Replacement: "staticcheck",
		Reason:      "varcheck was removed in golangci-lint v2, use staticcheck instead",
	},
	"structcheck": {
		Replacement: "staticcheck",
		Reason:      "structcheck was removed in golangci-lint v2, use staticcheck instead",
	},
	"gosimple": {
		Replacement: "staticcheck",
		Reason:      "gosimple was removed in golangci-lint v2, use staticcheck instead",
	},
	"exhaustivestruct": {
		Replacement: "exhaustive",
		Reason:      "exhaustivestruct was renamed to exhaustive in golangci-lint v2",
	},
	"interfacer": {
		Replacement: "staticcheck",
		Reason:      "interfacer was removed in golangci-lint v2, use staticcheck instead",
	},
	"maligned": {
		Replacement: "govet",
		Reason:      "maligned was removed in golangci-lint v2, use govet with fieldalignment instead",
	},
	"nosnakecase": {
		Replacement: "revive",
		Reason:      "nosnakecase was removed in golangci-lint v2, use revive instead",
	},
	"gomodguard": {
		Replacement: "gomodguard_v2",
		Reason:      "gomodguard is deprecated since golangci-lint v2.12.0, use gomodguard_v2 instead",
		MinVersion:  "v2.12.0",
	},

	// Removed in golangci-lint v2 — have replacements
	"exportloopref": {
		Replacement: "copyloopvar",
		Reason:      "exportloopref was removed in golangci-lint v2 (Go 1.22 fixed loop variable semantics), use copyloopvar instead",
	},
	"golint": {
		Replacement: "revive",
		Reason:      "golint was removed in golangci-lint v2, use revive instead",
	},
	"scopelint": {
		Replacement: "copyloopvar",
		Reason:      "scopelint was removed in golangci-lint v2, use copyloopvar instead",
	},
	"tenv": {
		Replacement: "usetesting",
		Reason:      "tenv was removed in golangci-lint v2, use usetesting instead",
	},

	// Removed in golangci-lint v2 — no direct replacement
	"ifshort": {
		Replacement: "",
		Reason:      "ifshort was removed in golangci-lint v2 with no direct replacement",
	},
	"execinquery": {
		Replacement: "",
		Reason:      "execinquery was removed in golangci-lint v2 with no direct replacement",
	},

	// v1 alternative names (renamed before v2)
	"gas": {
		Replacement: "gosec",
		Reason:      "gas was renamed to gosec",
	},
	"goerr113": {
		Replacement: "err113",
		Reason:      "goerr113 was renamed to err113",
	},
	"gomnd": {
		Replacement: "mnd",
		Reason:      "gomnd was renamed to mnd",
	},
	"logrlint": {
		Replacement: "loggercheck",
		Reason:      "logrlint was renamed to loggercheck",
	},
	"megacheck": {
		Replacement: "staticcheck",
		Reason:      "megacheck was renamed to staticcheck",
	},
	"vet": {
		Replacement: "govet",
		Reason:      "vet was renamed to govet",
	},
	"vetshadow": {
		Replacement: "govet",
		Reason:      "vetshadow was merged into govet",
	},
}

// DisabledLinters maps linter names that should never be recommended or enabled
// to the reason for disabling them. These linters are explicitly excluded from
// configuration by the tool.
// Linter management tiers
//
// Linters are managed across three maps with distinct, non-overlapping
// semantics. A linter lives in exactly one tier (or none, meaning it is a
// normal recommendable linter):
//
//  1. DisabledLinters      — forcibly disabled. Moved to linters.disable on
//     every run, never recommended, and must NOT have priority/reason entries
//     (they are excluded from the recommendation pipeline entirely).
//
//  2. NeverAutoEnableLinters — never added to an enable list by the tool, but
//     never stripped either. If a user manually enables one, it is respected:
//     safe default settings and test-file exclusions are still injected. Must
//     have priority + reason entries (manual enables need report data).
//
//  3. PragmaticNoiseLinters — enabled by default; dropped only when the user
//     passes --pragmatic. An opt-out escape hatch, not an unconditional
//     decision.
//
// The data integrity tests (data_integrity_test.go) and the standalone
// validator (scripts/validate_linter_data.go) enforce that these tiers stay
// disjoint and internally consistent.
var DisabledLinters = map[types.LinterName]string{
	"funcorder":   "provides minimal value and can be confusing for users",
	"noinlineerr": "conflicts with formatters (gofumpt, goimports) that reformat error handling expressions, causing noisy churn and contradictory findings",
}

// RedundantLinters maps linter names that are superseded by formatters.
var RedundantLinters = map[types.LinterName]types.LinterToFormatter{
	"lll": {
		Formatter: "golines",
		Reason:    "redundant when golines formatter is enabled (golines fixes long lines, lll only reports them)",
	},
}

// NeverAutoEnableLinters maps linter names that the tool will never add to an
// enable list, but will never strip from a config either. Unlike
// DisabledLinters (which are forcibly moved to the disable list), these are
// respected when a user has explicitly enabled them — the tool simply never
// recommends them. They still receive safe default settings and test-file
// exclusions (via DefaultExclusionRules) when manually enabled.
var NeverAutoEnableLinters = map[types.LinterName]string{
	"depguard":    "never auto-enabled; use library-policy for banned-library governance, but respect manual configuration for architectural enforcement (layer dependency rules, feature isolation) via file-pattern rules that library-policy cannot replicate",
	"exhaustruct": "highest-friction linter across 160 sibling projects (6.5 nolint ratio); never auto-enabled, but respected with curated stdlib excludes when added manually",
}

// PragmaticNoiseLinters is the set of high-friction linters dropped from the
// dynamic enable set when --pragmatic is used. These four have the highest
// friction ratios (gochecknoglobals 5.1, wrapcheck 1.8, ireturn 1.3,
// funlen 0.83) and are the most commonly cited sources of linting friction.
// They stay enabled by default; --pragmatic is an opt-in escape hatch for
// projects that find them too noisy. (exhaustruct, the former #1 at 6.5, is
// now in NeverAutoEnableLinters — never auto-enabled at all.)
//
// Tier boundary: these stay recommendable (not NeverAutoEnable) because, unlike
// exhaustruct, they function correctly without per-project tuning. Promoting any
// of them to NeverAutoEnable changes default lint coverage and needs explicit
// sign-off; --pragmatic remains the opt-out escape hatch.
var PragmaticNoiseLinters = map[types.LinterName]string{
	"gochecknoglobals": "fights standard Go patterns like registries and sentinels (friction 5.1, no config knobs)",
	"wrapcheck":        "demands every error be wrapped (friction 1.8)",
	"ireturn":          "conflicts with common interface-returning APIs (friction 1.3)",
	"funlen":           "default thresholds too strict vs. house style (friction 0.83)",
}
