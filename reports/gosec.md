# gosec Linter - Comprehensive Analysis

## What the Linter Does

**gosec** is a **CRITICAL static application security testing (SAST) tool** that inspects Go source code to identify security vulnerabilities, coding mistakes, and anti-patterns that could lead to security issues. It uses Abstract Syntax Tree (AST) analysis to scan code for over 100+ security rules across 6 major vulnerability categories.

### The Problem It Detects

Go code can have **critical security vulnerabilities** that compile and run correctly but expose the system to attacks:

- **SQL injection** (G201, G202) - User input directly in SQL queries
- **Cross-site scripting (XSS)** (G203) - Unescaped data in HTML templates
- **Path traversal** (G111, G304, G305) - User input in file operations
- **Command injection** (G204) - Tainted input in shell commands
- **Hardcoded credentials** (G101) - API keys, passwords in code
- **Weak cryptography** (G401-G407) - MD5, SHA1, DES, short RSA keys
- **Insecure random** (G404) - `math/rand` instead of `crypto/rand`
- **Improper file permissions** (G301, G302, G306) - World-readable sensitive files
- **Server-Side Request Forgery (SSRF)** (G107) - URLs from user input in HTTP requests
- **DoS vulnerabilities** (G110, G112, G114) - Missing timeouts, unlimited resource consumption

### How It Works

gosec analyzes Go code in three phases:

