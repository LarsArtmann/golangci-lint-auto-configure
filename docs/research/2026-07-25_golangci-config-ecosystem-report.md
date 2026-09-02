# `.golangci.yml` Ecosystem Report — Usage & AI-Model Preferences

**Date:** 2026-07-25
**Scope:** every `.golangci.*` config **outside** this repo, under `/home/lars/projects/` (sibling projects only; the current `golangci-lint-auto-configure` repo is excluded from counts).
**Method:** parsed all configs with PyYAML + ripgrepped `//nolint` directives across all non-vendored `.go` files, then cross-referenced.

---

## 1. Headline numbers

| Metric                                                 | Value                                                          |
| ------------------------------------------------------ | -------------------------------------------------------------- |
| Total `.golangci.*` files found                        | **576**                                                        |
| First-party project configs (no `vendor/`, no backups) | **160**                                                        |
| Vendored (third-party) configs under `vendor/`         | ~416                                                           |
| Text mentions of `.golangci.{yml,yaml}` in docs/code   | **2 030 files** (mostly `docs/status/*` reports + `AGENTS.md`) |
| Policy sidecars (`.golangci-lint-auto-configure.yml`)  | **0**                                                          |
| Total `//nolint:` directives in non-test Go code       | **5 201**                                                      |

---

## 2. How they are used

### 2.1 Format adoption: v2 is effectively universal

- **159 / 160 first-party configs are golangci-lint v2** (`version: "2"`) — **99.4 %**.
- The **single v1 holdout** is `archived/website-holger-hahn/.golangci.yml` (a dead project). There are **zero live v1 configs**.
- Takeaway: the v1→v2 migration is **complete** across this ecosystem. Any tooling that still defaults to emitting v1 is stale.

### 2.2 Opt-in model, not presets

- `linters.default`: **`none` in 95.6 %** (153), `standard` 5, `all` 2.
- Nobody relies on golangci-lint's built-in presets. Every project lists its linters explicitly under `linters.enable`.

### 2.3 The configs are overwhelmingly machine-generated

This is the single strongest signal. The fingerprints line up exactly with what `golangci-lint-auto-configure` emits:

- **88 configs share the byte-identical `enable` list of 109 linters.** No human hand-tunes 109 linters identically 88 times.
- **128 configs use the identical formatter set `{gci, gofumpt, goimports, golines}`.**
- **128 configs carry the exact issues pair `max-issues-per-linter: 50`, `max-same-issues: 10`** — precisely the values this tool's fixer injects (`AGENTS.md` gotcha #6).
- **101 configs disable nothing; 50 disable exactly `{noinlineerr}`** — matching `constants.DisabledLinters`.
- `BuildFlow/flake.nix` literally pulls `golangci-lint-auto-configure` in as a flake input (`v0.5.0`) and rewires `tools/go.mod` — direct confirmation of the generation path.

### 2.4 The "house style" stack

Consistent across ~150 projects:

```yaml
version: "2"
run:
  timeout: 5m          # 94.3 % (150/159)
  go: 1.26.4
  build-tags: [goexperiment.arenas, goexperiment.jsonv2, ...]
  allow-parallel-runners: true
  allow-serial-runners: true
  issues-exit-code: 1
linters:
  default: none
  enable: [ ... ~109 linters ... ]
  settings:                       # the auto-injected safe defaults (AGENTS.md #7)
    revive:     { rules: [{disabled: exported}, {disabled: package-comments}] }
    exhaustruct: { exclude: [...project-specific...] }
    cyclop:     { max-complexity: 12 }
    gocognit:   { min-complexity: 70 }
    funlen:     { lines: 200, statements: 100 }
    depguard / ireturn / makezero / varnamelen / gomoddirectives / testifylint / ginkgolinter ...
formatters:
  enable: [gci, gofumpt, goimports, golines]   # 128 configs
  settings: { golines: { ... }, gci: { ... } }
issues:
  max-issues-per-linter: 50
  max-same-issues: 10
```

### 2.5 Size & divergence

- Median config = **256 lines**; mean 303; max **1 568** (`auto-deduplicate`, inflated by a ~460-line `settings` block + 83 exclusion rules); min **9** (`cqrs-htmx/examples` — a deliberate minimal `{govet, staticcheck, unused}` config).
- Configs diverge mainly through **project-specific `exhaustruct.exclude` lists**, **`varnamelen.ignore-names`** (e.g. `[err, ok, tt, t, i, m, g]`), and **`linters.exclusions.rules`**.

