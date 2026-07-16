# Status: json/v2 Migration Fix + Build Repair

> **🔄 RETROACTIVE UPDATE — 2026-07-16**
>
> Most items from this report's "NOT STARTED" and "50 things" sections have since been completed:
>
> | Item                                                  | Status        | Details                                                                    |
> | ----------------------------------------------------- | ------------- | -------------------------------------------------------------------------- |
> | Commit all changes                                    | ✅ Done       | Commits f38ff5e, cfcb0df, 06d81a0                                          |
> | GOEXPERIMENT in CI workflows                          | ✅ Done       | Added to ci.yml (test-and-build, lint, govulncheck) + release.yml          |
> | `checks.race` GOEXPERIMENT inheritance                | ✅ Verified   | `nix flake check` passes — `old.env // {}` merge inherits GOEXPERIMENT     |
> | `.buildflow.yml` skip nixfmt-standalone               | ✅ Done       | Committed; BuildFlow 44/44 green                                           |
> | CONTRIBUTING.md with GOEXPERIMENT                     | ✅ Done       | Full rewrite with prerequisites                                            |
> | `docs/references/json-v2.md`                          | ✅ Done       | New reference doc created                                                  |
> | `docs/references/integrations.md` wire-format section | ✅ Done       | Wire→Report conversion table added                                         |
> | FEATURES.md json/v2 migration row                     | ✅ Done       | Added to Build & CI table                                                  |
> | TODO_LIST.md json/v2 item                             | ✅ Done       | Added to Completed section                                                 |
> | GOEXPERIMENT in shellHook echo                        | ✅ Done       | Visible in devShell banner                                                 |
> | gosec G204 nolints                                    | ❌ Still open | `loader.go` and `cmd_validate.go` still have 2 pre-existing gosec warnings |
>
> See `docs/status/2026-07-09_07-09_json-v2-complete-buildflow-green.md` for the follow-up session that resolved most items. Current status: `TODO_LIST.md`.

**Date:** 2026-07-09 06:29
**Session scope:** Fix BuildFlow failures (test-race, hierarchical-errors, nixfmt-standalone)
**Branch:** master (uncommitted working tree changes)
**Commit base:** a8ff465 (feat: migrate all fmt.Errorf to go-error-family structured errors)

---

## What Happened This Session

User ran `buildflow --fix --semantic --build-mode=full` — 3 steps failed:

- `test-race` (exit 1)
- `hierarchical-errors` (exit 75)
- `nixfmt-standalone` (exit 1)

Root cause traced to commit `a8ff465` which migrated `encoding/json` v1 → v2 **without**:

1. Enabling `GOEXPERIMENT=jsonv2` in any build environment
2. Fixing JSON tags for json/v2 case-sensitivity

---

## a) FULLY DONE ✅

| Item                                                               | Details                                                                                                                                                                                                                                                                                |
| ------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Root cause identified**                                          | Two bugs: missing `GOEXPERIMENT=jsonv2` env var + case-sensitivity mismatch in JSON tags                                                                                                                                                                                               |
| **GOEXPERIMENT=jsonv2 enabled in flake.nix**                       | Added to package build `env`, devShell `env`, and CI shell — 3 locations                                                                                                                                                                                                               |
| **vendorHash updated**                                             | `sha256-RgWyYJ4BF7n2lkzFYyXNcJQJkW1RRqEcBkxayFT1RoM=` (from go.mod dependency bumps)                                                                                                                                                                                                   |
| **Wire-format decoupling in analyzer.go**                          | Added `golangciLinterEntry` and `golangciFormatterEntry` structs with correct lowercase JSON tags matching golangci-lint's `linterHelp`/`formatterHelp` wire format, plus `toLinterInfo()`/`toFormatterInfo()` conversion methods and `convertLinters()`/`convertFormatters()` helpers |
| **LinterList tags fixed in loader.go**                             | `"enabled"`/`"disabled"` → `"Enabled"`/`"Disabled"` (golangci-lint uses capitalized wrapper keys)                                                                                                                                                                                      |
| **LinterInfo/FormatterInfo reverted to tag-free**                  | Report types must stay PascalCase per project tag policy (tagliatelle `json: pascal`) — verified by existing `json_tags_test.go`                                                                                                                                                       |
| **golangci-lint output verified against source**                   | Cross-referenced actual structs from `github.com/golangci/golangci-lint` main branch: `linterHelp`, `formatterHelp`, `BuildInfo`, `JSONResult`, `result.Issue` — all tags match wire format exactly                                                                                    |
| **All tests pass**                                                 | `GOEXPERIMENT=jsonv2 CGO_ENABLED=1 go test -race ./pkg/... ./internal/... ./cmd/...` — 16 packages, all green                                                                                                                                                                          |
| **nix build passes**                                               | `nix build` succeeds with new vendorHash + GOEXPERIMENT                                                                                                                                                                                                                                |
| **buildflow test-race + test-coverage + hierarchical-errors pass** | Confirmed in full buildflow run with `GOEXPERIMENT=jsonv2` in env                                                                                                                                                                                                                      |
| **AGENTS.md updated**                                              | Added GOEXPERIMENT requirement + wire-format decoupling documentation in Tech Stack section                                                                                                                                                                                            |
| **golangci-lint passes on changed code**                           | Only 2 pre-existing gosec G204 warnings remain (committed code, not from this session)                                                                                                                                                                                                 |

