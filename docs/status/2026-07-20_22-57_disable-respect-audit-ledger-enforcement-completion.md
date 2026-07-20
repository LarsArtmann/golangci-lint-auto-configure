# Status Report: Disable-Respect, Audit Ledger & Anti-Gaming Enforcement

> **Date:** 2026-07-20 23:00
> **Session goal:** Fix broken build, complete Pillars B+C, ship the three-pillar design
> **Verdict:** Build fixed, core features shipped, **critical test gaps remain**

---

## a) FULLY DONE (shipped, green, lint-clean)

### Build Fix (was the #1 blocker)

- Master was broken: commit `193b6b1` used `--no-verify`, leaving `newRunLedger` undefined
- Defined `newRunLedger` in `internal/cli/cmd_audit.go` with full RunContext construction
- `go build ./...` passes

### SA4023 Typed-Nil Interface Bugs (pre-existing, fixed)

- `RunFmtCommand` (`pkg/linter/command_runner.go:86`) **always returned non-nil error** — every `configure` run logged "golangci-lint fmt failed" even on success
- `validateConfig` (`internal/cli/cmd_validate.go:83`) **never returned nil** — the `validate` command always reported failure, the success path was unreachable
- Root cause: `WrapClassified(nil)` returns nil `*errorfamily.Error`, which boxes into a non-nil `error` interface
- Fixed with explicit `if err == nil { return nil }` guards before `WrapClassified`
- Documented the pitfall in `pkg/errors/errors.go` (comment on `WrapClassified`) and AGENTS.md Critical Gotcha #14

### Pillar B: Audit Ledger Package (`pkg/audit/`)

- Append-only JSONL ledger with `Recorder` interface + `NoopRecorder` fallback
- `NewRunID`, `RepoHashOf`, `ReadAll` with malformed-line skipping — all tested (prior session)
- **NEW this session:** `PurgeOlder`, `Clear`, `PurgeRetention`, `DefaultRetention` (90-day retention)
- **NEW tests:** `PurgeOlder` (3 specs), `Clear` (2 specs), `DefaultRetention` (1 spec) — all pass

### Pillar B: Audit CLI Subcommand (`internal/cli/cmd_audit.go`)

- `golangci-lint-auto-configure audit` with `--json`, `--since`, `--linter`, `--clear` flags
- Duration parsing supports both Go duration (`24h`) and day shorthand (`7d`)
- Table and JSON output formats
- Registered in `internal/cli/commands.go`

### Pillar B: `--no-audit` Flag

- `--no-audit` flag on `configure` command + `GOLANGCI_LINT_AUTO_CONFIGURE_NO_AUDIT` env var
- `noAudit` var moved to the shared `commands.go` var block (avoids `gochecknoglobals`)

### Pillar B: Broadened Recording

- `recordConfigChanges` now captures **formatter** enable/disable changes (in addition to linter enable/disable/settings)
- `configSnapshot` struct renamed from `linterSnapshot`, added `formatterEnable` and `formatterDisable` fields
- New audit actions: `ActionFormatterAddedToEnable`, `ActionFormatterRemovedFromEnable`

### Pillar C: Policy Package (`pkg/policy/`)

- Sidecar file `.golangci-lint-auto-configure.yml` with categorized disable justifications
- `Load`, `IsJustified`, `Justification` — all tested (7 specs, all pass)
- Categories: `false-positives`, `superseded`, `convention`, `performance`, `other`

### Pillar C: Re-Enable Enforcement (`pkg/linter/fixer_enforce.go`)

- When sidecar exists, re-enables linters in `linters.disable` without justification
- Tool-level disabled linters (`constants.DisabledLinters`) are always exempt
- Records `ActionReEnabled` in the audit ledger
- Runs only on non-dry-run successful saves

### Documentation

- **AGENTS.md:** Added Tech Stack entries for audit ledger + policy enforcement; Critical Gotchas #14 (typed-nil pitfall), #15 (sidecar enforcement), #16 (ledger best-effort)
- **README.md:** Added "Audit Trail" and "Disable-Reason Enforcement" sections with examples; added `audit` command and `--no-audit` flag to reference tables
- **Feedback doc** moved from `docs/feedback/new/` to `docs/feedback/resolved/`

### Lint Cleanup

- Fixed pre-existing gosec false positives in `pkg/config/loader.go` and `internal/cli/cmd_validate.go`
- Fixed pre-existing wsl_v5 issues in `pkg/config/merger_helpers.go`
- **Full project lint: 0 issues** (`golangci-lint run ./...`)

---

## b) PARTIALLY DONE

### `recordConfigChanges` coverage — PARTIAL

Records linter enable/disable, formatter enable/disable, and settings pruning. **Still misses:**

- Go version changes (`run.go`)
- Runner settings (`runners`)
- Build tags
- Output formats
- Issues settings (`max-issues-per-linter`, `max-same-issues`)
- Exclusion rules

