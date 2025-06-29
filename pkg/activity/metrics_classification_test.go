// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package activity

import (
	"testing"
	"time"
)

// TestMetricsClassificationThresholds tests the classification logic with various threshold scenarios
func TestMetricsClassificationThresholds(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	now := time.Now()

	tests := []struct {
		name     string
		metrics  *Metrics
		expected Status
		desc     string
	}{
		// Test StatusWorking conditions
		{
			name: "Working_WithUncommittedInsertions",
			metrics: &Metrics{
				Insertions:   5,
				Deletions:    0,
				FilesChanged: 1,
				LastCommitAt: now.Add(-30 * time.Minute),
				Status:       StatusIdle, // Status should be recalculated
			},
			expected: StatusWorking,
			desc:     "Agent with uncommitted insertions should be classified as working",
		},
		{
			name: "Working_WithUncommittedDeletions",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    3,
				FilesChanged: 1,
				LastCommitAt: now.Add(-30 * time.Minute),
				Status:       StatusIdle,
			},
			expected: StatusWorking,
			desc:     "Agent with uncommitted deletions should be classified as working",
		},
		{
			name: "Working_WithChangedFiles",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 2,
				LastCommitAt: now.Add(-30 * time.Minute),
				Status:       StatusIdle,
			},
			expected: StatusWorking,
			desc:     "Agent with changed files should be classified as working",
		},
		{
			name: "Working_WithRecentCommits_59Minutes",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-59 * time.Minute), // Just under 1 hour
				Status:       StatusIdle,
			},
			expected: StatusWorking,
			desc:     "Agent with commits within last hour should be classified as working",
		},
		{
			name: "Working_WithRecentCommits_ExactlyOneHour",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-60 * time.Minute), // Exactly 1 hour
				Status:       StatusIdle,
			},
			expected: StatusWorking,
			desc:     "Agent with commits exactly one hour ago should be classified as working",
		},

		// Test StatusStuck conditions
		{
			name: "Stuck_OldCommits_ExactlyTwoHours",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-2 * time.Hour), // Exactly 2 hours
				Status:       StatusIdle,
			},
			expected: StatusStuck,
			desc:     "Agent with no activity for exactly 2 hours should be classified as stuck",
		},
		{
			name: "Stuck_OldCommits_ThreeHours",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-3 * time.Hour),
				Status:       StatusIdle,
			},
			expected: StatusStuck,
			desc:     "Agent with no activity for 3 hours should be classified as stuck",
		},
		{
			name: "Stuck_VeryOldCommits_OneDay",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-24 * time.Hour),
				Status:       StatusIdle,
			},
			expected: StatusStuck,
			desc:     "Agent with no activity for a day should be classified as stuck",
		},

		// Test StatusIdle conditions
		{
			name: "Idle_RecentCommits_BetweenOneAndTwoHours",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-90 * time.Minute), // 1.5 hours
				Status:       StatusWorking,
			},
			expected: StatusIdle,
			desc:     "Agent with commits between 1-2 hours ago should be classified as idle",
		},
		{
			name: "Idle_NoCommitsEver",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: time.Time{}, // Zero time
				Status:       StatusWorking,
			},
			expected: StatusIdle,
			desc:     "Agent with no commits ever should be classified as idle",
		},
		{
			name:     "Idle_NilMetrics",
			metrics:  nil,
			expected: StatusIdle,
			desc:     "Nil metrics should be classified as idle",
		},

		// Test edge cases
		{
			name: "EdgeCase_JustOverOneHour",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-61 * time.Minute), // Just over 1 hour
				Status:       StatusWorking,
			},
			expected: StatusIdle,
			desc:     "Agent with commits just over one hour should transition to idle",
		},
		{
			name: "EdgeCase_JustUnderTwoHours",
			metrics: &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(-119 * time.Minute), // Just under 2 hours
				Status:       StatusWorking,
			},
			expected: StatusIdle,
			desc:     "Agent with commits just under two hours should be idle",
		},

		// Test priority - uncommitted changes should override time-based classification
		{
			name: "Priority_UncommittedChangesOverrideOldCommits",
			metrics: &Metrics{
				Insertions:   1,
				Deletions:    0,
				FilesChanged: 1,
				LastCommitAt: now.Add(-5 * time.Hour), // Very old commits
				Status:       StatusStuck,
			},
			expected: StatusWorking,
			desc:     "Uncommitted changes should override time-based stuck classification",
		},
		{
			name: "Priority_AllTypesOfUncommittedChanges",
			metrics: &Metrics{
				Insertions:   10,
				Deletions:    5,
				FilesChanged: 3,
				LastCommitAt: now.Add(-10 * time.Hour), // Very old commits
				Status:       StatusStuck,
			},
			expected: StatusWorking,
			desc:     "Multiple types of uncommitted changes should ensure working status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.ClassifyAtTime(tt.metrics, now)
			if result != tt.expected {
				t.Errorf("Expected status %v, got %v. %s", tt.expected, result, tt.desc)
			}
		})
	}
}

