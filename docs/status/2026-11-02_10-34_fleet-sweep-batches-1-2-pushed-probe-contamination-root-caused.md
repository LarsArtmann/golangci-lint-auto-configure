# Status: Fleet Convergence — T3/T11/T12-clarity landed, fleet sweep 20+3 repos pushed (batch 2 of 10), probe-contamination root cause found

**Date:** 2026-11-02 10:34
**Session:** Fleet-convergence execution (plan: `docs/planning/2026-11-02_03-11_SUPERB-fleet-convergence-and-v2-15-frontier.md`)
**Session paused mid-sweep for course decisions — see §g.**

## Executive summary

The SUPERB fleet-convergence plan is roughly 40% executed. All cheap tool-hardening
predecessors to the sweep (T1/T2/T3) are landed and CI-green. The fleet sweep (T5) is
**in flight: batches 1–2 of 10 pushed (23 sibling repos touched), 101 repos remaining**,
paused for three course decisions (§g). The session's biggest finding is a **falsified
upstream claim**: the "golangci-lint v2.15.0 lists every linter bare-dir" behavior
recorded in gotcha 8 never existed upstream — it was `/tmp/.golangci.yml` (an enable-108
experiment config) leaking into our probe via golangci-lint's upward config discovery.
The probe is now hermetic (root-cause fix, uncommitted), the T12 upstream issue was
correctly **not filed** (verify-before-filing gate caught the premise error), and the
stray config is trashed.

## a) Fully done (verified)

| Item | Evidence |
| --- | --- |
| D0 master green (recovered 3-way interleaved sessions, landed dead sessions' work) | CI green through `1cff856` |
| T1 stale project-specific linter retraction | `89a87b3`, pushed; fleet prediction confirmed: ~98/117 drifted configs carried arangolint/clickhouselint |
| T2 orphaned-settings prune + curated wrapcheck exemption | `1cff856` (landed by D0 recovery) |
| T3 declarative regression-loop detection + never-enable offers | `634b904` pushed; specs green (`desired_cycle_test.go`, fixer_test cycle context); CI green |
| T7/T8/T9/T10 v2.15 frontier (schema snapshot, version pins, wrapcheck, chain fixture) | `ef60545` + fix; dogfood gate green via co-tenant's `c01d745` |
| F10 README durable-suppression note + F32 version-claims sweep | `90efec4` pushed |
| T11 single-criterion validation version pin | Verified against upstream source: `excludeRuleMinConditionsCount = 2` present since ≥v1.38.0 (2021) and in every v2 release; CHANGELOG/spec/AGENTS updated (uncommitted with probe work) |
| T12 verify-before-filing outcome: DO NOT FILE | Premise falsified (see §d); no upstream issue filed; this is the skill working as designed |
| Fleet sweep Phase A+B: full classification of 175 config-bearing siblings | 143 eligible / 32 skipped-with-reasons; 100 drifted, 22 converged, 17 config+fmt, 4 fmt-only |
| Fleet sweep batches 1–2 + 3 schema heals | 23 repos pushed green-verified: accounting-pitch-deck ActaFlow ai-router-implementation AI-Speed-Test ai-task-prioritizer AI-von-Art-Bench architecture-diagrams ArtCode AutoCart auto-deduplicate bank-sync BerryBig branching-flow BuildFlow chats Code-To-CV-Agent cqrs-htmx (+heals on ai-router-implementation, AI-Speed-Test, auto-deduplicate) |

## b) Partially done

- **T5 fleet sweep (F16–F22):** batches 1–2 pushed (23/124 target repos). Remaining:
  80 drifted repos (batches 3–10), 17 config+fmt repos, 4 fmt-only repos, and 3
  verify-failures pending heal (blog, code-duplicate-analyzer, Code-Quality-Agent —
  same stale-v1-key class as the healed three: interfacebloat/forbidigo/exhaustive
  v2.15-shape renames).
- **Probe hermeticity fix (T12 root cause):** implemented + specs green
  (`TestProbePinsMinimalConfig`), docs corrected (AGENTS gotcha 8, CHANGELOG bullet),
  **uncommitted** — blocked on my commit discipline (see §e) and the daemon publishing
  co-tenant interleaves.
- **T6 fleet report:** data is accumulating in `/home/lars/.cache/glac-fleet/`
  (results.tsv, push-results.tsv, diffs/, skips.txt) but the report is unwritten.

## c) Not started

T4 (curated-default refresh with provenance, F11–F15), T14 (cosign v3 + unpin),
T16 (floating-installer audit), T17 (release runbook fold-in), T19 (Dependabot policy),
T15/T18 (USER-GATED artifact/GHCR work), T20 (v0.12.0 cut), T21 (docs tail: markdownlint
nested-config, FEATURES re-audit, table padding), T22 (decision packets F69–F76),
T13 (probe era fixtures — re-scoped by the contamination finding, see §e).

## d) Totally fucked up (caught, contained, or corrected)

1. **Gotcha 8's v2.15.0 claim was false** (inherited, now corrected): "108 of 115
   Enabled bare-dir" was `/tmp/.golangci.yml` discovered by golangci-lint's upward
   config search from our probe's temp cwd. Reproduced both ways (clean parent →
   Enabled=5 like v2.14; under /tmp → 108). The shipped >half-vocabulary guard was
   the safety net that correctly fired; docs now tell the true story; stray config
   trashed. **Consequence:** T12's upstream issue must never be filed as drafted —
   verify-before-filing Gate 1/4 failed exactly as designed.
