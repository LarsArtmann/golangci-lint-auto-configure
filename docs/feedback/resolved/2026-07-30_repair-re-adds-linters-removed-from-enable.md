# FEEDBACK: `repair` re-adds linters removed from `enable` when they're absent from both `enable` and `disable` — recurring regression loop with BuildFlow

> **Severity:** High — silently overrides committed config on every commit, breaks CI, wastes hours across 7+ sessions
> **Reported:** 2026-07-30
> **Reporter:** Lars Artmann (via Crush session)
> **Tool:** `golangci-lint-auto-configure` (invoked via BuildFlow pre-commit hook as `golangci-lint [root]:repair`)
> **Project affected:** `templ-components` (`github.com/larsartmann/templ-components`, 15 Go packages, golangci-lint v2)

---

## Summary

The previous feedback (2026-07-20, resolved 2026-07-25) fixed the case where
`repair` **rebuilt `disable` from scratch**, dropping user-disabled linters. The
fix made `repair` respect `linters.disable` and never re-add a linter listed
there.

**This is a different, unresolved scenario:** when a user removes a linter from
`enable` using the **standard golangci-lint v2 way** (simply omitting it — the
golangci-lint docs say linters default to disabled unless listed in `enable`),
and does NOT also add it to `disable`, the `repair` step treats the omission as
"user forgot" and re-adds it to `enable` on the next run. This creates an
infinite regression loop: user removes → repair re-adds → test catches → user
removes again → repair re-adds again.

---

## Reproduction

### The project: templ-components

This is a **templ UI component library** where three linters are fundamentally
incompatible:

| Linter | Why it's incompatible |
|--------|----------------------|
| `godoclint` | Demands exactly one `// Package` godoc per package; the repo intentionally documents per-file |
| `ireturn` | Every component returns `templ.Component` (an interface) by design; the linter's premise is antithetical to templ |
| `testableexamples` | `Example*` funcs render verbose HTML output that isn't asserted; noisy and version-dependent |

### Step 1 — User removes the linters from `enable` and deletes orphaned settings

The `.golangci.yml` `enable:` list has these three removed. The `ireturn:`
settings block is also deleted (it has no effect once ireturn is disabled).

The `disable:` list currently contains only `depguard`:

```yaml
  disable:
    - depguard
```

### Step 2 — BuildFlow pre-commit `repair` re-adds all three

On the next `git commit`, BuildFlow runs `golangci-lint-auto-configure repair`
(or `golangci-lint --fix`). The tool re-adds all three to `enable` and restores
the `ireturn:` settings block:

```diff
     - gocyclo
+    - godoclint
     - godot
     ...
-    - loggercheck
+    - ireturn
+    - loggercheck
     ...
+    - testableexamples
     - testifylint
     ...
+    ireturn:
+      allow:
+        - error
+        - empty
+        - anon
+        - stdlib
+        - generic
```

### Step 3 — CI fails

The project has a guard test (`TestGolangciDisabledLinters` in
`utils/lint_config_test.go`) that checks `.golangci.yml` for these three
linters in the enable list. It fails with 4 errors (3 linters + the `ireturn:`
settings block).

### Step 4 — User removes again → go to Step 2

This loop has occurred **7+ times** across development sessions in this project
alone. Each time, the developer must re-remove the linters, re-run tests, and
hope the commit goes through before the daemon reverts.

---

## Why the previous fix doesn't solve this

The 2026-07-25 fix made `repair` respect `linters.disable`. But in this project,
the three linters are **not in `disable`** — they were simply removed from
`enable`. In golangci-lint v2, removing a linter from `enable` is the
**documented way to disable it** (a linter is disabled by default unless listed
in `enable`). The user should not have to also list it in `disable` to signal
intent.

The tool's three-tier system in `pkg/constants/rules.go` does not include
`godoclint` or `testableexamples` in any tier:

| Tier | Linters | Behavior |
|------|---------|----------|
| `DisabledLinters` | `funcorder`, `noinlineerr`, `depguard` | Forcibly moved to `disable`, never recommended |
| `NeverAutoEnableLinters` | `exhaustruct` | Never auto-enabled, respected if manual |
| `PragmaticNoiseLinters` | `gochecknoglobals`, `wrapcheck`, `ireturn`, `funlen` | Enabled by default, dropped only with `--pragmatic` |

`ireturn` is in `PragmaticNoiseLinters` (enabled by default), and
`godoclint`/`testableexamples` are in no special tier at all — so the tool
treats them as "should be enabled" and re-adds them.

---

## What I expected

One or more of:

**(a)** `repair` does not re-add a linter to `enable` unless it was explicitly
removed from a known-good baseline by the tool itself. Removing a linter from
`enable` is an intentional act; treating it as an oversight is wrong by
default.

**(b)** `repair` detects the regression loop: if a linter was in `enable`,
removed by the user, and re-added by `repair`, and removed again — the tool
should stop re-adding it and emit a warning suggesting `disable` instead.

**(c)** The tool recommends adding project-specific incompatible linters to
`disable` proactively: when a linter is absent from both `enable` and `disable`
and the project has committed `.golangci.yml` without it for N commits, suggest
adding it to `disable` with a reason.

**(d)** A project-level sidecar (e.g., `.golangci-lint-auto-configure.yml`)
that lists linters to never re-add, with a required reason field. The existing
sidecar (for policy enforcement per AGENTS.md) could be extended.

**(e)** Add `godoclint` and `testableexamples` to a new tier or to
`NeverAutoEnableLinters`. These are niche linters with high false-positive
rates in certain project types (templ libraries, test-heavy projects). Unlike
`ireturn` (which is debatable), there is no scenario where `godoclint` adds
value to a templ project.

