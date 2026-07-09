# Status: json/v2 Migration Completion + BuildFlow 44/44 Green

**Date:** 2026-07-09 07:09 (updated 07:09 — follow-up CI + documentation session)
**Session scope:** Fix all BuildFlow failures, enable json/v2 end-to-end, verify all build systems
**Commits this session:** f38ff5e, cfcb0df, 06d81a0 (on top of a8ff465)
**Follow-up work:** CI workflows + all documentation (#12-17, #27-33) — uncommitted
**Untracked:** `.buildflow.yml` (new)
**BuildFlow:** 44/44 ✅ | **nix flake check:** all passed ✅ | **Tests:** 16/16 packages ✅

---

## Executive Summary

Commit a8ff465 migrated `encoding/json` v1→v2 but left the project broken: no `GOEXPERIMENT=jsonv2` anywhere and JSON tags that relied on v1 case-insensitive matching. This session completed the migration properly: enabled the experiment in flake.nix, fixed all wire-format JSON tags against golangci-lint's actual source, added a `.buildflow.yml` to fix nixfmt-standalone, and verified every build system passes.

---

## a) FULLY DONE ✅

| Item                                                | Commit / File                | Details                                                                                                                                                                                                                                                                                                     |
| --------------------------------------------------- | ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **GOEXPERIMENT=jsonv2 in flake.nix**                | 06d81a0                      | Added to package build `env`, devShell `env`, CI shell env — 3 locations                                                                                                                                                                                                                                    |
| **vendorHash updated**                              | 06d81a0                      | `sha256-RgWyYJ4BF7n2lkzFYyXNcJQJkW1RRqEcBkxayFT1RoM=` for new go.mod deps                                                                                                                                                                                                                                   |
| **Wire-format decoupling in analyzer.go**           | 06d81a0                      | Added `golangciLinterEntry` + `golangciFormatterEntry` structs matching golangci-lint's `linterHelp`/`formatterHelp` wire format (lowercase field tags). Added `toLinterInfo()`/`toFormatterInfo()` converters + `convertLinters()`/`convertFormatters()` helpers. Report types stay tag-free (PascalCase). |
| **LinterList wrapper tags fixed**                   | 06d81a0                      | `pkg/config/loader.go`: `"enabled"`/`"disabled"` → `"Enabled"`/`"Disabled"` (golangci-lint uses capitalized wrapper keys)                                                                                                                                                                                   |
| **JSON tags verified against golangci-lint source** | —                            | Cross-referenced actual structs: `linterHelp` (pkg/commands/help_linters.go), `formatterHelp` (pkg/commands/help_formatters.go), `BuildInfo` (pkg/commands/version.go), `JSONResult` + `result.Issue` (pkg/printers/json.go, pkg/result/issue.go). All tags match.                                          |
| **All tests pass**                                  | 06d81a0                      | `GOEXPERIMENT=jsonv2 CGO_ENABLED=1 go test -race ./pkg/... ./internal/... ./cmd/...` — 16 packages, 0 failures                                                                                                                                                                                              |
| **nix build passes**                                | 06d81a0                      | `nix build` succeeds with GOEXPERIMENT + new vendorHash                                                                                                                                                                                                                                                     |
| **nix flake check passes**                          | verified                     | All 3 checks (format, build, race) pass — `checks.race` inherits GOEXPERIMENT via `old.env // {}` merge                                                                                                                                                                                                     |
| **BuildFlow 44/44 green**                           | verified                     | `GOEXPERIMENT=jsonv2 buildflow --fix --semantic --build-mode=full` — all steps pass                                                                                                                                                                                                                         |
| **AGENTS.md updated**                               | 06d81a0                      | Added GOEXPERIMENT requirement + wire-format decoupling docs in Tech Stack                                                                                                                                                                                                                                  |
| **`.buildflow.yml` created**                        | untracked                    | Added `skip_steps: [nixfmt-standalone]` (86-88% failure rate scanning `.direnv/` cache), full default exclude list + `.direnv`                                                                                                                                                                              |
| **Previous status report**                          | docs/status/2026-07-09_06-29 | Written mid-session                                                                                                                                                                                                                                                                                         |

### Follow-up session (CI + Documentation)

| Item                               | File(s) changed                           | Details                                                                                                                                                                                       |
| ---------------------------------- | ----------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **GOEXPERIMENT in GitHub Actions** | `.github/workflows/ci.yml`, `release.yml` | Added `env: GOEXPERIMENT: jsonv2` to `test-and-build`, `lint`, `govulncheck`, and `release` jobs                                                                                              |
| **GOEXPERIMENT in shellHook echo** | `flake.nix`                               | Added `echo "  GOEXPERIMENT:   $GOEXPERIMENT"` to devShell shellHook for developer visibility                                                                                                 |
| **AGENTS.md gotchas updated**      | `AGENTS.md`                               | New gotcha #13 (`.buildflow.yml` skip rationale + CI GOEXPERIMENT), updated #3 (vendorHash + GOEXPERIMENT), fixed #4 (removed stale `allowGoReference = true`), added json-v2.md to ref table |
| **TODO_LIST.md updated**           | `TODO_LIST.md`                            | Added json/v2 migration to Completed section                                                                                                                                                  |
| **FEATURES.md updated**            | `FEATURES.md`                             | Added `encoding/json/v2` migration row to Build & CI table                                                                                                                                    |
| **Quality-sprint status updated**  | `docs/status/2026-07-07_23-03_*.md`       | Marked json/v2 question #1 as answered; updated items #11, #32, NOT STARTED section                                                                                                           |
| **CONTRIBUTING.md rewritten**      | `CONTRIBUTING.md`                         | Added GOEXPERIMENT prerequisites, Nix dev shell + manual setup, common commands                                                                                                               |
| **json/v2 reference doc created**  | `docs/references/json-v2.md`              | New file: GOEXPERIMENT config table, v1→v2 behavioral changes (case-sensitivity, nil slices, []byte, API), wire-format decoupling pattern                                                     |
| **integrations.md updated**        | `docs/references/integrations.md`         | Added JSON wire-format decoupling section with wire→Report conversion table                                                                                                                   |

---

## b) PARTIALLY DONE ⚠️

| Item                   | Status                                                             | What's Left                                                                                                                                      |
| ---------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **`.buildflow.yml`**   | Created and working (44/44) but **untracked** — not committed      | Needs `git add .buildflow.yml` + commit                                                                                                          |
| **golangci-lint lint** | Only 2 pre-existing gosec G204 warnings remain                     | These are in committed code (loader.go:247, cmd_validate.go:260) — predate this session. Could add `//nolint:gosec` but they're not our changes. |
| **Temp debug files**   | Two `/tmp/jsontest*.go` files created during debugging still exist | Need cleanup (cosmetic — they're in /tmp)                                                                                                        |

---

## c) NOT STARTED ❌

| Item                                             | Why                                                                                                           |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------- |
| **Commit `.buildflow.yml`**                      | User hasn't said "commit"                                                                                     |
| **Remove `/tmp/jsontest*.go`**                   | Cosmetic cleanup, not requested                                                                               |
| **Dedicated unit tests for wire-format structs** | `golangciLinterEntry`/`golangciFormatterEntry` parsing is only covered indirectly via fixer integration tests |

---

## d) TOTALLY FUCKED UP 💥

| Item                                                    | What Went Wrong                                                                                                                                                                                                                                                                  | Impact                                                                                                | Fixed?                              |
| ------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ----------------------------------- |
| **Tried reverting the json/v2 migration**               | Spent ~30 min reverting 7 files to v1, running tests, discovering buildflow's `go-auto-upgrade:repair` re-applied it automatically. Tried reverting again, same result.                                                                                                          | Wasted time. Should have first checked `git show HEAD:file` — the committed code already had json/v2. | ✅ Yes                              |
| **Added lowercase json tags directly to LinterInfo**    | Broke `json_tags_test.go` which asserts PascalCase output for Report types. Added tagliatelle exclusion as workaround, then realized the whole approach was wrong.                                                                                                               | Wasted ~15 min. Should have checked existing tests before editing types.                              | ✅ Yes                              |
| **`.buildflow.yml` initially dropped default excludes** | When creating `.buildflow.yml`, I wrote a minimal exclude list missing `reports/`, `__snapshots__`, `_templ.go`, lock files, etc. This caused `oxfmt:repair` to fail with permission errors scanning `reports/html/` (read-only generated CSS/JS).                               | Wasted ~10 min on a self-inflicted issue. Should have started from `buildflow config init` defaults.  | ✅ Yes                              |
| **Assumed the migration was uncommitted WIP**           | The git status snapshot showed modified .go files. I assumed someone started a json/v2 migration but didn't finish it. In reality, commit a8ff465 (later split into f38ff5e) already committed the migration — the modified files were from buildflow auto-fixes re-applying it. | Caused the initial wrong strategy (revert instead of complete).                                       | ✅ Yes                              |
| **Didn't check if commits were made**                   | Between the first status report and now, three commits were made (f38ff5e, cfcb0df, 06d81a0) — I didn't track when/whether these happened vs remaining uncommitted. The `.buildflow.yml` was left untracked.                                                                     | Minor confusion at status time.                                                                       | ⚠️ `.buildflow.yml` still untracked |

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Critical Process Failures

1. **`git show HEAD:<file>` BEFORE assuming changes are uncommitted.** I wasted 30+ minutes trying to revert a migration that was already committed. The working-tree modifications were from buildflow auto-fixes, not human-authored WIP. This is the single biggest process failure of the session.

2. **Check for existing enforcing tests before changing types.** `json_tags_test.go` explicitly enforces PascalCase for Report types. Reading that file first would have prevented the "add lowercase tags → break tests → revert → add wire-format structs" detour.

3. **Start from tool defaults, not from scratch.** When creating `.buildflow.yml`, I should have run `buildflow config init` to get the full default exclude list instead of writing a minimal one that dropped `reports/` and caused oxfmt failures.

4. **Track commit state throughout the session.** Three commits were made but I lost track of what was committed vs uncommitted. The `.buildflow.yml` was left untracked because I didn't verify at the end.

### Architectural Improvements

5. **The wire-format structs in analyzer.go should have dedicated tests.** Currently `golangciLinterEntry` and `golangciFormatterEntry` are only tested indirectly via integration tests that invoke real golangci-lint. A unit test with mock JSON would catch regressions faster and document the wire format contract.

6. **Consider extracting wire-format types to a dedicated file.** `pkg/linter/wire_format.go` would keep `analyzer.go` focused on analysis logic and make the wire-format contract more discoverable.

7. **The `convertLinters`/`convertFormatters` helpers are identical patterns.** They could use Go generics: `maps.Fn[[]T, []U]` or a generic `convertSlice` utility.

---

## f) Up to 50 Things We Should Get Done Next

### Immediate (this session's loose ends)

1. **Commit `.buildflow.yml`** — it's untracked and required for 44/44 buildflow
2. **Remove `/tmp/jsontest*.go` and `/tmp/jsontest2.go`** — debug temp files
3. **Verify the previous status report at `docs/status/2026-07-09_06-29` is still accurate** — or mark it superseded by this one

### Testing

4. **Add unit test for `golangciLinterEntry` JSON parsing** — mock golangci-lint `{"Enabled": [...], "Disabled": [...]}` JSON, verify all fields parse
5. **Add unit test for `golangciFormatterEntry` JSON parsing** — same for formatters
6. **Add unit test for `LinterList` in loader.go** — verify case-sensitive `"Enabled"` tag with real golangci-lint output shape
7. **Add JSON config round-trip test** — `marshalConfig(ConfigFormatJSON)` → `unmarshalConfig(ConfigFormatJSON)` under json/v2
8. **Add fuzz test for `ParseGolangciLintJSON`** — verify `GolangciLintIssue` tags work under v2 with edge cases
9. **Test `getAllLinterNames` function directly** — currently only indirect coverage via fixer tests
10. **Add integration test running the full CLI binary** — end-to-end `golangci-lint-auto-configure configure` with real config
11. **Benchmark json/v2 vs v1 for config parsing** — if performance matters for large configs

### Build / CI

12. ~~**Check GitHub Actions workflows for GOEXPERIMENT**~~ ✅ DONE — checked all 3 workflow files
13. ~~**Add `GOEXPERIMENT=jsonv2` to any CI workflows**~~ ✅ DONE — added to `ci.yml` (test-and-build, lint, govulncheck) + `release.yml`
14. ~~**Add `.buildflow.yml` to AGENTS.md gotchas**~~ ✅ DONE — new gotcha #13
15. ~~**Consider adding `GOEXPERIMENT` to the devShell `shellHook` echo**~~ ✅ DONE — added to shellHook
16. ~~**Verify `allowGoReference = true` doesn't conflict with GOEXPERIMENT**~~ ✅ DONE — `allowGoReference` was already removed (2026-06-29); no conflict. Updated AGENTS.md gotcha #4 to remove stale reference.
17. ~~**Document the vendorHash update procedure including GOEXPERIMENT**~~ ✅ DONE — updated AGENTS.md gotcha #3

### Code Quality

18. **Add `//nolint:gosec // trusted binary name from constants` to loader.go:247** — pre-existing G204 warning
19. **Add `//nolint:gosec // trusted binary name from constants` to cmd_validate.go:260** — pre-existing G204 warning
20. **Refactor `convertLinters`/`convertFormatters` to use generics** — eliminate duplication
21. **Extract wire-format types to `pkg/linter/wire_format.go`** — separate concerns
22. **Consider a shared `jsonMarshalIndent` helper** — the `json.Marshal(x, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))` pattern appears in 4 call sites
23. **Review `pkg/client/` package** — it was failing in the initial test run; verify it works now
24. **Review `examples/api-usage`** — ensure it compiles and runs with json/v2
25. **Check for any remaining `encoding/json` v1 imports** — search for non-v2 json usage
26. **Review `pkg/report/report_templ.go`** — buildflow modified it; verify correctness

### Documentation

27. ~~**Update TODO_LIST.md**~~ ✅ DONE — json/v2 migration added to Completed section
28. ~~**Update FEATURES.md**~~ ✅ DONE — json/v2 migration row added to Build & CI table
29. ~~**Update `docs/status/2026-07-07_23-03_quality-sprint`**~~ ✅ DONE — json/v2 question marked answered
30. ~~**Add CONTRIBUTING.md note about GOEXPERIMENT**~~ ✅ DONE — full rewrite with GOEXPERIMENT prerequisites
31. ~~**Document json/v2 behavioral changes**~~ ✅ DONE — new `docs/references/json-v2.md`
32. ~~**Update `docs/references/integrations.md`**~~ ✅ DONE — added wire-format decoupling section
33. ~~**Document the `.buildflow.yml` rationale**~~ ✅ DONE — AGENTS.md gotcha #13 + `.buildflow.yml` comments

### Architecture / Future-Proofing

34. **Monitor Go 1.27 release** — json/v2 may become stable, removing need for GOEXPERIMENT
35. **Consider whether Report types and Wire types should be in separate packages** — `pkg/types` for Report, `pkg/linter/wire` for external
36. **Evaluate if `go-finding` integration needs json/v2 updates** — consistency check
37. **Consider adding a pre-commit check for GOEXPERIMENT** — prevent breaking commits
38. **Consider `//go:build goexperiment.jsonv2` constraint** — or document env-level requirement
39. **Review `go.yaml.in/yaml/v3` interaction with json/v2 struct tags** — YAML uses `yaml` tags, should be unaffected, but verify
40. **Consider centralizing json/v2 options** — `jsontext.WithIndent` in a shared config

### Nix

41. ~~**Add GOEXPERIMENT to shellHook echo**~~ ✅ DONE — visibility for developers
42. **Consider a `checks.jsonv2` explicit check** — verify build works with the experiment
43. **Review `treefmt` programs** — ensure gofumpt/goimports work with json/v2 code
44. **Pin Go version that supports jsonv2** — ensure 1.26+ always
45. **Review flake.lock update** — nix-flake-update auto-updated inputs; verify expected

### Cleanup

46. **Remove dead `isParallelRunningError` helper** — already inlined by go-auto-upgrade in command_runner.go (committed)
47. **Review go.sum completeness** — `go mod tidy` ran, verify no missing entries
48. **Clean up any stale report files** — `reports/html/` had permission issues
49. **Review `.golangci.yml` tagliatelle exclusion removal** — verify no lint regressions on committed code
50. **Consider adding `.direnv/flake-inputs/` to `.gitignore`** — prevent it from being scanned by any tool

---

## g) Top 2 Questions I Cannot Answer Myself

### 1. Should `.buildflow.yml` be committed, and is `skip_steps: [nixfmt-standalone]` the right fix or a workaround?

The step has an 88% historical failure rate (buildflow's own telemetry). It runs raw `nixfmt .` which ignores buildflow's exclude patterns and scans `.direnv/flake-inputs/` (symlinked nix flake source caches containing third-party .nix files). The `nix-fmt` step (via treefmt) already handles Nix formatting correctly and respects excludes. **Is this a known buildflow issue you're fixing upstream, or should the project permanently skip this step?**

### 2. Were the three commits (f38ff5e, cfcb0df, 06d81a0) made by you or another tool during this session?

I see these commits in `git log` but I didn't explicitly create them through a commit command in our conversation. The commit messages follow the project's conventional-commit style and contain exactly the changes I made. **Did you commit these manually, or was there a hook/tool that committed them? I need to know so I don't accidentally amend or squash work I didn't create.**