---

## b) PARTIALLY DONE ⚠️

| Item                               | Status                                                                                                                             | What's Left                                                                                                                                                 |
| ---------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **buildflow full run**             | 38/45 steps pass with `GOEXPERIMENT=jsonv2` exported                                                                               | `nixfmt-standalone` still fails (pre-existing, 86% failure rate — scans `.direnv/flake-inputs/` cache of third-party .nix files)                            |
| **loader.go LinterList fix**       | Fixed the wrapper key casing (`"Enabled"`/`"Disabled"`)                                                                            | The `getAllLinterNames` function using this struct is not covered by a dedicated test — only indirectly via fixer tests                                     |
| **golangci-lint config exclusion** | Removed `pkg/types/types.go` from tagliatelle exclusion (was added then reverted when we went with wire-format decoupling instead) | The `.golangci.yml` diff removes a tagliatelle exclusion for `types.go` that existed in HEAD — needs verification this doesn't break lint on committed code |

---

## c) NOT STARTED ❌

| Item                                            | Why                                                                                                                                                                       |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Commit the changes**                          | User did not request a commit                                                                                                                                             |
| **CI pipeline verification**                    | No CI config found/checked — the `checks` in flake.nix have `race` check but it doesn't set `GOEXPERIMENT` in the `checks.race` overrideAttrs (only sets `CGO_ENABLED=1`) |
| **`.golangci.yml` GOEXPERIMENT for gosec G204** | 2 pre-existing gosec warnings on `exec.CommandContext` calls — not this session's scope but worth nolinting                                                               |
| **devShell direnv reload verification**         | The flake.nix env change requires direnv cache invalidation — tested manually but not verified the direnv auto-reload picks it up                                         |

---

## d) TOTALLY FUCKED UP 💥

| Item                                                         | What Went Wrong                                                                                                                                                                                    | Impact                                                                                                                                   |
| ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **First approach: tried reverting json/v2 migration**        | Spent significant effort reverting 7 files to v1, running tests, then discovering buildflow's `go-auto-upgrade:repair` re-applied the migration automatically. Tried reverting again, same result. | Wasted ~30 min. Should have first understood _why_ buildflow re-applied it (Lars's own `go-auto-upgrade` tool) before deciding strategy. |
| **Second approach: added lowercase json tags to LinterInfo** | Broke `json_tags_test.go` which asserts PascalCase output for Report types. Tried adding tagliatelle exclusion as workaround.                                                                      | Wasted ~15 min. Should have checked for existing tests enforcing the tag policy before changing tags.                                    |
| **Investigated `go fix` as cause**                           | Ran `GOEXPERIMENT=jsonv2 go fix ./...` thinking it re-applied json/v2 migration. It didn't. The real cause was buildflow's `go-auto-upgrade` step.                                                 | Minor confusion, ~5 min.                                                                                                                 |
| **Almost missed HEAD already had json/v2**                   | The committed code at a8ff465 already used json/v2 imports — I assumed the migration was uncommitted WIP. The _real_ issue was always just missing GOEXPERIMENT + tag casing.                      | This caused the initial revert strategy. Should have `git show HEAD:file` first.                                                         |

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Process Improvements

1. **Always `git show HEAD:<file>` before assuming changes are uncommitted.** The working tree diff in the env snapshot showed modified files, but those were from buildflow auto-fixes, not uncommitted user work. The actual committed code already had json/v2.

2. **Check for existing tests enforcing constraints before changing types.** `json_tags_test.go` enforces PascalCase for Report types. This was knowable before editing.

