# Status Report: Nix Flake Alignment & SystemNix Sync

**Date:** 2026-05-03 02:36
**Session Focus:** Align `flake.nix` with `SystemNix/pkgs/golangci-lint-auto-configure.nix`
**Previous Report:** 2026-05-01_05-28_branching-flow-analysis-complete.md

---

## Executive Summary

The project is in **excellent shape**. All tests pass (12/12 suites, 200+ specs), zero lint issues, clean build, and Nix flake builds successfully. This session focused on aligning the project's own `flake.nix` with the downstream `SystemNix` package definition — bringing consistency, reproducibility, and best practices to both.

---

## a) FULLY DONE

### This Session's Work

| Item | Status | Details |
|------|--------|---------|
| Source filtering in `flake.nix` | DONE | `cleanSourceWith` with explicit filter excludes docs, examples, scripts, .md/.yml/.lock files. Doc-only changes no longer trigger full rebuilds. |
| `proxyVendor = true` | DONE | Added to `flake.nix`. Now matches SystemNix. More reliable vendoring for large dependency trees. |
| `GOWORK = "off"` | DONE | Added to `flake.nix` env block. Safety guard against `go.work` interference. |
| Simplified go-finding patch | DONE | Replaced `cp -r` + `chmod` approach with direct nix store path reference: `replace ... => ${goFindingSrc}`. Cleaner, no file copying. |
| `subPackages` in SystemNix | DONE | Added `["cmd/golangci-lint-auto-configure"]` for explicit build scoping. |
| `vendorHash` alignment | DONE | Both `flake.nix` and SystemNix now produce identical vendor hash: `sha256-av2uH8xiTKkaYQtyb2oZNLXF4XoFfNCvvJ5D3/xmCtU=` |
| Nix build verification | DONE | `nix build` succeeds, binary runs with correct version output |
| Nix flake check | DONE | All checks passed |
| Go tests | DONE | 12/12 suites, all specs pass, 60.6% composite coverage |
| Lint | DONE | 0 issues |

### Previously Completed (Still Solid)

- **Core CLI tool** with 7 subcommands: configure, analyze, validate, report, migrate, install-hook, completion
- **119 linter priorities** with human-readable reasons
- **Formatter support** with priority levels (High, Medium, Low)
- **v1-to-v2 migration** (merged from golangci-config-migrator)
- **go-finding integration** — unified finding model, SARIF output, pipeline integration
- **Nix flake** — reproducible builds, dev shell, formatter (alejandra)
- **CI/CD pipeline** — Go 1.26, golangci-lint, Nix build, coverage upload to Codecov
- **HTML report generation** via templ
- **Project type detection** — CLI, Library, Web, API, Monorepo
- **Deprecated linter auto-replacement** (e.g., `wsl` -> `wsl_v5`)
- **go-finding published as v0.2.1** — no longer local-only dependency

---

## b) PARTIALLY DONE

| Item | Status | Gap |
|------|--------|-----|
| SystemNix version management | Partial | `flake.nix` uses dynamic `self.rev` / `"dev"`, but SystemNix pkg still has hardcoded `"0.1.0"`. Should inject from flake input metadata. |
| CI Go version matrix | Partial | CI only tests Go 1.26 now (was 1.25+1.26). AGENTS.md mentions 1.25 and 1.26 — but CI config was simplified to 1.26 only. Should clarify intended support matrix. |
| SystemNix uncommitted changes | Partial | `flake.lock` and `golangci-lint-auto-configure.nix` modified but not committed in SystemNix repo. |

---

## c) NOT STARTED

| Item | Priority | Notes |
|------|----------|-------|
| `TODO_LIST.md` | High | Does not exist. Project has 95 status reports but no consolidated TODO tracking. |
| `FEATURES.md` | High | Does not exist. No feature inventory document. |
| Release/tag v0.1.0 | Medium | Binary reports "dev" version. No git tags exist. SystemNix hardcodes "0.1.0" but no tag backs it. |
| DevShell `doCheck` in Nix | Low | Consider adding `doCheck = false` to flake.nix build (tests run via `just test`, not `nix build`) |
| Nix checks beyond build | Low | `checks.build` only — could add lint, test, format checks |
| go.work file | Low | No `go.work` exists yet, but `GOWORK = "off"` now guards against it |
| Trim old status reports | Low | 95 status reports in `docs/status/`. Consider archiving or cleaning up. |

