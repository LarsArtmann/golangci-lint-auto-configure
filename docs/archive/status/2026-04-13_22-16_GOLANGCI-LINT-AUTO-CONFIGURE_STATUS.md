# Comprehensive Status Report — 2026-04-13 22:16

**Project:** `golangci-lint-auto-configure`
**Date:** 2026-04-13 22:16 (Monday)
**Author:** AI Agent (via Crush)
**Context:** User rage-reaction to `golangci-lint-auto-configure configure && golangci-lint run --fix -v` failing with duplicate `linters:` key and `priority=high` default instead of `optional`.

---

## WORK STATUS

### a) FULLY DONE

| Task                                        | Status  | Notes                                                                                                                                                                                                  |
| ------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Codebase research — all critical files read | ✅ DONE | fixer.go, loader.go, merger.go, cmd*configure.go, cmd_builder.go, commands.go, fixer_preflight.go, fixer_config.go, fixer_formatters.go, all merger*\*.go, linter_priorities.go, types.go, analyzer.go |
| YAML parsing investigation                  | ✅ DONE | go.yaml.in/yaml/v3 correctly handles go-localfirst config — no bug in yaml library                                                                                                                     |
| Config struct analysis                      | ✅ DONE | types.Config and all sub-structs have correct yaml tags                                                                                                                                                |
| Priority default bug — ROOT CAUSE FOUND     | ✅ DONE | Global `--priority` default is `"high"` in `commands.go:200`, subcommand says `"optional"` — global wins                                                                                               |
| git history analysis (go-localfirst)        | ✅ DONE | Original duplicate key was in pre-8bf0088 commit, fixed in 8bf0088                                                                                                                                     |
| Build environment diagnosis                 | ✅ DONE | Nix Go conflict: `/nix/store/...go-1.26.0` vs actual Go 1.26.1                                                                                                                                         |
| Status report written                       | ✅ DONE | `docs/status/2026-04-10_11-40_COMPREHENSIVE_STATUS.md` (uncommitted)                                                                                                                                   |

### b) PARTIALLY DONE

| Task                                     | Status     | Blocker                                                                                 |
| ---------------------------------------- | ---------- | --------------------------------------------------------------------------------------- |
| Priority default bug — FIX               | 🔄 PARTIAL | Root cause found, fix is 1-line change in commands.go:200                               |
| YAML duplicate key bug — ROOT CAUSE      | 🔄 PARTIAL | Hypothesis: pre-flight multi-save chain + golangci-lint fmt integration corrupts config |
| Fix implementation — YAML duplicate key  | ⏸️ WAITING | Need to verify hypothesis with live test                                                |
| Fix implementation — single-save pattern | ⏸️ WAITING | Need to refactor fixer_preflight.go                                                     |
| Fix implementation — priority default    | ⏸️ WAITING | 1-line change ready                                                                     |

### c) NOT STARTED

| Task                                                                                     | Status |
| ---------------------------------------------------------------------------------------- | ------ |
| Fix: Change global `--priority` default from `"high"` to `"optional"` in commands.go:200 |
| Fix: Prevent YAML duplicate key — restructure pre-flight saves to single-save            |
| Fix: Broken Go build environment (Nix Go conflict)                                       |
| Test all fixes on go-localfirst project                                                  |
| Test all fixes on golangci-lint-auto-configure's own project                             |
| Commit all changes                                                                       |
| Push to remote                                                                           |

### d) TOTALLY FUCKED UP

| Issue                                   | Severity    | Detail                                                                                                |
| --------------------------------------- | ----------- | ----------------------------------------------------------------------------------------------------- |
| Global `--priority` default is `"high"` | 🔴 CRITICAL | `commands.go:200` — `golangci-lint-auto-configure configure` enables only ~30 linters instead of ~109 |
| Config corruption bug                   | 🔴 CRITICAL | `--fix` mode produces duplicate `linters:` key (unconfirmed root cause)                               |
| Build broken                            | 🔴 CRITICAL | Nix Go conflict makes `go test ./...` fail with "package encoding is not in std"                      |
| Uncommitted status report               | 🟡 MODERATE | `docs/status/2026-04-10_11-40_COMPREHENSIVE_STATUS.md` written but not committed                      |
| Confusing help text                     | 🟡 MODERATE | "priority=high" vs subcommand help "default: optional" is contradictory                               |

---

## ROOT CAUSE: Priority Default Bug

**File:** `internal/cli/commands.go:200`

```go
rootCmd.PersistentFlags().
    StringVar(&priority, "priority", "high", "Minimum priority level...")
```

The configure subcommand (`cmd_configure.go:103`) ALSO registers `--priority` with default `"optional"`:

```go
cmd.Flags().StringVar(&priority, "priority", "optional", "...")
```

Both flags bind to the SAME `priority` variable. Flag registration order:

1. `addSubCommands()` is called FIRST → subcommand flags register → `priority = "optional"`
2. `registerGlobalFlags()` is called SECOND → global flags register → `priority = "high"` ← LAST WINS

**Fix:** Change `commands.go:200` from `"high"` to `"optional"`.

---

## ROOT CAUSE: YAML Duplicate Key Bug

**Error:** `yaml: unmarshal errors: line 136: mapping key "linters" already defined at line 15`

