# Comprehensive Validation Analysis & Improvement Plan

**Date:** 2026-02-25 01:18\
**Session:** Multi-project validation analysis (147 projects scanned)\
**Status:** Analysis Complete - Action Plan Ready

---

## Executive Summary

Conducted comprehensive validation across **147 projects** using `golangci-lint-auto-configure validate` command. Identified **10 projects with validation failures (6.8% failure rate)** across **4 distinct error categories**. The analysis reveals that most failures stem from v1→v2 migration artifacts, deprecated settings, and type mismatches - all of which are **automatically fixable**.

**Key Insight:** The existing `migrate` command handles v1→v2 schema migration, but there's a gap for "v2 configs with deprecated/invalid settings" - configs that are technically v2 but contain invalid options. This calls for a new `repair` or `fix` command.

---

## Validation Results Summary

| Metric                  | Count        |
| ----------------------- | ------------ |
| Total Projects Scanned  | 147          |
| Valid Configurations    | 137 (93.2%)  |
| Failed Configurations   | 10 (6.8%)    |
| Distinct Error Patterns | 4            |
| Auto-Fixable Issues     | 10/10 (100%) |

---

## Detailed Failure Analysis

### Category 1: Deprecated/Invalid Linter Settings (4 projects)

**Projects:** `complaints-mcp`, `desire-secrets`, `ast-state-analyzer`, `cmdguard`

**Common Issues:**

| Setting                                   | Error                                                      | Fix                             |
| ----------------------------------------- | ---------------------------------------------------------- | ------------------------------- |
| `tagliatelle.use-field-name`              | `additionalProperties 'use-field-name' not allowed`        | Remove - property removed in v2 |
| `cyclop.skip-tests`                       | `additionalProperties 'skip-tests' not allowed`            | Remove - property removed       |
| `sloglint.key-naming-case: "UPPER_SNAKE"` | `value must be one of 'snake', 'kebab', 'camel', 'pascal'` | Change to valid enum value      |
| `modernize.disable: ["loop"]`             | `value must be one of [...]`                               | Use valid enum values           |
| `gochecksumtype.exhaustive`               | `additionalProperties 'exhaustive' not allowed`            | Remove                          |
| `fatcontext.ignore-len`                   | `additionalProperties 'ignore-len' not allowed`            | Remove                          |
| `gocritic.settings.appendAssign`          | `additionalProperties not allowed`                         | Remove nested settings          |
| `wrapcheck.ignoreSigRegexps`              | `additionalProperties not allowed`                         | Use correct property name       |
| `forbidigo.forbid[].p`                    | `additionalProperties 'p' not allowed`                     | Use `pattern:` not `p:`         |
| `gosec.excludes: null`                    | `got null, want array`                                     | Change to `[]` or valid array   |

**Root Cause:** Projects migrated to v2 schema but kept deprecated linter settings from v1.

---

### Category 2: Missing Version Field (3 projects)

**Projects:** `KeyCountdown`, `SwettySwipperWeb`, `go-structure-linter`

**Issues:**

- `KeyCountdown` & `SwettySwipperWeb`: Missing `version:` field entirely (v1 config structure)
- `go-structure-linter`: Has `version: "v1.26"` (explicit v1 version)

**Root Cause:** Never migrated from v1 to v2 schema.

**Fix:** Run existing `migrate` command which:

1. Creates backup
2. Runs `golangci-lint migrate`
3. Validates migrated config
4. Shows migration diff

---

### Category 3: Type Mismatch - string vs array (2 projects)

**Projects:** `StopTube`, `cmdguard`

**Error:**

```yaml
formatters:
  settings:
    goimports:
      local-prefixes: "github.com/myorg" # ERROR: got string, want array
```

**Fix:**

```yaml
formatters:
  settings:
    goimports:
      local-prefixes:
        - "github.com/myorg" # Array, not string
```

**Root Cause:** Schema change in v2 - `local-prefixes` changed from string to array.

---

### Category 4: Missing Required Field (1 project)

**Project:** `file-and-image-renamer`

**Error:** `run.timeout cannot be empty`

