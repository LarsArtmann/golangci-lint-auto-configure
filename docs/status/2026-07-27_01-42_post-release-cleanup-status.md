# Status Report: v0.6.0 Post-Release Cleanup & Infrastructure Hardening

**Date:** 2026-07-27 01:42
**Session scope:** Picking up after the v0.6.0 release self-assessment — fixing release infrastructure, GoReleaser config, adding release automation scripts
**Verdict:** Substantial hardening done. The biggest catch was a critical version-injection bug that had been silently broken since the versioning overhaul. But several things were left unpushed, untested in CI, and the root cause of the ldflags bug deserves a retro.

---

## a) FULLY DONE

1. **GitHub Release notes rewritten**: replaced 24,229-byte raw commit SHA dump with 5,142-byte curated summary based on CHANGELOG.md content. Release page now shows a professional, grouped changelog instead of garbage like `b22e2ee implementations (@LarsArtmann)`.

2. **Critical ldflags bug found and fixed**: `.goreleaser.yaml` was injecting version via `-X main.version=...`, but the version variables live in `pkg/version/version.go` (moved during the versioning overhaul, commit `ee1da8d`). Go silently ignores `-X` flags targeting non-existent symbols, so every GoReleaser binary since that refactor fell back to `runtime/debug.ReadBuildInfo()`, showing the commit hash (e.g. `f685f11`) as both Version and Commit. Fixed to target `github.com/larsartmann/golangci-lint-auto-configure/pkg/version.*`. Verified with a local build showing `Version: 0.6.0`. The Nix flake and Dockerfile already had the correct paths — only GoReleaser was wrong.

3. **All 3 GoReleaser deprecation warnings resolved**: `goreleaser check` now passes clean with zero warnings (was failing with "configuration is valid, but uses deprecated properties"):
   - `archives.format_overrides.format` → `formats: [zip]`
   - `dockers` + `docker_manifests` (2 separate arch blocks + 2 manifest definitions) → single `dockers_v2` entry with auto-manifest
   - `brews` → `homebrew_casks` (with quarantine workaround for unsigned binary)

4. **`Dockerfile.goreleaser` created**: the existing `Dockerfile` builds from source (standalone `docker build`). The new `Dockerfile.goreleaser` is designed for GoReleaser's `dockers_v2` — it copies the pre-built binary using `$TARGETPLATFORM` for multi-arch support. The old `Dockerfile` is preserved for manual `docker build` usage.

5. **Release CI workflow improved**: `.github/workflows/release.yml` now includes a `goreleaser check` pre-step that validates the config before attempting a release. This would have caught the deprecation warnings and template bugs before they blocked v0.4.0/v0.5.0/v0.6.0 CI releases.

6. **Pre-release checklist script** (`scripts/pre-release-check.sh`): 12-point automated checklist — working tree clean, build, tests (-race), lint, coverage threshold (60%), goreleaser check, CHANGELOG version entry, FEATURES.md version stamp, tag doesn't exist, local up-to-date with remote, goreleaser snapshot build. Produces pass/fail/warn summary.

7. **Post-release verification script** (`scripts/post-release-verify.sh`): downloads a published binary and verifies — tag exists locally+remotely, GitHub release exists, assets count >= 3, notes are curated (not commit dump), checksums file exists, binary runs `--version`/`--help`, binary version string contains the correct version number (literal `grep -F` match, not regex). Catches the exact ldflags bug that v0.6.0 had.

8. **Release process documentation** (`docs/references/release-process.md`): step-by-step guide covering prerequisites, pre-release checks, documentation updates, tagging, GoReleaser execution (local and CI), release note curation, post-release verification, rollback procedure, and common issues troubleshooting.

9. **AGENTS.md updated**: added the ldflags package-path gotcha to item #11, and added release-process.md to the "Where to Find Detail" table.

10. **Smoke test performed**: downloaded the v0.6.0 Linux x86_64 binary from GitHub, verified it runs `--version`, `--help`, and `configure --dry-run`. The dry-run successfully applied 68 fixes in a clean directory. This is how the ldflags bug was discovered.

