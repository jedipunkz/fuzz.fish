package ui

import "github.com/charmbracelet/lipgloss"

// Tokyo Night color palette
const (
	ColorCyan        = "#7dcfff"
	ColorPurple      = "#bb9af7"
	ColorForeground  = "#c0caf5"
	ColorYellow      = "#e0af68"
	ColorOrange      = "#ff9e64"
	ColorComment     = "#9aa5ce"
	ColorBlue        = "#7aa2f7"
	ColorPink        = "#f7768e" // Tokyo Night pink/magenta for highlights
	ColorSelectionBg = "#414868" // Tokyo Night selection color for better visibility
	ColorBorder      = "#565f89" // Tokyo Night gray for borders
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
