# dogsled Linter - Comprehensive Analysis

## What the Linter Does

**dogsled** checks for assignments with **too many blank identifiers (`_`)**. It flags code where you're discarding multiple return values using blank identifiers, which might indicate you're ignoring important data or using problematic APIs.

### The Problem It Detects

When a function returns multiple values and you discard several with `_`, it can indicate:

- You're ignoring important return values that should be handled
- The API returns too many values and could be simplified
- Code readability is suffering (lots of `_` reduces clarity)
- Potential bugs from ignored error values or data

### How It Works

dogsled counts the number of blank identifiers (`_`) in assignment statements and flags when the count exceeds the configured threshold.

**Default threshold**: 2 blank identifiers

### Examples

```go
// ❌ BAD: Too many blank identifiers (violates default threshold of 2)
a, _, _, err := getUserData()  // dogsled: assignment has 3 blank identifiers

// ❌ BAD: Often seen in auto-generated protocol buffer code
_, _, _, _ = someProtoMethod()  // 4 blank identifiers

// ✅ GOOD: Within limits (1-2 blank identifiers)
a, _, err := getUserData()     // OK: 1 blank identifier
a, b, c, _ := getData()        // OK: 1 blank identifier
a, b, _, _ := getData()        // OK: 2 blank identifiers

// ✅ GOOD: Explicitly handle all values (best practice)
a, b, c, err := getUserData()  // OK: no blank identifiers
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

- **Medium to large codebases** - Catch API design issues early
- **Team environments** - Ensure all returned values are considered
- **Libraries and APIs** - Discourage excessive return values
- **Codebases with quality gates** - Maintain code clarity standards
- **Projects with code reviews** - Objective metric for review discussions

**Scenarios:**

- **API design reviews** - Too many blank identifiers suggests API could be simpler
- **Error handling enforcement** - Ensure errors aren't discarded with `_`
- **Code clarity improvements** - Reduce visual noise from multiple `_`
- **Working with junior developers** - Teach them to consider all return values

### ❌ Disable For:

**Project Types:**

- **Small utilities** - Flexibility is more important
- **Proof-of-concept code** - Quick iteration is priority
- **Generated code** - Auto-generated code often has many return values
- **Protocol buffer code** - Protobuf generates methods with many returns
- **Interop/wrapper code** - May need to match external API signatures

**Specific Scenarios:**

- Working with **protocol buffer generated code** - Often has 3-4+ return values
- **Database ORM methods** - Some ORMs return many values for rows
- **Legacy code** - Refactoring would be too risky
- **Testing frameworks** - Test helpers may return multiple ignorable values

### Priority Assessment

- **Default Priority**: Medium (as defined in project linter data)
- **Value**: Catches API design issues and potential bugs
- **Effort**: Low - simple checks with minimal false positives
- **Recommendation**: **Enable** for most projects, exclude generated code

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    dogsled:
      # Maximum blank identifiers allowed (default: 2)
      max-blank-identifiers: 2
```

### `max-blank-identifiers` Option

- **Type**: `int`
- **Default**: `2`
- **Range**: `1` or higher
- **Description**: Maximum number of `_` allowed in assignment before triggering warning

**Behavior:**

- Setting of `2` means assignments with **3 or more** `_` will be flagged
- Setting of `3` means assignments with **4 or more** `_` will be flagged
- Setting of `1` means assignments with **2 or more** `_` will be flagged (very strict)

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# Most projects - catches most problematic cases
version: "2"
linters:
  settings:
    dogsled:
      max-blank-identifiers: 2 # Flag 3+ blank identifiers
```

#### ✅ Relaxed Configuration

```yaml
# For codebases with legitimate multi-value returns
version: "2"
linters:
  settings:
    dogsled:
      max-blank-identifiers: 3 # Flag 4+ blank identifiers
```

#### ✅ Very Strict Configuration

```yaml
# For teams wanting to minimize blank identifier usage
version: "2"
linters:
  settings:
    dogsled:
      max-blank-identifiers: 1 # Flag 2+ blank identifiers
```

#### ✅ Exclude Generated Code

```yaml
# Essential for protobuf and generated code
version: "2"
linters:
  settings:
    dogsled:
      max-blank-identifiers: 2

  exclusions:
    rules:
      - path: (.+)_generated\.go
        linters: [dogsled]
      - path: (.+)_pb\.go
        linters: [dogsled]
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

dogsled works well alongside:

| Linter            | Relationship        | Benefit                                                                                                                   |
| ----------------- | ------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **`errcheck`**    | Complementary       | Both catch error handling issues; errcheck ensures errors are checked, dogsled catches when errors are discarded with `_` |
| **`unused`**      | Complementary       | unused finds unused variables (when you should use a value), dogsled catches when you discard with `_` unnecessarily      |
| **`ineffassign`** | Compatible          | Both catch assignment-related issues                                                                                      |
| **`nolintlint`**  | Process improvement | If you frequently have to `//nolint:dogsled`, it may indicate the API is problematic                                      |
| **`gocritic`**    | Compatible          | Checks different code quality aspects                                                                                     |

**Complete Assignment Quality Suite:**

```yaml
linters:
  enable:
    - dogsled # Blank identifier usage
    - errcheck # Error checking
    - unused # Unused variables
    - ineffassign # Ineffective assignments

  settings:
    dogsled:
      max-blank-identifiers: 2
```

### 🔒 No Conflicts

- **No known conflicts** - dogsled focuses specifically on assignment patterns
- **Independent operation** - Doesn't overlap with other linter functionality
- **Safe to enable with all linters**

## Practical Examples

### ✅ Example 1: Refactoring to Reduce Blank Identifiers

