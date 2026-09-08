# Roadmap

Long-term direction and raw ideas not yet refined into bounded tasks. Actionable,
scoped work lives in `TODO_LIST.md`; shipped features live in `FEATURES.md`.

---

## Themes

### 1. Config propagation & round-trip fidelity

The biggest open architectural questions are about how the tool's evolving
defaults reach **existing** configs, not new ones.

- **`RuleKey()` merge strategy** — adding linters to `DefaultExclusionRules`
  only helps new/regenerated configs. The dedup key is `Path|Text|Source`, so
  88 machine-generated sibling configs are stuck on the old exclusion list until
  their rule is deleted and re-injected. A "rule merge" migration would
  propagate updates without rewriting user-trimmed rules.
- **YAML indentation preservation** — the tool reformats 2-space→4-space
  aggressively across the entire file, producing massive whitespace-only diffs
  that obscure real changes. The loader should preserve the input style or match it.
- **`--force-settings` / settings refresh** — the idempotency guarantee (never
  overwrite existing settings) is a trap for self-config: the repo's own
  `.golangci.yml` can never be cleaned up by re-running the tool. A
  settings-refresh path is needed.

### 2. CLI layer testability

- **`CommandContext` struct** — 9 package-level variables in
  `internal/cli/commands.go` (`priority`, `dryRun`, `verbose`, `quiet`,
  `configPath`, `noAudit`, `pragmatic`, `showDiff`, `detectedExtraFormatters`)
  act as implicit context. Extracting an explicit struct would enable parallel
  command execution and remove cross-test state leaks. (From the architecture
  review's #1 concern.)
- **Narrow interface adoption** — some CLI functions accept the concrete
  `*config.Loader` while others accept the `presetConfigLoader` interface.
  Standardising on the `ConfigReader`/`ConfigWriter` sub-interfaces everywhere
  would make every code path mockable.

### 3. Validation & schema alignment (ongoing)

- **`LinterMinVersions` accuracy audit** — cross-check `since` values against
  upstream golangci-lint release notes (the codegen generator helps but the
  hand-curated version gates still need periodic verification).
- **`DeprecatedLinters` target audit** — verify every replacement points to a
  linter that exists in the current golangci-lint v2. (Partially covered by
  `data_integrity_test.go`, but not against upstream accuracy.)

### 4. Build automation & process maturity

- **Auto-commit hook improvement** — the daemon mixes file types into generic
  `docs:` commits and swept up `.go`/`.yml`/`.nix` changes into doc commits.
  Either scope it to file-type-specific messages or make it refuse unexpected
  file types.
- **Status report lifecycle** — status reports in `docs/status/` and
  `docs/archive/status/` accumulate. A `docs/status/README.md` index now exists;
  establish an archive cadence (e.g. quarterly, or keep only the latest N/month).
- **Full `nix flake check` in CI** — currently only `--no-build` runs in CI
  (time constraints; a local full `nix build` was last validated 2026-09-08).
  Running the full check in CI would validate the hermetic build path on every push.

### 5. Error handling governance (largely complete)

- **Swallowed-error audit** — a prior pass found only 2 benign `defer Close()`
  sites; no log-and-continue anti-patterns. A deeper periodic audit would keep
  this honest as the codebase grows. Exit-code governance is already
  best-in-class: a single `os.Exit()` in `Main()` via `errorfamily.ExitCode()`.

---

## Explicit non-goals

- **Re-implementing golangci-lint** — this tool configures and optimizes
  golangci-lint; it does not replace its analysis engine.
- **A GUI** — the CLI + report outputs (HTML/JSON/SARIF) are the interface.
- **CBOR support** — report types are JSON-only; CBOR is not a target.
- **Active v1 config feature development** — v1 config support is
  maintenance-only (0 live v1 configs across 160 sibling projects, 99.4% are
  v2). The v1→v2 migrator (`migrate` subcommand) is kept functional but no new
  v1 features will be added. Bug fixes only.
- **Promoting the sidecar policy file** — `.golangci-lint-auto-configure.yml`
  has 0 adoption across 160 projects. The feature stays functional (backward
  compatible) but will not be actively promoted. The `--pragmatic` flag is the
  preferred friction-reduction mechanism going forward.
- **A typed `OutputConfig.Formats`** — investigated and deliberately kept as
  `map[string]any`. This tool round-trips arbitrary user config; a typed struct
  would risk dropping unknown YAML fields. The flexibility is the feature.
