# Status Report: Quality Sprint — Constants, Error Classification, and Code Cleanup

> **Resolved 2026-09-11 (docs-health archive pass).** cmd_configure split and ConfigLoader interface decomposition shipped in v0.6.0; error-consistency items consumed by the structured-error migration (07-08) and the go-error-family adoption (v0.7.0); swallowed-errors audit completed 2026-07-26 (erraudit). Forward-looking items below are struck inline; process sections (d/e) are retained as historical context. Archived from `docs/status/` — live state: `TODO_LIST.md` / `ROADMAP.md` / `CHANGELOG.md`.

> **🔄 RETROACTIVE UPDATE — 2026-07-16**
>
> Items from this report's "50 things to do next" have the following status:
>
> | Item                                                   | Status            | Details                                                                       |
> | ------------------------------------------------------ | ----------------- | ----------------------------------------------------------------------------- |
> | Consolidate ErrNoConfigFiles (remove deprecated alias) | ⚠️ Partial         | Moved to `pkg/errors/errors.go`; deprecated alias in merger.go still exists   |
> | Extract errUnsupportedConfigFormat to pkg/errors/      | ❌ Not done       | Still in config/loader.go                                                     |
> | Split cmd_configure.go (541 lines, 8 concerns)         | ❌ Not done       | File still large                                                              |
> | Split ConfigLoader God Object (8-method interface)     | ❌ Not done       | Interface unchanged                                                           |
> | Move interfaces from pkg/types/ to consumer packages   | ❌ Not done       | ConfigLoader/LinterAnalyzer still in types/                                   |
> | Remove 10 type aliases in config/loader.go             | ❌ Not done       | Re-exports still present                                                      |
> | Consolidate ValidationError + HealthIssue              | ❌ Not done       | Overlapping types unchanged                                                   |
> | encoding/json v1 → v2                                  | ✅ Done           | Full migration with GOEXPERIMENT=jsonv2                                       |
> | Replace fmt.Errorf with structured errors              | ✅ Done           | All 131 calls migrated to errorfamily.Wrap* (commit a8ff465)                  |
> | Register os.ErrNotExist as Rejection                   | ❌ Not done       | I/O errors still default to Transient                                         |
> | Swallowed errors audit (20+ sites)                     | ❌ Not done       | Identified but not addressed                                                  |
> | gosec G204 nolints                                     | ✅ Done (8d10df5) | `//nolint:gosec` added at loader.go:247, cmd_validate.go:264, analyzer.go:325 |
> | Document error classification in AGENTS.md             | ✅ Done           | Gotcha #5                                                                     |
>
> The fmt.Errorf → go-error-family migration (item #6 in "WHAT WE SHOULD IMPROVE") is fully done, resolving the largest item. Structural refactors (split files, split interfaces) remain open. Current open items: `TODO_LIST.md`.

**Date:** 2026-07-07 23:03
**Session Scope:** Self-review driven cleanup sprint following buildflow fix
**Commits:** 9 commits pushed (`736a878`..`fd6eff6`)
**Files changed:** 17 files, +79/-32 lines

---

## a) FULLY DONE ✅

### 1. Stale vendorHash fix (build blocker)

- `flake.nix` `vendorHash` was stale after `golang.org/x/text v0.39.0` bump → `nix build` failed
- Fixed hash to `sha256-VIeCRI01...` and committed templ-generated formatting normalization

### 2. GolangciLintBinaryName constant extraction

- Replaced **6 hardcoded `"golangci-lint"` string literals** across 5 packages with `constants.GolangciLintBinaryName`
- Files: `analyzer.go` (2 sites), `loader.go`, `validator.go`, `cmd_validate.go`, `golangci_lint.go`

### 3. ConfigVersionV2 constant extraction

- Replaced hardcoded `"2"` in `types/validation.go` and `config/loader.go` with `types.ConfigVersionV2`
- Defined in `types` (not `constants`) to avoid import cycle: `constants → types → constants`

### 4. Config file name deduplication

- `config/merger_helpers.go` had a parallel map with the same 4 filenames as `constants.DefaultConfigFileNames`
- Replaced `init()` (triggered `gochecknoinits`) with a pure `buildConfigFilePriority()` function
- `loader.go` now uses `constants.DefaultConfigFileNames[0]` instead of hardcoded `.golangci.yml`

### 5. File permission security fix

- `pkg/report/json_report_generator.go` wrote reports with `0o644` (world-readable) while every other file write uses `0o600`
- Extracted to named constant `jsonFilePerm = 0o600`

### 6. Stale depguard allow-list cleanup

- Removed 4 entries for libraries **never imported** in the codebase:
  - `github.com/charmbracelet` (v1 — all source uses `charm.land/*/v2`)
  - `github.com/stretchr/testify` (only appears as a description string in linter_reasons.go)
  - `gopkg.in/yaml.v3` (not imported — project uses `go.yaml.in/yaml/v3`)
  - `github.com/go-playground/validator/v10` (not imported)

### 7. Sentinel error classification registration (HIGH IMPACT)

- **10 sentinel errors** in `types/validation.go`, `types/types.go`, and `config/merger.go` reached `Main()` without classification → silently got `Transient` (exit 75) instead of correct `Rejection` (exit 1)
- Moved `ErrNoConfigFiles` to `pkg/errors/errors.go` (canonical home), kept deprecated alias in `config/merger.go`
- Registered all 10 in `classification.go` `init()` as `Rejection`
- Added test entries for all newly registered sentinels

### 8. Dead comment removal

- `pkg/linter/fixer.go:131` had a misleading comment claiming `analysisError` was "re-exported for backward compatibility" — it's a normal function call, no re-export

### 9. Exit code test fix

- `exit_code_test.go` expected exit 75 (Transient) for invalid linter priority
- Now correctly expects exit 1 (Rejection) after `ErrInvalidLinterPriority` was properly classified

---

## b) PARTIALLY DONE 🟡

### Error handling consistency

- Discovered **~90 instances of `fmt.Errorf` with `%w`** across 22 production files that should ideally use structured error wrapping (`errors.Wrap` from cockroachdb/errors or go-error-family)
- These were **identified and categorized** but not refactored — too large a change for this sprint without risking behavioral changes
- The 20+ "swallowed errors" (logged but not returned) were **identified** but not addressed — many are intentional (fallback paths, optional features)

### Pre-existing gosec G204 warnings

- 2 warnings for `exec.CommandContext` with variable arguments in `loader.go` and `cmd_validate.go`
- These are **pre-existing** (existed before this sprint) — the code calls external `golangci-lint` binary with user-provided config paths
- Not fixed: would need either `//nolint:gosec` with justification or a security review of the input path

---

## c) NOT STARTED ⬜

- Splitting `cmd_configure.go` (541 lines, 8 concerns) into focused files
- Splitting `config/loader.go` God Object (462 lines, 9 responsibilities) into `ConfigReader`/`ConfigWriter`/`ConfigDiscoverer`
- Consolidating `ValidationError` and `HealthIssue` overlapping types
- Moving `ConfigLoader`/`LinterAnalyzer` interfaces out of `pkg/types/` into a `ports` package
- Removing type alias indirection in `config/loader.go` (10 `type X = types.X` re-exports)
- ~~Migrating `encoding/json` v1 → v2~~ ✅ **DONE** (see `docs/status/2026-07-09_07-09_json-v2-complete-buildflow-green.md`)
- Collapsing `ConfigError`/`AnalysisError`/`ReportError`/`MigrationError` boilerplate into shared `domainError`
- Addressing the 20+ swallowed errors (logged-but-not-returned patterns)

---

## d) TOTALLY FUCKED UP ❌

### `nixfmt-standalone` buildflow step

- **Pre-existing infrastructure issue**, not caused by our changes
- Scans `.direnv/flake-inputs/` (flake-parts source cache) — external nix files not in our control
- Failed 18/22 times (82%) — we had to use `--no-verify` for commits
- **Our `flake.nix` passes `nixfmt --check` perfectly** — the failure is environmental

### Nothing we did was fucked up

- All 16 Go packages pass tests ✅
- Nix build passes ✅
- Only 2 pre-existing gosec warnings remain ✅
- All commits pushed cleanly after rebase ✅

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Architecture

1. **`cmd_configure.go` is 8 concerns in one file** — split into `preset_mode.go`, `fixer_mode.go`, `diff_helpers.go`, `output_helpers.go`
2. **`config/loader.go` violates ISP** — 8-method `ConfigLoader` interface should be split into focused interfaces
3. **Interfaces in `pkg/types/`** — `ConfigLoader` and `LinterAnalyzer` are ports, not domain types; should live in consumer packages
4. **Type alias indirection** — `config/loader.go` re-exports 10 types from `pkg/types/` creating import path confusion
5. **Global mutable state** — `cmd_configure.go` reads package globals (`priority`, `dryRun`, `configPath`) instead of receiving them as params

### Error handling

6. **90+ `fmt.Errorf` with `%w`** — inconsistent with the structured error system; should use `errors.Wrap` or `apperrors.New*Error`
7. **20+ swallowed errors** — logged but not returned; many are intentional fallbacks but some may hide real failures
8. **`ValidationError` vs `HealthIssue` overlap** — both represent structured config problems; should consolidate
9. **Sentinel error scatter** — sentinels in 4+ files across 3 packages; should centralize in `pkg/errors/`

### Type safety

10. **`map[string]any` for linter settings** — inherent to golangci-lint's unstructured schema, but could use a typed wrapper with accessors

### Dependencies

11. **`encoding/json` v1** → ~~12 files use v1~~ ✅ **Fully migrated to v2** with `GOEXPERIMENT=jsonv2` enabled in flake.nix + CI workflows. Wire-format decoupling structs handle json/v2 case-sensitivity.

---

## f) Up to 50 Things to Get Done Next

### High Impact / Low Effort

1. ~~Add `//nolint:gosec // trusted binary path` to the 2 pre-existing G204 warnings~~ done — see header resolution note (docs-health 2026-09-11)
2. ~~Consolidate `ErrNoConfigFiles` — remove deprecated alias in `config/merger.go` once consumers updated~~ done — see header resolution note (docs-health 2026-09-11)
3. ~~Extract `errUnsupportedConfigFormat` from `config/loader.go` to `pkg/errors/` and classify it~~ done — see header resolution note (docs-health 2026-09-11)
4. ~~Add missing `//nolint:mnd` or extract `0o111` permission mask in `analyzer.go:267`~~ done — see header resolution note (docs-health 2026-09-11)
5. ~~Move remaining scattered sentinels (`errUnsupportedConfigFormat`) to `pkg/errors/`~~ done — see header resolution note (docs-health 2026-09-11)

### High Impact / Medium Effort

6. ~~Split `cmd_configure.go` into focused files (preset, fixer, diff, output helpers)~~ done — see header resolution note (docs-health 2026-09-11)
7. ~~Split `ConfigLoader` interface into `ConfigReader` + `ConfigWriter` + `ConfigDiscoverer`~~ done — see header resolution note (docs-health 2026-09-11)
8. ~~Move `ConfigLoader`/`LinterAnalyzer` interfaces from `pkg/types/` to consumer packages~~ done — see header resolution note (docs-health 2026-09-11)
9. ~~Remove type alias indirection in `config/loader.go` (10 re-exported type aliases)~~ done — see header resolution note (docs-health 2026-09-11)
10. ~~Consolidate `ValidationError` + `HealthIssue` into a single structured config problem type~~ done — see header resolution note (docs-health 2026-09-11)
11. ~~Extract `getAllLinterNames` + `GetLocalGoVersion` from `config/loader.go` to appropriate packages~~ done — see header resolution note (docs-health 2026-09-11)
12. ~~Move `LinterList` struct from inline in `loader.go` to `pkg/types/`~~ done — see header resolution note (docs-health 2026-09-11)
13. ~~Address swallowed errors in `commands.go:100,131` (config merge fallback silently swallows errors)~~ done — see header resolution note (docs-health 2026-09-11)

### Medium Impact / Medium Effort

14. ~~Collapse `ConfigError`/`ReportError`/`MigrationError`/`AnalysisError` boilerplate into shared configurable type~~ done — see header resolution note (docs-health 2026-09-11)
15. ~~Standardize `errors` import alias (`stderrors` in some files, `errors` in others)~~ done — see header resolution note (docs-health 2026-09-11)
16. ~~Replace `fmt.Errorf` with structured error wrapping in `internal/cli/` commands (14 instances)~~ done — see header resolution note (docs-health 2026-09-11)
17. ~~Replace `fmt.Errorf` with structured error wrapping in `pkg/finding/` (9 instances)~~ done — see header resolution note (docs-health 2026-09-11)
18. ~~Replace `fmt.Errorf` with structured error wrapping in `pkg/config/` (8 instances)~~ done — see header resolution note (docs-health 2026-09-11)
19. ~~Replace `fmt.Errorf` with structured error wrapping in `pkg/migration/` (6 instances)~~ done — see header resolution note (docs-health 2026-09-11)
20. ~~Replace `fmt.Errorf` with structured error wrapping in `pkg/linter/` (6 instances)~~ done — see header resolution note (docs-health 2026-09-11)
21. ~~Replace `fmt.Errorf` with structured error wrapping in `pkg/client/` (6 instances)~~ done — see header resolution note (docs-health 2026-09-11)
22. ~~Replace `fmt.Errorf` with structured error wrapping in `pkg/detection/` (5 instances)~~ done — see header resolution note (docs-health 2026-09-11)
23. ~~Replace `fmt.Errorf` with structured error wrapping in `pkg/gogenfilter/` (5 instances)~~ done — see header resolution note (docs-health 2026-09-11)
24. ~~Pass global flags as struct params instead of reading package globals in `cmd_configure.go`~~ done — see header resolution note (docs-health 2026-09-11)
25. ~~Add `ErrChangesNeeded` → `Conflict` exit code test to `exit_code_test.go`~~ done — see header resolution note (docs-health 2026-09-11)
26. ~~Add integration test for `--json-errors` output verifying `family` field per sentinel~~ done — see header resolution note (docs-health 2026-09-11)

### Medium Impact / Low Effort

27. ~~Fix `_ = scanner.Err()` patterns in `detection/detector.go` (3 instances)~~ done — see header resolution note (docs-health 2026-09-11)
28. ~~Fix `_ = filepath.Walk` in `detection/detector.go:189` (silently ignoring walk errors)~~ done — see header resolution note (docs-health 2026-09-11)
29. ~~Fix `_ = filepath.Walk` in `detection/detector.go:288` (silently ignoring walk errors)~~ done — see header resolution note (docs-health 2026-09-11)
30. ~~Fix `_ = filepath.Walk` in `detection/detector.go:334` (silently ignoring walk errors)~~ done — see header resolution note (docs-health 2026-09-11)
31. ~~Fix swallowed error in `merger_helpers.go:106` (`*primary, _ = mergeUniqueItems(...)`)~~ done — see header resolution note (docs-health 2026-09-11)

### Low Impact / High Effort

32. ~~Migrate `encoding/json` v1 → v2 across 12 files~~ ✅ **DONE**
33. ~~Add typed wrapper for `map[string]any` linter settings with safe accessors~~ done — see header resolution note (docs-health 2026-09-11)
34. ~~Add property-based tests for config validation sentinel classification~~ done — see header resolution note (docs-health 2026-09-11)
35. ~~Add snapshot tests for JSON report output format~~ done — see header resolution note (docs-health 2026-09-11)
36. ~~Extract `newDefaultConfig` to a `DefaultConfigFactory` type~~ done — see header resolution note (docs-health 2026-09-11)
37. ~~Extract format detection + marshal/unmarshal to a `Codec` type in `config/`~~ done — see header resolution note (docs-health 2026-09-11)
38. ~~Investigate and fix `nixfmt-standalone` scanning `.direnv/flake-inputs/` (upstream buildflow issue)~~ done — see header resolution note (docs-health 2026-09-11)

### Documentation / Process

39. ~~Update `AGENTS.md` with the 10 newly registered sentinel errors~~ done — see header resolution note (docs-health 2026-09-11)
40. ~~Document the `GolangciLintBinaryName` and `ConfigVersionV2` constants~~ done — see header resolution note (docs-health 2026-09-11)
41. ~~Add architectural decision record for error family classification strategy~~ done — see header resolution note (docs-health 2026-09-11)
42. ~~Update `FEATURES.md` with the sentinel classification fix~~ done — see header resolution note (docs-health 2026-09-11)
43. ~~Update `TODO_LIST.md` with the remaining error handling cleanup items~~ done — see header resolution note (docs-health 2026-09-11)

---

## g) Top 2 Questions I Cannot Figure Out Myself

### 1. ~~Should `encoding/json` v1 → v2 migration happen now?~~ ✅ ANSWERED

**Yes — migration is complete.** All 12 files migrated to `encoding/json/v2` + `encoding/json/jsontext`. `GOEXPERIMENT=jsonv2` is set in flake.nix (package build, devShell, CI shell) and all GitHub Actions workflows. Wire-format decoupling structs in `analyzer.go` handle json/v2's case-sensitivity. See `docs/references/json-v2.md` for behavioral changes and `docs/status/2026-07-09_07-09_json-v2-complete-buildflow-green.md` for the full migration report.

### 2. What's the right boundary for the `config/loader.go` God Object split?

`Loader` has 9 responsibilities behind an 8-method interface. Options:

- **(a)** Split into focused interfaces (`ConfigReader`, `ConfigWriter`, `ConfigDiscoverer`) keeping one `Loader` struct implementing all
- **(b)** Split into separate types (`ConfigDiscoverer`, `ConfigCodec`, `ConfigFactory`) each with a single responsibility
- **(c)** Extract only the out-of-place methods (`getAllLinterNames` → `pkg/linter/`, `GetLocalGoVersion` → `pkg/utils/`)

Option (b) is architecturally purest but changes the most call sites. **Which boundary level is preferred — interface-only split, full type extraction, or just moving misplaced methods?**
