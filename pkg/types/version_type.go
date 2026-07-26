package types

import "golang.org/x/mod/semver"

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
