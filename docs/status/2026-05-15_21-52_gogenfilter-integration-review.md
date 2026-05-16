# Status Report: 2026-05-15 21:52 — gogenfilter Integration Review

## Session Goal

Integrate `github.com/LarsArtmann/gogenfilter/v3` to automatically detect and exclude auto-generated Go files (`*_templ.go`, sqlc, protobuf, etc.) from golangci-lint.

---

## A) FULLY DONE

1. **gogenfilter/v3 dependency added** — `go.mod` includes `github.com/LarsArtmann/gogenfilter/v3 v3.0.1`
2. **Scanner package created** — `pkg/gogenfilter/scanner.go` (324 lines):
   - `ScanProject()` walks project, detects generators, derives glob patterns
   - `MergeExclusionPaths()` deduplicates and sorts
   - `ExclusionPaths()` extracts path patterns from exclusions
   - Skips vendor/, hidden dirs, node_modules
   - sqlc config discovery via `gogenfilter.GetSQLOutputDirs()`
3. **BDD tests written** — `pkg/gogenfilter/scanner_test.go` (204 lines, 12 specs):
   - No generated files, templ, protobuf, multiple generators, vendor/hidden dirs, generic
   - `MergeExclusionPaths` dedup, `ExclusionPaths` extraction, `GeneratedExclusion.String()`
4. **Fixer integration wired** — `pkg/linter/fixer_config.go`:
   - `updateGeneratedExclusions()` method on `configUpdater`
   - Sets `linters.exclusions.generated: lax` and `formatters.exclusions.generated: lax`
   - Scans project, merges paths into both linter and formatter exclusions
5. **fixCounts updated** — Added `generated` field to `fixCounts` struct in `fixer.go`
6. **depguard allow-list updated** — `.golangci.yml` includes `github.com/LarsArtmann/gogenfilter/v3`
7. **AGENTS.md updated** — Documented gogenfilter integration, scanner, supported generators table
8. **All 14 test suites pass** — 254 specs, composite coverage 59.7%
9. **Build succeeds** — `go build ./...` clean

---

## B) PARTIALLY DONE

1. **Scanner pattern generation for sqlc** — Works but uses `|` pipe delimiter to join multiple directory patterns into a single string. golangci-lint `exclusions.paths` expects **individual list entries**, not pipe-delimited strings. This would produce an invalid/ignored exclusion pattern.
2. **Scanner pattern generation for oapi-codegen** — Same `|` pipe delimiter issue via `dirBasedPattern()`.
3. **Nix build** — `flake.nix` vendorHash is stale. Build fails with hash mismatch. Got: `sha256-FwIIwX5zZz1BCil2p9GTu+Xt/PyfYawbSiCSUXm60cI=`, need to update `vendorHash`.

---

## C) NOT STARTED

1. **Fix `successResult` message** — Doesn't include `counts.generated` in output at `pkg/linter/fixer_results.go:65`
2. **Fix early-return bypass** — `applyLintersFix()` at `fixer.go:179` returns `noFixesResult()` when `counts.total() == 0`, completely bypassing `updateGeneratedExclusions`. The generated exclusions only run when OTHER fixes exist.
3. **Test coverage for fixer_config.go generated exclusions** — No tests for `updateGeneratedExclusions` or `mergeExclusionPaths` in the linter package
4. **Integration/E2E test** — No test that runs `configure` end-to-end and verifies exclusion paths appear in the output config
5. **Scanner refactoring: `[]GeneratedExclusion` return** — The `deriveExclusionPatterns` should return `[]string` paths directly instead of wrapping in `GeneratedExclusion` then immediately unwrapping with `ExclusionPaths()`. Extra indirection with no benefit.
6. **Missing `exclusions.paths` in this project's own `.golangci.yml`** — The project has `report_templ.go` but doesn't exclude it via paths (relies only on `generated: lax`). Should dogfood the feature.
7. **Scanner should use `fs.FS` interface** — Currently hardcoded to `os.DirFS`. Should accept `fs.FS` for testability (like the config loader does).

---

## D) TOTALLY FUCKED UP

1. **Pipe-delimited path patterns** — `sqlcPatterns()` and `dirBasedPattern()` join multiple directory patterns with `|` (e.g., `"db/**|models/**"`). golangci-lint treats this as a single literal glob pattern — it will NEVER match any files. This is silently broken. The fix is to return **multiple individual paths** in the exclusion list, not concatenate them.
2. **Early-return prevents generated exclusions from ever running** — If a project already has all linters configured (counts.total() == 0), the generated exclusion scan is completely skipped. This is the most common case for existing projects. The feature literally can't work for its primary use case.

---

## E) WHAT WE SHOULD IMPROVE

