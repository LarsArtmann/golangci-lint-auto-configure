package constants

import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"

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
