# Decision Packets — 2026-11-02 (T22/F69–F76)

One file, five-minute reads. Each packet states the decision needed, the
options, a recommendation, and the cost of deciding now vs. later. Decisions
stay with the owner (Lars); nothing here is executed unilaterally.

## F69 — BuildFlow findings-gate posture

**Question**: should the repo's (currently absent) `.buildflow.yml` declare
per-gate failure policy, or does the standing all-advisory posture stay?

**State**: the repo deliberately has no `.buildflow.yml` (AGENTS gotcha 14/31).
BuildFlow's findings gate fails only on `error`; the standing warnings
(lo rewrites, lll-forbidden assets, single-stage Dockerfile, dprint fallbacks)
were triaged 2026-10-13 and are intentionally not fixed. The 2026-10-17
incident (stale in-process binary reporting phantom findings) showed the gate
can be wrong for reasons outside this repo.

**Options**:

1. Keep no-config posture (status quo): warnings stay advisory, errors gate.
2. Add `.buildflow.yml` with `fail_on:` allowlists per gate.
3. Add `.buildflow.yml` with `skip_steps:` for the noisy advisory tools.

**Recommendation**: option 1. The gate has never produced a false red on this
repo post-triage; a config file adds a surface to maintain for no observed
failures. Revisit if BuildFlow changes its default severity semantics or a
warning graduates to error.

**Cost of deciding now**: zero. Cost of never deciding: an occasional ignored
warning.

## F70 — M32 never-enable replacement bypass ratification

**Question**: ratify the M32 behavior — when a sidecar `never-enable` lists a
deprecation's REPLACEMENT, the deprecated predecessor is still removed but the
replacement is not added and its settings are not migrated.

**State**: shipped and spec-pinned (`fixer_never_enable_test.go`), documented
in AGENTS gotcha 2a and FEATURES. Rationale: replacing a linter the user
explicitly blocked is an auto-enable by another name. The behavior has been
live through the 2026-11-02 fleet sweep with zero complaints (0/157 sidecars
in the fleet means it has never even fired in practice).

**Options**: ratify as-is; or flip to "replacement added anyway".

**Recommendation**: ratify as-is. The bypass has never fired in the fleet; a
behavior flip would be speculative.

**Cost of deciding now**: one sentence of confirmation.

## F71 — AGENTS.md size posture

**Question**: AGENTS.md is 175 lines (BuildFlow preflight warns above 220 —
currently UNDER the threshold). Prune, or keep growing?

**State**: the 2026-11-02 session added gotcha 35 (fleet sweep protocol) and
extended gotchas 4, 8, 20, 24, 33. Growth is driven by incident-derived
protocol knowledge — exactly what AGENTS.md is for. The BuildFlow warning
fires at 220 lines (386-line warnings seen in sibling repos).

**Options**: keep (status quo); prune the "Where to Find Detail" tail into
docs/; split per-topic reference docs and shrink gotchas to one-liners.

**Recommendation**: keep. 175/220 leaves headroom; the doc's job is
session-context density, and every gotcha cites its spec. Prune trigger: when
the 220-line warning fires, promote the longest gotchas' _detail_ into
docs/references/ and keep their one-line rule here.

**Cost of deciding now**: zero; the trigger is mechanical.

## F72 — Declined-verdict owner review (veto candidates)

The TODO_LIST carries these `declined-for-now`/`declined` verdicts (strikethrough
rows). Owner review = confirm or veto each; silence ratifies:

1. `--explain-disabled` flag — declined-for-now (transparency surface not
   worth it; DOMAIN_LANGUAGE + analyze output own the rationale).
2. README "Disabled Linters" section — declined-for-now (README stays
   user-level).
3. `--list-disabled` subcommand — declined-for-now (no demand).
4. Integration test asserting `noinlineerr` in CLI disabled output — declined
   (covered class-wise by T18 dispatch specs).
5. Documenting the disabled mechanism in working-with-codebase.md — declined
   (AGENTS gotcha 11 owns it).
6. presets.go DisabledLinters convention comment — declined (table-enforced).
7. `--disable-reason` field in JSON reports — declined-for-now (minimal
   report surface).
8. `enable-all` + `disable` moving `noinlineerr` — declined-for-now (no
   observed real-world enable-all config).

Full rationale lives in TODO_LIST.md rows 6–45 (2026-10-07 precedent noted).

## F73 — KnownBadSettingsKeys scope

**Question**: does `constants.KnownBadSettingsKeys` stay strictly scoped to
keys THIS tool emitted wrongly, or widen to upstream soft-deprecations?

**State**: the map today holds one entry (`goconst.min-length` → `min-len`,
the 2026-09-11 schema bug). The 2026-11-02 fleet sweep found 15 post-push
schema heals of pre-existing v1-era keys (wrapcheck `defaultSuffixes`,
forbidigo `p:`, `analyze_types`, gomodguard v1 shapes, exhaustive reshapes…)
— none were tool-emitted, so NONE entered the map (policy held). They were
healed by manual `fix(config)` commits, inventoried in
`docs/status/2026-11-02_13-12_fleet-sweep-complete-143-repos.md`.

The alternative scope (auto-healing upstream-renamed keys like the v2.15
exhaustive reshapes) would duplicate golangci-lint's own config migration and
the `migrate` subcommand's `RenamedLinterSettings` — a second migration path
that can drift from upstream truth.

**Recommendation**: keep the strict scope. The sweep recipes are the manual
path; the exhaustive reshape class belongs in `migrate` (already partially
covered), and the v2.15 binary-vs-jsonschema drift (old names in the
committed schema, new names enforced by the binary) is upstream's to fix —
tracked as the T12-adjacent issue.

**Cost of deciding now**: confirms a policy that held under fire.

## F74 — flake.nix lint app (implemented)

`nix run .#lint [args…]` now runs golangci-lint with the committed config from
the repo root, with CGO/GOEXPERIMENT/GOPRIVATE set and a stable cache dir.
Added this session; `nix flake check` green.

## F75 — docs-health check-rows UnicodeDecodeError (staged upstream)

The fix belongs in the `crush-config` repo (the skill's owning repo), not
here — per the cross-project protocol it must be fixed at the source and
fan-out-verified, which is its own session. Staged as an upstream item; the
repro is a check-rows run over a repo with non-UTF-8 bytes in a status file.

## F76 — erraudit quarterly re-check calendar note

VERIFIED PRESENT: TODO_LIST row 43 records the 2026-10-07 quarterly re-check
(81 findings, 0 new bugs, all accounted) and the next due date ~2027-01. No
action needed until then.
