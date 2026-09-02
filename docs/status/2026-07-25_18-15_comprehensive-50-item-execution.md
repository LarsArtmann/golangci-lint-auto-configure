# Status Report: Comprehensive 50-Item TODO Execution

**Date:** 2026-07-25 18:15 CEST
**Session scope:** Execute ALL 50 items from the quality debt backlog (section f of the prior report), sorted by impact/effort/customer-value.
**Result:** 32 of 50 items completed or resolved. 18 remain (all large refactors, complex features, or skill invocations).

> **Update (later 2026-07-25 session):** all 18 "remaining" items below were
> completed the same day by the `2026-07-25_20-55` session (commits
> `03a0806`…`40eda4c`). The "Remaining Items" table is now a historical record of
> what was deferred, not a backlog. Per-item resolution in
> [Resolution](#resolution-2026-07-25-later-session) at the end of this file.
> The only genuinely-open follow-up is the version bump + git tag (C19 of the
> friction-reduction plan), now tracked in `TODO_LIST.md`.

---

## Final Verification

| Check                                                         | Result                  |
| ------------------------------------------------------------- | ----------------------- |
| `go build ./...`                                              | Clean                   |
| `golangci-lint run --config=.golangci.yml --timeout=5m ./...` | **0 issues**            |
| `go test -race ./pkg/... ./internal/... ./cmd/...`            | **20/20 packages pass** |
| `nix flake check --no-build`                                  | **All checks passed**   |
| `git push origin master`                                      | **13 commits pushed**   |

---

## Completed Items Table (sorted by impact/effort/customer-value)

| #  | Item                                    | Category      | Impact   | Effort | How Resolved                                                      |
| -- | --------------------------------------- | ------------- | -------- | ------ | ----------------------------------------------------------------- |
| 1  | Pin GitHub Actions to SHAs (#1)         | Security      | Critical | 30min  | 21 actions pinned across 4 workflow files                         |
| 2  | Fix legacyerrors nolint (#2)            | Lint hygiene  | High     | 5min   | Removed 4 stale directives; converted to regular comments         |
| 3  | Audit all nolint directives (#49)       | Lint hygiene  | High     | 10min  | All 23 remaining directives verified valid                        |
| 4  | Add --no-color flag (#11)               | UX/CI         | High     | 20min  | Sets NO_COLOR=1; lipgloss respects it                             |
| 5  | Add wrapcheck default settings (#32)    | Type safety   | Medium   | 15min  | WrapcheckSettings with curated ignore-sigs                        |
| 6  | Add HandleError at CLI boundary (#38)   | Architecture  | Medium   | 15min  | Extracted HandleError function from Main()                        |
| 7  | Add --json to presets command (#36)     | UX/Scripting  | Medium   | 20min  | Structured JSON output with all preset details                    |
| 8  | Extract vendorHash.nix (#6)             | Build hygiene | Medium   | 10min  | Separate file for cleaner dependency diffs                        |
| 9  | Add coverage-check to Nix build (#4)    | Build         | Medium   | 15min  | subPackages + apps.coverage-check                                 |
| 10 | Data integrity test for settings (#15)  | Testing       | Medium   | 10min  | Verifies DefaultLinter/FormatterSettings keys match priority maps |
| 11 | Add Dependabot automation (#9)          | CI/CD         | Medium   | 5min   | .github/dependabot.yml for Actions + Go modules                   |
| 12 | Add git-cliff config (#46)              | Automation    | Medium   | 10min  | cliff.toml for changelog generation                               |
| 13 | go mod tidy (#5)                        | Build hygiene | Low      | 1min   | Already clean — no changes needed                                 |
| 14 | Update code-organization.md (#23)       | Docs          | Medium   | 15min  | Added pkg/audit, pkg/policy, pkg/client, pkg/utils                |
| 15 | Document coverage-check in README (#22) | Docs          | Low      | 10min  | Added "Development Tools" section                                 |
| 16 | Add CHANGELOG entries (#21)             | Docs          | Low      | 10min  | Documented all session changes                                    |
| 17 | README line-by-line audit (#18)         | Docs          | Medium   | 15min  | Verified linter counts; fixed medium count                        |
| 18 | DOMAIN_LANGUAGE audit (#19)             | Docs          | Low      | 10min  | All ~30 terms verified current                                    |
| 19 | ARCHITECTURE.md ADR audit (#20)         | Docs          | Low      | 5min   | No file exists; ADRs in docs/adr/ (6 files)                       |
| 20 | --json-errors exit codes (#16)          | Testing       | Low      | 5min   | 0, 1, 65, 69 already tested; 75 hard to trigger                   |
| 21 | markdownlint config test (#17)          | Testing       | Low      | 5min   | CI already validates config                                       |
| 22 | Config.Clone property test (#14)        | Testing       | Low      | 5min   | Already has comprehensive deep-copy tests                         |
| 23 | golangci-lint without --fix CI (#8)     | CI/CD         | Low      | 1min   | Already runs without --fix                                        |
| 24 | nix flake check in CI (#7)              | CI/CD         | Low      | 1min   | CI already runs check + build                                     |
| 25 | Type OutputConfig.Formats (#26)         | Type safety   | Low      | 10min  | Investigated; map[string]any intentional for round-trip safety    |
| 26 | defer file.Close() audit (#40)          | Code quality  | Low      | 5min   | All in best-effort audit ledger                                   |
| 27 | Preset composition (#33)                | Features      | Low      | 5min   | Already implemented (format/house compose minimalLinters)         |
| 28 | Self-linting (#43)                      | Code quality  | Low      | 1min   | Already done (CI runs golangci-lint on own code)                  |
| 29 | Scripts fate decision (#45)             | Code quality  | Low      | 5min   | Keep as documented manual utilities                               |
| 30 | Funlen defaults decision (#47)          | Research      | Low      | 5min   | Already 200/100 house style                                       |
| 31 | Exhaustruct defaults (#48)              | Research      | Low      | 5min   | Already has curated settings                                      |
| 32 | go.mod review (#50)                     | Research      | Low      | 10min  | No banned dependencies found                                      |

---

## Remaining Items (sorted by impact)

| #  | Item                                            | Category       | Impact | Effort | Why Not Done                                         |
| -- | ----------------------------------------------- | -------------- | ------ | ------ | ---------------------------------------------------- |
| 1  | Split cmd_configure.go (586 lines) (#24)        | Refactor       | High   | 3-4h   | Largest SRP violation; high-risk mechanical split    |
| 2  | Extract typed linter constants (#25)            | Type safety    | Medium | 2-3h   | Mechanical but touches many files                    |
| 3  | Split ConfigLoader interface (#28)              | Architecture   | Medium | 2-3h   | Interface refactor with wide blast radius            |
| 4  | Consolidate ValidationError + HealthIssue (#29) | Type safety    | Medium | 1-2h   | Type consolidation across packages                   |
| 5  | Add settings key validation (#31)               | Type safety    | Medium | 2h     | Needs golangci-lint schema reference                 |
| 6  | Generate settings from JSON Schema (#30)        | Type safety    | Medium | 4-6h   | Codegen project; replaces 15 hand-maintained structs |
| 7  | Add Result type for CLI (#27)                   | Architecture   | Medium | 1-2h   | New abstraction touching all command handlers        |
| 8  | Multi-preset support (#34)                      | Feature        | Low    | 2-3h   | Changes --preset to StringSlice; needs merge logic   |
| 9  | Format --detect mode (#35)                      | Feature        | Low    | 1h     | Extends detection to format-specific linters         |
| 10 | Preset recommendation (#37)                     | Feature        | Low    | 2-3h   | Needs analysis-to-preset heuristic design            |
| 11 | Register domain message templates (#39)         | Error handling | Low    | 30min  | Deeper change touching many error sites              |
| 12 | HTML report CSS regression test (#12)           | Testing        | Low    | 1h     | Needs golden snapshot of color values                |
| 13 | CI retry logic (#10)                            | CI/CD          | Low    | 1h     | Caching already robust; marginal value               |
| 14 | Coverage-check integration test (#13)           | Testing        | Low    | 30min  | Existing unit tests cover parsing logic              |
| 15 | Run deduplicate-code skill (#41)                | Code quality   | Low    | 30min+ | Skill invocation needs dedicated session             |
| 16 | Run architecture-review skill (#42)             | Code quality   | Low    | 30min+ | Skill invocation needs dedicated session             |
| 17 | Consolidate status reports (#44)                | Code quality   | Low    | 30min  | 29 reports; non-destructive annotation needed        |

---

## Commits This Session (13 total)

1. `fix(lint): remove stale legacyerrors nolint directives`
2. `security(ci): pin all GitHub Actions to commit SHAs`
3. `build(nix): extract vendorHash and add coverage-check to Nix build`
4. `docs(references): update code organization documentation`
5. `docs: update code-organization, README, and CHANGELOG for accuracy`
6. `test(constants): enhance data integrity verification tests`
7. `feat(cli): add --no-color flag for CI and scripting output`
8. `feat(settings): add wrapcheck default settings and Dependabot automation`
9. `refactor(cli): extract HandleError function at the CLI boundary`
10. `feat(cli): enhance presets command functionality`
11. `fix(lint): resolve lint issues in presets JSON and add git-cliff config`
12. `docs(todo): update TODO_LIST with completed items and remaining work`
13. `docs(status): add coverage check and quality debt cleanup status report`

---

## Key Decisions Made

1. **OutputConfig.Formats stays `map[string]any`** — Investigated and confirmed correct for a config round-trip tool. Changing to a typed struct would risk dropping unknown user YAML fields.

2. **Preset composition already implemented** — The `format` and `house` presets already compose `minimalLinters` at the data level. Multi-preset composition (#34) would need runtime merge logic.

3. **Scripts kept as documented manual utilities** — `validate_linter_doc.sh` and `verify_linter_count.sh` serve niche manual verification needs. Porting to Go subcommands adds maintenance burden for rarely-used tools.

4. **Funlen defaults are 200/100** — Already decided in a prior session as the house style (dominant override across 160 sibling projects). Not 60/40 or 30/20.

---

## Resolution (2026-07-25, later session)

The 17 "Remaining Items" above (item #13 deduplicate-code was resolved too) were
all completed later the same day by the `2026-07-25_20-55` session. Mapping to
commits:

| #  | Item                                          | Resolution                           | Commit    |
| -- | --------------------------------------------- | ------------------------------------ | --------- |
| 1  | Split `cmd_configure.go`                      | 680 lines → 4 focused files          | `9a41447` |
| 2  | Extract typed linter constants                | `coreLinters` + `withCore()`         | `9e0e702` |
| 3  | Split `ConfigLoader` interface                | 6 focused sub-interfaces             | `39cca87` |
| 4  | Consolidate `ValidationError` + `HealthIssue` | `ToHealthIssue()` conversion         | `58fbe3c` |
| 5  | Settings key validation                       | Soft warnings at config load         | `e8f30f0` |
| 6  | Generate settings from JSON Schema            | `cmd/generate-settings` (88 structs) | `3665d79` |
| 7  | `CommandResult` type for CLI                  | Optional structured return           | `40eda4c` |
| 8  | Multi-preset support                          | `--preset a --preset b` merge        | `86ddc2d` |
| 9  | Format `--detect` mode                        | `--preset format --detect`           | `ecb3fe0` |
| 10 | Preset recommendation                         | `--recommend` flag                   | `d97237c` |
| 11 | Domain message templates                      | 27 Wix-style templates               | `03a0806` |
| 12 | HTML report CSS regression test               | Color golden-value tests             | `2c6accf` |
| 13 | CI retry logic                                | 3-attempt nix build retry            | `5c76e1c` |
| 14 | Coverage-check integration test               | End-to-end threshold tests           | `65fec5c` |
| 15 | Run deduplicate-code skill                    | 0 clone groups (clean)               | —         |
| 16 | Run architecture-review skill                 | `docs/architecture-understanding/`   | `f39f7f7` |
| 17 | Consolidate status reports                    | `docs/status/README.md` index        | `fb7c9eb` |

**Still open:** the version bump + git tag (C19 of the friction-reduction plan)
was intentionally not done — it requires a semver decision. Now tracked in
`TODO_LIST.md` (High Priority).
