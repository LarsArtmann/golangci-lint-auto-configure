# Release Process

This document describes the step-by-step process for cutting a release of golangci-lint-auto-configure.

## Prerequisites

- Go 1.26+ with `GOEXPERIMENT=jsonv2` set
- GoReleaser v2 installed (`nix develop` provides it)
- `gh` CLI authenticated
- `cosign` and `syft` installed (for signed releases with SBOMs)
- Docker installed (for container image publishing)
- Write access to the GitHub repository

## Pre-Release

### 1. Run the pre-release checklist

```bash
./scripts/pre-release-check.sh [version]
```

This script verifies:

- Working tree is clean
- Build, tests (-race), and lint pass
- Coverage meets the 60% threshold
- `goreleaser check` passes (no deprecation warnings)
- CHANGELOG.md has an entry for the target version
- FEATURES.md version stamp matches
- Tag does not already exist
- Local branch is up to date with origin
- Goreleaser snapshot build succeeds

**Do not proceed if any critical check fails.** Fix the issue first.

### 2. Update documentation

- Ensure CHANGELOG.md has the version under `## [X.Y.Z] - YYYY-MM-DD`
- Update FEATURES.md `**Version:**` line
- Clean TODO_LIST.md of resolved items
- Verify README.md is accurate

### 3. Commit and push all changes

```bash
git add -A
git commit -m "release: prepare vX.Y.Z"
git push origin master
```

Wait for CI to pass (if Actions budget allows).

### 4. Get explicit confirmation

**Never push a tag without explicit confirmation.** Tags trigger releases and are effectively irreversible. Present the release plan to the user and get a clear "yes" before proceeding.

## Cutting the Release

### 5. Create and push the annotated tag

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

This triggers the GitHub Actions release workflow (if budget allows).

### 6. If CI is unavailable, run GoReleaser locally

```bash
export GOEXPERIMENT=jsonv2
export GITHUB_TOKEN="$(gh auth token)"

goreleaser release --clean --skip=sbom,sign,docker
```

Add `--skip=nix,homebrew,scoop` if tap repos are not set up.

For a full signed release (requires cosign + syft):

```bash
goreleaser release --clean
```

### 7. Curate release notes

After the release is published, replace the auto-generated commit dump with curated content from CHANGELOG.md:

```bash
gh release edit vX.Y.Z --notes-file /tmp/release-notes.md
```

The release notes should be a curated summary, not a raw commit log. Use the CHANGELOG.md content as the basis.

## Post-Release

### 8. Run the post-release verification

```bash
./scripts/post-release-verify.sh X.Y.Z
```

This script verifies:

- Tag exists locally and on remote
- GitHub release exists with assets
- Release notes are curated (not a commit dump)
- Checksums file exists
- Downloaded binary runs and shows the correct version

### 9. Verify installation methods

- Check that archives, packages (.deb/.rpm/.apk), and checksums are present
- If Homebrew/Scoop/Nix publishing is enabled, verify the tap repos were updated
- If Docker publishing is enabled, verify the image was pushed to ghcr.io

### 10. Update project documentation

- Verify FEATURES.md version is correct
- Update TODO_LIST.md with any follow-up items
- Update AGENTS.md if any release-related gotchas were discovered

## Rollback

Tags and GitHub releases are difficult to fully revert. If a release has a critical bug:

1. **Do not delete the tag.** Users may have already pinned to it.
2. Mark the release as a prerelease or draft on GitHub.
3. Cut a patch release (vX.Y.Z+1) with the fix.
4. Add a note to the broken release's notes pointing to the patch.

## Common Issues

### GoReleaser fails with "syft not found"

Install syft: `nix profile install nixpkgs#syft` or skip SBOMs with `--skip=sbom`.

### GoReleaser fails with "cosign not found"

Install cosign: `nix profile install nixpkgs#cosign` or skip signing with `--skip=sign`.

### GoReleaser fails on Docker step

Ensure Docker daemon is running. Or skip with `--skip=docker`.

### Binary shows commit hash instead of version

This means the ldflags in `.goreleaser.yaml` are not targeting the correct package path. The version variables live in `github.com/larsartmann/golangci-lint-auto-configure/pkg/version`, not `main`. Verify the ldflags section of `.goreleaser.yaml`.

### GitHub Actions budget exhausted

If Actions are unavailable, run GoReleaser locally (step 6 above). All artifacts can be published from a local machine with the `GITHUB_TOKEN` environment variable set.

### `goreleaser check` fails on deprecations

Fix the deprecated config properties before releasing. See [GoReleaser deprecations](https://goreleaser.com/deprecations) for migration guides.

### Release run died after the tag was pushed (do NOT move the tag)

If the release workflow fails after `git push --tags` (the v0.11.0 case:
cosign signing broke after every artifact had already been built and
uploaded), the tag must NOT be deleted or re-pushed — the Go module proxy
may already have indexed it, and a moved tag poisons the version forever.
The proven local fallback:

1. Verify the tag is on master and master is green: `git log --oneline -1`
   and `gh run list --branch master --limit 3`.
2. Publish the GitHub release locally, skipping the steps that broke:
   `GORELEASER_CURRENT_TAG=vX.Y.Z GITHUB_TOKEN=$(gh auth token) goreleaser
   release --clean --skip=validate,sbom,sign,docker` (pick the skip list to
   match what already succeeded — GoReleaser re-uploads artifacts idempotently
   but does not deduplicate across its own partial runs).
3. Curate the release notes: `gh release edit vX.Y.Z --notes-file <file>`
   (GoReleaser's changelog dump is not the shipping notes).
4. Backfill anything skipped via dedicated workflows: dispatch
   `backfill-image.yml` with `tag=vX.Y.Z` for the GHCR image; SBOM/signature
   backfill runs the same GoReleaser command with only those steps from a
   machine with the required credentials.
5. Verify end-to-end: `./scripts/post-release-verify.sh X.Y.Z` (signature
   bundles, SBOMs, GHCR tag, cask, binary smoke test) and a clean
   `go get github.com/larsartmann/golangci-lint-auto-configure@vX.Y.Z`
   from outside the module.

### Cosign v3 bundle format

Releases are signed with cosign v3's REQUIRED bundle format: each artifact
gets a `<artifact>.sigstore.json` bundle containing the signature, the
Fulcio certificate, and the transparency-log material. There are no separate
`.sig`/`.pem` assets anymore. Verification:

```sh
cosign verify-blob \
  --bundle <artifact>.sigstore.json \
  --certificate-identity-regexp '^https://github.com/LarsArtmann/golangci-lint-auto-configure/\.github/workflows/release\.yml@[a-f0-9]+$' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  <artifact>
```

The signs config passes no service URLs: v3 fetches Fulcio/Rekor endpoints
from the TUF signing config, and explicitly passing them alongside it is the
error that broke the v0.11.0 release ("cannot specify service URLs and use
signing config").
