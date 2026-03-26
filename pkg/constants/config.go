package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

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

// RedundantFormatters are formatters that are superseded by other formatters.
var RedundantFormatters = map[types.FormatterName]string{
	"gofmt": "redundant when gofumpt is enabled",
}
