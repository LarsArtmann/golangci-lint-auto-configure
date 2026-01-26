// Example of using golangci-linter-auto-configure as a library
package main

import (
	"fmt"
	"os"

	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/client"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

func main() {
	fmt.Println("🔍 golangci-linter-auto-configure API Usage Example")
	fmt.Println()

	// Create client with verbose logging
	c := client.New(client.Options{
		Verbose: true,
	})

	// Analyze configuration
	configPath := ".golangci.yml"
	analysis, err := c.AnalyzeConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Analysis failed: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("✅ Analysis complete: %s\n", configPath)
	fmt.Printf("   - Enabled linters: %d\n", len(analysis.EnabledLinters))
	fmt.Printf("   - Disabled linters: %d\n", len(analysis.DisabledLinters))
	fmt.Printf("   - Enabled formatters: %d\n", len(analysis.EnabledFormatters))
	fmt.Printf("   - Disabled formatters: %d\n", len(analysis.DisabledFormatters))
	fmt.Printf("   - Recommendations: %d\n", len(analysis.LinterRecommendations))
	fmt.Println()

	// Show critical recommendations
	var critical []types.LinterRecommendation
	var high []types.LinterRecommendation

	for _, rec := range analysis.LinterRecommendations {
		switch rec.Priority {
		case types.LinterPriorityCritical:
			critical = append(critical, rec)
		case types.LinterPriorityHigh:
			high = append(high, rec)
		}
	}

	if len(critical) > 0 {
		fmt.Println("🚨 Critical linters (should ALWAYS be enabled):")
		for _, rec := range critical {
			fmt.Printf("   - %s\n", rec.Name)
		}
		fmt.Println()
	}

	if len(high) > 0 {
		fmt.Println("⚠️  High priority linters (recommended):")
		for _, rec := range high {
			fmt.Printf("   - %s: %s\n", rec.Name, rec.Reason)
		}
		fmt.Println()
	}

	// Show summary
	fmt.Println("📊 Summary:")
	fmt.Println(c.GetSummary(analysis))
}
