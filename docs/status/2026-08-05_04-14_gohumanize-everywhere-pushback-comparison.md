# Status Report — gohumanize "Everywhere" Pushback & Plugin System Comparison

**Date:** 2026-08-05 04:14
**Branch:** master (ahead of origin by 3 commits)
**Session scope:** Response to user's "I DO WANT it everywhere" pushback + technical comparison of golangci-lint's two plugin systems.

---

## Executive Summary

User rejected the project-specific gating design and asked for gohumanize "EVERYWHERE." Then pushed back a second time on the fundamental premise — "WHY do we need ANY custom build?" I produced an honest, evidence-based comparison of the two plugin systems (module plugins vs Go `plugin` package) using the official golangci-lint docs and the gohumanize-linter's actual code. Concluded: there's no third path that avoids the custom-build requirement, because gohumanize-linter was built on golangci-lint's recommended path. Three concrete options for "everywhere" surfaced as questions. Awaiting user direction. **No new code shipped this iteration** — this was an analysis-only round, but it's the kind of analysis that prevents shipping the wrong solution.

---

## a) FULLY DONE (this round — analysis-only)

1. **Re-fetched golangci-lint official docs** from `golangci/golangci-lint` GitHub raw content (not the JS-rendered site) for both:
   - `/docs/plugins/module-plugins.md` (confirmed module plugin requires `golangci-lint custom`)
   - `/docs/plugins/go-plugins.md` (confirmed legacy Go plugin system with `path: *.so`)
2. **Verified gohumanize-linter's actual plugin mechanism** by reading `go.mod`, `plugin/plugin.go`, and `plugin/plugin_integration_test.go` — confirmed it uses `github.com/golangci/plugin-module-register v0.1.2` (module plugin, not Go plugin).
3. **Produced a 12-dimension comparison table** covering: stock-binary support, build step, CGO, version matching, cross-platform, CI, golangci-lint's official stance, distribution, config schema, failure mode, discoverability, and source format.
4. **Identified three concrete options** for "everywhere" and surfaced them as questions to the user.

## a) FULLY DONE (cumulative — this session)

5. `gohumanize` added to `ProjectSpecificLinters` (`pkg/constants/config.go:69`) with tech key `"go-humanize"`.
6. `HasGoHumanize()` detector (`pkg/detection/detector.go`) following the established `HasClickHouse()` / `HasArangoDB()` pattern.
7. `GoHumanizeImports` import patterns (`pkg/detection/patterns.go:55-60`) listing `github.com/dustin/go-humanize`.
8. `hasTechnology("go-humanize")` case wired in `pkg/linter/categorizer.go:233`.
9. `LinterPriorities["gohumanize"] = LinterPriorityMedium` with inline rationale comment.
10. `LinterReasons["gohumanize"]` with human-readable explanation.
11. `Analyzer.SetProjectRoot()` public method for test-only project-root injection.
12. `TestDetector_HasGoHumanize` with 2 sub-tests (positive + negative control).
13. 3 Ginkgo specs for project-specific gating: fail-open on empty projectRoot, skip without dep, recommend with dep.
14. `setupProjectWithGoMod` test helper using `strings.Builder` (no `+=` loop, no dupl violation).
15. Data integrity specs in `pkg/constants/data_integrity_test.go` asserting gohumanize metadata.
16. `AGENTS.md` gotcha #30 documenting the integration design.
17. **Build verified**: `go build ./...` clean.
18. **Full test suite verified**: `go test -race ./...` — all 22 packages green.
19. **`golangci-lint` verified**: 0 issues across the entire codebase.
20. **Status report** at `docs/status/2026-08-05_03-48_gohumanize-project-specific-integration.md` covering the first iteration.

---

## b) PARTIALLY DONE

21. **Pushback response is honest but doesn't satisfy user.** The user wants everywhere. I gave them evidence that everywhere is operationally expensive (one-time custom build) but the actual friction is unavoidable. This is correct but doesn't move the needle — the user may want me to ship the gating removal + trust the user to handle the build, regardless.
22. **Question framing didn't elicit a clear next-step choice.** I asked three concrete options. The user's response was a _meta-question_ ("Why do we need any custom build?") rather than picking one of the three. I'm now blocked on a clarification of whether the user accepts the architectural constraint or wants me to find a creative workaround.
23. **No code shipped this round.** The technical analysis consumed a round but produced no commits toward "everywhere."

---

## c) NOT STARTED

