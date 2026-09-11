# Architecture Review: golangci-lint-auto-configure

**Date:** 2026-07-25
**Reviewer:** Automated (architecture-review skill)
**Scope:** Full codebase architecture assessment

---

## Executive Summary

The codebase exhibits **strong modular architecture** with clean package boundaries, interface-driven design, and zero harmful code duplication. The domain layer (`pkg/`) is well-separated from the CLI layer (`internal/cli/`). Key improvements this session include splitting the ConfigLoader God Object interface, splitting cmd_configure.go (680→4 files), and adding multi-preset support.

**Architecture Score: 8/10** — Production-ready with minor improvements possible.

---

## Package Structure (17 packages + 2 commands)

| Package           | Files | Role                         | Dependencies                                    |
| ----------------- | ----- | ---------------------------- | ----------------------------------------------- |
| `pkg/types`       | 9     | Domain types, interfaces     | None (leaf)                                     |
| `pkg/constants`   | 14    | Static linter data           | types                                           |
| `pkg/config`      | 6     | Config load/save/validate    | types, constants, errors                        |
| `pkg/linter`      | 18    | Analysis, fixing, formatting | types, constants, config, errors, audit, policy |
| `pkg/detection`   | 5     | Project type detection       | types                                           |
| `pkg/errors`      | 6     | Error classification         | types, go-error-family                          |
| `pkg/finding`     | 5     | Finding model + conversion   | types, finding-lib                              |
| `pkg/report`      | 6     | HTML/JSON/SARIF reports      | types, finding                                  |
| `pkg/audit`       | 3     | Audit ledger (JSONL)         | None                                            |
| `pkg/policy`      | 3     | Disable-reason enforcement   | types, constants                                |
| `pkg/client`      | 2     | golangci-lint binary client  | None                                            |
| `pkg/migration`   | 5     | v1→v2 config migration       | types, errors                                   |
| `pkg/diff`        | 2     | Config differ                | types                                           |
| `pkg/ui`          | 4     | Terminal output              | types                                           |
| `pkg/utils`       | 3     | Git, retry                   | errors                                          |
| `pkg/version`     | 2     | Version info                 | None                                            |
| `pkg/gogenfilter` | 2     | Generated file detection     | gogenfilter-lib                                 |
| `internal/cli`    | 16    | CLI commands                 | ALL pkg/*                                       |

### Dependency Direction

```text
internal/cli → pkg/linter → pkg/config → pkg/types (leaf)
                                    ↗
              pkg/constants ────────┘
              pkg/detection ────────┘
              pkg/errors ───────────┘
```

**Clean dependency graph:** No circular dependencies. `pkg/types` is the leaf with zero internal dependencies. `internal/cli` is the only consumer of all packages.

---

## Strengths

### 1. Interface-Driven Design

- `ConfigLoader` decomposed into 6 focused sub-interfaces (ConfigReader, ConfigWriter, ConfigDiscoverer, ConfigValidator, ConfigInspector, ConfigCreator)
- `LinterAnalyzer` interface for analysis
- `presetConfigLoader` narrow interface in CLI layer
- Consumers depend on the narrowest interface they need

### 2. Error Classification

- go-error-family integration with BSD sysexits (exit 1/65/69/75)
- 27 domain message templates registered (Wix-style What/Why/Fix/WayOut)
- Sentinel errors registered with Families (Rejection/Conflict/Corruption/Infrastructure)
- CLI HandleError renders user-friendly messages with fallback to slog

### 3. Test Infrastructure

- BDD testing (Ginkgo/Gomega) across all packages
- Integration tests with real binaries (exit code tests, coverage-check)
- Data integrity tests verifying constant consistency
- 20/20 packages passing with -race

### 4. Zero Duplication

- art-dupl reports 0 clone groups at threshold 5
- Core linter list extracted to shared `coreLinters` + `withCore()` pattern
- ConfigLoader composite interface eliminates interface duplication

### 5. Config Round-Trip Safety

- Config types use kebab YAML tags matching golangci-lint schema
- Report types use tag-free PascalCase for JSON output
- `map[string]any` preserved for unknown config fields

---

## Concerns

### 1. Package-Level CLI Globals (Medium)

**Location:** `internal/cli/commands.go` — `priority`, `dryRun`, `verbose`, `quiet`, `configPath`, `noAudit`, `pragmatic`, `showDiff`, `detectedExtraFormatters`

**Issue:** 9 package-level variables act as implicit context for CLI commands. This works but makes testing harder (state leaks between tests) and prevents parallel command execution.

**Recommendation:** Extract a `CommandContext` struct that carries flags explicitly. This is a medium-effort refactor that improves testability significantly.

### 2. Concrete Type Coupling in CLI (Low)

**Location:** `internal/cli/cmd_configure.go` — uses `*config.Loader` (concrete) and `*linter.Analyzer` (concrete) directly in function signatures.

**Issue:** Some functions accept the concrete `*config.Loader` while others accept the `presetConfigLoader` interface. This inconsistency means some code paths can't be tested with mocks.

**Recommendation:** Use `types.ConfigReader` / `types.ConfigWriter` sub-interfaces in function signatures that only need load/save.

### 3. Formatter/Linter Name String Type Safety (Low)

**Location:** Various files — linter names are `string` in some places and `types.LinterName` in others.

**Issue:** `types.LinterName` exists as a named string type but is used inconsistently. Bare strings appear in detection patterns and config loader fallback.

**Recommendation:** This is acceptable — the `coreLinters` extraction already reduced the duplication. Full migration to typed constants would add complexity for marginal benefit.

---

## Scalability Assessment

| Dimension                 | Rating    | Notes                                                                 |
| ------------------------- | --------- | --------------------------------------------------------------------- |
| **Package isolation**     | Excellent | 17 packages with clean boundaries, no cycles                          |
| **Interface segregation** | Good      | ConfigLoader split into 6 sub-interfaces; Fixer uses narrow interface |
| **Test coverage**         | Good      | BDD across all packages; integration tests for critical paths         |
| **Error handling**        | Excellent | go-error-family with domain templates; BSD sysexits                   |
| **Code duplication**      | Excellent | 0 clone groups (art-dupl)                                             |
| **Type safety**           | Good      | Config types typed; some string-typed linter names remain             |
| **DI consistency**        | Fair      | Mix of concrete types and interfaces in CLI layer                     |
| **Global state**          | Fair      | 9 package-level vars in CLI; manageable but not ideal                 |

---

## Composability Wins

1. **Multi-preset merge:** `--preset minimal --preset security` combines presets with deduplication
2. **Format detection:** `--preset format --detect` auto-enables swaggo based on project analysis
3. **Preset recommendation:** `--recommend` analyzes project and applies multiple presets
4. **Config sub-interfaces:** Consumers depend on minimal interfaces (ConfigReader, ConfigWriter, etc.)
5. **Error templates:** 27 domain codes resolve to user-friendly messages

---

## Action Roadmap

| Priority | Action                                                     | Effort | Impact |
| -------- | ---------------------------------------------------------- | ------ | ------ |
| 1        | Extract CommandContext struct to replace CLI globals       | 2-3h   | Medium |
| 2        | Use ConfigReader/ConfigWriter sub-interfaces in fixer mode | 1h     | Low    |
| 3        | Add settings key validation against golangci-lint schema   | 2h     | Medium |
| 4        | Generate linter settings from JSON Schema                  | 4-6h   | Medium |

---

## Conclusion

The architecture is solid and production-ready. The domain layer is clean, the CLI layer is well-organized (especially after the cmd_configure.go split), and the error handling is best-in-class. The main improvement opportunity is reducing package-level globals in the CLI layer, which would improve testability and enable parallel command execution.

---

## Action Roadmap status (2026-07-25, later session)

| # | Action                                               | Status  | Outcome                                                 |
| - | ---------------------------------------------------- | ------- | ------------------------------------------------------- |
| 1 | Extract `CommandContext` struct for CLI globals      | ❌ Open | Tracked in `TODO_LIST.md` (Medium Priority)             |
| 2 | Use ConfigReader/Writer sub-interfaces in fixer mode | ✅ Done | `ConfigLoader` split into 6 sub-interfaces (`39cca87`)  |
| 3 | Settings key validation against golangci-lint schema | ✅ Done | Soft warnings at config load (`e8f30f0`)                |
| 4 | Generate linter settings from JSON Schema            | ✅ Done | `cmd/generate-settings` produces 88 structs (`3665d79`) |

The "Architecture Score: 8/10" assessment stands. The concerns (CLI globals,
concrete-type coupling in a few signatures) remain the highest-leverage
improvements; #1 above is the only outstanding one of consequence.