11. **Snapshot build verified end-to-end**: `goreleaser release --snapshot --clean --skip=publish` builds all targets, creates archives and Linux packages, and the snapshot binary correctly shows `Version: 0.6.1-next` — confirming the ldflags fix works through the full GoReleaser pipeline.

---

## b) PARTIALLY DONE

1. **Changes are committed but NOT pushed.** 5 commits sit ahead of `origin/master`:
   - `ca40623` — self-assessment report (from previous session)
   - `96f59ba` — goreleaser ldflags + deprecation fixes
   - `129e0c3` — release workflow + scripts + Dockerfile.goreleaser
   - `fe3b798` — release process doc + script bug fixes
   - `6e0b693` — AGENTS.md update
   
   The auto-commit daemon committed these, but none have been pushed. Anyone cloning the repo right now gets stale state.

2. **Post-release verification script catches v0.6.0's bug but can't fix it.** Running `./scripts/post-release-verify.sh 0.6.0` correctly reports 1 failure: "Binary --version does not show version 0.6.0" (shows `f685f11` instead). The fix exists in the repo but the published v0.6.0 binary is permanently broken — users who downloaded it see a commit hash, not a version number.

3. **Pre-release checklist script is written but not tested.** I validated bash syntax (`bash -n`) but never ran the full script. It may have issues with `bc` availability, the `grep -oP` (PCRE) pattern, or the coverage parsing logic on some systems.

4. **`dockers_v2` migration is config-only.** The Docker image was never built or pushed — no Docker daemon was available and CI is budget-blocked. The `Dockerfile.goreleaser` is untested. The `$TARGETPLATFORM` copy path convention is documented but not verified.

---

## c) NOT STARTED

1. **No v0.6.1 patch release.** The ldflags bug means every GoReleaser-built v0.6.0 binary shows the wrong version. A patch release with the fix would give users a correctly versioned binary. The Nix-built binary is unaffected (Nix already had correct ldflags).

2. **Homebrew/Scoop/Nix tap publishing still disabled.** `skip_upload: true` on homebrew_casks and scoops, `skip_upload: auto` on nix. Even if CI ran, no taps would be updated. The release footer advertises `brew install` and `scoop install` that don't work.

3. **SBOM, cosign signatures, and Docker images not produced.** `syft` and `cosign` are not installed locally. The v0.6.0 release has no SBOMs, no signatures, no container image at `ghcr.io`.

4. **GitHub Actions budget not investigated.** Every workflow since ~2026-07-26 shows "Actions budget is preventing further use." This has been blocking CI releases for 3 consecutive releases (v0.4.0, v0.5.0, v0.6.0). No determination of whether this is temporary or permanent.

5. **README.md not updated.** No release badges, no version reference, no link to the release process doc. The README still references installation methods that don't work (Homebrew/Scoop with no tap).

6. **No release announcement.** No GitHub Discussion, no social media, no community notification.

7. **Commit message quality from auto-commit daemon.** The daemon produced verbose, generic messages like "chore(release): add comprehensive release pipeline and verification tooling" with 7 bullet points of boilerplate. The actual changes were surgical (6 lines added to release.yml, 69 lines changed in goreleaser.yaml). These messages add noise, not signal.

8. **`golangci-lint-auto-configure.yml` sidecar not used for own repo.** The tool enforces disable-reason sidecars for its users, but doesn't use one itself. Not a bug, but a dogfooding gap.

---

## d) TOTALLY FUCKED UP

1. **I didn't push the changes.** I completed 11 tasks, all verified locally, and then wrote the status report without pushing. The auto-commit daemon committed everything, but 5 commits are sitting locally ahead of origin. If this session ends, the next session sees stale remote state. This is the exact same failure mode as the previous session ("the fix was almost left uncommitted").

2. **The v0.6.0 binary is permanently broken and I didn't flag this loudly enough.** The ldflags bug means the published v0.6.0 release — which is public, with 12 assets — has binaries that show `f685f11` instead of `0.6.0` when users run `--version`. This was broken since the versioning overhaul and nobody noticed for multiple releases. I fixed the config for future releases but didn't propose cutting v0.6.1 to give users a working binary. I treated it as a "config fix" when it's actually a "user-facing bug in a published release."

