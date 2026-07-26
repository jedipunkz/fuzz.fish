package history

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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

	for i := end - 1; i >= start; i-- {
		// Commands may now contain real newlines; show them on one line so a
		// single entry cannot push the rest of the context out of the pane.
		cmd := strings.ReplaceAll(all[i].Cmd, "\n", " ")

		if i == idx {
			cursor := "→ "
			// Wrap active context line
			line := ui.ActiveContextStyle.Width(width).Render(cursor + cmd)
			sb.WriteString(line + "\n")
		} else {
			cursor := "  "
			// Truncate inactive lines to keep context compact. ansi.Truncate
			// measures display cells and cuts on rune boundaries, so narrow
			// panes and multibyte commands stay intact.
			maxWidth := width - lipgloss.Width(cursor)
			if maxWidth > 0 {
				cmd = ansi.Truncate(cmd, maxWidth, "…")
			}
			line := ui.InactiveContextStyle.Render(cursor + cmd)
			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}