**Root Cause:** Required field `run.timeout` is empty string.

**Fix:** Set valid duration: `timeout: "5m"`

---

## Architecture Analysis

### Existing Codebase Review

#### 1. Migrate Command (`internal/cli/cmd/migrate.go`)

**Status:** ✅ Implemented and working

**Functionality:**

- Detects v1 configs (via `version` field check)
- Creates `.v1-backup` before migration
- Runs `golangci-lint migrate` command
- Restores backup on failure
- Shows migration diff (version, linters, timeout changes)

**Gap Identified:** Only handles v1→v2 schema migration, not "v2 configs with invalid settings".

---

#### 2. Validation System (`pkg/types/types.go`)

**Status:** ✅ Well-structured

**Key Types:**

- `ValidationError` - Structured error with Field, Message, Line
- `ValidationResult` - Valid flag + errors array
- `ConfigLoader.ValidateConfig()` - Interface method

**Current Validation (`pkg/config/loader.go:198-213`):**

```go
func (l *Loader) ValidateConfig(config *Config) []error {
    var errs []error
    if config.Run.Timeout == "" {
        errs = append(errs, errors.NewConfigError("run.timeout cannot be empty", "", nil))
    }
    // ... basic checks only
    return errs
}
```

**Gap:** No schema validation, no deprecated setting detection, no type checking.

---

#### 3. Linter Data (`pkg/constants/linter_data.go`)

**Status:** ✅ Comprehensive

**Contains:**

- `LinterPriorities` - Maps linters to priority levels
- `LinterReasons` - Human-readable descriptions
- `DeprecatedLinters` - Maps deprecated → replacement
- `FormatterInfo` - Formatter metadata

**Gap:** No "Deprecated Settings" database for linter configuration options.

---

### Type Model Improvements Needed

#### Current Architecture

```go
// LintersConfig uses map[string]any for settings - loses type safety
type LintersConfig struct {
    Enable   []string       `yaml:"enable,omitempty"`
    Disable  []string       `yaml:"disable,omitempty"`
    Settings map[string]any `yaml:"settings,omitempty"`  // ⚠️ Untyped
}
```

**Problems:**

1. `Settings map[string]any` - No compile-time type safety
2. No validation at parse time
3. No schema enforcement
4. Can't detect deprecated settings

---

#### Proposed Architecture

**Option A: Add Validation Tags (Recommended for quick wins)**

Use `github.com/go-playground/validator/v10` for struct validation:

```go
type LintersSettings struct {
    Cyclop    *CyclopSettings    `yaml:"cyclop,omitempty" validate:"omitempty"`
    Gosec     *GosecSettings     `yaml:"gosec,omitempty" validate:"omitempty"`
    Forbidigo *ForbidigoSettings `yaml:"forbidigo,omitempty" validate:"omitempty"`
}

type CyclopSettings struct {
    MaxComplexity int  `yaml:"max-complexity" validate:"min=1,max=100"`
    // SkipTests removed - doesn't exist in v2
}

type GosecSettings struct {
    Excludes []string `yaml:"excludes" validate:"omitempty,dive,required"`
    // Cannot be null - must be array or omitted
}
```

**Pros:**

- Compile-time type safety
- Built-in validation
- Clear error messages
- Easy to maintain

**Cons:**

- Requires defining all linter settings structs
- More verbose

---

**Option B: JSON Schema Validation**

Use `github.com/santhosh-tekuri/jsonschema` to validate against golangci-lint's schema:

```go
// Load golangci-lint's JSON schema
schema, err := jsonschema.Compile("golangci-lint-schema.json")

// Validate config against schema
var config map[string]any
yaml.Unmarshal(data, &config)
err = schema.Validate(config)
```

**Pros:**

- Validates exact golangci-lint schema
- Catches all type mismatches
- No need to maintain parallel structs

**Cons:**

- Requires golangci-lint to expose schema (not currently available)
- Error messages less user-friendly
- Harder to customize

---

**Option C: Hybrid Approach (Recommended for long-term)**

Combine both approaches:

1. Use typed structs with validation tags for common settings
2. Use JSON schema validation for full coverage
3. Build "deprecated settings" database for auto-fix

