# Status Report: Issue #15 Capability Gate — Session Closeout & Brutal Self-Review

**Date:** 2026-10-16 19:46 CEST
**Scope:** This session only (issue #15 fix + everything noticed along the way). No unrelated research was done, per instruction.
**Repo:** golangci-lint-auto-configure @ master, HEAD `1f958bc` (auto-daemon commits), baseline at session start `b3409b8`.
**Session diff:** 20 files, +1342/−77 (includes pre-existing dirty `vendorHash.nix` that rode along — see d/e).

---

## a) FULLY DONE

1. **Issue #15 root cause fixed — recommendation side.** `Analyzer.ProbeKnownLinters` (`pkg/linter/probe.go`) probes the pinned binary config-independently (temp cwd, cached, fail-open). `Env.BinaryKnown` flows through `CollectEnvBase`; the desired renderer (`shouldRender`, `pkg/linter/desired.go`) withholds any linter the binary cannot name. Stock-binary projects with `dustin/go-humanize` no longer get a poison `enable` entry.
2. **Heal side (the "failure loop" from the issue).** `RemoveUnknownLinters` normalizer (`pkg/linter/normalize.go`) strips unknown names from **both** `linters.enable` and `linters.disable` in phase one, pre-analysis — verified empirically that both lists fail golangci-lint config load. Removal is audited through the existing `AuditActionsFor` projections (`removed-from-enable`/`removed-from-disable`).
3. **Advisory surfacing.** Withheld recommendations land as apply-mode `⚠️` warnings and one `unsupported-recommendation` finding per linter (`pkg/finding/evaluation.go`, new rule ID) — exactly the issue's "emit a finding instead of writing the entry". Verified firing end-to-end at `--priority optional`.
4. **Post-apply tripwire (issue's line of defense #2).** `Analyzer.VerifyConfigLoads` + `verifyWrittenConfig` in `persistConvergence`: after every real save, the fixer re-runs `golangci-lint linters` on the written config and warns loudly on unloadable output. Best-effort, check mode stays the gating surface.
5. **Analysis-failure fallback.** `collectAnalyzerEnabled` falls back to the probe's default-enabled set when `AnalyzeConfig` fails (poisoned config) — prevents bogus re-declaration churn in the very run that heals.
6. **End-to-end proof.** Built binary on a DiscordSync-style repro (go-humanize dep + gohumanize entry): BEFORE `golangci-lint linters` fails with `unknown linters: 'gohumanize'` → AFTER configure: healed, not re-added, config loads, second run idempotent. Dogfood: `configure --check` on this repo's own config = converged.
7. **Tests.** 6 new test files, ~23 new specs/tests: probe parsing/caching/fail-open, capability gate + Unsupported + settings gate, heal normalizer (table), env collection with fake probers, 4 Ginkgo e2e specs reproducing the issue (auto-skip on a custom binary that carries gohumanize), finding conversion specs. Full suite: 18 suites, exit 0. Constants/errors gates green (incl. the codes-convention gate after registering `probe.tempdir` + `evaluation.unsupported_finding`).
8. **Docs.** AGENTS.md: new gotcha 8 (capability gate), gotchas 6/7 corrected (legacy engine already deleted; settings now ride the policy table — old text referenced deleted `injectDefaultSettings`), Detail #1 rewritten (two-gate gohumanize story). FEATURES.md: gohumanize row updated + new capability-probe row. Policy-table rationale de-lie-ified (it promised H001/H002/H003/H007 settings that never existed).
9. **Lint hygiene on touched packages.** Fixed everything I introduced (funlen, gci, golines, nonamedreturns, wsl_v5, nolintlint, ginkgolinter, lll). Scoped lint on `pkg/linter` + `pkg/finding` + `pkg/errors` is clean except pre-existing lll in untouched files.

## b) PARTIALLY DONE

1. **Repo-wide lint state.** My packages are clean, but repo-wide `golangci-lint run` reports **45 pre-existing lll findings** (pkg/errors/templates.go, pkg/utils/retry.go, internal/cli flags, etc.). I did NOT check whether the CI lint job is red because of them. Scope-called, but unverified.
2. **The deadcode/gas/execinquery fail-open hazard.** Verified via debug run: with NO binary (fail-open), `Desired()` still renders removed-in-v2 records (`deadcode`, `gas`, `execinquery`, … — they have no `ValidTool.Max` bounds). The capability gate masks this whenever a binary is present, but the no-binary path keeps the exact poison class the issue describes. Analyzed, deliberately not fixed this session.
3. **Formatter-side capability coverage.** `RemoveUnknownLinters` covers linter lists only; there is no `formatters` probe and no `RemoveUnknownFormatters`. Whether any formatter policy record (e.g. `swaggo`) names a formatter absent from stock binaries is **unverified**. The post-apply tripwire would catch the breakage; the fixer would not heal it.
4. **Issue #15 lifecycle.** Fix implemented and committed locally (by the auto-daemon); issue remains OPEN, the "MANUALLY REVIEWED" checkbox unticked. No push (harness rule), no close comment posted yet.
5. **Pre-existing evaluate-vs-apply divergence.** `PlanDeviations` reports 45 enable-adds on configs that `configure --check` calls converged (pre-change: 66 — my gate shrank it). Diagnosed, measured, explained in the final message — NOT filed as an issue, NOT fixed.
6. **Docs surface coverage.** AGENTS.md + FEATURES.md updated; README.md (user-facing) and docs/DOMAIN_LANGUAGE.md (new terms: capability probe, BinaryKnown, Unsupported) not touched. CHANGELOG.md has no entry for the behavior change.

## c) NOT STARTED

1. Filing the evaluate-vs-apply divergence as a GitHub issue (evidence is in this session's transcript).
2. Investigating/fixing the fail-open removed-linter hazard (b2).
3. Formatter capability probe + `RemoveUnknownFormatters` (b3), if b3's premise verifies.
4. CHANGELOG.md entry, README note, DOMAIN_LANGUAGE.md terms.
5. `docs-health` HARVEST of section (f) into TODO_LIST.md/ROADMAP.md.
6. `nix flake check` / `nix build` — never ran this session (go build + go/ginkgo tests green; treefmt untouched area). The pre-existing `vendorHash.nix` change sitting in the tree is therefore **unvalidated by nix**.
7. `scripts/e2e-pin-matrix.sh` re-run (per-version invariants after a constants-table rationale change — rationale-only, low risk, still unproven).
8. CI status check (`gh run list`) — unknown whether the pre-existing lll findings already fail the lint workflow.
9. Push + squash decision + issue #15 close with repro evidence.

## d) TOTALLY FUCKED UP

**Nothing at this tier — the shipped work is green, verified, and coherent.** Honest near-misses, all caught and corrected in-session:

1. **Mid-list numbering churn in AGENTS.md.** I inserted the new gotcha as #8 in the middle of a 1–30 list, then wrote a renumber script that **crashed on its own assertion**, leaving duplicate numbering that a second pass had to fix. Should have appended at the end from the start.
2. **Mangled edit in normalize.go.** A careless `edit` ate a newline and glued a function signature to its first statement. Caught by immediate re-view, fixed in the next edit.
3. **Wrong-file edit.** Applied the ginkgolinter fix (issueConfig const) to `desired_capability_test.go` instead of `capability_e2e_test.go`; had to revert immediately.
4. **Wasted debugging round on the "missing" CLI warning.** 4–5 tool calls chasing why the CLI didn't emit the gohumanize warning. Root cause: `configure`'s priority default (`high` in `commands.go:237` vs `optional` in `cmd_configure.go:109` — a discrepancy I had *already seen in grep output* and not registered) filters Medium-priority gohumanize before the capability gate ever runs. I should have checked flag defaults first.
5. **Stale-binary confusion.** Part of the same chase: suspicion that `/tmp/glac-bin` was stale added noise. Lesson: rebuild before re-testing CLI behavior.
6. **History smear (not my act, but my problem to flag).** The auto-commit daemon split the feature across ~12 anonymous `chore: auto-commit N file(s)` commits. The feature is unreviewable from history; only a user-sanctioned squash fixes that now.

## e) WHAT WE SHOULD IMPROVE

1. **Defense in depth for removed linters:** derive `ValidTool.Max` from `SupersessionReason` at table init (or add explicit bounds) so fail-open mode can't render `deadcode`/`gas`/`execinquery`. Today the capability gate is the ONLY filter, and it is off exactly when there is no binary.
2. **Verify-before-trusting tool defaults:** when behavior differs between two entry points, diff their FLAG DEFAULTS first (the priority-default split cost an hour).
3. **Whole-function edits via `lsp_replace_symbol`**, not hand-matched `edit` — both mistakes this session (mangled line, wrong file) were hand-match failures.
4. **File issues the moment they're diagnosed.** The evaluate-vs-apply divergence has full evidence sitting in a chat message that no future session will read.
5. **End every CLI-behavior claim with a rebuilt binary.** Go test binaries and stale `go build` outputs diverge silently.
6. **Observability for silent fail-opens.** `collectBinaryKnown` returns nil on probe failure with no log; a debug line ("probe failed, capability gating off: …") would make fleet debugging possible.
7. **Warning volume control:** Unsupported warnings currently fire per record per run (15/run in fail-open-ish tables); consider dedupe/run-summary once item e1 lands.
8. **Probe robustness:** no explicit timeout beyond ctx on the probe subprocess; wire-shape drift (`linters --json`) silently yields fail-open. A canary fixture test + explicit timeout would harden it.
9. **Update TODO_LIST.md in the same change** when a tracked TODO's subject is touched (the legacy-deletion item referenced by gotcha 6 was verified done but left ticked-off nowhere — instant doc drift).

## f) UP TO 50 THINGS WE SHOULD GET DONE NEXT

Sorted by impact; first ~12 are real work, the rest are brainstorm/ROADMAP fuel (docs-health HARVEST should apply routing rigor).

1. **Derive `ValidTool.Max` from `SupersessionReason` at table load** (kills the fail-open deadcode/gas hazard at the root; e1). *(HIGH)*
2. **Verify + fix the pre-existing `vendorHash.nix` modification** sitting in the tree (buildflow-managed hash changed; I never validated it with `nix build` — could break every clone if stale). *(HIGH)*
3. **Check CI lint workflow status** (`gh run list`) against the 45 pre-existing lll findings — if red, that's a broken gate predating this session. *(HIGH)*
4. **File the evaluate-vs-apply divergence issue** (PlanDeviations 45 adds vs `--check` converged; evidence + repro commands in session transcript). *(HIGH)*
5. **Investigate the dual priority-flag default** (`commands.go:237` "high" vs `cmd_configure.go:109` "optional") — which is the real `configure`, and which default is intended? Changes whether Medium linters are ever recommended at defaults. *(HIGH)*
6. **Push + close issue #15** with the before/after repro evidence (user-gated). *(HIGH)*
7. **Squash the ~12 daemon commits** for this feature into one reviewable commit (user-gated). *(HIGH)*
8. **Formatter capability coverage:** verify whether `swaggo` formatter is bundled in stock binaries; if not, add a formatters probe + `RemoveUnknownFormatters`. *(HIGH if premise verifies)*
9. **Spec: `verifyWrittenConfig` warning fires** when the post-write config is unloadable (only the success path is pinned today). *(MED)*
10. **Spec: heal removals land in the audit ledger** as `removed-from-enable`/`removed-from-disable` (reasoned, not pinned). *(MED)*
11. **CHANGELOG.md entry** for the capability gate + heal behavior change. *(MED)*
12. **Run `nix flake check`** (format + build + tests gate) — the only house quality gate not executed this session. *(MED)*
13. Harvest section (f) into TODO_LIST.md / ROADMAP.md (docs-health HARVEST). *(MED)*
14. Close/remove the stale TODO_LIST.md item about legacy-engine deletion (verified deleted this session; AGENTS.md now says so — TODO_LIST may still track it). *(MED)*
15. README.md: user-facing note (plugin linters need a custom binary; new warning semantics). *(MED)*
16. Run `scripts/e2e-pin-matrix.sh` post-rationale-change. *(MED)*
17. Finding-report (HTML/JSON/SARIF) coverage spec for `unsupported-recommendation` (report path currently only indirectly exercised). *(MED)*
18. DOMAIN_LANGUAGE.md: capability probe / BinaryKnown / Unsupported / heal terms. *(LOW)*
19. Debug-level log in `collectBinaryKnown` on probe failure (observability). *(LOW)*
20. Run-summary or dedupe for Unsupported warnings (noise control at fleet scale). *(LOW)*
21. Explicit timeout guard on the probe subprocess (hung binary must not hang configure). *(LOW)*
22. Canary test pinning the `linters --json` wire shape (drift → silent fail-open today). *(LOW)*
23. Windows story for probe tests: fixture binary is a `#!/bin/sh` script — guard with `runtime.GOOS` or skip on Windows CI. *(LOW)*
24. Probe result info-log once per run ("probed golangci-lint 2.14.0: 115 linters known") for fleet debugging. *(LOW)*
25. Measure configure runtime delta from the +2 golangci-lint invocations (probe + verify) on a large fleet. *(LOW)*
26. Cross-check `RemoveUnknownLinters` ordering vs deprecation replacement in the normalizer chain (both mutate enable lists; interplay untested). *(LOW)*
27. Consider capability-gating `renderFormatterSet` symmetrical to linters (if item 8's premise verifies). *(LOW)*
28. Unify/document the duck-typed capability interfaces (`linterProber`, `configLoadVerifier`) beside `ledgerReader` in one "optional analyzer surface" doc note. *(LOW)*
29. Document the Unsupported projection chain (Desired → ConvergeResult → Evaluation) so the three carriers don't drift. *(LOW)*
30. Split `TestRemoveUnknownLinters` further if it grows (gocognit borderline at rewrite time). *(LOW)*
31. CLI-level e2e spec: `configure --check` exit code reflects the heal (handleCheckMode integration). *(LOW)*
32. Verify the Ginkgo `Skip` in capability_e2e actually triggers on a plugin-carrying binary (can't easily produce one; at least document how). *(LOW)*
33. Upstream suggestion to golangci-lint: make `config verify` catch unknown linter names (would obsolete this bug class at the schema layer) — verify-before-filing first. *(ROADMAP)*
34. Evaluate whether `Unsupported` should include below-threshold priorities in reports (product decision; today only at/above threshold). *(ROADMAP)*
35. Persist probe results across fleet runs (per-binary cache keyed by binary hash) if item 25 shows real cost. *(ROADMAP)*
36. Table-lint rule: policy records with `SupersessionReason` containing "removed" must carry `ValidTool.Max` (data-integrity spec so item 1 never regresses). *(MED — fold into item 1)*
37. Check whether `docs/DOMAIN_LANGUAGE.md` needs the gate vocabulary before external users see the finding rule IDs. *(LOW)*
38. Add the issue's `MANUALLY REVIEWED` checkbox tick to the #15 close checklist (user act). *(LOW)*
39. LSP-vs-CLI lint drift in dev experience: stale golangci_lint_ls diagnostics showed ghosts of fixed issues; consider `lsp_restart` after bulk edits or a documented note. *(LOW)*
40. Post-close: retro-annotate the 2026-10-14 status reports (docs-health ANNOTATE) that reference the pre-fix gohumanize behavior. *(LOW)*

## g) QUESTIONS I CANNOT FIGURE OUT MYSELF

1. **Push / close / squash sequencing:** Should I push and close #15 with the repro evidence now, or do you want to squash the ~12 auto-daemon commits into one clean, reviewable commit first? (Determines what the upstream issue links to.)
2. **`configure` priority default intent:** `commands.go:237` registers `--priority` default **"high"**, `cmd_configure.go:109` registers **"optional"**. Which command is the canonical user-facing `configure`, and which default is correct? (This decides whether Medium linters like gohumanize are recommendable at defaults at all — and whether my CLI warning is reachable without an explicit flag.)
3. **`vendorHash.nix` provenance:** The file was already dirty at session start (`sha256-zg+6…` → `sha256-YmWG…`, "Managed by buildflow nix-checker") and got committed by the daemon mid-session. Did you/buildflow change it intentionally, and should I validate it with `nix build` before this branch goes anywhere?

---

*Report format note: skill default is a styled HTML dashboard; the explicit `.md` path in your instruction wins (flagged per the skill's override rule). Section (f) exceeds the skill's Top-25 default per your "up to 50". Manual commit skipped per harness rule — the auto-commit daemon will pick this file up. HARVEST of (f) into TODO_LIST.md is pending your go-ahead since you asked me to wait.*
