# containedctx Linter - Comprehensive Analysis

## What the Linter Detects

**containedctx** detects when `context.Context` is used as a struct field, which is considered an anti-pattern in Go. Contexts should be passed explicitly as function parameters, not stored in structs or objects.

### The Anti-Pattern

Storing context in structs violates Go's context design principles:

- **Contexts are request-scoped**: They should flow through call chains, not be stored
- **Prevents proper cancellation**: Stored contexts can outlive their intended lifetime
- **Hides dependencies**: Makes it unclear which functions require a context
- **Concurrency issues**: Shared context in structs creates race conditions
- **Testing difficulties**: Harder to inject test contexts

### Example Violation

```go
// ❌ BAD: context.Context in struct field
type Client struct {
    ctx context.Context  // containedctx: found struct field context.Context
    httpClient *http.Client
}

func (c *Client) Fetch(url string) error {
    req, _ := http.NewRequestWithContext(c.ctx, "GET", url, nil)
    _, err := c.httpClient.Do(req)
    return err
}
```

### Correct Pattern

```go
// ✅ GOOD: Pass context as function parameter
type Client struct {
    httpClient *http.Client
}

func (c *Client) Fetch(ctx context.Context, url string) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    _, err := c.httpClient.Do(req)
    return err
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

- **Web APIs and HTTP servers** - Enforces proper context propagation through handlers
- **Microservices** - Critical for distributed request tracing and cancellation
- **CLI applications** - Clear request/command lifetimes
- **Libraries and SDKs** - Makes context dependencies explicit for users
- **Event processors** - Proper event-scoped context handling
- **Database repositories** - Explicit transaction contexts

**Team Scenarios:**

- Training new Go developers - Teaches proper context usage patterns
- Codebases with inconsistent context handling
- Projects migrating from contexts-in-structs pattern
- Teams wanting to enforce Go best practices

**Code Quality Goals:**

- Improving testability - Easier to inject test contexts
- Enhancing code clarity - Explicit dependencies
- Preventing goroutine leaks - Proper context lifecycle
- Supporting distributed tracing - Context flows naturally

### Priority Assessment

- **Default Priority**: Medium-High
- **Essential for**: Web services, APIs, microservices
- **Low priority for**: CLI tools with no concurrent operations, pure computation libraries

## When It Should Be Disabled

### ❌ Consider Disabling For:

**Legitimate Technical Constraints:**

1. **Framework implementations** - Must satisfy interfaces that require context storage

   ```go
   // webdav.File interface requires context storage
   type webdavFile struct {
       ctx context.Context //nolint:containedctx // required by webdav.File interface
       entry fs.File
   }
   ```

2. **Legacy API compatibility** - Interfacing with systems that require context in structs
3. **Specialized use cases** - Context lifetime genuinely matches struct lifetime

**Code Patterns:**

- **Test code** - Often has different context management patterns
- **Generated code** - May not follow best practices
- **Temporary migration code** - While refactoring gradually

### Better Alternative to Disabling

Instead of disabling globally, use **exclusions**:

```yaml
# .golangci.yml
linters:
  enable:
    - containedctx

issues:
  exclude-rules:
    # Exclude test files
    - linters: [containedctx]
      path: _test\.go

    # Exclude specific framework implementations
    - linters: [containedctx]
      path: internal/webdav/

    # Exclude generated code
    - linters: [containedctx]
      path: (.+)_generated\.go
```

Or use targeted `nolint` directives:

```go
type webdavFile struct {
    ctx context.Context //nolint:containedctx // required by webdav.File interface
    entry fs.File
}
```

## How It Should Be Configured

### Configuration Structure

**containedctx** has **no specific configuration options**. It's a simple check that only requires enabling:

```yaml
# .golangci.yml
linters:
  enable:
    - containedctx
```

### Recommended Configurations

**Standard Web Project:**

```yaml
version: "2"
linters:
  enable:
    - containedctx
    - contextcheck
    - noctx
    - errcheck
    - govet

issues:
  exclude-rules:
    # Exclude test files where patterns differ
    - linters: [containedctx]
      path: (.+)_test\.go
```

**Library/API Project:**

```yaml
version: "2"
linters:
  enable:
    - containedctx
    - contextcheck
    - wrapcheck
    - errorlint

issues:
  # Strict enforcement - no exceptions in library code
  exclude-use-default: false

  exclude-rules:
    # Only exclude truly unavoidable cases
    - linters: [containedctx]
      path: pkg/webdav/
```

**Comprehensive HTTP Service:**

```yaml
version: "2"
linters:
  enable:
    # Context safety
    - containedctx
    - contextcheck
    - noctx
    - fatcontext

    # HTTP correctness
    - bodyclose
    - canonicalheader

    # Security & errors
    - gosec
    - errcheck
    - errorlint
```

### Best Practices

1. **Enable alongside contextcheck** - They complement each other perfectly
2. **Exclude test files** - Tests often have legitimate reasons for different patterns
3. **Document exceptions** - Use `//nolint:containedctx` with explanatory comments
4. **Refactor gradually** - Use exclusions for legacy code during migration
5. **Educate the team** - Explain why contexts shouldn't be stored in structs