---

## Library Research

### Available Dependencies

| Library                    | Status              | Use Case                                    |
| -------------------------- | ------------------- | ------------------------------------------- |
| `github.com/goccy/go-yaml` | ✅ Already imported | YAML manipulation with comment preservation |
| `gopkg.in/yaml.v3`         | ✅ Already imported | Standard YAML parsing                       |
| `github.com/tidwall/gjson` | ✅ Already imported | JSON path queries                           |
| `github.com/tidwall/sjson` | ✅ Already imported | JSON modification                           |

### Recommended Additions

| Library                                  | Purpose                             | Priority |
| ---------------------------------------- | ----------------------------------- | -------- |
| `github.com/go-playground/validator/v10` | Struct validation with tags         | **P1**   |
| `github.com/santhosh-tekuri/jsonschema`  | JSON Schema validation              | P2       |
| `github.com/hashicorp/go-version`        | Semantic version parsing/comparison | P3       |

---

## Multi-Step Execution Plan

### Phase 1: Quick Wins (High Impact, Low Effort)

| #   | Task                                        | Effort | Impact | Status |
| --- | ------------------------------------------- | ------ | ------ | ------ |
| 1.1 | Add validation for empty `run.timeout`      | 10 min | High   | ⬜     |
| 1.2 | Add `local-prefixes` type fixer             | 15 min | High   | ⬜     |
| 1.3 | Create "Deprecated Settings" database       | 30 min | High   | ⬜     |
| 1.4 | Enhance error messages with fix suggestions | 20 min | Medium | ⬜     |
| 1.5 | Add pre-validation before schema check      | 15 min | Medium | ⬜     |

**Phase 1 Total:** ~90 minutes, fixes 80% of issues

---

### Phase 2: Repair Command (High Impact, Medium Effort)

| #   | Task                                  | Effort | Impact | Status |
| --- | ------------------------------------- | ------ | ------ | ------ |
| 2.1 | Design `repair` command interface     | 30 min | High   | ⬜     |
| 2.2 | Implement `pkg/repair/repair.go`      | 60 min | High   | ⬜     |
| 2.3 | Add deprecated settings auto-removal  | 45 min | High   | ⬜     |
| 2.4 | Add type mismatch auto-fix            | 30 min | High   | ⬜     |
| 2.5 | Add `local-prefixes` string→array fix | 20 min | High   | ⬜     |
| 2.6 | Wire into CLI commands                | 30 min | Medium | ⬜     |
| 2.7 | Write BDD tests                       | 60 min | Medium | ⬜     |

**Phase 2 Total:** ~4.5 hours, creates new repair capability

---

### Phase 3: Type Architecture (Medium Impact, High Effort)

| #   | Task                           | Effort | Impact | Status |
| --- | ------------------------------ | ------ | ------ | ------ |
| 3.1 | Add `validator/v10` dependency | 10 min | Medium | ⬜     |
| 3.2 | Define linter settings structs | 90 min | Medium | ⬜     |
| 3.3 | Add validation tags to types   | 60 min | Medium | ⬜     |
| 3.4 | Create validation middleware   | 45 min | Medium | ⬜     |
| 3.5 | Update loader with validation  | 30 min | Medium | ⬜     |
| 3.6 | Add tests for validation       | 60 min | Low    | ⬜     |

**Phase 3 Total:** ~5 hours, improves type safety

---

### Phase 4: Advanced Features (Low Impact, High Effort)

| #   | Task                                     | Effort  | Impact | Status |
| --- | ---------------------------------------- | ------- | ------ | ------ |
| 4.1 | Add `goccy/go-yaml` comment preservation | 90 min  | Low    | ⬜     |
| 4.2 | Create interactive repair mode           | 120 min | Low    | ⬜     |
| 4.3 | Add repair dry-run with diff             | 60 min  | Low    | ⬜     |
| 4.4 | Implement batch repair for all projects  | 90 min  | Low    | ⬜     |
| 4.5 | Add repair report generation             | 60 min  | Low    | ⬜     |

**Phase 4 Total:** ~7 hours, nice-to-have features

