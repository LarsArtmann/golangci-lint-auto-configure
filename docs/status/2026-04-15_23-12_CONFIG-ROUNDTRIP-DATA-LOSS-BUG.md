# Comprehensive Status Report: 2026-04-15 23:12

## Executive Summary

**Critical data-loss bug discovered** in the config roundtrip (load → modify → save). Top-level
`linters-settings:` YAML keys are **silently dropped** because `types.Config` has no field for them.
This caused real damage in downstream projects (project-dependency-graph, gogenfilter) where
depguard allow-lists were deleted, making depguard block ALL external imports.

---

## A) FULLY DONE ✓

### Test Baseline

| Suite            | Specs | Status  | Coverage |
| ---------------- | ----- | ------- | -------- |
| `pkg/config`     | 34    | ✅ PASS | 64.3%    |
| `pkg/linter`     | 35    | ✅ PASS | 79.3%    |
| `internal/cli`   | 19    | ✅ PASS | 11.0%    |
| All other suites | —     | ✅ PASS | —        |

### Existing Migration Infrastructure

The `pkg/migration/` package already handles `linters-settings` → `linters.settings` correctly:

- `migration.Config` has `LintersSettingsV1 map[string]any` (`config_types.go:14`)
- `migrateLintersSettings()` copies V1 settings → `Linters.Settings` (`migrations.go:14-31`)
- Full test coverage for this migration path

### Merger Infrastructure

`pkg/config/merger.go` already has `mergeSettingsMaps()` that correctly merges
`map[string]any` settings. The merger preserves settings — the problem is the loader.

---

## B) PARTIALLY DONE

### Investigation of the Bug

**What's done:**

