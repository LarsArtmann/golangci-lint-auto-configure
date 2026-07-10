# Deep Architecture, Configuration & Data Model Review

**Date:** 2026-07-10  
**Reviewer:** Crush (AI Senior Staff Engineering Partner)  
**Scope:** Architecture, data models, linter/formatter coverage, configuration handling  
**Benchmark:** [golangci-lint v2.12.2](https://github.com/golangci/golangci-lint) upstream — [linters config](https://golangci-lint.run/docs/linters/configuration/), [formatters config](https://golangci-lint.run/docs/formatters/configuration/)

---

## Executive Summary

This project is a **well-architected** Go CLI that auto-configures golangci-lint. The separation of concerns (constants as data, types as models, linter as analysis/fixing engine, config as I/O, migration as v1→v2 transformer) is clean and maintainable. The interface-based design enables testability, and the wire-format decoupling from report types is a mature pattern.

However, there are **critical data accuracy gaps** against the upstream golangci-lint v2.12.2: missing linters, incorrectly categorized entries, stale version whitelists, and a few data model weaknesses that should be addressed. The project tracks 113 of 114 upstream linters — one is missing entirely, and several removed-in-v2 linters are treated as active.

| Category                          | Findings                                                   | Severity                 |
| --------------------------------- | ---------------------------------------------------------- | ------------------------ |
| Linter coverage gaps              | 1 missing, 1 miscategorized, 6 missing deprecation entries | **High**                 |
| Formatter-as-linter confusion     | 2 entries in wrong map                                     | Medium                   |
| Stale migration version whitelist | v2.11/v2.12 not recognized                                 | Medium                   |
| Settings type safety              | `map[string]any` throughout                                | Low (architectural debt) |
| Architecture quality              | Excellent                                                  | —                        |
| Data model quality                | Good with room for improvement                             | —                        |

---

## 1. Linter Coverage Audit (vs Upstream v2.12.2)

### 1.1 Complete Linter Inventory

The upstream documents **114 linters** (including 2 deprecated: `gomodguard`, `wsl`). This project's `LinterPriorities` map tracks **114 entries**, but the entries are not all correctly categorized.

### 1.2 CRITICAL: Missing Linter — `clickhouselint`

**`clickhouselint`** (since v2.12.0) is a new linter that provides opinionated best practices for the ClickHouse client. It is **completely absent** from this project:

- Not in `LinterPriorities` ❌
- Not in `LinterReasons` ❌
- Not in `LinterMinVersions` ❌
- Not in any preset ❌
- Zero references in the entire codebase ❌

**Impact:** Users running `configure` with golangci-lint v2.12.0+ will never be recommended `clickhouselint`, even if they use ClickHouse. The linter will silently fall through to `LinterPriorityOptional` with a generic reason.

**Recommended fix:**

- Add to `LinterPriorities` as `LinterPriorityMedium` (specialized database linter)
- Add to `LinterReasons`: `"Opinionated best practices for ClickHouse client"`
- Add to `LinterMinVersions`: `"v2.12.0"`
- Consider adding to `reference` preset if ClickHouse projects are common

### 1.3 CRITICAL: `exportloopref` Treated as Active Linter

`exportloopref` was **removed in golangci-lint v2** (it's obsolete since Go 1.22's loop variable semantics). Yet this project lists it as:

- `LinterPriorities["exportloopref"] = LinterPriorityMedium` (`linter_priorities.go:85`)
- `LinterReasons["exportloopref"] = "Checks for pointers to enclosing loop variables"` (`linter_reasons.go:85`)

It is **NOT** in `DeprecatedLinters`, so no replacement is offered. Users who have it enabled will never be told to remove it.

**Impact:** Users with `exportloopref` in their config get no deprecation warning or auto-fix. The linter doesn't exist in v2, so golangci-lint itself will error.

**Recommended fix:**

- Remove from `LinterPriorities` and `LinterReasons`
- Add to `DeprecatedLinters`:
  ```go
  "exportloopref": {
      Replacement: "copyloopvar",
      Reason:      "exportloopref was removed in golangci-lint v2 (Go 1.22 fixed loop variable semantics), use copyloopvar instead",
  },
  ```

### 1.4 Missing Deprecation Entries (v1 Linters Removed in v2)

The `DeprecatedLinters` map (`rules.go`) correctly handles: `wsl`, `deadcode`, `varcheck`, `structcheck`, `gosimple`, `exhaustivestruct`, `interfacer`, `maligned`, `nosnakecase`, `gomodguard`.

**Missing entries** (all removed in v2):

| Removed Linter  | Replacement   | Notes                                                |
| --------------- | ------------- | ---------------------------------------------------- |
| `exportloopref` | `copyloopvar` | **Most impactful** — users may still have this       |
| `golint`        | `revive`      | Common in old configs                                |
| `scopelint`     | `copyloopvar` | Predecessor of exportloopref                         |
| `tenv`          | `usetesting`  | Test environment linter                              |
| `ifshort`       | _(none)_      | No direct replacement; just remove                   |
| `execinquery`   | _(none)_      | SQL-related; `unqueryvet` is the spiritual successor |

**Missing v1 alternative names** (renamed in v2, old names no longer valid):

| Old Name    | Current Name  |
| ----------- | ------------- |
| `gas`       | `gosec`       |
| `goerr113`  | `err113`      |
| `gomnd`     | `mnd`         |
| `logrlint`  | `loggercheck` |
| `megacheck` | `staticcheck` |
| `vet`       | `govet`       |
| `vetshadow` | `govet`       |

**Impact:** Users migrating from v1 configs with these linters get no guidance. The migration command handles some of this, but the fixer/analyzer won't catch them.

### 1.5 Formatter-as-Linter Confusion

In golangci-lint v2, `gofmt` and `gci` are **formatters**, not linters. Yet both appear in:

- `LinterPriorities` as `LinterPriorityMedium` (`linter_priorities.go:77-78`)
- `LinterReasons` with linter descriptions (`linter_reasons.go:77-78`)

**Impact:** Low — these entries are dead data. The analyzer calls `golangci-lint linters --json` which only returns actual linters, so `gofmt`/`gci` will never appear as disabled linters. But their presence in the linter maps is misleading and violates the single-source-of-truth principle.

**Recommended fix:** Remove `gofmt` and `gci` from `LinterPriorities` and `LinterReasons`. They are already correctly tracked in `FormatterInfo`, `FormatterPriorities`, and `FormatterReasons`.

---

## 2. Formatter Coverage Audit

### 2.1 Complete Formatter Inventory

The upstream documents **6 formatters**: `gci`, `gofmt`, `gofumpt`, `goimports`, `golines`, `swaggo`.

This project's `FormatterInfo` map tracks all 6. ✅ **Complete coverage.**

### 2.2 Formatter Priorities

| Formatter   | Project Priority | Assessment                                                                  |
| ----------- | ---------------- | --------------------------------------------------------------------------- |
| `gofumpt`   | High             | ✅ Correct — strictest formatter, superset of gofmt                         |
| `golines`   | High             | ✅ Correct — fixes long lines, high value                                   |
| `gofmt`     | Medium           | ⚠️ Should be Low — superseded by gofumpt (already in `RedundantFormatters`) |
| `goimports` | Medium           | ✅ Correct                                                                  |
| `gci`       | Medium           | ✅ Correct — import organization                                            |
| `swaggo`    | Low              | ✅ Correct — specialized for Swagger projects                               |

### 2.3 Formatter Settings Coverage

The upstream documents settings for 5 of 6 formatters (swaggo has none). This project's `DefaultFormatterSettings` only has defaults for `golines` (`max-len: 120`). This is acceptable — the other formatters have reasonable built-in defaults, and `gofumpt`/`goimports`/`gci` work well without explicit settings.

**Missing opportunity:** `gci` could benefit from a default `sections: [standard, default]` since the upstream default is `["standard", "default"]` and users who want `localmodule` section need to configure it. However, injecting this could override user intent, so the current conservative approach is defensible.

---

## 3. Data Model Review

### 3.1 Strengths

#### Strong Typing for Names

```go
type LinterName string     // Prevents typos, enables compile-time checking
type FormatterName string  // Same for formatters
```

These types are used consistently throughout constants, preventing the most common data integrity bug class (typos in linter names).

#### Config Type Tag Strategy

The three-family tag policy is well-designed:

| Family            | JSON                     | YAML/TOML | Rationale                          |
| ----------------- | ------------------------ | --------- | ---------------------------------- |
| Report types      | PascalCase (tag-free)    | n/a       | Go-native, zero tag cost           |
| Config types      | kebab                    | kebab     | Round-trips `.golangci.yml` schema |
| Wire-format types | as-is (matches upstream) | n/a       | External JSON compatibility        |

#### Wire-Format Decoupling

`golangciLinterEntry` and `golangciFormatterEntry` in `analyzer.go` are dedicated structs that match golangci-lint's JSON wire format, with conversion methods to domain types. This correctly handles the quirky wire format (capitalized wrapper keys `"Enabled"`/`"Disabled"` but lowercase field keys `"name"`, `"autoFix"`).

#### Generic Set Type

`types.Set[T]` provides type-safe set operations for linter/formatter management, used throughout the fixer pipeline.

### 3.2 Weaknesses

#### `map[string]any` for Linter/Formatter Settings — The Biggest Data Model Gap

```go
// config_types.go
type LintersConfig struct {
    Settings map[string]any `json:"settings,omitempty" ...`
}

type FormattersConfig struct {
    Settings map[string]any `json:"settings,omitempty" ...`
}
```

And in constants:

```go
var DefaultLinterSettings = map[types.LinterName]any{
    "depguard": map[string]any{
        "rules": map[string]any{
            "main": map[string]any{
                "allow": []string{"$gostd", "$module"},
            },
        },
    },
    // ...
}
```

**Problems:**

1. **No compile-time safety** — typos in setting keys (`"max-complexity"` vs `"max_complexity"`) are silent bugs
2. **No validation** — invalid values (e.g., `max-complexity: -1` for cyclop) are accepted
3. **No IDE autocomplete** — consumers can't discover available settings
4. **Brittle defaults** — the `DefaultLinterSettings` map uses untyped nested maps that are easy to get wrong

**Assessment:** This is a conscious trade-off. golangci-lint's own config uses `map[string]any` for settings, so matching that shape enables round-trip fidelity. Fully typing all 84 linters' settings would be a massive effort with ongoing maintenance burden. The current approach is pragmatic but creates a class of bugs that only manifest at runtime.

**Recommendation (long-term):** Consider generated typed settings from golangci-lint's JSON Schema, or at minimum, add a validation step that checks setting keys against a known-good list.

#### `LinterInfo` Mixes Concerns

```go
type LinterInfo struct {
    Name        LinterName
    Description string
    Groups      []string
    Fast        bool
    AutoFix     bool
    Deprecated  bool
    Since       string
    OriginalURL string
}
```

This struct serves dual purposes: it's both the **wire-format parse target** (from `golangci-lint linters --json`) and the **report/display type** (used in `ConfigAnalysis`). The conversion in `toLinterInfo()` is trivial now, but as the project evolves, these concerns may diverge.

**Assessment:** Acceptable for current scope. The conversion method `toLinterInfo()` provides the seam for future divergence.

#### `LinterReplacement.MinVersion` is Stringly-Typed

```go
type LinterReplacement struct {
    Replacement LinterName
    Reason      string
    MinVersion  string `json:",omitempty"`
}
```

`MinVersion` is a raw string compared via `semver.Compare()`. A typed `Version` struct with validation would be safer, but the current approach works because all values are controlled constants.

---

## 4. Architecture Review

### 4.1 Package Structure — Excellent

```
pkg/
├── constants/     # Static data: priorities, reasons, presets, rules, defaults
├── types/         # Domain models: Config, LinterInfo, Set[T], interfaces
├── linter/        # Analysis & fixing engine: Analyzer, Fixer, FormatterManager
├── config/        # Config I/O: Loader, Merger (load, save, discover, validate)
├── migration/     # v1→v2 transformer: Migrator, rules, validators
├── detection/     # Project detection: swaggo, generated files
├── finding/       # Finding model: SARIF/JSON output, converters
├── report/        # Report generation: HTML, JSON
├── errors/        # Error classification: go-error-family integration
├── version/       # Self-initializing version info
├── gogenfilter/   # Generated file detection (two-phase scanner)
├── ui/            # Terminal output formatting
├── client/        # HTTP client for version checks
└── utils/         # Git, retry utilities
```

**Assessment:** Clean separation. Each package has a single responsibility. Dependencies flow inward (constants → types ← linter/config/migration). No circular dependencies.

### 4.2 Interface Design — Well-Abstracted

```go
type ConfigLoader interface {
    LoadConfig(path string) (*Config, error)
    FindConfigFile(startDir string) (string, error)
    FindOrGetDefaultConfigPath(startDir string) string
    SaveConfig(config *Config, path string) error
    ValidateConfig(config *Config) []error
    GetLintersEnabled(config *Config) []string
    GetLintersDisabled(config *Config) []string
    CreateDefaultConfig(ctx context.Context) *Config
}

type LinterAnalyzer interface {
    AnalyzeConfig(ctx context.Context, configPath string) (*ConfigAnalysis, error)
    FindBinary(ctx context.Context) error
    CheckVersion(ctx context.Context) error
    GetDetectedVersion() string
    GetSummary(analysis *ConfigAnalysis) string
    GetLintersByPriority(...) []LinterRecommendation
}
```

**Assessment:** Interfaces are consumer-defined (ISP-compliant). Testability is excellent — both interfaces have mock implementations in tests. The `FS` interface in config/loader.go enables filesystem abstraction.

### 4.3 Fixer Pipeline — Well-Structured

The fixer follows a clear pipeline:

```
FixConfig
  ├── LoadConfig
  ├── detectVersion (for version-gated replacements)
  ├── runPreFlightChecks
  │     ├── preFixInvalidDurations
  │     ├── preFixVersion
  │     ├── preFixDeprecatedLinters (version-gated)
  │     └── preFixTypecheck
  ├── [dry-run early returns]
  └── analyzeAndFix
        ├── AnalyzeConfig (parallel linter + formatter parsing)
        └── applyLintersFix
              ├── applyAllFixes
              │     ├── replaceLinters (deprecated → successors)
              │     ├── EnableCoreFormatters (gci, gofumpt, goimports)
              │     ├── EnableGolinesFormatter
              │     ├── EnableSwaggoFormatter (conditional)
              │     ├── RemoveRedundantLinters (lll when golines enabled)
              │     ├── RemoveRedundantGofmt (when gofumpt enabled)
              │     └── enableRecommendedLinters (by priority)
              └── applyAndSave
                    ├── updateGoVersion
                    ├── updateRunnerSettings
                    ├── updateBuildTags (Go experiments)
                    ├── updateOutputFormats
                    ├── updateConfigFromSets
                    │     ├── injectDefaultSettings
                    │     └── injectDefaultFormatterSettings
                    ├── updateGeneratedExclusions (gogenfilter scan)
                    ├── updateExclusionRules
                    └── updateIssuesSettings
```

**Assessment:** The pipeline is comprehensive and idempotent. Each step is independently testable. The `fixCounts` struct tracks changes by category, and the `counts.total()==0` guard prevents empty writes.

**One concern:** The `fixCounts.normalization` counter is critical — every mutation in `applyAndSave` MUST increment it, or the `total()==0` guard silently discards all changes. This is a footgun documented in AGENTS.md but could be enforced with a wrapper pattern.

### 4.4 Migration System — Functional but Stale

The migration system handles v1→v2 transformation with rule-based property moves. It's well-structured with:

- `MigrationRules` struct for configurable mappings
- Separate functions for each migration step
- Post-migration validation via `golangci-lint config verify`

**Issue: Stale version whitelist.** The `validVersions()` function hardcodes:

```go
return types.NewSet("2", "2.8", "2.8.0", "2.9", "2.9.0", "2.10", "2.10.0", "2.10.1")
```

This is missing `"2.11"`, `"2.11.0"`, `"2.12"`, `"2.12.0"`, `"2.12.1"`, `"2.12.2"`. When `migrateVersion()` encounters `version: "2.12"` in a config, it won't find it in the valid set and will reset it to `"2"`. While `"2"` is the canonical value, the migrator shouldn't treat `"2.12"` as invalid — it should recognize it as a valid v2 version.

**Recommended fix:** Replace the whitelist with prefix matching:

```go
func (r *MigrationRules) IsValidVersion(version string) bool {
    return version == "2" || strings.HasPrefix(version, "2.")
}
```

### 4.5 Error Handling — Mature

The `go-error-family` integration provides:

- Sentinel error registration with Families (Rejection/Conflict/Transient/Corruption/Infrastructure)
- `ConfigError`, `ReportError`, `MigrationError` → always Rejection (type-level)
- `AnalysisError` → delegates to cause-chain sentinels
- BSD sysexits exit codes via `errorfamily.ExitCode(err)`
- `--json-errors` for structured error output

**Assessment:** This is above industry standard for a CLI tool. The error classification enables meaningful exit codes and machine-readable error reporting.

### 4.6 Version Checking — Robust

The version checker:

- Tries JSON output first, falls back to text parsing
- Retries on "parallel golangci-lint is running" errors
- Enforces minimum version (`v2.10.1`)
- Warns on version mismatch with tested version (`v2.12.2`)
- Supports version-gated linter recommendations via `LinterMinVersions`

**Assessment:** Well-designed. The retry logic handles the common case of multiple golangci-lint invocations in CI.

---

## 5. Default Settings Audit

### 5.1 Linter Defaults — Mostly Sound

| Linter            | Default                                | Upstream Default                   | Assessment                                        |
| ----------------- | -------------------------------------- | ---------------------------------- | ------------------------------------------------- |
| `depguard`        | allow `$gostd`, `$module`              | deny all non-stdlib                | ✅ Correct — prevents build breakage              |
| `ireturn`         | allow standard interfaces              | allow `anon, error, empty, stdlib` | ✅ Correct — matches upstream defaults            |
| `gocritic`        | disable `ifElseChain`                  | all stable checks enabled          | ✅ Reasonable — `ifElseChain` is noisy            |
| `exhaustruct`     | exclude `os/exec.Cmd`                  | no exclusions                      | ✅ Good — `exec.Cmd` has many optional fields     |
| `revive`          | disable `exported`, `package-comments` | no rules                           | ✅ Good — these are extremely noisy               |
| `varnamelen`      | ignore common short names              | no ignores                         | ✅ Excellent — prevents noise                     |
| `gomoddirectives` | `replace-local: true`                  | `replace-local: false`             | ⚠️ More permissive than upstream                  |
| `cyclop`          | `max-complexity: 12`                   | `max-complexity: 10`               | ⚠️ Slightly more lenient                          |
| `ginkgolinter`    | forbid focus + spec pollution          | all `false`                        | ✅ Good — catches test pollution                  |
| `testifylint`     | enable-all except `go-require`         | no defaults                        | ✅ Good — `go-require` is noisy for HTTP handlers |
| `makezero`        | `always: true`                         | `always: false`                    | ⚠️ Strict — may surprise users                    |

### 5.2 Missing Defaults Worth Considering

| Linter      | Suggested Default                  | Rationale                                  |
| ----------- | ---------------------------------- | ------------------------------------------ |
| `wrapcheck` | `report-internal-errors: false`    | Prevents noise from internal error returns |
| `funlen`    | `lines: 80` (upstream default: 60) | More realistic for modern Go code          |
| `mnd`       | `ignored-files: ["cmd/.*"]`        | CLI files legitimately have magic numbers  |

**Counter-argument:** Conservative defaults are safer. Adding more defaults increases the risk of overriding user intent. The current approach of only defaulting linters that **break builds without configuration** (depguard, ireturn) is defensible.

---

## 6. Presets Audit

### 6.1 Preset Coverage

| Preset        | Linter Count | Assessment                                                        |
| ------------- | ------------ | ----------------------------------------------------------------- |
| `minimal`     | 5            | ✅ Correct — core safety linters                                  |
| `standard`    | 8            | ✅ Good balance for most projects                                 |
| `strict`      | 17           | ✅ Good for CI/CD quality enforcement                             |
| `security`    | 1            | ⚠️ Only `gosec` — could add `gocritic` with security tags         |
| `performance` | 4            | ✅ Good — `ineffassign`, `prealloc`, `unconvert`, `perfsprint`    |
| `reference`   | 62           | ✅ All critical + high priority (verified by data integrity test) |

### 6.2 Missing Preset Opportunity

No preset includes **formatters**. A `format` preset that enables `gci`, `gofumpt`, `goimports` would be valuable since the fixer auto-enables these anyway. Currently formatters are only managed through the fixer flow.

---

## 7. Data Integrity & Testing

### 7.1 Existing Integrity Tests

The `data_integrity_test.go` enforces:

- `LinterMinVersions` entries exist in `LinterPriorities` ✅
- `reference` preset only contains linters from `LinterPriorities` ✅
- `reference` preset contains ALL critical + high priority linters ✅
- `DisabledLinters` entries don't appear in `LinterPriorities` or `LinterReasons` ✅
- `DisabledLinters` entries have non-empty reasons ✅

### 7.2 Missing Integrity Tests

| Test                                                               | What It Would Catch                                           |
| ------------------------------------------------------------------ | ------------------------------------------------------------- |
| `LinterPriorities` ⊆ `LinterReasons`                               | Linters with priority but no reason                           |
| `LinterReasons` ⊆ `LinterPriorities`                               | Reasons for linters not in priorities                         |
| `DeprecatedLinters` ∉ `LinterPriorities`                           | Deprecated linters that can still be recommended              |
| `FormatterPriorities` ⊆ `FormatterReasons`                         | Formatters with priority but no reason                        |
| `LinterPriorities` ∩ `FormatterInfo` = ∅                           | Formatters incorrectly in linter maps (would catch gofmt/gci) |
| `DeprecatedLinters` values exist in `LinterPriorities` or upstream | Replacement targets are valid linters                         |

The **last test** would have caught the `gofmt`/`gci`-in-linter-priorities issue automatically.

---

## 8. Cross-Cutting Concerns

### 8.1 Configuration Round-Trip Fidelity

The config loader/saver uses the same `Config` struct with kebab-case YAML tags, ensuring round-trip fidelity. The `KnownFields(false)` setting on the YAML decoder is correct — it allows unknown fields (for forward compatibility with new golangci-lint versions) without error.

**Risk:** Unknown fields are silently dropped on save. If golangci-lint adds a new top-level key, this tool will delete it when fixing the config. This is a known trade-off documented in the codebase.

### 8.2 Generated File Detection

The `gogenfilter` integration provides two-phase generated file detection (filename heuristic → content verification). This is more sophisticated than golangci-lint's built-in `strict`/`lax` modes and provides better exclusion patterns.

### 8.3 Go Experiment Build Tags

The fixer auto-injects Go experiment build tags (`goexperiment.arenas`, `goexperiment.jsonv2`, etc.) into `run.build-tags`. This is a thoughtful feature that enables analysis of code using experimental Go features, but it's opinionated — not all users want experiment tags in their config.

---

## 9. Recommendations — Prioritized

### P0 — Data Accuracy (Fix Immediately)

1. **Add `clickhouselint`** to `LinterPriorities`, `LinterReasons`, and `LinterMinVersions`
2. **Move `exportloopref` to `DeprecatedLinters`** with replacement `copyloopvar`
3. **Remove `gofmt` and `gci`** from `LinterPriorities` and `LinterReasons` (they're formatters)
4. **Fix `validVersions()`** in migration rules — use prefix matching instead of hardcoded list

### P1 — Deprecation Coverage (Fix Soon)

5. **Add missing v1 removed linters** to `DeprecatedLinters`: `golint`, `scopelint`, `tenv`, `ifshort`, `execinquery`
6. **Add v1 alternative names** to `DeprecatedLinters`: `gas`, `goerr113`, `gomnd`, `logrlint`, `megacheck`, `vet`, `vetshadow`

### P2 — Data Integrity (Improve Safeguards)

7. **Add cross-map integrity tests** (Section 7.2) to prevent formatter-in-linter-map regressions
8. **Add test: `LinterPriorities` keys ⊆ `LinterReasons` keys** and vice versa
9. **Add test: `DeprecatedLinters` keys ∉ `LinterPriorities`**

### P3 — Architecture Improvements (Nice to Have)

10. **Consider a `format` preset** that enables core formatters
11. **Long-term: typed linter settings** via code generation from JSON Schema
12. **Consider wrapping config mutations** in a counter-incrementing pattern to prevent the `normalization==0` footgun

---

## 10. Overall Assessment

| Dimension              | Score | Notes                                                     |
| ---------------------- | ----- | --------------------------------------------------------- |
| **Architecture**       | 9/10  | Clean separation, good interfaces, testable               |
| **Data Models**        | 7/10  | Strong typing for names; `map[string]any` is the weakness |
| **Linter Coverage**    | 6/10  | 1 missing, 1 miscategorized, 6+ missing deprecations      |
| **Formatter Coverage** | 9/10  | All 6 tracked, priorities sensible                        |
| **Error Handling**     | 9/10  | go-error-family integration, BSD exit codes               |
| **Testing**            | 8/10  | BDD specs, data integrity tests; missing cross-map checks |
| **Maintainability**    | 8/10  | Static data is easy to update; staleness is the risk      |

**Bottom line:** The architecture is excellent. The data accuracy gaps against upstream are the primary concern — they're easy to fix but currently mean the tool silently misses linters and misclassifies removed ones. Adding cross-map integrity tests would prevent future regressions of this class.

---

## Resolution Status (Updated 2026-07-10)

| Priority | Items                                                           | Status                                                                                                |
| -------- | --------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| **P0**   | clickhouselint, exportloopref, gofmt/gci removal, validVersions | **DONE**                                                                                              |
| **P1**   | Missing v1 removed linters, alternative names                   | **DONE**                                                                                              |
| **P2**   | Cross-map integrity tests, empty Replacement handling           | **DONE**                                                                                              |
| **P3.1** | Typed linter/formatter settings (SettingsConverter)             | **DONE** — `pkg/constants/linter_settings.go`                                                         |
| **P3.2** | configChangeRecorder closure-based counting                     | **DONE** — `pkg/linter/fixer_recorder.go`, applied consistently in `applyAllFixes` and `applyAndSave` |
| **P3.3** | Format preset (minimal linters + core formatters)               | **DONE** — `pkg/constants/presets.go`, `PresetFormatters` map                                         |
