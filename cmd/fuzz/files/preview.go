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

	// Path
	sb.WriteString(ui.LabelStyle.Render(entry.Path) + "\n\n")

	if entry.IsDir {
		// Directory listing
		sb.WriteString(ui.ContextHeaderStyle.Render("Contents") + "\n")
		listing := GetDirectoryListing(entry.Path)
		if listing != "" {
			sb.WriteString(listing)
		} else {
			sb.WriteString(ui.InactiveContextStyle.Render("  (empty)") + "\n")
		}
	} else {
		// File preview
		sb.WriteString(ui.ContextHeaderStyle.Render("Preview") + "\n")
		preview := utils.GetFilePreview(entry.Path, height-5)
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

	for _, entry := range entries {
		if count >= utils.MaxDirectoryEntries {
			sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  ... and %d more", len(entries)-utils.MaxDirectoryEntries)) + "\n")
			break
		}

		icon := GetFileIcon(entry.IsDir())
		sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  %s %s", icon, entry.Name())) + "\n")
		count++
	}

	return sb.String()
}
