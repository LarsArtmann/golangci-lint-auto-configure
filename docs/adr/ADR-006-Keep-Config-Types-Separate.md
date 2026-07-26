# ADR-006: Keep migration.Config and types.Config Separate

## Status

**Accepted** — 2026-07-26

## Context

The codebase has two `Config` struct types:

1. **`pkg/types/config_types.go`** (`types.Config`) — the canonical v2 config. Uses branded types (`LinterName`, `FormatterName`, `Version`), triple struct tags (`json`/`yaml`/`toml`), and the full v2 schema (linters under `linters.exclusions`, issues at top level).

2. **`pkg/migration/config_types.go`** (`migration.Config`) — the v1 migration config. Uses plain `string`/`[]string`, yaml-only tags, and the v1 schema (issues nested under `run`, `exclude-rules` at top level, `*bool`/`TriState` fields, deprecated `EnableAll`/`Fast` flags).

### Key Differences

| Aspect           | `types.Config`                   | `migration.Config`                                                                  |
| ---------------- | -------------------------------- | ----------------------------------------------------------------------------------- |
| Linter names     | `[]LinterName` (branded)         | `[]string`                                                                          |
| Formatter names  | `[]FormatterName` (branded)      | `[]string`                                                                          |
| Version          | `Version` (branded)              | `string`                                                                            |
| Struct tags      | `json`/`yaml`/`toml` (triple)    | `yaml` only                                                                         |
| Issues nesting   | Top-level `Issues` field         | Nested under `Run.RunIssues` (v1 layout)                                            |
| Exclude rules    | Under `Linters.Exclusions.Rules` | Top-level `ExcludeRules` (v1 layout)                                                |
| `*bool`/TriState | Not needed (v2 only)             | 7 fields use `TriState` for v1 compatibility                                        |
| v1-only fields   | None                             | `EnableAll`, `Fast`, `ExcludeFiles`, `ExcludeDirs`, 3 `*UseDefault` TriState fields |

### Why Two Types Exist

The migration package must parse v1 YAML configs that have a fundamentally different schema. v1 nests issues under `run`, has top-level exclude-rules, and uses `*bool` pointers for "use default" flags. The migration flow:

1. Parses v1 YAML into `migration.Config` (with custom `UnmarshalYAML` that hoists `issues.*` fields)
2. Transforms fields into v2 layout via `migrator.go`
3. Outputs `types.Config` (canonical v2)

## Decision

**Keep the two Config types separate. Do NOT unify them.**

### Rationale

1. **v1 is maintenance-only.** Per AGENTS.md: "0 live v1 configs across 160 sibling projects." The `migrate` subcommand is kept functional but no new v1 features will be added. Investing HIGH-risk effort to unify types for a dead format is negative ROI.

2. **The schemas are genuinely different.** v1 and v2 configs have different field layouts, different nesting, and different semantics. A unified type would require either:
   - Embedding v1-only fields in `types.Config` (complexity leak into the canonical type)
   - A shim type wrapping `types.Config` (essentially recreating `migration.Config` with extra indirection)

   Either option adds complexity without removing it.

3. **Branded types are a v2 concern.** `types.Config` uses `LinterName`/`FormatterName`/`Version` for compile-time safety. Forcing these onto the migration package would require converting at every boundary, adding friction to the migration code with no benefit (migration reads arbitrary strings from v1 configs).

4. **The migration flow already works correctly.** Tests pass, golden files match, and the migrator produces valid v2 output. Unification risks breaking this for zero user-visible improvement.

### Risk of Unification (if attempted)

- **HIGH risk to migration round-trip.** The custom `UnmarshalYAML` on `migration.Config` hoists v1 `issues.*` fields. Replicating this on a unified type is non-trivial.
- **YAML tag conflicts.** `types.Config` has triple tags (`json`/`yaml`/`toml`); `migration.Config` has yaml-only. Unifying would either add unused tags to migration or remove needed tags from types.
- **Test coverage gaps.** The migration package has golden file tests for v1→v2 conversion. Unification would invalidate these and require re-verification of every field mapping.

## Test Gaps Identified

If unification is ever revisited, these test areas need coverage BEFORE implementation:

1. **TriState round-trip:** Verify that all 7 `TriState` fields (`ExcludeDirUseDefault`, `ExcludeRulesUseDefault`, `ExcludeUseDefault`, `Run.SkipDirsUseDefault`, `RunIssues.ExcludeUseDefault`, `RunIssues.ExcludeRulesUseDefault`, `RunIssues.ExcludeDirUseDefault`) correctly map to v2 defaults (absent/unspecified).

2. **v1 issues hoisting:** The custom `UnmarshalYAML` hoists `issues.exclude-rules` and `issues.exclude-dirs` from nested-under-run to top-level. Verify every field is hoisted correctly.

3. **EnableAll/Fast deprecation:** v1 `linters.enable-all` and `linters.fast` flags are deprecated in v2. Verify they're dropped or converted during migration.

4. **Branded type conversion:** If unified, `[]string` → `[]LinterName` conversion must handle invalid linter names gracefully (the migration package currently passes raw strings through).

## Consequences

- Two Config types remain in the codebase. This is a known, documented divergence.
- The migration package maintains its own sub-struct names (`Run` vs `RunConfig`, `Linters` vs `LintersConfig`, etc.).
- No migration round-trip risk.
- If v1 support is ever dropped entirely, `migration.Config` and all its sub-types can be deleted in one pass.
