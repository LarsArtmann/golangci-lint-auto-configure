# ADR-012: Filesystem Abstraction for Testability

**Status:** Accepted — updated from original Afero-based design to a simpler custom interface.
**Date:** 2026-03-26

## Context

Configuration loading and saving requires file system operations. Testing file system behavior is difficult without abstraction.

## Decision

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

## Consequences

**Positive:**

- Testable with mock implementations of the `FS` interface
- No third-party dependency — keeps the dependency tree small
- Clear, minimal contract for what filesystem operations are needed

**Negative:**

- Custom interface must be maintained if new fs operations are needed

---
