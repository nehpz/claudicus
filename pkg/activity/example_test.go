// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package activity_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nehpz/claudicus/pkg/activity"
)

// Example shows how to use the activity package in practice
func Example() {
	// Create new metrics
	metrics := activity.NewMetrics()
	
	// Simulate some agent activity
	metrics.Commits = 2
	metrics.Insertions = 45
	metrics.Deletions = 12
	metrics.FilesChanged = 3
	metrics.LastCommitAt = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	metrics.Status = activity.StatusWorking
	
	// Check if active
	fmt.Printf("Agent is active: %v\n", metrics.IsActive())
	fmt.Printf("Total changes: %d\n", metrics.TotalChanges())
	fmt.Printf("Has commits: %v\n", metrics.HasCommits())
	
	// Convert to JSON
	jsonData, _ := json.MarshalIndent(metrics, "", "  ")
	fmt.Printf("JSON representation:\n%s\n", jsonData)
	
	// Output:
	// Agent is active: true
	// Total changes: 57
	// Has commits: true
	// JSON representation:
	// {
	//   "commits": 2,
	//   "insertions": 45,
	//   "deletions": 12,
	//   "files_changed": 3,
	//   "last_commit_at": "2025-01-15T10:30:00Z",
	//   "status": "working"
	// }
}

// ExampleStatusValidation shows how to validate status values
func ExampleStatus_IsValid() {
	statuses := []activity.Status{
		activity.StatusWorking,
		activity.StatusIdle,
		activity.StatusStuck,
		activity.Status("invalid"),
	}
	
	for _, status := range statuses {
		fmt.Printf("Status '%s' is valid: %v\n", status, status.IsValid())
	}
	
	// Output:
	// Status 'working' is valid: true
	// Status 'idle' is valid: true
	// Status 'stuck' is valid: true
	// Status 'invalid' is valid: false
}

// ExampleMetrics_IsActive shows different scenarios for activity detection
func ExampleMetrics_IsActive() {
	// Scenario 1: Working status (always active)
	m1 := &activity.Metrics{Status: activity.StatusWorking}
	fmt.Printf("Working status active: %v\n", m1.IsActive())
	
	// Scenario 2: Recent commit (active even if idle)
	m2 := &activity.Metrics{
		Status:       activity.StatusIdle,
		LastCommitAt: time.Now().Add(-2 * time.Minute),
	}
	fmt.Printf("Recent commit active: %v\n", m2.IsActive())
	
	// Scenario 3: Old commit (not active)
	m3 := &activity.Metrics{
		Status:       activity.StatusIdle,
		LastCommitAt: time.Now().Add(-10 * time.Minute),
	}
	fmt.Printf("Old commit active: %v\n", m3.IsActive())
	
	// Output:
	// Working status active: true
	// Recent commit active: true
	// Old commit active: false
}
