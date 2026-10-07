# Status Reports Index

Point-in-time snapshots from development sessions. Each report captures what was done, what was found, and what remained at that moment.

**For current project status**, see [`../../TODO_LIST.md`](../../TODO_LIST.md) and [`../../FEATURES.md`](../../FEATURES.md).

**For change history**, see [`../../CHANGELOG.md`](../../CHANGELOG.md).

> **Archive sweep (2026-09-11, docs-health pass):** the fully-resolved
> 2026-06/07 reports (JSON-v2 chain, error migration, 07-10 sweeps, audit-ledger
> pillars, omitempty fixes, and the matching planning docs) were annotated
> inline and moved to [`../archive/status/`](../archive/status/) and
> [`../archive/planning/`](../archive/planning/). **Second sweep (2026-09-14,
> pareto plan T26):** the 07-31 session review (all 50 section-f items routed
> or dropped) and the 08-05 humanize first-attempt report (H004 resolved via
> documented nolint) were annotated and archived. Reports listed below still
> carry open residue — their "next tasks" sections are the HARVEST source that
> feeds `TODO_LIST.md`.
>
> **Third sweep (2026-10-07, docs-health AUDIT):** all 48 non-archived
> 2026-0* docs were read and every numbered forward-looking item re-verified
> against the repo (~1,500 items across 5 triage passes). Eight fully-resolved
> or fully-routed files were annotated inline (every item struck or routed)
> and archived:
>
> | Archived file                                                              | Classification | Deciding reason                                                                    |
> | -------------------------------------------------------------------------- | -------------- | ---------------------------------------------------------------------------------- |
> | 2026-07-25_06-30 docs-health annotation pass                               | ARCHIVE        | 56/56 §e/§f items closed (1 routed: error-code registry → ROADMAP theme 5)         |
> | 2026-07-25_06-52 follow-up resolving open questions                        | ARCHIVE        | 50/50 §f items closed (1 routed to ROADMAP theme 5); §g answered                   |
> | 2026-07-25_17-31 coverage-check quality-debt cleanup                       | ARCHIVE        | 61/61 §c/§f items closed (2 declined: OutputConfig typing, CI retry)               |
> | 2026-07-25_18-15 comprehensive 50-item execution                           | ARCHIVE        | 17/17 remaining-item rows struck with commit hashes                                |
> | 2026-07-25_20-55 complete 17-item execution                                | ARCHIVE        | No forward-looking section; 2 stale claims inline-corrected                        |
> | 2026-07-06 pipeline comparison (research)                                  | ARCHIVE        | 5/5 recommendations resolved (3 shipped, 2 validated); banner + inline strikes     |
> | 2026-07-25 ecosystem report (research)                                     | ARCHIVE        | 7/7 recommendations shipped; residue routed to ROADMAP theme 2                     |
> | 2026-07-10 deep architecture review                                        | ARCHIVE        | 12/12 P0–P3 recommendations shipped; Resolution Status table verified              |
>
> **Fourth sweep wave (same day, 2026-10-07):** after routing, seven more
> reports became fully resolved (every numbered item struck or routed) and
> were archived: 2026-07-26_17-13 (50/50 closed), 2026-07-30_22-39 (50/50),
> 2026-07-30_23-21 (50/50), 2026-07-30_23-22 (50/50), 2026-09-11_13-27
> (30/30), 2026-09-14_11-22 (10/10 routed), 2026-09-14_11-34 (30/30 routed).
> The remaining 31 live reports carry partial strikes (done items struck,
> open items bare) plus sweep-verification banners where item-level strikes
> were not mechanically possible.
>
> The remaining 31 live reports keep bare (unstruck) open items — absence of a
> marker IS the open signal. Harvest from this sweep: 10 new TODO_LIST rows,
> ROADMAP theme-2/4 additions, 2 FEATURES ghost rows removed (auto-tag
> workflow, CI retry logic), AGENTS gotcha 13 version-drift fix.

## Lifecycle & Cadence

Status reports are point-in-time snapshots, never living documents. The
standing rule (established 2026-09-14, pareto plan T26.4):

1. **Write** a report at the end of any session that ships or decides
   something worth recording; its open residue goes into `TODO_LIST.md`
   immediately (harvest-at-creation).
2. **Resolve** a report once nothing in it is open: add a top annotation
   blockquote stating each open item's fate (shipped / routed to ROADMAP /
   consciously dropped), then `git mv` to `../archive/status/` and remove its
   row from the index below. Verify stale claims before annotating — a report
   saying "X is broken" is not evidence that X is still broken.
3. **Sweep** `docs/status/` at least quarterly, or whenever live reports
   exceed ~15, whichever comes first. The sweep is mechanical: for each
   report, re-verify open claims, annotate, archive.
4. **Never edit** an archived report's body (annotations go at the top only);
   historical text stays as written.

---

## Live Reports by Theme

### Quality Sprints & Sweeps (2026-07-25/26)

