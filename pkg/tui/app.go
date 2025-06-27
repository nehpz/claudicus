// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// RefreshMsg is sent by the ticker to refresh sessions without clearing screen
type RefreshMsg struct{}

// TickMsg wraps time.Time for ticker messages
type TickMsg time.Time

// App represents the main TUI application
type App struct {
	uzi           UziInterface
	list          *ListModel
	diffPreview   *DiffPreviewModel
	broadcastInput *BroadcastInputModel
	keys          KeyMap
	ticker        *time.Ticker
	width         int
	height        int
	loading       bool
	splitView     bool // Toggle between list-only and split view
}

// NewApp creates a new TUI application instance
func NewApp(uzi UziInterface) *App {
	// Initialize the list view
	list := NewListModel(80, 24) // Default size, will be updated on first render
	diffPreview := NewDiffPreviewModel(40, 24) // Default size, will be updated on first render
	broadcastInput := NewBroadcastInputModel()
	
	return &App{
		uzi:           uzi,
		list:          &list,
		diffPreview:   diffPreview,
		broadcastInput: broadcastInput,
		keys:          DefaultKeyMap(),
		ticker:        nil, // Will be created in Init
		loading:     true,
		splitView:   false, // Start in list view
	}
}

// tickEvery returns a command that sends TickMsg every duration
func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// refreshSessions returns a command that fetches sessions and sends RefreshMsg
func (a *App) refreshSessions() tea.Cmd {
	return func() tea.Msg {
		// Load sessions via UziInterface
		sessions, err := a.uzi.GetSessions()
		if err != nil {
			// For now, just return the refresh message even on error
			// In a production app, you might want to handle errors differently
			return RefreshMsg{}
		}
		
		// Update the list with new sessions
		a.list.LoadSessions(sessions)
		a.loading = false
		
		return RefreshMsg{}
	}
}

// Init implements tea.Model interface
func (a *App) Init() tea.Cmd {
	// Start the 2-second ticker and initial session load
	return tea.Batch(
		a.refreshSessions(), // Load sessions immediately
		tickEvery(2*time.Second), // Start ticker for smooth updates
	)
}