---

## Prioritized Action Items

### Immediate (Do First)

1. **Add `repair` command** - Address the gap between `migrate` (v1→v2) and `validate` (schema check)
2. **Create deprecated settings database** - Map removed/renamed settings for auto-fix
3. **Fix `local-prefixes` type** - String→array auto-conversion
4. **Enhance error messages** - Parse jsonschema errors, suggest fixes

### Short-term (This Week)

5. **Add struct validation** - Use `validator/v10` tags for type safety
6. **Pre-validation checks** - Catch empty timeout, type errors before schema validation
7. **Repair dry-run mode** - Show what would change without modifying

### Long-term (Next Sprint)

8. **Full linter settings structs** - Replace `map[string]any` with typed structs
9. **Comment preservation** - Use `goccy/go-yaml` to preserve YAML comments during repair
10. **Batch repair** - Run repair across all projects automatically

---

## Code Reuse Opportunities

### Existing Code That Fits

| Component                    | Location                      | Reusability                                 |
| ---------------------------- | ----------------------------- | ------------------------------------------- |
| `CreateBackup/RestoreConfig` | `pkg/config/loader.go`        | ✅ High - Repair needs backup/restore       |
| `ShowMigrationChanges`       | `internal/cli/cmd/migrate.go` | ✅ High - Repair needs diff view            |
| `ValidationError`            | `pkg/types/types.go`          | ✅ High - Repair errors use same format     |
| `FixConfig`                  | `pkg/linter/fixer.go`         | ⚠️ Medium - Similar pattern, different scope |
| `Diff`                       | `pkg/diff/differ.go`          | ✅ High - Show repair changes               |

### Pattern to Follow

The `migrate` command (`internal/cli/cmd/migrate.go`) provides an excellent template:

1. Create backup
2. Load config
3. Apply transformations
4. Save config
5. Show diff
6. Restore on failure

---

## Risk Assessment

| Risk                                   | Probability | Impact | Mitigation                              |
| -------------------------------------- | ----------- | ------ | --------------------------------------- |
| Breaking existing configs              | Low         | High   | Always backup before repair             |
| False positives in repair              | Medium      | Medium | Dry-run mode, user confirmation         |
| Maintenance burden of settings structs | Medium      | Low    | Auto-generate from golangci-lint source |
| golangci-lint schema changes           | High        | Medium | Version-aware repair, update database   |

---

## Success Metrics

- **Target:** Reduce validation failure rate from 6.8% to <1%
- **Coverage:** Handle 100% of identified error patterns
- **User Experience:** Clear error messages with actionable fixes
- **Safety:** Zero data loss (backup + restore on failure)

---

## Recommendations

### For Users (Immediate)

```bash
# For v1 configs - use migrate
golangci-lint-auto-configure migrate

# For v2 configs with errors - use repair (after implementation)
golangci-lint-auto-configure repair --dry-run  # Preview
golangci-lint-auto-configure repair            # Apply fixes
```

### For Developers

1. **Start with Phase 1** - Quick wins with immediate impact
2. **Reuse migrate patterns** - Copy backup/restore/diff flow
3. **Add validator library** - Foundation for type safety
4. **Document deprecated settings** - Build database incrementally

---

## Files Modified/Analyzed

| File                                                                | Action  | Notes                        |
| ------------------------------------------------------------------- | ------- | ---------------------------- |
| `docs/validation-analysis-2026-02-25.md`                            | Created | Raw analysis of 147 projects |
| `docs/status/2026-02-25_01-18_COMPREHENSIVE_VALIDATION_ANALYSIS.md` | Created | This report                  |

---

## Next Session Priorities

1. ✅ **Analysis Complete** - Understanding validation failures
2. ⬜ **Create repair command** - Core implementation (Phase 2)
3. ⬜ **Add deprecated settings DB** - Enable auto-fix (Phase 1.3)
4. ⬜ **Fix local-prefixes type** - Most common type error (Phase 1.2)
5. ⬜ **Enhance error messages** - Better UX (Phase 1.4)

---

_Generated with Crush - Assisted by Kimi K2.5_
