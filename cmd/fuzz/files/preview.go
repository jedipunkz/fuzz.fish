package files

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/utils"
)

// GeneratePreview generates a preview of the file entry for the TUI preview window
func GeneratePreview(entry Entry, width, height int) string {
	var sb strings.Builder

	// Header
	sb.WriteString(ui.HeaderStyle.Render("File Information") + "\n\n")

	// Path
	sb.WriteString(ui.LabelStyle.Render("Path") + "\n")
	sb.WriteString(ui.ContentStyle.Render(entry.Path) + "\n\n")

	// Type
	sb.WriteString(ui.LabelStyle.Render("Type") + "\n")
	if entry.IsDir {
		sb.WriteString(ui.ContentStyle.Render("Directory") + "\n\n")

		// Directory listing
		sb.WriteString(ui.ContextHeaderStyle.Render("Contents") + "\n")
		listing := GetDirectoryListing(entry.Path)
		if listing != "" {
			sb.WriteString(listing)
		} else {
			sb.WriteString(ui.InactiveContextStyle.Render("  (empty)") + "\n")
		}
	} else {
		sb.WriteString(ui.ContentStyle.Render("File") + "\n\n")

		// Size
		sb.WriteString(ui.LabelStyle.Render("Size") + "\n")
		sb.WriteString(ui.ContentStyle.Render(utils.FormatFileSize(entry.Size)) + "\n\n")

		// Permissions
		sb.WriteString(ui.LabelStyle.Render("Permissions") + "\n")
		sb.WriteString(ui.ContentStyle.Render(entry.Mode.String()) + "\n\n")

		// File preview
		sb.WriteString(ui.ContextHeaderStyle.Render("Preview") + "\n")
		preview := utils.GetFilePreview(entry.Path, height-15)
		if preview != "" {
			sb.WriteString(preview)
		} else {
			sb.WriteString(ui.InactiveContextStyle.Render("  (binary or empty file)") + "\n")
		}
	}

	return sb.String()
}

// GetDirectoryListing returns a formatted listing of a directory's contents
func GetDirectoryListing(path string) string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return ""
	}

	var sb strings.Builder
	count := 0
	maxEntries := 20

	for _, entry := range entries {
		if count >= maxEntries {
			sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  ... and %d more", len(entries)-maxEntries)) + "\n")
			break
		}

		icon := "📄"
		if entry.IsDir() {
			icon = "📁"
		}
		sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  %s %s", icon, entry.Name())) + "\n")
		count++
	}

	return sb.String()
}
