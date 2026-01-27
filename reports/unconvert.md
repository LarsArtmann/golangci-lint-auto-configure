# unconvert Linter - Comprehensive Analysis

## What It Does

The `unconvert` linter identifies and removes unnecessary type conversions in Go source code. It detects expressions like `T(x)` where the value `x` already has type `T`, making the conversion redundant and wasteful.

### The Problem It Detects

The linter finds the following anti-patterns:

1. **Identity conversions** - Converting a value to its own type (e.g., `int64(x)` when `x` is already `int64`)
2. **Redundant nested conversions** - Chains of conversions that could be simplified
3. **Unnecessary interface conversions** - Type assertions that convert to the same type
4. **Pointless numeric conversions** - Conversions between the same numeric types

**Why This Matters:**
- **Performance**: Unnecessary conversions waste CPU cycles
- **Memory Allocation**: Some conversions create temporary allocations
- **Code Clarity**: Redundant conversions obscure the actual intent
- **Code Maintenance**: Extra type conversions add unnecessary complexity
- **Readability**: Cleaner code without type noise is easier to understand

### How It Works

The linter analyzes code to identify type conversions where:
- The source type is already compatible with the target type
- The conversion is semantically unnecessary
- No type safety or semantic benefit is provided

When these conditions are met, it suggests removing the redundant conversion.

The analysis can be configured with two optional flags:
- `fast-math`: Removes conversions that force intermediate rounding (optimizes floating-point operations)
- `safe`: Uses a more conservative approach to reduce false positives (experimental)

### Examples

```go
// ❌ BAD: Unnecessary conversion (value is already int64)
i := int64(42)
result := int64(i) * 2

// ✅ GOOD: Direct use without conversion
i := int64(42)
result := i * 2
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
| Project Type | Priority | Justification |
|--------------|----------|----------------|
| **Performance-critical applications** | HIGH | Eliminates unnecessary CPU cycles |
| **High-throughput services** | HIGH | Reduces overhead in hot paths |
| **Production codebases** | HIGH | Maintains code cleanliness |
| **API servers** | MEDIUM | Improves efficiency |
| **Libraries** | MEDIUM | Better performance for library users |

**Specific Scenarios:**

**1. Performance-Critical Code**
- Hot paths in your application
- Code that processes large datasets
- Functions called frequently in loops
- Tight inner loops where every cycle counts

**Examples:**
```go
// Hot path optimization
func processData(data []int64) int64 {
    var sum int64
    for _, v := range data {
        // Unnecessary: v is already int64
        sum += int64(v)  // ❌ BAD
    }
    return sum
}

// Fixed
func processData(data []int64) int64 {
    var sum int64
    for _, v := range data {
        sum += v  // ✅ GOOD
    }
    return sum
}
```

**2. Code Cleanliness Initiatives**
- Codebases with accumulated technical debt
- Legacy code with many redundant conversions
- Teams focusing on code quality and maintainability

```go
// Legacy code with noise
func legacyFunction(i int32) {
    value := int32(i)  // ❌ Unnecessary
    process(value)
}

// Cleaned up code
func legacyFunction(i int32) {
    value := i  // ✅ Clean
    process(value)
}
```

### ❌ Disable For:

**Specific Scenarios:**

**1. External API Requirements**
Sometimes conversions are necessary for external API compatibility.

```yaml
# Disable for specific external API calls
linters:
  disable:
    - unconvert
```

**2. Explicit Type Documentation**
When conversions serve as self-documentation for developers.

### Priority Assessment

- **Default Priority**: MEDIUM
- **Value**: MEDIUM - Performance improvement, code cleanup
- **Effort**: LOW - Simple removal of redundant conversions
- **Recommendation**: RECOMMEND - Should be enabled for most projects

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    unconvert:
      # Remove conversions that force intermediate rounding
      # Type: bool
      # Default: false
      # Description: Optimizes floating-point operations by avoiding unnecessary precision loss
      fast-math: false

      # Be more conservative (experimental)
      # Type: bool
      # Default: false
      # Description: Reduces false positives but may miss some unnecessary conversions
      safe: false
```