1. **Return `[]string` paths instead of pipe-joined strings** — `deriveExclusionPatterns` should return multiple `GeneratedExclusion` entries when multiple directories are involved, not join them with `|`
2. **Move generated exclusion scan outside the `counts.total() == 0` guard** — Always scan, even when no other fixes are needed
3. **Use the existing `types.Set[string]` for merge logic** instead of custom `MergeExclusionPaths` — less code, same result
4. **Add `fs.FS` parameter to `ScanProject`** — match the project's interface-based design pattern
5. **Simplify `GeneratedExclusion`** — Consider removing it entirely and just returning `[]string` paths with a separate `map[string]string` for reasons. The wrapper type adds complexity with no clear benefit.
6. **Dogfood: add `exclusions.paths` to `.golangci.yml` for `*_templ.go`** — Validate our own feature works
7. **Update `successResult` to include generated count** — Simple fix, important for user feedback
8. **The `newFixCounts()` function is redundant** — Go zero-initializes all fields; `fixCounts{}` is equivalent to `newFixCounts()`

---

## F) TOP 25 THINGS TO DO NEXT

Sorted by **impact × effort** (highest impact, lowest effort first):

| #   | Task                                                                       | Impact   | Effort | Category       |
| --- | -------------------------------------------------------------------------- | -------- | ------ | -------------- |
| 1   | Fix pipe-delimited pattern join → return multiple paths                    | CRITICAL | S      | Bug fix        |
| 2   | Fix early-return bypass: always scan generated exclusions                  | CRITICAL | S      | Bug fix        |
| 3   | Update `successResult` to include `counts.generated`                       | HIGH     | XS     | Bug fix        |
| 4   | Update `flake.nix` vendorHash                                              | HIGH     | XS     | Build fix      |
| 5   | Simplify `newFixCounts()` → use `fixCounts{}`                              | LOW      | XS     | Cleanup        |
| 6   | Add `updateGeneratedExclusions` unit tests in linter package               | HIGH     | M      | Testing        |
| 7   | Add sqlc multi-dir pattern test                                            | HIGH     | S      | Testing        |
| 8   | Add oapi-codegen pattern test                                              | MEDIUM   | S      | Testing        |
| 9   | Dogfood: add `**/*_templ.go` to own `.golangci.yml` exclusions.paths       | MEDIUM   | XS     | Dogfood        |
| 10  | Refactor `ScanProject` to accept `fs.FS`                                   | MEDIUM   | S      | Architecture   |
| 11  | Simplify `MergeExclusionPaths` using `types.Set[string]`                   | LOW      | XS     | Cleanup        |
| 12  | Integration test: configure → verify exclusion paths in output             | HIGH     | M      | Testing        |
| 13  | Return `[]string` paths directly, eliminate `GeneratedExclusion` wrapper   | MEDIUM   | S      | Simplification |
| 14  | Move scanner to `pkg/gogenfilter` → consider merging into `pkg/detection`  | LOW      | M      | Architecture   |
| 15  | Add `--verbose` logging for scanner results                                | MEDIUM   | XS     | UX             |
| 16  | Handle error from `filter.FilterDetailed()` instead of silently continuing | MEDIUM   | XS     | Robustness     |
| 17  | Test: project with existing exclusion paths + new generated files          | MEDIUM   | S      | Testing        |
| 18  | Add `exclusions.generated: strict` option for stricter mode                | LOW      | M      | Feature        |
| 19  | Cache scan results across multiple configure runs                          | LOW      | M      | Performance    |
| 20  | Add `--skip-generated-scan` flag                                           | LOW      | S      | Feature        |
| 21  | Update `applyPreset` to also scan for generated files                      | HIGH     | S      | Feature gap    |
| 22  | Test generated exclusion with formatters (gci, goimports)                  | MEDIUM   | M      | Testing        |
| 23  | Add benchmark test for `ScanProject` on large codebases                    | LOW      | S      | Testing        |
| 24  | Consider using `gogenfilter.WithIncludePatterns` to scope scanning         | LOW      | S      | Optimization   |
| 25  | Document generated exclusion in README.md                                  | MEDIUM   | S      | Docs           |

---

## G) TOP #1 QUESTION

**Does `applyPreset()` in `cmd_configure.go` also need to scan for generated files?** Currently only the fixer flow (`runFixerMode`) triggers `updateGeneratedExclusions`. The preset flow (`handlePresetMode` → `applyPreset`) saves a fresh config with `cfg.Linters.Enable = linterNames` and `cfg.Linters.Disable = []string{}` — completely bypassing the generated exclusion scan. Users who use `--preset standard` will never get generated file exclusions.

---

## Build & Test Status

- **go build**: PASS
- **ginkgo -r**: 14/14 suites, 254/254 specs PASS, 59.7% coverage
- **nix build**: FAIL (stale vendorHash)
- **golangci-lint config verify**: PASS
- **self-validate**: PASS

## Commit Status

- Working tree: CLEAN (all changes committed)
- Latest commit: `435f961 docs: add Section 15 to audit report for automated health checks`
- gogenfilter commit: `06f0491 feat(fixer): add auto-generated file exclusion scanning via gogenfilter`
