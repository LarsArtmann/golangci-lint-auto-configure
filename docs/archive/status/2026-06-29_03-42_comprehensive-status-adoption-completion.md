# Comprehensive Status Report — go-error-family Adoption Completion

**Date:** 2026-06-29 03:42
**Branch:** `master` (clean, pushed to origin)
**Session commits:** 13 (`cb41ac3..a4de687`)
**Plan reference:** `docs/planning/2026-06-29_02-13_go-error-family-adoption-completion.md`

---

## A. Project Snapshot

| Metric                     | Value                                                                          |
| -------------------------- | ------------------------------------------------------------------------------ |
| Production LOC             | 11,638                                                                         |
| Test LOC                   | 8,351                                                                          |
| Direct dependencies        | 11 (3 LarsArtmann)                                                             |
| Test packages              | 19 (15 passing with tests, 4 no-test files)                                    |
| Test specs/suites          | 94                                                                             |
| CLI subcommands            | 7 (configure, analyze, validate, report, migrate, install-hook, completion)    |
| Custom error types         | 5 (ConfigError, AnalysisError, ReportError, MigrationError + domainError base) |
| Classified implementations | 3 (ConfigError, ReportError, MigrationError → Rejection)                       |
| Registered sentinels       | 11 + stdlib defaults                                                           |
| errorfamily usage sites    | 25                                                                             |
| Linters in priorities      | ~58                                                                            |
| Presets                    | 6 (minimal, standard, strict, security, performance, reference)                |
| Flake inputs               | 6 (nixpkgs, flake-parts, systems, treefmt-nix, goFindingSrc, gogenfilterSrc)   |

### Verification Gates (ALL GREEN)

| Gate      | Command                      | Result                                 |
| --------- | ---------------------------- | -------------------------------------- |
| Build     | `go build ./...`             | **PASS** (exit 0)                      |
| Vet       | `go vet ./...`               | **PASS** (clean)                       |
| Lint      | `golangci-lint run ./...`    | **0 issues** (was 34 at session start) |
| Tests     | `go test ./... -count=1`     | **All 18 test packages pass**          |
| Nix build | `nix build`                  | **PASS** (binary runs `--version`)     |
| Nix check | `nix flake check --no-build` | **all checks passed**                  |

### Coverage by Package

| Package           | Coverage  | Assessment | Δ from last report |
| ----------------- | --------- | ---------- | ------------------ |
| `pkg/errors`      | **96.6%** | Excellent  | +0.3% (new tests)  |
| `pkg/utils`       | 94.6%     | Excellent  | —                  |
| `pkg/diff`        | 94.6%     | Excellent  | —                  |
| `pkg/linter`      | 83.8%     | Good       | —                  |
| `pkg/constants`   | 80.0%     | Good       | —                  |
| `pkg/finding`     | 77.2%     | Acceptable | —                  |
| `pkg/migration`   | 75.3%     | Acceptable | —                  |
| `pkg/report`      | 72.1%     | Needs work | —                  |
| `pkg/ui`          | 67.7%     | Needs work | —                  |
| `pkg/detection`   | 67.1%     | Needs work | —                  |
| `pkg/gogenfilter` | 63.9%     | Needs work | —                  |
| `pkg/config`      | 63.4%     | Needs work | —                  |
| `pkg/types`       | 63.2%     | Needs work | —                  |
| `pkg/version`     | 51.4%     | Poor       | —                  |
| `internal/cli`    | **9.3%**  | **Crisis** | +0.5%              |
| `pkg/client`      | 0.0%      | Untested   | —                  |

---

## B. FULLY DONE (Complete & Verified)

### 1. Nix Build — Fixed & Green

**Problem:** `nix build` was broken because `go mod tidy` in `postPatch` needed network access, but the sandboxed main derivation has no DNS resolver (systemd-resolved at `127.0.0.1` is unreachable inside the sandbox).

**Root cause chain discovered:**

1. `go-finding` and `gogenfilter` are NOT cached on the Go module proxy (404) — they need direct git fetch.
2. `git ls-remote` fails inside the sandbox without auth for private repos.
3. The SSH flake inputs fetch the source, but `replace` directives make go.mod inconsistent.
4. `go mod tidy` reconciles go.mod but needs network — only the go-modules FOD has it (`__noChroot`).
5. The main derivation's `go build` then fails with "updates to go.mod needed".
6. `preBuild` (`go install templ@version + templ generate`) left a disallowed reference to the Go toolchain.

