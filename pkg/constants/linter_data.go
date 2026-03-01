package constants

import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"

// LinterPriorities maps linter names to their priority levels.
var LinterPriorities = map[types.LinterName]types.LinterPriority{
	// Critical linters - should ALWAYS be enabled
	"loggercheck": types.LinterPriorityCritical,
	"gosec":       types.LinterPriorityCritical,
	"errcheck":    types.LinterPriorityCritical,
	"staticcheck": types.LinterPriorityCritical,
	"govet":       types.LinterPriorityCritical,
	"errchkjson":  types.LinterPriorityCritical,
	"musttag":     types.LinterPriorityCritical,
	"sloglint":    types.LinterPriorityCritical,
	"nilerr":      types.LinterPriorityCritical,
	"noctx":       types.LinterPriorityCritical,

	// High value linters - recommended for most projects
	"wrapcheck":        types.LinterPriorityHigh,
	"errorlint":        types.LinterPriorityHigh,
	"prealloc":         types.LinterPriorityHigh,
	"unconvert":        types.LinterPriorityHigh,
	"ineffassign":      types.LinterPriorityHigh,
	"gocyclo":          types.LinterPriorityHigh,
	"funlen":           types.LinterPriorityHigh,
	"cyclop":           types.LinterPriorityHigh,
	"gocognit":         types.LinterPriorityHigh,
	"maintidx":         types.LinterPriorityHigh,
	"exhaustive":       types.LinterPriorityHigh,
	"exhaustruct":      types.LinterPriorityHigh,
	"goconst":          types.LinterPriorityHigh,
	"misspell":         types.LinterPriorityHigh,
	"revive":           types.LinterPriorityHigh,
	"nolintlint":       types.LinterPriorityHigh,
	"forcetypeassert":  types.LinterPriorityHigh,
	"gocritic":         types.LinterPriorityHigh,
	"unused":           types.LinterPriorityHigh,
	"bodyclose":        types.LinterPriorityHigh,
	"contextcheck":     types.LinterPriorityHigh,
	"dupl":             types.LinterPriorityHigh,
	"durationcheck":    types.LinterPriorityHigh,
	"errname":          types.LinterPriorityHigh,
	"gochecknoglobals": types.LinterPriorityHigh,
	"gochecknoinits":   types.LinterPriorityHigh,
	"gosmopolitan":     types.LinterPriorityHigh,
	"interfacebloat":   types.LinterPriorityHigh,
	"nestif":           types.LinterPriorityHigh,
	"nilnil":           types.LinterPriorityHigh,
	"nakedret":         types.LinterPriorityHigh,
	"paralleltest":     types.LinterPriorityHigh,
	"predeclared":      types.LinterPriorityHigh,
	"reassign":         types.LinterPriorityHigh,
	"rowserrcheck":     types.LinterPriorityHigh,
	"spancheck":        types.LinterPriorityHigh,
	"sqlclosecheck":    types.LinterPriorityHigh,
	"testifylint":      types.LinterPriorityHigh,
	"thelper":          types.LinterPriorityHigh,
	"unparam":          types.LinterPriorityHigh,
	"wastedassign":     types.LinterPriorityHigh,
	"copyloopvar":      types.LinterPriorityHigh,
	"ginkgolinter":     types.LinterPriorityHigh,
	"gochecksumtype":   types.LinterPriorityHigh,
	"intrange":         types.LinterPriorityHigh,
	"mirror":           types.LinterPriorityHigh,
	"perfsprint":       types.LinterPriorityHigh,
	"protogetter":      types.LinterPriorityHigh,
	"usetesting":       types.LinterPriorityHigh,
	"recvcheck":        types.LinterPriorityHigh,
	"nilnesserr":       types.LinterPriorityHigh,

	// Medium value linters - optional but recommended
	"dupword":                   types.LinterPriorityMedium,
	"godot":                     types.LinterPriorityMedium,
	"godox":                     types.LinterPriorityMedium,
	"goheader":                  types.LinterPriorityMedium,
	"gofmt":                     types.LinterPriorityMedium,
	"gci":                       types.LinterPriorityMedium,
	"varnamelen":                types.LinterPriorityMedium,
	"whitespace":                types.LinterPriorityMedium,
	"wsl_v5":                    types.LinterPriorityMedium,
	"grouper":                   types.LinterPriorityMedium,
	"dogsled":                   types.LinterPriorityMedium,
	"makezero":                  types.LinterPriorityMedium,
	"exportloopref":             types.LinterPriorityMedium,
	"asciicheck":                types.LinterPriorityMedium,
	"bidichk":                   types.LinterPriorityMedium,
	"containedctx":              types.LinterPriorityMedium,
	"decorder":                  types.LinterPriorityMedium,
	"depguard":                  types.LinterPriorityMedium,
	"forbidigo":                 types.LinterPriorityMedium,
	"godoclint":                 types.LinterPriorityMedium,
	"gomoddirectives":           types.LinterPriorityMedium,
	"gomodguard":                types.LinterPriorityMedium,
	"ireturn":                   types.LinterPriorityMedium,
	"lll":                       types.LinterPriorityMedium,
	"mnd":                       types.LinterPriorityMedium,
	"nlreturn":                  types.LinterPriorityMedium,
	"nonamedreturns":            types.LinterPriorityMedium,
	"promlinter":                types.LinterPriorityMedium,
	"tagliatelle":               types.LinterPriorityMedium,
	"testpackage":               types.LinterPriorityMedium,
	"tparallel":                 types.LinterPriorityMedium,
	"unqueryvet":                types.LinterPriorityMedium,
	"usestdlibvars":             types.LinterPriorityMedium,
	"asasalint":                 types.LinterPriorityMedium,
	"canonicalheader":           types.LinterPriorityMedium,
	"err113":                    types.LinterPriorityMedium,
	"exptostd":                  types.LinterPriorityMedium,
	"fatcontext":                types.LinterPriorityMedium,
	"funcorder":                 types.LinterPriorityMedium,
	"gocheckcompilerdirectives": types.LinterPriorityMedium,
	"goprintffuncname":          types.LinterPriorityMedium,
	"iface":                     types.LinterPriorityMedium,
	"imports":                   types.LinterPriorityMedium,
	"inamedparam":               types.LinterPriorityMedium,
	"iotamixing":                types.LinterPriorityMedium,
	"modernize":                 types.LinterPriorityMedium,
	"noinlineerr":               types.LinterPriorityMedium,
	"nosprintfhostport":         types.LinterPriorityMedium,
	"tagalign":                  types.LinterPriorityMedium,
	"testableexamples":          types.LinterPriorityMedium,

	// All other linters default to Optional
}

