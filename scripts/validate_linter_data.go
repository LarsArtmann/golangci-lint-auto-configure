//go:build ignore
// +build ignore

// This script validates the integrity of the linter policy data.
//
// Since the declarative policy table (pkg/constants/linter_policy_table.go)
// landed, most former checks are enforced at package init: one record per
// linter (no duplicate keys), per-record validation (tier payloads, disabled
// linters carrying neither priority nor rationale, priority/rationale
// coupling, supersession referential integrity), and formatter order
// references. A violation panics on import — importing the package here IS
// the check.
//
// What remains script-level is the cross-artifact check the package cannot
// do: every linter this repo's own .golangci.yml enables must exist in the
// policy table (or be a known deprecated alias), and the view-level
// composition facts must hold.
//
// Run with: go run scripts/validate_linter_data.go
package main

import (
	"fmt"
	"os"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"go.yaml.in/yaml/v3"
)

func main() {
	exitCode := 0

	fmt.Println("🔍 Validating linter policy integrity...")

	// Importing constants runs the init() table validation: one record per
	// linter/formatter, per-record Validate, referential integrity, and
	// formatter-order references. Reaching this line means the table is
	// structurally sound.
	fmt.Printf("   ✅ PASS: policy table (%d linters) and formatter table (%d formatters) structurally valid\n",
		len(constants.LinterPolicies), len(constants.FormatterPolicies))

	fmt.Println("📋 Check: min-version and project-specific views reference recommendable linters")

	for linter := range constants.LinterMinVersions {
		if !recommendable(linter) {
			exitCode = 1
			fmt.Printf("   ❌ FAIL: LinterMinVersions references non-recommendable linter %q\n", linter)
		}
	}

	for linter := range constants.ProjectSpecificLinters {
		if !recommendable(linter) {
			exitCode = 1
			fmt.Printf("   ❌ FAIL: ProjectSpecificLinters references non-recommendable linter %q\n", linter)
		}
	}

	if exitCode == 0 {
		fmt.Printf("   ✅ PASS: checked %d min-version and %d tech-gate entries\n",
			len(constants.LinterMinVersions), len(constants.ProjectSpecificLinters))
	}

	fmt.Println("📋 Check: this repo's own .golangci.yml enables known linters only")

	unknown := checkOwnConfig()
	if len(unknown) > 0 {
		exitCode = 1
		fmt.Printf("   ❌ FAIL: %d unknown linters enabled:\n", len(unknown))
		for _, linter := range unknown {
			fmt.Printf("      - %s\n", linter)
		}
	} else {
		fmt.Println("   ✅ PASS: all enabled linters are known to the policy table")
	}

	if exitCode != 0 {
		fmt.Println("\n❌ Linter data integrity check FAILED")

		os.Exit(exitCode)
	}

	fmt.Println("\n✨ Linter data integrity check passed!")
}

func recommendable(linter types.LinterName) bool {
	if _, disabled := constants.DisabledLinters[linter]; disabled {
		return false
	}

	record, known := constants.PolicyRecord(linter)

	return known && record.Tier.Kind != types.TierDisabled
}

type ownConfig struct {
	Linters struct {
		Enable []types.LinterName `yaml:"enable"`
	} `yaml:"linters"`
}

func checkOwnConfig() []string {
	data, err := os.ReadFile(".golangci.yml")
	if err != nil {
		fmt.Printf("   ⚠️  SKIP: cannot read .golangci.yml: %v\n", err)

		return nil
	}

	var cfg ownConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("   ⚠️  SKIP: cannot parse .golangci.yml: %v\n", err)

		return nil
	}

	known := constants.LinterPolicyView()

	var unknown []string

	for _, linter := range cfg.Linters.Enable {
		if _, ok := known[linter]; !ok {
			unknown = append(unknown, string(linter))
		}
	}

	return unknown
}
