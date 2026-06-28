package git

import (
	"strings"

	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// GeneratePreview generates a lightweight preview of the branch
func (b Branch) GeneratePreview(width, height int) string {
	var sb strings.Builder

	// Branch info
	sb.WriteString(ui.LabelStyle.Render("Branch") + "\n")
	sb.WriteString(ui.ContentStyle.Render(b.Name) + "\n\n")

	// Commit hash
	sb.WriteString(ui.LabelStyle.Render("Commit") + "\n")
	sb.WriteString(ui.ContentStyle.Render(b.LastCommit) + "\n\n")

	// Type
	sb.WriteString(ui.LabelStyle.Render("Type") + "\n")
	if b.IsCurrent {
		sb.WriteString(ui.ContentStyle.Render("Current branch") + "\n")
	} else if b.IsRemote {
		sb.WriteString(ui.ContentStyle.Render("Remote branch") + "\n")
	} else {
		sb.WriteString(ui.ContentStyle.Render("Local branch") + "\n")
	}

	return sb.String()
}

// GeneratePreview generates a lightweight preview of the worktree
func (w Worktree) GeneratePreview(width, height int) string {
	var sb strings.Builder

	sb.WriteString(ui.LabelStyle.Render("Path") + "\n")
	sb.WriteString(ui.ContentStyle.Render(w.Path) + "\n\n")

	sb.WriteString(ui.LabelStyle.Render("Branch") + "\n")
	sb.WriteString(ui.ContentStyle.Render(w.Branch) + "\n\n")

	sb.WriteString(ui.LabelStyle.Render("Commit") + "\n")
	sb.WriteString(ui.ContentStyle.Render(w.Head) + "\n")

	if w.IsCurrent {
		sb.WriteString("\n")
		sb.WriteString(ui.LabelStyle.Render("Type") + "\n")
		sb.WriteString(ui.ContentStyle.Render("Current worktree") + "\n")
	}

	return sb.String()
}
