package types_test

import (
	"fmt"
	"testing"

	. "github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

// makeStrings generates a slice of n strings.
func makeStrings(n int) []string {
	return makeStringsOffset(0, n)
}

// makeStringsOffset generates a slice of n strings starting from offset.
func makeStringsOffset(offset, n int) []string {
	result := make([]string, 0, n)
	for i := range n {
		result = append(result, fmt.Sprintf("element_%d", offset+i))
	}

	return result
}

// BenchmarkAdd benchmarks adding elements to sets of different sizes.
func BenchmarkAdd(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			set := NewSet(makeStrings(size)...)

			b.ResetTimer()

			for range b.N {
				set.Add("new-element")
			}
		})
	}
}

// BenchmarkContains benchmarks membership checks in sets of different sizes.
func BenchmarkContains(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			set := NewSet(makeStrings(size)...)
			element := fmt.Sprintf("element_%d", size/2) // Middle element

			b.ResetTimer()

			for range b.N {
				set.Contains(element)
			}
		})
	}
}

// BenchmarkUnion benchmarks union operations on sets of different sizes.
func BenchmarkUnion(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			primary := NewSet(makeStrings(size)...)
			secondary := NewSet(makeStringsOffset(size, size)...)

			b.ResetTimer()

			for range b.N {
				primary.Union(secondary)
			}
		})
	}
}

// BenchmarkIntersect benchmarks intersection operations on sets of different sizes.
func BenchmarkIntersect(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			primary := NewSet(makeStrings(size)...)
			secondary := NewSet(makeStringsOffset(size/2, size)...)

			b.ResetTimer()

			for range b.N {
				primary.Intersect(secondary)
			}
		})
	}
}

// BenchmarkDifference benchmarks difference operations on sets of different sizes.
func BenchmarkDifference(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			primary := NewSet(makeStrings(size)...)
			secondary := NewSet(makeStringsOffset(size/2, size)...)

			b.ResetTimer()

			for range b.N {
				primary.Difference(secondary)
			}
		})
	}
}

// BenchmarkIsSubset benchmarks subset checks on sets of different sizes.
func BenchmarkIsSubset(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			subset := NewSet(makeStrings(size / 2)...)
			superset := NewSet(makeStrings(size)...)

			b.ResetTimer()

			for range b.N {
				subset.IsSubset(superset)
			}
		})
	}
}
