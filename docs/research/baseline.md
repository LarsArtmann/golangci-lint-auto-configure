# Friction Baseline — 2026-07-25 (C0, before any changes)

Source: `/tmp/friction.py` over `/home/lars/projects/**/*.{go,yml,yaml}` (excluding this repo, vendored, and `*_test.go`).
This is the **before** snapshot. C18 will regenerate 5 sample configs with the new tool and re-measure to compute the delta.

## Per-linter friction (sorted by friction = nolint ÷ enabled)

| linter           | enabled | nolint | friction | verdict |
| ---------------- | ------: | -----: | -------: | ------- |
| exhaustruct      |     146 |    954 |     6.53 | HATED   |
| gochecknoglobals |     153 |    773 |     5.05 | HATED   |
| gosec            |     154 |    590 |     3.83 | HATED   |
| errcheck         |     143 |    443 |     3.10 | HATED   |
| wrapcheck        |     150 |    264 |     1.76 | HATED   |
| ireturn          |     136 |    171 |     1.26 | HATED   |
| recvcheck        |     148 |    147 |     0.99 | HATED   |
| contextcheck     |     149 |    139 |     0.93 | HATED   |
| exhaustive       |     154 |    129 |     0.84 | HATED   |
| funlen           |     151 |    126 |     0.83 | HATED   |

## Targets this plan addresses (DoD quantitative gates)

- exhaustruct: 954 → target **≥40% reduction** (≤ 572)
- gosec: 590 → target **≥25% reduction** (≤ 442)
- errcheck: 443 → target **≥20% reduction** (≤ 354)
- funlen: 126 → net reduction (200/100 default)

## Zero-friction linters (untouched — universally liked, nolint=0)

copyloopvar (154), errorlint (154), intrange (150), nakedret (150), nilnesserr (150),
durationcheck (149), gochecksumtype (149), sloglint (149), wastedassign (149),
loggercheck (148), mirror (148), paralleltest (148), protogetter (148), reassign (148),
ginkgolinter (146).

## Top absolute nolint (raw pain)

exhaustruct 954 · gochecknoglobals 773 · gosec 590 · errcheck 443 · wrapcheck 264 ·
ireturn 171 · recvcheck 147 · contextcheck 139 · exhaustive 129 · funlen 126 ·
nilerr 120 · forbidigo 113 · goconst 111 · gochecknoinits 88 · forcetypeassert 87.
