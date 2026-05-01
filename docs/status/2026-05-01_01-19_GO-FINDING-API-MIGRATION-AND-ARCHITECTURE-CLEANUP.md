# Status Report: go-finding API Migration & Architecture Cleanup

**Date:** 2026-05-01 01:19
**Session Start:** 2026-04-30 ~01:40
**Total Commits This Session:** 9
**Branch:** master (pushed to origin)

---

## Executive Summary

A routine fix for `ireturn` generics false positives snowballed into a comprehensive go-finding API migration, discovering and fixing 4 compilation errors, 6 deprecated API usages, an incorrect config key name, and multiple architectural issues. The project now builds cleanly, all 11 pkg test suites pass (70.2% coverage), and `golangci-lint config verify` passes. However, `golangci-lint run` reports ~60 issues (mostly exhaustruct, tagliatelle, varnamelen, funlen).

---

## A) FULLY DONE

### 1. ireturn Generic False Positive Fix

- **Commit:** `2596df3` (initial), `dbba67a` (corrected)
- Added `generic` to ireturn's allow list in `DefaultLinterSettings`
- **Critical fix caught:** Initial commit used `accept` (ireturn standalone key), but golangci-lint v2 schema uses `allow`. Fixed in `dbba67a`.
- Both `pkg/constants/config.go` and `.golangci.yml` corrected

### 2. go-finding Builder Pattern Migration (4 compile errors fixed)

- **Commit:** `b773b5f`
- `go-finding` added required `confidence float64` parameter to `NewFinding()`
- Migrated 3 files from raw `NewFinding()` + field mutation to `finding.NewBuilder()`:
  - `pkg/finding/golangci_lint.go`
  - `pkg/finding/diff_converter.go`
  - `internal/cli/cmd_validate.go`
- All Finding construction now goes through Builder pattern

### 3. Deprecated Tag → Tags Migration

- **Commit:** `273d6a9`
- `finding.Tag` (string) deprecated in favor of `finding.Tags` ([]Tag)
- Migrated 4 `WithTag()` calls to `WithTags(finding.Tag(...))` in `converter.go`
- Updated 5 test assertions in `converter_test.go`

### 4. diff_converter Full Builder Migration

- **Commit:** `a767c24`
- Replaced post-hoc `f.BeforeCode = ...` / `f.AfterCode = ...` with Builder's `WithBeforeCode()`/`WithAfterCode()`
- Zero Finding field mutations remain anywhere in the codebase

### 5. ErrorsToFindings Extraction

- **Commit:** `7976fb8`
- Added `ErrorsToFindings(errors []error, configPath string)` to `pkg/finding/converter.go`
- Refactored `cmd_validate.go` to use it — no longer creates findings directly
- All Finding construction centralized in `pkg/finding/`

### 6. LinterToCategory Map Refactor

- **Commit:** `39f673b`
- Converted 23-statement switch to `map[string]finding.Category`
- Fixes funlen warning (23 > 20 statements)
- Data-driven: adding a linter = one map entry

### 7. ConfigFormat Exhaustive Fix

- **Commit:** `587ce3d`
- Added explicit `ConfigFormatYAML` case in `unmarshalConfig()` switch
- Default branch now returns error for unknown formats (was silently defaulting to YAML)

### 8. DefaultLinterSettings Tests

- **Commit:** `8dbf5e3`
- 4 BDD tests for `injectDefaultSettings()`:
  - Depguard defaults injected when enabled without settings
  - ireturn defaults (incl. generic) injected when enabled
  - Existing settings preserved (not overwritten)
  - Multiple linters get defaults simultaneously

### 9. Project Config Fixes

- **Commit:** `dbba67a`
- Added `github.com/larsartmann/go-finding` to depguard allow list
- Added ireturn settings to `.golangci.yml`

---

## B) PARTIALLY DONE

### 1. golangci-lint Clean Run

