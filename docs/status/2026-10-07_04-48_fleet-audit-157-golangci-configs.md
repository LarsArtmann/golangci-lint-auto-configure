# Fleet Audit: 157 Sibling `.golangci.yml` Configs

**Date:** 2026-10-07 04:48 CEST
**Session scope:** Read all `~/projects/*/.golangci.yml` (157 of 283 projects), analyze against this tool's constants/curated defaults, verify every config against `golangci-lint v2.14.0 config verify`, and derive tool improvements. Plus the user's meta-question: what did I forget, what could be better?

**Method (repeatable, 4 steps):**

1. Glob `~/projects/*/.golangci.yml` → 157 files, parse with PyYAML → 0 errors, all `version: "2"`.
2. Structured extraction per config: `run.go`, timeout, `linters.default`, enable/disable sets, formatters, settings keys, exclusions, sidecar presence.
3. Diff against the tool's own truth: `DefaultLinterSettings`/`DefaultFormatterSettings` (pkg/constants/linter_settings.go), `ProjectSpecificLinters` gating, `DisabledLinters`, `DeprecatedLinters`, deprecation table.
4. Ground truth: `golangci-lint config verify` (v2.14.0, nix store binary) executed in each of the 157 project dirs.

This report is the "before" snapshot. The fleet sweep (already a tracked TODO row) should be followed by a re-audit to measure the delta.

---

## Headline Numbers

| Metric                                                   | Value                                                                                                   |
| -------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Projects scanned / with config                           | 283 / 157 (56 configs per 100 projects; 126 configless — Go vs non-Go split unmeasured)                 |
| Parse errors                                             | 0 (all v2 format)                                                                                       |
| `run.go` at patch level (pre-normalization)              | **148 / 157** (1.26.7×90, 1.27.1×34, 1.27.0×10, 1.26.4×7, 1.26.5×6, 1.26.8×1; only 9 clean major.minor) |
| **Fail `golangci-lint v2.14 config verify`**             | **21 / 157**                                                                                            |
| Deprecated `exhaustruct` in exclusion-rule linter lists  | 109 (86 of them already migrated to `exhaustruct_v5` — dead name left behind)                           |
| Still enabling deprecated v4 `exhaustruct`               | 7 (all stale run.go, all with v4 `exclude` settings)                                                    |
| `arangolint` enabled without arango dep in go.mod        | 123 (only 3 projects have the dep); `clickhouselint` 120 (2 have it)                                    |
| Settings blocks frozen at older curated defaults         | `varnamelen` missing current keys in 144; `mnd.ignored-files` in 136; `goconst.ignore-tests` in 22      |
| Orphaned settings (linter neither enabled nor disabled)  | 10 projects / 13 keys (incl. `funcorder` settings — funcorder is tool-disabled)                         |
| Sidecar `.golangci-lint-auto-configure.yml`              | **0 / 157**                                                                                             |
| Audit-ledger entries from real fleet runs                | **0 of 5,787** (117 distinct `repo_path`s, every one a `/tmp/ginkgo*` test dir)                         |
| Consensus enable set (≥150/157)                          | 100 linters; median enable-list size 108                                                                |
| Formatter quadruple `{gci, gofumpt, goimports, golines}` | 156/157 (137 bare, 19 +`swaggo`; tracking-papers has only gofumpt+goimports)                            |
| `goconst.min-length` bad-key remnants                    | 0 — `KnownBadSettingsKeys` self-heal proven fleet-wide                                                  |
| Tool-disabled linters in enable                          | 1 breach: typespec-sqlc has `nolineerr`→`noinlineerr` in enable with empty disable list                 |

**The 21 failing configs:** ai-router-implementation, AI-Speed-Test, ast-state-analyzer, BerryBig, code-duplicate-analyzer, Code-Quality-Agent, desire-secrets, go-appkit, go-auto-upgrade, go-wizard-sdk, nsfw-classifier, projects-management-automation, StopTube, storbi, template-air, template-github-actions, template-SECURITY, testing, universal-workflow, yt-history-intel, Zlota44.

