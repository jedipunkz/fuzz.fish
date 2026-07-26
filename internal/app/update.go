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

		m.validateCursor()
		m.updatePreview()

	case tea.KeyPressMsg:
		// Clear status message on any key press
		m.statusMsg = ""

		switch msg.String() {
		case "enter":
			if len(m.filtered) > 0 {
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

// switchToGitBranchMode switches to git branch mode (Ctrl+G)
func (m *model) switchToGitBranchMode() tea.Cmd {
	if m.mode == ModeGitBranch {
		return nil
	}

	m.mode = ModeGitBranch
	m.input.SetValue("")
	m.updatePlaceholder()
	m.previewCache = make(map[string]string)
	m.lastPreviewKey = ""

	if len(m.gitBranches) > 0 {
		m.loadItemsForMode()
		m.updateFilter("")
		m.resetCursorToBottom()
		m.updatePreview()
		return nil
	}

	// Async load branches
	m.loading = true
	m.filtered = nil
	m.allItems = nil
	m.allItemsStr = nil
	m.cursor = 0
	m.offset = 0
	return loadBranchesCmd()
}

// switchToHistoryMode switches directly to history mode (Ctrl+R)
func (m *model) switchToHistoryMode() tea.Cmd {
	if m.mode == ModeHistory {
		return nil
	}

	m.mode = ModeHistory
	m.input.SetValue("")
	m.updatePlaceholder()
	m.previewCache = make(map[string]string)
	m.lastPreviewKey = ""

	if len(m.historyEntries) > 0 {
		m.loadItemsForMode()
		m.updateFilter("")
		m.resetCursorToBottom()
		m.updatePreview()
		return nil
	}

	// Async load history: the initial load may still be in flight, in which
	// case its result no longer reaches this mode on its own.
	m.loading = true
	m.filtered = nil
	m.allItems = nil
	m.allItemsStr = nil
	m.cursor = 0
	m.offset = 0
	return loadHistoryCmd()
}

// switchToFilesMode switches to files mode (Ctrl+S)
func (m *model) switchToFilesMode() tea.Cmd {
	if m.mode == ModeFiles {
		return nil
	}

	m.mode = ModeFiles
	m.input.SetValue("")
	m.updatePlaceholder()
	m.previewCache = make(map[string]string)
	m.lastPreviewKey = ""

	if len(m.fileEntries) > 0 {
		m.loadItemsForMode()
		m.updateFilter("")
		m.resetCursorToBottom()
		m.updatePreview()
		return nil
	}

	// Async load files
	m.loading = true
	m.filtered = nil
	m.allItems = nil
	m.allItemsStr = nil
	m.cursor = 0
	m.offset = 0
	return loadFilesCmd()
}

// switchToWorktreeMode switches to git worktree mode (Ctrl+W)
func (m *model) switchToWorktreeMode() tea.Cmd {
	if m.mode == ModeWorktree {
		return nil
	}

	m.mode = ModeWorktree
	m.input.SetValue("")
	m.updatePlaceholder()
	m.previewCache = make(map[string]string)
	m.lastPreviewKey = ""

	if len(m.worktrees) > 0 {
		m.loadItemsForMode()
		m.updateFilter("")
		m.resetCursorToBottom()
		m.updatePreview()
		return nil
	}

	// Async load worktrees
	m.loading = true
	m.filtered = nil
	m.allItems = nil
	m.allItemsStr = nil
	m.cursor = 0
	m.offset = 0
	return loadWorktreesCmd()
}

// updatePlaceholder updates the input placeholder based on current mode
func (m *model) updatePlaceholder() {
	switch m.mode {
	case ModeHistory:
		m.input.Placeholder = ""
	case ModeGitBranch:
		m.input.Placeholder = ""
	case ModeFiles:
		m.input.Placeholder = ""
	case ModeWorktree:
		m.input.Placeholder = ""
	}
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

	var content string
	var cacheKey string

	switch m.mode {
	case ModeHistory:
		entry := item.Original.(history.Entry)
		content = entry.GeneratePreview(m.historyEntries, item.Index, m.viewport.Width(), m.viewport.Height())
	case ModeGitBranch:
		branch := item.Original.(git.Branch)
		cacheKey = branch.Name
		if cached, ok := m.previewCache[cacheKey]; ok {
			content = cached
		} else {
			content = branch.GeneratePreview(m.viewport.Width(), m.viewport.Height())
			m.previewCache[cacheKey] = content
		}
	case ModeFiles:
		entry := item.Original.(files.Entry)
		cacheKey = entry.Path
		if cached, ok := m.previewCache[cacheKey]; ok {
			content = cached
		} else {
			content = entry.GeneratePreview(m.viewport.Width(), m.viewport.Height())
			m.previewCache[cacheKey] = content
		}
	case ModeWorktree:
		wt := item.Original.(git.Worktree)
		cacheKey = wt.Path
		if cached, ok := m.previewCache[cacheKey]; ok {
			content = cached
		} else {
			content = wt.GeneratePreview(m.viewport.Width(), m.viewport.Height())
			m.previewCache[cacheKey] = content
		}
	}
	m.viewport.SetContent(content)
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
