# Architecture Decision Records

This document captures key architectural decisions made during the development of golangci-lint-auto-configure.

## ADR-001: Railway-Oriented Programming with Result Types

**Status:** Accepted  
**Date:** 2026-03-26

### Context

The codebase uses railway-oriented programming (ROP) patterns for handling operations that can fail. This pattern allows for clean composition of operations that may fail at any point.

### Decision

We use the `samber/mo` library which provides `Result[T]` types that wrap successful values or errors. Helper functions in `pkg/types/result.go` create typed result aliases:

```go
type MigrationResultType = mo.Result[*MigrationResult]
type ConfigResult = mo.Result[*Config]
```

Success and error helpers:

```go
func OkMigration(result *MigrationResult) MigrationResultType
func ErrMigration(err error) MigrationResultType
```

### Consequences

**Positive:**

- Explicit error handling at each step
- Composable operations
- No nil pointer dereference panics
- Clear success/failure paths

**Negative:**

- Additional verbosity in calling code
- Learning curve for contributors unfamiliar with ROP

---

## ADR-002: MigrationResult Uses Error Field Instead of Success Bool

**Status:** Accepted  
**Date:** 2026-03-26

### Context

Previously, `MigrationResult` had a `Success bool` field. This required checking both the returned error AND the Success field to determine if an operation succeeded.

### Decision

Changed `MigrationResult` to have an `Error error` field with helper methods:

```go
type MigrationResult struct {
    FixesApplied int      `json:"fixes_applied"`
    Message      string   `json:"message"`
    NextSteps    []string `json:"next_steps,omitempty"`
    Error        error    `json:"-"` // Not serialized to JSON
}

func (m *MigrationResult) IsSuccess() bool { return m.Error == nil }
func (m *MigrationResult) IsFailure() bool { return m.Error != nil }
```

### Consequences

**Positive:**

- Single source of truth for success/failure
- Error information preserved in the result struct
- Cleaner conditional logic (`if result.IsSuccess()` vs `if err == nil && result.Success`)

**Negative:**

- Requires updating all call sites that checked `Success` field

---

## ADR-003: Detector Caching for Project Type Analysis

**Status:** Accepted  
**Date:** 2026-03-26

### Context

The `Detector.Detect()` method performs expensive file system walks to analyze project structure. If called multiple times, it would redundantly re-scan the file system.

### Decision

Added thread-safe caching to the Detector:

```go
type Detector struct {
    rootDir string
    cache   ProjectType
    cached  bool
    mu      sync.Mutex
}

func (d *Detector) Detect() ProjectType {
    d.mu.Lock()
    if d.cached {
        d.mu.Unlock()
        return d.cache
    }
    d.mu.Unlock()

    projectType := d.detect()

    d.mu.Lock()
    d.cache = projectType
    d.cached = true
    d.mu.Unlock()

    return projectType
}
```

### Consequences

**Positive:**

- Eliminates redundant file system operations
- Thread-safe for concurrent access
- Transparent to callers

**Negative:**

- Slight memory overhead for cached value
- Cache cannot be invalidated (acceptable for CLI use case)

---

## ADR-004: Separate Error Types for Different Domains

**Status:** Accepted  
**Date:** 2026-03-26

### Context

The application has distinct error domains: configuration, analysis, and reporting. Generic errors lose important context.

### Decision

Created specific error types in `pkg/errors/errors.go`:

```go
type ConfigError struct {
    Message string
    Path    string
    Cause   error
}

type AnalysisError struct {
    Message string
    File    string
    Cause   error
}

type MigrationError struct {
    Message string
    Config  string
    Cause   error
}

type ReportError struct {
    Message string
    Path    string
    Cause   error
}
```

Helper functions allow type checking:

```go
func IsConfigError(err error) bool
func IsAnalysisError(err error) bool
func IsMigrationError(err error) bool
func IsReportError(err error) bool
```

### Consequences

**Positive:**

