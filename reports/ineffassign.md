# ineffassign Linter - Comprehensive Analysis

## What It Does

The `ineffassign` linter identifies assignments to existing variables that are never used after the assignment. It detects situations where a variable is assigned a value but that value is immediately overwritten or never read, making the assignment pointless and potentially masking bugs.

### The Problem It Detects

The linter finds the following anti-patterns:

1. **Immediate reassignment** - Variables assigned values that are immediately overwritten
2. **Lost error values** - Error variables reassigned without being checked or returned
3. **Unused return values** - Function return values assigned but never used
4. **Conditional overwrites** - Initial assignments in if/else branches that are immediately overwritten

**Why This Matters:**

- **Bug Prevention** - Lost errors and return values can mask critical failures
- **Code Clarity** - Removes confusing, dead assignments
- **Performance** - Eliminates unnecessary CPU cycles and memory operations
- **Maintainability** - Prevents code "rot" where assignments become unused over time
- **Test Reliability** - Catches test code that silently ignores errors

### How It Works

The linter performs a simple static analysis:

- Tracks variable assignments throughout a function
- Identifies when a variable is assigned but never read before the next assignment
- Reports these ineffectual assignments

The analysis has known limitations:

- **No type analysis** - Doesn't consider struct field assignments, method receivers, or channel assignments
- **Context-blind** - May miss some cases in complex control flow
- **Scope limitations** - Only analyzes within function boundaries

The linter can be configured to check escaping error variables with the `check-escaping-errors` option.

### Examples

```go
// ❌ BAD: Error is immediately overwritten
func process() error {
    data, err := getData()
    err = saveData(data)  // Lost original error from getData()
    return err
}

// ✅ GOOD: Handle original error first
func process() error {
    data, err := getData()
    if err != nil {
        return err
    }
    return saveData(data)
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
| Project Type | Priority | Justification |
|--------------|----------|----------------|
| **Production codebases** | HIGH | Prevents bugs from lost errors |
| **API services** | HIGH | Critical to handle errors correctly |
| **Database-heavy applications** | HIGH | Database errors must not be lost |
| **Test suites** | HIGH | Ensures tests properly verify failures |
| **All Go projects** | MEDIUM | Fundamental error handling hygiene |

**Specific Scenarios:**

**1. Error Handling Code**

- Functions that call multiple operations that can fail
- Code that processes external resources (files, network, databases)
- Functions that return error values

**Examples:**

```go
// Bad: Lost database connection error
func getUser(id int) (*User, error) {
    db, err := sql.Open("postgres", connStr)
    err = db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&user)  // Lost DB open error
    return &user, err
}

// Good: Handle connection error
func getUser(id int) (*User, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    err = db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&user)
    return &user, err
}
```

**2. Data Processing Pipelines**

- Multiple transformation steps where errors can occur
- Batch processing operations
- Data validation sequences

```go
// Bad: Lost validation errors
func processBatch(items []Item) error {
    for _, item := range items {
        err := validate(item)
        err = transform(item)  // Lost validation error
        err = save(item)       // Lost transform error
    }
    return nil
}

// Good: Handle each error
func processBatch(items []Item) error {
    for _, item := range items {
        if err := validate(item); err != nil {
            return fmt.Errorf("validation failed: %w", err)
        }
        if err := transform(item); err != nil {
            return fmt.Errorf("transform failed: %w", err)
        }
        if err := save(item); err != nil {
            return fmt.Errorf("save failed: %w", err)
        }
    }
    return nil
}
```

### ❌ Disable For:

**Specific Scenarios:**

**1. Code with Known False Positives**
If the linter produces many false positives in specific contexts.

**2. Intentional Overwrites**
Rare cases where overwrites are intentional for specific reasons (e.g., final fallback after multiple attempts).

### Priority Assessment

- **Default Priority**: HIGH
- **Value**: HIGH - Prevents serious bugs from lost errors
- **Effort**: LOW - Simple fix to handle or remove assignments
- **Recommendation**: ALWAYS - Should be enabled for all Go projects

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    ineffassign:
      # Check escaping variables of type error
      # Type: bool
      # Default: false
      # Description: Detects assignments to error variables that might escape the function, may cause false positives
      check-escaping-errors: false
```

