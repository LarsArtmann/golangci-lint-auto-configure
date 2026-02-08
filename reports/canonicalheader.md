# **canonicalheader** Linter Analysis

## **1. What the linter does**

The `canonicalheader` linter ensures that HTTP headers are written in their canonical MIME format when using methods on `http.Header`. It detects non-canonical header names and reports them with automatic fix suggestions.

### **Detailed Explanation**

HTTP headers have a standardized canonical format defined by RFC 2616 (superseded by RFC 7230), where:

- The first letter of each word is uppercase
- Letters following hyphens are uppercase
- All other letters are lowercase
- Example: `content-type` → `Content-Type`, `X-REQUEST-ID` → `X-Request-Id`

The `go` standard library automatically canonicalizes header keys when using `http.Header` methods (`Get`, `Set`, `Add`, `Del`, `Values`). However, using non-canonical forms in your code:

- Reduces readability and consistency
- Causes unnecessary canonicalization overhead
- Can lead to subtle bugs when directly accessing the header map

### **Detection Capabilities**

The linter identifies non-canonical header usage in:

```go
header.Get("content-type")     // Should be: "Content-Type"
header.Set("X-REQUEST-ID", "") // Should be: "X-Request-Id"
header.Add("cache_control", "") // Should be: "Cache-Control"
header.Del("AUTHORIZATION")     // Should be: "Authorization"
header.Values("ACCEPT-LANGUAGE") // Should be: "Accept-Language"
```

### **Key Features**

- **Automatic Fix**: Provides auto-fix capabilities (via `golangci-lint run --fix`)
- **Comprehensive Coverage**: Checks all `http.Header` methods
- **Precise Detection**: Distinguishes between method calls and direct map access
- **Zero Configuration**: Works out-of-the-box with sensible defaults

---

## **2. WHEN IT SHOULD BE ENABLED**

### **Critical Enable Scenarios**

Enable `canonicalheader` for **ALL** projects that:

| Project Type                | Rationale                                       | Priority       |
| --------------------------- | ----------------------------------------------- | -------------- |
| **Web APIs & HTTP Servers** | Directly manipulate HTTP headers frequently     | 🔴 Critical    |
| **Microservices**           | High volume of inter-service HTTP communication | 🔴 Critical    |
| **REST API Clients**        | Consistency with server expectations            | 🔴 Critical    |
| **Reverse Proxies**         | Critical to maintain header integrity           | 🔴 Critical    |
| **API Gateways**            | Handle diverse header patterns from clients     | 🔴 Critical    |
| **GraphQL Servers**         | HTTP layer interactions                         | 🟡 Recommended |
| **WebSocket Services**      | Uses HTTP upgrade headers                       | 🟡 Recommended |
| **OAuth/Auth Services**     | Heavy header-based authentication               | 🟡 Recommended |

### **Specific Use Cases**

**✅ Enable when:**

- Your codebase has more than 5 files using `http.Header`
- You maintain an API with external consumers
- Your team frequently works with `http.Request`/`http.Response`
- You're building libraries that wrap HTTP functionality
- Codebase has mixed header conventions
- Performance matters (avoiding redundant canonicalization)
- Code consistency is a priority

### **Real-World Example**

```go
// ❌ Without canonicalheader enabled (inconsistent)
func handleRequest(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("content-type", "application/json")
    w.Header().Set("X-REQUEST-ID", r.Header.Get("x-request-id"))
    w.Header().Set("cache_control", "no-cache")
}

// ✅ With canonicalheader (consistent & correct)
func handleRequest(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-Id", r.Header.Get("X-Request-Id"))
    w.Header().Set("Cache-Control", "no-cache")
}
```

---

## **3. WHEN IT SHOULD BE DISABLED**

### **Legitimate Disable Scenarios**

| Scenario                  | Rationale                          | Alternative                               |
| ------------------------- | ---------------------------------- | ----------------------------------------- |
| **Non-HTTP CLI tools**    | No HTTP header usage               | N/A                                       |
| **Pure data processing**  | No network I/O                     | N/A                                       |
| **Database libraries**    | Use database protocols, not HTTP   | N/A                                       |
| **System utilities**      | Focus on OS-level operations       | N/A                                       |
| **Custom header formats** | Intentionally non-standard headers | Apply `//nolint:canonicalheader` per-line |

### **⚠️ Conditional Disabling**

Disable **specific lines** (not the entire linter) when:

```go
// You're working with non-standard header formats
customHeaders := map[string]string{
    "x-custom-header": "value", // OK: Direct map access is not checked
}

// Intentionally using non-canonical form for testing
func TestNonStandardHeaders(t *testing.T) {
    h := http.Header{}
    h.Set("non-canonical", "test") // Use //nolint:canonicalheader here
}

// Migrating legacy code incrementally
```

### **What NOT to disable for**

- **"I don't want to fix old code"** → Use auto-fix: `golangci-lint run --fix`
- **"It doesn't catch bugs"** → It catches performance issues and improves maintainability
- **"Too many false positives"** → This linter has virtually zero false positives

---

## **4. HOW IT SHOULD BE CONFIGURED**

### **Basic Configuration**

```yaml
# .golangci.yml
linters:
  enable:
    - canonicalheader

  # No settings section needed - works with sensible defaults
```

### **Advanced Configuration Examples**

**Minimal HTTP Service:**

```yaml
version: "2"
linters:
  enable:
    - gosec
    - errcheck
    - staticcheck
    - canonicalheader # Critical for any HTTP service
```

**Comprehensive Web Project:**

```yaml
version: "2"
linters:
  enable:
    # All HTTP-related linters
    - bodyclose
    - canonicalheader
    - contextcheck
    - noctx

  settings:
    # No canonicalheader settings needed
```

