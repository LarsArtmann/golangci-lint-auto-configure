# errchkjson Linter - Comprehensive Analysis

## What the Linter Does

**errchkjson** checks that types passed to JSON encoding functions are appropriate. It reports unsupported types and identifies cases where error checking for JSON encoding/decoding can be omitted. This is a **CRITICAL** security and correctness linter for Go applications using JSON.

### The Problem It Detects

Go's `encoding/json` can encode/decode many types, but some types cause issues:

- **Unsupported types** in JSON encoding (chan, func, unsafe.Pointer)
- **Unexported struct fields** won't be marshaled
- **Nil pointer marshaling** may produce incorrect JSON
- **Error handling omitted** for JSON operations (when safe to do so)

### Examples

```go
// ❌ BAD: Unsupported type in JSON encoding
type Data struct {
    Value int
    Done  chan bool  // errchkjson: chan cannot be JSON encoded
}

func toJSON(d Data) ([]byte, error) {
    return json.Marshal(d)  // Will fail at runtime
}

// ✅ GOOD: Remove unsupported field
type Data struct {
    Value int
    Done  bool
}

func toJSON(d Data) ([]byte, error) {
    return json.Marshal(d)
}
```

```go
// ❌ BAD: Unexported field won't be in JSON
type User struct {
    ID    int    `json:"id"`
    name  string `json:"name"`  // errchkjson: unexported field
}

func toJSON(u User) ([]byte, error) {
    return json.Marshal(u)  // name won't be in JSON
}

// ✅ GOOD: Export the field
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
}

func toJSON(u User) ([]byte, error) {
    return json.Marshal(u)
}
```

```go
// ❌ BAD: Error not checked when it should be
func writeJSON(data interface{}) error {
    json.NewEncoder(os.Stdout).Encode(data)  // errchkjson: error not checked
}

// ✅ GOOD: Check the error
func writeJSON(data interface{}) error {
    return json.NewEncoder(os.Stdout).Encode(data)
}
```

```go
// ❌ BAD: Unsupported complex type
type Request struct {
    Handler func(http.ResponseWriter, *http.Request)  // errchkjson: func cannot be JSON encoded
}

func toJSON(r Request) ([]byte, error) {
    return json.Marshal(r)  // Will fail at runtime
}

// ✅ GOOD: Use serializable type
type Request struct {
    URL string
}

func toJSON(r Request) ([]byte, error) {
    return json.Marshal(r)
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**
- **All JSON APIs** - REST/GraphQL services
- **Web applications** - Any app serving/consuming JSON
- **Microservices** - Inter-service communication
- **CLI tools** - JSON configuration files
- **Configuration parsers** - JSON-based config
- **API clients** - Services consuming external JSON APIs

**Specific Scenarios:**
- **REST APIs** - Request/response bodies
- **GraphQL resolvers** - JSON responses
- **WebSocket messages** - JSON payload handling
- **Configuration files** - JSON config parsing
- **Database JSON storage** - JSON columns in PostgreSQL
- **Event streaming** - JSON event payloads

### ❌ Disable For:

**Specific Scenarios:**

**1. Non-JSON Projects**
```yaml
# Projects not using JSON
linters:
  enable:
    - errchkjson
```

**2. Protobuf/MessagePack**
- Using binary serialization instead of JSON
- Protobuf, Thrift, Avro projects

**3. Test Files with Intentional Errors**
```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errchkjson]
```

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** for JSON APIs - Prevents runtime panics
- **Effort**: Low - Simple checks, minimal false positives
- **Recommendation**: **ALWAYS ENABLE** for any code using JSON

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    errchkjson:
      # Report error returns when they can be omitted
      # Default: false
      check-error-free-encoding: false

      # Report error returns from MarshalIndent
      # Default: false
      check-error-free-encoding: false

      # Report error returns from Decode
      # Default: false
      check-error-free-decoding: false
```

### `check-error-free-encoding` Option

- **Type**: `bool`
- **Default**: `false`
- **Description**: Report error returns when JSON encoding is guaranteed error-free

**When to Enable:**
- Encoding primitive types (int, string, bool)
- Encoding slices of primitives
- Cases where you're certain encoding cannot fail

### `check-error-free-encoding` Options

- **check-error-free-marshaling** - Check json.Marshal()
- **check-error-free-encoding` - Check json.Encoder.Encode()

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)
```yaml
# Most JSON APIs
version: "2"
linters:
  settings:
    errchkjson:
      check-error-free-encoding: false
```

#### ✅ Strict Configuration
```yaml
# Security-critical, type-safe applications
version: "2"
linters:
  settings:
    errchkjson:
      check-error-free-encoding: false
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

errchkjson works excellently with:

| Linter | Relationship | Value |
|--------|--------------|---------|
| **errcheck** | Complementary | errcheck ensures JSON errors are checked, errchkjson ensures types are valid |
| **gosec** | Complementary | gosec: general security, errchkjson: JSON-specific security |
| **govet** | Complementary | govet: general issues, errchkjson: JSON-specific |
| **staticcheck** | Complementary | staticcheck: deep analysis, errchkjson: JSON type checking |
| **nilerr** | Complementary | nilerr: nil error returns, errchkjson: JSON encoding issues |
| **musttag** | Complementary | musttag: struct tags, errchkjson: JSON tag validation |

**Complete JSON Suite:**
```yaml
linters:
  enable:
    - errchkjson     # JSON type checking (CRITICAL)
    - errcheck        # Error handling (CRITICAL)
    - gosec           # Security (CRITICAL)
    - musttag         # Struct tags (CRITICAL)
    - staticcheck      # Deep analysis (CRITICAL)
    - nilerr           # Nil errors (CRITICAL)
```

