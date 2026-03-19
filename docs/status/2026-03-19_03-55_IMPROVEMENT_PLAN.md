# Improvement Plan - Post Phase 2 Refactoring

**Date:** 2026-03-19 03:55  
**Commit:** f06afab  
**Status:** Build passing, tests passing, Phase 2 complete

---

## Executive Summary

Phase 2 refactoring is complete. Build passes, all tests pass. Now focusing on:

1. **Critical:** Proper context propagation from cobra commands
2. **High Impact:** User-friendly validation error messages
3. **Medium Impact:** Documentation and code cleanup

---

## Analysis: What I Forgot / Could Do Better

### 1. Context Propagation Gap

**Current State:** Using `context.Background()` in CLI helpers

```go
// BAD - loses cancellation signals
defaultConfig := configLoader.CreateDefaultConfig(context.Background())
```

**Problem:** Users cannot cancel long-running operations (e.g., fetching linters) with Ctrl+C

**Solution:** Pass `cmd.Context()` from cobra through the call chain

### 2. Validation Error UX

**Current State:** Raw validator errors

```go
// User sees: "Key: 'Config.Version' Error:Field validation for 'Version' failed on the 'required' tag"
```

**Problem:** Not user-friendly for CLI output

**Solution:** Custom error message mapping

### 3. Stale LSP Diagnostics

**Current State:** gopls showing cached errors from before fixes

**Impact:** False positives in editor, confusing development experience

**Solution:** Restart LSP, clear module cache

---

## Multi-Step Execution Plan

### Phase A: Critical Fixes (Do First)

#### Step 1: Context Propagation from Cobra

**Impact:** HIGH - Enables cancellation, proper timeouts  
**Work:** MEDIUM  
**Files:** `internal/cli/cmd_*.go`

```go
// Change from:
RunE: func(cmd *cobra.Command, args []string) error {
    return runConfigure(...)
},

// Helper uses context.Background() - BAD

// To:
RunE: func(cmd *cobra.Command, args []string) error {
    return runConfigure(cmd.Context(), ...)
},

// Helper receives context - GOOD
func runConfigure(ctx context.Context, ...) error {
    configLoader.CreateDefaultConfig(ctx)
}
```

**Verification:**

- Build passes
- Tests pass
- Can cancel `golangci-linter-auto-configure configure` with Ctrl+C

---

#### Step 2: User-Friendly Validation Errors

**Impact:** HIGH - Better UX  
**Work:** LOW  
**Files:** `pkg/types/validation.go`, CLI commands

```go
// Add to validation.go:
func FormatValidationErrors(err error) string {
    // Convert validator.ValidationErrors to human-readable messages
    // "version is required" instead of "Key: 'Config.Version' Error:..."
}
```

**Verification:**

- Invalid config shows clear error message
- Users understand what to fix

---

### Phase B: Code Quality (Do Next)

#### Step 3: Remove Unused Builder.config Field

**Impact:** LOW - Cleaner code  
**Work:** LOW  
**Files:** `pkg/workflow/workflow.go`

**Verification:**

- Build passes
- No functionality changes

---

#### Step 4: Fix Remaining wrapcheck Errors

**Impact:** MEDIUM - Better error context  
**Work:** MEDIUM  
**Files:** `internal/cli/cmd_*.go`

**Verification:**

- golangci-lint shows fewer wrapcheck warnings

---

### Phase C: Documentation (Do Last)

#### Step 5: Document Result Type Patterns

**Impact:** MEDIUM - Better onboarding  
**Work:** MEDIUM  
**Files:** `AGENTS.md`

**Verification:**

- New contributors understand ROP patterns

---

## Top 25 Things To Get Done (Prioritized)

### Critical (This Session)

1. ✅ ~~Phase 2 refactoring~~ (DONE)
2. ⏳ Pass context from cobra commands to helpers
3. ⏳ Format validation errors for CLI output
4. ⏳ Restart LSP to clear stale diagnostics

