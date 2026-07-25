package ui

import (
	"os"
	"strings"
	"testing"
)

// TestColorConstantsGolden guards against accidental color value changes.
// These hex values are part of the visual design contract — changing them
// alters the CLI output appearance. Update this test only when intentionally
// rebranding.
func TestColorConstantsGolden(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"colorPrimary", colorPrimary, "#667eea"},
		{"colorCritical", colorCritical, "#dc3545"},
		{"colorHigh", colorHigh, "#fd7e14"},
		{"colorMedium", colorMedium, "#ffc107"},
		{"colorOptional", colorOptional, "#17a2b8"},
		{"colorSuccess", colorSuccess, "#28a745"},
		{"colorHeading", colorHeading, "#ffffff"},
		{"colorCodeFg", colorCodeFg, "#e0e0e0"},
		{"colorCodeBg", colorCodeBg, "#2d2d44"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q (if this is intentional, update the golden value)", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestReportTemplateColorConsistency verifies that the HTML report template
// uses the same color palette as the CLI output constants. This catches
// drift between the two presentation layers.
func TestReportTemplateColorConsistency(t *testing.T) {
	templateBytes, err := os.ReadFile("../report/report.templ")
	if err != nil {
		t.Skipf("report.templ not found: %v", err)
	}

	template := string(templateBytes)

	// These are the colors shared between CLI and HTML report.
	sharedColors := map[string]string{
		"primary":  colorPrimary,
		"critical": colorCritical,
		"high":     colorHigh,
		"medium":   colorMedium,
		"optional": colorOptional,
		"success":  colorSuccess,
	}

	for name, hex := range sharedColors {
		if !strings.Contains(template, hex) {
			t.Errorf("report.templ is missing the %s color %q — CLI and HTML report colors are out of sync", name, hex)
		}
	}
}

// TestPriorityBackgroundColor verifies the priority-to-color mapping is exhaustive.
func TestPriorityBackgroundColor(t *testing.T) {
	tests := []struct {
		priority int
		expected string
	}{
		{1, colorCritical},
		{2, colorHigh},
		{3, colorMedium},
		{4, colorOptional},
		{99, colorOptional}, // default
	}

	for _, tt := range tests {
		got := priorityBackgroundColor(tt.priority)
		if got != tt.expected {
			t.Errorf("priorityBackgroundColor(%d) = %q, want %q", tt.priority, got, tt.expected)
		}
	}
}

// TestPriorityName verifies the priority label mapping.
func TestPriorityName(t *testing.T) {
	tests := []struct {
		priority int
		expected string
	}{
		{1, "CRITICAL"},
		{2, "HIGH"},
		{3, "MEDIUM"},
		{4, "OPTIONAL"},
		{99, "OPTIONAL"}, // default
	}

	for _, tt := range tests {
		got := priorityName(tt.priority)
		if got != tt.expected {
			t.Errorf("priorityName(%d) = %q, want %q", tt.priority, got, tt.expected)
		}
	}
}
