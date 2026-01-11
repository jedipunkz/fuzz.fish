package ui

import "github.com/charmbracelet/lipgloss"

// Tokyo Night color palette
const (
	ColorCyan        = "#7dcfff"
	ColorPurple      = "#bb9af7"
	ColorForeground  = "#c0caf5"
	ColorYellow      = "#e0af68"
	ColorOrange      = "#ff9e64"
	ColorComment     = "#565f89"
	ColorBlue        = "#7aa2f7"
	ColorSelectionBg = "#543970" // Darker muted purple for selection
)

// Styles for preview window and TUI elements
var (
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorCyan)).
			Bold(true).
			Underline(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPurple))

	ContentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorForeground))

	ContextHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorYellow)).
				Bold(true)

	ActiveContextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorOrange)).
				Bold(true)

	InactiveContextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorComment))
)
