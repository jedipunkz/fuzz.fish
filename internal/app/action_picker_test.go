package app

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/jedipunkz/fuzz.fish/internal/config"
	"github.com/jedipunkz/fuzz.fish/internal/files"
	"github.com/jedipunkz/fuzz.fish/internal/git"
)

func pickerModel() model {
	return model{
		keys: map[string]string{
			"enter": config.ActionSelect,
			"esc":   config.ActionQuit,
			"up":    config.ActionUp,
			"down":  config.ActionDown,
		},
		pendingCommit: "abc1234",
		actionCursor:  0,
	}
}

func TestUpdateActionPicker_SelectRunsTemplate(t *testing.T) {
	m := pickerModel()
	updated, cmd := m.updateActionPicker(tea.KeyPressMsg{Code: tea.KeyEnter})

	next := updated.(model)
	want := "git show abc1234"
	if next.choice == nil || *next.choice != want {
		t.Errorf("choice = %v, want %q", next.choice, want)
	}
	if !next.commitIsCmd {
		t.Error("commitIsCmd = false, want true for a templated action")
	}
	if !next.quitting {
		t.Error("quitting = false, want true on select")
	}
	if cmd == nil {
		t.Error("cmd = nil, want the quit command")
	}
}

func TestUpdateActionPicker_SelectHashOnly(t *testing.T) {
	m := pickerModel()
	m.actionCursor = len(commitActions) - 1

	updated, cmd := m.updateActionPicker(tea.KeyPressMsg{Code: tea.KeyEnter})
	next := updated.(model)

	if next.choice == nil || *next.choice != "abc1234" {
		t.Errorf("choice = %v, want abc1234", next.choice)
	}
	if next.commitIsCmd {
		t.Error("commitIsCmd = true, want false for the hash-only action")
	}
	_ = cmd
}

func TestUpdateActionPicker_QuitClosesPicker(t *testing.T) {
	m := pickerModel()
	updated, cmd := m.updateActionPicker(tea.KeyPressMsg{Code: tea.KeyEscape})

	next := updated.(model)
	if next.pendingCommit != "" {
		t.Errorf("pendingCommit = %q, want empty", next.pendingCommit)
	}
	if next.quitting {
		t.Error("quitting = true, want false when the picker just closes")
	}
	if cmd != nil {
		t.Error("cmd must be nil when the picker closes")
	}
}

func TestUpdateActionPicker_CursorMovement(t *testing.T) {
	last := len(commitActions) - 1

	m := pickerModel()
	m.actionCursor = last
	updated, _ := m.updateActionPicker(tea.KeyPressMsg{Code: tea.KeyDown})
	if updated.(model).actionCursor != last {
		t.Error("down must not move past the last action")
	}

	m = pickerModel()
	m.actionCursor = last
	updated, _ = m.updateActionPicker(tea.KeyPressMsg{Code: tea.KeyUp})
	if updated.(model).actionCursor != last-1 {
		t.Errorf("actionCursor = %d, want %d", updated.(model).actionCursor, last-1)
	}

	m = pickerModel()
	updated, _ = m.updateActionPicker(tea.KeyPressMsg{Code: tea.KeyUp})
	if updated.(model).actionCursor != 0 {
		t.Error("up must not move above the first action")
	}

	m = pickerModel()
	updated, _ = m.updateActionPicker(tea.KeyPressMsg{Code: tea.KeyDown})
	if updated.(model).actionCursor != 1 {
		t.Errorf("actionCursor = %d, want 1", updated.(model).actionCursor)
	}
}

func TestCompleteSelectedItem(t *testing.T) {
	m := model{input: textinput.New()}
	m.filtered = []Item{{Text: "git status"}}
	m.cursor = 0

	(&m).completeSelectedItem()
	if got := m.input.Value(); got != "git status" {
		t.Errorf("input = %q, want git status", got)
	}
}

func TestSelectItem_Modes(t *testing.T) {
	tests := []struct {
		name string
		mode SearchMode
		item Item
		want string
	}{
		{
			"history",
			ModeHistory,
			Item{Text: "ls -la"},
			"ls -la",
		},
		{
			"files file",
			ModeFiles,
			Item{Original: files.Entry{Path: "d/a.txt"}},
			"d/a.txt",
		},
		{
			"worktree",
			ModeWorktree,
			Item{Original: git.Worktree{Path: "/repo/wt", Branch: "feat"}},
			"/repo/wt",
		},
		{
			"pull request",
			ModePullRequest,
			Item{Original: git.PullRequest{Number: 7, URL: "https://example.com/pr/7"}},
			"https://example.com/pr/7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{mode: tt.mode}
			m.filtered = []Item{tt.item}
			(&m).selectItem()

			if m.choice == nil || *m.choice != tt.want {
				t.Errorf("choice = %v, want %q", m.choice, tt.want)
			}
		})
	}
}

func TestSelectItem_FilesDirSetsFlag(t *testing.T) {
	m := model{mode: ModeFiles}
	m.filtered = []Item{{Original: files.Entry{Path: "docs", IsDir: true}}}

	(&m).selectItem()
	if m.choice == nil || *m.choice != "docs" {
		t.Errorf("choice = %v, want docs", m.choice)
	}
	if !m.choiceIsDir {
		t.Error("choiceIsDir = false, want true")
	}
}

func TestSelectItem_PullRequestKeepsPR(t *testing.T) {
	pr := git.PullRequest{Number: 7, URL: "https://example.com/pr/7"}
	m := model{mode: ModePullRequest}
	m.filtered = []Item{{Original: pr}}

	(&m).selectItem()
	if m.choicePR.Number != 7 {
		t.Errorf("choicePR.Number = %d, want 7", m.choicePR.Number)
	}
	if m.choicePR.URL != pr.URL {
		t.Errorf("choicePR.URL = %q, want %q", m.choicePR.URL, pr.URL)
	}
}
