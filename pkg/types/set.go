package types

import (
	"cmp"
	"slices"
)

// Set is a generic set of comparable values with O(1) lookups.
type Set[T comparable] map[T]struct{}

// NewSet creates a Set from a slice of items.
func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))

	for _, item := range items {
		s[item] = struct{}{}
	}

	return s
}

// Add inserts an item into the set.
func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

// Contains returns true if the item is in the set.
func (s Set[T]) Contains(item T) bool {
	_, ok := s[item]

	return ok
}

// Delete removes an item from the set.
func (s Set[T]) Delete(item T) {
	delete(s, item)
}

// Len returns the number of items in the set.
func (s Set[T]) Len() int {
	return len(s)
}

// ToSlice returns the set items as an unsorted slice.
func (s Set[T]) ToSlice() []T {
	result := make([]T, 0, len(s))

	for item := range s {
		result = append(result, item)
	}

	return result
}

// ToSortedSlice returns the set items as a sorted slice.
// The constraint cmp.Ordered ensures elements are sortable.
func ToSortedSlice[T cmp.Ordered](s Set[T]) []T {
	result := s.ToSlice()
	slices.Sort(result)

	return result
}
