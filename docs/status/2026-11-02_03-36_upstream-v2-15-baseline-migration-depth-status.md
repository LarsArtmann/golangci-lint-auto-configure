# Session Status: Upstream v2.15.0 Baseline — Migration Depth, Pin Bumps, Chain Fixture

> **Scope:** This session only — the resumption of
> `docs/status/2026-11-02_00-03_upstream-v2-audit-full-support-sweep-status.md`
> (the v2.0.0→v2.15.0 audit). That audit found and fixed supersession data,
> chain collapse, and two v2.15.0 breakages. This session executed the
> follow-up backlog that needed no user decision.
>
> **Concurrent-session note:** another session (00-18 invocation/OTEL
> self-review, see `docs/status/2026-11-02_00-18_…`) worked in this repo
> simultaneously. It rewrote `.golangci.yml` (a v1 experiment), edited
> `AGENTS.md`, `TODO_LIST.md`, `ROADMAP.md`, moved docs to `docs/archive/`,
> and created `pkg/telemetry/` (in flight — broke `go build ./cmd/...`
> transiently at 01:20, resolved when its files landed). I did not touch any
> of its in-flight files.

**Session summary:** migration-gap research verified against the official
guide AND the v2.15.0 JSON schema; schema snapshot regenerated at v2.15.0;
7 new migrate conversions + 3 explicit removals + exclusion-criterion
carry, proven end-to-end (`migrate` → `golangci-lint config verify`
LOADS-CLEAN on the v2.15.0 binary); all four CI pins + the
recommended-version constant bumped to v2.15.0; the wrapcheck 2.13 flood
(137 findings) diagnosed and a fix verified 137→0 on a config copy; new
parity fixture `exhaustivestruct-chain` pinning the one-run chain collapse;
read-only fleet scan quantifying blast radius. Tests green for everything
touched. The full suite, `nix fmt`, and `nix flake check` did NOT run this
session.

---

## Foreword: What did I forget / could do better / still improve?

**Forgot:**

- The CHANGELOG `[Unreleased]` "Added" edit — content written, but the edit
  was REJECTED because I read the file via `bash sed` instead of the View
  tool first. One edit away from done (see b1).
- The full test suite, `nix fmt`, `nix build`/`nix flake check` (skipped a
  second consecutive session).
- FEATURES.md/AGENTS.md/TODO_LIST.md updates for THIS session's work
  (previous session updated them for the audit; this session's additions
  are undocumented outside CHANGELOG-content-in-waiting).

**Could do better:**

- Probe deep-map structures with a throwaway internal test BEFORE writing
  conversion + spec — the gocritic nesting cost three debug cycles (d1).
- Verify exit codes before counting findings from external tools — my first
  "zero findings" read was a FAILED run (exit 3, the concurrent session's
  v1 config) that a grep silently misread as success.
- In multi-agent sessions, re-check `git status <file>` immediately before
  every edit, not just when contention is suspected (it worked for
  `ci.yml`/`version.go`; apply the discipline to docs too).

**Still improve:**

- The migrate subcommand still silently DROPS three structural user intents
  (`exclude-use-default`, `issues.exclude`, `output.format`) instead of
  converting them (c-item, question 3).
- CI has no canary that lints THIS repo with the NEWEST pinned binary, so
  the wrapcheck 2.13 flood sat undetected until a manual audit.

---

## a) FULLY DONE

1. **Research: v1→v2 migration ground truth, double-verified.** Fetched the
   official migration guide (golangci-lint.run/docs/product/migration-guide/)
   AND downloaded the v2.15.0 JSON schema
   (`jsonschema/golangci.jsonschema.json` at tag v2.15.0, 180 KB). Every
   mapping below is proven by the schema, not by the guide's prose:
   - `misspell.ignore-words` → `ignore-rules`; `predeclared.q` →
     `qualified-name` (renames)
   - `predeclared.ignore`: string → list (schema: `"type": "array"`)
   - `sloglint.context-only: true` → `context: all` (schema enum `["",
     "all", "scope"]`)
   - `govet.check-shadowing` gone; `enable` accepts `shadow`
   - `errcheck.exclude`/`errcheck.ignore`, `unused.exported-is-used`,
     `usestdlibvars.os-dev-null`/`syslog-priority`: ABSENT from the v2.15
     schema → must be dropped
   - gocritic ruleguard variable `${configDir}` → `${base-path}`
