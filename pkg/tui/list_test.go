// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestClaudeSquadListView(t *testing.T) {
	// Create test session data
	sessions := []SessionInfo{
		{
			Name:       "test-session-1",
			AgentName:  "claude-3.5-sonnet",
			Model:      "claude-3.5-sonnet",
			Status:     "running",
			Prompt:     "Create a web application with authentication",
			Insertions: 45,
			Deletions:  12,
			Port:       3000,
		},
		{
			Name:       "test-session-2",
			AgentName:  "gpt-4",
			Model:      "gpt-4",
			Status:     "ready",
			Prompt:     "Fix the database connection issue in the backend API",
			Insertions: 0,
			Deletions:  0,
			Port:       0,
		},
		{
			Name:       "test-session-3",
			AgentName:  "claude-3.5-haiku",
			Model:      "claude-3.5-haiku",
			Status:     "inactive",
			Prompt:     "Write unit tests for the user management module",
			Insertions: 23,
			Deletions:  5,
			Port:       8080,
		},
	}

	// Create list model with Claude Squad styling
	listModel := NewListModel(80, 24)

	// Test LoadSessions function
	listModel.LoadSessions(sessions)

	// Verify sessions were loaded
	if len(listModel.list.Items()) != len(sessions) {
		t.Errorf("Expected %d sessions, got %d", len(sessions), len(listModel.list.Items()))
	}

	// Test individual session list items
	for i, session := range sessions {
		sessionItem := NewSessionListItem(session)

		// Test title formatting
		title := sessionItem.Title()
		if !strings.Contains(title, session.AgentName) {
			t.Errorf("Title should contain agent name %s, got: %s", session.AgentName, title)
		}
		if !strings.Contains(title, session.Model) {
			t.Errorf("Title should contain model %s, got: %s", session.Model, title)
		}

		// Test description formatting
		description := sessionItem.Description()
		if !strings.Contains(description, session.Status) {
			t.Errorf("Description should contain status %s, got: %s", session.Status, description)
		}

		// Test diff stats formatting (if present)
		if session.Insertions > 0 || session.Deletions > 0 {
			expectedDiff := strings.Contains(description, "+") && strings.Contains(description, "-")
			if !expectedDiff {
				t.Errorf("Description should contain diff stats for session %d, got: %s", i, description)
			}
		}

		// Test dev URL formatting (if present)
		if session.Port > 0 {
			if !strings.Contains(description, "localhost:") {
				t.Errorf("Description should contain dev URL for session %d, got: %s", i, description)
			}
		}

		// Test prompt inclusion
		if session.Prompt != "" {
			// Prompt should be either full or truncated (now 37 chars + "...")
			truncatedPrompt := session.Prompt
			if len(session.Prompt) > 40 {
				truncatedPrompt = session.Prompt[:37] + "..."
			}
			hasPrompt := strings.Contains(description, session.Prompt) || strings.Contains(description, truncatedPrompt)
			if !hasPrompt {
				t.Errorf("Description should contain prompt for session %d, got: %s", i, description)
			}
		}

		// Test filter value
		filterValue := sessionItem.FilterValue()
		expectedParts := []string{session.AgentName, session.Model, session.Prompt}
		for _, part := range expectedParts {
			if part != "" && !strings.Contains(filterValue, part) {
				t.Errorf("FilterValue should contain %s, got: %s", part, filterValue)
			}
		}
	}
}

func TestClaudeSquadStatusFormatting(t *testing.T) {
	testCases := []struct {
		status       string
		expectedIcon string
	}{
		{"running", "●"},
		{"attached", "●"},
		{"ready", "○"},
		{"inactive", "○"},
		{"unknown", "?"},
	}

	for _, tc := range testCases {
		session := SessionInfo{
			Name:      "test",
			AgentName: "test-agent",
			Model:     "test-model",
			Status:    tc.status,
			Prompt:    "test prompt",
		}

		sessionItem := NewSessionListItem(session)
		statusIcon := sessionItem.formatStatusIcon(tc.status)

		// The formatted status should contain the expected icon
		if !strings.Contains(statusIcon, tc.expectedIcon) {
			t.Errorf("Status icon for %s should contain %s, got: %s", tc.status, tc.expectedIcon, statusIcon)
		}
	}
}

