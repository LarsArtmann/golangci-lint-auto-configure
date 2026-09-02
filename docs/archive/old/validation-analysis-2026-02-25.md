# Validation Analysis: 147 Projects Scan Results

**Date:** 2026-02-25\
**Command:** `golangci-lint-auto-configure validate`\
**Projects Scanned:** 147\
**Failures:** 10 projects (6.8% failure rate)

---

## Summary of Findings

Out of 147 projects scanned:

- **137 projects (93.2%)** - Configuration valid ✓
- **10 projects (6.8%)** - Configuration invalid ✗

### Failure Categories

| Category                           | Count | Projects                                                     |
| ---------------------------------- | ----- | ------------------------------------------------------------ |
| Deprecated/Invalid linter settings | 4     | complaints-mcp, desire-secrets, ast-state-analyzer, cmdguard |
| Missing version field              | 3     | KeyCountdown, SwettySwipperWeb, go-structure-linter          |
| Type mismatch (string vs array)    | 2     | StopTube, cmdguard                                           |
| Missing required field             | 1     | file-and-image-renamer                                       |

---

## Detailed Error Analysis

### 1. Deprecated/Invalid Linter Settings (4 projects)

**Most Common Issue** - Configuration contains settings that are no longer valid in golangci-lint v2.

#### complaints-mcp & desire-secrets (Identical errors)

Multiple deprecated settings:

```yaml
# INVALID: tagliatelle.use-field-name (property removed)
linters:
  settings:
    tagliatelle:
      use-field-name: true # REMOVED in v2

    # INVALID: cyclop.skip-tests (property removed)
    cyclop:
      skip-tests: true # REMOVED in v2

    # INVALID: modernize.disable values
    modernize:
      disable:
        - "loop" # INVALID: must be one of: 'any', 'bloop', 'fmtappendf', etc.

    # INVALID: sloglint.key-naming-case value
    sloglint:
      key-naming-case: "UPPER_SNAKE" # INVALID: must be 'snake', 'kebab', 'camel', or 'pascal'

    # INVALID: Additional properties not in schema
    gochecksumtype:
      exhaustive: true # additionalProperties 'exhaustive' not allowed
    fatcontext:
      ignore-len: 2 # additionalProperties 'ignore-len' not allowed
    gomodguard:
      local-replace-directives: false # additionalProperties not allowed
    gocritic:
      settings:
        appendAssign: { ... } # additionalProperties not allowed
    wrapcheck:
      ignoreSigRegexps: [...] # additionalProperties not allowed
```

#### ast-state-analyzer

```yaml
# INVALID: gosec.excludes must be array, not null
linters:
  settings:
    gosec:
      excludes: null # ERROR: got null, want array

    # INVALID: forbidigo.forbid items have extra property 'p'
    forbidigo:
      forbid:
        - p: "fmt\\.Print.*" # ERROR: additional properties 'p' not allowed
          msg: "use log instead"
      # CORRECT: Should be just the pattern string, or use 'pattern' key

    # INVALID: Removed settings
    nolintlint:
      allow-leading-space: true # REMOVED
    unused:
      check-exported: true # REMOVED
    exhaustive:
      check-generated: true # REMOVED
```

#### cmdguard

```yaml
# INVALID: gci.skip-generated not allowed
formatters:
  settings:
    gci:
      skip-generated: true # additionalProperties not allowed
```

**Root Cause:** These projects migrated from v1 to v2 but kept deprecated configuration options.

---

### 2. Missing Version Field (3 projects)

#### KeyCountdown & SwettySwipperWeb

```yaml
# ERROR: unsupported version of the configuration: ""
version: "" # MISSING or EMPTY
```

**Fix Required:**

```yaml
version: "2" # Required for v2 schema
```

#### go-structure-linter

```yaml
# ERROR: unsupported version of the configuration: "v1.26"
version: "v1.26" # OLD v1 CONFIG
```

**Fix Required:**

```yaml
version: "2" # Must migrate to v2 schema
```

---

### 3. Type Mismatch: string vs array (2 projects)

#### StopTube & cmdguard

```yaml
# ERROR: "formatters.settings.goimports.local-prefixes" does not validate:
#   got string, want array
formatters:
  settings:
    goimports:
      local-prefixes: "github.com/myorg" # WRONG: should be array
```

**Fix Required:**

```yaml
formatters:
  settings:
    goimports:
      local-prefixes: # Array, not string
        - "github.com/myorg"
        - "github.com/myorg/submodule"
```

---

### 4. Missing Required Field (1 project)

#### file-and-image-renamer

```yaml
# ERROR: run.timeout cannot be empty
run:
  timeout: "" # EMPTY - must have value
```

**Fix Required:**

```yaml
run:
  timeout: "5m" # Required format: duration string
```

---

## Patterns Discovered

### Pattern 1: Migration Artifacts

Projects that upgraded golangci-lint but didn't fully migrate their configs:

- Old v1 configs (`version: "v1.26"`)
- Empty version (`version: ""`)
- Removed linter settings still present

### Pattern 2: Schema Misunderstanding

Common type errors:

- `local-prefixes` as string instead of array
- `null` instead of `[]` for empty arrays
- Extra properties that don't exist in schema

### Pattern 3: Required Fields Missing

- `run.timeout` is required but often empty
- `version` field is mandatory in v2

---

## Recommendations for Tool Improvements

### Immediate Actions

1. **Add Auto-Migration for Common Issues**
   - Detect and fix `version: ""` → `version: "2"`
   - Detect and fix `local-prefixes: string` → `local-prefixes: [string]`
   - Remove known deprecated settings automatically

2. **Better Error Messages**
   - Parse jsonschema errors and provide human-readable fixes
   - Group related errors (e.g., "4 deprecated settings found")
   - Provide migration command suggestions

3. **Add Config Migration Command**
   ```bash
   golangci-lint-auto-configure migrate --from-v1
   ```

### Medium-Term Improvements

4. **Create Deprecated Settings Database**
   - Maintain mapping of removed/renamed settings
   - Include in LinterData for auto-fix

5. **Pre-Validation Checks**
   - Check for empty required fields before calling golangci-lint
   - Validate type correctness before schema validation

6. **Configuration Templates**
   - Generate v2-compliant templates for different project types
   - Include common valid settings examples

### Long-Term Enhancements

7. **Intelligent Config Repair**
   - Parse schema errors and auto-generate fixes
   - Interactive mode: "Found X issues. Fix automatically? [Y/n]"

8. **Version Compatibility Layer**
   - Detect v1 configs and offer to migrate
   - Support for version-specific validation

---

## Valid Config Example (from this project)

```yaml
version: "2"
run:
  timeout: 10m
  go: ""
  build-tags: []
  tests: true
linters:
  enable:
    - errcheck
    - gosec
    - staticcheck
    - govet
    # ... (full list)
  settings:
    cyclop:
      max-complexity: 15
    funlen:
      lines: 80
      statements: 50
formatters:
  enable:
    - golines
  settings:
    golines:
      max-len: 120
issues:
  max-issues-per-linter: 100
  max-same-issues: 15
```

---

## Action Items

- [ ] Add migration command to auto-fix v1 → v2 configs
- [ ] Create linter for detecting deprecated settings
- [ ] Add pre-validation for common type errors
- [ ] Improve error messages with specific fix suggestions
- [ ] Document all breaking changes between v1 and v2
- [ ] Create config templates for common project types
