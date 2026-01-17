package app

import (
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
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

		inputHeight := 3
		mainHeight := msg.Height - inputHeight
		if mainHeight < 0 {
			mainHeight = 0
		}
		m.mainHeight = mainHeight

		listWidth := int(float64(msg.Width) * 0.6) // 60% split for list, 40% for preview

		previewWidth := msg.Width - listWidth - 2
		m.listWidth = listWidth

		m.viewport.Width = previewWidth
		m.viewport.Height = mainHeight

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
		case tea.KeyCtrlR:
			// Toggle mode
			m.toggleMode()
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

// toggleMode toggles between history and git branch mode
func (m *model) toggleMode() {
	var newMode SearchMode
	if m.mode == ModeHistory {
		newMode = ModeGitBranch
		// Check if git repo
		if len(m.gitBranches) == 0 {
			// Try to collect branches if not already done?
			// But maybe we should have collected at start if possible.
			// Or check isGitRepo here.
			// Since we want seamless switching, we can try to fetch now.
			if !git.IsGitRepo() {
				// Cannot switch to git mode
				// Maybe show a flash message? For now just ignore or print to debug.
				return
			}
			branches := git.CollectBranches()
			if len(branches) == 0 {
				return
			}
			m.gitBranches = branches
		}
	} else {
		newMode = ModeHistory
	}

	m.mode = newMode
	m.input.SetValue("") // Clear input on switch

	// Update placeholder
	if m.mode == ModeHistory {
		m.input.Placeholder = "Search history... (Ctrl+R to switch)"
	} else {
		m.input.Placeholder = "Search branches... (Ctrl+R to switch)"
	}

	m.loadItemsForMode()
	m.updateFilter("")

	// Reset cursor to bottom explicitly after mode switch
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
	m.updatePreview()
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
		return
	}
	item := m.filtered[m.cursor]

	var content string
	if m.mode == ModeHistory {
		entry := item.Original.(history.Entry)
		// We need original index in historyEntries (which is Newest->Oldest).
		// Item.Index is index in allItems (Oldest->Newest).
		// But GeneratePreview expects index in the slice passed to it?
		// No, GeneratePreview takes (entry, allEntries, index, ...)
		// The index is used for Context.
		// We passed m.historyEntries (Newest->Oldest).
		// Item.Index is the index in m.historyEntries.

		content = history.GeneratePreview(entry, m.historyEntries, item.Index, m.viewport.Width, m.viewport.Height)
	} else {
		branch := item.Original.(git.Branch)
		content = git.GeneratePreview(branch, m.viewport.Width, m.viewport.Height)
	}
	m.viewport.SetContent(content)
}

// selectItem handles item selection
func (m *model) selectItem() {
	item := m.filtered[m.cursor]
	if m.mode == ModeHistory {
		res := item.Text
		m.choice = &res
	} else {
		branch := item.Original.(git.Branch)
		res := branch.Name
		if branch.IsRemote {
			parts := strings.SplitN(res, "/", 2)
			if len(parts) == 2 {
				res = parts[1]
			}
		}
		m.choice = &res
	}
}
