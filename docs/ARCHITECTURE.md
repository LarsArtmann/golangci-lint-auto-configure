# Architecture Decision Records

This document captures key architectural decisions made during the development of golangci-lint-auto-configure.

## ADR-001: Railway-Oriented Programming with Result Types

**Status:** Superseded — the `samber/mo` Result types were removed in favor of standard Go error returns. Kept as a historical record of the experiment.
**Date:** 2026-03-26

### Context

The codebase initially explored railway-oriented programming (ROP) patterns for handling operations that can fail.

### Decision

We used the `samber/mo` library which provided `Result[T]` types that wrap successful values or errors. Helper functions in `pkg/types/result.go` created typed result aliases.

### Consequences

**Positive:**

- Explicit error handling at each step
- Composable operations

**Negative:**

- Additional verbosity in calling code
- Learning curve for contributors unfamiliar with ROP

**Superseded by:** Standard Go `(value, error)` returns — simpler, more idiomatic, and what every Go developer expects. The ROP experiment added complexity without enough benefit to justify the non-standard pattern.

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
    FixesApplied int      // tag-free: PascalCase in JSON via json/v2
    Message      string
    NextSteps    []string `json:",omitempty"`
    DryRun       bool
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

## ADR-005: Filesystem Abstraction for Testability

**Status:** Accepted — updated from original Afero-based design to a simpler custom interface.
**Date:** 2026-03-26

### Context

Configuration loading and saving requires file system operations. Testing file system behavior is difficult without abstraction.

### Decision

Use a minimal `FS` interface (defined in `pkg/config/loader.go`) instead of a third-party filesystem library:

```go
type FS interface {
    ReadFile(name string) ([]byte, error)
    WriteFile(name string, data []byte, perm os.FileMode) error
    Stat(name string) (os.FileInfo, error)
    MkdirAll(path string, perm os.FileMode) error
}

type Loader struct {
    logger *log.Logger
    fs     FS
}

func NewLoader(logger *log.Logger) *Loader {
    return &Loader{
        logger: logger,
        fs:     osFS{},
    }
}

func NewLoaderWithFS(logger *log.Logger, fs FS) *Loader {
    return &Loader{
        logger: logger,
        fs:     fs,
    }
}
```

### Consequences

**Positive:**

- Testable with mock implementations of the `FS` interface
- No third-party dependency — keeps the dependency tree small
- Clear, minimal contract for what filesystem operations are needed

**Negative:**

- Custom interface must be maintained if new fs operations are needed

---

## ADR-006: Context Propagation for Cancellation

**Status:** Accepted  
**Date:** 2026-03-26

### Context

Long-running operations (like fetching linter lists) should be cancellable.

### Decision

Use `context.Context` for operations that may be long-running or need cancellation:

```go
func (l *Loader) getAllLinterNames(ctx context.Context) ([]string, error)
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
    Short: "Automatically configure and optimize golangci-lint",
    RunE: func(cmd *cobra.Command, _ []string) error {
        return runConfigure(cmd.Context(), logger, analyzer, configLoader,
            priority, preset, dryRun, configPath, check, showDiff)
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

### Fixer Module Split

The fixer has already been split into focused files:

- `fixer.go` — core orchestration
- `fixer_deprecated.go` — deprecated linter handling
- `fixer_config.go` — default settings injection
- `fixer_enforce.go` — disable-reason sidecar enforcement
- `fixer_formatters.go` — formatter management
- `fixer_preflight.go` — pre-flight normalization
- `fixer_recorder.go` — change counting/recording
- `fixer_results.go` — result aggregation
- `fixer_audit.go` — audit ledger integration

Further refinement could extract version-fixing and dry-run logic into their own files if `fixer.go` grows again.

### Generic Result Types

The codebase uses standard Go `(value, error)` returns. If a structured Result type becomes beneficial for carrying warnings/counts alongside errors, a project-specific `Result[T]` type (not `samber/mo`) could be introduced — see TODO_LIST.md for the proposal.
