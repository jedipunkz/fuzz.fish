package files

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Collect walks the directory tree and collects all files and directories
func Collect(root string) []Entry {
	var files []Entry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip hidden files/directories (except current dir)
		if path != root && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			relPath = path
		}

		// Skip current directory itself
		if relPath == "." {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		files = append(files, Entry{
			Path:  relPath,
			IsDir: d.IsDir(),
			Size:  info.Size(),
			Mode:  info.Mode(),
		})

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
	}

	return files
}
