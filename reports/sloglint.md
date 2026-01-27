# sloglint Linter - Comprehensive Analysis

## What the Linter Does

**sloglint** ensures **consistent code style when using log/slog**, Go's standard structured logging package introduced in Go 1.21. It checks for proper slog usage patterns, attribute naming, and common mistakes. This is a **CRITICAL** linter for applications using Go's modern structured logging.

### The Problem It Detects

sloglint identifies inconsistent or incorrect slog usage:

- **Missing context** - Not using slog.With() for adding attributes
- **Invalid attribute keys** - Keys not in snake_case
- **Stringified keys** - Using string keys instead of proper slog attributes
- **Redundant calls** - Adding same attribute multiple times
- **Wrong log level** - Using incorrect severity
- **Mixed logging frameworks** - Combining slog with other loggers inconsistently

### Examples

```go
// ❌ BAD: Missing context, stringified keys
import "log/slog"

func processUser(id string) {
    slog.Info("Processing user", "user_id", id)  // sloglint: use slog.With(), "user_id" should be snake_case

    data, err := fetchData(id)
    slog.Info("User processed", "success", err == nil)  // sloglint: stringified keys
}
```

```go
// ✅ GOOD: Proper slog usage with context
import "log/slog"

func processUser(id string) {
    slog.Info("Processing user",
        "user_id", id,
    )

    data, err := fetchData(id)
    if err != nil {
        slog.Error("Failed to fetch user",
            "user_id", id,
            "error", err,
        )
        return
    }

    slog.Info("User processed",
        "user_id", id,
        "success", true,
    )
}
```

```go
// ❌ BAD: Wrong attribute key format
slog.Info("User login", "UserID", 123, "EmailAddress", "test@example.com")
// sloglint: "UserID" should be snake_case ("user_id")

// ✅ GOOD: Snake_case attribute keys
slog.Info("User login",
    "user_id", 123,
    "email_address", "test@example.com",
)
```

```go
// ❌ BAD: Redundant attributes
slog.Info("Processing",
    "user_id", id,
    "user_id", id,  // sloglint: duplicate attribute
)

// ✅ GOOD: Remove duplicates
slog.Info("Processing", "user_id", id)
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **All applications using slog** - Essential for consistent logging
- **New Go projects** - Go 1.21+ projects using slog
- **Microservices** - Structured logging critical for distributed systems
- **Web APIs** - HTTP request/response logging
- **CLI tools** - User-facing applications with logging
- **Enterprise applications** - Structured logging for observability

**Specific Scenarios:**

**1. Migrated Projects**
- Transitioning from logrus/zap to slog
- Need consistent slog usage patterns
- Enforcing best practices across codebase

**2. New Projects with slog**
- Go 1.21+ projects
- Using structured logging from start
- Establishing logging standards

**3. Team Projects**
- Multiple developers working on same codebase
- Consistent logging style across team
- Code review enforcement for slog usage

**4. Observability-Critical Systems**
- Applications with high logging volume
- Systems requiring log aggregation (ELK, Splunk, etc.)
- Production systems needing structured logs

**5. API Development**
- REST/GraphQL services
- Request/response logging
- Debug/trace logging

### ❌ Disable For:

**Specific Scenarios:**

**1. Non-slog Projects**
```yaml
# Using other logging frameworks (logrus, zap, etc.)
linters:
  enable:
    - other-linter
    # NOT sloglint
```

**2. Legacy Go Projects**
- Go < 1.21 (no slog)
- Using old logging frameworks

**3. Test Files with Mocks**
```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [sloglint]
```

**4. Generated Code**
```yaml
run:
  skip-dirs:
    - generated
```

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** for slog projects - Ensures consistency, prevents mistakes
- **Effort**: Low - Simple checks, minimal false positives
- **Recommendation**: **ALWAYS ENABLE** for any code using slog

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    sloglint:
      # Only log levels to check (comma-separated)
      # Default: all levels (DEBUG, INFO, WARN, ERROR)
      levels: ""

      # Attr names to require (comma-separated)
      # Default: ""
      attr-only: ""

      # Attr names to forbid (comma-separated)
      # Default: ""
      attr-blacklist: ""

      # Allow mixing slog with other frameworks
      # Default: false
      mixed-args: false

      # Enforce snake_case for attribute keys
      # Default: true
      key-naming-case: snake

      # Require all arguments be attributes (not stringified)
      # Default: true
      no-unknown: "attr"

      # Add missing key/value pairs
      # Default: false
      kv-only-mode: false

      # Context-only mode: require slog.With() for attributes
      # Default: false
      context: "scope"

      # Static mode: check all slog calls (no runtime checks)
      # Default: false
      static: false
```

### `levels` Option

- **Type**: `string`
- **Default**: `""` (all levels)
- **Description**: Comma-separated list of log levels to check

