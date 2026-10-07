// Package constants provides version information for the tool and minimum version requirements.
package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// MinGolangCILintVersion is the minimum required version of golangci-lint.
// This is enforced by the version checker to ensure compatibility with
// the linter configurations and JSON output formats we expect.
// Bumped from v2.10.1 to v2.12.0 (2026-10-07): the curated goconst default
// injects `ignore-tests`, a settings key that only exists since v2.12.0.
// On older binaries the tool would write a config golangci-lint refuses to
// load at all — a classified refusal with an upgrade path is strictly safer.
// (exhaustruct_v5 needs v2.13.0 but is never auto-enabled, so it does not
// constrain the minimum.)
const MinGolangCILintVersion = "v2.12.0"

// ExpectedGolangCILintVersion is the recommended golangci-lint version.
// A warning is emitted when the detected version differs from this value,
// as the tool is tested and developed against this specific version.
// v2.13.2 was the first line built with Go 1.27 — older binaries reject
// configs whose run.go targets Go 1.27+. The tool is developed and tested
// against v2.14.0 (CI pin, schema snapshot), whose release binaries are
// also built with Go 1.27.
const ExpectedGolangCILintVersion = "v2.14.0"

// LinterMinVersions maps linter names to the minimum golangci-lint version
// that supports them. Linters not in this map are available in all versions.
// Used to skip recommendations for linters not yet available in the installed version.
var LinterMinVersions = map[types.LinterName]string{
	"gomodguard_v2":  "v2.12.0",
	"clickhouselint": "v2.12.0",
	"exhaustruct_v5": "v2.13.0",
}
