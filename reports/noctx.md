# noctx Linter - Comprehensive Analysis

## What the Linter Does

**noctx** detects functions that use a **non-inherited context** (context.Context) in Go. It identifies places where `context.Background()` or `context.TODO()` are used directly instead of accepting a context parameter from the caller. This is a **CRITICAL** linter for proper Go context propagation, especially in HTTP servers and microservices.

### The Problem It Detects

Go contexts are essential for:

- **Request cancellation** - Propagating cancel signals
- **Deadlines** - Enforcing timeouts
- **Traceability** - Distributed tracing and observability
- **Resource cleanup** - Ensuring goroutines exit cleanly
- **Authentication** - User context propagation

Using `context.Background()` or `context.TODO()` instead of accepting context breaks:

- **Request cancellation** - User cancels request, but operation continues
- **No timeout enforcement** - Operations run forever
- **Lost tracing** - Can't trace request through system
- **Resource leaks** - Goroutines don't receive cancel signal
- **Poor user experience** - Can't cancel long-running operations

### How It Works

noctx analyzes function definitions and calls:

1. **Function signature analysis** - Checks if function accepts context.Context parameter
2. **Context creation analysis** - Detects context.Background() or context.TODO() calls
3. **Call graph traversal** - Identifies functions that should accept context but don't
4. **Exclusion patterns** - Ignores main() functions, tests, etc.

### Examples

```go
// ❌ BAD: Function creates own context
func fetchUser(id int) (*User, error) {
    ctx := context.Background()  // noctx: function should accept context
    return db.QueryUser(ctx, id)
}
```

```go
// ✅ GOOD: Function accepts context
func fetchUser(ctx context.Context, id int) (*User, error) {
    return db.QueryUser(ctx, id)
}
```

```go
// ❌ BAD: HTTP handler creates context
func handleUser(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()  // noctx: should use r.Context()
    user, err := fetchUser(ctx, id)
    // ...
}
```

```go
// ✅ GOOD: HTTP handler uses request context
func handleUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // Propagates request cancellation
    user, err := fetchUser(ctx, id)
    // ...
}
```

```go
// ❌ BAD: Background goroutine with background context
func processData(data string) {
    go func() {
        ctx := context.Background()  // noctx: can't be cancelled
        processWithTimeout(ctx, data)
    }()
}
```

```go
// ✅ GOOD: Goroutine accepts context
func processData(ctx context.Context, data string) {
    go func(ctx context.Context) {
        processWithTimeout(ctx, data)
    }(ctx)
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

- **Web servers and APIs** - HTTP handlers must use request context
- **Microservices** - Inter-service calls need context propagation
- **Long-running operations** - Background jobs, batch processing
- **Database operations** - Queries should use context for cancellation
- **Network calls** - HTTP/gRPC calls should use context
- **File operations** - Should respect cancellation signals
- **CLI tools** - Context for global cancellation signals
- **Production systems** - Request timeouts and cancellation critical

**Specific Scenarios:**

**1. HTTP/GRPC Servers**

- All HTTP handlers
- gRPC service methods
- Middleware functions
- Request/response processing

**2. Database Operations**

- All queries and transactions
- Connection pool operations
- Migration scripts

**3. Microservice Communication**

- HTTP client calls
- gRPC client calls
- Message queue operations
- Event streaming

**4. Background Processing**

- Worker pools
- Scheduled jobs
- Async processing
- Batch operations

**5. API Clients**

- SDKs for external services
- HTTP wrapper functions
- gRPC client methods
- Database access layers

**6. CLI Tools with Signal Handling**

- Interactive CLI tools
- Long-running commands
- Batch processors
- Services with graceful shutdown

### ❌ Disable For:

**Specific Scenarios:**

**1. main() Functions**

```yaml
# Main entry point doesn't accept context
linters:
  enable:
    - noctx
linters-settings:
  noctx:
    # Don't report main() functions
    # Default: true
    check-main: false
```

**2. Test Files**

```yaml
# Tests intentionally use context.Background()
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [noctx]
```

**3. Generated Code**

```yaml
run:
  skip-dirs:
    - generated

issues:
  exclude-rules:
    - path: (.+)_pb\.go
      linters: [noctx]
