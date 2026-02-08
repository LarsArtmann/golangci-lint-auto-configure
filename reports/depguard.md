# depguard Linter - Comprehensive Analysis

## What the Linter Does

**depguard** controls which packages can be imported in your Go code. It allows you to create rules that:

- **Block deprecated or unmaintained packages**
- **Enforce architectural boundaries** (prevent circular dependencies)
- **Guide teams toward preferred alternatives**
- **Standardize library usage** across large codebases
- **Maintain consistency** in package dependencies

### Key Features

- **Rule-based configuration**: Flexible allow/deny lists
- **File pattern matching**: Different rules for different code locations
- **Multiple list modes**: Strict (whitelist), lax (blacklist), or original
- **Rich error messages**: Inform developers why a package is blocked and what to use instead
- **Standard library awareness**: Built-in handling of Go standard library packages

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

- **Large codebases** with multiple teams needing architectural boundaries
- **Enterprise projects** requiring standardized library choices
- **Monorepos** where dependency consistency is critical
- **Migration projects** (e.g., moving from deprecated packages)
- **Security-conscious applications** - Block vulnerable package versions

**Scenarios:**

- **Migrating between package versions** (e.g., `io/ioutil` → `io/os`)
- **Enforcing architectural layers** - Prevent circular dependencies
- **Blocking unmaintained packages** - e.g., `github.com/pkg/errors`
- **Standardizing on specific testing frameworks** - e.g., only `stretchr/testify`
- **Preventing direct database access** - Force use of repository pattern

### ❌ Disable For:

**Project Types:**

- **Small, single-maintainer projects** - Flexibility is more valuable
- **Open-source libraries** - Need to minimize dependencies
- **Rapid prototyping/experimentation** - Package restrictions hinder productivity
- **Research/academic code** - Need freedom to try different packages

**Scenarios:**

- **Early-stage startups** - Need to move fast and experiment
- **Hackathon projects** - Quick iteration is priority
- **Learning/exploration code** - Want to try various packages

## How It Should Be Configured

### Configuration Structure (v2)

depguard v2 uses a **rule-based configuration** with `list-mode`, `files`, `allow`, and `deny` sections:

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    depguard:
      rules:
        <rule-name>:
          list-mode: <strict|lax|original>
          files:
            - <file-patterns>
          allow:
            - <allowed-packages>
          deny:
            - pkg: <package-name>
              desc: <message-to-user>
```

### List Modes

**`strict`** - **Whitelist approach**: ONLY packages in the `allow` list are permitted

```yaml
rules:
  main:
    list-mode: strict
    allow:
      - $gostd
      - github.com/myorg/core
      - github.com/prometheus/client_golang/prometheus
    # deny list not needed in strict mode
```

**`lax`** - **Blacklist approach**: All packages are allowed EXCEPT those in the `deny` list

```yaml
rules:
  deprecated:
    list-mode: lax
    allow:
      - $gostd
    deny:
      - pkg: io/ioutil
        desc: "replaced by io and os packages since Go 1.16"
