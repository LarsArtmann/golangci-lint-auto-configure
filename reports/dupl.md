# dupl Linter - Comprehensive Analysis

## What the Linter Does

**dupl** detects **duplicate code fragments** in Go source code. It analyzes code to find sequences of tokens that appear in multiple places, which could potentially be refactored into reusable functions or structs. This helps maintain code quality, reduce bugs (by fixing issues in one place instead of many), and improve maintainability.

### The Problem It Detects

Code duplication leads to several problems:
- **Maintenance burden**: Bugs must be fixed in multiple places
- **Inconsistent fixes**: Changes might be applied inconsistently across duplicates
- **Increased code size**: Larger codebases are harder to understand and navigate
- **Hidden bugs**: Similar logic might behave differently in different locations
- **Technical debt**: Accumulates over time as more code is copy-pasted

### How It Works

dupl uses **AST-based tokenization** to analyze Go code:

1. **AST Tokenization**: Parses Go source code and converts it into a sequence of tokens from the Abstract Syntax Tree
2. **Suffixed Tree Construction**: Builds a suffixed tree data structure for efficient sequence matching
3. **Similarity Detection**: Finds sequences that share the same token pattern above the configured threshold
4. **Reporting**: Reports file locations, line numbers, and token counts of duplicate sequences

**Default threshold**: 150 tokens

### Examples

```go
// ❌ BAD: Duplicate code found (detected by dupl)

// File: user_service.go
func createUser(name, email string) (*User, error) {
    if name == "" {
        return nil, errors.New("name cannot be empty")
    }
    if email == "" {
        return nil, errors.New("email cannot be empty")
    }
    if !isValidEmail(email) {
        return nil, errors.New("invalid email format")
    }
    return &User{Name: name, Email: email}, nil
}

// File: product_service.go
func createProduct(name, sku string) (*Product, error) {
    if name == "" {
        return nil, errors.New("name cannot be empty")
    }
    if sku == "" {
        return nil, errors.New("sku cannot be empty")
    }
    if !isValidSKU(sku) {
        return nil, errors.New("invalid sku format")
    }
    return &Product{Name: name, SKU: sku}, nil
}

// ✅ GOOD: Refactored to remove duplication
func validateField(fieldName, value string) error {
    if value == "" {
        return fmt.Errorf("%s cannot be empty", fieldName)
    }
    return nil
}

func createUser(name, email string) (*User, error) {
    if err := validateField("name", name); err != nil {
        return nil, err
    }
    if err := validateField("email", email); err != nil {
        return nil, err
    }
    if !isValidEmail(email) {
        return nil, errors.New("invalid email format")
    }
    return &User{Name: name, Email: email}, nil
}

func createProduct(name, sku string) (*Product, error) {
    if err := validateField("name", name); err != nil {
        return nil, err
    }
    if err := validateField("sku", sku); err != nil {
        return nil, err
    }
    if !isValidSKU(sku) {
        return nil, errors.New("invalid sku format")
    }
    return &Product{Name: name, SKU: sku}, nil
}
```

