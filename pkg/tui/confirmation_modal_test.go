// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmationModal_New(t *testing.T) {
	modal := NewConfirmationModal()

	if modal == nil {
		t.Fatal("Expected modal to be created, got nil")
	}

	if modal.IsVisible() {
		t.Error("Expected new modal to be hidden")
	}

	if modal.message == "" {
		t.Error("Expected modal to have a default message")
	}
}

func TestConfirmationModal_SetVisible(t *testing.T) {
	modal := NewConfirmationModal()

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

func TestConfirmationModal_Update_WhenHidden(t *testing.T) {
	modal := NewConfirmationModal()

	// When modal is hidden, Update should return nil command
	_, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if cmd != nil {
		t.Error("Expected no command when modal is hidden")
	}
}

func TestConfirmationModal_Update_YesKey(t *testing.T) {
	modal := NewConfirmationModal()
	modal.SetVisible(true)
	modal.SetRequiredAgentName("test-agent")

	// Switch to kill-only mode using Tab
	_, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Error("Expected no command when switching modes")
	}

	// Type the correct agent name
	for _, r := range "test-agent" {
		_, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press Enter to confirm
	_, cmd = modal.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("Expected command to be returned after Enter press")
	}

	// Execute the command to get the message
	msg := cmd()
	if modalMsg, ok := msg.(ModalMsg); ok {
		if !modalMsg.Confirmed {
			t.Error("Expected confirmed modal message after Enter press")
		}
		if modalMsg.AgentName != "test-agent" {
			t.Errorf("Expected agent name 'test-agent', got '%s'", modalMsg.AgentName)
		}
	} else {
		t.Errorf("Expected ModalMsg, got %T", msg)
	}

	// Modal should be hidden after confirmation
	if modal.IsVisible() {
		t.Error("Expected modal to be hidden after confirmation")
	}
}

func TestConfirmationModal_Update_NoKey(t *testing.T) {
	// In the new modal design, 'n' key is treated as text input
	// Escape is used for cancellation
	modal := NewConfirmationModal()
	modal.SetVisible(true)

	// Press 'esc' key to cancel (not 'n' anymore)
	_, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Expected command to be returned after 'esc' press")
	}

	// Execute the command to get the message
	msg := cmd()
	if modalMsg, ok := msg.(ModalMsg); ok {
		if modalMsg.Confirmed {
			t.Error("Expected cancelled modal message after 'esc' press")
		}
	} else {
		t.Errorf("Expected ModalMsg, got %T", msg)
	}

	// Modal should be hidden after cancellation
	if modal.IsVisible() {
		t.Error("Expected modal to be hidden after cancellation")
	}
}

func TestConfirmationModal_Update_EscapeKey(t *testing.T) {
	modal := NewConfirmationModal()
	modal.SetVisible(true)

	// Press 'esc' key
	_, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("Expected command to be returned after 'esc' press")
	}

	// Execute the command to get the message
	msg := cmd()
	if modalMsg, ok := msg.(ModalMsg); ok {
		if modalMsg.Confirmed {
			t.Error("Expected cancelled modal message after 'esc' press")
		}
	} else {
		t.Errorf("Expected ModalMsg, got %T", msg)
	}

	// Modal should be hidden after cancellation
	if modal.IsVisible() {
		t.Error("Expected modal to be hidden after cancellation")
	}
}

func TestConfirmationModal_Update_OtherKeys(t *testing.T) {
	modal := NewConfirmationModal()
	modal.SetVisible(true)

	// In the new design, regular characters are input to the text field
	// Only special keys like enter/escape should generate commands
	keys := []string{"a", "x", "q", "1"}

	for _, key := range keys {
		_, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})

		// Text input might return commands for bubbles internal operations
		// The key test is that modal remains visible and operational

		// Modal should still be visible
		if !modal.IsVisible() {
			t.Errorf("Expected modal to remain visible after key '%s'", key)
		}
	}
}

func TestConfirmationModal_View_WhenHidden(t *testing.T) {
	modal := NewConfirmationModal()

	view := modal.View()
	if view != "" {
		t.Error("Expected empty view when modal is hidden")
	}
}

func TestConfirmationModal_View_WhenVisible(t *testing.T) {
	modal := NewConfirmationModal()
	modal.SetVisible(true)

	view := modal.View()
	if view == "" {
		t.Error("Expected non-empty view when modal is visible")
	}

	// View should contain expected elements
	if !containsAny(view, []string{"Confirmation", "Are you sure", "[Y]es", "[N]o", "ESC"}) {
		t.Error("Expected view to contain confirmation dialog elements")
	}
}

// Helper function to check if a string contains any of the provided substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
