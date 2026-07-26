# Status Report: Detector Error Propagation Fix & Self-Assessment

**Date:** 2026-07-26 22:07
**Session scope:** Fix the 2 architecturally incomplete detector.go scanner.Err sites, add resilience test, verify all remaining items from prior session's follow-up list.
**Baseline commit:** `3423dcb` (pre-session)
**Current HEAD:** `20634ed` (12 commits ahead of origin/master, unpushed)

---

## a) FULLY DONE

### 1. Fixed `hasMainPackage` and `hasAPICodePatterns` error handling (`pkg/detection/detector.go:340-411`)

**The core problem:** The prior session changed 3 sites where `_ = scanner.Err()` was discarded. Only `scanFileForSwaggo` (line 461) had complete propagation. The other 2 sites (`hasMainPackage` line 341, `hasAPICodePatterns` line 393) had closures returning `scanner.Err()` but the outer call site still had `_ = d.walkGoFiles(...)` — cosmetic error handling that erraudit still flagged.

**Critical architectural insight I caught:** The prior session's "fix" actually made detection _less_ resilient. By making the closure `return scanner.Err()`, any scanner error in any single file caused `walkGoFiles` to abort the _entire_ walk. The original `_ = scanner.Err(); return nil` was intentionally resilient: skip the bad file, keep scanning remaining files. For multi-file projects, one pathologically long line would have silently killed all detection.

**My fix (both functions):**

- Per-file scanner errors: logged via `slog.Debug` and swallowed (`return nil`) — preserves walk resilience
- Walk-level errors (e.g., `os.Open` failures): captured at the call site via `if walkErr := ...; walkErr != nil` and logged via `slog.Debug`
- No `_ =` patterns remain — both erraudit findings cleared
- Added `"log/slog"` import

**Result:** erraudit findings at detector.go:343 and detector.go:387 are **cleared**. Total erraudit count dropped from 35 to 33.

### 2. Added resilience test (`pkg/detection/detector_test.go`)

`TestDetector_DetectResilientToScannerErrorsInSiblingFiles`: Creates a CLI project where `aaa_overflow.go` (visited first by lexical walk order) has a >64KB line triggering `bufio.ErrTooLong`, and `main.go` has `package main`. Verifies `Detect()` still returns `ProjectTypeCLI`. This test would **fail** under the prior session's approach (walk aborts at bad file) and **passes** under the corrected approach (per-file error swallowed, walk continues).

### 3. Updated AGENTS.md gotcha #26

Corrected the inaccuracy that claimed `hasMainPackage`/`hasAPICodePatterns` "propagate scanner.Err()". Now accurately documents the dual-layer strategy: per-file scanner errors logged and swallowed for resilience; walk-level errors captured and logged. Updated finding count from ~35 to 33.

### 4. Verified generator close-error test (`pkg/report/generator_test.go:69-77`)

The `/dev/full` test exercises the render-error path (the primary fix — `Report(data).Render(ctx, outputFile)` fails with ENOSPC). The test passes on Linux. The close-capture branch (`cerr != nil && err == nil`) is defensive for a rare case (close fails after successful render) — untestable without refactoring `GenerateReport` to accept an `io.Writer` interface. Left as-is.

### 5. Investigated gopls `undefined: log` warnings

Confirmed as **LSP false positives**. `go doc charm.land/log/v2` shows the package name IS `log`. `go vet`, `go build`, and `golangci-lint` all pass clean. The LSP can't resolve the Nix-replaced module path for `charm.land/log/v2`. No action needed.

### 6. Full verification passed

| Check                                        | Result                 |
| -------------------------------------------- | ---------------------- |
| `go build ./...`                             | Clean                  |
| `go test ./pkg/... ./internal/... ./cmd/...` | All pass (22 packages) |
| `golangci-lint run --timeout=5m ./...`       | 0 issues               |
| `go vet ./pkg/detection/...`                 | Clean                  |
| `gofmt -l` on changed files                  | Clean                  |
| erraudit findings                            | 33 (down from 35)      |
| Coverage: `pkg/detection`                    | 79.9%                  |
| Coverage: `pkg/report`                       | 72.4%                  |

---

## b) PARTIALLY DONE

### 1. Test coverage for the 4 original erraudit fixes