// TestMetricsClassificationBoundaryConditions tests boundary conditions around time thresholds
func TestMetricsClassificationBoundaryConditions(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	now := time.Now()

	// Test millisecond precision around the 1-hour boundary
	boundaryTests := []struct {
		name     string
		offset   time.Duration
		expected Status
	}{
		{"59min59sec999ms", -59*time.Minute - 59*time.Second - 999*time.Millisecond, StatusWorking},
		{"60min0sec0ms", -60 * time.Minute, StatusWorking},
		{"60min0sec1ms", -60*time.Minute - 1*time.Millisecond, StatusIdle},
		{"60min1sec", -60*time.Minute - 1*time.Second, StatusIdle},
	}

	for _, tt := range boundaryTests {
		t.Run("OneHourBoundary_"+tt.name, func(t *testing.T) {
			metrics := &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(tt.offset),
				Status:       StatusIdle,
			}

			result := monitor.ClassifyAtTime(metrics, now)
			if result != tt.expected {
				t.Errorf("Expected status %v for offset %v, got %v", tt.expected, tt.offset, result)
			}
		})
	}

	// Test millisecond precision around the 2-hour boundary
	twoHourTests := []struct {
		name     string
		offset   time.Duration
		expected Status
	}{
		{"119min59sec999ms", -119*time.Minute - 59*time.Second - 999*time.Millisecond, StatusIdle},
		{"120min0sec0ms", -120 * time.Minute, StatusStuck},
		{"120min0sec1ms", -120*time.Minute - 1*time.Millisecond, StatusStuck},
		{"121min", -121 * time.Minute, StatusStuck},
	}

	for _, tt := range twoHourTests {
		t.Run("TwoHourBoundary_"+tt.name, func(t *testing.T) {
			metrics := &Metrics{
				Insertions:   0,
				Deletions:    0,
				FilesChanged: 0,
				LastCommitAt: now.Add(tt.offset),
				Status:       StatusIdle,
			}

			result := monitor.ClassifyAtTime(metrics, now)
			if result != tt.expected {
				t.Errorf("Expected status %v for offset %v, got %v", tt.expected, tt.offset, result)
			}
		})
	}
}

// TestMetricsClassificationStateTransitions tests all possible state transitions
func TestMetricsClassificationStateTransitions(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	now := time.Now()

	// Test Working -> Idle transition
	metrics := &Metrics{
		Insertions:   5,
		Deletions:    2,
		FilesChanged: 1,
		LastCommitAt: now.Add(-2 * time.Hour),
		Status:       StatusWorking,
	}

	// Initially working due to uncommitted changes
	result := monitor.ClassifyAtTime(metrics, now)
	if result != StatusWorking {
		t.Errorf("Expected StatusWorking with uncommitted changes, got %v", result)
	}

	// Clear uncommitted changes, should transition to stuck due to old commits
	metrics.Insertions = 0
	metrics.Deletions = 0
	metrics.FilesChanged = 0
	result = monitor.ClassifyAtTime(metrics, now)
	if result != StatusStuck {
		t.Errorf("Expected StatusStuck after clearing changes with old commits, got %v", result)
	}

	// Add recent commit, should transition to working
	metrics.LastCommitAt = now.Add(-30 * time.Minute)
	result = monitor.ClassifyAtTime(metrics, now)
	if result != StatusWorking {
		t.Errorf("Expected StatusWorking with recent commit, got %v", result)
	}

	// Age the commit beyond 1 hour but less than 2 hours, should be idle
	metrics.LastCommitAt = now.Add(-90 * time.Minute)
	result = monitor.ClassifyAtTime(metrics, now)
	if result != StatusIdle {
		t.Errorf("Expected StatusIdle with moderately old commit, got %v", result)
	}
}

