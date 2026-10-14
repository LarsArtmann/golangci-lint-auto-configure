package types

import (
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

// Version is a branded string representing a golangci-lint config schema version.
// It prevents accidental cross-substitution with arbitrary strings.
type Version string

// Valid returns true if the version is a recognized golangci-lint config schema version.
func (v Version) Valid() bool {
	return v == ConfigVersionV2
}

// Compare compares two Versions using semantic version ordering.
// Returns -1, 0, or 1 (like bytes.Compare).
// Both versions are normalized to "v" + version before comparison.
func (v Version) Compare(other Version) int {
	a := "v" + string(v)
	b := "v" + string(other)

	return semver.Compare(a, b)
}

// String returns the underlying string value.
func (v Version) String() string {
	return string(v)
}

// NormalizeGoMajorMinor reduces a Go version string ("go1.27.1", "1.27.1",
// "1.27") to its major.minor component ("1.27"). golangci-lint compares
// run.go against the Go version it was built with at major.minor granularity,
// and language semantics only change per minor release, so the patch component
// carries no meaning for linting. ok is false when the string does not look
// like a Go version (e.g. "devel", "1.x", "").
//
// Deprecated: use ParseGoVersion, which returns the GoVersion branded type
// with the same normalization semantics. This shim forwards to it and exists
// only until the call sites are migrated.
func NormalizeGoMajorMinor(version string) (string, bool) {
	parsed, err := ParseGoVersion(version)
	if err != nil {
		return "", false
	}

	return parsed.String(), true
}

// CompareGoMajorMinor compares two Go versions at major.minor granularity.
// Accepts both raw ("go1.27.1", "1.27") and already-normalized ("1.27") forms.
// Returns -1 when a is older, 0 when equal or either side is unparsable,
// and 1 when a is newer.
//
// Deprecated: use ParseGoVersion plus GoVersion.Compare, which makes the
// "unparsable means equal" hazard explicit at the call site. This shim
// forwards to the branded types and exists only until the call sites are
// migrated.
func CompareGoMajorMinor(a, b string) int {
	aParsed, aErr := ParseGoVersion(a)

	bParsed, bErr := ParseGoVersion(b)

	if aErr != nil || bErr != nil {
		return 0
	}

	return aParsed.Compare(bParsed)
}

func parseGoMajorMinor(version string) (int, int, bool) {
	version = strings.TrimPrefix(version, "go")

	majorStr, rest, found := strings.Cut(version, ".")
	if !found {
		return 0, 0, false
	}

	minorStr, _, _ := strings.Cut(rest, ".")

	major, err := strconv.Atoi(majorStr)
	if err != nil || major < 0 {
		return 0, 0, false
	}

	minor, err := strconv.Atoi(minorStr)
	if err != nil || minor < 0 {
		return 0, 0, false
	}

	return major, minor, true
}
