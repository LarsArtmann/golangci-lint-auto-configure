// Package policy loads and evaluates the disable-reason sidecar file
// (.golangci-lint-auto-configure.yml) that justifies intentional linter disables.
//
// When a sidecar file exists and contains a justification for a disabled linter,
// the tool respects the disable. When a sidecar exists but a disabled linter has
// no entry, the tool re-enables the linter (anti-gaming enforcement). When no
// sidecar exists, the tool respects all disables as before (backward compatible).
package policy

import (
	"fmt"
	"os"

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
	Disabled map[types.LinterName]DisableJustification `yaml:"disabled"`
}

// Load reads the sidecar file at path. Returns (nil, nil) when the file does
// not exist (no enforcement). Returns an error only for read or parse failures.
func Load(path string) (*Policy, error) { //nolint:erraudit // advisory: errors classified at command boundary via go-error-family, not per-function types
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil //nolint:nilnil // intentional: nil policy = no sidecar, nil error = success
		}

		return nil, fmt.Errorf("read policy file %q: %w", path, err)
	}

	var parsed Policy

	err = yaml.Unmarshal(data, &parsed)
	if err != nil {
		return nil, fmt.Errorf("parse policy file %q: %w", path, err)
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
