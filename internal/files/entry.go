package files

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// iconFile and iconDir are pre-allocated icon strings to avoid repeated allocation
const (
	iconFile = "📄"
	iconDir  = "📁"
)

// Entry represents a file or directory
type Entry struct {
	Path  string
	IsDir bool
}

// GetInfo returns file info (size and mode) for the entry
func (e Entry) GetInfo() (int64, fs.FileMode) {
	info, err := os.Stat(e.Path)
	if err != nil {
		return 0, 0
	}
	return info.Size(), info.Mode()
}

// Icon returns the appropriate icon for a file or directory
func (e Entry) Icon() string {
	if e.IsDir {
		return iconDir
	}
	return iconFile
}

// Format formats the entry for display in the TUI list
func (e Entry) Format() string {
	return e.Icon() + " " + e.Path
}

// GeneratePreview generates a preview of the file entry for the TUI preview window
func (e Entry) GeneratePreview(width, height int) string {
	var sb strings.Builder

	// Path
	sb.WriteString(ui.LabelStyle.Render(e.Path) + "\n\n")

	if e.IsDir {
		// Directory listing
		sb.WriteString(ui.ContextHeaderStyle.Render("Contents") + "\n")
		listing := e.DirectoryListing()
		if listing != "" {
			sb.WriteString(listing)
		} else {
			sb.WriteString(ui.InactiveContextStyle.Render("  (empty)") + "\n")
		}
	} else {
		// File preview
		sb.WriteString(ui.ContextHeaderStyle.Render("Preview") + "\n")
		preview := ui.GetFilePreview(e.Path, height-5)
		if preview != "" {
			sb.WriteString(preview)
		} else {
			sb.WriteString(ui.InactiveContextStyle.Render("  (binary or empty file)") + "\n")
		}
	}

	return sb.String()
}

// DirectoryListing returns a formatted listing of the entry's directory contents
func (e Entry) DirectoryListing() string {
	entries, err := os.ReadDir(e.Path)
	if err != nil {
		return ""
	}

	limit := ui.MaxDirectoryEntries
	if limit > len(entries) {
		limit = len(entries)
	}

	var sb strings.Builder
	sb.Grow(limit * 32) // pre-allocate for typical line length

	for i := 0; i < limit; i++ {
		dirEntry := entries[i]
		icon := iconFile
		if dirEntry.IsDir() {
			icon = iconDir
		}
		sb.WriteString(ui.InactiveContextStyle.Render("  "+icon+" "+dirEntry.Name()) + "\n")
	}

	if len(entries) > ui.MaxDirectoryEntries {
		sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  ... and %d more", len(entries)-ui.MaxDirectoryEntries)) + "\n")
	}

	return sb.String()
}