```

**4. Initialization Functions**

```yaml
# Some initialization functions legitimately create context
linters-settings:
  noctx:
    # Allow context.Background() in init functions
    # Default: []
    allow-init: true
```

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** - Ensures context propagation, enables request cancellation
- **Effort**: Low - Simple check, clear fix pattern
- **Recommendation**: **ALWAYS ENABLE** for HTTP servers, microservices, production code

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    noctx:
      # Check main() functions
      # Default: true
      check-main: true

      # Allow context.Background() in init functions
      # Default: []
      allow-init: []

      # Allow context.Background() in specific functions
      # Default: []
      allow-functions:
        - main\..*
        - (.*Test)\..*

      # Don't report context.Background() calls
      # Default: false
      disable-exported: false
```

### `check-main` Option

- **Type**: `bool`
- **Default**: `true`
- **Description**: Check if main() functions use non-inherited context

**When to Disable:**

- When main() legitimately creates initial context for background processes
- When checking main() causes too many false positives

**Example:**

```go
// With check-main: true, this would be flagged
func main() {
    ctx := context.Background()  // noctx: flagged
    process(ctx)
}

// With check-main: false, this is OK
func main() {
    ctx := context.Background()  // noctx: allowed
    process(ctx)
}
```

### `allow-init` Option

- **Type**: `[]string`
- **Default**: `[]`
- **Description**: Function names where context.Background() is allowed

**Use Cases:**

- Initialization functions
- Configuration loading functions
- Background worker setup

**Example:**

```yaml
allow-init:
  - init\..*
  - setup\..*
  - initialize\..*
```

### `allow-functions` Option

- **Type**: `[]string`
- **Default**: `[]`
- **Description**: Function name patterns allowed to use context.Background()

**Format:** Regular expression patterns

**Examples:**

```yaml
allow-functions:
  - main\..* # Allow in main functions
  - (.*Test)\..* # Allow in test functions
  - (.*Example)\..* # Allow in example functions
  - NewClient # Allow in client constructors
  - Connect # Allow in connection functions
```

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# Most production applications
version: "2"
linters:
  settings:
    noctx:
      check-main: true
      allow-init: []

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [noctx]
```

#### ✅ Web Server Configuration

```yaml
# HTTP/GRPC servers
version: "2"
linters:
  settings:
    noctx:
      check-main: true
      allow-init: []
      allow-functions:
        - main\..*
        - Run # Allow in Run() methods
        - Start # Allow in Start() methods
```

#### ✅ Microservice Configuration

```yaml
# Service-to-service communication
version: "2"
linters:
  settings:
    noctx:
      check-main: true
      allow-init: []
      allow-functions:
        - NewClient
        - Connect
        - Dial
```

#### ✅ Relaxed Configuration

```yaml
# CLI tools, simple scripts
version: "2"
linters:
  settings:
    noctx:
      check-main: false # Don't check main()
      allow-init:
        - main
```

#### ✅ Strict Configuration

```yaml
# Production-critical, no exceptions
version: "2"
linters:
  settings:
    noctx:
      check-main: true
      allow-init: []
      allow-functions: []
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

noctx works excellently with:

| Linter           | Relationship  | Value                                                       |
| ---------------- | ------------- | ----------------------------------------------------------- |
| **govet**        | Complementary | govet: context issues, noctx: context propagation           |
| **errcheck**     | Complementary | errcheck: error handling, noctx: context propagation        |
| **staticcheck**  | Complementary | staticcheck: SA2002 context, noctx: function-level          |
| **contextcheck** | Complementary | contextcheck: context propagation, noctx: context creation  |
| **gosec**        | Complementary | gosec: G107 SSRF, noctx: context for security               |
| **fatcontext**   | Complementary | fatcontext: nested contexts, noctx: missing context         |
| **spancheck**    | Complementary | spancheck: OpenTelemetry/Census, noctx: context propagation |

**Complete Context Suite:**

```yaml
linters:
  enable:
    - noctx # Non-inherited context (CRITICAL)
    - contextcheck # Context propagation (HIGH)
    - fatcontext # Nested contexts (HIGH)
    - spancheck # OpenTelemetry spans (HIGH)
    - errcheck # Error handling (CRITICAL)
    - gosec # Security (CRITICAL)
    - staticcheck # Deep analysis (CRITICAL)
```

