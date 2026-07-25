# FEEDBACK: `repair` re-enables explicitly disabled linters, overriding user intent

> **Severity:** High — silently undoes committed configuration changes on every commit
> **Reported:** 2026-07-20
> **Reporter:** Lars Artmann (via Crush session)
> **Tool:** `golangci-lint-auto-configure` (invoked via BuildFlow pre-commit hook as `golangci-lint-auto-configure repair` / `golangci-lint [root]:repair`)
> **Project affected:** `projects-management-automation` (159 Go repos, golangci-lint v2.12.2, Go 1.26.4)

---

## Summary

When a user **intentionally disables** a linter by removing it from
`linters.enable` in `.golangci.yml`, the `repair` command re-adds it on the
next run. The repair step treats "missing from `enable`" as "user forgot to
add it" rather than "user chose to exclude it." This silently overrides
committed configuration and produces hundreds of false-positive findings on
the very next commit.

## Reproduction

### Step 1 — User commits a config with 4 linters removed

The `projects-management-automation` project disabled 4 stylistic linters
(`mnd`, `varnamelen`, `ireturn`, `tagalign`) that produce false positives on
idiomatic Go in this codebase. The change was committed:

```bash
git show 210e035f -- .golangci.yml | head -40
```

After the commit, the committed `.golangci.yml` has **zero references** to
these 4 linters:

```bash
$ git show HEAD:.golangci.yml | grep -c 'ireturn\|mnd\|tagalign\|varnamelen'
0
```

### Step 2 — Next `git commit` triggers the pre-commit hook

BuildFlow runs `golangci-lint-auto-configure repair` as part of its
`golangci-lint [root]:repair` step. This re-adds all 4 linters to the
`enable` list:

```bash
$ git diff HEAD -- .golangci.yml | head -40
diff --git a/.golangci.yml b/.golangci.yml
@@ -74,11 +74,13 @@ linters:
     - interfacebloat
     - intrange
     - iotamixing
+    - ireturn
     - loggercheck
     - maintidx
     - makezero
     - mirror
     - misspell
+    - mnd
@@ -105,6 +107,7 @@ linters:
     - spancheck
     - sqlclosecheck
     - staticcheck
+    - tagalign
     - tagliatelle
@@ -117,6 +120,7 @@ linters:
     - unused
     - usestdlibvars
     - usetesting
+    - varnamelen
     - wastedassign
```

### Step 3 — Working tree now has 6 references to disabled linters

```bash
$ grep -c 'ireturn\|mnd\|tagalign\|varnamelen' .golangci.yml
6
```

The repair step also re-adds the now-orphaned `ireturn:` and `varnamelen:`
settings blocks (previously removed because the linters were disabled).

### Step 4 — Lint gate goes from 0 → 485 issues

With the linters re-enabled, the next `golangci-lint run` reports:

```
485 issues:
* golines: 12
* ireturn: 79
* mnd: 195
* tagalign: 86
* varnamelen: 125
```