```go
// ❌ BAD: Duplicate error handling patterns

func processPayment(orderID string) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    // ... payment processing ...

    if err != nil {
        tx.Rollback()
        return fmt.Errorf("payment processing failed: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    return nil
}

func processRefund(orderID string) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    // ... refund processing ...

    if err != nil {
        tx.Rollback()
        return fmt.Errorf("refund processing failed: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    return nil
}

// ✅ GOOD: Extract transaction handling
func withTransaction(fn func(tx *sql.Tx) error) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    return nil
}

func processPayment(orderID string) error {
    return withTransaction(func(tx *sql.Tx) error {
        // ... payment processing ...
        return err
    })
}

func processRefund(orderID string) error {
    return withTransaction(func(tx *sql.Tx) error {
        // ... refund processing ...
        return err
    })
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **Large codebases** (>10K lines) - Duplication accumulates over time
- **Team projects** (>3 developers) - Multiple contributors increase copy-paste risk
- **Long-lived projects** (>1 year) - Code patterns repeat over time
- **Libraries and APIs** - Public APIs benefit from consistent internal code
- **Enterprise applications** - Maintaining code quality is critical

**Scenarios:**
- **Code review process** - Catch duplicates before merge
- **Refactoring initiatives** - Identify opportunities for consolidation
- **Technical debt reduction** - Systematically eliminate duplication
- **Quality gates** - Prevent new duplicates in CI/CD
- **Onboarding new developers** - Encourage code reuse patterns

**Development Phases:**
- **Maintenance phase** - When codebase is stable and needs cleanup
- **Codebase reviews** - Periodic audits for technical debt
- **Feature development** - Prevent new duplicates from entering
- **Architecture redesign** - Find patterns that inform better abstractions

### ❌ Disable For:

**Project Types:**
- **Small utilities** (<1K lines) - Overhead not justified
- **Proof-of-concept code** - Speed over maintainability
- **Prototypes** - Code likely to change significantly
- **Single-developer projects** - Less risk of inconsistent patterns
- **Scripts and one-offs** - Duplication may be intentional

**Specific Scenarios:**
- **Test files** - Tests naturally have repetitive structures
- **Generated code** - Auto-generated code will always have patterns
- **Migration scripts** - One-time code where duplication is acceptable
- **Rapid prototyping** - Speed priority over quality
- **Legacy code refactoring** - Too many violations to fix at once

**Development Phases:**
- **Initial development** - Focus on functionality first
- **Spike solutions** - Experimental code where duplication is expected
- **Emergency fixes** - Speed is critical over code quality

### Priority Assessment

- **Default Priority**: Optional (not in linter_data.go)
- **Value**: Medium-High - Improves maintainability and reduces bugs
- **Effort**: Medium - Requires manual review and refactoring
- **Recommendation**: **Enable** for medium-to-large projects with threshold of 100-150

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    dupl:
      # Minimum tokens for duplicate detection (default: 150)
      threshold: 100
```

### `threshold` Option

- **Type**: `int`
- **Default**: `150`
- **Range**: `50-500+` (typical values)
- **Description**: Minimum number of tokens that must be duplicated for dupl to report an issue

**Behavior:**
- Lower threshold (50-100): Detects more duplicates but may have false positives
- Default threshold (150): Balanced detection rate
- Higher threshold (200+): Only reports substantial duplicates, fewer false positives

**Guidelines:**
- **100 tokens**: Good starting point for most projects
- **150 tokens**: Default, works well for many codebases
- **200+ tokens**: For large projects where only major duplicates matter

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Most projects - balanced detection
version: "2"
linters:
  settings:
    dupl:
      threshold: 100  # Detect medium to large duplicates
```

#### ✅ Strict Configuration
```yaml
# High-quality projects, fewer false positives acceptable
version: "2"
linters:
  settings:
    dupl:
      threshold: 50  # Detect even small duplicates
```

#### ✅ Relaxed Configuration
```yaml
# Large legacy codebases, focus on major issues
version: "2"
linters:
  settings:
    dupl:
      threshold: 200  # Only flag substantial duplicates
```

#### ✅ Exclude Test Files
```yaml
# Don't flag test duplication (intentionally repetitive)
version: "2"
linters:
  settings:
    dupl:
      threshold: 100

issues:
  exclude-rules:
    - linter: dupl
      path: (.+)_test\.go
```

#### ✅ Exclude Generated Code
```yaml
# Skip generated files (will always have patterns)
version: "2"
linters:
  settings:
    dupl:
      threshold: 100

run:
  skip-dirs:
    - generated
    - vendor

issues:
  exclude-rules:
    - linter: dupl
      path: (.+)_generated\.go
    - linter: dupl
      path: (.+)_pb\.go
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

dupl works well alongside:

| Linter | Relationship | Benefit |
|--------|--------------|---------|
| **`gocritic`** | Complementary | gocritic has code simplification checks; dupl finds actual duplicates |
| **`goconst`** | Complementary | goconst finds duplicate strings; dupl finds duplicate code sequences |
| **`ineffassign`** | Complementary | Both detect code quality issues |
| **`nolintlint`** | Process improvement | If you frequently use `//nolint:dupl`, consider raising threshold |
| **`revive`** | Compatible | General style linter vs specific duplicate detection |

**Complete Code Quality Suite:**
```yaml
linters:
  enable:
    - dupl          # Duplicate code detection
    - goconst       # Duplicate strings
    - gocritic      # Code simplifications
    - ineffassign   # Ineffective assignments
    - revive        # General code quality

  settings:
    dupl:
      threshold: 100
    goconst:
      min-len: 2
      min-occurrences: 2
```

### 🔒 No Conflicts

- **No known conflicts** - dupl focuses specifically on code duplication
- **Independent operation** - Doesn't overlap with other linter functionality
- **Safe to enable with all linters**

## Practical Examples

### ✅ Example 1: Database Query Patterns

```go
// ❌ BAD: Duplicate query construction
func getUserByID(id int) (*User, error) {
    query := `
        SELECT id, name, email, created_at
        FROM users
        WHERE id = ?
    `
    row := db.QueryRow(query, id)
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    return &user, err
}

func getUserByEmail(email string) (*User, error) {
    query := `
        SELECT id, name, email, created_at
        FROM users
        WHERE email = ?
    `
    row := db.QueryRow(query, email)
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    return &user, err
}

// ✅ GOOD: Extract query function
func queryUser(where string, args ...interface{}) (*User, error) {
    query := fmt.Sprintf(`
        SELECT id, name, email, created_at
        FROM users
        WHERE %s
    `, where)
    row := db.QueryRow(query, args...)
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
    return &user, err
}

func getUserByID(id int) (*User, error) {
    return queryUser("id = ?", id)
}

func getUserByEmail(email string) (*User, error) {
    return queryUser("email = ?", email)
}
```

### ✅ Example 2: HTTP Handler Patterns

```go
// ❌ BAD: Duplicate error response handling
func getUserHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    if userID == "" {
        http.Error(w, "user ID is required", http.StatusBadRequest)
        return
    }

    user, err := getUser(userID)
    if err != nil {
        http.Error(w, fmt.Sprintf("failed to get user: %v", err), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
    productID := r.URL.Query().Get("id")
    if productID == "" {
        http.Error(w, "product ID is required", http.StatusBadRequest)
        return
    }

    product, err := getProduct(productID)
    if err != nil {
        http.Error(w, fmt.Sprintf("failed to get product: %v", err), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(product)
}

// ✅ GOOD: Extract common patterns
func respondJSON(w http.ResponseWriter, data interface{}) {
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, status int) {
    http.Error(w, message, status)
}

func getQueryParam(r *http.Request, name string) string {
    return r.URL.Query().Get(name)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    userID := getQueryParam(r, "id")
    if userID == "" {
        respondError(w, "user ID is required", http.StatusBadRequest)
        return
    }

    user, err := getUser(userID)
    if err != nil {
        respondError(w, fmt.Sprintf("failed to get user: %v", err), http.StatusInternalServerError)
        return
    }

    respondJSON(w, user)
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
    productID := getQueryParam(r, "id")
    if productID == "" {
        respondError(w, "product ID is required", http.StatusBadRequest)
        return
    }

    product, err := getProduct(productID)
    if err != nil {
        respondError(w, fmt.Sprintf("failed to get product: %v", err), http.StatusInternalServerError)
        return
    }

    respondJSON(w, product)
}
```

### ✅ Example 3: Validation Logic