**Fix applied (commit `879cfe1`):**

- Conditional `go mod tidy`: runs only when `$name` contains `go-modules` (the FOD), skipped in the main derivation.
- `GOFLAGS=-mod=mod` set in the main derivation's `postPatch` — Go auto-reconciles from the FOD's proxy cache (no network).
- Removed `preBuild` entirely — committed `report_templ.go` to git (un-ignored `*_templ.go`).
- Added `allowGoReference = true` — Go embeds `GOROOT` in the binary.

### 2. All 34 golangci-lint Violations — Cleared (commit `1ab6dd0`)

| Linter         | Count | Fix                                                                                                            |
| -------------- | ----- | -------------------------------------------------------------------------------------------------------------- |
| makezero       | 20    | `make([]T, N)` + index-fill → `make([]T, 0, N)` + append                                                       |
| noctx          | 7     | `exec.Command` → `exec.CommandContext` in test helpers                                                         |
| nilerr         | 2     | Fixed genuine bug (file-open error silently swallowed in detector.go) + nolint for intentional walk-error skip |
| funlen         | 1     | Extracted `finalizeFixerResult` from `runFixerMode` (32→22 lines)                                              |
| wrapcheck      | 1     | Wrapped `types.ParseLinterPriority` error in `ParsePriorityParam`                                              |
| ginkgolinter   | 1     | Moved `standardTestLinters` into `BeforeEach`                                                                  |
| depguard       | 1     | Added go-error-family to `.golangci.yml` allowed imports                                                       |
| gochecknoinits | 1     | nolint for required `init()` in classification.go                                                              |

### 3. Error Classification — Complete (commits `efdd4d7`, `22651bf`)

All 4 custom error types now have explicit classification:

| Error Type       | Classification                    | Why                                                                                                          |
| ---------------- | --------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `ConfigError`    | `Classified` → Rejection (exit 1) | Always user-fault: bad file, parse error, validation failure                                                 |
| `ReportError`    | `Classified` → Rejection (exit 1) | Always user-fault: invalid format, unwritable path                                                           |
| `MigrationError` | `Classified` → Rejection (exit 1) | Always user-fault: unsupported version, malformed YAML                                                       |
| `AnalysisError`  | Cause-chain sentinels             | Heterogeneous: Infrastructure (binary-missing), Rejection (version-too-old), Corruption (unparseable output) |

**42 BDD specs** in `classification_test.go`, all passing at **96.6% coverage**.

**Architectural decision (resolves Section G):** `AnalysisError` deliberately does NOT implement `Classified` — its heterogeneous failures need fine-grained Families via cause-chain sentinels. A fixed Family would flatten the differentiation. Only always-user-fault errors get `Classified`.

### 4. Documentation — Synced & Accurate

- **README.md** — Added Exit Codes section with BSD sysexits table + CI/CD bash example.
- **FEATURES.md** — Removed justfile lie, fixed templ-generate reference, added MigrationError.
- **AGENTS.md** — Updated all 4 gotchas (nix build mechanics, templ output committed, error classification, go-finding replaces).
- **TODO_LIST.md** — Comprehensive cleanup: 16 completed items checked off, priorities re-sorted.

### 5. Pre-existing Items Verified Already Done

During investigation, discovered these "TODO" items were already implemented:

- `Config.Clone()` — exists in `pkg/types/clone.go:71`
- `DryRun bool` on `MigrationResult` — exists at `pkg/types/types.go:196`
- `errors.Join` for multi-finding — already in `pkg/finding/converter.go:211`
- `LinterMinVersions` validation test — exists in `data_integrity_test.go`
- `reference` preset validation test — exists in `data_integrity_test.go`
- `testifylint` default settings — exists in `pkg/constants/config.go:118`
- All 29 `fmt.Errorf` calls — **already use `%w`** (the "29 non-wrapping" count was a grep artifact from multi-line calls)

---

## C. PARTIALLY DONE

### 1. CLI Integration Test Coverage (9.3%)