These all go through `configChangeRecorder` in `applyAndSave` but are not snapshotted or diffed.

### Audit test coverage — PARTIAL

- `pkg/audit/ledger.go` core functions: **well tested** (prior session + this session)
- `pkg/policy/policy.go`: **well tested** (7 specs)
- **`internal/cli/cmd_audit.go`: ZERO tests** — `parseSinceDuration`, `filterAuditEntries`, `entryMatchesFilters`, `shortRunID`, `outputEntries`, `clearAuditLedger`, `auditDisabled`, `outputAuditJSON`, `outputAuditTable` are all untested
- **`pkg/linter/fixer_enforce.go`: ZERO tests** — `enforceDisableReasons`, `tryReEnableLinter`, `loadPolicy`, `isToolLevelDisabled` are all untested
- **`newRunLedger`: ZERO tests** — the function that was the original blocker has no test

### Ledger retention integration — PARTIAL

`PurgeRetention` is called in `newRunLedger`, but there's no test verifying the purge actually runs during a configure invocation (only unit tests on `PurgeOlder` directly).

---

## c) NOT STARTED

### Runtime Cost Analysis (Pillar C core feature)

The `Entry.FindingsHidden` field exists but is **always 0**. The plan was to run each disabled linter and count how many findings it would report — making the cost of disabling visible in the audit ledger. This is the quantitative backbone of the anti-gaming story ("you disabled `mnd` and hid 195 findings"). **Not implemented at all.**

### go-finding Export (Pillar C)

The summary planned `--output`/`--format` integration to emit findings as SARIF/JSON via `go-finding`. Not started.

### Example Sidecar File

No example `.golangci-lint-auto-configure.yml` in the `examples/` directory. Users have no template to copy.

### FEATURES.md Update

`FEATURES.md` has one mention of "audit" but no entries for the audit ledger subcommand, policy enforcement, or `--no-audit` flag. These are new features that belong in the feature inventory.

### TODO_LIST.md Update

No TODO entries for the remaining work (runtime cost analysis, broader recording, concurrency safety, etc.).

### DOMAIN_LANGUAGE.md Update

New domain terms (audit ledger, disable-reason sidecar, enforcement, re-enable, unjustified disable) are not in the domain glossary.

### Feedback Doc Resolution Note

Moved the feedback doc to `resolved/` but did not add a resolution annotation explaining what was implemented and what remains.

### Concurrency Safety

JSONL ledger has no file locking. Two parallel configure runs (e.g., BuildFlow running multiple repos) could interleave writes or one could purge while another appends. BuildFlow uses SQLite WAL for this; JSONL has no equivalent.

---

## d) TOTALLY FUCKED UP

### Nothing is "totally fucked up" — but these are serious gaps:

1. **Enforcement code has zero tests.** `enforceDisableReasons` is the most security-critical code in this feature (it decides what gets re-enabled). It has no test. An agent could disable a linter, the tool re-enables it, and nobody tested that the exempt list works correctly.

2. **CLI audit subcommand has zero tests.** `parseSinceDuration` handles user input with day shorthand — a parsing bug would silently filter incorrectly. Untested.

3. **`FindingsHidden` is a dead field.** It's in the JSON schema, in every entry, always zero. This is a lie — the schema promises data the tool never populates. Either implement it or remove it.

4. **I didn't run `nix build` or `nix flake check`.** AGENTS.md says `nix build` is the preferred build method. I only used `go build`. The Nix build may fail on `vendorHash` or other Nix-specific issues.

---

## e) WHAT WE SHOULD IMPROVE

1. **Test the enforcement logic.** This is the #1 priority. `enforceDisableReasons` needs specs for: no sidecar (all disables respected), sidecar with justification (respected), sidecar without justification (re-enabled), tool-level linter exempt, multiple linters mixed, dry-run does not enforce.

2. **Test the CLI audit command.** `parseSinceDuration` needs edge cases (empty, `7d`, `24h`, invalid, negative days). `filterAuditEntries` needs linter filter + time filter combinations.

3. **Remove or implement `FindingsHidden`.** A dead field in the schema is worse than no field.

4. **Run `nix build` and `nix flake check`.** These are the project's real build gates.

5. **Add file locking to the ledger.** `flock` on the file before append/purge. Prevents interleaved writes.

6. **Update FEATURES.md, TODO_LIST.md, DOMAIN_LANGUAGE.md.** These are the living docs that track project state.

7. **Add a resolution note to the feedback doc.** It was moved to `resolved/` but doesn't say what was resolved or what remains.

---

## f) Up to 50 Things to Get Done Next

### Critical (test gaps)

