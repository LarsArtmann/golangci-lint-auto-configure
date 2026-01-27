# contextcheck Linter - Comprehensive Analysis

## What the Linter Does

**contextcheck** checks whether functions properly propagate `context.Context` parameters throughout the call chain. It detects when a function receives a context but fails to pass it (or a derived context) to subsequent function calls that also accept contexts.

### The Problem It Solves

Proper context propagation is critical for:
- **Request cancellation**: Allowing clients to cancel long-running operations
- **Deadlines/timeouts**: Enforcing timeouts across call chains
- **Request-scoped values**: Passing request IDs, authentication tokens, tracing spans
- **Resource cleanup**: Ensuring proper cleanup when requests are cancelled

When contexts aren't propagated, these features break, leading to:
- Goroutine leaks from uncancellable operations
- Resource exhaustion from orphaned requests
- Broken distributed tracing
- Missing request timeouts

### How It Works

The linter performs static analysis to:
1. Identify functions that accept `context.Context` as a parameter
2. Track calls from those functions to other functions
3. Verify that called functions which accept contexts receive one
4. Report missing context propagation

### Example Detection

```go
// ❌ BAD: Context not propagated
func processRequest(ctx context.Context, id string) error {
    // Missing context parameter - contextcheck will flag this
    data, err := fetchData(id)  // Should be: fetchData(ctx, id)
    if err != nil {
        return err
    }
    return processData(data)    // Should be: processData(ctx, data)
}

// ✅ GOOD: Context properly propagated
func processRequest(ctx context.Context, id string) error {
    data, err := fetchData(ctx, id)
    if err != nil {
        return err
    }
    return processData(ctx, data)
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **Web APIs and HTTP servers** - Critical for request cancellation
- **Microservices** - Distributed tracing and timeout propagation
- **gRPC services** - Built-in context handling
- **Event processors** - Proper event-scoped cancellation
- **API clients and SDKs** - Allow users to control timeouts
- **Concurrent data pipelines** - Clean cancellation across stages
- **Database repositories** - Transaction timeout enforcement

**Scenarios:**
- High-traffic services where request cancellation is essential
- Codebases with complex call chains
- Projects using distributed tracing (OpenTelemetry, Jaeger)
- Services making multiple downstream calls per request
- Applications requiring strict timeout enforcement
- Code reviews frequently miss context propagation

### Priority Assessment

- **Default Priority**: High
- **Essential for**: HTTP/gRPC services, microservices
- **Recommended for**: Any project using contexts extensively
- **Less critical for**: Simple CLI tools, pure computation libraries

## When It Should Be Disabled

### ❌ Consider Disabling For:

**Project Types:**
- Simple CLI tools with no network operations
- Pure data processing without I/O
- Scripts without concurrency
- Single-purpose utilities with no call chains

**Specific Scenarios:**
- Legacy codebases where full refactoring is not feasible
- Generated code (protobuf, mock generators, etc.)
- Test code (often has different context management)
- Proof-of-concept or prototype code

### Better Alternative to Disabling

Use **exclusions** for specific patterns:

```yaml
# .golangci.yml
linters:
  enable:
    - contextcheck

issues:
  exclude-rules:
    # Exclude test files
    - linters: [contextcheck]
      path: (.+)_test\.go

    # Exclude generated code
    - linters: [contextcheck]
      path: (.+)_generated\.go

    # Exclude specific patterns
    - linters: [contextcheck]
      path: internal/prototypes/
```

Or use targeted `nolint` directives:

```go
// Legitimate case: Intentionally creating new context
func backgroundTask(ctx context.Context) {
    // Create new context for background work
    newCtx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go doWork(newCtx) //nolint:contextcheck // New background context
}
```

## How It Should Be Configured

### Configuration Options

**contextcheck** has **no specific configuration options** in golangci-lint v2:

```yaml
# .golangci.yml
linters:
  enable:
    - contextcheck
```

### Recommended Configurations

**Standard Web Service:**
```yaml
version: "2"
linters:
  enable:
    - contextcheck
    - noctx
    - containedctx
    - bodyclose
    - errcheck
    - staticcheck
    - govet

