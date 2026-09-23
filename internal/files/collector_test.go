package files

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEntriesFromLsFiles(t *testing.T) {
	tests := []struct {
		name          string
		out           string
		maxFiles      int
		want          []Entry
		wantTruncated bool
	}{
		{
			name:     "empty output",
			out:      "",
			maxFiles: 10,
			want:     []Entry{},
		},
		{
			name:     "directories derived from prefixes and deduplicated",
			out:      "README.md\x00a/b/one.go\x00a/b/two.go\x00a/c.go\x00",
			maxFiles: 10,
			want: []Entry{
				{Path: "README.md"},
				{Path: "a", IsDir: true},
				{Path: "a/b", IsDir: true},
				{Path: "a/b/one.go"},
				{Path: "a/b/two.go"},
				{Path: "a/c.go"},
			},
		},
		{
			name:     "dotfiles and dot directories are kept",
			out:      ".gitignore\x00.github/workflows/ci.yml\x00",
			maxFiles: 10,
			want: []Entry{
				{Path: ".gitignore"},
				{Path: ".github", IsDir: true},
				{Path: ".github/workflows", IsDir: true},
				{Path: ".github/workflows/ci.yml"},
			},
		},
		{
			name:          "cap reached on a file",
			out:           "a.go\x00b.go\x00c.go\x00",
			maxFiles:      2,
			want:          []Entry{{Path: "a.go"}, {Path: "b.go"}},
			wantTruncated: true,
		},
		{
			name:          "cap reached on a derived directory",
			out:           "a.go\x00b/c.go\x00",
			maxFiles:      1,
			want:          []Entry{{Path: "a.go"}},
			wantTruncated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, truncated := entriesFromLsFiles(tt.out, tt.maxFiles)
			if truncated != tt.wantTruncated {
				t.Errorf("truncated = %v, want %v", truncated, tt.wantTruncated)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d entries %+v, want %d %+v", len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("entry %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// initTestRepo creates a git repository in a temp directory containing the
// given files (paths relative to the repository root) and returns its path.
func initTestRepo(t *testing.T, paths ...string) string {
	t.Helper()

	dir := t.TempDir()
	for _, p := range paths {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll(%q) failed: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte("x\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) failed: %v", full, err)
		}
	}

	cmd := exec.Command("git", "init", "--initial-branch=main")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	return dir
}

func paths(entries []Entry) map[string]bool {
	m := make(map[string]bool, len(entries))
	for _, e := range entries {
		m[e.Path] = e.IsDir
	}
	return m
}

func TestCollect_GitRepo(t *testing.T) {
	dir := initTestRepo(t,
		".gitignore",
		".github/workflows/ci.yml",
		"main.go",
		"ignored.log",
		"node_modules/pkg/index.js",
	)
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\nnode_modules/\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(.gitignore) failed: %v", err)
	}

	c := NewCollector(dir)
	got := paths(c.Collect())

	if c.Truncated {
		t.Error("Truncated = true, want false")
	}
	for _, want := range []string{".gitignore", ".github/workflows/ci.yml", "main.go"} {
		if _, ok := got[want]; !ok {
			t.Errorf("%q missing from listing %v", want, got)
		}
	}
	if isDir, ok := got[".github/workflows"]; !ok || !isDir {
		t.Errorf("directory entry %q missing or not a directory: %v", ".github/workflows", got)
	}
	for _, unwanted := range []string{"ignored.log", "node_modules/pkg/index.js"} {
		if _, ok := got[unwanted]; ok {
			t.Errorf("%q should be excluded by .gitignore", unwanted)
		}
	}
}

func TestCollect_GitRepoTruncated(t *testing.T) {
	dir := initTestRepo(t, "a.go", "b.go", "c.go")

	c := NewCollector(dir)
	c.MaxFiles = 2
	entries := c.Collect()

	if len(entries) != 2 {
		t.Errorf("got %d entries, want 2", len(entries))
	}
	if !c.Truncated {
		t.Error("Truncated = false, want true")
	}
}

func TestCollect_NonRepoFallback(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{"visible.go", ".hidden"} {
		if err := os.WriteFile(filepath.Join(dir, p), []byte("x\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) failed: %v", p, err)
		}
	}

	c := NewCollector(dir)
	got := paths(c.Collect())

	if _, ok := got["visible.go"]; !ok {
		t.Errorf("visible.go missing from listing %v", got)
	}
	if _, ok := got[".hidden"]; ok {
		t.Error(".hidden should be skipped by the walk fallback")
	}
}

func TestCollect_WalkTruncated(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{"a.go", "b.go", "c.go"} {
		if err := os.WriteFile(filepath.Join(dir, p), []byte("x\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) failed: %v", p, err)
		}
	}

	c := NewCollector(dir)
	c.MaxFiles = 2
	entries := c.Collect()

	if len(entries) != 2 {
		t.Errorf("got %d entries, want 2", len(entries))
	}
	if !c.Truncated {
		t.Error("Truncated = false, want true")
	}
}
