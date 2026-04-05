package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// DefaultTimeout is the default run timeout value for golangci-lint configurations.
const DefaultTimeout = "5m"

// DefaultConfigFileNames is a list of default golangci-lint config file names.
var DefaultConfigFileNames = []string{
	".golangci.yml",
	".golangci.yaml",
	".golangci.toml",
	".golangci.json",
}

// FormattersManagedByBuildFlow are formatters that should be run by buildflow, not golangci-lint.
var FormattersManagedByBuildFlow = []types.FormatterName{
	"goimports",
	"gofumpt",
}

// CoreFormatters are the formatters enabled by default for all projects.
var CoreFormatters = []string{"gci", "gofumpt", "goimports"}

// FormatterOrder defines the canonical ordering of formatters in config output.
var FormatterOrder = []string{"gci", "goimports", "gofumpt", "golines", "swaggo"}

// RedundantFormatters are formatters that are superseded by other formatters.
var RedundantFormatters = map[types.FormatterName]string{
	"gofmt": "redundant when gofumpt is enabled",
}
