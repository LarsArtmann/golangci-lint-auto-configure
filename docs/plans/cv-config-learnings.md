# Improvement Plan: Learnings from CV `.golangci.yaml`

> Source: `~/projects/CV/.golangci.yaml` — a large, manually-tuned config across ~100 linters with
> architectural depguard rules, layered architecture, and battle-tested thresholds.

---

## Executive Summary

The CV config reveals **11 actionable improvements** across 4 categories:

1. **Settings enrichment** — our curated defaults are missing high-value fields and entire linter defaults
2. **Config health checks** — structural problems we don't detect yet
3. **Depguard policy** — blanket disable breaks architectural enforcement (decided: move to NeverAutoEnable)
4. **Pruning gap** — settings for non-enabled, non-disabled linters survive forever

Each item below has: **what**, **why**, **how**, **files**, and **effort**.

---

## Tier 1: High-Impact Friction Reduction

These reduce `nolint` noise across all projects when the tool auto-enables these linters.

### 1.1 Enrich `mnd` defaults: exclude test files

**What:** Add `IgnoredFiles: ["_test.go"]` to `MndSettings` and update `DefaultLinterSettings`.

**Why:** `mnd` (magic number detector) fires constantly in test code — `assert.Equal(t, 42, result)`,
table-driven test cases with literal values, benchmark constants. The CV config excludes `_test.go`
and `_fuzz_test.go`. Our current default only ignores numbers `0, 1, 2, 100`, leaving test files
as a major friction source.

**How:**
1. Add `IgnoredFiles []string` field to hand-maintained `MndSettings` struct
2. Set `IgnoredFiles: []string{"_test\\.go"}` in `DefaultLinterSettings`
3. Also add ignored numbers `3, 4, 5, 10, 1024` (common in code: powers of 2, small indices, percentages)
4. Add BDD specs verifying the new defaults
5. Run `templ generate` if needed, then `go test`

**Files:**
- `pkg/constants/linter_settings.go` — struct + defaults
- `pkg/constants/linter_settings_internal_test.go` — specs
- `pkg/linter/fixer_test.go` — integration specs

**Effort:** S (1-2 hours)

---

### 1.2 Enrich `wrapcheck` defaults: add sig regexps for stdlib

**What:** Add `IgnoreSigRegexps` field to `WrapcheckSettings` and populate with stdlib patterns.

**Why:** Our current `IgnoreSigs` list has 9 signatures (`.Errorf(`, `errors.New(`, etc.). The CV
config demonstrates that `ignore-sig-regexps` is far more powerful — it covers entire packages with
one entry: `slices\.`, `maps\.`, `sort\.`, `time\.`, `os\.`, `io\.`, `strings\.`, `strconv\.`,
`path\.`, `filepath\.`, `compress/gzip`, `net/http`, `go.opentelemetry.io/otel`. Without these,
`wrapcheck` fires on every stdlib call that returns an error, creating massive friction.

**How:**
1. Add `IgnoreSigRegexps []string` field to hand-maintained `WrapcheckSettings`
2. Set curated regexps in `DefaultLinterSettings`:
   ```
   encoding/json
   fmt\.
   errors\.
   slices\.
   maps\.
   sort\.
   time\.
   net/http
   os\.
   io\.
   strings\.
   strconv\.
   path\.
   filepath\.
   compress/gzip
   \(context\.Context\)\.Err
   ```
3. Note: keep existing `IgnoreSigs` as-is (they work alongside regexps)
4. Add BDD specs

**Files:**
- `pkg/constants/linter_settings.go` — struct + defaults
- `pkg/constants/linter_settings_internal_test.go` — specs

**Effort:** S (1 hour)

---

### 1.3 Enrich `errcheck` defaults: enable type-assertion checks

**What:** Add `CheckTypeAssertions: true` to `ErrcheckSettings`.

**Why:** Unchecked type assertions are a common source of panics. The CV config enables this.
Our current default only curates `ExcludeFunctions`. Adding `check-type-assertions: true` catches
`x.(T)` without `ok` check — a real safety improvement with low friction.