### fast-math Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: When true, removes conversions that force intermediate rounding in floating-point operations

**Format:** Boolean value

**Common Values:**
- `false` - Keep conversions that may affect floating-point precision (default)
- `true` - Remove conversions that force rounding (may change behavior in edge cases)

### safe Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: When true, uses a more conservative approach to reduce false positives

**Format:** Boolean value

**Common Values:**
- `false` - Standard analysis mode (default, more aggressive)
- `true` - Conservative mode, fewer false positives but may miss issues

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Balanced approach for most projects
version: "2"
linters:
  enable:
    - unconvert
  settings:
    unconvert:
      fast-math: false
      safe: false
```

#### ✅ Conservative Configuration
```yaml
# Fewer false positives
version: "2"
linters:
  enable:
    - unconvert
  settings:
    unconvert:
      fast-math: false
      safe: true
```

#### ✅ Aggressive Configuration
```yaml
# Maximum optimization (may change floating-point behavior)
version: "2"
linters:
  enable:
    - unconvert
  settings:
    unconvert:
      fast-math: true
      safe: false
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

unconvert works excellently with:

| Linter | Relationship | Value |
|--------|--------------|---------|
| **govet** | Static analysis | Catches other type-related issues |
| **staticcheck** | Type safety | Advanced type analysis |
| **ineffassign** | Performance | Detects other redundant operations |
| **gocritic** | Code quality | Finds other redundant patterns |

**Complete Code Quality Suite:**
```yaml
linters:
  enable:
    - unconvert      # Remove unnecessary conversions (MEDIUM)
    - govet          # Standard static analysis (CRITICAL)
    - staticcheck    # Advanced analysis (CRITICAL)
    - ineffassign    # Ineffective assignments (HIGH)
    - gocritic       # General code quality (HIGH)
```

### Example of Linter Synergy

```go
// unconvert catches:
i := int64(42)
result := int64(i) * 2

// ineffassign adds:
// Detects variables that are assigned but never used

// staticcheck adds:
// Advanced type checking and other issues
```

### 🔒 No Conflicts / Minimal Overlap

| Linter | Overlap | Recommendation |
|---------|----------|----------------|
| unconvert + **ineffassign** | Minimal | Both catch different redundancy issues |
| unconvert + **staticcheck** | Minor | Some overlap in type analysis |

**Why No Conflicts:**
- unconvert focuses on unnecessary type conversions
- ineffassign focuses on unused assignments
- staticcheck covers broader type and code quality issues
- Each linter provides unique value

## Practical Examples

### ✅ Example 1: Numeric Identity Conversion

```go
// ❌ BAD: Unnecessary conversion
func calculate(i int64) int64 {
    value := int64(i) * 2
    return value
}

// ✅ GOOD: Direct use
func calculate(i int64) int64 {
    value := i * 2
    return value
}
```

### ✅ Example 2: Function Parameter Conversion

```go
// ❌ BAD: Redundant conversion in function call
func processValue(v int64) {
    storeResult(int64(v))
}

// ✅ GOOD: Pass value directly
func processValue(v int64) {
    storeResult(v)
}
```

### ✅ Example 3: Nested Conversions

```go
// ❌ BAD: Unnecessary nested conversions
func transform(value int16) int64 {
    return int64(int32(int8(value)))
}

// ✅ GOOD: Single direct conversion
func transform(value int16) int64 {
    return int64(value)
}
```

### ✅ Example 4: Loop Counter Pattern

```go
// ❌ BAD: Unnecessary conversion in loop
func processItems(items []int32) {
    for i := range int32(len(items)) {
        handle(items[i])
    }
}

// ✅ GOOD: Direct range over slice
func processItems(items []int32) {
    for i := range items {
        handle(items[i])
    }
}
```

