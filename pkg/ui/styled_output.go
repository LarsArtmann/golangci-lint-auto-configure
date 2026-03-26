package ui

import "charm.land/lipgloss/v2"

// Color values as hex strings for compile-time safety.
// These are package-level constants to avoid magic numbers in code.

const (
	colorPrimary  = "#667eea"
	colorCritical = "#dc3545"
	colorHigh     = "#fd7e14"
	colorMedium   = "#ffc107"
	colorOptional = "#17a2b8"
	colorSuccess  = "#28a745"
	colorHeading  = "#ffffff"
	colorCodeFg   = "#e0e0e0"
	colorCodeBg   = "#2d2d44"
)

// SuccessMsg returns a success message.
func SuccessMsg(msg string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorSuccess)).
		Bold(true).
		Render("✓ " + msg)
}

// ErrorMsg returns an error message.
func ErrorMsg(msg string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorCritical)).
		Bold(true).
		Render("✗ " + msg)
}

// WarningMsg returns a warning message.
func WarningMsg(msg string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorHigh)).
		Bold(true).
		Render("⚠ " + msg)
}

// InfoMsg returns an info message.
func InfoMsg(msg string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorOptional)).
		Render("ℹ " + msg)
}

// PriorityBadge returns a styled priority badge.
func PriorityBadge(priority int) string {
	var style lipgloss.Style

	switch priority {
	case priorityCritical:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorHeading)).
			Background(lipgloss.Color(colorCritical)).
			Padding(0, 1)
	case priorityHigh:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorHeading)).
			Background(lipgloss.Color(colorHigh)).
			Padding(0, 1)
	case priorityMedium:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#212529")).
			Background(lipgloss.Color(colorMedium)).
			Padding(0, 1)
	default:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorHeading)).
			Background(lipgloss.Color(colorOptional)).
			Padding(0, 1)
	}

	name := priorityName(priority)

	return style.Render(name)
}

// Code returns styled code text.
func Code(text string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorCodeFg)).
		Background(lipgloss.Color(colorCodeBg)).
		Padding(0, 1).
		Render(text)
}

// SectionHeader returns a styled section header.
func SectionHeader(title string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorPrimary)).
		Bold(true).
		Render("━━ " + title + " ")
}

func priorityName(priority int) string {
	switch priority {
	case priorityCritical:
		return "CRITICAL"
	case priorityHigh:
		return "HIGH"
	case priorityMedium:
		return "MEDIUM"
	default:
		return "OPTIONAL"
	}
}
