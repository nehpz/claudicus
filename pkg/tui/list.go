// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	// Get status icon and activity bar using Claude Squad styling
	statusIcon := s.formatStatusIcon(s.session.Status)
	activityBar := s.formatActivityBar()

	// Format: [●] ▮▮▮ agent-name (model)
	return fmt.Sprintf("%s %s %s %s",
		statusIcon,
		activityBar,
		s.session.AgentName,
		ClaudeSquadAccentStyle.Render(fmt.Sprintf("(%s)", s.session.Model)))
}

// Description implements list.Item interface for sessions
func (s SessionListItem) Description() string {
	// Build description with status, diff stats, dev URL, last activity, and prompt
	var parts []string

	// Status with Claude Squad colors
	status := s.formatStatus(s.session.Status)
	parts = append(parts, status)

	// Git diff stats with Claude Squad green accent
	if s.session.Insertions > 0 || s.session.Deletions > 0 {
		diffStats := fmt.Sprintf("+%d/-%d", s.session.Insertions, s.session.Deletions)
		parts = append(parts, ClaudeSquadAccentStyle.Render(diffStats))
	}

	// Last activity time with muted styling
	if lastActivity := s.formatLastActivity(); lastActivity != "" {
		parts = append(parts, ClaudeSquadMutedStyle.Render(lastActivity))
	}

	// Dev server URL with Claude Squad accent
	if s.session.Port > 0 {
		devURL := fmt.Sprintf("localhost:%d", s.session.Port)
		parts = append(parts, ClaudeSquadAccentStyle.Render(devURL))
	}

	// Truncated prompt with muted styling
	prompt := s.session.Prompt
	if len(prompt) > 40 { // Reduced to make room for activity time
		prompt = prompt[:37] + "..."
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
		return ClaudeSquadMutedStyle.Render("○") // Muted gray
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

// getActivityStatus determines activity status based on last update time and diff stats
func (s SessionListItem) getActivityStatus() string {
	// Parse UpdatedAt timestamp, try multiple formats
	lastUpdate, err := time.Parse(time.RFC3339, s.session.UpdatedAt)
	if err != nil {
		// Fallback to simpler format
		if lastUpdate, err = time.Parse("2006-01-02T15:04:05Z", s.session.UpdatedAt); err != nil {
			// Fallback to CreatedAt if UpdatedAt is invalid
			if lastUpdate, err = time.Parse(time.RFC3339, s.session.CreatedAt); err != nil {
				if lastUpdate, err = time.Parse("2006-01-02T15:04:05Z", s.session.CreatedAt); err != nil {
					return "unknown"
				}
			}
		}
	}

	// Calculate time since last activity
	timeSince := time.Since(lastUpdate)

	// Activity classification rules:
	// 1. Recent activity (<=90s) OR has uncommitted changes = working
	// 2. No activity for >3 minutes AND no diffs = stuck
	// 3. Everything else = idle

	// If there are uncommitted changes, agent is working
	if s.session.Insertions > 0 || s.session.Deletions > 0 {
		return "working"
	}

	// If recent activity (within 90 seconds), agent is working
	if timeSince <= 90*time.Second {
		return "working"
	}

	// If no activity for >3 minutes and no diffs, agent is stuck
	if timeSince > 3*time.Minute {
		return "stuck"
	}

	// Otherwise, agent is idle
	return "idle"
}

// formatActivityBar returns a colored activity bar (▮▮▯ style)
func (s SessionListItem) formatActivityBar() string {
	activityStatus := s.getActivityStatus()

	switch activityStatus {
	case "working":
		// Green activity bar - fully active
		return ClaudeSquadAccentStyle.Render("▮▮▮")
	case "idle":
		// Yellow activity bar - some activity
		return WarningStyle.Render("▮▮▯")
	case "stuck":
		// Red activity bar - no progress
		return ErrorStyle.Render("▮▯▯")
	default:
		// Gray activity bar - unknown status
		return ClaudeSquadMutedStyle.Render("▯▯▯")
	}
}

// formatLastActivity returns a human-readable "time ago" string
func (s SessionListItem) formatLastActivity() string {
	// Parse UpdatedAt timestamp
	lastUpdate, err := time.Parse("2006-01-02T15:04:05Z", s.session.UpdatedAt)
	if err != nil {
		// Fallback to CreatedAt if UpdatedAt is invalid
		if lastUpdate, err = time.Parse("2006-01-02T15:04:05Z", s.session.CreatedAt); err != nil {
			return ""
		}
	}

	// Calculate time since last activity
	timeSince := time.Since(lastUpdate)

	// Format as human-readable duration
	if timeSince < time.Minute {
		return fmt.Sprintf("%d s ago", int(timeSince.Seconds()))
	} else if timeSince < time.Hour {
		return fmt.Sprintf("%d m ago", int(timeSince.Minutes()))
	} else if timeSince < 24*time.Hour {
		return fmt.Sprintf("%d h ago", int(timeSince.Hours()))
	} else {
		return fmt.Sprintf("%d d ago", int(timeSince.Hours()/24))
	}
}

// getActivityStatusStyle returns the appropriate style for the given activity status
// This method provides compatibility for test code that expects this interface
func (s SessionListItem) getActivityStatusStyle(activityStatus string) lipgloss.Style {
	switch activityStatus {
	case "working":
		return ClaudeSquadAccentStyle // Green for working
	case "idle":
		return WarningStyle // Yellow for idle
	case "stuck":
		return ErrorStyle // Red for stuck
	case "unknown":
		fallthrough
	default:
		return ClaudeSquadMutedStyle // Muted for unknown
	}
}

// FilterType represents the type of filter applied to the list
type FilterType int

const (
	FilterNone FilterType = iota
	FilterStuck
	FilterWorking
)

// ListModel wraps the bubbles list component with Claude Squad styling
type ListModel struct {
	list         list.Model
	width        int
	height       int
	allSessions  []SessionInfo // Store all sessions for filtering
	filterType   FilterType    // Current filter type
	stuckToggled bool          // Track if stuck filter is toggled on/off
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
	l.SetShowStatusBar(false)  // Hide the status bar to prevent double messages
	l.SetShowPagination(false) // Hide pagination for cleaner look when few items

	return ListModel{
		list:         l,
		width:        width,
		height:       height,
		allSessions:  []SessionInfo{},
		filterType:   FilterNone,
		stuckToggled: false,
	}
}

// LoadSessions loads session information and renders each row with agent name, status icon, diff stats, and dev URL
func (m *ListModel) LoadSessions(sessions []SessionInfo) {
	// Store all sessions for filtering
	m.allSessions = sessions

	// Apply current filter and update list
	m.applyFilter()
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

// ToggleStuckFilter toggles the stuck agents filter on/off
func (m *ListModel) ToggleStuckFilter() {
	if m.filterType == FilterStuck {
		// Turn off stuck filter
		m.filterType = FilterNone
		m.stuckToggled = false
	} else {
		// Turn on stuck filter
		m.filterType = FilterStuck
		m.stuckToggled = true
	}
	m.applyFilter()
}

// SetWorkingFilter sets the working agents filter
func (m *ListModel) SetWorkingFilter() {
	m.filterType = FilterWorking
	m.stuckToggled = false
	m.applyFilter()
}

// ClearFilter clears any active filter
func (m *ListModel) ClearFilter() {
	m.filterType = FilterNone
	m.stuckToggled = false
	m.applyFilter()
}

// GetFilterStatus returns a string describing the current filter status
func (m *ListModel) GetFilterStatus() string {
	switch m.filterType {
	case FilterStuck:
		return "Showing stuck agents only"
	case FilterWorking:
		return "Showing working agents only"
	default:
		return ""
	}
}

// applyFilter applies the current filter to sessions and updates the list
func (m *ListModel) applyFilter() {
	filteredSessions := m.filterSessions(m.allSessions)

	// Convert SessionInfo slice to list.Item slice
	items := make([]list.Item, len(filteredSessions))
	for i, session := range filteredSessions {
		items[i] = NewSessionListItem(session)
	}

	// Update the list with filtered items
	m.list.SetItems(items)
}

// filterSessions filters sessions based on the current filter type
func (m *ListModel) filterSessions(sessions []SessionInfo) []SessionInfo {
	if m.filterType == FilterNone {
		return sessions
	}

	var filtered []SessionInfo
	for _, session := range sessions {
		item := NewSessionListItem(session)
		activityStatus := item.getActivityStatus()

		switch m.filterType {
		case FilterStuck:
			if activityStatus == "stuck" {
				filtered = append(filtered, session)
			}
		case FilterWorking:
			if activityStatus == "working" {
				filtered = append(filtered, session)
			}
		}
	}

	return filtered
}

// Items returns the current list items for test compatibility
func (m *ListModel) Items() []list.Item {
	return m.list.Items()
}

// SetFilter sets the filter type for test compatibility
func (m *ListModel) SetFilter(filterType FilterType) {
	m.filterType = filterType
	m.applyFilter()
}
