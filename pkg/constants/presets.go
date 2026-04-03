package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// ValidPresets lists all valid preset names.
const ValidPresets = "minimal, standard, strict, security, performance"

// PresetLinters defines linter sets for different configuration presets.
var PresetLinters = map[string][]types.LinterName{
	"minimal": {
		"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
	},
	"standard": {
		"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
		"gocritic", "unused", "copyloopvar",
	},
	"strict": {
		"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
		"gocritic", "unused", "copyloopvar",
		"gocyclo", "funlen", "cyclop", "gocognit",
		"nestif", "maintidx",
		"dupl",
		"goconst", "misspell",
		"nolintlint", "godot", "godox",
	},
	"security": {
		"gosec",
	},
	"performance": {
		"ineffassign", "prealloc", "unconvert", "perfsprint",
	},
}

// PresetDescriptions explains what each preset is for.
var PresetDescriptions = map[string]string{
	"minimal":     "Essential linters only (5 linters) - Fastest, minimal false positives",
	"standard":    "Recommended for most projects (8 linters) - Good balance",
	"strict":      "Maximum linting (17 linters) - CI/CD, strict code quality",
	"security":    "Security-focused linters only",
	"performance": "Performance optimization linters",
}
