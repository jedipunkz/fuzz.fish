package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/muesli/termenv"
	"github.com/sahilm/fuzzy"
)

// item represents a search result item
type item struct {
	branch Branch
	index  int
}

type model struct {
	input    textinput.Model
	viewport viewport.Model
	branches []Branch
	filtered []item // Filtered items
	cursor   int    // Index of selected item in 'filtered'
	offset   int    // Scroll offset
	choice   *Branch
	quitting bool
	width    int
	height   int
	ready    bool

	// Layout
	listWidth  int
	mainHeight int
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		inputHeight := 3
		mainHeight := msg.Height - inputHeight
		if mainHeight < 0 {
			mainHeight = 0
		}
		m.mainHeight = mainHeight

		listWidth := int(float64(msg.Width) * 0.4) // Branch names shorter
		previewWidth := msg.Width - listWidth - 2
		m.listWidth = listWidth

		m.viewport.Width = previewWidth
		m.viewport.Height = mainHeight

		m.validateCursor()
		m.updatePreview()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				m.choice = &m.filtered[m.cursor].branch
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlY:
			if len(m.filtered) > 0 {
				_ = clipboard.WriteAll(m.filtered[m.cursor].branch.Name)
				m.quitting = true
				return m, tea.Quit
			}
		}
		switch msg.String() {
		case "down", "ctrl+n":
			if len(m.filtered) > 0 {
				m.cursor++
				if m.cursor >= len(m.filtered) {
					m.cursor = len(m.filtered) - 1
				}
				if m.cursor >= m.offset+m.mainHeight {
					m.offset = m.cursor - m.mainHeight + 1
				}
				m.updatePreview()
			}
			return m, nil
		case "up", "ctrl+p":
			if len(m.filtered) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = 0
				}
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
				m.updatePreview()
			}
			return m, nil
		}
	}

	oldValue := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	newValue := m.input.Value()
	if oldValue != newValue {
		m.filtered = filterBranches(m.branches, newValue)

		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
			m.offset = m.cursor - m.mainHeight + 1
			if m.offset < 0 {
				m.offset = 0
			}
		} else {
			m.cursor = 0
			m.offset = 0
		}
		m.updatePreview()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) validateCursor() {
	if len(m.filtered) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.offset < 0 {
		m.offset = 0
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.mainHeight {
		m.offset = m.cursor - m.mainHeight + 1
		if m.offset < 0 {
			m.offset = 0
		}
	}
}

func (m *model) updatePreview() {
	if len(m.filtered) == 0 {
		m.viewport.SetContent("")
		return
	}
	i := m.filtered[m.cursor]
	content := GeneratePreview(i.branch, m.viewport.Width, m.viewport.Height)
	m.viewport.SetContent(content)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	inputView := m.input.View()

	var listBuilder strings.Builder

	start := m.offset
	end := start + m.mainHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	visibleCount := end - start
	padding := m.mainHeight - visibleCount
	if padding > 0 {
		listBuilder.WriteString(strings.Repeat("\n", padding))
	}

	for i := start; i < end; i++ {
		item := m.filtered[i]
		renderItem(&listBuilder, m, i, item)
		if i < end-1 {
			listBuilder.WriteString("\n")
		}
	}

	listView := listBuilder.String()
	previewView := m.viewport.View()

	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(m.listWidth).Height(m.mainHeight).Render(listView),
		"  ",
		previewView,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		mainView,
		inputView,
	)
}

func renderItem(w io.Writer, m model, index int, i item) {
	width := m.listWidth
	if width <= 0 {
		return
	}

	var (
		cmdStyle lipgloss.Style
		cursor   string
	)

	isSelected := index == m.cursor

	if isSelected {
		cursor = "│"
		cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorCyan)).
			Background(lipgloss.Color(ui.ColorSelectionBg)).
			Bold(true)
	} else {
		cursor = " "
		cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorForeground))
	}

	name := i.branch.Name
	var icon string
	if i.branch.IsCurrent {
		icon = "*"
	} else if i.branch.IsRemote {
		icon = "R"
	} else {
		icon = " "
	}

	displayText := fmt.Sprintf("%s %s", icon, name)

	cursorStr := cursor + " "
	cursorWidth := lipgloss.Width(cursorStr)
	contentWidth := width - cursorWidth
	if contentWidth < 10 {
		contentWidth = 10
	}

	if len(displayText) > contentWidth {
		displayText = displayText[:contentWidth-1] + "…"
	}

	var line string
	if isSelected {
		fullContent := fmt.Sprintf("%-*s", contentWidth, displayText)
		renderedCursor := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorPurple)).
			Background(lipgloss.Color(ui.ColorSelectionBg)).
			Render(cursorStr)
		renderedContent := cmdStyle.Width(contentWidth).Render(fullContent)
		line = renderedCursor + renderedContent
	} else {
		renderedCursor := cursorStr
		renderedContent := cmdStyle.Render(fmt.Sprintf("%-*s", contentWidth, displayText))
		line = renderedCursor + renderedContent
	}

	fmt.Fprint(w, line)
}

func filterBranches(branches []Branch, query string) []item {
	if query == "" {
		items := make([]item, len(branches))
		for i := range branches {
			// Reverse: Index 0 is bottom (current branch typically)
			items[i] = item{branch: branches[len(branches)-1-i], index: len(branches) - 1 - i}
		}
		return items
	}

	src := make([]string, len(branches))
	for i, b := range branches {
		src[i] = b.Name
	}

	matches := fuzzy.Find(query, src)

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].Score < matches[j].Score // Low score first -> Best match last
	})

	items := make([]item, len(matches))
	for i, m := range matches {
		items[i] = item{branch: branches[m.Index], index: m.Index}
	}
	return items
}

// RunBranchSearch runs the interactive git branch search
func RunBranchSearch() {
	if !IsGitRepo() {
		fmt.Fprintln(os.Stderr, "Not a git repository")
		os.Exit(1)
	}

	branches := CollectBranches()
	if len(branches) == 0 {
		fmt.Fprintln(os.Stderr, "No branches found")
		os.Exit(1)
	}

	items := make([]item, len(branches))
	for i := range branches {
		items[i] = item{branch: branches[len(branches)-1-i], index: len(branches) - 1 - i}
	}

	ti := textinput.New()
	ti.Placeholder = "Search branches..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCyan))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorForeground))

	m := model{
		input:    ti,
		branches: branches,
		filtered: items,
		cursor:   len(items) - 1,
		viewport: viewport.New(0, 0),
	}

	// Open /dev/tty for interactive TUI
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
		branchName := m.choice.Name
		if m.choice.IsRemote {
			parts := strings.SplitN(branchName, "/", 2)
			if len(parts) == 2 {
				branchName = parts[1]
			}
		}
		fmt.Print(branchName)
	}
}

// IsGitRepo checks if the current directory is inside a git repository
func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}
