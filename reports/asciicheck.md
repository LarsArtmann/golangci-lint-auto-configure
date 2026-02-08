# asciicheck Linter - Comprehensive Analysis

## What the Linter Does

**asciicheck** is a simple but important linter that detects non-ASCII characters in Go identifiers (variable names, function names, type names, constants, etc.). It specifically checks for Unicode characters that visually resemble ASCII characters but are actually different code points.

### Detailed Explanation

The linter scans Go source code for identifiers containing non-ASCII Unicode characters. The most common issue it catches involves **homoglyphs** - characters from different Unicode scripts that look visually identical or very similar to ASCII characters.

**Key detection areas:**

- Variable names (local and global)
- Function and method names
- Type names (structs, interfaces, etc.)
- Constants
- Package names
- Field names in structs
- Interface method signatures

**Example of the problem it solves:**

```go
// This looks correct but contains a Cyrillic 'е' (U+0435) instead of Latin 'e' (U+0065)
type TеstStruct struct{}  // This will be flagged

func main() {
    s := TestStruct{}  // Compiler error: undefined: TestStruct
    fmt.Println(s)
}
```

The code appears to define `TestStruct` and use it, but the definition uses a Cyrillic 'е' while the usage uses a Latin 'e', causing a compilation error that's extremely difficult to spot visually.

## When It Should Be Enabled

### Always Enable For:

- **Open source projects** with contributors from multiple locales
- **Team environments** where developers use different keyboard layouts
- **Public APIs and libraries** to ensure maximum compatibility
- **Codebases with CI/CD pipelines** to catch issues early
- **Projects with strict coding standards** requiring ASCII-only identifiers
- **Educational code** and examples where clarity is paramount
- **Command-line tools** where identifiers might be typed by users
- **Any Go project where code is reviewed on GitHub/GitLab** (difficult to spot in PRs)

### Specific Use Cases:

1. **Multi-national teams**: Prevents "copy-paste" bugs when sharing code between developers using different language keyboards
2. **Code review environments**: Non-ASCII characters are nearly impossible to spot in diffs
3. **Generated code validation**: Ensures code generators produce ASCII-only identifiers
4. **Security-sensitive applications**: Prevents homoglyph attacks in identifier names
5. **Teaching/Educational code**: Ensures students can reproduce examples without encoding issues

### Project Types:

- **CLI applications**: Users may need to reference identifiers in commands
- **Libraries/SDKs**: Consumers should not deal with non-ASCII identifiers
- **Enterprise applications**: Enterprise coding standards typically require ASCII-only
- **Infrastructure/Cloud tools**: Configuration and scripting integrations benefit from ASCII-only identifiers

## When It Should Be Disabled

### Appropriate to Disable For:

- **Domain-specific applications** where native language identifiers improve readability (e.g., scientific computing with Greek symbols, financial applications with currency symbols)
- **Projects with non-English teams** that intentionally use native language for business domain terminology
- **Educational projects** specifically teaching Unicode in programming
- **Personal/small team projects** where all members use the same locale and are aware of the risks

### Specific Scenarios:

1. **Intentional Unicode Identifiers**: When your project specifically uses Unicode to improve domain model expressiveness (rare but valid)
2. **Scientific/Mathematical Code**: When using Greek letters (α, β, γ) or mathematical symbols for formulas
3. **Localization Tools**: Tools specifically designed to work with internationalization
4. **Legacy Codebases**: When dealing with existing code that intentionally uses non-ASCII identifiers and refactoring is not feasible

**Note**: Even in these cases, it's recommended to keep asciicheck enabled and use `//nolint:asciicheck` comments for specific exceptions rather than disabling it entirely.

## Configuration Options

### Configuration Options

**asciicheck** has **no configuration options**. It's a binary check - either identifiers contain only ASCII characters, or they don't.

### Usage in golangci-lint

**Basic enablement:**

```yaml
linters:
  enable:
    - asciicheck
```

**Complete example configuration:**

```yaml
version: "2"

linters:
  default: none
  enable:
    - asciicheck
    - errcheck
    - gosec
    - staticcheck
    # ... other linters

issues:
  # Suppress individual lines if absolutely necessary
  exclude-rules:
    # Example: Allow specific non-ASCII usage in a particular file
    - path: internal/domain/unicode_types.go
      linters:
        - asciicheck
      text: "non-ASCII identifier"
```

