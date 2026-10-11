package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse_Defaults(t *testing.T) {
	cfg, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	km := cfg.Keys
	for action, keys := range DefaultKeybinds {
		for _, key := range keys {
			if km[key] != action {
				t.Errorf("km[%q] = %q, want %q", key, km[key], action)
			}
		}
	}
}

func TestParse_Override(t *testing.T) {
	cfg, err := Parse([]byte("keybinds:\n  files: [ctrl+f]\n  copy: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	km := cfg.Keys
	if km["ctrl+f"] != ActionFiles {
		t.Errorf("ctrl+f = %q, want files", km["ctrl+f"])
	}
	if _, ok := km["ctrl+s"]; ok {
		t.Error("ctrl+s should no longer be bound")
	}
	if _, ok := km["ctrl+y"]; ok {
		t.Error("ctrl+y should be unbound")
	}
	if km["ctrl+r"] != ActionHistory {
		t.Errorf("ctrl+r = %q, want history (default kept)", km["ctrl+r"])
	}
}

func TestParse_Errors(t *testing.T) {
	for name, in := range map[string]string{
		"unknown action": "keybinds:\n  nope: [ctrl+a]\n",
		"unknown field":  "keybind:\n  files: [ctrl+f]\n",
		"duplicate key":  "keybinds:\n  files: [ctrl+r]\n",
		"relative dir":   "worktree_dir: gm/.worktrees\n",
	} {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestParse_WorktreeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := Parse([]byte("worktree_dir: ~/gm/.worktrees/\n"))
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "gm", ".worktrees"); cfg.WorktreeDir != want {
		t.Errorf("WorktreeDir = %q, want %q", cfg.WorktreeDir, want)
	}
}

func TestPath(t *testing.T) {
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(".config", "fuzz.fish", "fuzz.fish.yaml")
	if !strings.HasSuffix(path, want) {
		t.Errorf("Path() = %q, want suffix %q", path, want)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Keys) == 0 {
		t.Error("expected the default keymap for a missing config file")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "fuzz.fish")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fuzz.fish.yaml"), []byte("keybinds:\n  files: [ctrl+f]\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Keys["ctrl+f"] != ActionFiles {
		t.Errorf("ctrl+f = %q, want files", cfg.Keys["ctrl+f"])
	}
}

func TestLoad_InvalidFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "fuzz.fish")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fuzz.fish.yaml"), []byte("%%% not yaml\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "fuzz.fish.yaml") {
		t.Errorf("error = %v, want the config file path in the message", err)
	}
}
