# Memo: never-enable × deprecated-replacement bypass

**Decision needed:** should the deprecated-linter replacement path respect the
sidecar `never-enable` section? **Recommendation: yes — add the check; the
documented contract currently has a real, narrow hole.** (Guard + tests ≈ 30
min. Decline path documented below.)

## The trace (verified in code, 2026-10-07)

The sidecar contract (`pkg/policy/policy.go`, AGENTS "Where to Find Detail" item 2):
linters under `never-enable` "are never added to `enable`, period …
takes priority over all other mechanisms", enforced in both code paths:

- recommendation: `enableRecommendedLinters` (`pkg/linter/fixer.go:391`)
- anti-gaming enforcement: `tryReEnableLinter` (`pkg/linter/fixer_enforce.go:76`)

The **third** path that adds to `enable` is unchecked:
`deprecatedLinterHandler.applyReplacement` (`pkg/linter/fixer_deprecated.go:82`):
`linterSet.Add(replacement.Replacement)` — no `IsNeverEnable` check, no
`NeverAutoEnableLinters` check.

## Is the bypass real?

Yes, and it has a live instance today: `exhaustruct` is deprecated
(golangci-lint v2.13) with replacement `exhaustruct_v5`, which is in
`NeverAutoEnableLinters`. A user who puts `exhaustruct_v5` under
`never-enable` (e.g. "too much friction, don't ever add it") but still carries
deprecated `exhaustruct` in `enable` (86/157 fleet configs carry that exact
shape) gets `exhaustruct_v5` silently force-added by the migration — the one
outcome the sidecar promises cannot happen.

## The semantic split (why one check, not two)

- **Sidecar `never-enable` on the replacement → respect it.** It is the
  user's explicit, committed-to-git instruction. Fix: delete the deprecated
  predecessor (deprecation cleanup is still correct) but skip adding the
  replacement — the existing `logRemove` branch already implements exactly
  that shape.
- **`NeverAutoEnableLinters` on the replacement → keep bypassing it.**
  That list governs what the tool _auto-recommends_; a replacement of a
  linter the user manually enabled is a migration of their explicit choice,
  not an auto-enable. Documenting this distinction is part of the fix.

## Implementation sketch

In `applyReplacement`, before `linterSet.Add`:

```go
if f.pol != nil && f.pol.IsNeverEnable(replacement.Replacement) {
    h.logger.Infof("%sreplace deprecated linter: %s removed; replacement %s "+
        "is under never-enable — not added", dryRunActionPrefix(dryRun), linter, replacement.Replacement)
    return 1
}
```

(+ threading `pol`/handler access as needed, + BDD specs: bypass blocked,
predecessor still removed, settings NOT migrated, `--pragmatic` composition
unaffected, dry-run wording.)

## Decline path

If declined: no code change, but the AGENTS "never-enable takes priority over
all other mechanisms" sentence MUST be amended to "…except deprecation
migrations" — the current text is false under decline, and false docs are the
failure mode this project exists to prevent.

## Answer line

- [ ] Implement the guard (recommended)
- [ ] Decline + amend the docs instead
