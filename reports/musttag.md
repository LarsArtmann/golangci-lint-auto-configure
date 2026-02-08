# musttag Linter - Comprehensive Analysis

## What the Linter Does

**musttag** enforces struct tags for JSON/XML/YAML marshaling and unmarshaling in Go. It ensures that fields being serialized to or deserialized from external formats have the proper struct tags. This is a **CRITICAL** linter for data consistency and type safety in Go applications.

### The Problem It Detects

Go's encoding packages require struct tags:

- **Missing tags** - Fields won't be included in serialization
- **Incorrect tag names** - Causes API contract violations
- **Inconsistent tags** - Different fields use different naming conventions
- **Optional tags** - Causes silent failures when data missing
- **Omitted tags** - Values silently dropped during unmarshaling

### How It Works

musttag analyzes struct types and checks:

1. **Struct tag presence** - Ensures fields have tags when encoding
2. **Tag format validation** - Verifies tag syntax is correct
3. **Encoding function mapping** - Knows which functions require tags (json.Marshal, yaml.Unmarshal, etc.)

### Examples

```go
// ❌ BAD: Missing JSON tag
type User struct {
    ID    int
    Name   string  // musttag: field 'Name' should have JSON tag
    Email  string
}

func toJSON(u User) ([]byte, error) {
    return json.Marshal(u)  // Name won't be in JSON
}

// ✅ GOOD: All fields have tags
type User struct {
    ID    int    `json:"id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
}
```

```go
// ❌ BAD: Inconsistent optional tags
type Request struct {
    Page    int     `json:"page,omitempty"`  // Optional
    PageSize int     `json:"pageSize"`             // Not optional
    Search  string  `json:"search,omitempty"`  // Optional
}

// ✅ GOOD: Consistent optional tags
type Request struct {
    Page    int     `json:"page,omitempty"`
    PageSize int     `json:"pageSize,omitempty"`
    Search  string  `json:"search,omitempty"`
}
```

```go
// ❌ BAD: Missing YAML tag
type Config struct {
    Database string  // musttag: struct encoded to YAML
    Host     string
}

func saveConfig(c Config) error {
    data, _ := yaml.Marshal(c)  // Database, Host won't be in YAML
    return os.WriteFile("config.yaml", data, 0644)
}

// ✅ GOOD: All fields have tags
type Config struct {
    Database string `yaml:"database"`
    Host     string `yaml:"host"`
}
```

```go
// ❌ BAD: Wrong encoding package
type Product struct {
    ID    int    `json:"id"`
    Name   string `json:"name"`
    Price  float  `xml:"price"`  // musttag: using json.Marshal, not xml
}

func toXML(p Product) ([]byte, error) {
    return xml.Marshal(p)  // Name won't be in XML
}

// ✅ GOOD: Match encoding package tags
type Product struct {
    ID    int    `json:"id"`
    Name   string `json:"name"`
    Price  float  `json:"price"`  // Match json.Marshal
}
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

- **REST/GraphQL APIs** - All API request/response types
- **Web services** - HTTP JSON/XML payloads
- **Configuration files** - JSON/YAML/TOML config parsing
- **Database models** - ORM entities with tags
- **Microservices** - Inter-service communication
- **Public SDKs** - APIs consumed by others

**Specific Scenarios:**

**1. JSON APIs**

- REST endpoints
- GraphQL resolvers
- Webhook handlers
- API request/response types

**2. Configuration Files**

- Application config (JSON/YAML)
- Environment-specific configs
- User preferences
- Feature flags

**3. Database Models**

- PostgreSQL entities
- MySQL models
- MongoDB documents
- Redis structures

**4. XML Processing**

- SOAP services
- Sitemap generation
- XML-based protocols
- RSS feeds

**5. YAML Processing**

- Kubernetes manifests
- Docker Compose files
- CI/CD pipelines
- Helm charts

### ❌ Disable For:

**Specific Scenarios:**

**1. Non-Serialized Structs**

```yaml
# Structs used only in Go code
linters:
  enable:
    - musttag
linters-settings:
  musttag:
    functions:
      - go:.* # Exclude Go-only functions
```

**2. Test Mocks**

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [musttag]
```

**3. Protobuf/MessagePack**

```yaml
# Using binary serialization instead of JSON/XML/YAML
linters:
  enable:
    - musttag
linters-settings:
  musttag:
    functions: []
```

### Priority Assessment

- **Default Priority**: CRITICAL
- **Value**: **HIGHEST** for JSON/XML/YAML APIs - Prevents data loss
- **Effort**: Low - Simple tag addition, minimal false positives
- **Recommendation**: **ALWAYS ENABLE** for data serialization

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    musttag:
      # Functions that require struct tags
      # Default: all encoding functions
      functions:
        - json:.*
        - xml:.*
        - yaml:.*
        - toml:.*

      # Tag names to require
      # Default: all tags (json, xml, yaml, toml, etc.)
      tag-names:
        - json
        - xml
        - yaml

      # Auto-detect encoding functions (advanced)
      # Default: false
      auto-detect: false
```