### check-escaping-errors Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: When true, checks error variables that escape the function scope

**Format:** Boolean value

**Common Values:**

- `false` - Standard mode, checks only local assignments (default, recommended)
- `true` - Check escaping errors (may cause false positives in some patterns)

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# Balanced approach for most projects
version: "2"
linters:
  enable:
    - ineffassign
  settings:
    ineffassign:
      check-escaping-errors: false
```

#### ✅ Strict Configuration

```yaml
# More aggressive checking for error handling
version: "2"
linters:
  enable:
    - ineffassign
  settings:
    ineffassign:
      check-escaping-errors: true
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

ineffassign works excellently with:

| Linter          | Relationship      | Value                      |
| --------------- | ----------------- | -------------------------- |
| **errcheck**    | Error handling    | Catches unchecked errors   |
| **errchkjson**  | Type safety       | JSON-related type issues   |
| **govet**       | Static analysis   | Catches other code issues  |
| **staticcheck** | Advanced analysis | Comprehensive code quality |

**Complete Error Handling Suite:**

```yaml
linters:
  enable:
    - errcheck # Unchecked errors (CRITICAL)
    - ineffassign # Ineffectual assignments (HIGH)
    - errchkjson # JSON type safety (CRITICAL)
    - govet # Standard static analysis (CRITICAL)
    - staticcheck # Advanced analysis (CRITICAL)
```

### Example of Linter Synergy

```go
// ineffassign catches:
data, err := getData()
err = processData(data)  // Lost error from getData()

// errcheck catches:
getData()  // Unchecked error

// staticcheck adds:
// Additional analysis and suggestions
```

### 🔒 No Conflicts / Minimal Overlap

| Linter                     | Overlap | Recommendation                                  |
| -------------------------- | ------- | ----------------------------------------------- |
| ineffassign + **errcheck** | Minimal | Both handle different aspects of error handling |
| ineffassign + **govet**    | Minimal | Complementary static analysis                   |

**Why No Conflicts:**

- ineffassign focuses on overwritten/unused assignments
- errcheck focuses on unchecked return errors
- govet provides broader static analysis
- Each linter provides unique coverage

## Practical Examples

### ✅ Example 1: Lost Database Error

```go
// ❌ BAD: Connection error is lost
func getUser(id int) (*User, error) {
    db, err := sql.Open("postgres", connStr)
    rows, err := db.Query("SELECT * FROM users WHERE id = $1", id)
    // Connection error is lost
    return scanUser(rows)
}

// ✅ GOOD: Handle connection error
func getUser(id int) (*User, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    rows, err := db.Query("SELECT * FROM users WHERE id = $1", id)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    return scanUser(rows)
}
```

### ✅ Example 2: Conditional Configuration

```go
// ❌ BAD: Default config is immediately overwritten
func getConfig() *Config {
    cfg := getDefaultConfig()  // Ineffectual
    if os.Getenv("ENV") == "production" {
        cfg = getProductionConfig()
    }
    return cfg
}

// ✅ GOOD: Use var declaration
func getConfig() *Config {
    var cfg *Config
    if os.Getenv("ENV") == "production" {
        cfg = getProductionConfig()
    } else {
        cfg = getDefaultConfig()
    }
    return cfg
}
```

### ✅ Example 3: Multiple Operations

```go
// ❌ BAD: Errors from earlier operations are lost
func processFile(path string) error {
    f, err := os.Open(path)
    data, err := io.ReadAll(f)
    _, err = os.Create("output.txt")
    return err  // Only last error is returned
}

// ✅ GOOD: Handle each error
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open failed: %w", err)
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return fmt.Errorf("read failed: %w", err)
    }

    out, err := os.Create("output.txt")
    if err != nil {
        return fmt.Errorf("create failed: %w", err)
    }
    defer out.Close()

    _, err = out.Write(data)
    return err
}
```

### ✅ Example 4: Test Code