```go
// ❌ BAD: Repeated validation patterns
type CreateUserRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (r *CreateUserRequest) Validate() error {
    if r.Name == "" {
        return errors.New("name is required")
    }
    if len(r.Name) < 3 {
        return errors.New("name must be at least 3 characters")
    }
    if r.Email == "" {
        return errors.New("email is required")
    }
    if !isValidEmail(r.Email) {
        return errors.New("invalid email format")
    }
    if r.Password == "" {
        return errors.New("password is required")
    }
    if len(r.Password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    return nil
}

type CreateProductRequest struct {
    Name  string `json:"name"`
    SKU   string `json:"sku"`
    Price string `json:"price"`
}

func (r *CreateProductRequest) Validate() error {
    if r.Name == "" {
        return errors.New("name is required")
    }
    if len(r.Name) < 3 {
        return errors.New("name must be at least 3 characters")
    }
    if r.SKU == "" {
        return errors.New("sku is required")
    }
    if len(r.SKU) < 3 {
        return errors.New("sku must be at least 3 characters")
    }
    if r.Price == "" {
        return errors.New("price is required")
    }
    return nil
}

// ✅ GOOD: Extract validation helpers
type fieldValidator struct {
    name  string
    value string
}

func (fv *fieldValidator) Required() *fieldValidator {
    if fv.value == "" {
        return &fieldValidator{err: fmt.Sprintf("%s is required", fv.name)}
    }
    return fv
}

func (fv *fieldValidator) MinLength(min int) *fieldValidator {
    if len(fv.value) < min {
        return &fieldValidator{err: fmt.Sprintf("%s must be at least %d characters", fv.name, min)}
    }
    return fv
}

func (fv *fieldValidator) Error() error {
    return fv.err
}

func validateField(name, value string) *fieldValidator {
    return &fieldValidator{name: name, value: value}
}

func (r *CreateUserRequest) Validate() error {
    if err := validateField("name", r.Name).Required().MinLength(3).Error(); err != nil {
        return err
    }
    if err := validateField("email", r.Email).Required().Error(); err != nil {
        return err
    }
    if !isValidEmail(r.Email) {
        return errors.New("invalid email format")
    }
    if err := validateField("password", r.Password).Required().MinLength(8).Error(); err != nil {
        return err
    }
    return nil
}

func (r *CreateProductRequest) Validate() error {
    if err := validateField("name", r.Name).Required().MinLength(3).Error(); err != nil {
        return err
    }
    if err := validateField("sku", r.SKU).Required().MinLength(3).Error(); err != nil {
        return err
    }
    if err := validateField("price", r.Price).Required().Error(); err != nil {
        return err
    }
    return nil
}
```

### ✅ Example 4: Test Files (Legitimate Duplication)

```go
// Tests often have similar structure - use //nolint:dupl

func TestUser_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   User
        want    *User
        wantErr bool
    }{
        {"valid user", User{Name: "John", Email: "john@example.com"}, &User{Name: "John", Email: "john@example.com"}, false},
        {"empty name", User{Email: "john@example.com"}, nil, true},
        {"invalid email", User{Name: "John", Email: "invalid"}, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CreateUser(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("CreateUser() = %v, want %v", got, tt.want)
            }
        })
    }
} //nolint:dupl  // Test structure is intentionally repetitive

func TestProduct_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   Product
        want    *Product
        wantErr bool
    }{
        {"valid product", Product{Name: "Widget", SKU: "WID-001"}, &Product{Name: "Widget", SKU: "WID-001"}, false},
        {"empty name", Product{SKU: "WID-001"}, nil, true},
        {"invalid sku", Product{Name: "Widget", SKU: "invalid"}, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CreateProduct(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateProduct() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("CreateProduct() = %v, want %v", got, tt.want)
            }
        })
    }
} //nolint:dupl  // Test structure is intentionally repetitive
```

### ✅ Example 5: Configuration Tuning

