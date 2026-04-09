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
	result := make([]string, n)
	for i := range n {
		result[i] = fmt.Sprintf("element_%d", offset+i)
	}

	return result
}

// BenchmarkAdd benchmarks adding elements to sets of different sizes.
func BenchmarkAdd(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s := NewSet(makeStrings(size)...)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				s.Add("new-element")
			}
		})
	}
}

// BenchmarkContains benchmarks membership checks in sets of different sizes.
func BenchmarkContains(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s := NewSet(makeStrings(size)...)
			element := fmt.Sprintf("element_%d", size/2) // Middle element

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				s.Contains(element)
			}
		})
	}
}

// BenchmarkUnion benchmarks union operations on sets of different sizes.
func BenchmarkUnion(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s1 := NewSet(makeStrings(size)...)
			s2 := NewSet(makeStringsOffset(size, size)...)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				s1.Union(s2)
			}
		})
	}
}

// BenchmarkIntersect benchmarks intersection operations on sets of different sizes.
func BenchmarkIntersect(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s1 := NewSet(makeStrings(size)...)
			s2 := NewSet(makeStringsOffset(size/2, size)...)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				s1.Intersect(s2)
			}
		})
	}
}

// BenchmarkDifference benchmarks difference operations on sets of different sizes.
func BenchmarkDifference(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s1 := NewSet(makeStrings(size)...)
			s2 := NewSet(makeStringsOffset(size/2, size)...)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				s1.Difference(s2)
			}
		})
	}
}

// BenchmarkIsSubset benchmarks subset checks on sets of different sizes.
func BenchmarkIsSubset(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			s1 := NewSet(makeStrings(size / 2)...)
			s2 := NewSet(makeStrings(size)...)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				s1.IsSubset(s2)
			}
		})
	}
}
