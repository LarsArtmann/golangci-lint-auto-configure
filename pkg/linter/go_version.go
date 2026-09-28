package linter

import (
	"cmp"
	"strconv"
	"strings"
)

// normalizeGoMajorMinor reduces a Go version string ("go1.27.1", "1.27.1",
// "1.27") to its major.minor component ("1.27"). golangci-lint compares
// run.go against the Go version it was built with at major.minor granularity,
// and language semantics only change per minor release, so the patch component
// carries no meaning for linting. ok is false when the string does not look
// like a Go version (e.g. "devel", "1.x", "").
func normalizeGoMajorMinor(version string) (majorMinor string, ok bool) {
	major, minor, ok := parseGoMajorMinor(version)
	if !ok {
		return "", false
	}

	return strconv.Itoa(major) + "." + strconv.Itoa(minor), true
}

// compareGoMajorMinor compares two Go versions at major.minor granularity.
// Accepts both raw ("go1.27.1", "1.27") and already-normalized ("1.27") forms.
// Returns -1 when a is older, 0 when equal or either side is unparsable,
// and 1 when a is newer.
func compareGoMajorMinor(a, b string) int {
	aMajor, aMinor, aOK := parseGoMajorMinor(a)
	bMajor, bMinor, bOK := parseGoMajorMinor(b)

	if !aOK || !bOK {
		return 0
	}

	if c := cmp.Compare(aMajor, bMajor); c != 0 {
		return c
	}

	return cmp.Compare(aMinor, bMinor)
}

func parseGoMajorMinor(version string) (major, minor int, ok bool) {
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

	minor, err = strconv.Atoi(minorStr)
	if err != nil || minor < 0 {
		return 0, 0, false
	}

	return major, minor, true
}
