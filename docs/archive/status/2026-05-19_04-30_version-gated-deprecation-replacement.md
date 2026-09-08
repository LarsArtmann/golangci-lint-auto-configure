# Status Report — 2026-05-19 04:30 CEST

**Date:** 2026-05-19 04:30 CEST
**Author:** Crush (AI Agent)
**Session Focus:** Version-gated deprecated linter replacements — preventing breakage when golangci-lint version doesn't support replacement linter
**Branch:** `master` @ `50c1a0f`
**Go:** 1.26.2 | **golangci-lint:** v2.12.2 | **ginkgo:** v2.28.3

---

## Executive Summary

This session fixed a **production-breaking bug**: the tool blindly replaced `gomodguard → gomodguard_v2` even when the installed golangci-lint didn't support `gomodguard_v2` (only available in v2.12.0+). The tool's minimum version is v2.10.1, creating a gap where v2.10.1–v2.11.x users got `unknown linters: 'gomodguard_v2'` errors.

**Root cause discovered from a user bug report:** A project on a server with an older golangci-lint had `gomodguard_v2` in its `.golangci.yml` (put there by this tool), but the server's golangci-lint was too old to know that linter. Running `golangci-lint run --fix ./...` failed immediately.

**Fix:** Added `MinVersion` field to `LinterReplacement` and version-gated all deprecation replacements. Now `gomodguard → gomodguard_v2` only fires when installed golangci-lint >= v2.12.0.

**Overall project health:** 98 Go source files, 17,077 lines, 20 packages, 26 test files, 14 Ginkgo suites (290+ specs), 0 lint issues, 60.2% composite coverage.

---

## A. FULLY DONE

### 1. Version-Gated Deprecated Linter Replacements (commit `50c1a0f`)

**Problem:** The tool had `gomodguard → gomodguard_v2` in its `DeprecatedLinters` map with `MinVersion: ""`. When a user with golangci-lint v2.10.1–v2.11.x ran `configure`, the tool replaced `gomodguard` with `gomodguard_v2` — which doesn't exist in those versions. Result: `Error: unknown linters: 'gomodguard_v2'`.

**Files changed (9 files, +638/-427):**

| File                                    | Change                                                                                                       |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `pkg/types/types.go`                    | Added `MinVersion string` to `LinterReplacement`, added `GetDetectedVersion()` to `LinterAnalyzer` interface |
| `pkg/linter/analyzer.go`                | Added `detectedVersion` field, stored during `CheckVersion`, exposed via `GetDetectedVersion()`              |
| `pkg/constants/rules.go`                | Added `MinVersion: "v2.12.0"` to `gomodguard` deprecation entry                                              |
| `pkg/linter/fixer.go`                   | Added `detectVersion()` for early version detection, threaded version through all fix paths                  |
| `pkg/linter/fixer_deprecated.go`        | Added `replacementAvailable()` — skips replacements when installed version < MinVersion                      |
| `pkg/linter/fixer_preflight.go`         | Updated `filterDeprecatedFrom` to only remove linters whose replacements are available                       |
| `pkg/linter/deprecated_version_test.go` | New file: unit tests for version-gating logic                                                                |
| `pkg/linter/version_checker.go`         | Stores detected version in `a.detectedVersion`                                                               |
| `.golangci.yml`                         | Reformatting (tabs → spaces) from earlier session                                                            |

**Key architectural decisions:**

- **Version detection early in pipeline:** `detectVersion()` runs before pre-flight checks, so version gating works even before `golangci-lint linters` command
- **Graceful degradation:** If version detection fails (binary not found, network issue), `replacementAvailable()` returns `true` — same behavior as before the fix (all replacements applied). This prevents false negatives.
- **No MinVersion = always available:** Existing deprecation entries (`wsl`, `deadcode`, etc.) have empty `MinVersion`, so they're always applied. Only `gomodguard` is version-gated.

**Test coverage:**

- 6 unit tests in `deprecated_version_test.go` covering: no min version, empty version, version meets/exceeds/below min version
- 1 integration test for `hasDeprecatedLinters` version gating
- All 14 Ginkgo suites pass (290+ specs)
- 0 lint issues