**How:**
1. Add `CheckTypeAssertions bool` field to hand-maintained `ErrcheckSettings`
2. Set `CheckTypeAssertions: true` in `DefaultLinterSettings`
3. Add BDD specs

**Files:**
- `pkg/constants/linter_settings.go`
- `pkg/constants/linter_settings_internal_test.go`

**Effort:** XS (30 min)

---

### 1.4 Add curated defaults for complexity linters

**What:** Add `DefaultLinterSettings` entries for `gocognit`, `gocyclo`, `nestif`, `goconst`.

**Why:** These linters are in the "High" priority tier and get auto-enabled, but they have **no
curated defaults** — meaning they run with upstream defaults that are often too strict or too loose.
The CV config demonstrates sensible thresholds calibrated for real projects:

| Linter | Upstream default | CV value | Proposed default | Rationale |
|--------|-----------------|----------|-----------------|-----------|
| `gocognit` | 20 | 25 | **25** | Accounts for necessary domain complexity |
| `gocyclo` | 30 | 20 | **20** | Tighter than upstream, catches real complexity |
| `nestif` | 5 | 6 | **6** | Allows one extra level for guard clauses |
| `goconst` | min-length 3, min-occurrences 3 | min-length 4, min-occurrences 6, ignore-tests | **min-length 4, min-occurrences 5, ignore-tests: true** | Reduces noise from short constants and test repetitions |

**How:**
1. Add hand-maintained structs: `GocognitSettings`, `GocycloSettings`, `NestifSettings`, `GoconstSettings`
2. Register in `DefaultLinterSettings`
3. Add compile-time `SettingsConverter` compliance checks
4. Add BDD specs

**Files:**
- `pkg/constants/linter_settings.go` — 4 new structs + defaults
- `pkg/constants/linter_settings_internal_test.go` — specs

**Effort:** M (2-3 hours, 4 structs + tests)

---

## Tier 2: Config Health Checks

New health checks that detect structural problems in user configs.

### 2.1 Detect absolute paths in exclusion paths

**What:** Add a health check (Warning severity) when `linters.exclusions.paths` or
`formatters.exclusions.paths` contains absolute paths.

**Why:** The CV config has `/home/lars/projects/CV/internal/database/` baked in — this leaks the
developer's home directory, breaks portability across machines/CI, and won't match on other
developers' machines. Absolute paths in exclusion rules are almost always a mistake.

**How:**
1. Add new rule constant `RuleAbsolutePathExclusion = "absolute-path-exclusion"`
2. Add `checkAbsolutePathExclusions(cfg)` method to `ConfigHealth`
3. Check each path in `linters.exclusions.paths` and `formatters.exclusions.paths`
4. Flag if path starts with `/` (Unix) or matches `^[A-Z]:\\` (Windows)
5. Suggestion: "Use a relative path or glob pattern instead"
6. Register in `CheckConfigHealthWithCriticalLinters`
7. Add BDD specs

**Files:**
- `pkg/types/validation.go` — new check + rule constant
- `pkg/types/validation_test.go` — specs

**Effort:** S (1 hour)

---

### 2.2 Detect duplicate entries in exclusion rule linter lists

**What:** Add a health check (Warning severity) when a single exclusion rule has duplicate linter
entries in its `linters` list.

**Why:** The CV config has `exhaustruct` listed **3 times** in the same `_test.go` exclusion rule
(lines 401, 424, 425). This is a copy-paste error. While harmless at runtime (golangci-lint dedups
internally), it signals config neglect and confuses readers.

**How:**
1. Add new rule constant `RuleDuplicateExclusionLinter = "duplicate-exclusion-linter"`
2. Add `checkDuplicateExclusionLinters(cfg)` method to `ConfigHealth`
3. Iterate `cfg.Linters.Exclusions.Rules`, for each rule's `Linters` list, detect duplicates
4. Suggestion: "Remove duplicate entries from the exclusion rule"
5. Register in `CheckConfigHealthWithCriticalLinters`
6. Add BDD specs

**Files:**
- `pkg/types/validation.go`
- `pkg/types/validation_test.go`

