// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package activity

import (
	"time"
)

// Status represents the current activity status of an agent or process
type Status string

const (
	// StatusWorking indicates the agent is actively working
	StatusWorking Status = "working"
	
	// StatusIdle indicates the agent is idle and waiting for tasks
	StatusIdle Status = "idle"
	
	// StatusStuck indicates the agent appears to be stuck or blocked
	StatusStuck Status = "stuck"
)

// Metrics captures activity metrics for monitoring agent behavior
// across monitor, state reader, and UI components
type Metrics struct {
	// Git-related metrics
	Commits      int       `json:"commits"`       // Number of commits made
	Insertions   int       `json:"insertions"`    // Lines of code added
	Deletions    int       `json:"deletions"`     // Lines of code removed
	FilesChanged int       `json:"files_changed"` // Number of files modified
	LastCommitAt time.Time `json:"last_commit_at"` // Timestamp of most recent commit
	
	// Current status
	Status Status `json:"status"` // Current activity status
}

// NewMetrics creates a new Metrics instance with default values
func NewMetrics() *Metrics {
	return &Metrics{
		Commits:      0,
		Insertions:   0,
		Deletions:    0,
		FilesChanged: 0,
		LastCommitAt: time.Time{}, // Zero time indicates no commits yet
		Status:       StatusIdle,
	}
}

// IsActive returns true if the metrics indicate recent activity
func (m *Metrics) IsActive() bool {
	return m.Status == StatusWorking || 
		   (!m.LastCommitAt.IsZero() && time.Since(m.LastCommitAt) < 5*time.Minute)
}

// HasCommits returns true if any commits have been recorded
func (m *Metrics) HasCommits() bool {
	return m.Commits > 0
}

// TotalChanges returns the sum of insertions and deletions
func (m *Metrics) TotalChanges() int {
	return m.Insertions + m.Deletions
}

// String returns a string representation of the status
func (s Status) String() string {
	return string(s)
}

// IsValid checks if the status is one of the defined constants
func (s Status) IsValid() bool {
	switch s {
	case StatusWorking, StatusIdle, StatusStuck:
		return true
	default:
		return false
	}
}
