# ADR-013: Context Propagation for Cancellation

**Status:** Accepted\
**Date:** 2026-03-26

## Context

Long-running operations (like fetching linter lists) should be cancellable.

## Decision

Use `context.Context` for operations that may be long-running or need cancellation:

```go
func (l *Loader) getAllLinterNames(ctx context.Context) ([]string, error)
func (l *Loader) CreateDefaultConfig(ctx context.Context) *Config
func (l *Loader) IsGitRepo(ctx context.Context, startDir string) bool
```

## Consequences

**Positive:**

- Operations can be cancelled
- Timeouts can be applied
- Better resource management

**Negative:**

- Context must be passed through call stack
- Must handle context cancellation explicitly

---
