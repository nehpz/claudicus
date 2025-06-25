//go:build examples

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nehpz/claudicus/pkg/state"
)

func TestStateReaderWithMockData(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "uzi_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .uzi directory
	uziDir := filepath.Join(tempDir, ".uzi")
	if err := os.MkdirAll(uziDir, 0755); err != nil {
		t.Fatalf("Failed to create .uzi dir: %v", err)
	}

	// Create mock state data
	mockStates := createTestMockStates()

	// Write mock state to file
	stateFile := filepath.Join(uziDir, "state.json")
	data, err := json.MarshalIndent(mockStates, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal mock data: %v", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}

	t.Logf("Created mock state file: %s", stateFile)

	// Test the state reader
	reader := state.NewStateReader(tempDir)

	// Load all sessions
	sessions, err := reader.LoadSessions()
	if err != nil {
		t.Fatalf("Failed to load sessions: %v", err)
	}

	// Verify we loaded the expected number of sessions
	expectedSessionCount := len(mockStates)
	if len(sessions) != expectedSessionCount {
		t.Errorf("Expected %d sessions, got %d", expectedSessionCount, len(sessions))
	}

	t.Logf("Loaded Sessions (%d)", len(sessions))
	logTestSessions(t, sessions)

	// Get active sessions (would be none since tmux sessions don't exist)
	activeSessions, err := reader.GetActiveSessions()
	if err != nil {
		t.Fatalf("Failed to get active sessions: %v", err)
	}

	t.Logf("Active Sessions (%d)", len(activeSessions))
	logTestSessions(t, activeSessions)

	// Verify session data integrity
	for _, session := range sessions {
		if session.AgentName == "" {
			t.Error("Found session with empty AgentName")
		}
		if session.Model == "" {
			t.Error("Found session with empty Model")
		}
	}
}


func createTestMockStates() map[string]state.AgentState {
	now := time.Now()
	
	return map[string]state.AgentState{
		"agent-myproject-abc123-claude": {
			GitRepo:      "https://github.com/user/myproject.git",
			BranchFrom:   "main",
			BranchName:   "claude-myproject-abc123-1234567890",
			Prompt:       "Implement user authentication system",
			WorktreePath: "/tmp/worktrees/claude-myproject-abc123",
			Port:         3001,
			Model:        "claude-3-sonnet",
			CreatedAt:    now.Add(-2 * time.Hour),
			UpdatedAt:    now.Add(-30 * time.Minute),
		},
		"agent-myproject-abc123-cursor": {
			GitRepo:      "https://github.com/user/myproject.git",
			BranchFrom:   "main",
			BranchName:   "cursor-myproject-abc123-1234567891",
			Prompt:       "Fix TypeScript compilation errors",
			WorktreePath: "/tmp/worktrees/cursor-myproject-abc123",
			Port:         3002,
			Model:        "cursor-gpt-4",
			CreatedAt:    now.Add(-1 * time.Hour),
			UpdatedAt:    now.Add(-15 * time.Minute),
		},
		"agent-myproject-abc123-aider": {
			GitRepo:      "https://github.com/user/myproject.git",
			BranchFrom:   "main",
			BranchName:   "aider-myproject-abc123-1234567892",
			Prompt:       "Add comprehensive unit tests",
			WorktreePath: "/tmp/worktrees/aider-myproject-abc123",
			Port:         0, // No dev server
			Model:        "aider",
			CreatedAt:    now.Add(-30 * time.Minute),
			UpdatedAt:    now.Add(-5 * time.Minute),
		},
	}
}


func logTestSessions(t *testing.T, sessions []state.SessionInfo) {
	if len(sessions) == 0 {
		t.Log("No sessions found")
		return
	}

	// Log as formatted table
	t.Logf("%-15s %-15s %-10s %-8s %-20s %-40s", 
		"AGENT", "MODEL", "STATUS", "DIFF", "DEV_SERVER", "PROMPT")
	t.Log(strings.Repeat("-", 120))

	for _, session := range sessions {
		diff := ""
		if session.Insertions > 0 || session.Deletions > 0 {
			diff = fmt.Sprintf("+%d/-%d", session.Insertions, session.Deletions)
		}
		
		// Truncate prompt if too long
		prompt := session.Prompt
		if len(prompt) > 40 {
			prompt = prompt[:37] + "..."
		}

		devServer := session.DevServerURL
		if devServer == "" {
			devServer = "none"
		}

		t.Logf("%-15s %-15s %-10s %-8s %-20s %-40s",
			session.AgentName,
			session.Model,
			session.Status,
			diff,
			devServer,
			prompt,
		)
	}
}