### **Integration with CI/CD**

```yaml
# .github/workflows/lint.yml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v6
  with:
    version: latest
    args: --timeout=10m --fix

# Auto-fix canonical header issues in PRs
- name: Commit auto-fixes
  run: |
    if [[ $(git status --porcelain) ]]; then
      git add .
      git commit -m "fix: auto-correct canonical header format [skip ci]"
      git push
    fi
```

### **Editor Integration**

**VS Code settings.json:**

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.lintFlags": ["--fast", "--fix"]
}
```

---

## **5. HOW IT INTERFERES OR WORKS TOGETHER WITH OTHER LINTERS**

### **✅ Complementary Linters**

| Linter                    | Synergy                                          | Relationship Type     |
| ------------------------- | ------------------------------------------------ | --------------------- |
| **`staticcheck (S1035)`** | **Strong overlap** but different detection scope | ⚠️ Partial Redundancy |
| **`bodyclose`**           | Both guard HTTP correctness; work independently  | 🟢 Synergistic        |
| **`contextcheck`**        | HTTP + context safety; independent concerns      | 🟢 Complementary      |
| **`noctx`**               | HTTP request context awareness; no overlap       | 🟢 Complementary      |
| **`sloglint`**            | HTTP + logging; unrelated but both critical      | 🟢 Independent        |

### **⚠️ Relationship with staticcheck S1035**

Both detect canonical header issues but differ in scope:

```go
// S1035 detects: Redundant call to http.CanonicalHeaderKey
h.Set(http.CanonicalHeaderKey("key"), "value")

// canonicalheader detects: Non-canonical string literals
h.Set("content-type", "application/json")
```

**Best Practice:**

- Keep **both enabled** - they catch different patterns
- S1035 focuses on eliminating redundant calls
- canonicalheader ensures consistency in literal strings

### **🟢 Perfect Coordination Matrix**

```yaml
# Recommended HTTP Linter Set
linters:
  enable:
    # HTTP correctness
    - bodyclose # Ensure HTTP response bodies are closed
    - canonicalheader # Ensure canonical header format

    # Context safety
    - contextcheck # Check context propagation
    - noctx # Ensure HTTP requests have context

    # Security
    - gosec # Generic security checks

    # Error handling
    - wrapcheck # Wrap external errors
    - errorlint # Error handling best practices
```

### **Configuration Conflicts**

**NONE**. The `canonicalheader` linter:

- Has no configuration options
- Doesn't conflict with any other linter
- Cannot be misconfigured
- Works with all Go versions 1.22+

### **Performance Impact**

- **Very Low**: Only checks string literals in `http.Header` method calls
- **Fast**: Linear scan with minimal AST traversal
- **Auto-fixable**: Zero manual effort required

### **Migration Strategy**

**Step 1**: Enable with auto-fix in CI

```bash
golangci-lint run --fix --disable-all --enable=canonicalheader
```

**Step 2**: Review changes (all will be safe)

```bash
git diff  # All changes are header case corrections
```

**Step 3**: Enable permanently

```yaml
linters:
  enable:
    - canonicalheader
```

**Step 4**: Add to pre-commit hook

```yaml
# .pre-commit-config.yaml
- repo: https://github.com/golangci/golangci-lint
  hooks:
    - id: golangci-lint
      args: [--fix]
```

---

## **6. PRACTICAL EXAMPLES**

### **Before (Inconsistent)**

```go
func sendResponse(w http.ResponseWriter, data []byte) {
    w.Header().Set("content-type", "application/json")
    w.Header().Set("X-REQUEST-ID", getRequestID())
    w.Header().Set("cache_control", "public, max-age=3600")
    w.Header().Set("X-Custom-Header", "value")
    w.Header().Add("set-cookie", "session=abc123")
    w.WriteHeader(http.StatusOK)
    w.Write(data)
}
```

### **After (Fixed)**

```go
func sendResponse(w http.ResponseWriter, data []byte) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-Id", getRequestID())
    w.Header().Set("Cache-Control", "public, max-age=3600")
    w.Header().Set("X-Custom-Header", "value")
    w.Header().Add("Set-Cookie", "session=abc123")
    w.WriteHeader(http.StatusOK)
    w.Write(data)
}
```

### **Edge Cases It Handles**

```go
// ✅ Direct map access is NOT flagged (intentional)
headers := make(http.Header)
headers["non-canonical"] = []string{"value"} // OK

// ✅ Variables are NOT flagged (analyzer checks literals)
key := "content-type"
header.Get(key) // OK

// ❌ String literals ARE flagged
header.Get("content-type") // Flagged
```

---

## **7. FINAL RECOMMENDATIONS**

### **Enable by Default For:**

- ✅ All web projects
- ✅ All microservices
- ✅ All API clients
- ✅ Any project importing `"net/http"`

### **Configuration Priority:**

1. **Tier 1 (Critical)**: `gosec`, `errcheck`, `canonicalheader`
2. **Tier 2 (Recommended)**: `staticcheck`, `bodyclose`, `contextcheck`
3. **Tier 3 (Optional)**: `wrapcheck`, `errorlint`, `noctx`

### **Action Items**

```bash
# 1. Add to your configuration
echo "  - canonicalheader" >> .golangci.yml

# 2. Auto-fix existing issues
golangci-lint run --fix --disable-all --enable=canonicalheader

# 3. Review and commit
git add -A && git commit -m "style: enforce canonical HTTP header format"

# 4. Enable in CI
# Add to your .github/workflows/lint.yml
```

**Bottom Line**: Enable this linter in every Go project that touches HTTP. The cost is zero, the benefit is consistent, correct code.
