package app

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"github.com/jedipunkz/fuzz.fish/internal/history"
	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// Pre-computed styles to avoid per-render allocation (lipgloss.NewStyle is expensive)
var (
	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColorBorder))

	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	// Item list styles
	itemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorCyan)).
				Background(lipgloss.Color(ui.ColorSelectionBg)).
				Bold(true)

	itemNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorForeground))

	cursorSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorPurple)).
				Background(lipgloss.Color(ui.ColorSelectionBg))

	// Match highlight styles
	matchSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorPink)).
				Background(lipgloss.Color(ui.ColorSelectionBg)).
				Bold(true)

	matchNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorPink)).
				Bold(true)

	// Padding styles
	paddingSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ui.ColorSelectionBg))

	paddingNormalStyle = lipgloss.NewStyle()

	// Time ago styles
	timeAgoSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorTimeAgo)).
				Background(lipgloss.Color(ui.ColorSelectionBg))

	timeAgoNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ui.ColorTimeAgo))
)

// View renders the application view
func (m model) View() tea.View {
	if !m.ready {
		v := tea.NewView("Initializing...")
		v.AltScreen = true
		return v
	}

	inputView := m.input.View()

	// List View
	var listBuilder strings.Builder

	// Show loading indicator when async loading with no items yet
	if m.loading && len(m.filtered) == 0 {
		if m.mainHeight > 1 {
			listBuilder.WriteString(strings.Repeat("\n", m.mainHeight-1))
		}
		listBuilder.WriteString("Loading...")

		listView := listBuilder.String()
		previewView := m.viewport.View()

		// In lipgloss v2, Width/Height include borders, so add 2 for left+right / top+bottom borders
		listBox := boxStyle.Width(m.listWidth + 2).Height(m.mainHeight + 2).Render(listView)
		previewBox := boxStyle.Width(m.viewport.Width() + 2).Height(m.mainHeight + 2).Render(previewView)
		inputBox := boxStyle.Width(m.width).Padding(0, 1).Render(inputView)

		mainView := lipgloss.JoinHorizontal(lipgloss.Top, listBox, previewBox)
		v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, mainView, inputBox))
		v.AltScreen = true
		v.Cursor = m.inputCursor()
		return v
	}

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
		m.renderItem(&listBuilder, i, item)
		if i < end-1 {
			listBuilder.WriteString("\n")
		}
	}

	listView := listBuilder.String()
	previewView := m.viewport.View()

	// List pane with border
	// In lipgloss v2, Width/Height include borders, so add 2 for left+right / top+bottom borders
	listBox := boxStyle.
		Width(m.listWidth + 2).
		Height(m.mainHeight + 2).
		Render(listView)

	// Preview pane with border
	previewBox := boxStyle.
		Width(m.viewport.Width() + 2).
		Height(m.mainHeight + 2).
		Render(previewView)

	// Build input line with optional status message
	inputContent := inputView
	if m.statusMsg != "" {
		inputContent = inputView + "  " + warningStyle.Render(m.statusMsg)
	}

	// Input box with border
	inputBox := boxStyle.
		Width(m.width).
		Padding(0, 1).
		Render(inputContent)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		listBox,
		previewBox,
	)

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		mainView,
		inputBox,
	))
	v.AltScreen = true
	v.Cursor = m.inputCursor()
	return v
}

// inputCursor returns the textinput cursor offset to absolute screen coordinates.
func (m model) inputCursor() *tea.Cursor {
	c := m.input.Cursor()
	if c == nil {
		return nil
	}
	// X: left border (1) + left padding (1)
	c.Position.X += 2
	// Y: main panes height (mainHeight + 2 borders) + input box top border (1)
	c.Position.Y += m.mainHeight + 2 + 1
	return c
}

// renderItem renders a single item in the list
func (m model) renderItem(w io.Writer, index int, i Item) {
	width := m.listWidth
	if width <= 0 {
		return
	}

	isSelected := index == m.cursor

	var cmdStyle lipgloss.Style
	var cursor string
	if isSelected {
		cursor = "│"
		cmdStyle = itemSelectedStyle
	} else {
		cursor = " "
		cmdStyle = itemNormalStyle
	}

	text := i.Text

	// Calculate time ago string for history mode
	var timeAgo string
	switch m.mode {
	case ModeHistory:
		text = strings.ReplaceAll(text, "\n", " ")
		if entry, ok := i.Original.(history.Entry); ok && entry.When > 0 {
			timeAgo = formatTimeAgo(entry.When)
		}
	case ModeGitBranch:
		var icon string
		if i.IsCurrent {
			icon = "*"
		} else if i.IsRemote {
			icon = "R"
		} else {
			icon = " "
		}
		text = icon + " " + text
	case ModeFiles:
		var icon string
		if i.IsDir {
			icon = "📁"
		} else {
			icon = "📄"
		}
		text = icon + " " + text
	}

	cursorStr := cursor + " "
	cursorWidth := lipgloss.Width(cursorStr)

	// Reserve space for time ago display
	timeAgoWidth := 0
	if timeAgo != "" {
		timeAgoWidth = len(timeAgo) + 1 // +1 for spacing
	}

	contentWidth := width - cursorWidth - timeAgoWidth
	if contentWidth < 10 {
		contentWidth = 10
	}

	if len(text) > contentWidth {
		text = text[:contentWidth-1] + "…"
	}

	var renderedCursor string
	if isSelected {
		renderedCursor = cursorSelectedStyle.Render(cursorStr)
	} else {
		renderedCursor = cursorStr
	}

	// Build match set as bool slice for O(1) lookup without map overhead
	runes := []rune(text)
	var matchBits []bool
	if len(i.MatchedIndexes) > 0 {
		// Find max index to size the bool slice
		maxIdx := 0
		for _, idx := range i.MatchedIndexes {
			if idx > maxIdx {
				maxIdx = idx
			}
		}
		if maxIdx < len(runes) {
			matchBits = make([]bool, maxIdx+1)
			for _, idx := range i.MatchedIndexes {
				if idx < len(matchBits) {
					matchBits[idx] = true
				}
			}
		}
	}

	// Render text with match highlighting
	var textBuilder strings.Builder
	textBuilder.Grow(len(text) * 20) // estimate: each char may get ANSI escape codes
	for runeIdx, r := range runes {
		var charStyle lipgloss.Style
		isMatch := runeIdx < len(matchBits) && matchBits[runeIdx]
		if isMatch {
			if isSelected {
				charStyle = matchSelectedStyle
			} else {
				charStyle = matchNormalStyle
			}
		} else {
			charStyle = cmdStyle
		}
		textBuilder.WriteString(charStyle.Render(string(r)))
	}

	rendered := textBuilder.String()
	renderedWidth := lipgloss.Width(rendered)
	padWidth := contentWidth - renderedWidth
	if padWidth > 0 {
		if isSelected {
			rendered += paddingSelectedStyle.Render(strings.Repeat(" ", padWidth))
		} else {
			rendered += paddingNormalStyle.Render(strings.Repeat(" ", padWidth))
		}
	}

	// Render time ago
	var timeAgoRendered string
	if timeAgo != "" {
		if isSelected {
			timeAgoRendered = " " + timeAgoSelectedStyle.Render(timeAgo)
		} else {
			timeAgoRendered = " " + timeAgoNormalStyle.Render(timeAgo)
		}
	}

	_, _ = fmt.Fprint(w, renderedCursor+rendered+timeAgoRendered)
}

// formatTimeAgo formats a Unix timestamp as a relative time string
func formatTimeAgo(when int64) string {
	t := time.Unix(when, 0)
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
