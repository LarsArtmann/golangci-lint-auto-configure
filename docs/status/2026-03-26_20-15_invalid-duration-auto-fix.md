# Status Report: Invalid Duration Auto-Fix Feature

**Date:** 2026-03-26 20:15
**Author:** AI Assistant
**Status:** Feature Implemented, Tests Passing

---

## Summary

Implemented auto-fix functionality for invalid `run.timeout` duration fields in golangci-lint configurations. The tool now detects and fixes empty or malformed duration strings (e.g., `""`, `"invalid"`) before running analysis.

---

## Work Completed

### ✅ Fully Done

| Item                | Description                                                               | Files Changed                   |
| ------------------- | ------------------------------------------------------------------------- | ------------------------------- |
| Duration validation | Added `preFixInvalidDurations()` to detect invalid timeout values         | `pkg/linter/fixer_preflight.go` |
| Dry-run support     | Returns `(needsFixing bool, err)` to skip analysis in dry-run mode        | `pkg/linter/fixer.go`           |
| Result calculation  | Added `calculateDryRunResultWithInvalidDurations()` for dry-run reporting | `pkg/linter/fixer_preflight.go` |
| Test coverage       | 4 new tests for duration validation (empty, invalid, valid, dry-run)      | `pkg/linter/fixer_test.go`      |
| Integration         | Wired into `FixConfigResult()` before version check                       | `pkg/linter/fixer.go`           |

### ⚠️ Partially Done

| Item               | Status                                                      | Next Step                   |
| ------------------ | ----------------------------------------------------------- | --------------------------- |
| Clean up dead code | Identified `FixerPreflight` struct as unused                | Remove it                   |
| Untracked files    | `pkg/formatters/`, `pkg/migration/` exist but not committed | Review and commit or delete |

### ❌ Not Started

| Item                              | Priority | Effort |
| --------------------------------- | -------- | ------ |
| Centralized duration utility      | Medium   | 15min  |
| Custom validator for duration     | Low      | 30min  |
| Preflight check interface pattern | Low      | 1hr    |

---

## Architecture Observations

### Current State

```
pkg/linter/
├── fixer.go              # Main FixConfigResult logic
├── fixer_preflight.go    # Contains:
│                         #   - preFixVersion()
│                         #   - preFixInvalidDurations() [NEW]
│                         #   - preFixDeprecatedLinters()
│                         #   - calculateDryRunResultWithDeprecated()
│                         #   - calculateDryRunResultWithInvalidDurations() [NEW]
│                         #   - FixerPreflight struct [DEAD CODE]
│                         #   - NewFixerPreflight() [DEAD CODE]
│                         #   - EnsureVersion() [DEAD CODE - duplicate]
│                         #   - RemoveDeprecatedLinters() [DEAD CODE - duplicate]
└── fixer_test.go         # Tests
```

### Problems Identified

1. **Dead Code**: `FixerPreflight` struct and its methods are never called
2. **Inconsistent Dry-Run Handling**: `preFixVersion` doesn't save in dry-run, but old `preFixInvalidDurations` did
3. **No Duration Validation on Load**: Config loads invalid durations without error
4. **String Type for Duration**: `RunConfig.Timeout` is `string`, not `time.Duration`

---

## What I Forgot / Could Have Done Better

### 1. Dead Code Cleanup

- **Forgot**: The `FixerPreflight` struct at lines 204-312 is never used
- **Impact**: Confusing for future developers, increases maintenance burden
- **Fix**: Remove `FixerPreflight`, `NewFixerPreflight`, `EnsureVersion`, `RemoveDeprecatedLinters` from `fixer_preflight.go`

### 2. Consistent Preflight Pattern

- **Issue**: `preFixVersion` returns `error`, `preFixInvalidDurations` returns `(bool, error)`
- **Better**: Both should return `(needsFixing bool, err error)` for consistency
- **Impact**: More complex but enables skipping analysis for version issues too

### 3. Validation at Load Time

- **Forgot**: Could validate duration format when loading config
- **Better**: Add custom YAML unmarshaler or validator tag
- **Impact**: Catches errors earlier, better UX

### 4. Untracked Files

- **Issue**: `pkg/formatters/` and `pkg/migration/` exist but not in git
- **Action**: Review and either commit or delete

---

## Top 25 Next Steps (Sorted by Impact/Effort)

### High Impact, Low Effort (Do Now)

| #   | Task                               | Effort | Impact | Why                        |
| --- | ---------------------------------- | ------ | ------ | -------------------------- |
| 1   | Remove dead `FixerPreflight` code  | 5min   | High   | Clean code, less confusion |
| 2   | Test E2E with broken config        | 5min   | High   | Verify feature works       |
| 3   | Commit current changes             | 5min   | High   | Don't lose work            |
| 4   | Review untracked `pkg/migration/`  | 10min  | Medium | May be needed feature      |
| 5   | Review untracked `pkg/formatters/` | 10min  | Medium | May be needed feature      |

### High Impact, Medium Effort (Do Soon)