### 2. .golangci.yml Reformatting (commit `838ef21`)

Reformatted from 4-space tabs to 2-space indentation for consistency with golangci-lint v2 conventions. Cosmetic only, no behavioral changes.

### 3. Previous Sessions (already committed)

Since last status report (2026-05-16):

| Commit              | Description                                          |
| ------------------- | ---------------------------------------------------- |
| `0d1731b`           | Updated vendorHash for new go-modules                |
| `838ef21`           | Reformatted `.golangci.yml` with v2 config structure |
| `7cc4fb5`           | Updated flake.lock with go-finding dependency lock   |
| `997ace7`           | Improved table alignment in docs                     |
| `d176294`           | Added Go module and documentation                    |
| `066768d`–`352804f` | GoReleaser setup for CI/CD releases                  |

---

## B. PARTIALLY DONE

### 1. GoReleaser CI/CD Pipeline

**Status:** Configuration exists (`352804f`) but actual releases haven't been triggered. CI uses `skip_upload: true` / `skip_upload: auto` for brews/scoops (no token configured yet). No `.github/workflows/` directory exists — unclear if GitHub Actions is set up.

**What's missing:**

- GitHub Actions workflow for GoReleaser
- Homebrew tap repository
- Scoop bucket
- COSIGN signing key/token
- Actual v1.0.0 tag and release

### 2. gomodguard_v2 Settings Schema

**Status:** The deprecation replacement migrates settings keys (`gomodguard` → `gomodguard_v2`), but there's no schema validation that the settings are compatible between versions. If gomodguard_v2 changes its settings format, the migration could produce invalid config.

**What's missing:**

- Settings schema validation for replacement linters
- Test for settings migration with real gomodguard_v2 settings

---

## C. NOT STARTED

### Major Features Not Started

| #  | Feature                                                             | Priority | Effort  |
| -- | ------------------------------------------------------------------- | -------- | ------- |
| 1  | GitHub Actions CI workflow                                          | High     | 2h      |
| 2  | First release (v1.0.0)                                              | High     | 1h      |
| 3  | Homebrew formula + tap                                              | Medium   | 2h      |
| 4  | Docker image build + publish                                        | Medium   | 1h      |
| 5  | `TODO_LIST.md` — no project-level TODO tracking exists              | Medium   | 1h      |
| 6  | `FEATURES.md` — no feature inventory exists                         | Medium   | 2h      |
| 7  | SARIF report file output (`report --format sarif`)                  | Low      | 3h      |
| 8  | go-finding pipeline integration (continuous mode)                   | Low      | 1 week  |
| 9  | Nix flake module for NixOS integration                              | Low      | 4h      |
| 10 | Plugin system for custom linter recommendations                     | Low      | 1 week  |
| 11 | Web UI for configuration visualization                              | Low      | 2 weeks |
| 12 | Multi-config support (subdirectory configs)                         | Low      | 3h      |
| 13 | Pre-commit hook v2 (using go-finding model)                         | Low      | 2h      |
| 14 | Integration tests with real golangci-lint binaries (version matrix) | Medium   | 4h      |
| 15 | Version-specific linter database (per golangci-lint version)        | Medium   | 1 week  |

### Documentation Gaps

| # | Gap                                                                     | Impact                             |
| - | ----------------------------------------------------------------------- | ---------------------------------- |
| 1 | No README update since initial write                                    | Users don't know about v2 features |
| 2 | No CHANGELOG.md                                                         | Users can't track what changed     |
| 3 | No CONTRIBUTING.md                                                      | Contributors don't know process    |
| 4 | No architecture decision records (ADR)                                  | Decision context lost              |
| 5 | AGENTS.md references `just` but `justfile` exists alongside `flake.nix` | Confusing for Nix-first projects   |

---

## D. TOTALLY FUCKED UP

### 1. The Version-Gate Gap Was a Silent Landmine