Failure classes observed (not exhaustively root-caused — see b.1):

- v4 key under `exhaustruct_v5` (`exclude`) or bogus key (`include-unexported-declarations`) — go-appkit, go-auto-upgrade, projects-management-automation. Pre-dates the tool's rename pass; a fresh configure visit would likely repair.
- V1-era placement: `linters.settings.goimports` (nsfw-classifier) — migrator handles this; config simply unvisited.
- Hand/AI-edited schema junk: `revive.max-methods`, `revive.exported`, `err113` as a settings key (Zlota44), `mnd.ignored-numbers` as numbers not strings (Code-Quality-Agent), old ginkgolinter/usetesting/funlen/godot keys (template-air, universal-workflow, testing, template-SECURITY).

---

## a) FULLY DONE

1. **Inventory + parse:** 157/157 configs discovered, parsed, structured-extracted to JSON (0 errors, all v2).
2. **Structural analysis:** run.go / timeout / default / formatters / enable / disable / settings-key / sidecar distributions.
3. **Consensus + delta analysis:** 100-linter consensus set; every per-project deviation enumerated (adds and misses).
4. **Settings-drift diff vs current curated defaults:** 24 curated keys diffed across ~1,500 settings blocks; separated intentional per-project tuning (funlen 280/300/350, gocognit 45–70) from frozen-at-old-default blocks (varnamelen/mnd missing later-added keys).
5. **Exclusion-block analysis:** the 2-rule test-exclusion pattern's evolution traced (linter list grew 6 → 14 → +exhaustruct_v5 across config generations); bloat tail (go-dag-app 136 rules, 80 of them referencing `cmd/library-policy/...` paths that do not exist in go-dag-app — copy-paste rot from library-policy); dead names (`funcorder`, `noinlineerr`, `stylecheck`, `golines` listed as a _linter_) inside exclusion lists.
6. **Ground-truth verification:** `golangci-lint v2.14.0 config verify` executed in all 157 dirs → 136 OK / 21 FAIL, with per-config first-error capture.
7. **Gating cross-check:** `arangolint`/`clickhouselint` enables vs actual go.mod deps — proves the retraction gap (gate blocks adding, nothing removes).
8. **Deprecation scars:** 86 configs migrated to `exhaustruct_v5` but carrying dead `exhaustruct` in exclusion lists — a hard-break time bomb for when upstream removes the linter.
9. **Invariant breach found:** typespec-sqlc enables `noinlineerr` (a `DisabledLinters` member) with an empty disable list.
10. **Audit-ledger claim verified twice:** first pass had a buggy helper (see d.1); corrected pass confirms 5,787 entries / 117 repo_paths / all `/tmp/ginkgo*` / **zero real fleet runs ever**.
11. **TODO_LIST.md updated:** +5 new Medium-priority rows (exclusion cleanup on rename, curated-default provenance refresh, gated-linter retraction, orphan prune for standard-default, auto-durable never-enable sidecar) + fleet-sweep row evidence refreshed with this audit's numbers. Committed by the daemon (9b6c4d4), markdownlint green.

## b) PARTIALLY DONE

1. **Root-causing the 21 failures:** 3 exemplars deep-dived (exhaustruct_v5.exclude class, goimports-placement class, mnd-numbers class); the remaining ~18 are classified by error pattern only, not individually attributed to "tool-emitted-when-valid" vs "hand/AI edit".
2. **DB-linter overcount risk:** dep check is a go.mod substring scan (lowercased); it does not scan `.go` imports, and `hasTechnology` may use broader signals — "123 without dep" may slightly overcount. Directionally solid (3 vs 123), not exact.
3. **Audit methodology:** analysis scripts live in `/tmp/gci-analysis/` (ephemeral). The audit is documented here but not re-runnable from the repo. Persisting it is item f.41.
4. **`exhaustruct_v5` posture:** evidence gathered (116 configs + this repo enable it; policy says never-auto-enable, friction 6.5) but no recommendation made — needs Lars's call (question g.2).
5. **This report** — being written right now; harvest of its (f) list into TODO_LIST/ROADMAP follows after (c.12).