1. **AST Parsing**: Parses Go source code into Abstract Syntax Tree
2. **Pattern Matching**: Traverses AST looking for security anti-patterns (100+ rules)
3. **Vulnerability Reporting**: Reports findings with:
   - **CWE ID** (Common Weakness Enumeration)
   - **OWASP Category** (Top 10 security risks)
   - **Severity Level** (CRITICAL, HIGH, MEDIUM, LOW)
   - **Confidence Score** (0.0-1.0, how certain it's a vulnerability)

### Security Rule Categories

| Category              | Rule Range | Description                                  |
| --------------------- | ---------- | -------------------------------------------- |
| **G1xx - Misc**       | G101-G117  | Credentials, DoS, SSRF, HTTP issues          |
| **G2xx - Injection**  | G201-G204  | SQL injection, XSS, command injection        |
| **G3xx - Filesystem** | G301-G306  | Path traversal, file permissions, temp files |
| **G4xx - Crypto**     | G401-G407  | Weak algorithms, bad TLS, insecure random    |
| **G5xx - Blocklist**  | G501-G507  | Deprecated/vulnerable package imports        |
| **G6xx - Memory**     | G601       | Memory safety (range aliasing)               |

### Examples

```go
// ❌ VULNERABLE: SQL injection via string concatenation (G202)
func GetUser(id string) (*User, error) {
    query := "SELECT * FROM users WHERE id = " + id  // gosec: G202
    return db.Query(query)
}

// ✅ SECURE: Parameterized query
func GetUser(id string) (*User, error) {
    query := "SELECT * FROM users WHERE id = ?"  // gosec: OK
    return db.Query(query, id)
}
```

```go
// ❌ VULNERABLE: Hardcoded credentials (G101)
const (
    dbPassword = "f62e5bcda4fae4f82370da0c6f20697b8f8447ef"  // gosec: G101 (high entropy)
    apiKey = "sk_live_51d5b5b8a4f8d9e1c5a0f1a2b3c4d5e6f7a8b9"  // gosec: G101 (pattern match)
)

// ✅ SECURE: Environment variables or secret management
func getConfig() (string, error) {
    return os.Getenv("DB_PASSWORD"), nil
}
```

```go
// ❌ VULNERABLE: Path traversal (G304)
func ReadUserFile(filename string) ([]byte, error) {
    return os.ReadFile(filename)  // gosec: G304 - user controls path
}

// ✅ SECURE: Validate and sanitize path
func ReadUserFile(filename string) ([]byte, error) {
    cleanPath := filepath.Clean(filename)
    if !strings.HasPrefix(cleanPath, "/safe/basepath") {
        return nil, errors.New("unsafe filename")
    }
    return os.ReadFile(cleanPath)
}
```

```go
// ❌ VULNERABLE: XSS via templates (G203)
func RenderTemplate(name string) string {
    tmpl := `<h1>Hello, {{.Name}}</h1>`
    t, _ := template.New("test").Parse(tmpl)
    t.Execute(os.Stdout, struct{ Name: name })  // gosec: G203 - unescaped
}

// ✅ SECURE: Auto-escaping templates
func RenderTemplate(name string) string {
    tmpl := `<h1>Hello, {{.Name | html}}</h1>`  // gosec: OK
    t, _ := template.New("test").Parse(tmpl)
    t.Execute(os.Stdout, struct{ Name: name })
}
```

```go
// ❌ VULNERABLE: Weak random (G404)
func GenerateToken() string {
    n := rand.Intn(1000000)  // gosec: G404 - predictable
    return fmt.Sprintf("%d", n)
}

// ✅ SECURE: Cryptographically secure random
func GenerateToken() string {
    b := make([]byte, 8)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    return hex.EncodeToString(b)
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

| Project Type               | Priority | Justification                            |
| -------------------------- | -------- | ---------------------------------------- |
| **Web Services/APIs**      | CRITICAL | User input handling, XSS, SQLi           |
| **Microservices**          | CRITICAL | Distributed security, SSRF risks         |
| **REST/GraphQL APIs**      | CRITICAL | Input validation, injection vectors      |
| **CLI Tools**              | CRITICAL | File operations, path traversal          |
| **Authentication Systems** | CRITICAL | JWT, OAuth, session management           |
| **Financial Applications** | CRITICAL | PCI DSS compliance required              |
| **Healthcare Software**    | CRITICAL | HIPAA compliance required                |
| **Government/Military**    | CRITICAL | Classified information security          |
| **Payment Processing**     | CRITICAL | Financial transactions security-critical |
| **Database Tools**         | CRITICAL | SQL injection vulnerabilities            |
| **File Handling Apps**     | CRITICAL | Path traversal, file permissions         |
| **Public Libraries/SDKs**  | HIGH     | Security reputation depends on it        |
| **SaaS Platforms**         | CRITICAL | Multi-tenant security concerns           |

**Specific Scenarios:**

**1. Web Applications & APIs**

- REST/GraphQL endpoints handling user input
- File upload/download functionality
- Form processing and validation
- Session and cookie management

**2. Security-Sensitive Operations**

- Authentication/authorization logic
- Cryptographic operations (hashing, encryption, signing)
- Random number generation (tokens, IDs, nonces)
- File I/O operations (config, uploads, downloads)
- Network communication (HTTP, WebSocket, gRPC)
- Database transactions and queries

**3. Production Systems**

- Any code deployed to production
- User-facing applications
- Services with SLA requirements
- Infrastructure-as-Code tools

**4. CI/CD Pipelines**

- Build processes and deployment automation
- Code review automation
- Security scanning integration

### ❌ Disable For:

**Specific Scenarios (with Exclusions, Not Full Disabling):**

**1. Test Files** (`_test.go`)

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [gosec]
```

**2. Generated Code**

```yaml
run:
  skip-dirs:
    - generated
    - vendor
```

**3. Documentation/Example Code**

- Code samples in README.md
- Tutorial implementations
- Example code in API documentation

**4. Security Testing Utilities**

- Penetration testing tools
- Vulnerability research code
- Intentionally vulnerable examples

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** - Detects real security vulnerabilities
- **Effort**: Medium - May require refactoring for security
- **Recommendation**: **ALWAYS ENABLE** for any production code

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    gosec:
      # Run only specific rules (whitelist)
      # Default: all rules
      include:
        - G101 # Hardcoded credentials
        - G201 # SQL injection
        - G304 # Path traversal

      # Exclude specific rules (blacklist)
      # Default: []
      exclude:
        - G104 # Errors not checked (use errcheck instead)
        - G404 # Weak random (only for test files)

      # Exclude generated code
      # Default: false
      exclude-generated: true

      # Respect #nosec comments
      # Default: false
      nosec: false

      # Custom nosec tag format
      # Default: "#nosec"
      nosec-tag: "#nosec"

      # Sort by severity
      # Default: true
      sort: true

      # Confidence level threshold (0.0-1.0)
      # Default: 0.8
      confidence: 0.8

      # Severity threshold
      # Default: "low"
      severity: "medium"
```

### `include` Option

- **Type**: `[]string`
- **Default**: All rules
- **Description**: Only scan specified rules (whitelist approach)

**Use Cases:**

- Security-focused scanning (only critical rules)
- Incremental adoption (start with high-impact rules)
- Specific vulnerability type monitoring

**Example:**

```yaml
include:
  - G101 # Credentials
  - G201 # SQL injection
  - G202 # SQL injection (concat)
  - G203 # XSS
  - G204 # Command injection
  - G304 # Path traversal
```

### `exclude` Option

- **Type**: `[]string`
- **Default**: `[]`
- **Description**: Exclude specific rules from scanning

**Common Exclusions:**

```yaml
exclude:
  - G104 # Errors not checked (let errcheck handle)
  - G404 # Weak random (math/rand in tests)
  - G102 # Bind to all interfaces (K8s requirement)
```

### `exclude-generated` Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: Skip auto-generated code

**Generated Patterns Recognized:**

- Protocol buffer generated files (`*_pb.go`)
- Mockery generated files
- Wire dependency injection code
- Stringer generated code

### `confidence` Option

- **Type**: `float64`
- **Default**: `0.8`
- **Range**: `0.0-1.0`
- **Description**: Only report findings with confidence >= threshold

**Confidence Levels:**
| Confidence | Meaning | Action |
|------------|----------|--------|
| 0.9-1.0 | Almost certainly a vulnerability | Fix immediately |
| 0.7-0.9 | Likely a vulnerability | Review and fix |
| 0.5-0.7 | Possibly a vulnerability | Review manually |
| 0.0-0.5 | Might be a vulnerability | Consider fixing |

### `severity` Option

- **Type**: `string`
- **Default**: `"low"`
- **Values**: `"low"`, `"medium"`, `"high"`
- **Description**: Only report findings with severity >= threshold

**Severity Classification:**
| Level | Description | Example Rules |
|-------|-------------|---------------|
| CRITICAL | Direct security exploit, remote code execution | G201, G202, G204 |
| HIGH | Credential exposure, data breach | G101, G304, G111 |
| MEDIUM | DoS vectors, weak crypto | G110, G112, G401 |
| LOW | Minor security issues, code quality | G103, G104 |

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# Most production applications
version: "2"
linters:
  settings:
    gosec:
      exclude-generated: true
      confidence: 0.8
      severity: "medium"

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [gosec]
```

#### ✅ Strict Configuration

```yaml
# Security-critical applications, financial/healthcare software
version: "2"
linters:
  settings:
    gosec:
      exclude-generated: true
      confidence: 0.7
      severity: "low"
      exclude: []

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [gosec]
```

#### ✅ Focused Configuration (Critical Rules Only)

```yaml
# Focus on highest-impact vulnerabilities
version: "2"
linters:
  settings:
    gosec:
      include:
        - G101 # Credentials
        - G201 # SQL injection
        - G202 # SQL injection
        - G203 # XSS
        - G204 # Command injection
        - G304 # Path traversal
        - G404 # Weak random
```

#### ✅ Web API Configuration

```yaml
# HTTP server with user input handling
version: "2"
linters:
  settings:
    gosec:
      exclude-generated: true
      include:
        - G101 # Credentials
        - G107 # SSRF
        - G111 # Path traversal (http.Dir)
        - G201 # SQL injection
        - G202 # SQL injection
        - G203 # XSS
        - G204 # Command injection
        - G304 # Path traversal
      exclude:
        - G102 # Allow 0.0.0.0 for K8s

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [gosec]
```

#### ✅ Exclude Test Files

```yaml
# Tests have intentional false positives (mocks, test tokens)
version: "2"
linters:
  settings:
    gosec:
      exclude-generated: true
      exclude:
        - G104 # Errors not checked in tests
        - G404 # math/rand in tests

run:
  skip-dirs:
    - vendor

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [gosec]
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

gosec works excellently with:

| Linter          | Relationship  | Value                                                       |
| --------------- | ------------- | ----------------------------------------------------------- |
| **errcheck**    | Complementary | gosec G104 + errcheck = complete error handling             |
| **staticcheck** | Complementary | gosec finds security issues, staticcheck finds logic errors |
| **govet**       | Complementary | gosec finds high-level issues, govet finds AST-level        |
| **bidichk**     | Complementary | gosec G116 (Trojan Source) + bidichk (general BiDi)         |
| **bodyclose**   | Complementary | gosec DoS (G110) + bodyclose (resource leaks)               |
| **nilerr**      | Complementary | gosec G104 errors + nilerr nil pointer returns              |
| **noctx**       | Complementary | gosec G107 SSRF + noctx (missing context)                   |
| **errchkjson**  | Complementary | JSON marshaling security + JSON type safety                 |

**Complete Security Suite:**

```yaml
linters:
  enable:
    - gosec # Security vulnerabilities (CRITICAL)
    - errcheck # Error handling (CRITICAL)
    - staticcheck # Logic errors (CRITICAL)
    - govet # AST checks (CRITICAL)
    - bidichk # Unicode security (HIGH)
    - bodyclose # Resource leaks (HIGH)
    - nilerr # Nil error returns (CRITICAL)
    - noctx # HTTP security (CRITICAL)
```

### Example of Linter Synergy

```go
// gosec catches:
func processUserInput(input string) error {
    query := "SELECT * FROM users WHERE name = '" + input + "'"
    db.Exec(query)  // gosec: G202 - SQL injection
    return nil
}

// errcheck adds (after fixing SQLi):
func processUserInput(input string) error {
    query := "SELECT * FROM users WHERE name = ?"
    _, err := db.Exec(query, input)  // errcheck: error not checked
    return err
}

// staticcheck adds (after fixing errcheck):
func processUserInput(input string) error {
    query := "SELECT * FROM users WHERE name = ?"
    result, err := db.Exec(query, input)
    if err != nil {
        return err  // errcheck satisfied
    }
    rows, err := result.RowsAffected()
    if err != nil {  // staticcheck: unnecessary check
        return err
    }
    return nil
}
```

### 🔒 Minimal Overlap, No Conflicts

| Linter                  | Overlap                      | Recommendation                       |
| ----------------------- | ---------------------------- | ------------------------------------ |
| gosec + **errcheck**    | G104 overlaps error checking | Use both - gosec is security-focused |
| gosec + **staticcheck** | Some error analysis overlap  | Use both - different focus areas     |
| gosec + **govet**       | Different AST analysis       | Use both - comprehensive coverage    |

**Why No Conflicts:**

- gosec: Security-focused vulnerability detection
- errcheck: Error handling completeness
- staticcheck: Logic and correctness analysis
- Each linter covers different aspects of code quality

## Practical Examples

### ✅ Example 1: SQL Injection Protection

```go
// ❌ VULNERABLE: String concatenation in SQL
func FindUser(username string) (*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
    row := db.QueryRow(query)  // gosec: G201
    var user User
    row.Scan(&user.ID, &user.Name)
    return &user, nil
}

// ✅ SECURE: Parameterized query
func FindUser(username string) (*User, error) {
    query := "SELECT * FROM users WHERE username = ?"
    row := db.QueryRow(query, username)  // gosec: OK
    var user User
    err := row.Scan(&user.ID, &user.Name)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

### ✅ Example 2: XSS Prevention

```go
// ❌ VULNERABLE: Unescaped user input in HTML
func RenderUserComment(comment string) string {
    tmpl := `<div class="comment">{{.Comment}}</div>`
    t, _ := template.New("comment").Parse(tmpl)
    t.Execute(os.Stdout, struct{ Comment: comment })  // gosec: G203
    return comment
}

// ✅ SECURE: Auto-escaping template
func RenderUserComment(comment string) string {
    tmpl := `<div class="comment">{{.Comment | html}}</div>`  // gosec: OK
    t, _ := template.New("comment").Parse(tmpl)
    t.Execute(os.Stdout, struct{ Comment: comment })
    return comment
}
```

### ✅ Example 3: Path Traversal Prevention

```go
// ❌ VULNERABLE: User input in file path
func ServeFile(filename string, w http.ResponseWriter) {
    content, err := os.ReadFile(filename)  // gosec: G304
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.Write(content)
}

// ✅ SECURE: Validate and sanitize path
func ServeFile(filename string, w http.ResponseWriter) {
    cleanPath := filepath.Clean(filename)
    if strings.Contains(cleanPath, "..") {
        http.Error(w, "Invalid filename", 400)
        return
    }
    if !strings.HasPrefix(cleanPath, "/safe/files") {
        http.Error(w, "Access denied", 403)
        return
    }
    content, err := os.ReadFile(cleanPath)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.Write(content)
}
```

### ✅ Example 4: Secure Random Generation

```go
// ❌ VULNERABLE: Predictable random numbers
func GenerateSessionToken() string {
    token := rand.Intn(1000000000)  // gosec: G404
    return fmt.Sprintf("%d", token)
}

// ✅ SECURE: Cryptographically secure random
func GenerateSessionToken() string {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {  // gosec: OK
        return ""
    }
    return hex.EncodeToString(b)
}
```

### ✅ Example 5: Secure Cryptography

```go
// ❌ VULNERABLE: Weak hash algorithm
func HashPassword(password string) string {
    hash := sha1.Sum([]byte(password))  // gosec: G501, G401
    return hex.EncodeToString(hash)
}

// ✅ SECURE: Strong hash algorithm
func HashPassword(password string) string {
    hash := sha256.Sum256([]byte(password))  // gosec: OK
    return hex.EncodeToString(hash)
}
```

## Best Practices

1. **ALWAYS enable gosec** for any production code
2. **Set severity to "medium"** for production CI/CD
3. **Exclude test files** (`(.+)_test\.go`) - intentional false positives
4. **Exclude generated code** (`_pb.go`, `_generated.go`) - not your code
5. **Use path-based exclusions** over `#nosec` comments
6. **Set confidence threshold appropriately** - 0.8 is good balance
7. **Review findings seriously** - gosec detects real vulnerabilities
8. **Fix CRITICAL and HIGH severity issues first**
9. **Integrate with CI/CD** - block merges with new HIGH/CRITICAL issues
10. **Combine with complementary linters** - errcheck, staticcheck, govet
11. **Educate team** - explain why security issues matter
12. **Keep gosec updated** - new rules added as vulnerabilities discovered

## Common Scenarios and Solutions

### Scenario 1: Too Many False Positives

**Problem:** gosec reporting many issues in test files or generated code.

**Solution:** Exclude test files and generated code.

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [gosec]

run:
  skip-dirs:
    - generated
    - vendor
```

### Scenario 2: G104 (Errors Not Checked) vs errcheck

**Problem:** Both gosec and errcheck reporting errors not checked.

**Solution:** Let errcheck handle all error checking.

```yaml
linters:
  settings:
    gosec:
      exclude:
        - G104 # Let errcheck handle
```

### Scenario 3: G404 (Weak Random) in Tests

**Problem:** Tests use `math/rand` for reproducibility, gosec flags as weak.

**Solution:** Exclude test files.

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [gosec]
```

### Scenario 4: Security-Critical Path

**Problem:** Must ensure ALL security issues caught, no false negatives.

**Solution:** Use strictest settings.

```yaml
linters:
  settings:
    gosec:
      severity: "low"
      confidence: 0.7
      exclude: []
```

### Scenario 5: CI/CD Integration

**Problem:** Block merges with new security vulnerabilities.

**Solution:** Use severity threshold and fail on HIGH/CRITICAL.

```bash
# In CI/CD
golangci-lint run --enable=gosec --out-format=json | \
  jq '[.Severity == "HIGH" or .Severity == "CRITICAL"] | length' | \
  xargs -I {} sh -c 'if [ {} -gt 0 ]; then exit 1; fi'
```

## Summary

**gosec** is a **CRITICAL security linter** that detects real vulnerabilities:

- ✅ **Highest priority** - Security is critical for all production code
- ✅ **Comprehensive coverage** - 100+ rules across 6 categories
- ✅ **Real vulnerabilities** - Not just code style issues
- ✅ **CWE/OWASP mapping** - Industry-standard classification
- ✅ **Configurable** - Adjust strictness by severity and confidence
- ✅ **CI/CD integration** - Block merges with new vulnerabilities
- ⚠️ **May need exclusions** - Test files, generated code
- ⚠️ **Performance overhead** - Slower than other linters (2-5 min for medium projects)
- ⚠️ **Requires expertise** - Need to understand security implications

**Recommendation:** **ALWAYS ENABLE** with `severity: "medium"` and `exclude-generated: true` for all production code. Exclude test files (`(.+)_test\.go`). Combine with **errcheck**, **staticcheck**, and **govet** for comprehensive coverage. For security-critical applications (financial, healthcare), use `severity: "low"` to catch ALL issues. Integrate with CI/CD to block merges with new HIGH/CRITICAL severity findings.

**Top 3 Security Rules to Never Disable:**

1. **G101** (Hardcoded credentials) - Major security risk
2. **G201/G202** (SQL injection) - Critical data breach risk
3. **G304** (Path traversal) - File system security risk

---

**Reference:** https://github.com/securego/gosec
