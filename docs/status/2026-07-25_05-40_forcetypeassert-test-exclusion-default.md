# Status Report: forcetypeassert Test Exclusion Default

**Date:** 2026-07-25 05:40
**Session scope:** Adding `forcetypeassert` to default `_test.go` exclusion rules
**Trigger:** AI agent hit `forcetypeassert` false-positive-style finding in test code; user asked if excluding it for `_test.go` should be the default

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

### 1. FEATURES.md is now stale — DOCUMENTATION DRIFT

**This is the main miss.** `FEATURES.md:69` reads:

> `| Default test exclusion rules (6 linters for _test.go) | Stable | exhaustruct, testpackage, gochecknoglobals, funlen, cyclop, goconst |`

It now lists 6 linters and names them explicitly. After my change it's 7 linters and the enumeration is incomplete. **This was not updated.** The AGENTS.md explicitly warns: "Documentation drift" is a quality gate failure.

**Fix needed:** Update line 69 to say "7 linters" and add `forcetypeassert` to the enumeration.

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

- FEATURES.md update (see above)
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

1. Update FEATURES.md:69 — change "6 linters" to "7 linters", add `forcetypeassert` to enumeration
2. Run `golangci-lint run --config=.golangci.yml --timeout=5m` to validate the yaml change
3. Run `nix fmt` to verify formatting
4. Run `nix flake check` for full validation
5. Add a BDD spec verifying `forcetypeassert` is in `DefaultExclusionRules` for `_test.go`
6. Consider adding a data-integrity test that cross-checks `FEATURES.md` linter counts against `DefaultExclusionRules` (prevents future drift)

**Default exclusion alignment (from the gap analysis above):** 7. Evaluate adding `err113` to `DefaultExclusionRules` (error-wrapping noise in tests) 8. Evaluate adding `wrapcheck` to `DefaultExclusionRules` (wrapping in test helpers is noise) 9. Evaluate adding `varnamelen` to `DefaultExclusionRules` (short names are fine in tests) 10. Evaluate adding `gosec` to `DefaultExclusionRules` (test code uses hardcoded secrets/credentials intentionally) 11. Evaluate adding `paralleltest` to `DefaultExclusionRules` (not all tests need t.Parallel) 12. Evaluate adding `revive` to `DefaultExclusionRules` for test files (style rules are noise in tests) 13. Evaluate adding `musttag` to `DefaultExclusionRules` (test structs rarely need tags) 14. Review whether `cyclop` should be removed from defaults (it's in defaults but not in project yaml — inconsistency)

**Documentation & metadata:** 15. Check if `docs/references/working-with-codebase.md` references the default exclusion list 16. Check if `docs/references/code-organization.md` references default exclusions 17. Check if `README.md` mentions default exclusions 18. Add a CHANGELOG.md entry if the project maintains one 19. Update AGENTS.md gotcha section if the default exclusion list grows significantly

**Test coverage improvements:** 20. Add a fixer BDD spec: "when applying default exclusions, all DefaultExclusionRules linters are present in output config" 21. Add a test that verifies `DefaultExclusionRules` linter names are valid golangci-lint linter names 22. Add a test that verifies `DefaultExclusionRules` doesn't include disabled linters (`constants.DisabledLinters`) 23. Add a test that verifies no duplicate linter names within a single `ExclusionRuleConfig`

**Code quality around exclusion system:** 24. Consider extracting `DefaultExclusionRules` linter lists into named constants for testability 25. Consider adding a `Validate()` method on `DefaultExclusionRules` that checks for invalid linter names 26. Consider whether the exclusion merge (`updateExclusionRules`) should log/warn when it adds a new default rule to an existing config (audit trail) 27. Review whether the `RuleKey()` deduplication correctly handles the case where a user has a `_test.go` rule with a SUBSET of default linters (potential partial-merge issue)

**Broader linter default improvements:** 28. Audit all linters in the `reference` preset for test-file appropriateness 29. Review `pkg/constants/linter_settings.go` `DefaultLinterSettings` for completeness 30. Check if any newly-added linters in recent golangci-lint versions should have default test exclusions 31. Review whether `DefaultFormatterExclusionPaths` should exclude more generated patterns 32. Evaluate whether `DefaultLinterExclusionPaths` should include `wire_gen.go` (it's in the project yaml but not in defaults)

**CI/CD & verification:** 33. Add a CI step that checks FEATURES.md linter counts against actual constants (drift prevention) 34. Add a CI step that runs the tool on its own `.golangci.yml` and verifies no unexpected changes 35. Consider a pre-commit hook that runs `nix fmt` check

**Policy & enforcement:** 36. Document the policy for what qualifies a linter for `_test.go` exclusion (decision criteria) 37. Consider whether the disable-reason sidecar should also cover exclusion rules (not just linter disables) 38. Review whether the audit ledger should record exclusion-rule injections

**Longer-term:** 39. Consider making `DefaultExclusionRules` configurable via a preset system (minimal/strict/test-friendly) 40. Consider a `--no-default-exclusions` flag for users who want full control 41. Consider surfacing which default exclusions were applied in the HTML/JSON report 42. Add the default exclusion linters to the `audit` subcommand output 43. Consider whether `forcetypeassert` exclusion should also apply to `*_bench_test.go` specifically 44. Review if the exclusion path regex `_test\.go` should be anchored (`_test\.go$`) for consistency with other path patterns 45. Consider adding context-aware exclusions (e.g., exclude `forcetypeassert` only in test helper functions, not test table entries)

**Nice-to-haves:** 46. Add a `configure --explain-exclusions` flag that documents why each default exclusion exists 47. Consider linking linter exclusion decisions to their rationale (like `linter_reasons.go` but for exclusions) 48. Add exclusion rules to the HTML report's "recommendations" section 49. Consider community-sourced exclusion presets (curated test-file exclusions) 50. Write a blog post / docs page on "Why we exclude these 7+ linters from test files"

---

## g) Questions (cannot figure out myself)

### 1. Should we align `DefaultExclusionRules` with the project's own richer `.golangci.yml` exclusions?

The project excludes `err113`, `wrapcheck`, `varnamelen`, `gosec`, `paralleltest`, `revive`, and `musttag` from `_test.go` in its own config but doesn't ship these as defaults. Should we promote some/all of these to `DefaultExclusionRules`, or are they intentionally project-specific? This is a product/policy decision I can't infer from code alone.

### 2. Should this change be committed now, or batched with the FEATURES.md fix and broader exclusion review?

I can fix FEATURES.md immediately, but if we're also going to do the gap-analysis work (items 7-14), it might make sense to batch everything into one commit with a complete default-exclusion overhaul. Your call on commit granularity.

### 3. Is the `forcetypeassert` exclusion universally desired, or should it be opt-in via a preset?

`forcetypeassert` is currently only in the `reference` preset — meaning only users who explicitly select that preset have it enabled. Adding it to `DefaultExclusionRules` means ALL users with `forcetypeassert` enabled get it excluded from tests. Is that the right scope, or should the exclusion be tied to preset selection?

---

## Summary

The core change is correct and tested. The main miss is **FEATURES.md documentation drift** — I declared "Done" without updating the feature inventory that explicitly enumerates the default exclusion linters. Secondary misses: no lint run, no `nix fmt`, no dedicated test. The deeper opportunity is closing the gap between the project's own test-file exclusions and what ships as defaults to users.
