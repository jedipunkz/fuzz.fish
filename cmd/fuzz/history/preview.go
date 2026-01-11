package history

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/utils"
)

// GeneratePreview generates a preview of the history entry for the TUI preview window
func GeneratePreview(entry Entry, all []Entry, idx, width, height int) string {
	var sb strings.Builder

	// Header
	// sb.WriteString(ui.HeaderStyle.Render("Command") + "\n")
	// Wrap command to fit width
	// sb.WriteString(ui.ContentStyle.Copy().Width(width).Render(entry.Cmd))
	// sb.WriteString("\n\n")

	// Metadata
	// Time
	sb.WriteString(ui.LabelStyle.Render("Time") + "\n")
	sb.WriteString(ui.ContentStyle.Render(utils.FormatTime(entry.When)))
	sb.WriteString("\n")
	sb.WriteString(ui.ContentStyle.Faint(true).Render(utils.FormatRelativeTime(entry.When)))
	sb.WriteString("\n\n")

	// Dir
	if len(entry.Paths) > 0 {
		sb.WriteString(ui.LabelStyle.Render("Directory") + "\n")
		sb.WriteString(ui.ContentStyle.Render(FormatDir(entry.Paths[0])))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Context (commands before/after)
	sb.WriteString(ui.ContextHeaderStyle.Render("Context") + "\n")
	start := idx - 3
	if start < 0 {
		start = 0
	}
	end := idx + 4
	if end > len(all) {
		end = len(all)
	}

	for i := start; i < end; i++ {
		e := all[i]
		cmd := e.Cmd

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
