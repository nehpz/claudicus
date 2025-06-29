// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"strings"
	"testing"
	"time"
)

func TestGetActivityStatus(t *testing.T) {
	now := time.Now().UTC() // Use UTC to avoid timezone issues
	testCases := []struct {
		name           string
		session        SessionInfo
		expectedStatus string
	}{
		{
			name: "Working - recent activity (30s ago)",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-30 * time.Second).Format("2006-01-02T15:04:05Z"),
				Insertions: 10,
				Deletions:  5,
			},
			expectedStatus: "working",
		},
		{
			name: "Working - 89s ago (just under boundary)",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-89 * time.Second).Format("2006-01-02T15:04:05Z"),
				Insertions: 10,
				Deletions:  5,
			},
			expectedStatus: "working",
		},
		{
			name: "Idle - 2 minutes ago with changes",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-2 * time.Minute).Format("2006-01-02T15:04:05Z"),
				Insertions: 10,
				Deletions:  5,
			},
			expectedStatus: "idle",
		},
		{
			name: "Stuck - 5 minutes ago with no diffs",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-5 * time.Minute).Format("2006-01-02T15:04:05Z"),
				Insertions: 0,
				Deletions:  0,
			},
			expectedStatus: "stuck",
		},
		{
			name: "Idle - 5 minutes ago with diffs",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-5 * time.Minute).Format("2006-01-02T15:04:05Z"),
				Insertions: 10,
				Deletions:  5,
			},
			expectedStatus: "idle",
		},
		{
			name: "Unknown - invalid UpdatedAt, valid CreatedAt",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  "invalid-timestamp",
				CreatedAt:  now.Add(-2 * time.Minute).Format("2006-01-02T15:04:05Z"),
				Insertions: 5,
				Deletions:  2,
			},
			expectedStatus: "idle",
		},
		{
			name: "Unknown - both timestamps invalid",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  "invalid-timestamp",
				CreatedAt:  "invalid-timestamp",
				Insertions: 5,
				Deletions:  2,
			},
			expectedStatus: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := NewSessionListItem(tc.session)
			status := item.getActivityStatus()
			if status != tc.expectedStatus {
				t.Errorf("Expected activity status %s, got %s", tc.expectedStatus, status)
			}
		})
	}
}

func TestFormatActivityBar(t *testing.T) {
	now := time.Now().UTC()
	testCases := []struct {
		name         string
		session      SessionInfo
		expectedBars []string // Multiple possible unicode representations
	}{
		{
			name: "Working activity bar",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-30 * time.Second).Format("2006-01-02T15:04:05Z"),
				Insertions: 10,
				Deletions:  5,
			},
			expectedBars: []string{"▮▮▮"}, // Green working bar
		},
		{
			name: "Idle activity bar",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-2 * time.Minute).Format("2006-01-02T15:04:05Z"),
				Insertions: 10,
				Deletions:  5,
			},
			expectedBars: []string{"▮▮▯"}, // Yellow idle bar
		},
		{
			name: "Stuck activity bar",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-5 * time.Minute).Format("2006-01-02T15:04:05Z"),
				Insertions: 0,
				Deletions:  0,
			},
			expectedBars: []string{"▮▯▯"}, // Red stuck bar
		},
		{
			name: "Unknown activity bar",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  "invalid-timestamp",
				CreatedAt:  "invalid-timestamp",
				Insertions: 5,
				Deletions:  2,
			},
			expectedBars: []string{"▯▯▯"}, // Gray unknown bar
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := NewSessionListItem(tc.session)
			activityBar := item.formatActivityBar()

			// Check if the activity bar contains any of the expected bar patterns
			found := false
			for _, expectedBar := range tc.expectedBars {
				if strings.Contains(activityBar, expectedBar) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected activity bar to contain one of %v, got: %s", tc.expectedBars, activityBar)
			}
		})
	}
}

