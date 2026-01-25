package constants

import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"

// LinterPriorities maps linter names to their priority levels
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
	"wrapcheck":       types.LinterPriorityHigh,
	"errorlint":       types.LinterPriorityHigh,
	"prealloc":        types.LinterPriorityHigh,
	"unconvert":       types.LinterPriorityHigh,
	"ineffassign":     types.LinterPriorityHigh,
	"gocyclo":         types.LinterPriorityHigh,
	"funlen":          types.LinterPriorityHigh,
	"cyclop":          types.LinterPriorityHigh,
	"gocognit":        types.LinterPriorityHigh,
	"maintidx":        types.LinterPriorityHigh,
	"exhaustive":      types.LinterPriorityHigh,
	"exhaustruct":     types.LinterPriorityHigh,
	"goconst":         types.LinterPriorityHigh,
	"misspell":        types.LinterPriorityHigh,
	"revive":          types.LinterPriorityHigh,
	"nolintlint":      types.LinterPriorityHigh,
	"forcetypeassert": types.LinterPriorityHigh,

	// Medium value linters - optional but recommended
	"dupword":       types.LinterPriorityMedium,
	"godot":         types.LinterPriorityMedium,
	"godox":         types.LinterPriorityMedium,
	"goheader":      types.LinterPriorityMedium,
	"gofmt":         types.LinterPriorityMedium,
	"gci":           types.LinterPriorityMedium,
	"varnamelen":    types.LinterPriorityMedium,
	"lll":           types.LinterPriorityMedium,
	"whitespace":    types.LinterPriorityMedium,
	"grouper":       types.LinterPriorityMedium,
	"dogsled":       types.LinterPriorityMedium,
	"makezero":      types.LinterPriorityMedium,
	"thelper":       types.LinterPriorityMedium,
	"exportloopref": types.LinterPriorityMedium,
	"paralleltest":  types.LinterPriorityMedium,

	// All other linters default to Optional
}

// LinterReasons provides human-readable reasons for each linter recommendation
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
	"wrapcheck":       "Checks that errors returned from external packages are wrapped",
	"errorlint":       "Find code that will cause problems with error wrapping",
	"prealloc":        "Find slice declarations that could potentially be preallocated",
	"unconvert":       "Remove unnecessary type conversions",
	"ineffassign":     "Detect ineffectual assignments",
	"gocyclo":         "Compute cyclomatic complexities",
	"funlen":          "Detect long functions",
	"cyclop":          "Calculate cyclomatic complexities of functions",
	"gocognit":        "Compute cognitive complexities",
	"maintidx":        "Calculate maintenance index",
	"exhaustive":      "Check exhaustiveness of enum switch statements",
	"exhaustruct":     "Check if all struct fields are initialized",
	"goconst":         "Find repeated strings that could be replaced by a constant",
	"misspell":        "Find commonly misspelled English words in comments",
	"revive":          "Fast, configurable, extensible, flexible, and beautiful linter",
	"nolintlint":      "Reports ill-formed or insufficient nolint directives",
	"forcetypeassert": "Forces type assertions to be safe",

	// Medium value linters
	"dupword":       "Checks for duplicate words in the source code",
	"godot":         "Check if comments end in a period",
	"godox":         "Check for comments that begin with certain prefixes (TODO, FIXME, etc.)",
	"goheader":      "Checks that file headers conform to rules",
	"gofmt":         "Check whether code was gofmt-ed",
	"gci":           "Check that import order and formatting is correct",
	"varnamelen":    "Check that the length of variable names follows some rules",
	"lll":           "Check line length",
	"whitespace":    "Detection of leading and trailing whitespace",
	"grouper":       "Analyze expression groups",
	"dogsled":       "Checks assignments with too many identifiers",
	"makezero":      "Finds slice declarations with non-zero initial lengths",
	"thelper":       "Check for correct usage of test helper functions",
	"exportloopref": "Checks for pointers to enclosing loop variables",
	"paralleltest":  "Detects inappropriate usage of t.Parallel()",
}

// DefaultConfigFileNames is a list of default golangci-lint config file names
var DefaultConfigFileNames = []string{
	".golangci.yml",
	".golangci.yaml",
	".golangci.toml",
	".golangci.json",
}

// FormattersManagedByBuildFlow are formatters that should be run by buildflow, not golangci-lint
var FormattersManagedByBuildFlow = []types.LinterName{
	"goimports",
	"gofumpt",
}

// RedundantFormatters are formatters that are superseded by other formatters
var RedundantFormatters = map[types.LinterName]string{
	"gofmt": "redundant when gofumpt is enabled",
}
