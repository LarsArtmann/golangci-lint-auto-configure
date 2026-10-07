# Memo: `Detector.Detect()` error-path contract

**Decision needed:** keep `Detect() ProjectType` (nil-on-error) or change to
`(ProjectType, error)`? **Recommendation: keep the current signature and
document the contract.** This completes the twice-deferred decision with a
documented keep — no API change.

## Context

Signature (`pkg/detection/detector.go:129`): `func (d *Detector) Detect() ProjectType`.
Behavior (intentional, per AGENTS #26 erraudit review): detection is a
best-effort heuristic classifier — per-file scanner errors are logged at
debug and swallowed so one bad file cannot abort project detection; walk-level
errors are captured and logged at the call site. `HasSwaggo`, the one
error-propagating consumer surface, already returns `(bool, error)` with
classified transient errors. Deferred twice (2026-07-26 f20/g2).

## Consumers

`cmd_configure_preset.go:39` (preset recommendation path) and two internal
callers (`RecommendPresets`, preset base). All treat the result as a starting
heuristic that the user can override with an explicit `--preset` flag.

## Options

| Option                        | Cost                                                                                              | Benefit                                                     |
| ----------------------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| **A. Keep + document (rec.)** | One AGENTS/docs paragraph (below)                                                                    | Zero churn; callers stay simple; heuristic semantics stay honest |
| B. `(ProjectType, error)`     | Signature break → 3 call sites + mocks + tests; every caller must handle errors it cannot act on (no retry, no user remedy beyond picking `--preset` themselves — which the flag already offers) | Errors visible in types. |

## Why A

"Detection failed" has no actionable remedy other than what already exists:
the `--preset` override. An error return would push non-actionable handling
into every caller — ceremony without behavior change. The swallow is not
accidental; it is the documented resilience property (one unreadable file
must not break configure for an entire project). What was missing is the
contract in writing — supplied by this memo and AGENTS.

## Contract (adopt verbatim into AGENTS if ratified)

`Detect()` never fails hard. It returns the best heuristic guess from
whatever was readable; unreadable inputs are logged and skipped by design.
Callers needing verification-with-errors use `HasSwaggo` (the classified
surface). Users override detection with `--preset`.

## Answer line

- [ ] A — keep + document (no code change)
- [ ] B — break the signature to `(ProjectType, error)`