### ✅ Example 5: Interface Conversion

```go
// ❌ BAD: Unnecessary type assertion
func processData(data interface{}) int64 {
    if value, ok := data.(int64); ok {
        return int64(value)
    }
    return 0
}

// ✅ GOOD: Use asserted value directly
func processData(data interface{}) int64 {
    if value, ok := data.(int64); ok {
        return value
    }
    return 0
}
```

## Best Practices

1. **Trust the Linter** - Most conversions flagged by unconvert are truly unnecessary
2. **Review Before Removing** - Verify that the conversion doesn't serve a documentation purpose
3. **Consider Intent** - Some conversions explicitly show type transitions, even if technically unnecessary
4. **Profile Performance** - Measure actual performance impact before optimization
5. **Use with Type-Safe Linters** - Combine with staticcheck for comprehensive type analysis
6. **Document Exceptions** - If you keep a conversion, add a comment explaining why
7. **Enable in CI/CD** - Catch new unnecessary conversions as code is developed
8. **Review with Team** - Discuss any exceptions or false positives with your team
9. **Update Documentation** - Remove unnecessary conversions from examples and documentation
10. **Combine with Refactoring** - Use unconvert as part of code cleanup initiatives

## Common Scenarios and Solutions

### Scenario 1: Legacy Code Cleanup

**Problem:** Large legacy codebase with many accumulated redundant conversions.

**Solution:** Enable unconvert and address findings incrementally
```yaml
linters:
  enable:
    - unconvert
  settings:
    unconvert:
      fast-math: false
      safe: false
```

### Scenario 2: Performance Optimization

**Problem:** Hot paths with unnecessary type conversions causing performance issues.

**Solution:** Use aggressive configuration to maximize optimization
```yaml
linters:
  enable:
    - unconvert
  settings:
    unconvert:
      fast-math: true
      safe: false
```

### Scenario 3: Code Quality Initiative

**Problem:** Team wants to improve code cleanliness and maintainability.

**Solution:** Combine unconvert with other code quality linters
```yaml
linters:
  enable:
    - unconvert
    - ineffassign
    - gocritic
    - stylecheck
```

### Scenario 4: External API Requirements

**Problem:** Some conversions are necessary for external API compatibility.

**Solution:** Exclude specific files or functions with inline directives
```go
//nolint:unconvert
func externalAPICall() {
    // Keep conversion for API compatibility
    value := int64(internalValue)
    api.Send(value)
}
```

### Scenario 5: False Positive Prevention

**Problem:** unconvert flags conversions that serve as explicit type documentation.

**Solution:** Use safe mode or specific exclusions
```yaml
linters:
  enable:
    - unconvert
  settings:
    unconvert:
      fast-math: false
      safe: true
```

## Summary

**unconvert** is a **MEDIUM** priority linter that improves Go code quality by identifying and removing unnecessary type conversions:

- ✅ **Improved Performance** - Eliminates redundant CPU cycles and allocations
- ✅ **Cleaner Code** - Removes type noise and improves readability
- ✅ **Better Maintainability** - Reduces unnecessary complexity
- ✅ **Type Safety** - Maintains type safety while removing redundancy
- ✅ **Easy to Fix** - Simple removal of redundant conversions
- ✅ **Configurable** - Adjust strictness based on project needs
- ✅ **No Runtime Impact** - Pure static analysis
- ✅ **Works Well with Others** - Complements other type and performance linters
- ⚠️ **False Positives** - May flag conversions that serve documentation purposes
- ⚠️ **Intent Obscuration** - Some conversions explicitly show type transitions

**Recommendation:** **RECOMMEND** enabling unconvert for most Go projects as part of a comprehensive code quality and performance strategy. Use standard configuration initially, adjust safe mode if needed. Remove flagged conversions unless they serve an explicit documentation purpose.

---

**Reference:** https://github.com/mdempsky/unconvert