issues:
  exclude-rules:
    # Exclude test files
    - linters: [contextcheck]
      path: (.+)_test\.go
```

**Microservice with Strict Enforcement:**
```yaml
version: "2"
linters:
  enable:
    # Context safety
    - contextcheck
    - noctx
    - containedctx
    - fatcontext
    
    # HTTP correctness
    - bodyclose
    - canonicalheader
    
    # Security & errors
    - gosec
    - errcheck
    - errorlint
    - wrapcheck
    
issues:
  # Maximum strictness
  exclude-use-default: false
  
  exclude-rules:
    - linters: [contextcheck]
      path: (.+)_test\.go
```

**API Client Library:**
```yaml
version: "2"
linters:
  enable:
    - contextcheck
    - wrapcheck
    - errorlint
    - errcheck
    - staticcheck

issues:
  # Library code should be strict
  exclude-use-default: false
  
  # Only exclude test helpers
  exclude-rules:
    - linters: [contextcheck]
      path: internal/testhelpers/
```

### Best Practices

1. **Enable with related linters** - Works best with `noctx`, `containedctx`, `fatcontext`
2. **Exclude test files** - Tests often have legitimate reasons to handle contexts differently
3. **Use nolint with explanations** - Document why context isn't being propagated
4. **Refactor incrementally** - Add exclusions for legacy code during gradual migration
5. **Educate the team** - Explain context propagation patterns during code review

## How It Interferes or Works Together With Other Linters

### ✅ Synergistic Linters

**Essential Combinations:**
- **`noctx`**: Ensures HTTP requests have context - both validate context usage
- **`containedctx`**: Prevents context in structs - comprehensive context safety
- **`fatcontext`**: Detects nested contexts in loops - related context anti-patterns
- **`bodyclose`**: HTTP response cleanup - independent but related to request handling

**Complete HTTP Safety Suite:**
```yaml
linters:
  enable:
    # Context propagation
    - contextcheck
    - noctx
    - containedctx
    - fatcontext
    
    # HTTP correctness
    - bodyclose
    - canonicalheader
    
    # Error handling
    - errcheck
    - errorlint
    - wrapcheck
    
    # Security
    - gosec
```

**Example Synergy:**
```go
// contextcheck verifies we pass ctx to fetchData
// noctx verifies we don't use http.Get without context
// bodyclose verifies we close the response body
func (c *Client) GetResource(ctx context.Context, id string) (*Resource, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/resource/"+id, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()  // bodyclose validates
    
    return parseResponse(ctx, resp)  // contextcheck validates
}
```

### 🔒 Potential Conflicts

**No Known Conflicts:**
- contextcheck operates independently on call chain analysis
- No overlapping functionality with other linters
- Safe to enable with all other linters

### 📊 Performance Impact

- **Low Overhead**: Single-pass static analysis
- **Fast Execution**: No complex interprocedural analysis
- **Scales Well**: Linear with function call graph size
- **Recommended in CI/CD**: No performance concerns

## Practical Examples

### ✅ Correct - Context Propagation Chain

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Context from HTTP request
    ctx := r.Context()
    
    if err := processRequest(ctx, r.URL.Query().Get("id")); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}

func processRequest(ctx context.Context, id string) error {
    // Propagate to validate
    if err := validateID(ctx, id); err != nil {  // contextcheck validates this
        return fmt.Errorf("validation: %w", err)
    }
    
    // Propagate to fetch
    data, err := fetchData(ctx, id)  // contextcheck validates this
    if err != nil {
        return fmt.Errorf("fetch data: %w", err)
    }
    
    // Propagate to process
    return processData(ctx, data)  // contextcheck validates this
}

func validateID(ctx context.Context, id string) error {
    // Use context for timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
    defer cancel()
    
    // Simulate validation logic
    return validateInDatabase(timeoutCtx, id)
}

func fetchData(ctx context.Context, id string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/"+id, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return io.ReadAll(resp.Body)
}

func processData(ctx context.Context, data []byte) error {
    // Check if context cancelled
    if err := ctx.Err(); err != nil {
        return err
    }
    
    // Process data...
    return nil
}
```

