# ADR 001: YAML Dependency Decision

**Status:** Accepted  
**Date:** 2026-04-14  
**Author:** Crush AI Assistant

## Context

The project uses `go.yaml.in/yaml/v3` as its YAML library. This raised questions about:

1. Why not use the standard `gopkg.in/yaml.v3`?
2. Is this a fork with special features?
3. Is this intentional or a mistake?

## Investigation

The `go.yaml.in/yaml/v3` import path is a **vanity import URL** that resolves to:

- **Repository:** `github.com/yaml/go-yaml`
- **Branch/Tag:** `v3`
- **Current Version:** v3.0.4 (as of 2026-04-14)

This is the **official YAML v3 repository** maintained by the Go YAML team. The vanity URL `go.yaml.in` was likely chosen for:

- Cleaner import paths
- Stability (not tied to GitHub's infrastructure)
- Brand identity

## Decision

**KEEP using `go.yaml.in/yaml/v3` for the following reasons:**

1. **Official Source:** This IS the canonical source for YAML v3 in Go
2. **Maintenance:** Actively maintained (latest version from 2025-06-29)
3. **No Functional Difference:** Identical to `github.com/yaml/go-yaml`
4. **Clean Imports:** Vanity URLs are idiomatic in Go

## Consequences

### Positive

- Clean, memorable import paths
- Stable resolution (infrastructure independence)
- Official, maintained package

### Negative

- Requires explanation for contributors unfamiliar with vanity URLs
- Must be explicitly allowed in depguard rules
- Potential confusion with `gopkg.in/yaml.v3`

## Alternatives Considered

### Option 1: Migrate to `gopkg.in/yaml.v3`

- **Rejected:** This is the OLDER repository. `go.yaml.in/yaml/v3` (github.com/yaml/go-yaml) is the newer, actively maintained v3 line.

### Option 2: Migrate to `github.com/yaml/go-yaml`

- **Rejected:** The vanity URL is the documented import path for this library.

### Option 3: Use `sigs.k8s.io/yaml`

- **Rejected:** Kubernetes-specific wrapper, unnecessary dependency.

## Implementation Notes

The following files use `go.yaml.in/yaml/v3`:

- `pkg/migration/config_types.go`
- `pkg/migration/yaml_loader.go`
- `pkg/config/loader.go`

These are all appropriate usage locations for YAML parsing in this codebase.

## References

- Repository: https://github.com/yaml/go-yaml
- Go module proxy: https://proxy.golang.org/go.yaml.in/yaml/v3
- Vanity URL redirects to: https://github.com/yaml/go-yaml/tree/v3/
