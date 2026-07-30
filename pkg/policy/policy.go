// Package policy loads and evaluates the disable-reason sidecar file
// (.golangci-lint-auto-configure.yml) that justifies intentional linter disables
// and never-enable directives.
//
// When a sidecar file exists and contains a justification for a disabled linter,
// the tool respects the disable. When a sidecar exists but a disabled linter has
// no entry, the tool re-enables the linter (anti-gaming enforcement). When no
// sidecar exists, the tool respects all disables as before (backward compatible).
//
// The neverEnable section lists linters that the tool must never add to the
// enable list, even if they are absent from both enable and disable. This is the
// durable, committed-to-git signal for "I intentionally removed this linter from
// enable — don't re-add it." It breaks the regression loop where configure
// re-adds linters the user deliberately omitted (in golangci-lint v2, omitting a
// linter from enable is the documented way to disable it).
package policy

import (
	"os"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
	"go.yaml.in/yaml/v3"
)

// SidecarFileName is the name of the policy sidecar file placed alongside .golangci.yml.
const SidecarFileName = ".golangci-lint-auto-configure.yml"

// DisableCategory classifies why a linter was intentionally disabled.
type DisableCategory string

const (
	CategoryFalsePositives DisableCategory = "false-positives"
	CategorySuperseded     DisableCategory = "superseded"
	CategoryConvention     DisableCategory = "convention"
	CategoryPerformance    DisableCategory = "performance"
	CategoryOther          DisableCategory = "other"
)

// DisableJustification pairs a human-readable reason with a category.
type DisableJustification struct {
	Reason   string          `yaml:"reason"`
	Category DisableCategory `yaml:"category"`
}

// Policy is the parsed sidecar file. A nil *Policy means no sidecar was found
// (enforcement is inactive; all disables are respected).
type Policy struct {
	Disabled   map[types.LinterName]DisableJustification `yaml:"disabled"`
	NeverEnable map[types.LinterName]DisableJustification `yaml:"neverEnable"`
}

// Load reads the sidecar file at path. Returns (nil, nil) when the file does
// not exist (no enforcement). Returns an error only for read or parse failures.
func Load(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil //nolint:nilnil // intentional: nil policy = no sidecar, nil error = success
		}

		return nil, errorfamily.WrapTransientf(err, "policy.read",
			"read policy file %q", path)
	}

	var parsed Policy

	err = yaml.Unmarshal(data, &parsed)
	if err != nil {
		return nil, errorfamily.WrapRejectionf(err, "policy.parse",
			"parse policy file %q", path)
	}

	return &parsed, nil
}

// IsJustified reports whether the given linter has a justification entry.
// Returns false when the policy is nil (no sidecar) — callers should check
// for nil policy separately to distinguish "no enforcement" from "unjustified".
func (p *Policy) IsJustified(linter types.LinterName) bool {
	if p == nil {
		return false
	}

	_, ok := p.Disabled[linter]

	return ok
}

// Justification returns the justification for the given linter, or false.
func (p *Policy) Justification(linter types.LinterName) (DisableJustification, bool) {
	if p == nil {
		return DisableJustification{}, false
	}

	just, ok := p.Disabled[linter]

	return just, ok
}

// IsNeverEnable reports whether the given linter is listed in the neverEnable
// section of the sidecar. Linters in this section must never be added to the
// enable list by the tool, even when absent from both enable and disable.
// Returns false when the policy is nil (no sidecar).
func (p *Policy) IsNeverEnable(linter types.LinterName) bool {
	if p == nil {
		return false
	}

	_, ok := p.NeverEnable[linter]

	return ok
}

// NeverEnableJustification returns the justification for a never-enable entry,
// or false.
func (p *Policy) NeverEnableJustification(linter types.LinterName) (DisableJustification, bool) {
	if p == nil {
		return DisableJustification{}, false
	}

	just, ok := p.NeverEnable[linter]

	return just, ok
}
