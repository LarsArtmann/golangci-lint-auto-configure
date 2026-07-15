# Linter Documentation Template

## {{LINTER_NAME}} Linter - Comprehensive Analysis

## What It Does

[1-2 paragraphs explaining what the linter does, its core purpose, and the specific problems it detects]

### The Problem It Detects

[Detailed explanation of the specific issues, bugs, or anti-patterns this linter identifies]

**Why This Matters:**

- [Bullet points explaining why these issues are important]
- [Impact on code quality, security, performance, etc.]

### How It Works

[Technical explanation of how the linter analyzes code]
[Algorithm or pattern matching approach]
[Analysis scope and limitations]

### Examples

```go
// ❌ BAD: Example of issue this linter detects
[Brief explanation]

// ✅ GOOD: Example of how to fix the issue
[Brief explanation]
```

## When It Should Be Enabled

### ✅ Enable For:

**Project Types:**

| Project Type    | Priority               | Justification   |
| --------------- | ---------------------- | --------------- |
| [Project types] | [CRITICAL/HIGH/MEDIUM] | [Why important] |

**Specific Scenarios:**

**1. [Scenario name]**

- [When to enable]
- [Examples]

**2. [Scenario name]**

- [When to enable]
- [Examples]

### ❌ Disable For:

**Specific Scenarios:**

**1. [Scenario name]**

```yaml
[Example YAML exclusion]
```

**2. [Scenario name]**

- [Reason to disable]

### Priority Assessment

- **Default Priority**: [CRITICAL/HIGH/MEDIUM/OPTIONAL]
- **Value**: [HIGHEST/HIGH/MEDIUM/LOW] - [Why important]
- **Effort**: [Low/Medium/High] - [How easy to fix findings]
- **Recommendation**: [ALWAYS/RECOMMEND/CONSIDER] - [When to enable]

## How It Should Be Configured

### Configuration Options

```yaml
# golangci-lint v2 format
version: "2"

linters:
  settings:
    { { LINTER_NAME } }:
      # Option 1
      # Type: [type]
      # Default: [default value]
      # Description: [what it does]
      option-name: value

      # Option 2
      # Type: [type]
      # Default: [default value]
      # Description: [what it does]
      option-name: value
```

### [Option Name] Option

- **Type**: `[type]`
- **Default**: `[default]`
- **Description**: [Detailed explanation of what this option controls]

**Format:** [How to specify values]

**Common Values:**

- [Value 1] - [When to use]
- [Value 2] - [When to use]

### Recommended Configurations

#### ✅ Standard Configuration (Recommended)

```yaml
# [Use case description]
version: "2"
linters:
  settings:
    { { LINTER_NAME } }: [options]
```

#### ✅ Strict Configuration

```yaml
# [Use case description]
version: "2"
linters:
  settings:
    { { LINTER_NAME } }: [options]

issues:
  exclude-rules:
    - path: [pattern]
      linters: [{ { LINTER_NAME } }]
```

#### ✅ [Custom Configuration Name]

```yaml
# [Use case description]
version: "2"
linters:
  settings:
    { { LINTER_NAME } }: [options]
```

## How It Interferes or Works Together With Other Linters

### ✅ Complementary Linters

{{LINTER_NAME}} works excellently with:

| Linter        | Relationship        | Value              |
| ------------- | ------------------- | ------------------ |
| **[linter1]** | [relationship type] | [what it provides] |
| **[linter2]** | [relationship type] | [what it provides] |
| **[linter3]** | [relationship type] | [what it provides] |

**Complete [Category] Suite:**

```yaml
linters:
  enable:
    - { { LINTER_NAME } } # [Purpose] ([PRIORITY])
    - [other linter] # [Purpose] ([PRIORITY])
    - [other linter] # [Purpose] ([PRIORITY])
```

### Example of Linter Synergy

```go
// [linter1] catches:
[Code example with issue]

// [linter2] adds:
[Fixed code example]

// [linter3] adds:
[Further improved code example]
```

### 🔒 No Conflicts / Minimal Overlap

| Linter                         | Overlap         | Recommendation  |
| ------------------------------ | --------------- | --------------- |
| {{LINTER_NAME}} + **[linter]** | [What overlaps] | [How to handle] |

**Why No Conflicts:**

- [Explanation of why linters work together]
- [Each linter covers different aspects]

## Practical Examples

### ✅ Example 1: [Example Title]

```go
// ❌ BAD: [Brief description]
[Code example showing issue]
[Brief explanation]

// ✅ GOOD: [Brief description]
[Code example showing fix]
[Brief explanation]
```

### ✅ Example 2: [Example Title]

```go
// ❌ BAD: [Brief description]
[Code example showing issue]

// ✅ GOOD: [Brief description]
[Code example showing fix]
[Brief explanation]
```

### ✅ Example 3: [Example Title]

```go
// ❌ BAD: [Brief description]
[Code example showing issue]

// ✅ GOOD: [Brief description]
[Code example showing fix]
[Brief explanation]
```

### ✅ Example 4: [Example Title]

```go
// ❌ BAD: [Brief description]
[Code example showing issue]

// ✅ GOOD: [Brief description]
[Code example showing fix]
[Brief explanation]
```

### ✅ Example 5: [Example Title]

```go
// ❌ BAD: [Brief description]
[Code example showing issue]

// ✅ GOOD: [Brief description]
[Code example showing fix]
[Brief explanation]
```

## Best Practices

1. **[Guideline 1]** - [Explanation]
2. **[Guideline 2]** - [Explanation]
3. **[Guideline 3]** - [Explanation]
4. **[Guideline 4]** - [Explanation]
5. **[Guideline 5]** - [Explanation]
6. **[Guideline 6]** - [Explanation]
7. **[Guideline 7]** - [Explanation]
8. **[Guideline 8]** - [Explanation]
9. **[Guideline 9]** - [Explanation]
10. **[Guideline 10]** - [Explanation]

## Common Scenarios and Solutions

### Scenario 1: [Scenario Title]

**Problem:** [Description of issue]

**Solution:** [How to fix it]

```yaml
[Example YAML configuration]
```

### Scenario 2: [Scenario Title]

**Problem:** [Description of issue]

**Solution:** [How to fix it]

```yaml
[Example YAML configuration]
```

### Scenario 3: [Scenario Title]

**Problem:** [Description of issue]

**Solution:** [How to fix it]

```yaml
[Example YAML configuration]
```

### Scenario 4: [Scenario Title]

**Problem:** [Description of issue]

**Solution:** [How to fix it]

```yaml
[Example YAML configuration]
```

### Scenario 5: [Scenario Title]

**Problem:** [Description of issue]

**Solution:** [How to fix it]

```yaml
[Example YAML configuration]
```

## Summary

**{{LINTER_NAME}}** is a **[PRIORITY]** linter that [brief description]:

- ✅ **[Key benefit 1]** - [Explanation]
- ✅ **[Key benefit 2]** - [Explanation]
- ✅ **[Key benefit 3]** - [Explanation]
- ✅ **[Key benefit 4]** - [Explanation]
- ✅ **[Key benefit 5]** - [Explanation]
- ✅ **[Key benefit 6]** - [Explanation]
- ✅ **[Key benefit 7]** - [Explanation]
- ✅ **[Key benefit 8]** - [Explanation]
- ⚠️ **[Limitation 1]** - [Explanation]
- ⚠️ **[Limitation 2]** - [Explanation]

**Recommendation:** **[RECOMMENDATION]** [detailed guidance]. [Top 3 tips].

---

**Reference:** [GitHub repository or documentation link]
