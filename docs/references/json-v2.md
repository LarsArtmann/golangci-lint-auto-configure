# encoding/json/v2 Migration

The codebase uses `encoding/json/v2` + `encoding/json/jsontext` (experimental stdlib in Go 1.26). This document covers the behavioral changes, the `GOEXPERIMENT=jsonv2` requirement, and the wire-format decoupling pattern.

## GOEXPERIMENT=jsonv2

`encoding/json/v2` is behind the `GOEXPERIMENT=jsonv2` experiment flag in Go 1.26. Without it, any import of `encoding/json/v2` or `encoding/json/jsontext` fails at compile time:

```
build constraints exclude all Go files in encoding/json/v2
```

### Where it's configured

| Location                       | How                                                                         |
| ------------------------------ | --------------------------------------------------------------------------- |
| **flake.nix** package build    | `env.GOEXPERIMENT = "jsonv2"`                                               |
| **flake.nix** devShell         | `env.GOEXPERIMENT = "jsonv2"`                                               |
| **flake.nix** CI shell         | `GOEXPERIMENT = "jsonv2"`                                                   |
| **GitHub Actions ci.yml**      | `env: GOEXPERIMENT: jsonv2` on `test-and-build`, `lint`, `govulncheck` jobs |
| **GitHub Actions release.yml** | `env: GOEXPERIMENT: jsonv2` on `release` job                                |

When running `go` commands directly outside `nix develop`, export it manually:

```bash
export GOEXPERIMENT=jsonv2
```

## Behavioral Changes from v1 → v2

### Case-sensitive field matching

json/v1 matched struct fields **case-insensitively** by default. json/v2 is **case-sensitive** — if a JSON key is `"name"` but the struct field is `Name` with no tag, json/v2 will NOT match it.

**Impact:** All JSON tags must exactly match the wire format. This is why we added wire-format decoupling structs (see below).

### nil slices marshal as `[]`, not `null`

```go
// v1: []byte(nil) → null
// v2: []byte(nil) → []
```

json/v2 marshals nil slices as `[]` instead of `null`. This is generally safer for JSON API consumers, but is a behavioral change to be aware of.

### `[]byte` marshals as base64

json/v2 marshals `[]byte` as base64-encoded strings (same as v1, but v2 does NOT allow `null` for `[]byte` fields during unmarshal).

### Marshal/Unmarshal API changes

json/v2's `Marshal` and `Unmarshal` accept functional options instead of separate function variants:

```go
// v1
data, _ := json.MarshalIndent(v, "", "  ")

// v2
data, _ := json.Marshal(v, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
```

### No `json.Encoder` / `json.Decoder`

json/v2 uses `jsontext.Encoder` / `jsontext.Decoder` for streaming. The old `json.NewEncoder(w)` pattern becomes `jsontext.NewEncoder(w)`.

### time.Time formatting

json/v2 uses RFC 3339 with nanosecond precision by default (same as v1), but the internal representation changed. No action needed unless you rely on exact byte output.

## Wire-Format Decoupling Pattern

### The Problem

golangci-lint's JSON output uses a mix of casing conventions:

| JSON key                               | Type                                  | Source struct                               |
| -------------------------------------- | ------------------------------------- | ------------------------------------------- |
| `"Enabled"`, `"Disabled"`              | Capitalized wrapper keys              | `lintersHelp` (pkg/commands/linters.go)     |
| `"name"`, `"description"`, `"autoFix"` | lowercase field keys                  | `linterHelp` (pkg/commands/help_linters.go) |
| `"version"`, `"goVersion"`             | lowercase field keys                  | `BuildInfo` (pkg/commands/version.go)       |
| `"Issues"`, `"FromLinter"`             | Capitalized (Go field names, no tags) | `JSONResult`, `result.Issue`                |

json/v2 is case-sensitive. Our Report types (`LinterInfo`, `FormatterInfo` in `pkg/types/types.go`) are **tag-free** — they use Go field names (PascalCase) for JSON output, enforced by `json_tags_test.go`.

We cannot add lowercase tags to Report types without breaking the PascalCase policy. And we cannot parse golangci-lint's lowercase wire format without matching tags.

### The Solution

Dedicated wire-format structs in `pkg/linter/analyzer.go` that exactly match golangci-lint's output, with conversion methods to Report types:

```go
// Wire format — matches golangci-lint output exactly
type golangciLinterEntry struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Fast        bool     `json:"fast"`
    AutoFix     bool     `json:"autoFix"`
    Deprecated  bool     `json:"deprecated"`
    Since       string   `json:"since"`
    OriginalURL string   `json:"originalURL"`
}

// Converts wire format → Report type (tag-free PascalCase)
func (e golangciLinterEntry) toLinterInfo() types.LinterInfo { ... }
```

The flow: parse JSON into wire-format structs → convert to Report types → use throughout the application.

### Where this pattern is used

| File                            | Wire structs                                      | Converts to                               |
| ------------------------------- | ------------------------------------------------- | ----------------------------------------- |
| `pkg/linter/analyzer.go`        | `golangciLinterEntry`, `golangciFormatterEntry`   | `types.LinterInfo`, `types.FormatterInfo` |
| `pkg/config/loader.go`          | `LinterList` (with `"Enabled"`/`"Disabled"` tags) | internal processing                       |
| `pkg/finding/golangci_lint.go`  | `GolangciLintIssue` (Go field names as JSON keys) | `finding.Finding`                         |
| `pkg/linter/version_checker.go` | `golangciLintVersion` (lowercase tags)            | version comparison                        |

### External format types and tagliatelle

External format types (`LinterList`, `golangciLintOutput`, `golangciLintVersion`, `GolangciLintIssue`) are **excluded from tagliatelle** in `.golangci.yml` because their tags must match external JSON shapes, not the project's PascalCase/kebab-case policy.
