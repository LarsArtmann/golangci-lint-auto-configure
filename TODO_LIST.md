# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-25

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`; long-term ideas live in `ROADMAP.md`.

---

## High Priority

| Task                                                                                       | Impact                                       | Effort | Evidence                                                                                 |
| ------------------------------------------------------------------------------------------ | -------------------------------------------- | ------ | ---------------------------------------------------------------------------------------- |
| Split `cmd_configure.go` (581 lines, 8+ concerns) into focused files                       | High — largest SRP violation in the codebase | 3–4h   | `internal/cli/cmd_configure.go` is 581 lines; fixer was already split, configure was not |
| Extend docs-integrity test to cover ALL hardcoded counts in FEATURES.md (not just presets) | High — prevents all documentation drift      | 1h     | `pkg/constants/docs_integrity_test.go` infrastructure exists; only preset counts covered |

## Medium Priority

| Task                                                                | Impact                                                   | Effort | Evidence                                                                             |
| ------------------------------------------------------------------- | -------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------ |
| Adopt `HandleError` at the CLI boundary (replaces slog)             | Medium — structured error output at the system boundary  | 1–2h   | No `HandleError` in `internal/cli/` today                                            |
| Add `--no-color` flag for CI/scripting output                       | Medium — enables accurate plain-text examples in docs    | 1h     | README example output is hand-simplified because real output has ANSI codes          |
| Extract ARCHITECTURE.md inline ADRs to individual `docs/adr/` files | Medium — ADRs live in two places (split brain)           | 1–2h   | `docs/adr/` has 5 files; ARCHITECTURE.md has 8 inline ADRs                           |
| Extract linter/formatter name strings as typed `const` values       | Medium — eliminates goconst class of lint warnings       | 2–3h   | Linter names are bare strings in many files; typed `LinterName` exists but underused |
| Type `OutputConfig.Formats` (only two known shapes: `format: path`) | Medium — makes config parsing type-safe                  | 1h     | `pkg/types/config_types.go` uses `map[string]any` for formats                        |
| Split the 8-method `ConfigLoader` God Object interface              | Medium — too many concerns in one interface              | 2–3h   | `pkg/types/types.go:223` — combines 6 sub-interfaces                                 |
| Consolidate `ValidationError` + `HealthIssue` (overlapping types)   | Medium — type system duplication                         | 1–2h   | `pkg/types/types.go:199` and `pkg/types/validation.go:118`                           |
| Generate settings structs from golangci-lint's JSON Schema          | Medium — replaces hand-maintained structs with generated | 4–6h   | 13 typed structs in `pkg/constants/linter_settings.go` are hand-maintained           |

## Low Priority

| Task                                                                   | Impact | Effort | Evidence                                                  |
| ---------------------------------------------------------------------- | ------ | ------ | --------------------------------------------------------- |
| Register domain message templates for `errorfamily.New()` constructors | Low    | 30min  | `pkg/errors/classification.go` uses bare sentinels        |
| Implement preset composition (`format = minimal + formatters`)         | Low    | 1–2h   | `presets.go` format preset reuses minimalLinters directly |
| Add `--preset a --preset b` multi-preset support                       | Low    | 2–3h   | `--preset` currently accepts one value                    |
| Add `--detect` mode for the format preset (auto-enable swaggo)         | Low    | 1h     | `--detect` exists for configure but not format-specific   |
| Add `--backup` flag decision (always-on vs opt-in)                     | Low    | 30min  | Product decision needed from user                         |
| Conventional-commits-to-changelog automation (`git-cliff`)             | Low    | 2h     | CHANGELOG is hand-maintained                              |