// Update implements tea.Model interface
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle broadcast input when active
		if a.broadcastInput.IsActive() {
			switch {
			case key.Matches(msg, a.keys.Enter):
				// Execute broadcast and deactivate input
				message := a.broadcastInput.Value()
				a.broadcastInput.SetActive(false)
				
				if message != "" {
					return a, func() tea.Msg {
						err := a.uzi.RunBroadcast(message)
						if err != nil {
							// Handle error - for now just continue
							return nil
						}
						// Refresh sessions after broadcast
						return RefreshMsg{}
					}
				}
				return a, nil
				
			case key.Matches(msg, a.keys.Escape):
				// Cancel broadcast input
				a.broadcastInput.SetActive(false)
				return a, nil
				
			default:
				// Delegate to broadcast input
				var cmd tea.Cmd
				a.broadcastInput, cmd = a.broadcastInput.Update(msg)
				return a, cmd
			}
		}
		
		// Handle key events
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
			
		case key.Matches(msg, a.keys.Tab):
			// Toggle between list view and split view
			a.splitView = !a.splitView
			
			// When entering split view, load diff for selected session
			if a.splitView {
				if selected := a.list.SelectedSession(); selected != nil {
					a.diffPreview.LoadDiff(selected)
				}
			}
			return a, nil
			
		case key.Matches(msg, a.keys.Refresh):
			// Manual refresh
			a.loading = true
			return a, a.refreshSessions()
			
		case key.Matches(msg, a.keys.Enter):
			// Handle session selection/attachment
			if selected := a.list.SelectedSession(); selected != nil {
				// Attach to the selected session
				return a, func() tea.Msg {
					err := a.uzi.AttachToSession(selected.Name)
					if err != nil {
						// Handle error - for now just continue
						return nil
					}
					return tea.Quit // Exit TUI after attaching
				}
			}
			
		case key.Matches(msg, a.keys.Kill):
			// Handle agent kill command
			if selected := a.list.SelectedSession(); selected != nil {
				// Kill the selected agent session
				return a, func() tea.Msg {
					err := a.uzi.KillSession(selected.Name)
					if err != nil {
						// Handle error - for now just continue
						return nil
					}
					// Refresh sessions after kill
					return RefreshMsg{}
				}
			}
			
		case key.Matches(msg, a.keys.Broadcast):
			// Activate broadcast input prompt
			a.broadcastInput.SetActive(true)
			a.broadcastInput.SetWidth(a.width)
			return a, nil
		}
		
		// In split view, handle navigation differently
		if a.splitView {
			// Track previous selection
			prevSelected := a.list.SelectedSession()
			
			// Delegate navigation to the list
			var cmd tea.Cmd
			model, cmd := a.list.Update(msg)
			if listModel, ok := model.(ListModel); ok {
				*a.list = listModel
			}
			cmds = append(cmds, cmd)
			
			// If selection changed, update diff view
			if newSelected := a.list.SelectedSession(); newSelected != nil {
				if prevSelected == nil || prevSelected.Name != newSelected.Name {
					a.diffPreview.LoadDiff(newSelected)
				}
			}
			
			return a, tea.Batch(cmds...)
		} else {
			// In list view, delegate all key events to the list
			var cmd tea.Cmd
			model, cmd := a.list.Update(msg)
			if listModel, ok := model.(ListModel); ok {
				*a.list = listModel
			}
			return a, cmd
		}
		
	case tea.WindowSizeMsg:
		// Update dimensions
		a.width = msg.Width
		a.height = msg.Height
		
		if a.splitView {
			// In split view, allocate space for both list and diff
			listWidth := msg.Width / 2
			diffWidth := msg.Width - listWidth
			
			a.list.SetSize(listWidth, msg.Height-2)
			a.diffPreview.SetSize(diffWidth, msg.Height-2)
		} else {
			// In list view, use full width
			a.list.SetSize(msg.Width, msg.Height-2)
		}
		
		// Delegate to components for their own size handling
		var cmd tea.Cmd
		model, cmd := a.list.Update(msg)
		if listModel, ok := model.(ListModel); ok {
			*a.list = listModel
		}
		cmds = append(cmds, cmd)
		
		if a.splitView {
			// DiffPreview doesn't need to handle its own updates
		}
		
		return a, tea.Batch(cmds...)
		
	case TickMsg:
		// Ticker fired - refresh sessions smoothly without clearing screen
		return a, tea.Batch(
			a.refreshSessions(), // Refresh session data
			tickEvery(2*time.Second), // Schedule next tick
		)
		
	case RefreshMsg:
		// Sessions have been refreshed - no action needed
		// The list has already been updated in refreshSessions()
		return a, nil
		
	default:
		// Delegate other messages to appropriate components
		var cmd tea.Cmd
		model, cmd := a.list.Update(msg)
		if listModel, ok := model.(ListModel); ok {
			*a.list = listModel
		}
		cmds = append(cmds, cmd)
		
		return a, tea.Batch(cmds...)
	}
}

// View implements tea.Model interface - delegates to list view
func (a *App) View() string {
	// If we don't have proper dimensions yet, return a simple message
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}
	
	if a.splitView {
		// Split view: show list on left and diff on right
		listView := a.list.View()
		diffView := a.diffPreview.View()
		
		// Join horizontally with Claude Squad styling
		splitContent := lipgloss.JoinHorizontal(lipgloss.Top, listView, diffView)
		
		// Add status line if loading
		if a.loading {
			statusLine := ClaudeSquadMutedStyle.Render("Refreshing sessions...")
			return splitContent + "\n" + statusLine
		}
		
		// Add broadcast input if active
		content := splitContent
		if a.broadcastInput.IsActive() {
			broadcastView := a.broadcastInput.View()
			content = lipgloss.JoinVertical(lipgloss.Left, content, broadcastView)
		}
		
		return content
	} else {
		// List view: delegate to the list view for rendering
		listView := a.list.View()
		
		// Add a subtle status line if loading
		if a.loading {
			statusLine := ClaudeSquadMutedStyle.Render("Refreshing sessions...")
			listView = listView + "\n" + statusLine
		}
		
		// Add broadcast input if active
		if a.broadcastInput.IsActive() {
			broadcastView := a.broadcastInput.View()
			listView = lipgloss.JoinVertical(lipgloss.Left, listView, broadcastView)
		}
		
		return listView
	}
}
