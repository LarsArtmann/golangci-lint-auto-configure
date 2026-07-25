# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-07-25

Short- and mid-term actionable work. Completed items live in `CHANGELOG.md`; long-term ideas live in `ROADMAP.md`.

---

## High Priority

| Task | Impact | Effort | Evidence |
| ---- | ------ | ------ | -------- |
| Add tests for the audit/policy code paths (`cmd_audit.go`, `pkg/linter/fixer_enforce.go`, `newRunLedger`) | High — security-critical, currently ZERO tests | 2–3h | `internal/cli/cmd_audit.go`, `pkg/linter/fixer_enforce.go` have no `*_test.go` |
| Add exit-code integration tests for Infrastructure (69) and Corruption (65) | High | 1h | `internal/cli/exit_code_test.go` only covers exit 0 + exit 1 (invalid priority) |
| Increase CLI integration test coverage (~11%) | High | ongoing | integration tests exec the binary, which doesn't count toward `go test` coverage |

## Medium Priority

| Task | Impact | Effort | Evidence |
| ---- | ------ | ------ | -------- |
| Add a separate `golangci-lint run` (no `--fix`) CI step | Medium — catches unfixable issues BuildFlow's `--fix` mode silently swallows | 30min | BuildFlow false-negative documented in `docs/status/2026-07-10_17-38_*` §d |
| Convert `scripts/coverage-check.sh` to a Go test | Medium — more portable, testable | 1h | `scripts/coverage-check.sh` is still bash |
| Adopt `HandleError` at the CLI boundary (replaces slog) | Medium | 1–2h | no `HandleError` in `internal/cli/` today |
| Add a `funcorder` test gap closure | Medium | 30min | flagged in `docs/status/2026-07-10_14-13_*` |

## Low Priority

| Task | Impact | Effort | Evidence |
| ---- | ------ | ------ | -------- |
| Register domain message templates when adopting `errorfamily.New()` constructors | Low | 30min | `pkg/errors/classification.go` |