// LinterReasons provides human-readable reasons for each linter recommendation.
var LinterReasons = map[types.LinterName]string{
	// Critical linters
	"loggercheck": "Checks key value pairs for common logger libraries",
	"gosec":       "Inspects source code for security problems",
	"errcheck":    "Checks for unchecked errors in Go code",
	"staticcheck": "Advanced static analysis (finds bugs, performance issues)",
	"govet":       "Go vet's suspicious construct checks",
	"errchkjson":  "Checks types passed to json encoding functions",
	"musttag":     "Enforces struct tags for JSON/XML/YAML marshaling",
	"sloglint":    "Ensure consistent code style when using log/slog",
	"nilerr":      "Find code that returns nil even if it checks that error is not nil",
	"noctx":       "Check whether function uses a non-inherited context",

	// High value linters
	"wrapcheck":        "Checks that errors returned from external packages are wrapped",
	"errorlint":        "Find code that will cause problems with error wrapping",
	"prealloc":         "Find slice declarations that could potentially be preallocated",
	"unconvert":        "Remove unnecessary type conversions",
	"ineffassign":      "Detect ineffectual assignments",
	"gocyclo":          "Compute cyclomatic complexities",
	"funlen":           "Detect long functions",
	"cyclop":           "Calculate cyclomatic complexities of functions",
	"gocognit":         "Compute cognitive complexities",
	"maintidx":         "Calculate maintenance index",
	"exhaustive":       "Check exhaustiveness of enum switch statements",
	"exhaustruct":      "Check if all struct fields are initialized",
	"goconst":          "Find repeated strings that could be replaced by a constant",
	"misspell":         "Find commonly misspelled English words in comments",
	"revive":           "Fast, configurable, extensible, flexible, and beautiful linter",
	"nolintlint":       "Reports ill-formed or insufficient nolint directives",
	"forcetypeassert":  "Forces type assertions to be safe",
	"gocritic":         "Provides diagnostics for bugs, performance and style issues",
	"unused":           "Checks Go code for unused constants, variables, functions and types",
	"bodyclose":        "Checks whether HTTP response body is closed successfully",
	"contextcheck":     "Check whether the function uses a non-inherited context",
	"dupl":             "Detects duplicate fragments of code",
	"durationcheck":    "Check for two durations multiplied together",
	"errname":          "Checks that sentinel errors are prefixed with Err and error types with Error",
	"gochecknoglobals": "Check that no global variables exist",
	"gochecknoinits":   "Checks that no init functions are present in Go code",
	"gosmopolitan":     "Report i18n/l10n anti-patterns in Go codebase",
	"interfacebloat":   "Checks the number of methods inside an interface",
	"nestif":           "Reports deeply nested if statements",
	"nilnil":           "Checks that there is no simultaneous return of nil error and invalid value",
	"nakedret":         "Checks that functions with naked returns are not too long",
	"paralleltest":     "Detects missing usage of t.Parallel() method in Go tests",
	"predeclared":      "Find code that shadows predeclared identifiers",
	"reassign":         "Checks that package variables are not reassigned",
	"rowserrcheck":     "Checks whether Rows.Err of rows is checked successfully",
	"spancheck":        "Checks for mistakes with OpenTelemetry/Census spans",
	"sqlclosecheck":    "Checks that sql.Rows, sql.Stmt, sqlx.NamedStmt, pgx.Query are closed",
	"testifylint":      "Checks usage of github.com/stretchr/testify",
	"thelper":          "Detects test helpers which do not start with t.Helper()",
	"unparam":          "Reports unused function parameters",
	"wastedassign":     "Finds wasted assignment statements",
	"copyloopvar":      "Detects places where loop variables are copied",
	"ginkgolinter":     "Enforces standards of using ginkgo and gomega",
	"gochecksumtype":   "Run exhaustiveness checks on Go sum types",
	"intrange":         "Finds places where for loops could use integer range",
	"mirror":           "Reports wrong mirror patterns of bytes/strings usage",
	"perfsprint":       "Checks that fmt.Sprintf can be replaced with faster alternative",
	"protogetter":      "Reports direct reads from proto message fields when getters should be used",
	"usetesting":       "Reports uses of functions with replacement inside testing package",
	"recvcheck":        "Checks for receiver type consistency",
	"nilnesserr":       "Reports constructs that check err != nil but return different nil value error",

	// Medium value linters
	"dupword":                   "Checks for duplicate words in the source code",
	"godot":                     "Check if comments end in a period",
	"godox":                     "Check for comments that begin with certain prefixes (TODO, FIXME, etc.)",
	"goheader":                  "Checks that file headers conform to rules",
	"gofmt":                     "Check whether code was gofmt-ed",
	"gci":                       "Check that import order and formatting is correct",
	"varnamelen":                "Check that the length of variable names follows some rules",
	"whitespace":                "Detection of leading and trailing whitespace",
	"wsl_v5":                    "Enforces empty line separation between statements for better readability",
	"grouper":                   "Analyze expression groups",
	"dogsled":                   "Checks assignments with too many identifiers",
	"makezero":                  "Finds slice declarations with non-zero initial lengths",
	"exportloopref":             "Checks for pointers to enclosing loop variables",
	"asciicheck":                "Checks that code identifiers have no non-ASCII symbols",
	"bidichk":                   "Checks for dangerous unicode character sequences",
	"containedctx":              "Detects struct contained context.Context field",
	"decorder":                  "Check declaration order and count of types, constants, variables and functions",
	"depguard":                  "Checks if package imports are in list of acceptable packages",
	"forbidigo":                 "Forbids identifiers",
	"godoclint":                 "Checks Golang's documentation practice (godoc)",
	"gomoddirectives":           "Manage the use of replace, retract, and excludes directives in go.mod",
	"gomodguard":                "Allow and blocklist linter for direct Go module dependencies",
	"ireturn":                   "Accept Interfaces, Return Concrete Types",
	"lll":                       "Reports long lines",
	"mnd":                       "Detects magic numbers",
	"nlreturn":                  "Checks for new line before return and branch statements",
	"nonamedreturns":            "Reports all named returns",
	"promlinter":                "Check Prometheus metrics naming via promlint",
	"tagliatelle":               "Checks the struct tags",
	"testpackage":               "Linter that makes you use separate _test package",
	"tparallel":                 "Detects inappropriate usage of t.Parallel()",
	"unqueryvet":                "Detects SELECT * in SQL queries preventing performance issues",
	"usestdlibvars":             "Detects possibility to use variables/constants from Go standard library",
	"asasalint":                 "Check for pass []any as any in variadic func",
	"canonicalheader":           "Checks whether net/http.Header uses canonical header",
	"err113":                    "Check errors handling expressions",
	"exptostd":                  "Detects functions from golang.org/x/exp/ that can be replaced by std functions",
	"fatcontext":                "Detects nested contexts in loops and function literals",
	"funcorder":                 "Checks the order of functions, methods, and constructors",
	"gocheckcompilerdirectives": "Checks that go compiler directive comments are valid",
	"goprintffuncname":          "Checks that printf-like functions are named with f at the end",
	"iface":                     "Detect incorrect use of interfaces, helping avoid interface pollution",
	"imports":                   "Enforces consistent import aliases",
	"inamedparam":               "Reports interfaces with unnamed method parameters",
	"iotamixing":                "Checks if iotas are used in const blocks with other non-iota declarations",
	"modernize":                 "Suggests simplifications using modern Go language and library features",
	"noinlineerr":               "Disallows inline error handling",
	"nosprintfhostport":         "Checks for misuse of Sprintf to construct host with port in URL",
	"tagalign":                  "Check that struct tags are well aligned",
	"testableexamples":          "Checks if examples are testable (have expected output)",

	// Optional linters (not explicitly listed, will use default message)
	"arangolint":               "Opinionated best practices for arangodb client",
	"zerologlint":              "Detects wrong usage of zerolog that forgets to dispatch with Send or Msg",
	"embeddedstructfieldcheck": "Embedded types should be at top of field list with empty line separation",
}

