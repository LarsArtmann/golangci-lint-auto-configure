# Pareto Execution Session — v0.8.1 Shipped, Guards Installed, 10/27 Tasks Done

**Created:** 2026-09-11 13:27 CEST
**Source:** Execution of `docs/planning/2026-09-11_09-28_SUPERB-pareto-execution-plan.md` (user order: "NOW GET SHIT DONE — the whole TODO list").
**Session window:** ~10:40–13:27 CEST, same day as plan creation.

---

## Executive Summary

The plan's entire 4% core (T1 + T2 + T3 ≈ 64% of total value) is **shipped, tagged, and verified in production**: v0.8.1 was released end-to-end — the release pipeline succeeded for the first time after three consecutive failures, publishing the **first-ever GHCR image** — and the two repair mechanisms (key-normalization self-heal, CI schema gate) are on master with BDD guards. The full 20% tier is nearly complete (T4–T10: 7/7 done). 10 of 27 medium tasks are fully done and pushed; every landed change has green CI.

The session also **caught three real bugs** the plan predicted would be hiding: the public `go install` promise is broken without `GOEXPERIMENT=jsonv2` (README fixed), the pre-release coverage check mis-parsed a green 65.9% suite as 0.0% (fixed by delegating to the CI gate), and exhaustruct_v5 correctly flags a migration DTO the old linter silently ignored (repo config updated).

---

## a) FULLY DONE (verified, committed, pushed)

All of these are on `master` with green CI runs; the working tree is clean.