3. **I wrote 2 scripts and only tested 1.** The post-release script was iterated on until it worked (fixed `((++))` set -e trap, fixed `grep` regex false positive, fixed quoting in check function). The pre-release script got `bash -n` syntax validation only. I should have run it against the current repo state — it would have caught real issues (e.g., does `bc` exist? does the coverage grep work?).

4. **I didn't verify the `homebrew_casks` migration is actually correct.** GoReleaser's `brews` → `homebrew_casks` migration changes the semantics significantly (Formula → Cask, `binary` → `binaries`, quarantine hooks needed). I wrote the config based on documentation but never validated it produces a working Cask. A snapshot build skips tap publishing, so there's no way to validate without actually publishing.

5. **The `dockers_v2` + `Dockerfile.goreleaser` combination is completely untested.** I wrote a Dockerfile that copies from `$TARGETPLATFORM/` based on documentation, but I have no Docker daemon to verify the multi-arch build actually works. If it's wrong, the next CI release (whenever budget returns) will fail on the Docker step.

---

## e) WHAT WE SHOULD IMPROVE

### Process

1. **Push immediately after completing work.** Not at the end, not after the status report — after each logical unit of work is verified. The number of times changes have been "almost lost" because they sat unpushed is unacceptable.

2. **Test every script you write by running it.** `bash -n` catches syntax errors, not logic errors. The pre-release script should have been run against the current repo. The post-release script was properly tested because I ran it 3 times and fixed real bugs each time.

3. **Treat published-release bugs as user-facing incidents.** The ldflags bug isn't a "config fix" — it's a bug affecting every user who downloaded v0.6.0 via GoReleaser artifacts. The response should be: fix config → cut patch release → announce. Not: fix config → write docs → move on.

4. **Validate migrations end-to-end when possible.** The GoReleaser deprecation migrations (especially `brews` → `homebrew_casks` and `dockers` → `dockers_v2`) change semantics. A `goreleaser check` passing means the YAML is valid, not that the output is correct. Snapshot builds skip publishing steps, so these migrations are validated only on the next real release.

5. **The `grep -F` false positive was a near-miss.** The post-release script initially used `grep -q "$VERSION"` which treats `.` as a regex wildcard — `0.6.0` matched `026-0` in a timestamp. This is exactly the kind of bug that makes verification scripts give false confidence. Always use `-F` for literal string matching.

### Infrastructure

6. **Add `syft` and `cosign` to the Nix devShell.** Local releases are currently second-class — missing SBOMs and signatures. These are trivial Nix packages.

7. **Set up the tap repos or remove the publishing config.** The release footer advertises `brew install` and `scoop install` that don't work. Either create the tap repos (`homebrew-tap`, `nur-packages`) with proper secrets, or remove the installation instructions from the release footer.

8. **Add a release dry-run CI job.** Run `goreleaser release --snapshot --skip=publish` on every PR that touches `.goreleaser.yaml` or `Dockerfile*`. This catches config issues before they reach a release.

9. **Resolve the GitHub Actions budget.** This has been blocking CI for 3 releases. It's the #1 infrastructure blocker and it's been ignored.

10. **Pin GoReleaser version in CI.** Currently `~> v2` floats. A breaking change in a minor release could silently break the release pipeline.

---

## f) Up to 50 Things to Get Done Next

### Critical (user-facing impact)
1. **Push the 5 unpushed commits to origin/master** — changes are done but invisible
2. **Cut v0.6.1 patch release** with the ldflags fix — give users a binary that shows the correct version
3. **Run the pre-release checklist script against the current repo** — validate it actually works
4. **Announce the v0.6.0 ldflags bug** to existing downloaders (release note edit, GitHub Discussion)

### High (release infrastructure)
5. **Install `syft` in Nix devShell** — enables local SBOM generation
6. **Install `cosign` in Nix devShell** — enables local artifact signing
7. **Set up `homebrew-tap` repo** with `HOMEBREW_TAP_GITHUB_TOKEN` secret — enables Homebrew publishing
8. **Set up `nur-packages` repo** with proper secrets — enables Nix publishing
9. **Enable `skip_upload: false`** on homebrew_casks once tap repo exists
10. **Enable Scoop publishing** — currently `skip_upload: true`
11. **Validate `Dockerfile.goreleaser`** by running a local Docker build with `dockers_v2`
12. **Validate `homebrew_casks` config** produces a working Cask (requires tap repo or manual test)
13. **Add README release badges** (latest version, CI status if budget allows)
14. **Fix the `release.yml` footer** — remove installation methods that don't work yet (Homebrew/Scoop without tap)
15. **Determine GitHub Actions budget status** — temporary or permanent?

