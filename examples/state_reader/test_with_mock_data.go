package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/devflowinc/uzi/pkg/state"
)

func main() {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "uzi_test_*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .uzi directory
	uziDir := filepath.Join(tempDir, ".uzi")
	if err := os.MkdirAll(uziDir, 0755); err != nil {
		log.Fatalf("Failed to create .uzi dir: %v", err)
	}

	// Create mock state data
	mockStates := createMockStates()

	// Write mock state to file
	stateFile := filepath.Join(uziDir, "state.json")
	data, err := json.MarshalIndent(mockStates, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal mock data: %v", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		log.Fatalf("Failed to write state file: %v", err)
	}

	fmt.Printf("Created mock state file: %s\n", stateFile)
	fmt.Printf("Mock state content:\n%s\n\n", string(data))

	// Test the state reader
	reader := state.NewStateReader(tempDir)

	// Load all sessions
	sessions, err := reader.LoadSessions()
	if err != nil {
		log.Fatalf("Failed to load sessions: %v", err)
	}

	fmt.Printf("=== Loaded Sessions (%d) ===\n", len(sessions))
	printSessions(sessions)

	// Get active sessions (would be none since tmux sessions don't exist)
	activeSessions, err := reader.GetActiveSessions()
	if err != nil {
		log.Fatalf("Failed to get active sessions: %v", err)
	}

	fmt.Printf("\n=== Active Sessions (%d) ===\n", len(activeSessions))
	printSessions(activeSessions)
}

func createMockStates() map[string]state.AgentState {
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

func printSessions(sessions []state.SessionInfo) {
	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}

	// Print as formatted table
	fmt.Printf("%-15s %-15s %-10s %-8s %-20s %-40s\n", 
		"AGENT", "MODEL", "STATUS", "DIFF", "DEV_SERVER", "PROMPT")
	fmt.Println(strings.Repeat("-", 120))

	for _, session := range sessions {
		diff := fmt.Sprintf("+%d/-%d", session.Insertions, session.Deletions)
		
		// Truncate prompt if too long
		prompt := session.Prompt
		if len(prompt) > 40 {
			prompt = prompt[:37] + "..."
		}

		devServer := session.DevServerURL
		if devServer == "" {
			devServer = "none"
		}

		fmt.Printf("%-15s %-15s %-10s %-8s %-20s %-40s\n",
			session.AgentName,
			session.Model,
			session.Status,
			diff,
			devServer,
			prompt,
		)
	}

	// Also print as JSON for debugging
	fmt.Printf("\n=== JSON Output ===\n")
	jsonData, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}