2. **Schema snapshot bumped to v2.15.0.** Committed
   `pkg/constants/schema/golangci-lint.jsonschema.json` replaced with the
   v2.15.0 capture; `pkg/constants/linter_settings_generated.go`
   regenerated with `-schema-version=v2.15.0` (provenance header was
   previously `"unknown"` — now stamped correctly). Drift: exactly one new
   key, `exhaustruct_v5.allow-empty-blank-assignments`; no curated default
   affected; derived `MinGolangCILintVersion` stays v2.12.0 (verified by
   the constants test run).
3. **All remaining v1→v2 migrate conversions implemented** (`pkg/migration/`):
   `rules.go` (+3 removal groups, +2 renames),
   `migrations_linter_settings.go` steps 9–12 (string→list, context-only,
   check-shadowing with `shadow`-already-enabled dedup, ${configDir}), and
   `config_types.go` + `migrations.go`: v1 `ExcludeRule` now carries
   `path-except` and `source` — previously the
   `issues.exclude-rules` → `linters.exclusions.rules` move SILENTLY LOST
   both criteria, which could leave a rule golangci-lint refuses to load.
4. **7 new Ginkgo specs; pkg/migration 51/51 green.** Table-style specs
   mirror the upstream guide per conversion, including negative cases
   (`context-only: false` → nothing, `shadow` not appended twice, valid
   settings survive alongside dropped ones).
5. **E2E load-proof under the real v2.15.0 binary.** Built the CLI,
   migrated a v1 config containing EVERY legacy key from the research
   list, and `golangci-lint config verify` on v2.15.0 returned clean
   (`LOADS-CLEAN`), with all conversions visible in the written config
   (`context: all`, `enable: [shadow]`, `ignore-rules`,
   `qualified-name`, `${base-path}`, `source:` carried).
6. **CI baseline moved to v2.15.0.** All four `golangci-lint-action`
   `version:` pins (test :64, lint :162, schema-verify :201, parity :364)
   and the schema-verify regen command + drift message (both
   `-schema-version=` occurrences) bumped v2.14.0 → v2.15.0.
7. **`ExpectedGolangCILintVersion` → v2.15.0** (`pkg/constants/version.go`,
   comment updated); version-related tests in pkg/constants + pkg/linter
   green (the specs reference the constant dynamically — no hardcoded
   value to chase).
8. **wrapcheck 2.13 flood diagnosed and fix VERIFIED 137 → 0.** The v2.15
   binary's wrapcheck 2.13 flags in-module cross-package returns from
   `pkg/errors`/`go-error-family` constructors (house pattern: classified
   errors are returned, not re-wrapped). Fix = `ignore-sig-regexps` entry
   `github\.com/larsartmann/(golangci-lint-auto-configure/pkg/errors|go-error-family)\.`
   — verified against a copy of the HEAD config (`/tmp/self-head.yml`):
   137 → 0. NOT yet applied to the real `.golangci.yml` (b2).
9. **New parity fixture `exhaustivestruct-chain`.** Input enables
   `exhaustivestruct` with a v1 `exclude` settings block; the golden pins
   the one-run supersession-chain collapse
   (exhaustivestruct → exhaustruct → exhaustruct_v5) byte-exactly,
   including settings ported across BOTH hops (`exclude` →
   `ignore-patterns`). Parity green WITHOUT `UPDATE_PARITY` (deterministic);
   corpus README fixture table row added. Corpus now 12 fixtures.
10. **staticcheck S1016 fixed in my own new code:** the exclusion-rule
    carry first used a struct literal; since v1 `ExcludeRule` and v2
    `ExclusionRule` are now field-identical, staticcheck (2.16 in the
    v2.15 binary) requires the conversion `ExclusionRule(rule)`. Applied.
11. **Fleet impact scan (read-only, ~/projects):** 4 real (non-vendor)
    configs carry superseded linter names and/or camelCase wrapcheck keys
    (`template-arch-lint/template-configs`, `go-localsync/provider/github`,
    `go-plugin-mvp/marketplace`, `plugmarket/marketplace`); **577**
    `go-error-family` consumers → the wrapcheck 2.13 flood WILL be
    fleet-wide when consumers bump to v2.15.0. Data recorded for the
    curated-default decision (question 2).

