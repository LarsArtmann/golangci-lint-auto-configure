# Comprehensive Status Report — go-error-family Adoption

**Generated:** 2026-06-28 22:49
**Branch:** master @ `0501f54`
**Build:** ✅ Clean — `go build ./...` exits 0
**Tests:** ✅ All 15 packages green
**Coverage:** Range 8.8%–96.3% (weighted avg ~68%)

---

## Project Snapshot

| Metric                                | Value                                                                       |
| ------------------------------------- | --------------------------------------------------------------------------- |
| Production LOC                        | 11,847                                                                      |
| Test LOC                              | 8,313                                                                       |
| Direct dependencies                   | 14 (2 LarsArtmann private)                                                  |
| Error construction sites              | 173 (129 `fmt.Errorf`, 43 `errors.New`, 1 `errors.Join`)                    |
| `apperrors`/`errorfamily` usage sites | 37                                                                          |
| Test packages                         | 15 (all passing)                                                            |
| CLI subcommands                       | 7 (configure, analyze, validate, report, migrate, install-hook, completion) |

### Coverage by Package

| Package           | Coverage | Assessment       |
| ----------------- | -------- | ---------------- |
| `pkg/errors`      | 96.3%    | Excellent        |
| `pkg/diff`        | 94.6%    | Excellent        |
| `pkg/utils`       | 94.6%    | Excellent        |
| `pkg/linter`      | 83.9%    | Good             |
| `pkg/constants`   | 80.0%    | Good             |
| `pkg/finding`     | 77.2%    | Acceptable       |
| `pkg/migration`   | 75.3%    | Acceptable       |
| `pkg/report`      | 71.9%    | Needs work       |
| `pkg/ui`          | 67.7%    | Needs work       |
| `pkg/detection`   | 67.1%    | Needs work       |
| `pkg/gogenfilter` | 63.9%    | Needs work       |
| `pkg/config`      | 63.4%    | Needs work       |
| `pkg/types`       | 63.2%    | Needs work       |
| `pkg/version`     | 51.4%    | Poor             |
| `internal/cli`    | 8.8%     | **Critical gap** |

---

## A. FULLY DONE ✅

### 1. go-error-family v0.5.1 Dependency — Integrated

**Commit:** `efdd4d7`
**Files:** `go.mod`, `go.sum`

- Added `github.com/larsartmann/go-error-family v0.5.1` as direct dependency.
- `go mod tidy` clean. No local replace needed (published version).

### 2. Sentinel Classification Registration — Complete

**File:** `pkg/errors/classification.go`

All 11 sentinel errors + stdlib defaults registered with Families:

| Sentinel                    | Family         | Exit Code | Rationale                    |
| --------------------------- | -------------- | --------- | ---------------------------- |
| `ErrNotGitRepository`       | Rejection      | 1         | User must be in a git repo   |
| `ErrNotInGitWorkingTree`    | Rejection      | 1         | User must be in working tree |
| `ErrUnknownPreset`          | Rejection      | 1         | User passed invalid preset   |
| `ErrInvalidActivityContext` | Rejection      | 1         | Programming error            |
| `ErrVersionTooOld`          | Rejection      | 1         | User must upgrade            |
| `ErrConfigValidationFailed` | Rejection      | 1         | User's config is invalid     |
| `ErrHookAlreadyExists`      | Conflict       | 1         | State conflict               |
| `ErrChangesNeeded`          | Conflict       | 1         | Check-mode signal            |
| `ErrVersionParse`           | Corruption     | 65        | Broken installation output   |
| `ErrInvalidVersionFormat`   | Corruption     | 65        | Malformed version string     |
| `exec.ErrNotFound`          | Infrastructure | 69        | Binary missing from PATH     |

Also calls `RegisterStdlibDefaults` for context/sql/os errors.

### 3. ConfigError Classified Interface — Implemented

**File:** `pkg/errors/classification.go`

`ConfigError` implements `errorfamily.Classified` → always returns `Rejection`.
This is checked BEFORE sentinels in the classification chain, which is correct:
config errors are always user-fault regardless of the underlying cause.

### 4. CLI Boundary Exit Codes — Wired

**File:** `internal/cli/commands.go` (commit `efdd4d7`)

`Main()` now:

