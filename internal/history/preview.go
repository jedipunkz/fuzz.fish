package history

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// formatDir abbreviates a directory path by replacing the home directory with ~
func formatDir(path string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		path = strings.Replace(path, home, "~", 1)
	}
	return path
}

// GeneratePreview generates a preview of the history entry for the TUI preview window
func (e Entry) GeneratePreview(all []Entry, idx, width, height int) string {
	var sb strings.Builder

	// Metadata
	// Time
	sb.WriteString(ui.LabelStyle.Render("Time") + "\n")
	sb.WriteString(ui.ContentStyle.Render(ui.FormatTime(e.When)))
	sb.WriteString("\n")
	sb.WriteString(ui.ContentStyle.Render(ui.FormatRelativeTime(e.When)))
	sb.WriteString("\n\n")

	// Dir
	if len(e.Paths) > 0 {
		sb.WriteString(ui.LabelStyle.Render("Directory") + "\n")
		sb.WriteString(ui.ContentStyle.Render(formatDir(e.Paths[0])))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Context (commands before/after)
	sb.WriteString(ui.ContextHeaderStyle.Render("Context") + "\n")
	start := idx - ui.HistoryContextLinesBefore
	if start < 0 {
		start = 0
	}
	end := idx + ui.HistoryContextLinesAfter
	if end > len(all) {
		end = len(all)
	}

	for i := start; i < end; i++ {
		cmd := all[i].Cmd

		if i == idx {
			cursor := "→ "
			// Wrap active context line
			line := ui.ActiveContextStyle.Width(width).Render(cursor + cmd)
			sb.WriteString(line + "\n")
		} else {
			cursor := "  "
			// Truncate inactive lines to keep context compact
			maxWidth := width - lipgloss.Width(cursor)
			if maxWidth > 0 && len(cmd) > maxWidth {
				cmd = cmd[:maxWidth-3] + "..."
			}
			line := ui.InactiveContextStyle.Render(cursor + cmd)
			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}
