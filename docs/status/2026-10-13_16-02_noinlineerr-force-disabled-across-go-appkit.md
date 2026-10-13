# Status Report: noinlineerr force-disabled across go-appkit

| Field            | Value                                                                                    |
| ---------------- | ---------------------------------------------------------------------------------------- |
| Date             | 2026-10-13 16:02 CEST                                                                    |
| Session CWD      | `/home/lars/projects/golangci-lint-auto-configure`                                       |
| Repos touched    | `go-appkit` (11 files edited), `golangci-lint-auto-configure` (analysis only)            |
| Trigger          | User pasted a `noinlineerr` finding, then demanded: "I want noinlineerr FORCE disabled!" |
| Commit state     | go-appkit changes **uncommitted** at report time (auto-commit daemon may pick them up)   |
| Companion review | Brutal self-review answers folded into section (e) — single report per user instruction  |

## Session Summary

1. User pasted `noinlineerr` finding at `example/main.go:116` and asked why the linter exists.
2. I explained the rationale (ldez's style opinion vs. Effective Go's inline form) and noted the tool already lists `noinlineerr` in `constants.DisabledLinters`.
3. User demanded force-disable. I traced the finding to its real source project, then edited every affected config.

**Key discovery chain:**

- `noinlineerr` is tool-level force-disabled since commit `0c6c4c7` (`pkg/constants/rules.go:138-141`): moved to `linters.disable` on every configure run (`pkg/linter/fixer_config.go:430-439`), exempt from anti-gaming re-enable enforcement (`pkg/linter/fixer_enforce.go:108`), and stripped even from manually enabled configs since `linterSet` is seeded from the user's enable list (`pkg/linter/fixer.go:263`).
- The pasted finding came from `/home/lars/projects/go-appkit/flightrecorder/example/main.go:116`. **The code was already fixed** (plain `err := rec.Start()` assignment) — the pasted lint output was stale; only the linter remained enabled in config.
- go-appkit survey: root `.golangci.yml` already had `noinlineerr` in `disable:`; **11 module configs** still had it in `linters.enable`; `flightrecorderhealth/.golangci.yml` documented it as a deliberate divergence in its header comment.
- **Split brain found in this tool (not fixed, reported only):** the config-merger path (`pkg/config/merger_linters.go`, `mergeEnableDisable`) has zero knowledge of `DisabledLinters` — zero grep matches in `pkg/config`. The `merge` subcommand (`internal/cli/commands.go:98`) can union a tool-disabled linter back into an enable list.

## Changes Made (go-appkit)

- 11 module configs (`flightrecorder`, `otel`, `cqrs`, `realtime`, `errorpages`, `flightrecorderhealth`, `integration`, `health`, `docs`, `security`, `systemd`): removed `noinlineerr` from `linters.enable`, added `disable: [noinlineerr]` before the first `settings:` key — mirroring the root config's layout and exactly what the fixer's DisabledLinters pass would produce.
- `flightrecorderhealth/.golangci.yml`: dropped the now-false "noinlineerr additionally enabled" divergence bullet from the header comment.
- Root `.golangci.yml`: untouched (already correct).
- Verification: `golangci-lint config verify` passes on all 12 configs. Final diff: 11 files changed, +22/−12.

---

## a) FULLY DONE

| Item                                                                     | Evidence                                                                                                                               |
| ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| Located the true source of the pasted finding                            | `grep` for `rec.Start()` → `go-appkit/flightrecorder/example/main.go:116`, `go-appkit/otel/example/main.go:119`                        |
| Confirmed the code itself needs no change                                | `flightrecorder/example/main.go:116` already uses plain assignment; the finding was stale lint output                                  |
| Confirmed the tool's force-disable mechanism exists and is correct       | `pkg/constants/rules.go:138-141`, `pkg/linter/fixer_config.go:430-439`, `fixer_enforce.go:108`, `fixer.go:263` (all read this session) |
| Force-disabled `noinlineerr` in all 11 go-appkit module configs          | `git diff --stat`: 11 files, +22/−12; every config now has exactly one `noinlineerr` mention, under `disable:`                         |
| Updated stale divergence comment in `flightrecorderhealth/.golangci.yml` | Header bullet removed; remaining "no gci formatter" divergence left intact                                                             |
| Validated all 12 go-appkit configs still parse and verify                | `golangci-lint config verify`: OK for root + 11 modules                                                                                |
| Explained the linter's rationale honestly (turn 1)                       | Style opinion by ldez; Effective Go itself uses the inline form; tool treats it as noise                                               |

