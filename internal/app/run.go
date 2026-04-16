package app

import (
	"fmt"
	"os"
	"os/exec"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// Run starts the application
func Run() {
	ti := textinput.New()
	ti.Placeholder = "Search history... (Ctrl+G: git, Ctrl+S: files)"
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)
	s := textinput.DefaultDarkStyles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCyan))
	s.Focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorForeground))
	ti.SetStyles(s)

	m := model{
		mode:             ModeHistory,
		input:            ti,
		viewport:         viewport.New(),
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

	prog := tea.NewProgram(m, tea.WithInput(tty), tea.WithOutput(tty))
	finalModel, err := prog.Run()
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
				if m.fetchBranch {
					cmd := exec.Command("git", "pull", "origin", *m.choice)
					cmd.Stdin = tty
					cmd.Stdout = tty
					cmd.Stderr = tty
					if err := cmd.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "git pull failed: %v\n", err)
					}
				} else {
					fmt.Printf("BRANCH:%s", *m.choice)
				}
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
