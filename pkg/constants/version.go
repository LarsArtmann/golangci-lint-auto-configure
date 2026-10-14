package constants

import "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"

// MinGolangCILintVersion is the minimum required version of golangci-lint.
// This is enforced by the version checker to ensure compatibility with
// the linter configurations and JSON output formats we expect.
//
// DERIVED, not hand-maintained: the value is the maximum Since bound over
// every settings block the tool renders (the oldest binary that can load a
// rendered config). The goconst ignore-tests key (v2.12.0) is the current
// binding constraint; adding a curated setting newer than that bumps this
// automatically. Below-minimum binaries get the classified version.too_old
// refusal (upgrade path) instead of a config golangci-lint refuses to load.
var MinGolangCILintVersion = DerivedMinToolVersion().String()

// DerivedMinToolVersion computes the tool minimum from the policy table:
// the maximum Since bound over the settings blocks of every renderable
// record (recommendable and pragmatic-noise tiers — the tiers the renderer
// can emit). Zero means the table carries no versioned settings.
func DerivedMinToolVersion() types.GolangciVersion {
	var minimum types.GolangciVersion

	for _, record := range LinterPolicies {
		if record.Tier.Kind != types.TierRecommendable && record.Tier.Kind != types.TierPragmaticNoise {
			continue
		}

		for _, block := range record.Settings {
			if block.Since.IsZero() {
				continue
			}

			if minimum.IsZero() || block.Since.Compare(minimum) > 0 {
				minimum = block.Since
			}
		}
	}

	return minimum
}

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
// Used to skip recommendations for linters not yet available in the installed
// version. Derived from the policy table's ValidTool lower bounds.
var LinterMinVersions = linterMinVersionsView()
