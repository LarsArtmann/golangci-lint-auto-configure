# Status Report: Buildflow Failures Resolved (Generator Formatting + vendorHash)

**Date:** Fri Sep 11 06:38 CEST 2026
**Session scope:** Triage and fix the 3 buildflow failures from the user-pasted run (treefmt, go-auto-upgrade, test-coverage), plus the nix-build/vendorHash breakage behind them.
**Headline:** All 3 failures diagnosed; 2 fixed at root cause in-repo, 2 diagnosed as external-tool issues (buildflow timeout, migration tool self-revert). Repo-level gates all green: build, full `-race` suite, treefmt, lint on changed packages, all 4 flake checks.

---

## Meta: Direct Answers to Your Three Questions

**What did you forget?**
- I never re-ran buildflow itself — the tool that produced the original failure report. I verified repo-equivalents (`nix flake check`, `go test -race`, treefmt) but the original report is closed only by proxy, not end-to-end.
- I never investigated the "⚠ 9 tools unavailable (health check failed)" buildflow warning. Dismissed as environmental without running `--verbose`.
- I noticed two recurring gopls warnings (stdversion false positive on `json.Unmarshal`; "No packages found" for the `generate`-tagged file) and left both unfixed — only documented the second.
- I noticed golangci-lint now warns `exhaustruct is deprecated (since v2.13.0)` on every run and did nothing.
- My AGENTS.md gotcha #35 edit was still uncommitted at report time (auto-commit daemon pending).
- I never root-caused why `nix flake check` reported "running 0 flake checks" on the final run vs 2 earlier — I worked around it by explicitly building all 4 checks, but the anomaly is unexplained.

