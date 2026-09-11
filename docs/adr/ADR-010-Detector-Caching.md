# ADR-010: Detector Caching for Project Type Analysis

**Status:** Accepted\
**Date:** 2026-03-26

## Context

The `Detector.Detect()` method performs expensive file system walks to analyze project structure. If called multiple times, it would redundantly re-scan the file system.

## Decision

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

## Consequences

**Positive:**

- Eliminates redundant file system operations
- Thread-safe for concurrent access
- Transparent to callers

**Negative:**

- Slight memory overhead for cached value
- Cache cannot be invalidated (acceptable for CLI use case)

---
