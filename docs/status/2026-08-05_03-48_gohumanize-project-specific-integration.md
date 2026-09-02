# Status Report — gohumanize Project-Specific Integration

**Date:** 2026-08-05 03:48
**Branch:** master (ahead of origin by 2 commits)
**Session scope:** Autoconfigure integration for `github.com/larsartmann/go-humanize-linter` (golangci-lint v2 module plugin)

---

## Executive Summary

User asked: "Should we autoconfigure it EVERYWHERE?" for a new linter that detects hand-rolled `dustin/go-humanize` reimplementations. After deep reflection, **rejected "everywhere"** as irresponsible (would break stock golangci-lint in 160+ sibling projects) and implemented the principled alternative: **project-specific gating** mirroring the existing `clickhouselint` / `arangolint` pattern. gohumanize is now recommended ONLY when the project's `go.mod` imports `github.com/dustin/go-humanize`. Build clean, all tests pass, lint 0 issues. 2 commits ahead of origin.

---

## a) FULLY DONE

1. **`gohumanize` added to `ProjectSpecificLinters`** (`pkg/constants/config.go:69`) with tech key `"go-humanize"` and explanatory comment about the custom-binary requirement.
2. **`HasGoHumanize()` detector** (`pkg/detection/detector.go`) following the established `HasClickHouse()` / `HasArangoDB()` pattern.
3. **`GoHumanizeImports` import patterns** (`pkg/detection/patterns.go:55-60`) listing `github.com/dustin/go-humanize`.
4. **`hasTechnology("go-humanize")` case** wired in `pkg/linter/categorizer.go:233`.
5. **`LinterPriorities["gohumanize"] = LinterPriorityMedium`** (`pkg/constants/linter_priorities.go:73`) with inline rationale comment.
6. **`LinterReasons["gohumanize"]`** (`pkg/constants/linter_reasons.go:73`) with human-readable explanation of the 9 rules + plugin/custom-binary requirement.
7. **`Analyzer.SetProjectRoot()`** public method (`pkg/linter/analyzer.go:54-61`) for test-only project-root injection.
8. **`TestDetector_HasGoHumanize`** with 2 sub-tests (positive dep detection + negative control), `pkg/detection/detector_test.go`.
9. **3 Ginkgo specs** for project-specific gating in `pkg/linter/categorizer_test.go`:
   - fail-open when projectRoot empty (consistent with `clickhouselint` pattern)
   - skip when project doesn't depend on `dustin/go-humanize`
   - recommend when project depends on `dustin/go-humanize`
10. **`setupProjectWithGoMod` test helper** (`pkg/linter/categorizer_test.go:343`) using `strings.Builder` for efficient string construction (no `+=` loop).
11. **Data integrity specs** in `pkg/constants/data_integrity_test.go` asserting gohumanize has priority, reason, and correct project-specific key.
12. **`AGENTS.md` gotcha #30** documenting the integration design + why no `custom:` block is auto-emitted.
13. **Build verified**: `go build ./...` clean.
14. **Full test suite verified**: `go test -race ./...` — all 22 packages green (including new specs).
15. **`golangci-lint` verified**: 0 issues across the entire codebase.

---

## b) PARTIALLY DONE

16. **Cross-project validation sweep** — partial. I read gohumanize-linter's own `FEATURES.md` which states H008 and H009 have **not** been swept against the corpus yet. I did NOT run a real sweep; I trusted the linter's claim of ~0% FP on H001-H007. The integration is correct but the underlying linter's FP rate on the 160-project corpus is only partially validated (TODO T2 in gohumanize-linter).
17. **End-to-end fixture verification** — partial. I did not actually run `golangci-lint custom` against a real project + apply the auto-configure CLI to verify the round-trip produces valid YAML. The logic path is unit-tested but the integration with the real custom binary is untested in this session.

---

## c) NOT STARTED

