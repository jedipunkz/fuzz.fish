package files

import "fmt"

// GetFileIcon returns the appropriate icon for a file or directory
func GetFileIcon(isDir bool) string {
	if isDir {
		return "📁"
	}
	return "📄"
}

// FormatEntry formats a file entry for display in the TUI list
func FormatEntry(e Entry) string {
	return fmt.Sprintf("%s %s", GetFileIcon(e.IsDir), e.Path)
}
