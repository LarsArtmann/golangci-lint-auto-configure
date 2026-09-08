# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-30

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`;
long-term ideas live in `ROADMAP.md`. **This file contains OPEN work only** —
when a task ships, remove it here and record it in `CHANGELOG.md`.

---

## High Priority

_All high-priority items resolved in v0.6.0 — see CHANGELOG.md._

## Medium Priority

| Task                                                       | Impact                                                  | Effort | Evidence                                                                   |
| ---------------------------------------------------------- | ------------------------------------------------------- | ------ | -------------------------------------------------------------------------- |
| Consolidate `ARCHITECTURE.md` inline ADRs into `docs/adr/` | Medium — ADRs live in two places (split-brain)          | 1–2h   | `docs/ARCHITECTURE.md` has 8 inline ADRs; `docs/adr/` has 6 separate files |
| Full `README.md` claim-by-claim audit (~500 lines)         | Medium — repeated spot-checks, never line-by-line       | 2h     | Only Requirements / example output / CI / Related Projects verified        |

## Low Priority

| Task                                                            | Impact | Effort | Evidence                                                                         |
| --------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------------------------------- |
| Extend docs-integrity test to ALL hardcoded FEATURES.md counts  | Low    | 1h     | `pkg/constants/docs_integrity_test.go` only covers preset counts                 |
| Status report lifecycle policy (archive cadence)                | Low    | 30min  | `docs/status/README.md` index exists; no archive cadence policy                  |
| Multi-preset merge correctness tests (dedup, formatter union)   | Low    | 1h     | `--preset a --preset b` shipped without dedicated merge tests                    |
| Swallowed-error governance audit (erraudit + periodic re-check) | Low    | 1h     | 194 findings reviewed 2026-07-30; 141 `context_loss` are noise; re-run quarterly |
