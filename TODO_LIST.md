# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-09-14

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

The 2026-09-13/14 pareto execution (tasks T11–T25, see
`docs/status/2026-09-11_23-26_pareto-execution-t11-t25-thirteen-tasks-and-honest-scars.md`)
closed twelve more rows: sibling `min-length` sweep, GHCR v0.8.0 backfill,
branch/tag rulesets, README claim audit, ADR consolidation, gitleaks
full-history scan, schema-version awareness, `internal/cli` coverage +
suite speedup (116s → 48s parallel), `omitzero` migration, multi-preset
merge tests, `FindingsHidden` removal, `FixConfig` e2e integration test,
goconst-drift question, `auto-tag.yml` deletion, and the README CI/CD
section.

---

## High Priority

(none — all five closed 2026-09-11, see header note)

## Medium Priority

| Task                                                                                                                                                               | Impact                                                                                                                     | Effort | Evidence                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Release dry-run (`goreleaser release --snapshot --clean`) on PRs                                                                                                   | High — catches dockers_v2/buildx issues before a tag exists                                                                | M      | `docs/status/2026-09-11_08-33…md` e6/f8                                                                                                      |
| Confirm Dependabot green run — root cause fixed 2026-09-13 (`go-finding` negatively cached as 404 on proxy.golang.org; cleared, verified via proxy-only `go list`) | Medium — six consecutive red runs since Aug 2; fix applied and proxy-verified, the green scheduled run is still unproven   | S      | next scheduled run Sunday 2026-09-13 18:52 UTC; `docs/status/2026-09-11_23-26…md` T14                                                         |
| Buildflow findings-gate posture: e2e fails only on advisory gates (branching-flow 190, erraudit 44 — manual-review per AGENTS #26); test/lint/fmt steps green; "9 tools unavailable" health check is JS/TS+Python noise for a Go repo | Medium — decide per-gate policy in `.buildflow.yml` (`fail_on:`/`skip_steps:`) vs fixing findings — USER-GATED (open question 3) | M      | `docs/status/2026-09-11_06-38…md` d1/f1/f3; timeout claim resolved 2026-09-11: no step-level timeout exists (BuildFlow `step_options.go`), `go test -timeout=10m` binds |

## Low Priority

| Task                                                                                                   | Impact   | Effort   | Evidence                                                                                                  |
| ------------------------------------------------------------------------------------------------------ | -------- | -------- | --------------------------------------------------------------------------------------------------------- |
| Swallowed-error governance audit (erraudit + periodic re-check)                                        | Low      | 1h       | 194 findings reviewed 2026-07-30; 141 `context_loss` are noise; next quarterly re-check due ~2026-10      |
| Decide `Detect()` error-path contract (`(ProjectType, error)` vs nil-on-error)                         | Low      | decision | `docs/status/2026-07-26_22-07…md` f20/g2; deferred twice                                                  |
| Root-cause `nix flake check` "running 0 flake checks" vs 4 checks in eval                              | Low      | S        | `docs/status/2026-09-11_06-38…md` d3/f5                                                                   |
| Link checker (lychee) in CI + render-check struck archive tables                                       | Low      | S        | `docs/status/2026-09-09_02-08…md` f23; surfaced by the 2026-09-11 archive pass                            |
| Metadata checklist script (description/topics/badges/workflow-state/release-page in one `gh api` pass) | Low      | M        | `docs/status/2026-09-11_08-33…md` e8/f25                                                                  |
