# Status Report: Docs-Health Third Sweep — Full 2026-0* Audit, 8 Files Archived, Living Docs Rehabilitated

**Date:** 2026-10-07 02:57 CEST
**Session scope:** Execute the docs-health skill over ALL non-archived `2026-0*` docs (48 files): triage every forward-looking item, annotate + archive fully-resolved files, harvest residue into the living docs, verify all six living docs against code.
**Result:** 8 files fully annotated (~200 inline strikes) and archived with a manifest; 10 new TODO_LIST rows + ROADMAP additions harvested; 2 FEATURES ghost rows and 1 AGENTS version-drift fixed; all quality gates green.

---

## a) FULLY DONE ✅

### 1. All 48 non-archived `2026-0*` docs read and triaged

Five parallel triage passes extracted ~1,500 numbered forward-looking items
(sections f / "Remaining" / "NOT STARTED" / Tier lists / Open questions /
plan tasks) with code-level evidence per verdict: ~60% done, ~15% open,
~10% user-gated, ~15% obsolete/unverified. Files covered: 38 status reports,
1 planning doc (SUPERB pareto plan — already bannered COMPLETE), 3 research
docs, 1 review doc, 1 HTML review.

### 2. Eight fully-resolved files annotated inline and archived

Every numbered item carries a `done at` / `done — <evidence>` / `Won't
implement` / `NOT-DO` / routed verdict; top resolution blockquotes added;
`git mv` to `docs/archive/{status,research,reviews}/`; manifest table written
into `docs/status/README.md` (per the 2026-10-01 bulk-archive rule):

| Archived file                                          | Deciding reason                                                              |
| ------------------------------------------------------ | ----------------------------------------------------------------------------- |
| 2026-07-25_06-30 docs-health annotation pass           | 56/56 §e/§f closed (1 routed: error-code registry → ROADMAP theme 5)         |
| 2026-07-25_06-52 follow-up open questions              | 50/50 §f closed (1 routed); §g answered in-file                              |
| 2026-07-25_17-31 coverage-check cleanup                | 61/61 §c/§f closed (2 declined: OutputConfig typing, CI retry)               |
| 2026-07-25_18-15 comprehensive 50-item execution       | 17/17 remaining rows struck with the in-file commit hashes                   |
| 2026-07-25_20-55 complete 17-item execution            | No forward-looking section; 2 stale claims inline-corrected                  |
| 2026-07-06 pipeline comparison (research)              | 5/5 recommendations resolved (3 shipped, 2 validated)                        |
| 2026-07-25 ecosystem report (research)                 | 7/7 recommendations shipped; residue routed to ROADMAP theme 2               |
| 2026-07-10 deep architecture review                    | 12/12 P0–P3 recommendations shipped; Resolution table verified               |

All annotations ran through the skill scripts (`annotate-rows.py`,
`annotate-prose.py`, `annotate-status-items.py`) with dry-run before every
write; `check-rows.py` reports COMPLETE on all 8; the `grep '~~'`
completeness gate passes for every newly archived file. Live reports: 43 → 38.

### 3. Living docs verified and rehabilitated

- **FEATURES.md**: removed 2 ghost rows — `Auto-tag workflow
  FULLY_FUNCTIONAL` (workflow deleted 2026-09-13) and `CI retry logic
  FULLY_FUNCTIONAL` (no retry step exists in any workflow, verified by grep).
  Fixed stale "21 actions across 4 workflows" → 14 distinct actions across 5
  (counted). Added 4 rows from `[Unreleased]`/recent work: jsondeterminism
  release gate, atomic config writes, CI watchdog, branch/tag rulesets.
  Audit stamp → 2026-10-07.
- **AGENTS.md gotcha 13**: said the CI lint job runs golangci-lint v2.13.2
  "aligned with the devShell"; ci.yml pins v2.14.0. Fixed.