`internal/cli` has 16 test files but coverage is only 9.3%. The existing `cmd_configure_test.go` covers configure command paths. Missing: analyze, validate, report, migrate, install-hook integration tests. The test infrastructure (binary build + exec in `test_helpers_test.go`) is solid but needs more command coverage.

### 2. Exit Code Integration Tests

Unit-level classification tests verify `Classify(err)` → Family → `ExitCode(err)` → integer. But NO test verifies the full `Main()` → `errorfamily.ExitCode(err)` → `os.Exit()` path end-to-end. This requires either a test binary or extracting `Main()` into a testable function.

### 3. `--diff` + `--check` Interaction

The `--diff` + `--check` combination has a known behavior where diff shows nothing in check mode (the tool temporarily applies changes to compute the diff, then restores). This needs either a fix or documentation. No test covers this yet.

---

## D. TOTALLY FUCKED UP (Nothing)

The working tree is clean. No broken builds, no failing tests, no stale branches. All verification gates are green. The only "fuckup" was the earlier exploration of removing `postPatch` entirely (when I incorrectly thought the repos were public on the Go proxy — they're public on GitHub but NOT cached by proxy.golang.org, causing the direct-git fallback to fail in sandbox). This was corrected.

---

## E. WHAT WE SHOULD IMPROVE

### Architecture & Code Quality

1. **`internal/cli` is a 9.3% coverage crisis.** This is the user-facing entry point. Every command path should have integration tests. The test harness is ready, the tests just need writing.

2. **`pkg/client` has 0% coverage and unclear intent.** Is it a public API? Internal helper? If unused, delete it. If public, test it.

3. **LSP cache desync persists.** 33 stale lint diagnostics from the go-finding v1.0.0 breakage (fixed in `0501f54`) still show in gopls. The actual `golangci-lint run` is clean (0 issues), so this is purely a gopls cache issue. Needs `gopls restart`.

4. **No structured error output for CI.** The CLI logs errors via `slog.Error` with a `"family"` field, but CI consumers parsing output would benefit from `--json` error mode using `errorfamily.Error.JSON()`.

5. **Domain message templates not registered.** `HandleError()` would produce structured What/Why/Fix/WayOut output, but requires `errorfamily.Error` codes that our plain sentinels don't have. This is a deeper adoption decision.

6. **Coverage gaps in low-priority packages.** `pkg/version` (51.4%), `pkg/config` (63.4%), `pkg/types` (63.2%) all have untested edge cases (especially the `ReadBuildInfo()` fallback path).

### Process & Tooling

7. **BuildFlow hook auto-commits.** The pre-commit hook (`buildflow`) auto-commits changes with generic messages, bypassing manual commit control. This caused the initial commit (`124e46d`) to have a less-detailed message than intended.

8. **No CI badge in README.** GitHub Actions CI exists (`.github/workflows/ci.yml`) but no status badge in README.

9. **`go-finding` and `gogenfilter` on Go proxy.** These repos are public on GitHub but NOT cached by `proxy.golang.org`. This forces the replace-directive + SSH-flake-input workaround. Filing a request with the Go proxy or running `GOPROXY=proxy.golang.org go mod download` from a public CI might seed the cache.

---

## F. TOP 25 THINGS TO GET DONE NEXT

### Critical (blocks trust/CI)

| # | Task                                                                     | Impact | Effort | Package        |
| - | ------------------------------------------------------------------------ | ------ | ------ | -------------- |
| 1 | Add exit-code integration tests (full `Main()` → `os.Exit()` per Family) | 5      | 3h     | `internal/cli` |
| 2 | Increase `internal/cli` coverage from 9.3% to ≥40%                       | 5      | 8h     | `internal/cli` |
| 3 | Investigate & fix `--diff` + `--check` empty-diff behavior               | 4      | 2h     | `internal/cli` |
| 4 | Add `--check` mode integration tests (exit codes, flag combos)           | 4      | 3h     | `internal/cli` |

### High (error system maturity)

| # | Task                                                                      | Impact | Effort | Package        |
| - | ------------------------------------------------------------------------- | ------ | ------ | -------------- |
| 5 | Add `--json` error output flag (`errorfamily.Error.JSON()`)               | 4      | 3h     | `internal/cli` |
| 6 | Evaluate `HandleError` at CLI boundary (decision doc)                     | 3      | 2h     | docs/          |
| 7 | Register domain message templates (requires `errorfamily.New()` adoption) | 3      | 4h     | `pkg/errors`   |
| 8 | Add `pkg/version` `ReadBuildInfo()` fallback test                         | 3      | 1h     | `pkg/version`  |

### Medium (coverage & polish)

| #  | Task                                                                 | Impact | Effort | Package           |
| -- | -------------------------------------------------------------------- | ------ | ------ | ----------------- |
| 9  | Add `analyze` command integration test                               | 3      | 2h     | `internal/cli`    |
| 10 | Add `validate` command integration test                              | 3      | 2h     | `internal/cli`    |
| 11 | Add `report` command integration test                                | 3      | 2h     | `internal/cli`    |
| 12 | Add `migrate` command integration test                               | 3      | 2h     | `internal/cli`    |
| 13 | Add `install-hook` command integration test                          | 2      | 1h     | `internal/cli`    |
| 14 | Resolve `pkg/client` intent: public API vs internal (delete or test) | 3      | 1h     | `pkg/client`      |
| 15 | Increase `pkg/config` coverage from 63.4%                            | 2      | 2h     | `pkg/config`      |
| 16 | Increase `pkg/types` coverage from 63.2%                             | 2      | 2h     | `pkg/types`       |
| 17 | Increase `pkg/gogenfilter` coverage from 63.9%                       | 2      | 2h     | `pkg/gogenfilter` |
| 18 | Add `--diff` flag integration tests                                  | 2      | 1h     | `internal/cli`    |

### Low (polish & docs)

| #  | Task                                                          | Impact | Effort | Package         |
| -- | ------------------------------------------------------------- | ------ | ------ | --------------- |
| 19 | Add CI status badge to README.md                              | 2      | 15m    | `README.md`     |
| 20 | Seed go-finding/gogenfilter on Go module proxy                | 2      | 30m    | infra           |
| 21 | Reset LSP cache (gopls stale diagnostics)                     | 1      | 5m     | env             |
| 22 | Add `ginkgolinter` default settings if any exist              | 1      | 30m    | `pkg/constants` |
| 23 | Decide `vendor/` in formatter exclusions                      | 1      | 15m    | `pkg/constants` |
| 24 | Audit all FEATURES.md "Stable" entries for accuracy (ongoing) | 1      | 1h     | `FEATURES.md`   |
| 25 | Add benchmark targets to CI (prevent perf regressions)        | 1      | 1h     | `.github/`      |

---

## G. TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

### Should we adopt `errorfamily.HandleError()` at the CLI boundary, replacing `charm.land/log`?

**Context:** Currently `Main()` does:

```go
slog.Error("command failed", "err", err, "family", errorfamily.Classify(err).String())
os.Exit(errorfamily.ExitCode(err))
```

go-error-family's `HandleError()` provides structured What/Why/Fix/WayOut output:

```
What: Your golangci-lint config is missing the required version field.
Why: golangci-lint v2 requires `version: "2"` at the top level.
Fix:  Run `golangci-lint-auto-configure configure` to auto-fix.
WayOut: See https://golangci-lint.run/usage/configuration/ for v2 format.
```

**The tradeoff:**

| Option                                               | Pros                                                                               | Cons                                                                                                                                                                   |
| ---------------------------------------------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Keep slog**                                        | Consistent with existing charm.land output, zero migration, structured JSON fields | No What/Why/Fix guidance for users, errors are terse                                                                                                                   |
| **Adopt HandleError**                                | Beautiful structured error output, actionable guidance, better UX                  | Requires registering domain message templates (4h+), replaces charm.land output (two logging systems), changes CLI output format (breaking for scripts parsing output) |
| **Hybrid: slog for success, HandleError for errors** | Best of both                                                                       | Complexity, two code paths, inconsistent                                                                                                                               |

**Why I can't decide:** This is a **product/UX decision**, not a technical one. The technical work is clear (register templates, swap the call). But whether changing the error output format is worth the user-facing disruption depends on how consumers use the CLI today. If most users are in CI/CD parsing exit codes, HandleError is unnecessary. If users run it interactively, HandleError's guidance is a significant UX win.

**What would help me decide:** Knowing the primary usage pattern — interactive human vs CI script.

---

_Generated by Crush — 2026-06-29_