```go
// ❌ BAD: Test doesn't check setup errors
func TestProcess(t *testing.T) {
    input, err := createTestInput()
    err = process(input)  // Lost setup error
    assert.NoError(t, err)
}

// ✅ GOOD: Check setup errors
func TestProcess(t *testing.T) {
    input, err := createTestInput()
    require.NoError(t, err)  // Fail test if setup fails

    err = process(input)
    assert.NoError(t, err)
}
```

### ✅ Example 5: Loop Error Handling

```go
// ❌ BAD: Only last error is returned
func processItems(items []Item) error {
    var err error
    for _, item := range items {
        err = processItem(item)  // Each error overwrites the previous
    }
    return err  // Returns only the last error
}

// ✅ GOOD: Return immediately on error
func processItems(items []Item) error {
    for _, item := range items {
        if err := processItem(item); err != nil {
            return err  // Return first error
        }
    }
    return nil
}
```

## Best Practices

1. **Handle Errors Immediately** - Check and return errors as soon as they occur
2. **Use Short Declarations** - Use `:=` for first assignment to avoid reassignment patterns
3. **Explicit Ignore with `_`** - Use `_` when intentionally ignoring a return value
4. **Chain Errors with `%w`** - Use `fmt.Errorf` with `%w` for error wrapping
5. **Use `require.NoError` in Tests** - Fail tests immediately on setup errors
6. **Prefer Early Returns** - Return early on errors rather than accumulating them
7. **Separate Variables** - Use different variable names when dealing with multiple errors
8. **Check First Assignment** - Verify the first operation succeeds before proceeding
9. **Review All Findings** - Every ineffassign finding is worth investigating
10. **Combine with Errcheck** - Use both for comprehensive error handling coverage

## Common Scenarios and Solutions

### Scenario 1: Database Operations

**Problem:** Multiple database operations where errors can be lost.

**Solution:** Check each error before proceeding

```yaml
linters:
  enable:
    - ineffassign
```

### Scenario 2: File I/O Operations

**Problem:** File operations often have multiple error points.

**Solution:** Explicitly check each operation

```go
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()

data, err := io.ReadAll(f)
if err != nil {
    return err
}
```

### Scenario 3: API Client Operations

**Problem:** HTTP requests can fail at multiple stages.

**Solution:** Handle each stage separately

```go
req, err := http.NewRequest("GET", url, nil)
if err != nil {
    return err
}

resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
```

### Scenario 4: Test Setup

**Problem:** Test setup code losing errors.

**Solution:** Use require for setup errors

```go
input, err := createTestInput()
require.NoError(t, err)  // Fail test if setup fails

err = process(input)
assert.NoError(t, err)
```

### Scenario 5: Data Transformation

**Problem:** Pipeline with multiple transformation steps.

**Solution:** Check each step with context

```go
if err := validate(data); err != nil {
    return fmt.Errorf("validation: %w", err)
}
if err := transform(data); err != nil {
    return fmt.Errorf("transform: %w", err)
}
if err := save(data); err != nil {
    return fmt.Errorf("save: %w", err)
}
```

## Summary

**ineffassign** is a **HIGH** priority linter that improves Go code quality by identifying and preventing ineffectual assignments that can mask critical bugs:

- ✅ **Prevents Bug Loss** - Catches lost errors and return values before they cause production issues
- ✅ **Improves Error Handling** - Enforces proper error handling patterns
- ✅ **Code Clarity** - Removes confusing dead assignments
- ✅ **Performance** - Eliminates unnecessary operations
- ✅ **Easy to Fix** - Simple pattern changes to handle errors properly
- ✅ **Critical for Production** - Essential for reliable, bug-free code
- ✅ **Works Well with Others** - Complements errcheck and other error linters
- ✅ **No False Positives** - By design, reports only true issues
- ⚠️ **Type Analysis Limitations** - Doesn't analyze struct fields, method receivers, or channels
- ⚠️ **Context Limitations** - May miss some complex control flow cases

**Recommendation:** **ALWAYS** enable ineffassign for all Go projects. It's a fundamental linter that prevents serious bugs from lost errors and ineffective assignments. Combine with errcheck for comprehensive error handling coverage.

---

**Reference:** https://github.com/gordonklaus/ineffassign
