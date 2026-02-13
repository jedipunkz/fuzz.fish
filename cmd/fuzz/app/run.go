package app

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/muesli/termenv"
)

// Run starts the application
func Run() {
	ti := textinput.New()
	ti.Placeholder = "Search history... (Ctrl+G: git, Ctrl+S: files)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCyan))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorForeground))

	m := model{
		mode:             ModeHistory,
		input:            ti,
		viewport:         viewport.New(0, 0),
		previewCache:     make(map[string]string),
		lastPreviewIndex: -1,
		loading:          true,
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = tty.Close() }()

	lipgloss.SetColorProfile(termenv.NewOutput(tty).Profile)

	p := tea.NewProgram(m, tea.WithInput(tty), tea.WithOutput(tty), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(model); ok {
		if m.choice != nil {
			switch m.mode {
			case ModeHistory:
				fmt.Printf("CMD:%s", *m.choice)
			case ModeGitBranch:
				fmt.Printf("BRANCH:%s", *m.choice)
			case ModeFiles:
				if m.choiceIsDir {
					fmt.Printf("DIR:%s", *m.choice)
				} else {
					fmt.Printf("FILE:%s", *m.choice)
				}
			}
		}
	}
}