// TestMetricsClassificationConsistency tests that classification is consistent across multiple calls
func TestMetricsClassificationConsistency(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	now := time.Now()

	metrics := &Metrics{
		Insertions:   3,
		Deletions:    1,
		FilesChanged: 2,
		LastCommitAt: now.Add(-45 * time.Minute),
		Status:       StatusIdle,
	}

	// Call classify multiple times - should always return the same result
	results := make([]Status, 5)
	for i := 0; i < 5; i++ {
		results[i] = monitor.ClassifyAtTime(metrics, now)
	}

	// All results should be identical
	expected := StatusWorking
	for i, result := range results {
		if result != expected {
			t.Errorf("Call %d: Expected consistent result %v, got %v", i, expected, result)
		}
	}
}

// TestMetricsClassificationWithConcurrentAccess tests classification under concurrent access
func TestMetricsClassificationWithConcurrentAccess(t *testing.T) {
	monitor := NewAgentActivityMonitor()
	now := time.Now()

	metrics := &Metrics{
		Insertions:   1,
		Deletions:    0,
		FilesChanged: 1,
		LastCommitAt: now.Add(-30 * time.Minute),
		Status:       StatusIdle,
	}

	// Launch multiple goroutines that classify the same metrics
	const numGoroutines = 10
	results := make(chan Status, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			result := monitor.Classify(metrics)
			results <- result
		}()
	}

	// Collect all results
	var allResults []Status
	for i := 0; i < numGoroutines; i++ {
		allResults = append(allResults, <-results)
	}

	// All results should be the same
	expected := StatusWorking
	for i, result := range allResults {
		if result != expected {
			t.Errorf("Goroutine %d: Expected %v, got %v", i, expected, result)
		}
	}
}

// TestMetricsIsActiveThresholds tests the IsActive method thresholds
func TestMetricsIsActiveThresholds(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		metrics  *Metrics
		expected bool
		desc     string
	}{
		{
			name: "Active_StatusWorking",
			metrics: &Metrics{
				Status:       StatusWorking,
				LastCommitAt: now.Add(-10 * time.Hour), // Should be ignored when status is working
			},
			expected: true,
			desc:     "StatusWorking should always be active regardless of commit time",
		},
		{
			name: "Active_RecentCommit_4Minutes",
			metrics: &Metrics{
				Status:       StatusIdle,
				LastCommitAt: now.Add(-4 * time.Minute),
			},
			expected: true,
			desc:     "Commit within 5 minutes should be active",
		},
		{
			name: "Active_RecentCommit_ExactlyFiveMinutes",
			metrics: &Metrics{
				Status:       StatusIdle,
				LastCommitAt: now.Add(-5 * time.Minute),
			},
			expected: false, // Exactly 5 minutes should not be active (< not <=)
			desc:     "Commit exactly 5 minutes ago should not be active",
		},
		{
			name: "Inactive_OldCommit",
			metrics: &Metrics{
				Status:       StatusIdle,
				LastCommitAt: now.Add(-6 * time.Minute),
			},
			expected: false,
			desc:     "Commit older than 5 minutes should not be active",
		},
		{
			name: "Inactive_NoCommits",
			metrics: &Metrics{
				Status:       StatusStuck,
				LastCommitAt: time.Time{}, // Zero time
			},
			expected: false,
			desc:     "No commits should not be active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metrics.IsActive()
			if result != tt.expected {
				t.Errorf("Expected IsActive=%v, got %v. %s", tt.expected, result, tt.desc)
			}
		})
	}
}
