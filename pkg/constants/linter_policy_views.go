package constants

import (
	"fmt"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// Table invariants are enforced at init: a violation is a programming error
// in curated data and must fail fast (at test time and at first use), not
// surface as a corrupted view later.
//
//nolint:gochecknoinits // curated-data validation must fail fast at package init
func init() {
	validateLinterPolicyTable()
	validateFormatterPolicyTable()
}

// deriveLinterTiers rewrites every unmanaged Recommendable record whose
// friction reaches the threshold into pragmatic noise. Explicit management
// tiers (Disabled, NeverAutoEnable) win over measurement; the hand-maintained
// noise map is gone — the measurement IS the decision.
func deriveLinterTiers(records []types.LinterPolicy) []types.LinterPolicy {
	for i := range records {
		records[i].Tier = types.DeriveTier(records[i].Tier, records[i].Friction)
	}

	return records
}

func validateLinterPolicyTable() {
	seen := make(map[types.LinterName]bool, len(LinterPolicies))

	for _, record := range LinterPolicies {
		if err := record.Validate(); err != nil {
			panic(fmt.Sprintf("constants: invalid policy table: %v", err))
		}

		if seen[record.Name] {
			panic(fmt.Sprintf("constants: duplicate policy record for linter %q", record.Name))
		}

		seen[record.Name] = true

		for _, entry := range record.Settings {
			if entry.Settings == nil {
				panic(fmt.Sprintf("constants: linter %q has a nil settings block", record.Name))
			}
		}
	}

	// Referential integrity in a second pass: supersession targets may be
	// declared anywhere in the table (forward references are normal).
	for _, record := range LinterPolicies {
		if record.SupersededBy != "" && !seen[record.SupersededBy] {
			panic(fmt.Sprintf("constants: linter %q supersedes unknown linter %q", record.Name, record.SupersededBy))
		}
	}
}

func validateFormatterPolicyTable() {
	seen := make(map[types.FormatterName]bool, len(FormatterPolicies))

	for _, record := range FormatterPolicies {
		if err := record.Validate(); err != nil {
			panic(fmt.Sprintf("constants: invalid formatter policy table: %v", err))
		}

		if seen[record.Name] {
			panic(fmt.Sprintf("constants: duplicate formatter policy record for %q", record.Name))
		}

		seen[record.Name] = true
	}

	validateFormatterReferences(seen)
	validateFormatterOrder(seen)
}

func validateFormatterReferences(seen map[types.FormatterName]bool) {
	// Referential integrity, second pass (forward references are normal).
	for _, record := range FormatterPolicies {
		if record.SupersededBy != "" && !seen[record.SupersededBy] {
			panic(
				fmt.Sprintf(
					"constants: formatter %q superseded by unknown formatter %q",
					record.Name,
					record.SupersededBy,
				),
			)
		}
	}
}

func validateFormatterOrder(seen map[types.FormatterName]bool) {
	// FormatterOrder references known records; the converse does not hold —
	// gofmt is a known (redundant) formatter that is not part of the
	// canonical render order.
	for _, name := range FormatterOrder {
		if !seen[name] {
			panic(fmt.Sprintf("constants: FormatterOrder references unknown formatter %q", name))
		}
	}
}

// This file derives the legacy linter/formatter maps from the declarative
// policy table (linter_policy_table.go). The table is the single source of
// truth; these views exist so the ~40 existing consumers keep compiling and
// behaving identically (views-parity is pinned by views_parity_test.go
// against byte-goldens captured from the old hand-maintained maps).
//
// Every view returns a fresh map: callers may mutate their result without
// corrupting the table.

// PolicyRecord returns the table record for a linter, reporting whether the
// linter is known to the table.
func PolicyRecord(name types.LinterName) (types.LinterPolicy, bool) {
	for _, record := range LinterPolicies {
		if record.Name == name {
			return record, true
		}
	}

	return types.LinterPolicy{}, false
}

// LinterPolicyView returns a name-indexed view over the policy table.
func LinterPolicyView() map[types.LinterName]types.LinterPolicy {
	view := make(map[types.LinterName]types.LinterPolicy, len(LinterPolicies))
	for _, record := range LinterPolicies {
		view[record.Name] = record
	}

	return view
}

// linterPrioritiesView reconstructs the legacy priority map: one entry per
// linter whose priority is worth calling out (Optional is the implicit
// default and stays absent, matching the old literal).
func linterPrioritiesView() map[types.LinterName]types.LinterPriority {
	view := make(map[types.LinterName]types.LinterPriority)

	for _, record := range LinterPolicies {
		if record.Priority != types.LinterPriorityOptional {
			view[record.Name] = record.Priority
		}
	}

	return view
}

// linterReasonsView reconstructs the legacy rationale map.
func linterReasonsView() map[types.LinterName]string {
	view := make(map[types.LinterName]string)

	for _, record := range LinterPolicies {
		if record.Rationale != "" {
			view[record.Name] = record.Rationale
		}
	}

	return view
}

// deprecatedLintersView reconstructs the legacy deprecation map from the
// supersession fields. A supersession reason without a replacement names a
// linter removed with no direct successor.
func deprecatedLintersView() map[types.LinterName]types.LinterReplacement {
	view := make(map[types.LinterName]types.LinterReplacement)

	for _, record := range LinterPolicies {
		if record.SupersessionReason == "" {
			continue
		}

		view[record.Name] = types.LinterReplacement{
			Replacement: record.SupersededBy,
			Reason:      record.SupersessionReason,
			MinVersion:  record.SupersessionSince.String(),
		}
	}

	return view
}

// tierView returns the linters of one tier keyed by their tier rationale.
func tierView(kind types.TierKind) map[types.LinterName]string {
	view := make(map[types.LinterName]string)

	for _, record := range LinterPolicies {
		if record.Tier.Kind == kind {
			view[record.Name] = record.Tier.Rationale
		}
	}

	return view
}

// linterMinVersionsView reconstructs the legacy min-version map from the
// ValidTool lower bounds.
func linterMinVersionsView() map[types.LinterName]string {
	view := make(map[types.LinterName]string)

	for _, record := range LinterPolicies {
		if !record.ValidTool.Min.IsZero() {
			view[record.Name] = record.ValidTool.Min.String()
		}
	}

	return view
}

// projectSpecificLintersView reconstructs the legacy tech-gate map from the
// Gate payloads.
func projectSpecificLintersView() map[types.LinterName]string {
	view := make(map[types.LinterName]string)

	for _, record := range LinterPolicies {
		if record.Gate.Kind == types.GateRequiresTech {
			view[record.Name] = record.Gate.Tech
		}
	}

	return view
}

// defaultLinterSettingsView reconstructs the legacy curated-settings map
// (one settings block per linter; the table carries versioned blocks).
func defaultLinterSettingsView() map[types.LinterName]types.SettingsConverter {
	view := make(map[types.LinterName]types.SettingsConverter)

	for _, record := range LinterPolicies {
		if len(record.Settings) > 0 {
			view[record.Name] = record.Settings[0].Settings
		}
	}

	return view
}

// LintersInTier returns the linter names of one management tier, in table
// order. The future renderer consumes this directly instead of the legacy maps.
func LintersInTier(kind types.TierKind) []types.LinterName {
	var names []types.LinterName

	for _, record := range LinterPolicies {
		if record.Tier.Kind == kind {
			names = append(names, record.Name)
		}
	}

	return names
}
