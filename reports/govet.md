# govet Linter - Comprehensive Analysis

## What the Linter Does

**govet** is Go's built-in static analysis tool, integrated into `go vet` command. It performs a suite of checks on Go source code to detect suspicious constructs, bugs, and potential issues that the compiler doesn't catch. It's part of Go's standard tooling and is considered **CRITICAL priority**.

### The Problem It Detects

govet includes multiple analyzer passes that catch issues:

- **Struct tags** - Incorrect or inconsistent struct field tags
- **Printf usage** - Incorrect format strings in printf-like functions
- **Unreachable code** - Code paths that can never be executed
- **Marshal/Unmarshal issues** - Problems with JSON/XML encoding
- **Nil pointer checks** - Potential nil dereferences
- **Method signatures** - Incorrect receiver types, missing parameters
- **Control flow** - Infinite loops, impossible conditions

### How It Works

govet uses Go's compiler and type system to analyze code:

1. **Type checking** - Verifies types are used correctly
2. **Control flow analysis** - Finds unreachable code and other issues
3. **Struct tag validation** - Checks JSON/XML/YAML tags
4. **Printf format checking** - Validates format strings match arguments
5. **Method signature analysis** - Ensures methods can be called correctly

### Examples

```go
// ❌ BAD: Incorrect struct tag
type User struct {
    Name  string `json:"name"`  // govet: invalid struct tag
    Email string `json:"email"`
}

// ✅ GOOD: Correct struct tag
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

```go
// ❌ BAD: Wrong printf format
func logUser(id int, name string) {
    fmt.Printf("User: %d", id, name)  // govet: wrong type for %d
}

// ✅ GOOD: Correct printf format
func logUser(id int, name string) {
    fmt.Printf("User: %d %s", id, name)
}
```

```go
// ❌ BAD: Unreachable code
func process(x int) int {
    if x > 10 {
        return x * 2
    }
    return x
    fmt.Println("Unreachable")  // govet: unreachable code
}

// ✅ GOOD: Remove unreachable code
func process(x int) int {
    if x > 10 {
        return x * 2
    }
    return x
}
```

```go
// ❌ BAD: Self-assignment in range (Go < 1.22)
func process(items []Item) {
    for _, item := range items {
        item := item  // govet: self-assignment
    }
}

