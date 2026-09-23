package files

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
)

// Collector walks the directory tree and collects files and directories
type Collector struct {
	Root     string
	MaxFiles int
	SkipDirs map[string]bool

	// Truncated reports whether the last Collect hit MaxFiles and dropped
	// the remaining entries.
	Truncated bool
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

// Collect returns the entries below Root. Inside a git repository the listing
// comes from git, so .gitignore rules apply and dotfiles are included;
// elsewhere it falls back to walking the tree.
func (c *Collector) Collect() []Entry {
	c.Truncated = false

	if entries, ok := c.collectGit(); ok {
		return entries
	}
	return c.collectWalk()
}

// collectGit lists the tracked and untracked-but-not-ignored files under Root
// via git. It reports false when Root is not in a repository or git fails, so
// the caller can fall back to walking.
func (c *Collector) collectGit() ([]Entry, bool) {
	// Argument list, no shell. -z keeps paths verbatim instead of quoting
	// the ones with unusual characters.
	cmd := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z")
	cmd.Dir = c.Root
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}

	entries, truncated := entriesFromLsFiles(string(out), c.MaxFiles)
	c.Truncated = truncated
	return entries, true
}

// entriesFromLsFiles converts NUL-separated `git ls-files` output into file
// entries plus the directory entries implied by their path prefixes, since git
// lists files only. It stops at maxFiles and reports whether it did.
func entriesFromLsFiles(out string, maxFiles int) ([]Entry, bool) {
	entries := make([]Entry, 0, 512)
	seenDirs := make(map[string]bool)

	for _, path := range strings.Split(out, "\x00") {
		if path == "" {
			continue
		}

		for i := 0; i < len(path); i++ {
			if path[i] != '/' {
				continue
			}
			dir := path[:i]
			if seenDirs[dir] {
				continue
			}
			if len(entries) >= maxFiles {
				return entries, true
			}
			seenDirs[dir] = true
			entries = append(entries, Entry{Path: dir, IsDir: true})
		}

		if len(entries) >= maxFiles {
			return entries, true
		}
		entries = append(entries, Entry{Path: path, IsDir: false})
	}

	return entries, false
}

// collectWalk walks the directory tree, skipping dotfiles and known heavy
// directories. Used outside git repositories.
func (c *Collector) collectWalk() []Entry {
	// Pre-allocate with a reasonable initial capacity to reduce re-allocations
	files := make([]Entry, 0, 512)

	_ = filepath.WalkDir(c.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Check file limit
		if len(files) >= c.MaxFiles {
			c.Truncated = true
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
