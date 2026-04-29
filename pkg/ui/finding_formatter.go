package ui

import (
	"fmt"
	"strconv"
	"strings"

	finding "github.com/larsartmann/go-finding"
)

// FormatFindings renders go-finding Findings in styled terminal output.
func FormatFindings(findings []finding.Finding) string {
	if len(findings) == 0 {
		return InfoMsg("No findings")
	}

	var builder strings.Builder

	groups := finding.GroupByCategory(findings)

	for category, categoryFindings := range groups {
		builder.WriteString(SectionHeader(string(category)))
		fmt.Fprintf(&builder, " (%d)\n", len(categoryFindings))

		for _, f := range categoryFindings {
			severityBadge := severityBadge(f.Severity)
			fmt.Fprintf(&builder, "  %s [%s] %s: %s\n",
				severityBadge,
				f.Rule,
				f.Position,
				f.Message,
			)

			if f.Suggestion != "" {
				fmt.Fprintf(&builder, "    → %s\n", f.Suggestion)
			}
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

// FormatFindingsSummary renders a compact summary of findings.
func FormatFindingsSummary(findings []finding.Finding) string {
	if len(findings) == 0 {
		return InfoMsg("No findings")
	}

	bySeverity := finding.GroupBySeverity(findings)

	var parts []string
	if group, ok := bySeverity[finding.SeverityCritical]; ok {
		parts = append(parts, strconv.Itoa(len(group))+" critical")
	}

	if group, ok := bySeverity[finding.SeverityError]; ok {
		parts = append(parts, strconv.Itoa(len(group))+" error")
	}

	if group, ok := bySeverity[finding.SeverityWarning]; ok {
		parts = append(parts, strconv.Itoa(len(group))+" warning")
	}

	if group, ok := bySeverity[finding.SeverityInfo]; ok {
		parts = append(parts, strconv.Itoa(len(group))+" info")
	}

	return "Findings: " + strings.Join(parts, ", ")
}

func severityBadge(severity finding.Severity) string {
	switch severity {
	case finding.SeverityCritical:
		return "\U0001f534"
	case finding.SeverityError:
		return "\U0001f7e0"
	case finding.SeverityWarning:
		return "\U0001f7e1"
	case finding.SeverityInfo:
		return "\U0001f535"
	default:
		return "\u26aa"
	}
}
