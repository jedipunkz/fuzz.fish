package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "charm.land/bubbletea/v2"
	"github.com/jedipunkz/fuzz.fish/internal/files"
	"github.com/jedipunkz/fuzz.fish/internal/git"
	"github.com/jedipunkz/fuzz.fish/internal/history"
)

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case historyLoadedMsg:
		m.historyEntries = msg.entries
		if m.mode == ModeHistory {
			m.loading = false
			m.loadItemsForMode()
			m.updateFilter(m.input.Value())
		}
		return m, nil

	case branchesLoadedMsg:
		m.gitBranches = msg.branches
		if m.mode == ModeGitBranch {
			m.loading = false
			m.loadItemsForMode()
			m.updateFilter(m.input.Value())
		}
		return m, nil

	case filesLoadedMsg:
		m.fileEntries = msg.entries
		if m.mode == ModeFiles {
			m.loading = false
			if msg.truncated {
				m.statusMsg = "⚠ showing first " + strconv.Itoa(len(msg.entries)) + " files"
			}
			m.loadItemsForMode()
			m.updateFilter(m.input.Value())
		}
		return m, nil

	case worktreesLoadedMsg:
		m.worktrees = msg.worktrees
		if m.mode == ModeWorktree {
			m.loading = false
			m.loadItemsForMode()
			m.updateFilter(m.input.Value())
		}
		return m, nil

	case commitsLoadedMsg:
		m.commits = msg.commits
		if m.mode == ModeCommit {
			m.loading = false
			m.loadItemsForMode()
			m.updateFilter(m.input.Value())
		}
		return m, nil

	case filterTickMsg:
		if msg.query == m.pendingQuery {
			m.updateFilter(msg.query)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Input box: top border (1) + input line (1) + bottom border (1) = 3
		inputHeight := 3
		mainHeight := msg.Height - inputHeight
		if mainHeight < 0 {
			mainHeight = 0
		}
		// Subtract top/bottom borders from main panes
		m.mainHeight = mainHeight - 2
		if m.mainHeight < 0 {
			m.mainHeight = 0
		}

		// 60% split for list, 40% for preview
		listWidth := int(float64(msg.Width) * 0.6)
		// Subtract left/right borders from list width
		m.listWidth = listWidth - 2
		if m.listWidth < 0 {
			m.listWidth = 0
		}

		// Preview width: remaining space minus borders
		previewWidth := msg.Width - listWidth - 2
		if previewWidth < 0 {
			previewWidth = 0
		}

		m.viewport.SetWidth(previewWidth)
		m.viewport.SetHeight(m.mainHeight)

		// Recalculate offset to keep cursor at the bottom of the view
		if len(m.filtered) > 0 {
			m.offset = m.cursor - m.mainHeight + 1
			if m.offset < 0 {
				m.offset = 0
			}
		}

		// Previews are rendered for a fixed pane size, so every cached render
		// and the change-detection key are stale once the pane is resized.
		m.previewCache = make(map[string]string)
		m.lastPreviewKey = ""

		m.validateCursor()
		m.updatePreview()

	case tea.KeyPressMsg:
		// Clear status message on any key press
		m.statusMsg = ""

		// The action picker owns every key while it is open, so typing does
		// not leak into the search box behind it.
		if m.pendingCommit != "" {
			return m.updateActionPicker(msg)
		}

		switch msg.String() {
		case "enter":
			if len(m.filtered) > 0 {
				if m.mode == ModeCommit {
					m.pendingCommit = m.filtered[m.cursor].Text
					m.actionCursor = 0
					return m, nil
				}
				m.selectItem()
				m.quitting = true
				return m, tea.Quit
			}
		case "tab":
			if len(m.filtered) > 0 {
				m.completeSelectedItem()
			}
			return m, nil
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+y":
			if len(m.filtered) > 0 {
				_ = clipboard.WriteAll(m.filtered[m.cursor].Text)
				m.quitting = true
				return m, tea.Quit
			}
		case "ctrl+g":
			if m.mode == ModeGitBranch {
				// In GitBranch mode: pull current branch or show warning
				if len(m.filtered) > 0 && m.filtered[m.cursor].IsCurrent {
					branch := m.filtered[m.cursor].Original.(git.Branch)
					res := branch.Name
					m.choice = &res
					m.fetchBranch = true
					m.quitting = true
					return m, tea.Quit
				}
				m.statusMsg = "⚠ Select current branch to pull"
				return m, nil
			}
			cmd = m.switchToGitBranchMode()
			return m, cmd
		case "ctrl+s":
			// Switch to Files mode
			cmd = m.switchToFilesMode()
			return m, cmd
		case "ctrl+w":
			// Switch to Worktree mode
			cmd = m.switchToWorktreeMode()
			return m, cmd
		case "ctrl+x":
			// Switch to Commit mode
			cmd = m.switchToCommitMode()
			return m, cmd
		case "ctrl+r":
			// Switch to History mode
			cmd = m.switchToHistoryMode()
			return m, cmd
		case "down", "ctrl+n":
			if len(m.filtered) > 0 {
				m.cursor++
				if m.cursor >= len(m.filtered) {
					m.cursor = len(m.filtered) - 1
				}
				if m.cursor >= m.offset+m.mainHeight {
					m.offset = m.cursor - m.mainHeight + 1
				}
				m.updatePreview()
			}
			return m, nil
		case "up", "ctrl+p":
			if len(m.filtered) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = 0
				}
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
				m.updatePreview()
			}
			return m, nil
		}
	}

	oldValue := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	newValue := m.input.Value()
	if oldValue != newValue {
		m.pendingQuery = newValue
		cmds = append(cmds, tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
			return filterTickMsg{query: newValue}
		}))
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// switchMode moves to mode and resets the per-mode view state. cached says the
// mode's data is already in the model; otherwise load is returned so Bubble Tea
// fetches it and the *LoadedMsg handler finishes the switch.
func (m *model) switchMode(mode SearchMode, cached bool, load tea.Cmd) tea.Cmd {
	if m.mode == mode {
		return nil
	}

	m.mode = mode
	m.input.SetValue("")
	m.previewCache = make(map[string]string)
	m.lastPreviewKey = ""

	if cached {
		m.loading = false
		m.loadItemsForMode()
		m.updateFilter("")
		m.resetCursorToBottom()
		m.updatePreview()
		return nil
	}

	m.loading = true
	m.filtered = nil
	m.allItems = nil
	m.allItemsStr = nil
	m.cursor = 0
	m.offset = 0
	return load
}

