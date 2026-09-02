# Status Report: 2026-06-05 02:20 — Comprehensive Project Audit

## Executive Summary

Project is in strong shape. **All 15 test suites pass (61.1% composite coverage).** Build succeeds. Zero code duplication at threshold 50. Zero TODO/FIXME comments. `go vet` clean. **9 lint issues remain** (7 funlen, 1 noinlineerr, 1 varnamelen) — all pre-existing, none introduced since last report.

**Codebase:** 74 production files (10,811 lines), 24 test files (6,557 lines), 12 direct dependencies, 37 total. Version v0.2.0-12-gbfd0a12.

---

## a) FULLY DONE ✅

### CLI Commands (7/7)

| Command        | Status  | Coverage |
| -------------- | ------- | -------- |
| `configure`    | ✅ DONE | 8.2%\*   |
| `analyze`      | ✅ DONE | —        |
| `validate`     | ✅ DONE | —        |
| `report`       | ✅ DONE | —        |
| `migrate`      | ✅ DONE | —        |
| `install-hook` | ✅ DONE | —        |
| `completion`   | ✅ DONE | —        |

\*CLI integration tests run actual binary, coverage appears low but functionality is well-tested through `pkg/linter` (81.8%).

### Core Features (All Complete)

| Feature                                         | Status     | Package           |
| ----------------------------------------------- | ---------- | ----------------- |
| 119 linter priorities (4 tiers)                 | ✅ DONE    | `pkg/constants`   |
| Priority-based filtering (`--priority`)         | ✅ DONE    | `internal/cli`    |
| Dry-run mode (`--dry-run`)                      | ✅ DONE    | `internal/cli`    |
| CI check mode (`--check`)                       | ✅ DONE    | `internal/cli`    |
| Diff preview (`--diff`)                         | ✅ DONE    | `internal/cli`    |
| Deprecated linter auto-replacement              | ✅ DONE    | `pkg/linter`      |
| Version-gated deprecation (gomodguard_v2)       | ✅ DONE    | `pkg/linter`      |
| Typecheck linter removal                        | ✅ DONE    | `pkg/linter`      |
| Invalid timeout fix                             | ✅ DONE    | `pkg/linter`      |
| 9 default linter settings injection             | ✅ DONE    | `pkg/linter`      |
| Default formatter settings (golines)            | ✅ DONE    | `pkg/linter`      |
| Reference preset (60+ linters)                  | ✅ DONE    | `pkg/constants`   |
| gogenfilter/v3 dynamic scan (8+ generators)     | ✅ DONE    | `pkg/gogenfilter` |
| `generated: lax` auto-set                       | ✅ DONE    | `pkg/linter`      |
| Exclusion path deduplication                    | ✅ DONE    | `pkg/gogenfilter` |
| GOEXPERIMENT build tags auto-injection (5 tags) | ✅ DONE    | `pkg/linter`      |
| Go version auto-detection                       | ✅ DONE    | `pkg/config`      |
| Multiple golangci-lint binary detection         | ✅ DONE    | `pkg/linter`      |
| Version-aware feature flags                     | ✅ DONE    | `pkg/constants`   |
| Project type detection (5 types)                | ✅ DONE    | `pkg/detection`   |
| swaggo formatter detection (6 patterns)         | ✅ DONE    | `pkg/detection`   |
| Custom error types (4 domain types)             | ✅ DONE    | `pkg/errors`      |
| Result[T] railway-oriented type                 | ❌ REMOVED | `pkg/types`       |

### Default Settings Injection (9 linters)

depguard, ireturn, gocritic, exhaustruct, revive, varnamelen, gomoddirectives, cyclop, golines — all ✅ DONE.

### go-finding Integration (Complete)

LinterRecommendation/ValidationError/Diff/golangci-lint JSON → finding.Finding conversions, SARIF 2.1.0, Report JSON — all ✅ DONE.

### Migration v1 → v2 (Complete)

issues.exclude-rules → linters.exclusions.rules, directory/file exclusions, linter-specific settings, removed settings cleanup — all ✅ DONE.

### Reports (4 formats)

HTML (templ), JSON, SARIF, finding JSON — all ✅ DONE.

### Build & CI