## b) PARTIALLY DONE

1. **CHANGELOG `[Unreleased]` "Added" section.** Three entries fully
   drafted (v2.15.0 baseline; the migrate conversions; the chain fixture)
   — the file edit itself was rejected (View-before-edit violation: I read
   via `bash sed`). Retrying the edit is step 1 of the next session.
2. **Self-repo lint green under v2.15.0.** Achieved on a COPY of HEAD's
   config. The real `.golangci.yml` is mid-experiment by the concurrent
   session (rewritten to a v1 fixture config — `golangci-lint run` exits 3
   against it right now). Applying the verified wrapcheck fix is blocked
   on that file settling (question 1).
3. **Docs for this session's work.** CHANGELOG drafted (b1); AGENTS.md
   (Where-to-Find-Detail item 6 schema-snapshot date, gotcha 20 migrate
   sentence, new wrapcheck-2.13 fact), FEATURES.md rows, and TODO_LIST.md
   refresh not started — the latter two files are contended by the
   concurrent session.
4. **Repo self-lint:** my code contributes ZERO findings on v2.15.0. 12
   findings remain in the tree — ALL in the concurrent session's in-flight
   files (`pkg/telemetry/*`: funlen, gci, gocritic, misspell ×2,
   nolintlint, spancheck, wsl_v5; `internal/cli/commands.go`: contextcheck
   ×2). Theirs to finish, not mine to fix.

## c) NOT STARTED

- **Full test suite this session** — only pkg/migration (51/51),
  pkg/constants + pkg/linter version-subset, and the parity corpus ran.
  The 18-suite ginkgo sweep is pending.
- **`nix fmt`, `nix build`, `nix flake check`** — skipped a second
  consecutive session. go.mod untouched (flake hash unaffected), but the
  "twice-skipped" pattern is becoming a real verification gap.
- **Upstream issue for v2.15.0's bare-dir `linters` probe semantics**
  (lists every linter under Enabled; poisons naive consumers). User-gated
  via verify-before-filing.
- **Upstream issue candidate: wrapcheck 2.13 in-module cross-package
  reporting** — determine whether the flood is intended behavior before
  relying on ignore-config forever.
- **Release v0.12.0** — user-gated; blockers 1–8 of the next-items list.
- **Fleet sweep execution** — scan done, mutations need approval.
- **Remaining structural migrate conversions** (see question 3):
  `issues.exclude-use-default` → `linters.exclusions.presets`,
  `issues.exclude` (plain list) → `exclusions.rules`,
  `issues.include` → presets, `output.format` variants →
  `output.formats` keyed structure. Current code CLEARS these keys
  without converting, silently changing the lint surface for
  `exclude-use-default: false` configs.

## d) TOTALLY FUCKED UP

1. **gocritic ${configDir} conversion written at the wrong nesting depth.**
   I read `gocritic.settings.ruleguard.rules` as
   `gocritic["settings"]["rules"]` — dead code that silently didn't fire.
   Worse, my SPEC made the identical mistake, and the public-API E2E
   harness output initially looked plausible. Three debug cycles (spec
   failure → public-API reproduction → internal type-print) exposed it.
   The internal test (`fixes=0` + printing the map) was the moment of
   truth. Lesson recorded: probe structure first, assert second.
2. **False "zero findings" claim.** My first post-fix lint read grepped a
   run that had actually FAILED (exit 3 — the concurrent session's v1
   `.golangci.yml`) and reported zero wrapcheck. Caught by re-running with
   exit-code visibility before drawing conclusions. A near-miss for a
   wrong status claim.
3. **Three attempts for one E2E run.** Attempt 1: `cat > .golangci.yml`
   heredoc chained AFTER a failing `&&` build (concurrent telemetry
   breakage) → directory/config never created, migrate ran against
   nothing, output confusing. Attempt 2: `git add -A && git commit`
   BEFORE the config file existed → commit failed → `cat` skipped again.
   Attempt 3: file-first sequencing → clean. Sloppy shell sequencing, not
   tooling.
