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
