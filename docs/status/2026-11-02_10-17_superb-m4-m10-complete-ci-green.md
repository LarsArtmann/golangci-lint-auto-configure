# SUPERB M4-M10 Complete — Optional OTEL + 4/4 Audit Coverage + Master Green

**Date:** 2026-11-02 10:17
**Session:** Resumption of the interrupted 2026-11-02 OTEL session
(handoff: M1-M3 done, M4 ~50%, M5-M10 pending). This session executed M4
through M10 to completion. CI is green on the final commit `8bebd80`.
**Plan:** `docs/planning/2026-11-02_00-29_SUPERB-optional-otel-and-audit-coverage.md`
(gates checklist ticked with evidence; deviations recorded in §8).

## What landed (this session)

| Commit | Milestones | Content |
| --- | --- | --- |
| `98a74a1` | M4+M5 | Migrate OTEL span + audit coverage 4/4: `config-migrated` / `preset-applied` actions, `audit.OpenRunLedger`/`EnvVarDisable` (promoted from internal/cli's `newRunLedger`), migrate + preset wiring, `client.Options.Recorder`; binary-level ledger specs (XDG_CACHE_HOME-isolated) + client recorder spec |
| `044d75a` | M6+M7 | `SetFlagErrorFunc` → usage errors exit 1 (Rejection, `cli.flags` registered in `pkg/errors/codes.go`); exit-code pins (unknown flag → 1, `--check` changes-needed → 1) + installhook no-mutation spec |
| `156041a` | M8 | Docs: SECURITY (opt-in telemetry posture), README "Optional Telemetry" section, FEATURES rows, CHANGELOG, AGENTS gotcha 34 (telemetry design) + gotcha 17 refresh, ROADMAP OTEL decision, TODO_LIST row narrowed, plan errata, 00-18 false-claim annotation |
| `c01d745` | master-green | Converged own config for the concurrent session's tech-gated retraction normalizer (arangolint/clickhouselint out; CI dogfood gate had gone red on their `89a87b3`) |
| `8bebd80` | lint-red fix | Wrap-proof funlen restructure (see incident 2) — CI lint green |

## Verification (all green at final state)

- Focused specs 6/6: migrate ledger (exactly one `config-migrated` on non-dry; dry-run none),
  preset ledger (`preset-applied` recorded; check-mode none), unknown-flag exit 1,
  changes-needed exit 1, installhook writes only `.git/hooks/pre-commit`.
- Full `ginkgo -r -p -race --skip-package=cmd/` (19 suites) + `go test ./cmd/...`: green.
- golangci-lint on changed packages: 0 findings (after incident 2's re-fix).
- `nix fmt`: no-op; `nix build`: exit 0 (go.mod unchanged since f0726ef; vendorHash in sync).
- Dogfood: `--help` lists `--otel`; `configure --check` exit 0 after c01d745; earlier
  `--check --diff` runs restored the config byte-identically every time.
- CI `ci.yml` on `8bebd80`: success (4m47s) — lint + dogfood + schema gate + e2e pin matrix
  + parity corpus + nix flake check all green.

## Incidents this session (fixed)

1. **`cli.flags` unregistered** — full suite caught `TestEveryUsedCodeIsRegistered`;
   added to `pkg/errors/codes.go`. (Focused-package runs alone would have missed it.)
2. **CI lint red on my funlen fix** (`8bebd80`'s predecessor): the auto-format daemon
   (golines, max-len < 120) re-wrapped my single-line calls AFTER my lint run and BEFORE
   my pathspec commit picked up the working tree — re-inflating three functions past
   funlen 30. Lesson encoded in the commit message and applied: run `nix fmt` to a no-op
   BEFORE linting, then commit the linted bytes; structure functions ≤25 lines so wrapping
   has headroom.
3. **Master red from the concurrent session's `89a87b3`** (retraction normalizer without
   self-repo convergence; CI dogfood step demands `configure --check` exit 0). Fixed
   forward via the tool's own self-heal (`configure` on the repo, c01d745). The failed
   runs (41392780812, 41412798632) predate it; two intermediate runs were cancelled by
   newer pushes.
4. **Shared `/mnt/buildcache` link corruption twice** + `/tmp` at 100% (48G tmpfs):
   phantom "[Compilation failure]" and "no space left on device". All heavy gates ran
   with `GOCACHE=/dev/shm/glac-gocache TMPDIR=/dev/shm/glac-tmp`.

## Not mine (concurrent session, still in flight at 10:17)

Working tree carries their probe-hermeticity work (`pkg/linter/probe*.go`, `normalize.go`,
AGENTS gotcha 8 rewrite — the v2.15.0 "lists everything" theory was falsified; real cause
was ambient `/tmp/.golangci.yml` contaminating the probe), plus AGENTS/CHANGELOG edits.
Left untouched per the handoff; the daemon commits their work.

## Open (user-gated, unchanged)

- OTEL gRPC transport / metrics / logs: parked (ROADMAP records the decision).
- Audit run-lifecycle events (no-op/check/analyze visibility): TODO_LIST row, noise/retention tradeoff.
- CI telemetry hard-block policy: default remains zero-egress-by-default (SECURITY.md documents it).
