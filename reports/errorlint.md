# errorlint Linter - Comprehensive Analysis

## What Linter Does

**errorlint** detects issues with Go 1.13+ error handling patterns. It's a specialized linter that focuses on best practices introduced in Go 1.13's error wrapping enhancements. It catches:

- **Non-wrapped errors** that should use `fmt.Errorf()` with `%w` verb
- **Error comparisons** using `==` instead of `errors.Is()` and `errors.As()`
- **Type assertions** on errors that should use `errors.As()` instead of type switches

### The Problem It Detects

Go 1.13 introduced `fmt.Errorf()` with `%w` verb and `errors.Is()`/`errors.As()` functions for improved error handling. Not using these modern patterns makes error chains inconsistent and harder to debug.

**Example Scenarios:**

```go
// ❌ BAD: Error not wrapped with %w
if err := db.Query(); err != nil {
    return fmt.Errorf("query failed: %v", err)  // Loses err in chain
}

// ✅ GOOD: Error wrapped with %w
if err := db.Query(); err != nil {
    return fmt.Errorf("query failed: %w", err)  // Preserves err in chain
}
```

### How It Works

errorlint analyzes:

1. **`fmt.Errorf` calls** - Checks if external errors are wrapped with `%w`
2. **Error comparisons** - Detects `err == sentinelErr` instead of `errors.Is(err, sentinelErr)`
3. **Error type assertions** - Detects `err.(type)` instead of `errors.As(err, &type)`
4. **`fmt.Errorf` with multiple %w** - Checks for incorrect error wrapping patterns

**Analysis Scope:**

- Function bodies
- Error handling patterns
- Return statements with errors
- Error variable assignments

### Examples

```go
// ❌ BAD: External error not wrapped
func fetchUser(id int) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user: %v", err)
    }
    return user, nil
}

// ✅ GOOD: External error wrapped with %w
func fetchUser(id int) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user: %w", err)
    }
    return user, nil
}
```

```go
// ❌ BAD: Error comparison using ==
if err == sql.ErrNoRows {
    return ErrUserNotFound
}

// ✅ GOOD: Error comparison using errors.Is
if errors.Is(err, sql.ErrNoRows) {
    return ErrUserNotFound
}
```

```go
// ❌ BAD: Error type assertion
if e, ok := err.(*MyError); ok {
    return e
}

// ✅ GOOD: Error type assertion using errors.As
var e *MyError
if errors.As(err, &e) {
    return e
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

- **All Go 1.13+ projects** - errorlint depends on Go 1.13 features
- **Production applications** - Ensures consistent error handling
- **Web services and APIs** - Critical for debugging
- **Libraries and SDKs** - Public APIs should use modern error patterns
- **Microservices** - Consistent error handling across services
- **Projects with custom errors** - Type-safe error handling
- **Enterprise applications** - High quality standards required

**Specific Scenarios:**

**1. Codebases with Custom Error Types**

- Domain-specific errors
- Error wrapping libraries
- Multi-layer error handling

**2. Database-Heavy Applications**

- Repository layers wrapping database errors
- Service layers adding context

**3. HTTP Client Code**

- Wrapping HTTP errors with context
- External service integration

**4. File System Operations**

- Wrapping I/O errors
- Adding file path context

**5. Testing Error Chains**

- Verifying errors are properly wrapped
- Testing error unwrapping

### ❌ Disable For:

**Specific Scenarios:**

**1. Go Versions < 1.13**

- errorlint depends on Go 1.13 features
- Upgrade Go or use compatible linters

**2. Code with Intentional Non-Wrapped Errors**

```yaml
# Custom error handling not following %w pattern
linters-settings:
  errorlint:
    # Allow certain patterns
    allow-single-wrapped-error: true
```

**3. Test Files with Legacy Patterns**

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errorlint]
```

**4. Generated Code**

```yaml
run:
  skip-dirs:
    - generated
```

### Priority Assessment

- **Default Priority**: HIGH
- **Value**: **HIGH** - Enforces Go 1.13+ error handling best practices
- **Effort**: Low-Medium - Simple changes, requires understanding of new patterns
- **Recommendation**: **RECOMMEND** for all Go 1.13+ projects

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    errorlint:
      # Check errorf for single wrapped errors (default: false)
      allow-single-wrapped-error: false

      # Check for %w wrapping errors with different types (default: true)
      errorf: true

      # Check for errorf with multiple %w (default: true)
      errorf-multi: true

      # Check for error assertions (default: true)
      asserts: true

      # Check for error comparisons (default: true)
      comparison: true

      # Check for non-wrapped errors (default: true)
      default-is: true