```

**`original`** - **Backward compatibility**: Uses v1 behavior for migration

### File Pattern Variables

**`$all`** - Matches all Go files
**`$test`** - Matches test files (`*_test.go`)
**`!` prefix** - Negation (e.g., `!$test` excludes test files)
**Glob patterns** - `**/internal/**/*.go`, `cmd/**`, `pkg/service/*.go`

### Package Variables

**`$gostd`** - Matches all Go standard library packages (e.g., `fmt`, `io`, `net/http`)

### Example Configurations

#### ✅ Example 1: Block Deprecated Packages (Lax Mode)

```yaml
# .golangci.yml
version: "2"

linters:
  settings:
    depguard:
      rules:
        prevent_unmaintained_packages:
          list-mode: lax
          files:
            - $all
            - "!$test" # Exclude test files (tests can use any packages)
          allow:
            - $gostd
          deny:
            - pkg: io/ioutil
              desc: "replaced by io and os packages since Go 1.16"
            - pkg: github.com/pkg/errors
              desc: "use stdlib errors package instead"
            - pkg: github.com/satori/go.uuid
              desc: "use github.com/google/uuid instead"
            - pkg: github.com/gofrs/uuid
              desc: "use github.com/google/uuid instead"
```

#### ✅ Example 2: Enforce Strict Dependencies

```yaml
# .golangci.yml
version: "2"

linters:
  settings:
    depguard:
      rules:
        # Main code - only approved packages
        main_code:
          list-mode: strict
          files:
            - $all
            - "!$test"
          allow:
            - $gostd
            - github.com/myorg/core
            - github.com/myorg/config
            - github.com/prometheus/client_golang/prometheus
            - go.uber.org/zap
          deny: []

        # Test code - can use testing frameworks
        test_code:
          list-mode: strict
          files:
            - $test
          allow:
            - $gostd
            - github.com/stretchr/testify/assert
            - github.com/stretchr/testify/require
            - github.com/stretchr/testify/mock
            - github.com/myorg/testutil
```

#### ✅ Example 3: Architecture Layer Enforcement

```yaml
# .golangci.yml
version: "2"

linters:
  settings:
    depguard:
      rules:
        # Domain layer - cannot access infrastructure
        domain:
          list-mode: strict
          files:
            - "**/domain/**/*.go"
          allow:
            - $gostd
            - github.com/myorg/domain

        # Application layer - can access domain, limited infrastructure
        application:
          list-mode: strict
          files:
            - "**/application/**/*.go"
          allow:
            - $gostd
            - github.com/myorg/domain
            - github.com/myorg/application
            - github.com/myorg/shared

        # Infrastructure layer - can access everything
        infrastructure:
          list-mode: lax
          files:
            - "**/infrastructure/**/*.go"
          deny:
            - pkg: github.com/myorg/internal
              desc: "infrastructure should not depend on internal packages"
```

#### ✅ Example 4: Prevent Specific Patterns

```yaml
# .golangci.yml
version: "2"

linters:
  settings:
    depguard:
      rules:
        # No Ginkgo/Gomega (use standard tests)
        no_ginkgo:
          list-mode: lax
          files:
            - $all
          deny:
            - pkg: github.com/onsi/ginkgo
              desc: "use standard Go tests"
            - pkg: github.com/onsi/gomega
              desc: "use standard Go tests"

        # Modern crypto in crypto package
        modern_crypto:
          list-mode: lax
          files:
            - "**/crypto/**/*.go"
          deny:
            - pkg: crypto/md5
              desc: "use crypto/sha256 for new code"
            - pkg: crypto/rsa
              desc: "consider crypto/ed25519 instead"

        # No direct database in handlers
        no_direct_db:
          list-mode: lax
          files:
            - "**/handler/**/*.go"
            - "**/controller/**/*.go"
          deny:
            - pkg: database/sql
              desc: "use repository pattern, don't access DB directly"
            - pkg: github.com/lib/pq
              desc: "use repository pattern, don't access DB directly"
            - pkg: github.com/go-sql-driver/mysql
              desc: "use repository pattern, don't access DB directly"
```

#### ✅ Example 5: Block Specific Package Versions

```yaml
# .golangci.yml
version: "2"

linters:
  settings:
    depguard:
      rules:
        security:
          list-mode: lax
          files:
            - $all
          deny:
            - pkg: github.com/sirupsen/logrus
              desc: "logrus has known CVEs, use zap or slog instead"
            - pkg: github.com/gin-gonic/gin
              desc: "use http.ServeMux or chi for routing"
```

## How It Works With Other Linters

### ✅ Complementary Linters

**Perfect Partners:**

- **`gci`** **/ `goimports` ** - Organize imports consistently while depguard controls what can be imported
- **`gofmt` / `gofumpt`** - Format imports and code alongside import restrictions
- **`goconst`** - Find repeated string literals (often package names) for constants

**Complete Import Control Suite:**

```yaml
linters:
  enable:
    - depguard # What packages can be imported
    - gci # How imports are formatted
    - goconst # Repeated strings for constants
    - revive # General style enforcement

  settings:
    gci:
      sections:
        - standard
        - default
        - prefix(github.com/myorg)
