# bodyclose Linter - Comprehensive Analysis

## What the Linter Does

**bodyclose** detects unclosed HTTP response bodies in Go code, preventing resource leaks that can exhaust connection pools and cause production outages.

### Detailed Explanation

The linter statically analyzes Go programs to ensure that `http.Response.Body` is properly closed after HTTP requests. According to Go's HTTP documentation, the response body **must** be closed to release the underlying TCP connection back to the connection pool. Failure to close response bodies causes connection pool exhaustion, leading to new requests hanging indefinitely.

**Key detection areas:**
- HTTP client operations (`http.Get`, `http.Post`, `client.Do`, etc.)
- Response body handling in all code paths
- Proper `defer` statements for cleanup
- Error handling scenarios where response might be `nil`
- Conditional branches that might skip body closure

**Example of the problem it solves:**
```go
// BAD: Response body is never closed
func fetchData(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    return ioutil.ReadAll(resp.Body) // ❌ Body never closed
}

// GOOD: Response body is properly closed
func fetchData(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close() // ✅ Body closed after reading
    return ioutil.ReadAll(resp.Body)
}
```

**Advanced detection capabilities:**
The linter tracks response body references through:
- Function returns (when response is passed up the call stack)
- Struct fields (when response is stored in objects)
- Conditional flows (ensures closure in all branches)
- Error paths (ensures closure even when errors occur)

## When It Should Be Enabled

### Always Enable For:
- **Any project making HTTP requests** (clients, microservices, APIs, CLI tools)
- **High-throughput services** where connection pool exhaustion is critical
- **Cloud-native applications** with inter-service communication
- **Web scrapers, crawlers, or automation tools**
- **Projects using REST APIs, GraphQL, or HTTP-based SDKs**
- **Infrastructure tooling** that manages remote resources

### Specific Use Cases:
1. **Microservices architectures**: Prevents cascading failures from connection starvation
2. **API clients and SDKs**: Ensures proper resource management for library users
3. **Web applications**: Protects against backend service connection leaks
4. **Data pipeline workers**: Critical when making thousands of HTTP requests
5. **CLI tools with network operations**: Prevents user complaints about hanging commands
6. **Testing frameworks**: Ensures test HTTP servers don't leak connections
7. **Proxy/gateway services**: High connection volume makes this essential

### Project Types:
- **Web applications**: Backend services, APIs, middleware
- **Cloud services**: AWS/Azure/GCP clients, Kubernetes operators
- **DevOps tools**: CI/CD integrations, monitoring agents
- **Data engineering**: ETL processes, API consumers
- **Mobile backends**: Server-side request handling
- **Network utilities**: Load balancers, proxies, gateways

## When It Should Be Disabled

### Appropriate to Disable For:
- **Projects with zero HTTP usage** (pure computation, algorithms, data structures)
- **Embedded systems** without network stacks
- **Command-line tools** that only operate on local files
- **Cryptography/math libraries** with no network dependencies

