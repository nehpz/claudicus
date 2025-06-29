// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package activity

import (
	"context"
	"testing"
	"time"
)

func TestAgentActivityMonitor_NewAgentActivityMonitor(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	if monitor == nil {
		t.Fatal("Expected NewAgentActivityMonitor to return non-nil monitor")
	}

	if monitor.stateManager == nil {
		t.Error("Expected state manager to be initialized")
	}

	if monitor.metrics == nil {
		t.Error("Expected metrics map to be initialized")
	}

	if monitor.running {
		t.Error("Expected monitor to not be running initially")
	}
}

func TestAgentActivityMonitor_Start(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test starting monitor
	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("Expected Start to succeed, got error: %v", err)
	}

	if !monitor.running {
		t.Error("Expected monitor to be running after Start")
	}

	if monitor.ticker == nil {
		t.Error("Expected ticker to be initialized after Start")
	}

	// Test starting already running monitor
	err = monitor.Start(ctx)
	if err == nil {
		t.Error("Expected Start to fail when monitor is already running")
	}

	// Clean up
	monitor.Stop()
}

func TestAgentActivityMonitor_Stop(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test stopping non-running monitor (should not panic)
	monitor.Stop()

	// Start and then stop
	err := monitor.Start(ctx)
	if err != nil {
		t.Fatalf("Expected Start to succeed, got error: %v", err)
	}

	monitor.Stop()

	if monitor.running {
		t.Error("Expected monitor to not be running after Stop")
	}
}

func TestAgentActivityMonitor_Classify(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	now := time.Now()

	tests := []struct {
		name     string
		metrics  *Metrics
		expected Status
	}{
		{
			name:     "nil metrics should return idle",
			metrics:  nil,
			expected: StatusIdle,
		},
		{
			name: "uncommitted insertions should return working",
			metrics: &Metrics{
				Insertions:   10,
				Deletions:    0,
				FilesChanged: 1,
				LastCommitAt: now.Add(-30 * time.Minute),
			},
			expected: StatusWorking,
		},
		{
			name: "uncommitted deletions should return working",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    5,
				FilesChanged: 1,
				LastCommitAt: now.Add(-30 * time.Minute),
			},
			expected: StatusWorking,
		},
		{
			name: "uncommitted file changes should return working",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 2,
				LastCommitAt: now.Add(-30 * time.Minute),
			},
			expected: StatusWorking,
		},
		{
			name: "recent commit within hour should return working",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-30 * time.Minute),
			},
			expected: StatusWorking,
		},
		{
			name: "commit exactly 1 hour ago should return working",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-time.Hour + time.Second), // Just under 1 hour
			},
			expected: StatusWorking,
		},
		{
			name: "commit slightly over 1 hour ago should return idle",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-time.Hour - time.Minute),
			},
			expected: StatusIdle,
		},
		{
			name: "commit over 2 hours ago should return stuck",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-3 * time.Hour),
			},
			expected: StatusStuck,
		},
		{
			name: "commit just under 2 hours ago should return idle",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-2*time.Hour + time.Second),
			},
			expected: StatusIdle,
		},
		{
			name: "commit exactly 2 hours ago should return stuck",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-2 * time.Hour),
			},
			expected: StatusStuck,
		},
		{
			name: "no commits ever should return idle",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: time.Time{}, // Zero time
			},
			expected: StatusIdle,
		},
		{
			name: "uncommitted changes override old commit timestamp",
			metrics: &Metrics{
				Insertions:   5,
				Deletions:    2,
				FilesChanged: 1,
				LastCommitAt: now.Add(-5 * time.Hour), // Very old commit
			},
			expected: StatusWorking, // Should be working due to uncommitted changes
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := monitor.Classify(tc.metrics)
			if result != tc.expected {
				t.Errorf("Expected status %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestAgentActivityMonitor_parseShortstat(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	tests := []struct {
		name                 string
		output               string
		expectedInsertions   int
		expectedDeletions    int
		expectedFilesChanged int
	}{
		{
			name:                 "empty output",
			output:               "",
			expectedInsertions:   0,
			expectedDeletions:    0,
			expectedFilesChanged: 0,
		},
		{
			name:                 "whitespace only",
			output:               "   \n  \t  ",
			expectedInsertions:   0,
			expectedDeletions:    0,
			expectedFilesChanged: 0,
		},
		{
			name:                 "standard format with all stats",
			output:               " 3 files changed, 15 insertions(+), 7 deletions(-)",
			expectedInsertions:   15,
			expectedDeletions:    7,
			expectedFilesChanged: 3,
		},
		{
			name:                 "single file singular form",
			output:               " 1 file changed, 5 insertions(+), 2 deletions(-)",
			expectedInsertions:   5,
			expectedDeletions:    2,
			expectedFilesChanged: 1,
		},
		{
			name:                 "only insertions",
			output:               " 2 files changed, 10 insertions(+)",
			expectedInsertions:   10,
			expectedDeletions:    0,
			expectedFilesChanged: 2,
		},
		{
			name:                 "only deletions",
			output:               " 1 file changed, 8 deletions(-)",
			expectedInsertions:   0,
			expectedDeletions:    8,
			expectedFilesChanged: 1,
		},
		{
			name:                 "only file changes no content changes",
			output:               " 3 files changed",
			expectedInsertions:   0,
			expectedDeletions:    0,
			expectedFilesChanged: 3,
		},
		{
			name:                 "single insertion singular form",
			output:               " 1 file changed, 1 insertion(+)",
			expectedInsertions:   1,
			expectedDeletions:    0,
			expectedFilesChanged: 1,
		},
		{
			name:                 "single deletion singular form",
			output:               " 1 file changed, 1 deletion(-)",
			expectedInsertions:   0,
			expectedDeletions:    1,
			expectedFilesChanged: 1,
		},
		{
			name:                 "large numbers",
			output:               " 25 files changed, 1500 insertions(+), 800 deletions(-)",
			expectedInsertions:   1500,
			expectedDeletions:    800,
			expectedFilesChanged: 25,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			insertions, deletions, filesChanged := monitor.parseShortstat(tc.output)

			if insertions != tc.expectedInsertions {
				t.Errorf("Expected %d insertions, got %d", tc.expectedInsertions, insertions)
			}

			if deletions != tc.expectedDeletions {
				t.Errorf("Expected %d deletions, got %d", tc.expectedDeletions, deletions)
			}

			if filesChanged != tc.expectedFilesChanged {
				t.Errorf("Expected %d files changed, got %d", tc.expectedFilesChanged, filesChanged)
			}
		})
	}
}

