# Comprehensive Status Report — 2026-06-28

**Generated:** 2026-06-28 22:45
**Branch:** master (up to date with origin/master)
**Build:** ✅ Clean — `go build ./...` passes
**Tests:** ✅ All 15 packages green (`go test -race ./pkg/... ./internal/...`)

---

## A. FULLY DONE ✅

### 1. go-finding v1.0.0 API Breakage — FIXED

The dependency bump to go-finding v1.0.0 (commit `73c8650`) introduced branded types
that broke the build. All 9 call sites fixed:

| File                                    | Fix                                                                    |
| --------------------------------------- | ---------------------------------------------------------------------- |
| `pkg/finding/finding_builder.go:33-34`  | `string` → `finding.RuleName(...)`, `finding.ToolName(...)`            |
| `pkg/finding/diff_converter.go:31-32`   | Same branded-type wrapping                                             |
| `pkg/finding/golangci_lint.go:60-62`    | Same + `"golangci-lint"` → `finding.ToolName(...)`                     |
| `internal/cli/cmd_validate.go:217-219`  | Same branded-type wrapping                                             |
| `pkg/finding/helpers.go:40,52`          | `report.Findings` → `report.FindingsSnapshot()` (field now unexported) |
| `pkg/finding/converter_test.go:331-393` | `report.Findings` → `FindingsSnapshot()` (4 sites)                     |
| `pkg/finding/converter_test.go:415`     | `f.Rule != expectedRule` → `finding.RuleName(...)`                     |
| `pkg/finding/golangci_lint_test.go:172` | `f.Rule != linter` → `finding.RuleName(...)`                           |

**Verified:** Build clean, all tests pass with race detector.

### 2. AGENTS.md Rewrite — LEAN & CORRECT

Rewrote from ~290 bloated lines to ~75 lines of high-signal context.

**Critical correction:** The old AGENTS.md told agents to use `just build`, `just test`, etc.
**There is no justfile.** The build system is Nix flake + plain Go tooling. All `just`
references replaced with real commands (`nix build`, `go test -race`, `golangci-lint run`, etc.).

Also corrected: go-finding/gogenfilter are now published deps (no local `replace` in go.mod,
contrary to old docs).

### 3. Stale `just` References Purged from Living Docs

- `README.md` — Testing + Contributing sections
- `docs/references/working-with-codebase.md` — 10 command blocks fixed
- `docs/references/testing-style-and-patterns.md` — test/coverage commands fixed
- `docs/references/code-organization.md` — justfile → flake.nix

### 4. Error Family Integration (pre-existing, now committed)

`go-error-family v0.5.1` integrated for semantic exit codes (BSD sysexits.h):

- `pkg/errors/classification.go` — registers all sentinel errors to families (Rejection, Conflict, Corruption, Infrastructure)
- `internal/cli/commands.go` — `Main()` now classifies errors and exits with proper codes
- `pkg/errors/classification_test.go` — full BDD test coverage

---

## B. PARTIALLY DONE ⚠️

### 1. Error Family Integration — Wired but Not Fully Propagated

The classification system is registered and the CLI `Main()` uses it, BUT:

- Individual command handlers still return generic errors without wrapping in classified types
- No documentation yet on exit code meanings for CI/CD consumers
- `flake.nix` `vendorHash` will need updating after this commit (Nix build not yet verified)

### 2. templ Code Generation

- Works in Nix build (`preBuild` runs `templ generate`)
- No local convenience command — developers must remember to run `templ generate` manually
- Should be documented or wrapped

---

## C. NOT STARTED ❌

1. **`flake.nix` vendorHash update** — after go-error-family addition, Nix build will fail on hash mismatch
2. **Lint pass** — `golangci-lint run` not yet verified clean after all changes
3. **Nix build verification** — `nix build` not run (will fail on vendorHash until updated)
4. **CHANGELOG.md entry** — no entry for go-finding fix or error-family integration
5. **FEATURES.md update** — error classification feature not listed
6. **`nix flake check`** — not run

---

## D. TOTALLY FUCKED UP 💥

### 1. Documentation Drift Was Severe

The AGENTS.md was actively misleading every AI session:

- Instructed agents to use `just` commands that **do not exist** (no justfile in repo)
- Claimed `go-finding` uses local `replace` directive — **it doesn't anymore** (published v1.0.0)
- 290 lines of detail duplicated from `docs/references/` — violating the "AGENTS.md must be SHORT" principle
- This caused agents to waste time on non-existent workflows and misunderstand the dependency model

### 2. Build Was Broken on master

Commit `73c8650` bumped go-finding to v1.0.0 but did NOT update any call sites.
**The build was broken** — `go build ./...` failed with 5 compiler errors.
This means CI was either not running or not gating on build success.
The breakage affected: finding builder, diff converter, golangci-lint parser, validate command, all finding tests.

---

## E. WHAT WE SHOULD IMPROVE 🔧

