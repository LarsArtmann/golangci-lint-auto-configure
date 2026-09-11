# ADR-005: showDiff Parameter Refactor

**Date:** 2026-07-06
**Status:** Accepted

## Context

The `showDiff` flag was a package-level variable in `internal/cli/commands.go`, read directly by functions in `internal/cli/cmd_configure.go` (`runFixerMode`, `effectiveDryRunForCheckDiff`, `finalizeFixerResult`). This created a hidden dependency: functions read global state without it being passed as a parameter.

This violated the "explicit over implicit" principle and made the functions untestable in isolation — unit tests had to set package-level state before calling them.

## Decision

Thread `showDiff` as an explicit `bool` parameter through the entire call chain:

```text
newConfigureCommand → runDetectOrConfigure → runConfigure → runPresetOrFixer → runFixerMode → finalizeFixerResult
```

The package-level variable remains only as the cobra flag binding target. Its value is read once at the entry point and passed down.

## Consequences

- **Positive:** Functions are now pure with respect to `showDiff` — they can be unit-tested without setting global state. The dependency is visible in function signatures.
- **Positive:** Enables table-driven tests for `effectiveDryRunForCheckDiff` with different `showDiff` values.
- **Negative:** Function signatures are longer (one additional `bool` parameter on 6 functions).
- **Mitigated:** The entry point reads the flag value once, keeping the rest of the chain clean.