---

## Workaround (current)

The project can add all three to `linters.disable`:

```yaml
  disable:
    - depguard
    - godoclint
    - ireturn
    - testableexamples
```

This should prevent `repair` from re-adding them (per the 2026-07-25 fix). But
this is undocumented — the developer had to read the resolved feedback to learn
that `disable` is the signal the tool respects. Most users would simply remove
from `enable` and hit the same loop.

---

## Impact

- **7+ regression cycles** across development sessions in this project
- Each cycle costs 5-15 minutes (diagnose, fix, verify, commit, discover
  reversion, repeat)
- The CI guard test (`TestGolangciDisabledLinters`) catches it, but only after
  the commit — the working tree is already dirty
- The project's AGENTS.md has a 200+ word section documenting this regression,
  3 prevention layers (test + script + CI), and it STILL happens because the
  tool keeps re-adding the linters
- Developer trust in the tool is eroded — the tool is seen as fighting the user
  rather than helping

---

## Environment

- `golangci-lint-auto-configure`: version unknown (invoked via BuildFlow
  pre-commit hook, step name `golangci-lint [root]:repair`)
- `golangci-lint`: v2 (Go 1.26)
- BuildFlow: latest (auto-commit daemon)
- Project: `templ-components`, 15 Go packages, ~100 Go files
- Config: 200+ line `.golangci.yml` with 60+ enabled linters

---

## Reproduction commands (minimal)

```bash
# In any Go project with golangci-lint v2:
# 1. Remove a linter from linters.enable in .golangci.yml
#    (do NOT add it to linters.disable)
# 2. Commit the change
# 3. Run the repair step (as BuildFlow does)
golangci-lint-auto-configure repair
# 4. Observe: the linter is re-added to linters.enable
git diff .golangci.yml
```

---

_Same pattern as the 2026-07-20 feedback, but triggered by a different code
path: that one was `repair` rebuilding `disable` from scratch (dropping user
disables). This one is `repair` re-adding to `enable` when a linter is absent
from both lists. The 2026-07-25 fix for the former does not address the latter._

---

## Resolution (2026-07-30)

**FIXED — options (b) and (d) shipped.** Two complementary mechanisms now prevent
the regression loop where `configure` re-adds linters the user deliberately
removed from `enable`.

### (b) Automatic cycle detection via audit ledger

`enableRecommendedLinters` (`pkg/linter/fixer.go`) now queries the audit ledger
before adding a linter to `enable`. If the ledger has an `ActionAddedToEnable`
entry for the linter (meaning the tool previously auto-enabled it), the tool
skips re-adding and logs a warning:

```
⚠️  Skipping godoclint: was auto-enabled in a previous run and subsequently removed.
    To make this permanent, add it to linters.disable or .golangci-lint-auto-configure.yml under never-enable.
```

This breaks the loop automatically on the second cycle (first cycle: tool adds +
records; user removes; second cycle: tool detects + skips). The suppression is
recorded as `ActionSuppressedReEnable` in the ledger, visible via the `audit`
subcommand.

Implementation: `audit.PreviouslyAutoEnabled(path, repoHash)` reads the ledger
and returns the set of auto-enabled linters. `Ledger.PreviouslyAutoEnabled()`
and `NoopRecorder.PreviouslyAutoEnabled()` satisfy the `ledgerReader` interface
(duck-typed in the linter package). The `Fixer.reader` field is set in
`SetLedger` via type assertion.

Best-effort: requires the audit ledger (disabled with `--no-audit`), is
per-machine, and entries are purged after 90 days. When unavailable, behavior is
unchanged (linter is added as before).

### (d) Durable `never-enable` sidecar section

The sidecar `.golangci-lint-auto-configure.yml` now supports a `never-enable:`
section (kebab-case YAML key) alongside the existing `disabled:` section.
Linters listed there are never added to `enable` by any code path.

```yaml
never-enable:
  godoclint:
    reason: "demands per-package godoc; this repo documents per-file"
    category: convention
  ireturn:
    reason: "every component returns templ.Component by design"
    category: convention
  testableexamples:
    reason: "Example* funcs render verbose HTML output that isn't asserted"
    category: convention
```

This is the durable, committed-to-git signal that works across machines and CI.
When no sidecar exists, only the automatic cycle detection applies.

### Not implemented

- **(a)** Not viable — `configure` and `repair` are the same code path (no
  separate `repair` subcommand exists).
- **(c)** Superseded by (b)'s warning message, which proactively suggests
  `disable` or the sidecar.
- **(e)** Not implemented — `godoclint` and `testableexamples` may add value to
  some projects. The general mechanisms (b)+(d) cover them without a global
  policy change. Users can add them to `never-enable` per-project.

### Tests

- `TestEnableRecommendedLinters_NeverEnableSidecar` — sidecar never-enable skips linter
- `TestEnableRecommendedLinters_CycleDetection` — ledger-based cycle detection skips + records
- `TestEnableRecommendedLinters_CycleDetectionDryRun` — dry-run warns but doesn't record
- `TestEnableRecommendedLinters_NoReaderNoCycleDetection` — no reader = normal behavior
- `TestEnableRecommendedLinters_DisabledNotAffectedByCycle` — disabled linters skip before cycle check
- Policy tests for `never-enable` parsing, `IsNeverEnable`, `NeverEnableJustification`
- Audit tests for `PreviouslyAutoEnabled` (matching repoHash, cross-repo filtering, edge cases)
