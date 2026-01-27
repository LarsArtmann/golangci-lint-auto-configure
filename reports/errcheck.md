# errcheck Linter - Comprehensive Analysis

## What the Linter Does

**errcheck** is a **critical static analysis tool** that finds **silently ignored errors in Go code**. It identifies cases where functions return an error value that is neither checked nor explicitly discarded by assignment to `_`. This is one of the most important linters for Go development, as unchecked errors are the #1 source of production bugs and security vulnerabilities.

### The Problem It Detects

Go's error handling model requires explicit checking, but developers sometimes accidentally ignore errors:

```go
// ❌ SILENT FAILURE - Error completely ignored
file, _ := os.Open("config.yaml")  // If file doesn't exist, file == nil
readFile(file)  // Will panic, hard to debug

// ❌ ERROR IN GOROUTINE - Asynchronous failure invisible
go func() {
    saveUserData(user)  // Error silently lost
}()

// ❌ METHOD CALL ERROR IGNORED
writer.Write([]byte("important data"))  // Data loss, no indication
```

**Why This is Critical:**
- **System crashes** - Nil dereference when operation fails
- **Data corruption** - Writing to failed resources
- **Security vulnerabilities** - Weak cryptographic keys, exposed sensitive data
- **Silent failures** - Operations that appear to work but don't
- **Data loss** - Files not written, network requests not completed

### How It Works

errcheck analyzes all callable expressions (functions, methods) and ensures that:

1. **Error is assigned** to a variable for later checking
2. **Error is immediately checked** with `if err != nil`
3. **Error is explicitly discarded** using `_` to show intent

**Analysis Scope:**
- Function calls: `f()`
- Method calls: `obj.Method()`
- Type assertions: `x.(T)` (if enabled)
- All callables not in exclusion list

**What errcheck Does NOT Do:**
- Analyze whether errors are properly handled after assignment
- Check if error handling is correct (use `staticcheck` for this)
- Understand business logic or context
- Detect panics or runtime errors (use `govet` for these)

### Examples

```go
// ❌ BAD: Completely ignored error
func loadConfig() (*Config, error) {
    data, _ := os.ReadFile("config.yaml")  // errcheck: unchecked error
    var config Config
    json.Unmarshal(data, &config)
    return &config, nil
}

// ✅ GOOD: Error checked
func loadConfig() (*Config, error) {
    data, err := os.ReadFile("config.yaml")
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    return &config, nil
}
```

```go
// ❌ BAD: Error in goroutine
func processItems(items []Item) {
    for _, item := range items {
        go func(i Item) {
            saveToDatabase(i)  // errcheck: error ignored in goroutine
        }(item)
    }
}

// ✅ GOOD: Error handled in goroutine
func processItems(items []Item) {
    for _, item := range items {
        go func(i Item) {
            if err := saveToDatabase(i); err != nil {
                log.Printf("Failed to save %v: %v", i, err)
            }
        }(item)
    }
}
```

```go
// ❌ BAD: Type assertion error ignored
func handleValue(val interface{}) {
    str := val.(string)  // errcheck: type assertion error ignored
    fmt.Println(str)
}

// ✅ GOOD: Type assertion checked (with check-type-assertions: true)
func handleValue(val interface{}) {
    str, ok := val.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", val)
    }
    fmt.Println(str)
}
```

