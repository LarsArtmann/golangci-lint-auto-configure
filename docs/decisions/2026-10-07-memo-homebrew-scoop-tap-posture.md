# Memo: Homebrew/Scoop tap publishing posture

**Decision needed:** publish the wired-but-disabled tap blocks, or delete them?
**Recommendation: delete** unless there is an actual install-metric or user
signal demanding package-manager installs.

## Context

`.goreleaser.yaml` carries `homebrew_casks` and `scoop` blocks wired with
`skip_upload: true`, and `release.yml` has a PAT slot for tap publishing.
Neither has ever run. The GitHub Release page footer renders the
`brew install` / `scoop install` lines **even though they cannot work**
(upstream repos don't exist) — the release page currently advertises broken
install methods, which is the one genuinely bad part of keeping the blocks.

## Options

| Option                      | Cost                                                                                                                                                     | Benefit                                                            |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| **A. Delete blocks (rec.)** | ~30 min: remove both blocks + PAT slot + footer rows; one release cycle to see the clean page                                                            | Honest install story (`go install`, release binaries, GHCR docker) |
| B. Publish                  | Create + own `homebrew-tap` and `scoop-bucket` repos, flip `skip_upload`, supply PAT, test on real macOS/Windows, ongoing cask maintenance every release | `brew install` works                                               |
| C. Keep as-is               | None                                                                                                                                                     | Release footer keeps advertising installs that fail                |

## Why A

Taps without users are pure liability: two more repos, a PAT with repo-write
scope (supply-chain surface), and a per-release maintenance tax. The current
state (C) is dishonest-by-default. If a user ever asks for brew, B is a
half-day task — the blocks are recoverable from git history.

## Revisit

First brew/scoop request from a real user (issue, discussion, or download
pattern that suggests non-Linux use).
