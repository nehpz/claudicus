package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModalMsg struct {
	Confirmed     bool
	AgentName     string
	Prompt        string
	Model         string
	SpawnReplacement bool
}

type ConfirmationModal struct {
	visible           bool
	message           string
	requiredAgentName string
	textInput         textinput.Model
	mode              string // "kill" or "replace"
	promptInput       textinput.Model
	modelInput        textinput.Model
	currentStep       int // 0: agent name, 1: prompt, 2: model
}

func NewConfirmationModal() *ConfirmationModal {
	// Create text input for agent name confirmation
	ti := textinput.New()
	ti.Placeholder = "Enter agent name to confirm"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 30

	// Create text input for prompt
	promptInput := textinput.New()
	promptInput.Placeholder = "Enter prompt for replacement agent"
	promptInput.CharLimit = 500
	promptInput.Width = 50

	// Create text input for model
	modelInput := textinput.New()
	modelInput.Placeholder = "Enter model (e.g., claude:1)"
	modelInput.CharLimit = 50
	modelInput.Width = 30
	modelInput.SetValue("claude:1") // Default value

	return &ConfirmationModal{
		visible:     false,
		message:     "Kill & Replace Agent",
		textInput:   ti,
		promptInput: promptInput,
		modelInput:  modelInput,
		mode:        "replace", // Default to replace mode
		currentStep: 0,
	}
}

func (m *ConfirmationModal) Update(msg tea.Msg) (*ConfirmationModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.mode == "replace" {
				return m.handleReplaceStep()
			}
			// For simple kill mode
			if m.currentStep == 0 && strings.TrimSpace(m.textInput.Value()) == m.requiredAgentName {
				m.visible = false
				return m, func() tea.Msg {
					return ModalMsg{Confirmed: true, AgentName: m.requiredAgentName}
				}
			}
			
		case "esc":
			m.visible = false
			m.reset()
			return m, func() tea.Msg {
				return ModalMsg{Confirmed: false}
			}
			
		case "tab":
			// Switch between kill and replace modes in step 0
			if m.currentStep == 0 {
				if m.mode == "kill" {
					m.mode = "replace"
					m.message = "Kill & Replace Agent"
				} else {
					m.mode = "kill"
					m.message = "Kill Agent"
				}
			}
			
		default:
			// Handle text input updates based on current step
			var cmd tea.Cmd
			switch m.currentStep {
			case 0:
				m.textInput, cmd = m.textInput.Update(msg)
			case 1:
				m.promptInput, cmd = m.promptInput.Update(msg)
			case 2:
				m.modelInput, cmd = m.modelInput.Update(msg)
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m *ConfirmationModal) handleReplaceStep() (*ConfirmationModal, tea.Cmd) {
	switch m.currentStep {
	case 0:
		// Agent name confirmation step
		if strings.TrimSpace(m.textInput.Value()) == m.requiredAgentName {
			m.currentStep = 1
			m.textInput.Blur()
			m.promptInput.Focus()
			return m, nil
		}
		// Invalid agent name, stay on step 0
		return m, nil
		
	case 1:
		// Prompt input step
		if strings.TrimSpace(m.promptInput.Value()) != "" {
			m.currentStep = 2
			m.promptInput.Blur()
			m.modelInput.Focus()
			return m, nil
		}
		// Empty prompt, stay on step 1
		return m, nil
		
	case 2:
		// Model input step - complete the workflow
		if strings.TrimSpace(m.modelInput.Value()) != "" {
			m.visible = false
			agentName := m.requiredAgentName
			prompt := strings.TrimSpace(m.promptInput.Value())
			model := strings.TrimSpace(m.modelInput.Value())
			m.reset()
			return m, func() tea.Msg {
				return ModalMsg{
					Confirmed:        true,
					AgentName:        agentName,
					Prompt:           prompt,
					Model:            model,
					SpawnReplacement: true,
				}
			}
		}
		// Empty model, stay on step 2
		return m, nil
	}
	
	return m, nil
}

func (m *ConfirmationModal) reset() {
	m.currentStep = 0
	m.textInput.SetValue("")
	m.promptInput.SetValue("")
	m.modelInput.SetValue("claude:1")
	m.textInput.Focus()
	m.promptInput.Blur()
	m.modelInput.Blur()
}

func (m *ConfirmationModal) View() string {
	if !m.visible {
		return ""
	}

	// Create the modal content based on current step and mode
	title := ClaudeSquadAccentStyle.Render("⚠️  " + m.message)
	var content string
	
	// Step indicator
	stepIndicator := ""
	if m.mode == "replace" {
		stepIndicator = ClaudeSquadMutedStyle.Render(fmt.Sprintf("Step %d of 3", m.currentStep+1))
	}
	
	switch m.currentStep {
	case 0:
		// Agent name confirmation step
		message := ClaudeSquadPrimaryStyle.Render("Type agent name '" + m.requiredAgentName + "' to confirm:")
		modeHint := ""
		if m.mode == "kill" {
			modeHint = ClaudeSquadMutedStyle.Render("[TAB] to switch to Kill & Replace mode")
		} else {
			modeHint = ClaudeSquadMutedStyle.Render("[TAB] to switch to Kill Only mode")
		}
		
		inputView := m.textInput.View()
		escapeHint := ClaudeSquadMutedStyle.Render("(ESC to cancel)")
		
		contentParts := []string{title}
		if stepIndicator != "" {
			contentParts = append(contentParts, stepIndicator)
		}
		contentParts = append(contentParts, "", message, "", inputView, "", modeHint, escapeHint)
		content = lipgloss.JoinVertical(lipgloss.Center, contentParts...)
		
	case 1:
		// Prompt input step (replace mode only)
		message := ClaudeSquadPrimaryStyle.Render("Enter prompt for replacement agent:")
		inputView := m.promptInput.View()
		hint := ClaudeSquadMutedStyle.Render("[ENTER] to continue | [ESC] to cancel")
		
		content = lipgloss.JoinVertical(lipgloss.Center, title, stepIndicator, "", message, "", inputView, "", hint)
		
	case 2:
		// Model input step (replace mode only)
		message := ClaudeSquadPrimaryStyle.Render("Enter model for replacement agent:")
		inputView := m.modelInput.View()
		hint := ClaudeSquadMutedStyle.Render("[ENTER] to execute | [ESC] to cancel")
		
		content = lipgloss.JoinVertical(lipgloss.Center, title, stepIndicator, "", message, "", inputView, "", hint)
	}
	
	// Apply border with padding
	return ClaudeSquadBorderStyle.Copy().
		Width(70).
		Align(lipgloss.Center).
		Render(content)
}

func (m *ConfirmationModal) SetVisible(v bool) {
	m.visible = v
}

func (m *ConfirmationModal) IsVisible() bool {
	return m.visible
}

func (m *ConfirmationModal) SetRequiredAgentName(agentName string) {
	m.requiredAgentName = agentName
	m.reset() // Reset state when setting new agent name
}
