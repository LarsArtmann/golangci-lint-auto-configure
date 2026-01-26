# Example Configurations

This directory contains example `golangci-lint` configurations for different types of Go projects.

## Quick Start

```bash
# Copy the appropriate config to your project
cp examples/minimal.golangci.yml .golangci.yml
```

## Available Configurations

| File | Use Case | Linters | Complexity |
|------|----------|---------|------------|
| `minimal.golangci.yml` | Small projects, fast linting | 10 critical | Low |
| `standard.golangci.yml` | Most projects, balanced | 26 linters | Medium |
| `web-project.golangci.yml` | HTTP servers, APIs, microservices | 32 linters | Medium-High |
| `cli-project.golangci.yml` | Command-line tools, Cobra apps | 29 linters | Medium |
| `library.golangci.yml` | Reusable packages, SDKs | 34 linters | High |

## Configuration Details

### minimal.golangci.yml

**Best for:** Small projects, quick feedback, CI pipelines with time constraints

```yaml
# 10 critical linters only
# Fast execution (~30s)
# Focus on security and correctness
```

### standard.golangci.yml

**Best for:** Most production projects

```yaml
# 26 linters (critical + high value)
# Balanced coverage and speed
# Good for teams of any size
```

### web-project.golangci.yml

**Best for:** HTTP servers, REST APIs, microservices

```yaml
# Includes HTTP-specific linters:
# - contextcheck: Verifies context usage
# - bodyclose: Ensures HTTP response bodies are closed
# - Stricter function length limits
```

### cli-project.golangci.yml

**Best for:** Command-line tools, applications using Cobra/urfave-cli

```yaml
# Includes CLI-specific checks:
# - depguard: Restricts imports (no fmt for output)
# - containedctx: Checks for context in structs
# - Shorter function limits for readability
```

### library.golangci.yml

**Best for:** Reusable packages, open source libraries, SDKs

```yaml
# Strictest configuration
# 34 linters including:
# - varnamelen: Enforces meaningful variable names
# - interfacebloat: Prevents large interfaces
# - iface: Checks interface design
# Lower thresholds for complexity
```

## Customization

All configurations are starting points. Customize based on your needs:

```yaml
# Add to your .golangci.yml
linters:
  enable:
    - your-custom-linter
  disable:
    - linter-you-want-to-skip

linters-settings:
  your-linter:
    setting: value
```

## Validation

Verify your configuration works:

```bash
golangci-lint linters --config .golangci.yml
golangci-lint run --config .golangci.yml ./...
```