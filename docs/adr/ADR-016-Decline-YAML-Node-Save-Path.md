# ADR-016: Decline yaml.Node Comment-Preserving Save Path (For Now)

**Status:** Declined (revisit condition below)\
**Date:** 2026-10-07

## Context

`SaveConfig` unmarshals the config into `types.Config`, mutates the struct,
and re-marshals — so **user comments and blank-line separation are dropped on
every write**. Users notice when the tool reformats their hand-tuned configs.
A `yaml.Node`-based round-trip was proposed as the comment-preserving
alternative (plan M24).

## Spike Evidence (2026-10-07, measured on a commented fixture)

Pure `yaml.Node` unmarshal → mutate one scalar → re-marshal:

| Property                     | Result                                                        |
| ---------------------------- | ------------------------------------------------------------- |
| Head comments (section)      | **Preserved (4/4)**                                           |
| Line comments (value suffix) | **Preserved** (`go: "1.27" # patched by tool`)                 |
| Blank lines between sections | **Lost** — encoder compacts the document                       |
| Indentation                  | Encoder default is 4-space; `SetIndent(2)` restores 2-space     |
| Scalar mutation in place     | Works (value + tag + style updatable, comment retained)         |
| Sequence rendering           | Changes shape vs `SetIndent` settings; needs per-case verification |

## Why Declined

The spike proves **preservation**, but the save path's real workload is not a
round-trip — it is a **deep merge**: the fixer changes dozens of keys, adds
enable-list entries, injects settings maps, and appends exclusion rules across
`types.Config`. Doing that against a `yaml.Node` tree means re-implementing
structural merge (key positioning, sequence dedup, settings map ordering)
that the struct round-trip currently gets for free from YAML marshaling.

That merge layer is exactly the kind of rewrite that breaks 160 fleet configs
in subtle ways (ordering, quoting styles, sequence indentation) for a
cosmetic gain — the highest Verschlimmbesserung risk in the save path.

## Consequences

- Comments remain dropped on `configure` writes. The audit ledger +
  `rescued-run-go`/backup behaviors already mitigate the surprise.
- The `--force-settings` and idempotency machinery are unaffected.

## Revisit When

- A user files a real issue about comment loss (demand signal), or
- golangci-lint ships a Node-based config API we can lean on, or
- A contributor prototypes the deep-merge layer against the full fixer test
  suite (not just a round-trip) — the spike above is the starting point.