```

### ⚠️ Overlapping Linters

**Related but Different:**

- **`forbidigo`** - Blocks specific **function calls** (e.g., `fmt.Printf`) rather than imports
- **`importas`** - Enforces import **aliases** but doesn't block packages
- **`gomodguard`** - Controls module dependencies (go.mod) vs. import statements

**Key Differences:**

```go
// depguard blocks this:
import "github.com/pkg/errors"  // Blocked: import not allowed

// forbidigo blocks this:
import "fmt"
fmt.Printf(...)  // Blocked: function call not allowed

// importas enforces this:
import errors "github.com/pkg/errors"  // Must use this alias
```

### 🔒 No Conflicts

- **No known conflicts** - depguard operates independently on import statements
- **Can run with all other linters** safely
- **Focused scope** prevents overlapping functionality

## Practical Migration Strategy

### Step 1: Start with Lax Mode (Blacklist)

```yaml
# .golangci.yml
rules:
  deprecated:
    list-mode: lax
    deny:
      - pkg: io/ioutil
        desc: "use io or os instead"
```

### Step 2: Gradually Add More Rules

```yaml
# Add more packages to deny list
rules:
  deprecated:
    list-mode: lax
    deny:
      - pkg: io/ioutil
        desc: "use io or os instead"
      - pkg: github.com/pkg/errors
        desc: "use stdlib errors package"
```

### Step 3: Enable Strict Mode for Critical Packages

```yaml
# Add strict rules for specific packages
rules:
  deprecated:
    list-mode: lax
    deny: [...]

  internal:
    list-mode: strict
    files:
      - "**/domain/**/*.go"
    allow:
      - $gostd
      - github.com/myorg/domain
```

### Step 4: Document and Socialize

- Document allowed packages in developer docs
- Explain rationale for each restriction
- Provide migration guides for blocked packages
- Update onboarding materials

### Step 5: Monitor and Adjust

- Review depguard violations in PRs
- Adjust rules based on false positives
- Add new rules as needed
- Remove outdated restrictions

## Performance Impact

- **Minimal overhead**: Only checks import statements at parse time
- **Fast execution**: No complex analysis, just string matching
- **Negligible memory usage**: Small rule set loaded once
- **Recommended in CI/CD**: Zero performance concerns

## Common Pitfalls

**❌ Don't do this:**

```yaml
# Too restrictive - will block legitimate uses
rules:
  strict:
    list-mode: strict
    allow:
      - $gostd # Only stdlib - too restrictive!
```

**✅ Do this instead:**

```yaml
# More reasonable - allows stdlib + organization packages
rules:
  strict:
    list-mode: strict
    allow:
      - $gostd
      - github.com/myorg/**
      - github.com/prometheus/client_golang/prometheus
      - go.uber.org/zap
```

**❌ Don't do this:**

```yaml
# No description - developers won't know why
rules:
  deprecated:
    list-mode: lax
    deny:
      - pkg: github.com/pkg/errors
        # desc missing - unhelpful error message
```

**✅ Do this instead:**

```yaml
# Clear description helps developers
rules:
  deprecated:
    list-mode: lax
    deny:
      - pkg: github.com/pkg/errors
        desc: "use stdlib errors package instead (Go 1.13+)"
```

## Summary

**depguard** is a **powerful dependency control linter** that allows you to:

- ✅ Block deprecated and unmaintained packages
- ✅ Enforce architectural boundaries and layering
- ✅ Standardize library choices across teams
- ✅ Guide developers toward preferred alternatives
- ✅ Prevent security issues from vulnerable packages
- ✅ Maintain consistency in large codebases

**Recommendation**: **ENABLE** for large teams, enterprise projects, and codebases undergoing migrations. Start with lax mode and gradually add more restrictions. Use strict mode for critical packages that need tight control.

**Priority**: Medium-High for team environments, Low for solo projects

---

**Reference**: https://github.com/OpenPeeDeeP/depguard
