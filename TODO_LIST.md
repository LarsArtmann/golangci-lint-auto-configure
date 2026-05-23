# golangci-lint-auto-configure — TODO List

**Last Updated:** 2026-05-23

---

## Critical Priority

- [ ] Increase CLI integration test coverage (currently 9.0%)
- [ ] Enrich `CreateDefaultConfig()` output with default settings injection (not just exclusion paths/rules)

## High Priority

- [ ] Add report package tests (currently 0.0% — just added basic generator tests)
- [ ] Trim AGENTS.md from 912 to ≤377 lines (extract detailed docs to referenced files)
- [ ] Increase gogenfilter scanner coverage (currently 59.8%)
- [ ] Increase migration coverage (currently 66.8%)
- [ ] Remove `report_templ.go` from git tracking (generated file)

## Medium Priority

- [ ] Add `--check` mode for CI (exit 1 if config needs changes)
- [ ] Migrate `charmbracelet/fang` v1 → v2 (`charm.land/fang/v2`)
- [ ] Move `coverage.out` to `coverage/` directory
- [ ] Remove tracked binaries from git (`bin/`, root-level binary, `result`)
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
