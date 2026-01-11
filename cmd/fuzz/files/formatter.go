package files

import "fmt"

// FormatEntry formats a file entry for display in the TUI list
func FormatEntry(e Entry) string {
	icon := "📄"
	if e.IsDir {
		icon = "📁"
	}
	return fmt.Sprintf("%s %s", icon, e.Path)
}
