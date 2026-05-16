package finding

import (
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// linterCategories maps linter names to their finding category.
var linterCategories = map[string]finding.Category{
	// Security
	"gosec":      finding.CategorySecurity,
	"noctx":      finding.CategorySecurity,
	"errchkjson": finding.CategorySecurity,
	// Correctness
	"govet":         finding.CategoryCorrectness,
	"staticcheck":   finding.CategoryCorrectness,
	"errcheck":      finding.CategoryCorrectness,
	"nilerr":        finding.CategoryCorrectness,
	"ineffassign":   finding.CategoryCorrectness,
	"unconvert":     finding.CategoryCorrectness,
	"bodyclose":     finding.CategoryCorrectness,
	"contextcheck":  finding.CategoryCorrectness,
	"durationcheck": finding.CategoryCorrectness,
	// Performance
	"prealloc":   finding.CategoryPerformance,
	"perfsprint": finding.CategoryPerformance,
	"unparam":    finding.CategoryPerformance,
	// Complexity
	"gocyclo":        finding.CategoryComplexity,
	"cyclop":         finding.CategoryComplexity,
	"gocognit":       finding.CategoryComplexity,
	"maintidx":       finding.CategoryComplexity,
	"funlen":         finding.CategoryComplexity,
	"nestif":         finding.CategoryComplexity,
	"interfacebloat": finding.CategoryComplexity,
	"gocritic":       finding.CategoryComplexity,
	// Duplication
	"dupl":    finding.CategoryDuplication,
	"goconst": finding.CategoryDuplication,
	// Error handling
	"wrapcheck": finding.CategoryErrorHandling,
	"errorlint": finding.CategoryErrorHandling,
	"errname":   finding.CategoryErrorHandling,
	"nilnil":    finding.CategoryErrorHandling,
	// Style
	"misspell":   finding.CategoryStyle,
	"revive":     finding.CategoryStyle,
	"gofmt":      finding.CategoryStyle,
	"gci":        finding.CategoryStyle,
	"wsl_v5":     finding.CategoryStyle,
	"dupword":    finding.CategoryStyle,
	"godot":      finding.CategoryStyle,
	"lll":        finding.CategoryStyle,
	"whitespace": finding.CategoryStyle,
	"nlreturn":   finding.CategoryStyle,
	// Testing
	"paralleltest":     finding.CategoryTesting,
	"thelper":          finding.CategoryTesting,
	"testifylint":      finding.CategoryTesting,
	"ginkgolinter":     finding.CategoryTesting,
	"tparallel":        finding.CategoryTesting,
	"testpackage":      finding.CategoryTesting,
	"testableexamples": finding.CategoryTesting,
	// Type safety
	"exhaustive":      finding.CategoryTypeSafety,
	"exhaustruct":     finding.CategoryTypeSafety,
	"forcetypeassert": finding.CategoryTypeSafety,
	"musttag":         finding.CategoryTypeSafety,
	"gochecksumtype":  finding.CategoryTypeSafety,
	"copyloopvar":     finding.CategoryTypeSafety,
	"intrange":        finding.CategoryTypeSafety,
	// Structure
	"sloglint":      finding.CategoryStructure,
	"loggercheck":   finding.CategoryStructure,
	"gomodguard_v2": finding.CategoryStructure,
}

// LinterToCategory maps a linter name to its finding category.
// Returns CategoryConfiguration for unknown linters.
func LinterToCategory(name string) finding.Category {
	if cat, ok := linterCategories[name]; ok {
		return cat
	}

	return finding.CategoryConfiguration
}

// LinterNameToCategory wraps LinterToCategory for the types.LinterName type.
func LinterNameToCategory(name types.LinterName) finding.Category {
	return LinterToCategory(string(name))
}
