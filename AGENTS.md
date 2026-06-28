# Agent Guide: golangci-lint-auto-configure

Concise, enduring context for AI sessions. For detail, see the linked docs at the bottom.

## What This Is

Go CLI that auto-configures and optimizes golangci-lint configs: analyzes, detects missing linters, recommends settings, auto-fixes, replaces deprecated linters, migrates v1→v2, and emits HTML/JSON/SARIF/finding reports.

## Commands (IMPORTANT — there is NO justfile)

The build system is **Nix flake + plain Go tooling**. Do not use `just` — older docs referencing it are stale.

```bash
# Build
nix build                              # reproducible build (preferred)
go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure

# Test (Ginkgo BDD specs, but run via go test)
go test -race ./pkg/... ./internal/...
ginkgo -r --cover                      # alternative: ginkgo directly

# Lint
golangci-lint run --config=.golangci.yml --timeout=5m

# Format (treefmt via Nix)
nix fmt                                # format Go, Nix, templ
nix flake check                        # runs format check + build + tests

# Dev shell (provides Go 1.26, golangci-lint, templ, gopls, etc.)
nix develop
```

## Tech Stack (non-obvious points)

- **Go 1.26**, Cobra CLI, charm.land log/lipgloss/fang, templ HTML reports, `go.yaml.in/yaml/v3`.
- **Testing: Ginkgo v2 + Gomega (BDD)** — NOT standard `testing` style. Specs use `Describe`/`Context`/`It` + Gomega matchers. See `docs/references/testing-style-and-patterns.md`.
- **gogenfilter/v3**: auto-detects generated files to exclude from linting.
- **go-finding**: unified finding model (SARIF/JSON output).
- **go-error-family** (`v0.5.1`): structured error classification. Sentinel errors are registered with Families (Rejection/Conflict/Transient/Corruption/Infrastructure) in `pkg/errors/classification.go`. `ConfigError` implements the `Classified` interface (always Rejection). `Main()` uses `errorfamily.ExitCode(err)` for BSD sysexits exit codes instead of hardcoded `os.Exit(1)`.

## Critical Gotchas (read these — they bite)

1. **No justfile.** Despite older docs, all `just <cmd>` references are stale. Use Nix/Go directly (see Commands above).

2. **templ requires generation.** `.templ` files compile to Go. The Nix build runs `templ generate` in `preBuild`. For local builds after editing `pkg/report/report.templ`, run `templ generate` manually before `go build`.

3. **vendorHash update after go.mod changes.** `nix build` will fail with a hash mismatch. Procedure:

   ```bash
   go mod tidy
   nix build 2>&1 | rg "got:"   # copy the got: sha256
   # paste into flake.nix vendorHash
   nix build                     # rebuild
   ```

4. **go-finding & gogenfilter are private LarsArtmann repos.** `go.mod` uses published versions (no local replace). Nix fetches them via SSH flake inputs and injects `replace` directives in `postPatch` pointing to vendored copies. Local `go build` works with published versions directly.

5. **Error classification via go-error-family.** `pkg/errors/classification.go` has an `init()` that registers all sentinel errors with their `errorfamily.Family`. To add a new sentinel: add it to the map in that file. `ConfigError` implements `Classified` → `Rejection` (type-level, checked before sentinels). `AnalysisError` does NOT implement `Classified` — its sentinels in the cause chain (e.g. `ErrVersionTooOld`) handle classification. Exit codes: Rejection/Conflict → 1, Transient → 75, Corruption → 65, Infrastructure → 69.

6. **Fixer normalization counting.** Every config mutation in `applyAndSave` MUST increment `fixCounts.normalization` — otherwise the `counts.total()==0` guard silently discards changes. The fixer also injects `issues.max-issues-per-linter: 50` and `max-same-issues: 10` when absent (prevents golangci-lint's default `max-same-issues: 3` from hiding CI problems).

7. **Auto-injected linter safe defaults.** Linters that misbehave without explicit config get safe defaults injected by `injectDefaultSettings` (`pkg/linter/fixer_config.go`) when enabled and missing settings. Map lives in `pkg/constants/config.go` (`DefaultLinterSettings`). Idempotent; never overwrites existing user settings.

8. **Config auto-creation requires a git repo.** `configure` creates a default `.golangci.yml` if missing. Git is mandatory for version-control safety.

9. **DI is manual.** No `internal/di/` directory (older docs lie). Dependencies wired manually in CLI commands. No DI framework.

10. **Linter priority data is static constants.** Priorities live in `pkg/constants/linter_priorities.go` + reasons in `linter_reasons.go`. Not dynamically computed from golangci-lint. To add/change a linter, edit both files (and `presets.go`/`rules.go` if relevant).

11. **Versioning is self-initializing.** `pkg/version/` reads ldflags with `runtime/debug.ReadBuildInfo()` fallback. `cli.Version` self-inits — no manual setup. All build targets (Nix, CI) inject via ldflags.

## Where to Find Detail

| Topic                                                  | Location                                        |
| ------------------------------------------------------ | ----------------------------------------------- |
| Directory structure & patterns                         | `docs/references/code-organization.md`          |
| Adding commands/linters, common tasks, troubleshooting | `docs/references/working-with-codebase.md`      |
| BDD testing, code style, CI/CD                         | `docs/references/testing-style-and-patterns.md` |
| Error handling patterns                                | `docs/references/error-handling.md`             |
| gogenfilter & go-finding integration                   | `docs/references/integrations.md`               |
| User-facing usage                                      | `README.md`                                     |
| Feature inventory                                      | `FEATURES.md`                                   |
| Open work                                              | `TODO_LIST.md`                                  |
| Domain language                                        | `docs/DOMAIN_LANGUAGE.md`                       |