- **ROADMAP.md**: "vendorHash guard is in TODO_LIST" → shipped 2026-09-11;
  cadence bullet updated for this sweep; added `run.go` hardening cluster
  (1.27.0 ≡ 1.27 no-op, go.mod-driven rescue, `--fix` affordance, fuzz,
  doctor output), schema-verify extension to examples/+test configs, BuildFlow
  upstream feedback; routed 2 previously-untracked open questions (cross-repo
  write authorization, homebrew/Scoop tap posture).
- **TODO_LIST.md**: rebuilt — 6 Medium + 13 Low open rows, every row with
  impact/effort/evidence; all 9 report citations verified to resolve.
- **docs/status/README.md**: third-sweep banner + archive manifest; 5 archived
  rows removed from the index; 3 missing live reports added (09-11_23-26,
  09-14_11-34, 09-28_21-59); index now covers 38/38 live reports.

### 4. Quality gates green

`ginkgo -r -p -race` 18/18 suites pass (1m17s); `markdownlint-cli2` 0 issues
on all edited docs; `pkg/constants` + `pkg/types` tests pass (docs-integrity
specs intact); archive completeness gates pass.

---

## b) PARTIALLY DONE ⚠️

1. **Inline strikes in the 38 live reports.** I triaged them but struck
   nothing in live files — the archive candidates got the full treatment.
   Several live reports (07-10, 07-20, the exhaustruct quartet, the erraudit
   trio, the never-enable pair) are 60–90% done per the triage; striking their
   DONE items would expose the true residue. Deferred as lower value-per-
   annotation than the archive work.
2. **Verification depth.** Evidence came from 5 triage agents; I spot-checked
   the load-bearing claims (auto-tag deletion, house preset, coverage-check,
   ROADMAP routing rows, ci.yml retry absence, action counts, schema header)
   but ~30 items the agents marked "unverified" were never individually
   resolved — they are recorded as bare (open) in their files, which is the
   honest state, but the triage debt is real.
3. **Harvest ledger.** The harvest-guide wants a per-item disposition ledger
   in the pass report; I wrote a summary manifest instead of a row-per-item
   ledger (the manifest table covers the 8 archived files, not each harvested
   TODO/ROADMAP row).

---

## c) NOT STARTED

1. **Never-enable residue routing** — `replaceLinters` does not respect
   `never-enable` (fixer_deprecated.go, user-gated g1 from 23-22), no audit
   display test for `ActionSuppressedReEnable`, no own-repo sidecar dogfood.
   Flagged by triage, not yet in TODO_LIST.
2. **erraudit quarterly re-check** — due ~2026-10 per TODO_LIST; noticed,
   tracked, not executed.
3. **A curated session commit** — the daemon produced ~12 heuristic commits;
   I never wrote one deliberate commit message for the sweep.
4. **gohumanize README/3-tier doc items** — user-gated (strategy blocked
   since 2026-08-05); left bare deliberately.

---

## d) TOTALLY FUCKED UP 💀

1. **Malformed multiedit on FEATURES.md shipped a broken intermediate state.**
   My 4th edit was a bad no-op pattern (row + trailing newline) that merged
   the retry row and the Dependabot row onto ONE table line. Caught it on the
   read-back and fixed it in the next edit — but for one commit the table was
   corrupted. Read-back verification saved me; writing the edit correctly the
   first time would have saved the round trip.
2. **Hand-typed script specs bit me twice.** The annotate scripts explicitly
   warn "generate keys mechanically, never from memory" — I hand-typed prose
   specs for 4 files. Two `annotate-status-items` keys failed to match
   (`exportloopref`, `go-finding/pipel`); I burned 2 debug round trips
   (`cat -A` showed a stray trailing backtick I introduced myself) and worked
   around the second with a different substring instead of root-causing the
   mismatch. Dry-run-first is the only reason nothing was corrupted.
3. **Completeness gate misfire.** I ran `grep -rLn '~~'` across the ENTIRE
   archive dirs, producing a 130-line false alarm: ~120 pre-2026-09 archives
   legitimately lack inline markers (banner-annotation era). Should have
   scoped the gate to the 8 new files from the start.
4. **Citation path discipline.** TODO_LIST initially cited 3 reports under
   `docs/archive/` that are still LIVE (`07-27_01-42`, `09-11_06-38`,
   `09-11_08-33`). Fixed with sed — but I wrote the paths from memory right
   after doing the moves; inexcusable given I had the file listing in context.
