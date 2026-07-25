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
  linter that exists in the current golangci-lint v2.

### 3. Test depth & CI robustness

- **Property-based JSON round-trip tests** — fuzz/property tests asserting
  report types survive a marshal → unmarshal cycle with no field loss.
- **HTML report snapshot/golden tests** — guard against silent templ regressions.
- **CGO / `-race` in CI** — enable the race detector in the canonical CI gate.
- **golangci-lint version pinning in CI** — match the devShell version exactly.

### 4. UX & preset ergonomics

- **Preset composition** — `format` duplicates `minimal`'s linter list; support
  composing presets (`format = minimal + formatters`, `--preset a --preset b`).
- **`--detect` for the format preset** — auto-enable `swaggo` when Swagger is
  detected, mirroring the existing project-type detection.
- **`reference+format` combined preset** — for projects that want everything.

---

## Explicit non-goals

- **Re-implementing golangci-lint** — this tool configures and optimizes
  golangci-lint; it does not replace its analysis engine.
- **A GUI** — the CLI + report outputs (HTML/JSON/SARIF) are the interface.
- **CBOR support** — report types are JSON-only; CBOR is not a target.
