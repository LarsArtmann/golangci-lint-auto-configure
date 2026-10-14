# ADR-018: Two Version Axes with Branded Types

**Status:** Accepted\
**Date:** 2026-10-14

## Context

The codebase carried two version concepts as bare strings:

1. **golangci-lint versions** (`v2.12.0`) — gate settings availability,
   supersession, and the tool minimum.
2. **Go versions** (`1.27`, sometimes `1.27.1`) — gate language relevance
   (`run.go`, `Lang` intervals).

Both were compared with ad-hoc `semver` helpers and string surgery. The
resulting split brain: `NormalizeGoMajorMinor` returned `(string, bool)`,
comparisons lived in three packages, and nothing prevented passing a Go
version where a tool version was expected (the two share a format prefix but
different semantics — patch levels matter for one and not the other).

## Decision

Two **branded types** with distinct, minimal semantics:

- `types.GolangciVersion` — full semver core required, parse-only
  (`ParseGolangciVersion`), `AtLeast`/`Compare` methods. Patch levels are
  significant.
- `types.GoVersion` — major.minor only by construction: `ParseGoVersion`
  DROPS the patch level ("1.27.1" → "1.27") so a toolchain patch bump never
  rewrites `run.go`.

Cross-axis misuse is a compile error: `Interval[GolangciVersion]` and
`Interval[GoVersion]` are distinct instantiations of the generic
`Interval[V Ordered[V]]`, and an `Ordered[V]` constraint (comparable +
`Compare` + `String`) keeps the interval machinery generic without interface
boxing.

## Consequences

- `run.go` capping (`capRunGo`), min-version derivation
  (`DerivedMinToolVersion`), and validity gating all compare through one
  mechanism.
- The old string shims (`NormalizeGoMajorMinor`) are deprecated and their
  call sites migrated; the config wire format (`run.go: "1.27"`) is
  unchanged — the branding is in-memory only.
- Fuzz targets cover both parsers (`version_fuzz_test.go`).

## References

- `pkg/types/golangci_version.go`, `pkg/types/go_version.go`,
  `pkg/types/interval.go`
- Plan phase P1 (version kernel), T03-T05
