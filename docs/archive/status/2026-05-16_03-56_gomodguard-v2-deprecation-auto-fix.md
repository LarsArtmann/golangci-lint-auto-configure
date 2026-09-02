# Status Report — 2026-05-16 03:56 CEST

**Session Focus:** gomodguard → gomodguard_v2 deprecation auto-fix + settings migration

---

## A. FULLY DONE

### 1. gomodguard → gomodguard_v2 Deprecation Auto-Fix

**Problem:** golangci-lint v2.12.0 deprecated `gomodguard` in favor of `gomodguard_v2`. The tool showed warnings but had no auto-fix.

**Solution implemented across 7 files:**

| File                                 | Change                                                                                                                                                                                              |
| ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/constants/rules.go`             | Added `gomodguard` → `gomodguard_v2` to `DeprecatedLinters` map                                                                                                                                     |
| `pkg/constants/linter_priorities.go` | Added `gomodguard_v2` with Medium priority                                                                                                                                                          |
| `pkg/constants/linter_reasons.go`    | Added reason string for `gomodguard_v2`                                                                                                                                                             |
| `pkg/finding/categories.go`          | Added both `gomodguard` and `gomodguard_v2` to Structure category                                                                                                                                   |
| `pkg/linter/fixer_deprecated.go`     | **Enhanced**: `replaceLinters` now migrates linter settings keys (e.g., `linters.settings.gomodguard` → `linters.settings.gomodguard_v2`). Extracted `replaceOne` + added `migrateSettings` method. |
| `pkg/linter/fixer.go`                | Updated call site to pass `cfg` to `replaceLinters`                                                                                                                                                 |
| `pkg/linter/fixer_test.go`           | Added 3 BDD tests: dry-run detection, non-dry-run fix, settings migration                                                                                                                           |

**Key architectural improvement:** The `migrateSettings` method is generic — it benefits ALL future deprecation replacements, not just gomodguard. Previously, replacing a deprecated linter would leave orphaned settings under the old key name in the YAML.

### 2. Previously Completed (from earlier sessions)

- gogenfilter/v3 integration for auto-generated file exclusion scanning
- go-finding unified model integration (SARIF, JSON, finding formats)
- v1 → v2 config migration (merged from golangci-config-migrator)
- HTML/JSON/SARIF report generation
- 7 CLI subcommands (configure, analyze, validate, report, migrate, install-hook, completion)
- 119 linters categorized with priorities and reasons
- Pre-commit hook installation
- Nix flake build system
- BDD testing with Ginkgo v2 + Gomega

---

## B. PARTIALLY DONE

### 1. ✅ Own Config Fixed: `gomodguard` → `gomodguard_v2`

**FIXED in commit `4a5e1a3`** — Replaced deprecated `gomodguard` with `gomodguard_v2` in `.golangci.yml`. Deprecation warning eliminated.

### 2. ✅ `gochecknoglobals` Fixed in `pkg/ui/finding_formatter.go`

**FIXED in commit `4a5e1a3`** — Replaced global `severityData` variable with local `switch` statements in `severityLabel()` and `severityBadge()` functions. `golangci-lint run` now reports **0 issues**.

### 3. Coverage Gaps

| Package             | Coverage |
| ------------------- | -------- |
| `internal/cli`      | 8.5%     |
| `pkg/finding`       | 58.1%    |
| `pkg/gogenfilter`   | 62.5%    |
| `pkg/detection`     | 65.0%    |
| `pkg/migration`     | 66.8%    |
| `pkg/config`        | 65.9%    |
| `pkg/ui`            | 67.7%    |
| `pkg/types`         | 57.7%    |
| `pkg/version`       | 51.4%    |
| `cmd/`              | 0.0%     |
| `pkg/report`        | 0.0%     |
| `internal/cli/cmd/` | 0.0%     |

Strong coverage: `pkg/constants` (100%), `pkg/errors` (100%), `pkg/diff` (96.5%), `pkg/utils` (94.6%), `pkg/linter` (78.6%).

---

## C. NOT STARTED

1. **FEATURES.md** — No feature inventory document exists
2. **TODO_LIST.md** — No comprehensive TODO list document exists
3. **CONTEXT.md** — No domain context document exists
4. **ADR directory** — No Architecture Decision Records (`docs/adr/`)
5. ✅ **Dogfooding** — Fixed in commit `4a5e1a3` (`gomodguard` → `gomodguard_v2`, depguard cleanup)
6. **CI/CD Pipeline Updates** — GitHub Actions may need updates for golangci-lint v2.12.x compatibility
7. **Integration Tests** — No end-to-end integration tests that run the actual CLI binary against real configs
8. **Performance Benchmarking** — No benchmarks for analysis/fixing operations
9. **Documentation Site** — No generated API docs or user guide beyond README.md

---

## D. TOTALLY FUCKED UP

Nothing catastrophic. Clean sailing this session. However:

### Pre-existing Concerns

**RESOLVED (commit `4a5e1a3`):**

1. ✅ **Own config is stale** — Fixed: `gomodguard` replaced with `gomodguard_v2` in `.golangci.yml`
2. ✅ **`gochecknoglobals` fixed** — Refactored `severityData` global in `finding_formatter.go` to local `switch` statements
3. ✅ **Stale depguard entries removed** — `samber/mo` and `spf13/afero` removed from allowlist

**Remaining concerns:**

4. **go-finding local replace** — `replace github.com/larsartmann/go-finding => ../go-finding` in `go.mod` means local dev requires the sibling directory. CI/Nix builds handle this, but it's friction for new contributors.

---

## E. WHAT WE SHOULD IMPROVE

### High Impact

1. ✅ **Dogfood the tool** — Fixed in commit `4a5e1a3`: `gomodguard` → `gomodguard_v2`
2. ✅ **Fix the `gochecknoglobals` issue** — Fixed in commit `4a5e1a3`: refactored `severityData` to switch statements
3. **Create FEATURES.md** — Inventory all features for better project understanding
4. **Create TODO_LIST.md** — Comprehensive backlog from docs and code
5. **Improve test coverage** — Target `internal/cli` (8.5%) and `pkg/finding` (58.1%) especially

### Medium Impact

6. **Integration/E2E tests** — Test the actual binary against real golangci-lint configs
7. **ADR records** — Document key architectural decisions (go-finding integration, gogenfilter, BDD testing)
8. ✅ **Remove `samber/mo` from depguard allowlist** — Fixed in commit `4a5e1a3`
9. ✅ **Remove `spf13/afero` from depguard allowlist** — Fixed in commit `4a5e1a3`
10. **Stale exclusion rules audit** — The `.golangci.yml` has many per-file exclusions; some may be obsolete

### Lower Impact

11. **Version tagging** — No semver tags yet; tool is pre-release
12. **README.md refresh** — May not reflect all current features
13. **Report templates** — Only one HTML report template; could add more output formats

---

## F. Top #25 Things to Get Done Next

| #  | Priority | Task                                                                                | Impact                       |
| -- | -------- | ----------------------------------------------------------------------------------- | ---------------------------- |
| 1  | ✅ DONE  | Dogfood: Fix `gomodguard` → `gomodguard_v2` in own `.golangci.yml`                  | Commit `4a5e1a3`             |
| 2  | ✅ DONE  | Fix `gochecknoglobals` in `finding_formatter.go:69`                                 | Commit `4a5e1a3`, 0 issues   |
| 3  | HIGH     | Create `FEATURES.md` with honest feature inventory                                  | Project clarity              |
| 4  | HIGH     | Create `TODO_LIST.md` comprehensive backlog                                         | Execution roadmap            |
| 5  | HIGH     | Add integration/E2E tests for CLI binary                                            | Confidence in releases       |
| 6  | HIGH     | Improve `internal/cli` test coverage (currently 8.5%)                               | Core path coverage           |
| 7  | ✅ DONE  | Remove stale `samber/mo` from depguard allow list                                   | Commit `4a5e1a3`             |
| 8  | ✅ DONE  | Remove stale `spf13/afero` from depguard allow list                                 | Commit `4a5e1a3`             |
| 9  | HIGH     | Audit `.golangci.yml` exclusion rules for obsolescence                              | Config hygiene               |
| 10 | MEDIUM   | Improve `pkg/finding` coverage (58.1%)                                              | Finding pipeline reliability |
| 11 | MEDIUM   | Improve `pkg/types` coverage (57.7%)                                                | Core types reliability       |
| 12 | MEDIUM   | Create `docs/adr/` with architecture decision records                               | Knowledge preservation       |
| 13 | MEDIUM   | Add performance benchmarks for analysis/fixing                                      | Performance awareness        |
| 14 | MEDIUM   | Refresh `README.md` to reflect current features                                     | User documentation           |
| 15 | MEDIUM   | Add `CONTEXT.md` for domain language                                                | Onboarding                   |
| 16 | MEDIUM   | Verify CI/CD works with golangci-lint v2.12.x                                       | Pipeline health              |
| 17 | MEDIUM   | Add `gomodguard_v2` settings schema to migration rules                              | Complete migration path      |
| 18 | LOW      | Tag first semver release (v0.1.0)                                                   | Release management           |
| 19 | LOW      | Add changelog generation                                                            | Release documentation        |
| 20 | LOW      | Explore `golangci-lint` plugin system for tighter integration                       | Future architecture          |
| 21 | LOW      | Add `--verbose` flag output improvements                                            | Debugging experience         |
| 22 | LOW      | Review `pkg/report/` for 0% coverage                                                | Report reliability           |
| 23 | LOW      | Add example configs for `gomodguard_v2` in `examples/`                              | User guidance                |
| 24 | LOW      | Consider adding `deprecated-linters` command to list all known deprecations         | Discoverability              |
| 25 | LOW      | Evaluate moving from justfile to flake.nix for build tasks (per AGENTS.md guidance) | Build system alignment       |

---

## G. Top #1 Question I Cannot Figure Out Myself

**RESOLVED — Manual update was chosen.** The config was manually updated in commit `4a5e1a3` to replace `gomodguard` with `gomodguard_v2`, rather than running the full `configure` command which might make other changes. This preserves the project's config as a curated reference.

---

## Project Health Summary

| Metric                           | Value                      | Status              |
| -------------------------------- | -------------------------- | ------------------- |
| Go Version                       | 1.26+                      | ✅ Current          |
| golangci-lint Version            | v2.12.2                    | ✅ Current          |
| Tests                            | 14 suites, ALL PASS        | ✅ Green            |
| Composite Coverage               | 59.8%                      | ⚠️ Needs improvement |
| Lint Issues                      | 0 (all fixed)              | ✅ Clean            |
| Build                            | Clean                      | ✅                  |
| Packages                         | 20                         | ✅                  |
| Go Source Files                  | 96                         | ✅                  |
| Total Lines of Code              | ~16,100                    | ✅                  |
| Deprecated Linters in Own Config | 0 (`gomodguard_v2` in use) | ✅ Fixed            |
| Modified Files This Session      | 7                          | ✅ Focused changes  |
| New Tests This Session           | 3                          | ✅                  |

---

## Files Changed This Session

```
pkg/constants/rules.go              (+4 lines)  — gomodguard deprecation entry
pkg/constants/linter_priorities.go  (+1 line)   — gomodguard_v2 priority
pkg/constants/linter_reasons.go     (+1 line)   — gomodguard_v2 reason
pkg/finding/categories.go           (+3 lines)  — category mappings
pkg/linter/fixer.go                 (~1 line)   — pass cfg to replaceLinters
pkg/linter/fixer_deprecated.go      (+38 lines) — settings migration + refactor
pkg/linter/fixer_test.go            (+45 lines) — 3 new BDD tests
```

## Follow-up: Stale Issue Cleanup (commit `4a5e1a3`)

Fixed three items that had been pending across prior sessions:

```
.golangci.yml                       (-3 deps, +1 linter rename) — depguard + gomodguard_v2
pkg/ui/finding_formatter.go         (rewrite) — switch-based severity lookup, no globals
```

---

_Generated: 2026-05-16 03:56 CEST_
_Updated: 2026-05-16 05:56 CEST_
