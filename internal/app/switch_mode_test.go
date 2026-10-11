package app

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/jedipunkz/fuzz.fish/internal/git"
)

func TestInit_ReturnsLoadCmd(t *testing.T) {
	m := model{}
	if cmd := m.Init(); cmd == nil {
		t.Fatal("Init() = nil cmd, want the history load command")
	}
}

func TestSwitchMode_SameModeIsNoop(t *testing.T) {
	m := model{mode: ModeGitBranch}
	if cmd := (&m).switchToGitBranchMode(); cmd != nil {
		t.Error("switch to the current mode must return nil")
	}
	if m.loading {
		t.Error("loading must be untouched in the same mode")
	}
}

func TestSwitchMode_CacheEmptyStartsLoad(t *testing.T) {
	tests := []struct {
		name string
		run  func(*model) tea.Cmd
		want SearchMode
	}{
		{"git branch", (*model).switchToGitBranchMode, ModeGitBranch},
		{"files", (*model).switchToFilesMode, ModeFiles},
		{"worktree", (*model).switchToWorktreeMode, ModeWorktree},
		{"commit", (*model).switchToCommitMode, ModeCommit},
		{"pull request", (*model).switchToPullRequestMode, ModePullRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{mode: ModeHistory}
			cmd := tt.run(&m)
			if cmd == nil {
				t.Fatal("cmd = nil, want the load command")
			}
			if m.mode != tt.want {
				t.Errorf("mode = %v, want %v", m.mode, tt.want)
			}
			if !m.loading {
				t.Error("loading = false, want true while data is being fetched")
			}
			if m.filtered != nil {
				t.Error("filtered must be reset for the new mode")
			}
			if m.cursor != 0 || m.offset != 0 {
				t.Error("cursor and offset must be reset for the new mode")
			}
		})
	}
}

func TestSwitchMode_CachedDataRenders(t *testing.T) {
	m := model{
		mode:  ModeHistory,
		input: textinput.New(),
	}
	m.gitBranches = []git.Branch{{Name: "main"}, {Name: "feat/x"}}

	cmd := m.switchToGitBranchMode()
	if m.mode != ModeGitBranch {
		t.Errorf("mode = %v, want ModeGitBranch", m.mode)
	}
	if m.loading {
		t.Error("loading = true, want false with cached data")
	}
	if len(m.filtered) != 2 {
		t.Errorf("filtered = %d items, want 2", len(m.filtered))
	}
	if cmd == nil {
		t.Error("cmd = nil, want the filter/preview command batch")
	}
}