// switchToGitBranchMode switches to git branch mode (Ctrl+G)
func (m *model) switchToGitBranchMode() tea.Cmd {
	return m.switchMode(ModeGitBranch, len(m.gitBranches) > 0, loadBranchesCmd())
}

// switchToHistoryMode switches directly to history mode (Ctrl+R). The initial
// load may still be in flight, in which case its result no longer reaches this
// mode on its own, so an empty cache starts a fresh load.
func (m *model) switchToHistoryMode() tea.Cmd {
	return m.switchMode(ModeHistory, len(m.historyEntries) > 0, loadHistoryCmd())
}

// switchToFilesMode switches to files mode (Ctrl+S)
func (m *model) switchToFilesMode() tea.Cmd {
	return m.switchMode(ModeFiles, len(m.fileEntries) > 0, loadFilesCmd())
}

// switchToWorktreeMode switches to git worktree mode (Ctrl+W)
func (m *model) switchToWorktreeMode() tea.Cmd {
	return m.switchMode(ModeWorktree, len(m.worktrees) > 0, loadWorktreesCmd())
}

// switchToCommitMode switches to git commit mode (Ctrl+X)
func (m *model) switchToCommitMode() tea.Cmd {
	if m.mode != ModeCommit && !git.NewRepository(".").IsRepo() {
		m.statusMsg = "⚠ Not a git repository"
		return nil
	}
	return m.switchMode(ModeCommit, len(m.commits) > 0, loadCommitsCmd())
}

// updateActionPicker handles keys while the commit action picker is open.
func (m model) updateActionPicker(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		res := m.pendingCommit
		if tmpl := commitActions[m.actionCursor].Template; tmpl != "" {
			res = tmpl + " " + res
		}
		m.choice = &res
		m.commitIsCmd = commitActions[m.actionCursor].Template != ""
		m.quitting = true
		return m, tea.Quit
	case "esc", "ctrl+c":
		m.pendingCommit = ""
		return m, nil
	case "down", "ctrl+n":
		if m.actionCursor < len(commitActions)-1 {
			m.actionCursor++
		}
	case "up", "ctrl+p":
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	}
	return m, nil
}