**Common Values:**
- `DEBUG,INFO,WARN,ERROR`
- `INFO,WARN,ERROR` (exclude debug)
- `ERROR` (only errors)

**Example:**
```yaml
linters:
  settings:
    sloglint:
      levels: "INFO,WARN,ERROR"  # Don't check debug logs
```

### `attr-only` Option

- **Type**: `string`
- **Default**: `""`
- **Description**: Only check specified attribute names

**Example:**
```yaml
linters:
  settings:
    sloglint:
      attr-only: "user_id,request_id,trace_id"  # Only check these
```

### `attr-blacklist` Option

- **Type**: `string`
- **Default**: `""`
- **Description**: Forbid these attribute names

**Example:**
```yaml
linters:
  settings:
    sloglint:
      attr-blacklist: "password,token,secret"  # Never log sensitive data
```

### `mixed-args` Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: Allow mixing slog with other logging frameworks

**When to Enable:**
- Transitioning from old logger to slog
- Using multiple logging frameworks temporarily

**Example:**
```yaml
linters:
  settings:
    sloglint:
      mixed-args: true
```

### `key-naming-case` Option

- **Type**: `string`
- **Default**: `snake`
- **Values**: `snake`, `camel`, `kebab`, `pascal`

**Example:**
```yaml
linters:
  settings:
    sloglint:
      key-naming-case: snake  # Default, recommended
```

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Most projects using slog
version: "2"
linters:
  settings:
    sloglint:
      key-naming-case: snake
      mixed-args: false

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [sloglint]
```

#### ✅ Strict Configuration
```yaml
# Production systems, high consistency requirements
version: "2"
linters:
  settings:
    sloglint:
      key-naming-case: snake
      context: "scope"  # Require slog.With() for attributes
      no-unknown: "attr"

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [sloglint]
```

#### ✅ Permissive Configuration
```yaml
# During migration or development
version: "2"
linters:
  settings:
    sloglint:
      key-naming-case: snake
      mixed-args: true  # Allow mixing with other loggers
      kv-only-mode: true  # Allow non-attribute args
```

#### ✅ Production-Only Configuration
```yaml
# Only check INFO/WARN/ERROR (skip debug)
version: "2"
linters:
  settings:
    sloglint:
      levels: "INFO,WARN,ERROR"
      attr-blacklist: "password,token,secret,api_key"
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

sloglint works excellently with:

| Linter | Relationship | Value |
|--------|--------------|---------|
| **loggercheck** | Complementary | loggercheck checks logger key/value pairs, sloglint: slog-specific |
| **govet** | Complementary | govet: general printf issues, sloglint: slog-specific |
| **staticcheck** | Complementary | staticcheck: general code issues, sloglint: slog usage |
| **zerologlint** | Alternative | Similar linter for zerolog framework (use one or other) |

**Complete Logging Suite:**
```yaml
linters:
  enable:
    - sloglint       # slog-specific (CRITICAL)
    - loggercheck    # Key/value pairs (CRITICAL)
    - govet          # General checks (CRITICAL)
    - staticcheck     # Deep analysis (CRITICAL)
```

### 🔒 No Conflicts

- **No known conflicts** - sloglint focuses specifically on slog usage
- **Independent operation** - Doesn't overlap with other linter functionality
- **Safe to enable with all linters**

## Practical Examples

### ✅ Example 1: Request/Response Logging

```go
// ❌ BAD: Stringified keys, missing context
import "log/slog"

func handleRequest(r *http.Request) *User {
    slog.Info("Handling request",
        "request_id", r.Header.Get("X-Request-ID"),  // sloglint: use slog.With()
        "user_agent", r.UserAgent(),
    )

    user := getUser(r)
    slog.Info("User retrieved",
        "request_id", r.Header.Get("X-Request-ID"),  // sloglint: duplicate
        "user_id", user.ID,
    )
    return user
}
```

```go
// ✅ GOOD: Proper slog usage
import "log/slog"

func handleRequest(r *http.Request) *User {
    ctx := slog.With(
        "request_id", r.Header.Get("X-Request-ID"),
    )

    slog.InfoContext(ctx, "Handling request",
        "user_agent", r.UserAgent(),
    )

    user := getUser(r)
    slog.InfoContext(ctx, "User retrieved",
        "user_id", user.ID,
    )
    return user
}
```

### ✅ Example 2: Error Logging

```go
// ❌ BAD: Inconsistent error logging
import "log/slog"

func processPayment(payment *Payment) error {
    if err := payment.Process(); err != nil {
        slog.Error("Payment failed", "error", err)  // sloglint: use slog.With()
        return err
    }
    slog.Info("Payment succeeded", "payment_id", payment.ID)  // sloglint: use slog.With()
    return nil
}
```