## b) PARTIALLY DONE

| Item                                              | What works                                                               | What remains                                                                                                                                               | Effort |
| ------------------------------------------------- | ------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| "noinlineerr FORCE disabled" (the user's demand)  | Done for go-appkit: 11 configs moved to `disable:`, root already correct | Not done for the other ~160 sibling projects; not done for this tool's own repo (see c); the `merge` command can still resurrect it (split brain, unfixed) | M      |
| Verification that the finding is gone             | `config verify` green on all 12 configs                                  | Never re-ran `golangci-lint run` on `flightrecorder`/`otel` to empirically prove zero findings post-change                                                 | S      |
| Understanding why the user saw the finding at all | Root cause half-answered: code was already fixed, so output was stale    | Did not identify WHICH lint loop/cache served the stale result to the user                                                                                 | S      |
| Session memory capture                            | This report                                                              | go-appkit's `AGENTS.md`/`TODO_LIST.md` not updated with the new repo-wide invariant; `TODO_LIST.md` of this repo not updated with the merger split brain   | S      |

## c) NOT STARTED

| Item                                                                                     | Why not started                                                                                                            | Still wanted?                |
| ---------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| Ecosystem sweep: find `noinlineerr` in `linters.enable` across all ~160 sibling projects | Out of scope per user instruction ("do not research other stuff"); scope of the "FORCE" mandate is unknown → question (g)1 | Yes — likely the real intent |
| Check WHEN commit `0c6c4c7` landed                                                       | Not needed for the immediate edit; dates the exposure window (configs created before it may still enable the linter)       | Yes, cheap                   |
| Fix the merger split brain (`mergeEnableDisable` ignores `DisabledLinters`)              | Code change to this tool needs a scope decision → question (g)2                                                            | Yes                          |
| Check this tool's own `.golangci.yml` hygiene for noinlineerr                            | Discovered DURING report writing (see d); cleanup not started                                                              | Yes                          |
| Update go-appkit docs (`AGENTS.md`, `TODO_LIST.md`, `FEATURES.md`) for the new invariant | Report-first session; doc updates not yet sanctioned                                                                       | Yes                          |
| Commit the go-appkit changes                                                             | Harness rule: never commit without explicit user request; auto-commit daemon handles it                                    | Daemon-dependent             |
| HARVEST section (f) into `TODO_LIST.md`/`ROADMAP.md`                                     | User said report then wait                                                                                                 | Yes, next session            |

## d) TOTALLY FUCKED UP

**Nothing destroyed, nothing broken.** All 12 configs verify; the diff is minimal and reversible. Radical honesty about the near-misses and wrong calls:

1. **I overrode a documented, deliberate decision without asking.** `flightrecorderhealth/.golangci.yml` explicitly said "noinlineerr additionally enabled (newer linter, catches inline err decls)" — someone chose that on purpose. I removed it because your blanket demand postdates that choice, but I never confirmed the supersession. Severity: low (trivially revertible), but it is the one place I acted against a written intent. → question (g)3.
2. **Removing the linter removes real future value.** `noinlineerr` was catching something in go-appkit (at least historically). Force-disabling it means future inline-err regressions in those 11 modules go unflagged, with no compensating control. That is the correct execution of your instruction, but it should be said plainly: you traded coverage for consistency.
3. **Stale lint output reached you as if current.** The finding you pasted no longer exists in the code — some lint loop, cache, or CI artifact served an old result. Untreated; root cause unknown (which loop? which runner?).
4. **One wasted round trip.** I fired a batched edit+bash call without viewing `flightrecorderhealth/.golangci.yml` first; the edit tool correctly refused ("you must read the file before editing"). Trivial, but it violates my own read-before-edit rule.
5. **Turn-1 imprecision (not a lie, unverified claim).** I said the finding "most likely came from a manually-configured `linters.enable`". Turn 2 proved the enable part right but the _manual_ part is unproven — the config may have been written by an older tool version before `0c6c4c7` added the force-disable. Corrected here.

## e) WHAT WE SHOULD IMPROVE

1. **Split brain in this tool: fixer path vs merger path.** `updateConfigFromSets` enforces `DisabledLinters`; `mergeEnableDisable` (`pkg/config/merger_linters.go`) does not. Same invariant, two code paths, one drifted. This session's demand ("FORCE disabled") is only as strong as the weakest path. Fix: apply the DisabledLinters filter in `mergeLintersConfig` + BDD spec.
2. **No detection surface for stale configs.** If a config still enables a tool-disabled linter, nothing tells the user until they happen to run `configure`. `analyze` should report "linter X is in `enable` but tool-disabled" as a finding, so the 0c6c4c7-style decisions propagate visibly instead of silently.
3. **This tool doesn't fully eat its own dogfood.** Own `.golangci.yml`: `noinlineerr` is not enabled (good) but also not in `disable:` (unlike what the tool would produce downstream), and **7 inert exclusion-rule mentions remain** (`pkg/audit/`, `loader.go`, `analyzer.go`, `fixer.go`, `version_checker.go`) — dead config left over from before the disable decision. Exclusion pruning for unenabled linters doesn't exist (settings are pruned, exclusion rules are not).
4. **Cross-repo config drift is fought by hand.** 12 near-identical `.golangci.yml` files in go-appkit; one linter decision required 11 manual edits + 1 comment fix. The tool exists to do exactly this (`configure` per module), but it wasn't run here — I hand-edited to keep the diff surgical. Tradeoff recorded; a "sweep siblings" mode would make the ecosystem-wide decision cheap.
5. **Deliberate divergences need an expiry or registry.** flightrecorderhealth's divergence comment went stale the moment the user's policy changed. Divergences documented only in comments drift silently; they belong in the sidecar policy file or a checked list.
6. **Verification depth.** I verified config _validity_, not config _effect_. One `golangci-lint run` on the two modules with known findings would have closed the loop empirically.
7. **Search order.** Finding the source project: I `sed`-inspected 5 candidate `example/main.go` files before grepping `rec.Start()`. Grep-for-the-unique-string first would have been one step.
8. **Memory capture discipline.** The go-appkit invariant ("noinlineerr force-disabled repo-wide, decision 2026-10-13") lives only in this report so far. Project AGENTS.md is the right home; not yet written.

## f) Things We Should Get Done Next

Ranked by impact. Impact: Critical/High/Medium/Low. Effort: S <30min, M 30min–2h, L >2h.

**Closeout of this session's work**

| # | Task                                                                                                                            | Impact | Effort | Category      |
| - | ------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 1 | Run `golangci-lint run ./...` in `go-appkit/flightrecorder` and `otel` — prove zero noinlineerr findings post-change            | High   | S      | Quality       |
| 2 | Confirm go-appkit changes are committed (auto-commit daemon) and the diff is exactly the 11 config files                        | High   | S      | Cleanup       |
| 3 | Answer question (g)1 and, if approved, sweep all ~160 sibling projects for `noinlineerr` under `linters.enable`; fix every hit  | High   | M      | Cleanup       |
| 4 | Get the date of commit `0c6c4c7` (when noinlineerr became tool-disabled) to size how many sibling configs predate the guarantee | Medium | S      | Documentation |
| 5 | Grep go-appkit's `AGENTS.md`/`TODO_LIST.md`/`FEATURES.md`/`docs/` for noinlineerr mentions; fix stale text                      | Medium | S      | Documentation |
| 6 | Record the invariant in go-appkit's `AGENTS.md`: noinlineerr force-disabled repo-wide (decision 2026-10-13)                     | Medium | S      | Documentation |

**Tool gaps found this session (golangci-lint-auto-configure)**

| #  | Task                                                                                                                                                                        | Impact | Effort | Category      |
| -- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 7  | Enforce `DisabledLinters` in the merger path (`mergeLintersConfig`/`mergeEnableDisable`) so `merge` cannot resurrect tool-disabled linters; add BDD spec                    | High   | M      | Bug           |
| 8  | Add an `analyze` finding: "linter in `enable` is tool-disabled (`DisabledLinters`)" so stale configs surface without a configure run                                        | High   | M      | Feature       |
| 9  | After item 3's sweep: run this tool's `configure` across affected sibling modules instead of hand-editing (dogfood + churn measurement)                                     | High   | M      | Quality       |
| 10 | Verify `noinlineerr` complies with the gotcha-#10 invariants (no `LinterPriorities`/`LinterReasons` entries) — data-integrity tests should already pin this; confirm        | Medium | S      | Quality       |
| 11 | Regression spec: config with `noinlineerr` in enable → `configure` → assert moved to `disable` (may already exist in fixer tests; verify, add if missing)                   | Medium | M      | Quality       |
| 12 | Surface a user-visible warning (not just `Debugf`) when a DisabledLinters force-move fires (`fixer_config.go:433`)                                                          | Low    | S      | Feature       |
| 13 | Clean this repo's own config: move `noinlineerr` to `linters.disable` next to `depguard`, delete the 7 inert exclusion-rule mentions                                        | Medium | S      | Cleanup       |
| 14 | Consider pruning exclusion-rule entries whose linters are not enabled (settings already have `pruneUnenabledLinterSettings`; exclusion rules have nothing)                  | Medium | M      | Feature       |
| 15 | Document the user-facing guarantee in README/FEATURES: "DisabledLinters are moved to `linters.disable` on every configure run, in every code path" — true only after item 7 | Low    | S      | Documentation |
| 16 | Record the merger split brain in this repo's `TODO_LIST.md` (HARVEST input so it survives this report)                                                                      | High   | S      | Documentation |

**go-appkit follow-ups**

| #  | Task                                                                                                                                                                            | Impact | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ | ------------- |
| 17 | Decide whether remaining inline error handling in go-appkit deserves manual cleanup now that noinlineerr no longer flags regressions                                            | Medium | M      | Quality       |
| 18 | Add a CI/data guard in go-appkit: no `DisabledLinters`-class linter (at least noinlineerr) under any `enable:` — blocks copy-paste re-introduction across the 12 configs        | Medium | S      | Quality       |
| 19 | Note the force-disable decision in go-appkit's `TODO_LIST.md`/CHANGELOG                                                                                                         | Low    | S      | Documentation |
| 20 | Re-check flightrecorderhealth's remaining documented divergence ("no gci formatter") is still current                                                                           | Low    | S      | Documentation |
| 21 | Sandbox test: run `configure` on a COPY of one go-appkit module config; confirm it produces the same noinlineerr outcome as the hand edit (product parity check, no repo churn) | Medium | S      | Quality       |

**Process / observability**

| #  | Task                                                                                                                                                         | Impact | Effort | Category      |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ | ------ | ------------- |
| 22 | Identify which lint loop/cache served the stale finding to the user (finding referenced code fixed before the session); fix or document the staleness window | Medium | S      | Bug           |
| 23 | Standing rule for pasted linter findings: verify the cited file state first, explain rationale second (this session had it inverted)                         | Low    | S      | Process       |
| 24 | When editing multiple files in one project, batch read all targets before any edit call (prevents read-before-edit round-trip failures)                      | Low    | S      | Process       |
| 25 | Document the dogfood-vs-surgical-edit decision rule (when to run the tool on downstream repos vs hand-edit) in this repo's AGENTS.md                         | Low    | S      | Documentation |
| 26 | Run HARVEST next session: items 3, 7, 8, 16 into `TODO_LIST.md`; items 9, 14, 21 as ROADMAP fuel                                                             | High   | S      | Documentation |

I stop at 26 deliberately: every item above is derived from something actually observed this session. Filling out to 50 would mean inventing scope you told me not to research (items 3 and 9 will spawn the real follow-up list once the sweep runs).

## g) Questions I Cannot Figure Out Myself

1. **Scope of "FORCE disabled":** go-appkit only, or all ~160 sibling projects? I found the mechanism is per-configure-run; any sibling config written before commit `0c6c4c7` still enables noinlineerr and will keep flagging it. I can sweep, but that touches many repos you did not name — your mandate decides.
2. **Should I fix the merger split brain now?** The `merge` subcommand can re-introduce tool-disabled linters (e.g. noinlineerr) into an enable list, silently undoing today's work for any project that ever merges configs. Fixing means a code change in this tool (filter in `mergeLintersConfig` + BDD spec). Priority is yours — I don't know whether `merge` is in active use anywhere.
3. **flightrecorderhealth's divergence was deliberate** (documented: "noinlineerr additionally enabled, catches inline err decls"). Did your blanket force-disable supersede it, or do you want that one module to keep the linter and I revert its config?

---

_A point-in-time snapshot. Section (f) is the HARVEST input — it dies here if not pulled into `TODO_LIST.md` next session. Report committed by the auto-commit daemon (harness forbids manual commits without explicit request)._