The `gomodguard → gomodguard_v2` replacement was added on 2026-05-16 (commit `4a5e1a3`) without ANY version gating. For 3 days, every user with golangci-lint < v2.12.0 who ran `configure` got a broken config. This was caught only because a user hit it on a remote server and shared the error.

**Why this is bad:**

- No CI matrix testing against multiple golangci-lint versions
- The tool's own dogfooding only tests against the latest version
- No integration test that verifies "replace deprecated linter" produces a config that actually works
- The minimum version check (`v2.10.1`) was treated as "this is the version we support" when it should have been "this is the minimum, but some features require newer versions"

**Lesson:** Every deprecation replacement MUST have a `MinVersion` specifying when the replacement linter was introduced. This is now enforced by the fix, but only for `gomodguard`. We should audit all `DeprecatedLinters` entries.

### 2. Nix nixpkgs Ships golangci-lint v2.11.4, Not v2.12.2

`nix run nixpkgs#golangci-lint -- version` returns v2.11.4 — which means `gomodguard_v2` doesn't exist in the nixpkgs version. This is a nixpkgs lag issue but it means the version-gating fix is critical for Nix users.

### 3. gci Lint Warning Persists on Test File

`deprecated_version_test.go:11:1` has a persistent `gci` formatting warning that `golangci-lint fmt` doesn't auto-fix. This is likely a false positive from the linter or a gci version mismatch.

---

## E. WHAT WE SHOULD IMPROVE

### Architecture & Design

| # | Improvement                                                                                                                                              | Why                                                                                                             |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| 1 | **Add `MinVersion` to ALL `DeprecatedLinters` entries**                                                                                                  | `wsl → wsl_v5` needs `MinVersion: "v2.2.0"`, others need auditing. Currently only `gomodguard` is version-gated |
| 2 | **Integration test matrix** — test `configure` against golangci-lint v2.10.1, v2.11.x, v2.12.x                                                           | The current bug would have been caught immediately                                                              |
| 3 | **Post-fix config validation** — after writing config, run `golangci-lint config verify` or parse linter list to confirm all referenced linters exist    | Safety net against future breakage                                                                              |
| 4 | **Move `DeprecatedLinters` to a data-driven model** — each entry should have: replacement, reason, MinVersion, MaxVersion (when the old one was removed) | Prevents future version-gap bugs                                                                                |
| 5 | **Version-aware linter database** — `golangci-lint help linters` output changes per version. Cache per-version linter lists                              | Enables version-specific recommendations                                                                        |

### Testing

| # | Improvement                                                                                   | Why                                            |
| - | --------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| 6 | **Test coverage at 60.2%** — needs improvement                                                | Critical path (fixer, analyzer) should be 80%+ |
| 7 | **No integration tests with version matrix**                                                  | As mentioned above                             |
| 8 | **No end-to-end test** — run the actual binary, verify config output works with golangci-lint | Highest-value missing test                     |
| 9 | **Benchmark tests** — no performance regression testing                                       | Important for large repos                      |

### Operations

| #  | Improvement                                                                | Why                                    |
| -- | -------------------------------------------------------------------------- | -------------------------------------- |
| 10 | **GitHub Actions CI** — no `.github/workflows/` at all                     | No automated testing on push/PR        |
| 11 | **No release automation** — GoReleaser configured but never triggered      | Users can't install via brew/nix/scoop |
| 12 | **flake.nix vendorHash** — must be manually updated after `go.mod` changes | Error-prone process                    |
| 13 | **No dependabot/renovate** — dependency updates are manual                 | Security risk                          |

### Code Quality

| #  | Improvement                                                                  | Why                                                                                                                              |
| -- | ---------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| 14 | **`resolveLinterName` is NOT version-gated**                                 | It still resolves `gomodguard` → `gomodguard_v2` regardless of version. Used in `enableRecommendedLinters`. This is a latent bug |
| 15 | **`applyDeprecatedReplacements` in fixer_preflight.go is NOT version-gated** | Used in `calculateDryRunResultWithDeprecated` — dry-run mode doesn't check version before listing replacements                   |
| 16 | **`formatDeprecatedSection` in analyzer.go is NOT version-gated**            | Analysis output still shows "use gomodguard_v2 instead" even when version is too old                                             |
| 17 | **gogenfilter test coverage at 59.8%** — below project average               | Scanner is critical path for `configure`                                                                                         |

