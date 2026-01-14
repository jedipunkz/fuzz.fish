package app

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
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/git"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/history"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/muesli/termenv"
	"github.com/sahilm/fuzzy"
)

type SearchMode int

const (
	ModeHistory SearchMode = iota
	ModeGitBranch
)

// Item represents a search result item
type Item struct {
	Text      string
	Index     int         // Index in the original source slice
	Original  interface{} // The original object (history.Entry or git.Branch)
	IsCurrent bool        // For git branch (icon logic)
	IsRemote  bool        // For git branch (icon logic)
}

type model struct {
	mode     SearchMode
	input    textinput.Model
	viewport viewport.Model

	// Data sources
	historyEntries []history.Entry
	gitBranches    []git.Branch

	// Items state
	allItems []Item // All items for current mode (sorted newest/priority first)
	filtered []Item // Filtered items

	cursor   int
	offset   int
	choice   *string // Result string to print
	quitting bool

	width      int
	height     int
	ready      bool
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

		listWidth := int(float64(msg.Width) * 0.6) // 60% split for list, 40% for preview

		previewWidth := msg.Width - listWidth - 2
		m.listWidth = listWidth

		m.viewport.Width = previewWidth
		m.viewport.Height = mainHeight

		// Recalculate offset to keep cursor at the bottom of the view
		if len(m.filtered) > 0 {
			m.offset = m.cursor - m.mainHeight + 1
			if m.offset < 0 {
				m.offset = 0
			}
		}

		m.validateCursor()
		m.updatePreview()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				m.selectItem()
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlY:
			if len(m.filtered) > 0 {
				_ = clipboard.WriteAll(m.filtered[m.cursor].Text)
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyCtrlR:
			// Toggle mode
			m.toggleMode()
			return m, nil
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
		m.updateFilter(newValue)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) toggleMode() {
	newMode := ModeHistory
	if m.mode == ModeHistory {
		newMode = ModeGitBranch
		// Check if git repo
		if len(m.gitBranches) == 0 {
			// Try to collect branches if not already done?
			// But maybe we should have collected at start if possible.
			// Or check isGitRepo here.
			// Since we want seamless switching, we can try to fetch now.
			if !git.IsGitRepo() {
				// Cannot switch to git mode
				// Maybe show a flash message? For now just ignore or print to debug.
				return
			}
			branches := git.CollectBranches()
			if len(branches) == 0 {
				return
			}
			m.gitBranches = branches
		}
	} else {
		newMode = ModeHistory
	}

	m.mode = newMode
	m.input.SetValue("") // Clear input on switch

	// Update placeholder
	if m.mode == ModeHistory {
		m.input.Placeholder = "Search history... (Ctrl+R to switch)"
	} else {
		m.input.Placeholder = "Search branches... (Ctrl+R to switch)"
	}

	m.loadItemsForMode()
	m.updateFilter("")

	// Reset cursor to bottom explicitly after mode switch
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

func (m *model) loadItemsForMode() {
	m.allItems = []Item{}

	if m.mode == ModeHistory {
		// History: entries are Newest -> Oldest
		// We want Newest at Bottom.
		// Item[0] should be Oldest, Item[N] should be Newest.
		for i := range m.historyEntries {
			e := m.historyEntries[len(m.historyEntries)-1-i]
			m.allItems = append(m.allItems, Item{
				Text:     e.Cmd,
				Index:    len(m.historyEntries) - 1 - i,
				Original: e,
			})
		}
	} else {
		// Git: branches are collected.
		// Sort? CollectBranches usually returns some order.
		// We want Default/Current at bottom?
		// Let's assume input branches are standard.
		// We reverse them to put first item at bottom.
		for i := range m.gitBranches {
			b := m.gitBranches[len(m.gitBranches)-1-i]
			m.allItems = append(m.allItems, Item{
				Text:      b.Name,
				Index:     len(m.gitBranches) - 1 - i,
				Original:  b,
				IsCurrent: b.IsCurrent,
				IsRemote:  b.IsRemote,
			})
		}
	}
}

func (m *model) updateFilter(query string) {
	if query == "" {
		// Return all items (which are already in display order)
		m.filtered = make([]Item, len(m.allItems))
		copy(m.filtered, m.allItems)
	} else {
		// Fuzzy search
		src := make([]string, len(m.allItems))
		// We need search against original list order?
		// m.allItems is already reversed for display.
		// Usually we search against the "source of truth".
		// Let's search against m.allItems text.
		for i, item := range m.allItems {
			src[i] = item.Text
		}

		tokens := strings.Fields(query)
		if len(tokens) > 0 {
			matches := fuzzy.Find(tokens[0], src)

			for _, token := range tokens[1:] {
				if len(matches) == 0 {
					break
				}
				subset := make([]string, len(matches))
				for i, mat := range matches {
					subset[i] = src[mat.Index]
				}
				subMatches := fuzzy.Find(token, subset)
				newMatches := make(fuzzy.Matches, len(subMatches))
				for i, sm := range subMatches {
					newMatches[i] = matches[sm.Index]
				}
				matches = newMatches
			}

			// Sort logic
			sort.SliceStable(matches, func(i, j int) bool {
				scoreI := float64(matches[i].Score)
				scoreJ := float64(matches[j].Score)

				if m.mode == ModeHistory {
					total := float64(len(m.allItems))
					maxBonus := 100.0
					recencyI := float64(matches[i].Index) / total
					recencyJ := float64(matches[j].Index) / total
					scoreI += recencyI * maxBonus
					scoreJ += recencyJ * maxBonus
				}
				return scoreI < scoreJ
			})

			m.filtered = make([]Item, len(matches))
			for i, mat := range matches {
				m.filtered[i] = m.allItems[mat.Index]
			}
		} else {
			// Query is just whitespace, treat as empty
			m.filtered = make([]Item, len(m.allItems))
			copy(m.filtered, m.allItems)
		}
	}

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
	item := m.filtered[m.cursor]

	var content string
	if m.mode == ModeHistory {
		entry := item.Original.(history.Entry)
		// We need original index in historyEntries (which is Newest->Oldest).
		// Item.Index is index in allItems (Oldest->Newest).
		// But GeneratePreview expects index in the slice passed to it?
		// No, GeneratePreview takes (entry, allEntries, index, ...)
		// The index is used for Context.
		// We passed m.historyEntries (Newest->Oldest).
		// We need index in that slice.
		// Item.Index was calculated as len - 1 - originalIndex.
		// So originalIndex = len - 1 - Item.Index.

		origIndex := len(m.historyEntries) - 1 - item.Index
		content = history.GeneratePreview(entry, m.historyEntries, origIndex, m.viewport.Width, m.viewport.Height)
	} else {
		branch := item.Original.(git.Branch)
		content = git.GeneratePreview(branch, m.viewport.Width, m.viewport.Height)
	}
	m.viewport.SetContent(content)
}

func (m *model) selectItem() {
	item := m.filtered[m.cursor]
	if m.mode == ModeHistory {
		res := item.Text
		m.choice = &res
	} else {
		branch := item.Original.(git.Branch)
		res := branch.Name
		if branch.IsRemote {
			parts := strings.SplitN(res, "/", 2)
			if len(parts) == 2 {
				res = parts[1]
			}
		}
		m.choice = &res
	}
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	inputView := m.input.View()

	// List View
	var listBuilder strings.Builder

	// Determine visible range
	start := m.offset
	end := start + m.mainHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	// If items < height, offset is 0, we need to push items to bottom.
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

	// Add Mode Indicator or Help
	// Maybe inside input prompt?

	return lipgloss.JoinVertical(lipgloss.Left,
		mainView,
		inputView,
	)
}

func renderItem(w io.Writer, m model, index int, i Item) {
	width := m.listWidth
	if width <= 0 {
		return
	}

	var cmdStyle lipgloss.Style
	var cursor string

	isSelected := index == m.cursor
	if isSelected {
		cursor = "│"
		cmdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCyan)).Background(lipgloss.Color(ui.ColorSelectionBg)).Bold(true)
	} else {
		cursor = " "
		cmdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorForeground))
	}

	text := i.Text
	// Format text based on mode?
	// History: Replace newlines.
	// Git: Add icons.

	if m.mode == ModeHistory {
		text = strings.ReplaceAll(text, "\n", " ")
	} else {
		var icon string
		if i.IsCurrent {
			icon = "*"
		} else if i.IsRemote {
			icon = "R"
		} else {
			icon = " "
		}
		text = fmt.Sprintf("%s %s", icon, text)
	}

	cursorStr := cursor + " "
	cursorWidth := lipgloss.Width(cursorStr)
	contentWidth := width - cursorWidth
	if contentWidth < 10 {
		contentWidth = 10
	}

	if len(text) > contentWidth {
		text = text[:contentWidth-1] + "…"
	}

	if isSelected {
		fullContent := fmt.Sprintf("%-*s", contentWidth, text)
		renderedCursor := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorPurple)).Background(lipgloss.Color(ui.ColorSelectionBg)).Render(cursorStr)
		renderedContent := cmdStyle.Width(contentWidth).Render(fullContent)
		fmt.Fprint(w, renderedCursor+renderedContent)
	} else {
		renderedCursor := cursorStr
		renderedContent := cmdStyle.Render(fmt.Sprintf("%-*s", contentWidth, text))
		fmt.Fprint(w, renderedCursor+renderedContent)
	}
}

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

	m := model{
		mode:           ModeHistory,
		input:          ti,
		historyEntries: entries,
		viewport:       viewport.New(0, 0),
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
