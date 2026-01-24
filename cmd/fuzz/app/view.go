package app

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/history"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
)

// View renders the application view
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

	// Border style with rounded corners and gray color
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColorBorder))

	// List pane with border
	listBox := boxStyle.
		Width(m.listWidth).
		Height(m.mainHeight).
		Render(listView)

	// Preview pane with border
	previewBox := boxStyle.
		Width(m.viewport.Width).
		Height(m.mainHeight).
		Render(previewView)

	// Input box with border
	inputBox := boxStyle.
		Width(m.width-2).
		Padding(0, 1).
		Render(inputView)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		listBox,
		previewBox,
	)

	// Add Mode Indicator or Help
	// Maybe inside input prompt?

	return lipgloss.JoinVertical(lipgloss.Left,
		mainView,
		inputBox,
	)
}

// renderItem renders a single item in the list
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
	// Files: Add icons.

	// Calculate time ago string for history mode
	var timeAgo string
	if m.mode == ModeHistory {
		text = strings.ReplaceAll(text, "\n", " ")
		if entry, ok := i.Original.(history.Entry); ok && entry.When > 0 {
			timeAgo = formatTimeAgo(entry.When)
		}
	} else if m.mode == ModeGitBranch {
		var icon string
		if i.IsCurrent {
			icon = "*"
		} else if i.IsRemote {
			icon = "R"
		} else {
			icon = " "
		}
		text = fmt.Sprintf("%s %s", icon, text)
	} else if m.mode == ModeFiles {
		var icon string
		if i.IsDir {
			icon = "📁"
		} else {
			icon = "📄"
		}
		text = fmt.Sprintf("%s %s", icon, text)
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

	renderedCursor := cursorStr
	if isSelected {
		renderedCursor = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorPurple)).Background(lipgloss.Color(ui.ColorSelectionBg)).Render(cursorStr)
	}

	// Render text with match highlighting
	var textBuilder strings.Builder
	matchSet := make(map[int]bool)
	for _, idx := range i.MatchedIndexes {
		matchSet[idx] = true
	}

	runes := []rune(text)
	for runeIdx := 0; runeIdx < len(runes); runeIdx++ {
		var charStyle lipgloss.Style
		if matchSet[runeIdx] {
			// Matched character
			if isSelected {
				charStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ui.ColorPink)).
					Background(lipgloss.Color(ui.ColorSelectionBg)).
					Bold(true)
			} else {
				charStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(ui.ColorPink)).
					Bold(true)
			}
		} else {
			// Non-matched character
			charStyle = cmdStyle
		}
		textBuilder.WriteString(charStyle.Render(string(runes[runeIdx])))
	}

	rendered := textBuilder.String()
	renderedWidth := lipgloss.Width(rendered)
	padding := contentWidth - renderedWidth
	if padding > 0 {
		paddingStyle := lipgloss.NewStyle()
		if isSelected {
			paddingStyle = paddingStyle.Background(lipgloss.Color(ui.ColorSelectionBg))
		}
		rendered += paddingStyle.Render(strings.Repeat(" ", padding))
	}

	// Render time ago
	var timeAgoRendered string
	if timeAgo != "" {
		timeAgoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorTimeAgo))
		if isSelected {
			timeAgoStyle = timeAgoStyle.Background(lipgloss.Color(ui.ColorSelectionBg))
		}
		timeAgoRendered = " " + timeAgoStyle.Render(timeAgo)
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