### `functions` Option

- **Type**: `[]string`
- **Default**: All encoding functions
- **Description**: List of functions that require struct tags

**Format:** `package:Function` or `package:.*` (regex)

**Built-in Functions (Default):**

- `json:Marshal`, `json:MarshalIndent`, `json:Unmarshal`, `json:Decoder`
- `xml:Marshal`, `xml:Unmarshal`, `xml:Decoder`
- `yaml:Marshal`, `yaml:Unmarshal`
- `toml:Marshal`, `toml:Unmarshal`

**Custom Functions:**

```yaml
functions:
  - json:.*
  - xml:.*
  - encoding/json:.*
  - mypackage:CustomMarshal
```

### `tag-names` Option

- **Type**: `[]string`
- **Default**: All tags
- **Description**: List of tag names to require on fields

**Common Tags:**

- `json`
- `xml`
- `yaml`
- `toml`
- `bson` (MongoDB)
- `gorm` (ORM)

**Example:**

```yaml
tag-names:
  - json
  - yaml
```

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# Most production applications
version: "2"
linters:
  settings:
    musttag:
      functions:
        - json:.*
        - xml:.*
        - yaml:.*
```

#### ✅ JSON-Only Configuration

```yaml
# REST APIs, JSON databases
version: "2"
linters:
  settings:
    musttag:
      functions:
        - json:.*
      tag-names:
        - json
```

#### ✅ Multiple Encoding Configuration

```yaml
# Applications using JSON, XML, YAML
version: "2"
linters:
  settings:
    musttag:
      functions:
        - json:.*
        - xml:.*
        - yaml:.*
      tag-names:
        - json
        - xml
        - yaml
```

#### ✅ Exclude Test Files

```yaml
version: "2"
linters:
  settings:
    musttag:
      functions:
        - json:.*

issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [musttag]
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

musttag works excellently with:

| Linter          | Relationship  | Value                                                   |
| --------------- | ------------- | ------------------------------------------------------- |
| **govet**       | Complementary | govet: general struct tag issues, musttag: missing tags |
| **errchkjson**  | Complementary | errchkjson: type validation, musttag: tag presence      |
| **staticcheck** | Complementary | staticcheck: deep analysis, musttag: tag requirement    |
| **gosec**       | Complementary | gosec: security issues, musttag: data contract          |
| **tagalign**    | Complementary | tagalign: formatting, musttag: presence                 |
| **tagliatelle** | Complementary | tagliatelle: case validation, musttag: presence         |

**Complete Serialization Suite:**

```yaml
linters:
  enable:
    - musttag # Tag enforcement (CRITICAL)
    - errchkjson # JSON type safety (CRITICAL)
    - govet # General vet (CRITICAL)
    - staticcheck # Deep analysis (CRITICAL)
    - tagalign # Tag formatting (HIGH)
```

### 🔒 No Conflicts

- **No known conflicts** - musttag focuses on tag presence
- **Independent operation** - Doesn't overlap with other linter functionality
- **Safe to enable with all linters**

## Practical Examples

### ✅ Example 1: REST API Request/Response

```go
// ❌ BAD: Missing tags
type CreateUserRequest struct {
    Username string
    Password string
}

type CreateUserResponse struct {
    Success bool
    UserID  int
}

// ✅ GOOD: All fields tagged
type CreateUserRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type CreateUserResponse struct {
    Success bool `json:"success"`
    UserID  int  `json:"user_id,omitempty"`
}
```

### ✅ Example 2: Configuration File

```go
// ❌ BAD: Mixed encoding, missing YAML tags
type Config struct {
    Database string `json:"database"`  // Wrong tag for YAML
    Host     string `json:"host"`
    Port     int    `json:"port"`
}

func LoadConfig() (*Config, error) {
    data, _ := os.ReadFile("config.yaml")
    var config Config
    yaml.Unmarshal(data, &config)  // Database, Host not in YAML
    return &config, nil
}

// ✅ GOOD: Correct YAML tags
type Config struct {
    Database string `yaml:"database"`
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
}

func LoadConfig() (*Config, error) {
    data, err := os.ReadFile("config.yaml")
    if err != nil {
        return nil, err
    }
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    return &config, nil
}
```

### ✅ Example 3: Database ORM

