# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-25

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`;
long-term ideas live in `ROADMAP.md`. **This file contains OPEN work only** —
when a task ships, remove it here and record it in `CHANGELOG.md`.

---

## High Priority

| Task                                                    | Impact                                          | Effort | Evidence                                                                                  |
| ------------------------------------------------------- | ----------------------------------------------- | ------ | ----------------------------------------------------------------------------------------- |
| Cut a release (version bump + git tag)                  | High — ~30 unreleased commits since `v0.5.0`    | 30min  | `git tag` shows `v0.5.0` as latest; friction-reduction + quality-debt work is untagged    |
| Resolve the `format` preset formatter split-brain       | High — config inconsistency, user confusion     | 1h     | `pkg/constants/presets.go`: `format` = 3 formatters, `CoreFormatters` & `house` = 4       |
| Remove `EnableGolinesFormatter` dead code               | Medium — unreachable logic after CoreFormatters | 30min  | `pkg/linter/fixer_formatters.go:51`; bypassed by `EnableCoreFormatters` (`fixer.go:260`)  |
| Remove stale `G104` from the repo's own `.golangci.yml` | Medium — split-brain vs current defaults        | 15min  | `.golangci.yml:179` still lists `G104`; defaults now ship `G304, G115` only (idempotency) |

### Notes on the High-priority items

- **Release:** the friction-reduction Pareto plan (C19) was the only rollout
  task not executed. Decide patch (`v0.5.1`) vs minor (`v0.6.0`); additive
  defaults that change output argue for minor, but pre-1.0 semver is looser.
- **`format` split-brain:** `CoreFormatters` gained `golines` (now 4), the
  `house` preset has 4, but `PresetFormatters["format"]` still has 3. Either add
  `golines` to `format` or document it as the deliberate "minimal formatter"
  preset. Product decision.
- **Dead code:** two options — (a) remove `EnableGolinesFormatter` + its call
  site and accept golines as unconditional (matches the validated 128/160
  stack); (b) remove `golines` from `CoreFormatters` and keep the conditional
  recommendation path. Option (a) is recommended.
- **Stale `G104`:** the tool's idempotency guarantee means re-running
  `configure` won't remove it. Manually delete the `gosec:` settings block from
  `.golangci.yml`, then run the tool to re-inject the correct defaults.

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