**Project-specific configuration patterns:**

1. **Standard Projects** (Web applications, CLI tools, libraries):

```yaml
linters:
  enable:
    - asciicheck
```

2. **Monorepo with Domain-Specific Module**:

```yaml
version: "2"

linters:
  enable:
    - asciicheck

issues:
  exclude-rules:
    # Allow scientific notation in a specific package
    - path: pkg/mathematics/
      linters: [asciicheck]
      # Or more granular:
      # text: "identifier α contains non-ASCII character"
```

3. **With Autofix Warning**:

```yaml
version: "2"

linters:
  enable:
    - asciicheck

issues:
  exclude-use-default: false
  # Generate reports but don't fail CI for existing issues
  new-from-rev: HEAD~1
```

### Best Practices

1. **Always enable by default** in new projects
2. **Document exceptions**: Use `//nolint:asciicheck` with explanation comments
3. **Combine with code review**: Flag non-ASCII identifiers in style guides
4. **Educate team members**: Explain why this linter is important during onboarding
5. **Use in pre-commit hooks**: Catch issues before they enter version control

## How It Interferes or Works Together With Other Linters

### Synergistic Relationships

**Works Well With:**

- **`gofmt` / `gofumpt`**: ASCII-only identifiers align with standard Go formatting practices
- **`revive`**: Complements revive's style checks (e.g., `var-naming` rule)
- **`misspell`**: Both catch different types of text-related issues (non-ASCII vs spelling)
- **`gosimple` / `staticcheck`**: ASCII identifiers are simpler and more portable
- **`godox`**: Comments with TODO/FIXME should also use ASCII for consistency
- **`nolintlint`**: Ensures any `//nolint:asciicheck` directives are justified

**Example Synergy:**

```go
// Combined linting catches multiple issues:
type TеstStruct struct {  // asciicheck: non-ASCII in TеstStruct
    SomeField string // revive: exported type SomeField should have comment
    ΔValue    int    // asciicheck: non-ASCII in ΔValue
}
```

### Potential Conflicts

**Limited Conflicts:**

- **`asciicheck` has essentially no conflicts** because it operates on a very specific, non-overlapping concern
- No known linters encourage or require non-ASCII identifiers

**Special Considerations:**

1. **Internationalization (i18n) linter**: If a hypothetical linter checked for proper internationalization support, it might conflict philosophically, but not technically

2. **Generated Code**: May need exclusion in combination with other linters:

```yaml
issues:
  exclude-rules:
    - path: ".*_generated\.go$"
      linters:
        - asciicheck
        - misspell
        - gofmt
```

### Integration Patterns

**CI/CD Pipeline Integration:**

```bash
# In Makefile or CI script
lint:
    golangci-lint run --disable-all --enable=asciicheck,errcheck,gosec
    # asciicheck runs quickly and has no dependencies on other linters
```

**Pre-commit Hook:**

```yaml
# .pre-commit-config.yaml
- repo: https://github.com/golangci/golangci-lint
  rev: v1.64.5
  hooks:
    - id: golangci-lint
      args: ["--disable-all", "--enable=asciicheck,misspell,revive"]
```

### Performance Impact

- **Very low overhead**: Scans identifiers only, no complex analysis
- **Fast execution**: Typically completes in milliseconds even on large codebases
- **No dependencies**: Doesn't require type information or cross-package analysis
- **Ideal for quick feedback loops**: Can be run frequently during development

### Summary of Interactions

| Linter             | Relationship  | Reason                                                                |
| ------------------ | ------------- | --------------------------------------------------------------------- |
| `gofmt`, `gofumpt` | Complementary | Enforces ASCII-only identifiers which aligns with standard formatting |
| `revive`           | Complementary | Both enforce Go naming conventions and readability                    |
| `misspell`         | Complementary | Catches different text-related issues                                 |
| `staticcheck`      | Compatible    | No overlap; both improve code quality                                 |
| `nolintlint`       | Complementary | Ensures proper usage of nolint directives                             |
| `gci`              | Compatible    | Both contribute to clean, readable imports                            |

**Bottom Line**: asciicheck is a lightweight, conflict-free linter that should be enabled by default in virtually all Go projects. It provides significant value with zero configuration overhead and minimal performance cost.
