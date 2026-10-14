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
	validateRedundancyRelations()
}

// validateRedundancyRelations pins the relation table's referential
// integrity: relations may never dangle (a superseded entry or a superseder
// missing from its policy table is a curated-data bug, not a runtime
// condition).
func validateRedundancyRelations() {
	linters := make(map[types.LinterName]bool, len(LinterPolicies))
	for _, record := range LinterPolicies {
		linters[record.Name] = true
	}

	formatters := make(map[types.FormatterName]bool, len(FormatterPolicies))
	for _, record := range FormatterPolicies {
		formatters[record.Name] = true
	}

	for _, relation := range RedundancyRelations {
		if problem := relation.ValidateProblem(linters, formatters); problem != "" {
			panic("constants: " + problem)
		}
	}
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

// LinterSettings returns the base curated settings of a linter (the table's
// first versioned block, already mapped) and whether the table carries one.
func LinterSettings(name types.LinterName) (map[string]any, bool) {
	record, known := PolicyRecord(name)
	if !known || len(record.Settings) == 0 {
		return nil, false
	}

	return record.Settings[0].Settings.ToMap(), true
}

// FormatterSettings returns the base curated settings of a formatter (the
// table's first versioned block, already mapped) and whether the table has one.
func FormatterSettings(name types.FormatterName) (map[string]any, bool) {
	record, known := FormatterPolicyRecord(name)
	if !known || len(record.Settings) == 0 {
		return nil, false
	}

	return record.Settings[0].Settings.ToMap(), true
}

// LintersInTier returns the linter names of one management tier, in table
// order. The renderer consumes this directly instead of the legacy maps.
func LintersInTier(kind types.TierKind) []types.LinterName {
	var names []types.LinterName

	for _, record := range LinterPolicies {
		if record.Tier.Kind == kind {
			names = append(names, record.Name)
		}
	}

	return names
}
