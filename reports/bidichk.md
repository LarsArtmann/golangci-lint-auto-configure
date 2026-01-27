# bidichk Linter - Comprehensive Analysis

## What the Linter Does

**`bidichk`** (github.com/breml/bidichk) detects dangerous bidirectional Unicode control characters in Go source code that can be exploited for "Trojan Source" attacks. These invisible characters can alter the visual rendering of code, making malicious code appear benign to human reviewers while being executed differently by the compiler.

### The Security Risk It Mitigates

The linter catches the presence of bidirectional (BiDi) control characters that can be weaponized:

```go
// What reviewers see (visually):
func isAdmin() bool {
    return false /* check privileges */
}

// What actually executes (contains U+202E):
func isAdmin() bool {
    return false /* check privileges */ ;return true
}
```

In this example, the RIGHT-TO-LEFT-OVERRIDE (U+202E) character causes everything after it to be rendered right-to-left, hiding the malicious `;return true` that actually executes.

### Unicode Characters Detected

| Unicode | Name | Short Code | Security Impact |
|---------|------|------------|-----------------|
| U+202A | LEFT-TO-RIGHT-EMBEDDING | LRE | Alters text flow direction |
| U+202B | RIGHT-TO-LEFT-EMBEDDING | RLE | Alters text flow direction |
| U+202C | POP-DIRECTIONAL-FORMATTING | PDF | Resets direction to default |
| U+202D | LEFT-TO-RIGHT-OVERRIDE | LRO | Forces LTR direction |
| U+202E | RIGHT-TO-LEFT-OVERRIDE | RLO | Forces RTL direction (most dangerous) |
| U+2066 | LEFT-TO-RIGHT-ISOLATE | LRI | Isolates LTR text |
| U+2067 | RIGHT-TO-LEFT-ISOLATE | RLI | Isolates RTL text |
| U+2068 | FIRST-STRONG-ISOLATE | FSI | Isolates based on first strong character |
| U+2069 | POP-DIRECTIONAL-ISOLATE | PDI | Ends isolation |

### Key Characteristics
- **Type**: Security-focused static analysis linter
- **Speed**: Very fast (linear scan of source files)
- **Integration**: Part of golangci-lint v2+
- **Repository**: https://github.com/breml/bidichk
- **Category**: Trojan Source detection / Supply chain security

## When It Should Be Enabled

### ✅ MANDATORY FOR:

**Security-Critical Projects:**
- Cryptocurrency and blockchain applications
- Authentication/authorization systems
- Payment processing and financial software
- Security libraries and frameworks
- Code that handles sensitive data (PII, credentials)

**Open Source Projects:**
- Public repositories on GitHub/GitLab
- Projects accepting external contributions
- Libraries consumed by other projects
- Projects with multiple maintainers

**Enterprise/Controlled Environments:**
- Codebases requiring audit compliance
- Projects under regulatory oversight (SOC2, ISO27001)
- Government or defense-related software
- Healthcare applications (HIPAA)

**CI/CD & Code Review:**
- All projects using automated code review
- Organizations with formal security review processes
- Projects requiring signed commits/PRs

### PRIORITY ASSESSMENT:
- **Default Priority**: **HIGH** (Security)
- **Should be enabled**: For 99% of Go projects
- **Exception cases**: Only specific internationalization scenarios

## When It Should Be Disabled

### ❌ CONSIDER DISABLING WHEN:

**Legitimate BiDi Text Processing:**
```go
// Working with RTL language content intentionally
const arabicMessage = "مرحبا" // Contains RTL characters legitimately
```

**Internationalization Libraries:**
- Projects specifically handling Arabic, Hebrew, Persian text
- BiDi algorithm implementations
- Unicode text processing libraries

**Generated Code:**
- Protobuf/thrift generated code with embedded comments
- Swagger/OpenAPI generated clients
- ORM generated code
- Code from code generators that may include BiDi chars

**Documentation-Heavy Projects:**
- Projects with extensive RTL language documentation
- Comments and strings containing legitimate BiDi text

### Better Alternative to Disabling:

Use **exclusions** instead of disabling entirely:

```yaml
# .golangci.yml
linters:
  exclusions:
    rules:
      # Exclude specific files that legitimately use BiDi
      - linters: [bidichk]
        path: internal/i18n/

      # Exclude generated code
      - linters: [bidichk]
        path: (.+)_generated\.go

      # Exclude specific patterns
      - linters: [bidichk]
        path: pkg/rtl_text_processor/
```

Or use targeted `nolint` directives:

```go
// Package arabic provides Arabic text processing
//nolint:bidichk // This package intentionally handles RTL text
package arabic
```

## Configuration Options

### Basic Configuration Structure:

```yaml
# .golangci.yml (version 2+)
version: "2"
linters:
  enable:
    - bidichk
  settings:
    bidichk:
      # Individual character controls (default: all enabled)
      left-to-right-embedding: true
      right-to-left-embedding: true
      pop-directional-formatting: true
      left-to-right-override: true
      right-to-left-override: true
      left-to-right-isolate: true
      right-to-left-isolate: true
      first-strong-isolate: true
      pop-directional-isolate: true
```

