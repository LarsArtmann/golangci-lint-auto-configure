# Memo: cross-repo write authorization (fleet sweep)

**Decision needed:** may the agent commit+push `configure` fixes to sibling
repos, and under which protocol? **Recommendation: standing authorization,
gated on a one-time rehearsal, in batches of ≤10, each batch reported.**

## Scope of the work waiting on this

Fleet audit 2026-10-07 (157 sibling configs):

- 148 carry patch-level `run.go` (pre-normalization; harmless today, breaks
  when golangci-lint tightens parsing, and is the tool's headline v0.11 fix)
- 21 fail `golangci-lint v2.14 config verify` outright (highest urgency)
- 109 carry dead `exhaustruct` in exclusion lists post-migration
- 8 repos have uncommitted `min-length` → `min-len` repairs from 2026-09-13

## Options

| Option                          | Cost                                                                                                     | Risk                                        |
| ------------------------------- | -------------------------------------------------------------------------------------------------------- | -------------------------------------------- |
| **A. Standing auth + rehearsal (rec.)** | One authorization now; each batch arrives as a report (repos touched, diff stat, verify results) | Mitigated by rehearsal + per-repo diff gate |
| B. Per-batch explicit go        | ~16 interruptions (157 repos / 10 per batch); the work stalls repeatedly                                 | Lowest                                       |
| C. Never (tool runs locally only) | Fleet stays broken; the tool's actual purpose (real-world configs) goes unexercised                     | Configs keep drifting                        |

## The non-negotiable protocol (guardrail 1, applies under every option)

1. Per repo: `git clone` into a scratch dir → run the freshly built binary's
   `configure` → `git diff` inside the clone only.
2. Assert the diff touches ONLY expected keys (run.go normalization, min-len,
   exhaustruct exclusion cleanup). Any unexpected hunk → repo skipped, flagged.
3. Spot-verify ≥5 diffs per batch by hand before any push.
4. `golangci-lint config verify` on every touched repo after push.
5. Batch report lands in this repo (`docs/status/`), repos touched listed
   explicitly.

## Answer line

- [ ] A — standing authorization with the protocol above
- [ ] B — ask before every batch
- [ ] C — never touch siblings from here
