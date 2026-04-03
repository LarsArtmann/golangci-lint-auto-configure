package ui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/larsartmann/golangci-lint-auto-configure/pkg/types"
)

const (
	priorityCritical = 0
	priorityHigh     = 1
	priorityMedium   = 2
	priorityOptional = 3
	priorityCount    = 4
)

// FormatRecommendations formats linter recommendations for terminal output.
func FormatRecommendations(analysis *types.ConfigAnalysis) string {
	if len(analysis.LinterRecommendations) == 0 {
		return SuccessMsg("All recommended linters are already enabled")
	}

	priorityNames := []string{"Critical", "High Priority", "Medium Priority", "Optional"}

	return buildRecommendationsOutput(analysis, priorityNames)
}

func buildRecommendationsOutput(analysis *types.ConfigAnalysis, priorityNames []string) string {
	var output strings.Builder

	for priorityLevel := range priorityCount {
		group := filterByPriority(analysis.LinterRecommendations, priorityLevel)
		if len(group) == 0 {
			continue
		}

		output.WriteString("\n" + SectionHeader(priorityNames[priorityLevel]) + "\n\n")
		writePriorityGroup(&output, group, analysis)
	}

	return output.String()
}

func filterByPriority(recs []types.LinterRecommendation, priority int) []types.LinterRecommendation {
	var group []types.LinterRecommendation

	for _, rec := range recs {
		if int(rec.Priority) == priority {
			group = append(group, rec)
		}
	}

	return group
}

func writePriorityGroup(output *strings.Builder, group []types.LinterRecommendation, analysis *types.ConfigAnalysis) {
	for _, rec := range group {
		enabled := isLinterEnabled(rec.Name, analysis)
		status := getStatusSymbol(enabled)

		fmt.Fprintf(output, "  %s %s %s\n", status, PriorityBadge(int(rec.Priority)), Code(rec.Name.String()))
		fmt.Fprintf(output, "    %s\n\n", rec.Reason)
	}
}

func isLinterEnabled(recName types.LinterName, analysis *types.ConfigAnalysis) bool {
	return slices.ContainsFunc(analysis.EnabledLinters, func(l types.LinterInfo) bool {
		return l.Name.String() == recName.String()
	})
}

func getStatusSymbol(enabled bool) string {
	if enabled {
		return SuccessMsg("✓")
	}

	return WarningMsg("○")
}

// FormatSummary formats a summary for terminal output.
func FormatSummary(analysis *types.ConfigAnalysis) string {
	var output string

	output += "\n" + SectionHeader("Summary") + "\n\n"

	output += formatSummaryLine("Critical:", PriorityBadge(priorityCritical), analysis.CriticalCount)
	output += formatSummaryLine("High Priority:", PriorityBadge(priorityHigh), analysis.HighValueCount)
	output += formatSummaryLine("Medium Priority:", PriorityBadge(priorityMedium), analysis.MediumValueCount)
	output += formatSummaryLine("Optional:", PriorityBadge(priorityOptional), analysis.OptionalCount)
	output += fmt.Sprintf("  %-20s %s\n", "Enabled:", SuccessMsg(strconv.Itoa(len(analysis.EnabledLinters))))
	output += fmt.Sprintf("  %-20s %s\n", "Disabled:", ErrorMsg(strconv.Itoa(len(analysis.DisabledLinters))))

	return output
}

// FormatConfigHeader creates a styled header.
func FormatConfigHeader(configPath string) string {
	return "\n" + SectionHeader("golangci-lint Configuration") +
		"\n\n  Config: " + Code(configPath) + "\n\n"
}

// FormatDryRunWarning returns a dry-run warning.
func FormatDryRunWarning() string {
	return WarningMsg("DRY-RUN MODE - No changes will be made")
}

// FormatFixResult formats a fix result.
func FormatFixResult(result *types.MigrationResult) string {
	if result.IsSuccess() && result.FixesApplied == 0 {
		return SuccessMsg("No fixes needed - Configuration is optimal")
	}

	if result.IsSuccess() {
		return SuccessMsg(fmt.Sprintf("Applied %d fixes", result.FixesApplied)) +
			"\n\n  " + result.Message
	}

	return ErrorMsg("Fix failed") + "\n\n  " + result.Message
}

func formatSummaryLine(label, badge string, count int) string {
	return fmt.Sprintf("  %-20s %s\n", label, badge+" "+Code(strconv.Itoa(count)))
}
