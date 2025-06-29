// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"testing"
	"time"
)

func TestListModelFiltering(t *testing.T) {
	// Create a list model
	listModel := NewListModel(80, 24)

	// Create test sessions with different activity statuses
	now := time.Now()
	sessions := []SessionInfo{
		{
			Name:       "working-session",
			AgentName:  "claude-1",
			Model:      "claude-3.5-sonnet",
			Status:     "running",
			UpdatedAt:  now.Add(-30 * time.Second).Format(time.RFC3339),
			Insertions: 10,
			Deletions:  5,
		},
		{
			Name:       "stuck-session",
			AgentName:  "claude-2",
			Model:      "claude-3.5-sonnet",
			Status:     "ready",
			UpdatedAt:  now.Add(-5 * time.Minute).Format(time.RFC3339),
			Insertions: 0,
			Deletions:  0,
		},
		{
			Name:       "idle-session",
			AgentName:  "claude-3",
			Model:      "claude-3.5-sonnet",
			Status:     "ready",
			UpdatedAt:  now.Add(-2 * time.Minute).Format(time.RFC3339),
			Insertions: 0,
			Deletions:  0,
		},
	}

	// Load sessions
	listModel.LoadSessions(sessions)

	// Test initial state - should show all sessions
	if len(listModel.list.Items()) != 3 {
		t.Errorf("Expected 3 sessions initially, got %d", len(listModel.list.Items()))
	}

	// Test working filter
	listModel.SetWorkingFilter()
	if len(listModel.list.Items()) != 1 {
		t.Errorf("Expected 1 working session, got %d", len(listModel.list.Items()))
	}
	if listModel.GetFilterStatus() != "Showing working agents only" {
		t.Errorf("Expected working filter status, got: %s", listModel.GetFilterStatus())
	}

	// Test stuck filter toggle
	listModel.ToggleStuckFilter()
	if len(listModel.list.Items()) != 1 {
		t.Errorf("Expected 1 stuck session, got %d", len(listModel.list.Items()))
	}
	if listModel.GetFilterStatus() != "Showing stuck agents only" {
		t.Errorf("Expected stuck filter status, got: %s", listModel.GetFilterStatus())
	}

	// Test stuck filter toggle off
	listModel.ToggleStuckFilter()
	if len(listModel.list.Items()) != 3 {
		t.Errorf("Expected all 3 sessions when filter is toggled off, got %d", len(listModel.list.Items()))
	}
	if listModel.GetFilterStatus() != "" {
		t.Errorf("Expected no filter status when toggled off, got: %s", listModel.GetFilterStatus())
	}

	// Test clear filter
	listModel.SetWorkingFilter()
	listModel.ClearFilter()
	if len(listModel.list.Items()) != 3 {
		t.Errorf("Expected all 3 sessions after clear filter, got %d", len(listModel.list.Items()))
	}
	if listModel.GetFilterStatus() != "" {
		t.Errorf("Expected no filter status after clear, got: %s", listModel.GetFilterStatus())
	}
}

func TestActivityStatusClassification(t *testing.T) {
	now := time.Now()

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
				UpdatedAt:  now.Add(-30 * time.Second).Format(time.RFC3339),
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
				UpdatedAt:  now.Add(-89 * time.Second).Format(time.RFC3339),
				Insertions: 10,
				Deletions:  5,
			},
			expectedStatus: "working",
		},
		{
			name: "Idle - 2 minutes ago with no diffs",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-2 * time.Minute).Format(time.RFC3339),
				Insertions: 0,
				Deletions:  0,
			},
			expectedStatus: "idle",
		},
		{
			name: "Stuck - 5 minutes ago with no diffs",
			session: SessionInfo{
				Name:       "test-session",
				AgentName:  "claude",
				UpdatedAt:  now.Add(-5 * time.Minute).Format(time.RFC3339),
				Insertions: 0,
				Deletions:  0,
			},
			expectedStatus: "stuck",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			item := NewSessionListItem(tc.session)
			status := item.getActivityStatus()
			if status != tc.expectedStatus {
				t.Errorf("Expected status %s, got %s", tc.expectedStatus, status)
			}
		})
	}
}

func TestKeyMapFilterKeys(t *testing.T) {
	keyMap := DefaultKeyMap()

	// Test that filter keys exist
	if len(keyMap.FilterStuck.Keys()) == 0 {
		t.Error("FilterStuck key binding should have keys")
	}
	if len(keyMap.FilterWorking.Keys()) == 0 {
		t.Error("FilterWorking key binding should have keys")
	}

	// Test that filter keys are correctly bound
	if keyMap.FilterStuck.Keys()[0] != "f" {
		t.Errorf("Expected FilterStuck to be bound to 'f', got %s", keyMap.FilterStuck.Keys()[0])
	}
	if keyMap.FilterWorking.Keys()[0] != "w" {
		t.Errorf("Expected FilterWorking to be bound to 'w', got %s", keyMap.FilterWorking.Keys()[0])
	}
}
