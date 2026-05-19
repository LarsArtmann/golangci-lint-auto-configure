package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// DeprecatedLinters maps deprecated linter names to their recommended replacements.
var DeprecatedLinters = map[types.LinterName]types.LinterReplacement{
	"wsl": {
		Replacement: "wsl_v5",
		Reason:      "wsl is deprecated since golangci-lint v2.2.0, use wsl_v5 instead",
	},
	"deadcode": {
		Replacement: "staticcheck",
		Reason:      "deadcode was removed in golangci-lint v2, use staticcheck instead",
	},
	"varcheck": {
		Replacement: "staticcheck",
		Reason:      "varcheck was removed in golangci-lint v2, use staticcheck instead",
	},
	"structcheck": {
		Replacement: "staticcheck",
		Reason:      "structcheck was removed in golangci-lint v2, use staticcheck instead",
	},
	"gosimple": {
		Replacement: "staticcheck",
		Reason:      "gosimple was removed in golangci-lint v2, use staticcheck instead",
	},
	"exhaustivestruct": {
		Replacement: "exhaustive",
		Reason:      "exhaustivestruct was renamed to exhaustive in golangci-lint v2",
	},
	"interfacer": {
		Replacement: "staticcheck",
		Reason:      "interfacer was removed in golangci-lint v2, use staticcheck instead",
	},
	"maligned": {
		Replacement: "govet",
		Reason:      "maligned was removed in golangci-lint v2, use govet with fieldalignment instead",
	},
	"nosnakecase": {
		Replacement: "revive",
		Reason:      "nosnakecase was removed in golangci-lint v2, use revive instead",
	},
	"gomodguard": {
		Replacement: "gomodguard_v2",
		Reason:      "gomodguard is deprecated since golangci-lint v2.12.0, use gomodguard_v2 instead",
		MinVersion:  "v2.12.0",
	},
}

// DisabledLinters is a set of linters that should never be recommended or enabled.
// These linters are explicitly excluded from configuration by the tool.
var DisabledLinters = types.NewSet[types.LinterName]("funcorder")

// RedundantLinters maps linter names that are superseded by formatters.
var RedundantLinters = map[types.LinterName]types.LinterToFormatter{
	"lll": {
		Formatter: "golines",
		Reason:    "redundant when golines formatter is enabled (golines fixes long lines, lll only reports them)",
	},
}
