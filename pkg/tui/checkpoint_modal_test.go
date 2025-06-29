package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCheckpointModal_New(t *testing.T) {
	modal := NewCheckpointModal()

	if modal.IsVisible() {
		t.Error("Expected new modal to be hidden")
	}

	if modal.currentStep != CheckpointStepSelectAgent {
		t.Error("Expected new modal to start at agent selection step")
	}
}

func TestCheckpointModal_SetVisible(t *testing.T) {
	modal := NewCheckpointModal()

	// Test showing modal
	modal.SetVisible(true)
	if !modal.IsVisible() {
		t.Error("Expected modal to be visible after SetVisible(true)")
	}

	// Test hiding modal
	modal.SetVisible(false)
	if modal.IsVisible() {
		t.Error("Expected modal to be hidden after SetVisible(false)")
	}
}

func TestCheckpointModal_SetAgents(t *testing.T) {
	modal := NewCheckpointModal()
	agents := []SessionInfo{
		{Name: "agent-test-abc123-claude", AgentName: "claude", Status: "ready"},
		{Name: "agent-test-def456-cursor", AgentName: "cursor", Status: "running"},
	}

	modal.SetAgents(agents)

	if len(modal.agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(modal.agents))
	}
}

func TestCheckpointModal_Update_WhenHidden(t *testing.T) {
	modal := NewCheckpointModal()

	// When modal is hidden, Update should return no commands
	_, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if cmd != nil {
		t.Error("Expected no command when modal is hidden")
	}
}

func TestCheckpointModal_Update_AgentSelection(t *testing.T) {
	modal := NewCheckpointModal()
	modal.SetVisible(true)

	agents := []SessionInfo{
		{Name: "agent-test-abc123-claude", AgentName: "claude", Status: "ready"},
		{Name: "agent-test-def456-cursor", AgentName: "cursor", Status: "running"},
	}
	modal.SetAgents(agents)

	// Test navigation
	if modal.selectedIdx != 0 {
		t.Error("Expected initial selection at index 0")
	}

	// Navigate down
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if modal.selectedIdx != 1 {
		t.Error("Expected selection to move down to index 1")
	}

	// Navigate up
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if modal.selectedIdx != 0 {
		t.Error("Expected selection to move up to index 0")
	}
}

func TestCheckpointModal_Update_EnterProgression(t *testing.T) {
	modal := NewCheckpointModal()
	modal.SetVisible(true)

	agents := []SessionInfo{
		{Name: "agent-test-abc123-claude", AgentName: "claude", Status: "ready"},
	}
	modal.SetAgents(agents)

	// Press Enter to go to commit message step
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if modal.currentStep != CheckpointStepCommitMessage {
		t.Error("Expected to progress to commit message step")
	}
}

func TestCheckpointModal_View_WhenHidden(t *testing.T) {
	modal := NewCheckpointModal()

	view := modal.View()
	if view != "" {
		t.Error("Expected empty view when modal is hidden")
	}
}

func TestCheckpointModal_View_WhenVisible(t *testing.T) {
	modal := NewCheckpointModal()
	modal.SetVisible(true)
	modal.SetSize(80, 24)

	agents := []SessionInfo{
		{Name: "agent-test-abc123-claude", AgentName: "claude", Status: "ready"},
	}
	modal.SetAgents(agents)

	view := modal.View()
	if view == "" {
		t.Error("Expected non-empty view when modal is visible")
	}

	// View should contain expected elements
	if !strings.Contains(view, "Checkpoint Agent") {
		t.Error("Expected view to contain title")
	}
}
