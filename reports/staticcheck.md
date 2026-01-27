# staticcheck Linter - Comprehensive Analysis

## What the Linter Does

**staticcheck** is the **most advanced static analysis tool for Go**. It performs deep code analysis to find bugs, performance issues, suggest code simplifications, and point out dead code. It's a collection of multiple analysis passes that catch issues most other linters miss. It's the foundation of Go's static analysis ecosystem and is considered **CRITICAL priority**.

### The Problem It Detects

staticcheck includes three main components:

1. **staticcheck (SA* checks)** - Advanced bug detection and performance issues
2. **stylecheck (ST* checks)** - Code style improvements and common mistakes
3. **gosimple (S* checks)** - Code simplifications and redundancies

### Core Capabilities

**staticcheck (SA*):** Finds bugs, performance issues, and suspicious code:
- SA1001-Sa1019: Security and correctness issues
- SA2001-SA2099: Control flow and testing issues
- SA3001-SA3099: Performance problems
- SA4001-SA4029: Additional bug detection
- SA5001-SA5011: Control flow issues

**stylecheck (ST*):** Finds style issues and common mistakes:
- ST1000-ST1021: Common mistakes and stylistic issues
- ST1012-ST1019: Additional style checks

**gosimple (S*):** Suggests code simplifications:
- S1001-S1037: Code simplifications
- S1038-S1039: Additional simplifications

### Examples

```go
// ❌ BUG: Self-assignment (SA4001)
func process(x int) {
    x = x  // SA4001: self-assignment of x to x
    fmt.Println(x)
}

// ✅ FIXED
func process(x int) {
    x = x + 1  // Or remove the assignment entirely
    fmt.Println(x)
}
```

```go
// ❌ BUG: Wrong type in printf (SA5009)
func logUser(id int, name string) {
    fmt.Printf("User ID: %d Name: %s", name, id)  // SA5009: %d expects int, got string
}

// ✅ FIXED
func logUser(id int, name string) {
    fmt.Printf("User ID: %d Name: %s", id, name)
}
```

```go
// ❌ PERFORMANCE: Copying lock by value (SA2001)
type MutexWrapper struct {
    mu sync.Mutex  // struct contains sync.Mutex
}

func (w MutexWrapper) lock() {  // SA2001: method receiver should be pointer
    w.mu.Lock()
}
```

```go
// ❌ BUG: Nil dereference after check (SA5011)
func getValue(v *int) int {
    if v == nil {
        return *v  // SA5011: nil dereference in line it's guarded against
    }
    return *v
}

// ✅ FIXED
func getValue(v *int) int {
    if v == nil {
        return 0
    }
    return *v
}
```

```go
// ❌ SECURITY: Deprecated function (SA1019)
func hash(data []byte) []byte {
    return md5.Sum(data)  // SA1019: md5.Sum is deprecated
}

// ✅ FIXED
func hash(data []byte) []byte {
    h := sha256.New()
    h.Write(data)
    return h.Sum(nil)
}
```

```go
// ❌ STYLE: Unused parameter (ST1003)
func calculate(x, y int) int {
    return x * 2  // ST1003: parameter 'y' is unused
}
```

```go
// ❌ SIMPLIFICATION: Should use range (S1001)
func sum(nums []int) int {
    total := 0
    for i := 0; i < len(nums); i++ {  // S1001: should use range
        total += nums[i]
    }
    return total
}

// ✅ FIXED
func sum(nums []int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **All production applications** - Most critical linter after govet
- **Public libraries and SDKs** - Ensures API correctness and security
- **Security-sensitive code** - Catches deprecated crypto, insecure patterns
- **Performance-critical systems** - Identifies inefficiencies
- **Long-lived projects** - Prevents technical debt accumulation
- **Projects with multiple contributors** - Maintains code quality standards
- **Enterprise applications** - Comprehensive code quality coverage

**Specific Scenarios:**

**1. Production Code**
- Any code deployed to production
- User-facing applications
- Services with SLA requirements
- Microservices and distributed systems

**2. Public APIs**
- Libraries consumed by other developers
- SDKs for external services
- Framework libraries
- Public APIs with error handling contracts

**3. Security-Critical Code**
- Authentication and authorization logic
- Cryptographic operations
- Input validation and sanitization
- File and network operations

**4. Codebases with Legacy Code**
- Identifies deprecated patterns
- Points out security vulnerabilities
- Suggests modern alternatives

**5. High-Quality Standards**
- Projects requiring strict code review
- Codebases with automated testing
- Teams focusing on technical debt reduction

### ❌ Disable For:

**Specific Scenarios:**

**1. Generated Code**
```yaml
run:
  skip-dirs:
    - generated
    - vendor

issues:
  exclude-rules:
    - path: (.+)_pb\.go
    - path: (.+)_generated\.go
      linters: [staticcheck]
