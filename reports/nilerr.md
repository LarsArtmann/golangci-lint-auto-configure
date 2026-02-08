# nilerr Linter - Comprehensive Analysis

## What the Linter Does

**nilerr** finds code that returns **nil even if it checks that the error is not nil**. This is a subtle but critical bug where error handling exists, but the function returns nil for a value even when there's an error. This is a **CRITICAL** linter for Go error handling.

### The Problem It Detects

A common Go error handling pattern has a subtle bug:

```go
func getData() (*Data, error) {
    if errorCondition {
        return nil, err  // ❌ nilerr: nil returned with non-nil error
    }
    return &Data{}, nil
}
```

**Why This is Critical:**

- **Contract violation** - Go convention: error != nil should imply value is valid
- **Nil pointer panics** - Caller dereferences nil pointer expecting valid data
- **Hard to debug** - Error is checked, so panic happens later in code
- **Silent failures** - Data appears to load but is actually nil

### How It Works

nilerr analyzes functions that:

1. **Return error** - Function signature has `(T, error)` return
2. **Check error in return** - Returns `nil, err` pattern or similar
3. **Verify error not nil** - Error is checked before return (`if err != nil`)
4. **Detect nil value** - Value returned is nil when error is non-nil

### Examples

```go
// ❌ BAD: Returns nil value with non-nil error
func getUser(id int) (*User, error) {
    user, err := db.QueryUser(id)
    if err != nil {
        return nil, err  // nilerr: nil returned with non-nil error
    }
    return user, nil
}

// ✅ GOOD: Return nil error with nil value
func getUser(id int) (*User, error) {
    user, err := db.QueryUser(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

```go
// ❌ BAD: Multiple error checks, final return has nil value
func processData(data string) (*Result, error) {
    if data == "" {
        return nil, errors.New("empty data")  // nilerr: correct
    }

    result := parse(data)
    if result == nil {
        return nil, errors.New("parse failed")  // nilerr: correct
    }

    return nil, nil  // nilerr: nil returned with nil error
}
```

```go
// ❌ BAD: Early return with nil value, error checked
func loadData(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        log.Error(err)
        return nil, err  // nilerr: nil returned with non-nil error
    }
    return data, nil
}

