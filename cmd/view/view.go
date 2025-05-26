package view

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs      = flag.NewFlagSet("uzi view", flag.ExitOnError)
	CmdView = &ffcli.Command{
		Name:       "view",
		ShortUsage: "uzi view",
		ShortHelp:  "Interactive view of agent sessions",
		FlagSet:    fs,
		Exec:       executeView,
	}
)

type sessionItem struct {
	name   string
	status string
	output string
}

func (i sessionItem) FilterValue() string { return i.name }
func (i sessionItem) Title() string       { return i.name }
func (i sessionItem) Description() string { return i.status }

type model struct {
	list     list.Model
	viewport viewport.Model
	sessions []sessionItem
	ready    bool
	width    int
	height   int
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"})
)

func initialModel() model {
	sessions := getAgentSessions()
	
	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		items[i] = session
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Agent Sessions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	vp := viewport.New(0, 0)

	return model{
		list:     l,
		viewport: vp,
		sessions: sessions,
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.width = msg.Width
			m.height = msg.Height
			m.list.SetWidth(msg.Width / 3)
			m.list.SetHeight(msg.Height - 2)
			m.viewport.Width = (msg.Width * 2 / 3) - 2
			m.viewport.Height = msg.Height - 4
			m.ready = true
			
			if len(m.sessions) > 0 {
				m.viewport.SetContent(m.sessions[0].output)
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			return m, tea.Batch(
				func() tea.Msg {
					newSessions := getAgentSessions()
					items := make([]list.Item, len(newSessions))
					for i, session := range newSessions {
						items[i] = session
					}
					return refreshMsg{sessions: newSessions, items: items}
				},
			)
		}

		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		
		if selected := m.list.SelectedItem(); selected != nil {
			if session, ok := selected.(sessionItem); ok {
				m.viewport.SetContent(session.output)
			}
		}
		
		return m, cmd

	case refreshMsg:
		m.sessions = msg.sessions
		m.list.SetItems(msg.items)
		if len(m.sessions) > 0 {
			m.viewport.SetContent(m.sessions[0].output)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

type refreshMsg struct {
	sessions []sessionItem
	items    []list.Item
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	listView := lipgloss.NewStyle().
		Width(m.width / 3).
		Height(m.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Render(m.list.View())

	viewportContent := m.viewport.View()
	viewportView := lipgloss.NewStyle().
		Width((m.width * 2 / 3) - 2).
		Height(m.height - 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Render(viewportContent)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, listView, viewportView)

	statusBar := statusBarStyle.Width(m.width).Render(
		"Sessions: " + fmt.Sprintf("%d", len(m.sessions)) + 
		" | Press 'r' to refresh, 'q' to quit, ↑/↓ to navigate",
	)

	return lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar)
}

func getAgentSessions() []sessionItem {
	cmd := exec.Command("tmux", "ls")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return []sessionItem{{
			name:   "Error",
			status: "Failed to get sessions",
			output: fmt.Sprintf("Error listing tmux sessions: %v", err),
		}}
	}

	lines := strings.Split(out.String(), "\n")
	var sessions []sessionItem
	
	for _, line := range lines {
		if strings.HasPrefix(line, "agent-") {
			sessionName := strings.SplitN(line, ":", 2)[0]
			
			parts := strings.Split(line, " ")
			status := "unknown"
			if len(parts) > 1 {
				if strings.Contains(line, "(attached)") {
					status = "attached"
				} else {
					status = "detached"
				}
			}
			
			output := getSessionOutput(sessionName)
			
			sessions = append(sessions, sessionItem{
				name:   sessionName,
				status: status,
				output: output,
			})
		}
	}

	if len(sessions) == 0 {
		return []sessionItem{{
			name:   "No Sessions",
			status: "No agent sessions found",
			output: "No agent sessions are currently running.\n\nTo create an agent session, you might need to:\n1. Start your agent processes\n2. Ensure tmux sessions are named with 'agent-' prefix",
		}}
	}

	return sessions
}

func getSessionOutput(sessionName string) string {
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p")
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("Error capturing session output: %v", err)
	}
	
	output := out.String()
	if strings.TrimSpace(output) == "" {
		return fmt.Sprintf("Session '%s' is active but has no visible output.", sessionName)
	}
	
	return output
}

func executeView(ctx context.Context, args []string) error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}