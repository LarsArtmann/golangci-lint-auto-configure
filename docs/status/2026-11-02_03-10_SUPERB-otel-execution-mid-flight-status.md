# Status: SUPERB Optional-OTEL Execution — Mid-Flight Snapshot

> **Scope:** The 2026-11-02 "Make everything superb + optional OTEL support"
> execution, interrupted mid-M4. Plan of record:
> `docs/planning/2026-11-02_00-29_SUPERB-optional-otel-and-audit-coverage.md`.
> Point-in-time snapshot — stale by design.

## a) FULLY DONE

1. **Comprehensive plan** (pareto 1%/4%/20%/rest, master tasks M1–M10 at
   30–100 min, fine tasks F1–F46 at ≤12 min, both as sorted tables, mermaid
   execution graph, errata, locked design decisions, verification gates).
2. **Errata verified live before building anything:** the "75-vs-1 exit-code
   inconsistency" from the earlier self-review was a false alarm (75 = the
   unknown-flag error path; `ErrChangesNeeded` → Conflict → 1 is correct);
   the 61 wrapcheck findings are version skew (CI lint job green on pinned
   v2.14, local v2.15 noise); the red-master CI root cause is a one-word
   provenance mismatch in a generated file.
3. **M1 — master green fix:** regenerated `linter_settings_generated.go`
   (provenance header v2.15.0 → v2.14.0, exactly what the CI drift guard
   regenerates); schema fixture verified clean; picked up by the auto-commit
   daemon.
4. **M2 — `pkg/telemetry`:** strictly opt-in tracing (gate: `--otel` flag OR
   `GOLANGCI_LINT_AUTO_CONFIGURE_OTEL` env), OTLP/HTTP exporter reading
   standard `OTEL_EXPORTER_OTLP_*` envs, batch processor, resource with
   service name/version, best-effort warn-and-degrade (audit-ledger
   philosophy), 5s bounded shutdown. BDD specs green, including a wire-level
   e2e: an httptest OTLP receiver asserts the exported POST arrives as
   protobuf, and an inert-session spec proves zero egress without Setup.
   Deps added: otel v1.47.0 (API + SDK + otlptracehttp; grpc/genproto arrive
   transitively). `go 1.27` floor verified intact after tidy.
5. **M3 — CLI wiring:** `--otel` persistent flag; telemetry `Session`
   lifecycle in `Execute` (Setup inside a chained `PersistentPreRun` where
   the command name is known; Finish after fang returns, recording error
   family/code on the root span); `NewRootCommand(flags, session)`; in-process
   test caller updated. Build green, gofmt clean.
6. **M4 (half) — engine spans:** `FixConfig` span in `pkg/linter/fixer.go`
   (config.path, priority, dry_run, changes.total) with a package tracer
   constant; noop-by-default for embedded consumers (API-only import).

## b) PARTIALLY DONE

1. **M4 migrate span pending** — ctx threading planned through
   runMigrate → executeMigration → runMigrator (ADR-013 style).
2. **Verification debt:** pkg/telemetry suite + `go build ./...` green
   (BUILD_EXIT=0 verified at 03:10), but the full test/lint gates (M9) have
   not run since the M3/M4 edits.
3. **vendorHash.nix is stale** relative to the new go.mod/go.sum — any
   `nix build` before running `scripts/vendorhash-guard.sh --fix` will fail
   on the hash mismatch. Known, sequenced (F41), currently an open window.
4. **Parallel-session co-tenancy:** another session is actively editing
   `pkg/migration/*` (incl. a deleted internal test), `.golangci.yml`, and
   many archived docs. Nothing of mine touches those files; my commits stay
   pathspec-scoped. Noted, not yet stress-tested by an actual commit.

## c) NOT STARTED

- **M5** audit coverage: `audit.EnvVarDisable` export, `config-migrated` +
  `preset-applied` actions, ledger wiring for migrate + preset paths,
  `client.Options.Recorder` injection, plus three specs.
