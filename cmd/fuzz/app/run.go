package app

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/history"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/muesli/termenv"
)

// Run starts the application
func Run() {
	// Initial Load: History
	entries := history.Parse()

	ti := textinput.New()
	ti.Placeholder = "Search history... (Ctrl+R to switch)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCyan))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorForeground))

	// Initialize spinner with pink color for history mode
	sp := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorSpinnerHistory))),
	)

	m := model{
		mode:           ModeHistory,
		input:          ti,
		historyEntries: entries,
		viewport:       viewport.New(0, 0),
		spinner:        sp,
	}

	m.loadItemsForMode()
	m.updateFilter("")

	// Ensure cursor is at bottom for initial load
	if len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
		// Offset isn't fully calculable yet because height is unknown until WindowSizeMsg,
		// but we can set cursor index. WindowSizeMsg will fix offset.
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	lipgloss.SetColorProfile(termenv.NewOutput(tty).Profile)

	p := tea.NewProgram(m, tea.WithInput(tty), tea.WithOutput(tty), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(model); ok && m.choice != nil {
		if m.mode == ModeHistory {
			fmt.Printf("CMD:%s", *m.choice)
		} else {
			fmt.Printf("BRANCH:%s", *m.choice)
		}
	}
}
