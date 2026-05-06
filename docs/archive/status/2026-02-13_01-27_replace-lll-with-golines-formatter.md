# Status Report: Replace lll Linter with golines Formatter

**Date**: 2026-02-13 01:27
**Type**: Enhancement
**Status**: Complete

## Summary

Replaced the `lll` (line length limit) linter with the `golines` formatter to provide automatic line length fixing instead of just reporting violations.

## Problem

The `lll` linter only **reports** long lines but has no autofix capability. This meant developers had to manually break long lines, which is tedious and error-prone.

## Solution

Use `golines` formatter instead, which **automatically fixes** long lines by breaking them appropriately.

### golangci-lint v2 Formatters

golangci-lint v2 introduced a separate `formatters:` section in the configuration. Formatters are distinct from linters:

- **Linters**: Check code and report issues (no modification)
- **Formatters**: Actually modify code to fix issues

Available formatters: `gci`, `gofmt`, `gofumpt`, `goimports`, `golines`, `swaggo`

## Changes Made

### 1. `.golangci.yml`

```yaml
# Removed from linters.enable:
- lll

# Added new section:
formatters:
  enable:
    - golines
  settings:
    golines:
      max-len: 120
```

### 2. `pkg/constants/linter_data.go`

- Removed `lll` from `LinterPriorities` map
- Removed `lll` from `LinterReasons` map
- Removed `lll` from `PresetLinters["strict"]` slice
- Added `golines` to `FormatterPriorities` (High priority)
- Added `golines` to `FormatterReasons`
- Added new `RedundantLinters` map documenting why `lll` is superseded

### 3. `internal/cli/commands_test.go`

Fixed pre-existing test bug where the validate command test's sample config was missing the required `run.timeout` field.

## Usage

After these changes, running:

```bash
golangci-lint fmt ./...
```

Will automatically format code AND fix long lines exceeding 120 characters.

## Verification

- All 20 tests pass
- No `lll` warnings in lint output
- `golangci-lint fmt` works with golines formatter
- Coverage: 44.3% composite

## Files Changed

| File                            | Changes                                            |
| ------------------------------- | -------------------------------------------------- |
| `.golangci.yml`                 | Removed lll, added formatters section with golines |
| `go.mod`                        | Updated dependencies (go mod tidy)                 |
| `go.sum`                        | Updated checksums                                  |
| `internal/cli/commands_test.go` | Fixed test sample config (added run.timeout)       |
| `pkg/constants/linter_data.go`  | Updated linter/formatter metadata                  |

## Technical Details

### Why golines over lll?

| Feature              | lll        | golines       |
| -------------------- | ---------- | ------------- |
| Detects long lines   | Yes        | Yes           |
| Autofixes long lines | No         | Yes           |
| Type                 | Linter     | Formatter     |
| Action required      | Manual fix | Automatic fix |

### RedundantLinters Map

Added a new map to track linters that are superseded by formatters:

```go
var RedundantLinters = map[types.LinterName]string{
    "lll": "redundant when golines formatter is enabled (golines fixes long lines, lll only reports them)",
}
```

This can be used by the analyzer to recommend formatter alternatives to lint-only tools.

## Next Steps

Consider updating the analyzer to:

1. Detect when `lll` is enabled but `golines` is not
2. Suggest the formatter alternative
3. Potentially auto-replace `lll` with `golines` during configuration
