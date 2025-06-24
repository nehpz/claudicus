package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nehpz/claudicus/pkg/state"
)

func main() {
	// Get the current working directory (repository root)
	repoRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// If we're in the examples subdirectory, go up to the repo root
	if filepath.Base(repoRoot) == "state_reader" {
		repoRoot = filepath.Dir(filepath.Dir(repoRoot))
	}

	fmt.Printf("Reading state from: %s/.uzi/state.json\n", repoRoot)

	// Create a new state reader
	reader := state.NewStateReader(repoRoot)

	// Load all sessions
	sessions, err := reader.LoadSessions()
	if err != nil {
		log.Fatalf("Failed to load sessions: %v", err)
	}

	fmt.Printf("\n=== All Sessions (%d) ===\n", len(sessions))
	printSessions(sessions)

	// Get only active sessions
	activeSessions, err := reader.GetActiveSessions()
	if err != nil {
		log.Fatalf("Failed to get active sessions: %v", err)
	}

	fmt.Printf("\n=== Active Sessions (%d) ===\n", len(activeSessions))
	printSessions(activeSessions)
}

func printSessions(sessions []state.SessionInfo) {
	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}

	// Print as formatted table
	fmt.Printf("%-20s %-15s %-10s %-8s %-25s %-30s\n", 
		"AGENT", "MODEL", "STATUS", "DIFF", "DEV_SERVER", "PROMPT")
	fmt.Println(strings.Repeat("-", 120))

	for _, session := range sessions {
		diff := fmt.Sprintf("+%d/-%d", session.Insertions, session.Deletions)
		
		// Truncate prompt if too long
		prompt := session.Prompt
		if len(prompt) > 30 {
			prompt = prompt[:27] + "..."
		}

		fmt.Printf("%-20s %-15s %-10s %-8s %-25s %-30s\n",
			session.AgentName,
			session.Model,
			session.Status,
			diff,
			session.DevServerURL,
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
