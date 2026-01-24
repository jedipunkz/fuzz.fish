package files

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Maximum number of files to collect
const MaxFiles = 5000

// Directories to skip during collection
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"__pycache__":  true,
	".venv":        true,
	"venv":         true,
	".cache":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".nuxt":        true,
	"target":       true, // Rust/Java
}

// Collect walks the directory tree and collects all files and directories
func Collect(root string) []Entry {
	var files []Entry

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Check file limit
		if len(files) >= MaxFiles {
			return filepath.SkipAll
		}

		// Skip known heavy directories
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
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

		files = append(files, Entry{
			Path:  relPath,
			IsDir: d.IsDir(),
		})

		return nil
	})

	return files
}