5. **Wasted a foreground round trip** sleeping 65s to wait for the daemon
   instead of verifying content and moving on.

---

## e) WHAT WE SHOULD IMPROVE

1. **Decide annotate-vs-archive per file BEFORE annotating.** I triangulated
   readiness from triage output, but committing to the 8-file archive set up
   front would have avoided one mid-flight scope wobble.
2. **Always use `--emit-keys`, even for prose lists.** The two key failures
   were both hand-typing artifacts. The mechanical path exists; use it.
3. **Strike DONE items in live reports during sweeps**, not only in archive
   candidates — that is what makes the next sweep cheap (the residue becomes
   visible at a glance).
4. **Keep the harvest ledger as a standing artifact** (per-item disposition
   table), not a prose summary.
5. **Path citations must be copy-pasted from `ls` output**, never typed.
6. **The 2026-06-17 research doc still has an unbannered open item** (the
   `--normalize` timeout flag idea) — research docs should carry the same
   resolution banners status docs get.

---

## f) Up to 50 things we should get done next

### Immediate (this session's direct follow-ups)

1. Run the quarterly erraudit re-check (due ~2026-10; TODO_LIST row exists)
2. Second sweep mode: strike DONE items in the 38 live reports (start with
   the 07-10/07-20 linter-policy pair and the exhaustruct quartet)
3. Resolve the ~30 "unverified" triage items into real verdicts
4. Route the never-enable residue into TODO_LIST (replaceLinters guard,
   `ActionSuppressedReEnable` display test, own-repo sidecar dogfood)
5. Add a per-item harvest ledger to this report (or its successor)
6. Banner the 2026-06-17 cross-project research doc (2 open items remain)

### High-impact tracked work (already in TODO_LIST — verify, don't re-harvest)

7. Fleet sweep: re-run `configure` on ~160 siblings for patched-form `run.go`
8. Refresh committed schema snapshot to dated v2.13.x + `-schema-version` in CI
9. Dogfood gate: `configure --check` on own `.golangci.yml` in CI
10. Release dry-run (`goreleaser release --snapshot`) on PRs
11. Dependabot × vendorHash CI policy decision
12. Buildflow findings-gate posture decision
13. Remove inert `goexperiment.jsonv2` from `.golangci.yml` build-tags
14. Dockerfile slim variant: build as tag or delete the dead block
15. Pin golangci-lint version matrix in e2e (exercise the `run.go` cap path)
16. Scheduled gitleaks job
17. Surface `golangci-lint run` as a flake.nix check output
18. Coverage gate 60 → 65
19. GHCR hygiene (stray `:master` tag + backfill-image fresh-tag smoke)
20. README "Go version handling" section
21. Root-cause `nix flake check` "0 checks" puzzle
22. Decide `Detect()` error-path contract
23. lychee link checker in CI
24. Metadata checklist script

### User-gated decisions (ROADMAP open questions)

25. gohumanize strategy (project-specific vs everywhere; blocked since 08-05)
26. Cross-repo write authorization (8 sibling `min-len` repairs)
27. Homebrew/Scoop tap publishing posture (+ release-page footer cleanup)
28. Git history sanitization (names-only exposure, security dim closed)
29. Daemon-mangled commit messages (rebase vs notes vs leave)
30. Homepage + announcement posture (OSS vs portfolio)
31. Next release cut (`[Unreleased]` is sizeable: Go 1.27 readiness + atomic
    writes + jsondeterminism gate)

### Longer-term (ROADMAP themes — do not promote without refinement)

32. Error-code registry + convention test
33. ConfigReader/Writer narrow-interface adoption
34. Exclusion-merge observability (`--reset-exclusions`, `--show-merged-rules`)
35. Comment-preserving YAML round-trip + `--indent` flag
36. Schema-verify for `examples/` + `test.golangci.yml`
37. BuildFlow upstream language-filter issue
38. Website launch + demo GIF + announcement (post-posture)
39. Error-handling.md / testing-style docs: detection best-effort philosophy +
    /dev/full pattern (still missing from reference docs)
