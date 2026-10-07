# Memo: homepage + announcement posture

**Decision needed:** is this officially-maintained OSS or portfolio code?
**Recommendation: "portfolio-plus"** — public, best-effort support, no
announcements, homepage stays empty until a docs site actually ships.

## Context

One decision gates four queued items: the empty `homepage` repo field
(the standing ✗ in `scripts/metadata-check.sh`), the website launch
(website-launch pattern exists for sibling repos), any announcement
(r/golang, HN, X), and the community tier (issue/PR templates exist;
Discussions was already decided against). It also inherits into the
gohumanize strategy (see that memo).

## Options

| Option                                     | What it commits you to                                                                                | Homepage field                                  | Announcement  |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------- | ----------------------------------------------- | ------------- |
| **A. Portfolio-plus (rec.)**               | Issues welcome, best-effort triage (no SLA), PRs reviewed eventually. Templates already exist.        | Stays empty                                     | None          |
| B. Officially-maintained OSS               | Triaged issues with response expectations, release support promises, launch support, docs site upkeep | Docs-site URL                                   | r/golang + HN |
| C. Pure portfolio (archive-on-abandonment) | Nothing; repo is a code sample                                                                        | Empty; stale-content risk if abandoned silently | None          |

## Why A

The tool is real and used across ~160 sibling projects, but it is a
single-maintainer project. B's promises (response-time expectations, launch
support) are exactly the obligations that rot first under one maintainer.
A is honest: the repo stays genuinely maintained (it is — releases are cut,
CI is layered) without pretending to be a supported product. The
metadata-check homepage ✗ is reclassified as "expected" under A, or becomes
"all green" the day a docs site ships.

## Revisit

If a second regular contributor appears, or golangci-lint upstream links this
tool, B becomes viable — re-decide then.
