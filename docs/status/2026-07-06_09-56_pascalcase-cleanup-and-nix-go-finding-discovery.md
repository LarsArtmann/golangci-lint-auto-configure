# Status: PascalCase Migration Cleanup & go-finding Nix Discovery

> **🔄 RETROACTIVE UPDATE — 2026-07-16**
>
> The go-finding version mismatch (the session's main discovery) is fully resolved:
>
> | Item                                          | Status        | Details                                                                          |
> | --------------------------------------------- | ------------- | -------------------------------------------------------------------------------- |
> | go-finding upgrade                            | ✅ Done       | Now at v1.2.0 (this session reverted to v1.0.0; later session properly upgraded) |
> | nix fmt passes                                | ✅ Done       | treefmt formatting verified                                                      |
> | nix flake check passes                        | ✅ Done       | All 3 checks (format, build, race) pass                                          |
> | PascalCase migration fully complete           | ✅ Done       | All report types tag-free, enforced by tagliatelle                               |
> | Git add before nix build                      | ✅ Documented | Lessons learned from this session are captured in AGENTS.md                      |
> | examples/*.golangci.yml tagliatelle alignment | ❌ Not done   | Still `json: snake` — deliberate (examples for downstream projects)              |
> | flake.lock drift check in CI                  | ❌ Not done   |                                                                                  |
>
> This session's key insight — "when Nix bumps a dependency, adopt the new version fully, don't fight it" — became the project's go-finding upgrade strategy. Current open items: `TODO_LIST.md`.

**Date:** 2026-07-06 09:56
**Previous reports:** `docs/status/2026-07-06_05-54_json-pascalcase-tag-migration.md`, `docs/status/2026-07-06_06-21_pascalcase-migration-brutal-self-review.md`
**Session scope:** Fix remaining issues from PascalCase tag migration: revert slog keys, split types.go, add integration tests, run nix fmt + nix flake check, document findings.

---

## a) FULLY DONE

| #  | Task                                                                                           | Verification                                                                                                        |
| -- | ---------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| 1  | Reverted slog keys in `examples/api-usage/main.go` back to snake_case                          | Was scope creep — slog keys ≠ serialization tags                                                                    |
| 2  | Audited ALL structs with `json:"` tags across entire codebase                                  | `rg 'json:"' --type go` confirms 3 type families are correct                                                        |
| 3  | Split `pkg/types/types.go` → report types (stays) + `config_types.go` (new)                    | Config types with kebab tags extracted for precise tagliatelle enforcement                                          |
| 4  | Moved tagliatelle + tagalign exclusions from `types.go` → `config_types.go` in `.golangci.yml` | Report types now PascalCase-enforced without file-level exclusion                                                   |
| 5  | Updated AGENTS.md gotcha #12 table with file locations                                         | Reflects the split: `types.go` = report types, `config_types.go` = config types                                     |
| 6  | Added PascalCase JSON key integration tests for `analyze` and `report` commands                | `internal/cli/integration_test.go` — parse JSON output, assert `ConfigPath`, `EnabledLinters`, `CriticalCount` etc. |
| 7  | Fixed stale `just build` reference in integration test skip message                            | Changed to `go build -o bin/...`                                                                                    |
| 8  | Added CHANGELOG entry for breaking JSON key change                                             | BREAKING note in `[Unreleased] > Changed`                                                                           |
| 9  | Ran `nix fmt` (treefmt)                                                                        | 0 changes needed — all files properly formatted                                                                     |
| 10 | `go build ./...`                                                                               | **PASS**                                                                                                            |
| 11 | `golangci-lint run`                                                                            | **PASS** (0 issues)                                                                                                 |
| 12 | `go test -race ./pkg/... ./internal/...` (16 packages)                                         | **PASS** (all green)                                                                                                |
| 13 | Nix format check (`nix build .#checks.x86_64-linux.format`)                                    | **PASS**                                                                                                            |

---

## b) PARTIALLY DONE

| # | What                            | Status                                                              | Gap                                                                                            |
| - | ------------------------------- | ------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| 1 | `nix flake check`               | Format check passes; build/test/race checks fail                    | Pre-existing go-finding version mismatch (see d) #1-2). Documented in AGENTS.md but not fixed. |
| 2 | go-finding FilePath API upgrade | Attempted, reverted, then committed by a later session as `7d27530` | My session couldn't resolve the API + runtime output format change cleanly.                    |

---

## c) NOT STARTED

| # | Task                                                                        | Why                                                                      |
| - | --------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| 1 | CBOR support (`fxamacker/cbor`) with PascalCase tags                        | No CBOR library in project; policy-only decision                         |
| 2 | Property-based testing for JSON round-trip (marshal → unmarshal → equality) | Deferred — current unit tests cover key shape, not full round-trip       |
| 3 | JSON Schema export from report types                                        | Feature idea, not scoped                                                 |
| 4 | `--legacy-json-keys` flag for backward compatibility                        | Would require dual serialization paths                                   |
| 5 | Audit `examples/*.golangci.yml` tagliatelle config alignment                | These are example configs for other projects, not this tool's own config |

---

## d) TOTALLY FUCKED UP

| # | What                                                            | Impact                                                                                                                                                                                                                     | Root Cause                                                                                                                                                                                                                                                                                   | Resolution                                                                                                                                                                                                                                             |
| - | --------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | **Didn't stage `config_types.go` before `nix flake check`**     | Nix build failed with `undefined: Config` because `lib.fileset.unions` only sees git-tracked files                                                                                                                         | Nix flakes filter source by git index — untracked files are invisible. I should have `git add` before any Nix build.                                                                                                                                                                         | Fixed by staging. **Lesson: always `git add` new files before running `nix build` or `nix flake check`.**                                                                                                                                              |
| 2 | **Churned on go-finding FilePath API for 6+ tool calls**        | Wasted time applying casts, discovering they break v1.0.0, reverting, then the Nix FOD bumped go.mod to v1.1.0, breaking tests again.                                                                                      | I didn't understand the version mismatch root cause until late. The Nix FOD runs `go mod tidy` which bumps go.mod. The SSH-fetched go-finding HEAD has a newer API than the published v1.0.0 tag. v1.1.0 has the branded `FilePath` type AND changes SARIF/finding output format at runtime. | Reverted to v1.0.0 for local correctness. Later session (`7d27530`) properly upgraded to v1.1.0 with all FilePath casts + output format fixes. **Lesson: when Nix bumps a dependency, the right fix is to adopt the new version fully, not fight it.** |
| 3 | **`nix flake check` updated `flake.lock` as a side effect**     | The flake.lock got bumped to a newer go-finding SSH revision. I almost committed this unintentionally.                                                                                                                     | `nix flake check` evaluates and fetches inputs, updating the lock file as a side effect. I should have checked `git diff` before committing.                                                                                                                                                 | Reverted `flake.lock` with `git checkout HEAD -- flake.lock`. **Lesson: always `git diff` after `nix flake check` to catch lock file drift.**                                                                                                          |
| 4 | **Didn't notice go.mod was modified by Nix until tests failed** | The `git diff HEAD -- go.mod` showed `go-finding v1.0.0 → v1.1.0` and `golang.org/x/tools v0.46.0 → v0.47.0`. This was sitting in the working tree from the Nix FOD's `go mod tidy`. I only noticed when `go test` failed. | I ran `go build` and `golangci-lint` which silently used the module cache. Only `go test -race` actually compiled test binaries that exercised the runtime API differences.                                                                                                                  | Reverted go.mod/go.sum. **Lesson: check `git diff` on dependency files after ANY Nix command that might invoke `go mod tidy`.**                                                                                                                        |

---

## e) WHAT WE SHOULD IMPROVE

1. **Nix + untracked files workflow.** The Nix flake's `lib.fileset.unions` only sees git-tracked files. New files must be `git add`-ed before any `nix build` or `nix flake check`. This should be a documented checklist step, not discovered by failure.

2. **Nix lock file side effects.** `nix flake check` and `nix build` can silently update `flake.lock`. Always `git diff flake.lock` after running Nix commands. Consider `--no-update-lock-file` flag if supported.

3. **go-finding version strategy.** The project has a fundamental tension: go.mod uses published tags (v1.0.0), but Nix SSH-fetches HEAD which has a newer API. This split-brain means local development and CI/Nix build see different APIs. The proper fix (done in `7d27530`) is to adopt the latest version everywhere. But the process of getting there was painful — 6+ failed attempts.

4. **Test after EVERY change, not after a batch.** I made multiple edits across multiple files, then ran tests once. When 5 tests failed, I had to bisect to find which change broke them. Running tests after each logical change would have caught issues immediately.

5. **Integration test binary dependency.** The integration tests (`//go:build integration`) require a pre-built binary at `bin/golangci-lint-auto-configure`. The skip message now says `go build` instead of `just build`, but there's no Nix check that builds the binary for integration tests. The `nix flake check` test derivation doesn't build with `-tags=integration`.

6. **CHANGELOG discipline.** The CHANGELOG entry was added but never covers the full session's work (slog revert, types split, go-finding upgrade). A living CHANGELOG updated incrementally would be more accurate than a retroactive entry.

7. **Status report commit strategy.** Previous status reports were left as untracked files. This report should be explicitly committed or explicitly excluded.

8. **The `examples/*.golangci.yml` files still use `json: snake` for tagliatelle** — if the tool's own policy is PascalCase, the generated/example configs should reflect that. This is a consistency gap I noticed but didn't address.

---

## f) Next 25 things to get done

| #  | Task                                                                                                 | Impact | Effort | Type          |
| -- | ---------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1  | Verify go-finding v1.1.0 upgrade (commit `7d27530`) is fully correct — all SARIF/finding tests pass  | High   | 10 min | Verification  |
| 2  | Run `nix flake check` after go-finding v1.1.0 upgrade to see if build/test/race now pass             | High   | 10 min | Verification  |
| 3  | Update `vendorHash` in flake.nix if go-finding v1.1.0 changed dependencies                           | High   | 5 min  | Fix           |
| 4  | Add `git add` step to a pre-Nix-build checklist in AGENTS.md or working-with-codebase docs           | Medium | 5 min  | Documentation |
| 5  | Add MergeResult PascalCase serialization test (currently only ConfigAnalysis tested in unit tests)   | Low    | 5 min  | Testing       |
| 6  | Add JSONReport PascalCase serialization test                                                         | Low    | 5 min  | Testing       |
| 7  | Update `examples/*.golangci.yml` tagliatelle config to `json: pascal` for consistency                | Medium | 10 min | Consistency   |
| 8  | Narrow `_test.go` musttag exclusion to specific files or use per-site `//nolint`                     | Low    | 10 min | Precision     |
| 9  | Add `cbor: pascal` to tagliatelle rules in `.golangci.yml` (policy-first, before CBOR is added)      | Low    | 1 min  | Enforcement   |
| 10 | Document tag case policy in README.md (user-facing, since JSON output is a breaking change)          | Medium | 10 min | Documentation |
| 11 | Update `docs/references/testing-style-and-patterns.md` with struct tag case conventions              | Low    | 5 min  | Documentation |
| 12 | Add property-based test for JSON round-trip (marshal → unmarshal → equality) on all report types     | Medium | 20 min | Testing       |
| 13 | Audit all `nolint` directives project-wide for staleness                                             | Low    | 15 min | Cleanup       |
| 14 | Update FEATURES.md to mention PascalCase JSON as deliberate design choice                            | Low    | 5 min  | Documentation |
| 15 | Consider `--legacy-json-keys` flag for backward compatibility                                        | Low    | 30 min | Feature       |
| 16 | Add a lint rule or test that prevents new snake_case json tags from being added to report types      | Medium | 15 min | Enforcement   |
| 17 | Review whether `pkg/diff/differ.go` `Change` struct needs json tag alignment                         | Low    | 5 min  | Audit         |
| 18 | Verify `pkg/client/client.go` structs don't need tag changes                                         | Medium | 10 min | Audit         |
| 19 | Run integration tests with `-tags=integration` in CI (currently only run manually)                   | High   | 15 min | CI            |
| 20 | Add Nix check derivation for integration tests (`-tags=integration`)                                 | Medium | 20 min | CI            |
| 21 | Consider JSON Schema export from report types for API consumers                                      | Low    | 45 min | Feature       |
| 22 | Review `test.golangci.yml` — should it include tagliatelle?                                          | Low    | 5 min  | Consistency   |
| 23 | Add `flake.lock` drift check to CI (fail if `nix flake check` modifies lock file)                    | Medium | 15 min | CI            |
| 24 | Consolidate tagliatelle exclusions — many files have overlapping path rules that could be merged     | Low    | 10 min | Cleanup       |
| 25 | Add a `make verify` or Nix check that runs build + lint + test + format in one command for local dev | Medium | 10 min | DX            |

---

## g) Top #1 question I cannot figure out myself

**Should the `examples/*.golangci.yml` files be updated to `json: pascal`?**

These 5 example configs (`minimal`, `standard`, `library`, `web-project`, `cli-project`) are user-facing examples of what this tool generates. They currently have `json: snake` for tagliatelle. If the tool's own policy is PascalCase for report types, the example configs it ships should model that same policy.

BUT — these examples are for _other_ Go projects, not this tool itself. The tagliatelle config in an example project depends on what that project's JSON output needs to look like. A web API project might legitimately want snake_case JSON for browser compatibility. A CLI tool might want PascalCase.

I cannot determine whether the examples should:

- (a) Match this tool's own policy (PascalCase) for consistency, or
- (b) Show different policies per example to demonstrate flexibility, or
- (c) Stay as-is since they're examples for downstream projects, not this tool's config

**My recommendation:** Option (b) — keep `minimal` and `standard` as snake_case (most common Go web API convention), but add a comment or separate example showing PascalCase for CLI tools. This demonstrates the tool handles both.
