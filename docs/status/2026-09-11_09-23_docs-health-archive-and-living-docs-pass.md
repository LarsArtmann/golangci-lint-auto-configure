# Status Report: Docs-Health Audit — Living-Docs Overhaul + Archive Sweep

**Date:** 2026-09-11 09:23 CEST
**Session scope:** Execute the docs-health skill (AUDIT = BUILD + HARVEST + VERIFY + ANNOTATE + ARCHIVE) across ALL `**/2026-0*` docs; make the six living docs superb; annotate fully-resolved reports inline and archive them.
**Headline:** 74 active docs read and classified, 6 living docs rewritten/verified against code, 765 forward-looking items struck inline across 17 resolved reports, 28 files archived via `git mv` (all renames), markdownlint 41 files → 0 issues. One corruption-class bug caught in dry-run before it touched 765 lines.

**Commits:** this session's 35 changed paths landed in daemon commit `fd41881` ("61 changed file(s)" — interleaved with other work).

---

## a) FULLY DONE

| #  | What                                                                                                          | Evidence                                                                                                                                                                      |
|----|---------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| a1 | Full inventory: 230+ `2026-0*` files found; 74 active docs classified (58 status, 9 planning, 3 research, 1 review, 1 architecture-understanding, 2 feedback) | `find docs -name "2026-0*"`; four sub-agent reads (11+20+10+16 files) + direct reads of all six living docs and the three most recent reports                                  |
| a2 | Claim verification battery — 20+ spot-checks against code                                                     | `rg`-verified: Flags struct shipped (only 1 package-level var left, `internal/cli/commands.go:24`); exhaustruct still old name (`rules.go:154`); 8 inline ADRs vs 7 oddly-named files in `docs/adr/`; G104 gone from `.golangci.yml`; zero multi-preset merge tests; docs-integrity test covers presets only; `test.golangci.yml` is buildflow-used (kept); depguard in `NeverAutoEnableLinters` (`rules.go:152`); mnd `IgnoredNumbers` empty; 5 new settings structs exist (`linter_settings.go:63-67`); `pruneUnenabledLinterSettings` at `fixer_config.go:358`; health checks at `pkg/types/validation.go`; gohumanize gating at `config.go:69` + `categorizer.go:234` |
| a3 | **TODO_LIST.md rewritten** — 5 High / 13 Medium / 14 Low, every item with evidence; harvested both 09-11 reports' f-sections; removed the done markdown-lint row | `TODO_LIST.md` (written this session); prior stale rows (GoReleaser docker, MD029 lint) updated/removed per `ff88e51` + 09-11 evidence                                          |
| a4 | **CHANGELOG.md backfilled** — the entire August work was missing from v0.7.0; `[Unreleased]` was empty          | v0.7.0 now records gohumanize (+ H004 nolint), CV-config defaults (5 structs, mnd/wrapcheck/errcheck/varnamelen), depguard → NeverAutoEnable, `pruneUnenabledLinterSettings`, 2 health checks; `[Unreleased]` records the goconst `min-len` schema fix, CI rehabilitation, and the release buildx/login fix |
| a5 | **FEATURES.md de-drifted** — audit bumped to v0.8.0/2026-09-11; 12 missing features added; 2 wrong rows fixed | Fixed: depguard listed under `DisabledLinters` (actually NeverAutoEnable); mnd row still claimed `ignored-numbers: 0,1,2,100`. Added: `--force-settings`, YAML indent preservation, orphaned-settings pruning, gohumanize, gocognit/gocyclo/nestif/goconst/tagalign, gogenfilter v3 delegation, mockery/counterfeiter presets, config health checks |
| a6 | **ROADMAP.md rewritten** — 3 shipped themes pruned (Flags/CommandContext, `--force-settings`, YAML indent — all verified in code); added public-repo ops theme + "Open questions" section (3 user-gated questions) | `ROADMAP.md`; the old theme 2 claimed "9 package-level variables in `internal/cli/commands.go`" — reality: 1 (`Version`)                                                            |
| a7 | **README.md synced** — version recommendation v2.12.2+ → v2.13.2 (devShell/CI reality); zero-clone `nix run github:…` one-liner; no-SSH note | `README.md:58`, install section; closes 09-11 metadata report f20/f22 and 09-09 f13                                                                                            |
| a8 | **AGENTS.md fixes** — gotcha #13 rewritten (no `.buildflow.yml` exists; documents CI lint v2.13.2 + test-job binary install); #19 exhaustruct_v5 deprecation flagged; #30 gohumanize self-detection nolint documented | `AGENTS.md`; closes 09-11 f17 and the 08-07 report's open AGENTS item                                                                                                         |
| a9 | **ANNOTATE: 765 items struck inline** across the 17 fully-resolved status reports, each with a top-placed resolution banner citing per-file successor evidence | `/tmp/archive_pass.py` (dry-run first per skill mandate; banner inserted directly under H1, strikes limited to forward-looking c/f sections, d/e narrative left intact)            |
| a10| **ARCHIVE: 28 files `git mv`'d** — 17 status → `docs/archive/status/` (now 152), 9 planning → `docs/archive/planning/` (now 25), `PUBLIC_OR_PRIVATE.md` → `docs/archive/` (decision resolved 09-09) | `git status`: all 35 paths are renames (R) or doc edits; zero history loss; the one relative cross-link (`07-21_10-38` → `07-21_09-58`) still resolves because both moved together |
| a11| Planning docs bannered with per-file execution evidence                                                       | e.g. friction plan: "C16 → TODO_LIST; C19 shipped via v0.6.0–v0.8.0"; DISABLED-LINTERS plan: "commit `31df177`"                                                                 |
| a12| Targeted annotations in 3 recent KEEP reports — items THIS session completed are struck with evidence          | 09-11 metadata (f9 harvest, f17, f20, f22, f16-partial), 09-11 buildflow (f34 harvest, gotcha-#35 commit), 08-08 self-review (status-README index)                                |
| a13| `docs/status/README.md` index rebuilt — archive notice + 39 live reports regrouped into 7 themes               | `docs/status/README.md`                                                                                                                                                       |
| a14| Quality gates green — markdownlint 41 files 0 issues (via `nix develop -c markdownlint-cli2`); link integrity clean (no refs to moved files outside archive); no Go code touched | markdownlint output "0 issues in 0 files"; `rg` link sweep empty                                                                                                              |

## b) PARTIALLY DONE

| #  | Item                                            | Works now                                                                      | Remains open                                                                                                  | Effort |
|----|-------------------------------------------------|--------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|--------|
| b1 | Archive evidence fidelity                       | Every struck item carries concrete chain-level evidence ("resolved by successor report X / header note") | Per-item commit-hash attribution — daemon-commit history makes it unreliable; a precision pass with the skill's annotate scripts is possible | 1–2h per dozen reports |
| b2 | Root one-off reports relocation (09-11 f16)     | `PUBLIC_OR_PRIVATE.md` archived (decision executed 09-09)                      | `PARTS.md`, `PROJECT_SPLIT_EXECUTIVE_REPORT.md`, `BDD_TESTS_REVIEW.md` kept pending decision (TODO_LIST Low)    | decision |
| b3 | KEEP-report annotation                          | 3 of 39 live reports got targeted strikes for items closed today               | The other 36 still carry unstruck resolved-items; full ANNOTATE pass not started                                | M      |
| b4 | 08-05 humanize report annotation (08-07 open item "annotate 03-25 as resolved") | Identified                                                      | Not struck — the 08-07 report's item stays open                                                                | S      |
| b5 | ROADMAP idea-carryover completeness             | Core themes + open questions carried                                           | Small 07-31 F-list ideas (`--reset-exclusions`, `--indent` flag, fuzz tests) were dropped, not consciously routed | 30min  |
| b6 | Version-reference sweep (09-11 e9)              | README Requirements synced                                                     | Other doc version references (AGENTS #11 release notes, docs/references) not swept                             | S      |

## c) NOT STARTED

| #  | Item                                                   | Why                                                                     |
|----|--------------------------------------------------------|-------------------------------------------------------------------------|
| c1 | exhaustruct_v5 migration (code)                        | Docs session; tracked TODO_LIST High                                     |
| c2 | `min-length` key-normalization fixer pass (code)       | Critical; tracked TODO_LIST High                                         |
| c3 | CI schema-compat gate (code)                           | Critical; tracked TODO_LIST High                                         |
| c4 | BDD spec debt for the 08-08 features (tests)           | Tracked TODO_LIST Medium                                                 |
| c5 | The two HTML dashboards (`docs/reviews/`, `docs/architecture-understanding/`) | Reference material, not annotation targets; untouched by design |
| c6 | Re-verification of the ~150 pre-existing archive files | Out of scope; previous passes own them                                   |
| c7 | Committing by hand                                     | Daemon committed as `fd41881`; no manual commit needed                   |
| c8 | `001-yaml-dependency-decision.md` rename               | Flagged inside the ADR-consolidation TODO item only                      |
| c9 | gohumanize mention in README                           | Deliberate: niche dep-gated feature; FEATURES/CHANGELOG are its homes    |

## d) TOTALLY FUCKED UP

1. **The strike transform had a corruption-class bug — caught by the dry-run, not by luck.** `strike_cell2` originally struck `cells[1]` of the post-regex remainder, which is the **Impact** column, not the task text (regex `group(4)` starts after the number-cell's trailing pipe, so item text is `cells[0]`). Had I applied blind, 765 lines would read `| 3 | Fix warmup… | ~~Critical~~ done — … |` — evidence attached to the wrong column in 17 historical files. The skill's dry-run-first mandate plus a sample-line test caught it before any write. Lesson generalized: **unit-test text transforms on samples before dry-running them against real files** — the dry-run found it only because I inspected actual output, and a second bug class (regex subtleties) hides from counts alone.
2. **Read-before-edit rejected three times** (CHANGELOG, README, AGENTS.md) — I edited from project-context content, which the tool (correctly) doesn't count. This is the *same failure class* documented in the 09-09 and 09-11 reports ("View-before-edit is mechanical, not optional"). Three wasted round trips on a lesson I already knew.
3. **First FEATURES multiedit failed 8 of 9 edits** — I drafted aligned table rows from memory with doubled pipes (`||`) instead of copying the file's exact single-pipe rows. Root cause: composing "what the row should say" instead of "what the row says". Cost: one full retry cycle. Aligned-column tables make approximate matching guaranteed to fail; exact text must come from the file.
4. **Sub-agent quota/rate-limit failures** — first batch (36 files) died on "insufficient quota"; the retry hit a rate limit. Recovered by splitting into 10/16/9-file chunks. Two wasted round trips; the 36-file batch was over-sized for the quota budget from the start.
5. **QMD call wasted** — `mcp_qmd_multi_get` cannot see this repo's docs (index: 0 documents). The MCP instructions literally said so; I reached for it out of habit before falling back to filesystem tools.
6. **Generic evidence on some strikes** — a minority of struck items carry "done — see header resolution note" rather than item-specific evidence. Defensible (the banner carries the per-file verification story), but it dilutes the "so what?" test the skill demands; flagged as b1 rather than hidden.

## e) WHAT WE SHOULD IMPROVE

1. **View immediately before edit, every time** — three rejections this session on a rule documented twice in recent reports. It needs to be a reflex, not a memory: *no edit call without a View call of that exact file in the same session*.
2. **Sample-test transforms before dry-runs** — the strike bug was visible in a 3-line sample test; running that first would have saved the patch cycle.
3. **Never compose table rows from memory** — copy the exact row, then change only the cells that must change. Doubled-pipe typos don't survive copying.
4. **Size agent batches to the quota budget (~10 files)** — 36-file batches fail on billing/rate limits; chunking is cheap, failures are not.
5. **Check the QMD index state before using it** — 0 indexed documents means filesystem tools are the only path here.
6. **Record tradeoff decisions where the work lands** — the banner's "chain-level evidence" note is the model: the reader knows the evidence standard without archaeology.
7. **Keep the archive script if archive passes recur** — `/tmp/archive_pass.py` is throwaway; a parameterized `scripts/docs-health-archive.py` (corrected strike logic, dry-run default, per-file specs) would make the next sweep 30 minutes instead of a session.

## f) TOP 50 THINGS WE SHOULD GET DONE NEXT

Ranked; Effort: S <30min, M 30min–2h, L >2h. Items 1–20 mirror the fresh TODO_LIST (not restated in detail there).

| #  | Task                                                                                                         | Impact   | Effort |
|----|--------------------------------------------------------------------------------------------------------------|----------|--------|
| 1  | Cut v0.8.1 (min-len fix + CI rehab) and watch the release pipeline end-to-end                                 | Critical | S      |
| 2  | Key-normalization pass: rewrite `goconst.min-length` → `min-len` in user configs                              | Critical | M      |
| 3  | CI schema-compat gate: `golangci-lint config verify` over all injected defaults                               | Critical | M      |
| 4  | Daemon/vendorHash guard: `nix build` on go.mod/go.sum changes                                                 | Critical | M      |
| 5  | exhaustruct → exhaustruct_v5 migration (constants, excludes, deprecation table, data tests)                   | High     | M      |
| 6  | Generated-file drift guard in CI (`git diff --exit-code` after regenerate)                                    | High     | S      |
| 7  | CI-health watchdog (weekly workflow-state + master-green check)                                               | High     | S      |
| 8  | Release dry-run (`goreleaser --snapshot`) on PRs                                                              | High     | M      |
| 9  | Sibling sweep for emitted `min-length` keys                                                                   | High     | M      |
| 10 | Verify homebrew/scoop manifests published for v0.8.0                                                          | High     | S      |
| 11 | Backfill GHCR image for v0.8.0                                                                                | Medium   | S      |
| 12 | BDD spec debt for the 08-08 features + trim `varnamelen.IgnoreDecls`                                          | High     | M      |
| 13 | Branch/tag protection rulesets (`master`, `v*`)                                                               | Medium   | S      |
| 14 | gitleaks full-history scan                                                                                    | Medium   | M      |
| 15 | `go install …@latest` in a clean `GOMODCACHE`                                                                 | Medium   | S      |
| 16 | Dependabot Updates failures (08-23, 08-30, 09-06)                                                             | Medium   | S      |
| 17 | Buildflow `test-coverage` timeout + end-to-end re-run                                                         | Medium   | S–M    |
| 18 | Schema-version gate in `cmd/generate-settings`                                                                | Medium   | M      |
| 19 | `internal/cli` coverage (25.8%) + 116s `-race` suite                                                          | Medium   | M–L    |
| 20 | json/v2 `omitempty` → `omitzero` in `config_types.go`                                                         | Medium   | M      |
| 21 | ADR consolidation + `001-yaml-dependency-decision.md` naming fix                                              | Medium   | 1–2h   |
| 22 | Full README claim-by-claim audit                                                                              | Medium   | 2h     |
| 23 | Precision ANNOTATE pass over the 39 live reports (per-item hashes via annotate scripts)                       | Medium   | M      |
| 24 | Annotate `2026-08-05_03-25_humanize-linter-status.md` as resolved (08-07 open item)                           | Low      | S      |
| 25 | Decide fates of `PARTS.md` / `PROJECT_SPLIT_EXECUTIVE_REPORT.md` / `BDD_TESTS_REVIEW.md`                      | Low      | decision |
| 26 | Sweep remaining doc version references (AGENTS #11, docs/references) post-2.13.2 sync                          | Low      | S      |
| 27 | Consciously route or drop the small 07-31 F-list ideas (`--reset-exclusions`, `--indent`, fuzz tests)         | Low      | S      |
| 28 | README CI/CD section (what runs, what gates)                                                                  | Low      | S      |
| 29 | Extend docs-integrity test to all hardcoded FEATURES counts                                                   | Low      | 1h     |
| 30 | Status-report lifecycle cadence policy (this sweep was one-off)                                               | Low      | 30min  |
| 31 | Multi-preset merge correctness tests (verified still missing)                                                 | Low      | 1h     |
| 32 | Quarterly erraudit re-check (~2026-10 due)                                                                    | Low      | 1h     |
| 33 | Implement-or-remove `FindingsHidden` dead ledger field                                                        | Low      | 30min  |
| 34 | Full `FixConfig` sidecar + ledger integration test                                                            | Low      | M      |
| 35 | `Detect()` error-path contract decision                                                                       | Low      | decision |
| 36 | `linter_settings_generated.go` goconst drift decision                                                         | Low      | S      |
| 37 | Document cosign/SBOM artifact verification in README                                                          | Low      | S      |
| 38 | Confirm `gh workflow run ci.yml` dispatch works                                                               | Low      | S      |
| 39 | Root-cause `nix flake check` "0 checks" anomaly                                                               | Low      | S      |
| 40 | Decide `auto-tag.yml` fate (delete vs keep-disabled)                                                          | Low      | S      |
| 41 | Render-check a sample of struck archive tables in a Markdown renderer (pandoc/GitHub view)                    | Low      | S      |
| 42 | Consider committing the archive-pass script as `scripts/docs-health-archive.py` if sweeps recur               | Low      | S      |
| 43 | gohumanize row in README (only if you want it on the sales page — currently deliberate omission)              | Low      | S      |
| 44 | Add CI/CD-tooling entries convention note to CHANGELOG header (f38 of 09-11 report)                           | Low      | S      |
| 45 | Add `nix` topic decision + metadata checklist script (09-11 f25/f28)                                          | Low      | M      |
| 46 | Link-checker (lychee) in CI (09-09 f23, still open)                                                           | Low      | S      |
| 47 | Demo GIF/asciinema for README (09-09 f15, carried)                                                            | Low      | M      |
| 48 | Decide homepage target: website launch vs GitHub anchor (09-11 f26)                                           | Medium   | M      |
| 49 | Announcement posture decision (officially-maintained vs portfolio; gates the community tier)                  | Medium   | decision |
| 50 | Re-verify TODO_LIST state next session (daemon interleaving may drift evidence hashes)                        | Low      | S      |

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Evidence standard for the 765 struck items:** is chain-level evidence ("resolved by successor report / header note") acceptable for the archived reports, or do you want a per-item commit-hash pass? The daemon's heuristic commit history makes per-item attribution partly guesswork; I chose honest chain-level citations over fabricated precision. A precision pass is feasible (~1–2h per dozen reports) but I won't start it without knowing the standard you want.
2. **The three remaining root one-off reports** (`PARTS.md`, `PROJECT_SPLIT_EXECUTIVE_REPORT.md`, `BDD_TESTS_REVIEW.md`): act on them, archive them, or delete them? `PARTS.md`/`PROJECT_SPLIT_EXECUTIVE_REPORT.md` are undecided project-split proposals; `BDD_TESTS_REVIEW.md` has Sep-2 findings nobody has actioned. All three are referenced only by status reports (no code/build impact either way).
3. **Target state for `docs/status/`:** should it eventually hold ONLY reports with live residue (current state after this sweep: 39), with everything else archived as it resolves — or do you prefer keeping reports flat in `docs/status/` for a while and archiving in bigger periodic sweeps? This decides whether the next docs-health run should be continuous (annotate-as-you-go) or sweep-based (like today).

---

**Format note:** Markdown per session convention (`.md` reports throughout `docs/status/`).
**Handoff:** section (f) items 1–22 are already in `TODO_LIST.md` with evidence; 23–50 are the delta this report adds on top.