### Example of Linter Synergy

```go
// noctx catches:
func processRequest(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()  // noctx: should use r.Context()
    result, err := process(ctx)
    json.NewEncoder(w).Encode(result)
}

// contextcheck would also catch:
func processRequest(w http.ResponseWriter, r *http.Request) {
    result, err := process(r.Context())  // contextcheck: context not propagated to process()
    json.NewEncoder(w).Encode(result)
}

// spancheck would catch:
func processRequest(w http.ResponseWriter, r *http.Request) {
    result, err := process(r.Context())
    // spancheck: operation uses context but no span created
    json.NewEncoder(w).Encode(result)
}
```

### 🔒 No Conflicts

- **No known conflicts** - noctx focuses specifically on context propagation
- **Independent operation** - Doesn't overlap with other linter functionality
- **Safe to enable with all linters**

## Practical Examples

### ✅ Example 1: HTTP Handler

```go
// ❌ BAD: Handler creates own context
func handleUser(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()  // noctx: can't cancel this request
    user, err := fetchUser(ctx, id)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

```go
// ✅ GOOD: Handler uses request context
func handleUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // Propagates request cancellation
    user, err := fetchUser(ctx, id)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(user)
}

// When user cancels request, fetchUser is cancelled
```

### ✅ Example 2: gRPC Service

```go
// ❌ BAD: Service method creates own context
type Server struct{}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    dbCtx := context.Background()  // noctx: ignores request context
    user, err := s.db.QueryUser(dbCtx, req.UserId)
    if err != nil {
        return nil, err
    }
    return &pb.GetUserResponse{User: user}, nil
}
```

```go
// ✅ GOOD: Service method uses request context
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // Use request context directly
    user, err := s.db.QueryUser(ctx, req.UserId)
    if err != nil {
        return nil, err
    }
    return &pb.GetUserResponse{User: user}, nil
}

// Request cancellation propagated to database
```

### ✅ Example 3: Database Operation

```go
// ❌ BAD: Query function creates own context
func QueryUser(db *sql.DB, id int) (*User, error) {
    ctx := context.Background()  // noctx: can't cancel query
    return db.QueryUser(ctx, id)
}
```

```go
// ✅ GOOD: Query function accepts context
func QueryUser(ctx context.Context, db *sql.DB, id int) (*User, error) {
    return db.QueryUser(ctx, id)
}

// Caller controls timeout:
func handleUser(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    user, err := QueryUser(ctx, id)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

### ✅ Example 4: HTTP Client

```go
// ❌ BAD: Client creates own context
func CallAPI(url string) ([]byte, error) {
    ctx := context.Background()  // noctx: can't cancel this request
    req, _ := http.NewRequest("GET", url, nil)
    req = req.WithContext(ctx)
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

```go
// ✅ GOOD: Client function accepts context
func CallAPI(ctx context.Context, url string) ([]byte, error) {
    req, _ := http.NewRequest("GET", url, nil)
    req = req.WithContext(ctx)  // Use caller's context
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return io.ReadAll(resp.Body), nil
}

// Caller controls timeout:
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    data, err := CallAPI(ctx, "https://api.example.com/data")
    // ...
}
```

### ✅ Example 5: Background Worker

```go
// ❌ BAD: Worker creates background context
func StartWorker() {
    go func() {
        ctx := context.Background()  // noctx: worker can't be shut down
        for {
            processJob(ctx)
        }
    }()
}
```

```go
// ✅ GOOD: Worker accepts context
func StartWorker(ctx context.Context) {
    go func(ctx context.Context) {
        for {
            select {
            case <-ctx.Done():
                return  // Clean shutdown
            default:
                processJob(ctx)
            }
        }
    }(ctx)
}

