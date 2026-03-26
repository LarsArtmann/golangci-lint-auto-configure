package ui

import "charm.land/lipgloss/v2"

// Color definitions - intentional global constants for consistent UI styling.
//
//nolint:gochecknoglobals
var (
	PrimaryColor    = lipgloss.Color("#667eea")
	CriticalColor   = lipgloss.Color("#dc3545")
	HighColor       = lipgloss.Color("#fd7e14")
	MediumColor     = lipgloss.Color("#ffc107")
	OptionalColor   = lipgloss.Color("#17a2b8")
	SuccessColor    = lipgloss.Color("#28a745")
	MutedColor      = lipgloss.Color("#6c757d")
	HeadingColor    = lipgloss.Color("#ffffff")
	SubtextColor    = lipgloss.Color("#a0a0a0")
	CardBorderColor = lipgloss.Color("#4a4a6a")
)

// SuccessMsg returns a success message.
func SuccessMsg(msg string) string {
	return lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true).
		Render("✓ " + msg)
}

// ErrorMsg returns an error message.
func ErrorMsg(msg string) string {
	return lipgloss.NewStyle().
		Foreground(CriticalColor).
		Bold(true).
		Render("✗ " + msg)
}

// WarningMsg returns a warning message.
func WarningMsg(msg string) string {
	return lipgloss.NewStyle().
		Foreground(HighColor).
		Bold(true).
		Render("⚠ " + msg)
}

// InfoMsg returns an info message.
func InfoMsg(msg string) string {
	return lipgloss.NewStyle().
		Foreground(OptionalColor).
		Render("ℹ " + msg)
}

// PriorityBadge returns a styled priority badge.
func PriorityBadge(priority int) string {
	var style lipgloss.Style

	switch priority {
	case priorityCritical:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(CriticalColor).
			Padding(0, 1)
	case priorityHigh:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(HighColor).
			Padding(0, 1)
	case priorityMedium:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#212529")).
			Background(MediumColor).
			Padding(0, 1)
	default:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(OptionalColor).
			Padding(0, 1)
	}

	name := priorityName(priority)

	return style.Render(name)
}

// Code returns styled code text.
func Code(text string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0")).
		Background(lipgloss.Color("#2d2d44")).
		Padding(0, 1).
		Render(text)
}

// SectionHeader returns a styled section header.
func SectionHeader(title string) string {
	return lipgloss.NewStyle().
		Foreground(PrimaryColor).
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