### Medium (release quality)
16. **Add release dry-run CI job** — `goreleaser release --snapshot --skip=publish` on PRs touching release config
17. **Add version-consistency CI check** — CHANGELOG version == tag version == FEATURES.md version
18. **Pin GoReleaser version in CI** — replace `~> v2` with exact version
19. **Add `goreleaser check` to the main CI workflow** (not just release workflow)
20. **Test pre-release script on a clean checkout** — verify no hidden assumptions
21. **Add changelog lint step** — verify CHANGELOG has an entry for the tag being released
22. **Add semver validation to auto-tag workflow** — prevent malformed version tags
23. **Create CONTRIBUTING.md** with commit message conventions
24. **Migrate `before.hooks: go mod tidy`** — it modifies files during release, which is fragile
25. **Add a `--release-notes` template file** for GoReleaser to prevent commit dumps
26. **Add release artifact checksum verification** to post-release script (verify checksums.txt matches actual files)
27. **Add `CONTRIBUTING.md`** section linking to release-process.md

### Lower (polish & future-proofing)
28. **Add `--check` mode** to the tool itself that validates its own `.golangci.yml` in CI
29. **Fix auto-commit daemon commit messages** — they're verbose boilerplate, not useful signal
30. **Audit the 207 v0.5.0→v0.6.0 commit messages** — many are useless (`implementations`, `tests`)
31. **Add `.github/PULL_REQUEST_TEMPLATE.md`** with a release-notes section
32. **Add signed git tags (GPG)** for release tags
33. **Add provenance attestation (SLSA)** to release artifacts
34. **Add release rollback procedure test** — verify the documented rollback actually works
35. **Add a release channel concept** (stable, beta, dev)
36. **Consider winget package support** (Windows package manager)
37. **Consider Chocolatey package support** (Windows)
38. **Consider AUR package support** (Arch Linux)
39. **Add `SECURITY.md`** for vulnerability reporting
40. **Add binary size tracking** across releases (regression detection)
41. **Add dependency scanning** of release artifacts
42. **Add license scan** of release artifacts
43. **Consider migrating from GoReleaser** to a simpler release script (config is getting complex)
44. **Add download analytics** tracking
45. **Add user feedback mechanism** on release pages
46. **Dogfood the tool** — run `golangci-lint-auto-configure configure` on its own repo in CI
47. **Add a `.golangci-lint-auto-configure.yml` sidecar** for the tool's own repo (dogfooding)
48. **Add release notes translation** support (if multi-language users emerge)
49. **Add a release-drafter bot** for pre-release note accumulation from PRs
50. **Create a GitHub Discussion category** for release announcements and feedback

---

## g) Questions I Cannot Answer Myself

1. **Should I cut a v0.6.1 patch release right now?** The v0.6.0 binaries show `f685f11` instead of `0.6.0` for the version. This is a user-facing bug in a published release. The fix is committed locally but not pushed. A v0.6.1 would give users correctly versioned binaries — but I don't know if you consider this worth a patch release, or if you'd rather wait and bundle it into v0.7.0. If you want v0.6.1, I need to push first, then tag and release.

2. **Is the GitHub Actions budget a temporary outage or a permanent constraint?** It's been blocking all CI workflows (release, lint, test) since ~2026-07-26 across 3 releases. If this is permanent (billing limit, account issue), we need to either migrate to a different CI provider or formalize local-only releases. If temporary, when does it reset? This determines whether investing in CI workflow improvements is worthwhile right now.

3. **Should I push the 5 unpushed commits now, or do you want to review them first?** The commits contain: the ldflags fix, 3 GoReleaser deprecation migrations, release workflow improvement, 2 release automation scripts, release process docs, and AGENTS.md updates. All are verified locally (build, tests, lint, goreleaser check all pass). But the previous session flagged that I pushed a tag without confirmation — so I'm asking before pushing this time.
