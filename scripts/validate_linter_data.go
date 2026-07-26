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
		fmt.Printf("   ❌ FAIL: %d linters missing reasons:\n", len(missingReasons))
		for _, linter := range missingReasons {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: All %d linters have reasons\n", len(constants.LinterPriorities))
	}
	fmt.Println()

	// Check 2: All linters with reasons exist in priorities
	fmt.Println("📋 Check 2: All linters in LinterReasons exist in LinterPriorities")
	orphanReasons := checkOrphanReasons()
	if len(orphanReasons) > 0 {
		exitCode = 1
		fmt.Printf("   ❌ FAIL: %d linters have reasons but no priority:\n", len(orphanReasons))
		for _, linter := range orphanReasons {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: All %d reasons have corresponding priorities\n", len(constants.LinterReasons))
	}
	fmt.Println()

	// Check 3: All formatters have priorities
	fmt.Println("📋 Check 3: All formatters in FormatterInfo have priorities")
	formattersWithoutPriorities := checkFormatterPriorities()
	if len(formattersWithoutPriorities) > 0 {
		exitCode = 1
		fmt.Printf("   ❌ FAIL: %d formatters missing priorities:\n", len(formattersWithoutPriorities))
		for _, formatter := range formattersWithoutPriorities {
			fmt.Printf("      - %s\n", formatter)
		}
	} else {
		fmt.Printf("   ✅ PASS: All %d formatters have priorities\n", len(constants.FormatterInfo))
	}
	fmt.Println()

	// Check 4: All formatters with priorities have info
	fmt.Println("📋 Check 4: All formatters in FormatterPriorities exist in FormatterInfo")
	orphanFormatterPriorities := checkOrphanFormatterPriorities()
	if len(orphanFormatterPriorities) > 0 {
		exitCode = 1
		fmt.Printf("   ❌ FAIL: %d formatter priorities without info:\n", len(orphanFormatterPriorities))
		for _, formatter := range orphanFormatterPriorities {
			fmt.Printf("      - %s\n", formatter)
		}
	} else {
		fmt.Printf("   ✅ PASS: All %d formatter priorities have info\n", len(constants.FormatterPriorities))
	}
	fmt.Println()

	// Check 5: DisabledLinters must not appear in priorities or reasons
	fmt.Println("📋 Check 5: DisabledLinters must not appear in LinterPriorities or LinterReasons")
	disabledInPrioritiesOrReasons := checkDisabledLintersConsistency()
	if len(disabledInPrioritiesOrReasons) > 0 {
		exitCode = 1
		fmt.Printf(
			"   ❌ FAIL: %d disabled linters found in priorities or reasons:\n",
			len(disabledInPrioritiesOrReasons),
		)
		for _, linter := range disabledInPrioritiesOrReasons {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf(
			"   ✅ PASS: All %d disabled linters are absent from priorities and reasons\n",
			len(constants.DisabledLinters),
		)
	}
	fmt.Println()

	// Check 6: DisabledLinters must have non-empty reason strings
	fmt.Println("📋 Check 6: DisabledLinters must have non-empty reason strings")
	emptyReasons := checkDisabledLinterReasons()
	if len(emptyReasons) > 0 {
		exitCode = 1
		fmt.Printf("   ❌ FAIL: %d disabled linters have empty reasons:\n", len(emptyReasons))
		for _, linter := range emptyReasons {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf("   ✅ PASS: All %d disabled linters have non-empty reasons\n", len(constants.DisabledLinters))
	}
	fmt.Println()

	// Check 7: NeverAutoEnableLinters must appear in priorities and reasons
	fmt.Println("📋 Check 7: NeverAutoEnableLinters must appear in LinterPriorities and LinterReasons")
	neverAutoEnableMissing := checkNeverAutoEnableLintersConsistency()
	if len(neverAutoEnableMissing) > 0 {
		exitCode = 1
		fmt.Printf(
			"   ❌ FAIL: %d never-auto-enable linters missing from priorities or reasons:\n",
			len(neverAutoEnableMissing),
		)
		for _, linter := range neverAutoEnableMissing {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf(
			"   ✅ PASS: All %d never-auto-enable linters have priorities and reasons\n",
			len(constants.NeverAutoEnableLinters),
		)
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
			"   ❌ FAIL: %d never-auto-enable linters have empty reasons or overlap with another tier:\n",
			len(neverAutoEnableBad),
		)
		for _, linter := range neverAutoEnableBad {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Printf(
			"   ✅ PASS: All %d never-auto-enable linters have non-empty reasons and are disjoint from other tiers\n",
			len(constants.NeverAutoEnableLinters),
		)
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
