# Status: Pareto Execution T11–T25 — 13 More Tasks Done, Disk Wars, Honest Scars

**Created:** 2026-09-11, 23:26 CEST
**Session:** continuation of the 2026-09-11 pareto execution (T1–T10 + partial T11 were done in the
previous session; this session executed T11.2–T25).
**Branch:** `master`, all work pushed except the very last T25 batch (see d1).
**Verdict in one line:** 15 of 27 plan tasks are now fully done, 1 is one red spec away from done,
2 remain unstarted, and the session produced 3 real production fixes plus 2 workflow bugs that
I introduced and fixed myself.

---

## 1. What was forgotten / done badly (the blunt section)

1. **First GHCR backfill pushed the image as `:master` instead of `:v0.8.0`** — on
   `workflow_dispatch`, `GITHUB_REF_NAME` is the branch, not the input tag. The image content was
   correct (built from the v0.8.0 checkout) but the tag was wrong, and the verify step then
   exec'd `--version` as a command (image has CMD, no ENTRYPOINT) → exit 127. Both fixed in the
   same session, but the **stray `ghcr.io/...:master` tag still exists** — deleting a GHCR tag
   needs `delete:packages` scope which neither my `gh` token nor the workflow's GITHUB_TOKEN has.
2. **Disk-full chaos burned ~40 minutes.** /tmp is a 48G tmpfs that filled to 100% twice mid-task
   (test runs failing with `no space left on device`, pkg/audit reporting phantom "compilation
   failures"). Worse: my first cleanup **trashed files *within* tmpfs**, which moves them to
   `/tmp/.Trash-1000` — freeing *nothing*. I only understood this after `df` refused to move.
   Real fix was `trash-empty` (13G tmpfs trash) + trashing genuinely stale build dirs.
3. **Python-heredoc backslash mangling hit me twice more** (T25 test file) despite the known
   gotcha — corrupted a Go string literal with raw newlines. `edit`-tool fix took seconds; the
   heredoc attempts cost minutes. Lesson re-learned: never inline Go string literals via heredocs.
4. **T20 struct-tag alignment:** three manual alignment-guess rounds before remembering that
   `golangci-lint run --fix` applies the golines formatter directly. Should have been fix-first.
5. **Lint whack-a-mole from sed renames** (`binaryErr` → `buildErr` → `errBuild`) — one of the
   renames also created a shadowed variable (`output, buildErr := ...`) that `ineffassign`
   caught. Sloppy mechanical editing.
6. **T11.4 was claimed in the todo list from the previous session as pending when it was already
   partially stale** — minor bookkeeping churn recreating the todo state.
7. **The buildflow e2e goal "confirm 0 failures" (T16.4) was not achieved** — the findings gate
   fails on 235 advisory findings (branching-flow 190, erraudit 44) that this repo's policy
   (AGENTS #26) explicitly classifies as manual-review. I documented the posture decision as
   open instead of force-greening the gate by suppressing everything.
8. **Dependabot is fixed upstream but NOT proven green** — the proxy cache for
   `github.com/larsartmann/go-finding` is filled (verified via proxy-only `go list`), but the
   next scheduled Dependabot run is Sunday 2026-09-13. Until then this is a strong hypothesis,
   not a verified fix.

---

## 2. a) FULLY DONE (committed, pushed, verified)

| # | Task | Evidence |
|---|------|----------|
| 1 | **T11 Release-adjacent verification** | README artifact-verification section (cosign keyless command, SBOMs, GHCR pull) committed `b6b3bd7`; `gh workflow run ci.yml` dispatch run 34596025246 **success** (all 5 jobs green); TODO_LIST pruned of 12 closed rows (`96e2612`) |
| 2 | **T12 GHCR backfill for v0.8.0** | New `backfill-image.yml` workflow (reusable for any failed-release tag, never touches `latest`); v0.8.0 image live, **anonymous** `docker manifest inspect` shows amd64+arm64+attestations, `docker run … --version` prints `Version: 0.8.0`, commit `1f653cb` = tag v0.8.0 exactly. Runs: 34596612242 (failed, my 2 bugs), 34597444223 (**success**) |
| 3 | **T13 Branch/tag protection rulesets** | Ruleset 22912554 "master: no history rewrites" (deletion + non-fast-forward, DEFAULT_BRANCH) and 22912555 "release tags: immutable" (`refs/tags/v*`) — both **active**, verified via API. `auto-tag.yml` **deleted** (`ba66164`): it contradicted the release policy ("never push a tag without explicit confirmation") and had been disabled for 2 months |
| 4 | **T14 Dependabot root-cause + fix** | All 6 weekly runs since Aug 2 failed with `git_dependencies_not_reachable: go-finding`. Log forensics: proxy.golang.org 404'd the module AND git fallback 404'd through dependabot's proxy. The module is **now cached on proxy.golang.org** (all 31 versions listed; `GOPROXY=https://proxy.golang.org go list -m -versions` exits 0). Custom-manager for the `ci.yml` `version: v2.13.2` pin evaluated and **rejected** (no ecosystem matches a bare `version:` input); the schema gate remains the mitigation |
| 5 | **T15 gitleaks full-history scan** | `gitleaks git . --log-opts=--all` over **1,117 commits / 9.95 MB: zero leaks**. ROADMAP history question updated (`25c2c3a`): security dimension closed; only name-privacy/announcement remains |
| 6 | **T16 Buildflow** | `.buildflow.yml` **restored from git history** (`2cdeccf`) — it had vanished (auto-commit daemon clobber), re-arming the known-broken `nixfmt-standalone` step. Timeout myth **debunked in BuildFlow source**: test steps declare no step-level timeout (0 = unlimited); only `go test -timeout=10m` binds. "9 tools unavailable" = 8 JS/TS + 1 Python tools = language-filter noise for a Go repo |
| 7 | **T17 generate-settings schema awareness** | Provenance header in the generated file (snapshot version, tool minimum v2.10.1), `-schema-version` flag with loud newer-than-minimum warning, version-comparator BDD specs, drift-guard verified byte-identical (`a491a4d`), goconst-drift question resolved (reference file already has `min-len`), AGENTS updated |
| 8 | **T18 internal/cli coverage sprint** | 6 in-process dispatch specs (analyze json/sarif/finding/unknown-fallback, presets --json, bash completion) running the root cobra command with captured stdout — 0.5s instead of subprocess builds. 3 public-API FixConfig e2e specs with a REAL sidecar + REAL ledger file proving enforcement, never-enable priority, and backward compat |
| 9 | **T18 production fix (bonus)** | The enforcement path skipped never-enable linters **silently** — the audit ledger never recorded the decision. Now records `suppressed-re-enable` like the recommendation path (`e27b45f`); existing unit test updated to pin the new behavior |
| 10 | **T19 Suite speedup** | Root causes: `buildBinary()` rebuilt the binary per spec (50 call sites) → build-once `sync.Once`; **real data race** in `runAnalysisWithSpinner` (buffered done-channel let the spinner goroutine outlive the send and race the final stdout clear-write) → proper exit handshake. Result: `ginkgo -r -p -race` full suite **48.4s** (was ~116–150s serial); internal/cli 16–19s (was 71–104s). AGENTS documents the parallel command (`4f9a638`) |
| 11 | **T20 json/v2 omitzero migration** | 7 bool/int fields in `config_types.go` now `omitzero` on json tags only (yaml/toml keep v1 `omitempty`); pinning spec proves zeros vanish and non-zeros survive in `.golangci.json`; AGENTS #17 latent note closed (`29f6eb0`) |
| 12 | **T21 ADR consolidation** | 8 inline ADRs extracted verbatim to `docs/adr/ADR-008…015`, `001-yaml-dependency-decision.md` → `ADR-007` (git mv, 100% rename), ARCHITECTURE.md reduced to a linked index + Future Considerations. One scheme: 15 ADRs, `ADR-NNN-slug.md` (`a2cfee6`) |
| 13 | **T22+T23 README audit (both parts)** | Every checkable claim verified against the built binary + constants. **8 fixes**: missing exhaustruct_v5 deprecation entry, "--pragmatic drops 5" → 4, flags table missing `--pragmatic/--recommend/--quiet/--json-errors`, preset list missing `house`, Medium count 48 → **51** (+3 linters in details), exit-code table's wrong `EX_USAGE` label for code 1, parallel test command first, new **CI/CD section** (5 workflows, schema gate, rulesets). Verified-correct left untouched: 9 commands, exclusion defaults (matches schema fixture), critical/high counts 11/50, "100+ linters" (actual 112), exit codes 1/65/69/75, `report --format json` (`9cf8a83`) |
| 14 | **T24 Sibling sweep** | 9 hits across ~296 repos; **8 repaired** (goconst `min-length` → `min-len`, goconst-context-verified per file), **1 correctly skipped** (template-SECURITY's hit is `varnamelen.min-length`, a valid schema key). Repaired config verified with `golangci-lint config verify` exit 0 (erraudit spot check) |

## 3. b) PARTIALLY DONE

| # | Item | Exact remaining work |
|---|------|----------------------|
| 1 | **T25 Small-code-fixes bundle** (~90% done, uncommitted) | DONE: `FindingsHidden` dead ledger field removed; duplicate `errUnsupportedFormat`/`errUnsupportedConfigFormat` sentinels deduped (marshal path now wraps the rejection properly); multi-preset merge specs written; **real dedup bug found and fixed** (extra formatters were appended after set conversion → duplicates possible). REMAINING: my third spec fails on a type mismatch in the assertion (`ContainElements("gci")` vs `[]types.FormatterName`) — needs `types.FormatterName(...)` + 2 lint nits (wsl_v5, gci), then full suite + commit. `pkg/audit` + `pkg/config` suites already green; staged files: `cmd_configure_preset.go` + new internal test |
| 2 | **Dependabot green-run confirmation** | Waits for the Sunday 2026-09-13 18:52 UTC scheduled run; everything else verified |
| 3 | **T16.4 buildflow e2e green** | Runs clean through 35 steps; only the findings gate (advisory) blocks. Needs a posture decision (see question 3) |
| 4 | **3 user-gated questions from the previous session** | Still unanswered (see section g) — homebrew-tap publish, v0.8.2 timing, release-page README |

## 4. c) NOT STARTED

| # | Item | Note |
|---|------|------|
| 1 | **T26 Docs-hygiene bundle** | Annotate 08-05 humanize report, route 07-31 F-ideas, version-ref sweep, status cadence policy, TODO prune |
| 2 | **T27 Decisions + docs-integrity extension** | PARTS/PROJECT_SPLIT/BDD_TESTS_REVIEW fates, homepage routing, docs-integrity test beyond preset counts, final plan-status report |
| 3 | **Plan-file annotation (guardrail 7)** | The pareto plan was supposed to get docs-health ANNOTATEd incrementally as tasks completed — never started; all 15 completions are only in git history + status reports |
| 4 | **BuildFlow upstream feedback** | The "9 tools unavailable" health check should filter by project languages — investigated, not reported to the BuildFlow repo |

## 5. d) TOTALLY FUCKED UP (and current state)

| # | Item | State |
|---|------|-------|
| 1 | **T25 tree is red** | One failing spec (type-mismatch assertion) + 2 lint nits sit in staged/uncommitted files; the auto-commit daemon may commit them as-is. 3-line fix, not yet applied because you said report-then-wait |
| 2 | **Stray `ghcr.io/larsartmann/golangci-lint-auto-configure:master` tag** | Contains v0.8.0 content under a misleading tag. Cannot delete: needs `delete:packages` (token lacks it). Inert but confusing |
| 3 | **8 sibling repos have uncommitted `.golangci.yml` repairs** | Code-Quality-Agent, crush-daily, erraudit, go-idempotency, licenseforge, plugmarket, go-cqrs-lite, go-finding. Verified correct, but not committed (relied on their auto-commit daemons; no cross-repo push authorization) |
| 4 | **/tmp hygiene** | Now 52% used (was 100%). Remaining hogs are ANOTHER session's live builds (monitor365-client.lSVL3W 9.8G, mv-verify, mv-target) — deliberately untouched. `t10-gomodcache` failed to delete from trash (left behind, harmless) |
| 5 | **CI-state uncertainty on latest push** | `9cf8a83` (README) + daemon commits pushed without watching a full CI run to the end this session; dispatched-run proof exists only up to `cf3021f`-era + run 34596025246. Low risk (docs-only), unverified regardless |

## 6. e) WHAT WE SHOULD IMPROVE

1. **Test the dispatch-path details of new workflows before shipping them** — a dry-run of
   `GITHUB_REF_NAME` semantics (or a workflow-lint step) would have caught the `:master` tag bug.
2. **Delete-then-verify workflow steps**: every new workflow should end with an assertion that
   survives its own first run (my verify step ran against the wrong tag).
3. **Check `df` before long test batches on this machine** — tmpfs 48G fills silently from
   concurrent sessions; a pre-flight guard (or moving GOCACHE/GOMODCACHE checks) would save
   entire debug cycles.
4. **Never trash tmpfs files as a space-freeing strategy** — `trash` on tmpfs copies within the
   same filesystem. Use `trash-empty` or accept it.
5. **Use `golangci-lint run --fix` as the FIRST resort for format findings**, not the last.
6. **Stop using bash heredocs for Go string literals entirely** — the edit tool is strictly
   better; this session's two corruptions were both predictable.
7. **Cross-repo work needs an explicit authorization model** — 8 sibling repos sit in limbo
   (fixed but uncommitted) because push authority is unclear.
8. **Build-once caching and process-level test parallelism** should be the default assumption
   for new integration tests (write in-process first, subprocess only when the boundary IS the
   process).
9. **The findings-gate posture needs a written policy** — "erraudit is manual-review" lives in
   AGENTS #26, but buildflow's gate enforces error-severity anyway; the two sources disagree.
10. **GHCR tag hygiene** needs a decision (scope grant or accepted-wart note in release-process docs).

## 7. f) Up to 50 things to get done next (ordered by impact)

**Finish what this session started**
1. Fix the T25 spec assertion type + 2 lint nits, run full suite, commit (30 min).
2. Watch the Sunday 2026-09-13 Dependabot run; if green, prune the TODO row + record in AGENTS.
3. Remove the stray `:master` GHCR tag (needs `delete:packages` PAT or a one-shot workflow).
4. Decide + execute buildflow findings-gate posture (`fail_on: critical` won't help — erraudit findings ARE critical; likely `skip_steps: [erraudit, branching-flow]` or fix the 44).
5. Commit/push the 8 sibling `.golangci.yml` repairs (pending question 1).
6. T26.1: annotate `2026-08-05_03-25_humanize-linter-status.md` as resolved.
7. T26.2: route the dropped 07-31 F-ideas into ROADMAP or consciously drop.
8. T26.3: version-reference sweep (AGENTS #11, docs/references).
9. T26.4: status cadence policy into `docs/status/README.md` + final TODO prune.
10. T27.1: decision note on PARTS/PROJECT_SPLIT/BDD_TESTS_REVIEW fates.
11. T27.2: homepage + announcement posture routing.
12. T27.3: extend docs-integrity test beyond preset counts (FEATURES.md counts).
13. T27.4: final plan-status report + docs-health ANNOTATE the pareto plan file (guardrail 7 — overdue for all 15 done tasks).
14. Mark T25/T26/T27 rows closed in TODO_LIST.md.

**Verify & guard**
15. Full CI watch on the current master head (post-README-audit).
16. Cut v0.8.2 (README fixes, exhaustruct migration, varnamelen trim, omitzero, dedup fix — pending question 2).
17. Backfill-image workflow: smoke-test with a throwaway tag before the next real need.
18. Add a CI job that asserts the schema-fixture ALSO round-trips as JSON (omitzero class bugs).
19. gitleaks: add the full-history scan as a scheduled CI job (currently manual-only).
20. Dependabot: add `groups` for golang.org/x minor bumps to cut PR noise.
21. Wire `buildflow diff` into PR checks so only changed-file findings gate PRs.

**From the README audit (verified-stale adjacent docs)**
22. Sweep docs/references/*.md for the same stale claims the README had (5-noise linters, 48-medium counts, missing house preset).
23. Regenerate the schema snapshot from live golangci-lint 2.13.2 (`-schema-version` flag now exists) and refresh exhaustruct_v5 keys in the reference file.
24. Record the "no v0.7.x image backfill" decision in release-process.md (currently only in commit history).
25. Document the GHCR tag model (release tags + latest; master-tag wart) in release-process.md.

**Test debt (from T18 leftovers)**
26. cmd/migrate.go: 13 functions at 0% — the migrate subcommand is user-facing and completely untested.
27. cmd/ completion + installhook: NewCompletionCommand/generateCompletion/runInstallHook at 0%.
28. cmd_report.go: runReport/resolveReportConfig/writeReport at 0%.
29. cmd_configure_config.go: prepareConfigFile/ensureConfigFile/mapKeys at 0%.
30. internal/cli coverage re-measure after T18/T25 (18.7 delta report never produced).
31. Ginkgo parallel labels: mark subprocess-heavy suites so `-p` scheduling stays optimal.

**Upstream / ecosystem**
32. File the BuildFlow language-filter issue for the tools-unavailable health check.
33. Suggest BuildFlow surface step-level timeout config (`.buildflow.yml` per-step override).
34. Consider a `docs/adr/ADR-016` for the findings-gate policy decision once made.
35. Sibling sweep part 2: check siblings for OTHER KnownBadSettingsKeys-class drift (the map will grow).
36. Add the 8 repaired siblings to the audit ledger's cross-repo story (or a status note).

**Hygiene**
37. `nix flake check` "running 0 flake checks" root-cause (open Low TODO, untouched).
38. `FindingsHidden` removal follow-up: bump the audit JSONL schema note (consumers may expect the field).
39. Link sweep for the renamed `001-yaml-dependency-decision.md` (TODO/status docs reference the old name — historical, but a redirect note in the file would help).
40. `erraudit` quarterly re-check is due ~2026-10 (calendar item).
41. Metadata checklist script (description/topics/badges in one gh api pass — open Low TODO).
42. Link checker (lychee) in CI (open Low TODO).
43. `Detect()` error-path contract decision (deferred twice, still open).
44. Homebrew cask + scoop publish story (pending previous question 1).
45. Release-page README warning for v0.8.1 (pending previous question 3).
46. Move buildBinary() shared dir cleanup into TestMain (temp dirs accumulate per run).
47. Consider `Serial`-labeling the two os.Stdout-capturing in-process suites if future parallel flakes appear.
48. `.buildflow.yml`: add `golangci-cache`/tmp paths to excludes if buildflow starts scanning /tmp artifacts.
49. AGENTS: record the trash-on-tmpfs lesson (it burned this session and a previous one).
50. Celebrate: 15/27 plan tasks done in one day — then close the plan file with the final report (T27.4).

## 8. g) Questions I cannot answer myself

1. **Cross-repo write authorization:** The 8 sibling `.golangci.yml` repairs (goconst
   `min-length` → `min-len`, verified against `golangci-lint config verify`) are sitting
   uncommitted in Code-Quality-Agent, crush-daily, erraudit, go-idempotency, licenseforge,
   plugmarket, go-cqrs-lite, and go-finding. Should I commit (and push?) them, or do their
   auto-commit daemons own that? Same scope question as the old homebrew-tap PAT question —
   if you want me doing cross-repo writes, a PAT with the right scopes would unblock both.
2. **v0.8.2 now or batch?** Since v0.8.1 shipped, master accumulated: README overhaul, exhaustruct_v5
   migration, varnamelen trim, omitzero JSON fix, the multi-preset formatter-dedup fix, and the
   FindingsHidden/sentinel cleanups. Cut v0.8.2 now (users get the omitzero + dedup fixes), or
   batch into v0.9.0? (This was question 2 from the 13-27 status report — still unanswered.)
3. **Buildflow findings-gate policy:** The e2e gate fails on 235 advisory findings
   (branching-flow 190, erraudit 44 critical-severity, go-structure-linter 1) while AGENTS #26
   says erraudit is manual-review, not a gate. Do I (a) add `erraudit`/`branching-flow` to
   `skip_steps` in `.buildflow.yml`, (b) drop gate severity via `fail_on: critical`-plus-fixes,
   (c) fix/annotate the 44 erraudit findings for real, or (d) leave the gate red and documented?
   I cannot decide this for you because it defines what "green" means for YOUR local buildflow.

---

*All claims above verified this session unless marked as pending (Dependabot Sunday run, final CI
watch). Commit chain: `96e2612 → b6b3bd7 → 39169d4 → 14e74d0 → ba66164 → 25c2c3a → 2cdeccf →
a491a4d → e27b45f → 4f9a638 → 29f6eb0 → a2cfee6 → 9cf8a83` (+ interleaved daemon commits).*
