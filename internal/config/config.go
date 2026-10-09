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

	"go.yaml.in/yaml/v3"
)

// Action names, used as keys under `keybinds:` in the config file.
const (
	ActionHistory   = "history"
	ActionGitBranch = "git_branch"
	ActionFiles     = "files"
	ActionWorktree  = "worktree"
	ActionCommit    = "commit"
	ActionSelect    = "select"
	ActionComplete  = "complete"
	ActionCopy      = "copy"
	ActionQuit      = "quit"
	ActionUp        = "up"
	ActionDown      = "down"
)

// DefaultKeybinds maps each action to the keys bound to it when the config
// file does not override that action. Keys use Bubble Tea's KeyPressMsg.String()
// notation (e.g. "ctrl+r", "enter", "up").
var DefaultKeybinds = map[string][]string{
	ActionHistory:   {"ctrl+r"},
	ActionGitBranch: {"ctrl+g"},
	ActionFiles:     {"ctrl+s"},
	ActionWorktree:  {"ctrl+w"},
	ActionCommit:    {"ctrl+x"},
	ActionSelect:    {"enter"},
	ActionComplete:  {"tab"},
	ActionCopy:      {"ctrl+y"},
	ActionQuit:      {"esc", "ctrl+c"},
	ActionUp:        {"up", "ctrl+p"},
	ActionDown:      {"down", "ctrl+n"},
}

type file struct {
	Keybinds map[string][]string `yaml:"keybinds"`
}

// Path returns the config file location.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "fuzz.fish", "fuzz.fish.yaml"), nil
}

// LoadKeymap reads the config file and returns a key → action lookup table.
// A missing file yields the defaults.
func LoadKeymap() (map[string]string, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return ParseKeymap(nil)
	}
	if err != nil {
		return nil, err
	}
	km, err := ParseKeymap(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return km, nil
}

// ParseKeymap builds the key → action table from YAML. Actions listed in the
// YAML replace their default keys; an empty list unbinds the action.
func ParseKeymap(data []byte) (map[string]string, error) {
	var f file
	if len(bytes.TrimSpace(data)) > 0 {
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		if err := dec.Decode(&f); err != nil {
			return nil, err
		}
	}

	for action := range f.Keybinds {
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
		keys, ok := f.Keybinds[action]
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
