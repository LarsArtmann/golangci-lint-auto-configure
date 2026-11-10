# Status: Fleet-Convergence Execution — D0 Green, T1/T2 Landed, T3 Mid-Flight

> **Scope:** This session only — executing the fleet-convergence plan of
> record (`docs/planning/2026-11-02_03-11_SUPERB-fleet-convergence-and-v2-15-frontier.md`,
> 22 master tasks / 76 fine tasks) under the "whole TODO list" standing
> order. Point-in-time snapshot: 2026-11-02 06:09 CET.

---

## Foreword: What did I forget / could do better / still improve?

**Forgot:**

- **Verify the STAGED snapshot, not the last-tested tree.** Commit ef60545
  shipped the 03-36 session's structural migrate code that landed AFTER my
  last lint run — 7 lint findings + an all-fixtures parity failure went to
  master. `git add -A` blindly staged another session's mid-flight work
  between my verification and my commit. The recovery cost one extra
  push/CI cycle and the fix batch became 1cff856.
- **89a87b3 (T1) is committed but NOT pushed** — session paused for this
  report before pushing; CI has not seen it.
- The full 19-suite race run has not re-run since T1 landed (scoped
  pkg/linter + pkg/audit suites only).

**Could do better:**

- Never trust a dead session's status report over a live test run: the
  03-36 report claimed the version specs "reference the constant
  dynamically — no hardcoded value to chase" — `version_checker_test.go`
  hardcoded v2.14.0 and failed.
- One multiedit on converge.go fired a wrong replacement (edit matched
  planConvergence, wrote a bogus body). Caught by reading the result
  immediately; still a wasted cycle. Batch edits on files under concurrent
  modification deserve one-at-a-time discipline.
- Push cadence: I stacked a new push while the previous run was still red —
  the two failure sets merged and diagnosis got noisier. Push after the
  previous run's verdict whenever possible.

**Still improve:**

- The auto-commit daemon publishes ALL local commits with any push —
  "exclude in-flight files from my commit" is not a real containment
  strategy. The working posture that survived: make the entire tree green,
  commit everything coherently, push.
- This timeline's Go 1.27 lacks `context.WithTimeoutWithoutCancel` —
  compose `WithTimeout(WithoutCancel(ctx), d)` instead (noted for future
  sessions; cost one compile round).

## a) FULLY DONE

1. **D0 — master green (CI: all 3 workflows pass on 1cff856).** Landed the
   dead sessions' completed work after making it green: v2.15.0 baseline
   (schema pins, `ExpectedGolangCILintVersion`, four CI action pins),
   complete v1→v2 migrate depth (renames, type changes, structural
   sections: `issues.exclude`, `exclude-use-default`, skip-dirs defaults,
   `output.format`), the `exhaustivestruct-chain` parity fixture, the
   OTEL session's M2–M4 (`pkg/telemetry`, CLI wiring, FixConfig span),
   and the wrapcheck 2.13 self-lint fix. Four curated commits:
   ef60545, f0726ef, f3e8801, e17261e; recovery batch 1cff856.
2. **T2 — orphaned-settings prune liveness guard** (in 1cff856). Settings
   for linters that RUN via the standard default set (`unused`,
   `ineffassign` under `linters.default: standard`, the v2 default) are no
   longer pruned as dead weight; `default: none` behavior byte-identical;
   dead exported `PruneOrphanSettings` deleted. Failing-spec-first e2e +
   regression pin.
3. **T7 + T8 + T10 — landed** (by the 03-36 session, verified and shipped
   by me): schema snapshot v2.15.0 + drift-guard alignment, version
   frontier (constant + CI pins; version-checker specs now derive from
   constants), exhaustivestruct parity fixture.
4. **T1 — stale project-specific linter retraction** (89a87b3, local).
   `RemoveStaleTechGatedLinters` normalizer (arangolint/clickhouselint/
   gohumanize whose tech is definitively absent), audited as
   `retracted-project-specific` (reason-prefix mapping), tech-unknown
   fails open, disable entries untouched. Unit table + e2e specs;
   tool-own parity golden regenerated with EXACTLY the two retractions —
   the tool's own config is the first healed fleet member.
