# ADR-009: MigrationResult Uses Error Field Instead of Success Bool

**Status:** Accepted\
**Date:** 2026-03-26

## Context

Previously, `MigrationResult` had a `Success bool` field. This required checking both the returned error AND the Success field to determine if an operation succeeded.

## Decision

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

## Consequences

**Positive:**

- Single source of truth for success/failure
- Error information preserved in the result struct
- Cleaner conditional logic (`if result.IsSuccess()` vs `if err == nil && result.Success`)

**Negative:**

- Requires updating all call sites that checked `Success` field

---
