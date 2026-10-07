# Memo: daemon-mangled commit messages (mid-2026-07)

**Decision needed:** rewrite, annotate, or leave the garbled commit messages?
**Recommendation: leave as-is.** No action.

## Context

The auto-commit daemon produced truncated/garbled messages on a few mid-2026-07
commits (`5b91141`, `df5f006`, `dc2b0d6` era). The daemon's message generator
has since been fixed; only the historical artifacts remain. The affected
commits are ordinary "chore: auto-commit" noise — no content is undocumented
that CHANGELOG/docs don't cover.

## Options

| Option                    | Cost                                                                                                  | Benefit                                |
| ------------------------- | ------------------------------------------------------------------------------------------------------ | -------------------------------------- |
| **A. Leave (rec.)**       | None.                                                                                                  | History stays immutable.               |
| B. `git notes` annotations| Notes are local-only by default and do not push with normal flows; every clone needs `git fetch origin 'refs/notes/*:refs/notes/*'`; effort for near-zero readers. | In-place explanation without SHAs changing. |
| C. Rebase/filter rewrite  | All descendant SHAs change → tags re-created/re-signed, cosign attestations stale, forks diverge, force-push. | Cosmetic fix nobody asked for.         |

## Why A

The messages are noise on noise commits; the information value of fixing them
is near zero, and C destroys the signed-artifact chain (same reasoning as the
history-sanitization memo). B costs ongoing fetch ceremony for a benefit only
visible to someone running archaeology — who can read the surrounding docs.

## Revisit

Never for its own sake. Fold into the history purge memo only if that
decision flips (one rewrite is marginally cheaper than two).
