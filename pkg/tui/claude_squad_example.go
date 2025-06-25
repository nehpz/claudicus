//go:build examples

// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"fmt"
	"log"
)

// ExampleClaudeSquadListView demonstrates the Claude Squad styled list view
// with the LoadSessions function for rendering sessions with agent names,
// status icons, diff stats, and dev URLs.
func ExampleClaudeSquadListView() {
	// Sample session data similar to what would come from Uzi
	sessions := []SessionInfo{
		{
			Name:        "web-dev-session",
			AgentName:   "claude-3.5-sonnet",
			Model:       "claude-3.5-sonnet",
			Status:      "running",
			Prompt:      "Create a React application with TypeScript and Tailwind CSS",
			Insertions:  127,
			Deletions:   34,
			Port:        3000,
		},
		{
			Name:        "api-dev-session",
			AgentName:   "gpt-4",
			Model:       "gpt-4",
			Status:      "attached",
			Prompt:      "Build a REST API with authentication and database integration",
			Insertions:  89,
			Deletions:   12,
			Port:        8080,
		},
		{
			Name:        "testing-session",
			AgentName:   "claude-3.5-haiku",
			Model:       "claude-3.5-haiku",
			Status:      "ready",
			Prompt:      "Write comprehensive unit tests for the user service",
			Insertions:  45,
			Deletions:   7,
			Port:        0,
		},
		{
			Name:        "refactor-session",
			AgentName:   "gpt-4o",
			Model:       "gpt-4o",
			Status:      "inactive",
			Prompt:      "Refactor the legacy codebase to improve performance and maintainability",
			Insertions:  0,
			Deletions:   0,
			Port:        0,
		},
	}

	// Create a list model with Claude Squad styling
	listModel := NewListModel(100, 20)
	
	// Load sessions using the LoadSessions function
	// This renders each row with agent name, status icon, diff stats, and dev URL
	listModel.LoadSessions(sessions)

	fmt.Println("=== Claude Squad Styled Session List ===")
	fmt.Println("Each row shows: [status_icon] agent_name (model)")
	fmt.Println("                status | diff_stats | dev_url | prompt...")
	fmt.Println()

	// Demonstrate what each session would look like in the TUI
	for i, session := range sessions {
		sessionItem := NewSessionListItem(session)
		
		fmt.Printf("%d. %s\n", i+1, sessionItem.Title())
		fmt.Printf("   %s\n", sessionItem.Description())
		fmt.Println()
	}

	// Show how the list model can be used to get selected session
	fmt.Println("List model features:")
	fmt.Printf("- Total sessions loaded: %d\n", len(listModel.list.Items()))
	fmt.Printf("- List dimensions: %dx%d\n", listModel.width, listModel.height)
	
	// In a real TUI application, you would call listModel.View() to render
	// the list and handle user interactions through the Update method
	
	fmt.Println("\nColor scheme (Claude Squad):")
	fmt.Printf("- Primary text: %s (white)\n", ClaudeSquadPrimary)
	fmt.Printf("- Accent color: %s (Claude Squad green)\n", ClaudeSquadAccent)
	fmt.Printf("- Background: %s (deep black)\n", ClaudeSquadDark)
	fmt.Printf("- Muted text: %s (gray)\n", ClaudeSquadMuted)
}

// RunClaudeSquadExample runs the Claude Squad list view example
func RunClaudeSquadExample() {
	fmt.Println("Running Claude Squad list view example...")
	
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error in Claude Squad example: %v", r)
		}
	}()
	
	ExampleClaudeSquadListView()
}
