package app

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"github.com/jedipunkz/fuzz.fish/internal/git"
)

func TestUpdatePreviewAsync(t *testing.T) {
	m := model{
		mode:         ModeCommit,
		viewport:     viewport.New(),
		previewCache: map[string]string{},
		width:        80,
		listWidth:    40,
		mainHeight:   10,
		ready:        true,
	}
	m.viewport.SetWidth(40)
	m.viewport.SetHeight(10)
	m.filtered = []Item{{Index: 0, Original: git.Commit{Hash: "a1b2c3d", Subject: "commit subject"}}}
	m.previewGen = 1

	// Commit previews generate in a tea.Cmd: without it the update loop
	// blocks on the git show subprocess.
	cmd := m.updatePreview()
	if cmd == nil {
		t.Fatal("commit preview should generate asynchronously, got no command")
	}
	if !strings.Contains(m.viewport.View(), "loading preview") {
		t.Errorf("viewport shows %q while the preview generates, want the loading hint", m.viewport.View())
	}

	// The result lands: render it and cache it.
	msg := cmd().(previewReadyMsg)
	updated, _ := m.Update(previewReadyMsg{gen: msg.gen, key: msg.key, content: "PREVIEW-CONTENT"})
	m = updated.(model)
	if !strings.Contains(m.viewport.View(), "PREVIEW-CONTENT") {
		t.Errorf("viewport shows %q after the preview arrived", m.viewport.View())
	}
	if m.previewCache[msg.key] != "PREVIEW-CONTENT" {
		t.Errorf("previewCache[%q] = %q, want PREVIEW-CONTENT", msg.key, m.previewCache[msg.key])
	}

	// A render from a previous resize generation must be dropped entirely:
	// it was sized for a pane that no longer exists.
	updated, _ = m.Update(previewReadyMsg{gen: 0, key: msg.key, content: "STALE"})
	m = updated.(model)
	if strings.Contains(m.viewport.View(), "STALE") {
		t.Error("viewport rendered a preview from a previous resize generation")
	}
	if m.previewCache[msg.key] != "PREVIEW-CONTENT" {
		t.Errorf("previewCache[%q] = %q, want the stale result to be discarded", msg.key, m.previewCache[msg.key])
	}
}
