# Roadmap

Long-term direction and raw ideas not yet refined into bounded tasks. Actionable,
scoped work lives in `TODO_LIST.md`; shipped features live in `FEATURES.md`.

---

## Themes

### 1. Validation & schema alignment (ongoing)

- **`LinterMinVersions` accuracy audit** — cross-check `since` values against
  upstream golangci-lint release notes (the codegen generator helps but the
  hand-curated version gates still need periodic verification).
- **`DeprecatedLinters` target audit** — verify every replacement points to a
  linter that exists in the current golangci-lint v2. (Partially covered by
  `data_integrity_test.go`, but not against upstream accuracy.)
- **Schema-version sync as a discipline** — the `min-length` incident
  (2026-09-11) showed the gap between the generator's schema snapshot (v2.13)
  and the tool's advertised minimum (v2.10.1). The concrete CI gate lives in
  TODO_LIST; the durable idea is treating "every injected default must verify
  against the minimum supported golangci-lint" as an invariant, not a one-off fix.

### 2. Config propagation & settings refresh

- **Settings refresh scope** — `--force-settings` covers `configure`; the
  detect/recommend paths do not accept it yet, so analysis-only runs can still
  surface stale settings they cannot fix.
- **Settings-key exhaustiveness** — unknown-key soft warnings exist at load
  time; a complementary "known keys the tool *should* manage but doesn't" audit
  (driven by `cmd/generate-settings` output) would catch silently-unmanaged
  settings.

### 3. CLI layer testability

- **Narrow interface adoption** — `CommandResult` + the `Flags` struct landed
  (2026-07-26); some CLI functions still accept the concrete `*config.Loader`
  where the `ConfigReader`/`ConfigWriter` sub-interfaces would make every code
  path mockable.
- **`internal/cli` suite performance** — 116s with `-race` is the direct cause
  of buildflow timeout-class failures; profile, parallelize, trim sleeps.

### 4. Build automation & process maturity

- **Auto-commit daemon quality** — the daemon mixes file types into generic
  commits, bumped go.mod/go.sum without `vendorHash.nix` (breaking `nix build`
  for all clones), and produces `heuristic` messages that are now public
  history. The concrete vendorHash guard is in TODO_LIST; the broader theme is
  making daemon commits deliberate (Renovate-style PRs, file-type-scoped
  messages, or push heartbeats).
- **Status report lifecycle cadence** — the 2026-09-11 docs-health sweep
  archived the fully-resolved 2026-06/07 reports; a standing rule (quarterly,
  or keep-latest-N) would keep `docs/status/` containing only live residue.
- **Multi-system flake checks** — `nix flake check --all-systems` currently
  omits aarch64-linux/darwin targets; either gate explicitly or extend.

### 5. Error handling governance (largely complete)

- **Swallowed-error audit cadence** — quarterly `erraudit` re-check (last full
  review 2026-07-30: 194 findings, all triaged as intentional/idiomatic). Exit
  code governance remains best-in-class: a single `os.Exit()` in `Main()` via
  `errorfamily.ExitCode()`.
- **Error-code registry** — ~40 ad-hoc error codes exist without a convention
  test; a registry would prevent drift (raised 2026-07-08, never prioritized).

### 6. Public-repo operations & launch

The repo went public 2026-09-09 and got metadata + CI rehabilitation
2026-09-11. Remaining launch-tier ideas, gated on user decisions:

- **Website launch** — homepage field is deliberately empty; a docs site would
  fill it (website-launch pattern exists for sibling repos).
- **Demo GIF/asciinema in README** — the tool's before/after config diff is
  inherently visual; no demo asset exists.
- **Announcement** (r/golang, HN, X) — depends on support posture: officially
  maintained OSS (triaged issues, response expectations) vs portfolio code.
- **Community tier** — issue/PR templates are the cheap part; Discussions was
  decided against (sibling consistency), revisit only if the support posture
  changes.

---

## Open questions (user-gated, not tasks)

- **Git history sanitization**: sibling-project references exist in pre-cleanup
  history (accepted 2026-09-09). Reopen as a `git filter-repo` purge, or close
  as "acceptable forever"? Everything downstream (gitleaks scope, announcement)
  depends on this.
- **gohumanize strategy**: project-specific (current, dep-gated on
  `dustin/go-humanize`) vs everywhere (requires custom-binary story for stock
  golangci-lint users). Analysis in `docs/status/2026-08-05_04-14…md`; decision
  blocked since 2026-08-05.
- **Daemon-mangled commit messages**: several mid-2026-07 commits have
  truncated/garbled messages (`5b91141`, `df5f006`, `dc2b0d6` era). Rewrite via
  rebase (public-history risk), annotate via `git notes`, or leave as-is?

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