1. Classifies the error via `errorfamily.Classify(err)`
2. Logs the error with its Family for observability: `slog.Error("CLI execution failed", "error", err, "family", family.String())`
3. Exits with `errorfamily.ExitCode(err)` instead of hardcoded `os.Exit(1)`

**Before:** Every error → exit 1.
**After:** Rejection→1, Conflict→1, Transient→75, Corruption→65, Infrastructure→69.

### 5. Classification Tests — Full BDD Coverage

**File:** `pkg/errors/classification_test.go` (commit `efdd4d7`)

24 Ginkgo tests covering:

- All 11 sentinel → Family mappings (DescribeTable)
- All 4 exit code assertions (DescribeTable)
- Wrapped sentinel chain walking (double-wrapped)
- ConfigError Classified override (even when wrapping Infrastructure cause)
- AnalysisError classification via sentinel in cause chain
- Nil error → exit 0

**Result:** `pkg/errors` coverage at 96.3%.

### 6. go-finding v1.0.0 API Breakage — Fixed (parallel commit)

**Commit:** `0501f54`

The go-finding v1.0.0 bump introduced branded types (`RuleName`, `ToolName`) that
broke 9 call sites. All fixed. Build clean, all tests pass.

### 7. AGENTS.md Rewrite — Lean + Correct

**Commit:** `0501f54` (parallel agent)

Rewritten from ~290 bloated lines to ~75 lines. Critical fix: removed all `just`
command references (no justfile exists). My go-error-family documentation entries
survived the rewrite (lines 39 and 58).

---

## B. PARTIALLY DONE ⚠️

### 1. Error Family Deep Adoption — Boundary Only

The classification system is wired at the CLI boundary (`Main()`), but the 129
`fmt.Errorf` construction sites throughout the codebase still create generic
errors without Family classification. They work because sentinels in their cause
chains are classified, but:

- **No error codes** — most `fmt.Errorf` calls don't attach a machine-readable code
- **No structured context** — errors carry string messages but not key-value context
- **No message templates** — go-error-family's What/Why/Fix/WayOut system is unused

**Current state:** 37 `apperrors`/`errorfamily` usage sites vs 173 total error constructors (21% adoption).

### 2. Classified Interface — Only ConfigError

Only `ConfigError` implements `Classified`. The other three custom error types
(`AnalysisError`, `ReportError`, `MigrationError`) do NOT implement it — they
rely entirely on sentinel classification through their cause chains. This works
but is fragile: an `AnalysisError` without a recognizable sentinel defaults to
Transient, which may not be correct.

### 3. Exit Code Test Assertions — Stale

`internal/cli/cmd_configure_test.go:149` asserts `exitErr.ExitCode()` equals 1.
This still passes (Rejection and Conflict both map to exit 1), but tests don't
verify the NEW differentiated codes (65 for Corruption, 69 for Infrastructure).

### 4. Documentation — Partially Updated

- `AGENTS.md` — updated with go-error-family section ✅
- `FEATURES.md` — has error handling section ✅, but still references "justfile recipes" as Stable ❌
- `README.md` — no mention of exit codes or error families ❌
- `docs/references/error-handling.md` — not checked/updated ❌

---

## C. NOT STARTED ❌

1. **`flake.nix` vendorHash update** — Nix build WILL FAIL. The hash doesn't account for go-error-family's transitive dependencies.
2. **`nix build` verification** — not run since go-error-family was added.
3. **`nix flake check`** — not run.
4. **`golangci-lint run`** — lint pass not verified after all changes.
5. **Domain message templates** — go-error-family's `RegisterTemplate` system for What/Why/Fix/WayOut is completely unused. No domain-specific error messages.
6. **AnalysisError/ReportError/MigrationError Classified** — only ConfigError done.
7. **HandleError integration** — the library's `HandleError()` provides structured CLI output (What/Why/Fix), but we only use `ExitCode()`. The project uses `charm.land/log` instead.
8. **errors.Join for multi-finding failures** — TODO_LIST mentions this. Converter returns first error only.
9. **Exit code documentation** — CI/CD consumers have no reference for what exit 65 vs 69 vs 75 means.
10. **JSON error output** — `errorfamily.Error.JSON()` exists for API boundaries but is unused.
11. **`--verbose`/`--quiet` flags** — no control over error output detail level.
12. **LSP cache reset** — LSP still shows 13 stale errors from the go-finding breakage that was fixed in `0501f54`.

