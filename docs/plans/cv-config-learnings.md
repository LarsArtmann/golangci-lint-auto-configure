# Comprehensive Implementation Plan: CV Config Learnings

> **Source:** `~/projects/CV/.golangci.yaml` analysis
> **Sorted by:** Impact → Customer Value → Effort (Pareto)

---

## Task Table (sorted by priority)

| #   | Task                                                                                              | Impact   | Effort | Category |
| --- | ------------------------------------------------------------------------------------------------- | -------- | ------ | -------- |
| 1   | Enrich `mnd` defaults: add `IgnoredFiles` field + test-file exclusion + more numbers              | High     | S      | Settings |
| 2   | Enrich `wrapcheck` defaults: add `IgnoreSigRegexps` field + stdlib regexps                        | High     | S      | Settings |
| 3   | Enrich `errcheck` defaults: add `CheckTypeAssertions` field + enable it                           | High     | XS     | Settings |
| 4   | Enrich `varnamelen` defaults: add `IgnoreDecls`, `MaxDistance`, `MinNameLength`                   | High     | S      | Settings |
| 5   | Add `gocognit` settings struct + curated default (min-complexity: 25)                             | Medium   | XS     | Settings |
| 6   | Add `gocyclo` settings struct + curated default (min-complexity: 20)                              | Medium   | XS     | Settings |
| 7   | Add `nestif` settings struct + curated default (min-complexity: 6)                                | Medium   | XS     | Settings |
| 8   | Add `goconst` settings struct + curated default (min-length: 4, min-occurrences: 5, ignore-tests) | Medium   | XS     | Settings |
| 9   | Move depguard from `DisabledLinters` to `NeverAutoEnableLinters`                                  | High     | M      | Policy   |
| 10  | Add depguard to `LinterPriorities` + `LinterReasons` (required by validator)                      | High     | XS     | Policy   |
| 11  | Update depguard test in `fixer_test.go` (no longer forcibly disabled)                             | High     | XS     | Policy   |
| 12  | Update depguard test in `fixer_enforce_test.go` (now never-auto-enable)                           | High     | XS     | Policy   |
| 13  | Add `checkAbsolutePathExclusions` health check                                                    | Medium   | S      | Health   |
| 14  | Add `checkDuplicateExclusionLinters` health check                                                 | Medium   | S      | Health   |
| 15  | Add `pruneUnenabledLinterSettings` to fixer                                                       | Medium   | S      | Fixer    |
| 16  | Add `tagalign` settings struct + curated default ordering                                         | Low      | S      | Settings |
| 17  | Run full test suite + validate linter data                                                        | Critical | XS     | Verify   |
| 18  | Update AGENTS.md with all changes                                                                 | Medium   | XS     | Docs     |

---

## Phase 1: Settings Enrichment (Tasks 1-4)

### Task 1: `mnd` — add test-file exclusion + more numbers (~10 min)

**Files:** `pkg/constants/linter_settings.go`

1. Add `IgnoredFiles []string` to `MndSettings` struct
2. Update `DefaultLinterSettings["mnd"]` with `IgnoredFiles: []string{"_test\\.go"}` and expand `IgnoredNumbers`
3. Verify build

### Task 2: `wrapcheck` — add sig regexps (~10 min)

**Files:** `pkg/constants/linter_settings.go`

1. Add `IgnoreSigRegexps []string` to `WrapcheckSettings` struct
2. Update defaults with curated stdlib regexps

### Task 3: `errcheck` — add type-assertion checks (~5 min)

**Files:** `pkg/constants/linter_settings.go`

1. Add `CheckTypeAssertions bool` to `ErrcheckSettings` struct
2. Set `CheckTypeAssertions: true` in defaults

### Task 4: `varnamelen` — add typed declarations (~10 min)

**Files:** `pkg/constants/linter_settings.go`

1. Add `IgnoreDecls []string`, `MaxDistance int`, `MinNameLength int` to `VarnamelenSettings`
2. Update defaults with typed declarations and thresholds

---

## Phase 2: Complexity Linter Defaults (Tasks 5-8)

### Task 5-8: Add 4 new settings structs (~5 min each)

**Files:** `pkg/constants/linter_settings.go`

For each of `gocognit`, `gocyclo`, `nestif`, `goconst`:

1. Add hand-maintained struct with yaml tags
2. Add `ToMap()` method
3. Add compile-time `SettingsConverter` check
4. Add entry in `DefaultLinterSettings`

---

## Phase 3: Depguard Policy Change (Tasks 9-12)

### Task 9: Move depguard between maps (~5 min)

**Files:** `pkg/constants/rules.go`

1. Remove `"depguard"` from `DisabledLinters`
2. Add to `NeverAutoEnableLinters` with architectural-enforcement reason

### Task 10: Add depguard priority + reason (~5 min)

**Files:** `pkg/constants/linter_priorities.go`, `pkg/constants/linter_reasons.go`

1. Add `"depguard": types.LinterPriorityMedium` to priorities
2. Add `"depguard": "..."` to reasons

### Task 11: Update fixer_test.go depguard test (~5 min)

**Files:** `pkg/linter/fixer_test.go`

1. Change test: depguard should NOT be forcibly moved to disable anymore
2. Verify depguard stays in enable when manually configured

### Task 12: Update fixer_enforce_test.go depguard test (~5 min)

**Files:** `pkg/linter/fixer_enforce_test.go`

1. Update test description from "forcibly disabled" to "never-auto-enable"

---

## Phase 4: New Health Checks (Tasks 13-14)

### Task 13: Absolute path detection (~10 min)

**Files:** `pkg/types/validation.go`

1. Add `RuleAbsolutePathExclusion` constant
2. Add `checkAbsolutePathExclusions()` method
3. Register in `CheckConfigHealthWithCriticalLinters`

### Task 14: Duplicate exclusion linter detection (~10 min)

**Files:** `pkg/types/validation.go`

1. Add `RuleDuplicateExclusionLinter` constant
2. Add `checkDuplicateExclusionLinters()` method
3. Register in `CheckConfigHealthWithCriticalLinters`

---

## Phase 5: Orphaned Settings Pruning (Task 15)

### Task 15: Prune settings for non-enabled linters (~10 min)

**Files:** `pkg/linter/fixer_config.go`

1. Add `pruneUnenabledLinterSettings()` function
2. Call from `updateConfigFromSets`

---

## Phase 6: Tagalign (Task 16)

### Task 16: Add tagalign default ordering (~10 min)

**Files:** `pkg/constants/linter_settings.go`

1. Add `TagalignSettings` struct
2. Add compile-time check
3. Add to `DefaultLinterSettings`

---

## Phase 7: Verification (Tasks 17-18)

### Task 17: Full test + validation (~10 min)

1. `go build ./...`
2. `go test -race ./pkg/... ./internal/...`
3. `go run ./scripts/validate_linter_data.go`

### Task 18: Update AGENTS.md (~10 min)

**Files:** `AGENTS.md`

1. Update gotcha #10 (depguard tier change)
2. Document new health checks
3. Document new settings defaults