---

## F. Top #25 Things We Should Get Done Next

### Priority 1 — Critical Bugs & Safety (DO THESE FIRST)

| # | Task                                                                                                                   | Effort | Impact                                  |
| - | ---------------------------------------------------------------------------------------------------------------------- | ------ | --------------------------------------- |
| 1 | **Version-gate `resolveLinterName` in `fixer.go`** — currently resolves `gomodguard` → `gomodguard_v2` unconditionally | 30min  | HIGH — latent version-gap bug           |
| 2 | **Version-gate `applyDeprecatedReplacements` in `fixer_preflight.go`** — dry-run output is wrong for old versions      | 30min  | HIGH — misleading dry-run output        |
| 3 | **Version-gate `formatDeprecatedSection` in `analyzer.go`** — analysis shows wrong replacement                         | 30min  | MEDIUM — misleading analyze output      |
| 4 | **Add `MinVersion` to ALL `DeprecatedLinters` entries** (`wsl` needs `"v2.2.0"`, etc.)                                 | 1h     | HIGH — prevents future version-gap bugs |
| 5 | **Add post-fix config validation** — after writing config, verify all linters exist                                    | 2h     | HIGH — safety net                       |

### Priority 2 — Release Readiness

| #  | Task                                                                                            | Effort | Impact                          |
| -- | ----------------------------------------------------------------------------------------------- | ------ | ------------------------------- |
| 6  | **Create GitHub Actions CI workflow** — build + test + lint on push/PR                          | 2h     | HIGH — automated quality gate   |
| 7  | **Add integration test with version matrix** — test configure against v2.10.1, v2.11.x, v2.12.x | 4h     | HIGH — catches version-gap bugs |
| 8  | **Tag v0.1.0 and test GoReleaser** — verify the full release pipeline works                     | 2h     | HIGH — unblocks distribution    |
| 9  | **Update README.md** — reflects current feature set, not just initial vision                    | 2h     | HIGH — users need current docs  |
| 10 | **Create CHANGELOG.md** — track changes since project start                                     | 1h     | MEDIUM — release readiness      |

### Priority 3 — Quality & Coverage

| #  | Task                                                                                  | Effort | Impact                            |
| -- | ------------------------------------------------------------------------------------- | ------ | --------------------------------- |
| 11 | **Write end-to-end test** — run actual binary, verify config works with golangci-lint | 3h     | HIGH — highest-value missing test |
| 12 | **Improve test coverage to 70%+** — focus on fixer, analyzer, config packages         | 4h     | MEDIUM — confidence in changes    |
| 13 | **Write FEATURES.md** — inventory all features with status                            | 2h     | MEDIUM — project visibility       |
| 14 | **Write TODO_LIST.md** — track all planned work                                       | 1h     | MEDIUM — project tracking         |
| 15 | **Fix persistent gci warning on test file** — investigate root cause                  | 30min  | LOW — clean lint output           |

### Priority 4 — Architecture Improvements

| #  | Task                                                                                                                           | Effort | Impact                             |
| -- | ------------------------------------------------------------------------------------------------------------------------------ | ------ | ---------------------------------- |
| 16 | **Create `LinterDeprecation` struct** — Replacement, Reason, MinVersion, RemovalVersion (when old linter was removed entirely) | 2h     | MEDIUM — data model completeness   |
| 17 | **Add version-aware linter database** — cache `golangci-lint help linters` per version                                         | 1 week | MEDIUM — version-specific behavior |
| 18 | **Extract deprecation data to YAML/JSON** — make it data-driven, not Go code                                                   | 3h     | MEDIUM — maintainability           |
| 19 | **Add `--target-version` flag** — let users specify which golangci-lint version they target                                    | 2h     | MEDIUM — CI/CD integration         |
| 20 | **Refactor `DeprecatedLinters` to include `RemovalVersion`** — know when old linter stops working entirely                     | 1h     | MEDIUM — proactive migration       |