```

### `allow-single-wrapped-error` Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: Allow `fmt.Errorf` with single `%w` error

**When to Enable:**

- When custom error handling strategy doesn't use multiple wraps
- When transitioning to new patterns gradually

### `errorf` Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Check `fmt.Errorf` format strings

### `errorf-multi` Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Check `fmt.Errorf` with multiple `%w` verbs

### `asserts` Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Check error assertions

### `comparison` Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Check error comparisons using `==`

### `default-is` Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Check non-wrapped external errors

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# Most Go 1.13+ projects
version: "2"
linters:
  settings:
    errorlint:
      errorf: true
      errorf-multi: true
      asserts: true
      comparison: true
      default-is: true

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errorlint]
```

#### ✅ Strict Configuration

```yaml
# High-quality standards
version: "2"
linters:
  settings:
    errorlint:
      errorf: true
      errorf-multi: true
      asserts: true
      comparison: true
      default-is: true
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

errorlint works excellently with:

| Linter          | Relationship  | Value                                                                       |
| --------------- | ------------- | --------------------------------------------------------------------------- |
| **wrapcheck**   | Complementary | wrapcheck: external error wrapping, errorlint: error wrapping patterns (%w) |
| **errcheck**    | Complementary | errcheck: errors not checked, errorlint: correct wrapping patterns          |
| **goerr113**    | Complementary | goerr113: error expressions, errorlint: error comparison (errors.Is)        |
| **nilerr**      | Complementary | nilerr: nil errors with non-nil values, errorlint: error wrapping           |
| **staticcheck** | Complementary | staticcheck: SA5009, errorlint: %w usage                                    |
| **nilnesserr**  | Complementary | nilnesserr: impossible nils, errorlint: error type checks                   |

**Complete Error Handling Suite:**

```yaml
linters:
  enable:
    - errorlint # Error wrapping patterns (HIGH)
    - wrapcheck # External error wrapping (HIGH)
    - errcheck # Error handling (CRITICAL)
    - goerr113 # Error expressions (MEDIUM)
    - nilerr # Nil errors (CRITICAL)
    - nilnil # Simultaneous nils (MEDIUM)
    - staticcheck # Deep analysis (CRITICAL)
```

### Example of Linter Synergy

```go
// errcheck catches:
func process() error {
    _, err := fetchData()
    return err  // errcheck: error not checked
}

// errorlint adds (after errcheck fixed):
func process() error {
    _, err := fetchData()
    return fmt.Errorf("failed to fetch: %v", err)  // errorlint: should use %w
}

// wrapcheck adds (after errorlint fixed):
func process() error {
    _, err := fetchData()
    return err  // wrapcheck: error from external package not wrapped
}

// All linters satisfied:
func process() error {
    _, err := fetchData()
    return fmt.Errorf("failed to fetch: %w", err)
}
```

### 🔒 Minimal Overlap, No Conflicts

| Linter                      | Overlap                   | Recommendation                                   |
| --------------------------- | ------------------------- | ------------------------------------------------ |
| errorlint + **wrapcheck**   | Both check error wrapping | Use both - complementary                         |
| errorlint + **staticcheck** | Some SA5009 overlap       | Use both - errorlint more specific               |
| errorlint + **errcheck**    | Different focus           | Use both - errorlint wrapping, errcheck checking |

**Why No Conflicts:**

- errorlint: Specific Go 1.13+ error patterns
- wrapcheck: External package error wrapping
- Each covers different aspects of error handling

## Practical Examples

### ✅ Example 1: Database Repository

```go
// ❌ BAD: Repository doesn't wrap database errors
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) FindByID(id int) (*User, error) {
    row := r.db.QueryRow("SELECT * FROM users WHERE id = ?", id)
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to scan user: %v", err)  // errorlint: should use %w
    }
    return &user, nil
}

// ✅ GOOD: Repository wraps errors with %w
func (r *UserRepository) FindByID(id int) (*User, error) {
    row := r.db.QueryRow("SELECT * FROM users WHERE id = ?", id)
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to scan user %d: %w", id, err)
    }
    return &user, nil
}
```

### ✅ Example 2: Service Layer

```go
// ❌ BAD: Service doesn't wrap repository errors
type UserService struct {
    repo *UserRepository
}

func (s *UserService) GetUser(id int) (*User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %v", err)  // errorlint: should use %w
    }
    return user, nil
}

// ✅ GOOD: Service wraps errors with context
func (s *UserService) GetUser(id int) (*User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return user, nil
}
```

### ✅ Example 3: Error Comparison

```go
// ❌ BAD: Direct comparison
func process(err error) error {
    if err == sql.ErrNoRows {  // errorlint: should use errors.Is
        return ErrUserNotFound
    }
    return nil
}

// ✅ GOOD: Use errors.Is
func process(err error) error {
    if errors.Is(err, sql.ErrNoRows) {
        return ErrUserNotFound
    }
    return nil
}
```

### ✅ Example 4: Type Assertion

```go
// ❌ BAD: Type assertion
func process(err error) error {
    if e, ok := err.(*ValidationError); ok {  // errorlint: should use errors.As
        return e
    }
    return nil
}

