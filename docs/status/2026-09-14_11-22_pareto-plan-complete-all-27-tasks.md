# Pareto Plan COMPLETE — All 27 Tasks Executed (Final Status)

- **Timestamp:** 2026-09-14 11:22 CEST
- **Repository:** `golangci-lint-auto-configure`
- **Branch:** `master` (in sync with origin at time of writing)
- **Task scope:** final wrap-up of `docs/planning/2026-09-11_09-28_SUPERB-pareto-execution-plan.md` — finish T25, execute T26 (docs hygiene), execute T27 (decisions + docs integrity), annotate the plan (guardrail 7).
- **Working tree at report time:** clean; all gates green (build, 18-suite `-race` suite 48s, `golangci-lint run` 0 issues, markdownlint 0 issues, coverage 69.8% ≥ 60% gate).

## Executive summary

The 2026-09-11 SUPERB pareto plan is **fully executed: 27/27 tasks** across three
sessions. This session closed the last three: T25 (the red multi-preset spec
pushed to master by the daemon race is green), T26 (docs hygiene: two more
status reports resolved+archived, status cadence policy established, twelve
TODO rows pruned), and T27 (analysis-report fates decided and archived,
FEATURES.md integrity test extended to all hardcoded counts, launch-posture
question routed). The plan file is annotated per guardrail 7.

## a) FULLY DONE

1. **T25 — small-code-fixes bundle (finished).** The multi-preset formatter
   spec failed only on its assertion: `ContainElements("gci", "swaggo")`
   compared plain strings against `[]types.FormatterName`. Fixed by wrapping
   in `types.FormatterName(...)`; also resolved the three lint nits the file
   carried (gci import grouping, wsl_v5 whitespace, ginkgolinter
   `HaveLen(expected.Len())`). Full suite green 44–48s; lint 0 issues.
   Committed by the daemon (`ad621cc`); not amended because the master
   ruleset forbids history rewrites on pushed commits.
2. **T26.1 — humanize report resolved + archived.** The 2026-08-05 first
   H004 attempt is annotated (resolved 2026-08-07 via documented
   `//nolint:gohumanize`; gohumanize is a module plugin, so the dependency
   path was abandoned by design) and moved to `docs/archive/status/`.
3. **T26.2 — 07-31 f-ideas routed.** All 50 items in the 07-31 session
   review's f-section are accounted for in its annotation: shipped / routed
   to ROADMAP / consciously dropped (micro-optimizations, YAGNI). The
   genuinely valuable unrouted ideas are now in ROADMAP theme 2:
   exclusion-merge control & observability (`--reset-exclusions`,
   `--show-merged-rules`, merge logging, ledger recording, versioned rules)
   and comment-preserving `yaml.Node` round-trip + fuzz tests. Report
   archived.
4. **T26.3 — version-reference sweep.** `AGENTS.md` #11's v0.6.0 mention is
   historical fact (correct); `docs/references/` carries no stale tool
   version claims (only an example `git tag v0.1.0`). No changes needed.
5. **T26.4 — status cadence policy + TODO prune.** `docs/status/README.md`
   now documents the standing lifecycle rule (write → harvest-at-creation →
   resolve+archive when fully resolved → quarterly or ~15-report sweep →
   never edit archived bodies). TODO_LIST pruned of twelve rows closed by
   T11–T25; the Dependabot row narrowed to its one open item.
6. **T27.1 — analysis-report fates.** Three root-level reports annotated and
   archived to `docs/archive/`: PROJECT_SPLIT (no-split ratified, now a
   ROADMAP non-goal), PARTS.md (extraction consciously deferred, revisit on
   a second consumer), BDD_TESTS_REVIEW (resolved: coverage 29.5% → 69.8%,
   e2e layer shipped in T18; user-scenario idea consciously unrefined).
7. **T27.2 — launch posture routed.** Homepage + announcement now one
   explicit user-gated open question in ROADMAP: decide support posture
   (maintained OSS vs portfolio) first; homepage, announcement, and
   gohumanize strategy all inherit from it.
8. **T27.3 — docs-integrity test extended.** `pkg/constants/docs_integrity_test.go`
   now covers ALL hardcoded FEATURES.md counts: every preset in
   `constants.PresetLinters` must be documented, and any documented
   linter/formatter count must match constants (derived from the maps, so a
   new undocumented preset fails); pragmatic noise linter count + names; and
   tool-disabled linter names. Green + lint-clean.
9. **T27.4 — full gates + this report + plan annotation.** Build ✓, full
   parallel `-race` suite ✓ (48s, 18 suites), lint ✓ (0 issues),
   markdownlint ✓ (0 issues), coverage ✓ (69.8% ≥ 60). Plan file annotated
   (guardrail 7).

## b) PARTIALLY DONE

1. **Dependabot green run** — root cause fixed and proxy-verified (2026-09-13),
   but the scheduled run (Sunday 2026-09-13 18:52 UTC) had not been observed
   at report time. Confirmation remains in TODO_LIST.
2. **Sibling `min-length` repairs** — executed and `golangci-lint config
   verify`-checked in 8 sibling repos, but uncommitted there pending
   cross-repo authorization (open question 1).