```go
// ❌ Before: Too many blank identifiers (3)
func processUser(id string) error {
    user, _, _, err := getUserData(id)  // dogsled: 3 blank identifiers
    if err != nil {
        return err
    }

    return process(user)
}

// ✅ After: Handle or remove unnecessary returns
func processUser(id string) error {
    // Refactored getUserData to return fewer values
    user, err := getUserData(id)  // OK: no blank identifiers
    if err != nil {
        return err
    }

    return process(user)
}
```

### ✅ Example 2: Protocol Buffer Generated Code

```go
// Common in protobuf generated code - needs exclusion
// pkg/proto/user.pb.go

// ❌ Generated code often has this pattern
func (m *User) Reset() {
    *m = User{}  // OK: no blank identifiers in this method
}

func (m *User) String() string {
    b, _ := proto.Marshal(m)  // OK: 1 blank identifier
    return string(b)
}

// ⚠️ This is common in generated code (depending on protobuf version)
_, _, _ = someInternalFunc()  // dogsled would flag this (3 blank identifiers)

// Solution: Exclude generated files
# .golangci.yml
version: "2"

linters:
  settings:
    dogsled:
      max-blank-identifiers: 2

  exclusions:
    rules:
      - path: (.+)_pb\.go
        linters: [dogsled]
      - path: (.+)_generated\.go
        linters: [dogsled]
```

### ✅ Example 3: Database Row Scanning

```go
// ❌ Database scan with many ignored columns
func getUserName(id string) (string, error) {
    var name string
    row := db.QueryRow("SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?", id)

    // 3 blank identifiers - might be legitimate or API could be better
    err := row.Scan(&name, _, _, _, _)  // dogsled: 3 blank identifiers
    return name, err
}

// ✅ Better: Refactor query to only SELECT needed columns
func getUserName(id string) (string, error) {
    var name string
    row := db.QueryRow("SELECT name FROM users WHERE id = ?", id)

    err := row.Scan(&name)  // OK: no blank identifiers
    return name, err
}
```

### ✅ Example 4: Test Helpers

```go
package testutil

// Helper function - might legitimately return multiple ignorable values
func CreateTestUser(t *testing.T) (*User, string, string, string, string) {
    // Creates user, returns user + temp password + 3 tokens
    // In tests, often only need the user
    return user, password, token1, token2, token3
}

// In test:
func TestUser(t *testing.T) {
    user, _, _, _, _ := CreateTestUser(t)  // dogsled: 4 blank identifiers

    // Only use user in test
    assert.Equal(t, "test@example.com", user.Email)
}

// Solutions:

// 1. Exclude test files
# .golangci.yml
exclusions:
  rules:
    - path: (.+)_test\.go
      linters: [dogsled]

// 2. Or refactor helper
func CreateTestUser(t *testing.T) *User {
    // Return just what's needed
    return user
}

// 3. Or create specific helper
func CreateTestUserWithTokens(t *testing.T) (*User, string, string) {
    // Return commonly needed values
    return user, password, token1
}
```

### ✅ Example 5: Error Handling with Multiple Returns

```go
// ❌ Error handling - might ignore multiple values
func processFile(path string) error {
    f, _, _, err := openFile(path)  // dogsled: 3 blank identifiers
    if err != nil {
        return err
    }
    defer f.Close()

    return process(f)
}

// ✅ Better: Refactor to meaningful returns
func openFile(path string) (*os.File, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    return f, nil
}

func processFile(path string) error {
    f, err := openFile(path)  // OK: no blank identifiers
    if err != nil {
        return err
    }
    defer f.Close()

    return process(f)
}
```

## Best Practices

1. **Start with default (2)**: Good balance for most codebases
2. **Refactor rather than increase**: Try to simplify APIs before raising the threshold
3. **Exclude generated code**: Protobuf, mock, and other generated files often need exceptions
4. **Use `_` sparingly**: Consider if the API could return fewer values
5. **Check error paths**: Don't use `_` for errors - handle them properly
6. **Document exclusions**: Use `//nolint:dogsled` with explanatory comment

## Common Scenarios and Solutions

### Scenario 1: Protocol Buffers

```go
// Generated protobuf code often has this pattern
_, _, _, _ = x.XXX_unrecognized  // Common in older protobuf versions

// Solution: Exclude protobuf files
exclusions:
  rules:
    - path: (.+)_pb\.go
      linters: [dogsled]
```

### Scenario 2: Time Package

```go
// Time package methods often return multiple values
_, month, day := t.Date()  // Might be acceptable

// Solution: Consider if you need date parts
// Better: t.Month() and t.Day() if that's all you need
month := t.Month()
day := t.Day()
```

### Scenario 3: Map Operations

```go
// Map operations sometimes have multiple returns
value, _, _, found := complexMap.Lookup(key)

// Solution: Either accept it (if legitimate) or refactor map type
```

## Summary

**dogsled** is a **focused linter** that identifies assignments with excessive blank identifier usage:

- ✅ **Catches API design issues** - Functions returning too many values
- ✅ **Improves code clarity** - Fewer blank identifiers = cleaner code
- ✅ **Error handling helper** - Discourages discarding important values
- ✅ **Low overhead** - Simple check, minimal false positives
- ✅ **Configurable** - Adjust threshold based on your needs
- ⚠️ **May need exclusions** - Generated code often requires exceptions

**Recommendation**: **ENABLE** for most projects with `max-blank-identifiers: 2` (default). Exclude generated files (protobuf, mocks) and test files if needed. Consider increasing to `3` if working with code that legitimately returns many values.

---

**Reference**: https://github.com/charithe/dogsled
