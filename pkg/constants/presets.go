package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// ValidPresets lists all valid preset names.
const ValidPresets = "minimal, standard, strict, security, performance, reference"

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
	"reference": {
		"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
		"loggercheck", "errchkjson", "musttag", "sloglint", "nilerr", "noctx",
		"wrapcheck", "errorlint", "prealloc", "unconvert",
		"gocyclo", "funlen", "cyclop", "gocognit", "maintidx",
		"exhaustive", "exhaustruct", "goconst", "misspell",
		"revive", "nolintlint", "forcetypeassert", "gocritic", "unused",
		"bodyclose", "contextcheck", "dupl", "durationcheck", "errname",
		"gochecknoglobals", "gochecknoinits", "gosmopolitan", "interfacebloat",
		"nestif", "nilnil", "nakedret", "predeclared", "reassign",
		"rowserrcheck", "spancheck", "sqlclosecheck", "testifylint",
		"thelper", "unparam", "wastedassign", "copyloopvar",
		"ginkgolinter", "gochecksumtype", "intrange", "mirror",
		"perfsprint", "protogetter", "usetesting", "recvcheck", "nilnesserr",
		"zerologlint", "paralleltest",
	},
}

// PresetDescriptions explains what each preset is for.
var PresetDescriptions = map[string]string{
	"minimal":     "Essential linters only (5 linters) - Fastest, minimal false positives",
	"standard":    "Recommended for most projects (8 linters) - Good balance",
	"strict":      "Maximum linting (17 linters) - CI/CD, strict code quality",
	"security":    "Security-focused linters only",
	"performance": "Performance optimization linters",
	"reference":   "All critical + high priority linters (recommended starting point)",
}