```go
// ❌ BAD: Security vulnerability
func generateKey() (*rsa.PrivateKey, error) {
    key, _ := rsa.GenerateKey(2048)  // errcheck: security risk
    return key, nil  // Returns nil key on error
}

// ✅ GOOD: Secure error handling
func generateKey() (*rsa.PrivateKey, error) {
    key, err := rsa.GenerateKey(2048)
    if err != nil {
        return nil, fmt.Errorf("failed to generate key: %w", err)
    }
    return key, nil
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

| Project Type | Priority | Justification |
|--------------|----------|----------------|
| **Web Services/APIs** | CRITICAL | User-facing, reliability critical |
| **CLI Tools** | CRITICAL | User experience depends on errors |
| **Libraries/SDKs** | CRITICAL | Public APIs need correct error handling |
| **Database Tools** | CRITICAL | Data integrity at risk |
| **System Tools** | HIGH | System stability |
| **Microservices** | CRITICAL | Distributed systems require robust error handling |
| **Financial/Security Software** | CRITICAL | Security vulnerabilities from ignored errors |
| **One-off Scripts** | LOW | Optional, still recommended |

**Specific Scenarios:**

**1. Production Applications**
- Any code deployed to production
- User-facing applications
- Services processing customer data
- APIs with SLA requirements

**2. Security-Sensitive Code**
- Authentication/authorization logic
- Cryptographic operations (hashing, encryption, signing)
- Random number generation
- File operations (reading config, user uploads)
- Network communication
- Database transactions

**3. Data Integrity Critical Systems**
- Financial transaction processing
- Data persistence layers
- Backup systems
- Data migration tools
- Logging and monitoring

**4. Public Libraries/SDKs**
- Code consumed by other developers
- APIs with error handling contracts
- SDKs for external services
- Framework libraries

**5. Long-Lived Projects**
- Projects maintained for months/years
- Multiple contributors
- Codebases growing over time

**6. Code with Heavy External Dependencies**
- Many database operations
- Network calls to external APIs
- File I/O operations
- System calls

### ❌ Disable For:

**Specific Scenarios (with Exclusions, Not Full Disabling):**

**1. Test Files** (`_test.go`)
- Tests often use simplified error handling
- Mock functions may intentionally ignore errors
- Benchmarks prioritize performance

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errcheck]
```

**2. Generated Code**
- Protobuf-generated code
- Mock-generated code
- API client libraries
- Code generation tools

```yaml
run:
  skip-dirs:
    - generated
    - vendor

issues:
  exclude-rules:
    - path: (.+)_pb\.go
    - path: (.+)_generated\.go
      linters: [errcheck]
```

**3. Specific Known-Safe Functions**

Functions documented to never return errors:

```yaml
linters-settings:
  errcheck:
    exclude-functions:
      # Buffer operations cannot fail (in-memory)
      - (*bytes.Buffer).Write
      - (*bytes.Buffer).WriteString
      - (*bytes.Buffer).WriteByte

      # fmt.State writes cannot fail
      - io.WriteString(fmt.State)

      # Logger sync is best-effort
      - (github.com/go-kit/log.Logger).Sync
      - (*go.uber.org/zap.Logger).Sync
      - (*logrus.Logger).Sync

      # HTTP responses where failure is acceptable
      - (*net/http.ResponseWriter).Write
```

**4. Graceful Degradation Patterns**

Where continuing despite error is intentional:

```go
// Best-effort cleanup in defer
defer func() {
    // Don't fail if file already removed
    _ = os.Remove(tempFile)
}()

// Attempt logging, don't fail
_ = debugLog(debugData)  // Add comment explaining why
```

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** - Prevents production bugs and security vulnerabilities
- **Effort**: Low - Simple check, minimal false positives
- **Recommendation**: **ALWAYS ENABLE** for all production code

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    errcheck:
      # Report unchecked errors in type assertions
      # Default: false
      check-type-assertions: false

      # Report errors assigned to blank identifier
      # Default: false
      check-blank: false

      # Disable built-in exclusion list (functions documented to never error)
      # Default: false
      disable-default-exclusions: false

      # List of functions to exclude from checking
      # Default: []
      exclude-functions:
        - fmt:.*
        - io.Copy(os.Stdout)

      # Display function signature (debugging)
      # Default: false
      verbose: false
```

### `check-type-assertions` Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: Report unchecked errors in type assertions

**When to Enable:**
- Type safety is critical
- Working with interface{}
- Public API code
- Security-sensitive operations

**Example:**
```go
// With check-type-assertions: true
val := getValue()
str := val.(string)  // errcheck: unchecked error

// Safe alternative
str, ok := val.(string)
if !ok {
    return fmt.Errorf("expected string")
}
```

### `check-blank` Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: Report errors assigned to blank identifier `_`

**When to Enable:**
- Strict error handling policy
- No intentional error discarding
- Security-critical code
- Public libraries

**Example:**
```go
// With check-blank: true
_, _ = os.OpenFile(path, os.O_RDONLY, 0644)  // errcheck reports