---

## d) TOTALLY FUCKED UP

**Nothing is fucked up.** The project is in clean, working state:
- `git status`: clean, no uncommitted changes
- `just build`: succeeds
- `just test`: 12/12 suites pass
- `just lint`: 0 issues
- `nix build`: succeeds
- `nix flake check`: all checks passed

The only "close call" was the `proxyVendor` change initially causing build failures because the vendorHash was stale. This was resolved by computing the correct hash.

---

## e) WHAT WE SHOULD IMPROVE

### High Impact

1. **Create `TODO_LIST.md`** — 95 status reports but no consolidated action tracking. Hard to know what's pending without reading all reports.
2. **Create `FEATURES.md`** — Feature inventory is essential for any tool claiming to be production-ready.
3. **Dynamic version in SystemNix** — The hardcoded `"0.1.0"` in SystemNix will drift. Inject from flake input's `lastModifiedDate` or `shortRev`.
4. **Git tag v0.1.0** — If SystemNix references "0.1.0", the tag should exist in this repo.
5. **Commit SystemNix changes** — `golangci-lint-auto-configure.nix` and `flake.lock` are modified but uncommitted in SystemNix.

### Medium Impact

6. **CI Go version matrix** — Decide: support 1.25+1.26 (update CI) or 1.26 only (update AGENTS.md).
7. **Nix devShell tests** — Add `nix run .#test` or similar to CI alongside `go test`.
8. **Test coverage gaps** — CLI Commands suite at 9.4%, Set suite at 39.7%, Finding package at 58.1%. These drag the composite down.
9. **Reduce status report bloat** — 95 reports is excessive. Consider a cleanup policy (archive older than 30 days?).
10. **Flake check could be richer** — Only checks build. Add lint and format checks.

### Low Impact

11. **alejandra formatting** — `flake.nix` formatter is alejandra. Verify the file is formatted correctly.
12. **`justfile` deprecation note** — AGENTS.md says justfile is deprecated in favor of flake.nix, but this project still uses justfile extensively. Reconcile.
13. **CI Nix job requires SSH key** — Document this more prominently; Nix build will fail without `SSH_DEPLOY_KEY` secret.
14. **Templ generation** — `templ generate` not in justfile. Needs manual run when `.templ` files change.

---

## f) Top 25 Things We Should Get Done Next