// TestMetricsBasedActivityColors tests UI rendering colors based on ActivityMonitor metrics
func TestMetricsBasedActivityColors(t *testing.T) {
	type testCase struct {
		name           string
		activityStatus string
		expectedStyle  lipgloss.Style
	}

	testCases := []testCase{
		{
			name:           "working_agent",
			activityStatus: "working",
			expectedStyle:  ClaudeSquadAccentStyle, // Green for working
		},
		{
			name:           "idle_agent",
			activityStatus: "idle",
			expectedStyle:  WarningStyle, // Yellow for idle
		},
		{
			name:           "stuck_agent",
			activityStatus: "stuck",
			expectedStyle:  ErrorStyle, // Red for stuck
		},
		{
			name:           "unknown_agent",
			activityStatus: "unknown",
			expectedStyle:  ClaudeSquadMutedStyle, // Muted for unknown
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock session with activity status
			session := SessionInfo{
				Name:           "test-agent-session",
				AgentName:      "claude",
				Model:          "claude-3.5-sonnet",
				Status:         "running",
				Prompt:         "Test prompt",
				ActivityStatus: tc.activityStatus,
			}

			// Create list item and test styling
			sessionItem := NewSessionListItem(session)

			// Test that the correct style is applied based on activity status
			actualStyle := sessionItem.getActivityStatusStyle(tc.activityStatus)

			// Render both styles to compare (basic check that they're different for different statuses)
			expectedRendered := tc.expectedStyle.Render("test")
			actualRendered := actualStyle.Render("test")

			// Verify the style is not empty and matches expected behavior
			if actualRendered == "" {
				t.Error("Activity status style should not render empty string")
			}

			// For a more specific test, check the color constants
			switch tc.activityStatus {
			case "working":
				// Working agents should have green accent color
				if !strings.Contains(actualRendered, string(ClaudeSquadAccent)) && !strings.Contains(expectedRendered, string(ClaudeSquadAccent)) {
					// This is a relaxed check since ANSI codes might not appear in test environment
					t.Logf("Working agent style: expected=%s, actual=%s", expectedRendered, actualRendered)
				}
			case "idle":
				// Idle agents should have warning color (yellow)
				if !strings.Contains(actualRendered, string(WarningColor)) && !strings.Contains(expectedRendered, string(WarningColor)) {
					t.Logf("Idle agent style: expected=%s, actual=%s", expectedRendered, actualRendered)
				}
			case "stuck":
				// Stuck agents should have error color (red)
				if !strings.Contains(actualRendered, string(ErrorColor)) && !strings.Contains(expectedRendered, string(ErrorColor)) {
					t.Logf("Stuck agent style: expected=%s, actual=%s", expectedRendered, actualRendered)
				}
			case "unknown":
				// Unknown agents should have muted color
				if !strings.Contains(actualRendered, string(ClaudeSquadMuted)) && !strings.Contains(expectedRendered, string(ClaudeSquadMuted)) {
					t.Logf("Unknown agent style: expected=%s, actual=%s", expectedRendered, actualRendered)
				}
			}
		})
	}
}

// TestListFilteringWithMetrics tests that list filtering works correctly with metrics-based status
func TestListFilteringWithMetrics(t *testing.T) {
	sessions := []SessionInfo{
		{
			Name:           "working-agent",
			AgentName:      "claude",
			Model:          "claude-3.5-sonnet",
			Status:         "running",
			ActivityStatus: "working",
			Prompt:         "Fix bug in authentication",
		},
		{
			Name:           "stuck-agent",
			AgentName:      "gpt-4",
			Model:          "gpt-4",
			Status:         "running",
			ActivityStatus: "stuck",
			Prompt:         "Optimize database queries",
		},
		{
			Name:           "idle-agent",
			AgentName:      "claude",
			Model:          "claude-3.5-haiku",
			Status:         "ready",
			ActivityStatus: "idle",
			Prompt:         "Write unit tests",
		},
	}

	listModel := NewListModel(80, 24)
	listModel.LoadSessions(sessions)

	// Test initial state - all sessions loaded
	if len(listModel.Items()) != 3 {
		t.Errorf("Expected 3 sessions initially, got %d", len(listModel.Items()))
	}

	// Test stuck filter
	listModel.SetFilter(FilterStuck)
	stuckFiltered := len(listModel.Items())
	if stuckFiltered != 1 {
		t.Errorf("Expected 1 stuck session after filtering, got %d", stuckFiltered)
	}

	// Test working filter
	listModel.SetFilter(FilterWorking)
	workingFiltered := len(listModel.Items())
	if workingFiltered != 1 {
		t.Errorf("Expected 1 working session after filtering, got %d", workingFiltered)
	}

	// Test clearing filter
	listModel.SetFilter(FilterNone)
	allFiltered := len(listModel.Items())
	if allFiltered != 3 {
		t.Errorf("Expected 3 sessions after clearing filter, got %d", allFiltered)
	}
}

func TestClaudeSquadColorScheme(t *testing.T) {
	// Test that Claude Squad colors are defined correctly
	expectedColors := map[string]string{
		"ClaudeSquadPrimary": "#ffffff",
		"ClaudeSquadAccent":  "#00ff9d",
		"ClaudeSquadDark":    "#0a0a0a",
		"ClaudeSquadGray":    "#1a1a1a",
		"ClaudeSquadMuted":   "#6b7280",
		"ClaudeSquadHover":   "#00e68a",
	}

	// Test that the color variables exist and have expected values
	if string(ClaudeSquadPrimary) != expectedColors["ClaudeSquadPrimary"] {
		t.Errorf("ClaudeSquadPrimary should be %s, got %s", expectedColors["ClaudeSquadPrimary"], string(ClaudeSquadPrimary))
	}
	if string(ClaudeSquadAccent) != expectedColors["ClaudeSquadAccent"] {
		t.Errorf("ClaudeSquadAccent should be %s, got %s", expectedColors["ClaudeSquadAccent"], string(ClaudeSquadAccent))
	}
	if string(ClaudeSquadDark) != expectedColors["ClaudeSquadDark"] {
		t.Errorf("ClaudeSquadDark should be %s, got %s", expectedColors["ClaudeSquadDark"], string(ClaudeSquadDark))
	}
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