### 2.6 How they are invoked

- **50 / 50 `flake.nix` lint steps** run `golangci-lint run ./...` (the `lint` output / `nix fmt`-adjacent check).
- Version is pinned per-flake (no shared lock across projects).
- `output.formats` is **unset in 100 %** — everyone uses the default text output via CLI, never a configured reporter.
- **0 GitHub-Actions-native references** to the config path were found — linting lives in Nix, not in standalone `.github/workflows` lint jobs.

---

## 3. What AI Models DO NOT like (friction evidence)

The strongest, least-opinionated evidence is the **`//nolint` directive density** — i.e. how often the code author (mostly AI-assisted sessions) suppresses a linter relative to how widely it is enabled. **Friction ratio = nolint-count ÷ enable-count.** Higher = more pain per adoption.

| Linter             | enabled in | `//nolint` | friction | verdict                                                                                                                              |
| ------------------ | ---------: | ---------: | -------: | ------------------------------------------------------------------------------------------------------------------------------------ |
| `exhaustruct`      |        146 |    **952** |      6.5 | **Most-hated.** Forces exhaustive struct literals on every `http.Server{}`, `Cmd{}`, etc. AI emits floods of `//nolint:exhaustruct`. |
| `gochecknoglobals` |        153 |    **773** |      5.1 | Fights standard Go patterns (package-level registries, sentinels, `var ErrX = …`). Constantly suppressed.                            |
| `gosec`            |        154 |    **589** |      3.8 | Security linter with many false positives (hardcoded creds, weak crypto on test fixtures). Heavily nolinted.                         |
| `errcheck`         |        143 |    **443** |      3.1 | Legit but noisy on `defer x.Close()`, `fmt.Fprint`, ignored returns.                                                                 |
| `wrapcheck`        |        150 |        265 |      1.8 | Demands every returned error be wrapped; tedious in glue code.                                                                       |
| `ireturn`          |        136 |        171 |      1.3 | "Don't return interfaces" — conflicts with common Go API design.                                                                     |
| `recvcheck`        |        148 |        147 |     0.99 | Receiver type-naming nits.                                                                                                           |
| `contextcheck`     |        149 |        139 |     0.93 | Misfires on context plumbing in handlers.                                                                                            |
| `exhaustive`       |        154 |        129 |     0.84 | Exhaustive switch/enums — noisy on intentionally-default switches.                                                                   |
| `funlen`           |        151 |        126 |     0.83 | Function length; fights table-driven tests and long constructors.                                                                    |
| `forbidigo`        |        138 |        111 |     0.80 | Forbidden identifiers (e.g. `fmt.Println`) — suppressed in main/scripts.                                                             |

> Small-sample artifact: `lll` shows friction 20.0 because it is enabled in only 1 config but accrued 20 nolints there — not ecosystem-wide. Exclude it from general conclusions.

**Pattern:** the linters AI models rebel against are the **style/boilerplate-enforcing** ones (`exhaustruct`, `gochecknoglobals`, `ireturn`, `recvcheck`, `wrapcheck`) and the **high-false-positive** ones (`gosec`). They don't resist linters that catch real defects.

### Side effects AI models produce under these linters

1. **Suppress-spam:** `//nolint:exhaustruct` on nearly every struct literal in glue code (952 occurrences).
2. **Global-var gymnastics:** moving package-level vars into `init()` or `var _ = …` tricks to appease `gochecknoglobals`.
3. **Wrap-everything noise:** `wrapcheck` pressure produces shallow `fmt.Errorf("…: %w", err)` chains that add no information.
4. **Sidecar abandonment:** the disable-justification sidecar (`.golangci-lint-auto-configure.yml`) exists in **0 projects** — the anti-gaming enforcement feature is unused in practice.

---

## 4. What AI Models DO like (zero-friction linters)

These are enabled in **~150 configs each and receive ZERO `//nolint` directives** — they catch real bugs and never get in the way:

`copyloopvar`, `errorlint`, `intrange`, `nakedret`, `nilnesserr`, `durationcheck`, `gochecksumtype`, `sloglint`, `wastedassign`, `loggercheck`, `mirror`, `protogetter`, `reassign`, `ginkgolinter`, `unconvert`, `unparam`, `gosec` (painful but kept), `misspell`.

