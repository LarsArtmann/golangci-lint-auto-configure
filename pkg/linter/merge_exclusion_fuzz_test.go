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

// FuzzMergeExclusionLinters pins the union contract on duplicate-heavy input:
// the result is dedup(existing) followed by the new defaults, carries no
// duplicates, and re-merging is a fixed point.
func FuzzMergeExclusionLinters(f *testing.F) {
	f.Add("gosec,gosec,errcheck", "govet,gosec")
	f.Add("", "govet")
	f.Add("errcheck", "")
	f.Add("a,a,a", "a,a")
	f.Add("x,,y", ",z")
	f.Add("linters with spaces , x", "x")

	f.Fuzz(func(t *testing.T, existingRaw, defaultsRaw string) {
		existing := splitLinterList(existingRaw)
		defaults := splitLinterList(defaultsRaw)

		merged := mergeExclusionLinters(existing, defaults)

		want := dedupInOrder(existing)

		seen := make(map[string]bool, len(want))

		for _, s := range want {
			seen[s] = true
		}

		for _, d := range defaults {
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

		if again := mergeExclusionLinters(merged, defaults); len(again) != len(merged) {
			t.Fatalf("re-merge is not a fixed point: %v -> %v", merged, again)
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
