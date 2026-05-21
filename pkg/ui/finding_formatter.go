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

		for _, finding := range categoryFindings {
			severityBadge := severityBadge(finding.Severity)
			fmt.Fprintf(
				&builder, "  %s [%s] %s: %s\n",
				severityBadge,
				finding.Rule,
				finding.Position,
				finding.Message,
			)

			if finding.Suggestion != "" {
				fmt.Fprintf(&builder, "    → %s\n", finding.Suggestion)
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

	for _, severity := range []finding.Severity{
		finding.SeverityCritical,
		finding.SeverityError,
		finding.SeverityWarning,
		finding.SeverityInfo,
	} {
		if group, ok := bySeverity[severity]; ok {
			parts = append(parts, strconv.Itoa(len(group))+" "+severityLabel(severity))
		}
	}

	return "Findings: " + strings.Join(parts, ", ")
}

func severityLabel(severity finding.Severity) string {
	return severityString(severity, "critical", "error", "warning", "info", "unknown")
}

func severityBadge(severity finding.Severity) string {
	return severityString(severity, "\U0001f534", "\U0001f7e0", "\U0001f7e1", "\U0001f535", "\u26aa")
}

func severityString(severity finding.Severity, critical, err, warning, info, fallback string) string {
	switch severity {
	case finding.SeverityCritical:
		return critical
	case finding.SeverityError:
		return err
	case finding.SeverityWarning:
		return warning
	case finding.SeverityInfo:
		return info
	default:
		return fallback
	}
}
