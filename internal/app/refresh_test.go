package app

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/jedipunkz/fuzz.fish/internal/git"
)

func TestRefreshMode_RunsLoaderForCurrentMode(t *testing.T) {
	tests := []struct {
		name string
		mode SearchMode
	}{
		{"history", ModeHistory},
		{"git branch", ModeGitBranch},
		{"files", ModeFiles},
		{"worktree", ModeWorktree},
		{"commit", ModeCommit},
		{"pull request", ModePullRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{mode: tt.mode, filtered: []Item{{Text: "stale"}}}
			cmd := m.refreshMode()
			if cmd == nil {
				t.Fatal("cmd = nil, want the load command")
			}
			if m.mode != tt.mode {
				t.Errorf("mode = %v, want %v", m.mode, tt.mode)
			}
			if !m.loading {
				t.Error("loading = false, want true while data is being refetched")
			}
			if m.filtered != nil || m.allItems != nil {
				t.Error("items must be reset while reloading")
			}
			if m.cursor != 0 || m.offset != 0 {
				t.Error("cursor and offset must be reset for the reloaded data")
			}
			if len(m.previewCache) != 0 || m.lastPreviewKey != "" {
				t.Error("preview cache must be cleared for the reloaded data")
			}
		})
	}
}

func TestRefreshMode_WhileLoadingIsNoop(t *testing.T) {
	m := model{mode: ModeGitBranch, loading: true, filtered: []Item{{Text: "x"}}}
	if cmd := m.refreshMode(); cmd != nil {
		t.Error("cmd = non-nil, want nil while a load is already in flight")
	}
	if len(m.filtered) != 1 {
		t.Error("filtered must be untouched by a no-op refresh")
	}
}

func TestRefreshMode_KeepsQueryAndRendersReloadedData(t *testing.T) {
	m := model{mode: ModeHistory, input: textinput.New()}
	m.gitBranches = []git.Branch{{Name: "main"}, {Name: "feat/x"}}
	cmd := m.switchToGitBranchMode()
	m = runCmd(m, cmd)
	if len(m.filtered) != 2 {
		t.Fatalf("filtered = %d items, want 2 before the reload", len(m.filtered))
	}
	m.input.SetValue("feat")
	m = runCmd(m, m.updateFilter(m.input.Value()))

	// The branch list is cached, so a mode switch alone would keep the stale
	// data; refresh re-runs the loader instead. The returned load command is
	// not executed (it shells out to git); the reload is simulated by feeding
	// the completion message directly.
	if cmd := m.refreshMode(); cmd == nil {
		t.Fatal("cmd = nil, want the load command from the refresh")
	}
	m = runCmd(m, func() tea.Msg {
		return branchesLoadedMsg{branches: []git.Branch{{Name: "feat/new"}, {Name: "main"}}}
	})
	if len(m.filtered) != 1 || m.filtered[0].Text != "feat/new" {
		t.Errorf("filtered = %v, want only feat/new from the reloaded data", itemTexts(m.filtered))
	}
	if m.input.Value() != "feat" {
		t.Errorf("query = %q, want the pre-refresh query kept", m.input.Value())
	}
}

func TestUpdate_RefreshKeyRunsLoader(t *testing.T) {
	m := model{
		mode:         ModeGitBranch,
		input:        textinput.New(),
		keys:         map[string]string{"f5": "refresh"},
		previewCache: map[string]string{},
	}
	m.gitBranches = []git.Branch{{Name: "main"}}
	m = runCmd(m, m.switchToGitBranchMode())

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyF5})
	if cmd == nil {
		t.Fatal("cmd = nil, want the load command from the refresh key")
	}
	m, ok := updated.(model)
	if !ok {
		t.Fatal("Update must return the model")
	}
	if !m.loading {
		t.Error("loading = false, want true after the refresh key")
	}
}

func itemTexts(items []Item) []string {
	texts := make([]string, len(items))
	for i, item := range items {
		texts[i] = item.Text
	}
	return texts
}
