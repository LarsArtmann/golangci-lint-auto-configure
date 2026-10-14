package types

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// Parse-error sentinels for the golangci-lint tool-version axis.
var (
	ErrEmptyGolangciVersion    = errors.New("types: empty golangci-lint version")
	ErrInvalidGolangciVersion  = errors.New("types: invalid golangci-lint version")
)

// GolangciVersion is a golangci-lint tool version (semantic version,
// canonical form "vX.Y.Z"). Parse-only construction: the zero value means
// "no version" and is only ever produced by a zero literal — Parse rejects
// empty and unparsable input. Comparison against GoVersion is a compile
// error: the two axes answer different questions (tool existence/schema
// facts vs language semantics).
type GolangciVersion struct {
	canonical string
}

// ParseGolangciVersion parses a golangci-lint version string. Accepted forms:
// "v2.14.0", "2.14.0", "v2.14.0-rc1" — semver-valid with all three
// major.minor.patch components (golangci-lint always reports full versions;
// "v2" or "v2.14" are rejected as ambiguous). The stored form is always the
// canonical "v"-prefixed string.
func ParseGolangciVersion(raw string) (GolangciVersion, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return GolangciVersion{}, ErrEmptyGolangciVersion
	}

	canonical := "v" + strings.TrimPrefix(trimmed, "v")

	if !semver.IsValid(canonical) || !hasFullSemverCore(canonical) {
		return GolangciVersion{}, fmt.Errorf("%w: %q", ErrInvalidGolangciVersion, raw)
	}

	return GolangciVersion{canonical: canonical}, nil
}

// hasFullSemverCore reports whether the canonical semver has all three
// major.minor.patch components before any prerelease suffix.
func hasFullSemverCore(canonical string) bool {
	core := canonical[1:]

	if idx := strings.IndexAny(core, "-+"); idx >= 0 {
		core = core[:idx]
	}

	return strings.Count(core, ".") == 2
}

// MustParseGolangciVersion is ParseGolangciVersion for curated data where a
// parse failure is a programming error (compile-time constants, tests).
func MustParseGolangciVersion(raw string) GolangciVersion {
	v, err := ParseGolangciVersion(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// IsZero reports whether the version is the zero value ("no version").
func (v GolangciVersion) IsZero() bool { return v.canonical == "" }

// String returns the canonical "vX.Y.Z" form, "" for the zero value.
func (v GolangciVersion) String() string { return v.canonical }

// Compare orders two versions semantically. The zero value sorts before every
// parsed version, giving the type a total order (zero = "no version yet").
func (v GolangciVersion) Compare(other GolangciVersion) int {
	switch {
	case v.IsZero() && other.IsZero():
		return 0
	case v.IsZero():
		return -1
	case other.IsZero():
		return 1
	default:
		return semver.Compare(v.canonical, other.canonical)
	}
}

// AtLeast reports whether v is >= other. A zero receiver is never at least
// anything (unknown versions satisfy no minimum).
func (v GolangciVersion) AtLeast(other GolangciVersion) bool {
	if v.IsZero() {
		return false
	}

	return v.Compare(other) >= 0
}
