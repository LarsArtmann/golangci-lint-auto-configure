# Status Reports Index

Point-in-time snapshots from development sessions. Each report captures what was done, what was found, and what remained at that moment.

**For current project status**, see [`../../TODO_LIST.md`](../../TODO_LIST.md) and [`../../FEATURES.md`](../../FEATURES.md).

**For change history**, see [`../../CHANGELOG.md`](../../CHANGELOG.md).

> **Archive sweep (2026-09-11, docs-health pass):** the fully-resolved
> 2026-06/07 reports (JSON-v2 chain, error migration, 07-10 sweeps, audit-ledger
> pillars, omitempty fixes, and the matching planning docs) were annotated
> inline and moved to [`../archive/status/`](../archive/status/) and
> [`../archive/planning/`](../archive/planning/). Reports listed below still
> carry open residue — their "next tasks" sections are the HARVEST source that
> feeds `TODO_LIST.md`.

---

## Live Reports by Theme

### Quality Sprints & Sweeps (2026-07-25/26)

| Date       | Report                                        | Key Outcome                                              |
| ---------- | --------------------------------------------- | -------------------------------------------------------- |
| 2026-07-25 | `05-40_forcetypeassert-test-exclusion-default`  | Test exclusion default                                   |
| 2026-07-25 | `06-30_docs-health-and-old-docs-annotation-pass`| Old-docs annotation pass                                 |
| 2026-07-25 | `06-52_follow-up-resolving-docs-health-open-questions` | Open questions resolved                            |
| 2026-07-25 | `07-03_test-coverage-sprint-status`             | Enforcement + audit CLI tests                            |
| 2026-07-25 | `07-35_docs-health-todo-sweep-self-review`      | 15/50 sweep items                                        |
| 2026-07-25 | `14-01_50-item-todo-list-second-sweep`          | 16/50 sweep items                                        |
| 2026-07-25 | `14-01_friction-reduction-plan-execution`       | `--pragmatic`, house preset, funlen 200/100              |
| 2026-07-25 | `14-29_quality-debt-cleanup`                    | CoreFormatters, gosec/errcheck alignment                 |
| 2026-07-25 | `17-31_coverage-check-quality-debt-cleanup`     | Ghost script deleted, parser specs                       |
| 2026-07-25 | `18-15_comprehensive-50-item-execution`         | 32/50 executed                                           |
| 2026-07-25 | `20-55_complete-17-item-execution`              | Remaining 17 executed                                    |
| 2026-07-26 | `06-33_docs-health-and-old-docs-second-pass`    | Second annotation pass                                   |
| 2026-07-26 | `09-41_superb-plan-phase-1-3-execution`         | Branded types, decoupling                                |
| 2026-07-26 | `10-06_deduplication-to-zero-session-report`    | art-dupl → 0 clones                                      |
| 2026-07-26 | `16-37_superb-plan-phase-4-execution`           | Flags struct, SettingsMap                                |
| 2026-07-26 | `17-13_phase-4-cleanup-comprehensive`           | Dead code, 12 SettingsMap specs                          |

### Linter Policy & Tiers (2026-07-10 → 07-26)

| Date       | Report                                          | Key Outcome                                |
| ---------- | ----------------------------------------------- | ------------------------------------------ |
| 2026-07-10 | `11-34_noinlineerr-disabled-formatter-conflict` | noinlineerr disabled                       |
| 2026-07-20 | `22-57_disable-respect-audit-ledger-enforcement-completion` | Ledger + sidecar enforcement shipped |
| 2026-07-26 | `20-07_exhaustruct-never-auto-enable-tier`      | NeverAutoEnable tier                       |
| 2026-07-26 | `20-28_exhaustruct-followup-cleanup-pass`       | Validation checks 7–8                      |
| 2026-07-26 | `20-43_exhaustruct-followup-brutal-self-review` | Test renames, authorship confession        |

### Error Handling Reviews (2026-07-26)

| Date       | Report                                              | Key Outcome                          |
| ---------- | --------------------------------------------------- | ------------------------------------ |
| 2026-07-26 | `21-24_erraudit-full-review-brutal-self-assessment` | 4 real bugs fixed, nolint reverted   |
| 2026-07-26 | `21-40_erraudit-followup-test-coverage-and-verification` | Error-path tests                |
| 2026-07-26 | `22-07_detector-error-propagation-fix-and-self-assessment` | Scanner error propagation |

### Releases (2026-07-27)

| Date       | Report                                     | Key Outcome                              |
| ---------- | ------------------------------------------ | ---------------------------------------- |
| 2026-07-27 | `01-25_v0.6.0-release-self-assessment`     | v0.6.0 cut; brutal process review        |
| 2026-07-27 | `01-42_post-release-cleanup-status`        | ldflags fix, GoReleaser deprecations     |

### Regression-Loop Prevention & Pareto (2026-07-30/31)

| Date       | Report                                              | Key Outcome                             |
| ---------- | --------------------------------------------------- | --------------------------------------- |
| 2026-07-30 | `22-39_regression-loop-prevention-never-enable-cycle-detection` | `never-enable` + cycle detection |
| 2026-07-30 | `23-21_PARETO-SESSION-REVIEW`                       | YAML indent, `--force-settings`         |
| 2026-07-30 | `23-22_never-enable-enforcement-fix-and-review`     | Enforcement bypass fixed                |
| 2026-07-31 | `03-40_SESSION-2-REVIEW`                            | Tab handling, merger RuleKey dedup      |

### gohumanize & CV-Config Learnings (2026-08)

| Date       | Report                                          | Key Outcome                              |
| ---------- | ----------------------------------------------- | ---------------------------------------- |
| 2026-08-05 | `03-25_humanize-linter-status`                  | Failed first H004 attempt (honest)       |
| 2026-08-05 | `03-48_gohumanize-project-specific-integration` | Dep-gated gohumanize shipped             |
| 2026-08-05 | `04-14_gohumanize-everywhere-pushback-comparison` | Module-plugin vs Go-plugin analysis    |
| 2026-08-07 | `08-58_gohumanize-linter-h004-resolved`         | H004 resolved via documented nolint      |
| 2026-08-08 | `01-07_cv-config-learnings-implementation`      | 11 CV-derived default improvements       |
| 2026-08-08 | `01-23_cv-config-learnings-brutal-self-review`  | BDD-spec debt surfaced                   |

### Public Launch & CI Rehabilitation (2026-09)

| Date       | Report                                            | Key Outcome                                   |
| ---------- | ------------------------------------------------- | --------------------------------------------- |
| 2026-09-09 | `02-08_going-public-launch-and-sanitization`      | Repo made public; sanitization battery        |
| 2026-09-11 | `06-38_buildflow-failures-resolved`               | Generator formatting, vendorHash triage       |
| 2026-09-11 | `08-33_github-metadata-and-ci-rehabilitation-status` | Metadata, CI re-enable, goconst schema fix |
