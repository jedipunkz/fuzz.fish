package history

import (
	"fmt"
	"io"
	"os"
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
	entry          Entry
	index          int   // Index in the original entries slice
	matchedIndexes []int // Indexes of matched characters for highlighting
}

type model struct {
	input    textinput.Model
	viewport viewport.Model
	entries  []Entry
	filtered []item // Filtered items
	cursor   int    // Index of selected item in 'filtered'
	offset   int    // Scroll offset (index of the top item being displayed in the list view)
	choice   *Entry
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

		// Layout: Input (3 lines) + Main Area
		inputHeight := 3
		mainHeight := msg.Height - inputHeight
		if mainHeight < 0 {
			mainHeight = 0
		}
		m.mainHeight = mainHeight

		listWidth := int(float64(msg.Width) * 0.6)
		previewWidth := msg.Width - listWidth - 2 // -2 for border/gap
		m.listWidth = listWidth

		m.viewport.Width = previewWidth
		m.viewport.Height = mainHeight

		// Ensure cursor and offset are valid
		m.validateCursor()
		m.updatePreview()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				m.choice = &m.filtered[m.cursor].entry
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlY:
			if len(m.filtered) > 0 {
				_ = clipboard.WriteAll(m.filtered[m.cursor].entry.Cmd)
				m.quitting = true
				return m, tea.Quit
			}
		}

		// Custom navigation
		switch msg.String() {
		case "down", "ctrl+n":
			if len(m.filtered) > 0 {
				m.cursor++
				if m.cursor >= len(m.filtered) {
					m.cursor = len(m.filtered) - 1
				}
				// Scroll logic: if cursor moves out of view (bottom), move offset
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
				// Scroll logic: if cursor moves out of view (top), move offset
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
		m.filtered = filterEntries(m.entries, newValue)

		// Reset state on filter change
		// We want to select the last item (Newest/Best Match)
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
			// Set offset so that the cursor is visible (ideally at bottom)
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
	// Adjust offset if needed
	if m.offset < 0 {
		m.offset = 0
	}
	// Ensure cursor is visible?
	// Not strictly necessary here unless window resized drastically, but good practice.
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
	content := GeneratePreview(i.entry, m.entries, i.index, m.viewport.Width, m.viewport.Height)
	m.viewport.SetContent(content)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Render input
	inputView := m.input.View()

	// Render list manually
	var listBuilder strings.Builder

	// Determine visible range
	start := m.offset
	end := start + m.mainHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	// Padding calculation for "Bottom Alignment" behavior
	// If we have fewer items than mainHeight, we render them, but we want them at the bottom.
	// Actually, the visible range logic above (offset=0 if items < height) handles content.
	// We just need to prepend newlines if necessary.

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

	// Ensure list view matches height exactly
	// (strings.Repeat logic above ensures we have at least mainHeight lines if padding > 0)
	// But lipgloss might help ensure exact box size.
	// Actually, if we just output text, it should be fine.

	// Render Preview
	previewView := m.viewport.View()

	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(m.listWidth).Height(m.mainHeight).Render(listView),
		"  ", // Gap
		previewView,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		mainView,
		inputView,
	)
}