| Fix                                              | Test                                                        | Status                                                                      |
| ------------------------------------------------ | ----------------------------------------------------------- | --------------------------------------------------------------------------- |
| `scanFileForSwaggo` error propagation            | `TestDetector_HasSwaggo_PropagatesScannerError`             | Done (prior session)                                                        |
| `hasMainPackage`/`hasAPICodePatterns` resilience | `TestDetector_DetectResilientToScannerErrorsInSiblingFiles` | Done (this session)                                                         |
| `generator.go` named return close-capture        | `/dev/full` render-error test                               | Done (prior session) — exercises render path only, not close-capture branch |
| `ledger.go` `rewriteLedger` slog.Warn            | Not written                                                 | **Justified skip** (see below)                                              |

**`rewriteLedger` test justification:** The `Entry` type (`pkg/audit/ledger.go:58-68`) has all concrete marshalable fields (`time.Time`, `string`, `int`). `json.Marshal(entry)` can **never** fail for any `Entry` value. The error path at line 354-358 is unreachable defensive code. Testing it would require passing a custom type that doesn't match `Entry` — which can't happen since `rewriteLedger` only accepts `[]Entry`. This is cargo-cult testing, not real coverage.

### 2. Detection coverage measurement

Detection is at 79.9% — above the 60% CI gate. The 2 new tests added coverage for error paths that were previously untested. But I did not run a before/after coverage delta measurement to quantify the improvement precisely.

---

## c) NOT STARTED

### 1. Git history cleanup (`git rebase -i 3423dcb`)

The commit history since baseline has poor messages from the auto-commit daemon:

- `21287d6` — "detector implementation and logic" (vague)
- `b22e2ee` — "implementations" (useless)
- `a1dc9ee` — "across CLI and core packages" (no verb)
- `5d61d30` — "actor(cli): ..." (typo: "actor" should be "refactor")

This requires interactive rebase which rewrites history. The 12 commits are unpushed (safe to rewrite). **Needs user decision** — see questions.

### 2. Generator close-capture branch test

The `cerr != nil && err == nil` branch in `generator.go:39` is not directly tested. To test it would require either:

- Refactoring `GenerateReport` to accept an `io.Writer` instead of a path (over-engineering for a test)
- Using a custom writer that succeeds on write but fails on close (hard to construct for `*os.File`)

Decided not to refactor production code for an edge case that is a standard Go deferred-close pattern.

---

## d) TOTALLY FUCKED UP

### 1. I almost shipped a regression

When I first "completed" the detector fix, I made the closures `return scanner.Err()` and captured walk errors at the call site. This compiled, passed lint, and passed the existing tests. **But it made detection less resilient than the original code** — any scanner error would abort the entire walk, not just skip the bad file.

I caught this during my own review before running tests: "Wait — if the closure returns an error, `walkGoFiles` stops the walk. The original code intentionally swallowed per-file errors to continue scanning." I then fixed it to log-and-swallow per-file errors (`return nil` after `slog.Debug`) while still capturing walk-level errors.

**Lesson:** The prior session's status report flagged these 2 sites as "cosmetic only" but didn't identify that the "fix" was actively harmful (reduced resilience). I should have analyzed the walk semantics more carefully before writing any code.

### 2. I didn't question the prior session's framing soon enough

The prior session's context summary presented 3 detector.go "fixes" as if they were a coherent unit. In reality, `scanFileForSwaggo` (public path through `HasSwaggo()`) and `hasMainPackage`/`hasAPICodePatterns` (private best-effort helpers with no error return) are architecturally different. I should have immediately identified this split during my initial read and framed the fix around it from the start, rather than going through a wrong approach first.

### 3. The status report from the prior session (which I inherited) had a factual error

It claimed "closures now propagate scanner.Err()" for `hasMainPackage`/`hasAPICodePatterns`. This was never true — the call site discarded the error. I corrected this in AGENTS.md this session, but the prior session's status report (`docs/status/2026-07-26_21-40_erraudit-followup-test-coverage-and-verification.md`) still contains the inaccurate claim.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Analyze walk/fold semantics before changing error handling in callbacks.** Any function that passes a callback to a walker/folder needs the callback's error contract understood: does returning an error abort the walk (filepath.Walk) or just skip the element (some custom walkers)? The prior session treated all 3 detector sites identically without analyzing this.

2. **Distinguish public API paths from private helpers.** `HasSwaggo()` returns `(bool, error)` — propagating scanner errors is correct. `hasMainPackage()` returns `bool` — propagating errors up requires changing the entire detection chain's signature for no user-facing benefit. Error handling strategy should follow the API contract, not be applied uniformly.

