# wrapcheck Linter - Comprehensive Analysis

## What the Linter Does

**wrapcheck** checks that errors returned from external packages are wrapped before being returned. In Go 1.13+, error wrapping using `fmt.Errorf` with `%w` is the recommended way to add context to errors and create an error chain. Without wrapping, debugging becomes difficult as you lose the call stack context, and error handling becomes inconsistent across a codebase.

### The Problem It Detects

Go projects often call external packages (database drivers, HTTP clients, etc.) that return errors. Without wrapping these errors, you lose context about what operation was being performed when the error occurred. This makes debugging difficult and violates Go 1.13+ best practices for error handling.

**Example Scenario:**
```go
// ❌ BAD: Error from external package not wrapped
func (s *Service) GetUser(id int) (*User, error) {
    user, err := s.db.Query(id)  // err from external package (sql.DB)
    if err != nil {
        return nil, err  // wrapcheck: error from external package not wrapped
    }
    return user, nil
}

// When this error is logged, we only see "no such user"
// We don't know which service function failed or the user ID
```

**Why This Matters:**
- **Lost context** - Don't know what operation failed
- **Poor debugging** - Can't trace error through call stack
- **Inconsistent error handling** - Some errors wrapped, some not
- **Poor user experience** - Generic error messages
- **Compliance issue** - Violates Go 1.13+ error handling best practices

### How It Works

wrapcheck analyzes function return statements and checks:

1. **Return statement analysis** - Identifies functions returning `(T, error)`
2. **Error source tracking** - Determines if error is from external package
3. **Wrap detection** - Checks if error is wrapped with `%w` or `errors.Wrap()`
4. **Exclusion list** - Allows certain functions to be excluded from wrapping requirement

**Excluded by Default:**
- `errors.New()` and `errors.New()` are creating new errors, not wrapping
- `fmt.Errorf()` (without `%w`) creates new error, doesn't wrap
- Standard library functions that don't need wrapping

### Examples

```go
// ❌ BAD: External error not wrapped
func (s *Service) FetchUser(id int) (*User, error) {
    user, err := s.client.Get(fmt.Sprintf("/users/%d", id))
    if err != nil {
        return nil, err  // wrapcheck: error from external package not wrapped
    }
    return user, nil
}

// ✅ GOOD: External error wrapped with context
func (s *Service) FetchUser(id int) (*User, error) {
    user, err := s.client.Get(fmt.Sprintf("/users/%d", id))
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user %d: %w", id, err)
    }
    return user, nil
}
```

```go
// ❌ BAD: Database error not wrapped
func (r *Repository) Save(user *User) error {
    if err := r.db.Create(user); err != nil {
        return err  // wrapcheck: error from external package not wrapped
    }
    return nil
}

// ✅ GOOD: Database error wrapped
func (r *Repository) Save(user *User) error {
    if err := r.db.Create(user); err != nil {
        return fmt.Errorf("failed to save user %s: %w", user.ID, err)
    }
    return nil
}
```

