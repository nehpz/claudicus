package view

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"uzi/pkg/state"

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
func (i sessionItem) Description() string { 
	var statusColor lipgloss.Color
	var statusIcon string
	
	switch i.status {
	case "attached":
		statusColor = lipgloss.Color("#22C55E") // green
		statusIcon = "●"
	case "detached":
		statusColor = lipgloss.Color("#F59E0B") // amber  
		statusIcon = "○"
	case "error":
		statusColor = lipgloss.Color("#EF4444") // red
		statusIcon = "✗"
	default:
		statusColor = lipgloss.Color("#6B7280") // gray
		statusIcon = "◯"
	}
	
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	return statusStyle.Render(statusIcon + " " + i.status)
}

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
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3B82F6")).
			Padding(0, 1).
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#F9FAFB"}).
			Background(lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"})

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"})

	listBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"})

	viewportBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}).
			Padding(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	noSessionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)
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
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
	)
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
		case "d":
			if selected := m.list.SelectedItem(); selected != nil {
				if session, ok := selected.(sessionItem); ok {
					return m, tea.Batch(
						func() tea.Msg {
							deleteSession(session.name)
							newSessions := getAgentSessions()
							items := make([]list.Item, len(newSessions))
							for i, session := range newSessions {
								items[i] = session
							}
							return refreshMsg{sessions: newSessions, items: items}
						},
					)
				}
			}
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

	case tickMsg:
		return m, tea.Batch(
			tickCmd(),
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
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

type refreshMsg struct {
	sessions []sessionItem
	items    []list.Item
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) View() string {
	if !m.ready {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true)
		return "\n  " + loadingStyle.Render("Initializing...")
	}

	listView := listBorderStyle.
		Width(m.width / 3).
		Height(m.height - 2).
		Render(m.list.View())

	viewportContent := m.viewport.View()
	viewportView := viewportBorderStyle.
		Width((m.width * 2 / 3) - 2).
		Height(m.height - 4).
		Render(viewportContent)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, listView, viewportView)

	sessionCount := len(m.sessions)
	var sessionCountColor lipgloss.Color
	if sessionCount == 0 {
		sessionCountColor = lipgloss.Color("#6B7280") // gray
	} else {
		sessionCountColor = lipgloss.Color("#22C55E") // green
	}
	
	sessionCountStyle := lipgloss.NewStyle().Foreground(sessionCountColor).Bold(true)
	statusBar := statusBarStyle.Width(m.width).Render(
		"Sessions: " + sessionCountStyle.Render(fmt.Sprintf("%d", sessionCount)) + 
		" | Press " + helpStyle.Render("r") + " to refresh, " + 
		helpStyle.Render("d") + " to delete, " + 
		helpStyle.Render("q") + " to quit, " + 
		helpStyle.Render("↑/↓") + " to navigate",
	)

	return lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar)
}

func getAgentSessions() []sessionItem {
	stateManager := state.NewStateManager()
	if stateManager == nil {
		return []sessionItem{{
			name:   "Error",
			status: "error",
			output: errorStyle.Render("Error creating state manager"),
		}}
	}

	activeSessions, err := stateManager.GetActiveSessionsForRepo()
	if err != nil {
		return []sessionItem{{
			name:   "Error", 
			status: "error",
			output: errorStyle.Render("Error getting active sessions for repo: ") + err.Error(),
		}}
	}

	var sessions []sessionItem
	
	for _, sessionName := range activeSessions {
		cmd := exec.Command("tmux", "ls")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		
		status := "detached"
		if err == nil {
			lines := strings.Split(out.String(), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, sessionName+":") {
					if strings.Contains(line, "(attached)") {
						status = "attached"
					}
					break
				}
			}
		}
		
		output := getSessionOutput(sessionName)
		
		sessions = append(sessions, sessionItem{
			name:   sessionName,
			status: status,
			output: output,
		})
	}

	if len(sessions) == 0 {
		return []sessionItem{{
			name:   "No Sessions",
			status: "empty",
			output: noSessionsStyle.Render("No agent sessions are currently running for this git workspace.\n\n") +
				"To create an agent session for this workspace:\n" +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Render("1. Run 'uzi prompt' to start a new session\n") +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Render("2. Ensure you're in a git repository"),
		}}
	}

	return sessions
}

func getSessionOutput(sessionName string) string {
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p")
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return errorStyle.Render("Error capturing session output: ") + err.Error()
	}
	
	output := out.String()
	if strings.TrimSpace(output) == "" {
		return noSessionsStyle.Render(fmt.Sprintf("Session '%s' is active but has no visible output.", sessionName))
	}
	
	return output
}

func deleteSession(sessionName string) {
	cmd := exec.Command("uzi", "kill", sessionName)
	cmd.Run()
}

func executeView(ctx context.Context, args []string) error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}