// DefaultConfigFileNames is a list of default golangci-lint config file names.
var DefaultConfigFileNames = []string{
	".golangci.yml",
	".golangci.yaml",
	".golangci.toml",
	".golangci.json",
}

// FormatterInfo provides metadata about formatters.
var FormatterInfo = map[types.FormatterName]types.FormatterInfo{
	"gci": {
		Name:        "gci",
		Description: "Organizes import statements with additional rules",
		AutoFix:     true,
	},
	"gofmt": {
		Name:        "gofmt",
		Description: "Standard Go code formatting",
		AutoFix:     true,
	},
	"gofumpt": {
		Name:        "gofumpt",
		Description: "Enhanced Go formatting with stricter rules",
		AutoFix:     true,
	},
	"goimports": {
		Name:        "goimports",
		Description: "Formats code and manages import statements",
		AutoFix:     true,
	},
	"golines": {
		Name:        "golines",
		Description: "Formats code and fixes long lines",
		AutoFix:     true,
	},
	"swaggo": {
		Name:        "swaggo",
		Description: "Formats Swagger/OpenAPI documentation comments",
		AutoFix:     true,
	},
}

// FormatterPriorities defines priority levels for formatters.
var FormatterPriorities = map[types.FormatterName]types.FormatterPriority{
	// High priority - recommended for most projects
	"gofumpt":   types.FormatterPriorityHigh,
	"golines":   types.FormatterPriorityHigh,
	"gofmt":     types.FormatterPriorityMedium,
	"goimports": types.FormatterPriorityMedium,
}