```go
// ❌ BAD: Missing tags
type User struct {
    ID        int
    Username  string
    Email     string
    CreatedAt time.Time
}

// ✅ GOOD: ORM tags
type User struct {
    ID        int       `gorm:"primaryKey;column:id"`
    Username  string     `gorm:"column:username"`
    Email     string     `gorm:"column:email;unique"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
}
```

### ✅ Example 4: XML Processing

```go
// ❌ BAD: Missing XML tags
type Order struct {
    ID       string
    Items    []Item
    Total    float64
}

func (o *Order) ToXML() ([]byte, error) {
    return xml.Marshal(o)  // ID, Items, Total not in XML
}

// ✅ GOOD: XML tags
type Order struct {
    ID       string `xml:"id,attr"`
    Items    []Item `xml:"items>item"`
    Total    float64 `xml:"total,attr"`
}
```

### ✅ Example 5: Optional Fields

```go
// ❌ BAD: Inconsistent omitempty
type UpdateUserRequest struct {
    ID       int    `json:"id"`
    Username  string `json:"username,omitempty"`  // Optional
    Email     string  // Not optional - inconsistent

func UpdateUser(req UpdateUserRequest) error {
    data, _ := json.Marshal(req)
    // Email always sent, even if empty string
    return http.Post(url, "application/json", bytes.NewReader(data))
}
```

## Best Practices

1. **ALWAYS enable musttag** for JSON/XML/YAML APIs
2. **Tag all fields** in structs used with encoding functions
3. **Use omitempty consistently** - Either all optional fields or none
4. **Match encoding package tags** - Use `json:` tags with json.Marshal
5. **Use snake_case for JSON tags** - Go field names PascalCase, JSON snake_case
6. **Use omitempty for pointers** - Better than value types with optional
7. **Exclude test files** - Test mocks often don't need tags
8. **Combine with errchkjson** - musttag for presence, errchkjson for types
9. **Review struct tags in code reviews** - Ensure consistency
10. **Use tag aliases** - When JSON field name differs from Go field

## Common Scenarios and Solutions

### Scenario 1: Legacy Code Without Tags

**Problem:** Many existing structs don't have tags.

**Solution:** Add tags systematically, starting with most-used structs.

```bash
# Find structs used with encoding
grep -r "json.Marshal\|xml.Marshal\|yaml.Marshal" ./...
grep -r "struct {" ./...

# Add tags to all found structs
```

### Scenario 2: Mixed Encoding Usage

**Problem:** Struct used with both JSON and XML.

**Solution:** Use multiple tags.

```go
type Product struct {
    ID    int    `json:"id" xml:"id,attr"`
    Name   string `json:"name" xml:"name"`
    Price  float `json:"price" xml:"price"`
}
```

### Scenario 3: API Versioning

**Problem:** Different API versions use different field names.

**Solution:** Use tag aliases.

```go
type User struct {
    ID    int    `json:"id" xml:"user-id"`           // v1 API
    UserID int    `json:"user_id" xml:"userId"`    // v2 API
}
```

### Scenario 4: Partial Updates

**Problem:** Want to update only some fields.

**Solution:** Use omitempty on all optional fields.

```go
type UpdateUserRequest struct {
    ID      int    `json:"id"`
    Username string `json:"username,omitempty"`
    Email    string `json:"email,omitempty"`
}

// Only non-empty fields will be in JSON
```

### Scenario 5: Test Files

**Problem:** Test structs flagged for missing tags.

**Solution:** Exclude test files.

```yaml
issues:
  exclude-rules:
    - path: (.+)_test\.go
      linters: [musttag]
```

## Summary

**musttag** is a **CRITICAL** linter for data serialization:

- ✅ **Highest priority** for JSON/XML/YAML APIs - Prevents data loss
- ✅ **Prevents silent failures** - Tags missing cause values to be dropped
- ✅ **Ensures API contracts** - Correct field names in serialized data
- ✅ **Simple to use** - Just add tags to struct fields
- ✅ **Configurable** - Adjust which encoding functions/tags to check
- ✅ **Comprehensive** - Covers JSON, XML, YAML, TOML, etc.
- ✅ **Low overhead** - Simple static analysis
- ⚠️ **May need exclusions** - Test files, Go-only structs
- ⚠️ **Encoding-specific** - Only relevant for code using encoding packages

**Recommendation:** **ALWAYS ENABLE** for any code using JSON/XML/YAML/TOML encoding (REST APIs, GraphQL, configuration files). Ensure **all struct fields** used with encoding functions have appropriate tags matching the encoding package. Use **omitempty consistently** on optional fields. Combine with **errchkjson** for complete type safety. Exclude test files (`(.+)_test\.go`).

**Top 3 Configuration Tips:**

1. Match encoding package tags (use `json:` with json.Marshal)
2. Use snake_case in JSON tags (Go field: Email, JSON tag: "email")
3. Make all optional fields pointers with omitempty

---

**Reference:** https://github.com/go-simpler/musttag
