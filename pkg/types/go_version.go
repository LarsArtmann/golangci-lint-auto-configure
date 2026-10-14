package types

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
)

// ErrInvalidGoVersion is wrapped by every Go-version parse failure so callers
// can classify without string matching.
var ErrInvalidGoVersion = errors.New("types: invalid Go version")

// GoVersion is a Go language version at major.minor granularity. The patch
// component is unrepresentable by design: language semantics change per
// minor release, golangci-lint compares run.go at major.minor granularity,
// and "1.27.2" and "1.27" are the same language. Parse-only construction;
// the zero value means "no version" and Parse rejects major 0 so the zero
// value is unreachable from input. Comparison against GolangciVersion is a
// compile error: the two axes answer different questions.
type GoVersion struct {
	major int
	minor int
}

// ParseGoVersion parses a Go version string at major.minor granularity.
// Accepted forms: "go1.27.2", "1.27.2", "1.27", "go1.27" — a patch component
// is accepted and dropped (it carries no language meaning). Rejected:
// "devel", "1.x", "0.x" (the zero value is reserved for "no version"),
// empty and non-numeric input.
func ParseGoVersion(raw string) (GoVersion, error) {
	major, minor, ok := parseGoMajorMinor(raw)
	if !ok || major == 0 {
		return GoVersion{}, fmt.Errorf("%w: %q (want go1.N, 1.N, or 1.N.P)", ErrInvalidGoVersion, raw)
	}

	return GoVersion{major: major, minor: minor}, nil
}

// MustParseGoVersion is ParseGoVersion for curated data where a parse failure
// is a programming error.
func MustParseGoVersion(raw string) GoVersion {
	v, err := ParseGoVersion(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// IsZero reports whether the version is the zero value ("no version").
func (v GoVersion) IsZero() bool { return v.major == 0 && v.minor == 0 }

// String returns the "1.27" form golangci-lint configs use for run.go, ""
// for the zero value.
func (v GoVersion) String() string {
	if v.IsZero() {
		return ""
	}

	return strconv.Itoa(v.major) + "." + strconv.Itoa(v.minor)
}

// Compare orders two versions by major, then minor. The zero value sorts
// before every parsed version, giving the type a total order.
func (v GoVersion) Compare(other GoVersion) int {
	switch {
	case v.IsZero() && other.IsZero():
		return 0
	case v.IsZero():
		return -1
	case other.IsZero():
		return 1
	case v.major != other.major:
		return cmp.Compare(v.major, other.major)
	default:
		return cmp.Compare(v.minor, other.minor)
	}
}

// AtLeast reports whether v is >= other. A zero receiver is never at least
// anything (unknown versions satisfy no minimum).
func (v GoVersion) AtLeast(other GoVersion) bool {
	if v.IsZero() {
		return false
	}

	return v.Compare(other) >= 0
}