```

**2. Legacy Code in Migration**
- When fixing issues would require complete rewrite
- Gradual adoption approach: enable on new code only
- Use path-based exclusions for legacy modules

**3. Experimental/Unstable Code**
- Code that may change significantly
- Prototypes and proof-of-concepts
- Research and spike solutions

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** - Catches bugs other linters miss, ensures code quality
- **Effort**: Low - Automatic analysis, minimal configuration
- **Recommendation**: **ALWAYS ENABLE** for all production Go code

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    staticcheck:
      # List of checks to enable
      # Default: all checks
      checks: ["all"]

      # Initialisms for acronyms (stylecheck only)
      # Default: ["API", "ASCII", "CPU", "CSS", "DNS", "EOF", ...]
      initialisms:
        - ID
        - JSON
        - URL

      # Whitelist for dot imports (stylecheck only)
      # Default: []
      dot-import-whitelist:
        - github.com/golang/protobuf/proto

      # Whitelist for HTTP status codes (stylecheck only)
      # Default: []
      http-status-code-whitelist:
        - 200
        - 404
```

### `checks` Option

- **Type**: `[]string`
- **Default**: `["all"]` (all checks enabled)
- **Description**: Control which staticcheck checks run

**Format Options:**
- `["all"]` - Enable all checks (recommended)
- `["SA*"]` - All SA (staticcheck) checks
- `["ST*"]` - All ST (stylecheck) checks
- `["S*"]` - All S (gosimple) checks
- `["SA1001", "SA1019"]` - Specific checks
- `["all", "-SA2001"]` - All except specific checks

**Commonly Disabled Checks:**
```yaml
checks:
  - all
  - -SA1019  # Allow deprecated functions during migration
  - -SA2006  # Allow printf-style in tests
  - -ST1005  # Allow certain unused params
```

### `initialisms` Option

- **Type**: `[]string`
- **Default**: `["API", "ASCII", "CPU", "CSS", "DNS", "EOF", ...]`
- **Description**: Capitalization of acronyms in comments (stylecheck only)

**Example:**
```yaml
initialisms:
  - API
  - JSON
  - URL
  - ID
  - SQL
```

### `dot-import-whitelist` Option

- **Type**: `[]string`
- **Default**: `[]`
- **Description**: Allow specific packages to use dot imports

**Example:**
```yaml
dot-import-whitelist:
  - github.com/golang/protobuf/proto
  - github.com/golang/protoc-gen-go
```

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Most production applications
version: "2"
linters:
  settings:
    staticcheck:
      checks: ["all"]

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [staticcheck]
    - path: (.+)_generated\.go
      linters: [staticcheck]
```

#### ✅ Relaxed Configuration
```yaml
# Legacy codebases, gradual adoption
version: "2"
linters:
  settings:
    staticcheck:
      checks:
        - all
        - -SA1019  # Allow deprecated functions
        - -SA2006  # Allow printf in tests
        - -ST1005  # Allow unused params in handlers

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [staticcheck]
```

#### ✅ Security-Focused Configuration
```yaml
# Security-critical applications
version: "2"
linters:
  settings:
    staticcheck:
      checks:
        - SA1001
        - SA1012
        - SA1019
        - SA9001
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

staticcheck works excellently with:

| Linter | Relationship | Value |
|--------|--------------|---------|
| **govet** | Complementary | govet: AST-level, staticcheck: deeper analysis |
| **errcheck** | Complementary | errcheck: unchecked errors, staticcheck: error handling correctness |
| **gosec** | Complementary | gosec: security issues, staticcheck: deprecated/crypto |
| **ineffassign** | Complementary | ineffassign: ineffective assignments, staticcheck: more comprehensive |
| **gosimple** | Part of staticcheck | Already included in staticcheck |
| **stylecheck** | Part of staticcheck | Already included in staticcheck |
| **nilerr** | Complementary | nilerr: nil errors, staticcheck: SA5011 |
| **revive** | Complementary | revive: style, staticcheck: advanced analysis |
| **depguard** | Complementary | depguard: imports, staticcheck: code quality |

**Complete Code Quality Suite:**
```yaml
linters:
  enable:
    - staticcheck      # Advanced analysis (CRITICAL)
    - govet           # Standard vet (CRITICAL)
    - errcheck         # Error handling (CRITICAL)
    - gosec            # Security (CRITICAL)
    - ineffassign      # Ineffective assignments (HIGH)
    - nilerr           # Nil errors (CRITICAL)
    - bodyclose        # Resource leaks (HIGH)
```

### 🔒 Minimal Overlap, No Conflicts

| Linter | Overlap | Recommendation |
|---------|----------|----------------|
| staticcheck + **gosimple** | Complete overlap | gosimple is part of staticcheck, don't enable separately |
| staticcheck + **stylecheck** | Complete overlap | stylecheck is part of staticcheck, don't enable separately |
| staticcheck + **errcheck** | SA5000 (unreachable) vs errcheck | Different focus, use both |
| staticcheck + **govet** | Some AST analysis overlap | Use both - complementary |

**Why No Conflicts:**
- staticcheck: Advanced deep analysis
- govet: Standard Go vet checks
- errcheck: Error handling completeness
- Each linter covers different aspects or depth

## Practical Examples

### ✅ Example 1: Fixing Performance Issues

