# Status Report: forcetypeassert Test Exclusion Default

**Date:** 2026-07-25 05:40
**Session scope:** Adding `forcetypeassert` to default `_test.go` exclusion rules
**Trigger:** AI agent hit `forcetypeassert` false-positive-style finding in test code; user asked if excluding it for `_test.go` should be the default

---

> **PARTIALLY RESOLVED (docs-health sweep 2026-10-07).** Section-f items 1-6
> are struck inline below (all shipped). Items 7-50 are inline runs inside
> paragraph lines the strikethrough tooling cannot target; their fates,
> verified this sweep: exclusion evaluations 8 (wrapcheck) and 10 (gosec)
> adopted into `DefaultExclusionRules`; 7/9/11/12/13 evaluated, not adopted;
> 14 (cyclop review) still open. Documentation checks 15-17 open; 18 done
> (CHANGELOG); 19 obsolete (conditional). Integrity tests 21-23 done
> (`data_integrity_test.go`); 33-34 done (CI counts + tool-on-own-config).
> 29 done (typed settings coverage). 37 obsolete (sidecar non-goal). The
> rest (20, 24-28, 30-32, 35-36, 38-50) remain open ideas — bare below.
>
> **CLOSED (2026-11-02, docs-health fifth sweep):** every previously-bare
> item now carries an inline verdict above — adopted/evaluated via friction
> evidence (7-14), covered by existing gates (20-23, 26-27, 29, 33-34, 38,
> 42-43), or consciously declined (the rest). The three `g` questions are
> answered inline. **Fully resolved — archived.**

---

## What Was Done

Added `forcetypeassert` to the default linters excluded from `_test.go` files, in two places:

1. **`pkg/constants/config.go:78`** — `DefaultExclusionRules[0].Linters`: added `"forcetypeassert"` to the slice. This is the canonical default injected into all user configs via `CreateDefaultConfig` and the fixer flow (`updateExclusionRules`).

2. **`.golangci.yml:237`** — the project's own `_test.go` exclusion rule: added `forcetypeassert` for dogfooding consistency.

### Verification Done

- `go build ./...` — passed
- `go test -race ./pkg/constants/... ./pkg/linter/... ./pkg/config/...` — all passed

### Justification

Unchecked type assertions in test code are an intentional pattern: panicking on a wrong type fails the test immediately with a clear stack trace. Forcing `x, ok := y.(T); if !ok { t.Fatal(...) }` adds noise with zero safety benefit in tests. Consistent with the 6 linters already excluded from `_test.go` by default.

---

## a) FULLY DONE

- [x] `forcetypeassert` added to `DefaultExclusionRules` in `pkg/constants/config.go`
- [x] `forcetypeassert` added to `.golangci.yml` project config
- [x] `go test -race` on affected packages (`constants`, `linter`, `config`) — green

---

## b) PARTIALLY DONE / FORGOTTEN

### 1. ~~FEATURES.md is now stale — DOCUMENTATION DRIFT~~ → RESOLVED

> **Resolution (2026-07-25):** FEATURES.md line 80 now correctly reads "7 linters for `_test.go`" with `forcetypeassert` in the enumeration. Fixed during the docs-health pass (commit in batch `9ff74db`→`7c02302`). The claim below is preserved for historical context.

**Original claim (now resolved):** `FEATURES.md:69` read:

> `| Default test exclusion rules (6 linters for _test.go) | Stable | exhaustruct, testpackage, gochecknoglobals, funlen, cyclop, goconst |`

It listed 6 linters and named them explicitly. After the forcetypeassert change it's 7 linters and the enumeration was incomplete. This was not updated at the time of writing. **The AGENTS.md explicitly warns: "Documentation drift" is a quality gate failure.** The fix was applied in the subsequent docs-health session.

### 2. Did NOT run `golangci-lint run`

I modified `.golangci.yml` but never ran the linter to verify:

- The YAML is valid
- The project code passes linting with the new config
- No new findings surface

Tests passing ≠ lint passing.

### 3. Did NOT run `nix fmt` / `nix flake check`

AGENTS.md specifies these as the standard verification. I only ran bare `go test`. The `nix fmt` step (treefmt) verifies formatting of Go, Nix, and YAML files.

### 4. No BDD test added