2. **First Phase-B pass: 129/143 CLONE_FAILED** — /tmp tmpfs (48G) was full from other
   sessions' artifacts. Relocated the sweep workspace to disk-backed
   `~/.cache/glac-fleet/`. Lesson recorded: check `df` before mass-clone operations.
3. **My diff classifier had two bugs** (embedded newlines in gate rows; retraction
   grep matching exclusion-rule linter lists). Both fixed; final classification was
   hand-verified via a full-diff anomaly scan + representative full reads.
4. **CI went red once mid-session** (a1a2cc3): co-tenant's landed internal/cli work
   had 3 funlen violations. Before I touched it, the co-tenant session woke and fixed
   it themselves (`8bebd80`) — CI green again. No double-fix shipped.
5. **Sibling pre-commit hooks fired on sweep pushes:** BuildFlow clones/repos run
   heavy pre-commit pipelines on commit (cqrs-htmx: 1m51s pipeline; BuildFlow clone's
   build gate fails in clones due to missing `../go-sibling` replace dirs). Pushes
   still succeeded, but this slows batches and can fail future commits — needs a
   decision (§g Q1).

## e) What was stupid / what I'd do differently

- **Trusted gotcha 8 for hours** (built the whole T12 filing plan on it) before
  reproducing the probe myself. Inherited claims from prior sessions are evidence,
  not truth — the 03-36 lesson repeated. The falsification cascade (probe → docs →
  planned upstream issue) was caught only because verify-before-filing demands
  source-level reproduction.
- **Built the sweep binary and clones inside /tmp tmpfs** without checking free space
  (129 wasted clones). Should have started disk-backed.
- **Underestimated sibling pre-commit hooks** in the push protocol — the memo's
  protocol didn't mention them; discovered live.
- **Improvements found:** (1) probe hermeticity (done, uncommitted); (2) the yaml
  writer loses comments and flips quote style on rewrite (fleet-visible: wrapcheck
  exemption comments dropped, `"` → `'` churn) — candidate tool improvement;
  (3) a `configure --check --format finding` mode that actually emits findings would
  have made the pre-scan scriptable (the flag is analyze-only today);
  (4) stale-v1-key heals (interfacebloat/forbidigo/exhaustive) suggest a future
  `KnownBadSettingsKeys`-adjacent normalizer for v2 configs carrying v1 keys —
  needs the F73 policy decision first.

## f) Next things (ordered, ~50)

**Sweep completion (T5/T6):**
1. Heal 3 verify-failures (blog, code-duplicate-analyzer, Code-Quality-Agent) — same recipe as the healed three
2. Batches 3–10: 80 drifted repos (10/batch, spot-verify 5 diffs each)
3. 17 config+fmt repos (two-commit split: config / format)
4. 4 fmt-only repos (clean-wizard, emeet-pixyd, go-datastar, go-taskqueue)
5. F22: fold in 00-03 f9 grep scan (deadcode/varcheck/structcheck/exhaustivestruct/interfacer/nosnakecase counts)
6. F23: fleet convergence status report (per-bucket totals, scars, skip list)
7. F24: harvest into TODO/CHANGELOG/FEATURES; update fleet-audit numbers
8. F25: cross-repo audit-ledger story note
9. Re-run `configure --check` pre-scan post-sweep to prove fleet-wide convergence
10. Record the sibling-behavior surprises (pre-commit hooks, quote-style churn) in the fleet report

