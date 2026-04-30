package finding

import (
	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// LinterToCategory maps a linter name to its finding category.
// This is the single source of truth for the linter→category mapping
// used by both the recommendation converter and the golangci-lint JSON parser.
func LinterToCategory(name string) finding.Category {
	switch name {
	case "gosec", "noctx", "errchkjson":
		return finding.CategorySecurity
	case "govet", "staticcheck", "errcheck", "nilerr", "ineffassign",
		"unconvert", "bodyclose", "contextcheck", "durationcheck":
		return finding.CategoryCorrectness
	case "prealloc", "perfsprint", "unparam":
		return finding.CategoryPerformance
	case "gocyclo", "cyclop", "gocognit", "maintidx", "funlen",
		"nestif", "interfacebloat", "gocritic":
		return finding.CategoryComplexity
	case "dupl", "goconst":
		return finding.CategoryDuplication
	case "wrapcheck", "errorlint", "errname", "nilnil":
		return finding.CategoryErrorHandling
	case "misspell", "revive", "gofmt", "gci", "wsl_v5",
		"dupword", "godot", "lll", "whitespace", "nlreturn":
		return finding.CategoryStyle
	case "paralleltest", "thelper", "testifylint", "ginkgolinter",
		"tparallel", "testpackage", "testableexamples":
		return finding.CategoryTesting
	case "exhaustive", "exhaustruct", "forcetypeassert", "musttag",
		"gochecksumtype", "copyloopvar", "intrange":
		return finding.CategoryTypeSafety
	case "sloglint", "loggercheck":
		return finding.CategoryStructure
	default:
		return finding.CategoryConfiguration
	}
}

// LinterNameToCategory wraps LinterToCategory for the types.LinterName type.
func LinterNameToCategory(name types.LinterName) finding.Category {
	return LinterToCategory(string(name))
}