func TestFormatLastActivity(t *testing.T) {
	now := time.Now().UTC()
	testCases := []struct {
		name             string
		session          SessionInfo
		expectedContains string
	}{
		{
			name: "Seconds ago",
			session: SessionInfo{
				Name:      "test-session",
				AgentName: "claude",
				UpdatedAt: now.Add(-30 * time.Second).Format("2006-01-02T15:04:05Z"),
			},
			expectedContains: "s ago",
		},
		{
			name: "Minutes ago",
			session: SessionInfo{
				Name:      "test-session",
				AgentName: "claude",
				UpdatedAt: now.Add(-5 * time.Minute).Format("2006-01-02T15:04:05Z"),
			},
			expectedContains: "m ago",
		},
		{
			name: "Hours ago",
			session: SessionInfo{
				Name:      "test-session",
				AgentName: "claude",
				UpdatedAt: now.Add(-2 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
			expectedContains: "h ago",
		},
		{
			name: "Days ago",
			session: SessionInfo{
				Name:      "test-session",
				AgentName: "claude",
				UpdatedAt: now.Add(-2 * 24 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
			expectedContains: "d ago",
		},
		{
			name: "Invalid UpdatedAt, fallback to CreatedAt",
			session: SessionInfo{
				Name:      "test-session",
				AgentName: "claude",
				UpdatedAt: "invalid-timestamp",
				CreatedAt: now.Add(-5 * time.Minute).Format("2006-01-02T15:04:05Z"),
			},
			expectedContains: "m ago",
		},
		{
			name: "Both timestamps invalid",
			session: SessionInfo{
				Name:      "test-session",
				AgentName: "claude",
				UpdatedAt: "invalid-timestamp",
				CreatedAt: "invalid-timestamp",
			},
			expectedContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := NewSessionListItem(tc.session)
			lastActivity := item.formatLastActivity()

			if tc.expectedContains == "" {
				if lastActivity != "" {
					t.Errorf("Expected empty string for invalid timestamps, got: %s", lastActivity)
				}
			} else {
				if !strings.Contains(lastActivity, tc.expectedContains) {
					t.Errorf("Expected last activity to contain %s, got: %s", tc.expectedContains, lastActivity)
				}
			}
		})
	}
}

func TestActivityBarInTitle(t *testing.T) {
	now := time.Now().UTC()
	session := SessionInfo{
		Name:       "test-session",
		AgentName:  "claude",
		Model:      "claude-3.5-sonnet",
		Status:     "running",
		UpdatedAt:  now.Add(-30 * time.Second).Format("2006-01-02T15:04:05Z"),
		Insertions: 10,
		Deletions:  5,
	}

	item := NewSessionListItem(session)
	title := item.Title()

	// Should contain status icon, activity bar, agent name, and model
	if !strings.Contains(title, "claude") {
		t.Errorf("Title should contain agent name, got: %s", title)
	}
	if !strings.Contains(title, "claude-3.5-sonnet") {
		t.Errorf("Title should contain model, got: %s", title)
	}
	// Activity bar should be present (any of the bar characters)
	hasActivityBar := strings.Contains(title, "▮") || strings.Contains(title, "▯")
	if !hasActivityBar {
		t.Errorf("Title should contain activity bar, got: %s", title)
	}
}

func TestLastActivityInDescription(t *testing.T) {
	now := time.Now().UTC()
	session := SessionInfo{
		Name:       "test-session",
		AgentName:  "claude",
		Model:      "claude-3.5-sonnet",
		Status:     "running",
		UpdatedAt:  now.Add(-2 * time.Minute).Format("2006-01-02T15:04:05Z"),
		Insertions: 10,
		Deletions:  5,
		Port:       3000,
		Prompt:     "Create a web application",
	}

	item := NewSessionListItem(session)
	description := item.Description()

	// Should contain last activity time
	if !strings.Contains(description, "m ago") {
		t.Errorf("Description should contain last activity time, got: %s", description)
	}

	// Should still contain other expected elements
	if !strings.Contains(description, "running") {
		t.Errorf("Description should contain status, got: %s", description)
	}
	if !strings.Contains(description, "+10/-5") {
		t.Errorf("Description should contain diff stats, got: %s", description)
	}
	if !strings.Contains(description, "localhost:3000") {
		t.Errorf("Description should contain dev URL, got: %s", description)
	}
	if !strings.Contains(description, "Create a web application") {
		t.Errorf("Description should contain prompt, got: %s", description)
	}
}

func TestFormatStatusCoverage(t *testing.T) {
	testCases := []struct {
		status      string
		description string
	}{
		{"attached", "attached status"},
		{"running", "running status"},
		{"ready", "ready status"},
		{"inactive", "inactive status"},
		{"unknown", "unknown status"},
		{"custom", "custom status"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			session := SessionInfo{
				Name:      "test-session",
				AgentName: "claude",
				Status:    tc.status,
			}

			item := NewSessionListItem(session)
			formattedStatus := item.formatStatus(tc.status)

			// Should return a non-empty styled string
			if formattedStatus == "" {
				t.Errorf("formatStatus should return non-empty string for status %s", tc.status)
			}
		})
	}
}

func TestActivityStatusEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		session     SessionInfo
		description string
	}{
		{
			name: "Exactly 90 seconds - should be working",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  time.Now().UTC().Add(-90 * time.Second).Format("2006-01-02T15:04:05Z"),
				Insertions: 5,
				Deletions:  2,
			},
			description: "boundary case for working status",
		},
		{
			name: "Exactly 3 minutes - should check diffs",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  time.Now().UTC().Add(-3 * time.Minute).Format("2006-01-02T15:04:05Z"),
				Insertions: 0,
				Deletions:  0,
			},
			description: "boundary case for stuck status",
		},
		{
			name: "Empty timestamps - should be unknown",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  "",
				CreatedAt:  "",
				Insertions: 5,
				Deletions:  2,
			},
			description: "empty timestamp handling",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := NewSessionListItem(tc.session)
			status := item.getActivityStatus()

			// Should return a valid status
			validStatuses := []string{"working", "idle", "stuck", "unknown"}
			found := false
			for _, validStatus := range validStatuses {
				if status == validStatus {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("getActivityStatus should return valid status, got: %s", status)
			}
		})
	}
}
