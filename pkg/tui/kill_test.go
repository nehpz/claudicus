// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nehpz/claudicus/pkg/tui"
)

// MockUziInterface for testing kill functionality
type MockUziInterface struct {
	killedSessions []string
	shouldFail     bool
}

func (m *MockUziInterface) GetSessions() ([]SessionInfo, error) {
	return []SessionInfo{
		{Name: "test-session-1", AgentName: "agent1", Status: "ready"},
		{Name: "test-session-2", AgentName: "agent2", Status: "running"},
	}, nil
}

func (m *MockUziInterface) GetSessionState(sessionName string) (*state.AgentState, error) {
	return nil, nil // Not used in kill tests
}

func (m *MockUziInterface) GetSessionStatus(sessionName string) (string, error) {
	return "ready", nil
}

func (m *MockUziInterface) AttachToSession(sessionName string) error {
	return nil // Not used in kill tests
}

func (m *MockUziInterface) KillSession(sessionName string) error {
	if m.shouldFail {
		return errors.New("mock kill failure")
	}
	m.killedSessions = append(m.killedSessions, sessionName)
	return nil
}

func (m *MockUziInterface) RefreshSessions() error {
	return nil
}

func (m *MockUziInterface) RunPrompt(agents string, prompt string) error {
	return nil
}

func (m *MockUziInterface) RunBroadcast(message string) error {
	return nil
}

func (m *MockUziInterface) RunCommand(command string) error {
	return nil
}

func TestKillAgentHandling(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)
	
	// Initialize the app
	app.Init()
	
	// Load mock sessions
	app.list.LoadSessions([]SessionInfo{
		{Name: "test-session-1", AgentName: "agent1", Status: "ready"},
		{Name: "test-session-2", AgentName: "agent2", Status: "running"},
	})
	
	// Select the first session (cursor should be at index 0 by default)
	selectedSession := app.list.SelectedSession()
	if selectedSession == nil {
		t.Fatal("Expected a selected session, got nil")
	}
	
	if selectedSession.Name != "test-session-1" {
		t.Errorf("Expected selected session to be 'test-session-1', got '%s'", selectedSession.Name)
	}
	
	// Simulate kill key press
	killKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	
	// Test that kill key is properly handled
	if !key.Matches(killKeyMsg, app.keys.Kill) {
		t.Error("Kill key should match the Kill binding")
	}
	
	// Update the app with kill key message
	_, cmd := app.Update(killKeyMsg)
	
	// Execute the command to trigger the kill
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			// The command should return a RefreshMsg after successful kill
			if _, ok := msg.(RefreshMsg); !ok {
				t.Errorf("Expected RefreshMsg after kill, got %T", msg)
			}
		}
	}
	
	// Verify that the mock received the kill command
	if len(mockUzi.killedSessions) != 1 {
		t.Errorf("Expected 1 killed session, got %d", len(mockUzi.killedSessions))
	}
	
	if mockUzi.killedSessions[0] != "test-session-1" {
		t.Errorf("Expected killed session to be 'test-session-1', got '%s'", mockUzi.killedSessions[0])
	}
}

func TestKillAgentHandlingNoSelection(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)
	
	// Initialize the app without any sessions
	app.Init()
	
	// Don't load any sessions - list should be empty
	app.list.LoadSessions([]SessionInfo{})
	
	// Simulate kill key press with no selection
	killKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	
	// Update the app with kill key message
	_, cmd := app.Update(killKeyMsg)
	
	// Command should be nil since there's no selection
	if cmd != nil {
		t.Error("Expected no command when no session is selected")
	}
	
	// Verify that no kill command was sent
	if len(mockUzi.killedSessions) != 0 {
		t.Errorf("Expected 0 killed sessions when none selected, got %d", len(mockUzi.killedSessions))
	}
}

func TestKillAgentHandlingError(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}, shouldFail: true}
	app := NewApp(mockUzi)
	
	// Initialize and load a session
	app.Init()
	app.list.LoadSessions([]SessionInfo{
		{Name: "test-session-1", AgentName: "agent1", Status: "ready"},
	})
	
	// Simulate kill key press
	killKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	
	// Update the app with kill key message
	_, cmd := app.Update(killKeyMsg)
	
	// Execute the command - it should handle the error gracefully
	if cmd != nil {
		msg := cmd()
		// Should return nil when kill fails, not crash
		if msg != nil {
			t.Errorf("Expected nil message on kill error, got %T", msg)
		}
	}
	
	// Verify that the kill was attempted but failed
	if len(mockUzi.killedSessions) != 0 {
		t.Errorf("Expected 0 killed sessions on failure, got %d", len(mockUzi.killedSessions))
	}
}
