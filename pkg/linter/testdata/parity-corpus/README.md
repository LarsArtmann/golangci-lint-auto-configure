# Parity Corpus

Byte-level behavior snapshots of the configure engine, captured **before** the
declarative-reconciler refactor (plan:
`docs/planning/2026-10-13_17-45_SUPERB-declarative-reconciler.md`). Every later
phase must keep this corpus green until its intentional, documented flip.

## Layout

```text
<fixture>/
  input.golangci.yml                      fixture config fed to FixConfig
  input.golangci-lint-auto-configure.yml  optional policy sidecar
  manifest.json                           env assumptions + flags (see below)
  golden/config.yml                       saved config bytes after the run
  golden/audit.jsonl                      audit entries, timestamps scrubbed
  golden/result.json                      MigrationResult summary
```

## Fixtures

| Fixture                 | Exercises                                                                  |
| ----------------------- | -------------------------------------------------------------------------- |
| `bare-minimal`          | default enable-set rendering, formatter injection, run/issues/output defaults |
| `typical`               | partial enables, respected disable (no sidecar), preserved settings, old run.go untouched |
| `deprecated-linters`    | deprecation replacement migration + settings-block porting + bad-key rename |
| `bad-settings-keys`     | KnownBadSettingsKeys self-heal (`goconst.min-length` → `min-len`)          |
| `overspecified-run-go`  | run.go rescue (repair + `rescued-run-go` audit entry)                      |
| `redundant-formatters`  | gofmt/golines redundancy removal, exclusion-rule merge + linter rename     |
| `enforced-sidecar`      | anti-gaming re-enable of unjustified disables (sidecar present)            |
| `never-enable-sidecar`  | never-enable overlay wins over enforcement; suppression audited            |
| `pragmatic`             | `--pragmatic` gate with all four noise linters enabled                     |
| `tool-own`              | this repo's own config at corpus creation; converged-config idempotency witness |
| `schema-fixture`        | every curated default at once (copy of the constants schema fixture)       |

## Manifest fields

- `priority`, `dryRun`, `forceSettings`, `pragmatic` — flags threaded to the fixer/analyzer.
- `localGoVersion` — the value `SetGoVersionProvider` returns (fixed, so goldens do not
  drift with the developer toolchain).
- `requiresToolAtLeast` — the run refuses (with a clear message) when the environment's
  golangci-lint is older; a newer binary that changes output means the goldens must be
  consciously regenerated.

## Running

```bash
go test ./pkg/linter/ -run TestParityCorpus            # parity gate
UPDATE_PARITY=1 go test ./pkg/linter/ -run TestParityCorpus   # regenerate goldens
```

Regenerated goldens must be reviewed like production output: they become the
contract every later refactor phase is judged against.