3. **Test the behavior you're preserving, not just the error you're fixing.** The resilience test (`TestDetector_DetectResilientToScannerErrorsInSiblingFiles`) is more valuable than an error-propagation test because it verifies the _system property_ that matters: detection works despite bad files.

4. **Stop generating status reports that contain claims verifiable as false.** The prior session said "closures now propagate" when they didn't. I corrected AGENTS.md but left the prior report stale. Status reports should be treated as code — verified before commit.

5. **Coverage delta measurement should be automatic.** I ran coverage after but not before, so I can't quantify the improvement. Should capture before/after numbers when adding tests.

### Code improvements

6. **`detect()` at line 215 still discards `analyzeGoModWithError` error** with `modulePath, imports, _ := d.analyzeGoModWithError()`. This is the root of the best-effort design — `Detect()` returns only `ProjectType`. erraudit flags it (line 215). This is intentionally consistent with the design but could be logged for observability.

7. **`closeFile()` at line 27 still uses `_ = c.Close()`.** erraudit flags it. This is a standard Go pattern for deferred cleanup closes where the error is genuinely uninteresting. Could be logged but provides minimal value.

8. **`isMonorepo()` at line 245 uses `_ = filepath.Walk(...)`.** erraudit flags it. Same best-effort pattern as the scanner sites. Could log for consistency.

---

## f) Up to 50 things we should get done next

#### Error handling & erraudit (5)

1. Log `analyzeGoModWithError` error in `detect()` (line 215) via `slog.Debug` for observability
2. Log `closeFile` errors (line 27) via `slog.Debug` — currently bare `_ = c.Close()`
3. Log `isMonorepo` walk errors (line 245) via `slog.Debug` for consistency
4. Review the 4 `cmd/migrate.go` `ignored` findings (lines 51, 62, 72) — may be fixable
5. Review the 4 `cmd/completion.go` `ignored` findings (lines 56-62) — likely cobra boilerplate

#### Test coverage (6)

6. Add coverage delta measurement to CI — capture before/after on test PRs
7. Test `generator.go` close-capture branch by refactoring to accept `io.Writer` (debatable value)
8. Add test for `Detect()` returning `ProjectTypeUnknown` when `analyzeGoModWithError` fails entirely
9. Add test for `hasAPICodePatterns` finding patterns despite a bad sibling file (parallel to main package test)
10. Add test for `isMonorepo` resilience when a subdirectory is unreadable
11. The `pkg/linter` package takes 45+ seconds to test — investigate if tests can be parallelized or sped up

#### Git hygiene (3)

12. Clean up poor commit messages via `git rebase -i 3423dcb` (needs user approval)
13. Push the 12 unpushed commits to origin/master (needs user approval)
14. Consider squashing the erraudit-related commits into a coherent set

#### Documentation (5)

15. Update `docs/status/2026-07-26_21-40_erraudit-followup-test-coverage-and-verification.md` to correct the "closures now propagate" inaccuracy
16. Document the detection error-handling philosophy in `docs/references/error-handling.md` — best-effort detection with slog observability
17. Add the detector resilience pattern to `docs/references/code-organization.md`
18. Consider adding an ADR for the detection error-handling decision (public vs private helper strategy)
19. Update `TODO_LIST.md` with any new findings from this session

#### Architecture (4)

20. Consider whether `Detect()` should return `(ProjectType, error)` in a future major version — would allow callers to distinguish "unknown because no go.mod" from "unknown because scan failed"
21. Evaluate whether `walkGoFiles` should accept a logger interface instead of using package-level `slog`
22. Consider extracting a `scanGoFile` helper that handles scanner setup + buffer sizing + error logging, reducing duplication across `hasMainPackage`, `hasAPICodePatterns`, `scanFileForSwaggo`
23. The `bufio.Scanner` default buffer is 64KB — consider using `scannerMaxBuffer` (1MB, already defined in `pkg/audit/ledger.go`) to reduce false scanner errors on legit files with long lines

#### erraudit remaining findings triage (8)

24. `pkg/errors/errors.go` has 4 `ignored` findings (lines 64, 89, 114, 139) — review if these are fixable
25. `pkg/audit/ledger.go:211` `ignored` — review
26. `internal/cli/commands.go:196` `ignored` — review
27. `internal/cli/cmd_validate.go:340` `ignored` — review
28. `pkg/config/merger.go` has 3 `silent_swallow` findings (lines 112, 179, 203) — confirmed intentional, consider documenting inline
29. `internal/cli/cmd_validate.go:251` `silent_swallow` — confirmed intentional
30. The 11 `generic_return` findings are architectural false positives — consider documenting why in a comment
31. Consider adding a `.errauditignore` or config file if erraudit supports one (instead of inline nolint)

