# ADR-001: Set[T] Generic Type for Collection Operations

**Status:** Accepted  
**Date:** 2026-04-09  
**Author:** Lars Artmann (@larsartmann)

---

## Context

Our codebase frequently performs set-like operations on collections of linters, formatters, file paths, and other data. We need a clean, type-safe way to handle these operations without duplication or boilerplate.

Prior to this decision, we used:

1. **Slices with manual deduplication**: Required repeated loops and `sort.Search` for membership checks
2. **`map[T]struct{}`**: Standard Go pattern but lacks expressiveness and method chaining
3. **Manual iteration**: Error-prone and scattered across codebase

---

## Decision

We will create a generic `Set[T comparable]` type with comprehensive set operations.

### Implementation

```go
// pkg/types/set.go
type Set[T comparable] struct {
    elements map[T]struct{}
}

// Constructor
func NewSet[T comparable](items ...T) Set[T] { ... }

// Core operations
func (s Set[T]) Add(item T) Set[T]          // Returns new set (immutable)
func (s Set[T]) Contains(item T) bool
func (s Set[T]) Remove(item T) Set[T]       // Returns new set (immutable)
func (s Set[T]) Slice() []T                 // Extract to slice
func (s Set[T]) Len() int

// Set operations
func (s Set[T]) Union(other Set[T]) Set[T]
func (s Set[T]) Intersect(other Set[T]) Set[T]
func (s Set[T]) Difference(other Set[T]) Set[T]
func (s Set[T]) IsSubset(other Set[T]) bool
func (s Set[T]) IsSuperset(other Set[T]) bool
```

### Key Design Decisions

1. **Generic Type**: `Set[T comparable]` provides type safety
2. **Immutable by Default**: `Add`/`Remove` return new sets to prevent side effects
3. **Zero Dependencies**: Pure Go, no external libraries needed
4. **Performance**: O(1) membership checks via underlying map

---

## Consequences

### Positive

- **Type Safety**: Compiler catches mixing of different set types
- **Composability**: Chain operations: `set1.Union(set2).Intersect(set3)`
- **Clarity**: `set.Contains(x)` vs `slices.Contains(slice, x)`
- **Testability**: Easy to assert on set equality

### Negative

- **Memory**: Slightly more allocations than raw maps
- **Learning Curve**: Team must learn Set API

---

## Usage Examples

### Before (manual slice operations)

```go
linters := []string{"gosec", "errcheck", "staticcheck"}
if !slices.Contains(linters, "gosec") {
    linters = append(linters, "gosec")
}
```

### After (with Set[T])

```go
linters := types.NewSet("gosec", "errcheck", "staticcheck")
linters = linters.Add("gosec") // No-op, already exists
```

### Set Operations

```go
// Union of enabled linters
allEnabled := criticalLinters.Union(optionalLinters)

// Difference to find missing linters
missing := recommendedLinters.Difference(currentlyEnabled)

// Intersection for common functionality
common := projectALinters.Intersect(projectBLinters)
```

---

## Migration Strategy

1. **Phase 1**: Introduce Set[T] for new code
2. **Phase 2**: Refactor high-traffic areas (merger.go, analyzer.go)
3. **Phase 3**: Audit remaining slice operations for Set candidacy

---

## Alternatives Considered

| Alternative                        | Pros                  | Cons                              | Verdict  |
| ---------------------------------- | --------------------- | --------------------------------- | -------- |
| Keep using slices                  | Simple, familiar      | O(n) contains, verbose            | Rejected |
| `map[T]struct{}`                   | Standard Go, fast     | Verbose, no operations            | Rejected |
| **github.com/deckarep/golang-set** | Mature, battle-tested | External dependency, less control | Rejected |
| **samber/mo.Set**                  | Functional style      | Overkill for our needs            | Rejected |

---

## References

- [Set Implementation](../../pkg/types/set.go)
- [Set Tests](../../pkg/types/set_test.go)
- Related ADRs: ADR-002 (CommandBuilder Pattern)

---

_Accepted by: Lars Artmann_  
_Date: 2026-04-09_
