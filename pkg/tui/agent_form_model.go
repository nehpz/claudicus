package tui

// Imports required for the model functions
import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormStep represents the current step in the form
type FormStep int

const (
	StepAgentType FormStep = iota
	StepCount
	StepPrompt
	StepComplete
)

// AgentFormModel for creating new agents interactively
type AgentFormModel struct {
	agentType   textinput.Model
	count       textinput.Model
	prompt      textinput.Model
	currentStep FormStep
	active      bool
	error       string
	width       int
	height      int
}

// NewAgentFormModel creates and initializes an AgentFormModel
func NewAgentFormModel() AgentFormModel {
	agentType := textinput.NewModel()
	agentType.Placeholder = "Agent type (claude, cursor, codex, gemini)"
	agentType.Focus()

	count := textinput.NewModel()
	count.Placeholder = "Number of agents (1-10)"
	count.CharLimit = 2

	prompt := textinput.NewModel()
	prompt.Placeholder = "Enter your prompt..."

	return AgentFormModel{
		agentType:   agentType,
		count:       count,
		prompt:      prompt,
		currentStep: StepAgentType,
		active:      false,
		error:       "",
	}
}

// SetActive sets the form's active state
func (m *AgentFormModel) SetActive(active bool) {
	m.active = active
	if active {
		m.currentStep = StepAgentType
		m.agentType.Focus()
		m.count.Blur()
		m.prompt.Blur()
		m.error = ""
	}
}

// IsActive returns whether the form is currently active
func (m *AgentFormModel) IsActive() bool {
	return m.active
}

// SetSize updates the form dimensions
func (m *AgentFormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.agentType.Width = width - 20
	m.count.Width = width - 20
	m.prompt.Width = width - 20
}

// Update handles form input and navigation
func (m AgentFormModel) Update(msg tea.Msg) (AgentFormModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.active = false
			return m, nil

		case "enter":
			return m.handleStepNavigation()

		case "tab":
			return m.nextStep()

		case "shift+tab":
			return m.prevStep()
		}
	}

	// Update the current input field
	var cmd tea.Cmd
	switch m.currentStep {
	case StepAgentType:
		m.agentType, cmd = m.agentType.Update(msg)
	case StepCount:
		m.count, cmd = m.count.Update(msg)
	case StepPrompt:
		m.prompt, cmd = m.prompt.Update(msg)
	}

	return m, cmd
}

// handleStepNavigation handles enter key navigation between steps
func (m AgentFormModel) handleStepNavigation() (AgentFormModel, tea.Cmd) {
	switch m.currentStep {
	case StepAgentType:
		if err := m.validateAgentType(); err != nil {
			m.error = err.Error()
			return m, nil
		}
		m.error = ""
		return m.nextStep()

	case StepCount:
		if err := m.validateCount(); err != nil {
			m.error = err.Error()
			return m, nil
		}
		m.error = ""
		return m.nextStep()

	case StepPrompt:
		if err := m.validatePrompt(); err != nil {
			m.error = err.Error()
			return m, nil
		}
		m.error = ""
		m.currentStep = StepComplete
		m.active = false
		return m, func() tea.Msg {
			return AgentFormSubmitMsg{
				AgentType: m.agentType.Value(),
				Count:     m.count.Value(),
				Prompt:    m.prompt.Value(),
			}
		}
	}

	return m, nil
}

// nextStep moves to the next form step
func (m AgentFormModel) nextStep() (AgentFormModel, tea.Cmd) {
	switch m.currentStep {
	case StepAgentType:
		m.currentStep = StepCount
		m.agentType.Blur()
		m.count.Focus()
	case StepCount:
		m.currentStep = StepPrompt
		m.count.Blur()
		m.prompt.Focus()
	}
	return m, nil
}

