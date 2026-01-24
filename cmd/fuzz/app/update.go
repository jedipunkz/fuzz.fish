package app

import (
	"os"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/files"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/git"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/history"
)

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
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

		m.viewport.Width = previewWidth
		m.viewport.Height = m.mainHeight

		// Recalculate offset to keep cursor at the bottom of the view
		if len(m.filtered) > 0 {
			m.offset = m.cursor - m.mainHeight + 1
			if m.offset < 0 {
				m.offset = 0
			}
		}

		m.validateCursor()
		m.updatePreview()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				m.selectItem()
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlY:
			if len(m.filtered) > 0 {
				_ = clipboard.WriteAll(m.filtered[m.cursor].Text)
				m.quitting = true
				return m, tea.Quit
			}
		case tea.KeyCtrlG:
			// Switch to GitBranch mode
			m.switchToGitBranchMode()
			return m, nil
		case tea.KeyCtrlF:
			// Switch to Files mode
			m.switchToFilesMode()
			return m, nil
		case tea.KeyCtrlR:
			// Go back to history mode
			m.switchToHistoryMode()
			return m, nil
		}
		switch msg.String() {
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
		m.updateFilter(newValue)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// switchToGitBranchMode switches to git branch mode (Ctrl+G)
func (m *model) switchToGitBranchMode() {
	if m.mode == ModeGitBranch {
		// Already in git branch mode, nothing to do
		return
	}

	// Check if git repo
	if len(m.gitBranches) == 0 {
		if !git.IsGitRepo() {
			return
		}
		branches := git.CollectBranches()
		if len(branches) == 0 {
			return
		}
		m.gitBranches = branches
	}

	m.mode = ModeGitBranch
	m.input.SetValue("") // Clear input on switch

	// Update placeholder
	m.updatePlaceholder()

	// Reset preview state when switching modes
	m.lastPreviewIndex = -1

	m.loadItemsForMode()
	m.updateFilter("")

	// Reset cursor to bottom explicitly after mode switch
	m.resetCursorToBottom()
	m.updatePreview()
}

// switchToHistoryMode switches directly to history mode (Ctrl+R)
func (m *model) switchToHistoryMode() {
	if m.mode == ModeHistory {
		// Already in history mode, nothing to do
		return
	}

	m.mode = ModeHistory
	m.input.SetValue("") // Clear input on switch

	// Update placeholder
	m.updatePlaceholder()

	// Reset preview state when switching modes
	m.lastPreviewIndex = -1

	m.loadItemsForMode()
	m.updateFilter("")

	// Reset cursor to bottom explicitly after mode switch
	m.resetCursorToBottom()
	m.updatePreview()
}

// switchToFilesMode switches to files mode (Ctrl+F)
func (m *model) switchToFilesMode() {
	if m.mode == ModeFiles {
		// Already in files mode, nothing to do
		return
	}

	// Collect files from current directory if not already done
	if len(m.fileEntries) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return
		}
		entries := files.Collect(cwd)
		if len(entries) == 0 {
			return
		}
		m.fileEntries = entries
	}

	m.mode = ModeFiles
	m.input.SetValue("") // Clear input on switch

	// Update placeholder
	m.updatePlaceholder()

	// Clear preview cache and index when switching modes
	m.previewCache = make(map[string]string)
	m.lastPreviewIndex = -1

	m.loadItemsForMode()
	m.updateFilter("")

	// Reset cursor to bottom explicitly after mode switch
	m.resetCursorToBottom()
	m.updatePreview()
}

// updatePlaceholder updates the input placeholder based on current mode
func (m *model) updatePlaceholder() {
	switch m.mode {
	case ModeHistory:
		m.input.Placeholder = "Search history... (Ctrl+G: git, Ctrl+F: files)"
	case ModeGitBranch:
		m.input.Placeholder = "Search branches... (Ctrl+R: history, Ctrl+F: files)"
	case ModeFiles:
		m.input.Placeholder = "Search files... (Ctrl+R: history, Ctrl+G: git)"
	}
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

// updatePreview updates the preview pane content
func (m *model) updatePreview() {
	if len(m.filtered) == 0 {
		m.viewport.SetContent("")
		m.lastPreviewIndex = -1
		return
	}

	// Skip update if cursor hasn't moved
	if m.cursor == m.lastPreviewIndex {
		return
	}
	m.lastPreviewIndex = m.cursor

	item := m.filtered[m.cursor]

	var content string
	var cacheKey string

	switch m.mode {
	case ModeHistory:
		entry := item.Original.(history.Entry)
		content = history.GeneratePreview(entry, m.historyEntries, item.Index, m.viewport.Width, m.viewport.Height)
	case ModeGitBranch:
		branch := item.Original.(git.Branch)
		content = git.GeneratePreview(branch, m.viewport.Width, m.viewport.Height)
	case ModeFiles:
		entry := item.Original.(files.Entry)
		// Use cache for file previews
		cacheKey = entry.Path
		if cached, ok := m.previewCache[cacheKey]; ok {
			content = cached
		} else {
			content = files.GeneratePreview(entry, m.viewport.Width, m.viewport.Height)
			// Cache the result
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
	}
}