// Intentional case (use with care)
// Best-effort cleanup
defer func() {
    _, _ = os.Remove(tempFile)  // Add explanatory comment
}()
```

### `disable-default-exclusions` Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: Disable built-in exclusion list

**Built-in Exclusions Include:**
- fmt package print functions (fmt.Print, fmt.Sprint, etc.)
- Buffer operations (bytes.Buffer.Write, etc.)
- Logger methods

**When to Enable:**
- Maximum strictness required
- Want to verify all potential error sources
- Security-critical projects

### `exclude-functions` Option

- **Type**: `[]string`
- **Default**: `[]`
- **Description**: List of functions to exclude from checking

**Format:** `receiver` or `package.path` or `package.Type.Method`

**Common Exclusions:**
```yaml
exclude-functions:
  # Standard library functions
  - fmt:.*
  - io.Copy(os.Stdout)
  - io.Copy(os.Stderr)

  # Buffer operations (infallible)
  - (*bytes.Buffer).Write
  - (*bytes.Buffer).WriteString
  - (*strings.Builder).Write
  - (*strings.Builder).WriteString

  # Logger sync (best-effort)
  - (*logrus.Logger).Sync
  - (*go.uber.org/zap.Logger).Sync
  - (github.com/go-kit/log.Logger).Sync

  # HTTP response writes (handled by framework)
  - (*net/http.ResponseWriter).Write
  - (*net/http.ResponseWriter).WriteHeader

  # Context cancellation (intentional)
  - (*context.Context).Done

  # Project-specific
  - (*github.com/yourorg/yourpkg.Cleanup)
```

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Most production applications
version: "2"
linters:
  settings:
    errcheck:
      check-type-assertions: true
      check-blank: false
      disable-default-exclusions: false
      exclude-functions:
        - io.Copy(os.Stdout)
        - io.Copy(os.Stderr)
        - (*bytes.Buffer).Write
        - (*bytes.Buffer).WriteString

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errcheck]
```

#### ✅ Strict Configuration
```yaml
# Security-critical, public APIs, financial software
version: "2"
linters:
  settings:
    errcheck:
      check-type-assertions: true
      check-blank: true
      disable-default-exclusions: false
      exclude-functions:
        - (*bytes.Buffer).Write
```

#### ✅ Relaxed Configuration
```yaml
# Rapid prototyping, learning Go
version: "2"
linters:
  settings:
    errcheck:
      check-type-assertions: false
      check-blank: false
      disable-default-exclusions: false

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errcheck]
```

#### ✅ Web API Configuration
```yaml
# HTTP server with strict security
version: "2"
linters:
  settings:
    errcheck:
      check-type-assertions: true
      check-blank: false
      exclude-functions:
        # Logging to stdout is best-effort
        - io.Copy(os.Stdout)
        - io.Copy(os.Stderr)

        # Response writes handled by HTTP framework
        - (*net/http.ResponseWriter).Write
        - (*net/http.ResponseWriter).WriteHeader

        # Context cancellation is expected
        - (*context.Context).Done

        # Infallible operations
        - (*bytes.Buffer).Write
        - (*bytes.Buffer).WriteString

issues:
  exclude-rules:
    - path: (.+)_test\.go
    - path: generated/
      linters: [errcheck]
```

#### ✅ CLI Tool Configuration
```yaml
# Command-line interface tool
version: "2"
linters:
  settings:
    errcheck:
      check-type-assertions: false
      check-blank: false
      exclude-functions:
        # fmt print functions are expected
        - fmt:.*
        - log:.*

        # File cleanup in defer is best-effort
        - os.Remove
        - os.RemoveAll

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errcheck]
```

