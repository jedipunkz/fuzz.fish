package app

import (
	"testing"

	"github.com/jedipunkz/fuzz.fish/internal/git"
)

// selectModel builds a git-branch-mode model with one selectable item.
func selectModel(branch git.Branch) model {
	return model{
		mode: ModeGitBranch,
		filtered: []Item{{
			Text:      branch.Name,
			Index:     0,
			Original:  branch,
			IsRemote:  branch.IsRemote,
			IsCurrent: branch.IsCurrent,
		}},
		cursor: 0,
	}
}

func TestSelectItem_RemoteNameWithSlash(t *testing.T) {
	m := selectModel(git.Branch{
		Name:     "gitlab/team/feat/x",
		IsRemote: true,
		Remote:   "gitlab/team",
	})
	m.selectItem()

	if m.choice == nil || *m.choice != "feat/x" {
		t.Fatalf("choice = %v, want feat/x", m.choice)
	}
}

func TestSelectItem_SimpleRemoteName(t *testing.T) {
	m := selectModel(git.Branch{
		Name:     "origin/main",
		IsRemote: true,
		Remote:   "origin",
	})
	m.selectItem()

	if m.choice == nil || *m.choice != "main" {
		t.Fatalf("choice = %v, want main", m.choice)
	}
}

func TestSelectItem_UnconfiguredRemoteFallback(t *testing.T) {
	m := selectModel(git.Branch{
		Name:     "origin/main",
		IsRemote: true,
	})
	m.selectItem()

	if m.choice == nil || *m.choice != "main" {
		t.Fatalf("choice = %v, want main", m.choice)
	}
}

func TestSelectItem_LocalBranchKeepsName(t *testing.T) {
	m := selectModel(git.Branch{Name: "feat/nested/path"})
	m.selectItem()

	if m.choice == nil || *m.choice != "feat/nested/path" {
		t.Fatalf("choice = %v, want feat/nested/path", m.choice)
	}
}