1. **CI must gate on `go build ./...`** — a broken build landed on master and stayed. The CI workflow runs `go build -v ./...` but it somehow passed (possibly cached, or the breakage was introduced after the last green CI run on the exact commit). Need to verify CI actually blocks merges on build failure.

2. **AGENTS.md discipline** — keep it under 100 lines. Push detail to `docs/references/`. The global AGENTS.md rule exists for a reason and was being violated.

3. **Dependency bump hygiene** — when bumping a dependency with breaking API changes, the same PR/commit MUST include the call-site fixes. Never land a dep bump that breaks the build.

4. **Single source of truth for commands** — the `just` lie propagated to 4+ doc files. Consider a single `docs/commands.md` that everything else links to, or generate command docs from a script.

5. **templ workflow** — add a `templ generate` step to README dev instructions or create a Nix devShell hook that watches `.templ` files.

6. **Exit code documentation** — now that error-family gives semantic exit codes, document them in README for CI/CD users (exit 1 = user fault, exit 65 = corruption, exit 69 = infrastructure).

---

## F. TOP 25 THINGS TO DO NEXT 🎯

Sorted by Impact × Customer-Value ÷ Effort.

| #  | Task                                                                              | Impact   | Effort | Category           |
| -- | --------------------------------------------------------------------------------- | -------- | ------ | ------------------ |
| 1  | Update `flake.nix` vendorHash for go-error-family                                 | Critical | 5min   | Blocks Nix build   |
| 2  | Run `golangci-lint run` and fix any new lint issues                               | High     | 15min  | Quality gate       |
| 3  | Run `nix build` + `nix flake check` to verify full pipeline                       | High     | 10min  | CI readiness       |
| 4  | Add CHANGELOG.md entry for go-finding fix + error-family                          | Medium   | 10min  | Release prep       |
| 5  | Update FEATURES.md with error classification feature                              | Medium   | 10min  | Docs               |
| 6  | Verify CI actually blocks on `go build` failure (audit ci.yml)                    | High     | 15min  | Process            |
| 7  | Add exit code documentation to README (BSD sysexits table)                        | Medium   | 15min  | UX                 |
| 8  | Add `.editorconfig` or treefmt check for templ files in CI                        | Low      | 10min  | Quality            |
| 9  | Propagate error-family classification to all command handlers                     | High     | 30min  | Feature completion |
| 10 | Add integration test: verify exit codes for common failure modes                  | Medium   | 20min  | Test coverage      |
| 11 | Clean up `docs/status/` archive — 100+ old status reports cluttering              | Low      | 15min  | Hygiene            |
| 12 | Audit all remaining `just` references in archived docs (or add disclaimer)        | Low      | 5min   | Hygiene            |
| 13 | Add `templ generate` to README quickstart or devShell hook                        | Low      | 10min  | DX                 |
| 14 | Review `pkg/errors/classification.go` — are all sentinel errors covered?          | Medium   | 15min  | Completeness       |
| 15 | Add `--dry-run` exit code test for `configure --check` (exit 1 on changes needed) | Medium   | 15min  | CI contract        |
| 16 | Verify SARIF output still works after finding API changes                         | High     | 15min  | Feature            |
| 17 | Check `examples/` configs are still valid v2 format                               | Low      | 10min  | Docs               |
| 18 | Run `go mod tidy` to ensure no unused deps linger                                 | Low      | 5min   | Hygiene            |
| 19 | Add pre-push git hook option (currently only pre-commit)                          | Low      | 15min  | DX                 |
| 20 | Review `flake.lock` — is go-error-family properly pinned?                         | Medium   | 5min   | Reproducibility    |
| 21 | Consider adding `nix develop` automatic `templ generate` on enter                 | Low      | 10min  | DX                 |
| 22 | Audit `internal/di/` references in old docs (doesn't exist)                       | Low      | 5min   | Docs               |
| 23 | Add version-gated deprecation tests for gomodguard_v2                             | Medium   | 20min  | Test coverage      |
| 24 | Review duplicate ConfigError/AnalysisError field patterns                         | Low      | 15min  | Tech debt          |
| 25 | Plan next minor version release (v0.3.0?) with error-family feature               | Medium   | 20min  | Release            |

---

## G. TOP QUESTION I CANNOT FIGURE OUT MYSELF 🤔

**#1: How did commit `73c8650` (go-finding v1.0.0 bump) land on master with a broken build?**

The CI workflow `.github/workflows/ci.yml` runs `go build -v ./...` in the `test-and-build` job.
If that passed, either:

- (a) CI wasn't triggered for that commit
- (b) CI ran but the job is non-blocking (the `summary` job uses `if: always()` which could mask failures)
- (c) The build was broken after the CI run via a subsequent commit
- (d) CI caching masked the failure

I cannot determine which without access to GitHub Actions run history. This is important because if CI is non-blocking, **every future dep bump can break master silently**.

**Action needed:** Check the GitHub Actions tab for the `73c8650` commit run status, and audit whether the `summary` job's `if: always()` is masking real failures.
