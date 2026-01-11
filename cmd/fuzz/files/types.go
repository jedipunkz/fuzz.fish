package files

import "io/fs"

// Entry represents a file or directory
type Entry struct {
	Path  string
	IsDir bool
	Size  int64
	Mode  fs.FileMode
}
