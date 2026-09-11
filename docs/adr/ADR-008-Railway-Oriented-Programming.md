# ADR-008: Railway-Oriented Programming with Result Types

**Status:** Superseded — the `samber/mo` Result types were removed in favor of standard Go error returns. Kept as a historical record of the experiment.
**Date:** 2026-03-26

## Context

The codebase initially explored railway-oriented programming (ROP) patterns for handling operations that can fail.

## Decision

We used the `samber/mo` library which provided `Result[T]` types that wrap successful values or errors. Helper functions in `pkg/types/result.go` created typed result aliases.

## Consequences

**Positive:**

- Explicit error handling at each step
- Composable operations

**Negative:**

- Additional verbosity in calling code
- Learning curve for contributors unfamiliar with ROP

**Superseded by:** Standard Go `(value, error)` returns — simpler, more idiomatic, and what every Go developer expects. The ROP experiment added complexity without enough benefit to justify the non-standard pattern.

---