#### ✅ Library/Public SDK Configuration
```yaml
# Code consumed by others
version: "2"
linters:
  settings:
    errcheck:
      check-type-assertions: true
      check-blank: false
      disable-default-exclusions: false
      exclude-functions:
        - (*bytes.Buffer).Write

issues:
  # No exclusions for public code
  exclude-use-default: false
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

errcheck works excellently with:

| Linter | Relationship | Value |
|--------|--------------|---------|
| **errorlint** | Complementary | errcheck: "error ignored" → errorlint: "errors should be wrapped with %w" |
| **wrapcheck** | Complementary | errcheck: "error checked" → wrapcheck: "error not wrapped" |
| **staticcheck** | Complementary | errcheck: "error assigned" → staticcheck: "error not handled correctly" |
| **gosec** | Complementary | errcheck: "error ignored" → gosec: "potential security issue" |
| **nilerr** | Complementary | errcheck: "error not checked" → nilerr: "returned nil error with non-nil value" |
| **nilnil** | Complementary | errcheck: "error not checked" → nilnil: "simultaneous nil error and value" |
| **forcetypeassert** | Complementary | errcheck: "type assertion error ignored" → forcetypeassert: "must use comma-ok" |
| **govet** | Complementary | errcheck: "error handling" → govet: "suspicious constructs" |

**Complete Error Handling Suite:**
```yaml
linters:
  enable:
    - errcheck        # Unchecked errors (CRITICAL)
    - errorlint       # Error wrapping patterns (HIGH)
    - wrapcheck       # Error wrapping from external packages (HIGH)
    - nilerr          # Nil error returns (CRITICAL)
    - nilnil          # Simultaneous nil returns (MEDIUM)
    - forcetypeassert # Type assertion safety (HIGH)
    - staticcheck      # Deeper error analysis (CRITICAL)
    - gosec           # Security implications (CRITICAL)
```

### Example of Linter Synergy

```go
// errcheck catches:
func processData(data []byte) error {
    _, err := json.Unmarshal(data, &result)  // errcheck: ignored
    return nil
}

// errorlint adds (after errcheck fixed):
func processData(data []byte) error {
    data, err := json.Unmarshal(data, &result)
    if err != nil {
        return err  // errcheck satisfied
    }
    return nil  // errorlint: errors should be wrapped with context
}

// wrapcheck adds (after errorlint fixed):
func processData(data []byte) error {
    data, err := json.Unmarshal(data, &result)
    if err != nil {
        return fmt.Errorf("failed to unmarshal: %w", err)  // wrapcheck satisfied
    }
    return nil
}

// All linters satisfied:
func processData(data []byte) error {
    data, err := json.Unmarshal(data, &result)
    if err != nil {
        return fmt.Errorf("failed to unmarshal JSON data: %w", err)
    }
    return nil
}
```

### 🔒 Minimal Overlap, No Conflicts

| Linter | Overlap | Recommendation |
|---------|----------|----------------|
| errcheck + **errorlint** | Both check error handling | Use both - complementary |
| errcheck + **gosec** | Both catch error issues | Use both - gosec has security focus |
| errcheck + **staticcheck** | Both check errors | Use both - staticcheck is deeper analysis |
| errcheck + **nilerr** | Both catch error issues | Use both - different patterns |

**Why No Conflicts:**
- errcheck: Ensures errors are not ignored
- errorlint: Ensures errors are wrapped properly
- staticcheck: Analyzes if error handling is correct
- Each linter catches different aspects of error handling

### Configuration to Avoid Duplicate Reporting

```yaml
issues:
  exclude-rules:
    # Don't double-report from errcheck and gosec
    - linters: [errcheck, gosec]
      text: "G104"

    # Allow errcheck to take precedence for simple cases
    - linters: [staticcheck]
      text: "SA4001"
```

## Practical Examples

### ✅ Example 1: Database Operations

```go
// ❌ BAD: Unchecked database error
func getUser(id string) (*User, error) {
    row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
    var user User
    _ = row.Scan(&user.ID, &user.Name)  // errcheck: ignored
    return &user, nil
}

// ✅ GOOD: Proper error handling
func getUser(id string) (*User, error) {
    row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
    var user User
    err := row.Scan(&user.ID, &user.Name)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("user not found: %s", id)
        }
        return nil, fmt.Errorf("failed to scan user %s: %w", id, err)
    }
    return &user, nil
}
```

### ✅ Example 2: HTTP Handlers

```go
// ❌ BAD: Silent failure in HTTP handler
func getUserHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    user, _ := getUser(id)  // errcheck: ignored

    json.NewEncoder(w).Encode(user)  // Might panic if user is nil
}

// ✅ GOOD: Proper error handling
func getUserHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    user, err := getUser(id)
    if err != nil {
        http.Error(w, fmt.Sprintf("failed to get user: %v", err), http.StatusInternalServerError)
        return
    }

    if err := json.NewEncoder(w).Encode(user); err != nil {
        log.Printf("failed to encode user: %v", err)
    }
}
```

### ✅ Example 3: File Operations

```go
// ❌ BAD: Unchecked file error
func saveConfig(config *Config) error {
    data, err := json.Marshal(config)
    _ = os.WriteFile("config.json", data, 0644)  // errcheck: ignored
    return nil
}