24. **Remove the project-specific gating** (`ProjectSpecificLinters` entry + `HasGoHumanize()` + `hasTechnology` case) if user wants true everywhere. Currently blocking on user direction.
25. **Decide on priority level** for everywhere mode (Medium vs Critical). My previous question surfaced this — no answer yet.
26. **Decide on curated subset** (auto-inject `enable: "H001,H002,H003,H007"` settings) — yes/no.
27. **Add `gohumanize` to the `reference` preset** as an explicit "always include" entry, separate from any recommendation logic.
28. **Build `.custom-gcl.yml` template generator** — a subcommand or template that emits the build config so users have a one-liner to get the custom binary.
29. **Update gohumanize-linter's docs** to recommend the auto-configure integration (out of scope but not flagged to user).
30. **Run `nix flake check`** — bypassed all session.
31. **Run `cmd/coverage-check`** — bypassed all session.
32. **Run `markdownlint-cli2`** on AGENTS.md change.
33. **Update `FEATURES.md`** to list the gohumanize integration.
34. **Update `TODO_LIST.md`** with new "everywhere" follow-up work.
35. **Update `CHANGELOG.md`**.
36. **Update `README.md`** to mention the third project-specific linter OR the new everywhere recommendation.
37. **Cross-project sweep** on 160 sibling repos to count `github.com/dustin/go-humanize` usage.
38. **Real `golangci-lint custom` build verification** — never run end-to-end.
39. **Audit the fail-open asymmetry** in `hasTechnology` between bundled linters (clickhouselint/arangolint) and plugin linters (gohumanize).
40. **Add BDD end-to-end fixer test** with gohumanize in the full pipeline.
41. **Add audit-ledger event** `ActionRecommendedGohumanize` for telemetry.
42. **Document the 3-tier linter management system** (DisabledLinters / NeverAutoEnableLinters / ProjectSpecificLinters).
43. **Add an integration test for HasGoHumanize edge cases** — vendor directory detection, replace directives, transitive deps.

---

## d) TOTALLY FUCKED UP

44. **Didn't read the gohumanize-linter AGENTS.md or the module plugin section thoroughly on round 1.** I should have caught the "module plugin = mandatory custom build" fact from gohumanize-linter's own AGENTS.md immediately. Instead I read the .custom-gcl.yml + .golangci.custom.yml which made the constraint visible but I framed it as a "blocker" rather than "this is the architecture, design within it." The user's "WHY do we need ANY custom build" was a fair callout that I should have pre-empted with a comparison table on round 1.
45. **Framed the comparison as a wall, not as options.** The 12-dimension table I shipped is good but I led with "no third path exists" — that's a defensive framing. A better framing: "here are 3 paths to 'everywhere', each with different operational costs, pick one." The user then had to push back twice to get me to enumerate options. I should have enumerated first.
46. **The original question I asked the user had character limits that truncated my reasoning.** The first question's choice descriptions hit a 200-char limit and got rejected. I shrunk them and lost nuance. Could have just laid out the options as a markdown table in the prompt and asked the user to pick by number.
47. **Pivoted too late on the user's "I want it everywhere" stance.** The first round I built the gating with extensive tests (commit `a0ade48`). Now that the user has rejected it, that work is partially wasted — it'll need to be unwound if we go to true everywhere. Should have asked the user the _strategy_ question BEFORE writing 8 files of tests.

---

## e) WHAT WE SHOULD IMPROVE

48. **Surface architectural constraints as OPTIONS, not BLOCKERS, on the first response.** "Module plugins require X" → "here are 3 ways to ship gohumanize: project-specific, opt-in via flag, always-on with curated subset. Each has these costs. Which do you want?"
49. **Read the upstream linter's AGENTS.md FIRST** when integrating a third-party linter. Saves 2 round-trips.
50. **Default behavior should be a separate decision from test coverage.** Build the minimum first (one config entry), get the user's strategic direction, THEN build the tests. I built tests first because the existing patterns invited it, but the patterns exist for established decisions, not for new ones still in flux.
51. **Plugin mechanism taxonomy should be in `AGENTS.md`** as a gotcha. We now have THREE patterns in the codebase (bundled linter, module plugin via custom binary, Go plugin via .so). Each has different config schemas, different build flows, different failure modes. A future contributor adding the 4th plugin linter will hit the same confusion I did. Document it.
52. **The `hasTechnology` fail-open asymmetry needs explicit docs.** Three different linter categories, three different correct failure modes. Currently the code returns `true` (fail-open) for ALL cases except the detector itself, which fails closed. For module plugins this is fine; for bundled linters it's double-fail-open. The comment says "fail-open" but doesn't qualify what "open" means per category.
53. **Should we have a `--plugin-linters=auto|none|all` flag?** A meta-flag that controls how aggressively the tool surfaces plugin linters. Default `auto` = current behavior (gating on dep). `all` = everywhere. `none` = never. This would give users a clean knob without code changes.
54. **The 160 sibling project claim is unverified.** I've been citing it throughout the session as if it's a hard constraint, but I haven't actually run a sweep to count how many sibling projects exist, how many use go-humanize, and how many have broken stock golangci-lint when a module plugin gets added. The "160 projects" number came from the AGENTS.md; treat as approximate.

---