- **Status:** ~60 issues remaining from `golangci-lint run`
- **Breakdown:**
  | Linter | Count | Category |
  |--------|-------|----------|
  | varnamelen | 12 | Variable naming |
  | exhaustruct | 9 | Struct field exhaustiveness |
  | tagliatelle | 8 | Tag naming conventions |
  | funlen | 8 | Function length |
  | wsl_v5 | 4 | Whitespace style |
  | gosec | 3 | Security |
  | noinlineerr | 3 | Error handling |
  | gci | 2 | Import ordering |
  | wrapcheck | 2 | Error wrapping |
  | testpackage | 2 | Test package naming |
  | gochecknoglobals | 2 | Global variables |
  | forbidigo | 1 | Forbidden functions |
  | nlreturn | 1 | Newline returns |
  | revive | 1 | Various |
  | prealloc | 1 | Preallocation |
  | err113 | 1 | Dynamic errors |
- Many are in test files (exempted by exclusion rules but some slip through)

---

## C) NOT STARTED

### 1. samber/mo Removal

- `mo.Result[T]` used in `pkg/types/result.go` and `pkg/config/loader.go`
- Every `.Get()` call immediately unwraps — pure overhead
- Would remove ~100 lines of boilerplate + a dependency

### 2. Type-Safe DefaultLinterSettings

- Currently `map[string]any` — zero type safety
- Could use structured types with validation per-linter
- Would catch config key mistakes at compile time

### 3. Default Settings for More Linters

Only `depguard` and `ireturn` have defaults. Common linters that need them:

- `gocritic` — noisy without `enabled-tags`/`disabled-checks`
- `revive` — needs rule configuration
- `exhaustruct` — needs exclusion patterns
- `gomodguard` — needs `allowed`/`blocked` modules
- `funlen` — default thresholds too strict for some projects

### 4. CLI Integration Test Fix (NixOS)

- 19/23 CLI integration tests fail with `exec format error`
- Pre-existing issue — binary built in temp dir can't execute on NixOS
- Not related to any code changes

### 5. go-finding CI Publishing

- Local `replace` directive: `replace github.com/larsartmann/go-finding => ../go-finding`
- CI will fail without sibling checkout or published module

---

## D) TOTALLY FUCKED UP

### Nothing catastrophic.

The worst issue was the `accept` vs `allow` key name mistake in the initial ireturn fix (`2596df3`). The standalone ireturn linter uses `accept`, but golangci-lint v2's YAML schema wraps it as `allow`. This means any config generated by the tool before the fix (`dbba67a`) would fail `golangci-lint config verify`. Caught and fixed within the same session.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **Eliminate `map[string]any` for DefaultLinterSettings** — Replace with typed structs per linter. This prevents silently generating invalid configs (like the accept/allow bug).

2. **Remove `samber/mo` dependency** — The `mo.Result[T]` pattern adds boilerplate with zero benefit. Every call site immediately unwraps with `.Get()`. Direct `(T, error)` returns are more idiomatic Go.

3. **Consolidate Finding construction** — Already mostly done. `buildFinding()` helper is unexported and encapsulated in `pkg/finding/`. All external consumers use exported converter functions.

### Code Quality

4. **Add exhaustruct exclusions for `finding.Position`** — 9 warnings about missing `Line`, `Column`, `Offset` fields. Most call sites intentionally only set `File`. Add to `.golangci.yml` exhaustruct excludes.

5. **Add exclusion rules for test files** — Several funlen, varnamelen, gochecknoglobals warnings are in test helpers. Already partially covered but incomplete.

6. **Fix `forbidigo` violation** — `cmd_validate.go:161` uses `fmt.Println` for SARIF output. Should use a writer or structured output instead.

### Developer Experience

7. **Fix CLI integration tests on NixOS** — 19 tests fail due to binary format issue. Need to investigate the temp binary build approach.

8. **Document the `accept` vs `allow` gotcha** — Add to AGENTS.md to prevent future confusion with ireturn settings.

---

## F) Top 25 Things To Do Next

Sorted by impact × effort (highest first):