// ✅ GOOD: Proper error handling
func saveConfig(config *Config) error {
    data, err := json.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }

    if err := os.WriteFile("config.json", data, 0644); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }
    return nil
}
```

### ✅ Example 4: Cryptographic Operations (CRITICAL)

```go
// ❌ BAD: Security vulnerability
func generateToken(user User) (string, error) {
    key, _ := hmac.New(hmac.NewSHA256, []byte("secret"))  // errcheck: ignored
    message := fmt.Sprintf("%s|%s|%d", user.ID, user.Email, time.Now().Unix())
    sig := hex.EncodeToString(key.Sum([]byte(message)))
    return sig, nil  // Token is weak/invalid if error
}

// ✅ GOOD: Secure error handling
func generateToken(user User) (string, error) {
    key, err := hmac.New(hmac.NewSHA256, []byte("secret"))
    if err != nil {
        return "", fmt.Errorf("failed to create HMAC key: %w", err)
    }

    message := fmt.Sprintf("%s|%s|%d", user.ID, user.Email, time.Now().Unix())
    sig := hex.EncodeToString(key.Sum([]byte(message)))
    return sig, nil
}
```

### ✅ Example 5: Background Operations

```go
// ❌ BAD: Error lost in goroutine
func processAsync(items []Item) {
    for _, item := range items {
        go func(i Item) {
            saveItem(i)  // errcheck: error ignored
        }(item)
    }
}

// ✅ GOOD: Error handled in goroutine
func processAsync(items []Item) {
    for _, item := range items {
        go func(i Item) {
            if err := saveItem(i); err != nil {
                log.Printf("Failed to save item %v: %v", i, err)
            }
        }(item)
    }
}
```

### ✅ Example 6: Error Wrapping Chain

```go
// ❌ BAD: No context in error
func loadData(id string) (*Data, error) {
    record, err := db.Get(id)
    if err != nil {
        return nil, err  // Lost context about what operation failed
    }

    data, err := process(record)
    if err != nil {
        return nil, err  // Lost context
    }

    return data, nil
}

// ✅ GOOD: Proper error wrapping chain
func loadData(id string) (*Data, error) {
    record, err := db.Get(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get record %s: %w", id, err)
    }

    data, err := process(record)
    if err != nil {
        return nil, fmt.Errorf("failed to process record %s: %w", id, err)
    }

    return data, nil
}
```

## Best Practices

1. **ALWAYS enable errcheck** for production code - it's CRITICAL priority
2. **Check errors immediately** after assignment, not later
3. **Wrap errors with context** - use `fmt.Errorf("...: %w", err)`
4. **Use blank identifier intentionally** - add comments explaining why
5. **Exclude test files** - tests have different error handling patterns
6. **Exclude generated code** - add to skip-dirs and exclude-rules
7. **Keep exclusion list minimal** - only exclude truly safe functions
8. **Document every exclusion** - explain why this function is safe
9. **Review exclusions quarterly** - remove when no longer needed
10. **Combine with errorlint** and **wrapcheck** for complete coverage
11. **Never ignore cryptographic errors** - major security risk
12. **Never ignore file permission errors** - leads to silent failures
13. **Never ignore network errors** - affects system reliability

## Common Scenarios and Solutions

### Scenario 1: Intentional Error Discarding

**Problem:** Some operations where failure is acceptable and you don't want to clutter code.

**Solution:** Use `_` with explanatory comment.

```go
// Best-effort cleanup
defer func() {
    // Don't fail if file already removed or permissions changed
    _ = os.Remove(tempFile)
}()

// Debug logging - don't fail if logging fails
_ = debugLog(debugInfo)  // Intentional, documented
```

### Scenario 2: Buffer Operations

**Problem:** errcheck flags in-memory buffer operations.

**Solution:** These are safe to exclude.

```yaml
linters-settings:
  errcheck:
    exclude-functions:
      - (*bytes.Buffer).Write
      - (*bytes.Buffer).WriteString
      - (*strings.Builder).Write
      - (*strings.Builder).WriteString
```

### Scenario 3: Test Files

**Problem:** All tests flagged with errcheck violations.

**Solution:** Exclude test files.

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errcheck]
```

