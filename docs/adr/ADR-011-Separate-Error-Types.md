# ADR-011: Separate Error Types for Different Domains

**Status:** Accepted\
**Date:** 2026-03-26

## Context

The application has distinct error domains: configuration, analysis, and reporting. Generic errors lose important context.

## Decision

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

## Consequences

**Positive:**

- Rich error context with file paths, config names, etc.
- Type-safe error handling
- Easier debugging with structured error information

**Negative:**

- More error types to maintain
- Slightly more verbose error creation

---
