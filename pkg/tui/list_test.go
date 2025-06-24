// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"strings"
	"testing"
)

func TestClaudeSquadListView(t *testing.T) {
	// Create test session data
	sessions := []SessionInfo{
		{
			Name:        "test-session-1",
			AgentName:   "claude-3.5-sonnet",
			Model:       "claude-3.5-sonnet",
			Status:      "running",
			Prompt:      "Create a web application with authentication",
			Insertions:  45,
			Deletions:   12,
			Port:        3000,
		},
		{
			Name:        "test-session-2",
			AgentName:   "gpt-4",
			Model:       "gpt-4",
			Status:      "ready",
			Prompt:      "Fix the database connection issue in the backend API",
			Insertions:  0,
			Deletions:   0,
			Port:        0,
		},
		{
			Name:        "test-session-3",
			AgentName:   "claude-3.5-haiku",
			Model:       "claude-3.5-haiku",
			Status:      "inactive",
			Prompt:      "Write unit tests for the user management module",
			Insertions:  23,
			Deletions:   5,
			Port:        8080,
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
			// Prompt should be either full or truncated
			hasPrompt := strings.Contains(description, session.Prompt) || 
						 strings.Contains(description, session.Prompt[:min(len(session.Prompt), 47)])
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
		status string
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
		// (we can't test exact equality due to styling)
		if !strings.Contains(statusIcon, tc.expectedIcon) {
			t.Errorf("Status icon for %s should contain %s, got: %s", tc.status, tc.expectedIcon, statusIcon)
		}
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