```go
// ❌ INEFFICIENT: Copying slices (SA4006)
func filterEven(nums []int) []int {
    result := []int{}
    for _, n := range nums {
        if n%2 == 0 {
            result = append(result, n)  // SA4006: identity conversion
        }
    }
    return result
}

// ✅ OPTIMIZED
func filterEven(nums []int) []int {
    result := make([]int, 0, len(nums))
    i := 0
    for _, n := range nums {
        if n%2 == 0 {
            result[i] = n
            i++
        }
    }
    return result
}
```

### ✅ Example 2: Fixing Control Flow Issues

```go
// ❌ BUG: Always receives same value (SA2019)
func getMinMax(nums []int) (int, int) {
    min, max := nums[0], nums[0]
    for _, n := range nums {
        if n < min {
            min = n
        }
        if n > max {
            max = n
        }
    }
    return min, min  // SA2019: returns min for both values
}
```

### ✅ Example 3: Fixing Nil Pointer Issues

```go
// ❌ BUG: Nil dereference (SA5011)
func processItem(item *Item) {
    if item != nil {
        return item.Name
    }
    return item.Name  // SA5011: nil dereference in line it's guarded
}
```

### ✅ Example 4: Code Simplification

```go
// ❌ COMPLEX: Could use strings.Contains (S1005)
func hasPrefix(s, prefix string) bool {
    if len(s) < len(prefix) {
        return false
    }
    for i := 0; i < len(prefix); i++ {  // S1005: could use strings.HasPrefix
        if s[i] != prefix[i] {
            return false
        }
    }
    return true
}

// ✅ SIMPLIFIED
func hasPrefix(s, prefix string) bool {
    return strings.HasPrefix(s, prefix)
}
```

### ✅ Example 5: Fixing Deprecated Code

```go
// ❌ DEPRECATED: Using old crypto (SA1019)
func encrypt(password string) []byte {
    block, _ := aes.NewCipher(key)  // SA1019: aes.NewCipher deprecated
    // ...
}
```

## Best Practices

1. **ALWAYS enable staticcheck** for all production Go code
2. **Use `checks: ["all"]`** for maximum coverage
3. **Review SA4xxx issues** (performance) - can significantly impact code
4. **Review SA5xxx issues** (control flow) - often indicate logic errors
5. **Review SA1xxx issues** (security) - deprecated crypto is major risk
6. **Exclude generated code** - doesn't apply to auto-generated files
7. **Don't enable gosimple or stylecheck separately** - they're part of staticcheck
8. **Use initialisms whitelist** to customize for your project
9. **Consider enabling -SA2006** for better printf checking
10. **Combine with govet** - standard and advanced vetting

## Common Scenarios and Solutions

### Scenario 1: Legacy Code with Many Issues

**Problem:** staticcheck reporting hundreds of findings in legacy code.

**Solution:** Gradual adoption with exclusions.
```yaml
# Start with critical issues only
linters:
  settings:
    staticcheck:
      checks:
        - SA1xxx  # Security issues
        - SA5xxx  # Control flow
        - SA4xxx  # Performance
```

### Scenario 2: False Positives in Tests

**Problem:** Test code flagged for issues that don't apply.

**Solution:** Exclude test files.
```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [staticcheck]
```

### Scenario 3: Deprecated Functions During Migration

**Problem:** Code transitioning from old APIs to new ones.

**Solution:** Temporarily disable specific checks.
```yaml
linters:
  settings:
    staticcheck:
      checks:
        - all
        - -SA1019  # Allow deprecated during migration
```

### Scenario 4: Performance-Critical Code

**Problem:** Need to catch all performance issues.

**Solution:** Enable all SA4xxx performance checks.
```yaml
linters:
  settings:
    staticcheck:
      checks:
        - SA3001
        - SA4002
        - SA4006
        - SA4007
        - SA4010
```

## Summary

**staticcheck** is a **CRITICAL** linter that provides advanced static analysis:

- ✅ **Highest value** - Catches bugs other linters miss
- ✅ **Deep analysis** - Goes beyond simple pattern matching
- ✅ **Performance focus** - Identifies inefficiencies (SA4xxx)
- ✅ **Security focus** - Catches deprecated crypto, insecure patterns
- ✅ **Code quality** - Suggests simplifications, style improvements
- ✅ **Three components** - staticcheck, stylecheck, gosimple in one
- ✅ **Minimal configuration** - Works well with default settings
- ⚠️ **May be slow** - Large projects (several minutes)
- ⚠️ **May need exclusions** - Generated code, legacy patterns

**Recommendation:** **ALWAYS ENABLE** with `checks: ["all"]` for all production code. Exclude test files (`(.+)_test\.go`) and generated code. **Don't enable gosimple or stylecheck separately** as they're included. Combine with **govet**, **errcheck**, and **gosec** for comprehensive coverage. For large legacy codebases, consider gradual adoption with selective check enabling.

**Top 3 Benefits:**
1. Catches bugs that compile correctly but fail at runtime
2. Identifies performance issues before they impact production
3. Suggests code simplifications that improve maintainability

---

**Reference:** https://github.com/dominikh/go-tools/tree/HEAD/staticcheck
