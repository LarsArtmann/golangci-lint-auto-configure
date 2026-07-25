# Roadmap

Long-term direction and raw ideas not yet refined into bounded tasks. Actionable,
scoped work lives in `TODO_LIST.md`; shipped features live in `FEATURES.md`.

---

## Themes

### 1. Type safety & data-model rigor

- **Linter name constants** — linter/formatter names are bare strings across 6+
  files in `pkg/constants/`. Extracting them as typed `const` values would
  eliminate the entire `goconst` class of warnings and prevent typos.
- **Type `OutputConfig.Formats`** — currently `map[string]any`; only two known
  shapes exist (`format: path`). A dedicated type would make invalid states
  unrepresentable.
- **Result type for CLI commands** — commands return bare `error`; a structured
  `Result` could carry warnings, counts, and findings alongside the error.
- **Generate settings structs from golangci-lint's JSON Schema** — replace the
  hand-maintained typed settings structs with generated ones to stay aligned
  with upstream automatically.

### 2. Validation & schema alignment

- **Settings schema validation** — validate user-provided settings keys against
  golangci-lint's schema at config load time (JSON Schema or key allowlist).
- **`LinterMinVersions` accuracy audit** — cross-check `since` values against
  upstream golangci-lint release notes.
- **`DeprecatedLinters` target audit** — verify every replacement points to a
  linter that exists in the current golangci-lint v2. (Partially covered:
  `data_integrity_test.go` checks existence in `LinterPriorities`, but not
  accuracy against upstream.)

### 3. UX & preset ergonomics

- **Preset composition** — `format` duplicates `minimal`'s linter list; support
  composing presets (`format = minimal + formatters`, `--preset a --preset b`).
- **`--detect` for the format preset** — auto-enable `swaggo` when Swagger is
  detected, mirroring the existing project-type detection.
- **`reference+format` combined preset** — for projects that want everything.
- **`--backup` flag decision** — always-on backup vs opt-in (product decision).

### 4. Build automation & CI maturity

- **CHANGELOG automation** — adopt `git-cliff` or similar to auto-generate
  version sections from conventional commits at tag time. Eliminates the class
  of "missing CHANGELOG entry" problems.
- **Auto-commit hook improvement** — the auto-commit daemon mixes file types
  into generic "docs:" commits. Either scope it to file-type-specific messages
  or make it refuse unexpected file types.
- **Status report lifecycle** — 100+ status reports in `docs/status/` and
  `docs/archive/status/` are accumulating. Establish a convention: archive
  quarterly, or keep only the latest N per month.

### 5. Error handling governance

- **Swallowed-error audit** — 20+ sites identified in prior reports where errors
  are logged but not propagated. Systematic audit + fix pass.
- **Error-code governance registry** — ~40 ad-hoc exit/error codes exist across
  CLI commands with no central registry or test. Build a governance table.

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