4. **Edit-tool contract violation.** Read CHANGELOG via `bash sed`, then
   tried to edit → rejected. Wasted a round trip and left the deliverable
   incomplete at wait-time.
5. **`/tmp/mig-e2e` vanished between calls** (cause unknown — possibly
   tmp hygiene); had to recreate. Contributing factor: the mess in (3).

## e) WHAT WE SHOULD IMPROVE

1. **Exit-code gate before interpretation:** never count/summarize output
   from a non-zero-exit external tool run.
2. **Structure-probe-first for deep maps:** throwaway internal test that
   prints the nested types BEFORE writing conversion code and specs.
3. **Sequencing:** create files BEFORE fallible `&&` chains; prefer `;`
   after file-creating steps.
4. **Contended-file protocol:** `git status <file>` immediately before
   every edit in a multi-agent session.
5. **View tool for anything I will edit** — bash reads don't satisfy the
   contract and hide line numbers I'd need anyway.
6. **Schema-verified migrations as the bar:** every mapping in this
   session was proven against the v2.15.0 schema; the four structural
   conversions still missing deserve the same treatment, not prose-guesses.
7. **Newest-pin canary in CI:** add "run golangci-lint on this repo at the
   newest matrix cell" so floods like wrapcheck 2.13 surface at bump time,
   not audit time.
8. **Migration provenance:** `removedLinterSettings()` entries carry no
   source/version annotations; add WHY per entry (guide + schema version).

## f) 50 things to do next (prioritized; ← = blocked on user)

**Finish this sweep (do first):**

1. Apply the drafted CHANGELOG `[Unreleased]` Added-section edit (b1).
2. Apply the verified wrapcheck `ignore-sig-regexps` fix to
   `.golangci.yml` once the concurrent session settles; re-lint → 0
   findings from my code. ← q1
3. AGENTS.md: Where-to-Find-Detail item 6 (snapshot v2.14.0 → v2.15.0 +
   CI command), gotcha 20 append (migrate conversions complete),
   wrapcheck-2.13 fact + house fix.
4. FEATURES.md: migrate-conversions row + v2.15.0-baseline row.
5. TODO_LIST.md refresh: harvest items 1–50 below, close audit items done.
6. Full suite: `ginkgo -r -p -race` (all 18 suites) after 1–5.
7. `nix fmt` + confirm treefmt green on all touched files.
8. `nix build` + `nix flake check` — the twice-skipped gate.

**Migrate completeness (schema-verified, one PR each):**
9. `issues.exclude-use-default: false` → explicit empty
`linters.exclusions.presets` (preserve the "no default exclusions"
intent). ← q3
10. `issues.exclude` (plain string list) → `exclusions.rules` path rules.
11. `issues.include` → `linters.exclusions.presets` conversion.
12. `output.format` (line-number/colored-*/tab variants) → `output.formats`
keyed structure instead of silent drop.
13. `issues.exclude-generated(-strict)` → `linters.exclusions.generated`.
14. Migrate: emit a summary WARNING listing every key dropped with no
successor (`errcheck.exclude`, …) so loss is visible, never silent.
15. `scripts/e2e-pin-matrix.sh`: add "migrated config loads" step per cell.
16. Extend the parity corpus (or a migrate corpus) to cover the NEW
conversions — migrate currently has no golden snapshots at all.
17. Unify `normalizeLocalPrefixes` + `normalizeStringToSlice` into one
string→list helper.
18. Per-entry provenance (source version/link) for `removedLinterSettings`.
19. Table-driven weird-YAML-type test for `migrateLinterSettings` (string
where list expected, nil maps, etc.).
20. Split `migrateLinterSettings` (now 12 steps) per-linter before it hits
the nolint ceiling.
21. Consider `type ExcludeRule = ExclusionRule` alias now that the structs
are identical (staticcheck already exploits convertibility).

