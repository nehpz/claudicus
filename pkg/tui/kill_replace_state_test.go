// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStateCleanupAfterKillReplace(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)

	// Initialize the app
	app.Init()

	// Load mock sessions
	app.list.LoadSessions([]SessionInfo{{
		Name:      "agent-session-1",
		AgentName: "claude",
		Status:    "running",
	}})

	// Simulate the kill key press
	killKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	app.Update(killKeyMsg)

	// Enter the correct agent name to proceed
	for _, r := range "claude" {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		app.confirmModal.Update(charMsg)
	}

	// Confirm the action with enter
	enterKeyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, confirmCmd := app.confirmModal.Update(enterKeyMsg)

	if confirmCmd != nil {
		msg := confirmCmd()
		if modalMsg, ok := msg.(ModalMsg); ok {
			if !modalMsg.Confirmed {
				t.Error("Expected confirmed modal message")
			}
			// Process the ModalMsg
			_, killCmd := app.Update(modalMsg)
			if killCmd != nil {
				killResultMsg := killCmd()
				if _, ok := killResultMsg.(RefreshMsg); !ok {
					t.Errorf("Expected RefreshMsg after kill, got %T", killResultMsg)
				}
			} else {
				t.Error("Expected killCmd to be returned after confirming modal")
			}
		} else {
			t.Errorf("Expected ModalMsg, got %T", msg)
		}
	} else {
		t.Error("Expected confirmCmd to be returned after pressing enter")
	}

	// Check that all sessions were cleared
	if len(app.list.Items()) != 0 {
		t.Errorf("Expected zero sessions after cleanup, got %d", len(app.list.Items()))
	}

	// Verify the sessions killed were as expected
	if len(mockUzi.killedSessions) != 1 || mockUzi.killedSessions[0] != "agent-session-1" {
		t.Errorf("Expected 'agent-session-1' to be killed, got %v", mockUzi.killedSessions)
	}
}

func TestGitWorktreeAndTmuxCleanup(t *testing.T) {
	// This test would verify mocking of git worktree removal and tmux session cleanup
	// For now, we'll test the interface expectations
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)

	// Initialize the app
	app.Init()

	// Load mock sessions
	app.list.LoadSessions([]SessionInfo{
		{Name: "agent-test-session", AgentName: "claude", Status: "ready"},
	})

	// Simulate kill workflow
	killKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	app.Update(killKeyMsg)

	// Switch to kill-only mode using tab
	tabKeyMsg := tea.KeyMsg{Type: tea.KeyTab}
	app.confirmModal.Update(tabKeyMsg)

	// Enter correct agent name
	for _, r := range "claude" {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		app.confirmModal.Update(charMsg)
	}

	// Confirm kill
	enterKeyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := app.confirmModal.Update(enterKeyMsg)

	// Process the confirmation
	if cmd != nil {
		msg := cmd()
		if modalMsg, ok := msg.(ModalMsg); ok && modalMsg.Confirmed {
			// Process the kill command
			app.Update(modalMsg)

			// Verify that the mock interface's KillSession was called
			if len(mockUzi.killedSessions) != 1 {
				t.Errorf("Expected 1 killed session, got %d", len(mockUzi.killedSessions))
			}

			// In a real scenario, this would also test:
			// - Git worktree removal via exec.Command calls
			// - Tmux session cleanup via tmux kill-session
			// - Directory cleanup
			// But since we're using a mock, we verify the interface was called correctly
		}
	}
}

func TestReplaceStateConsistency(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)

	// Initialize the app
	app.Init()

	// Load mock sessions
	app.list.LoadSessions([]SessionInfo{
		{Name: "agent-project-abc123-claude", AgentName: "claude", Status: "ready"},
	})

	// Show the modal
	killKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	app.Update(killKeyMsg)

	// Step 1: Enter correct agent name
	for _, r := range "claude" {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		app.confirmModal.Update(charMsg)
	}

	// Move to step 2
	enterKeyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	app.confirmModal.Update(enterKeyMsg)

	// Step 2: Enter prompt
	testPrompt := "Fix the authentication bug"
	for _, r := range testPrompt {
		charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		app.confirmModal.Update(charMsg)
	}

	// Move to step 3
	app.confirmModal.Update(enterKeyMsg)

	// Step 3: Confirm model (should be pre-filled)
	_, cmd := app.confirmModal.Update(enterKeyMsg)

	// Should generate a ModalMsg with replacement details
	if cmd != nil {
		msg := cmd()
		if modalMsg, ok := msg.(ModalMsg); ok {
			if !modalMsg.SpawnReplacement {
				t.Error("Expected replacement spawn in replace mode")
			}
			if modalMsg.Prompt != testPrompt {
				t.Errorf("Expected prompt '%s', got '%s'", testPrompt, modalMsg.Prompt)
			}

			// Process the modal message to trigger the replacement
			_, replaceCmd := app.Update(modalMsg)
			if replaceCmd != nil {
				// This would trigger both kill and spawn operations
				// In our mock, we can verify the kill happened
				if len(mockUzi.killedSessions) != 1 {
					t.Errorf("Expected 1 killed session, got %d", len(mockUzi.killedSessions))
				}
			}
		}
	}
}