- ✅ Root cause identified and reproduced
- ✅ Verified depguard blocks all imports without settings (both gogenfilter, project-dependency-graph)
- ✅ Traced the exact code path: `LoadConfig → yaml.Unmarshal (drops linters-settings) → modify → yaml.Marshal (gone)`
- ✅ Confirmed `updateConfigFromSets()` does NOT touch `Settings` (good — it's not the destroyer)
- ✅ Confirmed `golangci-lint fmt` does NOT destroy `linters-settings` (it's our tool's load/save cycle)

**What's NOT done:**

- ❌ No fix implemented yet
- ❌ No tests for roundtrip fidelity of settings
- ❌ No tests for top-level `linters-settings` preservation

---

## C) NOT STARTED

1. **Fix: Add `LintersSettingsV1` field to `types.Config`**
   - Capture top-level `linters-settings` during YAML unmarshal
   - Mirror the pattern from `migration.Config`

2. **Fix: Auto-migrate V1 settings to `linters.settings` in `config/loader.go`**
   - In `LoadConfigResult()`, after unmarshal: merge `LintersSettingsV1` → `cfg.Linters.Settings`
   - Log a warning when this migration occurs

3. **Tests: Roundtrip fidelity for linter settings**
   - Test that `linters-settings:` at top level survives load → save
   - Test that `linters.settings:` nested survives load → modify → save
   - Test that depguard rules, gomoddirectives, funlen thresholds all survive

4. **Tests: Merger roundtrip fidelity**
   - Verify `Merger.MergeConfigs()` preserves settings from all input configs

5. **Fix: Restore project-dependency-graph depguard settings**
   - The commit `41dd6eb` deleted legitimate depguard allow-lists
   - Need to restore or recreate them with proper v2 nesting

6. **Fix: Add depguard settings to gogenfilter**
   - gogenfilter has depguard enabled with zero config → blocks its only dep

---

## D) TOTALLY FUCKED UP 💥

### The Bug: Silent `linters-settings` Data Loss

**Severity: CRITICAL — P0**

| Aspect                | Detail                                                                                     |
| --------------------- | ------------------------------------------------------------------------------------------ |
| **Root cause**        | `types.Config` has no `linters-settings` YAML field                                        |
| **Impact**            | Any config with top-level `linters-settings:` loses ALL linter settings on roundtrip       |
| **Affected path**     | `configure` command (load → analyze → fix → save)                                          |
| **Silent failure**    | No error, no warning — settings just vanish                                                |
| **Real-world damage** | project-dependency-graph lost depguard allow-lists; gogenfilter has depguard with no rules |

**The exact code path:**

```
types.Config struct {
    Version    string           `yaml:"version"`      ← OK
    Run        RunConfig        `yaml:"run"`          ← OK
    Output     OutputConfig     `yaml:"output"`       ← OK
    Linters    LintersConfig    `yaml:"linters"`      ← OK
    Formatters FormattersConfig `yaml:"formatters"`   ← OK
    Issues     IssuesConfig     `yaml:"issues"`       ← OK
    // NO field for top-level "linters-settings" ← 💀
}
```

When YAML has:

```yaml
linters-settings:       ← SILENTLY IGNORED by yaml.Unmarshal
    depguard:
        rules: ...
```

After save → gone. Config now has `depguard` enabled but **no rules** → blocks everything.

### Downstream Impact Verified

```
# project-dependency-graph (after the bad commit)
$ golangci-lint run --enable-only depguard ./...
discover.go:11:2: import 'golang.org/x/mod/modfile' is not allowed (depguard)
main.go:9:2: import 'github.com/larsartmann/cmdguard/pkg/cmdguard/v2' is not allowed (depguard)
main.go:10:2: import 'github.com/spf13/cobra' is not allowed (depguard)

# gogenfilter (depguard added without settings)
$ golangci-lint run --enable-only depguard ./...
example_test.go:6:2: import 'github.com/LarsArtmann/gogenfilter' is not allowed (depguard)
sqlc.go:11:2: import 'github.com/go-faster/yaml' is not allowed (depguard)
```

---

## E) WHAT WE SHOULD IMPROVE

### Architecture Issues

1. **Two separate `Config` structs doing the same job**
   - `types.Config` (used by configure/fixer) — does NOT capture `linters-settings`
   - `migration.Config` (used by migrate) — DOES capture `linters-settings`
   - These should be unified, or `types.Config` should learn about V1 keys

2. **No roundtrip fidelity tests**
   - There's no test that loads a config, saves it, and verifies all fields survive
   - This is a missing quality gate that would have caught the bug immediately

3. **`map[string]any` for settings is a free-for-all**
   - `LintersConfig.Settings` is `map[string]any` — no type safety
   - But this is hard to fix without a code generator from golangci-lint schemas
   - Low priority; the immediate fix is preservation, not type safety

4. **No warning when config data is lost**
   - The YAML library silently ignores unknown keys
   - We should use `yaml.UnmarshalStrict()` or at least log a warning for unknown keys
   - Alternatively: add a `DisallowedKeys` strict-mode option

### Process Issues

5. **`depguard` added to both projects without any settings**
   - depguard with zero config = allow nothing = break everything
   - The tool should either: (a) not enable depguard without settings, or (b) add sensible defaults
   - This is a linter-data quality issue in `pkg/constants/linter_data.go`

---

## F) TOP 25 THINGS WE SHOULD GET DONE NEXT

### Priority 1: Stop The Bleeding (P0 — Data Loss)

| #   | Task                                                          | Effort | Impact               |
| --- | ------------------------------------------------------------- | ------ | -------------------- |
| 1   | Add `LintersSettingsV1` field to `types.Config`               | S      | Prevents data loss   |
| 2   | Auto-migrate V1→V2 settings in `config/loader.go`             | S      | Fixes the root cause |
| 3   | Add roundtrip fidelity tests (load→save preserves all fields) | M      | Prevents regression  |
| 4   | Restore project-dependency-graph depguard allow-lists         | S      | Fixes broken project |
| 5   | Add depguard settings (allow $gostd, $module) to gogenfilter  | S      | Fixes broken project |

### Priority 2: Prevent Future Occurrences (P1)

| #   | Task                                                                      | Effort | Impact              |
| --- | ------------------------------------------------------------------------- | ------ | ------------------- |
| 6   | Warn when YAML has unknown top-level keys                                 | M      | Early detection     |
| 7   | Add depguard to "needs settings" list — don't auto-enable bare            | S      | Prevents footgun    |
| 8   | Add linter-data validation: flag linters that require settings            | M      | Systematic fix      |
| 9   | Add integration test: `configure` on real project with `linters-settings` | M      | End-to-end coverage |
| 10  | Consider using `yaml.Node` for roundtrip to preserve comments/formatting  | L      | Better UX           |

### Priority 3: Architecture Cleanup (P2)

| #   | Task                                                         | Effort | Impact                       |
| --- | ------------------------------------------------------------ | ------ | ---------------------------- |
| 11  | Unify `types.Config` and `migration.Config`                  | M      | Single source of truth       |
| 12  | Extract `LintersSettingsV1` migration to shared helper       | S      | Code reuse                   |
| 13  | Add merger roundtrip fidelity tests                          | S      | Confidence in merges         |
| 14  | Add `RecommendedLinterSettings` constant to `pkg/constants/` | M      | Single source of truth       |
| 15  | Type-safe settings structs (at least for critical linters)   | L      | IDE support, typo prevention |

### Priority 4: Test Coverage (P2)

| #   | Task                                                          | Effort | Impact                 |
| --- | ------------------------------------------------------------- | ------ | ---------------------- |
| 16  | Increase `internal/cli` coverage from 11% to 50%+             | M      | Critical path coverage |
| 17  | Add `pkg/client` tests (0% coverage)                          | S      | Client is untested     |
| 18  | Add `pkg/report` tests (0% coverage)                          | S      | Report gen is untested |
| 19  | Add config roundtrip tests for all formats (YAML, TOML, JSON) | M      | Format safety          |
| 20  | Add fuzz tests for config loading                             | M      | Edge case discovery    |

### Priority 5: Polish & DX (P3)

| #   | Task                                                                 | Effort | Impact                 |
| --- | -------------------------------------------------------------------- | ------ | ---------------------- |
| 21  | Fix LSP warnings in `internal/cli/cmd_configure.go` (14 warnings)    | S      | Clean IDE experience   |
| 22  | Fix deprecated `cobra.ExactValidArgs()` usage                        | S      | API hygiene            |
| 23  | Add `just watch` for auto-test on file change                        | S      | Developer productivity |
| 24  | Consider `github.com/goccy/go-yaml` for comment-preserving roundtrip | M      | Better UX              |
| 25  | Add schema validation against golangci-lint JSON schema              | L      | Full correctness       |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**Should depguard be auto-enabled at all?**

depguard with zero settings is a footgun — it defaults to "deny everything except stdlib".
This means enabling depguard without explicit allow-rules will break most projects immediately.

Three possible approaches:

1. **Don't auto-enable depguard** — Remove it from recommended linters, let users opt in
2. **Auto-enable with sensible defaults** — Always add `allow: [$gostd, $module]` when enabling depguard
3. **Conditional auto-enable** — Only enable depguard if the config already has `linters.settings.depguard`

**My recommendation:** Option 2 (auto-enable with sensible defaults) — it's the best balance of
safety and convenience. `$gostd` + `$module` covers 90% of projects, and users can tighten rules later.

But this is a product decision that depends on your philosophy for this tool.

---

## Codebase Health Metrics

| Metric                 | Value                    |
| ---------------------- | ------------------------ |
| Production code        | 7,345 lines              |
| Test code              | 3,513 lines              |
| Test ratio             | 47.8%                    |
| Packages with tests    | 10/12 (83%)              |
| Packages without tests | `client`, `report`       |
| Go version             | 1.26.0                   |
| `go vet`               | ✅ Clean                 |
| All tests              | ✅ Pass                  |
| Open bugs              | 1 critical (this report) |

---

_Generated by Crush on 2026-04-15T23:12:00+02:00_