**Baseline hygiene:**
22. Docs-wide v2.14.0 → v2.15.0 reference sweep (README, FEATURES,
docs/references/*).
23. Curated wrapcheck default decision → if approved: policy-table change,
parity regen, schema-fixture regen, MinGolangCILintVersion re-check,
fleet rollout note. ← q2
24. Upstream issue: v2.15 bare-dir probe semantics (verify-before-filing).
25. Upstream issue candidate: wrapcheck 2.13 in-module flood intent.
26. Release v0.12.0 (after 1–8): changelog cut, tag, GoReleaser, proxy
verify, GHCR backfill. ← user-gated
27. BuildFlow flake-input refresh after push + `nix run .#reinstall` +
`BUILDFLOW_NO_RESULT_CACHE=1` verify (gotcha 32 — the stale-binary
loop will otherwise flag phantom findings).
28. CI drift guard: fail when ci.yml pin ≠ ExpectedGolangCILintVersion ≠
generated-file header version (single-source-of-truth test).
29. Schema-verify job: assert the generated-file header version matches
the `-schema-version` flag it was built with.
30. coverage-check run for pkg/migration after new code (`-min=65` gate).
31. Lychee link-check for the new CHANGELOG migration-guide link.
32. Differential test: upstream's own `golangci-lint migrate` vs ours on a
shared v1 fixture set (compatibility witness).
33. Spec: emptied multi-linter exclusion rule where only SOME linters are
superseded (partial rename keeps rule alive) — verify existing
behavior is pinned.
34. Spec: `normalizeStringToSlice` nil/empty handling.
35. Pin matrix: document why v2.11.x has no cell (below-min refusal
already proven at v2.10.1).

**Fleet (read-only scans are free; mutations need approval):**
36. Supervised configure run on the 4 flagged fleet configs. ← approval
37. Quantify wrapcheck-2.13 exposure across the 577 errorfamily consumers
(estimate configs needing the ignore entry).
38. Manual look at `template-arch-lint/template-configs/.golangci.yml`
(matched BOTH scans — superseded linters AND camelCase keys).
39. House snippet for the wrapcheck `ignore-sig-regexps` entry in
crush-config/docs so every repo adopts it identically.
40. Fleet-wide `migrate --dry-run` audit (read-only) for v1-only keys
beyond my scan list.
41. go-error-family docs: "returns are pre-classified — wrapcheck config"
pattern for consumers.
42. Post-sweep drift inventory: `configure --check` across all 160 repos.

**Docs/process:**
43. docs/references/release-process.md: "bump pins + schema snapshot +
ExpectedGolangCILintVersion together" checklist entry.
44. AGENTS.md gotcha candidate: wrapcheck 2.13 in-module cross-package
flagging + house fix.
45. Cross-link this report ↔ the 00-18 telemetry report (shared files:
`.golangci.yml`, `commands.go`, `TODO_LIST.md`).
46. docs/status/README.md index: add this report.
47. After telemetry lands: lint their files green (the 12 findings) —
coordinate, don't duplicate.
48. Watch golangci-lint releases ≥ v2.16.0; schedule the next audit
trigger at release time, not ad-hoc.
49. `migrate` command help text: document the new conversions + dropped
keys so users see the behavior without reading source.
50. Consider publishing the schema-capture procedure as a script
(`scripts/capture-golangci-schema.sh <tag>`) — it's now done twice by
hand.

## g) Questions I cannot answer myself

1. **`.golangci.yml` contention:** the concurrent session replaced it with
   a v1 experiment (plus telemetry wiring in `commands.go`). When it
   settles — may I apply the verified wrapcheck fix to whatever config
   lands, or should that fix ride their experiment to avoid a clobber
   war? (I did not touch the file; the verified fix sits in
   `/tmp/self-head.yml`.)
2. **Wrapcheck-2.13 fleet policy:** house `ignore-sig-regexps` for
   errorfamily/`pkg-errors` constructor returns — should it become a
   CURATED DEFAULT rendered into all ~160 fleet configs (rewrites rendered
   configs + parity goldens + schema fixture), or stay per-repo
   self-config? 577 errorfamily consumers say the flood is fleet-wide.
3. **Migrate structural conversions (q3 above):** the subcommand is
   maintenance-only, yet `exclude-use-default: false`, plain
   `issues.exclude`, and `output.format` are currently DROPPED silently —
   silently changing the user's lint surface. Convert them properly now,
   or keep drop-semantics and add a loud warning listing what was lost?

---

_Point-in-time snapshot 2026-11-02 03:36 CET — stale by design. Waiting for
instructions; items 1–8 are the pre-approved continuation default._
