# Comprehensive Improvement Plan - golangci-lint-auto-configure
**Date:** 2026-03-19 08:37  
**Status:** Production Ready, Incremental Improvements Planned

---

## What I Forgot / Could Do Better

### 1. Context Propagation Analysis
The `contextcheck` warnings at `commands.go:88` and `generator.go:39` are **false positives**:
- Context IS properly passed through `fang.Execute(ctx, ...)` 
- Cobra commands receive context via `cmd.Context()`
- The linter doesn't trace through third-party libraries properly

**Decision:** Mark with `//nolint:contextcheck` after verification

### 2. Type Safety Improvements
Current Config types use `map[string]any` for settings which loses type safety:
```go
// Current - unsafe
Settings map[string]any `yaml:"settings,omitempty"`

// Better - typed
Settings LinterSettings `yaml:"settings,omitempty"`
```

### 3. Missing Libraries That Could Help

| Library | Purpose | Impact |
|---------|---------|--------|
| `github.com/pelletier/go-toml/v2` | TOML config support | Medium |
| `github.com/spf13/afero` | Filesystem abstraction for tests | High |
| `github.com/mitchellh/mapstructure` | Struct mapping from maps | Medium |
| `github.com/caarlos0/env/v11` | Environment variable config | Low |

### 4. Architecture Improvements

**Current Issues:**
- `ConfigLoader` interface is large (8 methods) - violates Interface Segregation
- Error types could be more specific
- No unified configuration source (file, env, flags)

---

## Multi-Step Execution Plan (Sorted by Impact/Work)

### Phase 1: High Impact, Low Effort (Do First)

#### Step 1.1: Fix contextcheck false positives [5min]
**Files:** `internal/cli/commands.go`, `pkg/report/generator.go`  
**Impact:** Clean lint output  
**Work:** Add `//nolint:contextcheck` with explanation

```go
//nolint:contextcheck // Context is passed through fang.Execute which linter doesn't trace
rootCmd := NewRootCommand()
```

#### Step 1.2: Add filesystem abstraction for testing [15min]
**File:** `pkg/config/loader.go`  
**Impact:** Better testability  
**Work:** Use `afero.Fs` interface

```go
type Loader struct {
    logger *log.Logger
    fs     afero.Fs  // Add this
}

func NewLoader(logger *log.Logger) *Loader {
    return &Loader{
        logger: logger,
        fs:     afero.NewOsFs(),  // Default to real filesystem
    }
}

// NewLoaderWithFS for testing
func NewLoaderWithFS(logger *log.Logger, fs afero.Fs) *Loader {
    return &Loader{logger: logger, fs: fs}
}
```

#### Step 1.3: Add TOML config support [20min]
**File:** `pkg/config/loader.go`  
**Impact:** User flexibility  
**Work:** Detect format and use appropriate decoder

```go
func (l *Loader) detectFormat(path string) ConfigFormat {
    switch filepath.Ext(path) {
    case ".toml":
        return ConfigFormatTOML
    case ".json":
        return ConfigFormatJSON
    default:
        return ConfigFormatYAML
    }
}
```

### Phase 2: Medium Impact, Medium Effort

#### Step 2.1: Split ConfigLoader interface [30min]
**File:** `pkg/types/types.go`  
**Impact:** Better interface design  
**Work:** Split into focused interfaces

```go
// ConfigReader - read operations
type ConfigReader interface {
    LoadConfig(path string) (*Config, error)
    FindConfigFile(startDir string) (string, error)
}

// ConfigWriter - write operations
type ConfigWriter interface {
    SaveConfig(config *Config, path string) error
}

// ConfigValidator - validation operations
type ConfigValidator interface {
    ValidateConfig(config *Config) []error
}
```

#### Step 2.2: Add structured validation errors [25min]
**File:** `pkg/types/validation.go`  
**Impact:** Better UX for validation failures

```go
type ValidationResult struct {
    Valid   bool
    Errors  []FieldError
    Warnings []FieldWarning
}

type FieldError struct {
    Field   string
    Value   interface{}
    Rule    string
    Message string
}

func (r ValidationResult) Error() string {
    // Format errors for display
}
```