| # | Item | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Commit SystemNix changes (flake.lock + .nix) | HIGH | 2min | Housekeeping |
| 2 | Create `TODO_LIST.md` from all status reports | HIGH | 1hr | Documentation |
| 3 | Create `FEATURES.md` feature inventory | HIGH | 1hr | Documentation |
| 4 | Tag v0.1.0 release | HIGH | 5min | Release |
| 5 | Dynamic version in SystemNix (inject from flake input) | HIGH | 30min | Cross-repo |
| 6 | Improve CLI Commands test coverage (9.4% -> 50%+) | HIGH | 2hr | Testing |
| 7 | Improve Set package test coverage (39.7% -> 70%+) | MEDIUM | 1hr | Testing |
| 8 | Improve Finding package coverage (58.1% -> 70%+) | MEDIUM | 1hr | Testing |
| 9 | Add Nix lint/format checks to `flake.nix` checks | MEDIUM | 30min | Nix |
| 10 | Update AGENTS.md with flake.nix changes from this session | MEDIUM | 15min | Documentation |
| 11 | Clarify Go version support matrix (1.25+1.26 vs 1.26-only) | MEDIUM | 5min | CI |
| 12 | Add `doCheck = false` to flake.nix (tests via just, not nix) | LOW | 2min | Nix |
| 13 | Archive old status reports (>30 days old) | LOW | 15min | Housekeeping |
| 14 | Add `templ generate` to justfile | LOW | 5min | Tooling |
| 15 | Reconcile justfile vs flake.nix for build tasks | LOW | 1hr | Architecture |
| 16 | Document CI SSH key requirement prominently | LOW | 5min | CI |
| 17 | Add integration test for `nix build` output | LOW | 30min | Testing |
| 18 | Verify alejandra formatting on flake.nix | LOW | 2min | Nix |
| 19 | Add `--version` flag integration test | LOW | 15min | Testing |
| 20 | Consider removing `flake-utils` dependency (use nixpkgs lib) | LOW | 30min | Nix |
| 21 | Add pre-commit hook that checks `nix flake check` | LOW | 15min | Tooling |
| 22 | Create CONTRIBUTING.md | LOW | 1hr | Documentation |
| 23 | Add CHANGELOG.md | LOW | 1hr | Documentation |
| 24 | Investigate if `go-finding` replace directive can be removed entirely | LOW | 30min | Dependencies |
| 25 | Clean up unused `client/` package (only README.md inside) | LOW | 5min | Housekeeping |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the `go-finding` local replace directive in `go.mod` be removed entirely now that go-finding is published as v0.2.1?**

Context: The `go.mod` still contains `replace github.com/larsartmann/go-finding => ../go-finding`. This means:
- Local dev works via sibling directory
- Nix builds override it via `postPatch` (both flake.nix and SystemNix)
- CI would need `../go-finding` to exist or `GOPRIVATE`/`GONOSUMCHECK` config

If go-finding v0.2.1 is stable and published, removing the replace directive would:
- Simplify `go.mod` and all Nix postPatch hacks
- Make CI simpler (no SSH key needed for go-finding)
- Remove the `GOWORK` / `GOPRIVATE` workarounds

But I cannot determine if go-finding is still under rapid development where local replace is needed for iteration speed. This is a product decision.

---

## Test Coverage Breakdown

| Suite | Specs | Coverage | Status |
|-------|-------|----------|--------|
| CLI Commands | 23 | 9.4% | PASS (low coverage) |
| Config | 39 | 66.0% | PASS |
| Experiments | 6 | 100.0% | PASS |
| Detection | 11 | 65.0% | PASS |
| Diff | 16 | 96.4% | PASS |
| Errors | 20 | 100.0% | PASS |
| Finding | 19 | 58.1% | PASS (low coverage) |
| Analyzer | 39 | 79.8% | PASS |
| Migration | 37 | 66.8% | PASS |
| Set | 20 | 39.7% | PASS (low coverage) |
| UI Formatter | 16 | 65.6% | PASS |
| Utils | 16 | 94.6% | PASS |
| **Composite** | **262** | **60.6%** | **ALL PASS** |

## Codebase Stats

| Metric | Value |
|--------|-------|
| Go source files | 91 |
| Test files | 22 |
| Total lines of Go code | 15,239 |
| Packages | 13 (pkg/) + 2 (internal/) |
| CLI commands | 7 |
| Linters tracked | 119 |
| Status reports | 95 |

## Changes This Session

### `flake.nix` (this repo)
- Added `cleanSourceWith` source filtering (26-line filter block)
- Added `proxyVendor = true`
- Updated `vendorHash` from `sha256-4ooM...` to `sha256-av2u...`
- Converted `env.CGO_ENABLED = 0` to `env = { CGO_ENABLED = 0; GOWORK = "off"; }`
- Simplified `postPatch` from `cp -r` + `chmod` to direct store path reference

### `pkgs/golangci-lint-auto-configure.nix` (SystemNix)
- Added `subPackages = ["cmd/golangci-lint-auto-configure"]`

### Build Verification
- `nix build`: succeeds
- `nix flake check`: all checks passed
- `just build`: succeeds
- `just test`: 12/12 suites, 262 specs, all pass
- `just lint`: 0 issues
