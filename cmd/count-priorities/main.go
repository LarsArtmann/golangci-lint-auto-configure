package main

import (
	"fmt"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/constants"
)

func main() {
	counts := map[string]int{}
	for _, p := range constants.LinterPriorities {
		counts[string(p)]++
	}
	fmt.Println("total:", len(constants.LinterPriorities))
	for k, v := range counts {
		fmt.Println(k, v)
	}
	fmt.Println("presets:", len(constants.PresetDescriptions))
	for name := range constants.PresetDescriptions {
		fmt.Print(name, " ")
	}
	fmt.Println()
	fmt.Println("exclusion rules:", len(constants.DefaultExclusionRules))
	fmt.Println("deprecated:", len(constants.DeprecatedLinters))
	for old, dep := range constants.DeprecatedLinters {
		fmt.Printf("  %s -> %s (min %s)\n", old, dep.Replacement, dep.MinVersion)
	}
}