| Item                     | Status  |
| ------------------------ | ------- |
| Nix flake build          | ✅ DONE |
| justfile recipes         | ✅ DONE |
| GitHub Actions CI matrix | ✅ DONE |
| Pre-commit hooks         | ✅ DONE |
| templ generate in Nix    | ✅ DONE |
| Version via ldflags      | ✅ DONE |
| Auto-tag workflow        | ✅ DONE |

### Code Quality

| Item                          | Status                     |
| ----------------------------- | -------------------------- |
| Semantic deduplication audit  | ✅ DONE (0 clones at t=50) |
| Zero TODO/FIXME comments      | ✅ DONE                    |
| `go vet` clean                | ✅ DONE                    |
| Custom `Result[T]` (no mo)    | ✅ DONE                    |
| Custom `config.FS` (no afero) | ✅ DONE                    |
| `report_templ.go` untracked   | ✅ DONE                    |

---

## b) PARTIALLY DONE ⚠️

### 1. Test Coverage (Overall 61.1%)

| Package            | Coverage | Verdict      |
| ------------------ | -------- | ------------ |
| `pkg/diff`         | 96.5%    | ✅ Excellent |
| `pkg/errors`       | 95.8%    | ✅ Excellent |
| `pkg/utils`        | 94.6%    | ✅ Excellent |
| `pkg/constants`    | 80.0%    | ✅ Good      |
| `pkg/linter`       | 81.8%    | ✅ Good      |
| `pkg/report`       | 71.9%    | ⚠️ Adequate   |
| `pkg/ui`           | 67.7%    | ⚠️ Adequate   |
| `pkg/migration`    | 66.8%    | ⚠️ Adequate   |
| `pkg/config`       | 64.5%    | ⚠️ Adequate   |
| `pkg/detection`    | 62.8%    | ⚠️ Adequate   |
| `pkg/gogenfilter`  | 59.8%    | ⚠️ Needs work |
| `pkg/types`        | 58.8%    | ⚠️ Needs work |
| `pkg/finding`      | 50.0%    | ⚠️ Needs work |
| `pkg/version`      | 51.4%    | ⚠️ Needs work |
| `internal/cli`     | 8.2%     | 🔴 Critical  |
| `pkg/client`       | 0.0%     | 🔴 No tests  |
| `internal/cli/cmd` | 0.0%     | 🔴 No tests  |

**~45 exported functions lack dedicated direct tests** (many have indirect coverage through integration tests).

### 2. Lint Cleanliness (9 issues remain)

| File                                | Linter      | Details                                     |
| ----------------------------------- | ----------- | ------------------------------------------- |
| `internal/cli/cmd_configure.go:70`  | funlen      | `newConfigureCommand` 33 > 30 lines         |
| `internal/cli/cmd_configure.go:164` | funlen      | `runConfigure` 31 > 30 lines                |
| `internal/cli/cmd_configure.go:217` | funlen      | `runFixerMode` 46 > 30 lines                |
| `pkg/config/loader.go:316`          | funlen      | `CreateDefaultConfig` 33 > 30 lines         |
| `pkg/finding/converter.go:128`      | funlen      | `DeprecatedLintersToFindings` 32 > 30 lines |
| `pkg/finding/diff_converter.go:11`  | funlen      | `ChangesToFindings` 34 > 30 lines           |
| `pkg/linter/categorizer.go:33`      | funlen      | `shouldSkipLinter` 32 > 30 lines            |
| `pkg/detection/detector.go:142`     | noinlineerr | Inline error in `if` condition              |
| `internal/cli/cmd_report.go:176`    | varnamelen  | Variable `r` too short                      |

### 3. `--diff` in `--check` Mode

The `--diff` flag shows nothing when combined with `--check` (dry-run mode) because the file isn't saved. The fixer computes the result in-memory but doesn't expose the modified config object.

### 4. AGENTS.md Size

Still at 912 lines — TODO_LIST.md says ≤377 target. Contains detailed reference info that should be in separate docs.

---

## c) NOT STARTED ❌

### From TODO_LIST.md (Open Items)