| #   | Task                                                                | Impact            | Effort |
| --- | ------------------------------------------------------------------- | ----------------- | ------ |
| 1   | Add `finding.Position` to exhaustruct exclusions in `.golangci.yml` | Clean diagnostics | 2 min  |
| 2   | Fix `forbidigo` in `cmd_validate.go` (fmt.Println → writer)         | Correctness       | 5 min  |
| 3   | Fix `err113` in `loader.go` (dynamic error → sentinel)              | Lint clean        | 5 min  |
| 4   | Add `ErrorsToFindings` to `converter.go` as exported                | Already done      | —      |
| 5   | Fix import ordering in `cmd_validate.go` (gci)                      | Clean lint        | 1 min  |
| 6   | Add default settings for `gocritic`, `exhaustruct`, `revive`        | User experience   | 30 min |
| 7   | Type-safe `DefaultLinterSettings` (replace `map[string]any`)        | Architecture      | 1 hr   |
| 8   | Remove `samber/mo` dependency                                       | Simplification    | 1 hr   |
| 9   | Document `accept` vs `allow` gotcha in AGENTS.md                    | Knowledge         | 5 min  |
| 10  | Fix `funlen` warnings (8 functions)                                 | Lint clean        | 30 min |
| 11  | Fix `tagliatelle` warnings (8 tags)                                 | Lint clean        | 20 min |
| 12  | Fix `varnamelen` warnings (12 variables)                            | Lint clean        | 30 min |
| 13  | Fix `wsl_v5` warnings (4 locations)                                 | Lint clean        | 15 min |
| 14  | Fix `gosec` warnings (3 locations)                                  | Security          | 20 min |
| 15  | Fix `noinlineerr` warnings (3 locations)                            | Lint clean        | 15 min |
| 16  | Fix `gochecknoglobals` warnings (2 locations)                       | Lint clean        | 10 min |
| 17  | Fix `testpackage` warnings (2 locations)                            | Lint clean        | 10 min |
| 18  | Fix `wrapcheck` warnings (2 locations)                              | Lint clean        | 10 min |
| 19  | Fix `nlreturn` warning (1 location)                                 | Lint clean        | 2 min  |
| 20  | Fix `revive` warning (1 location)                                   | Lint clean        | 5 min  |
| 21  | Fix `prealloc` warning (1 location)                                 | Lint clean        | 2 min  |
| 22  | Fix `fixCounts` exhaustruct in `fixer.go`                           | Lint clean        | 5 min  |
| 23  | Fix CLI integration tests on NixOS                                  | Test reliability  | 1 hr   |
| 24  | Publish `go-finding` module / fix CI                                | CI/CD             | 2 hr   |
| 25  | Update AGENTS.md with session learnings                             | Knowledge         | 15 min |

---

## G) Top #1 Question I Cannot Figure Out Myself

**The CLI integration tests fail with `exec format error` on NixOS (19/23 tests).**

The tests build a binary to a temp directory and try to execute it. On NixOS, this fails with:

```
fork/exec /tmp/ginkgo2711724685/golangci-lint-auto-configure: exec format error
```

I don't know if this is:

- A NixOS dynamic linker issue (the binary can't find its interpreter)
- A test infrastructure issue (how the binary is built/copied)
- A `go build` configuration issue
- Something that works on "normal" Linux but not NixOS

This is the single biggest blocker for CI confidence. All 11 `pkg/` test suites pass (311 specs), but the CLI integration tests are effectively broken in this environment.

---

## Test Results Summary

| Suite                     | Specs   | Status                                       |
| ------------------------- | ------- | -------------------------------------------- |
| Config                    | 39      | PASS                                         |
| Experiments               | 6       | PASS                                         |
| Errors                    | 20      | PASS                                         |
| Linter (Analyzer + Fixer) | 39      | PASS                                         |
| Migration                 | 37      | PASS                                         |
| Set                       | 20      | PASS                                         |
| Finding                   | 16      | PASS                                         |
| Detection                 | 21      | PASS                                         |
| Diff                      | 24      | PASS                                         |
| Report                    | 53      | PASS                                         |
| Utils                     | 16      | PASS                                         |
| **Total pkg**             | **311** | **ALL PASS**                                 |
| CLI Integration           | 23      | 4 PASS / 19 FAIL (NixOS `exec format error`) |

**Composite coverage:** 70.2%
**Build:** Clean
**`golangci-lint config verify`:** Pass
**`golangci-lint run`:** ~60 issues (mostly style/naming)
