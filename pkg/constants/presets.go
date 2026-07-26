package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// ValidPresets lists all valid preset names.
const ValidPresets = "minimal, standard, strict, security, performance, reference, format, house"

// minimalLinters is the essential linter set shared by the minimal and format presets.
var minimalLinters = []types.LinterName{
	"errcheck", "gosec", "govet", "staticcheck", "ineffassign",
}

// PresetLinters defines linter sets for different configuration presets.
var PresetLinters = map[string][]types.LinterName{
	"minimal": minimalLinters,
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
		"exhaustive", "goconst", "misspell",
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
	"format": minimalLinters, // format composes minimal's linters and adds its own formatters (see PresetFormatters)
	"house":  minimalLinters, // house composes minimal's linters and adds the canonical formatter quadruple (see PresetFormatters)
}

// PresetFormatters defines formatter sets for presets that include formatters.
// Currently only the "format" preset enables formatters explicitly.
var PresetFormatters = map[string][]types.FormatterName{
	"format": {"gci", "goimports", "gofumpt", "golines"},
	"house":  {"gci", "goimports", "gofumpt", "golines"},
}

// PresetDescriptions explains what each preset is for.
var PresetDescriptions = map[string]string{
	"minimal":     "Essential linters only (5 linters) - Fastest, minimal false positives",
	"standard":    "Recommended for most projects (8 linters) - Good balance",
	"strict":      "Maximum linting (20 linters) - CI/CD, strict code quality",
	"security":    "Security-focused linters only",
	"performance": "Performance optimization linters",
	"reference":   "All critical + high priority linters (recommended starting point)",
	"format":      "Core formatters + essential linters (5 linters, 4 formatters) - Code formatting setup",
	"house":       "House formatter stack (4 formatters: gci, goimports, gofumpt, golines) + essential linters - validated winning stack across 128/160 projects",
}