### 🔒 No Conflicts

- **No known conflicts** - errchkjson focuses specifically on JSON encoding/decoding
- **Independent operation** - Doesn't overlap with other linter functionality
- **Safe to enable with all linters**

## Practical Examples

### ✅ Example 1: REST API Response

```go
// ❌ BAD: Unexported field
type APIResponse struct {
    Success bool   `json:"success"`
    data    UserData `json:"data"`  // errchkjson: unexported
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    resp := APIResponse{Success: true, data: getUserData()}
    json.NewEncoder(w).Encode(resp)  // data won't be in JSON
}

// ✅ GOOD: Export the field
type APIResponse struct {
    Success bool      `json:"success"`
    Data    UserData `json:"data"`
}
```

### ✅ Example 2: JSON Configuration

```go
// ❌ BAD: Error not checked
func loadConfig(filename string) (*Config, error) {
    data, _ := os.ReadFile(filename)  // errchkjson: error not checked (may be safe)
    var config Config
    json.Unmarshal(data, &config)
    return &config, nil
}

// ✅ GOOD: Check error
func loadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    return &config, nil
}
```

### ✅ Example 3: Webhook Payload

```go
// ❌ BAD: Unsupported type
type Webhook struct {
    ID     string `json:"id"`
    Action func() `json:"action"`  // errchkjson: func cannot be JSON encoded
}

func sendWebhook(w Webhook) error {
    payload, _ := json.Marshal(w)  // Runtime panic
    _, err := http.Post(url, "application/json", bytes.NewReader(payload))
    return err
}

// ✅ GOOD: Use serializable action
type Webhook struct {
    ID     string `json:"id"`
    Action string `json:"action"`  // Serializable
}
```

### ✅ Example 4: Event Stream

```go
// ❌ BAD: Channel in struct
type Event struct {
    ID    string `json:"id"`
    Data  chan []byte `json:"data"`  // errchkjson: chan cannot be encoded
}

func streamEvents(events <-chan Event) error {
    for e := range events {
        data, _ := json.Marshal(e)  // Runtime panic
        fmt.Println(string(data))
    }
    return nil
}

// ✅ GOOD: Remove channel, use slice
type Event struct {
    ID    string `json:"id"`
    Data  []byte `json:"data"`
}
```

### ✅ Example 5: JSON Field Tags

```go
// ❌ BAD: Multiple issues
type User struct {
    id       int     `json:id`       // errchkjson: unexported
    password string `json:password`
    Metadata interface{} `json:"metadata"` // errchkjson: interface{} is fine
}

// ✅ GOOD: Fix all issues
type User struct {
    ID       int     `json:"id"`
    Password string `json:"password"`
    Metadata interface{} `json:"metadata"`
}
```

## Best Practices

1. **ALWAYS enable errchkjson** for JSON APIs
2. **Export all fields** that should be in JSON
3. **Check JSON encoding/decoding errors** unless explicitly safe
4. **Remove unsupported types** (chan, func, unsafe.Pointer) from structs
5. **Use pointer receivers** for JSON methods to avoid value copying
6. **Validate JSON structure** before encoding
7. **Handle JSON decoding errors** with context
8. **Use json.Encoder** for streaming JSON
9. **Use json.Decoder** for streaming JSON parsing
10. **Combine with errcheck** for complete error handling

## Common Scenarios and Solutions

### Scenario 1: Large Struct with Many Fields

**Problem:** Unexported field buried in large struct.

**Solution:** Use errchkjson to catch it.
```yaml
linters:
  settings:
    errchkjson:
      check-error-free-encoding: false
```

### Scenario 2: Error-Free Encoding

**Problem:** Encoding primitive types always succeeds, errors can be omitted.

**Solution:** Use check-error-free-encoding: false to catch all.
```yaml
linters:
  settings:
    errchkjson:
      check-error-free-encoding: false
```

### Scenario 3: Test Mock Data

**Problem:** Test structs intentionally have unexported fields.

**Solution:** Exclude test files.
```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [errchkjson]
```

### Scenario 4: JSON in Goroutines

**Problem:** JSON operations in goroutines, errors hard to check.

**Solution:** Check errors in goroutine.
```go
go func() {
    data, err := json.Marshal(obj)
    if err != nil {
        log.Printf("Failed to marshal: %v", err)
        return
    }
    sendToKafka(data)
}()
```

## Summary

**errchkjson** is a **CRITICAL** linter for JSON-focused Go applications:

- ✅ **Highest priority** for JSON APIs - Prevents runtime panics
- ✅ **Type safety** - Ensures only JSON-serializable types are used
- ✅ **Error handling** - Catches omitted error checks
- ✅ **Low overhead** - Simple type checking
- ✅ **Comprehensive** - Covers encoding and decoding
- ✅ **Configurable** - Adjust strictness for error checking
- ⚠️ **JSON-specific** - Only relevant for JSON code
- ⚠️ **May need exclusions** - Test files, error-free encoding

**Recommendation:** **ALWAYS ENABLE** for any code using JSON (REST APIs, GraphQL, configuration files, event streaming). Combine with **errcheck** for complete error handling coverage. Use with **musttag** to ensure correct JSON struct tags. Exclude test files (`(.+)_test\.go`). Set `check-error-free-encoding: false` to catch all error cases.

**Top 3 Configuration Tips:**
1. Keep default settings (check-error-free-encoding: false)
2. Export all fields that should be in JSON
3. Check all JSON encoding/decoding errors

---

**Reference:** https://github.com/breml/errchkjson