- Rich error context with file paths, config names, etc.
- Type-safe error handling
- Easier debugging with structured error information

**Negative:**

- More error types to maintain
- Slightly more verbose error creation

---

## ADR-005: Afero Filesystem Abstraction

**Status:** Accepted  
**Date:** 2026-03-26

### Context

Configuration loading and saving requires file system operations. Testing file system behavior is difficult without abstraction.

### Decision

Use `spf13/afero` for filesystem operations:

```go
type Loader struct {
    logger *log.Logger
    fs     afero.Fs
}

func NewLoader(logger *log.Logger) *Loader {
    return &Loader{
        logger: logger,
        fs:     afero.NewOsFs(),
    }
}

func NewLoaderWithFS(logger *log.Logger, fs afero.Fs) *Loader {
    return &Loader{
        logger: logger,
        fs:     fs,
    }
}
```

### Consequences

**Positive:**

- Testable with in-memory filesystems (afero.MemMapFs)
- Consistent filesystem abstraction
- Easy to mock for unit tests

**Negative:**

- Additional dependency
- Slight performance overhead vs direct os calls

---

## ADR-006: Context Propagation for Cancellation

**Status:** Accepted  
**Date:** 2026-03-26

### Context

Long-running operations (like fetching linter lists) should be cancellable.

### Decision

Use `context.Context` for operations that may be long-running or need cancellation:

```go
func (l *Loader) GetAllLinterNames(ctx context.Context) ([]string, error)
func (l *Loader) CreateDefaultConfig(ctx context.Context) *Config
func (l *Loader) IsGitRepo(ctx context.Context, startDir string) bool
```

### Consequences

**Positive:**

- Operations can be cancelled
- Timeouts can be applied
- Better resource management

**Negative:**

- Context must be passed through call stack
- Must handle context cancellation explicitly

---

## ADR-007: Strong Type Aliases for Linter and Formatter Names

**Status:** Accepted  
**Date:** 2026-03-26

### Context

Linter and formatter names are strings, which are prone to typos.

### Decision

Create strong type aliases:

```go
type LinterName string

func (ln LinterName) String() string {
    return string(ln)
}

type FormatterName string

func (fn FormatterName) String() string {
    return string(fn)
}
```

### Consequences

**Positive:**

- Type safety for linter/formatter names
- Compile-time error for typos
- Self-documenting code

**Negative:**

- Need to convert when interacting with external systems
- Slight verbosity in type conversions

---

## ADR-008: Command-Line Interface with Cobra

**Status:** Accepted  
**Date:** 2026-03-26

### Context

The application is a CLI tool with multiple subcommands.

### Decision

Use `spf13/cobra` for CLI framework:

```go
cmd := &cobra.Command{
    Use:   "configure",
    Short: "Auto-configure golangci-lint",
    RunE: func(cmd *cobra.Command, _ []string) error {
        return runConfigure(cmd.Context(), logger, analyzer, configLoader, priority, preset, dryRun, configPath)
    },
}
```

### Consequences

**Positive:**

- Standard Go CLI patterns
- Built-in help, flags, and subcommand support
- Well-maintained library

**Negative:**

- Additional dependency
- Some deprecated APIs in older versions

---

## Future Considerations

### Potential Split of fixer.go

The `fixer.go` file is ~470 lines and handles multiple responsibilities:

- Version fixing
- Deprecated linter handling
- Dry-run calculations
- Actual config application

Future refactoring could extract:

- `fixer_version.go` - version field fixing
- `fixer_deprecated.go` - deprecated linter handling
- `fixer_dryrun.go` - dry-run calculations
- `fixer_apply.go` - actual application logic

### Generic ConfigResult Types

The codebase already uses `samber/mo` Result types. Further abstraction could use Go 1.18+ generics to reduce boilerplate:

```go
type ConfigResult[T any] = mo.Result[T]
```

However, the current approach with typed aliases provides sufficient type safety without complexity.
