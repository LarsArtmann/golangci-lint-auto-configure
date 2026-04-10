# Migration to Nix Flakes — Proposal

**Date:** 2026-04-09
**Status:** Draft
**Author:** Agent-assisted analysis

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current State Analysis](#2-current-state-analysis)
3. [Why Nix Flakes](#3-why-nix-flakes)
4. [Proposed Architecture](#4-proposed-architecture)
5. [Migration Phases](#5-migration-phases)
6. [File-by-File Breakdown](#6-file-by-file-breakdown)
7. [Justfile Integration](#7-justfile-integration)
8. [CI/CD Migration Strategy](#8-cicd-migration-strategy)
9. [Risks and Mitigations](#9-risks-and-mitigations)
10. [Decision Matrix](#10-decision-matrix)
11. [Acceptance Criteria](#11-acceptance-criteria)

---

## 1. Executive Summary

This proposal outlines a phased migration of **golangci-lint-auto-configure** to Nix Flakes for **fully reproducible builds, pinned developer toolchains, and hermetic CI/CD**. The project currently relies on ad-hoc tool installation (Go 1.26, golangci-lint v2, ginkgo, templ, just) with no version-locking mechanism beyond `go.mod`. Nix Flakes would eliminate "works on my machine" issues, make onboarding instant (`nix develop`), and produce bit-for-bit reproducible binaries across macOS and Linux.

**Key outcomes:**

- `nix develop` provides the complete dev environment (Go 1.26, golangci-lint, ginkgo, templ, just, jq, git)
- `nix build` produces the CLI binary with ldflags-injected version
- CI/CD uses `nix flake check` for deterministic linting and testing
- Existing `just` commands remain unchanged — Nix provides the tools, `just` orchestrates them
- No disruption to non-Nix users (fallback paths preserved)

---

## 2. Current State Analysis

### 2.1 Build Toolchain Inventory

Every tool required to build, test, and lint this project, as extracted from the Justfile, scripts, and CI config:

| Tool              | Version / Source                | Used By                             | How Installed Today                  |
| ----------------- | ------------------------------- | ----------------------------------- | ------------------------------------ |
| **Go**            | 1.26 (go.mod: `go 1.26.0`)      | Build, test, lint, install          | Manual / `actions/setup-go@v5`       |
| **golangci-lint** | v2.10.1+ (min enforced in code) | `just lint`, CI, pre-commit         | Manual / `golangci-lint-action@v9`   |
| **ginkgo**        | v2.28.1 (indirect via go.mod)   | `just test` (`ginkgo -r --cover`)   | `go install` / comes with ginkgo dep |
| **templ**         | v0.3.1001 (go.mod)              | `templ generate` (report templates) | Manual `go install`                  |
| **just**          | Any                             | All `just` commands                 | Manual (`brew install just`, etc.)   |
| **git**           | Any                             | Version ldflags, pre-commit hooks   | System package                       |
| **jq**            | Any                             | `scripts/verify_linter_count.sh`    | Manual                               |
| **bc**            | Any                             | `scripts/verify_linter_count.sh`    | System package                       |
| **pre-commit**    | Any                             | `.pre-commit-config.yaml`           | `pip install pre-commit`             |
| **gofmt/gofumpt** | Via Go / golangci-lint          | `just fmt`, `just fmt-check`        | Bundled with Go                      |
| **golines**       | Via golangci-lint formatters    | `.golangci.yml` formatters          | Bundled via golangci-lint            |
| **gci**           | Via golangci-lint formatters    | `.golangci.yml` formatters          | Bundled via golangci-lint            |

### 2.2 Current Build Commands (from Justfile)

```
build:          GOWORK=off GOTOOLCHAIN=local go build -o bin/golangci-lint-auto-configure ./cmd/golangci-lint-auto-configure
test:           GOWORK=off GOTOOLCHAIN=local ginkgo -r --cover
test-coverage:  GOWORK=off GOTOOLCHAIN=local go test ./... -coverprofile=coverage.out -covermode=atomic
coverage-html:  GOWORK=off GOTOOLCHAIN=local go test ./... + go tool cover -html
lint:           GOWORK=off GOTOOLCHAIN=local golangci-lint run --config .golangci.yml
install-local:  go build -ldflags "-X main.version=$VERSION" -o "$GOPATH/bin/..." ./cmd/...
fmt:            go fmt ./...
fmt-check:      test -z "$(gofmt -l .)"
tidy:           GOWORK=off GOTOOLCHAIN=local go mod tidy
deps:           GOWORK=off GOTOOLCHAIN=local go mod download
```

**Observations:**

- `GOWORK=off GOTOOLCHAIN=local` is used everywhere — suggests a deliberate avoidance of Go workspace / auto-toolchain-fetch behavior. Nix makes this unnecessary (toolchain is hermetic).
- Version injection via ldflags: `-X main.version=$VERSION` — must be replicated in `nix build`.
- The `install-local` recipe uses a bash shebang script inside justfile — Nix build should replicate the ldflags pattern.

### 2.3 Shell Scripts

| Script                           | Dependencies                                 | Purpose                                           |
| -------------------------------- | -------------------------------------------- | ------------------------------------------------- |
| `scripts/pre-commit-hook.sh`     | `golangci-lint-auto-configure`, bash         | Pre-commit analysis hook                          |
| `scripts/validate_linter_doc.sh` | bash, coreutils                              | Validates linter doc files meet quality standards |
| `scripts/verify_linter_count.sh` | bash, `golangci-lint`, `jq`, `bc`, coreutils | Verifies linter doc count matches golangci-lint   |

**Note:** `validate_linter_doc.sh` and `verify_linter_count.sh` contain hardcoded paths (`/Users/larsartmann/...`). These should be parameterized regardless of the Nix migration.

### 2.4 CI/CD Pipeline (`.github/workflows/ci.yml`)

- **Matrix:** Go 1.25 + 1.26 on `ubuntu-latest`
- **Steps:** `actions/setup-go@v5` → `go mod download` → `go build` → `go test -race` → coverage → `codecov`
- **Lint job:** `golangci/golangci-lint-action@v9` with `version: latest`
- **Pain point:** `golangci-lint-action@v9` with `version: latest` is not reproducible — different CI runs may use different linter versions.

### 2.5 Docker Build (`Dockerfile`)

- Multi-stage: `golang:1.26-alpine` (builder) → `golangci/golangci-lint:2.1.5-alpine` (runtime)
- Runtime image includes `golangci-lint` — useful for using this tool in CI pipelines
- Currently pins golangci-lint at v2.1.5 (outdated vs. v2.10.1+ requirement in code)

### 2.6 Existing Nix Presence

**None.** No `flake.nix`, `flake.lock`, `default.nix`, `shell.nix`, or any Nix-related files exist.

---

## 3. Why Nix Flakes

### 3.1 Problems Solved

| Problem                        | Current State                                               | With Nix Flakes                                    |
| ------------------------------ | ----------------------------------------------------------- | -------------------------------------------------- |
| Tool version drift             | "Install Go 1.26, golangci-lint v2, ginkgo, templ, just..." | `nix develop` — everything pinned in `flake.lock`  |
| CI reproducibility             | `golangci-lint-action@v9` with `version: latest`            | `nix flake check` — exact same binaries every run  |
| Onboarding friction            | README lists manual install steps                           | Clone → `nix develop` → ready                      |
| "Works on my machine"          | Different Go versions, missing tools                        | Identical environments across all machines         |
| Cross-platform consistency     | macOS vs Linux tool differences                             | Nix normalizes both                                |
| Build artifact reproducibility | `go build` depends on local Go version                      | `nix build` produces bit-for-bit identical outputs |

### 3.2 What Nix Flakes Will NOT Replace

- **Justfile** — `just` remains the task runner. Nix provides the tools, `just` orchestrates them.
- **go.mod** — Still the source of truth for Go dependencies.
- **.golangci.yml** — Still the linter configuration.
- **Pre-commit hooks** — Still work, but can optionally use Nix-managed tools.
- **Dockerfile** — Can be simplified (use `nix build` as builder stage) but not required initially.

### 3.3 Trade-offs

| Pro                           | Con                                               |
| ----------------------------- | ------------------------------------------------- |
| Full reproducibility          | Nix learning curve for contributors               |
| Instant onboarding            | `flake.lock` adds a file to maintain              |
| Hermetic CI/CD                | Longer initial `nix develop` (subsequent: cached) |
| Multi-platform builds         | Nix must be installed on all dev machines         |
| Version pinning for all tools | `vendorHash` must be updated on `go.mod` changes  |

---

## 4. Proposed Architecture

### 4.1 File Structure (New Files)

```
golangci-lint-auto-configure/
├── flake.nix              # Main flake definition
├── flake.lock             # Pinned dependency versions (auto-generated)
├── nix/
│   └── packages/
│       └── default.nix    # Package definition (extracted for readability)
```

### 4.2 Flake Outputs

```
flake.nix
├── inputs
│   ├── nixpkgs (nixos-unstable)
│   └── flake-utils (optional, for multi-system support)
├── outputs
│   ├── packages.<system>.default     # The CLI binary
│   ├── packages.<system>.docker      # Docker image (optional, Phase 3)
│   ├── devShells.<system>.default    # Full dev environment
│   ├── checks.<system>               # Test + lint checks
│   ├── apps.<system>.default         # Run the CLI
│   └── overlays.default              # Overlay for other flakes
```

### 4.3 Proposed `flake.nix` (Reference Implementation)

```nix
{
  description = "Automatically configure and optimize golangci-lint configurations";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        # Version from git tag, fallback to "dev"
        version = builtins.substring 0 7 (self.rev or "dev");

        # The main CLI package
        golangci-lint-auto-configure = pkgs.buildGoModule rec {
          pname = "golangci-lint-auto-configure";
          inherit version;

          src = ./.;

          # Update this hash when go.mod/go.sum change:
          #   nix-build -A packages.<system>.default 2>&1 | tail -1
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

          subPackages = [ "cmd/golangci-lint-auto-configure" ];

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
          ];

          CGO_ENABLED = 0;

          meta = with pkgs.lib; {
            description = "Automatically configure and optimize golangci-lint configurations";
            homepage = "https://github.com/LarsArtmann/golangci-lint-auto-configure";
            license = licenses.mit;
            mainProgram = "golangci-lint-auto-configure";
          };
        };

      in
      {
        packages.default = golangci-lint-auto-configure;

        apps.default = {
          type = "app";
          program = "${golangci-lint-auto-configure}/bin/golangci-lint-auto-configure";
        };

        devShells.default = pkgs.mkShell {
          inputsFrom = [ golangci-lint-auto-configure ];

          packages = with pkgs; [
            # Build tools
            go_1_26
            just

            # Linting
            golangci-lint

            # Testing
            ginkgo

            # Templating
            templ

            # Utility
            jq
            git
            pre-commit

            # Go tooling
            gotools
            gopls
          ];

          shellHook = ''
            echo "golangci-lint-auto-configure dev shell"
            echo "  Go:         $(go version)"
            echo "  golangci-lint: $(golangci-lint version --short 2>/dev/null || echo 'N/A')"
            echo "  ginkgo:     $(ginkgo version 2>/dev/null || echo 'N/A')"
            echo "  just:       $(just --version 2>/dev/null || echo 'N/A')"
          '';
        };

        checks = {
          build = golangci-lint-auto-configure;

          lint = pkgs.runCommand "lint-check" { } ''
            cd ${self}
            ${pkgs.golangci-lint}/bin/golangci-lint run --config .golangci.yml
            touch $out
          '';

          test = pkgs.runCommand "test-check" { } ''
            cd ${self}
            ${pkgs.ginkgo}/bin/ginkgo -r --cover
            touch $out
          '';
        };
      }
    );
}
```

### 4.4 Systems Supported

Per `flake-utils.lib.eachDefaultSystem`:

- `x86_64-linux`
- `aarch64-linux`
- `x86_64-darwin` (Intel Mac)
- `aarch64-darwin` (Apple Silicon)

---

## 5. Migration Phases

### Phase 1: Foundation (Minimum Viable Flake)

**Goal:** `nix develop` works. `nix build` works.

**Tasks:**

| #   | Task                                                                             | Est. Effort |
| --- | -------------------------------------------------------------------------------- | ----------- |
| 1.1 | Create `flake.nix` with `buildGoModule`, `devShells.default`                     | 2h          |
| 1.2 | Run `nix build` to generate initial `vendorHash`                                 | 15min       |
| 1.3 | Commit `flake.nix` + `flake.lock`                                                | 5min        |
| 1.4 | Update `.gitignore` to include `result` symlink                                  | 5min        |
| 1.5 | Verify `nix develop` provides all tools (go, ginkgo, golangci-lint, just, templ) | 30min       |
| 1.6 | Verify `just build && just test && just lint` all work inside `nix develop`      | 30min       |

**Deliverables:**

- `flake.nix` with `packages.default` and `devShells.default`
- `flake.lock` with pinned nixpkgs
- Updated `.gitignore`

### Phase 2: CI/CD Integration

**Goal:** GitHub Actions uses Nix for deterministic builds.

**Tasks:**

| #   | Task                                                                                    | Est. Effort |
| --- | --------------------------------------------------------------------------------------- | ----------- |
| 2.1 | Add `nix develop` verification step to CI                                               | 1h          |
| 2.2 | Add `nix flake check` job to CI                                                         | 1h          |
| 2.3 | Cache `/nix/store` in GitHub Actions using `cachix` or `nix-community/cache-nix-action` | 2h          |
| 2.4 | Keep existing CI jobs as fallback (dual-mode)                                           | 30min       |
| 2.5 | Optionally add Cachix binary cache for PR builds                                        | 1h          |

**Deliverables:**

- Updated `.github/workflows/ci.yml` with Nix-based job
- Optional: `cachix` integration for binary cache

### Phase 3: Advanced Features

**Goal:** Docker image via Nix, pre-commit integration, cross-compilation.

**Tasks:**

| #   | Task                                                                     | Est. Effort |
| --- | ------------------------------------------------------------------------ | ----------- |
| 3.1 | Create `nix/packages/docker.nix` for Nix-built Docker image              | 2h          |
| 3.2 | Add `overlays.default` for other flakes to consume                       | 30min       |
| 3.3 | Integrate `direnv` support (`.envrc` with `use flake`)                   | 30min       |
| 3.4 | Cross-compilation targets (`nix build .#packages.aarch64-linux.default`) | 1h          |
| 3.5 | Parameterize hardcoded paths in shell scripts                            | 1h          |
| 3.6 | Add `nix run . -- analyze` as alias                                      | 15min       |

**Deliverables:**

- Nix-built Docker image (replaces or supplements Dockerfile)
- `.envrc` for automatic shell activation
- Cross-platform binary builds

### Phase 4: Cleanup and Optimization

**Goal:** Remove duplication, finalize migration.

**Tasks:**

| #   | Task                                                                   | Est. Effort |
| --- | ---------------------------------------------------------------------- | ----------- |
| 4.1 | Remove `GOWORK=off GOTOOLCHAIN=local` from Justfile (Nix handles this) | 15min       |
| 4.2 | Update AGENTS.md with Nix commands                                     | 30min       |
| 4.3 | Update README.md with Nix onboarding instructions                      | 30min       |
| 4.4 | Remove or deprecate Dockerfile if Nix Docker build is sufficient       | 1h          |
| 4.5 | Add `nix fmt` integration using `nixpkgs-fmt` or `alejandra`           | 30min       |

---

## 6. File-by-File Breakdown

### 6.1 Files to Create

| File                       | Purpose                                                  |
| -------------------------- | -------------------------------------------------------- |
| `flake.nix`                | Main flake: inputs, outputs, packages, devShells, checks |
| `flake.lock`               | Auto-generated: pins nixpkgs and all flake inputs        |
| `nix/packages/default.nix` | (Optional) Extracted package definition for readability  |

### 6.2 Files to Modify

| File                             | Change                                                         | Phase |
| -------------------------------- | -------------------------------------------------------------- | ----- |
| `.gitignore`                     | Add `result` (Nix build symlink)                               | 1     |
| `.github/workflows/ci.yml`       | Add Nix-based CI job                                           | 2     |
| `justfile`                       | Remove `GOWORK=off GOTOOLCHAIN=local` (Nix makes it redundant) | 4     |
| `AGENTS.md`                      | Add Nix commands section                                       | 4     |
| `README.md`                      | Add Nix onboarding section                                     | 4     |
| `scripts/validate_linter_doc.sh` | Parameterize hardcoded path                                    | 3     |
| `scripts/verify_linter_count.sh` | Parameterize hardcoded path                                    | 3     |

### 6.3 Files Unchanged

| File                      | Reason                                           |
| ------------------------- | ------------------------------------------------ |
| `go.mod` / `go.sum`       | Still source of truth for Go dependencies        |
| `.golangci.yml`           | Linter configuration unchanged                   |
| `.pre-commit-config.yaml` | Still works, tools now from Nix                  |
| `.pre-commit-hooks.yaml`  | Published hooks unchanged                        |
| `Dockerfile`              | Kept as fallback, optionally replaced in Phase 3 |
| All `.go` files           | No code changes needed                           |
| All `_test.go` files      | No test changes needed                           |

---

## 7. Justfile Integration

### 7.1 Current State

The Justfile assumes all tools are on `$PATH`. With Nix, `nix develop` puts them there — **no Justfile changes required** for Phase 1.

### 7.2 Optional Enhancements

```just
# Add these optional recipes to the Justfile:

# Enter the Nix development shell
shell:
    nix develop

# Build with Nix (reproducible)
nix-build:
    nix build
    echo "Binary: ./result/bin/golangci-lint-auto-configure"

# Run all Nix checks
nix-check:
    nix flake check

# Update Nix flake inputs
nix-update:
    nix flake update

# Generate/update vendorHash after go.mod changes
nix-vendor-hash:
    #!/bin/bash
    set -e
    echo "Building to calculate vendorHash..."
    nix build 2>&1 | tail -5
```

### 7.3 Environment Variable Cleanup

After Phase 4, these can be removed from the Justfile since Nix provides the correct Go version:

```
# Before (current):
build:
    @GOWORK=off GOTOOLCHAIN=local go build -o bin/...

# After (Phase 4):
build:
    @go build -o bin/...
```

---

## 8. CI/CD Migration Strategy

### 8.1 Dual-Mode Approach (Recommended)

Keep existing CI jobs alongside Nix jobs during migration. This ensures zero disruption.

```yaml
# Proposed addition to .github/workflows/ci.yml
nix-check:
  name: Nix Flake Check
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - uses: DeterminateSystems/nix-installer-action@main

    - uses: DeterminateSystems/magic-nix-cache-action@main

    - name: Check flake
      run: nix flake check --no-build

    - name: Build with Nix
      run: nix build

    - name: Run tests via Nix
      run: nix develop --command just test

    - name: Run lint via Nix
      run: nix develop --command just lint
```

### 8.2 Caching Strategy

| Layer           | Tool                                 | Purpose                              |
| --------------- | ------------------------------------ | ------------------------------------ |
| Nix store       | `magic-nix-cache-action` or `cachix` | Cache `/nix/store` between CI runs   |
| Go modules      | Handled by `buildGoModule`           | `vendorHash` ensures reproducibility |
| Build artifacts | Nix derivation cache                 | Incremental builds within Nix        |

### 8.3 Migration Timeline

| Milestone            | Trigger                                            |
| -------------------- | -------------------------------------------------- |
| Phase 2 complete     | Nix CI runs alongside existing CI (both must pass) |
| Confidence threshold | After 2 weeks of dual-mode passing                 |
| Existing CI removal  | Optional — keep as fallback indefinitely           |

---

## 9. Risks and Mitigations

### 9.1 Technical Risks

| Risk                                                     | Likelihood              | Impact         | Mitigation                                                                                                                        |
| -------------------------------------------------------- | ----------------------- | -------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `vendorHash` mismatch after `go.mod` changes             | High (every dep update) | Low (5min fix) | Document the workflow: `go mod tidy` → `nix build` → copy hash                                                                    |
| Nix not available on contributor machines                | Medium                  | Medium         | Keep Justfile working without Nix (fallback mode)                                                                                 |
| `buildGoModule` doesn't handle `local replace` in go.mod | Medium                  | Medium         | The project has `replace github.com/LarsArtmann/universal-workflow => /Users/...` — must be removed or handled via `overrideMods` |
| Long initial `nix develop` time                          | Low                     | Low            | Subsequent runs are cached; `cachix` for CI                                                                                       |
| Nixpkgs Go version lags behind                           | Low                     | Medium         | Use `go_1_26` explicitly; can override with `fetchurl` if needed                                                                  |

### 9.2 The `local replace` Problem

**Current `go.mod` contains:**

```
replace github.com/LarsArtmann/universal-workflow => /Users/larsartmann/projects/universal-workflow
```

This will **break** `nix build` because Nix builds in a sandbox without access to local paths.

**Solutions (in order of preference):**

1. **Remove the local replace** — if `universal-workflow` is published or can be referenced by commit hash
2. **Add `universal-workflow` as a flake input** — reference the repo directly
3. **Use `overrideMods`** in `buildGoModule` to inject the local dependency
4. **Keep separate `go.work` for local dev** — Nix uses `GOWORK=off` (already in Justfile)

**Recommended:** Option 2 — add as flake input:

```nix
inputs = {
  universal-workflow = {
    url = "github:LarsArtmann/universal-workflow";
    flake = false;
  };
};
```

Then vendor it into the Go module cache during build.

### 9.3 Organizational Risks

| Risk                               | Likelihood | Impact | Mitigation                                                 |
| ---------------------------------- | ---------- | ------ | ---------------------------------------------------------- |
| Contributor resistance to Nix      | Medium     | Low    | Nix is opt-in — `just` commands still work                 |
| Maintenance burden of `flake.lock` | Low        | Low    | `nix flake update` is a single command                     |
| CI binary cache costs              | Low        | Low    | GitHub Actions cache is free; Cachix free tier is generous |

---

## 10. Decision Matrix

### `buildGoModule` vs `buildGoApplication` (gomod2nix)

| Criterion          | `buildGoModule`                                 | `buildGoApplication` (gomod2nix)           |
| ------------------ | ----------------------------------------------- | ------------------------------------------ |
| Simplicity         | **High** — single `vendorHash`                  | Low — requires `gomod2nix.toml` generation |
| Reproducibility    | High                                            | Higher (explicit per-module hashes)        |
| Maintenance        | Low — update `vendorHash` on dep changes        | Medium — regenerate `gomod2nix.toml`       |
| Ecosystem adoption | **Standard** — used by Grafana, Headscale, etc. | Niche — used by smaller projects           |
| `go.mod` changes   | Update one hash                                 | Regenerate full TOML file                  |

**Recommendation:** `buildGoModule` — simpler, widely adopted, lower maintenance burden.

### `flake-utils` vs `flake-parts` vs raw `nixpkgs.lib.genAttrs`

| Criterion       | `flake-utils`      | `flake-parts`                 | Raw `genAttrs`            |
| --------------- | ------------------ | ----------------------------- | ------------------------- |
| Simplicity      | **High** — drop-in | Medium — module system        | Medium — more boilerplate |
| Extensibility   | Low                | **High** — module composition | Medium                    |
| Community usage | Very high          | Growing                       | Low                       |
| Dependency      | Extra input        | Extra input                   | None                      |

**Recommendation:** `flake-utils` for Phase 1. Consider `flake-parts` if the project grows complex flake outputs.

---

## 11. Acceptance Criteria

### Phase 1 Complete When

- [ ] `nix develop` provides: `go`, `ginkgo`, `golangci-lint`, `just`, `templ`, `jq`, `git`
- [ ] `nix build` produces `./result/bin/golangci-lint-auto-configure`
- [ ] The built binary reports correct version via `--version`
- [ ] `just build`, `just test`, `just lint` all pass inside `nix develop`
- [ ] `flake.lock` is committed
- [ ] `.gitignore` includes `result`
- [ ] Non-Nix workflow (`just build` without Nix) still works unchanged

### Phase 2 Complete When

- [ ] CI has a Nix-based job that runs `nix build` and `nix flake check`
- [ ] Nix CI job caches `/nix/store` between runs
- [ ] Both Nix and non-Nix CI jobs pass

### Phase 3 Complete When

- [ ] `nix build .#docker` produces a Docker image
- [ ] `.envrc` with `use flake` works for automatic shell activation
- [ ] Shell scripts have parameterized paths (no hardcoded `/Users/...`)

### Phase 4 Complete When

- [ ] `AGENTS.md` includes Nix commands
- [ ] `README.md` includes Nix onboarding section
- [ ] Justfile no longer needs `GOWORK=off GOTOOLCHAIN=local`
- [ ] `nix fmt` formats `.nix` files

---

## Appendix A: Quick Reference Commands

```bash
# Development
nix develop                          # Enter dev shell
nix develop --command just test      # Run tests in Nix shell
nix develop --command just lint      # Run linter in Nix shell

# Building
nix build                            # Build the CLI binary
./result/bin/golangci-lint-auto-configure --version

# Running
nix run . -- analyze                 # Run the CLI directly

# Checking
nix flake check                      # Run all checks (build, test, lint)

# Maintenance
nix flake update                     # Update all flake inputs
nix build --rebuild                  # Rebuild after go.mod changes → new vendorHash

# Formatting
nix fmt                              # Format .nix files (if configured)
```

## Appendix B: Onboarding Instructions (for README)

```markdown
## Quick Start with Nix (Recommended)

1. [Install Nix](https://nixos.org/download.html)
2. Enable Flakes: `mkdir -p ~/.config/nix && echo 'experimental-features = nix-command flakes' >> ~/.config/nix/nix.conf`
3. Clone this repo
4. Run `nix develop` — you're ready to go

All tools (Go, ginkgo, golangci-lint, just, templ) are provided automatically.

## Quick Start without Nix

Install manually:

- Go 1.26+
- golangci-lint v2.10+
- ginkgo (`go install github.com/onsi/ginkgo/v2/ginkgo@latest`)
- just (`brew install just` or `go install github.com/casey/just@latest`)
- templ (`go install github.com/a-h/templ/cmd/templ@latest`)

Then: `just build && just test && just lint`
```

## Appendix C: `vendorHash` Update Workflow

When `go.mod` or `go.sum` changes:

```bash
# 1. Tidy dependencies
just tidy

# 2. Attempt build — it will fail with the expected hash
nix build 2>&1 | tail -5

# 3. Copy the "got:" hash from the error output
# 4. Update vendorHash in flake.nix
# 5. Rebuild
nix build
```

Example error output:

```
error: hash mismatch in fixed-output derivation '/nix/store/...':
         specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
            got:    sha256-BcC62Sx4QdpAoWr5JdI5VDWtZbW2RgHf5m9Bz...
```

Copy the `got:` value into `flake.nix` as the new `vendorHash`.
