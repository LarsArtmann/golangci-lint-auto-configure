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

	// Removed in golangci-lint v2 — have replacements
	"exportloopref": {
		Replacement: "copyloopvar",
		Reason:      "exportloopref was removed in golangci-lint v2 (Go 1.22 fixed loop variable semantics), use copyloopvar instead",
	},
	"golint": {
		Replacement: "revive",
		Reason:      "golint was removed in golangci-lint v2, use revive instead",
	},
	"scopelint": {
		Replacement: "copyloopvar",
		Reason:      "scopelint was removed in golangci-lint v2, use copyloopvar instead",
	},
	"tenv": {
		Replacement: "usetesting",
		Reason:      "tenv was removed in golangci-lint v2, use usetesting instead",
	},

	// Removed in golangci-lint v2 — no direct replacement
	"ifshort": {
		Replacement: "",
		Reason:      "ifshort was removed in golangci-lint v2 with no direct replacement",
	},
	"execinquery": {
		Replacement: "",
		Reason:      "execinquery was removed in golangci-lint v2 with no direct replacement",
	},

	// v1 alternative names (renamed before v2)
	"gas": {
		Replacement: "gosec",
		Reason:      "gas was renamed to gosec",
	},
	"goerr113": {
		Replacement: "err113",
		Reason:      "goerr113 was renamed to err113",
	},
	"gomnd": {
		Replacement: "mnd",
		Reason:      "gomnd was renamed to mnd",
	},
	"logrlint": {
		Replacement: "loggercheck",
		Reason:      "logrlint was renamed to loggercheck",
	},
	"megacheck": {
		Replacement: "staticcheck",
		Reason:      "megacheck was renamed to staticcheck",
	},
	"vet": {
		Replacement: "govet",
		Reason:      "vet was renamed to govet",
	},
	"vetshadow": {
		Replacement: "govet",
		Reason:      "vetshadow was merged into govet",
	},
}

// DisabledLinters maps linter names that should never be recommended or enabled
// to the reason for disabling them. These linters are explicitly excluded from
// configuration by the tool.
var DisabledLinters = map[types.LinterName]string{
	"funcorder":   "provides minimal value and can be confusing for users",
	"noinlineerr": "conflicts with formatters (gofumpt, goimports) that reformat error handling expressions, causing noisy churn and contradictory findings",
	"depguard":    "superseded by the dedicated library-policy tool (github.com/LarsArtmann/library-policy) which provides AST-based banned-library governance across all projects; depguard's per-config allow-list model is redundant and weaker",
}

// RedundantLinters maps linter names that are superseded by formatters.
var RedundantLinters = map[types.LinterName]types.LinterToFormatter{
	"lll": {
		Formatter: "golines",
		Reason:    "redundant when golines formatter is enabled (golines fixes long lines, lll only reports them)",
	},
}

// PragmaticNoiseLinters is the set of high-friction linters dropped from the
// dynamic enable set when --pragmatic is used. These five have the highest
// friction ratios (exhaustruct 6.5, gochecknoglobals 5.1, ireturn 1.3,
// wrapcheck 1.8, funlen 0.83) and are the most commonly cited sources of
// linting friction. They stay enabled by default; --pragmatic is an opt-in
// escape hatch for projects that find them too noisy.
var PragmaticNoiseLinters = map[types.LinterName]string{
	"exhaustruct":      "forces exhaustive struct literals on every http.Server{}, Cmd{}, etc. (friction 6.5)",
	"gochecknoglobals": "fights standard Go patterns like registries and sentinels (friction 5.1, no config knobs)",
	"wrapcheck":        "demands every error be wrapped (friction 1.8)",
	"ireturn":          "conflicts with common interface-returning APIs (friction 1.3)",
	"funlen":           "default thresholds too strict vs. house style (friction 0.83)",
}
