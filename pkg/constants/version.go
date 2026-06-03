// Package constants provides version information for the tool and minimum version requirements.
package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// MinGolangCILintVersion is the minimum required version of golangci-lint.
// This is enforced by the version checker to ensure compatibility with
// the linter configurations and JSON output formats we expect.
const MinGolangCILintVersion = "v2.10.1"

// ExpectedGolangCILintVersion is the recommended golangci-lint version.
// A warning is emitted when the detected version differs from this value,
// as the tool is tested and developed against this specific version.
const ExpectedGolangCILintVersion = "v2.12.2"

// LinterMinVersions maps linter names to the minimum golangci-lint version
// that supports them. Linters not in this map are available in all versions.
// Used to skip recommendations for linters not yet available in the installed version.
var LinterMinVersions = map[types.LinterName]string{
	"gomodguard_v2": "v2.12.0", //nolint:goconst // linter name used across maps
}