1. Write BDD specs for `enforceDisableReasons` — no sidecar case
2. Write BDD specs for `enforceDisableReasons` — justified disable respected
3. Write BDD specs for `enforceDisableReasons` — unjustified disable re-enabled
4. Write BDD specs for `enforceDisableReasons` — tool-level linter exempt
5. Write BDD specs for `enforceDisableReasons` — dry-run does not enforce
6. Write BDD specs for `tryReEnableLinter` — recording ActionReEnabled in ledger
7. Write BDD specs for `loadPolicy` — missing file, malformed YAML, valid file
8. Write unit tests for `parseSinceDuration` — empty, days, hours, invalid
9. Write unit tests for `filterAuditEntries` — linter filter, time filter, combined
10. Write unit tests for `shortRunID` — normal, short, edge cases
11. Write unit tests for `auditDisabled` — flag, env var, both, neither
12. Write integration test: configure run → ledger has entries → audit subcommand reads them
13. Write test: `newRunLedger` creates a working ledger when path is available
14. Write test: `newRunLedger` returns NoopRecorder when `--no-audit`

### Pillar C completion

15. Implement runtime cost analysis — run each disabled linter, count findings
16. Populate `FindingsHidden` in audit entries with real counts
17. Add `--show-cost` flag to `configure` (or run always in `analyze`)
18. Integrate go-finding for SARIF/JSON export of hidden findings
19. Add `--output`/`--format` flags to emit hidden-finding report

### Broader audit recording

20. Snapshot and diff Go version changes
21. Snapshot and diff runner settings
22. Snapshot and diff build tags
23. Snapshot and diff output formats
24. Snapshot and diff issues settings (`max-issues-per-linter`, `max-same-issues`)
25. Snapshot and diff exclusion rules

### Robustness

26. Add file locking (`flock`) to ledger append
27. Add file locking to `PurgeOlder` rewrite
28. Consider SQLite backend for concurrent-write safety (BuildFlow pattern)
29. Add ledger size cap or rotation (beyond 90-day purge)
30. Add `--ledger-path` flag to override default location

### CLI polish

31. Add `audit --summary` mode (counts by action, by linter, by repo)
32. Add `audit --repo <path>` filter by repository
33. Add `audit --action <type>` filter by action type
34. Add `audit --follow` / `--watch` for tailing new entries
35. Add exit codes to `audit` (e.g., exit 1 if unjustified disables found)

### Documentation

36. Add example `.golangci-lint-auto-configure.yml` to `examples/`
37. Update FEATURES.md with audit ledger + enforcement features
38. Update TODO_LIST.md with remaining Pillar C work
39. Update DOMAIN_LANGUAGE.md with new terms
40. Add resolution note to the moved feedback doc
41. Update `docs/references/working-with-codebase.md` with audit/policy sections
42. Update prior status report (`docs/status/2026-07-20_12-28_*`) with cross-reference

### Build verification

43. Run `nix build` and fix any issues
44. Run `nix flake check` and fix any issues
45. Update `vendorHash` in `flake.nix` if go.mod changed

### Policy features

46. Add `policy init` subcommand to scaffold a sidecar from current disables
47. Add `policy validate` subcommand to check sidecar syntax
48. Add policy linting (category must match a known enum, reason must be non-empty)
49. Add sidecar schema validation against actual disabled linters (warn on stale entries)
50. Add `--strict-enforcement` flag to make re-enable a hard error (exit 1) instead of silent fix

---

## g) Questions for the User

### Q1: Runtime cost analysis — when should it run?

Running each disabled linter to count hidden findings takes minutes on large repos (159 repos, 610 Go files). Should it:

- **(a)** Run on every `configure` (slow but always visible)
- **(b)** Run only on `analyze` (explicit, opt-in)
- **(c)** Run behind a `--show-cost` flag (opt-in on any command)

I cannot decide this because it trades coverage against speed, and the right answer depends on how the tool is invoked in your BuildFlow pipeline (pre-commit hook = must be fast; manual `analyze` = can be slow).

### Q2: Sidecar committed or gitignored?

Should `.golangci-lint-auto-configure.yml` be:

- **(a)**Committed to git (team-shared, reviewable, agent can't silently remove it without a visible diff)
- **(b)** Gitignored (per-developer, personal preferences)

I chose "committed" in the README, but this is a product decision. If committed, agents must add entries to justify disables (visible in review). If gitignored, enforcement is per-developer and provides no team-level anti-gaming protection.

### Q3: `FindingsHidden` — implement now or remove the field?

The `FindingsHidden` field is in the audit `Entry` struct but always 0. Should I:

- **(a)** Implement runtime cost analysis now to populate it (requires Q1 answer, takes significant work)
- **(b)** Remove the field until runtime cost analysis is built (cleaner schema, no dead fields)
- **(c)** Leave it as-is (field exists, shows intent, populated later)

I lean toward (b) — a dead field is a lie in the schema — but removing it means changing the JSONL format which is already written to disk by prior runs.