// FormatterReasons provides human-readable reasons for each formatter recommendation.
var FormatterReasons = map[types.FormatterName]string{
	"gofumpt":   "Enhanced Go formatting with stricter rules than gofmt",
	"gofmt":     "Standard Go code formatting (consider gofumpt for stricter formatting)",
	"goimports": "Formats code and automatically manages import statements",
	"gci":       "Organizes import statements with custom section rules",
	"golines":   "Formats code and fixes long lines by breaking them",
	"swaggo":    "Formats Swagger/OpenAPI documentation comments",
}

// FormattersManagedByBuildFlow are formatters that should be run by buildflow, not golangci-lint.
var FormattersManagedByBuildFlow = []types.FormatterName{
	"goimports",
	"gofumpt",
}

// RedundantFormatters are formatters that are superseded by other formatters.
var RedundantFormatters = map[types.FormatterName]string{
	"gofmt": "redundant when gofumpt is enabled",
}

// DeprecatedLinters maps deprecated linter names to their recommended replacements.
var DeprecatedLinters = map[types.LinterName]types.LinterReplacement{
	"wsl": {
		Replacement: "wsl_v5",
		Reason:      "wsl is deprecated since golangci-lint v2.2.0, use wsl_v5 instead",
	},
}

// RedundantLinters maps linter names that are superseded by formatters.
var RedundantLinters = map[types.LinterName]string{
	"lll": "redundant when golines formatter is enabled (golines fixes long lines, lll only reports them)",
}

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
