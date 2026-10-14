package types_test

import (
	"testing"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// FuzzNormalizeGoMajorMinor pins the normalization contract: a parsable
// version reduces to major.minor and re-normalizing the result is a fixed
// point; an unparsable version yields ("", false).
func FuzzNormalizeGoMajorMinor(f *testing.F) {
	f.Add("go1.27.1")
	f.Add("1.27")
	f.Add("1.27.0")
	f.Add("devel")
	f.Add("go1.28-0f9a9bc")
	f.Add("1")
	f.Add("")
	f.Add("1.x")
	f.Add("go")

	f.Fuzz(func(t *testing.T, version string) {
		normalized, ok := types.NormalizeGoMajorMinor(version)
		if !ok {
			if normalized != "" {
				t.Fatalf("unparsable %q must return empty, got %q", version, normalized)
			}

			return
		}

		if normalized == "" {
			t.Fatalf("parsable %q must not normalize to empty", version)
		}

		again, okAgain := types.NormalizeGoMajorMinor(normalized)
		if !okAgain || again != normalized {
			t.Fatalf("normalization is not idempotent: %q -> %q -> (%q, %v)", version, normalized, again, okAgain)
		}
	})
}

// FuzzCompareGoMajorMinor pins the comparison contract: antisymmetry,
// reflexivity, patch-invariance (a version equals its own normalization), and
// agreement with normalization.
func FuzzCompareGoMajorMinor(f *testing.F) {
	f.Add("1.27", "1.26")
	f.Add("1.27.0", "go1.27.9")
	f.Add("1.9", "1.10")
	f.Add("devel", "1.27")
	f.Add("2.0", "1.99")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, a, b string) {
		ab := types.CompareGoMajorMinor(a, b)
		ba := types.CompareGoMajorMinor(b, a)

		if ab != -ba {
			t.Fatalf("antisymmetry violated: Compare(%q,%q)=%d, Compare(%q,%q)=%d", a, b, ab, b, a, ba)
		}

		if aa := types.CompareGoMajorMinor(a, a); aa != 0 {
			t.Fatalf("reflexivity violated: Compare(%q,%q)=%d", a, a, aa)
		}

		normalizedA, okA := types.NormalizeGoMajorMinor(a)
		if okA && types.CompareGoMajorMinor(a, normalizedA) != 0 {
			t.Fatalf("patch invariance violated: %q != its normalization %q", a, normalizedA)
		}

		normalizedB, okB := types.NormalizeGoMajorMinor(b)
		if okA && okB {
			if want := types.CompareGoMajorMinor(normalizedA, normalizedB); ab != want {
				t.Fatalf("patch component changed the verdict: Compare(%q,%q)=%d, want %d",
					a, b, ab, want)
			}
		}
	})
}

// FuzzParseGoVersion pins the branded-type parse contract: parse-or-error is
// total, a parsed version re-parses from its own String form (round-trip),
// and patch input collapses to the same major.minor value.
func FuzzParseGoVersion(f *testing.F) {
	f.Add("go1.27.1")
	f.Add("1.27")
	f.Add("1.27.0")
	f.Add("devel")
	f.Add("go1.28-0f9a9bc")
	f.Add("1")
	f.Add("")
	f.Add("1.x")
	f.Add("go")
	f.Add("0.42")
	f.Add("2.0.0")

	f.Fuzz(func(t *testing.T, version string) {
		parsed, err := types.ParseGoVersion(version)
		if err != nil {
			if !parsed.IsZero() {
				t.Fatalf("unparsable %q must return the zero value, got %q", version, parsed.String())
			}

			return
		}

		if parsed.String() == "" {
			t.Fatalf("parsable %q must not stringify to empty", version)
		}

		again, errAgain := types.ParseGoVersion(parsed.String())
		if errAgain != nil || again != parsed {
			t.Fatalf("round-trip broken: %q -> %q -> (%q, %v)", version, parsed.String(), again.String(), errAgain)
		}

		if !parsed.AtLeast(parsed) {
			t.Fatalf("AtLeast violated for reflexive case: %q", parsed.String())
		}
	})
}

// FuzzParseGolangciVersion pins the tool-version parse contract: parse-or-error
// is total, canonical strings are fixed points, and Compare is antisymmetric.
func FuzzParseGolangciVersion(f *testing.F) {
	f.Add("v2.14.0")
	f.Add("2.12.0")
	f.Add("v2.14.0-rc1")
	f.Add("v2")
	f.Add("2.14")
	f.Add("")
	f.Add("abc")
	f.Add("v2.x.0")

	f.Fuzz(func(t *testing.T, version string) {
		parsed, err := types.ParseGolangciVersion(version)
		if err != nil {
			if !parsed.IsZero() {
				t.Fatalf("unparsable %q must return the zero value, got %q", version, parsed.String())
			}

			return
		}

		again, errAgain := types.ParseGolangciVersion(parsed.String())
		if errAgain != nil || again != parsed {
			t.Fatalf("round-trip broken: %q -> %q -> (%q, %v)", version, parsed.String(), again.String(), errAgain)
		}

		other, otherErr := types.ParseGolangciVersion(parsed.String())
		if otherErr != nil {
			t.Fatalf("re-parse failed: %v", otherErr)
		}

		if c := parsed.Compare(other); c != 0 {
			t.Fatalf("reflexivity violated: Compare(%q,%q)=%d", parsed.String(), other.String(), c)
		}
	})
}
