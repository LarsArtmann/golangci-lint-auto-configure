package linter

import (
	"strings"
	"testing"
)

func dedupInOrder(in []string) []string {
	seen := make(map[string]bool, len(in))

	out := make([]string, 0, len(in))

	for _, s := range in {
		if !seen[s] {
			seen[s] = true

			out = append(out, s)
		}
	}

	return out
}

// FuzzUnionStrings pins the default-exclusion-rule union contract on
// duplicate-heavy input: the result is existing (in order) followed by the
// new entries, carries no duplicates, and re-unioning is a fixed point.
func FuzzUnionStrings(f *testing.F) {
	f.Add("gosec,gosec,errcheck", "govet,gosec")
	f.Add("", "govet")
	f.Add("errcheck", "")
	f.Add("a,a,a", "a,a")
	f.Add("x,,y", ",z")
	f.Add("linters with spaces , x", "x")

	f.Fuzz(func(t *testing.T, existingRaw, additionsRaw string) {
		existing := splitLinterList(existingRaw)
		additions := splitLinterList(additionsRaw)

		merged, _ := unionStrings(existing, additions)

		want := dedupInOrder(existing)

		seen := make(map[string]bool, len(want))

		for _, s := range want {
			seen[s] = true
		}

		for _, d := range additions {
			if !seen[d] {
				seen[d] = true

				want = append(want, d)
			}
		}

		if len(merged) != len(want) {
			t.Fatalf("length mismatch: got %v, want %v", merged, want)
		}

		for i := range want {
			if merged[i] != want[i] {
				t.Fatalf("position %d: got %q, want %q (merged=%v, want=%v)", i, merged[i], want[i], merged, want)
			}
		}

		mergedAgain, _ := unionStrings(merged, additions)
		if len(mergedAgain) != len(merged) {
			t.Fatalf("re-union is not a fixed point: %v -> %v", merged, mergedAgain)
		}
	})
}

func splitLinterList(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}

	return out
}