**Effort:** S (1 hour)

---

### 2.3 Prune orphaned settings for non-enabled linters

**What:** Extend the pruning logic to remove settings blocks for linters that are neither in
`enable` nor in `disable` — they're simply absent.

**Why:** Currently `pruneDisabledLinterSettings` only removes settings for linters in the `disable`
list. But with `default: none` (which our tool injects), linters not in `enable` don't run — their
settings are dead weight. The CV config has `lll` settings (`line-length: 140`) but `lll` is not in
`enable` or `disable`. These settings persist forever, confusing readers and bloating the config.

**How:**
1. Add new function `pruneUnenabledLinterSettings(cfg, enabledLinters, logger)` in `fixer_config.go`
2. Logic: for each key in `cfg.Linters.Settings`, if the key is not in `enabledLinters` and not in
   `DisabledLinters`, delete it
3. Call from `updateConfigFromSets` after `pruneDisabledLinterSettings`
4. Count as a config change in the recorder
5. Add BDD specs for: settings for enabled linter preserved, settings for disabled linter pruned
   (existing), settings for absent linter pruned (new)
6. Edge case: don't prune settings for linters in `DisabledLinters` that are also in
   `disable` list (already handled by `pruneDisabledLinterSettings`)

**Files:**
- `pkg/linter/fixer_config.go` — new prune function
- `pkg/linter/fixer_test.go` — specs

**Effort:** S (1-2 hours)

---

## Tier 3: Depguard Policy Change

### 3.1 Move depguard from `DisabledLinters` to `NeverAutoEnableLinters`

**What:** Remove `depguard` from `DisabledLinters` and add it to `NeverAutoEnableLinters`.

**Why:** The blanket disable breaks projects using depguard for architectural enforcement
(layer dependency rules, feature isolation, service-layer import constraints). `library-policy`
cannot replicate file-pattern-based rules. The CV project demonstrates this use case with 7 depguard
rules enforcing a 3-layer architecture. Moving to `NeverAutoEnable` means:
- Never auto-recommended (no behavioral change for configure runs)
- Respected when manually configured (settings preserved, not pruned)
- Not forcibly moved to disable list
- Can still be disabled via sidecar if desired

**How:**
1. Remove `"depguard"` entry from `DisabledLinters` in `pkg/constants/rules.go`
2. Add to `NeverAutoEnableLinters`:
   ```go
   "depguard": "never auto-enabled; use library-policy for banned-library governance, but respect manual configuration for architectural enforcement (layer rules, feature isolation) via file-pattern rules",
   ```
3. **Must** add entries to `LinterPriorities` and `LinterReasons` (required by data integrity tests
   for NeverAutoEnable linters)
4. Remove depguard exemption from `isToolLevelManaged` in `fixer_enforce.go` — depguard is now
   subject to normal NeverAutoEnable treatment (which is exempt from enforcement but NOT forcibly
   disabled)
   - Actually: `isToolLevelManaged` checks both `DisabledLinters` and `NeverAutoEnableLinters`, so
     depguard stays exempt from policy enforcement. No change needed here.
5. Update the `depguard` replacement/migration handling: since it's no longer in `DisabledLinters`,
   the fixer won't forcibly move it to disable. But `gomodguard_v2` (the v2 successor of
   `gomodguard`, a different linter) is unaffected.
6. Remove `DepguardSettings` from `linter_settings_generated.go` pruning — actually it should stay
   since it's generated from the schema. The fixer just won't prune it anymore since depguard is no
   longer in `DisabledLinters`.
7. Update BDD specs:
   - Remove "should move depguard from enable to disable" test in `fixer_test.go`
   - Add "depguard is never auto-enabled but respected when manually configured" test
   - Update `fixer_enforce_test.go` — depguard still tool-level managed via NeverAutoEnable
8. Run `go run ./scripts/validate_linter_data.go` to verify data integrity
9. Update AGENTS.md gotcha #10 to reflect the change