```go
// ✅ GOOD: Context-based error logging
import "log/slog"

func processPayment(payment *Payment) error {
    if err := payment.Process(); err != nil {
        slog.Error("Payment failed",
            slog.String("error", err.Error()),
            slog.Int64("amount", payment.Amount),
            slog.String("payment_id", payment.ID),
        )
        return err
    }
    slog.Info("Payment succeeded",
        slog.Int64("amount", payment.Amount),
        slog.String("payment_id", payment.ID),
    )
    return nil
}
```

### ✅ Example 3: Structured Logging with Attributes

```go
// ❌ BAD: Snake_case not enforced
import "log/slog"

func logAPICall(method, path, status, duration_ms int64) {
    slog.Info("API call",
        "Method", method,     // sloglint: should be snake_case
        "Path", path,         // sloglint: should be snake_case
        "Status", status,       // sloglint: should be snake_case
        "Duration", duration_ms,  // sloglint: should be snake_case
    )
}
```

```go
// ✅ GOOD: Proper snake_case attributes
import "log/slog"

func logAPICall(method, path, status, duration_ms int64) {
    slog.Info("API call",
        "method", method,
        "path", path,
        "status", status,
        "duration_ms", duration_ms,
    )
}
```

### ✅ Example 4: Debug vs Production Logging

```go
// ❌ BAD: Same style for debug and production
import "log/slog"

func processItem(item *Item) error {
    slog.Debug("Processing item", "item_id", item.ID)  // Stringified keys
    slog.Info("Item processed", "item_id", item.ID)    // Stringified keys
    return nil
}
```

```go
// ✅ GOOD: Strict for production, relaxed for debug
import "log/slog"

func processItem(item *Item) error {
    if debugMode {
        slog.Debug("Processing item",
            "item_id", fmt.Sprint(item.ID),  // Allowed for debug
        )
    }
    slog.Info("Item processed",
        "item_id", item.ID,  // Strict attribute
    )
    return nil
}
```

## Best Practices

1. **ALWAYS enable sloglint** for slog projects
2. **Use slog.With() for context** - Add attributes efficiently
3. **Use snake_case for keys** - Follow slog conventions
4. **Don't stringify keys** - Use proper slog attributes (slog.String, slog.Int, etc.)
5. **Avoid duplicate attributes** - Each attribute should appear once
6. **Exclude test files** - Test mocks often have intentional violations
7. **Use appropriate log levels** - Debug for development, Info/Warn/Error for production
8. **Never log sensitive data** - Use attr-blacklist to prevent logging passwords/tokens
9. **Be consistent** - Use same patterns across codebase
10. **Combine with loggercheck** - Complete logging quality coverage

## Common Scenarios and Solutions

### Scenario 1: Migration from Other Loggers

**Problem:** Transitioning to slog, many inconsistent patterns.

**Solution:** Enable sloglint to enforce consistency.
```yaml
linters:
  settings:
    sloglint:
      mixed-args: true  # Temporarily allow mixing during migration
```

### Scenario 2: Too Many False Positives in Tests

**Problem:** Test code flagged for intentional non-slog patterns.

**Solution:** Exclude test files.
```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [sloglint]
```

### Scenario 3: Sensitive Data Logging

**Problem:** Risk of logging passwords/tokens.

**Solution:** Use attr-blacklist.
```yaml
linters:
  settings:
    sloglint:
      attr-blacklist: "password,token,secret,api_key,auth_token,session_id"
```

### Scenario 4: Debug Logging in Production

**Problem:** Too much debug noise in production logs.

**Solution:** Only check INFO/WARN/ERROR levels.
```yaml
linters:
  settings:
    sloglint:
      levels: "INFO,WARN,ERROR"
```

## Summary

**sloglint** is a **CRITICAL** linter for Go applications using slog:

- ✅ **Highest priority** for slog projects - Ensures consistency
- ✅ **Prevents mistakes** - Catches common slog usage errors
- ✅ **Enforces standards** - snake_case keys, proper attribute usage
- ✅ **Simple to use** - Minimal configuration needed
- ✅ **Comprehensive** - Covers all slog usage patterns
- ✅ **Configurable** - Adjust strictness for different environments
- ⚠️ **slog-specific** - Only relevant for code using slog
- ⚠️ **May need exclusions** - Test files, debug logging

**Recommendation:** **ALWAYS ENABLE** with default settings for any code using slog (`log/slog`). Ensure **snake_case** for all attribute keys. Use **slog.With()** for adding context/attributes efficiently. Combine with **loggercheck** for complete logging quality coverage. Exclude test files (`(.+)_test\.go`). Use `attr-blacklist` to prevent logging sensitive data (passwords, tokens, etc.).

**Top 3 Configuration Tips:**
1. Enable `key-naming-case: snake` (default)
2. Use `slog.With()` for context/attributes
3. Combine with loggercheck for comprehensive logging quality

---

**Reference:** https://github.com/go-simpler/sloglint