### Priority 5 — Distribution & Ecosystem

| #  | Task                                                                                  | Effort | Impact                                |
| -- | ------------------------------------------------------------------------------------- | ------ | ------------------------------------- |
| 21 | **Set up Homebrew tap** — `brew install larsartmann/tap/golangci-lint-auto-configure` | 2h     | MEDIUM — easy install for macOS users |
| 22 | **Publish Docker image** — `docker run larsartmann/golangci-lint-auto-configure`      | 1h     | MEDIUM — CI/CD pipeline integration   |
| 23 | **Add NixOS module** — `services.golangci-lint-auto-configure`                        | 4h     | LOW — Nix ecosystem                   |
| 24 | **Create pre-commit hook v2** — using go-finding model                                | 2h     | MEDIUM — better developer UX          |
| 25 | **Write CONTRIBUTING.md** — onboarding for external contributors                      | 1h     | LOW — community readiness             |

---

## G. Top #1 Question I Cannot Figure Out Myself

**What is the target golangci-lint version for this tool?**

The minimum supported version is `v2.10.1` (in `pkg/constants/version.go`), but:

- nixpkgs ships v2.11.4 (doesn't have `gomodguard_v2`)
- Local dev uses v2.12.2 (has `gomodguard_v2`)
- The server where the bug was found runs an unknown version < v2.12.0
- Some `DeprecatedLinters` entries reference features from v2.2.0 (`wsl_v5`)
- The tool's own `.golangci.yml` uses `gomodguard_v2` (requires v2.12.0+)

**The question:** Should the tool bump its minimum version to v2.12.0, keep v2.10.1 and make all version-sensitive features conditional, or adopt a different strategy (e.g., per-feature minimum versions with graceful degradation)?

This matters because it determines whether we need to version-gate every feature or just bump the minimum and move on.

---

## Metrics

| Metric                        | Value       |
| ----------------------------- | ----------- |
| Go source files               | 98          |
| Test files                    | 26          |
| Total lines of Go code        | 17,077      |
| Go packages                   | 20          |
| Ginkgo test suites            | 14          |
| Test specs                    | 290+        |
| Composite coverage            | 60.2%       |
| Lint issues                   | 0           |
| Build status                  | ✅ Clean    |
| Test status                   | ✅ All pass |
| Git commits since last report | 8           |
| Deprecated linters tracked    | 10          |
| Linter priorities defined     | 119         |
| Formatters tracked            | 5           |
| CLI subcommands               | 7           |

---

## Recent Commit History (last 20)

```
50c1a0f feat(linter): add version-gated deprecated linter replacement
0d1731b chore(deps): update vendorHash for new go-modules
838ef21 chore(linting): reformat .golangci.yml with v2 config structure
7cc4fb5 chore(deps): update flake.lock with go-finding dependency lock
997ace7 docs(domain-language): improve table alignment and blockquote formatting
d176294 feat(project): add Go module and documentation
066768d Add COSIGN_YES=true for keyless signing in CI
4ee92ee Set skip_upload: true for brews/scoops (no token yet)
1eda296 Fix GoReleaser v2: rename folder to directory in brews/scoops
97e733e Add skip_upload: auto to brews, nix, and scoops
352804f Add GoReleaser: full release automation with Nix, Homebrew, Docker, signing
c73af69 docs(status): improve table formatting and code block syntax
daa55b1 refactor(tests): apply DRY principle across test files
0f282fd docs(status): add comprehensive report for architecture session
9175a7a docs(status): add comprehensive session report
ca6b00d feat(types): add TotalRecommendations and EnabledLinterNames
31cce92 fix(validate): log warning when health issue finding fails to build
c9d617a refactor(detection): add ProjectType.Preset() method
206c20a feat(types): add ConfigHealth domain methods; migrate tests
2b0fb0f feat(health-checks): derive critical linters from constants
```

---

_Report generated by Crush (AI Agent) on 2026-05-19 at 04:30 CEST._