| Priority | Item                                                          |
| -------- | ------------------------------------------------------------- |
| Critical | Increase CLI integration test coverage (currently 8.2%)       |
| High     | Trim AGENTS.md from 912 to ≤377 lines                         |
| High     | Remove `report_templ.go` from git tracking → **DONE** (stale) |
| Medium   | Add `output.formats: {}` to default config → **DONE** (stale) |
| Medium   | Add reference preset → **DONE** (stale)                       |
| Medium   | Consider swaggo improvements → **DONE** (stale)               |
| Medium   | Add benchmarking → **DONE** (stale)                           |
| Medium   | Document RE2 patterns in README → **DONE** (stale)            |
| Medium   | Add `--check` mode → **DONE** (stale)                         |
| Medium   | Add `--diff` flag → **DONE** (stale)                          |

**TODO_LIST.md is significantly stale** — 7 of 10 open items are actually completed.

### Genuinely Not Started

1. **Typed linter settings structs** — Currently `map[string]any` in `DefaultLinterSettings`
2. **`Config.Clone()` method** — Using JSON marshal/unmarshal in `cmd_configure.go`
3. **`DryRun bool` field on `MigrationResult`** — API clarity
4. **`errors.Join` for multi-finding failures** — Currently returns first error only
5. **`--diff` support in dry-run mode** — Requires `FixConfig` API change
6. **`--format sarif` with `--check` mode** — CI integration gap
7. **`golangci-lint fmt` integration** — Auto-format in check mode
8. **Migrate justfile → flake.nix apps** — Per global AGENTS.md preference
9. **`swaggo` formatter default settings** — Detection done, settings not injected
10. **Typed formatter settings** — Same `map[string]any` issue as linters

---

## d) TOTALLY FUCKED UP 💥

### 1. TODO_LIST.md Is Stale (7/10 Items Already Done)

The TODO list hasn't been updated since 2026-05-23. 7 of 10 "open" items were completed in the June sprint. This is misleading and wastes time checking status.

### 2. `internal/cli` Coverage at 8.2%

The CLI package is the user-facing surface of the entire tool. 8.2% coverage means most command paths, flag handling, error paths, and output formatting are untested. The `--check` and `--diff` features have zero tests. This is the single biggest quality gap.

### 3. `pkg/client` Has Zero Tests

`pkg/client/client.go` exposes `New()`, `SimpleFix()`, `SimpleAnalyze()` — a public API surface with zero test coverage. Anyone using this as a library is flying blind.

### 4. `internal/cli/cmd` Has Zero Tests

`migrate`, `install-hook`, and `completion` commands have no test coverage at all.

### 5. `--diff` + `--check` Is Broken by Design

Two flags that were explicitly built and documented don't work together. Users will try this combination and get nothing. This needs either a fix or a clear warning/error message.

### 6. Lint Has 9 Failures But CI Passes

The project has 9 lint violations but presumably the CI is configured to allow them or runs different settings. This creates a false sense of cleanliness — `just lint` fails locally.

### 7. `LinterMinVersions` Not Validated Against `LinterPriorities`

The `LinterMinVersions` map could reference linters that don't exist in `LinterPriorities`. No compile-time or test-time check catches this. Silent misconfiguration.

---

## e) WHAT WE SHOULD IMPROVE 🎯

### Architecture

1. **Extract `Config.Clone()` method** — Replace JSON marshal/unmarshal hack in `cmd_configure.go` with a proper deep-copy. Makes cloning testable and reusable.

2. **Extract `findingBuilder` helper** — The `pkg/finding/converter.go` has repetitive `buildFinding(NewBuilder(...).With...())` + error check + append patterns. A builder pattern or generic helper would cut 50+ lines and reduce error-proneness.

3. **Typed linter settings** — Replace `map[string]any` with typed structs for `DefaultLinterSettings`. Compile-time safety over runtime map lookups.

4. **`MigrationResult` should have `DryRun bool` field** — Currently ambiguous whether `FixesApplied` means "would fix" or "did fix".

5. **`FixConfig` should return the modified config** — Enables `--diff` in `--check` mode without saving. Changes API but solves the broken-by-design interaction.

### Testing

6. **CLI integration tests** — The 8.2% coverage is the #1 quality gap. Test `--check`, `--diff`, `--dry-run`, error paths, flag combinations.

7. **`pkg/client` tests** — Zero coverage on the public API. Even basic smoke tests would help.

8. **`--check` mode tests** — Zero tests for a CI-facing feature. This is the most critical untested path.

9. **`LinterMinVersions` validation test** — Ensure all version-gated entries exist in `LinterPriorities`.