---

## D. TOTALLY FUCKED UP 💥

### 1. FEATURES.md "justfile recipes | Stable" — ACTIVELY WRONG

**File:** `FEATURES.md:163`

The table says:

```
| justfile recipes | Stable | Primary build interface |
```

**THERE IS NO JUSTFILE.** Commit `0501f54` established this and purged `just`
references from AGENTS.md, README.md, and docs/references/. But FEATURES.md was
NOT updated. Anyone reading FEATURES.md will try `just build` and get
`error: no justfile found`.

### 2. LSP Diagnostics Cache — Poisoned

The LSP shows 13 compiler errors in `pkg/finding/` and `internal/cli/cmd_validate.go`
that were FIXED in commit `0501f54`. `go build ./...` passes clean. The LSP
has not re-indexed since the fix. This creates false alarm and makes it hard to
distinguish real errors from stale ones.

### 3. Git Stat-Cache Desync

Git status reported `go.sum` as modified, but `git diff HEAD -- go.sum` shows
0 lines. This is a filesystem stat-cache desync — the file content is identical
to HEAD but the stat metadata changed (likely from `go mod tidy` touching the
file without changing content).

---

## E. WHAT WE SHOULD IMPROVE 🎯

### Error Architecture

1. **Eliminate 29 non-wrapping `fmt.Errorf` calls** — these break error chains by not using `%w`. Callers can't `errors.Is` or `errors.As` through them. Every `fmt.Errorf` with a cause error should use `%w`.

2. **Implement Classified on remaining error types** — AnalysisError, ReportError, MigrationError should all implement `Classified` like ConfigError. The question is: what Family? (See Section G.)

3. **Register domain message templates** — `config.not_found`, `version.too_old`, `git.not_repository` etc. should have What/Why/Fix templates so `HandleError` can produce structured CLI output instead of raw error strings.

4. **Consider HandleError for the CLI boundary** — currently `Main()` does `slog.Error(...)`. go-error-family's `HandleError()` provides What/Why/Fix/WayOut formatted for users. The tradeoff: it replaces charm.land/log output with its own format.

### Test Coverage

5. **`internal/cli` at 8.8% coverage is a crisis** — this is the user-facing entry point. Every command path should have integration tests. The existing `cmd_configure_test.go` integration tests are good but only cover configure.

6. **Exit code integration tests** — no test verifies that a Corruption error produces exit 65, or Infrastructure produces exit 69. The unit tests in `classification_test.go` verify classification, but not the full `Main()` → `os.Exit()` path.

7. **`pkg/version` at 51.4%** — the `runtime/debug.ReadBuildInfo()` fallback paths are untested.

### Code Quality

8. **FEATURES.md audit** — at minimum, fix the "justfile recipes" lie. Audit all "Stable" entries for accuracy.

9. **Multi-error handling** — `pkg/finding/converter.go` returns the first error only. Should use `errors.Join` and let `Classify` pick the worst Family.

10. **Error context propagation** — most `fmt.Errorf` calls add string context but not key-value context. go-error-family's `WithContext("path", path)` enables structured logging and diagnostic rules.

---

## F. TOP 25 THINGS TO GET DONE NEXT 📋

### Critical (blocks CI/CD)

1. **Fix FEATURES.md "justfile recipes"** → change to "Nix flake build | Stable"
2. **Update `flake.nix` vendorHash** for go-error-family — `nix build` is broken until this is done
3. **Run `nix build`** to verify the full Nix pipeline works with the new dependency
4. **Run `nix flake check`** — verify all Nix checks pass
5. **Run `golangci-lint run`** — verify lint is clean after all changes

### High (error system maturity)

6. **Fix the 29 non-wrapping `fmt.Errorf` calls** — add `%w` where there's a cause error
7. **Implement `Classified` on `AnalysisError`** — decide default Family (see Section G)
8. **Implement `Classified` on `ReportError`** → Rejection (report failures are user-fault)
9. **Implement `Classified` on `MigrationError`** → Rejection (migration failures are user-fault)
10. **Register domain message templates** — `config.not_found`, `version.too_old`, `git.not_repository`
11. **Add exit code integration tests** — verify full `Main()` → `os.Exit()` path for each Family
12. **Document exit codes in README.md** — table of exit code → meaning for CI/CD consumers