// ✅ GOOD: Use new variable name
func process(items []Item) {
    for _, item := range items {
        processedItem := item
    }
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **All Go projects** - govet is Go's standard vet tool
- **Production applications** - Essential for catching bugs before deployment
- **Libraries and APIs** - Ensures correct method signatures
- **CLI tools** - Catches format string issues
- **Web services** - Validates struct tags for JSON/XML

**Specific Scenarios:**

**1. All Production Code**
- Any code deployed to production
- User-facing applications
- Services with SLA requirements

**2. Code with Struct Tags**
- JSON encoding/decoding
- YAML configuration parsing
- XML serialization
- Database ORM models

**3. Printf/Logging Code**
- Custom logging functions
- Formatted output
- Debug print statements
- Error message formatting

**4. Interface and Method Code**
- Public APIs
- Interface implementations
- Method definitions
- Receiver types

**5. Control Flow**
- Complex conditional logic
- Loop constructs
- Error handling paths

### ❌ Disable For:

**Specific Scenarios:**

**1. Code with Known False Positives**
```yaml
# Only when absolutely necessary and documented
issues:
  exclude-rules:
    - text: "composite literal uses unkeyed fields"
      linters: [govet]
```

**2. Test Files (Rarely Needed)**
```yaml
# govet usually doesn't need exclusion for tests
# Only if using specific test patterns that trigger false positives
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [govet]
```

**3. Generated Code**
```yaml
run:
  skip-dirs:
    - generated
    - vendor
```

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** - Go's standard vet tool, catches compiler-missed bugs
- **Effort**: Low - Built into Go toolchain, no setup
- **Recommendation**: **ALWAYS ENABLE** for all Go projects

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    govet:
      # Additional analyzer checks
      # Default: []
      enable:
        - atomic
        - bools
        - buildtags
        - composites
        - copylocks
        - deepequal
        - embedded
        - errorsas
        - fieldalignment
        - findcall
        - httpresponse
        - ifaceassert
        - loopclosure
        - lostcancel
        - nilfunc
        - printf
        - reflectvaluecompare
        - shadow
        - shifts
        - sloglint
        - stdmethods
        - stringintconv
        - structtag
        - testinggoroutine
        - tests
        - unmarshal
        - unreachable
        - unsafeptr
        - unusedresult
        - unusedwrite

      # Analyzer tags to run (comma-separated)
      # Default: ""
      tags: ""

      # All analyzers (disable default analyzers)
      # Default: false
      enable-all: false

      # Minimum Go version
      # Default: ""
      min-compatibility: ""

      # Maximum Go version
      # Default: ""
      max-compatibility: ""

      # Custom analyzer settings
      # Default: {}
      settings: {}
```

### Key Analyzer Options

**Atomic**
- Checks for common mistakes using sync/atomic

**Bools**
- Checks for misuse of booleans

**Composite**
- Checks for unkeyed composite literals

**Copylocks**
- Checks for locks being copied by value

**Fieldalignment**
- Checks for wasted space due to struct field alignment

**Printf**
- Checks for consistent formatting of printf-style functions

**Shadow**
- Checks for variable shadowing

**Structtag**
- Checks that struct field tags have correct format

**Unreachable**
- Checks for unreachable code

**Unsafe**
- Checks for misuse of unsafe.Pointer

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Most production applications
version: "2"
linters:
  settings:
    govet:
      enable:
        - atomic
        - copylocks
        - fieldalignment
        - printf
        - shadow
        - structtag
        - unreachable
        - unusedresult
```

#### ✅ Strict Configuration
```yaml
# Security-critical, high-quality standards
version: "2"
linters:
  settings:
    govet:
      enable-all: true
```

#### ✅ Performance-Focused Configuration
```yaml
# Performance-critical applications
version: "2"
linters:
  settings:
    govet:
      enable:
        - copylocks
        - fieldalignment
        - printf
        - shadow
```

#### ✅ Minimal Configuration
```yaml
# Use only default analyzers
version: "2"
linters:
  settings:
    govet: {}
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

govet works excellently with:

| Linter | Relationship | Value |
|--------|--------------|---------|
| **staticcheck** | Complementary | govet: AST-level, staticcheck: deeper analysis |
| **errcheck** | Complementary | govet: Printf, errcheck: all errors |
| **gosec** | Complementary | govet: unsafe pointers, gosec: security |
| **nilerr** | Complementary | govet: nil issues, nilerr: nil returns |
| **bodyclose** | Complementary | govet: resource leaks, bodyclose: HTTP bodies |
| **goconst** | Complementary | govet: Printf, goconst: repeated strings |

**Complete Analysis Suite:**
```yaml
linters:
  enable:
    - govet           # Standard Go vet (CRITICAL)
    - staticcheck      # Advanced analysis (CRITICAL)
    - errcheck         # Error handling (CRITICAL)
    - gosec            # Security (CRITICAL)
    - nilerr           # Nil errors (CRITICAL)
    - bodyclose         # Resource leaks (HIGH)
```

### 🔒 Minimal Overlap, No Conflicts

| Linter | Overlap | Recommendation |
|---------|----------|----------------|
| govet + **staticcheck** | Some Printf overlap | Use both - complementary |
| govet + **revive** | Some style overlap | Use both - different focus |
| govet + **errcheck** | Some error overlap | Use both - errcheck is more comprehensive |

**Why No Conflicts:**
- govet: Standard Go tooling, AST-level analysis
- staticcheck: Deeper analysis, more checks
- Each linter covers different aspects of code correctness

## Practical Examples

### ✅ Example 1: Struct Tag Fixes

```go
// ❌ BAD: Missing tag or incorrect format
type Product struct {
    ID    string `json:"id"`
    Name  string `json:name`   // govet: missing quotes
    Price float `json:price,required`
}

// ✅ GOOD: Correct tags
type Product struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Price float `json:"price"`
    Tags  []string `json:"tags,omitempty"`
}
```

### ✅ Example 2: Printf Formatting

```go
// ❌ BAD: Wrong format or wrong number of args
func logData(id int, name string, email string) {
    fmt.Printf("User: %d %s %d", id, name, email)  // govet: wrong type for 3rd arg
}

// ✅ GOOD: Correct format
func logData(id int, name string, email string) {
    fmt.Printf("User: %d %s %s", id, name, email)
}
```

### ✅ Example 3: Unreachable Code

```go
// ❌ BAD: Code after return
func process(x int) int {
    if x > 0 {
        return x
    }
    return 0
    fmt.Println("This is unreachable")  // govet: unreachable
}

