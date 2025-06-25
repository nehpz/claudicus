//go:build examples

// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

// ExampleTmuxUsage demonstrates how the TUI list view can use tmux discovery
// to highlight attached/active sessions
func ExampleTmuxUsage() error {
	// Create Uzi CLI interface with tmux discovery
	uziCLI := NewUziCLI()
	
	// Get sessions with tmux attachment information
	sessions, tmuxMapping, err := uziCLI.GetSessionsWithTmuxInfo()
	if err != nil {
		return fmt.Errorf("failed to get sessions with tmux info: %w", err)
	}
	
	// Convert sessions to list items
	var listItems []list.Item
	for _, session := range sessions {
		// Note: tmuxMapping available but not used in simplified API
		_ = tmuxMapping // Ignore for now
		
		// Create session list item with Claude Squad styling
		sessionItem := NewSessionListItem(session)
		listItems = append(listItems, sessionItem)
	}
	
	// Create list model with the enhanced items
	listModel := NewListModel(80, 24)
	
	// Print example of what the TUI would display
	fmt.Println("=== Uzi TUI Session List (with tmux highlighting) ===")
	
	for i, item := range listItems {
		if sessionItem, ok := item.(SessionListItem); ok {
			// Use the formatted status icon from the session item
			status := sessionItem.formatStatusIcon(sessionItem.session.Status)
			
			fmt.Printf("%d. %s %s\n", 
				i+1, 
				status,
				sessionItem.Title())
			fmt.Printf("   %s\n", sessionItem.Description())
		}
	}
	
	// Show activity summary
	attachedCount, err := uziCLI.GetAttachedSessionCount()
	if err == nil {
		fmt.Printf("\nAttached sessions: %d\n", attachedCount)
	}
	
	// Show sessions grouped by activity
	sessionsByActivity, err := uziCLI.GetTmuxSessionsByActivity()
	if err == nil {
		fmt.Printf("\nSession Activity Summary:\n")
		for activity, sessions := range sessionsByActivity {
			fmt.Printf("  %s: %d sessions\n", activity, len(sessions))
		}
	}
	
	_ = listModel // Would be used in actual TUI
	
	return nil
}

// ExampleTmuxDiscoveryOnly demonstrates using just the tmux discovery helper
func ExampleTmuxDiscoveryOnly() error {
	// Create tmux discovery helper
	tmuxDiscovery := NewTmuxDiscovery()
	
	// Get all tmux sessions
	allSessions, err := tmuxDiscovery.GetAllSessions()
	if err != nil {
		return fmt.Errorf("failed to get tmux sessions: %w", err)
	}
	
	fmt.Println("=== All Tmux Sessions ===")
	for name, session := range allSessions {
		status := "○"
		if session.Attached {
			status = "🔗"
		} else if session.Activity == "active" {
			status = "●"
		}
		
		fmt.Printf("%s %s (%d windows, %d panes) - %s\n", 
			status, name, session.Windows, session.Panes, session.Activity)
		
		if len(session.WindowNames) > 0 {
			fmt.Printf("   Windows: %v\n", session.WindowNames)
		}
	}
	
	// Get only Uzi sessions
	uziSessions, err := tmuxDiscovery.GetUziSessions()
	if err != nil {
		return fmt.Errorf("failed to get Uzi sessions: %w", err)
	}
	
	fmt.Printf("\n=== Uzi Sessions Only ===\n")
	for name, session := range uziSessions {
		fmt.Printf("- %s (attached: %v, activity: %s)\n", 
			name, session.Attached, session.Activity)
	}
	
	return nil
}

// RunExamples runs both examples to demonstrate the functionality
func RunExamples() {
	fmt.Println("Running tmux discovery examples...")
	
	if err := ExampleTmuxDiscoveryOnly(); err != nil {
		log.Printf("Error in tmux discovery example: %v", err)
	}
	
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")
	
	if err := ExampleTmuxUsage(); err != nil {
		log.Printf("Error in TUI usage example: %v", err)
	}
}