### Medium (test coverage)

13. **Increase `internal/cli` coverage** from 8.8% to ≥40% — add integration tests for analyze, validate, report, migrate
14. **Increase `pkg/version` coverage** from 51.4% — test `ReadBuildInfo()` fallback
15. **Add `--check` mode integration tests** — exit codes, flag combinations
16. **Add `--diff` + `--check` interaction test** — diff shows nothing in check mode (bug?)
17. **Add `errors.Join` for multi-finding failures** in converter.go

### Low (polish)

18. **Add `Config.Clone()` method** — replace JSON marshal/unmarshal hack
19. **Add `DryRun bool` field on `MigrationResult`**
20. **Add `pkg/client` smoke tests** or clarify intent (public API vs internal)
21. **Add `testifylint` default settings** (TODO_LIST)
22. **Evaluate `HandleError` for structured CLI output** — What/Why/Fix/WayOut at the boundary
23. **Add `--json` error output flag** — machine-readable errors for scripts using `errorfamily.Error.JSON()`
24. **Audit all "Stable" entries in FEATURES.md** for accuracy
25. **Reset LSP cache** — stale diagnostics showing fixed errors

---

## G. TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF 🤔

### Should AnalysisError implement `Classified` at the type level, or keep delegating to its cause chain?

**Context:** `ConfigError` implements `Classified` → always `Rejection`. This is correct because config errors are ALWAYS user-fault regardless of underlying cause. The classification is deterministic and context-free.

**The problem with AnalysisError:** It's used for semantically different failures:

| Scenario                           | Current Classification   | Via What?                   |
| ---------------------------------- | ------------------------ | --------------------------- |
| golangci-lint not in PATH          | Infrastructure (exit 69) | `exec.ErrNotFound` sentinel |
| Version too old                    | Rejection (exit 1)       | `ErrVersionTooOld` sentinel |
| Version output unparseable         | Corruption (exit 65)     | `ErrVersionParse` sentinel  |
| Generic golangci-lint run failure  | Transient (exit 75)      | Default (no sentinel match) |
| JSON parse failure of linters list | Transient (exit 75)      | Default (no sentinel match) |

If I implement `Classified` returning a FIXED family on `AnalysisError`:

- ✅ Deterministic — every AnalysisError classifies the same way
- ❌ LOSES the sentinel-based differentiation that currently gives us exit 65 for Corruption and exit 69 for Infrastructure

If I DON'T implement it (current state):

- ✅ Cause-chain sentinels give correct fine-grained classification
- ❌ AnalysisError without a sentinel cause defaults to Transient — is that right?

**The core question: what should the DEFAULT family be for an AnalysisError that has no registered sentinel in its cause chain?**

- **Transient** (current): "golangci-lint crashed, maybe retry will work" — fail-open
- **Rejection**: "your config/project caused the analysis to fail" — user-fault
- **Infrastructure**: "the analysis tooling is broken" — system-fault

I lean toward keeping the current approach (no `Classified` on AnalysisError, rely on sentinels + Transient default) because analysis failures are genuinely heterogeneous. But I can't decide without knowing the user's intent for CI/CD consumers: do they RETRY on exit 75, or do they treat all non-zero exits the same?

**What I need from you:** A decision on the adoption strategy. Should we:

- **(A) Stay boundary-only** — exit codes work, sentinels handle known cases, don't touch the 129 `fmt.Errorf` calls
- **(B) Deep adopt** — migrate `fmt.Errorf` → `errorfamily.Wrap*()` with codes, context, and message templates across all 129 sites
- **(C) Targeted adopt** — migrate only the ~40 call sites in `pkg/linter/` and `pkg/config/` (the most error-dense packages), leave the rest as-is

This determines the entire roadmap.

---

## Commit History (This Session)

| Commit    | Description                                                           |
| --------- | --------------------------------------------------------------------- |
| `efdd4d7` | feat(errors): integrate go-error-family for semantic CLI exit codes   |
| `0501f54` | fix: repair go-finding v1.0.0 API breakage and restore lean AGENTS.md |

---

_Generated by Crush — comprehensive status review session._