3. **Understand tooling before fighting it.** buildflow's `go-auto-upgrade` is Lars's own tool designed to migrate to json/v2. Reverting its work is futile — the correct response is completing the migration properly.

4. **The `checks.race` in flake.nix likely needs GOEXPERIMENT too.** It overrides attrs but only sets `CGO_ENABLED=1`, not `GOEXPERIMENT=jsonv2`. `nix flake check` would fail on the race check.

### Code Improvements

5. **`pkg/config/loader.go` LinterList struct uses anonymous struct embedding.** Could be a named type for clarity and reuse, but this is cosmetic.

6. **The `convertLinters`/`convertFormatters` helpers in analyzer.go could be generics.** They're identical except for types. A generic `convertSlice[T, U any](in []T, fn func(T) U) []U` would eliminate duplication.

7. **No dedicated test for `LinterList` JSON parsing.** The `getAllLinterNames` function in loader.go parses golangci-lint output with the case-sensitive `"Enabled"` tag fix, but only has indirect coverage via fixer tests. A unit test with mock JSON would be safer.

---

## f) Up to 50 Things We Should Get Done Next

### Critical (blocks correctness)

1. **Add `GOEXPERIMENT=jsonv2` to `checks.race` in flake.nix** — currently only sets `CGO_ENABLED=1`, would fail `nix flake check`
2. **Verify `nix flake check` passes** — likely fails on race check due to missing GOEXPERIMENT
3. **Commit all changes** — working tree has 7 files changed, uncommitted
4. **Verify direnv auto-reload picks up GOEXPERIMENT from flake.nix** — users entering `nix develop` should get it automatically

### High Priority (prevents future breakage)

5. **Add unit test for `LinterList` JSON parsing in loader.go** — verify case-sensitive `"Enabled"` tag works with real golangci-lint output
6. **Add unit test for `golangciLinterEntry`/`golangciFormatterEntry` parsing** — verify all fields parse correctly under json/v2
7. **Fix or exclude `nixfmt-standalone` in buildflow** — 86% failure rate scanning `.direnv/flake-inputs/` is unacceptable noise; configure buildflow to skip `.direnv/`
8. **Add `//nolint:gosec // trusted binary name from constants` to the 2 G204 warnings** — `loader.go:247` and `cmd_validate.go:260` use `exec.CommandContext` with constant binary names, not user input

### Medium Priority (code quality)

9. **Refactor `convertLinters`/`convertFormatters` to use generics** — eliminate the duplicated loop pattern
10. **Consider extracting wire-format types to `pkg/linter/wire_format.go`** — keeps `analyzer.go` focused on analysis logic
11. **Update `docs/references/integrations.md` with json/v2 wire-format documentation** — future maintainers need to know why there are two layers of types
12. **Update `docs/status/2026-07-07_23-03_quality-sprint-constants-errors-cleanup.md`** — the json/v2 question "Is the v2 API stable enough?" is now answered: yes, adopted with GOEXPERIMENT
13. **Review `pkg/constants/experiments.go`** — the `goexperiment.jsonv2` entry is now actively used, not just listed
14. **Add `GOEXPERIMENT=jsonv2` to any GitHub Actions / CI workflows** — search for `.github/workflows/` CI configs
15. **Consider adding a `//go:build goexperiment.jsonv2` constraint file** — or document that the project requires the experiment at the environment level

### Low Priority (polish)

16. **Remove the dead `isParallelRunningError` function** — buildflow's `stdlibwrappers` migration already inlined it in `command_runner.go` (committed in HEAD, not this session's change, but worth confirming)
17. **Consider adding `GOEXPERIMENT=jsonv2` to `go.test` settings in `.vscode/settings.json` or equivalent** — if the project uses VS Code
18. **Document the json/v2 behavioral changes** — nil slices → `[]{}` not `null`, `[]byte` → base64, case-sensitive matching; relevant for any future JSON config round-tripping
19. **Consider whether `ConfigFormatJSON` path in `loader.go:marshalConfig` needs testing** — it now uses `json/v2.Marshal` with `jsontext.WithIndent` options
20. **Review if `go-auto-upgrade` should skip `stdlibwrappers` migrator** — it inlined `isParallelRunningError`, reducing readability slightly; may want to keep helper functions for complex logic

### Documentation