// ✅ GOOD: Remove unreachable code
func process(x int) int {
    if x > 0 {
        return x
    }
    return 0
}
```

### ✅ Example 4: Field Alignment

```go
// ❌ BAD: Poor field alignment wastes space
type BadStruct struct {
    A bool
    B bool
    C int8
    D int64
}

// ✅ GOOD: Proper alignment reduces padding
type GoodStruct struct {
    A bool
    B bool
    C int64
    D int8
}
```

### ✅ Example 5: Lock Copying

```go
// ❌ BAD: Lock copied by value
func processData(mu sync.Mutex) {
    lockMu := mu  // govet: copylocks
    lockMu.Lock()
    // ... process
    lockMu.Unlock()
}

// ✅ GOOD: Use pointer to lock
func processData(mu *sync.Mutex) {
    mu.Lock()
    // ... process
    mu.Unlock()
}
```

### ✅ Example 6: Variable Shadowing

```go
// ❌ BAD: Variable shadowed
func process(x int) int {
    for i := 0; i < 10; i++ {
        x := i * 2  // govet: shadow - x from outer scope
    }
    return x
}

// ✅ GOOD: Use different variable name
func process(x int) int {
    for i := 0; i < 10; i++ {
        y := i * 2  // Different name
    }
    return x
}
```

## Best Practices

1. **ALWAYS enable govet** - It's Go's standard vet tool
2. **Run before commits** - Part of standard Go workflow
3. **Use with staticcheck** - Deeper analysis beyond govet
4. **Enable atomic analyzer** - For code using sync/atomic
5. **Enable copylocks analyzer** - Prevents lock-related bugs
6. **Enable fieldalignment analyzer** - Reduces memory waste
7. **Enable printf analyzer** - Catches format string issues
8. **Enable shadow analyzer** - Prevents variable shadowing bugs
9. **Enable structtag analyzer** - Validates JSON/XML/YAML tags
10. **Combine with errcheck** - Complete error handling coverage

## Common Scenarios and Solutions

### Scenario 1: JSON Encoding Issues

**Problem:** Struct tags incorrect, JSON marshaling fails.

**Solution:** Use structtag analyzer.
```yaml
linters:
  settings:
    govet:
      enable:
        - structtag
```

### Scenario 2: Printf Formatting Bugs

**Problem:** Wrong format strings cause runtime panics.

**Solution:** Use printf analyzer.
```yaml
linters:
  settings:
    govet:
      enable:
        - printf
```

### Scenario 3: Locking Bugs

**Problem:** Locks copied by value cause deadlocks.

**Solution:** Use copylocks analyzer.
```yaml
linters:
  settings:
    govet:
      enable:
        - copylocks
```

### Scenario 4: Performance Issues

**Problem:** Poor struct alignment wastes memory and CPU cache.

**Solution:** Use fieldalignment analyzer.
```yaml
linters:
  settings:
    govet:
      enable:
        - fieldalignment
```

## Summary

**govet** is Go's **standard static analysis tool**:

- ✅ **Highest priority** - Built into Go toolchain, essential
- ✅ **Comprehensive** - Multiple analyzers covering many issue types
- ✅ **Fast execution** - Optimized for performance
- ✅ **Built-in** - No installation or configuration needed
- ✅ **Official** - Part of Go's standard tooling
- ✅ **Compatible** - Works with all Go code
- ✅ **Configurable** - Enable/disable specific analyzers
- ✅ **Complementary** - Works with staticcheck, errcheck, gosec
- ⚠️ **May need additional analyzers** - Default doesn't enable all
- ⚠️ **Requires Go installation** - Relies on Go compiler

**Recommendation:** **ALWAYS ENABLE** with recommended analyzers (atomic, copylocks, fieldalignment, printf, shadow, structtag). Combine with **staticcheck** for deeper analysis. Use with **errcheck** for complete error handling. Part of standard Go development workflow (`go vet`, `go test`, `go build`).

**Top 3 Analyzers to Enable:**
1. **structtag** - Validates JSON/XML/YAML tags
2. **printf** - Catches format string bugs
3. **copylocks** - Prevents lock-related bugs

---

**Reference:** https://pkg.go.dev/cmd/vet
