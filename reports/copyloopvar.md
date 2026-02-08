# copyloopvar Linter - Comprehensive Analysis

## What the Linter Does

The `copyloopvar` linter detects **unnecessary manual copying of loop variables** in Go 1.22 and later. It identifies and flags code that manually copies loop variables (`v := v`) which is no longer necessary due to Go 1.22's improved loop variable semantics.

### The Go 1.22 Change

**Before Go 1.22**, loop variables (`i`, `v` in `for i, v := range slice`) were **reused across iterations**, causing a notorious bug where closures or goroutines would capture the same variable address, leading to all iterations seeing the final value:

```go
// BROKEN in Go < 1.22
for _, val := range values {
    go func() {
        fmt.Println(val) // BUG: All goroutines print the last value!
    }()
}
```

**Go 1.22 fixed this** by making each loop iteration create new variables. The workaround (`val := val`) is now unnecessary:

```go
// Go 1.22+ - this workaround is now redundant
for _, val := range values {
    val := val  // ⚠️ copyloopvar: The copy of the 'for' variable "val" can be deleted
    go func() {
        fmt.Println(val)  // Works correctly without manual copying
    }()
}
```

### Detection Behavior

The linter identifies:

- **Exact copies**: `v := v`
- **Index copies**: `i := i`
- **Variable renaming**: `_v := v`, `item := v`

With `check-alias: true`, it additionally flags aliased copies where the copy variable name differs from the original.

## When It Should Be Enabled

### ✅ Strongly Recommended For:

- **Go 1.22+ projects** - Eliminates redundant code and improves readability
- **Codebases migrating from older Go versions** - Identifies outdated workarounds
- **Teams with mixed experience levels** - Prevents junior developers from adding unnecessary copies
- **Libraries and frameworks** - Ensures modern, clean code examples
- **Projects using goroutines in loops** - Removes defensive copying patterns
- **CI/CD pipelines** - Catches copy operations during code review

### 📦 Specific Project Types:

- Microservices with concurrent processing
- Web frameworks and HTTP handlers
- Data processing pipelines
- Test suites with parallel execution
- CLI tools with background processing
- Any project using `go` keywords inside loops

## When It Should Be Disabled

### ❌ Disable When:

- **Using Go < 1.22** - The linter is automatically disabled by golangci-lint (feature detection)
- **Maintaining backward compatibility** - Code must compile on older Go versions
- **Legacy codebases without upgrade plans** - Keep existing patterns if not modernizing
- **Generated code** - If code generators target older Go versions
- **Educational materials** - Teaching pre-1.22 behavior requires showing the old pattern

### ⚠️ Caution:

- **Transitional periods**: If your team hasn't fully migrated mental models to Go 1.22+ semantics
- **Mixed-version builds**: Projects built with both pre-1.22 and post-1.22 toolchains

## Configuration Options

### Configuration Structure

```yaml
# golangci-lint v2 format
linters:
  settings:
    copyloopvar:
      check-alias: false # Default: false
```

```yaml
# golangci-lint v1 format
linters-settings:
  copyloopvar:
    check-alias: false
```

### `check-alias` Option

- **Default: `false`**

When `false`: Only flags exact-name copies (`v := v`)

When `true`: Also flags renamed copies (`_v := v`, `item := v`)

### Example Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# .golangci.yml
version: "2"
linters:
  enable:
    - copyloopvar
  settings:
    copyloopvar:
      check-alias: true # Catches more cases
```

#### ✅ Minimal Configuration

```yaml
# .golangci.yml
version: "2"
linters:
  enable:
    - copyloopvar
  # Uses default: check-alias: false
```

#### ✅ With Go Version Build Tags (Explicit)

```yaml
# .golangci.yml
version: "2"
run:
  go: "1.22"

linters:
  enable:
    - copyloopvar
  settings:
    copyloopvar:
      check-alias: true
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

| Linter                  | Relationship  | Benefit                                         |
| ----------------------- | ------------- | ----------------------------------------------- |
| **`govet` (copylocks)** | Complementary | Both catch closure/scoping issues; no overlap   |
| **`staticcheck`**       | Complementary | Different focus areas; can run together         |
| **`gocritic`**          | Compatible    | Different categories of issues                  |
| **`ineffassign`**       | Compatible    | Both flag unused assignments                    |
| **`wastedassign`**      | Compatible    | Similar goals, different detection targets      |
| **`paralleltest`**      | Synergistic   | For tests using `t.Parallel()` with range loops |

### ❌ No Conflicts

- No known linters produce contradictory warnings
- No overlapping message patterns
- Can safely run with all standard golangci-lint presets

### 🔍 Detection Gaps Filled

```go
// This is caught by copyloopvar
for _, v := range items {
    v := v  // ⚠️ Flagged: can be deleted
    process(v)
}

// This is NOT caught (but might be caught by other linters)
for _, v := range items {
    process(&v)  // Potentially still buggy if process stores the pointer
}
```

### Real-World Detection Examples

**Before (Go < 1.22 workaround):**

```go
// cmd/process.go
func processItems(items []Item) []func() {
    var handlers []func()
    for _, item := range items {
        item := item  // Workaround for loop variable capture
        handlers = append(handlers, func() {
            fmt.Printf("Processing: %s\n", item.Name)
        })
    }
    return handlers
}
```

**After (Go 1.22+ with copyloopvar):**

```go
// cmd/process.go
func processItems(items []Item) []func() {
    var handlers []func()
    for _, item := range items {
        // No manual copy needed - Go 1.22 handles it
        handlers = append(handlers, func() {
            fmt.Printf("Processing: %s\n", item.Name)
        })
    }
    return handlers
}
```

**Linter Output:**

```
cmd/process.go:6:3: The copy of the 'for' variable "item" can be deleted (Go 1.22+)
```

### Migration Strategy

1. **Upgrade to Go 1.22+** in `go.mod`
2. **Enable copyloopvar** in golangci-lint
3. **Run linter** to identify all redundant copies
4. **Remove manual copies** one by one
5. **Test thoroughly** - behavior should be identical (but cleaner)
6. **Update team documentation** about new Go 1.22 semantics

### Performance Impact

- **Negligible runtime impact** (removes unnecessary assignments)
- **Slight compile-time improvement** (less code to compile)
- **Readability improvement** (less defensive boilerplate)
- **Binary size reduction** (minor, but measurable in tight loops)

### Version Compatibility Notes

- golangci-lint v1.57.0+ required
- Automatically disabled on Go < 1.22
- Works with both v1 and v2 golangci-lint config formats
- No impact on generated code (respects `// Code generated` comments)

## Summary

**copyloopvar** is a **Go 1.22+ specific linter** that modernizes code by removing defensive loop variable copying patterns that are no longer necessary. It:

- ✅ Identifies redundant copy operations
- ✅ Improves code readability
- ✅ Reduces technical debt
- ✅ Educates developers about Go 1.22 changes
- ✅ Has negligible performance impact
- ✅ Auto-fixable (via `golangci-lint run --fix`)

**Recommendation**: **ENABLE** for all Go 1.22+ projects. The linter will be automatically disabled on older Go versions. Use `check-alias: true` for maximum benefit.

---

**Reference**: https://github.com/karamaru-alpha/copyloopvar
