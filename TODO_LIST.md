# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-10-07

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`;
long-term ideas live in `ROADMAP.md`. **This file contains OPEN work only** —
when a task ships, remove it here and record it in `CHANGELOG.md`.

Harvest history: the 2026-09-11 High tier + 20% tier closed with v0.8.1
(`docs/archive/status/2026-09-11_13-27…md`); the 2026-09-13/14 pareto execution closed
fifteen more rows (T11–T25, `docs/status/2026-09-11_23-26…md`). The 2026-10-07
docs-health sweep re-verified every row below against the repo and harvested
the Go-1.27 frontier (`docs/status/2026-09-28_21-59…md`) plus residue that
three July reports surfaced but no row ever tracked (dogfood gate, flake lint
output, gitleaks cadence).

---

## High Priority

(none)

## Medium Priority

| Task                                                                                                                                                                     | Impact                                                                                                                                                 | Effort | Evidence                                                                                                                                            |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Release dry-run (`goreleaser release --snapshot --clean`) on PRs                                                                                                         | High — catches dockers_v2/buildx issues before a tag exists                                                                                            | M      | `docs/status/2026-07-27_01-42…md` f16                                                                                                       |
| Confirm Dependabot green run — root cause fixed 2026-09-13 (`go-finding` negatively cached as 404 on proxy.golang.org; cleared, verified via proxy-only `go list`)       | Medium — scheduled runs work (2026-09-27 batch opened 6 PRs), but every `go_modules` PR fails CI on the vendorHash drift check, which Dependabot can't commit | S      | failed run 36342326659: only failing step "Build with Nix (vendorHash guard)"; needs a CI policy decision (auto-commit on dependabot branches vs `continue-on-error`); `github_actions` bumps pass fine |
| Buildflow findings-gate posture: e2e fails only on advisory gates (branching-flow 190, erraudit 44 — manual-review per AGENTS #26); test/lint/fmt steps green            | Medium — decide per-gate policy in `.buildflow.yml` (`fail_on:`/`skip_steps:`) vs fixing findings — USER-GATED (ROADMAP open question 3)               | M      | `docs/status/2026-09-11_06-38…md` d1/f1/f3; no step-level timeout exists (BuildFlow `step_options.go`), `go test -timeout=10m` binds       |
| Fleet sweep: re-run `configure` across ~160 sibling projects so patched-form `run.go` values get normalized/capped (highest real-world impact of the Go-1.27 work)       | Medium — 88 machine-generated sibling configs still carry pre-Go-1.27 `run.go` forms                                                                   | M      | `docs/status/2026-09-28_21-59…md` c1/f14; `KnownBadSettingsKeys` self-heal covers settings keys, not `run.go`                                       |
| Dogfood gate: run this tool's `configure --check` against the repo's own `.golangci.yml` in CI                                                                           | Medium — the tool never verifies its own config in CI; surfaced by three July reports, never tracked                                                   | S      | `.github/workflows/ci.yml` has no such step; `cmd_check_test.go` proves the flag works                                                              |

## Low Priority

| Task                                                                                                   | Impact | Effort    | Evidence                                                                                             |
| ------------------------------------------------------------------------------------------------------ | ------ | --------- | ---------------------------------------------------------------------------------------------------- |
| Swallowed-error governance audit (erraudit + periodic re-check)                                        | Low    | 1h        | Quarterly re-check done 2026-10-07: 81 type-aware findings (81 vs 218 at 2026-07-30 — erraudit's own context_loss noise shrank), 0 new real bugs, all accounted (25 sentinels + 4 coverage-check + 6 coverage-check fmt.Errorf + 1 Join + 16 idiomatic ignored + 29 noise); next due ~2027-01 |
| Decide `Detect()` error-path contract (`(ProjectType, error)` vs nil-on-error)                         | Low    | decision  | `docs/status/2026-07-26_22-07…md` f20/g2; deferred twice                                            |
| ~~Root-cause `nix flake check` "running 0 flake checks" vs 4 checks in eval~~ RESOLVED 2026-10-07: unreproducible on nix 2.34.8 — `nix flake check` enumerates and checks all 5 x86_64-linux derivations (build/format/race/treefmt/vendor-hash) + packages + apps, green; eval shows 3 systems × 5 checks. Historical 2-vs-0 flapping likely stale eval on the older nix or mid-session flake.lock mutation by the buildflow nix-checker; re-open only if it reproduces | Low    | S         | `docs/status/2026-09-11_06-38…md` d3/f5                                                     |
| Link checker (lychee) in CI + render-check struck archive tables                                       | Low    | S         | `docs/status/2026-09-09_02-08…md` f23                                                               |
| Metadata checklist script (description/topics/badges/workflow-state/release-page in one `gh api` pass) | Low    | M         | `docs/status/2026-09-11_08-33…md` e8/f25                                                    |
| Remove inert `goexperiment.jsonv2` from `.golangci.yml` `build-tags` (inert on Go 1.27); flake/CI env removal is USER-GATED | Low    | XS        | `.golangci.yml:8`; `docs/status/2026-09-28_21-59…md` f18/g1                                        |
| Dockerfile: slim variant — build it as an extra tag or delete the dead commented block (`FROM alpine … slim`) | Low    | S         | `Dockerfile:62`; `docs/status/2026-09-28_21-59…md` f20                                             |
| Pin a golangci-lint version matrix in e2e CI to exercise the `run.go` cap path against an older binary (e.g. v2.12.2) | Low    | M         | `docs/status/2026-09-28_21-59…md` c7/f13; cap path only tested against the pinned v2.14.0          |
| Scheduled gitleaks job (full-history scan ran once manually 2026-09-11: 0 secrets over 1,117 commits)  | Low    | S         | ROADMAP theme 6 / `docs/status/2026-09-11_23-26…md` f21                                            |
| Surface `golangci-lint run` as a `flake.nix` check output (checks currently: race only)                | Low    | S         | `flake.nix` checks block; surfaced by 2026-07-25 friction report f34                               |
| ~~Coverage gate 60 → 65 (suite total 69.8% after T18/T19; headroom exists)~~ DONE 2026-10-07: total 72.5%, `-min=65` in ci.yml + pre-release-check + README (per-package weakest internal/cli 49.4% — total-gated, not blocking) | Low    | XS        | `.github/workflows/ci.yml` `-min=65`; `docs/archive/status/2026-09-14_11-34…md` f13                        |
| GHCR hygiene: delete stray `:master` tag (needs `delete:packages` scope — USER-GATED) + smoke-test backfill-image on a fresh tag | Low    | S         | `docs/archive/status/2026-09-14_11-34…md` c5/c15                                                           |
| README "Go version handling" user-facing section (how `run.go` is written, capped, and rescued)        | Low    | S         | `docs/status/2026-09-28_21-59…md` f19; CHANGELOG [Unreleased] carries the behavior                  |
