# Memo: Git history sanitization — purge or close forever?

**Decision needed:** keep pre-cleanup history as-is, or run a `git filter-repo`
purge of sibling-project references. **Recommendation: close as acceptable
forever.** No action.

## Context

Pre-cleanup history (before 2026-09-09) contains sibling-project _names_.
The security dimension is closed: full-history gitleaks scans (2026-09-11 over
1,117 commits; 2026-10-07 over 1,243 commits) found **zero secrets**. All
referenced sibling repos are themselves public. The only remaining exposure is
name-privacy: the history reveals which projects the author worked on and in
what order.

## Options

| Option                            | Cost                                                                                                                                                                                                                                            | Benefit                                                        |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| **A. Close as acceptable (rec.)** | None.                                                                                                                                                                                                                                           | No history rewrite; tags, cosign signatures, forks stay valid. |
| B. `git filter-repo` name purge   | Every commit SHA changes → v0.x tags must be re-created and re-signed; cosign/SBOM attestations reference old digests and go stale; all clones/forks diverge; force-push required; open PRs must be rebased; ~1–2h careful work + verification. | Names gone from history.                                       |

## Why A

The names are public repos discoverable from the author's GitHub profile
anyway. Filter-repo buys a privacy marginal anyone can defeat in one profile
click, at the cost of invalidating every signed artifact this project ships
(the supply-chain story is a core feature of v0.11+).

## Revisit

Only if a sibling repo goes private AND its name is sensitive. Then scope the
purge to that name only.
