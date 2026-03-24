package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/types"
)

// LinterItem represents a linter with formatted output.
type LinterItem struct {
	Name     string
	Priority int
	Reason   string
	Enabled  bool
}

// FormatLinterRecommendations formats linter recommendations for terminal output.
func FormatLinterRecommendations(recommendations []types.LinterRecommendation, enabledLinters []string) string {
	if len(recommendations) == 0 {
		return SuccessStyle.Render("✓ All recommended linters are already enabled!")
	}

	var builder strings.Builder

	// Group by priority
	priorities := make(map[int][]LinterItem)
	for _, rec := range recommendations {
		enabled := false
		for _, e := range enabledLinters {
			if e == rec.Name.String() {
				enabled = true
				break
			}
		}
		priorities[int(rec.Priority)] = append(priorities[int(rec.Priority)], LinterItem{
			Name:     rec.Name.String(),
			Priority: int(rec.Priority),
			Reason:   rec.Reason,
			Enabled:  enabled,
		})
	}

	// Output each priority group
	priorityNames := []string{"Critical", "High Priority", "Medium Priority", "Optional"}
	priorityColors := []lipgloss.Style{CriticalBadgeStyle, HighBadgeStyle, MediumBadgeStyle, OptionalBadgeStyle}

	for i := 0; i <= 3; i++ {
		items, exists := priorities[i]
		if !exists || len(items) == 0 {
			continue
		}

		builder.WriteString("\n")
		builder.WriteString(SectionDivider(priorityNames[i]))
		builder.WriteString("\n")

		for _, item := range items {
			badge := priorityColors[i].Render(GetPriorityName(i))
			status := ""
			if item.Enabled {
				status = SuccessStyle.Render("✓")
			} else {
				status = WarningStyle.Render("○")
			}

			builder.WriteString(fmt.Sprintf("  %s %s %s\n", status, badge, item.Name))
			builder.WriteString(SubheadingStyle.Render(item.Reason))
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// FormatSummary formats a summary for terminal output.
func FormatSummary(analysis *types.ConfigAnalysis) string {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(SectionDivider("Summary"))
	builder.WriteString("\n\n")

	// Stats in a grid-like format
	stats := []struct {
		label string
		value int
		style lipgloss.Style
	}{
		{"Critical", analysis.CriticalCount, CriticalBadgeStyle},
		{"High Priority", analysis.HighValueCount, HighBadgeStyle},
		{"Medium Priority", analysis.MediumValueCount, MediumBadgeStyle},
		{"Optional", analysis.OptionalCount, OptionalBadgeStyle},
		{"Enabled", len(analysis.EnabledLinters), SuccessStyle},
		{"Disabled", len(analysis.DisabledLinters), ErrorStyle},
	}

	for _, stat := range stats {
		valueStr := stat.style.Render(fmt.Sprintf("%d", stat.value))
		labelStr := BodyStyle.Render(stat.label)
		builder.WriteString(fmt.Sprintf("  %-15s %s\n", labelStr, valueStr))
	}

	// Add enabled/disabled linters list
	if len(analysis.EnabledLinters) > 0 {
		builder.WriteString("\n")
		builder.WriteString(HeadingStyle.Render("Enabled Linters"))
		builder.WriteString("\n")
		for _, linter := range analysis.EnabledLinters {
			builder.WriteString(fmt.Sprintf("  %s %s\n", SuccessStyle.Render("✓"), CodeStyle.Render(linter.Name.String())))
		}
	}

	if len(analysis.DisabledLinters) > 0 {
		builder.WriteString("\n")
		builder.WriteString(HeadingStyle.Render("Disabled Linters"))
		builder.WriteString("\n")
		for _, linter := range analysis.DisabledLinters {
			builder.WriteString(fmt.Sprintf("  %s %s\n", ErrorStyle.Render("✗"), CodeStyle.Render(linter.Name.String())))
		}
	}

	return builder.String()
}

// FormatFixResult formats a fix result for terminal output.
func FormatFixResult(result *types.MigrationResult) string {
	var builder strings.Builder

	if result.Success {
		if result.FixesApplied == 0 {
			builder.WriteString(SuccessStyle.Render("✓ No fixes needed"))
			builder.WriteString(" - ")
			builder.WriteString(BodyStyle.Render("Configuration is already optimal"))
		} else {
			builder.WriteString(SuccessStyle.Render(fmt.Sprintf("✓ Applied %d fixes", result.FixesApplied)))
			builder.WriteString("\n")
			builder.WriteString(SubheadingStyle.Render(result.Message))
		}
	} else {
		builder.WriteString(ErrorStyle.Render("✗ Fix failed"))
		builder.WriteString("\n")
		builder.WriteString(ErrorStyle.Render(result.Message))
	}

	return builder.String()
}

// FormatConfigHeader creates a styled header for configuration output.
func FormatConfigHeader(configPath string) string {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(TitleStyle.Render("golangci-lint Configuration"))
	builder.WriteString("\n\n")
	builder.WriteString(SubheadingStyle.Render("Config: " + configPath))
	builder.WriteString("\n")

	return builder.String()
}

// FormatDryRunWarning creates a styled dry-run warning.
func FormatDryRunWarning() string {
	return WarningStyle.Render("⚠ DRY-RUN MODE - No changes will be made")
}
