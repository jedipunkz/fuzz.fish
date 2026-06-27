package history

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Parser reads and parses Fish shell history
type Parser struct {
	Path     string // history file path
	CacheDir string // optional cache directory override, mainly for tests
}

type cacheFile struct {
	Version int       `json:"version"`
	Meta    cacheMeta `json:"meta"`
	Entries []Entry   `json:"entries"`
}

type cacheMeta struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mod_time"`
	Dev     uint64 `json:"dev"`
	Inode   uint64 `json:"inode"`
}

const cacheVersion = 1

// NewParser returns a Parser with the default Fish history file path
func NewParser() *Parser {
	home, err := os.UserHomeDir()
	if err != nil {
		return &Parser{}
	}
	return &Parser{
		Path: filepath.Join(home, ".local", "share", "fish", "fish_history"),
	}
}

// Parse reads and parses the Fish shell history file
func (p *Parser) Parse() []Entry {
	if p.Path == "" {
		return []Entry{}
	}

	info, err := os.Stat(p.Path)
	if err != nil {
		return []Entry{}
	}

	meta := p.cacheMeta(info)
	if entries, ok := p.readCache(meta); ok {
		if currentInfo, err := os.Stat(p.Path); err == nil && p.cacheMeta(currentInfo) == meta {
			return entries
		}
	}

	file, err := os.Open(p.Path)
	if err != nil {
		return []Entry{}
	}
	defer file.Close() //nolint:errcheck

	entries := parseReader(file)
	if currentInfo, err := os.Stat(p.Path); err == nil && p.cacheMeta(currentInfo) == meta {
		p.writeCache(meta, entries)
	}
	return entries
}

func (p *Parser) cacheMeta(info os.FileInfo) cacheMeta {
	path := p.Path
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}

	dev, inode := fileIdentity(info)
	return cacheMeta{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime().UnixNano(),
		Dev:     dev,
		Inode:   inode,
	}
}

func fileIdentity(info os.FileInfo) (uint64, uint64) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0
	}
	return uint64(stat.Dev), uint64(stat.Ino)
}

func (p *Parser) readCache(meta cacheMeta) ([]Entry, bool) {
	path := p.cachePath()
	if path == "" {
		return nil, false
	}

	_ = os.Chmod(path, 0o600)
	file, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer file.Close() //nolint:errcheck

	var cached cacheFile
	if err := json.NewDecoder(file).Decode(&cached); err != nil {
		return nil, false
	}
	if cached.Version != cacheVersion || cached.Meta != meta {
		return nil, false
	}
	if cached.Entries == nil {
		return []Entry{}, true
	}
	return cached.Entries, true
}

func (p *Parser) writeCache(meta cacheMeta, entries []Entry) {
	path := p.cachePath()
	if path == "" {
		return
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return
	}

	tmp, err := os.CreateTemp(dir, "history-cache-*.tmp")
	if err != nil {
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) //nolint:errcheck

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return
	}

	cached := cacheFile{
		Version: cacheVersion,
		Meta:    meta,
		Entries: entries,
	}
	enc := json.NewEncoder(tmp)
	if err := enc.Encode(cached); err != nil {
		_ = tmp.Close()
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}

	_ = os.Chmod(path, 0o600)
	if err := os.Rename(tmpPath, path); err != nil {
		return
	}
	_ = os.Chmod(path, 0o600)
}

func (p *Parser) cachePath() string {
	cacheDir := p.CacheDir
	if cacheDir == "" {
		cacheHome := os.Getenv("XDG_CACHE_HOME")
		if cacheHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}
			cacheHome = filepath.Join(home, ".cache")
		}
		cacheDir = filepath.Join(cacheHome, "fuzz.fish")
	}
	return filepath.Join(cacheDir, "history-cache.json")
}

// parseReader parses Fish shell history entries from an io.Reader.
// This is exported for testing purposes.
func parseReader(r io.Reader) []Entry {
	var entries []Entry
	var current *Entry
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if strings.HasPrefix(line, "- cmd: ") {
			if current != nil {
				entries = append(entries, *current)
			}
			current = &Entry{
				Cmd:     strings.TrimPrefix(line, "- cmd: "),
				CmdLine: lineNum,
			}
		} else if current != nil {
			if strings.HasPrefix(line, "  when: ") {
				whenStr := strings.TrimPrefix(line, "  when: ")
				when, err := strconv.ParseInt(whenStr, 10, 64)
				if err == nil {
					current.When = when
				}
			} else if strings.HasPrefix(line, "    - ") {
				path := strings.TrimPrefix(line, "    - ")
				current.Paths = append(current.Paths, path)
			}
		}
	}

	if current != nil {
		entries = append(entries, *current)
	}

	// Reverse to show newest first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	// Deduplicate commands - keep only the newest occurrence
	seen := make(map[string]bool)
	deduplicated := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if !seen[entry.Cmd] {
			seen[entry.Cmd] = true
			deduplicated = append(deduplicated, entry)
		}
	}

	return deduplicated
}
