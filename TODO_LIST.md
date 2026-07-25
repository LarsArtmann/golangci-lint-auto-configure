# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-25

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`; long-term ideas live in `ROADMAP.md`.

---

## High Priority

| Task                                                                                       | Impact                                       | Effort | Evidence                                                                                 |
| ------------------------------------------------------------------------------------------ | -------------------------------------------- | ------ | ---------------------------------------------------------------------------------------- |
| Split `cmd_configure.go` (586 lines, 8+ concerns) into focused files                       | High — largest SRP violation in the codebase | 3–4h   | `internal/cli/cmd_configure.go` is 586 lines; fixer was already split, configure was not |
| Extend docs-integrity test to cover ALL hardcoded counts in FEATURES.md (not just presets) | High — prevents all documentation drift      | 1h     | `pkg/constants/docs_integrity_test.go` infrastructure exists; only preset counts covered |

## Medium Priority

| Task                                                              | Impact                                                   | Effort | Evidence                                                                             |
| ----------------------------------------------------------------- | -------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------ |
| Extract linter/formatter name strings as typed `const` values     | Medium — eliminates goconst class of lint warnings       | 2–3h   | Linter names are bare strings in many files; typed `LinterName` exists but underused |
| Split the 8-method `ConfigLoader` God Object interface            | Medium — too many concerns in one interface              | 2–3h   | `pkg/types/types.go:223` — combines 6 sub-interfaces                                 |
| Consolidate `ValidationError` + `HealthIssue` (overlapping types) | Medium — type system duplication                         | 1–2h   | `pkg/types/types.go:199` and `pkg/types/validation.go:118`                           |
| Generate settings structs from golangci-lint's JSON Schema        | Medium — replaces hand-maintained structs with generated | 4–6h   | 15 typed structs in `pkg/constants/linter_settings.go` are hand-maintained           |
| Add settings key validation against golangci-lint schema at load  | Medium — catches typos in user config before runtime     | 2h     | No schema validation in `pkg/config/loader.go` today                                 |

## Low Priority

| Task                                                                   | Impact | Effort | Evidence                                                 |
| ---------------------------------------------------------------------- | ------ | ------ | -------------------------------------------------------- |
| Register domain message templates for `errorfamily.New()` constructors | Low    | 30min  | `pkg/errors/classification.go` uses bare sentinels       |
| Add `--preset a --preset b` multi-preset support                       | Low    | 2–3h   | `--preset` currently accepts one value                   |
| Add `--detect` mode for the format preset (auto-enable swaggo)         | Low    | 1h     | `--detect` exists for configure but not format-specific  |
| Add `--backup` flag decision (always-on vs opt-in)                     | Low    | 30min  | Product decision needed from user                        |
| Add preset recommendation based on project analysis                    | Low    | 2–3h   | Detection exists but doesn't feed into preset selection  |
| Add HTML report CSS regression test                                    | Low    | 1h     | Color values in `pkg/ui/styled_output.go` are unverified |
| Add CI retry logic for flaky golangci-lint cache steps                 | Low    | 1h     | Cache failures occasionally cause CI flakiness           |

## Completed This Session (2026-07-25)

These items were resolved during the comprehensive quality debt cleanup:

- ~~Adopt `HandleError` at the CLI boundary~~ — Done (`internal/cli/commands.go`)
- ~~Add `--no-color` flag~~ — Done (sets NO_COLOR=1 env var)
- ~~Type `OutputConfig.Formats`~~ — Investigated; `map[string]any` is intentional for config round-trip safety
- ~~Conventional commits/changelog automation~~ — Done (`cliff.toml` added)
- ~~Implement preset composition~~ — Done (format/house presets compose minimalLinters)
- ~~Extract ARCHITECTURE.md inline ADRs~~ — Done (no ARCHITECTURE.md exists; ADRs in `docs/adr/`)
