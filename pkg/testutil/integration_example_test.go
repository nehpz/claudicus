// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package testutil_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nehpz/claudicus/pkg/testutil/fsmock"
	"github.com/nehpz/claudicus/pkg/testutil/timefreeze"
)

// Example showing how fsmock and timefreeze work together for state persistence testing
func TestIntegratedStatePersistence(t *testing.T) {
	// Set up temporary filesystem
	fs := fsmock.NewTempFS(t)

	// Set up time control
	freeze := timefreeze.NewWithTime(t, timefreeze.TestTime)

	// Create state directory structure
	fs.MkdirAll("state/sessions", 0755)

	// Simulate agent state
	type AgentState struct {
		Name      string    `json:"name"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Duration  string    `json:"duration"`
	}

	// Create initial state
	initialState := AgentState{
		Name:      "test-agent",
		Status:    "starting",
		CreatedAt: freeze.Now(),
		UpdatedAt: freeze.Now(),
	}

	// Save initial state
	stateData, err := json.MarshalIndent(initialState, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal state: %v", err)
	}

	statePath := "state/sessions/test-agent.json"
	err = fs.WriteFileString(statePath, string(stateData), 0644)
	if err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}

	// Verify initial state was saved
	if !fs.Exists(statePath) {
		t.Fatal("State file should exist")
	}

	// Advance time and update state
	freeze.Advance(5 * time.Minute)

	// Load and update state
	loadedData, err := fs.ReadFileString(statePath)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	var loadedState AgentState
	err = json.Unmarshal([]byte(loadedData), &loadedState)
	if err != nil {
		t.Fatalf("Failed to unmarshal state: %v", err)
	}

	// Update state
	loadedState.Status = "running"
	loadedState.UpdatedAt = freeze.Now()
	loadedState.Duration = freeze.Now().Sub(loadedState.CreatedAt).String()

	// Save updated state
	updatedData, err := json.MarshalIndent(loadedState, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal updated state: %v", err)
	}

	err = fs.WriteFileString(statePath, string(updatedData), 0644)
	if err != nil {
		t.Fatalf("Failed to write updated state: %v", err)
	}

	// Advance time again for completion
	freeze.Advance(10 * time.Minute)

	// Final state update
	loadedData, err = fs.ReadFileString(statePath)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	err = json.Unmarshal([]byte(loadedData), &loadedState)
	if err != nil {
		t.Fatalf("Failed to unmarshal state: %v", err)
	}

	loadedState.Status = "completed"
	loadedState.UpdatedAt = freeze.Now()
	loadedState.Duration = freeze.Now().Sub(loadedState.CreatedAt).String()

	finalData, err := json.MarshalIndent(loadedState, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal final state: %v", err)
	}

	err = fs.WriteFileString(statePath, string(finalData), 0644)
	if err != nil {
		t.Fatalf("Failed to write final state: %v", err)
	}

	// Verify final state
	finalContent, err := fs.ReadFileString(statePath)
	if err != nil {
		t.Fatalf("Failed to read final state: %v", err)
	}

	var finalState AgentState
	err = json.Unmarshal([]byte(finalContent), &finalState)
	if err != nil {
		t.Fatalf("Failed to unmarshal final state: %v", err)
	}

	// Assertions
	if finalState.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", finalState.Name)
	}

	if finalState.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", finalState.Status)
	}

	// Test deterministic time calculations
	expectedDuration := 15 * time.Minute // 5 + 10 minutes
	actualDuration := finalState.UpdatedAt.Sub(finalState.CreatedAt)

	if actualDuration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, actualDuration)
	}

	// Test that duration string matches
	if finalState.Duration != actualDuration.String() {
		t.Errorf("Duration string doesn't match: expected %s, got %s",
			actualDuration.String(), finalState.Duration)
	}

	t.Logf("Successfully tested state persistence with deterministic time and filesystem")
	t.Logf("State file path: %s", fs.Path(statePath))
	t.Logf("Final state: %+v", finalState)
}

// Example showing how to test session history with both utilities
func TestSessionHistoryTracking(t *testing.T) {
	fs := fsmock.NewTempFS(t)
	freeze := timefreeze.NewWithTime(t, timefreeze.TestTime)

	// Create session history directory
	fs.MkdirAll("history/sessions", 0755)

	sessionName := "test-session"
	sessionStart := freeze.Now()

	// Track session events over time
	events := []struct {
		event string
		delay time.Duration
	}{
		{"session_created", 0},
		{"agent_started", 30 * time.Second},
		{"first_response", 2 * time.Minute},
		{"user_interaction", 5 * time.Minute},
		{"task_completed", 10 * time.Minute},
		{"session_ended", 12 * time.Minute},
	}

	var timeline []map[string]interface{}

	for _, event := range events {
		freeze.AdvanceTo(sessionStart.Add(event.delay))

		entry := map[string]interface{}{
			"timestamp": freeze.Now(),
			"event":     event.event,
			"elapsed":   freeze.Now().Sub(sessionStart).String(),
		}
		timeline = append(timeline, entry)

		// Save individual event log
		eventData, _ := json.MarshalIndent(entry, "", "  ")
		eventFile := fmt.Sprintf("history/sessions/%s_%s.json", sessionName, event.event)
		fs.WriteFileString(eventFile, string(eventData), 0644)
	}

	// Save complete timeline
	timelineData, err := json.MarshalIndent(timeline, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal timeline: %v", err)
	}

	timelineFile := fmt.Sprintf("history/sessions/%s_timeline.json", sessionName)
	err = fs.WriteFileString(timelineFile, string(timelineData), 0644)
	if err != nil {
		t.Fatalf("Failed to write timeline: %v", err)
	}

	// Verify all event files exist
	for _, event := range events {
		eventFile := fmt.Sprintf("history/sessions/%s_%s.json", sessionName, event.event)
		if !fs.Exists(eventFile) {
			t.Errorf("Event file %s should exist", eventFile)
		}
	}

	// Verify timeline file exists
	if !fs.Exists(timelineFile) {
		t.Fatal("Timeline file should exist")
	}

	// Test timeline content
	loadedTimeline, err := fs.ReadFileString(timelineFile)
	if err != nil {
		t.Fatalf("Failed to read timeline: %v", err)
	}

	var parsedTimeline []map[string]interface{}
	err = json.Unmarshal([]byte(loadedTimeline), &parsedTimeline)
	if err != nil {
		t.Fatalf("Failed to unmarshal timeline: %v", err)
	}

	if len(parsedTimeline) != len(events) {
		t.Errorf("Expected %d timeline entries, got %d", len(events), len(parsedTimeline))
	}

	// Verify session duration
	totalDuration := freeze.Now().Sub(sessionStart)
	expectedDuration := 12 * time.Minute

	if totalDuration != expectedDuration {
		t.Errorf("Expected session duration %v, got %v", expectedDuration, totalDuration)
	}

	// List all files created
	entries, err := fs.ListDir("history/sessions")
	if err != nil {
		t.Fatalf("Failed to list session files: %v", err)
	}

	expectedFiles := len(events) + 1 // events + timeline
	if len(entries) != expectedFiles {
		t.Errorf("Expected %d files, got %d", expectedFiles, len(entries))
	}

	t.Logf("Successfully tracked session history with %d events", len(events))
	t.Logf("Session duration: %v", totalDuration)
	t.Logf("Files created: %d", len(entries))
}