### Configuration Options Explained:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `left-to-right-embedding` | `bool` | `true` | Enable detection of U+202A (LRE) |
| `right-to-left-embedding` | `bool` | `true` | Enable detection of U+202B (RLE) |
| `pop-directional-formatting` | `bool` | `true` | Enable detection of U+202C (PDF) |
| `left-to-right-override` | `bool` | `true` | Enable detection of U+202D (LRO) |
| `right-to-left-override` | `bool` | `true` | Enable detection of U+202E (RLO) |
| `left-to-right-isolate` | `bool` | `true` | Enable detection of U+2066 (LRI) |
| `right-to-left-isolate` | `bool` | `true` | Enable detection of U+2067 (RLI) |
| `first-strong-isolate` | `bool` | `true` | Enable detection of U+2068 (FSI) |
| `pop-directional-isolate` | `bool` | `true` | Enable detection of U+2069 (PDI) |

### Recommended Configurations:

**Default (Check All Dangerous Characters):**
```yaml
linters:
  enable:
    - bidichk
  # No settings needed - checks all characters by default
```

**Paranoid Security (Explicitly Enable All):**
```yaml
linters:
  settings:
    bidichk:
      left-to-right-embedding: true
      right-to-left-embedding: true
      pop-directional-formatting: true
      left-to-right-override: true
      right-to-left-override: true
      left-to-right-isolate: true
      right-to-left-isolate: true
      first-strong-isolate: true
      pop-directional-isolate: true
```

**Minimal (Check Only Most Dangerous):**
```yaml
linters:
  settings:
    bidichk:
      # Only check override characters (most dangerous)
      left-to-right-override: true
      right-to-left-override: true
```

## How It Interferes or Works Together With Other Linters

### ✅ SYNERGIES:

**Security Linters:**
-  **`gosec`**  : bidichk catches BiDi attacks while gosec catches other security issues (SQL injection, XSS, etc.)
-  **`exportloopref`**  : Both protect against subtle bugs that can be exploited
-  **`nilerr`**  : Combined security posture - catch error handling bypasses

**Code Quality:**
-  **`govet`**  : Complements govet's suspicious construct checks
-  **`staticcheck`**  : Advanced static analysis + BiDi security = comprehensive coverage
-  **`errcheck`**  : Security requires proper error handling

**Example Security Workflow:**
```go
// gosec checks for hardcoded credentials
// errcheck ensures errors aren't ignored
// bidichk ensures no hidden code via BiDi chars
func connectToDB(connectionString string) (*sql.DB, error) {
    db, err := sql.Open("mysql", connectionString)
    if err != nil {
        return nil, err // errcheck validates this
    }

    // gosec flags if credentials are hardcoded above
    return db, nil
}
```

### 🔒 CONFLICTS:

**No known conflicts** - bidichk operates purely on source code character level

**No overlaps** with other linters' functionality - unique security focus

### 📊 PERFORMANCE IMPACT:

- **Minimal**: Simple linear scan of source files
- **Negligible overhead**: No AST parsing required
- **Recommended in CI/CD**: Zero impact on build times
- **Memory**: Constant memory usage

## Practical Examples

### Real-World Attack Caught by bidichk:

```go
// Malicious code with U+202E embedded:
func validateUser(userID string) bool {
    // Check if user is authorized‮⁩/*‮ } ⁦if 0 != 0 {
    return false
    // */return true;
}
```

**What this looks like visually** (comments appear to disable the return):
```go
func validateUser(userID string) bool {
    // Check if user is authorized/* } if 0 != 0 {
    return false
    // */return true;
}
```

**What actually executes** (U+202E reverses text rendering):
```go
func validateUser(userID string) bool {
    // Check if user is authorized
    return true; /* } if 0 != 0 {
    return false
    // */
}
```

### False Positive Handling:

```go
// Legitimate use in i18n package
const welcomeMessage = "Welcome ‫مَرْحَبًا" // Arabic greeting

// Solution 1: nolint directive
const welcomeMessage = "Welcome ‫مَرْحَبًا" //nolint:bidichk // Arabic RTL text

// Solution 2: Move to excluded file/directory
// Solution 3: Use escape sequences instead of raw characters
const welcomeMessage = "Welcome \u202Bمَرْحَبًا\u202C"
```

### Configuration for Mixed Projects:

```yaml
# For projects with legitimate BiDi usage
linters:
  exclusions:
    rules:
      # Exclude i18n packages entirely
      - linters: [bidichk]
        path: pkg/i18n/

      # Exclude specific files
      - linters: [bidichk]
        path: internal/locale/messages.go

      # Exclude test files (often contain examples)
      - linters: [bidichk]
        path: (.+)_test\.go
```

## Summary

**bidichk** is a **critical security linter** that protects against Trojan Source attacks via bidirectional Unicode control characters. It is:

- **Essential** for security-conscious projects
- **Zero-cost** in terms of performance
- **Simple** to configure and use
- **Unique** - no other linter provides this protection
- **Complementary** to all other linters

**Recommendation**: **ENABLE** for all Go projects by default. Only exclude specific files/directories that legitimately require BiDi text handling. This linter represents a crucial defense-in-depth measure for modern software supply chain security.
