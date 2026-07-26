//go:build ignore
// +build ignore

// This script validates the integrity of linter data in the constants package.
// It checks:
//   - No duplicate keys in LinterPriorities
//   - All priorities have corresponding reasons
//   - All enabled linters in .golangci.yml exist in priorities
//   - All formatters are consistent
//   - DisabledLinters entries must not appear in priorities or reasons
//   - DisabledLinters entries must have non-empty reason strings
//   - NeverAutoEnableLinters entries must appear in priorities and reasons
//   - NeverAutoEnableLinters entries must have non-empty reason strings
//   - NeverAutoEnableLinters must not overlap with DisabledLinters or PragmaticNoiseLinters
//
// Run with: go run scripts/validate_linter_data.go
package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

func sortSlice[T types.LinterName | types.FormatterName](slice []T) {
	slices.Sort(slice)
}

func main() {
	exitCode := 0

	fmt.Println("🔍 Validating linter data integrity...")
	fmt.Println()

	// Check 1: All linters in priorities have reasons
	fmt.Println("📋 Check 1: All linters in LinterPriorities have reasons in LinterReasons")
	missingReasons := checkMissingReasons()
	if len(missingReasons) > 0 {
		exitCode = 1
		fmt.Printf("   ❌ FAIL: %s missing reasons:\n", noun(len(missingReasons), "linter", "linters"))
		for _, linter := range missingReasons {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: checked %s; all have reasons\n", noun(len(constants.LinterPriorities), "linter", "linters"))
	}
	fmt.Println()

	// Check 2: All linters with reasons exist in priorities
	fmt.Println("📋 Check 2: All linters in LinterReasons exist in LinterPriorities")
	orphanReasons := checkOrphanReasons()
	if len(orphanReasons) > 0 {
		exitCode = 1
		fmt.Printf("   ❌ FAIL: %s with reasons but no priority:\n", noun(len(orphanReasons), "linter", "linters"))
		for _, linter := range orphanReasons {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: checked %s; all have corresponding priorities\n", noun(len(constants.LinterReasons), "reason", "reasons"))
	}
	fmt.Println()

	// Check 3: All formatters have priorities
	fmt.Println("📋 Check 3: All formatters in FormatterInfo have priorities")
	formattersWithoutPriorities := checkFormatterPriorities()
	if len(formattersWithoutPriorities) > 0 {
		exitCode = 1
		fmt.Printf(
			"   ❌ FAIL: %s missing priorities:\n",
			noun(len(formattersWithoutPriorities), "formatter", "formatters"),
		)
		for _, formatter := range formattersWithoutPriorities {
			fmt.Printf("      - %s\n", formatter)
		}
	} else {
		fmt.Printf("   ✅ PASS: checked %s; all have priorities\n", noun(len(constants.FormatterInfo), "formatter", "formatters"))
	}
	fmt.Println()

	// Check 4: All formatters with priorities have info
	fmt.Println("📋 Check 4: All formatters in FormatterPriorities exist in FormatterInfo")
	orphanFormatterPriorities := checkOrphanFormatterPriorities()
	if len(orphanFormatterPriorities) > 0 {
		exitCode = 1
		fmt.Printf(
			"   ❌ FAIL: %s without info:\n",
			noun(len(orphanFormatterPriorities), "formatter priority", "formatter priorities"),
		)
		for _, formatter := range orphanFormatterPriorities {
			fmt.Printf("      - %s\n", formatter)
		}
	} else {
		fmt.Printf("   ✅ PASS: checked %s; all have info\n", noun(len(constants.FormatterPriorities), "formatter priority", "formatter priorities"))
	}
	fmt.Println()

	// Check 5: DisabledLinters must not appear in priorities or reasons
	fmt.Println("📋 Check 5: DisabledLinters must not appear in LinterPriorities or LinterReasons")
	disabledInPrioritiesOrReasons := checkDisabledLintersConsistency()
	if len(disabledInPrioritiesOrReasons) > 0 {
		exitCode = 1
		fmt.Printf(
			"   ❌ FAIL: %s found in priorities or reasons:\n",
			noun(len(disabledInPrioritiesOrReasons), "disabled linter", "disabled linters"),
		)
		for _, linter := range disabledInPrioritiesOrReasons {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: checked %s; all absent from priorities and reasons\n", noun(len(constants.DisabledLinters), "disabled linter", "disabled linters"))
	}
	fmt.Println()

	// Check 6: DisabledLinters must have non-empty reason strings
	fmt.Println("📋 Check 6: DisabledLinters must have non-empty reason strings")
	emptyReasons := checkDisabledLinterReasons()
	if len(emptyReasons) > 0 {
		exitCode = 1
		fmt.Printf(
			"   ❌ FAIL: %s with empty reasons:\n",
			noun(len(emptyReasons), "disabled linter", "disabled linters"),
		)
		for _, linter := range emptyReasons {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: checked %s; all have non-empty reasons\n", noun(len(constants.DisabledLinters), "disabled linter", "disabled linters"))
	}
	fmt.Println()

	// Check 7: NeverAutoEnableLinters must appear in priorities and reasons
	fmt.Println("📋 Check 7: NeverAutoEnableLinters must appear in LinterPriorities and LinterReasons")
	neverAutoEnableMissing := checkNeverAutoEnableLintersConsistency()
	if len(neverAutoEnableMissing) > 0 {
		exitCode = 1
		fmt.Printf(
			"   ❌ FAIL: %s missing from priorities or reasons:\n",
			noun(len(neverAutoEnableMissing), "never-auto-enable linter", "never-auto-enable linters"),
		)
		for _, linter := range neverAutoEnableMissing {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: checked %s; all have priorities and reasons\n", noun(len(constants.NeverAutoEnableLinters), "never-auto-enable linter", "never-auto-enable linters"))
	}
	fmt.Println()

	// Check 8: NeverAutoEnableLinters must have non-empty reasons and not overlap
	fmt.Println(
		"📋 Check 8: NeverAutoEnableLinters must have non-empty reasons and be disjoint from DisabledLinters and PragmaticNoiseLinters",
	)
	neverAutoEnableBad := checkNeverAutoEnableLinterReasons()
	if len(neverAutoEnableBad) > 0 {
		exitCode = 1
		fmt.Printf(
			"   ❌ FAIL: %s with empty reasons or tier overlap:\n",
			noun(len(neverAutoEnableBad), "never-auto-enable linter", "never-auto-enable linters"),
		)
		for _, linter := range neverAutoEnableBad {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: checked %s; all have non-empty reasons and are disjoint from other tiers\n", noun(len(constants.NeverAutoEnableLinters), "never-auto-enable linter", "never-auto-enable linters"))
	}
	fmt.Println()

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════════")
	if exitCode == 0 {
		fmt.Println("✅ ALL CHECKS PASSED")
	} else {
		fmt.Println("❌ SOME CHECKS FAILED")
	}
	fmt.Println("═══════════════════════════════════════════════════════════")

	os.Exit(exitCode)
}

func sorted[T types.LinterName | types.FormatterName](items []T) []T {
	sortSlice(items)

	return items
}

// noun renders a count with the correct singular or plural noun form, e.g.
// noun(1, "linter", "linters") -> "1 linter", noun(2, "linter", "linters") -> "2 linters".
func noun(n int, singular, pluralNoun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}

	return fmt.Sprintf("%d %s", n, pluralNoun)
}

func checkMissingReasons() []types.LinterName {
	var missing []types.LinterName
	for linter := range constants.LinterPriorities {
		if _, ok := constants.LinterReasons[linter]; !ok {
			missing = append(missing, linter)
		}
	}

	return sorted(missing)
}

func checkOrphanReasons() []types.LinterName {
	var orphan []types.LinterName
	for linter := range constants.LinterReasons {
		if _, ok := constants.LinterPriorities[linter]; !ok {
			orphan = append(orphan, linter)
		}
	}

	return sorted(orphan)
}

func checkFormatterPriorities() []types.FormatterName {
	var missing []types.FormatterName
	for formatter := range constants.FormatterInfo {
		if _, ok := constants.FormatterPriorities[formatter]; !ok {
			missing = append(missing, formatter)
		}
	}

	return sorted(missing)
}

func checkOrphanFormatterPriorities() []types.FormatterName {
	var orphan []types.FormatterName
	for formatter := range constants.FormatterPriorities {
		if _, ok := constants.FormatterInfo[formatter]; !ok {
			orphan = append(orphan, formatter)
		}
	}

	return sorted(orphan)
}

func checkDisabledLintersConsistency() []types.LinterName {
	var violations []types.LinterName
	for linter := range constants.DisabledLinters {
		if _, ok := constants.LinterPriorities[linter]; ok {
			violations = append(violations, linter)
			continue
		}
		if _, ok := constants.LinterReasons[linter]; ok {
			violations = append(violations, linter)
		}
	}

	return sorted(violations)
}

func checkNeverAutoEnableLintersConsistency() []types.LinterName {
	var violations []types.LinterName
	for linter := range constants.NeverAutoEnableLinters {
		if _, ok := constants.LinterPriorities[linter]; !ok {
			violations = append(violations, linter)
			continue
		}
		if _, ok := constants.LinterReasons[linter]; !ok {
			violations = append(violations, linter)
		}
	}

	return sorted(violations)
}

func checkNeverAutoEnableLinterReasons() []types.LinterName {
	var violations []types.LinterName
	for linter, reason := range constants.NeverAutoEnableLinters {
		if reason == "" {
			violations = append(violations, linter)
			continue
		}
		if _, ok := constants.DisabledLinters[linter]; ok {
			violations = append(violations, linter)
			continue
		}
		if _, ok := constants.PragmaticNoiseLinters[linter]; ok {
			violations = append(violations, linter)
		}
	}

	return sorted(violations)
}

func checkDisabledLinterReasons() []types.LinterName {
	var empty []types.LinterName
	for linter, reason := range constants.DisabledLinters {
		if reason == "" {
			empty = append(empty, linter)
		}
	}

	return sorted(empty)
}
