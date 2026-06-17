# Status Report: BuildFlow test-race + test-coverage Fix Sprint

**Date:** 2026-06-17 17:04  
**Trigger:** `buildflow --fix --semantic -p --build-mode=full` failed with 2 steps: `test-race` (exit 2) and `test-coverage` (exit 1)  
**Session goal:** Make it work — diagnose and fix both test failures.

---

## Executive Summary

Three root causes were found, all fixed. The failures were **not** caused by actual test failures or data races — they were caused by **environment configuration** (CGO disabled), **test framework inconsistency** (4 packages missing Ginkgo suites), and a **BuildFlow parallel execution bug** (test-race and test-coverage compiling to the same directories simultaneously).

| Root Cause | Impact | Fix Location | Status |
|---|---|---|---|
| `CGO_ENABLED=0` in devShell blocks race detector | `test-race` exits 2: "-race requires cgo" | `flake.nix:135` | ✅ Fixed |
| 4 packages lack Ginkgo suite files | BuildFlow uses `go test -parallel=N`, Ginkgo rejects it | 4× `suite_test.go` | ✅ Fixed |
| test-race + test-coverage compile simultaneously | Binary corruption: "no such file or directory" | BuildFlow source: step deps | ✅ Fixed (source), ⚠️ not deployed |

---

## a) FULLY DONE ✅

| # | Work Item | Files | Impact |
|---|-----------|-------|--------|
| 1 | **Diagnose CGO_ENABLED=0** — DevShell exported `CGO_ENABLED=0`, but `-race` requires CGO. The race detector links a C-based runtime. gcc was already in the shell but unused. | `flake.nix:135` | Unblocks `test-race` in devShell and CI |
| 2 | **Fix CGO_ENABLED in devShell** — Changed `CGO_ENABLED = "0"` → `CGO_ENABLED = "1"`. The package build still uses `CGO_ENABLED=0` (line 93) for static binaries. The Nix `race` check already used `CGO_ENABLED=1` (line 169). | `flake.nix` | Race detector works in `nix develop` |
| 3 | **Diagnose missing Ginkgo suites** — BuildFlow's `shouldUseGinkgoRunner()` returns false if ANY test package lacks a file importing `onsi/ginkgo`. 4 packages (`pkg/detection`, `pkg/diff`, `pkg/finding`, `pkg/ui`) had zero Ginkgo presence, causing fallback to `go test -parallel=N`. Ginkgo intentionally FAILS with `-parallel`: "Go test's implementation of parallelization does not actually parallelize Ginkgo specs." | Analysis only | Identified exact mechanism |
| 4 | **Add Ginkgo suite files to 4 packages** — Created `suite_test.go` with `RegisterFailHandler(Fail)` + `RunSpecs()` in each package, matching the existing pattern in `pkg/constants/suite_test.go` and `pkg/types/suite_test.go`. | `pkg/detection/suite_test.go`, `pkg/diff/suite_test.go`, `pkg/finding/suite_test.go`, `pkg/ui/suite_test.go` | BuildFlow now uses `ginkgo run` for ALL packages |
| 5 | **Diagnose parallel compilation conflict** — When `test-race` and `test-coverage` run simultaneously, both invoke `ginkgo run` which compiles test binaries to the same package directories. Concurrent compilation causes: "testing: cannot use -test.coverprofile because test binary was not built with coverage enabled" and "fork-exec ... migration.test: no such file or directory". | Analysis only | Identified exact mechanism |
| 6 | **Fix BuildFlow parallel conflict** — Added dependency `registerStepDeps(StepTestCoverage, StepTestRace)` so test-coverage waits for test-race to finish. Both are in Phase 8 (Testing), so the DAG scheduler now serializes them. | `/home/lars/projects/BuildFlow/domain/step_definitions.go:171` | Prevents concurrent compilation corruption |
| 7 | **Verify all 15 Ginkgo suites pass individually** — Ran `buildflow -s test-race` and `buildflow -s test-coverage` separately, both exit 0. All 316+ specs pass with race detector and coverage enabled. | Verification | Tests are green |
| 8 | **Verify race tests pass with CGO_ENABLED=1** — `CGO_ENABLED=1 go test -race ./...` passes all packages in 28s. No actual data races detected. | Verification | Code is race-clean |

---

## b) PARTIALLY DONE 🟡

