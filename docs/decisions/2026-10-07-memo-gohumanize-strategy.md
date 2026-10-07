# Memo: gohumanize recommendation strategy

**Decision needed:** who should get the `gohumanize` linter recommendation?
**Recommendation: keep the current dep-gated project-specific posture.**
No code change.

## Context

`gohumanize` is a golangci-lint v2 _module plugin_: stock golangci-lint cannot
run it. A config that enables it requires a custom binary built via
`golangci-lint custom` + `.custom-gcl.yml`. The tool currently recommends it
ONLY when the project's `go.mod` imports `github.com/dustin/go-humanize`, and
emits only the bare `linters.enable: [gohumanize]` entry — never the
`linters.settings.custom` block (which would break stock binaries at load).
Analysis: `docs/status/2026-08-05_04-14…md`. Blocked since 2026-08-05.

## Breakage matrix

| Posture                          | Stock-binary user                                                                                                                         | Custom-binary user            | Maintenance                                                                    |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- | ------------------------------------------------------------------------------ |
| **A. Dep-gated (current, rec.)** | No recommendation → no break                                                                                                              | Gets recommendation, works    | One gate map entry + detector method                                           |
| B. Everywhere                    | Config enablement for a linter their binary can't load → `golangci-lint` **refuses to load the config** unless they build a custom binary | Works                         | Requires a documented custom-binary story + docs for every recommended project |
| C. Drop recommendation entirely  | No break                                                                                                                                  | Loses curated H001–H007 rules | One-time removal                                                               |

## Why A

B converts a curated nicety into a load-time hard failure for anyone who
hasn't read the plugin docs — the exact failure class this tool exists to
prevent. A scopes the risk to projects already depending on `go-humanize`
(highest prior that they run a plugin-capable setup). C throws away working
value for users who already have the binary.

## Revisit

If golangci-lint ever bundles gohumanize (watch release notes), flip to
B-trivially: remove the gate, keep the recommendation.