```go
// Scenario: Large legacy codebase with many duplicates

// Initial run shows 5000+ duplicate findings
// Too many to fix - need to adjust threshold

// Step 1: Increase threshold to focus on major issues
// .golangci.yml
version: "2"
linters:
  settings:
    dupl:
      threshold: 200  # Focus on substantial duplicates

// Step 2: Fix the largest duplicates first
// After fixing 50 major duplicates, reduce threshold
threshold: 150  # Now tackle medium-sized duplicates

// Step 3: Continue systematic cleanup
threshold: 100  # Finally address smaller duplicates

// Step 4: Maintain strictness
threshold: 100  # Keep this level for ongoing development
```

## Best Practices

1. **Start with higher threshold** (150-200) and lower gradually
2. **Review duplicates in batches** during focused refactoring sessions
3. **Use `//nolint:dupl`** for intentional duplicates with explanatory comment
4. **Consider the cost/benefit** - Some small duplicates aren't worth fixing
5. **Extract abstractions** rather than blindly consolidating code
6. **Exclude test files** - Tests naturally have repetitive patterns
7. **Exclude generated code** - Generated code will always have patterns
8. **Document why duplicates are intentional** if using `//nolint:dupl`
9. **Use as a refactoring tool**, not a blocker for all PRs
10. **Balance strictness** with team productivity

## Common Scenarios and Solutions

### Scenario 1: Too Many False Positives

**Problem:** dupl flags many small code snippets that aren't worth refactoring.

**Solution:** Increase the threshold.
```yaml
linters-settings:
  dupl:
    threshold: 200  # Only flag substantial duplicates
```

### Scenario 2: Test File Duplication

**Problem:** All test files are flagged because they use similar patterns.

**Solution:** Exclude test files or use `//nolint:dupl`.
```yaml
issues:
  exclude-rules:
    - linter: dupl
      path: (.+)_test\.go
```

### Scenario 3: Generated Code Patterns

**Problem:** Protobuf-generated code is flagged as duplicate.

**Solution:** Exclude generated directories.
```yaml
run:
  skip-dirs:
    - generated
    - vendor

issues:
  exclude-rules:
    - linter: dupl
      path: (.+)_pb\.go
    - linter: dupl
      path: (.+)_generated\.go
```

### Scenario 4: Intentional Domain-Specific Duplication

**Problem:** Business logic in different modules happens to be similar.

**Solution:** Use `//nolint:dupl` with comment explaining why.
```go
// These two functions look similar but serve different business purposes
// The duplication is intentional to keep domain logic clear.
//nolint:dupl
func processPaymentA(...) { ... }

//nolint:dupl
func processPaymentB(...) { ... }
```

### Scenario 5: Legacy Code with Thousands of Duplicates

**Problem:** Enabling dupl shows 5000+ violations - impossible to fix all at once.

**Solution:** Gradual approach.
```yaml
# Phase 1: Start with high threshold (300)
dupl:
  threshold: 300  # Fix only the most egregious duplicates

# Phase 2: Reduce threshold after cleanup (200)
dupl:
  threshold: 200  # Next batch

# Phase 3: Normal strictness (100)
dupl:
  threshold: 100  # Maintain going forward
```

## Summary

**dupl** is a **valuable code quality tool** that identifies duplicate code sequences:

- ✅ **Finds real duplication** - AST-based analysis is more accurate than text comparison
- ✅ **Fast execution** - Suitable for CI/CD pipelines
- ✅ **Improves maintainability** - Reduces code that needs updating when bugs are found
- ✅ **Identifies refactoring opportunities** - Large duplicates are good candidates for extraction
- ✅ **Configurable threshold** - Adjust strictness based on project needs
- ⚠️ **May have false positives** - Code that's similar by design
- ⚠️ **Requires manual review** - Cannot automatically fix issues
- ⚠️ **Context-agnostic** - Doesn't understand business logic

**Recommendation:** **ENABLE** for medium-to-large projects with `threshold: 100-150`. Exclude test files (`(.+)_test\.go`) and generated code. Start with higher threshold (200) if codebase has many existing duplicates, then gradually reduce to 100-150 as you clean up. Use `//nolint:dupl` for intentional duplicates with explanatory comments.

---

**Reference:** https://github.com/mibk/dupl
