# Post-Migration Cleanup — go-finding Published Dependency

**Date:** 2026-05-01 04:24
**Scope:** Remove go-finding local replace, simplify flake.nix, remove GOWORK=off

---

## Summary

Removed the `go-finding` local `replace` directive from `go.mod` now that the
package is published at `github.com/larsartmann/go-finding` (v0.2.1). This
simplifies the entire build chain: no sibling directory dependency, no go.work
conflicts, `GOWORK=off` no longer needed.

## Changes

### go.mod

- Removed `replace github.com/larsartmann/go-finding => ../go-finding`
- Updated to `github.com/larsartmann/go-finding v0.2.1` (published on GitHub)
- Requires `GOPRIVATE=github.com/larsartmann/go-finding` for private repo access

### flake.nix

- Kept `goFindingSrc` flake input (Nix sandbox can't fetch private repos via HTTPS)
- Simplified `postPatch`: now appends replace directive instead of sed'ing existing one
- Updated `vendorHash` to `sha256-4ooMHZbq+FnNCRoQcgofmqA1eQ8ELngMUq1Z4fnsSYI=`

### justfile

- Removed all `GOWORK=off` prefixes (no go.work conflict without local replace)
- Kept `GOTOOLCHAIN=local` (prevents auto-toolchain downloads)

### internal/cli/commands_test.go

- Added `GOPRIVATE` and `GONOSUMCHECK` env vars to `buildBinary()` helper
- Required because go-finding is a private GitHub repo and `go build` inside
  tests needs to authenticate via SSH (from `~/.gitconfig` URL rewrite)

## Verification

| Check                                | Result                                    |
| ------------------------------------ | ----------------------------------------- |
| `just build`                         | ✅                                        |
| `just test` (all 23 CLI + 12 suites) | ✅ 60.6% coverage                         |
| `just lint`                          | ✅ 0 issues                               |
| `nix build`                          | ✅ produces static binary                 |
| `nix run . -- --version`             | ✅                                        |
| `nix develop` tools                  | ✅ go, ginkgo, golangci-lint, just, templ |
