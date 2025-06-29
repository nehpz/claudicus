// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nehpz/claudicus/pkg/state"
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

func (m *MockUziInterface) RunCheckpoint(agentName string, message string) error {
	return nil // Mock implementation
}

func (m *MockUziInterface) SpawnAgent(prompt, model string) (string, error) {
	// Mock implementation - return a fake session name
	return "agent-test-abc123-new-spawned", nil
}

func (m *MockUziInterface) SpawnAgentInteractive(opts string) (<-chan struct{}, error) {
	// Mock implementation - return a channel that immediately signals completion
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	close(ch)
	return ch, nil
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

	// Update the app with kill key message (should show modal)
	_, cmd := app.Update(killKeyMsg)

	// Command should be nil since we're just showing the modal
	if cmd != nil {
		t.Error("Expected no command when showing modal, just modal state change")
	}

	// Verify modal is visible
	if !app.confirmModal.IsVisible() {
		t.Error("Expected confirmation modal to be visible after 'k' press")
	}

	// Extract agent name and set it up properly
	agentName := extractAgentName("test-session-1")
	app.confirmModal.SetRequiredAgentName(agentName)

	// Switch to kill-only mode
	tabKeyMsg := tea.KeyMsg{Type: tea.KeyTab}
	app.confirmModal.Update(tabKeyMsg)

	// Type the correct agent name
	for _, r := range agentName {
		charKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		app.confirmModal.Update(charKeyMsg)
	}

	// Press Enter to confirm
	confirmKeyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, confirmCmd := app.confirmModal.Update(confirmKeyMsg)

	// Execute the confirmation command
	if confirmCmd != nil {
		msg := confirmCmd()
		if modalMsg, ok := msg.(ModalMsg); ok {
			if !modalMsg.Confirmed {
				t.Error("Expected confirmed modal message")
			}
			// Now process the ModalMsg
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
		t.Error("Expected confirmCmd to be returned after pressing 'y'")
	}

	// Verify that the mock received the kill command
	if len(mockUzi.killedSessions) != 1 {
		t.Errorf("Expected 1 killed session, got %d", len(mockUzi.killedSessions))
	}

	if len(mockUzi.killedSessions) > 0 && mockUzi.killedSessions[0] != "test-session-1" {
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

	// Simulate kill key press (should show modal)
	killKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, cmd := app.Update(killKeyMsg)

	// Should show modal, no command yet
	if cmd != nil {
		t.Error("Expected no command when showing modal")
	}

	// Simulate 'y' key press to confirm
	confirmKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, confirmCmd := app.Update(confirmKeyMsg)

	// Execute the confirmation command
	if confirmCmd != nil {
		msg := confirmCmd()
		if modalMsg, ok := msg.(ModalMsg); ok && modalMsg.Confirmed {
			// Now process the ModalMsg (this should try to kill and fail)
			_, killCmd := app.Update(modalMsg)
			if killCmd != nil {
				killResultMsg := killCmd()
				// Should return nil when kill fails, not crash
				if killResultMsg != nil {
					t.Errorf("Expected nil message on kill error, got %T", killResultMsg)
				}
			}
		}
	}

	// Verify that the kill was attempted but failed
	if len(mockUzi.killedSessions) != 0 {
		t.Errorf("Expected 0 killed sessions on failure, got %d", len(mockUzi.killedSessions))
	}
}

func TestKillAgentHandlingCancel(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)

	// Initialize and load a session
	app.Init()
	app.list.LoadSessions([]SessionInfo{
		{Name: "test-session-1", AgentName: "agent1", Status: "ready"},
	})

	// Simulate kill key press (should show modal)
	killKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, cmd := app.Update(killKeyMsg)

	// Should show modal, no command yet
	if cmd != nil {
		t.Error("Expected no command when showing modal")
	}

	// Verify modal is visible
	if !app.confirmModal.IsVisible() {
		t.Error("Expected confirmation modal to be visible after 'k' press")
	}

	// Simulate 'esc' key press to cancel (not 'n' in the new design)
	cancelKeyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cancelCmd := app.confirmModal.Update(cancelKeyMsg)

	// Execute the cancel command
	if cancelCmd != nil {
		msg := cancelCmd()
		if modalMsg, ok := msg.(ModalMsg); ok {
			if modalMsg.Confirmed {
				t.Error("Expected cancelled modal message, got confirmed")
			}
			// Now process the ModalMsg (should do nothing)
			_, noCmd := app.Update(modalMsg)
			if noCmd != nil {
				t.Error("Expected no command after modal cancellation")
			}
		} else {
			t.Errorf("Expected ModalMsg, got %T", msg)
		}
	}

	// Verify modal is no longer visible
	if app.confirmModal.IsVisible() {
		t.Error("Expected confirmation modal to be hidden after cancellation")
	}

	// Verify that no kill command was sent
	if len(mockUzi.killedSessions) != 0 {
		t.Errorf("Expected 0 killed sessions after cancellation, got %d", len(mockUzi.killedSessions))
	}
}
