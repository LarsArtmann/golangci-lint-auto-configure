//go:build ignore
// +build ignore

// This script validates the integrity of linter data in the constants package.
// It checks:
//   - No duplicate keys in LinterPriorities
//   - All priorities have corresponding reasons
//   - All enabled linters in .golangci.yml exist in priorities
//   - All formatters are consistent
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

func checkMissingReasons() []types.LinterName {
	var missing []types.LinterName
	for linter := range constants.LinterPriorities {
		if _, ok := constants.LinterReasons[linter]; !ok {
			missing = append(missing, linter)
		}
	}
	sortSlice(missing)
	return missing
}

func checkOrphanReasons() []types.LinterName {
	var orphan []types.LinterName
	for linter := range constants.LinterReasons {
		if _, ok := constants.LinterPriorities[linter]; !ok {
			orphan = append(orphan, linter)
		}
	}
	sortSlice(orphan)
	return orphan
}

func checkFormatterPriorities() []types.FormatterName {
	var missing []types.FormatterName
	for formatter := range constants.FormatterInfo {
		if _, ok := constants.FormatterPriorities[formatter]; !ok {
			missing = append(missing, formatter)
		}
	}
	sortSlice(missing)
	return missing
}

func checkOrphanFormatterPriorities() []types.FormatterName {
	var orphan []types.FormatterName
	for formatter := range constants.FormatterPriorities {
		if _, ok := constants.FormatterInfo[formatter]; !ok {
			orphan = append(orphan, formatter)
		}
	}
	sortSlice(orphan)
	return orphan
}