All 485 are false positives in this codebase (magic numbers in file
permissions `0o644` and display widths; idiomatic short names like `tc`,
`r`, `wg`, `mu`; factory functions returning interfaces — the core DI
pattern; tag alignment that conflicts with `golines`'s 100-char limit).

---

## Why this is wrong

### 1. The tool cannot distinguish "forgot" from "chose not to"

Removing a linter from `enable` is the **documented way** to disable a
linter in golangci-lint v2. There is no signal in the config that says "I
deliberately excluded this." The repair step assumes omission is always an
oversight.

### 2. `linters.disable:` may or may not help — undocumented

golangci-lint v2 supports an explicit `linters.disable:` section. Hypothesis:
adding a linter to `disable` tells `repair` "I know about this and I'm
choosing to disable it," so it won't re-add to `enable`. **This is
untested** — I could not find documentation confirming the behavior. If
`disable` works, the tool should recommend it when a user removes a linter
from `enable`.

### 3. The repair fights the user on every commit

Every commit that touches Go code triggers the hook, which re-adds the
linters. The user must then run `git restore .golangci.yml` after every
commit. This is fragile and error-prone — easy to forget, easy to
accidentally commit the re-added linters.

### 4. The "auto-fix" makes the lint gate worse, not better

The purpose of `repair` is presumably to keep the config healthy. In this
case, it takes a green gate (0 issues) and makes it red (485 issues) on
every commit. This is the opposite of helpful.

---

## What I expected

One of:

**(a)** `repair` respects user intent: if a linter is not in `enable` and
not in `disable`, it should not be added to `enable` without confirmation.
Prompt the user or emit a warning instead of silently modifying.

**(b)** `repair` respects explicit `disable`: if I add a linter to
`linters.disable:`, `repair` should never re-add it to `enable`. This gives
the user a way to signal intent.

**(c)** `repair` provides a config flag to opt out of re-adding linters:
e.g., `repair --no-add-linters` or `repair --preserve-disabled`.

**(d)** BuildFlow provides a way to skip the repair step for specific files:
e.g., a `.buildflowignore` or `buildflow.yaml` entry that excludes
`.golangci.yml` from repair.

---

## Impact on the project

This single behavior is **blocking all subsequent work** on an 18-task
Pareto execution plan. Every task commit triggers the hook, which re-adds
the disabled linters, which makes the lint gate report 485 false positives,
which obscures whether the task introduced real issues. The team is forced
to either:

- Accept the fight (run `git restore .golangci.yml` after every commit)
- Bypass the hook (`git commit --no-verify`, which skips ALL checks)
- Fix all 485 false positives mechanically (2–4 hours of busywork that
  makes the codebase **worse** — e.g., extracting `0o644` to a named
  constant adds ceremony without clarity)

None of these is a good outcome.

---

## Environment

- `golangci-lint-auto-configure`: version unknown (invoked via BuildFlow
  `66ac27f`, step name `golangci-lint [root]:repair`)
- `golangci-lint`: v2.12.2 (built with go1.26.4)
- BuildFlow: `66ac27f`
- Project: `projects-management-automation`, 68 Go packages, 610 Go files
- Config: 667-line `.golangci.yml` with 100+ path-specific exclusion rules

---

## Reproduction commands (minimal)

```bash
# In any Go project with golangci-lint v2:
# 1. Remove a linter from linters.enable in .golangci.yml
# 2. Commit the change
# 3. Run the repair step
golangci-lint-auto-configure repair
# 4. Observe: the linter is re-added to linters.enable
git diff .golangci.yml
```

---

## Suggested fix priority

**High.** This is a daily friction point for any team that intentionally
disables linters. The current behavior actively harms the lint gate's
usefulness and forces workarounds that bypass safety checks.

If `linters.disable:` already solves this, **document it prominently** in
the README — the current docs don't mention how to permanently exclude a
linter in a way that survives `repair`.

---

_This feedback was generated during a Pareto execution plan session where
the lint gate fight consumed approximately 40 minutes of debugging time
across tasks T5, T6, and T7._

---

## Resolution (2026-07-25)

**FIXED — option (b) shipped.** `repair`/`configure` now respects `linters.disable` and never re-adds a disabled linter to `enable`.

- **Root cause fixed:** `updateConfigFromSets` (`pkg/linter/fixer_config.go`) rebuilt `linters.disable` from scratch on every run, dropping every user-disabled linter. It now preserves, dedups, and sorts the existing `disable` list. Commits `193b6b1` (preserve + audit) and `8d10df5` (policy enforcement + ledger).
- **Orphaned settings pruned:** when a linter is disabled, its `settings.<linter>` block is removed too (so re-enabling later starts clean) — `pruneDisabledLinterSettings` in `fixer_config.go`.
- **Anti-gaming (Pillar C):** a `.golangci-lint-auto-configure.yml` sidecar can justify intentional disables; without one, all disables are still respected (backward compatible). Tool-level disabled linters (`constants.DisabledLinters`) are always exempt.
- **Audit ledger:** every config mutation is recorded to `~/.cache/golangci-lint-auto-configure/audit.jsonl` (query via the `audit` subcommand; disable with `--no-audit`).
- **Documented:** AGENTS.md gotcha #15 ("Disable-reason enforcement is opt-in via sidecar").

**Not implemented:** option (c) `--no-add-linters` (superseded by the disable-respect + sidecar design, which makes it unnecessary). Options (a)/(d) were overtaken by the (b) implementation.
