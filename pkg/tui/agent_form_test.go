package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAgentFormBasicFunctionality(t *testing.T) {
	form := NewAgentFormModel()
	
	// Test initial state
	if form.IsActive() {
		t.Error("Form should not be active initially")
	}
	
	// Test activation
	form.SetActive(true)
	if !form.IsActive() {
		t.Error("Form should be active after SetActive(true)")
	}
	
	// Test validation with empty fields
	if err := form.Validate(); err == nil {
		t.Error("Form should fail validation with empty fields")
	}
}

func TestAgentFormStepNavigation(t *testing.T) {
	form := NewAgentFormModel()
	form.SetActive(true)
	
	// Should start on agent type step
	if form.currentStep != StepAgentType {
		t.Errorf("Expected to start on StepAgentType, got %v", form.currentStep)
	}
	
	// Test step navigation
	form.agentType.SetValue("claude")
	form, _ = form.nextStep()
	
	if form.currentStep != StepCount {
		t.Errorf("Expected to be on StepCount after nextStep, got %v", form.currentStep)
	}
}

func TestAgentFormValidation(t *testing.T) {
	form := NewAgentFormModel()
	
	// Test valid agent type
	form.agentType.SetValue("claude")
	if err := form.validateAgentType(); err != nil {
		t.Errorf("Valid agent type should pass validation: %v", err)
	}
	
	// Test invalid agent type
	form.agentType.SetValue("invalid")
	if err := form.validateAgentType(); err == nil {
		t.Error("Invalid agent type should fail validation")
	}
	
	// Test valid count
	form.count.SetValue("2")
	if err := form.validateCount(); err != nil {
		t.Errorf("Valid count should pass validation: %v", err)
	}
	
	// Test invalid count
	form.count.SetValue("15")
	if err := form.validateCount(); err == nil {
		t.Error("Count > 10 should fail validation")
	}
	
	// Test valid prompt
	form.prompt.SetValue("Create a React component for user authentication")
	if err := form.validatePrompt(); err != nil {
		t.Errorf("Valid prompt should pass validation: %v", err)
	}
	
	// Test invalid prompt (too short)
	form.prompt.SetValue("short")
	if err := form.validatePrompt(); err == nil {
		t.Error("Short prompt should fail validation")
	}
}

func TestAgentFormSubmission(t *testing.T) {
	form := NewAgentFormModel()
	form.SetActive(true)
	
	// Fill in valid data
	form.agentType.SetValue("claude")
	form.count.SetValue("2")
	form.prompt.SetValue("Create a React component for user authentication")
	
	// Navigate to prompt step
	form.currentStep = StepPrompt
	
	// Test form submission
	updatedForm, cmd := form.handleStepNavigation()
	
	if updatedForm.IsActive() {
		t.Error("Form should be inactive after successful submission")
	}
	
	if cmd == nil {
		t.Error("Submit should return a command")
	}
	
	// Execute the command to get the message
	msg := cmd()
	submitMsg, ok := msg.(AgentFormSubmitMsg)
	if !ok {
		t.Error("Command should return AgentFormSubmitMsg")
	}
	
	if submitMsg.AgentType != "claude" {
		t.Errorf("Expected AgentType 'claude', got '%s'", submitMsg.AgentType)
	}
	
	if submitMsg.Count != "2" {
		t.Errorf("Expected Count '2', got '%s'", submitMsg.Count)
	}
	
	if submitMsg.Prompt != "Create a React component for user authentication" {
		t.Errorf("Expected correct prompt, got '%s'", submitMsg.Prompt)
	}
}

func TestProgressModalBasicFunctionality(t *testing.T) {
	modal := NewProgressModal()
	
	// Test initial state
	if modal.IsActive() {
		t.Error("Modal should not be active initially")
	}
	
	// Test activation
	modal.SetActive(true)
	if !modal.IsActive() {
		t.Error("Modal should be active after SetActive(true)")
	}
	
	// Test step progression
	modal.NextStep()
	if modal.currentStep != ProgressStepCreateTmux {
		t.Errorf("Expected ProgressStepCreateTmux, got %v", modal.currentStep)
	}
}

func TestProgressModalKeyHandling(t *testing.T) {
	modal := NewProgressModal()
	modal.SetActive(true)
	
	// Test escape when not complete
	escMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updatedModal, _ := modal.Update(escMsg)
	
	if !updatedModal.IsActive() {
		t.Error("Modal should remain active when not complete")
	}
	
	// Test escape when complete
	modal.currentStep = ProgressStepComplete
	updatedModal, _ = modal.Update(escMsg)
	
	if updatedModal.IsActive() {
		t.Error("Modal should close when complete and escape is pressed")
	}
}

func TestAgentFormKeyHandling(t *testing.T) {
	form := NewAgentFormModel()
	form.SetActive(true)
	
	// Test escape key
	escMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}}
	escMsg.Type = tea.KeyEsc
	updatedForm, _ := form.Update(escMsg)
	
	if updatedForm.IsActive() {
		t.Error("Form should close when escape is pressed")
	}
	
	// Test enter key with invalid data
	form.SetActive(true)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedForm, _ = form.Update(enterMsg)
	
	if updatedForm.error == "" {
		t.Error("Form should show error when submitting invalid data")
	}
}
