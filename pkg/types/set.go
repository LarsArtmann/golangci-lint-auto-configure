package types

import (
	"cmp"
	"slices"
)

// Set is a generic set of comparable values with O(1) lookups.
type Set[T comparable] map[T]struct{}

// NewSet creates a Set from a slice of items.
func NewSet[T comparable](items ...T) Set[T] {
	result := make(Set[T], len(items))

	for _, item := range items {
		result[item] = struct{}{}
	}

	return result
}

// NewSetWithFunc creates a Set by transforming indices using the provided function.
func NewSetWithFunc[T comparable](n int, fn func(int) T) Set[T] {
	result := make(Set[T], n)

	for i := range n {
		result[fn(i)] = struct{}{}
	}

	return result
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

// IsEmpty returns true if the set has no items.
func (s Set[T]) IsEmpty() bool {
	return len(s) == 0
}

// Clone returns a shallow copy of the set.
func (s Set[T]) Clone() Set[T] {
	result := make(Set[T], len(s))

	for item := range s {
		result[item] = struct{}{}
	}

	return result
}

// Union returns a new set containing all items from both sets.
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := make(Set[T], len(s)+len(other))

	for item := range s {
		result[item] = struct{}{}
	}

	for item := range other {
		result[item] = struct{}{}
	}

	return result
}

// Difference returns a new set containing items in s that are not in other.
func (s Set[T]) Difference(other Set[T]) Set[T] {
	result := make(Set[T])

	for item := range s {
		if !other.Contains(item) {
			result[item] = struct{}{}
		}
	}

	return result
}

// Intersect returns a new set containing items present in both sets.
func (s Set[T]) Intersect(other Set[T]) Set[T] {
	result := make(Set[T])

	for item := range s {
		if other.Contains(item) {
			result[item] = struct{}{}
		}
	}

	return result
}

// Equal returns true if both sets contain exactly the same items.
func (s Set[T]) Equal(other Set[T]) bool {
	return len(s) == len(other) && s.containsAll(other)
}

// IsSubset returns true if all items in s are contained in other.
// Returns true for empty sets.
func (s Set[T]) IsSubset(other Set[T]) bool {
	return len(s) <= len(other) && s.containsAll(other)
}

// containsAll returns true if all items in s are contained in other.
func (s Set[T]) containsAll(other Set[T]) bool {
	for item := range s {
		if !other.Contains(item) {
			return false
		}
	}

	return true
}

// IsSuperset returns true if all items in other are contained in s.
// Returns true if other is empty.
func (s Set[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}

// IsProperSubset returns true if s is a subset of other and not equal.
func (s Set[T]) IsProperSubset(other Set[T]) bool {
	return s.IsSubset(other) && !s.Equal(other)
}

// IsProperSuperset returns true if s is a superset of other and not equal.
func (s Set[T]) IsProperSuperset(other Set[T]) bool {
	return s.IsSuperset(other) && !s.Equal(other)
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
// Returns nil for empty sets. The constraint cmp.Ordered ensures elements are sortable.
func ToSortedSlice[T cmp.Ordered](s Set[T]) []T {
	if len(s) == 0 {
		return nil
	}

	result := s.ToSlice()
	slices.Sort(result)

	return result
}