```go
// ❌ BAD: Chain of unwrapped errors
func process() error {
    data, err := readFile()
    if err != nil {
        return err  // wrapcheck: error from external package not wrapped
    }
    if err := processData(data); err != nil {
        return err  // wrapcheck: error from external package not wrapped
    }
    return nil
}

// ✅ GOOD: Each error wrapped with context
func process() error {
    data, err := readFile()
    if err != nil {
        return fmt.Errorf("failed to read file: %w", err)
    }
    if err := processData(data); err != nil {
        return fmt.Errorf("failed to process data: %w", err)
    }
    return nil
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **All production applications** - Critical for debugging and observability
- **Web services and APIs** - Many external package calls (databases, HTTP clients)
- **Microservices** - Inter-service communication needs context in errors
- **REST/GraphQL APIs** - HTTP client calls to external services
- **CLI tools** - Better error messages for users
- **Libraries and SDKs** - Public APIs should wrap errors consistently
- **Database-heavy applications** - Many database driver calls
- **Multi-service applications** - Consistency across services

**Specific Scenarios:**

**1. Database Operations**
- All queries, inserts, updates, deletes
- Transaction operations
- Connection pool operations

**2. HTTP Client Calls**
- REST API calls
- GraphQL queries
- Webhook requests
- External service integrations

**3. File System Operations**
- Reading config files
- Writing data files
- Directory operations
- Archive extraction

**4. Background Processing**
- Worker pools
- Async operations
- Scheduled jobs
- Event handlers

**5. Public APIs**
- Library functions calling external code
- SDK methods wrapping external services
- Framework code integrating databases/HTTP clients

### ❌ Disable For:

**Specific Scenarios:**

**1. Test Files with Intentionally Unwrapped Errors**
```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [wrapcheck]
```

**2. Packages That Create New Errors**
```yaml
# Packages that only create new errors, not wrap
linters-settings:
  wrapcheck:
    ignore-sig: (.*Error)\.New\(
    ignore-sig: (.*Error)\.Newf\(
    ignore-sig: errors\.New\(
```

**3. Functions That Intentionally Return External Errors**
```yaml
# Specific functions where wrapping is not needed
linters-settings:
  wrapcheck:
    extra-ignore-sigs:
      - (.*Test).*\..*(
      - (.*Mock).*\..*(
```

**4. Generated Code**
```yaml
run:
  skip-dirs:
    - generated
    - vendor

issues:
  exclude-rules:
    - path: (.+)_generated\.go
      linters: [wrapcheck]
```

**5. Legacy Code Under Migration**
```yaml
# When migrating to error wrapping, temporarily exclude
linters-settings:
  wrapcheck:
    ignore-package-globs:
      - github.com/legacy/internal/*
```

### Priority Assessment

- **Default Priority**: HIGH
- **Value**: **HIGH** - Improves debugging, ensures Go 1.13+ best practices
- **Effort**: Low-Medium - Adding `%w` is simple but requires thought for good messages
- **Recommendation**: **RECOMMEND** for all production code

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    wrapcheck:
      # Additional ignored function signatures (extends defaults)
      # Default: []
      ignore-sig:
        - .Errorf(
        - errors.New(
        - errors.Unwrap(
        - .Wrap(
        - .Wrapf(

      # Additional ignored signatures (extends defaults)
      # Default: []
      extra-ignore-sigs:
        - (.*Test).*\..*(
        - (.*Mock).*\..*(
        - NewClient.*\(

      # Ignore errors from specific packages (glob patterns)
      # Default: []
      ignore-package-globs:
        - encoding/*
        - github.com/my/internal/*

      # Report errors from internal packages (default: false)
      # Default: false
      report-internal-errors: true
```

### `ignore-sig` Option

- **Type**: `[]string`
- **Default**: [`.Errorf(`, `errors.New(`, `errors.Unwrap(`, `.Wrap(`, `.Wrapf(`]
- **Description**: Function signatures that don't need error wrapping

**Built-in Exclusions:**
- Functions that create new errors (not wrap existing)
- Error constructor functions
- Functions that intentionally unwrap errors

**Custom Exclusions:**
```yaml
ignore-sig:
  - (.*Error)\.New\(      # Error constructors
  - (.*Error)\.Errorf\(   # Error format constructors
  - (.*Error)\.Wrap\(     # Explicit wrap functions
  - (.*Error)\.Unwrap\(   # Explicit unwrap functions
  - (.*Wrapper)\.Wrap\(   # Wrapper functions
```

### `extra-ignore-sigs` Option

- **Type**: `[]string`
- **Default**: `[]`
- **Description**: Additional ignored signatures on top of defaults

**Use Cases:**
- Test functions that intentionally return external errors
- Mock functions
- Wrapper functions that already handle wrapping

**Example:**
```yaml
extra-ignore-sigs:
  - (.*Test).*\..*(
  - (.*Mock).*\..*(
  - NewClient.*\(        # Client constructors
  - Connect.*\(          # Connection functions
```

### `ignore-package-globs` Option

- **Type**: `[]string`
- **Default**: `[]`
- **Description**: Ignore errors from specific packages (glob patterns)

**Use Cases:**
- Internal packages that wrap errors consistently
- Third-party packages with known error handling
- Packages being migrated

**Format:** Glob pattern for package paths

**Example:**
```yaml
ignore-package-globs:
  - github.com/my/internal/*        # Ignore internal packages
  - github.com/legacy/db/*       # Ignore legacy DB wrapper
  - github.com/vendor/errors/*     # Ignore error utils
```

### `report-internal-errors` Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Report errors from internal packages (usually disabled)

**When to Enable:**
- When internal packages should also follow wrapping rules
- When internal code quality needs enforcement

**When to Disable:**
- When internal packages use different error handling strategy
- When internal functions are wrappers that handle wrapping

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Most production applications
version: "2"
linters:
  settings:
    wrapcheck:
      # Use default exclusions (error constructors)
      ignore-sig: []

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [wrapcheck]
```

#### ✅ Strict Configuration
```yaml
# High-quality standards, no exclusions
version: "2"
linters:
  settings:
    wrapcheck:
      # Require wrapping of all external errors
      ignore-sig: []
      extra-ignore-sigs: []
      ignore-package-globs: []
      report-internal-errors: true

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [wrapcheck]
```

#### ✅ Database-Focused Configuration
```yaml
# Applications with heavy database usage
version: "2"
linters:
  settings:
    wrapcheck:
      # Ignore specific database patterns
      extra-ignore-sigs:
        - (.*DB).*\..*(
        - (.*Query).*\(

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [wrapcheck]
```

#### ✅ HTTP Client Configuration
```yaml
# Microservices with many HTTP calls
version: "2"
linters:
  settings:
    wrapcheck:
      # Ignore HTTP client constructor
      extra-ignore-sigs:
        - NewClient.*\(
        - Do.*\(

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [wrapcheck]
```

#### ✅ Internal Package Exclusion
```yaml
# Internal packages already wrap errors
version: "2"
linters:
  settings:
    wrapcheck:
      ignore-package-globs:
        - github.com/my/internal/*
        - github.com/my/wrappers/*

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [wrapcheck]
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

wrapcheck works excellently with:

| Linter | Relationship | Value |
|--------|--------------|---------|
| **errcheck** | Complementary | errcheck: errors not checked, wrapcheck: errors not wrapped |
| **errorlint** | Complementary | errorlint: error wrapping patterns (%w, errors.As), wrapcheck: external errors |
| **staticcheck** | Complementary | staticcheck: deep error analysis, wrapcheck: external error wrapping |
| **gosec** | Complementary | gosec: security issues, wrapcheck: error context for security events |
| **goerr113** | Complementary | goerr113: error expressions, wrapcheck: external error wrapping |
| **errchkjson** | Complementary | errchkjson: JSON type safety, wrapcheck: error wrapping for JSON |
| **nilerr** | Complementary | nilerr: nil errors with non-nil values, wrapcheck: external errors |
| **errorfmt** | Complementary | errorfmt: error format strings, wrapcheck: error wrapping |
| **goimports** | Complementary | goimports: import order, wrapcheck: code quality |
| **revive** | Complementary | revive: general style, wrapcheck: error handling |

**Complete Error Handling Suite:**
```yaml
linters:
  enable:
    - wrapcheck       # External error wrapping (HIGH)
    - errcheck        # Error handling (CRITICAL)
    - errorlint       # Error wrapping patterns (HIGH)
    - nilerr          # Nil error returns (CRITICAL)
    - staticcheck      # Deep analysis (CRITICAL)
    - gosec           # Security (CRITICAL)
    - goerr113        # Error expressions (MEDIUM)
```

### Example of Linter Synergy

```go
// errcheck catches:
func (s *Service) GetUser(id int) (*User, error) {
    user, _ := s.db.Query(id)  // errcheck: error not checked
    return user, nil
}

// wrapcheck adds (after errcheck fixed):
func (s *Service) GetUser(id int) (*User, error) {
    user, err := s.db.Query(id)
    if err != nil {
        return user, err  // wrapcheck: error from external package not wrapped
    }
    return user, nil
}

// errorlint adds (after wrapcheck fixed):
func (s *Service) GetUser(id int) (*User, error) {
    user, err := s.db.Query(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)  // errorlint: errors should be wrapped with %w
    }
    return user, nil
}

// All linters satisfied:
func (s *Service) GetUser(id int) (*User, error) {
    user, err := s.db.Query(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return user, nil
}
```

### 🔒 Minimal Overlap, No Conflicts

| Linter | Overlap | Recommendation |
|---------|----------|----------------|
| wrapcheck + **errorlint** | Both check error wrapping | Use both - wrapcheck focuses on external packages |
| wrapcheck + **errcheck** | errcheck finds more, wrapcheck is subset | Use both - complement each other |
| wrapcheck + **goerr113** | Both check error handling | Use both - different focus |

**Why No Conflicts:**
- wrapcheck: Specifically checks external error wrapping
- errorlint: Checks general error wrapping patterns
- errcheck: Checks if errors are checked at all
- Each linter covers different aspects of error handling

## Practical Examples

### ✅ Example 1: Database Repository

```go
// ❌ BAD: Repository doesn't wrap database errors
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) FindByID(id int) (*User, error) {
    user, err := r.db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, err  // wrapcheck: error from sql.DB not wrapped
    }
    return user, nil
}

// ✅ GOOD: Repository wraps database errors with context
func (r *UserRepository) FindByID(id int) (*User, error) {
    user, err := r.db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, fmt.Errorf("failed to find user by id %d: %w", id, err)
    }
    return user, nil
}

// Usage: Clear error context
user, err := repo.FindByID(123)
if err != nil {
    log.Printf("Failed to find user: %v", err)  // Shows: "failed to find user by id 123: no such user"
    return
}
```

### ✅ Example 2: HTTP Client

```go
// ❌ BAD: HTTP client doesn't wrap errors
type APIClient struct {
    client *http.Client
    baseURL string
}

func (c *APIClient) FetchUser(id int) (*User, error) {
    url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
    resp, err := c.client.Get(url)
    if err != nil {
        return nil, err  // wrapcheck: error from http.Client not wrapped
    }
    defer resp.Body.Close()
    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }
    return &user, nil
}

// ✅ GOOD: HTTP client wraps errors with context
func (c *APIClient) FetchUser(id int) (*User, error) {
    url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
    resp, err := c.client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user %d from %s: %w", id, c.baseURL, err)
    }
    defer resp.Body.Close()
    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("failed to decode user response: %w", err)
    }
    return &user, nil
}
```

### ✅ Example 3: Background Worker

```go
// ❌ BAD: Worker doesn't wrap errors
func processWorker(ctx context.Context, jobs <-chan Job) {
    for job := range jobs {
        err := job.Process(ctx)
        if err != nil {
            log.Error(err)  // wrapcheck: error from external package not wrapped
        }
    }
}

// ✅ GOOD: Worker wraps errors with job context
func processWorker(ctx context.Context, jobs <-chan Job) {
    for job := range jobs {
        err := job.Process(ctx)
        if err != nil {
            log.Printf("failed to process job %s: %v", job.ID, err)  // wrapcheck: error from external package not wrapped
            return
        }
    }
}
```

### ✅ Example 4: File Operations

```go
// ❌ BAD: File loader doesn't wrap errors
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err  // wrapcheck: error from os package not wrapped
    }
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err  // wrapcheck: error from json package not wrapped
    }
    return &config, nil
}

// ✅ GOOD: File loader wraps errors with path context
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
    }
    return &config, nil
}
```

### ✅ Example 5: Multi-Service Call Chain

```go
// ❌ BAD: Service doesn't wrap errors from multiple services
type OrderService struct {
    dbService *DBService
    paymentService *PaymentService
    inventoryService *InventoryService
}

func (s *OrderService) CreateOrder(order *Order) error {
    // Step 1: Validate with DB
    if err := s.dbService.Validate(order); err != nil {
        return err  // wrapcheck: error from external package not wrapped
    }

    // Step 2: Process payment
    if err := s.paymentService.Process(order); err != nil {
        return err  // wrapcheck: error from external package not wrapped
    }

    // Step 3: Update inventory
    if err := s.inventoryService.Update(order); err != nil {
        return err  // wrapcheck: error from external package not wrapped
    }

    return nil
}

// ✅ GOOD: Service wraps each error with context
func (s *OrderService) CreateOrder(order *Order) error {
    // Step 1: Validate with DB
    if err := s.dbService.Validate(order); err != nil {
        return fmt.Errorf("order validation failed: %w", err)
    }

    // Step 2: Process payment
    if err := s.paymentService.Process(order); err != nil {
        return fmt.Errorf("payment processing failed for order %s: %w", order.ID, err)
    }

    // Step 3: Update inventory
    if err := s.inventoryService.Update(order); err != nil {
        return fmt.Errorf("inventory update failed for order %s: %w", order.ID, err)
    }

    return nil
}
```

## Best Practices

1. **ALWAYS enable wrapcheck** for production code - Critical for debugging
2. **Wrap all external errors** - Add context about what operation failed
3. **Use `fmt.Errorf` with `%w`** - Standard Go 1.13+ error wrapping
4. **Provide meaningful context** - Include operation name, IDs, values in error message
5. **Keep error messages concise** - Don't over-wrapping with redundant context
6. **Wrap at each layer** - Don't let external errors propagate unwrapped
7. **Combine with errcheck** - Ensure errors are also checked
8. **Combine with errorlint** - Ensure correct wrapping patterns
9. **Test error chains** - Verify errors unwrap correctly for debugging
10. **Document wrapping strategy** - Establish consistent patterns across codebase

## Common Scenarios and Solutions

### Scenario 1: Database Layer

**Problem:** Repository returns raw database errors.

**Solution:** Wrap at repository layer with context.
```yaml
linters:
  settings:
    wrapcheck: {}
```

### Scenario 2: HTTP Client Layer

**Problem:** Client returns raw HTTP errors.

**Solution:** Wrap at client layer with request context.
```yaml
linters:
  settings:
    wrapcheck:
      extra-ignore-sigs:
        - NewClient.*\(  # Don't wrap client constructor
```

### Scenario 3: Multiple External Calls

**Problem:** Function makes multiple external calls, doesn't wrap errors.

**Solution:** Wrap each error with specific context at each step.
```go
// Each step wrapped with context
if err := step1(); err != nil {
    return fmt.Errorf("step 1 failed: %w", err)
}
if err := step2(); err != nil {
    return fmt.Errorf("step 2 failed: %w", err)
}
```

### Scenario 4: Internal Packages

**Problem:** Internal packages already wrap errors, wrapcheck flags them.

**Solution:** Exclude internal packages or disable report-internal-errors.
```yaml
linters:
  settings:
    wrapcheck:
      ignore-package-globs:
        - github.com/my/internal/*
        - github.com/my/wrappers/*
```

### Scenario 5: Test Mocks

**Problem:** Test mocks intentionally return unwrapped external errors.

**Solution:** Exclude test functions.
```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [wrapcheck]
linters:
  settings:
    wrapcheck:
      extra-ignore-sigs:
        - (.*Test).*\..*(
        - (.*Mock).*\..*(
```

## Summary

**wrapcheck** is a **HIGH** priority linter for external error wrapping:

- ✅ **Improves debugging** - Adds context to errors from external packages
- ✅ **Ensures best practices** - Enforces Go 1.13+ error wrapping patterns
- ✅ **Consistent error handling** - Prevents mix of wrapped/unwrapped errors
- ✅ **Configurable** - Can exclude specific functions/packages as needed
- ✅ **Complementary** - Works excellently with errcheck, errorlint, staticcheck
- ✅ **Low overhead** - Simple check, fast execution
- ⚠️ **May need exclusions** - Test mocks, error constructors, internal packages
- ⚠️ **Thoughtful wrapping required** - Good error messages require effort

**Recommendation:** **RECOMMEND** for all production code, especially services, APIs, and applications with external package calls. Wrap all errors from external packages using `fmt.Errorf()` with `%w` and meaningful context. Use `ignore-sig` and `extra-ignore-sigs` to exclude error constructors and specific functions. Exclude test files (`(.+)_test\.go`). Combine with **errcheck** and **errorlint** for complete error handling coverage.

**Top 3 Configuration Tips:**
1. Keep default exclusions (error constructors like `errors.New()`)
2. Use `extra-ignore-sigs` for test mocks and client constructors
3. Wrap errors at service/repository layers, not at top level

---

**Reference:** https://github.com/tomarrell/wrapcheck
