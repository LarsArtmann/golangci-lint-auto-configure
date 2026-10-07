# Memo: buildflow findings-gate posture

**Decision needed:** should buildflow's e2e step gate (fail) on its advisory
findings, and should the repo carry a `.buildflow.yml` to encode that?
**Recommendation: keep "no `.buildflow.yml`" (as decided 2026-09), treat e2e
advisory failures as expected-and-documented, CI remains the only gate.**

## Context

A local buildflow e2e run currently ends red on two advisory gates only:
branching-flow (~190 findings) and erraudit (44, manually reviewed per AGENTS #26 — 0 real bugs across three quarterly reviews). Test/lint/fmt steps are
green. There is no `.buildflow.yml` — an earlier one was removed deliberately
(AGENTS gotcha 13: treefmt owns formatting). No buildflow job exists in CI;
`ci.yml` is the authoritative gate. BuildFlow has no step-level timeout
option (`step_options.go`), so e2e timing is bound by `go test -timeout=10m`.

## Options

| Option                                    | Cost                                                                                          | Benefit                                                       |
| ----------------------------------------- | ----------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| **A. No config, documented red (rec.)**   | None beyond this memo; AGENTS records "e2e advisory red = expected"                              | No config surface to maintain; no duplicate of AGENTS #26 policy |
| B. Author `.buildflow.yml` with `fail_on:` pinned to build/test/lint/fmt | Must learn/maintain the schema; re-introduces a file that was removed on purpose; policy now lives in two places | Local runs exit 0 honestly instead of always-red-expected      |
| C. Fix the 234 findings to green the gate | Already triaged: 0 real bugs (erraudit #26); branching-flow is advisory by design                  | Not available at any reasonable cost                            |

## Why A

The gate that matters (CI) already encodes the real policy: tests, lint,
coverage 65, dogfood, pin matrix. Buildflow is the local convenience loop;
its e2e red is informative (the numbers moved) not gating (nothing blocks).
B would make local green *by configuration* — the same result as ignoring
the red, with a new file to keep in sync with BuildFlow releases.

## Answer line

- [ ] A — keep no-config, document expected-red
- [ ] B — author `.buildflow.yml` (I'll learn the schema and write it)