**What could you have done better?**
- Should have checked for a buildflow config (repo-level `.buildflow.yml` is referenced by AGENTS.md #13 but absent; global config location never looked up) before concluding the timeout was unfixable in-repo.
- Should have run the full-repo `golangci-lint run`, not just the two changed packages.
- Should have written the drift guard (regenerate + `git diff --exit-code` in CI) in the same session instead of deferring it to the task list — the generator can silently rot against the schema otherwise.
- Could have coordinated with the auto-commit daemon: it committed my generator fix (8e86636) mid-session, interleaving history confusingly.

**What could you still improve?**
- The daemon/go.mod/vendorHash footgun needs a systemic fix (CI check or daemon awareness), not the manual procedure in AGENTS.md #3 — it broke the build this session and will recur.
- The `internal/cli` suite takes 116s with `-race`; it is the direct cause of the buildflow timeout class of failures.
- The generated-file workflow deserves a discoverability pass: `//go:generate` line, docs outside AGENTS.md, drift guard.

---

## a) FULLY DONE

| # | Item | Evidence | Scope |
|---|------|----------|-------|
| 1 | Generator root-cause fix: `cmd/generate-settings` now runs output through `go/format` before writing and emits `struct{}` for empty settings structs (gofumpt rule) | Commit `8e86636`; `gofumpt -l` clean; new tests 2/2 pass | `cmd/generate-settings/main.go` |
| 2 | Regenerated `linter_settings_generated.go` (88 structs) — now gofumpt-clean; whole repo treefmt-clean | `nix fmt`: traversed 510 files, **0 changed** | `pkg/constants/linter_settings_generated.go` |
| 3 | `vendorHash.nix` updated after the daemon's go.mod/go.sum dependency bumps (go-finding v1.10.0, ginkgo v2.32.1, ultraviolet, go-runewidth) | `nix build` OK after `sha256-FeWRnA7f...`; commit `3617f93` | `vendorHash.nix` |
| 4 | Regression test for the generator (gofmt-clean output; empty-struct `struct{}`) | `go test ./cmd/generate-settings/` ok; commit `3617f93` | `cmd/generate-settings/main_test.go` (new) |
| 5 | go-auto-upgrade failure triaged: transient — the migration tool restored its own backups; no `samber/lo` anywhere in tree | `rg "samber/lo" go.mod pkg/types/clone.go` → no matches; `go build ./...` OK | none needed |
| 6 | test-coverage failure triaged: not a test failure — buildflow tool timeout (~26s) kills a suite where `internal/cli` alone takes **116s** with `-race` | Full suite passes locally: 21 packages ok | diagnosis only |
| 7 | Full verification pass | `go build ./...`; `go test -race ./pkg/... ./internal/... ./cmd/...` all ok; `golangci-lint run` on changed pkgs: 0 issues; `nix build` + all 4 flake checks (`build`, `format`, `race`, `treefmt`) build green | repo-wide |
| 8 | AGENTS.md gotcha #35 documenting the generator contract (formatted-at-generation-time; `//go:build generate` semantics; expected gopls warning) | Written this session; uncommitted at report time | `AGENTS.md` |

## b) PARTIALLY DONE

| Item | Works now | Remains open | Blocker | Effort |
|------|-----------|--------------|---------|--------|
| Original buildflow report closure | All repo-level gates green (build/tests/treefmt/flake/lint-changed) | A green end-to-end buildflow run; the `test-coverage` step will fail again on a cold cache until the tool timeout is raised | Timeout lives in buildflow config (global or missing in-repo) — not located this session | S–M |
| AGENTS.md documentation | Gotcha #35 content written | Commit (daemon pending) | none | S |
| Diagnostics hygiene | Both gopls warnings root-understood and documented | gopls `buildFlags`/`directoryFilters` config change to silence them | none, just not done | S |
| `nix flake check` trust | Explicit `nix build .#checks.x86_64-linux.{build,format,treefmt,race}` verified green | Explanation for the "running 0 flake checks" vs 2-check discrepancy between runs | unknown — needs investigation | S |

## c) NOT STARTED

Nothing was started for (planned, zero code touched):

- **Buildflow `test-coverage` timeout fix** — blocked on locating the buildflow config (repo `.buildflow.yml` absent; global config not researched per session scope).
- **Generated-file drift guard in CI** — clearly needed (schema updates + committed reference file), not begun.
- **`exhaustruct` → `exhaustruct_v5` migration** — new deprecation warning observed this session; not begun; porting the curated stdlib excludes (AGENTS.md #19) is required when done.
- **json/v2 `omitempty` → `omitzero` in `pkg/types/config_types.go`** — pre-existing latent bug (AGENTS.md #17); untouched, correctly out of scope for this session.
- **Resolution of the two pre-existing CI failures documented at v0.8.0** (`ff88e51` / TODO_LIST.md) — not looked at this session.

The full ranked backlog is section (f).

## d) TOTALLY FUCKED UP

1. **buildflow `test-coverage` is structurally broken for this repo.** The tool timeout killed `go` after 26s while the suite needs ≥116s (`internal/cli` alone) with `-race`. Every cold-cache run fails regardless of code health. Workaround: warm cache, which CI sandbox runs do not reliably have. Real fix: raise the timeout in buildflow config (location not yet found).
2. **Auto-commit daemon + vendorHash interplay silently breaks `nix build` for everyone.** The daemon committed go.mod/go.sum bumps (`8645cd3`) without updating `vendorHash.nix`; the next `nix build` on any machine fails with a hash mismatch until someone applies the manual AGENTS.md #3 procedure. Severity: propagates broken builds to all clones and CI. Root cause: daemon heuristic has no vendorHash awareness. This will recur.
3. **`nix flake check` reporting is untrustworthy right now.** Same session, same flake: "running 2 flake checks" then later "running 0 flake checks", while `nix eval .#checks.x86_64-linux` lists 4 (`build`, `format`, `race`, `treefmt`). Both runs ended "all checks passed", so green is real — but a primary verification gate that nondeterministically reports zero checks is a trust problem. Unknown cause.
4. **`go-auto-upgrade` performs migrations that break compilation, then self-reverts.** It attempted a `samber/lo` migration (package not in this module), failed, restored 13 files. Net effect: wasted compute and a misleading red signal in every report. The tool appears misconfigured/scoped incorrectly for this repo.

## e) WHAT WE SHOULD IMPROVE

| Pattern | Pain | Concrete fix |
|---------|------|--------------|
| Manual vendorHash procedure after daemon dep bumps | Broken `nix build` on every bump; bit this session | CI step (or pre-push hook) that runs `nix build` on go.mod/go.sum changes and posts the `got:` hash; or teach the daemon |
| Verifying via proxy instead of the failing tool | Original report never closed end-to-end | Always re-run the actual failing pipeline as the final step; treat equivalence-checks as intermediate only |
| Gopls noise (stdversion, build-tag file) | Recurring warnings mislead agents into wrong "fixes" (e.g. bumping go directive) | Add `GOEXPERIMENT=jsonv2` to gopls buildFlags; exclude `linter_settings_generated.go` via directoryFilters |
| Deprecation warnings left running | `exhaustruct` deprecation prints on every lint; rot accumulates | Migrate to `exhaustruct_v5` with the curated excludes ported |
| Slow `-race` suite (116s) | Direct cause of timeout-class CI/tool failures | Profile and parallelize `internal/cli` specs; trim sleeps |
| Undocumented dev tooling | `cmd/generate-settings` was discoverable only via AGENTS.md | Add to `docs/references/working-with-codebase.md`; add `//go:generate` line |
| Mid-session daemon commits | Interleaved history, confusing attribution | Batch or gate daemon commits during active agent sessions |

## f) NEXT TASKS (ranked; HARVEST input for TODO_LIST.md / ROADMAP.md)

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Raise buildflow `test-coverage` tool timeout (suite needs ≥5 min cold with `-race`) via global config or committed `.buildflow.yml` | Critical | S | Quality |
| 2 | Fix daemon/vendorHash footgun: CI (or pre-push hook) runs `nix build` on go.mod/go.sum changes and surfaces the `got:` hash | Critical | M | Bug |
| 3 | Re-run full buildflow end-to-end; confirm 0 failures (close the original report) | Critical | S | Quality |
| 4 | Add generated-file drift guard: regenerate in CI, `git diff --exit-code` on `linter_settings_generated.go` | High | S | Quality |
| 5 | Root-cause `nix flake check` "running 0 flake checks" vs 4 checks present in eval | High | S | Bug |
| 6 | Investigate buildflow "9 tools unavailable" health check with `--verbose`; fix PATH or document | High | S | Quality |
| 7 | Disable/scope the `go-auto-upgrade` buildflow tool for this repo (its migrations break compilation then self-revert) | High | S | Quality |
| 8 | Resolve the two pre-existing CI failures documented at v0.8.0 (TODO_LIST.md via `ff88e51`) | High | M | Bug |
| 9 | Migrate `exhaustruct` → `exhaustruct_v5`; port curated stdlib excludes (AGENTS.md #19) | Medium | M | Quality |
| 10 | Fix latent json/v2 `omitempty` → `omitzero` in `pkg/types/config_types.go` (AGENTS.md #17) | Medium | M | Bug |
| 11 | Speed up `internal/cli` suite (116s with `-race`): profile, parallelize, trim | Medium | L | Quality |
| 12 | Add `GOEXPERIMENT=jsonv2` to gopls buildFlags; exclude generated file via directoryFilters | Medium | S | Quality |
| 13 | Run full-repo `golangci-lint run` (this session covered changed packages only) | Medium | S | Quality |
| 14 | Verify the wsl v5 path end-to-end: is generated `WslV5Settings` ever emitted by `injectDefaultSettings`, or only curated `WslSettings`? Add BDD spec | Medium | M | Feature |
| 15 | Document `cmd/generate-settings` in `docs/references/working-with-codebase.md`; add `//go:generate` line | Medium | S | Documentation |
| 16 | Add unit tests for generator helpers (`schemaTypeToGo`, `toPascalCase` edge cases) | Low | S | Quality |
| 17 | Determinism guard: run generator twice, assert byte-identical output | Low | S | Quality |
| 18 | CHANGELOG entry for the generator gofmt fix + regenerated file | Medium | S | Documentation |
| 19 | Confirm daemon-injected `.golangci.yml` additions (goconst/nestif/tagalign in `8645cd3`) were intentional self-config | Low | S | Documentation |
| 20 | Commit AGENTS.md gotcha #35 (pending daemon) | Medium | S | Documentation |
| 21 | Refresh or delete stale root `coverprofile.out` (dated Jun 18) | Low | S | Cleanup |
| 22 | Act on or archive `BDD_TESTS_REVIEW.md` findings (Sep 2) | Low | M | Cleanup |
| 23 | Add `--version` smoke check (ldflags correctness) to flake checks | Low | S | Quality |
| 24 | Add `nix run .#coverage-check` smoke test to flake checks | Low | S | Quality |
| 25 | Audit flake.lock update policy: scheduled `nix flake update` job vs daemon heuristic | Low | M | Quality |
| 26 | Consider Renovate/Dependabot so Go dep bumps are deliberate PRs, not daemon heuristic commits | Medium | M | Quality |
| 27 | Add pre-push hook running `nix flake check` to catch vendorHash/treefmt drift before remote | Medium | S | Quality |
| 28 | Investigate buildflow cache: 0% hit rate (36 misses) this run — why cold? | Low | S | Quality |
| 29 | Decide `examples/api-usage` story: own module/tests, or stop it polluting coverage output | Low | S | Cleanup |
| 30 | Clarify whether flake `race` check should cover `./cmd/...` tests too | Low | M | Quality |
| 31 | Make flake checks multi-system or explicitly gate (`--all-systems` currently omits aarch64/darwin) | Low | S | Quality |
| 32 | Verify go-finding v1.10.0 / ginkgo v2.32.1 bumps against integration docs; update referenced versions | Low | S | Documentation |
| 33 | Prune AGENTS.md (28KB): move narrative to docs/references, keep gotchas terse | Low | M | Documentation |
| 34 | Harvest this report's section (f) into TODO_LIST.md / ROADMAP.md via docs-health HARVEST | High | S | Documentation |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Where does the buildflow `test-coverage` tool timeout live, and may I change it?** The repo has no `.buildflow.yml` (AGENTS.md #13 references one; it's absent). If it's in your global buildflow config, I need its path and your OK to raise the timeout (the suite legitimately needs ~2 min warm, 5+ cold with `-race`). Alternatively: should I create a committed `.buildflow.yml` override in-repo so every machine/CI gets the fix?
2. **Are the two pre-existing CI failures documented at the v0.8.0 tag push (`ff88e51`) still wanted on the backlog — and are they safe to reproduce now that the daemon bumped go-finding/ginkgo, or do you already know they're environment-specific?** I deliberately did not research them this session.
3. **`exhaustruct` is deprecated upstream (v2.13.0 → `exhaustruct_v5`).** AGENTS.md #19 records curated friction-driven excludes for the old linter. Do you want the migration now (porting the excludes), or do you want to drop `exhaustruct` from this repo's own config entirely?

---

**Process note:** Section (f) is the primary input for `docs-health` HARVEST into TODO_LIST.md / ROADMAP.md. Per your instruction, I am waiting for instructions before harvesting.

**Format note:** The status-report skill's canonical output is a styled HTML dashboard; you explicitly requested `.md`, so this report is Markdown (one-off override, not propagated into the skill).