**Tool truth + tail of the frontier:**
11. Commit the hermetic-probe work (probe.go + tests + AGENTS/CHANGELOG corrections)
12. T13 re-scoped: pin the guard spec against BOTH eras + the contaminated shape (rename fixture semantics, add a hermetic-probe integration spec)
13. Re-verify the "21 broken configs" fleet-audit number against sweep data (verify-failures so far: 6)
14. Upstream-adjacent: file the REAL upstream issue candidate — golangci-lint v2.15.0's jsonschema file is stale vs its binary (exhaustive settings) — verify-then-file properly
15. Consider yaml comment preservation (big; needs design)
16. Consider quote-style preservation in the yaml writer (smaller)
17. `configure --check --format finding` wiring or docs fix
18. interfacebloat/forbidigo/exhaustive v1-key heal: decide policy (F73 packet) then maybe implement as normalizer

**T4 (curated refresh with provenance):**
19. F11 design `# managed-by:` stamp
20. F12 stamp write + comparison helper
21. F13 refresh pass (stamp-matched → current curated; diverged untouched + audited)
22. F14 specs (round-trip, tuning preservation, --force-settings wins)
23. F15 CHANGELOG + AGENTS gotcha

**Release/ops trust cluster:**
24. T14/F42 cosign v3 signs migration in .goreleaser.yaml
25. F43 unpin cosign-release; goreleaser check + local snapshot sign
26. F44 release-dry-run exercises sign step
27. F45 CHANGELOG + release-process.md note
28. T16/F50 inventory latest-resolving workflow steps
29. F51 pin installers; runbook table
30. F52 action-SHA bump automation decision
31. T17/F53 fold tag-failed runbook into release-process.md
32. F54 extend post-release-verify.sh (signatures, SBOMs, cask, GHCR)
33. F55 dry-run extended script against v0.11.0
34. T19/F58 Dependabot × vendorHash policy (prefer continue-on-error on dependabot actors)
35. F59 implement + watch one scheduled run green
36. T15/F46-F49 (USER-GATED) cosign sign-blob backfill 12 assets, SBOMs, release-note caveat, GHCR smoke
37. T18/F56-F57 (USER-GATED) GHCR :master deletion + backfill-image throwaway-tag test

**v0.12.0:**
38. F60 CHANGELOG versioned section, FEATURES stamp, TODO prune
39. F61 pre-release-check.sh (TMPDIR-safe)
40. F62 full gates: ginkgo -r -p -race, lint, nix build, dogfood --check, vendorhash guard
41. F63 tag + push + watch pipeline + curated notes
42. F64 post-release verify + clean go-get proxy check

**Docs/truth tail:**
43. F65/F66 markdownlint nested-config investigation + fix or upstream issue
44. F67 FEATURES full re-audit (earn 2026-11-02 stamp)
45. F68 restore going-public b/c table padding (cosmetic)

**Decision packets + calendar (T22):**
46. F69 buildflow findings-gate posture packet
47. F70 M32 never-enable ratification packet
48. F71 AGENTS.md size posture packet
49. F73 KnownBadSettingsKeys scope packet (now enriched by the sweep heals)
50. F74 flake.nix golangci-lint check output; F75 upstream skill-tooling check-rows fix; F76 erraudit ~2027-01 calendar note; F72 declined-verdict owner-review grep

## g) Questions (max 3)

1. **Sweep continuation with sibling pre-commit hooks live?** Pushes into siblings
   trigger their BuildFlow pre-commit pipelines (slow; one build gate fails in
   clones by design). Continue as-is (hooks run, slower), or commit with
   `--no-verify` in the clones (faster, skips their gates — their CI still runs
   post-push)?
2. **Do T15 (cosign/SBOM backfill on the published v0.11.0 release) and T18 (delete
   the stray GHCR `:master` tag) fall under this session's blanket "do everything"
   order, or do you want to pull those specific triggers yourself?** They mutate a
   shipped release and delete a package tag.
3. **v0.12.0 timing:** cut right after the sweep + T4 (shipping fleet-value), with
   T14–T19 landing post-release? Or hold the tag until the whole trust cluster is in?

## Handoff state

- Working tree: hermetic-probe work + T11 doc updates **uncommitted** (probe.go,
  probe_internal_test.go, normalize.go, fixer_deprecated_exclusions_test.go, AGENTS.md,
  CHANGELOG.md). Co-tenant session ACTIVE (their `8bebd80` funlen fix at 09:12; their
  10-17 status report untracked). Commit pathspec-scoped when resuming.
- Sweep workspace: `/home/lars/.cache/glac-fleet/` (inventory, results.tsv,
  push-results.tsv, diffs/, clones/, skips.txt, phaseA.sh, phaseB.sh, batch_push.sh).
- Machine: /tmp tmpfs is FULL (other sessions); use `~/.cache` paths and
  `GOCACHE=/home/lars/.cache/go-build-crush`. Load average was 136 — schedule heavy
  work thoughtfully.