| Date       | Report                                                 | Key Outcome                                 |
| ---------- | ------------------------------------------------------ | ------------------------------------------- |
| 2026-07-25 | `05-40_forcetypeassert-test-exclusion-default`         | Test exclusion default                      |
| 2026-07-25 | `07-03_test-coverage-sprint-status`                    | Enforcement + audit CLI tests               |
| 2026-07-25 | `07-35_docs-health-todo-sweep-self-review`             | 15/50 sweep items                           |
| 2026-07-25 | `14-01_50-item-todo-list-second-sweep`                 | 16/50 sweep items                           |
| 2026-07-25 | `14-01_friction-reduction-plan-execution`              | `--pragmatic`, house preset, funlen 200/100 |
| 2026-07-25 | `14-29_quality-debt-cleanup`                           | CoreFormatters, gosec/errcheck alignment    |
| 2026-07-26 | `06-33_docs-health-and-old-docs-second-pass`           | Second annotation pass                      |
| 2026-07-26 | `09-41_superb-plan-phase-1-3-execution`                | Branded types, decoupling                   |
| 2026-07-26 | `10-06_deduplication-to-zero-session-report`           | art-dupl → 0 clones                         |
| 2026-07-26 | `16-37_superb-plan-phase-4-execution`                  | Flags struct, SettingsMap                   |

### Linter Policy & Tiers (2026-07-10 → 07-26)

| Date       | Report                                                      | Key Outcome                          |
| ---------- | ----------------------------------------------------------- | ------------------------------------ |
| 2026-07-10 | `11-34_noinlineerr-disabled-formatter-conflict`             | noinlineerr disabled                 |
| 2026-07-20 | `22-57_disable-respect-audit-ledger-enforcement-completion` | Ledger + sidecar enforcement shipped |
| 2026-07-26 | `20-07_exhaustruct-never-auto-enable-tier`                  | NeverAutoEnable tier                 |
| 2026-07-26 | `20-28_exhaustruct-followup-cleanup-pass`                   | Validation checks 7–8                |
| 2026-07-26 | `20-43_exhaustruct-followup-brutal-self-review`             | Test renames, authorship confession  |

### Error Handling Reviews (2026-07-26)

| Date       | Report                                                     | Key Outcome                        |
| ---------- | ---------------------------------------------------------- | ---------------------------------- |
| 2026-07-26 | `21-24_erraudit-full-review-brutal-self-assessment`        | 4 real bugs fixed, nolint reverted |
| 2026-07-26 | `21-40_erraudit-followup-test-coverage-and-verification`   | Error-path tests                   |
| 2026-07-26 | `22-07_detector-error-propagation-fix-and-self-assessment` | Scanner error propagation          |

### Releases (2026-07-27)

| Date       | Report                                 | Key Outcome                          |
| ---------- | -------------------------------------- | ------------------------------------ |
| 2026-07-27 | `01-25_v0.6.0-release-self-assessment` | v0.6.0 cut; brutal process review    |
| 2026-07-27 | `01-42_post-release-cleanup-status`    | ldflags fix, GoReleaser deprecations |

### Regression-Loop Prevention & Pareto (2026-07-30/31)

| Date       | Report                                                          | Key Outcome                      |
| ---------- | --------------------------------------------------------------- | -------------------------------- |

### gohumanize & CV-Config Learnings (2026-08)

| Date       | Report                                            | Key Outcome                         |
| ---------- | ------------------------------------------------- | ----------------------------------- |
| 2026-08-05 | `03-48_gohumanize-project-specific-integration`   | Dep-gated gohumanize shipped        |
| 2026-08-05 | `04-14_gohumanize-everywhere-pushback-comparison` | Module-plugin vs Go-plugin analysis |
| 2026-08-07 | `08-58_gohumanize-linter-h004-resolved`           | H004 resolved via documented nolint |
| 2026-08-08 | `01-07_cv-config-learnings-implementation`        | 11 CV-derived default improvements  |
| 2026-08-08 | `01-23_cv-config-learnings-brutal-self-review`    | BDD-spec debt surfaced              |

### Public Launch & CI Rehabilitation (2026-09)

| Date       | Report                                                     | Key Outcome                                                                         |
| ---------- | ---------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| 2026-09-09 | `02-08_going-public-launch-and-sanitization`               | Repo made public; sanitization battery                                              |
| 2026-09-11 | `06-38_buildflow-failures-resolved`                        | Generator formatting, vendorHash triage                                             |
| 2026-09-11 | `08-33_github-metadata-and-ci-rehabilitation-status`       | Metadata, CI re-enable, goconst schema fix                                          |
| 2026-09-11 | `09-23_docs-health-archive-and-living-docs-pass`           | Living-docs overhaul + 28-file archive sweep                                        |
| 2026-09-11 | `23-26_pareto-execution-t11-t25-thirteen-tasks-and-honest-scars` | T11–T25: GHCR backfill, rulesets, gitleaks, ADR consolidation, README audit |
| 2026-09-28 | `21-59_go-1.27-readiness-shipped-and-reviewed`             | Go 1.27 readiness: `run.go` major.minor + cap + rescue; live frontier               |