### Scenario 4: Generated Code

**Problem:** Protobuf or mock-generated code has many violations.

**Solution:** Exclude generated directories.

```yaml
run:
  skip-dirs:
    - generated
    - vendor

issues:
  exclude-rules:
    - path: (.+)_pb\.go
    - path: (.+)_generated\.go
      linters: [errcheck]
```

### Scenario 5: Graceful Degradation

**Problem:** Operation should continue even if it fails.

**Solution:** Handle error explicitly but don't fail.

```go
// Attempt to write metrics, don't fail if it doesn't work
if err := metrics.Write(); err != nil {
    log.Printf("Failed to write metrics: %v", err)
    // Continue execution
}
```

### Scenario 6: Too Many False Positives

**Problem:** errcheck reporting many issues that are acceptable.

**Solution:** Add exclusions gradually.

```yaml
# Step 1: Start with defaults
linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: false

# Step 2: Add exclusions as needed
linters-settings:
  errcheck:
    exclude-functions:
      - io.Copy(os.Stdout)      # Logging
      - (*bytes.Buffer).Write      # Safe operation
```

### Scenario 7: Security-Critical Path

**Problem:** Must ensure NO errors are ignored in security code.

**Solution:** Use strictest configuration.

```yaml
linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
    disable-default-exclusions: false
    exclude-functions: []  # No exclusions
```

## Security-Specific Considerations

### Critical Security Errors

**ALWAYS check these errors:**

```go
// ❌ CRITICAL: Cryptographic operations
key, _ := rsa.GenerateKey(2048)  // Weak key
hash, _ := hmac.New(...)  // Invalid hash

// ✅ SECURE
key, err := rsa.GenerateKey(2048)
if err != nil {
    return nil, fmt.Errorf("cryptographic failure: %w", err)
}
```

```go
// ❌ CRITICAL: Random number generation
rand, _ := rand.Int(rand.Reader, max)  // Not random

// ✅ SECURE
rand, err := rand.Int(rand.Reader, max)
if err != nil {
    return 0, fmt.Errorf("random generation failed: %w", err)
}
```

```go
// ❌ CRITICAL: File permissions
f, _ := os.OpenFile(config, os.O_RDONLY, 0644)  // Wrong permissions or missing file

// ✅ SECURE
f, err := os.OpenFile(config, os.O_RDONLY, 0644)
if err != nil {
    return fmt.Errorf("failed to open config: %w", err)
}
```

### Security Checklist

Before disabling errcheck for security-sensitive code:

- [ ] Are cryptographic operations checking errors?
- [ ] Is random number generation checking errors?
- [ ] Are file permission operations checking errors?
- [ ] Are network connections checking errors?
- [ ] Is authentication/authorization checking errors?
- [ ] Is input validation checking errors?
- [ ] Are database operations checking errors?
- [ ] Are encryption/decryption operations checking errors?

## Summary

**errcheck** is a **CRITICAL** linter that prevents production bugs and security vulnerabilities:

- ✅ **Highest priority** - Most important Go linter
- ✅ **Prevents silent failures** - Catches #1 source of Go bugs
- ✅ **Security-critical** - Prevents major vulnerabilities
- ✅ **Low false positives** - Very accurate
- ✅ **Fast execution** - Minimal performance impact
- ✅ **Configurable** - Adjust strictness based on needs
- ✅ **Complementary** - Works excellently with errorlint, wrapcheck
- ⚠️ **May need exclusions** - For test files, generated code, safe functions
- ⚠️ **Requires manual review** - Cannot automatically fix issues

**Recommendation:** **ALWAYS ENABLE** with `check-type-assertions: true` for all production code. Exclude test files (`(.+)_test\.go`) and generated code. Exclude only documented safe functions (buffer operations, fmt print functions, etc.). Combine with **errorlint** and **wrapcheck** for complete error handling coverage. For security-critical code, use strictest settings (`check-blank: true`, minimal exclusions).

**Top 3 Configuration Tips:**
1. Enable `check-type-assertions: true` for type safety
2. Exclude only truly safe functions (buffers, fmt, logging)
3. Combine with errorlint and wrapcheck for complete error handling

---

**Reference:** https://github.com/kisielk/errcheck
