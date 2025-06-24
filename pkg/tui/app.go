// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
)

// RefreshMsg is sent by the ticker to refresh sessions without clearing screen
type RefreshMsg struct{}

// TickMsg wraps time.Time for ticker messages
type TickMsg time.Time

// App represents the main TUI application
type App struct {
	uzi      UziInterface
	list     *ListModel
	keys     KeyMap
	ticker   *time.Ticker
	width    int
	height   int
	loading  bool
}

// NewApp creates a new TUI application instance
func NewApp(uzi UziInterface) *App {
	// Initialize the list view
	list := NewListModel(80, 24) // Default size, will be updated on first render
	
	return &App{
		uzi:     uzi,
		list:    &list,
		keys:    DefaultKeyMap(),
		ticker:  nil, // Will be created in Init
		loading: true,
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key events
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
			
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
		}
		
		// Delegate other key events to the list
		var cmd tea.Cmd
		model, cmd := a.list.Update(msg)
		if listModel, ok := model.(ListModel); ok {
			*a.list = listModel
		}
		return a, cmd
		
	case tea.WindowSizeMsg:
		// Update dimensions
		a.width = msg.Width
		a.height = msg.Height
		
		// Update list size (leave space for potential status bar)
		a.list.SetSize(msg.Width, msg.Height-2)
		
		// Delegate to list for its own size handling
		var cmd tea.Cmd
		model, cmd := a.list.Update(msg)
		if listModel, ok := model.(ListModel); ok {
			*a.list = listModel
		}
		return a, cmd
		
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
		// Delegate other messages to the list
		var cmd tea.Cmd
		model, cmd := a.list.Update(msg)
		if listModel, ok := model.(ListModel); ok {
			*a.list = listModel
		}
		return a, cmd
	}
}

// View implements tea.Model interface - delegates to list view
func (a *App) View() string {
	// If we don't have proper dimensions yet, return a simple message
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}
	
	// Delegate to the list view for rendering
	listView := a.list.View()
	
	// Add a subtle status line if loading
	if a.loading {
		statusLine := ClaudeSquadMutedStyle.Render("Refreshing sessions...")
		return listView + "\n" + statusLine
	}
	
	return listView
}
