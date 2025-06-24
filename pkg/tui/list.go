// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// SessionListItem represents a session in the TUI list with Claude Squad styling
type SessionListItem struct {
	session SessionInfo
}

// NewSessionListItem creates a new session list item
func NewSessionListItem(session SessionInfo) SessionListItem {
	return SessionListItem{
		session: session,
	}
}

// Title implements list.Item interface for sessions
func (s SessionListItem) Title() string {
	// Get status icon using Claude Squad styling
	statusIcon := s.formatStatusIcon(s.session.Status)
	
	// Format: [●] agent-name (model)
	return fmt.Sprintf("%s %s %s", 
		statusIcon,
		s.session.AgentName,
		ClaudeSquadAccentStyle.Render(fmt.Sprintf("(%s)", s.session.Model)))
}

// Description implements list.Item interface for sessions
func (s SessionListItem) Description() string {
	// Build description with status, diff stats, dev URL, and prompt
	var parts []string
	
	// Status with Claude Squad colors
	status := s.formatStatus(s.session.Status)
	parts = append(parts, status)
	
	// Git diff stats with Claude Squad green accent
	if s.session.Insertions > 0 || s.session.Deletions > 0 {
		diffStats := fmt.Sprintf("+%d/-%d", s.session.Insertions, s.session.Deletions)
		parts = append(parts, ClaudeSquadAccentStyle.Render(diffStats))
	}
	
	// Dev server URL with Claude Squad accent
	if s.session.Port > 0 {
		devURL := fmt.Sprintf("localhost:%d", s.session.Port)
		parts = append(parts, ClaudeSquadAccentStyle.Render(devURL))
	}
	
	// Truncated prompt with muted styling
	prompt := s.session.Prompt
	if len(prompt) > 50 {
		prompt = prompt[:47] + "..."
	}
	if prompt != "" {
		parts = append(parts, ClaudeSquadMutedStyle.Render(prompt))
	}
	
	return strings.Join(parts, " │ ")
}

// FilterValue implements list.Item interface for sessions
func (s SessionListItem) FilterValue() string {
	return s.session.AgentName + " " + s.session.Model + " " + s.session.Prompt
}

// formatStatusIcon returns a styled status icon using Claude Squad colors
func (s SessionListItem) formatStatusIcon(status string) string {
	switch status {
	case "attached":
		return ClaudeSquadAccentStyle.Render("●") // Claude Squad green
	case "running":
		return ClaudeSquadAccentStyle.Render("●") // Claude Squad green 
	case "ready":
		return ClaudeSquadAccentStyle.Render("○") // Claude Squad green outline
	case "inactive":
		return ClaudeSquadMutedStyle.Render("○")  // Muted gray
	default:
		return ClaudeSquadMutedStyle.Render("?")
	}
}

// formatStatus returns a styled status string using Claude Squad colors
func (s SessionListItem) formatStatus(status string) string {
	switch status {
	case "attached":
		return ClaudeSquadAccentStyle.Render("attached")
	case "running":
		return ClaudeSquadAccentStyle.Render("running")
	case "ready":
		return ClaudeSquadPrimaryStyle.Render("ready")
	case "inactive":
		return ClaudeSquadMutedStyle.Render("inactive")
	default:
		return ClaudeSquadMutedStyle.Render(status)
	}
}

// ListModel wraps the bubbles list component with Claude Squad styling
type ListModel struct {
	list    list.Model
	width   int
	height  int
}

// NewListModel creates a new list model with Claude Squad styling
func NewListModel(width, height int) ListModel {
	// Create custom delegate for Claude Squad styling
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = ClaudeSquadSelectedStyle
	delegate.Styles.SelectedDesc = ClaudeSquadSelectedDescStyle
	delegate.Styles.NormalTitle = ClaudeSquadNormalTitleStyle
	delegate.Styles.NormalDesc = ClaudeSquadNormalDescStyle
	delegate.Styles.DimmedTitle = ClaudeSquadMutedStyle
	delegate.Styles.DimmedDesc = ClaudeSquadMutedStyle

	// Create list with custom delegate
	l := list.New([]list.Item{}, delegate, width, height)
	l.Title = "Agent Sessions"
	l.Styles.Title = ClaudeSquadHeaderStyle
	l.Styles.TitleBar = ClaudeSquadHeaderBarStyle
	
	// Customize the empty state message
	l.SetShowStatusBar(false) // Hide the status bar to prevent double messages
	l.SetShowPagination(false) // Hide pagination for cleaner look when few items
	
	return ListModel{
		list:   l,
		width:  width,
		height: height,
	}
}

// LoadSessions loads session information and renders each row with agent name, status icon, diff stats, and dev URL
func (m *ListModel) LoadSessions(sessions []SessionInfo) {
	// Convert SessionInfo slice to list.Item slice
	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		items[i] = NewSessionListItem(session)
	}
	
	// Update the list with new items
	m.list.SetItems(items)
}

// SetSize updates the dimensions of the list
func (m *ListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// SelectedSession returns the currently selected session, if any
func (m ListModel) SelectedSession() *SessionInfo {
	if item := m.list.SelectedItem(); item != nil {
		if sessionItem, ok := item.(SessionListItem); ok {
			return &sessionItem.session
		}
	}
	return nil
}

// Init implements tea.Model interface
func (m ListModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model interface
func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle additional key events here if needed
		switch msg.String() {
		case "enter":
			// Could emit a custom message for session selection
			if selected := m.SelectedSession(); selected != nil {
				// Return a command to handle session selection
				return m, tea.Sequence(tea.Printf("Selected: %s", selected.Name))
			}
		}
	case tea.WindowSizeMsg:
		// Update size when window changes
		m.SetSize(msg.Width, msg.Height-4) // Account for header/footer
	}
	
	var cmd tea.Cmd
	listModel, cmd := m.list.Update(msg)
	m.list = listModel
	return m, cmd
}

// View implements tea.Model interface
func (m ListModel) View() string {
	// Check if list is empty and provide custom empty state
	if len(m.list.Items()) == 0 {
		emptyMessage := ClaudeSquadMutedStyle.Render("No active agent sessions")
		headerView := ClaudeSquadHeaderStyle.Render("Agent Sessions")
		
		// Calculate padding to center the message
		maxWidth := m.width - 4 // Account for border padding
		padding := (maxWidth - len("No active agent sessions")) / 2
		if padding < 0 {
			padding = 0
		}
		
		emptyView := fmt.Sprintf("%s\n\n%s%s", 
			headerView,
			strings.Repeat(" ", padding),
			emptyMessage)
		
		return ClaudeSquadBorderStyle.Render(emptyView)
	}
	
	return ClaudeSquadBorderStyle.Render(m.list.View())
}
