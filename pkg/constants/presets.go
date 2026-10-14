package constants

import (
	"fmt"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// ValidPresets lists all valid preset names.
const ValidPresets = "minimal, standard, strict, security, performance, reference, format, house"

// Presets are QUERIES over the policy table, not parallel copies of it.
//
// Most presets are declared compositions: they name table records directly
// (validated at init — a preset naming an unknown linter panics, so preset
// and table can never drift apart). The reference preset is a true query:
// every Critical- or High-priority recommendable record. The old hardcoded
// reference list could silently drift from LinterPriorities — that drift is
// now structurally impossible.
//
// init() (linter_policy_views.go) enforces that every named record exists;
// presetQueriesParity specs pin the derived lists against the frozen goldens.

// presetPrioritySet is the declared core shared by minimal, format, and house.
var presetPrioritySet = []types.LinterName{
	"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
}

// presetQueries maps preset name to the query producing its linter set, in
// deterministic (table) order.
var presetQueries = map[string]func(view map[types.LinterName]types.LinterPolicy) []types.LinterName{
	"minimal": names(presetPrioritySet),
	"standard": compose(presetPrioritySet,
		"gocritic", "unused", "copyloopvar"),
	"strict": compose(presetPrioritySet,
		"gocritic", "unused", "copyloopvar",
		"gocyclo", "funlen", "cyclop", "gocognit",
		"nestif", "maintidx",
		"dupl",
		"goconst", "misspell",
		"nolintlint", "godot", "godox"),
	"security":    names([]types.LinterName{"gosec"}),
	"performance": names([]types.LinterName{"ineffassign", "prealloc", "unconvert", "perfsprint"}),
	"reference":   byPriority(types.LinterPriorityCritical, types.LinterPriorityHigh),
	"format":      names(presetPrioritySet),
	"house":       names(presetPrioritySet),
}

// PresetLinters resolves every preset query against the policy table.
// Derived at init; the parity specs pin the result against the frozen goldens.
var PresetLinters = resolvePresetLinters()

func resolvePresetLinters() map[string][]types.LinterName {
	view := make(map[types.LinterName]types.LinterPolicy, len(LinterPolicies))
	for _, record := range LinterPolicies {
		view[record.Name] = record
	}

	resolved := make(map[string][]types.LinterName, len(presetQueries))
	for name, query := range presetQueries {
		resolved[name] = query(view)
	}

	return resolved
}

// names builds a query returning the named records in the given order,
// validating every name against the view.
func names(want []types.LinterName) func(map[types.LinterName]types.LinterPolicy) []types.LinterName {
	return func(view map[types.LinterName]types.LinterPolicy) []types.LinterName {
		out := make([]types.LinterName, 0, len(want))
		for _, name := range want {
			if _, known := view[name]; !known {
				panic(fmt.Sprintf("constants: preset references unknown linter %q", name))
			}

			out = append(out, name)
		}

		return out
	}
}

// compose builds names(first, rest...) — the minimal set plus additions.
func compose(
	first []types.LinterName,
	rest ...types.LinterName,
) func(map[types.LinterName]types.LinterPolicy) []types.LinterName {
	return names(append(append([]types.LinterName{}, first...), rest...))
}

// byPriority builds a query returning every recommendable record whose
// priority is one of the given levels, in table order.
func byPriority(levels ...types.LinterPriority) func(map[types.LinterName]types.LinterPolicy) []types.LinterName {
	wanted := make(map[types.LinterPriority]bool, len(levels))
	for _, level := range levels {
		wanted[level] = true
	}

	return func(view map[types.LinterName]types.LinterPolicy) []types.LinterName {
		var out []types.LinterName

		for _, record := range LinterPolicies {
			if !wanted[record.Priority] {
				continue
			}

			if record.Tier.Kind == types.TierDisabled {
				continue
			}

			out = append(out, record.Name)
		}

		return out
	}
}

// PresetFormatters resolves the formatter sets for presets that include
// formatters, derived from the formatter policy table (CoreFormatters).
var PresetFormatters = map[string][]types.FormatterName{
	"format": CoreFormattersInOrder(),
	"house":  CoreFormattersInOrder(),
}

// PresetDescriptions explains what each preset is for, with counts derived
// from the resolved linter sets so the prose can never disagree with the data.
var PresetDescriptions = map[string]string{
	"minimal": fmt.Sprintf(
		"Essential linters only (%d linters) - Fastest, minimal false positives",
		len(PresetLinters["minimal"]),
	),
	"standard": fmt.Sprintf(
		"Recommended for most projects (%d linters) - Good balance",
		len(PresetLinters["standard"]),
	),
	"strict": fmt.Sprintf(
		"Maximum linting (%d linters) - CI/CD, strict code quality",
		len(PresetLinters["strict"]),
	),
	"security":    "Security-focused linters only",
	"performance": "Performance optimization linters",
	"reference": fmt.Sprintf(
		"All critical + high priority linters (%d linters, derived from the policy table) (recommended starting point)",
		len(PresetLinters["reference"]),
	),
	"format": fmt.Sprintf(
		"Core formatters + essential linters (%d linters, %d formatters) - Code formatting setup",
		len(PresetLinters["format"]),
		len(PresetFormatters["format"]),
	),
	"house": fmt.Sprintf(
		"House formatter stack (%d formatters: goimports, gofumpt, golines) + essential linters - gci dropped from the stack 2026-10-11",
		len(PresetFormatters["house"]),
	),
}
