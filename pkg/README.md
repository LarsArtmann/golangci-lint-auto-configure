# pkg/ - Public API Package

This directory contains the public API for `golangci-lint-auto-configure`, designed for integration into other Go applications.

## Quick Start

```go
import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/client"

// Simple one-line analysis
summary, err := client.SimpleAnalyze(client.Options{
    Verbose: true,
}, ".golangci.yml")
if err != nil {
    log.Fatal(err)
}
fmt.Println(summary)
```

## Package Overview

### client - High-Level API (Recommended)

The `client` package provides a simplified, production-ready API that abstracts away internal complexity.

```go
import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/client"

c := client.New(client.Options{
    Verbose: true,
})

// Analyze configuration
analysis, err := c.AnalyzeConfig(".golangci.yml")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d disabled linters\n", len(analysis.DisabledLinters))
fmt.Printf("Found %d disabled formatters\n", len(analysis.DisabledFormatters))
```

### config - Configuration Management

Load, save, and validate golangci-lint configuration files.

```go
import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"

loader := config.NewLoader(logger)

cfg, err := loader.LoadConfig(".golangci.yml")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Timeout: %s\n", cfg.Run.Timeout)
fmt.Printf("Enabled linters: %v\n", cfg.Linters.Enable)

// Save configuration
err = loader.SaveConfig(cfg, ".golangci.yml.new")
```

### linter - Analysis Engine

Analyze configurations and generate recommendations.

```go
import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/linter"

analyzer := linter.NewAnalyzer(logger)

analysis, err := analyzer.AnalyzeConfig(".golangci.yml")
if err != nil {
    log.Fatal(err)
}

fmt.Println(analyzer.FormatRecommendations(analysis))
```

### types - Type Definitions

Core types used throughout the API.

```go
import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"

// LinterName provides type safety
var linterName types.LinterName = "gosec"

// FormatterName provides type safety
var formatterName types.FormatterName = "gofumpt"
```

### report - Report Generation

Generate HTML and JSON reports from analysis results.

```go
import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/report"

generator := report.NewGenerator(logger)
err := generator.GenerateReport(analysis, "report.html")
```

### constants - Configuration Data

Constants for linter and formatter metadata.

```go
import "github.com/larsartmann/golangcli-linter-auto-configure/pkg/constants"

// Access linter priorities
priority := constants.LinterPriorities["gosec"]

// Access formatter info
formatter := constants.FormatterInfo["gofumpt"]
```

### finding - go-finding Integration

Convert analysis results to the unified [go-finding](https://github.com/larsartmann/go-finding) data model.

```go
import appfinding "github.com/larsartmann/golangci-lint-auto-configure/pkg/finding"

// Convert analysis to go-finding Report
report := appfinding.AnalysisToReport(analysis, "1.0.0")

// Export as SARIF for CI/CD integration
sarif, _ := report.ToSARIF()

// Parse golangci-lint JSON output
findings, _ := appfinding.ParseGolangciLintJSON(jsonData)

// Filter by category
security := finding.Filter(findings, finding.ByCategory(finding.CategorySecurity))
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/charmbracelet/log"
    "github.com/larsartmann/golangcli-linter-auto-configure/pkg/client"
    "github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

func main() {
    // Create client
    c := client.New(client.Options{
        Verbose: true,
    })

    // Analyze configuration
    analysis, err := c.AnalyzeConfig(".golangci.yml")
    if err != nil {
        log.Fatal(err)
    }

    // Print summary
    fmt.Println(c.GetSummary(analysis))

    // Check for recommended linters
    for _, rec := range analysis.LinterRecommendations {
        if rec.Priority >= types.LinterPriorityHigh {
            fmt.Printf("RECOMMENDATION: Enable %s - %s\n", rec.Name, rec.Reason)
        }
    }

    // Check for recommended formatters
    for _, rec := range analysis.FormatterRecommendations {
        if rec.Priority >= types.FormatterPriorityMedium {
            fmt.Printf("RECOMMENDATION: Enable %s - %s\n", rec.Name, rec.Reason)
        }
    }
}
```

## Feature Support

### ✅ Implemented Features

- [x] Linter analysis and recommendations
- [x] Formatter analysis and recommendations (v2.8+)
- [x] Configuration loading and validation
- [x] HTML and JSON report generation
- [x] SARIF 2.1.0 report output (via go-finding)
- [x] go-finding unified data model integration
- [x] golangci-lint JSON output parsing to Findings
- [x] High-level client API with context support
- [x] Strong typing with LinterName and FormatterName
- [x] Priority-based recommendations (Critical/High/Medium/Optional)

### 🚧 Planned Features

- [ ] Configuration auto-fixing API
- [ ] Bulk analysis of multiple configurations
- [ ] Custom formatter/linter definitions
- [ ] Integration with CI/CD systems

## Version Compatibility

- **golangci-lint**: v2.10.1 or higher required
- **Go**: 1.21 or higher required
- **Breaking Changes**: Public API is stable, but internal packages may change

## Error Handling

All operations return structured errors:

```go
analysis, err := c.AnalyzeConfig(".golangci.yml")
if err != nil {
    if errors.Is(err, errors.ErrNotFound) {
        log.Fatal("Config file not found")
    }
    log.Fatalf("Analysis failed: %v", err)
}
```

## Performance Considerations

- Analysis typically takes 1-3 seconds
- All operations are synchronous (use goroutines if needed)
- No caching - each call executes fresh analysis
- KeepAlive not required for short-lived analysis

## Best Practices

1. **Use the client package** for most integrations
2. **Handle errors gracefully** - configuration files may be malformed
3. **Respect priority levels** - Critical linters should always be enabled
4. **Consider formatter recommendations** - code formatting improves maintainability
5. **Generate reports for CI/CD** - HTML reports provide good visibility

## License

See /LICENSE for license information.