The project uses Ginkgo BDD specs and values test coverage. No existing test asserts on the exact contents of `DefaultExclusionRules` (confirmed via grep), so nothing broke. But there's also no test that verifies `forcetypeassert` is in the default exclusions. A data-integrity-style test or a fixer BDD spec ("when configuring defaults, forcetypeassert is excluded from test files") would lock in the policy decision.

---

## c) NOT STARTED

- ~~FEATURES.md update~~ ✅ Done (docs-health pass, 2026-07-25)
- BDD test for the new exclusion behavior
- `golangci-lint run` validation
- `nix fmt` / `nix flake check` validation
- CHANGELOG.md entry (if the project maintains one for behavior changes)
- AGENTS.md gotcha update (optional — gotcha #7 doesn't enumerate exclusion linters, so not strictly needed)

---

## d) TOTALLY FUCKED UP

Nothing catastrophically broken. The code change is correct and tests pass. But the documentation drift (FEATURES.md) is a self-inflicted quality gate miss that should have been caught before declaring "Done."

---

## e) WHAT WE SHOULD IMPROVE

### Broader Observation: Default vs Project Exclusion Gap

The project's own `.golangci.yml` excludes **MORE** linters from `_test.go` than `DefaultExclusionRules` provides to users:

| Linter              | In `.golangci.yml` _test.go exclusion | In `DefaultExclusionRules` |
| ------------------- | ------------------------------------- | -------------------------- |
| exhaustruct         | yes                                   | yes                        |
| testpackage         | yes                                   | yes                        |
| gochecknoglobals    | yes                                   | yes                        |
| funlen              | yes                                   | yes                        |
| goconst             | yes                                   | yes                        |
| **forcetypeassert** | **yes (just added)**                  | **yes (just added)**       |
| unused              | yes (separate rule)                   | yes (separate rule)        |
| **err113**          | **yes**                               | **no**                     |
| **wrapcheck**       | **yes**                               | **no**                     |
| **varnamelen**      | **yes**                               | **no**                     |
| **gosec**           | **yes**                               | **no**                     |
| **paralleltest**    | **yes**                               | **no**                     |
| **revive**          | **yes**                               | **no**                     |
| **musttag**         | **yes**                               | **no**                     |
| cyclop              | no                                    | yes                        |

The project benefits from richer test-file exclusions than it ships to users. `err113`, `wrapcheck`, and `varnamelen` are strong candidates for the defaults — they are noisy in tests for the same structural reasons as the already-included linters. This gap should be reviewed.

### Process Improvements

1. **Pre-declare a checklist.** When modifying default linter lists, the checklist should always include: constant, yaml, FEATURES.md, tests, lint, nix fmt. I jumped to "Done" after tests only.
2. **Dogfooding verification.** After changing `.golangci.yml`, always run `golangci-lint run` — the tool should lint its own config change.
3. **FEATURES.md as a quality gate.** FEATURES.md line counts and linter enumerations are a common drift point. Consider grepping FEATURES.md for hardcoded counts after any default-list change.

---

## f) Up to 50 Things to Get Done Next

**Immediate fixes from this session:**

1. ~~Update FEATURES.md:69 — change "6 linters" to "7 linters", add `forcetypeassert` to enumeration~~ done (docs-health pass 2026-10-07)
2. ~~Run `golangci-lint run --config=.golangci.yml --timeout=5m` to validate the yaml change~~ done (docs-health pass 2026-10-07)
3. ~~Run `nix fmt` to verify formatting~~ done (docs-health pass 2026-10-07)
4. ~~Run `nix flake check` for full validation~~ done (docs-health pass 2026-10-07)
5. ~~Add a BDD spec verifying `forcetypeassert` is in `DefaultExclusionRules` for `_test.go`~~ done (docs-health pass 2026-10-07)
6. ~~Consider adding a data-integrity test that cross-checks `FEATURES.md` linter counts against `DefaultExclusionRules` (prevents future drift)~~ done (docs-health pass 2026-10-07)

**Default exclusion alignment (from the gap analysis above):** ~~7. Evaluate adding `err113` to `DefaultExclusionRules` (error-wrapping noise in tests)~~ evaluated, not adopted (10-07). ~~8. Evaluate adding `wrapcheck` to `DefaultExclusionRules` (wrapping in test helpers is noise)~~ done — adopted (C2, 2026-07-25 friction session). ~~9. Evaluate adding `varnamelen` to `DefaultExclusionRules` (short names are fine in tests)~~ evaluated, not adopted. ~~10. Evaluate adding `gosec` to `DefaultExclusionRules` (test code uses hardcoded secrets/credentials intentionally)~~ done — adopted (C2). ~~11. Evaluate adding `paralleltest` to `DefaultExclusionRules` (not all tests need t.Parallel)~~ evaluated, not adopted. ~~12. Evaluate adding `revive` to `DefaultExclusionRules` for test files (style rules are noise in tests)~~ evaluated, not adopted. ~~13. Evaluate adding `musttag` to `DefaultExclusionRules` (test structs rarely need tags)~~ evaluated, not adopted. ~~14. Review whether `cyclop` should be removed from defaults (it's in defaults but not in project yaml — inconsistency)~~ done — dissolved: the repo's own config is tool-converged now (self-hosted, v0.11.0), so "project yaml vs defaults" can no longer drift (2026-11-02)

**Documentation & metadata:** ~~15. Check if `docs/references/working-with-codebase.md` references the default exclusion list~~ declined — AGENTS owns the pointers (2026-11-02). ~~16. Check if `docs/references/code-organization.md` references default exclusions~~ declined — same. ~~17. Check if `README.md` mentions default exclusions~~ done — README "Exclusion Patterns" section lists the injected defaults. ~~18. Add a CHANGELOG.md entry if the project maintains one~~ done (CHANGELOG). ~~19. Update AGENTS.md gotcha section if the default exclusion list grows significantly~~ obsolete (conditional).

**Test coverage improvements:** ~~20. Add a fixer BDD spec: "when applying default exclusions, all DefaultExclusionRules linters are present in output config"~~ done — schema-fixture + convergence specs pin the injected defaults. ~~21. Add a test that verifies `DefaultExclusionRules` linter names are valid golangci-lint linter names~~ done — capability probe validates names against the binary's vocabulary; unknown names are removed. ~~22. Add a test that verifies `DefaultExclusionRules` doesn't include disabled linters (`constants.DisabledLinters`)~~ done (`data_integrity_test.go`). ~~23. Add a test that verifies no duplicate linter names within a single `ExclusionRuleConfig`~~ done (`data_integrity_test.go` + `checkDuplicateExclusionLinters` health check)

**Code quality around exclusion system:** ~~24. Consider extracting `DefaultExclusionRules` linter lists into named constants for testability~~ declined — curated data lives on the policy table now. ~~25. Consider adding a `Validate()` method on `DefaultExclusionRules` that checks for invalid linter names~~ declined — init-time table validation + capability probe cover it. ~~26. Consider whether the exclusion merge (`updateExclusionRules`) should log/warn when it adds a new default rule to an existing config (audit trail)~~ done — `--show-merged-rules` + `exclusion-rule-merged` audit action (M25). ~~27. Review whether the `RuleKey()` deduplication correctly handles the case where a user has a `_test.go` rule with a SUBSET of default linters (potential partial-merge issue)~~ done — RuleKey merge unions linter lists (AGENTS detail item 5)

**Broader linter default improvements:** ~~28. Audit all linters in the `reference` preset for test-file appropriateness~~ declined — friction evidence governs test-exclusion membership now. ~~29. Review `pkg/constants/linter_settings.go` `DefaultLinterSettings` for completeness~~ done — the CI schema-fixture gate proves every default is schema-valid; membership is governed by the policy table. ~~30. Check if any newly-added linters in recent golangci-lint versions should have default test exclusions~~ declined — per-linter friction evidence governs. ~~31. Review whether `DefaultFormatterExclusionPaths` should exclude more generated patterns~~ declined-for-now — gogenfilter derives generated-file patterns dynamically. ~~32. Evaluate whether `DefaultLinterExclusionPaths` should include `wire_gen.go` (it's in the project yaml but not in defaults)~~ declined-for-now — gogenfilter's generator presets (wire) cover it

**CI/CD & verification:** ~~33. Add a CI step that checks FEATURES.md linter counts against actual constants (drift prevention)~~ done — `docs_integrity_test.go` runs in CI. ~~34. Add a CI step that runs the tool on its own `.golangci.yml` and verifies no unexpected changes~~ done — dogfood gate (2026-10-07). ~~35. Consider a pre-commit hook that runs `nix fmt` check~~ declined-for-now — CI treefmt/format checks cover the team surface

**Policy & enforcement:** ~~36. Document the policy for what qualifies a linter for `_test.go` exclusion (decision criteria)~~ declined — the criteria are the friction evidence carried on each policy-table record, not a standalone doc. ~~37. Consider whether the disable-reason sidecar should also cover exclusion rules (not just linter disables)~~ declined — sidecar scope is an explicit non-goal (ROADMAP). ~~38. Review whether the audit ledger should record exclusion-rule injections~~ done — audited as `exclusion-rule-merged` with the added linters named

**Longer-term:** ~~39. Consider making `DefaultExclusionRules` configurable via a preset system (minimal/strict/test-friendly)~~ declined-for-now. ~~40. Consider a `--no-default-exclusions` flag for users who want full control~~ declined-for-now. ~~41. Consider surfacing which default exclusions were applied in the HTML/JSON report~~ declined-for-now — report surface stays minimal. ~~42. Add the default exclusion linters to the `audit` subcommand output~~ done — `exclusion-rule-merged` ledger entries name them. ~~43. Consider whether `forcetypeassert` exclusion should also apply to `*_bench_test.go` specifically~~ done — the `_test\.go` path pattern already matches `*_bench_test.go` files. ~~44. Review if the exclusion path regex `_test\.go` should be anchored (`_test\.go$`) for consistency with other path patterns~~ declined — unanchored RE2 full-path matching works; re-anchoring would churn every config. ~~45. Consider adding context-aware exclusions (e.g., exclude `forcetypeassert` only in test helper functions, not test table entries)~~ declined — YAGNI

**Nice-to-haves:** ~~46. Add a `configure --explain-exclusions` flag that documents why each default exclusion exists~~ declined-for-now. ~~47. Consider linking linter exclusion decisions to their rationale (like `linter_reasons.go` but for exclusions)~~ declined — friction evidence on the table is the rationale. ~~48. Add exclusion rules to the HTML report's "recommendations" section~~ declined-for-now. ~~49. Consider community-sourced exclusion presets (curated test-file exclusions)~~ declined. ~~50. Write a blog post / docs page on "Why we exclude these 7+ linters from test files"~~ declined

---

## g) Questions (cannot figure out myself)

### 1. Should we align `DefaultExclusionRules` with the project's own richer `.golangci.yml` exclusions?

The project excludes `err113`, `wrapcheck`, `varnamelen`, `gosec`, `paralleltest`, `revive`, and `musttag` from `_test.go` in its own config but doesn't ship these as defaults. Should we promote some/all of these to `DefaultExclusionRules`, or are they intentionally project-specific? This is a product/policy decision I can't infer from code alone.

> **RESOLVED:** partially — `wrapcheck`, `gosec`, `errcheck` (C2) and `ireturn`, `recvcheck`, `contextcheck`, `exhaustive` (C2b) were promoted on friction evidence (2026-07-25); `err113`/`varnamelen`/`paralleltest`/`revive`/`musttag` were evaluated and consciously not adopted (2026-10-07). Friction evidence is the standing criterion.

### 2. Should this change be committed now, or batched with the FEATURES.md fix and broader exclusion review?

I can fix FEATURES.md immediately, but if we're also going to do the gap-analysis work (items 7-14), it might make sense to batch everything into one commit with a complete default-exclusion overhaul. Your call on commit granularity.

> **RESOLVED:** moot — committed and superseded by three months of subsequent history (the FEATURES fix landed via the docs-health pass).

### 3. Is the `forcetypeassert` exclusion universally desired, or should it be opt-in via a preset?

`forcetypeassert` is currently only in the `reference` preset — meaning only users who explicitly select that preset have it enabled. Adding it to `DefaultExclusionRules` means ALL users with `forcetypeassert` enabled get it excluded from tests. Is that the right scope, or should the exclusion be tied to preset selection?

> **RESOLVED:** universal — `DefaultExclusionRules` applies whenever the linter is enabled, independent of preset; pinned by the docs-integrity + schema-fixture specs.

---

## Summary

The core change is correct and tested. The main miss is **FEATURES.md documentation drift** — I declared "Done" without updating the feature inventory that explicitly enumerates the default exclusion linters. Secondary misses: no lint run, no `nix fmt`, no dedicated test. The deeper opportunity is closing the gap between the project's own test-file exclusions and what ships as defaults to users.