21. **Update FEATURES.md** — json/v2 migration is now complete and working
22. **Update TODO_LIST.md** — remove the json/v2 "blocked on ecosystem readiness" item if present
23. **Add a CONTRIBUTING.md note about GOEXPERIMENT** — new contributors need to know
24. **Update `docs/references/error-handling.md`** — if json/v2 changes error handling for JSON parse errors
25. **Consider adding a `.envrc` note** — `direnv reload` needed after flake.nix env changes

### Testing

26. **Add integration test that runs the full binary end-to-end** — `go test` with `GOEXPERIMENT=jsonv2` executing the CLI against a real `.golangci.yml`
27. **Add test for JSON config file round-trip** — `marshalConfig(ConfigFormatJSON)` → `unmarshalConfig(ConfigFormatJSON)` with json/v2
28. **Benchmark json/v2 vs v1** — if performance matters for large configs
29. **Test `ParseGolangciLintJSON` with real golangci-lint `run --out-format json` output** — verify `GolangciLintIssue` tags still work under v2
30. **Add fuzz test for JSON parsing of golangci-lint output** — edge cases in wire format

### Architecture

31. **Consider whether Report types and Wire types should be in separate packages** — `pkg/types` for Report, `pkg/linter/wire` for external API shapes
32. **Evaluate if all `encoding/json` v1 imports should be migrated** — search for any remaining `encoding/json` (non-v2) imports
33. **Review the `go-finding` integration** — does it use json/v1 or v2? Consistency check
34. **Consider a `jsonformat` package** — centralize json/v2 Marshal options (indent prefix/width) used in 4 call sites
35. **Review `examples/api-usage`** — ensure it compiles and works with json/v2

### Nix / Build

36. **Consider adding a `checks.jsonv2` flake check** — explicitly verify the build works with GOEXPERIMENT=jsonv2
37. **Add `GOEXPERIMENT` to `shellHook` echo** — show it in the dev shell banner alongside Go/golangci-lint/templ versions
38. **Consider pinning Go version that supports jsonv2** — ensure Go 1.26+ is always used
39. **Review if `allowGoReference = true` interacts with GOEXPERIMENT** — the embedded GOROOT might need the experiment compiled in
40. **Document vendorHash update procedure in AGENTS.md gotcha #3** — add the `GOEXPERIMENT=jsonv2` note to the procedure

### Cleanup

41. **Remove `/tmp/jsontest*.go` temp files** — created during debugging
42. **Review `pkg/report/report_templ.go`** — buildflow's `templ-generate` step modified it; verify it's correct
43. **Check if `go.sum` changes are complete** — `go mod tidy` ran, but verify no missing entries
44. **Review if `flake.lock` changes are expected** — nix-flake-update auto-updated inputs
45. **Consider adding `GOEXPERIMENT=jsonv2` to `go.toolchain` settings** — if Go modules support experiment flags

### Future-Proofing

46. **Monitor Go 1.27 release** — json/v2 may become stable (no longer experimental), at which point GOEXPERIMENT can be removed
47. **Consider migrating YAML parsing to a v2-compatible path** — `go.yaml.in/yaml/v3` uses reflection; verify it works with json/v2 struct tags (it should, as YAML uses `yaml` tags not `json`)
48. **Consider adding a `go generate` directive** — to auto-detect if GOEXPERIMENT is missing and warn
49. **Review `pkg/client/` package** — it was failing in the initial test run; verify it compiles with json/v2
50. **Consider a pre-commit hook that checks GOEXPERIMENT** — prevent commits that would break the build without it

---

## g) Top 2 Questions I Cannot Answer Myself

### 1. Should `checks.race` in flake.nix also set `GOEXPERIMENT=jsonv2`?

The `checks.race` overrideAttrs at `flake.nix:190` only sets `CGO_ENABLED=1` but not `GOEXPERIMENT`. Since the parent derivation now sets `GOEXPERIMENT=jsonv2` in `env`, does `old.env` include it in the override? If `overrideAttrs` merges rather than replaces, it should inherit. But I cannot verify this without running `nix flake check` (which takes ~5min). **Should I run `nix flake check` to verify, or do you know off-hand if overrideAttrs inherits parent env?**

### 2. Is the `nixfmt-standalone` failure in buildflow a known issue you're tracking?

It has an 86% failure rate (buildflow's own warning) because it scans `.direnv/flake-inputs/` cache of third-party .nix files. The `nix-fmt` step (via treefmt) passes fine. **Should I configure buildflow to exclude `.direnv/` from `nixfmt-standalone`, or is this already tracked elsewhere?**
