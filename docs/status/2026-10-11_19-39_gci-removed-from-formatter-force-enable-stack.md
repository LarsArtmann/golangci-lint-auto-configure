# Status Report: gci Removed From Formatter Force-Enable Stack

- **Date:** 2026-10-11 19:39 CEST
- **Repo:** `golangci-lint-auto-configure` (branch `master`)
- **Session scope:** One task — "many of my projects do not like `gci` being enabled; remove it from the force enable list" — implemented end-to-end, verified, docs synced.
- **Format note:** Skill default is a styled HTML dashboard; the user explicitly demanded `.md` at this path, so `.md` wins (one-off override, not propagated into the skill).

---

## TL;DR

`gci` is no longer force-enabled anywhere by default. The default formatter stack is now `{goimports, gofumpt, golines}` for both `CoreFormatters` (non-preset configure runs) and the `format`/`house` presets. gci remains a fully supported formatter: it stays in `FormatterOrder`/`FormatterData`, manually-enabled configs keep it (the fixer only ever adds core formatters, never removes non-core ones), and it is still recommended in analyze reports when detected-but-disabled. All gates green: full test suite (`-race`), lint, build, dogfood `configure --check`.

Evidence commits (auto-commit daemon, heuristic messages):
- `7b117d5` (19:34) — `pkg/constants/config.go`, `pkg/constants/presets.go`, `pkg/constants/data_integrity_test.go`
- `acbe791` (19:36) — `AGENTS.md`, `FEATURES.md`, `internal/cli/cmd_configure_internal_test.go`, `internal/cli/cmd_configure_preset_internal_test.go`

---

## a) FULLY DONE

