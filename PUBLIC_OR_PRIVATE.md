# Public or Private? — Decision Analysis

**Project:** golangci-lint-auto-configure
**Date:** 2026-05-04
**Current State:** Private (0 stars, 0 forks, 0 issues)
**Recommendation:** Make public, with conditions

---

## Project Summary

A Go CLI tool that automatically configures, optimizes, and maintains golangci-lint configurations. It analyzes existing configs, detects missing linters with a 4-tier priority system (Critical/High/Medium/Optional), auto-fixes issues, replaces deprecated linters, migrates v1→v2 configs, and generates SARIF/JSON/HTML reports. Ships with an SDK-style `pkg/client` for library usage. 7,486 lines of production Go code, 4,373 lines of tests (60.4% coverage), 461 commits, MIT licensed, single contributor.

---

## PRO — Arguments for Making Public

### 1. Fills a genuine gap in the Go ecosystem

No equivalent tool exists. Sourcegraph search across all public repos returns zero results for "golangci-lint auto configure linter priority recommend." The closest thing is copying example configs from golangci-lint's own repo — which is manual, unversioned, and unmaintained. Every Go project that uses golangci-lint has the same problem: which linters to enable, with what settings, and how to keep up with deprecations. This tool solves that systematically.

### 2. Linter priority knowledge is genuinely useful

The curated priority data in `pkg/constants/linter_priorities.go` (119 linters) with human-readable reasons (`linter_reasons.go`) is a non-trivial, expert-level curation that no other tool provides. Making this accessible benefits the entire Go community. Even if someone doesn't use the CLI, the knowledge that `gosec`, `errcheck`, `staticcheck`, `govet`, `errchkjson`, `musttag`, `sloglint`, `nilerr`, `noctx`, and `loggercheck` are "always enable" linters — with rationale — is valuable.

### 3. Production-ready quality

- **MIT licensed** already — the legal framework for open-source is in place
- **CI/CD complete** — GitHub Actions with Go 1.26, golangci-lint, Nix flake checks, Codecov
- **BDD-tested** — Ginkgo/Gomega test suites, 60.4% composite coverage
- **Well-documented** — README with usage examples, AGENTS.md with full architecture, ADRs, examples for 5 project types, API usage example
- **Nix flake** — reproducible builds out of the box
- **Pre-commit hooks** — `.pre-commit-hooks.yaml` ready for community consumption
- **SARIF support** — CI/CD integrable with GitHub Code Scanning from day one

### 4. Low risk, high potential upside

- **No secrets** in the codebase (confirmed — `runtimesecret` references are Go experiment feature flags, not credentials)
- **No proprietary algorithms** — the value is in curation and automation, not secrecy
- **Single contributor** — no risk of accidentally exposing someone else's private work
- **No customer data** — it's a CLI tool that operates on local configs
- **Community contributions** could improve linter coverage, add presets, and maintain deprecation mappings

### 5. go-finding dependency is no longer a blocker

The local `replace` directive has been removed from `go.mod`. The Nix build injects it via `postPatch` from a private flake input, but the published module path (`github.com/larsartmann/go-finding v0.3.0`) works for public consumers as long as go-finding is also made public (or the dependency is published to a public registry). This is a precondition — see below.

### 6. Marketing and portfolio value

For the author: a well-crafted open-source tool in the Go ecosystem with real utility demonstrates engineering quality, domain expertise in developer tooling, and the ability to ship production-grade software. This is significantly more valuable as a public portfolio piece than as a private repo.

---

## CONTRA — Arguments Against Making Public

### 1. go-finding is private

**Blocking issue.** `go-finding` is a dependency fetched via `git+ssh://` from a private GitHub repo. For `go install github.com/larsartmann/golangci-lint-auto-configure@latest` to work for public users, `go-finding` must also be public (or at minimum, the module must be resolvable). This is the single hard blocker.

### 2. Documentation bloat

- **100 status reports** in `docs/status/` — these are AI-generated session summaries with timestamps, not useful for public consumers
- **19 planning documents** in `docs/planning/` — internal working documents
- These would create noise and make the repo look messy/unfinished to visitors

### 3. Single contributor, high maintenance surface

119 linters with priorities and reasons need ongoing maintenance as golangci-lint adds/removes/renames linters. Without community contributions, the knowledge base will slowly drift out of date. Public visibility creates implicit pressure to maintain it.

### 4. No issue/PR process defined

- No `CONTRIBUTING.md` (README has a basic Contributing section but no formal process)
- No issue templates, PR templates, or code of conduct
- No release process documented

### 5. Test coverage is 60.4%

Not terrible, but below what you'd want for a public, widely-used tool. Key packages like `internal/cli` and `pkg/migration` may have lower coverage. Public scrutiny may highlight gaps.

### 6. No versioned release yet

No GitHub releases, no tags, no `goreleaser` config. The `version` variable is injected via ldflags but defaults to "dev". A public tool needs proper versioning.

---

## Conditions for Going Public

### Must (blocking)

| # | Condition | Why |
|---|-----------|-----|
| 1 | **Make go-finding public** (or remove dependency) | `go install` will fail without it |
| 2 | **Delete or archive docs/status/ and docs/planning/** | Noise, not useful publicly |
| 3 | **Tag v0.1.0 release** | Users need a version to pin to |

### Should (highly recommended)

| # | Condition | Why |
|---|-----------|-----|
| 4 | Add `CONTRIBUTING.md` with PR/issue process | Set expectations for contributors |
| 5 | Add GitHub release with binaries (goreleaser or similar) | Lower the barrier to trying the tool |
| 6 | Verify all example configs are valid with latest golangci-lint | First impressions matter |

### Nice to have

| # | Condition | Why |
|---|-----------|-----|
| 7 | Improve test coverage to 70%+ | Confidence for public consumers |
| 8 | Add issue/PR templates | Professionalism |
| 9 | Set up Renovate/Dependabot | Signal active maintenance |
| 10 | Create a short demo GIF/asciinema for README | Increase adoption |

---

## Conditional Recommendation

**Make it public — but execute the "Must" conditions first.**

This project has genuine, unique value in the Go ecosystem. No other tool does what it does. The quality is solid enough for a v0.1 release. The main blocker is the go-finding dependency; once that's resolved (either by making go-finding public or extracting the interface), this should be open-sourced.

**Timeline suggestion:** If go-finding can be made public, the remaining "Must" items (cleanup + tag) are 30 minutes of work. The project could be public within a day.

**If go-finding must stay private:** The `pkg/finding/` package and the SARIF output would need to be either extracted behind an interface with a stub, or the go-finding dependency would need to be published as a public Go module (separate from the source repo being public). This adds 1-2 days of refactoring but is feasible.

---

*Analysis generated from full codebase review: 7,486 LOC production, 4,373 LOC test, 461 commits, 12 test suites, all passing.*
