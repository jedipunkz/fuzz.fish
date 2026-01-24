package files

import (
	"io/fs"
	"os"
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
