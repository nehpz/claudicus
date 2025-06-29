package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CheckpointStep int

const (
	CheckpointStepSelectAgent CheckpointStep = iota
	CheckpointStepCommitMessage
	CheckpointStepProgress
	CheckpointStepComplete
)

type CheckpointModal struct {
	visible      bool
	currentStep  CheckpointStep
	agents       []SessionInfo // Available agents for selection
	selectedIdx  int           // Index of selected agent
	commitInput  textinput.Model
	progressText string
	conflicts    []string
	spinner      spinner.Model
	error        string
	width        int
	height       int
	completed    bool
}

// CheckpointMsg is sent when checkpoint operation is initiated
type CheckpointMsg struct {
	AgentName     string
	CommitMessage string
}

// CheckpointProgressMsg is sent during git rebase progress
type CheckpointProgressMsg struct {
	Output    string
	IsError   bool
	Conflicts []string
}

// CheckpointCompleteMsg is sent when checkpoint is complete
type CheckpointCompleteMsg struct {
	Success bool
	Error   string
}

func NewCheckpointModal() CheckpointModal {
	// Create text input for commit message
	commitInput := textinput.New()
	commitInput.Placeholder = "Enter commit message"
	commitInput.Focus()
	commitInput.CharLimit = 100
	commitInput.Width = 50

	// Create spinner for progress
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return CheckpointModal{
		visible:     false,
		currentStep: CheckpointStepSelectAgent,
		commitInput: commitInput,
		spinner:     s,
		selectedIdx: 0,
	}
}

func (m *CheckpointModal) SetVisible(visible bool) {
	m.visible = visible
	if visible {
		m.reset()
	}
}

func (m *CheckpointModal) IsVisible() bool {
	return m.visible
}

func (m *CheckpointModal) SetAgents(agents []SessionInfo) {
	m.agents = agents
	if len(agents) > 0 && m.selectedIdx >= len(agents) {
		m.selectedIdx = 0
	}
}

func (m *CheckpointModal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *CheckpointModal) SetProgress(output string, isError bool, conflicts []string) {
	m.progressText = output
	m.conflicts = conflicts
	if isError {
		m.error = output
	}
}

func (m *CheckpointModal) SetComplete(success bool, errorMsg string) {
	if success {
		m.currentStep = CheckpointStepComplete
		m.completed = true
	} else {
		m.error = errorMsg
	}
}

func (m *CheckpointModal) reset() {
	m.currentStep = CheckpointStepSelectAgent
	m.selectedIdx = 0
	m.commitInput.SetValue("")
	m.progressText = ""
	m.conflicts = nil
	m.error = ""
	m.completed = false
}

