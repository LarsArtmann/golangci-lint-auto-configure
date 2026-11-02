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
> | Archived file                                        | Classification | Deciding reason                                                                |
> | ---------------------------------------------------- | -------------- | ------------------------------------------------------------------------------ |
> | 2026-07-25_06-30 docs-health annotation pass         | ARCHIVE        | 56/56 §e/§f items closed (1 routed: error-code registry → ROADMAP theme 5)     |
> | 2026-07-25_06-52 follow-up resolving open questions  | ARCHIVE        | 50/50 §f items closed (1 routed to ROADMAP theme 5); §g answered               |
> | 2026-07-25_17-31 coverage-check quality-debt cleanup | ARCHIVE        | 61/61 §c/§f items closed (2 declined: OutputConfig typing, CI retry)           |
> | 2026-07-25_18-15 comprehensive 50-item execution     | ARCHIVE        | 17/17 remaining-item rows struck with commit hashes                            |
> | 2026-07-25_20-55 complete 17-item execution          | ARCHIVE        | No forward-looking section; 2 stale claims inline-corrected                    |
> | 2026-07-06 pipeline comparison (research)            | ARCHIVE        | 5/5 recommendations resolved (3 shipped, 2 validated); banner + inline strikes |
> | 2026-07-25 ecosystem report (research)               | ARCHIVE        | 7/7 recommendations shipped; residue routed to ROADMAP theme 2                 |
> | 2026-07-10 deep architecture review                  | ARCHIVE        | 12/12 P0–P3 recommendations shipped; Resolution Status table verified          |
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

> **Fifth sweep (2026-11-02, docs-health AUDIT):** re-verified every bare
> item in the remaining 2026-0* corpus against the post-v0.11.0 tree (legacy
> engine + compat views + `rules.go`/`linter_reasons.go` deleted; tier enum on
> the policy table; lll redundancy removal accounted). Newly-resolved items
> struck inline with per-item evidence; files whose every item carries a
> verdict were archived (manifest in `../archive/status/`). TODO_LIST/FEATURES/
> README/CHANGELOG/AGENTS rehabilitated in the same pass (AGENTS gotchas 11/20
> and detail items 2/4/5 de-drifted against the current code layout).
>
> **Fifth-sweep archive manifest (32 files, all `git mv` renames):**
>
> | Group (files) | Classification | Deciding reason |
> | ------------- | -------------- | --------------- |
> | 2026-07-10 noinlineerr, 2026-07-20 disable-respect | ARCHIVE | every item struck/declined; both `g` questions answered inline |
> | 2026-07-25/26 sprints (05-40, 07-03, 07-35, 14-01 ×2, 14-29, 06-33, 09-41, 10-06, 16-37) | ARCHIVE | residual ideas consciously declined or covered by standing gates; v0.6.0 questions answered |
> | 2026-07-26 exhaustruct trio (20-07, 20-28, 20-43) | ARCHIVE | tier model landed on the policy table; friction-derived tier systematized the residual decisions |
> | 2026-07-26 erraudit trio (21-24, 21-40, 22-07) | ARCHIVE | erraudit posture decided (manual-review, AGENTS #26); detection contract memo |
> | 2026-07-27 releases (01-25, 01-42) | ARCHIVE | all follow-ups done/declined; five releases shipped since |
> | 2026-08 gohumanize trio + CV pair (03-48, 04-14, 08-58, 01-07, 01-23) | ARCHIVE | gohumanize strategy decided by memo + capability probe; CV learnings all shipped/declined |
> | 2026-09 launch & pareto (02-08, 06-38, 08-33, 09-23, 23-26) | ARCHIVE | decisions closed by the 2026-10-07 memos; remaining work routed to TODO/ROADMAP |
> | 2026-09-28 go-1.27 | ARCHIVE | frontier work shipped in v0.9.0+; residue routed |
> | planning/2026-09-11 SUPERB pareto plan | ARCHIVE | COMPLETE banner (2026-09-14); all 27 task rows struck |

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

All 2026-0* reports were fully resolved by the 2026-11-02 sweep (every
numbered item struck, routed, or consciously declined) and archived to
[`../archive/status/`](../archive/status/). The live reports below are the
2026-10 era onward.

### Declarative Reconciler & v0.11.0 Era (2026-10 → 11)

| Date       | Report                                              | Key Outcome                                                        |
| ---------- | --------------------------------------------------- | ------------------------------------------------------------------ |
| 2026-10-11 | `19-39_gci-removed-from-formatter-force-enable-stack` | gci dropped from CoreFormatters/house                              |
| 2026-10-13 | `16-02_noinlineerr-force-disabled-across-go-appkit`  | noinlineerr forced disabled fleet-wide                             |
| 2026-10-13 | `17-22_buildflow-gate-errors-fixed-standing-warnings-triaged` | Gate errors fixed; 7 warning classes triaged             |
| 2026-10-13 | `17-50_exclusion-sweep-status-and-self-review`      | Exclusion-rules deprecation sweep shipped                          |
| 2026-10-14 | `14-59_declarative-reconciler-execution-status`     | ADR-017 reconciler P1–P4                                            |
| 2026-10-14 | `16-32_declarative-reconciler-p5-closeout-through-p8` | P5–P8 closeout                                                    |
| 2026-10-14 | `18-34_legacy-engine-and-compat-view-deletion`      | Legacy engine deleted; reconciler is the only engine                |
| 2026-10-16 | `19-46_issue-15-capability-gate-session-closeout`   | Capability probe gates recommendations on the pinned binary         |
| 2026-10-16 | `22-50_always-zero-issue-caps-status-and-self-review` | Issue caps always converge to 0                                   |
| 2026-10-17 | `04-16_buildflow-finding-gate-incident-fix`         | mergeNormalizerChanges union fix (check/apply drift dead)           |
| 2026-10-17 | `06-04_v0.11.0-release-and-buildflow-pin-bump-status` | v0.11.0 shipped; cosign pinned; BuildFlow pin bumped              |
| 2026-11-02 | `00-03_upstream-v2-audit-full-support-sweep-status` | Upstream v2 audit: full linter-support sweep                        |
| 2026-11-02 | `00-18_audit-invocations-cqrs-otel-self-review`     | Audit invocations + CQRS/OTel self-review                          |
| 2026-11-02 | `01-19_docs-health-fifth-sweep-all-2026-0x-resolved` | Fifth docs-health sweep: 32 files resolved + archived, living docs rehabilitated |
