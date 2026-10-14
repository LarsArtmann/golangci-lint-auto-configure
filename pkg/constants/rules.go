package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// The legacy linter-management maps are derived from the policy table
// (linter_policy_table.go). A linter lives in exactly one tier (or none,
// meaning it is a normal recommendable linter):
//
//  1. DisabledLinters          — forcibly disabled. Moved to linters.disable
//     on every run, never recommended, and must NOT have priority/reason
//     entries (they are excluded from the recommendation pipeline entirely).
//
//  2. NeverAutoEnableLinters   — never added to an enable list by the tool,
//     but never stripped either. If a user manually enables one, it is
//     respected: safe default settings and test-file exclusions are still
//     injected. Must have priority + reason entries (manual enables need
//     report data).
//
//  3. PragmaticNoiseLinters    — enabled by default; dropped only when the
//     user passes --pragmatic. An opt-out escape hatch, not an unconditional
//     decision.
//
// The tier-disjointness invariant is a structural property of the table (one
// Tier field per record); the data-integrity specs assert it on the table.

// DeprecatedLinters maps deprecated linter names to their recommended
// replacements. Derived from the table's supersession fields.
var DeprecatedLinters = deprecatedLintersView()

// DisabledLinters maps linter names that should never be recommended or enabled
// to the reason for disabling them. These linters are explicitly excluded from
// configuration by the tool.
var DisabledLinters = tierView(types.TierDisabled)

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
var NeverAutoEnableLinters = tierView(types.TierNeverAutoEnable)

// PragmaticNoiseLinters is the set of high-friction linters dropped from the
// dynamic enable set when --pragmatic is used. They stay enabled by default;
// --pragmatic is an opt-in escape hatch for projects that find them too noisy.
var PragmaticNoiseLinters = tierView(types.TierPragmaticNoise)
