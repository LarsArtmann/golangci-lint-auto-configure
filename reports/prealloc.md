# prealloc Linter - Comprehensive Analysis

## What It Does

The `prealloc` linter identifies slice declarations that could potentially be pre-allocated to improve performance. It detects situations where slices are built incrementally through loops without specifying an initial capacity, causing multiple memory allocations and copies as the slice grows.

### The Problem It Detects

The linter finds the following anti-patterns:

1. **Zero-capacity slice declarations** followed by loops that append elements
2. **Inefficient slice growth** where Go's internal capacity doubling strategy causes unnecessary allocations
3. **Missing capacity hints** when the final size is known or can be reasonably estimated

**Why This Matters:**
- **Performance**: Preallocating eliminates multiple memory allocations and copies
- **Reduced GC Pressure**: Single allocation reduces garbage collection overhead
- **Memory Efficiency**: Avoids temporary over-allocation during slice growth
- **Predictable Performance**: Eliminates performance spikes during slice growth
- **Throughput**: Critical for high-performance, high-throughput applications

### How It Works

The linter analyzes code for patterns where:
- A slice is declared with zero or unspecified capacity
- The slice is immediately followed by a loop (range, for, or simple)
- Each iteration appends elements to the slice
- The loop structure allows preallocation (no complex control flow)

When these conditions are met, it suggests preallocating with the expected final capacity using `make([]T, 0, capacity)`.

The analysis can be configured to:
- Check only simple loops (no returns/breaks/continues/gotos)
- Include or exclude range loops
- Include or exclude for loops

### Examples

```go
// ❌ BAD: Zero capacity slice, will allocate multiple times
var result []int
for _, v := range source {
    result = append(result, v)
}

// ✅ GOOD: Preallocate with known capacity
result := make([]int, 0, len(source))
for _, v := range source {
    result = append(result, v)
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
| Project Type | Priority | Justification |
|--------------|----------|----------------|
| **High-throughput services** | HIGH | Critical for performance under load |
| **Data processing pipelines** | HIGH | Transforms large datasets efficiently |
| **API servers** | MEDIUM | Improves response time and reduces latency |
| **CLI tools** | MEDIUM | Faster execution improves user experience |
| **Libraries** | MEDIUM | Better performance for library users |

**Specific Scenarios:**

**1. Performance-Critical Code**
- Hot paths in your application
- Code that processes large datasets
- Functions called frequently in loops
- API handlers that build response arrays

**Examples:**
```go
// API response building
func (s *Service) GetUsers(ctx context.Context) ([]User, error) {
    users, err := s.repo.GetAll(ctx)
    if err != nil {
        return nil, err
    }
    result := make([]User, 0, len(users))  // Preallocate
    for _, u := range users {
        result = append(result, User{...})
    }
    return result, nil
}
```

**2. Data Transformation Functions**
- Converting one data structure to another
- Filtering and mapping operations
- Building complex nested structures

```go
// Data transformation
func transformEvents(events []RawEvent) []ProcessedEvent {
    processed := make([]ProcessedEvent, 0, len(events))  // Preallocate
    for _, e := range events {
        processed = append(processed, ProcessEvent(e))
    }
    return processed
}
```

### ❌ Disable For:

**Specific Scenarios:**

**1. Unpredictable Data Sizes**
When the final size is highly variable and unknown, preallocation may waste memory.

```yaml
# Disable when dealing with unpredictable data volumes
linters:
  disable:
    - prealloc
```

**2. Code Where Readability Outweighs Performance**
In non-performance-critical code, simpler syntax may be preferred.

**3. Development/Testing Environments**
May add noise during rapid iteration. Enable only before performance optimization phases.

### Priority Assessment

- **Default Priority**: MEDIUM
- **Value**: HIGH - Performance improvements can be substantial (2-5x faster, fewer allocations)
- **Effort**: LOW - Simple mechanical fix
- **Recommendation**: CONSIDER - Enable after profiling indicates performance issues

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    prealloc:
      # Report pre-allocation suggestions only on simple loops
      # Type: bool
      # Default: true
      # Description: Only report on loops without returns/breaks/continues/gotos
      simple: true

      # Report pre-allocation suggestions on range loops
      # Type: bool
      # Default: true
      # Description: Include range loops in analysis
      range-loops: true

      # Report pre-allocation suggestions on for loops
      # Type: bool
      # Default: false
      # Description: Include regular for loops (disabled by default due to complexity)
      for-loops: false
```