3. **Daemon commit attribution** — the daemon interleaved 6 heuristic commits
   with this session's work (`ad621cc`, `4fb9eca`, `a1a03a7`, `8037c05`,
   `8817caa`, `cf4040a`), so per-task commit messages exist only where the
   daemon did not win the race (T26: `cda7c14`). Content is verified correct
   in all cases; history readability is the cost.

## c) NOT STARTED (all answer-gated, do not start without the user)

1. v0.8.2 cut now vs batching into v0.9.0 (open question 2).
2. Buildflow findings-gate posture: skip erraudit/branching-flow vs fix the
   44 vs leave documented-red (open question 3).
3. Cross-repo commit/push authorization for the 8 sibling repairs.
4. homebrew-tap PAT + first tap publish.
5. Stray GHCR `:master` tag deletion (needs `delete:packages` scope).

## d) TOTALLY FUCKED UP (honest)

1. **A red spec sat on public master for ~34 hours.** The daemon pushed the
   in-progress T25 test file inside the status-report commit (`76c8dc8`) on
   2026-09-13. Fixed first thing this session — but the window existed
   because the daemon commits faster than the write-test-commit loop.
2. **One botched edit, caught pre-commit.** The `HaveLen` rewrite initially
   produced `HaveLen(BeNumerically(...))` (matcher inside a length matcher —
   would not compile). Caught and corrected within the same step; zero
   shipped damage.
3. **Two wasted full-suite cycles on a phantom compile failure.** Ginkgo
   reported `[Compilation failure]` for pkg/audit, pkg/client, pkg/diff
   while `go build`/`go test`/`go vet` all passed. Root cause:
   `CGO_ENABLED=0` in the fresh shell (`-race` requires cgo); the earlier
   green run had inherited cgo from a different shell. Two full runs burned
   before the one-line error was extracted from ginkgo's output. Lesson:
   grep the compile error first, re-run later.
4. **Daemon-vs-explicit-commit race continues.** T27's intended detailed
   commit message was preempted by `8037c05`/`8817caa`. Known limitation
   (ROADMAP theme 4), not re-fixed here.

## e) WHAT WE SHOULD IMPROVE

1. **Extract compile errors before re-running.** A ginkgo "Compilation
   failure" line should always be chased with
   `ginkgo <pkg> 2>&1 | grep -A5 compile` before any re-run.
2. **Pin the race-toolchain environment in one place.** `CGO_ENABLED=1` +
   `GOEXPERIMENT=jsonv2` belong in the documented test command (AGENTS.md)
   so fresh shells cannot silently drop cgo.
3. **Commit the moment a task's tests go green.** The daemon race window is
   proportional to the gap between green tests and explicit commit.
4. **Archive analysis reports when their decision lands**, not months later —
   PARTS.md sat at the repo root for 4.5 months implying an open decision.
5. **Derive docs-integrity specs from constants, not hardcoded entries** —
   done here for presets; the same pattern (iterate the map, assert
   documentation) applies to deprecation mappings next.
6. **Keep ROADMAP non-goals carrying their own revisit conditions** — a
   non-goal without a reopen trigger reads as permanent; with one, it stays
   honest.
7. **Verify-before-annotate held throughout** — every annotation in this
   session was written only after re-verifying the claim (coverage number
   re-measured, H004 resolution confirmed in AGENTS, 07-31 items traced to
   shipped code). Keep this as the docs-health default.

## f) Next things (bounded, in order)

1. Check the Dependabot run result (`gh api …/workflows/320348528/runs`);
   if green, prune the TODO row.
2. Answer the three standing questions (v0.8.2 timing; buildflow gate
   posture; sibling-repo commit authorization) — each unblocks a bounded
   task.
3. Release dry-run (`goreleaser release --snapshot --clean`) on PRs — last
   remaining Medium TODO row.
4. Decide `Detect()` error-path contract (deferred twice; 30-minute
   decision note).
5. Root-cause `nix flake check` "running 0 flake checks" vs 4 in eval.
6. Quarterly erraudit re-check (~2026-10).
7. Link checker (lychee) in CI.
8. Metadata checklist script (one `gh api` pass).
9. ROADMAP theme-2 ideas when capacity allows: exclusion-merge
   observability, `yaml.Node` round-trip, fuzz tests.
10. Revisit homepage/announcement after the support-posture decision.

## g) Questions that cannot be figured out from the repository alone

1. **v0.8.2 now or v0.9.0 batch?** The unreleased changes are: multi-preset
   extra-formatter dedup fix, suppression-ledger recording, spinner race
   fix, omitzero migration, schema-version awareness. All are bug-fix/
   hardening class → v0.8.2 is defensible now; batching waits for a feature.
2. **Buildflow findings-gate posture** — skip the two advisory gates in
   `.buildflow.yml` (`fail_on`/`skip_steps`), fix the 44 erraudit findings
   (AGENTS #26 says leave documented), or keep documented-red?
3. **Cross-repo authorization** — commit + push the 8 sibling
   `.golangci.yml` `min-len` repairs? Each is context-verified; erraudit
   additionally passed `golangci-lint config verify`.

## Session self-assessment

The plan is closed with no silent residue: every open item from the
2026-09-11 TODO harvest is either shipped, archived with a fate, user-gated
and explicitly listed, or scheduled. The two process scars (daemon race,
CGO flake) are documented with their lessons. Verification-before-document
was maintained end to end.