func renderItem(w io.Writer, m model, index int, i item) {
	// Calculate available width
	width := m.listWidth
	if width <= 0 {
		return
	}

	// Define styles
	var (
		cmdStyle lipgloss.Style
		cursor   string
	)

	isSelected := index == m.cursor

	if isSelected {
		cursor = "│" // Active cursor
		cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorCyan)).
			Background(lipgloss.Color(ui.ColorSelectionBg)).
			Bold(true)
	} else {
		cursor = " "
		cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorForeground))
	}

	// Prepare content
	cmd := i.entry.Cmd
	cmd = strings.ReplaceAll(cmd, "\n", " ")

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
		renderedCursor := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorPurple)).
			Background(lipgloss.Color(ui.ColorSelectionBg)).
			Render(cursorStr)

		// Render command with match highlighting
		var cmdBuilder strings.Builder
		matchSet := make(map[int]bool)
		for _, idx := range i.matchedIndexes {
			matchSet[idx] = true
		}

		// Convert cmd string to runes for proper indexing
		runes := []rune(cmd)
		for runeIdx := 0; runeIdx < len(runes); runeIdx++ {
			var charStyle lipgloss.Style
			if matchSet[runeIdx] {
				// Matched character: Orange color
				charStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ui.ColorOrange)).
					Background(lipgloss.Color(ui.ColorSelectionBg)).
					Bold(true)
			} else {
				// Non-matched character: Cyan color
				charStyle = cmdStyle
			}
			cmdBuilder.WriteString(charStyle.Render(string(runes[runeIdx])))
		}

		// Pad to full width
		rendered := cmdBuilder.String()
		renderedWidth := lipgloss.Width(rendered)
		padding := contentWidth - renderedWidth
		if padding > 0 {
			paddingStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(ui.ColorSelectionBg))
			rendered += paddingStyle.Render(strings.Repeat(" ", padding))
		}

		line = renderedCursor + rendered
	} else {
		renderedCursor := cursorStr

		// Render command with match highlighting for non-selected items
		var cmdBuilder strings.Builder
		matchSet := make(map[int]bool)
		for _, idx := range i.matchedIndexes {
			matchSet[idx] = true
		}

		// Convert cmd string to runes for proper indexing
		runes := []rune(cmd)
		for runeIdx := 0; runeIdx < len(runes); runeIdx++ {
			var charStyle lipgloss.Style
			if matchSet[runeIdx] {
				// Matched character: Yellow color
				charStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ui.ColorYellow)).
					Bold(true)
			} else {
				// Non-matched character: Default color
				charStyle = cmdStyle
			}
			cmdBuilder.WriteString(charStyle.Render(string(runes[runeIdx])))
		}

		line = renderedCursor + cmdBuilder.String()
	}

	fmt.Fprint(w, line)
}

func filterEntries(entries []Entry, query string) []item {
	tokens := strings.Fields(query)
	var matches fuzzy.Matches

	// Generate source slice for fuzzy matching
	src := make([]string, len(entries))
	for i, e := range entries {
		src[i] = e.Cmd
	}

	if len(tokens) == 0 {
		// No query: Return all items
		// We want Newest at Bottom. 'entries' is Newest->Oldest.
		// So we reverse it: Oldest (index 0) ... Newest (index N)
		items := make([]item, len(entries))
		for i := range entries {
			// index in 'filtered' is i
			// index in 'entries' is len-1-i
			items[i] = item{entry: entries[len(entries)-1-i], index: len(entries) - 1 - i, matchedIndexes: nil}
		}
		return items
	}

	matches = fuzzy.Find(tokens[0], src)

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
			// Merge matched indexes from both matches
			origMatch := matches[sm.Index]
			mergedIndexes := make([]int, 0, len(origMatch.MatchedIndexes)+len(sm.MatchedIndexes))
			mergedIndexes = append(mergedIndexes, origMatch.MatchedIndexes...)
			mergedIndexes = append(mergedIndexes, sm.MatchedIndexes...)

			newMatches[i] = fuzzy.Match{
				Str:            origMatch.Str,
				Index:          origMatch.Index,
				MatchedIndexes: mergedIndexes,
				Score:          origMatch.Score + sm.Score,
			}
		}
		matches = newMatches
	}

	totalEntries := float64(len(entries))
	maxBonus := 100.0

	sort.SliceStable(matches, func(i, j int) bool {
		idxI := matches[i].Index
		recencyI := (totalEntries - float64(idxI)) / totalEntries
		scoreI := float64(matches[i].Score) + (recencyI * maxBonus)

		idxJ := matches[j].Index
		recencyJ := (totalEntries - float64(idxJ)) / totalEntries
		scoreJ := float64(matches[j].Score) + (recencyJ * maxBonus)

		// Ascending sort (Lower score first) -> Best match (highest score) last
		return scoreI < scoreJ
	})

	items := make([]item, len(matches))
	for i, m := range matches {
		items[i] = item{entry: entries[m.Index], index: m.Index, matchedIndexes: m.MatchedIndexes}
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

	// Initial items: entries is Newest->Oldest
	// We want to show Newest at Bottom.
	items := make([]item, len(entries))
	for i := range entries {
		items[i] = item{entry: entries[len(entries)-1-i], index: len(entries) - 1 - i}
	}

	ti := textinput.New()
	ti.Placeholder = "Search history..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCyan))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorForeground))

	m := model{
		input:    ti,
		entries:  entries,
		filtered: items,
		cursor:   len(items) - 1, // Start at bottom (Newest)
		viewport: viewport.New(0, 0),
	}
	// Initial offset logic will be handled in first Update/Resize or we set it here approximately
	// But mainHeight is unknown until WindowSizeMsg.

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