40. detector.go split + detection suite Ginkgo migration
41. SettingsMap helper surface (String/Merge/HasKey/GetString) — build or
    explicitly decline
42. `b.N` → `b.Loop()` modernization (7 live gopls warnings)
43. Audit-CLI surface gap (flock, `--summary`/`--repo`/`--action` filters,
    `--ledger-path`) — build or explicitly decline
44. swall-residue curation: bulk-classify the ~40% of live-report residue that
    looks like YAGNI (fuzz-for-internal-helpers, audit flags, micro-logging)
    as explicit `Won't implement` with reasons — shrinks the live set
45. check-rows.py over live annotated files after any partial strike pass
46. Consider keep-latest-N for `docs/status/` instead of quarterly-only
47. Curated (non-daemon) commit for the next docs sweep
48. Archive-eligible next sweep: the two 09-14 pareto wrap-up reports (once
    the cross-repo + homebrew questions land in ROADMAP — now done, so only
    per-item strikes remain)
49. Verify README badges/links once the next release tags
50. Celebrate: the repo's living docs are, for the first time, fully
    cross-verified against code in a single pass.

---

## g) Questions I cannot figure out myself

1. **Live-report annotation policy:** for the 38 live reports, do you want
   DONE items struck inline NOW (partial annotations exposing residue), or
   should live files stay bare until they are fully archivable in one pass?
   The first is more informative but means files live longer in a
   half-annotated state.
2. **YAGNI curation authority:** a large share of the open residue is
   speculative micro-work (audit-CLI flags, fuzz tests for internal helpers,
   micro-logging). I can bulk-classify these as explicit `Won't implement`
   with reasons (shrinking the live set dramatically), but that closes product
   options. Do you want me to propose the kill-list for your sign-off, decide
   it myself, or leave everything open?
3. **Commit hygiene for this sweep:** the daemon produced ~12 heuristic
   commits for this session. Should I squash them into one curated commit
   (`docs: third docs-health sweep — archive 8 resolved reports, fix ghost
   features, harvest residue`), or is daemon history acceptable? Squashing
   rewrites unpushed public-repo history, so it is your call.

---

## Summary

The third docs-health sweep closed the loop the 09-11/09-14 sweeps started:
every non-archived 2026-0* doc is now triaged with per-item verdicts, the
fully-resolved tail (8 files, ~200 items) is annotated and archived with a
manifest, and the six living docs are cross-verified against code — two ghost
features and one version-drift lie were caught and fixed in the process. The
remaining 38 live reports carry genuinely open residue, now visible and
harvested; the next sweep gets cheaper if live DONE items get struck too.

---

## Postscript (same session, after user decisions)

User answered the three questions: (1) strike DONE items in live reports now,
(2) I classify speculative residue as declined-for-now (vetoable), (3) keep
daemon history. Executed:

- **Live strikes:** 24 live reports fully or partially struck — ~700 inline
  verdicts added (p = verified this pass, w = Won't-implement with reason,
  n = NOT-DO, r = routed with destination, done items struck, open items
  bare). Remaining live reports (09-09, 06-38, 08-33, 09-23, 23-26) carry
  sweep-verification banners instead of item strikes (their residue was
  already routed; item-level strikes deferred as low marginal value).
- **Second archive wave (7 files, all fully resolved/routed):** 17-13 (50/50),
  22-39 (50/50), 23-21 (50/50), 23-22 (50/50), 13-27 (30/30), 09-14_11-22
  (10/10 routed), 09-14_11-34 (30/30 routed). Live reports: 38 → 31.
- **Known scar:** check-rows.py reports one false-positive on the 11-34 table
  separator (all 30 data rows verified struck by direct grep); the routed-row
  → struck-row conversion needed three regex passes (padding + extra cells) —
  hand-rolled conversion is exactly what the "use --emit-keys" rule warns
  about, and it cost 4 extra round trips.
- Citation paths in TODO_LIST re-verified after the moves (9/9 resolve).