### simple Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Only report suggestions on simple loops without complex control flow (returns, breaks, continues, gotos)

**Format:** Boolean value

**Common Values:**
- `true` - Only check simple, straightforward loops (recommended)
- `false` - Check all loops including those with complex control flow

### range-loops Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Include range loops in the analysis

**Format:** Boolean value

**Common Values:**
- `true` - Check range loops (recommended, most common use case)
- `false` - Exclude range loops from analysis

### for-loops Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: Include regular for loops in analysis

**Format:** Boolean value

**Common Values:**
- `false` - Exclude for loops (default, for loops are often more complex)
- `true` - Include for loops (only if you understand the implications)

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Balanced approach for most projects
version: "2"
linters:
  settings:
    prealloc:
      simple: true
      range-loops: true
      for-loops: false
```

#### ✅ Strict Configuration
```yaml
# More aggressive checking for performance-critical code
version: "2"
linters:
  settings:
    prealloc:
      simple: false
      range-loops: true
      for-loops: false
```

#### ✅ Maximum Coverage
```yaml
# Check all possible loops (may produce false positives)
version: "2"
linters:
  settings:
    prealloc:
      simple: false
      range-loops: true
      for-loops: true

issues:
  exclude-rules:
    - path: ".*_test.go"
      linters:
        - prealloc
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

prealloc works excellently with:

| Linter | Relationship | Value |
|--------|--------------|---------|
| **govet** | Static analysis | Catches other performance issues |
| **ineffassign** | Performance | Detects ineffective assignments |
| **staticcheck** | Performance | Advanced performance analysis |
| **predeclared** | Code quality | Improves code readability |

**Complete Performance Suite:**
```yaml
linters:
  enable:
    - prealloc       # Slice pre-allocation (MEDIUM)
    - govet          # Standard static analysis (CRITICAL)
    - ineffassign    # Ineffective assignments (HIGH)
    - staticcheck    # Advanced analysis (CRITICAL)
```

### Example of Linter Synergy

```go
// prealloc catches:
var result []int
for _, v := range source {
    result = append(result, v)
}

// staticcheck adds:
// "consider preallocating result with capacity len(source)"

// govet adds:
// Additional static checks for other issues
```

### 🔒 No Conflicts / Minimal Overlap

| Linter | Overlap | Recommendation |
|---------|----------|----------------|
| prealloc + **ineffassign** | Minimal | Both catch different performance issues |

**Why No Conflicts:**
- prealloc focuses on slice allocation patterns
- ineffassign focuses on unused assignments
- Each linter covers different aspects of performance optimization

## Practical Examples

### ✅ Example 1: Basic Range Loop

```go
// ❌ BAD: No preallocation, multiple allocations
var users []User
for _, u := range rawUsers {
    users = append(users, transformUser(u))
}

// ✅ GOOD: Preallocate with known capacity
users := make([]User, 0, len(rawUsers))
for _, u := range rawUsers {
    users = append(users, transformUser(u))
}
```

### ✅ Example 2: Filter Operation

```go
// ❌ BAD: No capacity hint
var activeUsers []User
for _, u := range allUsers {
    if u.IsActive {
        activeUsers = append(activeUsers, u)
    }
}

// ✅ GOOD: Preallocate with estimate (may overallocate but fewer allocations)
activeUsers := make([]User, 0, len(allUsers))
for _, u := range allUsers {
    if u.IsActive {
        activeUsers = append(activeUsers, u)
    }
}
```

### ✅ Example 3: Map to Slice Conversion

```go
// ❌ BAD: No preallocation
var values []int
for _, v := range dataMap {
    values = append(values, v)
}

// ✅ GOOD: Preallocate with map size
values := make([]int, 0, len(dataMap))
for _, v := range dataMap {
    values = append(values, v)
}
```

### ✅ Example 4: Nested Data Structure

```go
// ❌ BAD: No preallocation for nested slices
type Response struct {
    Users []User `json:"users"`
}

func buildResponse(users []*db.User) Response {
    var result []User
    for _, u := range users {
        result = append(result, User{
            ID:   u.ID,
            Name: u.Name,
        })
    }
    return Response{Users: result}
}

// ✅ GOOD: Preallocate at known size
func buildResponse(users []*db.User) Response {
    result := make([]User, 0, len(users))
    for _, u := range users {
        result = append(result, User{
            ID:   u.ID,
            Name: u.Name,
        })
    }
    return Response{Users: result}
}
```