**Files:**
- `pkg/constants/rules.go` — move between maps
- `pkg/constants/linter_priorities.go` — add priority
- `pkg/constants/linter_reasons.go` — add reason
- `pkg/linter/fixer_test.go` — update specs
- `pkg/linter/fixer_enforce_test.go` — update specs
- `AGENTS.md` — update gotcha #10

**Effort:** M (2 hours, careful cross-file coordination)

---

## Tier 4: Polish (Lower Priority)

### 4.1 Enrich `varnamelen` defaults: typed declarations

**What:** Add `IgnoreDecls` field to `VarnamelenSettings` and expand defaults.

**Why:** The CV config uses typed `ignore-decls` entries (`db *sql.DB`, `w http.ResponseWriter`,
`tx *sql.Tx`, etc.) which are more precise than bare name lists. Our current default has `IgnoreNames`
with short names, but typed declarations prevent false positives on legitimate short names in specific
contexts. Also missing: `MaxDistance: 15`, `MinNameLength: 2`.

**How:**
1. Add `IgnoreDecls []string`, `MaxDistance int`, `MinNameLength int` fields to `VarnamelenSettings`
2. Set curated values in `DefaultLinterSettings`
3. Add BDD specs

**Effort:** S (1 hour)

---

### 4.2 Add `tagalign` curated default ordering

**What:** Add `DefaultLinterSettings` entry for `tagalign` with a curated tag order.

**Why:** Struct tags should follow a consistent order. The CV config demonstrates a sensible order:
`binding, json, yaml, xml, toml, validate, mapstructure`. Without a default, `tagalign` runs with
no ordering preference, allowing inconsistent tag order across a codebase.

**How:**
1. Add `TagalignSettings` hand-maintained struct with `Align`, `Order`, `Sort` fields
2. Register in `DefaultLinterSettings` with curated order
3. Add BDD specs

**Effort:** S (1 hour)

---

### 4.3 Add `lll` to `RedundantLinters` warning (already done — verify)

**What:** Verify that `lll` is properly flagged as redundant when `golines` is enabled.

**Why:** The CV config has `lll` settings but doesn't enable it. Our tool already has `lll` in
`RedundantLinters` (redundant with `golines`). This is working — no action needed, just verification.

**Status:** Already implemented. `lll` → redundant with `golines` formatter. ✓

**Effort:** None

---

## Implementation Order (Pareto)

| Phase | Items | Impact | Effort |
|-------|-------|--------|--------|
| **Phase 1** | 1.1 (mnd test exclusion), 1.2 (wrapcheck regexps), 1.3 (errcheck type-assertions) | Highest friction reduction per effort | ~3 hours |
| **Phase 2** | 3.1 (depguard policy change) | Unblocks architectural enforcement projects | ~2 hours |
| **Phase 3** | 2.1 (absolute path detection), 2.2 (duplicate exclusion linters), 2.3 (orphaned settings pruning) | Config health quality | ~3 hours |
| **Phase 4** | 1.4 (complexity linter defaults) | Reduces friction for 4 High-priority linters | ~3 hours |
| **Phase 5** | 4.1 (varnamelen typed decls), 4.2 (tagalign ordering) | Polish | ~2 hours |

**Total estimated effort:** ~13 hours across 5 phases.

---

## What We Chose NOT to Do (and Why)

| Item | Reason |
|------|--------|
| Add `funlen` threshold of 120/80 (CV's value) | Our 200/100 was calibrated across 160 projects. CV's tighter thresholds are project-specific. |
| Add `depguard` smart detection (file-pattern vs deny-list) | Decided: move to NeverAutoEnable instead. Simpler, no detection complexity. |
| Add `cyclop` threshold change (CV uses 15, we use 12) | Our value was calibrated across sibling projects. 12 is tighter and catches more. |
| Add `gosec` G104 exclusion | Our gosec defaults already exclude G304 and G115. G104 (unhandled errors) is already covered by `errcheck`. |
| Add `staticcheck` SA1019 disable | Disabling deprecation warnings globally is project-specific. Our default enables all staticcheck checks. |
| Add `wrapcheck` `ignore-package-globs` | `ignore-sig-regexps` (item 1.2) is more precise and covers the same ground. |
