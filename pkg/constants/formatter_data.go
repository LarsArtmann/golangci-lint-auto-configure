package constants

import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"

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
