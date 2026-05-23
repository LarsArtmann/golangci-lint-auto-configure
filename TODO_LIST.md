# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-05-23

---

## Critical Priority

- [ ] Increase CLI integration test coverage (currently 9.0%)

## High Priority

- [ ] Trim AGENTS.md from 912 to ≤377 lines (extract detailed docs to referenced files)
- [ ] Increase gogenfilter scanner coverage (currently 59.8%)
- [ ] Increase migration coverage (currently 66.8%)
- [ ] Remove `report_templ.go` from git tracking (generated file)

## Medium Priority

- [ ] Add `--check` mode for CI (exit 1 if config needs changes)
- [ ] Consider adding `output.formats: {}` to default config
- [ ] Add preset to apply reference config (project-dependency-graph pattern)

## Low Priority

- [ ] Add `swaggo` formatter detection improvements
- [ ] Add benchmarking for analyzer and fixer
- [ ] Document exclusion pattern syntax (RE2 regex) in README
- [ ] Add `--diff` flag to show config changes before applying
- [ ] Decide whether vendor/ should be in formatter exclusions
- [ ] Add `ginkgolinter` default settings if any exist
- [ ] Add `testifylint` default settings

## Completed

- [x] Fix error types: use distinct structs instead of type aliases for errors.As discrimination
- [x] Remove 3 dead backward-compat validation functions from pkg/types/validation.go
- [x] Simplify updateExclusionRules from O(n\*m) to O(n+m) using set-based lookup
- [x] Remove redundant newFixCounts() constructor (Go zero-init is sufficient)
- [x] Add report package tests (HTML and JSON generators)
- [x] Migrate charmbracelet/fang v1 → v2 (charm.land/fang/v2)
- [x] Enrich CreateDefaultConfig() with default settings, exclusion paths, rules
- [x] Add comprehensive default linter settings (revive, varnamelen, gomoddirectives, cyclop)
- [x] Add default formatter settings (golines max-len: 120)
- [x] Add default exclusion rules for test files
- [x] Add default exclusion paths for \_templ.go and vendor/
- [x] Fix all lint violations (funlen, varnamelen, exhaustruct, gci, golines, gocritic)
- [x] Create FEATURES.md feature audit