func TestAgentActivityMonitor_UpdateAll(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	now := time.Now()

	// Set up some test metrics
	monitor.metrics = map[string]*Metrics{
		"session1": {
			Commits:      5,
			Insertions:   10,
			Deletions:    3,
			FilesChanged: 2,
			LastCommitAt: now,
			Status:       StatusWorking,
		},
		"session2": {
			Commits:      0,
			Insertions:   0,
			Deletions:    0,
			FilesChanged: 0,
			LastCommitAt: time.Time{},
			Status:       StatusIdle,
		},
	}

	result := monitor.UpdateAll()

	// Check that we get the right number of sessions
	if len(result) != 2 {
		t.Errorf("Expected 2 sessions in result, got %d", len(result))
	}

	// Check session1 metrics
	session1, exists := result["session1"]
	if !exists {
		t.Fatal("Expected session1 to exist in result")
	}

	if session1.Commits != 5 {
		t.Errorf("Expected session1 commits to be 5, got %d", session1.Commits)
	}

	if session1.Status != StatusWorking {
		t.Errorf("Expected session1 status to be %s, got %s", StatusWorking, session1.Status)
	}

	// Check session2 metrics
	session2, exists := result["session2"]
	if !exists {
		t.Fatal("Expected session2 to exist in result")
	}

	if session2.Status != StatusIdle {
		t.Errorf("Expected session2 status to be %s, got %s", StatusIdle, session2.Status)
	}

	// Verify it's a deep copy by modifying original and checking copy
	monitor.metrics["session1"].Commits = 999
	if session1.Commits != 5 {
		t.Error("Expected result to be a deep copy, but original modification affected result")
	}
}

func TestAgentActivityMonitor_getOrCreateMetrics(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	// Test creating new metrics
	metrics1 := monitor.getOrCreateMetrics("new-session")
	if metrics1 == nil {
		t.Fatal("Expected getOrCreateMetrics to return non-nil metrics")
	}

	if metrics1.Status != StatusIdle {
		t.Errorf("Expected new metrics to have status %s, got %s", StatusIdle, metrics1.Status)
	}

	// Test getting existing metrics
	metrics2 := monitor.getOrCreateMetrics("new-session")
	if metrics1 != metrics2 {
		t.Error("Expected getOrCreateMetrics to return same instance for existing session")
	}

	// Verify it was stored in the map
	if len(monitor.metrics) != 1 {
		t.Errorf("Expected 1 metrics entry, got %d", len(monitor.metrics))
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusWorking, "working"},
		{StatusIdle, "idle"},
		{StatusStuck, "stuck"},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			if string(tc.status) != tc.expected {
				t.Errorf("Expected status string %s, got %s", tc.expected, string(tc.status))
			}
		})
	}
}

// Test edge cases and boundary conditions for classification logic
func TestAgentActivityMonitor_ClassifyEdgeCases(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	now := time.Now()

	tests := []struct {
		name     string
		metrics  *Metrics
		expected Status
	}{
		{
			name: "recent commit with unchanged files still working",
			metrics: &Metrics{
				Commits:      1,
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-30 * time.Minute),
			},
			expected: StatusWorking,
		},
		{
			name: "future timestamp should still work",
			metrics: &Metrics{
				Commits:      1,
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(5 * time.Minute), // Future time
			},
			expected: StatusWorking,
		},
		{
			name: "very old commit should be stuck",
			metrics: &Metrics{
				Commits:      5,
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-24 * time.Hour), // 24 hours ago
			},
			expected: StatusStuck,
		},
		{
			name: "zero values with valid timestamp should be idle",
			metrics: &Metrics{
				Commits:      0,
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-90 * time.Minute), // 1.5 hours ago
			},
			expected: StatusIdle,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := monitor.Classify(tc.metrics)
			if result != tc.expected {
				t.Errorf("Expected status %s, got %s", tc.expected, result)
			}
		})
	}
}