**Verified:** The yaml library (go.yaml.in/yaml/v3) correctly marshal/unmarshal-cycles the go-localfirst config without producing duplicate keys.

**Hypothesis:** The pre-flight fix chain calls `SaveConfig()` up to 4 times in sequence. After the final save, `golangci-lint fmt --config=<path>` is called. This `fmt` command might produce YAML that breaks subsequent reads.

**The pre-flight chain:**

```go
// fixer_preflight.go — up to 4 saves before final applyAndSave:
preFixInvalidDurations()    // SAVE 1
preFixVersion()             // SAVE 2
preFixDeprecatedLinters()   // SAVE 3
preFixTypecheck()           // SAVE 4
// Then applyAndSave()       // SAVE 5 (final)
```

**Fix:** Restructure to collect all changes and save ONCE at the end.

---

## BUILD ENVIRONMENT BUG: Nix Go Conflict

```
/nix/store/5ajixjk279m40yf6x96xxlnvw1wg6hq3-go-1.26.0/share/go/src/encoding: package encoding is not in std
```

System Go is `/run/current-system/sw/bin/go` (v1.26.1) but GOPATH/GOROOT points to broken Nix Go installation.

**Fix:** Set `GOPATH` and `GOROOT` explicitly, or use `GOROOT_FINAL` environment variable.

---

## TOP #25 THINGS TO GET DONE NEXT

1. [FIX] Change `commands.go:200` `--priority` default from `"high"` to `"optional"`
2. [FIX] Refactor `fixer_preflight.go` to single-save pattern — collect all changes, save once
3. [FIX] Add post-save yaml validation — load saved config to verify it's parseable
4. [TEST] Test fixer on go-localfirst project — does it produce duplicate key?
5. [TEST] Test `golangci-lint fmt --config=<file>` — does it produce parseable YAML?
6. [BUILD] Fix Nix Go conflict — set explicit GOPATH/GOROOT or use `go env`
7. [ARCH] Remove priority system entirely — auto-detect project type, enable ALL sensible linters
8. [ARCH] Remove `RunFmtCommand` call after save — it's risky, produces unexpected format changes
9. [ARCH] Remove redundant global `--priority` flag from configure subcommand
10. [CLEANUP] Commit the uncommitted status report
11. [CLEANUP] Remove dead code in merger files
12. [TEST] Add yaml round-trip integration tests for all config scenarios
13. [TEST] Add BDD test: run fixer → save → load → verify parseable by golangci-lint
14. [DOCS] Update `configureLong` to match actual behavior
15. [DOCS] Update README to explain priority system (or remove it)
16. [FEATURE] Add `--validate-only` mode — don't save, just validate
17. [FEATURE] Add `--smart-detect` that auto-selects linters based on project type
18. [BUG] Investigate if `golangci-lint fmt` produces v1 vs v2 yaml incompatible output
19. [CI] Add CI step that runs golangci-lint-auto-configure on itself
20. [CI] Add CI step that verifies output config can be parsed by golangci-lint linters
21. [TEST] Test with `golangci-lint run --fix` on projects after configuration
22. [ARCH] Consider making fixer read-only by default with `--apply` flag to save
23. [PERF] Profile pre-flight chain — is 4 separate saves slow on large configs?
24. [CLEANUP] Remove unused `dryRun` parameter in `handlePresetMode`
25. [TEST] Test on a project with ALL linters already enabled — verify no-op is correct

---

## TOP #1 QUESTION I CAN NOT FIGURE OUT

**Does `golangci-lint fmt --config=<file>` produce YAML that `go.yaml.in/yaml/v3` cannot re-unmarshal?**

I tested `go.yaml.in/yaml/v3` (v3.0.4) directly — it correctly marshal/unmarshal-cycles the go-localfirst config. BUT the golangci-lint tool itself uses a DIFFERENT yaml library internally. When `golangci-lint fmt` runs, it might produce YAML in a format that breaks the go.yaml.in/yaml/v3 unmarshaler on the next read cycle.

Specifically: does `golangci-lint fmt` change the yaml in a way that corrupts the `output.formats: {}` or `linters.settings` or `formatters` sections?

**To answer this, I need to:**

1. Save a known-good config
2. Run `golangci-lint fmt --config=<file>` on it
3. Check if the output is still valid yaml
4. Check if it can be re-unmarshaled by go.yaml.in/yaml/v3

---

## SUMMARY

| Issue                                    | Status               | Fix Complexity                          |
| ---------------------------------------- | -------------------- | --------------------------------------- |
| Priority default `"high"` → `"optional"` | ROOT CAUSE FOUND     | **1-line fix**                          |
| YAML duplicate key corruption            | ROOT CAUSE SUSPECTED | Multi-save chain → single-save refactor |
| Build broken (Nix Go)                    | ROOT CAUSE FOUND     | Environment variable fix                |
| Uncommitted status report                | PENDING              | `git add` + commit                      |
| Remove priority system                   | NOT STARTED          | Medium refactor                         |
| Auto-detect project type                 | NOT STARTED          | Medium new feature                      |

**Immediate action items:**

1. Fix priority default (1 line)
2. Refactor pre-flight saves to single-save
3. Fix Nix Go build environment
4. Commit status report
5. Test fixes on go-localfirst project
