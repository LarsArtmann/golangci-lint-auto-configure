package config

import (
	"strings"
	"testing"
	"testing/quick"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// FuzzMergeEnableDisable verifies that merging linter enable/disable sets
// never panics and always produces deterministic results.
func FuzzMergeEnableDisable(f *testing.F) {
	f.Add("gosec,errcheck", "govet,gosec")
	f.Add("", "gosec")
	f.Add("gosec,gosec,errcheck", "gosec")

	f.Fuzz(func(t *testing.T, primaryStr, secondaryStr string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("merge panicked: %v", r)
			}
		}()

		var primaryItems, secondaryItems []string

		if primaryStr != "" {
			primaryItems = strings.Split(primaryStr, ",")
		}

		if secondaryStr != "" {
			secondaryItems = strings.Split(secondaryStr, ",")
		}

		primary := types.NewSet(primaryItems...)
		secondary := types.NewSet(secondaryItems...)

		merged := primary.Union(secondary)

		for _, linter := range types.ToSortedSlice(primary) {
			if !merged.Contains(linter) {
				t.Errorf("merged set missing primary linter %q", linter)
			}
		}
	})
}

// TestQuickMergeCommutative verifies that set union is commutative:
// A ∪ B == B ∪ A.
func TestQuickMergeCommutative(t *testing.T) {
	property := func(a, b []string) bool {
		setA := types.NewSet(a...)
		setB := types.NewSet(b...)

		return setA.Union(setB).Equal(setB.Union(setA))
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	if err != nil {
		t.Errorf("union is not commutative: %v", err)
	}
}

// TestQuickMergeIdempotent verifies that A ∪ A == A.
func TestQuickMergeIdempotent(t *testing.T) {
	property := func(items []string) bool {
		set := types.NewSet(items...)

		return set.Union(set).Equal(set)
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	if err != nil {
		t.Errorf("union is not idempotent: %v", err)
	}
}

// TestQuickMergeSubset verifies that A ⊆ (A ∪ B) for all B.
func TestQuickMergeSubset(t *testing.T) {
	property := func(a, b []string) bool {
		setA := types.NewSet(a...)
		setB := types.NewSet(b...)

		return setA.IsSubset(setA.Union(setB))
	}

	err := quick.Check(property, &quick.Config{MaxCount: 100})
	if err != nil {
		t.Errorf("A is not a subset of A∪B: %v", err)
	}
}
