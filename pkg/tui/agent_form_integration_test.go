package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewAgentKeyBinding(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)

	// Initialize the app
	app.Init()
	app.width = 80
	app.height = 24

	// Test that 'n' key activates the agent form
	nKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}

	// Verify form is initially inactive
	if app.agentForm.IsActive() {
		t.Error("Agent form should not be active initially")
	}

	// Update with 'n' key
	updatedApp, cmd := app.Update(nKeyMsg)

	// Cast back to App type
	app = updatedApp.(*App)

	// Verify form is now active
	if !app.agentForm.IsActive() {
		t.Error("Agent form should be active after 'n' key press")
	}

	// Verify a command was returned (spinner tick)
	if cmd == nil {
		t.Error("Expected a command to be returned for spinner")
	}
}

func TestFullAgentCreationWorkflow(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)

	// Initialize the app
	app.Init()
	app.width = 80
	app.height = 24

	// Step 1: Press 'n' to open form
	nKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedApp, _ := app.Update(nKeyMsg)
	app = updatedApp.(*App)

	// Step 2: Fill in agent type
	app.agentForm.agentType.SetValue("claude")

	// Step 3: Move to count field
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedApp, _ = app.Update(enterMsg)
	app = updatedApp.(*App)

	// Step 4: Fill in count
	app.agentForm.count.SetValue("2")

	// Step 5: Move to prompt field
	updatedApp, _ = app.Update(enterMsg)
	app = updatedApp.(*App)

	// Step 6: Fill in prompt
	app.agentForm.prompt.SetValue("Create a React component for user authentication")

	// Step 7: Submit form
	updatedApp, cmd := app.Update(enterMsg)
	app = updatedApp.(*App)

	// The form should become inactive immediately upon submission
	if app.agentForm.IsActive() {
		t.Error("Agent form should be inactive after submission")
	}

	// Execute the command to trigger the AgentFormSubmitMsg
	if cmd != nil {
		msg := cmd()

		// Should get an AgentFormSubmitMsg
		if submitMsg, ok := msg.(AgentFormSubmitMsg); ok {
			if submitMsg.AgentType != "claude" {
				t.Errorf("Expected AgentType 'claude', got '%s'", submitMsg.AgentType)
			}
			if submitMsg.Count != "2" {
				t.Errorf("Expected Count '2', got '%s'", submitMsg.Count)
			}
			if submitMsg.Prompt != "Create a React component for user authentication" {
				t.Errorf("Unexpected prompt: %s", submitMsg.Prompt)
			}

			// Now process the AgentFormSubmitMsg through the app
			updatedApp, _ = app.Update(submitMsg)
			app = updatedApp.(*App)

			// Now verify form is no longer active
			if app.agentForm.IsActive() {
				t.Error("Agent form should be inactive after submission")
			}

			// Verify progress modal is active
			if !app.progressModal.IsActive() {
				t.Error("Progress modal should be active after form submission")
			}
		} else {
			t.Errorf("Expected AgentFormSubmitMsg, got %T", msg)
		}
	} else {
		t.Error("Expected a command to be returned from form submission")
	}
}

func TestAgentFormEscapeKey(t *testing.T) {
	mockUzi := &MockUziInterface{killedSessions: []string{}}
	app := NewApp(mockUzi)

	// Initialize the app
	app.Init()
	app.width = 80
	app.height = 24

	// Press 'n' to open form
	nKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedApp, _ := app.Update(nKeyMsg)
	app = updatedApp.(*App)

	// Verify form is active
	if !app.agentForm.IsActive() {
		t.Error("Agent form should be active after 'n' key press")
	}

	// Press escape to close form
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedApp, _ = app.Update(escMsg)
	app = updatedApp.(*App)

	// Verify form is now inactive
	if app.agentForm.IsActive() {
		t.Error("Agent form should be inactive after escape key")
	}

	// Verify progress modal is not active
	if app.progressModal.IsActive() {
		t.Error("Progress modal should not be active after escape")
	}
}