10. **`reference` preset validation test** — Ensure all preset linters exist in `LinterPriorities`.

### Code Quality

11. **Fix 7 funlen violations** — Extract helpers from long functions. The 46-line `runFixerMode` is the worst offender.

12. **Fix `noinlineerr` in detector.go** — Trivial 2-minute fix.

13. **Fix `varnamelen` in cmd_report.go** — Rename `r` → `report`. 30-second fix.

14. **Use `errors.Join` for multi-finding failures** — Currently loses errors.

### Documentation

15. **Update TODO_LIST.md** — 7/10 items are stale. Needs a full refresh.

16. **Trim AGENTS.md** — Target ≤377 lines. Extract detailed reference to separate docs.

17. **Document `--check` and `--diff` in README** — New features not in user-facing docs.

18. **Document `reference` preset in README** — Users don't know it exists.

### Build

19. **Migrate justfile → flake.nix apps** — Per global AGENTS.md, justfile is deprecated in LarsArtmann projects.

20. **Update flake.nix vendorHash** — Ensure reproducible after any go.mod changes.

---

## f) Top #25 Things to Do Next (Sorted by Impact × Effort)

| #  | Priority | Item                                                             | Est.  | Impact          |
| -- | -------- | ---------------------------------------------------------------- | ----- | --------------- |
| 1  | CRITICAL | Update TODO_LIST.md (7/10 items stale)                           | 30min | Truth           |
| 2  | CRITICAL | Fix `varnamelen` in cmd_report.go (`r` → `report`)               | 2min  | Lint clean      |
| 3  | CRITICAL | Fix `noinlineerr` in detector.go                                 | 2min  | Lint clean      |
| 4  | HIGH     | Add `--check` mode integration tests                             | 1h    | Correctness     |
| 5  | HIGH     | Add `LinterMinVersions` validation test                          | 15min | Correctness     |
| 6  | HIGH     | Validate `reference` preset against `LinterPriorities`           | 15min | Correctness     |
| 7  | HIGH     | Fix funlen: extract `runFixerMode` helpers (46→≤30 lines)        | 30min | Code quality    |
| 8  | HIGH     | Fix funlen: extract `newConfigureCommand` helpers (33→≤30)       | 15min | Code quality    |
| 9  | HIGH     | Fix funlen: extract `CreateDefaultConfig` helpers (33→≤30)       | 15min | Code quality    |
| 10 | HIGH     | Add `Config.Clone()` method, remove JSON marshal hack            | 30min | Architecture    |
| 11 | HIGH     | Add `--diff` flag integration tests                              | 1h    | Correctness     |
| 12 | MEDIUM   | Add tests for ginkgolinter/testifylint default settings          | 30min | Correctness     |
| 13 | MEDIUM   | Fix funlen: `DeprecatedLintersToFindings` (32→≤30)               | 15min | Code quality    |
| 14 | MEDIUM   | Fix funlen: `ChangesToFindings` (34→≤30)                         | 15min | Code quality    |
| 15 | MEDIUM   | Fix funlen: `shouldSkipLinter` (32→≤30)                          | 15min | Code quality    |
| 16 | MEDIUM   | Extract `findingBuilder` helper in converter.go                  | 1h    | Architecture    |
| 17 | MEDIUM   | Add `pkg/client` smoke tests                                     | 1h    | Coverage        |
| 18 | MEDIUM   | Add `--check` + `--diff` interaction handling (warn or fix)      | 1h    | UX              |
| 19 | MEDIUM   | Document `--check`, `--diff`, `reference` preset in README       | 15min | Docs            |
| 20 | MEDIUM   | Use `errors.Join` for multi-finding failures in converter        | 30min | Robustness      |
| 21 | LOW      | Add `DryRun bool` to `MigrationResult`                           | 15min | Type model      |
| 22 | LOW      | Validate `LinterMinVersions` entries exist in `LinterPriorities` | 15min | Correctness     |
| 23 | LOW      | Trim AGENTS.md from 912 to ≤377 lines                            | 2h    | Maintainability |
| 24 | LOW      | Typed linter settings structs (replace `map[string]any`)         | 4h    | Architecture    |
| 25 | LOW      | Migrate justfile → flake.nix apps                                | 2h    | Build           |

