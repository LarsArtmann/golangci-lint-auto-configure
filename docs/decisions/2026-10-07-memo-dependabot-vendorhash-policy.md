# Memo: Dependabot vendorHash CI policy

**Decision needed:** how should `go_modules` Dependabot PRs pass the Nix
vendorHash drift check? **Recommendation: auto-repair + auto-commit via the
existing `vendorhash-guard.sh --fix`, with `contents: write` on
Dependabot-triggered runs only.**

## Context

Every `go_modules` Dependabot PR fails CI on exactly one step: "Build with Nix
(vendorHash guard)". Dependabot cannot update `vendorHash.nix` (it only knows
go.mod/go.sum), and the guard's `git diff --exit-code vendorHash.nix` drift
check turns the stale hash into a red build. Example: failed run 36342326659.
`github_actions` bumps pass fine (no go.mod involvement).

## Options

| Option                                                           | Mechanics                                                                                                                                                       | Tradeoff                                                                                                                                                               |
| ---------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **A. Guard --fix + commit on PRs (rec.)**                        | ci.yml nix job: when `github.event.pull_request.user.login == 'dependabot[bot]'` and guard --fix changes the file, commit it to the PR branch with GITHUB_TOKEN | First run red, commit triggers a second green run (standard two-run pattern). Token scope limited to `contents: write` and the guard only ever writes `vendorHash.nix` |
| B. `continue-on-error` on the vendorHash step for Dependabot PRs | Step failure doesn't fail the job on bot branches                                                                                                               | Silent gate erosion: an actually-broken hash merges green. Reject.                                                                                                     |
| C. Drop vendorHash from CI                                       | Build Nix only on master                                                                                                                                        | Loses the pre-merge Nix gate for human PRs too. Reject.                                                                                                                |

## Why A

The guard already does the repair (`--fix` is proven, CI-tested) — the only
missing piece is committing its output on bot branches where a human can't.
The blast radius is one generated file. Human PRs keep the strict
fail-with-instructions behavior (the guard prints the exact paste-in hash).

## Implementation sketch (≈15 lines in ci.yml)

```yaml
- name: Build with Nix (vendorHash guard)
  run: scripts/vendorhash-guard.sh --fix
- name: Commit vendorHash repair (Dependabot PRs only)
  if: >-
    failure() &&
    github.event.pull_request.user.login == 'dependabot[bot]'
  run: |
    git diff --exit-code vendorHash.nix || {
      git config user.name "dependabot-vendorhash-fix"
      git config user.email "noreply@github.com"
      git add vendorHash.nix
      git commit -m "fix: update vendorHash for go.mod bump"
      git push
    }
```

Requires job-level `permissions: contents: write` — but scoped by the `if`
condition to Dependabot branches only. (Alternative with zero token elevation:
leave A's commit step out and keep PRs red with instructions; a human
comments the paste-in hash and Dependabot rebase-picks it up. Honest but
keeps manual toil on every bump.)

## Answer line

- [ ] A — auto-repair + commit on Dependabot branches
- [ ] A' — auto-repair printed, human pastes (no token elevation)
- [ ] B/C — accept the stated tradeoff
