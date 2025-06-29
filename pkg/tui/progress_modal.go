package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressStep represents a step in the progress
type ProgressStep int

const (
	ProgressStepSetupWorktree ProgressStep = iota
	ProgressStepCreateTmux
	ProgressStepStartAgent
	ProgressStepComplete
)

// ProgressModal shows the progress of agent creation
type ProgressModal struct {
	active      bool
	currentStep ProgressStep
	steps       []string
	width       int
	height      int
	spinner     []string
	spinnerIdx  int
	message     string
	error       string
}

// NewProgressModal creates a new progress modal
func NewProgressModal() ProgressModal {
	return ProgressModal{
		active: false,
		steps: []string{
			"Setting up git worktree...",
			"Creating tmux session...",
			"Starting agent...",
			"Complete!",
		},
		spinner:    []string{"|", "/", "-", "\\"},
		spinnerIdx: 0,
	}
}

// SetActive sets the modal's active state
func (m *ProgressModal) SetActive(active bool) {
	m.active = active
	if active {
		m.currentStep = ProgressStepSetupWorktree
		m.error = ""
		m.message = ""
	}
}

// IsActive returns whether the modal is currently active
func (m *ProgressModal) IsActive() bool {
	return m.active
}

// SetSize updates the modal dimensions
func (m *ProgressModal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// NextStep advances to the next step
func (m *ProgressModal) NextStep() {
	if m.currentStep < ProgressStepComplete {
		m.currentStep++
	}
}

// SetError sets an error message and stops progress
func (m *ProgressModal) SetError(err string) {
	m.error = err
}

// SetMessage sets a custom message
func (m *ProgressModal) SetMessage(msg string) {
	m.message = msg
}

// Update handles progress modal updates
func (m ProgressModal) Update(msg tea.Msg) (ProgressModal, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.currentStep == ProgressStepComplete || m.error != "" {
				m.active = false
				return m, nil
			}
		}
	case SpinnerTickMsg:
		m.spinnerIdx = (m.spinnerIdx + 1) % len(m.spinner)
		return m, spinnerTick()
	}

	return m, nil
}

// View renders the progress modal
func (m ProgressModal) View() string {
	if !m.active {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ClaudeSquadAccent).
		Padding(1, 2).
		Width(m.width / 2).
		Align(lipgloss.Center)

	title := ClaudeSquadPrimaryStyle.Bold(true).Render("Creating Agent")
	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n\n")

	// Show error if present
	if m.error != "" {
		content.WriteString(ErrorStyle.Render("Error: " + m.error))
		content.WriteString("\n\n")
		content.WriteString(ClaudeSquadMutedStyle.Render("Press Esc to close"))
		return style.Render(content.String())
	}

	// Show progress steps
	for i, step := range m.steps {
		stepNum := ProgressStep(i)
		if stepNum < m.currentStep {
			// Completed step
			content.WriteString(ClaudeSquadAccentStyle.Render("✓ " + step))
		} else if stepNum == m.currentStep {
			// Current step with spinner
			spinner := m.spinner[m.spinnerIdx]
			content.WriteString(ClaudeSquadPrimaryStyle.Render(spinner + " " + step))
		} else {
			// Future step
			content.WriteString(ClaudeSquadMutedStyle.Render("• " + step))
		}
		content.WriteString("\n")
	}

	// Show custom message if present
	if m.message != "" {
		content.WriteString("\n")
		content.WriteString(ClaudeSquadMutedStyle.Render(m.message))
		content.WriteString("\n")
	}

	// Show completion message
	if m.currentStep == ProgressStepComplete {
		content.WriteString("\n")
		content.WriteString(ClaudeSquadAccentStyle.Render("Agent created successfully!"))
		content.WriteString("\n")
		content.WriteString(ClaudeSquadMutedStyle.Render("Press Esc to close"))
	}

	return style.Render(content.String())
}

// SpinnerTickMsg is sent to update the spinner
type SpinnerTickMsg struct{}

// spinnerTick returns a command that sends a SpinnerTickMsg after a short delay
func spinnerTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

// ProgressStepMsg is sent when a progress step is completed
type ProgressStepMsg struct {
	Step    ProgressStep
	Message string
}

// ProgressErrorMsg is sent when an error occurs during progress
type ProgressErrorMsg struct {
	Error string
}

// ProgressCompleteMsg is sent when the entire process is complete
type ProgressCompleteMsg struct {
	SessionName string
}