func (m CheckpointModal) Update(msg tea.Msg) (CheckpointModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.currentStep {
		case CheckpointStepSelectAgent:
			switch msg.String() {
			case "up", "k":
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			case "down", "j":
				if m.selectedIdx < len(m.agents)-1 {
					m.selectedIdx++
				}
			case "enter":
				if len(m.agents) > 0 {
					m.currentStep = CheckpointStepCommitMessage
					m.commitInput.Focus()
				}
			case "esc":
				m.visible = false
				return m, nil
			}

		case CheckpointStepCommitMessage:
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.commitInput.Value()) != "" && len(m.agents) > 0 {
					m.currentStep = CheckpointStepProgress
					m.commitInput.Blur()
					// Start spinner
					cmds = append(cmds, m.spinner.Tick)
					// Send checkpoint message
					selectedAgent := m.agents[m.selectedIdx]
					return m, tea.Batch(append(cmds, func() tea.Msg {
						return CheckpointMsg{
							AgentName:     selectedAgent.AgentName,
							CommitMessage: strings.TrimSpace(m.commitInput.Value()),
						}
					})...)
				}
			case "esc":
				m.currentStep = CheckpointStepSelectAgent
				m.commitInput.Blur()
			default:
				var cmd tea.Cmd
				m.commitInput, cmd = m.commitInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case CheckpointStepProgress:
			switch msg.String() {
			case "esc":
				if m.completed {
					m.visible = false
					return m, nil
				}
				// Can't escape during progress
			}

		case CheckpointStepComplete:
			switch msg.String() {
			case "enter", "esc":
				m.visible = false
				return m, nil
			}
		}

	case CheckpointProgressMsg:
		m.SetProgress(msg.Output, msg.IsError, msg.Conflicts)
		if len(msg.Conflicts) > 0 {
			// Handle conflicts - for now just display them
		}
		cmds = append(cmds, m.spinner.Tick)

	case CheckpointCompleteMsg:
		m.SetComplete(msg.Success, msg.Error)

	case spinner.TickMsg:
		if m.currentStep == CheckpointStepProgress && !m.completed {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m CheckpointModal) View() string {
	if !m.visible {
		return ""
	}

	var content string
	title := ClaudeSquadAccentStyle.Render("🔄 Checkpoint Agent")

	switch m.currentStep {
	case CheckpointStepSelectAgent:
		content = m.renderAgentSelection()

	case CheckpointStepCommitMessage:
		selectedAgent := ""
		if len(m.agents) > 0 && m.selectedIdx < len(m.agents) {
			selectedAgent = m.agents[m.selectedIdx].AgentName
		}
		content = fmt.Sprintf("Agent: %s\n\n%s\n\n%s",
			ClaudeSquadSelectedStyle.Render(selectedAgent),
			m.commitInput.View(),
			ClaudeSquadMutedStyle.Render("Press Enter to commit, Esc to go back"))

	case CheckpointStepProgress:
		content = m.renderProgress()

	case CheckpointStepComplete:
		if m.error != "" {
			content = fmt.Sprintf("❌ Error: %s\n\n%s",
				ErrorStyle.Render(m.error),
				ClaudeSquadMutedStyle.Render("Press Enter or Esc to close"))
		} else {
			content = fmt.Sprintf("✅ %s\n\n%s",
				ClaudeSquadAccentStyle.Render("Checkpoint completed successfully!"),
				ClaudeSquadMutedStyle.Render("Press Enter or Esc to close"))
		}
	}

	// Create modal container
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Width(60).
		Align(lipgloss.Center)

	modalContent := fmt.Sprintf("%s\n\n%s", title, content)
	modal := modalStyle.Render(modalContent)

	// Center the modal
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func (m CheckpointModal) renderAgentSelection() string {
	if len(m.agents) == 0 {
		return ErrorStyle.Render("No agents available for checkpoint")
	}

	var items []string
	items = append(items, ClaudeSquadMutedStyle.Render("Select an agent to checkpoint:"))
	items = append(items, "")

	for i, agent := range m.agents {
		prefix := "  "
		style := ClaudeSquadPrimaryStyle
		if i == m.selectedIdx {
			prefix = "▶ "
			style = ClaudeSquadSelectedStyle
		}
		
		status := agent.Status
		if status == "" {
			status = "unknown"
		}
		
		line := fmt.Sprintf("%s%s (%s)", prefix, agent.AgentName, status)
		items = append(items, style.Render(line))
	}

	items = append(items, "")
	items = append(items, ClaudeSquadMutedStyle.Render("↑/↓ to navigate, Enter to select, Esc to cancel"))

	return strings.Join(items, "\n")
}

func (m CheckpointModal) renderProgress() string {
	var lines []string
	
	selectedAgent := ""
	commitMsg := ""
	if len(m.agents) > 0 && m.selectedIdx < len(m.agents) {
		selectedAgent = m.agents[m.selectedIdx].AgentName
	}
	commitMsg = strings.TrimSpace(m.commitInput.Value())

	lines = append(lines, fmt.Sprintf("Agent: %s", ClaudeSquadSelectedStyle.Render(selectedAgent)))
	lines = append(lines, fmt.Sprintf("Message: %s", ClaudeSquadPrimaryStyle.Render(commitMsg)))
	lines = append(lines, "")

	if m.error != "" {
		lines = append(lines, ErrorStyle.Render("❌ Error: "+m.error))
	} else if m.completed {
		lines = append(lines, ClaudeSquadAccentStyle.Render("✅ Checkpoint completed!"))
	} else {
		lines = append(lines, fmt.Sprintf("%s Running git rebase...", m.spinner.View()))
	}

	if len(m.conflicts) > 0 {
		lines = append(lines, "")
		lines = append(lines, ErrorStyle.Render("⚠️  Conflicts detected:"))
		for _, conflict := range m.conflicts {
			lines = append(lines, "  "+ClaudeSquadMutedStyle.Render(conflict))
		}
	}

	if m.progressText != "" {
		lines = append(lines, "")
		lines = append(lines, ClaudeSquadMutedStyle.Render("Output:"))
		// Limit output length to avoid huge modals
		outputLines := strings.Split(m.progressText, "\n")
		maxLines := 10
		if len(outputLines) > maxLines {
			outputLines = outputLines[len(outputLines)-maxLines:]
		}
		for _, line := range outputLines {
			if strings.TrimSpace(line) != "" {
				lines = append(lines, "  "+ClaudeSquadMutedStyle.Render(line))
			}
		}
	}

	if m.completed {
		lines = append(lines, "")
		lines = append(lines, ClaudeSquadMutedStyle.Render("Press Enter or Esc to close"))
	}

	return strings.Join(lines, "\n")
}