### ✅ Example 5: Processing with Early Exit

```go
// ❌ BAD: No preallocation even with early exit
var matches []string
for _, s := range strings {
    if s == target {
        matches = append(matches, s)
        break
    }
}

// ✅ GOOD: Preallocate is still beneficial even with early exit
matches := make([]string, 0, len(strings))
for _, s := range strings {
    if s == target {
        matches = append(matches, s)
        break
    }
}
```

## Best Practices

1. **Profile Before Optimizing** - Always profile your application before enabling this linter to identify actual bottlenecks
2. **Use make() with Capacity** - Always use `make([]T, 0, capacity)` for preallocation (0 length, non-zero capacity)
3. **Estimate Conservatively** - When final size is unknown, estimate conservatively to avoid memory waste
4. **Prioritize Hot Paths** - Focus optimization on code paths that execute frequently
5. **Consider Tradeoffs** - Balance performance gains against code readability
6. **Document Why** - Add comments when preallocation is used for performance reasons
7. **Test After Changes** - Verify functionality and performance improvements after optimization
8. **Disable in Tests** - Consider excluding test files if preallocation adds noise
9. **Monitor Memory Usage** - Ensure preallocation doesn't cause excessive memory consumption
10. **Use with Other Performance Linters** - Combine with staticcheck and govet for comprehensive performance analysis

## Common Scenarios and Solutions

### Scenario 1: API Response Building

**Problem:** API endpoints building large JSON responses suffer from slow response times.

**Solution:** Preallocate response slices
```yaml
linters:
  enable:
    - prealloc
  settings:
    prealloc:
      simple: true
      range-loops: true
      for-loops: false
```

### Scenario 2: Data Pipeline Processing

**Problem:** ETL pipelines processing millions of records are slow.

**Solution:** Aggressive preallocation for all data transformations
```yaml
linters:
  enable:
    - prealloc
  settings:
    prealloc:
      simple: false
      range-loops: true
      for-loops: true
```

### Scenario 3: High-Quality Code Without Performance Requirements

**Problem:** Preallocation adds complexity to simple utility code.

**Solution:** Disable prealloc for non-critical code
```yaml
linters:
  disable:
    - prealloc
```

### Scenario 4: Library Code

**Problem:** Want to optimize library performance but don't want noise.

**Solution:** Enable with strict rules, exclude tests
```yaml
linters:
  enable:
    - prealloc
  settings:
    prealloc:
      simple: true
      range-loops: true
      for-loops: false

issues:
  exclude-rules:
    - path: ".*_test.go"
      linters:
        - prealloc
```

### Scenario 5: Migrated Code

**Problem:** Legacy code has many performance issues, overwhelming suggestions.

**Solution:** Start with standard config, gradually enable more options
```yaml
# Phase 1: Basic optimization
linters:
  enable:
    - prealloc
  settings:
    prealloc:
      simple: true
      range-loops: true
      for-loops: false

# Phase 2: After initial fixes, enable more aggressive checks
linters:
  enable:
    - prealloc
  settings:
    prealloc:
      simple: false
      range-loops: true
      for-loops: false
```

## Summary

**prealloc** is a **MEDIUM** priority linter that improves Go application performance by identifying opportunities to preallocate slice capacity:

- ✅ **2-5x Performance Improvement** - Eliminates multiple allocations and copies
- ✅ **Reduced Memory Usage** - Single allocation instead of multiple growing allocations
- ✅ **Lower GC Pressure** - Fewer allocations mean less garbage collection overhead
- ✅ **Predictable Performance** - Avoids performance spikes during slice growth
- ✅ **Easy to Fix** - Simple mechanical change to add capacity hints
- ✅ **Well-Understood Pattern** - Standard Go optimization technique
- ✅ **Configurable** - Can adjust strictness based on project needs
- ✅ **Valuable for Hot Paths** - Critical for performance-critical code
- ⚠️ **Premature Optimization** - Should only be used after profiling identifies performance issues
- ⚠️ **Potential for Over-Allocation** - May waste memory if final size is unknown

**Recommendation:** **CONSIDER** enabling prealloc for projects where performance is important, but only after profiling has identified bottlenecks. Use standard configuration initially, then adjust based on project needs. Balance performance gains against code complexity.

---

**Reference:** https://github.com/alexkohler/prealloc