### High Priority (Today)

5. Remove unused `Builder.config` field
6. Fix wrapcheck in `cmd_configure.go:159`
7. Fix wrapcheck in `cmd_migrate.go:83`
8. Add examples to AGENTS.md for Result types
9. Document context propagation pattern

### Medium Priority (This Week)

10. Fix varnamelen warnings (9 instances)
11. Fix noinlineerr warnings (15 instances)
12. Add nolint:exhaustruct to cobra.Command
13. Fix tagalign issues in types.go
14. Fix staticcheck SA4010 in fixer.go

### Lower Priority (Future)

15-25. Various lint cleanup items (goconst, godox, funlen, etc.)

---

## Architecture Improvements to Consider

### 1. Context-Aware Result Types

**Question:** How do we combine `mo.Result` with context cancellation?

**Options:**

| Option                    | Pros          | Cons                 |
| ------------------------- | ------------- | -------------------- |
| Check ctx in each FlatMap | Simple        | Verbose, error-prone |
| Wrap context in payload   | Explicit      | Pollutes types       |
| Create ContextResult type | Clean         | Duplicates mo.Result |
| Use uniflow library       | Purpose-built | New dependency       |

**Recommendation:** Research `larsartmann/uniflow` library

### 2. Validation Error Types

**Current:** Returns `[]error` from `ValidateConfig()`

**Better:** Return structured validation result

```go
type ValidationResult struct {
    Valid  bool
    Errors []FieldError
}

type FieldError struct {
    Field   string
    Value   interface{}
    Message string
    Tag     string
}
```

### 3. Operation Timeouts

**Current:** No timeout configuration

**Better:** Add `--timeout` flag per operation type

```bash
golangci-linter-auto-configure configure --fetch-timeout=30s --git-timeout=5s
```

---

## What We Should Improve

### 1. Error Message Quality

**Current:** Technical validator messages  
**Target:** Clear, actionable user guidance

### 2. Context Awareness

**Current:** Operations can't be cancelled  
**Target:** Full cancellation support via Ctrl+C

### 3. Documentation

**Current:** Result types not documented  
**Target:** Clear patterns in AGENTS.md

### 4. Dead Code

**Current:** Unused fields and parameters  
**Target:** Clean, minimal code

---

## Top #1 Question

### How do we properly handle context cancellation with Railway-Oriented Programming?

**Context:**

We use `samber/mo.Result` for error handling chains:

```go
result := loader.LoadConfigResult(path).
    FlatMap(func(cfg *Config) mo.Result[*Config] {
        // What if ctx is cancelled here?
        return processConfig(cfg)
    })
```

**The Problem:**

- `mo.Result` has no context awareness
- Each step should check `ctx.Err()` but that's verbose
- We want cancellation to propagate through the chain

**Potential Solutions:**

1. **Manual checks:** Check `ctx.Err()` in each lambda
2. **Context wrapper:** Include ctx in Result payload
3. **Custom Result type:** `ContextResult[T]` with context methods
4. **Switch to uniflow:** Use `larsartmann/uniflow` instead of `mo`

**What I Need:**

- Guidance on idiomatic context + ROP patterns
- Review of uniflow library capabilities
- Decision on whether to stay with mo or switch

---

## Current Status

| Metric | Status                             |
| ------ | ---------------------------------- |
| Build  | ✅ PASSING                         |
| Tests  | ✅ ALL PASSING (5 packages)        |
| Lint   | ⚠️ 211 warnings (non-critical)     |
| LSP    | ⚠️ Stale diagnostics (cache issue) |

---

## Next Actions

1. **Immediate:** Fix context propagation in CLI commands
2. **Today:** Add validation error formatting
3. **This Week:** Clean up remaining lint issues
4. **Decision:** Evaluate uniflow vs mo for context handling

---

_Assisted-by: Crush <crush@charm.land>_  
_Date: 2026-03-19 03:55_