---

## g) Top #1 Question I Cannot Figure Out Myself

**What is the intended public API surface for `pkg/client`?**

The `pkg/client/client.go` exposes `Client`, `New()`, `SimpleFix()`, `SimpleAnalyze()` with an `Options` struct. But:

- It has zero tests
- It's not imported by any other package in the codebase
- It's not documented in README or FEATURES.md
- No examples exist in `examples/` for library usage

**Is this package:**

1. **A public library API** meant for programmatic use by external consumers? (Needs tests, docs, semver guarantees)
2. **Internal scaffolding** that should move to `internal/`? (Remove from public surface)
3. **A work-in-progress** that's not ready yet? (Add to FEATURES.md as partial)

This matters because it affects the next sprint priorities: if it's public API, it needs tests urgently (item #17 above). If it's internal, we should move it and stop worrying about coverage.

---

## Package Health Matrix

| Package           | Files | Coverage | Lint Issues | Duplication | Verdict     |
| ----------------- | ----- | -------- | ----------- | ----------- | ----------- |
| `pkg/diff`        | 3     | 96.5%    | 0           | 0           | 🟢 Healthy  |
| `pkg/errors`      | 2     | 95.8%    | 0           | 0           | 🟢 Healthy  |
| `pkg/utils`       | 4     | 94.6%    | 0           | 0           | 🟢 Healthy  |
| `pkg/constants`   | 7     | 80.0%    | 0           | 0           | 🟢 Healthy  |
| `pkg/linter`      | 18    | 81.8%    | 1           | 0           | 🟢 Healthy  |
| `pkg/report`      | 4     | 71.9%    | 0           | 0           | 🟡 OK       |
| `pkg/ui`          | 4     | 67.7%    | 0           | 0           | 🟡 OK       |
| `pkg/migration`   | 7     | 66.8%    | 0           | 0           | 🟡 OK       |
| `pkg/config`      | 9     | 64.5%    | 1           | 0           | 🟡 OK       |
| `pkg/detection`   | 4     | 62.8%    | 1           | 0           | 🟡 OK       |
| `pkg/gogenfilter` | 2     | 59.8%    | 0           | 0           | 🟡 OK       |
| `pkg/types`       | 7     | 58.8%    | 0           | 0           | 🟡 OK       |
| `pkg/finding`     | 8     | 50.0%    | 2           | 0           | 🟠 Needs    |
| `pkg/version`     | 2     | 51.4%    | 0           | 0           | 🟠 Needs    |
| `internal/cli`    | 12    | 8.2%     | 4           | 0           | 🔴 Critical |
| `pkg/client`      | 1     | 0.0%     | 0           | 0           | 🔴 No tests |

## Test Suite Summary

```
Ginkgo ran 15 suites in 16.5s — ALL PASS
Composite coverage: 61.1% of statements
```

| Suite           | Specs | Coverage |
| --------------- | ----- | -------- |
| Linter          | ?     | 81.8%    |
| Config          | ?     | 64.5%    |
| Constants       | ?     | 80.0%    |
| Detection       | ?     | 62.8%    |
| Diff            | ?     | 96.5%    |
| Errors          | ?     | 95.8%    |
| Finding         | ?     | 50.0%    |
| Gogenfilter     | ?     | 59.8%    |
| Migration       | ?     | 66.8%    |
| Report          | ?     | 71.9%    |
| Types           | ?     | 58.8%    |
| UI              | 16    | 67.7%    |
| Utils           | 16    | 94.6%    |
| Version         | 6     | 51.4%    |
| CLI Integration | ?     | 8.2%     |

## Build Status

- **Build:** ✅ Passes (vv0.2.0-12-gbfd0a12)
- **Tests:** ✅ All 15 suites pass
- **Lint:** ❌ 9 issues (7 funlen, 1 noinlineerr, 1 varnamelen)
- **Vet:** ✅ Clean
- **Duplication:** ✅ Zero at threshold 50
- **Dependencies:** 12 direct, 37 total

## What Changed Since Last Report (2026-06-03)

Only 1 commit since last status report:

- `bfd0a12` — docs: reformat RE2 exclusion patterns section in README and refresh status tables

No functional changes. Codebase is stable and awaiting next sprint.

---

_Generated by Crush — 2026-06-05 02:20_