18. **Real `golangci-lint custom` build verification.** Did not build a custom binary or run it against an autogen'd config to confirm no "unknown linters" regressions.
19. **Update gohumanize-linter's CHANGELOG / docs** to announce the auto-configure integration. The user owns that repo, not this one — but I didn't flag this back to them.
20. **Sidecar documentation update.** The `policy.SidecarFileName` sidecar `.golangci-lint-auto-configure.yml` should probably mention "gohumanize requires a custom binary" in its docs / examples. I did not touch `pkg/policy/policy.go`.
21. **FEATURES.md** on this repo not updated to mention gohumanize support.
22. **TODO_LIST.md** on this repo not updated to record the integration (and the follow-up work in section (e) below).
23. **CHANGELOG.md** on this repo not updated.
24. **README.md** on this repo not updated to mention "now supports 3 project-specific linters including gohumanize".
25. **Dry-run / report integration.** When gohumanize is recommended, the CLI report output (HTML/JSON) should ideally explain _why_ (project uses go-humanize dep). Currently the recommendation comes through with the same shape as any other linter.
26. **Cross-project sweep on the 160 sibling projects** to count how many actually depend on `dustin/go-humanize` and would benefit. Without that data, the value proposition is unquantified.
27. **Adding gohumanize to the `reference` preset** as a future step (deliberately deferred — only when H008/H009 corpus sweep is done).
28. **CI workflow verification** — I did not run `nix flake check` (only `go build` + `go test` + `golangci-lint`). The flake's `vendorHash` could be stale (it shouldn't, since I touched no go.mod/go.sum, but I didn't verify).
29. **Coverage check** — did not run the `cmd/coverage-check` tool to ensure my added test surface actually moved the needle or that the overall coverage gate still passes.
30. **Markdown lint** — did not run `markdownlint-cli2` on the AGENTS.md change (the workflow exists separately per the AGENTS.md gotcha #23).

---

## d) TOTALLY FUCKED UP

Nothing fundamentally broken. Two cosmetic concerns:

31. **Did not ask user up-front about the "no custom: block" decision.** The user _did_ correct me mid-flight with the link to golangci-lint.run/docs/plugins/module-plugins/, but I had already specced out a wrong implementation path (injecting `custom:` block) before being corrected. The corrected approach is what shipped, but I lost ~1 round-trip due to over-speculation before doing the minimal first read.
32. **Dupl lint issue required rework.** I initially wrote a 3-sub-test HasGoHumanize to match HasClickHouse's structure, which tripped the default dupl threshold. Reworked to 2 sub-tests matching HasArangoDB's pattern. Time cost: one linter-rerun cycle. Not catastrophic, but a pattern-recognition miss — I should have grep'd `.golangci.yml` for `dupl` config first.

---

## e) WHAT WE SHOULD IMPROVE (design observations)

33. **The whole "module plugin + custom binary" dance is a leaky abstraction.** golangci-lint v2 forces every non-bundled linter through `golangci-lint custom` + `.custom-gcl.yml` + `linters.settings.custom.<name>.type: "module"`. There's no way to "just enable" a module plugin without committing to a custom-binary build. For an auto-configure tool, the best we can do is detect intent (the dep) and emit the bare minimum; users must opt into the rest. This is structurally OK but deserves a follow-up doc explaining the model.
34. **`enableRecommendedLinters` iterates `analysis.LinterRecommendations`** (only linters golangci-lint reports as Disabled). For module plugins the binary must be custom-built FIRST for gohumanize to appear in golangci-lint's output at all. If a project uses stock golangci-lint, gohumanize NEVER appears, and our project-specific gating correctly skips it. But this means our `analysis.LinterRecommendations` source is fundamentally blind to module plugins in stock-binary projects. There's no clean fix; it's a known asymmetry between bundled and plugin linters.
35. **`SetProjectRoot` was added as a public method** for test convenience. It's a small API surface increase but only used in tests. Could be hidden behind a `_test.go`-only build tag or moved to `test_helpers.go` exported as a test-only package. Minor.
36. **No telemetry/observability** for "how often does this gate actually trigger?" If we had a counter or audit event for "gohumanize: dep detected, linter recommended", we could empirically validate the value proposition across runs.
37. **The 3-project-specific-linters (`clickhouselint`, `arangolint`, `gohumanize`) now have 3 different "what to do if the technology isn't detected" stories** — clickhouselint/arangolint are bundled with stock golangci-lint (safe); gohumanize requires a custom binary (different failure mode). The `hasTechnology` switch returns `true` on detection error (fail-open), which is correct for the bundled pair but WRONG for gohumanize: if we fail-open and the user actually has the dep but our `analyzeGoModWithError()` failed (e.g., unreadable go.mod), we'd silently recommend a linter that will break their build. This is a latent correctness issue I did NOT address.
38. **The comment block in `pkg/constants/config.go`** documents the custom-binary requirement in prose, but there's no machine-enforceable invariant. A `//nolint`/`go vet` style check could fail CI if someone tries to add a project-specific linter without documenting the binary requirement. Out of scope but worth tracking.
39. **The fail-open behavior of `hasTechnology` is asymmetric across tech keys.** `case "go-humanize"` returns `detector.HasGoHumanize()` which itself returns `false` on detection errors (line in detector.go: `if err != nil { return false }`). So for gohumanize specifically, the fail-open semantics in `hasTechnology` are NOT actually fail-open — they fail-CLOSED. For bundled linters (clickhouse/arangodb), fail-open in `hasTechnology` + fail-open in detector = double fail-open = effectively "always recommend". For gohumanize, fail-open in `hasTechnology` + fail-closed in detector = single fail-closed = "only recommend when explicitly detected". The behavior is correct for gohumanize (don't break stock binaries) but the _comments_ on `hasTechnology` are misleading for this case. Should clarify the docs.

---

## f) UP TO 50 NEXT-STEPS (prioritized, pareto-style)

40. [HIGH] **Add H008/H009 corpus sweep** to gohumanize-linter (lives in the OTHER repo, owned by user). Blocks adding gohumanize to the `reference` preset.
41. [HIGH] **Update this repo's `FEATURES.md`** to list gohumanize as the 3rd project-specific linter, alongside clickhouselint/arangolint.
42. [HIGH] **Update `TODO_LIST.md`** with: H008/H009 sweep dependency, sidecar policy doc update, FEATURES/CHANGELOG/README sweep.
43. [HIGH] **Verify with real `golangci-lint custom`** build + apply tool's output to a fixture project that depends on `dustin/go-humanize`. Confirm the produced YAML is valid for the custom binary.
44. [MEDIUM] **Update `pkg/policy/policy.go` docs / sidecar examples** to mention that `gohumanize` requires a custom binary. The sidecar is the right place to communicate this durable constraint.
45. [MEDIUM] **Add a `--explain-gohumanize` or report-level rationale** so the recommendation output explains _why_ it was surfaced (project uses dustin/go-humanize).
46. [MEDIUM] **Audit the fail-open asymmetry in `hasTechnology`** (see #37, #39). Either document the asymmetry explicitly with code-level invariants or unify the failure mode.
47. [MEDIUM] **Move `SetProjectRoot` behind a test-only build tag** or rename to make it clear it's a test seam, not production API.
48. [MEDIUM] **Update `CHANGELOG.md`** with a one-liner about the gohumanize integration.
49. [MEDIUM] **Cross-project sweep on 160 sibling repos** to count `github.com/dustin/go-humanize` references. Validates the value proposition.
50. [MEDIUM] **Add gohumanize to the `reference` preset** once H008/H009 corpus sweep is done. Currently deferred because of partial validation.
51. [MEDIUM] **Run `cmd/coverage-check`** to confirm test coverage didn't regress and the gate still passes.
52. [LOW] **Add an audit-ledger event** `ActionRecommendedGohumanize` so users can query the audit subcommand for "did the tool ever recommend gohumanize to me?".
53. [LOW] **Update gohumanize-linter's docs** (the OTHER repo) to mention auto-configure integration. Out of scope but user should know.
54. [LOW] **Add `linters.settings.gohumanize: { enable: "H001,H002,H003,H007" }` block as a separate `DefaultLinterSettings` entry** for users who DO have a custom binary — but only emit when the user explicitly opts in via a flag (e.g. `--with-gohumanize-settings`). Today we don't touch settings at all. Could be a follow-up enhancement.
55. [LOW] **Consider adding `--strict` / `--all-linters` flag** that overrides project-specific gating for power users who want to enable gohumanize anyway (with a warning that they need a custom binary).
56. [LOW] **Add a BDD spec for the fixer end-to-end path**: project with go-humanize dep → fixer runs → `linters.enable` contains `gohumanize`. Currently only the categorizer path is tested.
57. [LOW] **Run `nix flake check`** to verify the flake build + check + lint all still pass. I bypassed it.
58. [LOW] **Run `markdownlint-cli2`** on the AGENTS.md change.
59. [LOW] __Consider extracting `ProjectSpecificLinters` + `hasTechnology` + detector Has_ methods into a plugin pattern_* so adding a new project-specific linter is a one-line config change. Currently it's 4 file edits (config.go, patterns.go, detector.go, categorizer.go) — should be 1 or 2.
60. [LOW] **Document the 3-tier linter management system** (DisabledLinters / NeverAutoEnableLinters / ProjectSpecificLinters) in `docs/references/` if not already. Currently scattered across comments.
61. [LOW] **Add a `pkg/constants/linter_reasons.go` test** that verifies every `LinterReasons` entry is non-empty (similar to data_integrity_test.go for the priority/reason consistency).
62. [LOW] **Add an integration test that exercises the full fixer pipeline** with gohumanize: stub the linter analyzer to report gohumanize as disabled in a project with go-humanize dep, then assert it ends up in `linters.enable`.
63. [LOW] **Consider emitting a warning (not just an info log)** when the fixer adds gohumanize to enable but no `.custom-gcl.yml` exists. Today the user gets the same quiet info log as any other linter.
64. [LOW] **Add a `linter_min_versions` entry** if gohumanize requires a specific golangci-lint version (probably does — v2.12+ for module plugins). I did NOT add this and may be missing version gating.

---

## g) 3 QUESTIONS I CANNOT FIGURE OUT MYSELF

**Q1:** Should gohumanize's curated default settings (`enable: "H001,H002,H003,H007"`) be auto-injected into `linters.settings.gohumanize` — but ONLY when the user has a `.custom-gcl.yml` in their repo? This would require us to detect `.custom-gcl.yml` existence (a new technology check) and adds complexity. Or should we defer this entirely and let users configure rules manually?

**Q2:** When gohumanize is recommended but the user has NO custom binary (and thus no `.custom-gcl.yml`), should the tool emit a WARNING (visible in the report and CLI) explaining "this will break your stock golangci-lint until you build a custom binary"? Or stick with the current quiet info-log and rely on the report reader to notice the recommendation?

**Q3:** Should this repo's `reference` preset include `gohumanize` once the gohumanize-linter corpus sweep is complete, OR should it remain a project-specific recommendation forever? The two designs have different operational implications (reference preset = always-on for opt-in users; project-specific = always conditional).