## c) NOT STARTED

1. **Fleet sweep execution** (tracked TODO row) — the single highest-impact action; fixes run.go normalization, injects missing defaults, migrates the 7 exhaustruct configs, likely repairs several of the 21.
2. **All five new tool features** recorded in TODO_LIST today (exclusion cleanup on rename; curated-default provenance refresh; gated-linter retraction; orphan prune for standard-default; auto-durable never-enable sidecar).
3. **Schema pre-flight in `configure --check`** using the committed jsonschema snapshot — would have caught all 21 failures before golangci-lint errors (untracked, new idea from this session).
4. **Dead-exclusion-rule detector** — flag exclusion rules whose `path` matches no file in the repo (go-dag-app's 80 copied rules prove demand) (untracked).
5. **Fleet/batch mode** (`configure --recursive` or a `fleet` subcommand with per-repo drift report) — the tool has every capability (configure, validate, migrate) and zero fleet execution (untracked).
6. **First sidecar anywhere** — 0/157 exist; the never-enable mechanism is unexercised in the wild.
7. **templ-components protection** — hand-removed `ireturn`/`godoclint`/`testableexamples` will be re-added on next configure visit (90-day ledger suppressions don't cover it, 0 sidecars).
8. **`.yaml` variant + nested-config completeness check** — this audit covered exactly `~/projects/*/.golangci.yml` per spec; `.golangci.yaml` and subdirectory configs unexamined.
9. **Configless-project classification** — which of the 126 are Go (sweep targets) vs non-Go (out of scope).
10. **Post-sweep re-audit** — before/after metrics; this report is the "before".
11. **Fixable-now ratio** — how many of the 21 failures today's tool repairs unattended vs needs new features (drives sweep sequencing, question g.1).
12. **Harvest of this report's (f) list** into TODO_LIST.md/ROADMAP.md (docs-health HARVEST) — deliberately deferred until Lars confirms priorities.

## d) TOTALLY FUCKED UP

Nothing damaged user data or the repo — every item below is verification/methodology sloppiness that I then closed, but it should not have shipped in the first place:

1. **Unverified generalization (the bad one).** I asserted "all 5,787 ledger entries are /tmp test dirs" after inspecting exactly ONE entry, while my counting helper looked for wrong keys (`config_path`/`path`/`file` instead of the actual `repo_path`) and bucketed everything under `'?'` ("distinct projects touched: 1"). The conclusion happened to be right, but the method was garbage; a corrected pass (117 distinct paths, all `/tmp/ginkgo*`, 0 from `~/projects`) was only run during this report's preparation. Sample-then-generalize is exactly the failure mode the verify-external-claims skill exists for — I applied it to inbound claims but not to my own intermediate results.
2. **Crashing scripts shipped to "production":** two Python heredocs died on a missing `import json` (NameError) — re-runs of code I had already written twice. Also a KeyError from a buggy `$defs` one-liner before the working `find_defs` fallback, and a shell typo (`$go-dag-app-cmd`) that executed garbage. Three avoidable error cycles.
3. **TODO_LIST edit fell back to whitespace-equivalent matching** (first row edit didn't match exactly) and I did not visually re-read the rendered table until after the daemon committed it (9b6c4d4 — now verified correct, 6 insertions/1 deletion, markdownlint green). Lucky ordering: verify-after-commit instead of verify-before-commit.
4. **Ephemeral evidence chain:** all extraction/analysis scripts are in `/tmp`, so the audit's numbers are not reproducible from the repo once /tmp clears. The report records conclusions, not the pipeline.
5. **Scope gaps accepted silently in the final summary:** the answer said "123 configs without dep" without the go.mod-only caveat, and presented patch-level run.go (148) without noting that a subset of those may have been _caused_ by running golangci-lint binaries built with patch-version toolchains (the cap logic targets this, but the 90×1.26.7 cluster strongly suggests unvisited configs, not tool writes).

## e) WHAT WE SHOULD IMPROVE

**The tool (root cause of the fleet mess: it adds, but never renews, retracts, or executes at fleet scale):**

1. **Execution gap is the #1 problem.** Every capability exists (configure, validate, migrate, --check); the fleet has never been touched (0 ledger entries). A fleet mode + one sweep fixes run.go (148), the 7 exhaustruct configs, the typespec-sqlc breach, and most schema failures in one pass.
2. **Renewal gap:** idempotency guard freezes injected settings blocks forever (144× varnamelen, 136× mnd); `--force-settings` is all-or-nothing and would nuke legitimate per-project tuning. Needs provenance (stamp tool-owned blocks; refresh only blocks still equal to an older curated default).
3. **Retraction gap:** gating (`ProjectSpecificLinters`) prevents adding arangolint/clickhouselint but nothing removes them (123/120 stale enables). Same class: `exhaustruct` rename leaves dead names in exclusion lists (86 configs).
4. **Durability gap:** ledger suppressions are per-machine and 90-day; sidecars are the durable signal and adoption is 0/157. Auto-write never-enable entries on suppression.
5. **Detection gap:** 21 configs are schema-invalid and only golangci-lint notices at lint time. The committed jsonschema + existing `validate` subcommand could pre-flight every configure/--check run and report exact bad keys with fix suggestions.

**My process (from d):**

6. Verify claims at the moment of assertion, not during report writing — especially aggregate claims ("all X are Y") derived from samples.
7. Persist audit scripts into the repo (reproducibility), and run `git diff` immediately after every edit as a standard verification step.
8. Root-cause all failure instances, not three exemplars, before publishing class attributions.

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Sorted by impact. Rows marked ✅ are already tracked in TODO_LIST.md (added/updated this session); the rest are harvest candidates for docs-health. Effort: XS/S/M/L.

| #  | Task                                                                                                                                                                    | Impact   | Effort   | Tracked |
| -- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | -------- | ------- |
| 1  | Run the fleet sweep: `configure` across all 157 projects (fixes run.go ×148, exhaustruct ×7, typespec-sqlc breach, missing default injections)                          | Critical | M        | ✅      |
| 2  | Add schema pre-flight to `configure --check` using the committed jsonschema; report exact invalid keys before golangci-lint errors (would catch all 21 failing configs) | Critical | M        | —       |
| 3  | Exclusion-list cleanup during deprecation renames (remove old linter name from `exclusions.rules[].linters`; repairs 86 configs)                                        | High     | M        | ✅      |
| 4  | Curated-default provenance stamping + refresh of tool-owned blocks (unfreezes varnamelen ×144, mnd ×136, goconst ×22)                                                   | High     | L        | ✅      |
| 5  | Gated-linter retraction pass: remove arangolint/clickhouselint when dep absent (×123/×120), respecting sidecar + audit entry                                            | High     | M        | ✅      |
| 6  | Auto-write `never-enable` sidecar entries when ledger cycle-detection suppresses a re-enable (makes templ-components-class fixes durable)                               | High     | M        | ✅      |
| 7  | Orphaned-settings prune for `default: standard` configs (10 projects; funcorder/exhaustruct dead settings)                                                              | Medium   | S        | ✅      |
| 8  | Dead-exclusion-rule detector: flag rules whose path matches no repo file (go-dag-app: 80 copied rules)                                                                  | Medium   | M        | —       |
| 9  | Fleet mode subcommand / `--recursive` with per-repo drift report (schema-invalid, stale run.go, deprecated linters, orphans)                                            | High     | L        | —       |
| 10 | Compute fixable-now ratio for the 21 failures (tool repairs unattended vs needs features) to sequence #1 vs #3/#4                                                       | Medium   | S        | —       |
| 11 | Root-cause all 21 failures individually; attribute tool-emitted vs hand/AI edit; feed patterns into KnownBadSettingsKeys candidates                                     | Medium   | M        | —       |
| 12 | Create the first sidecar (dogfood `.golangci-lint-auto-configure.yml` in this repo) and exercise never-enable end-to-end                                                | Medium   | S        | —       |
| 13 | templ-components: add never-enable sidecar entries for ireturn/godoclint/testableexamples before any configure run re-adds them                                         | Medium   | XS       | —       |
| 14 | Regression test: exhaustruct→exhaustruct_v5 migration must also rewrite exclusion-rule linter lists (guards #3)                                                         | Medium   | S        | —       |
| 15 | BDD specs for the retraction pass (#5) incl. sidecar-respect and audit-entry assertions                                                                                 | Medium   | S        | —       |
| 16 | Decide `exhaustruct_v5` house posture (enabled in 116 configs + this repo vs never-auto-enable policy); document decision in LinterReasons (question g.2)               | Medium   | decision | —       |
| 17 | Persist the fleet-audit scripts as a repo script (`scripts/fleet-audit.py` or equivalent) so re-audits are one command                                                  | Medium   | S        | —       |
| 18 | Post-sweep re-audit → before/after metrics appended to this report's successor                                                                                          | Medium   | S        | —       |
| 19 | Extend exclusion-linter-name validation: warn when a name is not a known golangci-lint linter (catches golines-in-linters class)                                        | Medium   | S        | —       |
| 20 | Sweep the stale 2-rule test-exclusion default: older configs carry the 6-linter generation; refresh via provenance feature (#4) scope                                   | Low-Med  | M        | —       |
| 21 | Classify the 126 configless projects Go vs non-Go to size the sweep and decide auto-create candidates                                                                   | Low-Med  | S        | —       |
| 22 | Check `.golangci.yaml` (alternate extension) and nested configs for completeness of future audits                                                                       | Low      | XS       | —       |
| 23 | Normalize timeout outliers (2m/3m/15m ×6) or document them as intentional                                                                                               | Low      | XS       | —       |
| 24 | Confirm golines 180/140 outliers (PapDashboard, Rolls-Royce-mtuGoHelpCenter-golang) are intentional tuning; if not, reset to 120                                        | Low      | XS       | —       |
| 25 | go-hotspot: `default: none` + 14 dropped consensus linters — confirm intentional slim profile or re-sync                                                                | Low      | S        | —       |
| 26 | tracking-papers: 6-linter hand config, missing quadruple formatters — prime configure candidate                                                                         | Low-Med  | XS       | —       |
| 27 | ast-state-analyzer: partially visited config (missing 35 consensus linters) — configure candidate                                                                       | Low-Med  | XS       | —       |
| 28 | Investigate whether `hasTechnology` should also scan `.go` imports (hardens the dep checks behind #5)                                                                   | Low-Med  | M        | —       |
| 29 | `validate` UX: when schema verify fails, print the offending YAML path + a suggested fix line (many of the 21 are trivial renames)                                      | Medium   | M        | —       |
| 30 | Document per-project tuning etiquette in README (which keys the tool preserves; when `--force-settings` is safe)                                                        | Low      | S        | —       |
| 31 | lll posture: enabled in 11 configs, orphaned settings in 2 (desire-secrets, invoices) — decide keep-exotic vs prune-everywhere                                          | Low      | S        | —       |
| 32 | depguard: enabled in 15, disabled in 102, orphaned settings in 2 (bank-sync, ledger) — document the architectural-enforcement pattern                                   | Low      | S        | —       |
| 33 | nsfw-classifier: repair `linters.settings.goimports` → `formatters.settings.goimports` (migrator already supports; needs the visit)                                     | Low-Med  | XS       | —       |
| 34 | code-duplicate-analyzer: fix `testifylint` unknown `p` key (hand-edit artifact)                                                                                         | Low      | XS       | —       |
| 35 | Code-Quality-Agent: fix `mnd.ignored-numbers` numbers → strings                                                                                                         | Low      | XS       | —       |
| 36 | Consider `validate --format json` machine output so fleet tooling can consume failures                                                                                  | Low      | S        | —       |
| 37 | Fold the test-exclusion rule's `exhaustruct` cleanup into #3's regression tests (both exclusion surfaces)                                                               | Low      | S        | —       |
| 38 | Add fleet stats (config count, verify pass rate) to `audit` subcommand summary — makes sweep progress queryable                                                         | Low      | M        | —       |
| 39 | Consider a `--report` flag on configure that emits the drift table this audit produced (HTML/JSON per repo)                                                             | Low      | M        | —       |
| 40 | AGENTS.md: no update needed for ephemeral fleet stats, but consider one line pointing at the audit report for fleet-state context                                       | Low      | XS       | —       |
| 41 | Delete `/tmp/gci-analysis` scratch after persisting (#17) so no stale copies circulate                                                                                  | Low      | XS       | —       |
| 42 | e2e-pin-matrix: add a cell asserting the schema pre-flight (#2) rejects one known-bad key                                                                               | Low      | S        | —       |
| 43 | Explore golangci-lint's own `--fix`-style config normalize as an alternative repair path for hand-edited junk keys                                                      | Low      | M        | —       |
| 44 | template-SECURITY / template-air / testing / universal-workflow: repair classes mapped in this report (blocks for the sweep's commit messages)                          | Low      | XS       | —       |
| 45 | Check whether 1.26.x-pinned projects (16 configs) are actually on Go 1.26 toolchains or just stale configs (affects cap behavior during sweep)                          | Low-Med  | M        | —       |
| 46 | Harvest this (f) list into TODO_LIST.md/ROADMAP.md via docs-health HARVEST after Lars's call on g.1–g.3                                                                 | Low-Med  | S        | —       |
| 47 | Verify sweep-created commits don't trip per-repo CI vendorHash guards (config-only changes shouldn't, but confirm)                                                      | Low      | S        | —       |
| 48 | Consider surfacing "your config was copy-pasted from another project" detection (shared-rule fingerprints) — go-dag-app/library-policy case                             | Low      | L        | —       |
| 49 | Add the audit's "before" table to the fleet-sweep TODO row's evidence (done for headline numbers; link this report)                                                     | Low      | XS       | —       |
| 50 | Re-check AGENTS gotcha 27's claim "0 live v1 configs" — this audit confirms it for `~/projects/*/.golangci.yml` (157/157 v2); note it in the report lineage             | Low      | XS       | —       |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Fleet sweep sequencing & authority:** May I run the sweep across all 157 repos now, and if so — before or after features #3 (exclusion cleanup) and #5 (retraction) land? Running now means two mutation passes over ~150 repos; running after means one clean pass but leaves 21 configs schema-invalid (lint-breaking) in the meantime. Which tradeoff do you want?
2. **`exhaustruct_v5` posture:** 116 sibling configs plus this repo's own config have it _enabled_, while the tool's stated policy is never-auto-enable (friction 6.5). Is the enablement deliberate house style I should respect (and document as such), or should the sweep actively drop it and record a `never-enable` sidecar entry?
3. **Provenance mechanism preference for #4:** YAML comment stamps inside `.golangci.yml` (self-contained, but adds tool-owned noise to every config you also hand-edit) vs. recording tool-owned block hashes in the sidecar file (pure `.golangci.yml`, but a second state file to keep in sync)? I cannot decide this without knowing which file you consider the source of truth when the two disagree.

---

_Point-in-time snapshot. Successor report after the fleet sweep should re-run steps 1–4 of the method above and diff against the Headline Numbers table._