// completeSelectedItem fills the input field with the currently selected item's text
func (m *model) completeSelectedItem() {
	text := m.filtered[m.cursor].Text
	m.input.SetValue(text)
	m.input.CursorEnd()
	m.updateFilter(text)
}

// resetCursorToBottom resets the cursor to the bottom of the list
func (m *model) resetCursorToBottom() {
	if len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
		m.offset = m.cursor - m.mainHeight + 1
		if m.offset < 0 {
			m.offset = 0
		}
	} else {
		m.cursor = 0
		m.offset = 0
	}
}

// validateCursor ensures the cursor is within valid bounds
func (m *model) validateCursor() {
	if len(m.filtered) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	if m.offset < 0 {
		m.offset = 0
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.mainHeight {
		m.offset = m.cursor - m.mainHeight + 1
		if m.offset < 0 {
			m.offset = 0
		}
	}
}

// previewKey identifies the item a preview was rendered for. History previews
// include surrounding commands, so they also depend on the item's position in
// the source slice.
func previewKey(mode SearchMode, item Item) string {
	return strconv.Itoa(int(mode)) + "\x00" + strconv.Itoa(item.Index) + "\x00" + item.Text
}

// updatePreview updates the preview pane content
func (m *model) updatePreview() {
	if len(m.filtered) == 0 {
		m.viewport.SetContent("")
		m.lastPreviewKey = ""
		return
	}

	item := m.filtered[m.cursor]

	// Skip update only when the same item is still selected. Keying this on the
	// cursor position alone kept a stale preview whenever filtering changed the
	// item sitting at that position.
	key := previewKey(m.mode, item)
	if key == m.lastPreviewKey {
		return
	}
	m.lastPreviewKey = key

	width, height := m.viewport.Width(), m.viewport.Height()

	var content string
	switch m.mode {
	case ModeHistory:
		entry := item.Original.(history.Entry)
		content = entry.GeneratePreview(m.historyEntries, item.Index, width, height)
	case ModeGitBranch:
		branch := item.Original.(git.Branch)
		content = m.cachedPreview(branch.Name, func() string { return branch.GeneratePreview(width, height) })
	case ModeFiles:
		entry := item.Original.(files.Entry)
		content = m.cachedPreview(entry.Path, func() string { return entry.GeneratePreview(width, height) })
	case ModeCommit:
		c := item.Original.(git.Commit)
		content = m.cachedPreview(c.Hash, func() string { return c.GeneratePreview(".", width, height) })
	case ModeWorktree:
		wt := item.Original.(git.Worktree)
		content = m.cachedPreview(wt.Path, func() string { return wt.GeneratePreview(width, height) })
	}
	m.viewport.SetContent(content)
}

// cachedPreview returns the stored render for key, generating and storing it on
// a miss. History previews are deliberately not routed through here: they depend
// on the entry's neighbours, not on the entry alone.
func (m *model) cachedPreview(key string, generate func() string) string {
	if cached, ok := m.previewCache[key]; ok {
		return cached
	}
	content := generate()
	m.previewCache[key] = content
	return content
}

// selectItem handles item selection
func (m *model) selectItem() {
	item := m.filtered[m.cursor]
	switch m.mode {
	case ModeHistory:
		res := item.Text
		m.choice = &res
	case ModeGitBranch:
		branch := item.Original.(git.Branch)
		res := branch.Name
		if branch.IsRemote {
			parts := strings.SplitN(res, "/", 2)
			if len(parts) == 2 {
				res = parts[1]
			}
		}
		m.choice = &res
	case ModeFiles:
		entry, ok := item.Original.(files.Entry)
		if ok {
			res := entry.Path
			m.choice = &res
			m.choiceIsDir = entry.IsDir
		}
	case ModeWorktree:
		wt, ok := item.Original.(git.Worktree)
		if ok {
			res := wt.Path
			m.choice = &res
		}
	}
}