### Specific Scenarios:
1. **Known false positive patterns**: When using HTTP wrapper libraries that handle cleanup internally (see issue #30)
2. **Custom HTTP client abstractions**: When your codebase uses a wrapper that guarantees body closure
3. **Legacy codebases**: When retrofitting would be prohibitively expensive AND you have runtime monitoring
4. **Short-lived processes**: When the process exits after single requests (e.g., Cron jobs), though still recommended
5. **Generated code**: When generated clients have known issues and manual review is impractical

**Critical Warning**: Even in these cases, consider using `//nolint:bodyclose` comments for specific exceptions rather than disabling globally.

## How It Should Be Configured

### Configuration Options

**bodyclose has no configuration options** - it uses static analysis with predetermined rules.

### Basic enablement:
```yaml
linters:
  enable:
    - bodyclose
```

### Complete example configurations:

**Standard Web Project:**
```yaml
version: "2"

linters:
  default: none
  enable:
    - bodyclose
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - gosec
    - noctx  # Complements bodyclose by ensuring request contexts

issues:
  # Exclude test files where response mocking is common
  exclude-rules:
    - path: _test\.go
      linters:
        - bodyclose
      text: "response body must be closed"

  # Don't fail CI on existing issues, but prevent new ones
  new-from-rev: HEAD~1
```

**API Client Library:**
```yaml
version: "2"

linters:
  enable:
    - bodyclose
    - errcheck
    - staticcheck
    - govet

  settings:
    errcheck:
      # Ensure Close errors are checked in library code
      check-type-assertions: true

issues:
  # Strict enforcement - no exceptions in library code
  exclude-use-default: false

  # Only exclude truly unavoidable cases with inline comments
  exclude-rules:
    - path: internal/testhelpers/
      linters: [bodyclose]
```

**Microservices with High Concurrency:**
```yaml
version: "2"

linters:
  default: none
  enable:
    - bodyclose
    - noctx      # Critical for cancellation
    - errcheck   # Check Close() errors
    - govet
    - staticcheck
    - gosimple
    - ineffassign

issues:
  # Maximum strictness for connection-sensitive code
  max-issues-per-linter: 0
  max-same-issues: 0

  # Exclude only vendored/generated code
  exclude-rules:
    - path: vendor/
      linters: [bodyclose]
```

### Best Practices

1. **Always use `defer resp.Body.Close()` immediately after checking errors**
2. **Check response for nil before deferring**: `if resp != nil && resp.Body != nil`
3. **Combine with `noctx` linter** for comprehensive HTTP request safety
4. **Use in pre-commit hooks** to catch issues before commit
5. **Pair with connection pool monitoring** in production as a safety net
6. **Document HTTP client patterns** in your project's style guide

## How It Interferes or Works Together With Other Linters

### Synergistic Relationships

**Works Excellently With:**
- **`noctx`**: Ensures HTTP requests use contexts for cancellation AND bodies are closed
- **`errcheck`**: Verifies that `resp.Body.Close()` errors are handled (critical for production)
- **`gosec`**: Catches security issues; bodyclose prevents DoS via connection exhaustion
- **`staticcheck`**: Comprehensive bug detection including resource leaks
- **`gosimple`**: Simplifies code patterns; bodyclose ensures they're correct

**Example Synergy:**
```go
func fetchWithContext(ctx context.Context, url string) ([]byte, error) {
    // noctx would flag this if context wasn't used
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }

    // bodyclose ensures this exists
    defer resp.Body.Close()

    // errcheck ensures we handle ReadAll error
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("reading response: %w", err)
    }

    // gosec might flag insecure URLs, staticcheck finds other issues
    return body, nil
}
```

### Potential Conflicts

**Known Issues (GitHub Issue #30):**
The linter has reported false positives when:
- HTTP wrapper libraries handle body closure internally
- Response bodies are passed to functions that guarantee closure
- Mock/testing frameworks simulate responses

**Example false positive pattern:**
```go
// Might be flagged incorrectly if wrapper closes internally
func fetch(url string) ([]byte, error) {
    return myHTTPWrapper.Get(url) // Body closed inside wrapper
}
```

**Mitigation strategies:**
1. Use `//nolint:bodyclose` with explanatory comment
2. Configure exclude rules for wrapper libraries
3. Wrap with helper that satisfies the linter

### Integration Patterns

**CI/CD Pipeline Integration:**
```yaml
# In GitHub Actions
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v3
  with:
    version: latest
    args: --disable-all --enable=bodyclose,noctx,errcheck,gosec,staticcheck
```

**Pre-commit Hook:**
```yaml
# .pre-commit-config.yaml
- repo: https://github.com/golangci/golangci-lint
  rev: v1.54.2
  hooks:
    - id: golangci-lint
      args: ['--disable-all', '--enable=bodyclose,noctx,errcheck']
```

### Performance Impact

- **Low overhead**: Single-pass static analysis
- **Fast execution**: Typically <100ms on medium-sized projects
- **No external dependencies**: Pure Go static analysis
- **Scales well**: Analysis time grows linearly with codebase size

### Summary of Interactions

| Linter | Relationship | Type | Reason |
|--------|--------------|------|--------|
| `noctx` | **Essential** | Synergistic | HTTP safety: cancellation + resource cleanup |
| `errcheck` | **Highly Recommended** | Synergistic | Ensures Close() errors aren't ignored |
| `gosec` | **Recommended** | Complementary | Security + resource exhaustion prevention |
| `staticcheck` | **Recommended** | Complementary | General bug detection |
| `gosimple` | **Compatible** | Neutral | Code quality without overlap |
| `interfacer` | **Compatible** | Neutral | Different concerns |

**Bottom Line**: bodyclose is a **critical** linter for any Go project making HTTP requests. When combined with `noctx` and `errcheck`, it provides comprehensive protection against HTTP-related resource leaks. The minor risk of false positives is far outweighed by the prevention of production outages from connection pool exhaustion.

---

**Reference:** https://github.com/timakin/bodyclose