## f) UP TO 50 NEXT-STEPS (prioritized)

55. **[BLOCKED — needs user answer]** Decide strategy: remove gating → everywhere (Medium, no curated subset); or → everywhere (Critical, with curated subset); or → opt-in flag; or → leave project-specific and document the architecture better.
56. **[HIGH, after decision]** Implement chosen strategy. If "remove gating": strip the `ProjectSpecificLinters` entry + `HasGoHumanize()` + `hasTechnology("go-humanize")` case + the 3 categorizer specs + the 2 detection specs (or keep specs as positive coverage of the detector still working in isolation). If "opt-in flag": add `--with-gohumanize` to flags struct + categorizer path.
57. **[HIGH, after decision]** Add `gohumanize` to the `reference` preset explicitly, so `--preset reference` users get it as part of the curated set.
58. **[HIGH]** Update `FEATURES.md`, `TODO_LIST.md`, `CHANGELOG.md`, `README.md` to reflect the chosen strategy. The "project-specific" framing in `AGENTS.md` gotcha #30 will need rewriting if we go everywhere.
59. **[HIGH]** Ship a `gocilint-template` or document the exact `golangci-lint custom` command + `.custom-gcl.yml` content needed so users can build the custom binary in one step.
60. **[MEDIUM]** Add a `--plugin-linters=auto|none|all` meta-flag (per improvement #53) for forward-proofing future module plugin integrations.
61. **[MEDIUM]** Audit and document the fail-open asymmetry (per improvement #52) — add explicit code comments per technology case explaining the correct failure mode.
62. **[MEDIUM]** End-to-end fixer test: project with go-humanize dep → full fixer pipeline → assert `linters.enable` contains `gohumanize`. Currently only categorizer is unit-tested.
63. **[MEDIUM]** Real-world sweep: count `dustin/go-humanize` imports across `~/projects/`. Validates "everywhere" value proposition empirically.
64. **[MEDIUM]** Add `linters_min_versions` entry for gohumanize (requires golangci-lint v2.x for module plugin support) — protects against accidentally recommending for old golangci-lint versions.
65. **[MEDIUM]** Add audit-ledger event `ActionRecommendedGohumanize` for telemetry.
66. **[MEDIUM]** Document the 3-tier linter management system in `docs/references/` — currently scattered across comments.
67. **[LOW]** Add integration test for `HasGoHumanize` edge cases: vendor dir, replace directives, transitive deps.
68. **[LOW]** Run `nix flake check` end-to-end (formatting + build + lint + test).
69. **[LOW]** Run `cmd/coverage-check` to verify test coverage gate.
70. **[LOW]** Run `markdownlint-cli2` on the AGENTS.md change.
71. **[LOW]** Consider extracting `ProjectSpecificLinters` + `hasTechnology` + detector `Has*` methods into a plugin pattern so adding a new project-specific linter is one config entry, not 4 file edits.
72. **[LOW]** Add a BDD spec for the fixer's `linters.settings.gohumanize` block emission (if we decide to inject curated settings).
73. **[LOW]** Consider whether `--pragmatic` should drop gohumanize (currently no — it's not in `PragmaticNoiseLinters`).
74. **[LOW]** Cross-check `DisabledLinters` / `NeverAutoEnableLinters` / `PragmaticNoiseLinters` to confirm gohumanize is correctly absent from all three (currently correct, but no test asserts it).
75. **[LOW]** Verify `data_integrity_test.go` "gohumanize linter metadata" specs still apply after strategy change.
76. **[LOW]** Consider whether `policy.go` sidecar should support a `module-plugins:` section that documents required custom-binary builds.

---

## g) 3 QUESTIONS I CANNOT FIGURE OUT MYSELF

**Q1:** Which "everywhere" do you want? Remove gating entirely (always-on, no curated subset, no dep check), OR keep some structure (curated subset + Critical priority + opt-in via flag)? The "everywhere" is ambiguous between "I want it recommended always" vs "I want it added to everyone's config unconditionally." These have different blast radii.

**Q2:** If everywhere, should the auto-configure tool ALSO emit the `.custom-gcl.yml` build config + a `make custom-gcl` or `nix run .#build-custom-gcl` target, so users get a one-command path to a working custom binary? This adds significant scope (we become a wrapper around golangci-lint's build system) but solves the friction that prompted your pushback.

**Q3:** Are you OK with shipping the integration in a state where "if you run `golangci-lint` (stock binary) on an auto-configured project that uses gohumanize, it errors out with 'unknown linters: gohumanize'"? Because that's the unavoidable consequence of "everywhere" with the current golangci-lint architecture. Either (a) accept the breakage as user error, (b) emit a sidecar / CI guard that warns "you need a custom binary", or (c) push gohumanize-linter to ship a Go-plugin fallback. (a) is what we'd ship today; (b) and (c) require new work.