## How It Interferes or Works Together With Other Linters

### ✅ Synergistic Linters

**Complementary Linters:**

- **`contextcheck`**: **Essential pairing** - containedctx checks struct fields, contextcheck verifies function parameter propagation
- **`noctx`**: Ensures HTTP requests use context - different but related concerns
- **`fatcontext`**: Detects nested contexts in loops - builds on context awareness
- **`wrapcheck`**: Context propagation + error wrapping = comprehensive error handling
- **`errorlint`**: Proper error handling with context cancellation

**Example Synergy:**

```go
// contextcheck ensures this passes context through call chain
// containedctx ensures we don't store it in the Client struct
// noctx ensures HTTP requests use the context
func (c *Client) Process(ctx context.Context, id string) error {
    data, err := c.fetchData(ctx, id)  // contextcheck validates this
    if err != nil {
        return err
    }
    return c.processData(ctx, data)  // contextcheck validates this
}
```

### ⚠️ Related Linters

**Code Complexity:**

- Often excluded alongside `gocyclo`, `gocognit`, `funlen` in test files
- Different concern but similar enforcement patterns

**Naming Conventions:**

- Works well with `revive` (var-naming rules)
- Complements `stylecheck` for overall code style

### 🔒 Potential Conflicts

**No Known Conflicts:**

- containedctx is narrowly focused with no overlapping functionality
- Safe to enable with all other linters
- Only checks struct field declarations

### 📊 Performance Impact

- **Minimal Overhead**: Simple AST check for struct field types
- **Fast Execution**: No complex dataflow analysis
- **Scales Linearly**: Proportional to number of struct definitions
- **Recommended in CI/CD**: Zero performance concerns

## Practical Examples

### ✅ Correct Pattern - Context as Parameter

```go
type UserService struct {
    db *sql.DB
    logger *zap.Logger
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // Context flows through call chain
    query := "SELECT id, name FROM users WHERE id = ?"
    row := s.db.QueryRowContext(ctx, query, id)

    var user User
    if err := row.Scan(&user.ID, &user.Name); err != nil {
        s.logger.ErrorContext(ctx, "failed to get user", zap.Error(err))
        return nil, err
    }
    return &user, nil
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    // Context for transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // ... use tx with context
    return tx.Commit()
}
```

### ❌ Anti-Pattern - Context in Struct

```go
// Don't do this!
type UserService struct {
    ctx context.Context  // containedctx would flag this
    db  *sql.DB
}

func (s *UserService) GetUser(id string) (*User, error) {
    // Unclear where context came from
    // Difficult to test with different contexts
    // Potential goroutine leaks
    row := s.db.QueryRowContext(s.ctx, query, id)
    // ...
}
```

### 📝 Justified Exception - WebDAV Implementation

```go
// Some interfaces require context storage
type fileHandler struct {
    ctx context.Context //nolint:containedctx // webdav.File interface requires this
    fs  http.FileSystem
}

// Required to satisfy webdav.File interface
func (f *fileHandler) Stat() (os.FileInfo, error) {
    // Must use stored context for WebDAV compatibility
    return f.fs.StatContext(f.ctx, f.path)
}
```

### 📝 Justified Exception - Long-lived Workers

```go
// When context matches struct lifetime exactly
type BackgroundProcessor struct {
    ctx    context.Context //nolint:containedctx // worker lifetime matches context
    cancel context.CancelFunc
    queue  chan Job
}

func NewProcessor() *BackgroundProcessor {
    ctx, cancel := context.WithCancel(context.Background())
    p := &BackgroundProcessor{
        ctx:    ctx,
        cancel: cancel,
        queue:  make(chan Job, 100),
    }
    go p.run()
    return p
}

func (p *BackgroundProcessor) run() {
    // Context lifetime exactly matches this goroutine
    <-p.ctx.Done()
}

func (p *BackgroundProcessor) Stop() {
    p.cancel() // Cancels the stored context
}
```

## Common Questions

**Q: Why can't I store context in a struct?**
A: Contexts are request-scoped and should flow through call chains. Storing them creates unclear lifetimes, makes testing harder, and can cause goroutine leaks.

**Q: What about database transaction objects?**
A: Transactions should also accept context in methods, not store it. Pass context to `BeginTx`, `QueryContext`, etc.

**Q: Are there ANY valid exceptions?**
A: Very few - primarily when implementing interfaces that require it (like webdav.File) or when the struct lifetime exactly matches context lifetime.

**Q: How do I refactor existing code?**
A: 1) Remove context from struct, 2) Add context parameter to methods, 3) Update all call sites to pass context.

## Summary

**containedctx** is a **valuable linter** that enforces proper Go context usage patterns. It prevents:

- Unclear context lifetimes
- Testing difficulties
- Potential goroutine leaks
- Hidden dependencies

**Recommendation**: **ENABLE** for all Go projects, especially web services, APIs, and microservices. Use exclusions for the rare legitimate exceptions and test files. Works best when combined with `contextcheck` and `noctx` for comprehensive context safety.

---

**Reference**: https://github.com/sivchari/containedctx