- **M6** flag errors → Rejection/exit 1 via `SetFlagErrorFunc` + spec.
- **M7** installhook never-mutates-config spec; ErrChangesNeeded exit pin
  (existing `exit_code_test.go` not yet inspected).
- **M8** docs: SECURITY (opt-in telemetry paragraph), README (OTEL section),
  FEATURES, CHANGELOG, AGENTS (telemetry gotcha + gotcha-17 caveat refresh +
  lint-skew note), ROADMAP (OTEL decision), TODO_LIST (row partial close).
- **M9** full gates: `ginkgo -r -p -race` full suite, golangci-lint baseline
  diff vs CURRENT master (the parallel session moved it), vendorhash-guard
  --fix, `nix build`, dogfood `configure --check`.
- **M10** curated commits (user asked for VERY DETAILED messages), push
  (explicitly requested), CI green verification.

## d) TOTALLY FUCKED UP

1. **Nothing broken, but two live risk windows:** (i) `pkg/linter/fixer.go`
   carried uncompiled OTEL edits across the interruption — only just proven
   green; (ii) go.mod changed while vendorHash.nix is stale — nix build is
   red-on-demand until F41. Both sequenced deliberately, but they are open.
2. **The earlier self-review's exit-code claim was wrong** and its status
   report (`2026-11-02_00-18_…self-review.md` §b.5/§f.6) still carries the
   false "75-vs-1 inconsistency" — correction lives only in the plan's
   errata so far. Needs a docs-health ANNOTATE pass, not yet done.
3. **Plan deviation unrecorded:** F13 (recursive RunE wrap) was replaced by
   the simpler PersistentPreRun chaining during M3 — the plan file still
   describes the old approach. Must be amended when execution resumes.
