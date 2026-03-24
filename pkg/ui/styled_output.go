package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions
var (
	// Primary brand colors
	PrimaryColor   = lipgloss.Color("#667eea")
	SecondaryColor = lipgloss.Color("#764ba2")

	// Priority colors
	CriticalColor = lipgloss.Color("#dc3545")
	HighColor     = lipgloss.Color("#fd7e14")
	MediumColor   = lipgloss.Color("#ffc107")
	OptionalColor = lipgloss.Color("#17a2b8")
	SuccessColor  = lipgloss.Color("#28a745")
	MutedColor    = lipgloss.Color("#6c757d")

	// Text colors
	TextColor    = lipgloss.Color("#eaeaea")
	HeadingColor = lipgloss.Color("#ffffff")
	SubtextColor = lipgloss.Color("#a0a0a0")

	// Background
	BgColor         = lipgloss.Color("#1a1a2e")
	CardBgColor     = lipgloss.Color("#16213e")
	CardBorderColor = lipgloss.Color("#4a4a6a")
)

// Text styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(HeadingColor).
			Background(PrimaryColor).
			Padding(1, 2).
			Width(60).
			Align(lipgloss.Center)

	HeadingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(HeadingColor).
			Padding(0, 1).
			MarginBottom(1)

	SubheadingStyle = lipgloss.NewStyle().
			Foreground(SubtextColor).
			PaddingLeft(2)

	BodyStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	CodeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0e0e0")).
			Background(lipgloss.Color("#2d2d44")).
			Padding(0, 1)
)

// Priority badge styles
var (
	CriticalBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ffffff")).
				Background(CriticalColor).
				Padding(0, 1).
				MarginRight(1)

	HighBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(HighColor).
			Padding(0, 1).
			MarginRight(1)

	MediumBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#212529")).
				Background(MediumColor).
				Padding(0, 1).
				MarginRight(1)

	OptionalBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ffffff")).
				Background(OptionalColor).
				Padding(0, 1).
				MarginRight(1)
)

// Status styles
var (
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(CriticalColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(HighColor).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(OptionalColor)
)

// Card styles provide styled card output.
var (
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(CardBorderColor).
			Background(CardBgColor).
			Padding(1, 2).
			MarginBottom(1)

	CardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(HeadingColor).
			MarginBottom(1)
)

// Box styles for sections provide styled box output.
var (
	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		MarginBottom(1)
)

// GetPriorityBadge returns the appropriate badge style for a priority level.
func GetPriorityBadge(priority int) lipgloss.Style {
	switch priority {
	case 0:
		return CriticalBadgeStyle
	case 1:
		return HighBadgeStyle
	case 2:
		return MediumBadgeStyle
	default:
		return OptionalBadgeStyle
	}
}

// GetPriorityColor returns the appropriate color for a priority level.
func GetPriorityColor(priority int) lipgloss.Color {
	switch priority {
	case 0:
		return CriticalColor
	case 1:
		return HighColor
	case 2:
		return MediumColor
	default:
		return OptionalColor
	}
}

// GetPriorityName returns the human-readable name for a priority level.
func GetPriorityName(priority int) string {
	switch priority {
	case 0:
		return "CRITICAL"
	case 1:
		return "HIGH"
	case 2:
		return "MEDIUM"
	default:
		return "OPTIONAL"
	}
}

// HorizontalRule creates a visual divider.
func HorizontalRule() string {
	return lipgloss.NewStyle().
		Foreground(CardBorderColor).
		Render("─────────────────────────────────────────────────────")
}

// SectionDivider creates a decorated section divider.
func SectionDivider(title string) string {
	if title == "" {
		return HorizontalRule()
	}
	style := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)
	return style.Render("━━ "+title+" ") + lipgloss.NewStyle().
		Foreground(CardBorderColor).
		Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