| # | Item | Status | What remains |
|---|------|--------|-------------|
| 1 | **Deploy fixed BuildFlow binary** | Source fixed, binary not installed | The system-installed buildflow (`/nix/store/072cklhz...`) is old. Need to rebuild and install via `nix build` or `go install`. Nix build fails with vendorHash mismatch. |
| 2 | **Full buildflow green run** | Individual steps pass, full run not yet verified | Need to run the complete `buildflow --fix --semantic -p --build-mode=full` end-to-end with the fixed binary to confirm zero failures. |
| 3 | **Dependency updates committed** | `go.mod`/`go.sum` updated by buildflow's `go-mod-update` step | gogenfilter v3.1.0→v3.2.0, go-finding v0.6.1→v0.8.0, go-toml v2.3.1→v2.4.0. Need to update Nix vendorHash if building with Nix. |

---

## c) NOT STARTED ⬜

| # | Item | Why | Impact |
|---|------|-----|--------|
| 1 | Update Nix `vendorHash` after go.mod changes | Nix build will fail until hash is updated. Run `nix build`, copy `got:` hash into `flake.nix`. | Blocks reproducible Nix builds |
| 2 | Add `reports/` to `.gitignore` | BuildFlow generates `reports/coverage.out`, `reports/html/` — currently untracked and cluttering git status | Cleanliness |
| 3 | Fix `oxfmt` step failure | Local buildflow run showed `oxfmt: exit status 2` — unrelated to test fixes but blocks full green run | Medium — formatting tool issue |
| 4 | Update BuildFlow nix vendorHash | BuildFlow's own `flake.nix` vendorHash is stale after the dependency fix | Blocks rebuilding BuildFlow via Nix |

---

## d) TOTALLY FUCKED UP! 🔥

| # | What went wrong | Root cause | Mitigation |
|---|----------------|------------|------------|
| 1 | **System-installed BuildFlow can't be replaced easily** | It's a Nix store path (`/run/current-system/sw/bin/buildflow` → `/nix/store/072cklhz...`), immutable. Can't `cp` over it. | Need to rebuild system flake or install to `$GOPATH/bin` and adjust PATH |
| 2 | **Nix build of BuildFlow fails with vendorHash mismatch** | `specified: sha256-ytVY...` vs `got: sha256-RNHO...` — the vendorHash in BuildFlow's `flake.nix` is stale | Update vendorHash in BuildFlow's flake.nix |
| 3 | **Local `go build` binary missing `-p` flag** | Different cobra/fang version in source vs installed binary. The `-p` shorthand for `--parallel` isn't registered. | Use `--build-mode=full` without `-p`, or align cobra versions |
| 4 | **Initial wrong diagnosis: assumed CGO was the only issue** | Fixed CGO_ENABLED but tests still failed. Had to dig through 3 layers of root causes. | Systematic `buildflow -s <step>` isolation revealed the Ginkgo/parallel issues |

---

## e) WHAT WE SHOULD IMPROVE! 💡

1. **Consistent test framework across ALL packages** — Every package with tests should have a Ginkgo suite file. The 4 missing packages (`detection`, `diff`, `finding`, `ui`) had standard `testing.T` tests alongside Ginkgo tests in sibling packages. This inconsistency caused BuildFlow to downgrade the runner. A CI check should enforce "all test packages use Ginkgo or none do."

2. **CGO_ENABLED should be 1 in devShells by default** — The race detector is a critical safety tool. Setting `CGO_ENABLED=0` in the devShell makes `-race` impossible without manual override. The package build can stay `CGO_ENABLED=0` for static binaries, but the devShell should enable CGO.

3. **BuildFlow should detect test-coverage/test-race conflict** — These two steps fundamentally cannot run in parallel because they compile test binaries to the same locations with different flags (`-race` vs `-cover`). This dependency should be hardcoded in BuildFlow's step definitions (now done in source, needs deployment).

4. **`.buildflow.yml` should exist in this project** — The project has no `.buildflow.yml`, so it uses global defaults. A project-specific config could set `max_concurrency: 1` for the Testing phase or document the known parallel limitation.

5. **`reports/` directory should be gitignored** — Generated test coverage artifacts are currently untracked. Add to `.gitignore`.

6. **Test suite files are trivial but critical** — The `suite_test.go` files we added are empty Ginkgo entry points (0 specs). They exist solely to make BuildFlow detect the package as "Ginkgo." This is fragile — a comment or convention should document why these files exist.

---

## f) Top #25 Things We Should Get Done Next!

### Priority 1: Unblock (do immediately)

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 1 | **Commit the 3 fixes** (flake.nix CGO, 4 suite_test.go, go.mod updates) | 5 min | Unblocks all future work |
| 2 | **Add `reports/` to `.gitignore`** | 1 min | Clean git status |
| 3 | **Deploy fixed BuildFlow binary** — `cd ~/projects/BuildFlow && go install ./cmd/buildflow` or update system flake | 5 min | Makes parallel test fix active |
| 4 | **Update Nix vendorHash** in this project's `flake.nix` after go.mod changes | 3 min | Unblocks `nix build`, `nix flake check` |