// ✅ GOOD: Use errors.As
func process(err error) error {
    var e *ValidationError
    if errors.As(err, &e) {
        return e
    }
    return nil
}
```

### ✅ Example 5: Multiple Wrapped Errors

```go
// ❌ BAD: Only one error wrapped, others not
func processAll() error {
    err1 := step1()
    if err1 != nil {
        return fmt.Errorf("step1 failed: %w", err1)  // Only err1 in chain
    }
    err2 := step2()
    if err2 != nil {
        return fmt.Errorf("step2 failed: %v", err2)  // err2 not wrapped
    }
    return nil
}

// ✅ GOOD: All errors in chain
func processAll() error {
    err1 := step1()
    if err1 != nil {
        return fmt.Errorf("step1 failed: %w", err1)
    }
    err2 := step2()
    if err2 != nil {
        return fmt.Errorf("step2 failed: %w", err2)  // All errors wrapped
    }
    return nil
}
```

## Best Practices

1. **ALWAYS use `fmt.Errorf()` with `%w`** - Preserves error chain
2. **Use `errors.Is()` for comparisons** - Don't use `==` for errors
3. **Use `errors.As()` for type assertions** - Don't use type assertions
4. **Wrap at each layer** - Don't let errors propagate unwrapped
5. **Add context in error messages** - Include operation name, IDs, values
6. **Combine with wrapcheck** - External package wrapping
7. **Combine with errcheck** - Complete error handling
8. **Test error chains** - Verify errors unwrap correctly
9. **Don't wrap sentinel errors** - Only wrap errors from external packages
10. **Use consistent wrapping patterns** - Establish team standards

## Common Scenarios and Solutions

### Scenario 1: Repository Layer

**Problem:** Repository returns raw database errors without wrapping.

**Solution:** Wrap database errors with repository-specific context.

```go
// Repository level
return nil, fmt.Errorf("failed to find user %d: %w", id, dbErr)
```

### Scenario 2: Service Layer

**Problem:** Service doesn't wrap repository errors.

**Solution:** Wrap repository errors with service-specific context.

```go
// Service level
return nil, fmt.Errorf("failed to get user %d: %w", id, repoErr)
```

### Scenario 3: HTTP Client Layer

**Problem:** Client returns HTTP errors without wrapping.

**Solution:** Wrap HTTP errors with request context.

```go
// Client level
return nil, fmt.Errorf("HTTP GET %s failed: %w", url, httpErr)
```

### Scenario 4: Multiple External Calls

**Problem:** Function makes multiple external calls, only wraps last error.

**Solution:** Wrap each error with step-specific context.

```go
// Each step wrapped
if err := step1(); err != nil {
    return fmt.Errorf("step1 failed: %w", err)
}
```

### Scenario 5: Custom Error Types

**Problem:** Code has custom error types, doesn't use errors.As.

**Solution:** Use errors.As for type-safe error handling.

```go
// Custom error type checking
var e *MyError
if errors.As(err, &e) {
    return e
}
```

## Summary

**errorlint** is a **HIGH** priority linter for Go 1.13+ error handling:

- ✅ **Enforces modern patterns** - Go 1.13+ error wrapping (%w, errors.Is, errors.As)
- ✅ **Improves debugging** - Preserves error chains for better error tracing
- ✅ **Consistent error handling** - Prevents mixed wrapped/unwrapped patterns
- ✅ **Type-safe** - Encourages errors.As for custom error types
- ✅ **Simple to use** - Clear fix patterns, minimal configuration
- ✅ **Comprehensive** - Covers %w wrapping, comparisons, type assertions
- ✅ **Complementary** - Works with wrapcheck, errcheck, goerr113
- ⚠️ **Go 1.13+ required** - Depends on modern Go features
- ⚠️ **Thoughtful wrapping** - Good error messages require effort
- ⚠️ **May need exclusions** - Test files, custom error handling

**Recommendation:** **RECOMMEND** for all Go 1.13+ projects. Always use `fmt.Errorf()` with `%w` when wrapping errors from external packages. Use `errors.Is()` for error comparisons instead of `==`. Use `errors.As()` for type assertions instead of type switches. Wrap errors at each layer (repository, service, handler) with context about operation, IDs, paths. Exclude test files (`(.+)_test\.go`) when using legacy patterns. Combine with **wrapcheck** for external error wrapping and **errcheck** for complete error handling coverage.

**Top 3 Configuration Tips:**

1. Use default settings - Already well-tuned for Go 1.13+ patterns
2. Enable all checks - errorf, errorf-multi, asserts, comparison, default-is
3. Combine with wrapcheck for complete external error wrapping coverage

---

**Reference:** https://github.com/polyfloyd/go-errorlint