4. **LSP noise repeatedly polluted edit feedback** (transient "can't load
   config: unsupported version 1", stale wsl/wrapcheck warnings) during
   concurrent edits; falling back to `go build`/repo-config lint was right,
   but cost cycles double-checking phantoms.
5. **One gofmt violation slipped into M3** (misindented `finishCtx` block) —
   caught immediately by `gofmt -l`, fixed. Minor, still sloppy.
6. **M1 rode the auto-commit daemon** instead of a curated detailed commit —
   the user explicitly asked for detailed messages; remaining milestones
   must not repeat this.

## e) WHAT WE SHOULD IMPROVE

1. **Dependency-surface honesty:** otlptracehttp drags grpc + genproto
   transitively; before docs land (M8), decide-and-document whether that is
   accepted or whether the exporter set should be trimmed.
2. **Commit cadence on big features:** curated commit per M-chunk instead of
   letting the daemon swallow milestone work into heuristic messages.
3. **Re-baseline before gates:** the lint baseline must be re-taken against
   CURRENT master (the parallel session changed `.golangci.yml` and
   pkg/migration), not this morning's snapshot.
4. **LSP flakiness protocol:** when golangci-lint LSP wedges under
   concurrent edits, `lsp_restart` early instead of re-reading diagnostics.

## f) Things to get done next (remaining plan; live source: plan file F16–F46)

| #  | Task (≤12 min each)                                                                          | Master |
| -- | --------------------------------------------------------------------------------------------- | ------ |
| 1  | Migrate span: thread ctx runMigrate→executeMigration→runMigrator; attrs dry_run/skip_validation/fixes_applied | M4 |
| 2  | Lint-clean pass over all new span/telemetry code (wsl, wrapcheck)                              | M4     |
| 3  | Export `audit.EnvVarDisable`; alias the cli const to it                                        | M5     |
| 4  | Add `ActionConfigMigrated` + `ActionPresetApplied` actions with doc comments                   | M5     |
| 5  | Wire ledger into migrate command; emit `config-migrated` on non-dry success                    | M5     |
| 6  | Wire ledger into preset path; emit `preset-applied` on non-dry save                            | M5     |
| 7  | `client.Options.Recorder` + `SetLedger` when non-nil + privacy/cycle-detection doc comment     | M5     |
| 8  | Spec: migrate non-dry writes exactly one `config-migrated`; dry-run none                        | M5    |
| 9  | Spec: preset run writes `preset-applied`; check-mode preset writes none                         | M5    |
| 10 | Spec: client with Recorder emits entries on apply                                              | M5     |
| 11 | `rootCmd.SetFlagErrorFunc` → `WrapRejection("cli.flags")` — usage errors exit 1                | M6     |
| 12 | Spec: unknown flag exits 1 with family=rejection (was 75)                                       | M6    |
| 13 | Spec: installhook writes only the git hook, never `.golangci.yml`                               | M7    |
| 14 | Inspect exit_code_test.go; pin ErrChangesNeeded→1 if unpinned                                  | M7     |
| 15 | SECURITY.md telemetry paragraph (strictly opt-in, standard envs, default local-only)           | M8     |
| 16 | README "Optional OpenTelemetry" section + example                                               | M8    |
| 17 | FEATURES.md: OTEL row + audit-coverage rows                                                     | M8    |
| 18 | CHANGELOG: Added (OTEL, audit coverage, recorder) / Fixed (flag-error exit 1)                  | M8     |
| 19 | AGENTS.md: telemetry gotcha + gotcha-17 caveat refresh + lint-skew note                         | M8    |
| 20 | ROADMAP OTEL decision line; TODO_LIST coverage-row partial close                                | M8    |
| 21 | Focused suites: telemetry, cli, client, audit                                                    | M9   |
| 22 | Full `ginkgo -r -p -race` + cmd package tests                                                    | M9    |
| 23 | golangci-lint baseline diff vs CURRENT master = zero new findings                                | M9    |
| 24 | `scripts/vendorhash-guard.sh --fix`; commit vendorHash.nix WITH go.mod/go.sum                   | M9    |
| 25 | `nix build` + dogfood `configure --check` exit 0                                                 | M9    |
| 26 | `nix fmt`; final `git status` review of MY files only                                            | M9    |
| 27 | Curated commits per chunk, pathspec-scoped, very detailed messages                              | M10   |
| 28 | `git push` (explicitly requested)                                                                | M10   |
| 29 | `gh run` verify: ci.yml green incl. schema gate                                                  | M10   |
| 30 | Amend plan file: record F13→PersistentPreRun deviation                                           | new   |
| 31 | docs-health ANNOTATE the 00-18 self-review's false exit-code claim                               | new   |
| 32 | Decide/document the grpc transitive dependency acceptance                                        | new   |

(routed: the audit-coverage row already lives in TODO_LIST.md; nothing new to
harvest from this snapshot beyond the plan file itself.)

## g) Questions I can NOT figure out myself

1. **OTLP transport scope:** HTTP-only (as built) or also wire OTLP/gRPC for
   collector environments that only speak gRPC? HTTP keeps the config surface
   minimal; gRPC would roughly double the exporter wiring.
2. **Security posture depth:** is default-off + explicit opt-in sufficient,
   or do you additionally want telemetry hard-blocked in CI/root contexts
   unless force-enabled? That paranoia level is policy, not engineering.
3. **Parallel-session work in the push path:** the other session has
   uncommitted `pkg/migration` changes (incl. a deleted internal test) and
   touched `.golangci.yml`; the auto-commit daemon will commit them and my
   requested push will publish them. Intended to ship together, or should the
   push wait until that session lands its own curated commit?

---

**Session files so far:** `docs/planning/2026-11-02_00-29_SUPERB-…md`,
`pkg/telemetry/{telemetry.go,telemetry_test.go,suite_test.go}` (new),
`pkg/constants/linter_settings_generated.go` (regenerated),
`internal/cli/{commands.go,flags.go,cmd_inprocess_test.go}`,
`pkg/linter/fixer.go`, `go.mod`/`go.sum`.