### Priority 2: Verify end-to-end

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 5 | **Run full `buildflow --fix --semantic -p --build-mode=full --no-tui`** with fixed binary | 2 min | Confirm zero failures |
| 6 | **Run `nix flake check`** to verify Nix integrity | 3 min | CI readiness |
| 7 | **Run `just lint`** to verify golangci-lint passes with new suite files | 1 min | Code quality gate |
| 8 | **Run `just test`** (ginkgo -r --cover) to verify all suites | 1 min | Test gate |

### Priority 3: Fix BuildFlow project

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 9 | **Update BuildFlow's own vendorHash** in `~/projects/BuildFlow/flake.nix` | 3 min | Unblocks `nix build` for BuildFlow |
| 10 | **Commit BuildFlow test dependency fix** (`registerStepDeps(StepTestCoverage, StepTestRace)`) | 2 min | Permanent fix in BuildFlow |
| 11 | **Add test for the parallel conflict** in BuildFlow's test suite | 15 min | Prevent regression |
| 12 | **Consider adding `.buildflow.yml`** to this project with documented settings | 10 min | Reproducibility |

### Priority 4: Code quality

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 13 | **Investigate `oxfmt` failure** — exits status 2 in local buildflow run | 10 min | Formatting gate |
| 14 | **Add Ginkgo suite entry point consistency check** — script/CI rule ensuring all packages with tests have Ginkgo presence | 30 min | Prevents future "missing suite" issues |
| 15 | **Convert remaining standard `testing.T` tests to Ginkgo** — 15 non-Ginkgo test files exist across packages. Full conversion would make all tests BDD-style. | 2-4 hours | Consistency, but low ROI |
| 16 | **Document the CGO/race/ginkgo interaction** in AGENTS.md gotchas | 10 min | Future agent knowledge |
| 17 | **Fix gopls warnings** — 3 `writestring` warnings in `detector_test.go` and `migrator_test.go` | 5 min | Code cleanliness |

### Priority 5: Strategic improvements

| # | Task | Effort | Impact |
|---|------|--------|--------|
| 18 | **Add coverage threshold gate** — BuildFlow has a TODO for failing if total coverage < 75%. This project is at ~65% average. | 1 hour | Quality enforcement |
| 19 | **Investigate dual golangci-lint binaries** — Tests warn about duplicate PATH entries (nix store + system) | 15 min | Test noise reduction |
| 20 | **Consider workspace-based test isolation** — Each ginkgo run could use a temp build dir to avoid compilation conflicts | 2 hours | Architectural improvement |
| 21 | **Add `--keep-going` to buildflow test runs** — So test-race failure doesn't block test-coverage visibility | 5 min | Better debugging |
| 22 | **Update `internal/cli/cmd_configure_internal_test.go`** — Still uses `package cli` (internal), not `cli_test`. Inconsistent with other CLI tests. | 30 min | Test organization |
| 23 | **Review if `test.golangci.yml` indentation change** (4-space → 2-space from buildflow formatting) is desirable | 5 min | Config consistency |
| 24 | **Consider adding `CGO_ENABLED=1` to CI devShell** as well, not just local | 5 min | CI race detection |
| 25 | **Run `brutal-self-review` skill** on the full session work for critical assessment | 15 min | Quality assurance |

---

## g) Top #1 Question I Cannot Figure Out Myself 🤔

**How should the fixed BuildFlow binary be deployed to replace the system-installed one?**

The system-installed buildflow is at `/run/current-system/sw/bin/buildflow` → `/nix/store/072cklhzs4c227xjk2azi0imn4x3il98-buildflow-e4c384f/bin/buildflow`. This is an immutable Nix store path managed by the system flake (likely `~/.config/home-manager` or `/etc/nixos`). 

Options I considered:
1. **`go install ./cmd/buildflow`** → Installs to `$GOPATH/bin` (`~/go/bin`), but this may not take precedence over the Nix system path in `$PATH`.
2. **Update the system flake** → Rebuild BuildFlow input in the system flake to point to the fixed commit. This is the "correct" Nix way but requires knowing where the system flake lives.
3. **`nix build` and symlink** → BuildFlow's `nix build` fails with vendorHash mismatch.
4. **Just use the `/tmp/buildflow-fixed` binary** → Works for testing but not persistent.

**What is the canonical way BuildFlow updates are deployed on this system?** Is it via home-manager, a system flake, or `go install`? This determines whether I should update a flake input, rebuild the system, or just `go install`.
