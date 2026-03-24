# Linting Issues Fix Plan

## Summary

- **Total Issues:** 162
- **Categories:** 23

## Prioritization Matrix

| Priority | Category         | Count | Impact           | Effort | Strategy                    |
| -------- | ---------------- | ----- | ---------------- | ------ | --------------------------- |
| P0       | gomoddirectives  | 2     | CI BLOCKER       | LOW    | Exclude in config           |
| P0       | godox            | 15    | Code Quality     | LOW    | Remove or add context       |
| P1       | gochecknoglobals | 25    | Architecture     | HIGH   | Refactor to options structs |
| P1       | forbidigo        | 1     | Style            | LOW    | Replace with logger         |
| P1       | paralleltest     | 18    | Test Quality     | MEDIUM | Add t.Parallel()            |
| P2       | tagliatelle      | 14    | API Consistency  | MEDIUM | Fix JSON tags               |
| P2       | revive           | 14    | Various          | LOW    | Fix per issue               |
| P2       | funcorder        | 4     | Code Style       | LOW    | Reorder methods             |
| P2       | varnamelen       | 10    | Readability      | LOW    | Rename variables            |
| P2       | goconst          | 3     | Maintainability  | LOW    | Extract constants           |
| P3       | noinlineerr      | 15    | Code Style       | LOW    | Refactor error handling     |
| P3       | wrapcheck        | 13    | Error Handling   | LOW    | Wrap errors                 |
| P3       | funlen           | 3     | Code Quality     | MEDIUM | Split functions             |
| P3       | gocognit         | 2     | Complexity       | MEDIUM | Refactor                    |
| P3       | gosec            | 7     | Security         | LOW    | Fix per issue               |
| P3       | mnd              | 3     | Maintainability  | LOW    | Extract constants           |
| P3       | testpackage      | 5     | Test Quality     | LOW    | Rename packages             |
| P3       | prealloc         | 1     | Performance      | LOW    | Preallocate slice           |
| P3       | nestif           | 1     | Code Quality     | LOW    | Simplify logic              |
| P3       | noctx            | 1     | Best Practice    | LOW    | Use CommandContext          |
| P3       | nilerr           | 1     | Bug              | LOW    | Fix error handling          |
| P3       | ireturn          | 2     | Interface Design | MEDIUM | Return concrete types       |
| P3       | godoclint        | 2     | Documentation    | LOW    | Consolidate docs            |

## Execution Plan

### Phase 1: Quick Wins (0-30 min)

| #   | Task                                            | Category   | Issues | Est Time |
| --- | ----------------------------------------------- | ---------- | ------ | -------- |
| 1   | Fix forbidigo - replace fmt.Println with logger | forbidigo  | 1      | 2 min    |
| 2   | Fix goconst - extract string literals           | goconst    | 3      | 5 min    |
| 3   | Fix mnd - extract magic numbers                 | mnd        | 3      | 5 min    |
| 4   | Fix varnamelen - rename short vars              | varnamelen | 10     | 10 min   |
| 5   | Fix nestif - simplify nested if                 | nestif     | 1      | 3 min    |
| 6   | Fix nilerr - fix error handling                 | nilerr     | 1      | 3 min    |
| 7   | Fix noctx - use CommandContext                  | noctx      | 1      | 3 min    |
| 8   | Fix prealloc - preallocate slice                | prealloc   | 1      | 2 min    |

### Phase 2: Low-Hanging Fruit (30-60 min)

| #   | Task                                    | Category    | Issues | Est Time |
| --- | --------------------------------------- | ----------- | ------ | -------- |
| 9   | Fix godox - add context or remove TODOs | godox       | 15     | 20 min   |
| 10  | Fix funcorder - reorder methods         | funcorder   | 4      | 10 min   |
| 11  | Fix testpackage - rename test packages  | testpackage | 5      | 10 min   |
| 12  | Fix godoclint - consolidate docs        | godoclint   | 2      | 5 min    |
| 13  | Fix revive (easy ones)                  | revive      | ~10    | 15 min   |

### Phase 3: Refactoring (60-120 min)

| #   | Task                                      | Category     | Issues | Est Time |
| --- | ----------------------------------------- | ------------ | ------ | -------- |
| 14  | Fix noinlineerr - refactor error handling | noinlineerr  | 15     | 25 min   |
| 15  | Fix wrapcheck - wrap errors properly      | wrapcheck    | 13     | 20 min   |
| 16  | Fix gosec - fix security issues           | gosec        | 7      | 15 min   |
| 17  | Fix paralleltest - add t.Parallel()       | paralleltest | 18     | 20 min   |
| 18  | Fix funlen - split long functions         | funlen       | 3      | 30 min   |
| 19  | Fix gocognit - reduce complexity          | gocognit     | 2      | 25 min   |

### Phase 4: Major Refactoring (120+ min)

| #   | Task                                       | Category         | Issues | Est Time |
| --- | ------------------------------------------ | ---------------- | ------ | -------- |
| 20  | Fix ireturn - return concrete types        | ireturn          | 2      | 30 min   |
| 21  | Fix gochecknoglobals - refactor to options | gochecknoglobals | 25     | 60 min   |
| 22  | Fix tagliatelle - fix JSON tags            | tagliatelle      | 14     | 40 min   |
| 23  | Fix remaining revive issues                | revive           | ~4     | 15 min   |

### Phase 5: Configuration

| #   | Task                                | Category        | Issues | Est Time |
| --- | ----------------------------------- | --------------- | ------ | -------- |
| 24  | Configure gomoddirectives exclusion | gomoddirectives | 2      | 5 min    |

---

## Critical Path

1. **gomoddirectives** → CI blocker, must exclude
2. **godox** → Low effort, high impact on code quality
3. **goconst + mnd** → Quick wins for maintainability
4. **gochecknoglobals** → Major architectural improvement
5. **paralleltest** → Test quality improvement
6. **tagliatelle** → API consistency

---

_Generated: 2026-03-24 06:00_
