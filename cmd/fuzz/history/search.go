package history

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/muesli/termenv"
	"github.com/sahilm/fuzzy"
)

// item implements list.Item
type item struct {
	entry Entry
	index int // Index in the original entries slice
}

func (i item) Title() string       { return i.entry.Cmd }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.entry.Cmd }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	// Calculate available width
	width := m.Width()
	if width <= 0 {
		return
	}

	// Define styles
	var (
		cmdStyle lipgloss.Style
		cursor   string
	)

	isSelected := index == m.Index()

	if isSelected {
		cursor = "│" // Active cursor
		// Selected style: Bright foreground, distinct background
		cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorCyan)).
			Background(lipgloss.Color(ui.ColorSelectionBg)).
			Bold(true)
	} else {
		cursor = " "
		// Normal style
		cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorForeground))
	}

	// Prepare content
	cmd := i.entry.Cmd
	cmd = strings.ReplaceAll(cmd, "\n", " ")

	// Layout calculation
	// Cursor (1 char + 1 space padding) + Content

	cursorStr := cursor + " "
	cursorWidth := lipgloss.Width(cursorStr)

	contentWidth := width - cursorWidth
	if contentWidth < 10 {
		contentWidth = 10
	}

	cmdWidth := contentWidth

	// Truncate command if needed
	if len(cmd) > cmdWidth {
		cmd = cmd[:cmdWidth-1] + "…"
	}

	// Render
	var line string

	if isSelected {
		// For selected item, we want the background to span the entire content width
		// Construct the full string first
		fullContent := fmt.Sprintf("%-*s", cmdWidth, cmd)

		// Render cursor with selection background for continuity, or keep distinct?
		// Let's keep cursor distinct foreground, but same background
		renderedCursor := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorPurple)).
			Background(lipgloss.Color(ui.ColorSelectionBg)).
			Render(cursorStr)

		renderedContent := cmdStyle.Width(contentWidth).Render(fullContent)
		line = renderedCursor + renderedContent
	} else {
		// For normal items
		renderedCursor := cursorStr

		// Render parts separately to apply different text colors
		renderedCmd := cmdStyle.Render(fmt.Sprintf("%-*s", cmdWidth, cmd))

		line = renderedCursor + renderedCmd
	}

	fmt.Fprint(w, line)
}

type model struct {
	list     list.Model
	input    textinput.Model
	viewport viewport.Model
	entries  []Entry
	choice   *Entry
	quitting bool
	width    int
	height   int
	ready    bool
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

		// Layout: Input (3 lines) + Main Area
		// Main Area: List (60%) | Preview (40%)

		inputHeight := 3
		mainHeight := msg.Height - inputHeight

		listWidth := int(float64(msg.Width) * 0.6)
		previewWidth := msg.Width - listWidth - 2 // -2 for border/gap

		m.list.SetSize(listWidth, mainHeight)
		m.viewport.Width = previewWidth
		m.viewport.Height = mainHeight

		// Refresh preview on resize
		m.updatePreview()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if i, ok := m.list.SelectedItem().(item); ok {
				m.choice = &i.entry
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		}
		// Custom navigation
		switch msg.String() {
		case "down", "ctrl+n":
			m.list.CursorDown()
			m.updatePreview()
			return m, nil
		case "up", "ctrl+p":
			m.list.CursorUp()
			m.updatePreview()
			return m, nil
		}
	}

	oldValue := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	newValue := m.input.Value()
	if oldValue != newValue {
		m.list.SetItems(filterEntries(m.entries, newValue))
		m.list.ResetSelected()
		m.updatePreview()
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) updatePreview() {
	sel := m.list.SelectedItem()
	if sel == nil {
		m.viewport.SetContent("")
		return
	}
	i, ok := sel.(item)
	if !ok {
		return
	}

	content := GeneratePreview(i.entry, m.entries, i.index, m.viewport.Width, m.viewport.Height)
	m.viewport.SetContent(content)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Render input
	inputView := m.input.View()

	// Render list and preview side-by-side
	listView := m.list.View()
	previewView := m.viewport.View()

	// Add a border or separator to preview if desired, or just space
	// Currently just side by side
	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		listView,
		"  ", // Gap
		previewView,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		inputView,
		mainView,
	)
}

func filterEntries(entries []Entry, query string) []list.Item {
	tokens := strings.Fields(query)
	if len(tokens) == 0 {
		items := make([]list.Item, len(entries))
		for i, e := range entries {
			items[i] = item{entry: e, index: i}
		}
		return items
	}

	src := make([]string, len(entries))
	for i, e := range entries {
		src[i] = e.Cmd
	}

	matches := fuzzy.Find(tokens[0], src)

	for _, token := range tokens[1:] {
		if len(matches) == 0 {
			break
		}

		subset := make([]string, len(matches))
		for i, m := range matches {
			subset[i] = src[m.Index]
		}

		subMatches := fuzzy.Find(token, subset)

		newMatches := make(fuzzy.Matches, len(subMatches))
		for i, sm := range subMatches {
			newMatches[i] = matches[sm.Index]
		}
		matches = newMatches
	}

	// Calculate custom score: Fuzzy Score + Time Bonus
	// fuzzy score is usually 0 to len(str) or so.
	// We want to give a bonus for recent items.
	// Let's normalize time.
	// We can use the index in 'entries' as a proxy for recency if 'entries' is sorted by time (newest first).
	// Parser() sorts entries newest first. So index 0 is the newest.
	//
	// Score calculation:
	// Base Score = matches[i].Score
	// Time Bonus = (Total Entries - Index) / Total Entries * Max Bonus
	// Let's say Max Bonus is equivalent to matching 2-3 characters?

	// Since fuzzy.Match.Score is unexported or hard to modify directly in the struct locally without copying,
	// we will use a custom sort.

	totalEntries := float64(len(entries))
	maxBonus := 100.0 // Increased weight to prioritize recency significantly

	sort.SliceStable(matches, func(i, j int) bool {
		// Calculate score for i
		idxI := matches[i].Index
		recencyI := (totalEntries - float64(idxI)) / totalEntries
		scoreI := float64(matches[i].Score) + (recencyI * maxBonus)

		// Calculate score for j
		idxJ := matches[j].Index
		recencyJ := (totalEntries - float64(idxJ)) / totalEntries
		scoreJ := float64(matches[j].Score) + (recencyJ * maxBonus)

		// Descending sort
		return scoreI > scoreJ
	})

	items := make([]list.Item, len(matches))
	for i, m := range matches {
		items[i] = item{entry: entries[m.Index], index: m.Index}
	}
	return items
}

// RunSearch runs the interactive history search
func RunSearch() {
	entries := Parse()
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No history found")
		os.Exit(1)
	}

	items := make([]list.Item, len(entries))
	for i, e := range entries {
		items[i] = item{entry: e, index: i}
	}

	l := list.New(items, itemDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "Search history..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCyan))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorForeground))

	m := model{
		list:     l,
		input:    ti,
		entries:  entries,
		viewport: viewport.New(0, 0),
	}

	// Open /dev/tty for interactive TUI
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	// Force lipgloss to use the TTY's color profile, as os.Stdout might be a pipe
	lipgloss.SetColorProfile(termenv.NewOutput(tty).Profile)

	p := tea.NewProgram(m, tea.WithInput(tty), tea.WithOutput(tty), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(model); ok && m.choice != nil {
		fmt.Print(m.choice.Cmd)
	}
}