// ✅ GOOD: All returns follow Go error convention
func loadData(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        log.Error(err)
        return nil, err
    }
    return data, nil
}
```

```go
// ❌ BAD: Error checked, but returns nil in success path
func fetchUser(id int) (*User, error) {
    user := &User{ID: id}
    if err := db.Load(user); err != nil {
        return user, err  // nilerr: returns non-nil user with error
    }
    return nil, nil  // nilerr: nil user returned
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

- **All Go applications** - Error handling is fundamental to Go
- **Production systems** - Prevents runtime panics
- **APIs and services** - Ensures error contracts are followed
- **Libraries and SDKs** - Public APIs must follow conventions
- **Microservices** - Distributed systems rely on proper error handling
- **Web applications** - HTTP handlers return values with errors

**Specific Scenarios:**

**1. All Functions Returning (T, error)**

- Database operations
- File I/O operations
- Network calls
- Business logic functions
- API handlers

**2. Codebases with Multiple Developers**

- Prevents this subtle bug pattern
- Enforces Go error conventions
- Code review quality

**3. Error-Intensive Code**

- Many error-prone operations
- Complex error handling chains
- Multiple error sources

**4. APIs with Error Contracts**

- Public libraries
- SDKs for external services
- Framework code
- Plugin systems

**5. Production-Critical Systems**

- Financial applications
- Healthcare software
- Security systems
- Any system where panics are unacceptable

### ❌ Disable For:

**Specific Scenarios:**

**1. Code Not Returning (T, error)**

```yaml
# Functions with different return signatures
linters:
  enable:
    - nilerr
linters-settings:
  nilerr:
    # Only check functions returning (T, error)
    # Default: true
    check-return-value: true

    # Don't check functions that only return error
    # Default: true
    check-assigns: false
```

**2. Test Files (May Intentionally Violate)**

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [nilerr]
```

**3. Generated Code**

```yaml
run:
  skip-dirs:
    - generated

issues:
  exclude-rules:
    - path: (.+)_pb\.go
      linters: [nilerr]
```

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** - Prevents nil pointer panics, enforces Go conventions
- **Effort**: Low - Simple check, minimal false positives
- **Recommendation**: **ALWAYS ENABLE** for all Go code returning errors

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    nilerr:
      # Check function return values
      # Default: true
      check-return-value: true

      # Check assignments (only works with assign)
      # Default: true
      check-assigns: false

      # Exclude functions with specific names
      # Default: []
      exclude-functions:
        - (.*Test).*
        - (.*Mock).*
```

### `check-return-value` Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Check functions returning `(T, error)`

**When to Disable:**

- Functions with different return signatures
- Code that doesn't return errors

### `check-assigns` Option

- **Type**: `bool`
- **Default**: `true` (limited implementation)
- **Description**: Check assignments of `(T, error)` (limited to specific cases)

**Note:** This option has limited implementation in nilerr, primarily for return value checking.

### `exclude-functions` Option

- **Type**: `[]string`
- **Default**: `[]`
- **Description**: Exclude functions matching patterns

**Format:** Regular expression patterns

**Examples:**

```yaml
exclude-functions:
  - (.*Test).* # Exclude test functions
  - (.*Mock).* # Exclude mock functions
  - (.*Example).* # Exclude example functions
  - (.*Benchmark).* # Exclude benchmarks
```

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# Most production applications
version: "2"
linters:
  settings:
    nilerr:
      check-return-value: true
      check-assigns: false

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [nilerr]
```

#### ✅ Strict Configuration

```yaml
# Security-critical, high-quality standards
version: "2"
linters:
  settings:
    nilerr:
      check-return-value: true
      check-assigns: true

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [nilerr]
```

#### ✅ Focused Configuration

```yaml
# Only check return values (most common issue)
version: "2"
linters:
  settings:
    nilerr:
      check-return-value: true
```

#### ✅ Exclude Test Functions

```yaml
# Test code may intentionally violate
version: "2"
linters:
  settings:
    nilerr:
      check-return-value: true
      exclude-functions:
        - (.*Test).*
        - (.*Mock).*

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [nilerr]
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

nilerr works excellently with:

| Linter          | Relationship  | Value                                                                  |
| --------------- | ------------- | ---------------------------------------------------------------------- |
| **errcheck**    | Complementary | errcheck: errors not checked, nilerr: wrong error/return combinations  |
| **nilnil**      | Complementary | nilnil: simultaneous nil error/value, nilerr: nil value with error     |
| **nilnesserr**  | Complementary | nilnesserr: impossible nil combinations, nilerr: wrong return patterns |
| **staticcheck** | Complementary | staticcheck: SA5011 nil deref, nilerr: nil return with error           |
| **govet**       | Complementary | govet: nil dereference, nilerr: error/return contract                  |
| **errorlint**   | Complementary | errorlint: error wrapping, nilerr: error handling correctness          |
| **wrapcheck**   | Complementary | wrapcheck: external error wrapping, nilerr: error return correctness   |

**Complete Error Handling Suite:**

```yaml
linters:
  enable:
    - errcheck # Unchecked errors (CRITICAL)
    - nilerr # Nil value with error (CRITICAL)
    - nilnil # Simultaneous nil error/value (MEDIUM)
    - nilnesserr # Impossible nils (MEDIUM)
    - staticcheck # Deep analysis (CRITICAL)
    - govet # Standard vet (CRITICAL)
    - errorlint # Error wrapping (HIGH)
    - wrapcheck # External error wrapping (HIGH)
```

### Example of Linter Synergy

```go
// nilerr catches:
func process(id int) (*Data, error) {
    data, err := db.GetData(id)
    if err != nil {
        return nil, err  // nilerr: nil returned with non-nil error
    }
    return data, nil
}

// errcheck would catch:
func process(id int) (*Data, error) {
    data, _ := db.GetData(id)  // errcheck: error not checked
    return data, nil
}

// nilnil would catch:
func process(id int) (*Data, error) {
    data, err := db.GetData(id)
    if err != nil {
        return nil, err
    }
    return nil, nil  // nilnil: simultaneous nil error and value
}

// staticcheck SA5011 would catch:
func process(id int) (*Data, error) {
    data, err := db.GetData(id)
    if err != nil {
        return data, err  // staticcheck: nil dereference in line it's guarded
    }
    return data, nil
}
```

### 🔒 No Conflicts

- **No known conflicts** - nilerr focuses specifically on error/return patterns
- **Independent operation** - Doesn't overlap with other linter functionality
- **Safe to enable with all linters**

## Practical Examples

### ✅ Example 1: Database Query

```go
// ❌ BAD: Nil user returned with error
func getUserByID(id int) (*User, error) {
    user, err := db.QueryUser(id)
    if err != nil {
        return nil, err  // nilerr: nil user with non-nil error
    }
    return user, nil
}

// ✅ GOOD: Follow Go error convention
func getUserByID(id int) (*User, error) {
    user, err := db.QueryUser(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}

// Caller (won't panic):
func handler(w http.ResponseWriter, r *http.Request) {
    user, err := getUserByID(123)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, "User: %s", user.Name)  // Safe to dereference
}
```

### ✅ Example 2: File Operation

```go
// ❌ BAD: Nil bytes returned with error
func readFile(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err  // nilerr: nil bytes with non-nil error
    }
    return data, nil
}

// ✅ GOOD: Proper error handling
func readFile(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    return data, nil
}

// Caller (won't panic):
func processFile(path string) {
    data, err := readFile(path)
    if err != nil {
        log.Error(err)
        return
    }
    fmt.Printf("Data: %s", string(data))  // Safe to use
}
```

### ✅ Example 3: HTTP Request

```go
// ❌ BAD: Nil response with error
func fetchUser(id int) (*User, error) {
    resp, err := http.Get(fmt.Sprintf("https://api.example.com/users/%d", id))
    if err != nil {
        return nil, err  // nilerr: nil user with non-nil error
    }
    defer resp.Body.Close()
    var user User
    json.NewDecoder(resp.Body).Decode(&user)
    return &user, nil
}

// ✅ GOOD: Check response status
func fetchUser(id int) (*User, error) {
    resp, err := http.Get(fmt.Sprintf("https://api.example.com/users/%d", id))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }
    return &user, nil
}
```

### ✅ Example 4: Multiple Error Checks

```go
// ❌ BAD: Complex logic, final return violates convention
func validateUser(user *User) (*User, error) {
    if user.Name == "" {
        return user, errors.New("name required")  // nilerr: non-nil user with error
    }
    if user.Email == "" {
        return user, errors.New("email required")
    }
    return user, nil
}

// ✅ GOOD: All returns follow convention
func validateUser(user *User) (*User, error) {
    if user.Name == "" {
        return nil, errors.New("name required")
    }
    if user.Email == "" {
        return nil, errors.New("email required")
    }
    return user, nil
}
```

### ✅ Example 5: Builder Pattern

```go
// ❌ BAD: Partial build, final return violates convention
func buildResponse(data string) (*Response, error) {
    resp := &Response{}
    if data == "" {
        return resp, errors.New("data required")  // nilerr: non-nil resp with error
    }
    resp.Data = data
    return resp, nil
}

// ✅ GOOD: Complete validation
func buildResponse(data string) (*Response, error) {
    resp := &Response{}
    if data == "" {
        return nil, errors.New("data required")
    }
    resp.Data = data
    return resp, nil
}
```

## Best Practices

1. **ALWAYS enable nilerr** for code returning (T, error)
2. **Follow Go error convention**: `error != nil` should imply `value != nil`
3. **Return nil, nil** only when value is nil and error is nil
4. **Return nil, err** when error is non-nil (value can be nil or non-nil, but error should be nil)
5. **Check all error paths** - Ensure all returns with error have nil value
6. **Use helper functions** to ensure consistent error handling
7. **Be explicit** - Don't rely on side effects for error handling
8. **Document unusual patterns** - Add comments if intentionally violating convention
9. **Combine with errcheck** - Complete error handling coverage
10. **Test error paths** - Verify nilerr findings with unit tests

## Common Scenarios and Solutions

### Scenario 1: Database Query Pattern

**Problem:** Common pattern where query fails but nil is returned.

**Solution:** Always return nil, err together.

```go
// ❌ Wrong pattern
user, err := db.Query(id)
if err != nil {
    return nil, err
}
return user, err  // Bug: user might be nil even with error

// ✅ Correct pattern
user, err := db.Query(id)
if err != nil {
    return nil, err
}
return user, nil  // Correct: user valid only when err == nil
```

### Scenario 2: Multiple Validation Steps

**Problem:** Early validation returns partial object, later returns error.

**Solution:** Return nil, nil in early returns, or validate all before returning.

```go
// ❌ Bug: Partial object returned with error
func validate(user *User) (*User, error) {
    if user.Name == "" {
        return user, errors.New("name required")
    }
    if user.Email == "" {
        return user, errors.New("email required")
    }
    return user, nil
}

// ✅ Fix 1: Nil in early returns
func validate(user *User) (*User, error) {
    if user.Name == "" {
        return nil, errors.New("name required")
    }
    if user.Email == "" {
        return nil, errors.New("email required")
    }
    return user, nil
}

// ✅ Fix 2: Validate all before returning
func validate(user *User) (*User, error) {
    var errors []error
    if user.Name == "" {
        errors = append(errors, errors.New("name required"))
    }
    if user.Email == "" {
        errors = append(errors, errors.New("email required"))
    }
    if len(errors) > 0 {
        return nil, errors.Join(errors, "; ")
    }
    return user, nil
}
```

### Scenario 3: Retry Logic

**Problem:** Retry loop with nil return on error.

**Solution:** Ensure all returns from retry follow convention.

```go
// ❌ Bug: Return nil on last retry error
func fetchWithRetry(url string) ([]byte, error) {
    for i := 0; i < 3; i++ {
        data, err := http.Get(url)
        if err == nil {
            return data, nil
        }
        if i == 2 {
            return nil, err  // nilerr: nil data with non-nil error
        }
    }
    return nil, errors.New("all retries failed")
}
```

### Scenario 4: Conditional Logic

**Problem:** Complex conditionals cause return pattern violation.

**Solution:** Simplify logic, ensure all error paths return nil value.

```go
// ❌ Bug: Conditional return violates convention
func processData(input string) (*Result, error) {
    if shouldFail(input) {
        return &Result{}, errors.New("invalid input")  // nilerr: non-nil result with error
    }
    return process(input), nil
}

// ✅ Fix: All error paths return nil
func processData(input string) (*Result, error) {
    if shouldFail(input) {
        return nil, errors.New("invalid input")
    }
    return process(input), nil
}
```

### Scenario 5: Test Mocks

**Problem:** Test mocks intentionally violate convention.

**Solution:** Exclude test files.

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [nilerr]
```

## Summary

**nilerr** is a **CRITICAL** linter for Go error handling:

- ✅ **Highest priority** - Prevents nil pointer panics, enforces Go conventions
- ✅ **Prevents critical bugs** - Catches subtle error handling mistakes
- ✅ **Enforces Go convention** - Ensures `error != nil` implies `value != nil`
- ✅ **Simple check** - Low false positive rate
- ✅ **Comprehensive** - Covers all functions returning (T, error)
- ✅ **Configurable** - Can exclude specific functions
- ✅ **Complementary** - Works with errcheck, staticcheck, nilnil
- ⚠️ **May need exclusions** - Test files, generated code
- ⚠️ **Go-specific** - Only relevant for Go error handling

**Recommendation:** **ALWAYS ENABLE** for all Go code returning (T, error). Follow **Go error convention**: return `(nil, err)` when error is non-nil (value can be nil or non-nil), return `(value, nil)` when error is nil. Combine with **errcheck** for complete error handling coverage. Exclude test files (`(.+)_test\.go`) if needed. Use **helper functions** to ensure consistent error handling patterns across codebase.

**Top 3 Guidelines:**

1. Never return non-nil value with non-nil error
2. Always return nil value when error is nil
3. Return nil, err (nil error, nil value) only when both are nil

---

**Reference:** https://github.com/gostaticanalysis/nilerr