// prevStep moves to the previous form step
func (m AgentFormModel) prevStep() (AgentFormModel, tea.Cmd) {
	switch m.currentStep {
	case StepCount:
		m.currentStep = StepAgentType
		m.count.Blur()
		m.agentType.Focus()
	case StepPrompt:
		m.currentStep = StepCount
		m.prompt.Blur()
		m.count.Focus()
	}
	return m, nil
}

// validateAgentType validates the agent type input
func (m *AgentFormModel) validateAgentType() error {
	value := strings.TrimSpace(m.agentType.Value())
	if value == "" {
		return errors.New("Agent type is required")
	}

	validTypes := []string{"claude", "cursor", "codex", "gemini", "random"}
	for _, validType := range validTypes {
		if strings.EqualFold(value, validType) {
			return nil
		}
	}

	return fmt.Errorf("Invalid agent type. Valid types: %s", strings.Join(validTypes, ", "))
}

// validateCount validates the count input
func (m *AgentFormModel) validateCount() error {
	value := strings.TrimSpace(m.count.Value())
	if value == "" {
		return errors.New("Count is required")
	}

	count, err := strconv.Atoi(value)
	if err != nil {
		return errors.New("Count must be a number")
	}

	if count < 1 || count > 10 {
		return errors.New("Count must be between 1 and 10")
	}

	return nil
}

// validatePrompt validates the prompt input
func (m *AgentFormModel) validatePrompt() error {
	value := strings.TrimSpace(m.prompt.Value())
	if value == "" {
		return errors.New("Prompt is required")
	}

	if len(value) < 10 {
		return errors.New("Prompt must be at least 10 characters long")
	}

	return nil
}

// Validate ensures that all form data is valid
func (m *AgentFormModel) Validate() error {
	if err := m.validateAgentType(); err != nil {
		return err
	}
	if err := m.validateCount(); err != nil {
		return err
	}
	if err := m.validatePrompt(); err != nil {
		return err
	}
	return nil
}

// View renders the form
func (m AgentFormModel) View() string {
	if !m.active {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ClaudeSquadAccent).
		Padding(1, 2).
		Width(m.width - 4)

	title := ClaudeSquadPrimaryStyle.Bold(true).Render("Create New Agent")
	stepInfo := ClaudeSquadMutedStyle.Render(fmt.Sprintf("Step %d of 3", int(m.currentStep)+1))
	header := fmt.Sprintf("%s\n%s\n", title, stepInfo)

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n")

	// Agent Type input
	if m.currentStep == StepAgentType {
		content.WriteString(ClaudeSquadAccentStyle.Render("► Agent Type:"))
	} else {
		content.WriteString("  Agent Type:")
	}
	content.WriteString("\n")
	content.WriteString("  " + m.agentType.View())
	content.WriteString("\n\n")

	// Count input
	if m.currentStep == StepCount {
		content.WriteString(ClaudeSquadAccentStyle.Render("► Count:"))
	} else {
		content.WriteString("  Count:")
	}
	content.WriteString("\n")
	content.WriteString("  " + m.count.View())
	content.WriteString("\n\n")

	// Prompt input
	if m.currentStep == StepPrompt {
		content.WriteString(ClaudeSquadAccentStyle.Render("► Prompt:"))
	} else {
		content.WriteString("  Prompt:")
	}
	content.WriteString("\n")
	content.WriteString("  " + m.prompt.View())
	content.WriteString("\n\n")

	// Error display
	if m.error != "" {
		content.WriteString(ErrorStyle.Render("Error: " + m.error))
		content.WriteString("\n\n")
	}

	// Instructions
	instructions := ClaudeSquadMutedStyle.Render(
		"Enter: Next step  •  Tab: Next field  •  Shift+Tab: Previous field  •  Esc: Cancel")
	content.WriteString(instructions)

	return style.Render(content.String())
}

// AgentFormSubmitMsg is sent when the form is submitted
type AgentFormSubmitMsg struct {
	AgentType string
	Count     string
	Prompt    string
}
