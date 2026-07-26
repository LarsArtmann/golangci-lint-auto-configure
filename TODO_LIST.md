# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-27

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`;
long-term ideas live in `ROADMAP.md`. **This file contains OPEN work only** —
when a task ships, remove it here and record it in `CHANGELOG.md`.

---

## High Priority

_All high-priority items resolved in v0.6.0 — see CHANGELOG.md._

## Medium Priority

| Task                                                                | Impact                                                  | Effort | Evidence                                                                                      |
| ------------------------------------------------------------------- | ------------------------------------------------------- | ------ | --------------------------------------------------------------------------------------------- |
| `RuleKey()` merge strategy for default-exclusion propagation        | Medium — new linters don't reach existing configs       | 2–3h   | `RuleKey()` dedup key is `Path\|Text\|Source`; 88 machine-generated configs stuck on old list |
| YAML indentation preservation in config output                      | Medium — massive whitespace diffs obscure changes       | 3–4h   | Tool reformats 2-space→4-space aggressively                                                   |
| `--force-settings` flag (re-inject defaults over existing settings) | Medium — solves the idempotency trap for self-config    | 2h     | No way to refresh stale settings in an existing config today                                  |
| Consolidate `ARCHITECTURE.md` inline ADRs into `docs/adr/`          | Medium — ADRs live in two places (split-brain)          | 1–2h   | `docs/ARCHITECTURE.md` has 8 inline ADRs; `docs/adr/` has 6 separate files                    |
| Extract a `CommandContext` struct for CLI globals                   | Medium — 9 package-level vars hinder testing            | 2–3h   | `internal/cli/commands.go`: `priority`, `dryRun`, `verbose`, `quiet`, …                       |
| Full `README.md` claim-by-claim audit (~500 lines)                  | Medium — repeated spot-checks, never line-by-line       | 2h     | Only Requirements / example output / CI / Related Projects verified                           |
| Run the full `nix flake check` (with build) at least once           | Medium — hermetic build path unvalidated for 4 sessions | 15min  | Only `--no-build` run recently; needs SSH for private flake inputs                            |

## Low Priority

| Task                                                           | Impact | Effort | Evidence                                                         |
| -------------------------------------------------------------- | ------ | ------ | ---------------------------------------------------------------- |
| Extend docs-integrity test to ALL hardcoded FEATURES.md counts | Low    | 1h     | `pkg/constants/docs_integrity_test.go` only covers preset counts |
| Status report lifecycle policy (archive cadence)               | Low    | 30min  | `docs/status/README.md` index exists; no archive cadence policy  |
| Multi-preset merge correctness tests (dedup, formatter union)  | Low    | 1h     | `--preset a --preset b` shipped without dedicated merge tests    |
| Swallowed-error governance audit (deeper than the 2-site pass) | Low    | 1h     | Prior audit found only 2 benign `defer Close()` sites            |
| `shortRunID` panic guard (`parts[2][:4]` without length check) | Low    | 15min  | `internal/cli/cmd_audit.go`; accepts arbitrary input             |