// Caller controls worker lifecycle:
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()  // Cancel on exit

    StartWorker(ctx)

    // Wait for shutdown signal
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt, os.Kill)
    <-sig
}
```

## Best Practices

1. **ALWAYS enable noctx** for HTTP servers, microservices, production code
2. **Accept context.Context as first parameter** - After receiver in methods
3. **Use r.Context() in HTTP handlers** - Propagates request cancellation
4. **Pass context to all database operations** - Enables query cancellation
5. **Pass context to all HTTP/gRPC calls** - Enables request timeout
6. **Pass context to background goroutines** - Enables goroutine cancellation
7. **Use context.WithTimeout() for deadlines** - Enforces operation timeouts
8. **Use context.WithCancel() for manual cancellation** - Enables graceful shutdown
9. **Don't create context.Background() in handlers** - Breaks request cancellation
10. **Combine with contextcheck** - Complete context propagation coverage

## Common Scenarios and Solutions

### Scenario 1: Database Layer Without Context

**Problem:** Database functions create own contexts.

**Solution:** Add context.Context parameter to all database functions.

```go
// ❌ Wrong pattern
func QueryUser(db *sql.DB, id int) (*User, error) {
    ctx := context.Background()
    return db.QueryContext(ctx, id)
}

// ✅ Correct pattern
func QueryUser(ctx context.Context, db *sql.DB, id int) (*User, error) {
    return db.QueryContext(ctx, id)
}
```

### Scenario 2: HTTP Handler Wrapper

**Problem:** Middleware or wrapper functions use context.Background().

**Solution:** Extract context from http.Request.

```go
// ❌ Wrong pattern
func withTimeout(fn func()) error {
    ctx := context.Background()
    return fn(ctx)
}

// ✅ Correct pattern
func withTimeout(fn func(context.Context) error) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    return fn(ctx)
}
```

### Scenario 3: gRPC Stub Implementation

**Problem:** gRPC stub methods don't accept context.

**Solution:** Accept context and pass to implementations.

```go
// ❌ Wrong pattern
func (s *server) GetUser(req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    ctx := context.Background()
    return s.impl.GetUser(ctx, req)
}

// ✅ Correct pattern
func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    return s.impl.GetUser(ctx, req)
}
```

### Scenario 4: Background Job Processing

**Problem:** Jobs use context.Background(), can't be cancelled.

**Solution:** Pass context to job workers.

```go
// ❌ Wrong pattern
func ProcessJobs(jobs []Job) {
    for _, job := range jobs {
        go func(j Job) {
            ctx := context.Background()
            j.Process(ctx)
        }(job)
    }
}

// ✅ Correct pattern
func ProcessJobs(ctx context.Context, jobs []Job) {
    for _, job := range jobs {
        go func(ctx context.Context, j Job) {
            j.Process(ctx)
        }(ctx, job)
    }
}
```

### Scenario 5: Test Files with Intentional Background

**Problem:** Tests use context.Background() intentionally.

**Solution:** Exclude test files.

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [noctx]
```

## Summary

**noctx** is a **CRITICAL** linter for proper Go context propagation:

- ✅ **Highest priority** - Ensures request cancellation, timeouts work
- ✅ **Prevents critical issues** - Operations can't be cancelled, poor UX
- ✅ **Enforces best practice** - Context propagation is Go pattern
- ✅ **Simple check** - Clear fix pattern: accept context parameter
- ✅ **Comprehensive** - Covers all function and goroutine patterns
- ✅ **Essential for microservices** - Proper distributed tracing and cancellation
- ✅ **Configurable** - Can exclude main(), init(), specific functions
- ✅ **Complementary** - Works with contextcheck, fatcontext, spancheck
- ⚠️ **May need exclusions** - main() functions, test files, init functions
- ⚠️ **Go-specific** - Only relevant for Go context.Context

**Recommendation:** **ALWAYS ENABLE** for HTTP servers, gRPC services, microservices, and any production code. Ensure **all functions** that perform I/O, make network calls, or run for extended time **accept context.Context as first parameter**. Use `r.Context()` in HTTP handlers. Pass context to database queries, HTTP clients, and background goroutines. Use `context.WithTimeout()` and `context.WithCancel()` for deadline and cancellation control. Combine with **contextcheck**, **fatcontext**, and **spancheck** for complete context propagation coverage.

**Top 3 Guidelines:**

1. Accept context.Context as first parameter in all I/O functions
2. Use r.Context() in HTTP handlers, not context.Background()
3. Pass context to all goroutines and long-running operations

---

**Reference:** https://github.com/sonatard/noctx
