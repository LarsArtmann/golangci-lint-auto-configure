# decorder Linter - Comprehensive Analysis

## What the Linter Does

**decorder** enforces **declaration order** and **grouping rules** for Go code elements (constants, variables, types, functions). It ensures code is organized in a consistent, predictable structure that improves readability and maintainability.

### Origins

- **Repository**: [gitlab.com/bosi/decorder](https://gitlab.com/bosi/decorder)
- **Integrated**: golangci-lint v1.44.0 (January 2022)
- **Purpose**: Code organization and structure enforcement

### What It Checks

**1. Declaration Order**: Enforces that code elements appear in a specified sequence

**2. Multiple Declaration Count**: Flags when you have multiple declarations of the same type that could be grouped

**3. Init Function Placement**: Ensures `init()` functions appear in the correct location

### Example Violations

```go
package example

// ❌ Violation: func declared before type
func doSomething() {}

// ❌ Violation: type after function (if order is enforced)
type MyStruct struct{}

// ❌ Violation: multiple var declarations (should be grouped)
var a int
var b string

// ❌ Violation: init() not first (if init-first check enabled)
var x = 10

func init() {}  // Would be flagged
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **Large codebases** (>10K LOC) where consistent organization improves readability
- **Enterprise projects** with strict coding standards
- **Team environments** with multiple developers
- **Codebases undergoing standardization** or refactoring
- **New projects** where you want to establish clear organizational patterns
- **Open source libraries** - Consistent structure helps contributors

**Scenarios:**
- **Code reviews** - Objective standard for code organization
- **Onboarding** - Predictable structure helps new team members
- **Consistency** - Eliminates debates about where to place declarations
- **Code generation** - Ensures generators follow the same patterns

### ❌ Disable For:

**Project Types:**
- **Small projects** (<1K LOC) where organization is less critical
- **Personal/solo projects** - Flexibility is more valuable
- **Prototypes and experiments** - Speed over structure
- **Generated code** - May follow different conventions
- **Code with mixed styles** - Would generate too many violations

**Philosophical Reasons:**
- Teams that value flexibility over rigid structure
- Projects where code location is determined by domain logic
- Codebases following different organizational principles (e.g., grouping by feature)

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    decorder:
      # Declaration order sequence (default: [const,var,func,type])
      dec-order:
        - const
        - var
        - func
        - type
      
      # Ignore underscore vars in count checks (default: false)
      ignore-underscore-vars: false
      
      # Disable all declaration count checks (default: true)
      disable-dec-num-check: true
      
      # Disable specific type count checks:
      disable-type-dec-num-check: false    # Multiple type declarations
      disable-const-dec-num-check: false   # Multiple const declarations  
      disable-var-dec-num-check: false     # Multiple var declarations
      
      # Disable declaration order check (default: true)
      disable-dec-order-check: true
      
      # Disable init() function must be first check (default: true)
      disable-init-func-first-check: true
```

### Configuration Options Explained

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `dec-order` | `[]string` | `["const","var","func","type"]` | Sequence for code organization |
| `ignore-underscore-vars` | `bool` | `false` | Skip underscore vars (like `_ int`) in count checks |
| `disable-dec-num-check` | `bool` | `true` | Disable all "multiple declarations" checks |
| `disable-type-dec-num-check` | `bool` | `false` | Disable multiple type declaration check |
| `disable-const-dec-num-check` | `bool` | `false` | Disable multiple const declaration check |
| `disable-var-dec-num-check` | `bool` | `false` | Disable multiple var declaration check |
| `disable-dec-order-check` | `bool` | `true` | Disable order checking |
| `disable-init-func-first-check` | `bool` | `true` | Disable init() placement checking |

### Common Configuration Patterns

#### ✅ Minimal Configuration (Count Checks Only)
```yaml
# Flag multiple declarations but ignore order
version: "2"
linters:
  enable:
    - decorder
  
  settings:
    decorder:
      disable-dec-num-check: false
      disable-dec-order-check: true
      disable-init-func-first-check: true
```

#### ✅ Order-Focused Configuration
```yaml
# Enforce declaration order but ignore grouping
version: "2"
linters:
  enable:
    - decorder
  
  settings:
    decorder:
      dec-order: [const, var, type, func]
      disable-dec-num-check: true
      disable-dec-order-check: false
      disable-init-func-first-check: true
```

#### ✅ Strict Configuration (All Checks)
```yaml
# Match Go standard library style
version: "2"
linters:
  enable:
    - decorder
  
  settings:
    decorder:
      dec-order: [const, var, type, func]
      ignore-underscore-vars: true
      disable-dec-num-check: false
      disable-dec-order-check: false
      disable-init-func-first-check: false
```

#### ✅ Relaxed Configuration
```yaml
# Only check var grouping, ignore everything else
version: "2"
linters:
  enable:
    - decorder
  
  settings:
    decorder:
      disable-dec-num-check: false
      disable-var-dec-num-check: false
      disable-type-dec-num-check: true
      disable-const-dec-num-check: true
      disable-dec-order-check: true
      disable-init-func-first-check: true
```

### Best Practices

1. **Start small**: Enable only declaration count checks initially
2. **Use standard order**: `[const, var, type, func]` follows Go conventions
3. **Exclude test files**: Tests often have different organization patterns
4. **Exclude generated code**: Generated code may not follow conventions
5. **Document custom orders**: If using non-standard order, document why
6. **Use nolint sparingly**: Only for truly exceptional cases

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

decorder **complements** these linters by checking different aspects of code structure:

-  **`gofmt` / `gofumpt`**  : Formatting (whitespace) vs. decorder's organization (declaration order)
-  **`gci`**  : Import organization vs. decorder's code declaration organization
-  **`godot`**  : Comment formatting vs. decorder's code structure
-  **`nonamedreturns`**  : Function style vs. decorder's function placement

**Complete Code Style Suite:**
```yaml
linters:
  enable:
    # Organization
    - decorder
    - gci
    
    # Formatting
    - gofmt
    - gofumpt
    
    # Style
    - godot
    - nonamedreturns
```

### ⚠️ Related Linters

**Minimal Overlap:**
-  **`revive` (max-public-structs)**  : Both touch type declarations but different concerns
-  **`wsl` (whitespace)**  : Both care about code structure but focus on different aspects
-  **`gocritic`**  : May have some stylistic overlap but complementary overall

**Independent Operation:**
- No known conflicts or contradictory warnings
- Can safely run with all other linters
- Focused scope prevents overlapping functionality

### 📊 Performance Impact

- **Very low overhead**: Single-pass AST analysis
- **Fast execution**: Only checks declaration order and counts
- **Negligible memory usage**: No complex data structures
- **Recommended in CI/CD**: Zero performance impact

## Practical Examples

### ✅ Before and After: Declaration Order

```go
// ❌ Before: Chaotic order (violates dec-order: [const,var,type,func])
package config

func Load() (*Config, error) { }

type Config struct { }

var DefaultPort int

const Version = "1.0.0"

var DebugMode bool

func init() { }
```

```go
// ✅ After: Well-organized
package config

const Version = "1.0.0"

var (
    DefaultPort int
    DebugMode   bool
)

type Config struct { }

func init() { }

func Load() (*Config, error) { }
```

### ✅ Before and After: Multiple Declarations Grouped

```go
// ❌ Before: Scattered declarations
package server

var Port int

var Host string

var Timeout time.Duration

type Handler struct { }

type Middleware struct { }

var Logger *zap.Logger
```

```go
// ✅ After: Grouped declarations
package server

var (
    Port    int
    Host    string
    Timeout time.Duration
    Logger  *zap.Logger
)

type (
    Handler     struct { }
    Middleware  struct { }
)
```

### ✅ Using Exclusions for Necessary Deviations

```go
// Some interfaces require context in structs (like webdav.File)
package webdav

type file struct {
    ctx context.Context //nolint:containedctx // Required by webdav.File interface
    fs  http.FileSystem
}

// But we can still organize properly
type fileSystem struct {
    root string
}

const (
    opOpen   = "OPEN"
    opClose  = "CLOSE"
)

var (
    ErrNotExist = errors.New("file does not exist")
    ErrNotDir   = errors.New("not a directory")
)

func init() {
    // Initialize webdav handlers
}
```

## Common Questions

**Q: Does decorder check private vs public declarations?**
A: No - it treats all declarations the same, regardless of visibility.

**Q: Can I enforce different orders for different packages?**
A: No - decorder uses a single global configuration for the entire project.

**Q: Does decorder check interface vs struct order?**
A: No - it treats all `type` declarations the same.

**Q: Should I exclude test files?**
A: Yes - tests often have different organizational needs.

## Default State

**IMPORTANT: decorder is DISABLED by default** in golangci-lint, and even when enabled, most checks are OFF:

```yaml
# Default behavior when enabled (all checks disabled)
linters:
  settings:
    decorder:
      disable-dec-num-check: true
      disable-dec-order-check: true
      disable-init-func-first-check: true
```

**You must explicitly opt into each type of check.** This conservative approach means you have full control over what decorder enforces.

## Summary

**decorder** is a **specialized linter** that enforces code organization standards. It:

- ✅ Encourages consistent declaration patterns
- ✅ Improves code readability through structure
- ✅ Helps large teams maintain consistent style
- ✅ Has zero performance impact
- ✅ Integrates well with other style linters
- ⚠️ Is disabled by default (requires explicit configuration)
- ⚠️ May be too rigid for some projects

**Recommendation**: **ENABLE** only if your team values strict, consistent declaration ordering and is willing to maintain it. For most projects, **keep disabled** or enable only the declaration count checks (which are less controversial).

**Verdict**: Most useful for large enterprise codebases, open source libraries, and teams with strict style guidelines. Overkill for small projects and rapid iteration codebases.

---

**Reference**: https://gitlab.com/bosi/decorder
