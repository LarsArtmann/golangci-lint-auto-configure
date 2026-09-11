# ADR-014: Strong Type Aliases for Linter and Formatter Names

**Status:** Accepted\
**Date:** 2026-03-26

## Context

Linter and formatter names are strings, which are prone to typos.

## Decision

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

## Consequences

**Positive:**

- Type safety for linter/formatter names
- Compile-time error for typos
- Self-documenting code

**Negative:**

- Need to convert when interacting with external systems
- Slight verbosity in type conversions

---