#### Step 2.3: Implement config builder pattern [30min]
**New File:** `pkg/config/builder.go`  
**Impact:** Easier config construction

```go
type Builder struct {
    config *Config
}

func NewBuilder() *Builder {
    return &Builder{config: DefaultConfig()}
}

func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
    b.config.Run.Timeout = timeout.String()
    return b
}

func (b *Builder) WithLinter(name string) *Builder {
    b.config.Linters.Enable = append(b.config.Linters.Enable, name)
    return b
}

func (b *Builder) Build() *Config {
    return b.config
}
```

### Phase 3: Lower Impact, Higher Effort (Future)

#### Step 3.1: Add environment variable support [45min]
**Impact:** 12-factor app compliance  
**Work:** Integrate `caarlos0/env` for env-based config

#### Step 3.2: Refactor Fixer to reduce complexity [60min]
**File:** `pkg/linter/fixer.go`  
**Impact:** Maintainability  
**Work:** Extract strategies for different fix types

```go
type FixStrategy interface {
    Apply(cfg *Config, analysis *ConfigAnalysis) error
}

type DeprecatedLinterStrategy struct{}
type RedundantLinterStrategy struct{}
type RecommendedLinterStrategy struct{}
```

#### Step 3.3: Add configuration schema generation [40min]
**Impact:** Better IDE support  
**Work:** Generate JSON Schema from types

---

## Libraries Research Summary

### Already Using (Well-Established)
- ✅ `samber/mo` - Railway-oriented programming
- ✅ `charmbracelet/log` - Structured logging
- ✅ `cobra` - CLI framework
- ✅ `validator/v10` - Struct validation
- ✅ `go-playground` - Validation ecosystem

### Could Add
| Library | Stars | Maturity | License |
|---------|-------|----------|---------|
| spf13/afero | 5k+ | Stable | Apache 2.0 |
| pelletier/go-toml | 4k+ | Stable | MIT |
| mitchellh/mapstructure | 7k+ | Stable | MIT |
| caarlos0/env | 4k+ | Stable | MIT |

---

## Architecture Improvements

### Current: Monolithic ConfigLoader
```
ConfigLoader (8 methods)
├── LoadConfig
├── SaveConfig
├── FindConfigFile
├── FindOrGetDefaultConfigPath
├── EnsureGitRepo
├── ValidateConfig
├── GetLintersEnabled
└── GetLintersDisabled
```

### Proposed: Segregated Interfaces
```
ConfigReader (2 methods)
├── LoadConfig
└── FindConfigFile

ConfigWriter (1 method)
└── SaveConfig

ConfigValidator (1 method)
└── ValidateConfig

GitChecker (1 method)
└── EnsureGitRepo

ConfigDefaults (1 method)
└── CreateDefaultConfig
```

**Benefits:**
- Easier mocking in tests
- Clearer dependencies
- Follows Interface Segregation Principle

---

## Risk Assessment

| Change | Risk | Mitigation |
|--------|------|------------|
| Filesystem abstraction | Low | Default to OS filesystem |
| Interface splitting | Medium | Keep old interface as composite |
| TOML support | Low | Feature addition only |
| Validation refactor | Medium | Keep backward compatible |

---

## Recommended Next Steps

### Immediate (Today)
1. ✅ Mark contextcheck false positives
2. ✅ Add afero filesystem abstraction
3. ✅ Run full test suite

### This Week
4. Add TOML config support
5. Split ConfigLoader interface
6. Update AGENTS.md with new patterns

### Next Sprint
7. Add config builder pattern
8. Implement structured validation errors
9. Research environment variable support

---

## Verification Checklist

After each change:
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes (19 tests)
- [ ] Branching-flow score remains ≥ 99
- [ ] Documentation updated
- [ ] Commit with descriptive message

---

*Plan generated: 2026-03-19 08:37*  
*Current Status: 99/100 branching-flow score, all tests passing*
