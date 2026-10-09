// Package config loads user settings from ~/.config/fuzz.fish/fuzz.fish.yaml.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Action names, used as keys under `keybinds:` in the config file.
const (
	ActionHistory     = "history"
	ActionGitBranch   = "git_branch"
	ActionFiles       = "files"
	ActionWorktree    = "worktree"
	ActionCommit      = "commit"
	ActionPullRequest = "pull_request"
	ActionSelect      = "select"
	ActionComplete    = "complete"
	ActionCopy        = "copy"
	ActionQuit        = "quit"
	ActionUp          = "up"
	ActionDown        = "down"
)

// DefaultKeybinds maps each action to the keys bound to it when the config
// file does not override that action. Keys use Bubble Tea's KeyPressMsg.String()
// notation (e.g. "ctrl+r", "enter", "up").
var DefaultKeybinds = map[string][]string{
	ActionHistory:     {"ctrl+r"},
	ActionGitBranch:   {"ctrl+g"},
	ActionFiles:       {"ctrl+s"},
	ActionWorktree:    {"ctrl+w"},
	ActionCommit:      {"ctrl+x"},
	ActionPullRequest: {"ctrl+j"},
	ActionSelect:      {"enter"},
	ActionComplete:    {"tab"},
	ActionCopy:        {"ctrl+y"},
	ActionQuit:        {"esc", "ctrl+c"},
	ActionUp:          {"up", "ctrl+p"},
	ActionDown:        {"down", "ctrl+n"},
}

type file struct {
	Keybinds    map[string][]string `yaml:"keybinds"`
	WorktreeDir string              `yaml:"worktree_dir"`
}

// Config is the resolved user configuration.
type Config struct {
	Keys        map[string]string // key → action
	WorktreeDir string            // absolute path, "" when unset
}

// Path returns the config file location.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "fuzz.fish", "fuzz.fish.yaml"), nil
}

// Load reads the config file. A missing file yields the defaults.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Parse(nil)
	}
	if err != nil {
		return Config{}, err
	}
	cfg, err := Parse(data)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

// Parse resolves the configuration from YAML.
func Parse(data []byte) (Config, error) {
	var f file
	if len(bytes.TrimSpace(data)) > 0 {
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		if err := dec.Decode(&f); err != nil {
			return Config{}, err
		}
	}

	keys, err := buildKeymap(f.Keybinds)
	if err != nil {
		return Config{}, err
	}
	dir, err := expandDir(f.WorktreeDir)
	if err != nil {
		return Config{}, err
	}
	return Config{Keys: keys, WorktreeDir: dir}, nil
}

// expandDir expands a leading "~" and requires an absolute path.
func expandDir(dir string) (string, error) {
	if dir == "" {
		return "", nil
	}
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, dir[1:])
	}
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("worktree_dir must be an absolute path or start with ~/: %q", dir)
	}
	return filepath.Clean(dir), nil
}

// buildKeymap builds the key → action table. Actions listed in overrides
// replace their default keys; an empty list unbinds the action.
func buildKeymap(overrides map[string][]string) (map[string]string, error) {
	for action := range overrides {
		if _, ok := DefaultKeybinds[action]; !ok {
			return nil, fmt.Errorf("unknown keybind action %q", action)
		}
	}

	// Sorted so a duplicate-key error names the same pair on every run.
	actions := make([]string, 0, len(DefaultKeybinds))
	for action := range DefaultKeybinds {
		actions = append(actions, action)
	}
	sort.Strings(actions)

	km := make(map[string]string)
	for _, action := range actions {
		keys, ok := overrides[action]
		if !ok {
			keys = DefaultKeybinds[action]
		}
		for _, key := range keys {
			if other, dup := km[key]; dup {
				return nil, fmt.Errorf("key %q is bound to both %q and %q", key, other, action)
			}
			km[key] = action
		}
	}
	return km, nil
}
