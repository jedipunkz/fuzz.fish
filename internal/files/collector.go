package files

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Collector walks the directory tree and collects files and directories
type Collector struct {
	Root     string
	MaxFiles int
	SkipDirs map[string]bool
}

// NewCollector creates a Collector with sensible defaults
func NewCollector(root string) *Collector {
	return &Collector{
		Root:     root,
		MaxFiles: 5000,
		SkipDirs: map[string]bool{
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
		},
	}
}

// Collect walks the directory tree and returns all collected entries
func (c *Collector) Collect() []Entry {
	var files []Entry

	_ = filepath.WalkDir(c.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Check file limit
		if len(files) >= c.MaxFiles {
			return filepath.SkipAll
		}

		// Skip known heavy directories
		if d.IsDir() {
			if c.SkipDirs[d.Name()] {
				return filepath.SkipDir
			}
		}

		// Skip hidden files/directories (except current dir)
		if path != c.Root && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(c.Root, path)
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
