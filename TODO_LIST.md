# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-30

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`;
long-term ideas live in `ROADMAP.md`. **This file contains OPEN work only** —
when a task ships, remove it here and record it in `CHANGELOG.md`.

---

## High Priority

_All high-priority items resolved in v0.6.0 — see CHANGELOG.md._

## Medium Priority

| Task                                                       | Impact                                                                                                                                                                                            | Effort | Evidence                                                                                                                                                                                                                                       |
| ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| exhaustruct → exhaustruct_v5 migration                     | Medium — golangci-lint v2.13 deprecates `exhaustruct` ("Replaced by exhaustruct_v5"); glac's priority table, exclusion rules, and curated settings reference the old name and will warn on v2.13+ | 2–3h   | golangci-lint v2.13 startup warning observed 2026-09-10 in `golangci-lint run` across LarsArtmann repos; constants touched: `linter_priorities.go`, `linter_reasons.go`, `linter_settings.go`, `config.go` (DefaultExclusionRules), `rules.go` |
| Consolidate `ARCHITECTURE.md` inline ADRs into `docs/adr/` | Medium — ADRs live in two places (split-brain)                                                                                                                                                    | 1–2h   | `docs/ARCHITECTURE.md` has 8 inline ADRs; `docs/adr/` has 6 separate files                                                                                                                                                                     |
| Full `README.md` claim-by-claim audit (~500 lines)         | Medium — repeated spot-checks, never line-by-line                                                                                                                                                 | 2h     | Only Requirements / example output / CI / Related Projects verified                                                                                                                                                                            |

## Low Priority

| Task                                                                                                                                                                                                                                                                               | Impact | Effort | Evidence                                                                         |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------- |
| Extend docs-integrity test to ALL hardcoded FEATURES.md counts                                                                                                                                                                                                                     | Low    | 1h     | `pkg/constants/docs_integrity_test.go` only covers preset counts                 |
| Status report lifecycle policy (archive cadence)                                                                                                                                                                                                                                   | Low    | 30min  | `docs/status/README.md` index exists; no archive cadence policy                  |
| Multi-preset merge correctness tests (dedup, formatter union)                                                                                                                                                                                                                      | Low    | 1h     | `--preset a --preset b` shipped without dedicated merge tests                    |
| Swallowed-error governance audit (erraudit + periodic re-check)                                                                                                                                                                                                                    | Low    | 1h     | 194 findings reviewed 2026-07-30; 141 `context_loss` are noise; re-run quarterly |
| GoReleaser docker publish fails on every release: `--attest=type=sbom` unsupported by the CI docker driver ("Attestation is not supported for the docker driver... turn on the containerd image store") — GH Releases get no binaries; v0.7.0/v0.7.1/v0.8.0 all failed identically | Medium | 1h     | `gh run view 34463730693` (v0.8.0), 34446767394 (v0.7.1)                         |
| Markdown Lint red on master: AGENTS.md MD029 ordered-list prefix style (30., 31., ... expected 1., 2., ...)                                                                                                                                                                        | Low    | 30min  | `gh run view 34463729656` (v0.8.0 push); failing since 2026-09-09                |