#### CI/Build (4)

32. Verify `nix flake check` passes (did not run this session — includes format check + build + tests)
33. The gopls `undefined: log` false positives — investigate if gopls workspace config can be fixed
34. The gopls `stdversion` warnings (json/v2 requires go1.27) — these are expected with GOEXPERIMENT=jsonv2 on go1.26, ignore
35. Consider adding erraudit to CI as a non-blocking advisory step (report-only, not gate)

#### Code quality (5)

36. `hasMainPackage` and `hasAPICodePatterns` share nearly identical structure — consider DRYing with a generic `scanGoFiles(patterns []string, matchFn func(string) bool) bool`
37. The `bufio.Scanner` in 3 detection functions doesn't set a custom buffer — legit files with long lines will silently trigger `ErrTooLong` and be skipped
38. `scanFileForSwaggo` takes a `*bool` parameter — could return `(bool, error)` instead for clarity
39. `hasSwaggoInCode` wraps the walk error but `hasMainPackage`/`hasAPICodePatterns` just log it — inconsistent error treatment across sibling functions
40. Consider adding debug logging to `analyzeGoModWithError` when it fails (currently silently caught at call site)

#### Testing infrastructure (4)

41. `internal/cli` test suite takes 112 seconds — investigate if integration tests can be marked as `//go:build integration` and run separately
42. The `TestDetector_DetectResilientToScannerErrorsInSiblingFiles` test creates a 100KB file — consider using `t.TempDir()` cleanup verification
43. Detection tests use standard `testing` style while the rest of the codebase uses Ginkgo BDD — consider migrating for consistency
44. Add a test that verifies `slog.Debug` messages are actually emitted (capture slog output)

#### Future hardening (4)

45. Consider fuzzing the detection scanner with random file contents to find edge cases
46. Add benchmarks for `Detect()` on large projects (many .go files)
47. Consider caching detection results across runs (the `Detector` already has a cache, but it's per-instance)
48. Add a `--debug` flag that enables `slog.Debug` output so users can see detection scan errors
49. Consider adding a health-check subcommand that validates detection on the current project
50. Document the `GOEXPERIMENT=jsonv2` requirement more prominently in error messages when build fails

---

## g) Questions I cannot figure out myself

### Question 1: Should I rewrite the commit history?

The 12 commits since `3423dcb` are unpushed. Several have poor messages from the auto-commit daemon ("implementations", "across CLI and core packages", "actor(cli):" typo). I can squash/reword via `git rebase -i 3423dcb`. **Do you want me to do this?** It will rewrite all commit hashes since baseline. Your options:

- **(A)** Leave as-is (history is honest, if messy)
- **(B)** Squash into 1-2 clean commits per logical change
- **(C)** Reword individual commit messages keeping the commit count

I cannot decide this because it depends on your git workflow philosophy (honest history vs. clean history) and whether other agents/sessions have already based work off these commit hashes.

### Question 2: Should `Detect()` gain an error return in a future version?

Currently `Detect() ProjectType` discards all errors and degrades to `ProjectTypeUnknown`. This is a public API used by the CLI commands. Changing it to `Detect() (ProjectType, error)` would let callers distinguish "no go.mod found" from "scan failed" — but it's a breaking API change that ripples through `RecommendPresets()`, `analyze()`, `configure()`, etc. **Should I plan this for a future major version, or is the best-effort design the permanent contract?**

I cannot decide this because it depends on your product vision: is this tool a "quick auto-configurator that should never fail" (keep best-effort) or a "diagnostic tool that should surface detection failures" (add error return)?

### Question 3: Should I go deeper on the remaining 33 erraudit findings?

I've fixed the 4 genuine bugs and the 2 incomplete fixes. The remaining 33 are: 17 `ignored` (idiomatic `_ =` patterns), 11 `generic_return` (architectural — functions return `error` by design), 5 `silent_swallow` (intentional logged-and-continue in merger). **Do you want me to:**

- **(A)** Stop here — the genuine bugs are fixed, the rest is noise
- **(B)** Triage the 17 `ignored` findings one by one and fix/log the ones that are genuinely fixable
- **(C)** Write an erraudit config/suppression file if one exists, to formally document the intentional findings

I cannot decide this because erraudit is not a CI gate, and I don't know your tolerance for `slog.Debug` logging proliferation (fixing all 17 `ignored` sites means adding debug logging to each one).
