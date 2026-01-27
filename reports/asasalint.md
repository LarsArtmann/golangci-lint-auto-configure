# asasalint Linter - Comprehensive Analysis

## What the Linter Does

**`asasalint`** (github.com/alingse/asasalint) detects a common but subtle bug in Go code where a slice of `[]interface{}` or `[]any` is passed as a single argument to a variadic function that expects `...interface{}` or `...any`.

### The Problem It Solves

The linter catches this mistake:
```go
// WRONG: Passing slice as single argument
args := []interface{}{"error", "timeout"}
logger.Error(args)  // slice passed as ONE argument
```

Instead of the correct:
```go
// CORRECT: Expanding slice into multiple arguments
args := []interface{}{"error", "timeout"}
logger.Error(args...)  // ellipsis expands slice
```

### Key Characteristics
- **Type**: Static analysis linter
- **Speed**: Fast (uses Go's type information)
- **Integration**: Part of golangci-lint v2+
- **Repository**: https://github.com/alingse/asasalint

## When It Should Be Enabled

### ✅ RECOMMENDED FOR:

**All Go Projects** that:
- Use variadic functions extensively (logging, formatting, database operations)
- Have multiple developers (catches common mistakes)
- Use reflection-based APIs that accept `...interface{}`
- Maintain code quality standards

**Specific Use Cases:**
- **Logging frameworks**: `log.Printf()`, `zap.Logger`, `slog.Logger`, `logrus`
- **Database operations**: SQL query builders with variadic parameters
- **Message formatting**: `fmt.Printf()` family, internationalization APIs
- **Custom APIs**: Any variadic functions accepting `...interface{}`

**Example Scenario:**
```go
// High-risk code that benefits from asasalint
func processEvent(event string, args ...interface{}) {
    logger.Info("Event: "+event, args)  // BUG: args should be args...
    metrics.Record(event, args)         // BUG: args should be args...
}
```

### Priority Assessment:
- **Default Priority**: Medium-High (Optional but recommended)
- **Critical for**: Projects with heavy logging/metrics usage
- **Low priority for**: Simple CLI tools with minimal variadic function usage

## When It Should Be Disabled

### ❌ CONSIDER DISABLING WHEN:

**Code Patterns:**
- Project rarely uses variadic functions
- Heavily relies on passing slices intentionally as single arguments to variadic functions
- Uses code generation that produces false positives

**Edge Cases:**
- **False positives in legacy code**: When migrating large existing codebases
- **Wrapper functions**: When intentionally wrapping variadic calls
```go
// Intentional pattern - not a bug
func wrapper(args ...interface{}) {
    // Pass slice as single argument to another variadic
    // (this might be intentional but is flagged)
    baseFunc(args)  // May need nolint directive
}
```

### Better Alternative to Disabling:
Instead of disabling entirely, use **exclusions**:
```yaml
# .golangci.yml
linters:
  exclusions:
    rules:
      - linters: [asasalint]
        path: (.+)_test\.go  # Exclude tests if needed
      - linters: [asasalint]
        path: internal/generated/.*  # Exclude generated code
```

## Configuration Options

### Basic Configuration Structure:

```yaml
# .golangci.yml (version 2+)
version: "2"
linters:
  enable:
    - asasalint
  settings:
    asasalint:
      # Exclude specific function patterns from checking
      exclude:
        - Append
        - \.Wrapf

      # Use built-in exclusions (default: true)
      # Built-in exclusions typically include common wrapper functions
      use-builtin-exclusions: true
```

### Configuration Options Explained:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `exclude` | `[]string` | `[]` | Regex patterns for function names/methods to exclude |
| `use-builtin-exclusions` | `bool` | `true` | Whether to use built-in exclusions for common cases |

### Recommended Configurations:

**Standard Project:**
```yaml
linters:
  settings:
    asasalint:
      use-builtin-exclusions: true
```

**Library with Wrapper Functions:**
```yaml
linters:
  settings:
    asasalint:
      exclude:
        - MyCustomWrapper
        - internal/wrap\.Wrap.*
      use-builtin-exclusions: true
```

**Strict Mode (No Exclusions):**
```yaml
linters:
  settings:
    asasalint:
      use-builtin-exclusions: false
```

## Interaction with Other Linters

### ✅ SYNERGIES:

**Complementary Linters:**
- **`govet`**: Catches different classes of bugs, works well together
- **`staticcheck`**: Advanced static analysis, no overlap with asasalint
- **`errcheck`**: Ensures errors are checked, asasalint ensures correct error passing
- **`loggercheck`**: Validates logger key-value pairs, asasalint ensures slice expansion

**Example Workflow:**
```go
// govcheck catches format string issues
// errcheck ensures errors are checked
// asasalint ensures slices are expanded
func handleError(err error, context []interface{}) {
    if err != nil {
        logger.Error("operation failed", context)  // asasalint would flag this
        return err  // errcheck validates this
    }
}
```

### 🔒 CONFLICTS:
- **No known conflicts** - asasalint operates independently on type checking
- Does not overlap with other linters' functionality

### 📊 PERFORMANCE IMPACT:
- **Minimal**: Uses Go's type information efficiently
- **No measurable slowdown** in typical projects
- **Recommended in CI/CD**: Safe to enable with no performance concerns

## Practical Examples

### Real-World Bug Caught by asasalint:

```go
// Before asasalint
func logError(msg string, fields ...zap.Field) {
    err := db.Query("SELECT ...")
    if err != nil {
        logger.Error(msg, fields)  // BUG: fields not expanded
    }
}

// After asasalint detects issue
func logError(msg string, fields ...zap.Field) {
    err := db.Query("SELECT ...")
    if err != nil {
        logger.Error(msg, fields...)  // FIXED: proper expansion
    }
}
```

### False Positive Handling:

```go
// If you intentionally pass slice as single argument
func collectMetrics(name string, values []float64) {
    // Might be intentional to pass slice as one metric value
    recordMetric(name, values)  // asasalint may flag this
}

// Solutions:
// 1. Use nolint directive
recordMetric(name, values) //nolint:asasalint // intentional single argument

// 2. Add to exclusions
// 3. Refactor API to be clearer
```

## Summary

**asasalint** is a **valuable, low-cost linter** that catches a common bug pattern in Go code. It's particularly valuable for projects using logging frameworks, database libraries, and other APIs with variadic `...interface{}` parameters. The linter has minimal configuration needs, no performance impact, and complements other linters well. **Recommended for most Go projects** with appropriate exclusions for generated code or intentional patterns.
