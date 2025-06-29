// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package activity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	if m.Commits != 0 {
		t.Errorf("Expected Commits to be 0, got %d", m.Commits)
	}

	if m.Insertions != 0 {
		t.Errorf("Expected Insertions to be 0, got %d", m.Insertions)
	}

	if m.Deletions != 0 {
		t.Errorf("Expected Deletions to be 0, got %d", m.Deletions)
	}

	if m.FilesChanged != 0 {
		t.Errorf("Expected FilesChanged to be 0, got %d", m.FilesChanged)
	}

	if !m.LastCommitAt.IsZero() {
		t.Errorf("Expected LastCommitAt to be zero time, got %v", m.LastCommitAt)
	}

	if m.Status != StatusIdle {
		t.Errorf("Expected Status to be %s, got %s", StatusIdle, m.Status)
	}
}

func TestMetricsIsActive(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *Metrics
		expected bool
	}{
		{
			name: "working status should be active",
			metrics: &Metrics{
				Status: StatusWorking,
			},
			expected: true,
		},
		{
			name: "idle status should not be active",
			metrics: &Metrics{
				Status: StatusIdle,
			},
			expected: false,
		},
		{
			name: "recent commit should be active",
			metrics: &Metrics{
				Status:       StatusIdle,
				LastCommitAt: time.Now().Add(-2 * time.Minute),
			},
			expected: true,
		},
		{
			name: "old commit should not be active",
			metrics: &Metrics{
				Status:       StatusIdle,
				LastCommitAt: time.Now().Add(-10 * time.Minute),
			},
			expected: false,
		},
		{
			name: "zero time should not be active",
			metrics: &Metrics{
				Status:       StatusIdle,
				LastCommitAt: time.Time{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metrics.IsActive()
			if result != tt.expected {
				t.Errorf("Expected IsActive() to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMetricsHasCommits(t *testing.T) {
	tests := []struct {
		name     string
		commits  int
		expected bool
	}{
		{"zero commits", 0, false},
		{"one commit", 1, true},
		{"multiple commits", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{Commits: tt.commits}
			result := m.HasCommits()
			if result != tt.expected {
				t.Errorf("Expected HasCommits() to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMetricsTotalChanges(t *testing.T) {
	tests := []struct {
		name       string
		insertions int
		deletions  int
		expected   int
	}{
		{"no changes", 0, 0, 0},
		{"only insertions", 10, 0, 10},
		{"only deletions", 0, 5, 5},
		{"mixed changes", 15, 8, 23},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				Insertions: tt.insertions,
				Deletions:  tt.deletions,
			}
			result := m.TotalChanges()
			if result != tt.expected {
				t.Errorf("Expected TotalChanges() to return %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusWorking, "working"},
		{StatusIdle, "idle"},
		{StatusStuck, "stuck"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("Expected Status.String() to return %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{"working is valid", StatusWorking, true},
		{"idle is valid", StatusIdle, true},
		{"stuck is valid", StatusStuck, true},
		{"unknown is invalid", Status("unknown"), false},
		{"empty is invalid", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("Expected Status.IsValid() to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMetricsJSONSerialization(t *testing.T) {
	// Create a metrics instance with sample data
	original := &Metrics{
		Commits:      3,
		Insertions:   150,
		Deletions:    75,
		FilesChanged: 5,
		LastCommitAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		Status:       StatusWorking,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal metrics to JSON: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled Metrics
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to metrics: %v", err)
	}

	// Verify all fields are preserved
	if unmarshaled.Commits != original.Commits {
		t.Errorf("Expected Commits %d, got %d", original.Commits, unmarshaled.Commits)
	}

	if unmarshaled.Insertions != original.Insertions {
		t.Errorf("Expected Insertions %d, got %d", original.Insertions, unmarshaled.Insertions)
	}

	if unmarshaled.Deletions != original.Deletions {
		t.Errorf("Expected Deletions %d, got %d", original.Deletions, unmarshaled.Deletions)
	}

	if unmarshaled.FilesChanged != original.FilesChanged {
		t.Errorf("Expected FilesChanged %d, got %d", original.FilesChanged, unmarshaled.FilesChanged)
	}

	if !unmarshaled.LastCommitAt.Equal(original.LastCommitAt) {
		t.Errorf("Expected LastCommitAt %v, got %v", original.LastCommitAt, unmarshaled.LastCommitAt)
	}

	if unmarshaled.Status != original.Status {
		t.Errorf("Expected Status %s, got %s", original.Status, unmarshaled.Status)
	}
}

func TestMetricsJSONFields(t *testing.T) {
	m := &Metrics{
		Commits:      1,
		Insertions:   10,
		Deletions:    5,
		FilesChanged: 2,
		LastCommitAt: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		Status:       StatusWorking,
	}

	jsonData, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Failed to marshal metrics: %v", err)
	}

	jsonStr := string(jsonData)

	// Check that all expected JSON fields are present
	expectedFields := []string{
		"\"commits\":",
		"\"insertions\":",
		"\"deletions\":",
		"\"files_changed\":",
		"\"last_commit_at\":",
		"\"status\":",
	}

	for _, field := range expectedFields {
		if !containsString(jsonStr, field) {
			t.Errorf("Expected JSON to contain field %s, but it was missing from: %s", field, jsonStr)
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