5. **Curated wrapcheck errorfamily default promoted** (in 1cff856, third
   session's work verified and shipped): org-wide
   `github\.com/larsartmann/([^/]+/pkg/errors|go-error-family)\.` rides
   the wrapcheck policy record; goldens + schema fixture + BDD spec
   regenerated; CHANGELOG + AGENTS updated.
6. **Living docs kept truthful as work landed**: CHANGELOG Unreleased
   (v2.15 baseline, migrate depth, curated wrapcheck, retraction, prune
   guard, OTEL, wrapcheck self-lint), AGENTS gotchas 7/8/14/20 +
   Where-to-Find-Detail 1/6 refreshed. markdownlint 0/62.

## b) PARTIALLY DONE

1. **T3 — declarative regression-loop detection + durable suppression
   offers.** Implemented but UNVERIFIED: `WithCycleSuppressed` renderer
   option + `DesiredConfig.CycleSuppressed` + 3-return `renderEnableSet`;
   `ConvergeOptions.CycleSuppressed`, `ConvergeResult.CycleSuppressed`,
   5-return `planConvergence`, audit loop + `cycleSuppressedReason`;
   fixer `ledgerReader` duck-type + `previouslyAutoEnabled` +
   `cycleSuppressionCandidates`; `warnCycleSuppressed` offer block (the
   copy-paste never-enable sidecar snippet, non-interactive, nothing
   auto-written); evaluation.go call sites updated. MISSING: build+test
   verification (last build was mid-edit), captureRecorder needs the
   `PreviouslyAutoEnabled()` method + e2e spec (ledger shows suppressed
   re-enable; config stays free of the linter), renderer unit spec,
   userEnabled-wins case, docs (CHANGELOG/AGENTS gotcha 2), commit.
2. **T9 — wrapcheck skew.** Self-repo fix landed AND the fleet decision
   got absorbed by the curated default (better than the plan's "document
   the pin" option). Remaining: the 03-36 session's suggested UPSTREAM
   issue about wrapcheck 2.13 in-module cross-package reporting (folded
   into T12's verify-then-file queue, optional).
3. **A parallel session is LIVE** in `internal/cli/` + `pkg/migration`
   (their OTEL plan M4-tail/M5/M6/M7: migrate ctx threading, audit
   coverage, flag-error exit codes, exit-code pins; 10 files, 238
   insertions in flight at 05:45). Their lint debt is theirs; my commits
   are pathspec-scoped around it.

## c) NOT STARTED

- **T5 fleet sweep** (~160 siblings; the 1%→51% item) + **T6 fleet report**.
- **T4** curated-default refresh with provenance (hash-stamped blocks).
- **T11** single-criterion validation version pin; **T12** upstream v2.15
  bare-dir issue (verify → file, github-voice); **T13** probe-era fixtures.
- **T14** cosign v3 signs migration + unpin; **T16** floating-installer +
  action-SHA audit; **T17** release runbook fold-in + extended
  post-release-verify; **T19** Dependabot vendorHash policy.
- **T15** v0.11.0 artifact backfill + **T18** GHCR hygiene (USER-GATED).
- **T20** cut v0.12.0; **T21** docs/truth tail (markdownlint nested-config
  probe, FEATURES re-audit stamp, going-public table padding);
  **T22** decision packets (buildflow posture, M32, AGENTS size,
  declined-verdict review, KnownBadSettings scope, flake check output,
  check-rows upstream fix, erraudit calendar note).
- Dogfood: run the new binary's `configure` on THIS repo to land the two
  real retractions (arangolint, clickhouselint) the parity golden now
  predicts.

## d) TOTALLY FUCKED UP

1. **Second red master of the day (my push ef60545..e17261e at 04:20):**
   lint job 7 findings in unlinted structural-conversion code + parity
   failing on ALL fixtures (I had staged regenerated goldens while the
   constants that produce them were NOT staged — golden/engine mismatch).
   Fixed in 1cff856 (green 04:35). Root cause: verified the wrong
   snapshot, staged blindly, pushed without watching the prior run.
2. **Botched multiedit on converge.go during T3** (wrong-target
   replacement mangled planConvergence; repaired by re-reading). No
   lasting damage; discipline note recorded above.
3. **The `.golangci.yml` v1-experiment left by the dead OTEL session**
   broke every local lint/LSP invocation (exit 3 / "unsupported version
   1") until I restored HEAD's config (experiment preserved at
   /tmp/golangci-v1-experiment-backup.yml). Cost several confusing
   diagnostic cycles before I connected the symptom to the file.

## e) WHAT WE SHOULD IMPROVE

1. **Snapshot discipline**: the verification gate must run against the
   exact staged/committed content (`git stash -k`-style or verify after
   staging), not "the tree as of N minutes ago" in a multi-session repo.
2. **Multi-session co-tenancy protocol is implicit — make it explicit**:
   a tiny AGENTS section ("verify AFTER staging; never `git add -A`
   another session's in-flight files without running gates over them")
   would have prevented d1.
3. **Dead-session reports are evidence, not truth**: every claim that
   gates a push (specs dynamic, tests green) gets re-run, never trusted.
4. **A canary that lints with the NEWEST binary** (the 03-36 session's
   own suggestion): the wrapcheck flood sat undetected until a manual
   audit; a matrix lint cell would have caught it months earlier.
5. **Untracked source is invisible to nix**: `pkg/telemetry/` untracked
   made the vendorHash FOD resolve `pkg/telemetry` against the proxy and
   fail opaquely. Stage (or commit) source before any nix build.

## f) Things to get done next (live queue, execution order per plan)

| #   | Task                                                                                              | Master      |
| --- | ------------------------------------------------------------------------------------------------- | ----------- |
| 1   | Finish T3: build, captureRecorder.PreviouslyAutoEnabled + e2e + unit specs, docs, commit, push    | T3          |
| 2   | Watch CI on 89a87b3 (T1) + T3 push; full 19-suite race re-run                                     | gates       |
| 3   | Dogfood: run built CLI `configure` on THIS repo → land arangolint/clickhouselint retractions      | T5-pre      |
| 4   | F16 fleet pre-scan script: per-repo `configure --check --dry-run` bucket counts                   | T5          |
| 5   | F17 record expectation table across ~/projects/*                                                  | T5          |
| 6–9 | F18–F21 fleet sweep batches 1–4 (20/40/40/rest + skip-list with reasons)                          | T5          |
| 10  | F22 fold in 00-03 f9 scan (deadcode/varcheck/structcheck/exhaustivestruct/interfacer/nosnakecase) | T5          |
| 11  | F23–F25 fleet report + harvest into TODO/CHANGELOG/FEATURES + cross-repo audit-ledger note        | T6          |
| 12  | F26–F29 schema v2.15 residual: MinGolangCILintVersion derivation re-check (verify 03-36 claim)    | T7          |
| 13  | F30–F32 T8 residue: README requirements/claims sweep for version names                            | T8          |
| 14  | F37–F38 bisect upstream for the single-criterion rejection version; CHANGELOG + spec note         | T11         |
| 15  | F39–F40 verify v2.15 bare-dir behavior at source, then file upstream issue (github-voice)         | T12         |
| 16  | F41 commit v2.14-style + v2.15-style probe fixtures; parameterize guard spec                      | T13         |
| 17  | F11–F15 T4: provenance stamp (`# managed-by:` marker), refresh pass, specs, CHANGELOG + gotcha    | T4          |
| 18  | F42–F45 cosign v3 signs migration, unpin, snapshot sign run, release-dry-run gate                 | T14         |
| 19  | F50–F52 floating-installer inventory, pin them, action-SHA bump automation decision               | T16         |
| 20  | F53–F55 release-process.md tag-failure runbook + extended post-release-verify + dry-run vs v0.11  | T17         |
| 21  | F58–F59 Dependabot policy (prefer continue-on-error for dependabot actors) + green scheduled run  | T19         |
| 22  | F46–F49 USER-GATED: cosign backfill sigs, syft SBOMs, release-notes caveat, GHCR smoke            | T15         |
| 23  | F56–F57 USER-GATED: GHCR :master deletion, backfill-image throwaway dispatch                      | T18         |
| 24  | F60–F64 v0.12.0: CHANGELOG cut, pre-release checklist, full gates, tag, watch, post-verify        | T20         |
| 25  | F65–F66 markdownlint nested-config repro (boolean vs object keys)                                 | T21         |
| 26  | F67 FEATURES full re-audit → earn the 2026-11-02 stamp                                            | T21         |
| 27  | F68 restore going-public b/c table padding (cosmetic, archive)                                    | T21         |
| 28  | F69–F73 decision packets: buildflow posture, M32, AGENTS size, declined-verdict veto list, KBSK   | T22         |
| 29  | F74 flake.nix golangci-lint check output; F75 check-rows upstream fix; F76 erraudit calendar      | T22         |
| 30  | Absorb the live parallel session's landing (M5–M7) into docs/CHANGELOG when it commits            | co-tenant   |
| 31  | README "Optional OpenTelemetry" section + SECURITY telemetry paragraph (OTEL session's M8 debt)   | M8-debt     |
| 32  | OTEL session's M9/M10 residue: their full-gates + curated push when their tree settles            | M9-M10-debt |

## g) Questions I can NOT figure out myself

1. **Fleet sweep trigger (T5):** the plan locks "fleet execution stays
   user-triggered" per the 2026-10-07 cross-repo memo. Does "NOW GET SHIT
   DONE! The WHOLE TODO LIST!" constitute the trigger for the ~160-repo
   sweep (real commits in every sibling), or should T5 stay staged until
   you say sweep?
2. **USER-GATED scopes (T15/T18):** may I run the cosign signature + SBOM
   backfill for the 12 v0.11.0 assets and attempt the GHCR `:master`
   package-tag deletion with the current `gh` credentials this session?
3. **v0.12.0 vs the live session (T20):** the cut wants a settled tree,
   but `internal/cli` is being actively rewritten right now. Cut v0.12.0
   from the last green commit while their work is out, or sequence T20
   after their landing?

---

**Session commits:** ef60545, f0726ef, f3e8801, e17261e, 1cff856 (pushed,
CI green) · 89a87b3 (local, not yet pushed). **Files in flight (T3):**
pkg/linter/{desired.go, converge.go, fixer.go, fixer_converge.go,
evaluation.go}.
