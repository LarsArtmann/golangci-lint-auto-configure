# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-09-11

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`;
long-term ideas live in `ROADMAP.md`. **This file contains OPEN work only** —
when a task ships, remove it here and record it in `CHANGELOG.md`.

Harvested from `docs/status/2026-09-11_06-38_buildflow-failures-resolved.md`
(section f) and `docs/status/2026-09-11_08-33_github-metadata-and-ci-rehabilitation-status.md`
(section f) on 2026-09-11. All High items + the 20% tier were closed on
2026-09-11 by the v0.8.1 delivery (key-normalization self-heal, CI schema
gate, release pipeline end-to-end, vendorHash guard, exhaustruct_v5
migration, spec debt, drift guards, watchdog, `go install` e2e) — see
`docs/status/2026-09-11_13-27_pareto-execution-v081-shipped-and-guards-installed.md`.

---

## High Priority

(none — all five closed 2026-09-11, see header note)

## Medium Priority

| Task                                                                                | Impact                                                                                     | Effort | Evidence                                                                                                                                        |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Release dry-run (`goreleaser release --snapshot --clean`) on PRs                     | High — catches dockers_v2/buildx issues before a tag exists                                | M      | `docs/status/2026-09-11_08-33…md` e6/f8                                                                                                        |
| Sweep ~160 sibling projects for emitted `goconst.min-length` and repair              | High — downstream configs are hard-broken on golangci-lint <2.13                           | M      | `docs/status/2026-09-11_08-33…md` c4/f4                                                                                                        |
| Backfill GHCR image for v0.8.0 (buildx build + push from existing tag)               | Medium — container installs blocked for all three latest releases                          | S      | `docs/status/2026-09-11_08-33…md` d2/c3/f7                                                                                                     |
| Branch/tag protection rulesets for `master` + `v*`                                   | Medium — tag push rights currently unguarded on a public repo                              | S      | `docs/status/2026-09-11_08-33…md` c8/f14/f33                                                                                                   |
| Full `README.md` claim-by-claim audit (~570 lines)                                   | Medium — repeated spot-checks, never line-by-line                                          | 2h     | carried since 2026-07-25; only Requirements/example output/CI/Related verified                                                                  |
| Consolidate `ARCHITECTURE.md` inline ADRs into `docs/adr/` + fix naming split (8 inline ADRs vs 6 numbered files + 1 oddly-named `001-yaml-dependency-decision.md`) | Medium — ADRs live in two places with inconsistent naming conventions                      | 1–2h   | verified 2026-09-11: `docs/ARCHITECTURE.md` (8 inline) vs `docs/adr/` (7 files, 2 naming schemes)                                              |
| gitleaks full-history scan (tip was swept 2026-09-09, history was not)               | Medium — public repo; history contains pre-sanitization sibling references (accepted risk, but should be quantified) | M      | `docs/status/2026-09-09_02-08…md` c5/f5                                                                                                        |
| Dependabot Updates workflow failures (2026-08-23, 08-30, 09-06)                      | Medium — three consecutive red runs unnoticed                                              | S      | `docs/status/2026-09-11_08-33…md` c6/f13                                                                                                       |
| Buildflow findings-gate posture: e2e fails only on advisory gates (branching-flow 190, erraudit 44 — manual-review per AGENTS #26); test/lint/fmt steps green; "9 tools unavailable" health check is JS/TS+Python noise for a Go repo | Medium — decide per-gate policy in `.buildflow.yml` (`fail_on:`/`skip_steps:`) vs fixing findings | M      | `docs/status/2026-09-11_06-38…md` d1/f1/f3; timeout claim resolved 2026-09-11: no step-level timeout exists (BuildFlow `step_options.go`), `go test -timeout=10m` binds                 |
| Schema-version awareness note/gate in `cmd/generate-settings`                        | Medium — regenerated against v2.13 schema while tool advertises v2.10.1 minimum (root cause of the `min-length` incident) | M      | `docs/status/2026-09-11_08-33…md` d1/f21                                                                                                       |
| `internal/cli` coverage (25.8%, weakest package): BDD specs for untested command paths + speed up the 116s `-race` suite | Medium — largest test-debt pocket; suite length causes buildflow timeout-class failures    | M–L    | `docs/status/2026-09-11_08-33…md` f24; `docs/status/2026-09-11_06-38…md` f11                                                                   |
| json/v2 `omitempty` → `omitzero` in `pkg/types/config_types.go`                      | Medium — JSON-format config output emits `false`/`0` for zero values; YAML output unaffected | M      | AGENTS.md gotcha #17 (latent); re-confirmed open 2026-09-11 (`docs/status/2026-09-11_06-38…md` f10)                                            |

## Low Priority

| Task                                                                  | Impact | Effort | Evidence                                                                                                        |
| ---------------------------------------------------------------------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------- |
| Extend docs-integrity test to ALL hardcoded FEATURES.md counts        | Low    | 1h     | `pkg/constants/docs_integrity_test.go` covers preset counts only (verified 2026-09-11: single `Describe`)       |
| Status-report lifecycle cadence policy (archive sweep done 2026-09-11; cadence remains) | Low    | 30min  | `docs/status/README.md` index exists; establish quarterly-or-N-reports rule                                     |
| Multi-preset merge correctness tests (dedup, formatter union)         | Low    | 1h     | verified 2026-09-11: no test combines two presets (`internal/cli/*_test.go` grep empty)                         |
| Swallowed-error governance audit (erraudit + periodic re-check)       | Low    | 1h     | 194 findings reviewed 2026-07-30; 141 `context_loss` are noise; next quarterly re-check due ~2026-10            |
| Implement-or-remove `FindingsHidden` dead (always-zero) ledger field  | Low    | 30min  | `docs/status/2026-07-20_22-57…md` open decision                                                                |
| Integration test: full `FixConfig` flow with real sidecar + ledger (never-enable end-to-end) | Low    | M      | recurring open item since 2026-07-25 (`docs/status/2026-07-30_23-22…md` c1)                                    |
| Decide `Detect()` error-path contract (`(ProjectType, error)` vs nil-on-error) | Low    | decision | `docs/status/2026-07-26_22-07…md` f20/g2; deferred twice                                                      |
| Decide `linter_settings_generated.go` goconst drift (regenerate + annotate vs exclude goconst) | Low    | S      | `docs/status/2026-09-11_08-33…md` b4/f18                                                                       |
| Root-cause `nix flake check` "running 0 flake checks" vs 4 checks in eval | Low    | S      | `docs/status/2026-09-11_06-38…md` d3/f5                                                                        |
| Add a short "CI/CD" section to README (what runs, what gates)          | Low    | S      | `docs/status/2026-09-11_08-33…md` f46                                                                          |
| Act on or archive `BDD_TESTS_REVIEW.md`, `PARTS.md`, `PROJECT_SPLIT_EXECUTIVE_REPORT.md` findings | Low    | M      | `docs/status/2026-09-11_06-38…md` f22; f16 (root reports are one-off analyses, undecided)                      |
| Decide fate of `auto-tag.yml` (delete vs keep-disabled; currently orphaned) | Low    | S      | `docs/status/2026-09-11_08-33…md` c7/f12                                                                       |
| Link checker (lychee) in CI + render-check struck archive tables                       | Low    | S      | `docs/status/2026-09-09_02-08…md` f23; surfaced by the 2026-09-11 archive pass                                  |
| Metadata checklist script (description/topics/badges/workflow-state/release-page in one `gh api` pass) | Low    | M      | `docs/status/2026-09-11_08-33…md` e8/f25                                                                        |