| # | Item | Evidence | Scope |
|---|------|----------|-------|
| 1 | `gci` removed from `CoreFormatters` (the force-enable list for every non-preset `configure` run) | commit `7b117d5`; `CoreFormatters alignment` spec now asserts exactly `{goimports, gofumpt, golines}` and passes | `pkg/constants/config.go:35-39` |
| 2 | `format` and `house` presets + `PresetDescriptions` updated to the 3-formatter stack (forced by the `house == CoreFormatters` alignment spec — same length, subset) | commit `7b117d5`; Constants Suite 103/103 green | `pkg/constants/presets.go` |
| 3 | Format-preset canonical-order test rewritten: was a strict prefix check against `FormatterOrder` (implicitly coupled to gci being first), now checks the preset is a sorted subsequence of `FormatterOrder` — the actual invariant | commit `7b117d5`; spec passes | `pkg/constants/data_integrity_test.go` |
| 4 | New negative assertion: format preset must NOT contain gci (regression guard for this exact revert) | commit `7b117d5` | same file |
| 5 | CLI tests updated: preset-union test (3 entries), dedup-extras test now uses `gofumpt`+`swaggo` (gci no longer a preset formatter, so it could no longer exercise dedup), `applyPreset("format")` expectations, YAML-output assertions | commit `acbe791`; CLI Commands Suite 72/72 green | `internal/cli/cmd_configure_internal_test.go`, `internal/cli/cmd_configure_preset_internal_test.go` |
| 6 | Living docs synced: `FEATURES.md` preset table (5 linters + **3** formatters, both presets), formatter-management row (3-formatter stack, gci dropped 2026-10-11), `AGENTS.md` gotcha #20 rewritten with the removal rationale and the deliberate gci retention in `FormatterOrder`/`FormatterData` | commit `acbe791`; both docs-integrity spec families (FEATURES count lockstep, formatter/preset tables) green | `FEATURES.md`, `AGENTS.md` |
| 7 | Full verification loop executed: `ginkgo -r -p -race` over `pkg/... internal/... cmd/...` — ALL suites SUCCESS; `go build ./...` + `go vet` clean; `golangci-lint run` on both changed packages → **0 issues**; `cmd/...` go tests ok | session log; commands reproducible from AGENTS.md §Commands | whole repo |
| 8 | Dogfood gate verified locally: `configure --check --config .golangci.yml` → "No fixes needed - Configuration is optimal" (this repo's own config still has gci enabled; the fixer respects manually-enabled formatters, so check stays idempotent) | session run output | repo config |
| 9 | Schema fixture confirmed unaffected: `DefaultFormatterSettings` only contains `golines` (gci never had curated settings), so the committed schema-fixture did not drift — `schema_gate_test.go` green without regeneration | Constants Suite run | `pkg/constants/testdata/schema-fixture/` |
| 10 | Verified no-op areas before deciding not to touch them: `FormatterOrder` (config-output ordering — gci kept), `formatter_data.go` (gci priority/reason kept), `categorizer_test.go` gci cases (dynamic recommendation path, unaffected), `cmd/generate-schema-fixture` (reads settings maps, not `CoreFormatters`) | code read during session | various |

## b) PARTIALLY DONE

| Item | What works | What remains | Effort |
|------|-----------|--------------|--------|
| Doc-wide consistency sweep | FEATURES.md (spec-locked) and AGENTS.md are updated and provably in lockstep | `README.md` usage examples were **not** audited for a shown 4-formatter `configure` output or a formatter-stack claim (README is not covered by any docs-integrity spec, so nothing fails CI if it is stale). `docs/DOMAIN_LANGUAGE.md` lists gci as an example formatter — still true, no change needed | S |
| Behavior-change record | Rationale is durably recorded in AGENTS.md gotcha #20 + FEATURES.md rows | No CHANGELOG entry yet (release-time concern); the daemon's heuristic commit messages carry no semantics, so the "why" lives only in docs, not history | S |
| E2E confidence | Full unit/spec suite + dogfood check ran locally | The CI `e2e-pin-matrix` job (asserts `configure --check` idempotency + config loadability across pinned golangci-lint v2.10.1–v2.14.0) was **not** run locally this session; analysis says it must pass (fixer only adds), but it is unproven until CI or a local run | M |

## c) NOT STARTED

| Item | Why not started |
|------|-----------------|
| Rollout sweep of existing sibling configs — the ~128 projects that previously had gci force-enabled into their `.golangci.yml` still contain it; this change affects only **future** configure runs | Session scope was the tool's default, not a 128-repo migration; needs a decision (see question g-2) |
| Formatter-side opt-out mechanism — linters have the sidecar `never-enable` section; **formatters have no equivalent at all**. If a project dislikes `goimports` tomorrow, the only option is editing every config by hand | Noticed during session; design work (sidecar schema extension) not started |
| Decision on the recommendation path: `CategorizeFormatters` still recommends gci (in analyze reports) when it is disabled-but-available. This does not auto-enable anything, but it means reports will keep suggesting the formatter the default stack just dropped | Needs a product decision (see question g-1) |
| `CoreFormatters ⊆ FormatterOrder` spec — the invariant "order matches FormatterOrder" (comment on `config.go`) is enforced for the format preset (subsequence check, this session) but has **no direct spec for `CoreFormatters` itself** | Gap noticed while rewriting the preset test; not written |
| HARVEST of this report's section (f) into `TODO_LIST.md`/`ROADMAP.md` | User instruction: write report, then WAIT |

## d) TOTALLY FUCKED UP

Nothing shipped broken — all gates green. Radical-honesty items:

1. **Two compile errors introduced mid-edit, caught by gopls, fixed in the next edit.** The rewritten format-preset order spec initially used `Contain(-1)` (nonexistent matcher — Gomega has `ContainElement`) and a missing `slices` import. LSP diagnostics flagged both immediately; never reached a test run. Lesson: the file's import block was read but the matcher set was assumed instead of checked.
2. **Commit history has no semantic record of a user-facing behavior change.** The auto-commit daemon landed the change as `chore: auto-commit 3 changed file(s) (heuristic)` / `4 changed file(s)`. Any future `git log` archaeology finds nothing about why the default formatter stack shrank. Mitigation exists (AGENTS.md #20, this report), but history itself is uninformative — and amending daemon commits is not worth the risk.
3. **The "validated winning stack across 128/160 sibling projects" claim was empirically stale.** The data came from the 2026-06/07 cross-project audits; the user's lived experience ("many of my projects do not like gci") contradicted the locked-in constant. The tool encoded a snapshot as a permanent truth — and its own data-integrity tests then made the stale claim *hard to change* (house==Core length lock, FEATURES count lock). The locks worked as designed but locked in an unverified-forever assumption.
4. **Preset description strings are not spec-locked.** `PresetDescriptions["format"]` says "(5 linters, 3 formatters)" as free text; FEATURES.md counts are locked by specs, but these in-code description strings could drift from `PresetLinters`/`PresetFormatters` with no failing test.

## e) WHAT WE SHOULD IMPROVE

1. **Formatter symmetry with the linter policy system.** Linters got `never-enable` (sidecar), regression-loop detection via the audit ledger, and anti-gaming enforcement. Formatters have none of that — `CoreFormatters` is force-added unconditionally with no opt-out. Today's task (remove gci) was a code change; the same preference tomorrow should be a one-line sidecar entry. Suggested fix: extend `.golangci-lint-auto-configure.yml` with a `never-enable` formatter section, checked in `EnableCoreFormatters`.
2. **Distinguish "validated" from "audited once."** Constants documenting empirical claims (the 128/160 stack) should carry their audit date in the comment (they now do, incidentally) and ideally a spec-anchored provenance pointer to the research doc. A claim without a re-validation date rots silently — this session proved it.
3. **README is outside the docs-integrity net.** FEATURES.md counts and the deprecation-migration table are spec-enforced; README examples and claims are not. One stale README screenshot/example already survives this session unchecked. Fix: extend `docs_integrity_test.go` to lock the README's preset/formatter claims (or explicitly mark README examples as illustrative).
4. **Force-enable lists should be data, checked against their order index.** This session's prefix-check failure showed a hidden coupling: the format-preset test silently depended on `FormatterOrder[0] == gci`. The rewritten subsequence check is robust, but a `CoreFormatters ⊆ FormatterOrder` spec would close the same class of gap for the core list.
5. **Post-daemon commit hygiene.** When a session makes a user-facing behavior change, immediately note the intended semantic commit message somewhere durable (CHANGELOG Unreleased section is the natural place) so the release process can reconstruct the "why" that the daemon's heuristic messages destroy.
6. **Categorizer recommendations vs. default stack should share a source.** After this change, the fixer's default (no gci) and the analyzer's recommendations (gci still recommended when detected-disabled) can disagree. If that divergence is intended (recommendations are opt-in), document it; if not, gate categorizer formatter recommendations by the same core/redundant data.

## f) NEXT TASKS (session-grounded brainstorm; user cap: 50 — 30 honest items here, the remaining space intentionally left empty rather than padded with un-researched filler)

| # | Task | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Decide + execute gci rollout for existing sibling configs (sweep vs. leave) | High | L | Cleanup |
| 2 | Add formatter `never-enable` sidecar support (`EnableCoreFormatters` + policy schema + specs) | High | M | Feature |
| 3 | Add `CoreFormatters ⊆ FormatterOrder` data-integrity spec | Medium | S | Quality |
| 4 | Spec-lock `PresetDescriptions` counts against `PresetLinters`/`PresetFormatters` | Medium | S | Quality |
| 5 | Add CHANGELOG "Unreleased" entry: gci removed from default stack, migration note for re-adding | High | S | Documentation |
| 6 | Audit README for stale 4-formatter examples/claims; update or mark illustrative | Medium | S | Documentation |
| 7 | Extend docs-integrity specs to README preset/formatter claims | Medium | M | Quality |
| 8 | Run `scripts/e2e-pin-matrix.sh` locally (or await CI) to prove idempotency across pinned golangci-lint versions post-change | High | M | Quality |
| 9 | Decide whether categorizer should still recommend gci in analyze reports; implement decision | Medium | S | Feature |
| 10 | Re-run the cross-project formatter audit (refresh the 128/160 claim, re-earn or drop the word "validated") | Medium | L | Quality |
| 11 | Consider a `--formatters` override flag / `--no-default-formatters` for per-run control | Medium | M | Feature |
| 12 | Decide whether this repo's own `.golangci.yml` should drop gci (dogfood the new stack; accepts an import-reformat diff) | Low | S | Cleanup |
| 13 | HARVEST this report into `TODO_LIST.md` (items 1-11) / `ROADMAP.md` (items 10-12) | High | S | Documentation |
| 14 | Add release-note copy for the behavior change (who is affected: everyone relying on default configure; what to do: add gci back manually or via sidecar once #2 lands) | Medium | S | Documentation |
| 15 | Write the migration doc: "I want gci back" one-pager (README section) | Medium | S | Documentation |
| 16 | Check `docs/references/working-with-codebase.md` / `code-organization.md` for quadruple references (not grepped this session — md-wide `\bgci\b` grep covered them implicitly: no hits outside archive/status/research + the files already updated) — confirm and close | Low | S | Documentation |
| 17 | Add a spec that `house` preset linters remain `minimalLinters` (composition invariant currently only implied by code comment) | Low | S | Quality |
| 18 | Evaluation: should `FormatterOrder` drop gci if no first-party config uses it anymore? (Counter-argument: third-party users; ordering list is harmless) — decide and document | Low | S | Cleanup |
| 19 | Tag a patch/minor release carrying this behavior change after #5/#14 | Medium | M | Release |
| 20 | Dogfood note: run `configure` (non-check) on a scratch project to eyeball the generated 3-formatter config end-to-end (unit tests cover it; one human eyeball is cheap) | Low | S | Quality |
| 21 | Consider extracting the "formatter stack" into a named preset constant reused by CoreFormatters + format + house (currently the triple is written in 3 places: `config.go`, `presets.go` x2 — the alignment specs make drift loud, but a single constant makes it impossible) | Medium | S | Refactoring |
| 22 | Verify `cmd_presets` output for house/format shows 3 formatters (dynamic map read — code-verified this session, no runtime eyeball yet) | Low | S | Quality |
| 23 | Sweep `docs/status/` + TODO_LIST for older items invalidated by this change (e.g., any TODO mentioning the gci quadruple as current truth) | Low | S | Cleanup |
| 24 | Follow-up: does any sibling *tooling* (buildflow nix-checker, repo templates) also inject gci independently of this tool? (Out of session scope; flagged by the "many projects" framing) | Medium | M | Bug |
| 25 | Add the audit-ledger action type for formatter force-enables (linters record enables; formatter mutations via `rec.formatter` are counted but not ledgéd — asymmetry noticed at `fixer.go:314`) | Medium | M | Feature |
| 26 | Consider making the removed-gci rationale greppable from the code (constant `// RemovedFromDefaultStack: 2026-10-11` on the gci FormatterData entry) so future audits find it without AGENTS.md | Low | S | Documentation |
| 27 | Update `docs/research/2026-06-17_cross-project-config-audit-and-fixer-improvements.md` successor note pointing to this decision (it currently recommends the quadruple as "Core formatters") | Low | S | Documentation |
| 28 | Property: `configure --check` idempotency on a config that has gci + all core formatters (exactly this repo's config) — covered by dogfood run; convert the manual verification into a committed spec | Low | S | Quality |
| 29 | Question the `house` preset description wording ("House stack across 128/160 projects") — the 128/160 number described the quadruple; the triple's real-world validation is the user's complaint, not an audit | Low | S | Documentation |
| 30 | Re-check `README.md` line 232 (`--preset format # Core formatters + essential linters`) — description still accurate post-change; no edit needed unless formatter count appears elsewhere (fold into #6) | Low | S | Documentation |

*(Items 31-50: intentionally unused. Filling them would require researching sibling repos, CI history, and TODO_LIST — explicitly out of scope for this session per instruction. Better 30 grounded items than 20 fabricated ones.)*

## g) QUESTIONS ONLY YOU CAN ANSWER

1. **Should gci also disappear from analyze *recommendations*?** The fixer no longer force-enables it, but `CategorizeFormatters` still recommends gci whenever it is disabled-but-available, so reports will keep suggesting the formatter the default stack just dropped. I could not determine from code or docs whether recommendation ≠ default-stack membership is intended policy (recommendations are opt-in) or an oversight. Your call: silence gci in recommendations too, or keep recommending and let users opt in?
2. **Sweep the ~128 existing sibling configs, or future-only?** Every project that ever ran `configure` before today still carries force-enabled gci in its committed `.golangci.yml`. I cannot know which of those repos are active, whether their import grouping is now gci-shaped (removing gci without reformatting leaves configs enforcing nothing but rendering imports in gci's layout), or whether you want a scripted sweep (and if so: remove-only, or remove + reformat + verify per repo). What is the rollout intent?
3. **Does this repo's own `.golangci.yml` stay on gci?** I deliberately left it (line 496, `formatters.enable`) because your instruction targeted the tool's force-enable list, and removing gci from this repo would trigger an import reformat across the codebase. But the dogfood story is now inconsistent: the tool ships a triple while its own repo lints with a quadruple. Dogfood the new stack (accept the reformat diff), or keep the quadruple here as a live test of the "manually-enabled gci keeps working" path?

---

## Verification Receipts (reproducible)

```bash
export CGO_ENABLED=1 GOEXPERIMENT=jsonv2
ginkgo -r -p -race --skip-package=cmd/ ./pkg/... ./internal/... ./cmd/   # ALL suites SUCCESS
go test ./cmd/...                                                        # ok
golangci-lint run --config=.golangci.yml ./pkg/constants/... ./internal/cli/...   # 0 issues
go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure
./bin/golangci-lint-auto-configure configure --check --config .golangci.yml       # No fixes needed
```

Commits: `7b117d5` (constants + specs), `acbe791` (docs + CLI tests) — both by the auto-commit daemon.

**Status: report written. Waiting for instructions.** (Section (f) is HARVEST-ready for `docs-health` on your go.)