**Pattern:** models happily accept linters that are **mechanical correctness checks** (loop-var capture, error wrapping/lint, nil-flow, range/int conversions, slog usage, duplicate assignments) — they almost never suppress them.

---

## 5. Concrete recommendations (for the auto-configure tool)

1. **Reconsider enabling `exhaustruct` by default.** It is the #1 source of AI nolint-spam (6.5 friction). Either keep it off the default enable-list or auto-populate a generous `exhaustruct.exclude` for stdlib (`os/exec.Cmd`, `net/http.Server`, `net/http.Request`, `time.Ticker`, …) — most projects already hand-curate exactly this list.
2. **`gochecknoglobals` + `ireturn` + `wrapcheck`** are the next-tier friction generators. Worth offering a "strict vs. pragmatic" preset so pragmatic projects aren't pushed into 500+ nolints.
3. **The issues pair `(50, 10)` is already de-facto standard** — keep injecting it (gotcha #6 is validated by 128 configs).
4. **The sidecar policy feature has 0 adoption.** Either promote it (docs/defaults) or de-emphasize it — currently it is dead surface area.
5. **`gosec` needs curated ignores.** 589 nolints (mostly test fixtures / hardcoded dev secrets) suggest the tool could auto-add a test-path exclusion or a `gosec.excludes` for common benign findings.
6. **Formatter quadruple `{gci, gofumpt, goimports, golines}`** is clearly the winning house stack — lock it in as the single recommended formatter preset.
7. **v1 support can move to maintenance-only.** Live v1 configs = 0. Migration effort is better spent on v2 ergonomics.

---

## 6. Methodology & reproducibility

- **Configs located** via `find /home/lars/projects -name ".golangci.*"` (the `glob` tool skips dotfiles, so `find` was required).
- **Parsed** with `python3` + PyYAML 6.0.3; **0 parse errors** across 160 configs.
- **nolint extraction** via ripgrep `//\s*nolint:\s*[a-z0-9_,\s]+` over non-vendored, non-test `.go` files (5 201 total directives).
- **Vendor configs excluded** from all counts (they are third-party libraries like `spf13/cobra`, `samber/lo`, `charm.land/*` and would skew the "house style" signal).
- Analysis scripts: `/tmp/analyze_golangci.py`, `/tmp/analyze_clusters.py`, `/tmp/friction.py`.

---

## Actions taken (2026-07-25)

This report's §5 recommendations drove the friction-reduction Pareto plan
(`docs/planning/2026-07-25_07-56_*`), which was executed the same day. Outcome
per recommendation:

| # | Recommendation                                               | Action taken                                                                                                        |
| - | ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------- |
| 1 | Reconsider `exhaustruct` default; auto-populate excludes     | ✅ Kept enabled; `ExhaustructSettings.Exclude` expanded to 14 stdlib structs. `--pragmatic` drops it on demand.     |
| 2 | "strict vs pragmatic" for gochecknoglobals/ireturn/wrapcheck | ✅ `--pragmatic` flag drops the 5 highest-noise linters from the dynamic enable set. Defaults unchanged.            |
| 3 | Keep injecting issues pair `(50, 10)`                        | ✅ Validated (128/160 configs); unchanged.                                                                          |
| 4 | Sidecar policy: promote or de-emphasize                      | ✅ De-emphasized (0 adoption; ROADMAP non-goal; `--pragmatic` is the preferred mechanism).                          |
| 5 | `gosec` curated ignores / test-path exclusion                | ✅ `GosecSettings` with G304/G115 excludes; gosec added to `_test.go` exclusions. (G104 later removed — too broad.) |
| 6 | Lock formatter quadruple as a preset                         | ✅ `house` preset = `{gci, gofumpt, goimports, golines}`; `CoreFormatters` aligned to the same 4.                   |
| 7 | v1 support → maintenance-only                                | ✅ Declared in ROADMAP non-goals + AGENTS (0 live v1 configs).                                                      |

**Measurement:** before/after nolint deltas are recorded in
`docs/research/validation-delta.md` (errcheck −27.3%, gosec −23.7%,
exhaustruct −2.3% — the exhaustruct target was unrealistic because 95.7% of its
nolints target project-specific domain types, not stdlib structs).

**Open follow-up:** the `RuleKey()` dedup means these default improvements only
reach **new** configs; the 88 machine-generated sibling configs keep the old
lists until re-injected (see ROADMAP "Config propagation & round-trip fidelity").