### ❌ Incorrect - Missing Context Propagation

```go
func processRequest(ctx context.Context, id string) error {
    // ❌ contextcheck would flag these:
    
    data, err := fetchData(id)  // Missing ctx parameter
    if err != nil {
        return err
    }
    
    return processData(data)    // Missing ctx parameter
}

func fetchData(id string) ([]byte, error) {
    // This function should accept context
    req, err := http.NewRequest("GET", "https://api.example.com/"+id, nil)  // noctx would also flag
    if err != nil {
        return nil, err
    }
    
    // No context timeout, cancellation, or tracing
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return io.ReadAll(resp.Body)
}

func processData(data []byte) error {
    // Can't check for cancellation
    // Can't access request-scoped values
    
    // Process data...
    return nil
}
```

### 📝 Handling Exceptions

**Creating New Contexts (Legitimate):**
```go
func processRequest(ctx context.Context, id string) error {
    // Create timeout context - this is legitimate
    timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Use timeout context for downstream call
    data, err := fetchData(timeoutCtx, id)  // ✅ contextcheck accepts this
    if err != nil {
        return err
    }
    
    return processData(ctx, data)  // ✅ Original context passed
}
```

**Background Operations:**
```go
func processRequest(ctx context.Context, id string) error {
    // Start background work - new context is OK
    bgCtx := context.WithoutCancel(ctx)  // Detach for background
    
    go func() {
        // Background processing shouldn't be cancelled by request
        saveToStorage(bgCtx, id)  //nolint:contextcheck // Background context
    }()
    
    // Continue with request context
    return fetchData(ctx, id)
}
```

**Test-Specific Patterns:**
```go
func TestProcessRequest(t *testing.T) {
    // Test might want to control context explicitly
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Test helper might use different context
    data := setupTestData(t)  //nolint:contextcheck // Test setup uses different ctx
    
    err := processRequest(ctx, "test-id")
    assert.NoError(t, err)
}
```

## Common Patterns and Solutions

### Pattern 1: Chain of Responsibility

```go
// Each handler receives and propagates context
func (h *Handler) Handle(ctx context.Context, req Request) (Response, error) {
    // Validate
    if err := h.validator.Validate(ctx, req); err != nil {
        return Response{}, err
    }
    
    // Process
    result, err := h.processor.Process(ctx, req.Data)
    if err != nil {
        return Response{}, err
    }
    
    // Persist
    if err := h.repo.Save(ctx, result); err != nil {
        return Response{}, err
    }
    
    return Response{Result: result}, nil
}
```

### Pattern 2: Database Repository

```go
type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    row := r.db.QueryRowContext(ctx, query, id)
    
    var user User
    if err := row.Scan(&user.ID, &user.Name, &user.Email); err != nil {
        return nil, fmt.Errorf("query user: %w", err)
    }
    return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    query := "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)"
    
    // Context for creating
    if _, err := r.db.ExecContext(ctx, query, user.ID, user.Name, user.Email); err != nil {
        return fmt.Errorf("create user: %w", err)
    }
    return nil
}
```

### Pattern 3: HTTP Client Wrapper

```go
type APIClient struct {
    baseURL    string
    httpClient *http.Client
}

func (c *APIClient) GetResource(ctx context.Context, id string) (*Resource, error) {
    url := c.baseURL + "/resources/" + id
    
    // Use context for request
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    
    // Request inherits cancellation, timeout, and values from context
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("execute request: %w", err)
    }
    defer resp.Body.Close()
    
    var resource Resource
    if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    
    return &resource, nil
}
```

## Summary

**contextcheck** is a **critical linter** for modern Go applications. It ensures:
- ✅ Proper context cancellation propagation
- ✅ Timeout enforcement across call chains
- ✅ Request-scoped value availability
- ✅ Resource cleanup on cancellation
- ✅ Distributed tracing support

**Recommendation**: **ENABLE** for all Go projects that use contexts, especially web services, APIs, and microservices. Works best when combined with `noctx`, `containedctx`, and `fatcontext` for comprehensive context safety. Use exclusions for test files and generated code.

---

**Reference**: https://github.com/kkHAIKE/contextcheck