| #   | Task                                               | Effort | Impact | Why                       |
| --- | -------------------------------------------------- | ------ | ------ | ------------------------- |
| 6   | Add duration validation tag to `RunConfig.Timeout` | 15min  | High   | Catch errors at load time |
| 7   | Create `IsValidDuration()` utility in `pkg/types`  | 15min  | Medium | Reusable validation       |
| 8   | Make `preFixVersion` return `(bool, error)`        | 20min  | Medium | Consistent pattern        |
| 9   | Add duration validation to `ValidateConfig()`      | 20min  | High   | Comprehensive validation  |
| 10  | Document preflight check pattern in AGENTS.md      | 15min  | Medium | Knowledge transfer        |

### Medium Impact, Medium Effort (Nice to Have)

| #   | Task                                      | Effort | Impact | Why                     |
| --- | ----------------------------------------- | ------ | ------ | ----------------------- |
| 11  | Create `PreflightCheck` interface         | 30min  | Medium | Extensible architecture |
| 12  | Add preflight check registry              | 30min  | Medium | Plugin pattern          |
| 13  | Add more duration fields (if any)         | 15min  | Low    | Complete coverage       |
| 14  | Add integration test for invalid duration | 20min  | Medium | E2E coverage            |
| 15  | Update CLI help text for configure        | 10min  | Low    | Better UX               |

### Lower Priority (Future)

| #   | Task                                       | Effort | Impact | Why               |
| --- | ------------------------------------------ | ------ | ------ | ----------------- |
| 16  | Use `time.Duration` type instead of string | 1hr    | Medium | Type safety       |
| 17  | Custom YAML unmarshaler for durations      | 45min  | Medium | Parse on load     |
| 18  | Add `--skip-preflight` flag                | 30min  | Low    | Advanced control  |
| 19  | Add preflight summary to output            | 20min  | Low    | Better reporting  |
| 20  | Add preflight metrics                      | 30min  | Low    | Observability     |
| 21  | Refactor to functional options             | 1hr    | Low    | Modern pattern    |
| 22  | Add preflight check tests                  | 30min  | Medium | Coverage          |
| 23  | Document in README                         | 15min  | Low    | User docs         |
| 24  | Add to examples/                           | 10min  | Low    | Examples          |
| 25  | Create ADR for preflight pattern           | 30min  | Low    | Architecture docs |

---

## Type Model Improvements

### Current

```go
type RunConfig struct {
    Timeout string `yaml:"timeout" validate:"required"`
    // ...
}
```

### Option 1: Custom Validator (Recommended)

```go
type RunConfig struct {
    Timeout string `yaml:"timeout" validate:"required,duration"`
    // ...
}

// In validation.go
func validateDuration(fl validator.FieldLevel) bool {
    _, err := time.ParseDuration(fl.Field().String())
    return err == nil
}
```

### Option 2: Dedicated Type (More Complex)

```go
type Duration string

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
    var s string
    if err := value.Decode(&s); err != nil {
        return err
    }
    if _, err := time.ParseDuration(s); err != nil {
        return fmt.Errorf("invalid duration %q: %w", s, err)
    }
    *d = Duration(s)
    return nil
}

type RunConfig struct {
    Timeout Duration `yaml:"timeout" validate:"required"`
}
```

**Recommendation**: Option 1 (custom validator) - simpler, uses existing validation framework.

---

## Libraries to Consider

| Library                              | Use Case                                    | Verdict                   |
| ------------------------------------ | ------------------------------------------- | ------------------------- |
| `go-playground/validator`            | Already used, add custom duration validator | ✅ Use                    |
| `time.ParseDuration`                 | Standard library, already using             | ✅ Use                    |
| `github.com/invopop/yaml`            | Better YAML unmarshaling with validation    | ❌ Overkill               |
| `github.com/go-ozzo/ozzo-validation` | Alternative validation                      | ❌ Already have validator |

---

## Test Results

```
✅ pkg/linter - 21 specs, all passing
✅ pkg/config - 20 specs, all passing
✅ pkg/detection - 7 specs, all passing
✅ pkg/diff - 10 specs, all passing
✅ internal/cli - 19 specs, all passing
✅ pkg/formatters - 12 specs, all passing
✅ pkg/styled_output - 13 specs, all passing

Composite coverage: 57.0%
```

---

## Files Changed

| File                            | Changes                                               |
| ------------------------------- | ----------------------------------------------------- |
| `pkg/linter/fixer.go`           | +8 lines (call preFixInvalidDurations, handle result) |
| `pkg/linter/fixer_preflight.go` | +42 lines (new function, dry-run result)              |
| `pkg/linter/fixer_test.go`      | +79 lines (4 new tests)                               |
| `internal/cli/cmd/migrate.go`   | Minor refactor (unrelated)                            |

---

## Top #1 Question I Cannot Answer

**Question**: What is the intended purpose of `pkg/migration/` and `pkg/formatters/` directories? They exist as untracked files but I don't know if they're work-in-progress features that should be committed or experimental code that should be deleted.

**Context**:

- `pkg/migration/validator.go` has compiler errors (undefined `Migrator`)
- Both directories are untracked in git
- No documentation about their purpose

**Action Needed**: User input on whether to:

1. Fix and commit these directories
2. Delete them
3. Add to `.gitignore` as work-in-progress

---

## Next Immediate Actions

1. ✅ All tests passing
2. 🔄 Commit current changes
3. ⏳ Remove dead `FixerPreflight` code
4. ⏳ Review untracked directories
5. ⏳ Test E2E with broken config
6. ⏳ Push to remote