| # | Task | Proof / Evidence |
|---|------|------------------|
| a1 | **T2 — Key-normalization pass** (`goconst.min-length` → `min-len` self-heal) | `normalizeKnownBadSettingsKeys` in `pkg/linter/fixer_config.go`; `constants.KnownBadSettingsKeys` map; 5 BDD specs (rewrite, unrelated keys preserved, good-key-wins, guard intact, no re-fire) in `pkg/linter/fixer_normalize_test.go`; verified with the real binary + `golangci-lint config verify` on a fixture (`SCHEMA-VALID`); AGENTS gotcha 7. Commit `6d8d907` |
| a2 | **T3 — CI schema-compat gate** | `cmd/generate-schema-fixture` (serialized through the tool's own `SaveConfig` path), committed fixture `pkg/constants/testdata/schema-fixture/.golangci.yml`, drift-guard + coverage specs in `pkg/constants/schema_gate_test.go`, new `schema-verify` CI job (drift check + `golangci-lint config verify`); fixture verified `SCHEMA-VALID` against live golangci-lint 2.13.2. Commit `bec0a98` |
| a3 | **T1 — Release v0.8.1** | CHANGELOG `0.8.1` section; FEATURES stamp v0.8.1; pre-release check **9 passed / 0 failed**; `nix build` green; annotated tag `v0.8.1` pushed; **Release workflow SUCCESS (3m35s, run 34583857045)**; `post-release-verify.sh` **9/9 passed** (binary reports 0.8.1 — ldflags correct); curated release notes published; checksums.sig/pem + SBOMs + .sig/.pem for all assets present |
| a4 | **First-ever GHCR image published** | GoReleaser metadata: `ghcr.io/larsartmann/golangci-lint-auto-configure:v0.8.1` + `:latest`, digest `sha256:a37b3c4d…`, linux/amd64 + arm64; anonymous `docker manifest inspect` succeeds (OCI image index). The buildx fix (`333c747`) is **proven** after failing on v0.7.0/v0.7.1/v0.8.0 |
| a5 | **Cosign keyless signature verified** | `cosign verify-blob --signature checksums.txt.sig --certificate checksums.txt.pem --certificate-identity-regexp …release[.]yml@refs/tags/v.* --certificate-oidc-issuer https://token.actions.githubusercontent.com checksums.txt` → **Verified OK** (cosign 3.1.3 from nixpkgs) |
| a6 | **T4 — exhaustruct → exhaustruct_v5 migration** | `DeprecatedLinters` entry (MinVersion v2.13.0), `KnownBadSettingsKeys` extended (`exclude` → `ignore-patterns`) so migrated v4 settings blocks self-heal; priorities/reasons/NeverAutoEnable/test-exclusions/finding-categories/LinterMinVersions/data-tests/categorizer-tests/fixer-tests all moved to v5; curated stdlib patterns ported verbatim; v5 key names **verified against the live 2.13.2 schema** (probed invalid keys, confirmed `ignore-patterns`); repo's own `.golangci.yml` migrated → **deprecation warning gone, 0 issues**; new deprecation DescribeTable entry. Commit `da9a433` |
| a7 | **T5 — vendorHash guard** | `scripts/vendorhash-guard.sh` (check mode = exact fix instructions; `--fix` = writes got-hash + re-builds); **proven end-to-end on a REAL hash mismatch** in a throwaway flake with a clean store (check mode failed with expected-vs-got, fix mode restored and rebuilt green); CI nix job wired with `--fix` + `git diff --exit-code` drift check; AGENTS gotcha 3 rewritten. Commits `818d980`, guard commit, `278bdce` |
| a8 | **T6 — Generated-file drift guard** | schema-verify job now regenerates `linter_settings_generated.go` and fails on any diff; **verified locally**: corrupted `min-len` → regeneration restored the file byte-identically; AGENTS #6 addendum. Commit `278bdce` |
| a9 | **T7 — CI-health watchdog** | `scripts/ci-watchdog.sh` (checks workflow `state=active` + latest **completed** master run `conclusion=success`; idempotent `ci-health` issue on drift) + `.github/workflows/ci-watchdog.yml` (weekly cron + `workflow_dispatch` with `force_drift` dry mode); **forced-drift dispatch run succeeded in CI (12s, run 34589893881)**; fixed a false-positive on in-flight runs (`status=completed` filter) discovered during local testing. Commit `ff0f96f` |
| a10 | **T8 — BDD spec debt (all 12 features)** | `pkg/constants/linter_settings_bdd_test.go`: exact wire output for gocognit/gocyclo/nestif (25/20/6), goconst (`min-len` NOT `min-length`), tagalign (curated ordering), mnd (ignored-files present, ignored-numbers intentionally absent), wrapcheck (16 regexps + 9 sigs), errcheck (check-type-assertions + 16 excludes), varnamelen (thresholds + decls + names); `pkg/types/health_test.go`: absolute-path exclusions (Unix + Windows `C:\`/`C:/` + portable globs), duplicate exclusion linters; `pkg/linter/fixer_config_internal_test.go`: `pruneUnenabledLinterSettings` (prune / keep-tool-level-disabled / empty no-op). Commit `03a3a53` |
| a11 | **T9 — varnamelen trim** | `c *gin.Context` removed from shared `ignore-decls` (no httpx/koanf entries existed in code); spec updated to stdlib-only with an explicit negative assertion; sibling sweep: 1 hit (`linter-autoconfigure-sdk`, manual per-project gin decl — left as-is, correct ownership); CHANGELOG `[Unreleased]` note. Commit `a97a221` |
| a12 | **T10 — go install e2e @v0.8.1** | Clean `GOMODCACHE` install: modules + deps downloaded through the public proxy; installed binary runs `--version` and full `analyze` (11 critical / 49 high / 45 medium / 1 optional on a fixture; clean user-friendly error without a config). **Found the README install promise broken** (see b1) and fixed it. Commit `f0985c0` |
| a13 | **Pre-release coverage check bugfix** | `scripts/pre-release-check.sh` grepped every per-package `coverage:` line and bc-mis-parsed the multi-line value → green 65.9% suite reported "0.0% < 60%" and blocked the release; now delegates to `cmd/coverage-check` (the same gate CI runs). This was the release's only failing gate |
| a14 | **CI green throughout** | Every push's CI run green (latest: run 34591180815 success); the one `cancelled` run (34589582357) was normal concurrency supersession by the next push, not a failure |

## b) PARTIALLY DONE

| # | Item | State | What remains |
|---|------|-------|--------------|
| b1 | **GOEXPERIMENT install gap (from T10)** | README fixed (`GOEXPERIMENT=jsonv2 go install …`); verified the env makes install succeed | Structural fix (vendoring json/v2 or waiting for Go to stabilize json/v2) is a roadmap decision; also the released **v0.8.1 README on GitHub still shows the old instructions** — the fix is on master for v0.8.2 |
| b2 | **T11 — Release-adjacent verification** | 11.1 done (tap repo exists public, `Casks/`/`Scoop/` absent — consistent with `skip_upload: true`; cross-repo push would need a PAT, documented below); 11.2 done (cosign verify **proven**, README section drafted mentally but the edit was interrupted); 11.3 (`gh workflow run ci.yml` dispatch test) and 11.4 (TODO prune) **not started** | Finish README artifact-verification section; dispatch-test ci.yml; prune TODO rows |
| b3 | **Cosign docs freshness** | GoReleaser still emits `.sig`+`.pem` (cosign 3.x deprecates `--certificate` in favor of `--bundle`); the verified command works today | Optionally migrate GoReleaser to bundle output; low priority |
| b4 | **Fixer deprecation coverage for settings blocks** | T4 relies on the `KnownBadSettingsKeys` pass to rename `exclude` → `ignore-patterns` after block migration; covered by normalization specs | No dedicated spec asserts the full exhaustruct→v5 *settings-block* migration end-to-end (linter rename + key rename in one run) — cheap add |
| b5 | **gopls diagnostics debt (observed, not acted on)** | A persistent stale gopls error (`fixer_config.go:364 NoNewVar`) and several `b.Loop`/`stdversion` warnings appear all session even though builds/tests/lint are green — every time verified against the authoritative CLI | Restart/`lsp_restart` investigation is a local-tooling chore, not a code problem |

## c) NOT STARTED (T12–T27, per plan order)

| Task | Note |
|------|------|
| T12 GHCR backfill for v0.8.0 | Needs the v0.8.1-proven buildx path; decision on v0.7.x backfill open |
| T13 Branch/tag protection rulesets | `gh api` rulesets; `auto-tag.yml` fate decision |
| T14 Dependabot failures (08-23/30, 09-06) | Root-cause from run logs; ci.yml golangci-lint pin custom-manager eval |
| T15 gitleaks full-history scan | Needs gitleaks binary (nixpkgs); triage + ROADMAP risk note |
| T16 Buildflow timeout/e2e/"9 tools" | buildflow config location still unknown |
| T17 generate-settings schema-version awareness | + goconst-drift decision (regenerate+annotate vs exclude) — note: T3's fixture gate changes the context for this decision |
| T18 internal/cli coverage sprint | Includes the FixConfig sidecar + ledger never-enable e2e |
| T19 Suite speedup (116s → ≤60s) | Profiling + parallelize; informs buildflow timeout |
| T20 json/v2 `omitempty` → `omitzero` | `pkg/types/config_types.go`; needs JSON-output goldens |
| T21 ADR consolidation | 8 inline ADRs → `docs/adr/`, rename `001-yaml-dependency-decision.md` |
| T22/T23 README audit part 1/2 | Now audits against the **released v0.8.1** binary |
| T24 Sibling sweep for emitted `min-length` keys | T2 fix shipped; sweep + repair ~160 repos |
| T25 Small-code-fixes bundle | FindingsHidden, multi-preset tests, `errUnsupportedFormat` registration |
| T26 Docs-hygiene bundle | Annotate 08-05 report, 07-31 F-ideas routing, version sweep, cadence policy |
| T27 Decisions + docs-integrity extension | PARTS/PROJECT_SPLIT/BDD_TESTS_REVIEW fates; homepage posture |
| Plan-file annotation | The plan file must be docs-health-ANNOTATEd as tasks complete (guardrail 7) — not yet started |
| TODO_LIST consolidation | Many plan rows are now closable (T1/T2/T3/T8/T9 etc.) — one prune pass pending |

## d) TOTALLY FUCKED UP

Nothing is broken or lost. Two self-inflicted stumbles, both caught and corrected within the session:

1. **Local FOD-caching rabbit hole (T5, ~20 min).** I tried to prove the vendorHash guard by corrupting the real repo's `vendorHash.nix` — `nix build` kept succeeding with a wrong-but-structurally-valid hash because the store-cached FOD output stayed valid (0 deletable paths; `--rebuild` only re-checked validity). I misread this as the guard being untestable before pivoting to the throwaway-flake test, which is the correct method and produced a REAL mismatch. Lesson recorded: Nix FOD verification on a warm store cannot prove hash-mismatch handling; test in a clean store.
2. **Heredoc backslash corruption.** The bash tool eats one level of backslashes, so python heredocs with `\t`/`\\n` literals twice corrupted Go test text (a duplicated `"\n"` fragment in a DescribeTable entry; failed asserts elsewhere). Fixed by switching to exact-text `edit`/`multiedit` for anything containing backslashes, and one corrupted line was repaired by direct `edit`. No damage reached git.

Also worth naming (not a fuckup but a near-miss): the first `go install` test silently "passed" because of exit-code masking (`… | tail` made the pipeline exit 0 while the install had failed). The AGENTS pipeline-masking rule was applied retroactively: re-ran without pipes to get the true failure.

## e) WHAT WE SHOULD IMPROVE

1. **Test hash-sensitive/build-sensitive logic against clean stores, not the dev machine** — warm caches (Nix store, Go build cache) invalidate half the failure modes you're guarding against.
2. **Never trust piped exit codes** — the `exit=$?` after `cmd | tail` pattern lied twice this session; use `set -o pipefail` or capture before piping.
3. **Fix README install claims at release time, not after** — T10's discovery (GOEXPERIMENT) should have been caught by T22's audit if the audit ran before the release; the plan's ordering (release first, audit later) trades discoverability for delivery — fine, but then the fix must land in the *next* patch release, which needs a TODO row (it has one now).
4. **Daemon interleaving makes per-file commit attribution fuzzy** — 5 daemon commits landed mid-work; my commits occasionally caught only a tail of files. Consider a `.git/hooks` daemon-aware guard or a pre-commit check that blocks `chore: auto-commit` when a semantic commit is in progress (user decision — daemon is user-owned).
5. **gopls stale diagnostics** — repeated false errors wasted verification cycles; run the CLI linter as the single source of truth earlier.
6. **Verify-before-document worked — keep doing it** — the cosign `Verified OK` and `SCHEMA-VALID` proofs made the docs trustworthy; no doc claims ship unproven.

## f) NEXT TASKS (up to 50, in execution order)

**Immediate (finish the interrupted task + tier completion):**
1. Finish T11.2: write the README artifact-verification section (cosign verify command proven in a5 + SBOM note + docker pull snippet).
2. T11.3: `gh workflow run ci.yml` + confirm dispatch path green.
3. T11.4: prune the TODO_LIST rows T1–T11 close (one consolidated pass).
4. T12: backfill GHCR image for v0.8.0 using the proven buildx path; decide v0.7.x backfill.
5. T13: master + `v*` rulesets via `gh api`; decide `auto-tag.yml` fate.
6. T14: read the 3 failed dependabot runs; fix; evaluate custom manager for the ci.yml golangci-lint pin.
7. T15: gitleaks full-history scan + triage + quantified risk note into ROADMAP.
8. T16: locate buildflow config; raise `test-coverage` timeout; "9 tools" health check; full e2e.
9. T17: generate-settings schema-version note + warning; decide goconst drift (fixture gate from T3 is the new context).
10. T18: internal/cli coverage sprint (cmd_check/cmd_analyze/cmd_configure_config/cmd_presets + FixConfig sidecar e2e).
11. T19: profile the 116s `-race` suite; target ≤60s.
12. T20: omitzero migration in `config_types.go` with per-field round-trip decisions + goldens.
13. T21: ADR consolidation (extract 8 inline ADRs, rename odd file, linked index, link sweep).
14. T22: README audit part 1 against the released v0.8.1.
15. T23: README audit part 2 + CI/CD section.
16. T24: sibling sweep for emitted `min-length` keys (~160 repos) + repair with the T2 fix.
17. T25: FindingsHidden decision, multi-preset union tests, `errUnsupportedFormat` registration.
18. T26: docs-hygiene bundle (08-05 annotation, 07-31 ideas routing, version sweep, cadence policy).
19. T27: PARTS/PROJECT_SPLIT/BDD_TESTS_REVIEW fates + homepage posture + docs-integrity extension.
20. Docs-health ANNOTATE the plan file as tasks complete (guardrail 7) — do it incrementally, not at the end.

**Follow-ups surfaced by this session (new, evidence-backed):**
21. Cut v0.8.2 to ship the README GOEXPERIMENT fix + exhaustruct_v5 migration (both are master-only until the next tag).
22. Add an end-to-end spec: exhaustruct→v5 linter rename + settings-block migration + `exclude`→`ignore-patterns` in one FixConfig run (b4).
23. Enabling homebrew/scoop publishing requires a PAT with `repo` scope on `homebrew-tap` — decide posture (enable with PAT in release workflow vs keep `skip_upload: true` and drop the tap repo).
24. Consider GoReleaser cosign `--bundle` migration (b3) before cosign 3.x drops `.pem` support.
25. Consider a structural fix for the GOEXPERIMENT install gap (vendor json/v2, or Go's eventual json/v2 stabilization) — ROADMAP.
26. Release-notes curation is now proven — codify it into `scripts/post-release-verify.sh` as a check (it already checks notes exist; add "not a commit dump" heuristic).
27. Watchdog currently covers ci.yml only — extend to release.yml + markdown-lint.yml states.
28. schema-verify job could also run `golangci-lint config verify` against `test.golangci.yml` + `examples/standard.golangci.yml` (example configs rot independently).
29. `nix flake update` + validated rebuild (Scheduled table — after v0.8.1 stabilizes).
30. Quarterly erraudit re-check (~2026-10, Scheduled table).

## g) QUESTIONS FOR YOU (cannot answer myself)

1. **Homebrew/Scoop tap posture:** the `homebrew-tap` repo exists and is public, but GoReleaser has `skip_upload: true` — was that deliberate (no PAT available / don't want tap maintenance) or leftover intent to publish? If you give the release workflow a fine-grained PAT with write access to `homebrew-tap`, I'll flip `skip_upload` and validate on the next release; if deliberate, should I delete `homebrew-tap` or leave it dormant?
2. **v0.8.2 timing:** master now carries three user-visible improvements beyond v0.8.1 (README GOEXPERIMENT fix, exhaustruct→v5 auto-migration, varnamelen stdlib-only trim). Cut v0.8.2 immediately to get these to users, or batch it with the remaining tier tasks (T12–T27) for a v0.9.0?
3. **The v0.8.1 release README on GitHub still shows the broken `go install` command** (fix is on master). Since releases pin their README, do you want me to (a) cut v0.8.2 promptly (fix ships), (b) edit the GitHub release description with a one-line warning pointing to master, or (c) leave it — Nix/ghcr users are unaffected and go-install users will find the master README?

---

**Bottom line:** the 1% and 4% tiers are delivered and *proven in production* (release succeeded, image live, signatures verified, self-heal works). The remaining 16 tasks are the non-differentiating 80% — well-scoped, order-free, and unblocked. Say the word and I continue at T11.2/T12.
