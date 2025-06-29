// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nehpz/claudicus/pkg/activity"
	"github.com/nehpz/claudicus/pkg/config"
	"gopkg.in/yaml.v3"
)

// RefreshMsg is sent by the ticker to refresh sessions without clearing screen
type RefreshMsg struct{}

// TickMsg wraps time.Time for ticker messages
type TickMsg time.Time

// App represents the main TUI application
type App struct {
	uzi             UziInterface
	list            *ListModel
	diffPreview     *DiffPreviewModel
	broadcastInput  *BroadcastInputModel
	confirmModal    *ConfirmationModal
	checkpointModal CheckpointModal
	agentForm       AgentFormModel
	progressModal   ProgressModal
	keys            KeyMap
	ticker          *time.Ticker
	activityMonitor *activity.AgentActivityMonitor
	monitorCtx      context.Context
	monitorCancel   context.CancelFunc
	width           int
	height          int
	loading         bool
	splitView       bool // Toggle between list-only and split view
}

// NewApp creates a new TUI application instance
func NewApp(uzi UziInterface) *App {
	// Initialize the list view
	list := NewListModel(80, 24)               // Default size, will be updated on first render
	diffPreview := NewDiffPreviewModel(40, 24) // Default size, will be updated on first render
	broadcastInput := NewBroadcastInputModel()
	confirmModal := NewConfirmationModal()
	checkpointModal := NewCheckpointModal()
	agentForm := NewAgentFormModel()
	progressModal := NewProgressModal()

	activityMonitor := activity.NewAgentActivityMonitor()
	// Create context for the monitor with cancellation
	monitorCtx, monitorCancel := context.WithCancel(context.Background())

	// Start the activity monitor
	err := activityMonitor.Start(monitorCtx)
	if err != nil {
		// Log error but don't panic, allow app to work without monitor
		monitorCancel()
	}

	return &App{
		uzi:             uzi,
		list:            &list,
		diffPreview:     diffPreview,
		broadcastInput:  broadcastInput,
		confirmModal:    confirmModal,
		checkpointModal: checkpointModal,
		agentForm:       agentForm,
		progressModal:   progressModal,
		keys:            DefaultKeyMap(),
		ticker:          nil, // Will be created in Init
		activityMonitor: activityMonitor,
		monitorCtx:      monitorCtx,
		monitorCancel:   monitorCancel,
		loading:         true,
		splitView:       false, // Start in list view
	}
}

// tickEvery returns a command that sends TickMsg every duration
func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// refreshSessions returns a command that fetches sessions and sends RefreshMsg
func (a *App) refreshSessions() tea.Cmd {
	return func() tea.Msg {
		// Load sessions via UziInterface
		sessions, err := a.uzi.GetSessions()
		if err != nil {
			// For now, just return the refresh message even on error
			// In a production app, you might want to handle errors differently
			return RefreshMsg{}
		}
		// Get metrics from activity monitor (safe call - won't block)
		monitorMetrics := make(map[string]*activity.Metrics)
		if a.activityMonitor != nil {
			monitorMetrics = a.activityMonitor.UpdateAll()
		}

		// Merge monitor metrics into sessions (create new slice to avoid mutation)
		updatedSessions := make([]SessionInfo, len(sessions))
		copy(updatedSessions, sessions)
		for i, session := range updatedSessions {
			if metrics, exists := monitorMetrics[session.Name]; exists {
				updatedSessions[i].Insertions = metrics.Insertions
				updatedSessions[i].Deletions = metrics.Deletions
				// Map activity status to session status
				updatedSessions[i].Status = string(metrics.Status)
			}
		}
		sessions = updatedSessions

		// Update the list with new sessions
		a.list.LoadSessions(sessions)
		a.loading = false

		return RefreshMsg{}
	}
}

// Init implements tea.Model interface
func (a *App) Init() tea.Cmd {
	// Start the 2-second ticker and initial session load
	return tea.Batch(
		a.refreshSessions(),      // Load sessions immediately
		tickEvery(2*time.Second), // Start ticker for smooth updates
	)
}

// Update implements tea.Model interface
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirmation modal when visible
		if a.confirmModal != nil && a.confirmModal.IsVisible() {
			var modalCmd tea.Cmd
			a.confirmModal, modalCmd = a.confirmModal.Update(msg)
			return a, modalCmd
		}

		// Handle checkpoint modal when visible
		if a.checkpointModal.IsVisible() {
			var modalCmd tea.Cmd
			a.checkpointModal, modalCmd = a.checkpointModal.Update(msg)
			return a, modalCmd
		}

		// Handle agent form when active
		if a.agentForm.IsActive() {
			var formCmd tea.Cmd
			a.agentForm, formCmd = a.agentForm.Update(msg)
			return a, formCmd
		}

		// Handle progress modal when active
		if a.progressModal.IsActive() {
			var progressCmd tea.Cmd
			a.progressModal, progressCmd = a.progressModal.Update(msg)
			return a, progressCmd
		}

		// Handle broadcast input when active
		if a.broadcastInput.IsActive() {
			switch {
			case key.Matches(msg, a.keys.Enter):
				// Execute broadcast and deactivate input
				message := a.broadcastInput.Value()
				a.broadcastInput.SetActive(false)

				if message != "" {
					return a, func() tea.Msg {
						err := a.uzi.RunBroadcast(message)
						if err != nil {
							// Handle error - for now just continue
							return nil
						}
						// Refresh sessions after broadcast
						return RefreshMsg{}
					}
				}
				return a, nil

			case key.Matches(msg, a.keys.Escape):
				// Cancel broadcast input
				a.broadcastInput.SetActive(false)
				return a, nil

			default:
				// Delegate to broadcast input
				var cmd tea.Cmd
				a.broadcastInput, cmd = a.broadcastInput.Update(msg)
				return a, cmd
			}
		}

		// Handle key events
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit

		case key.Matches(msg, a.keys.Tab):
			// Toggle between list view and split view
			a.splitView = !a.splitView

			// When entering split view, load diff for selected session
			if a.splitView {
				if selected := a.list.SelectedSession(); selected != nil {
					a.diffPreview.LoadDiff(selected)
				}
			}
			return a, nil

		case key.Matches(msg, a.keys.Config):
			// Load the `uzi.yaml` configuration
			cfg, err := config.LoadConfig("uzi.yaml")
			if err != nil {
				// If file doesn't exist create a starter file
				if os.IsNotExist(err) {
					starterConfig := config.DefaultConfig()
					if err := writeStarterConfig("uzi.yaml", starterConfig); err != nil {
						return a, func() tea.Msg { return err }
					}
				} else {
					// Handle other errors loading config
					return a, func() tea.Msg { return err }
				}
			}

			// Open YAML editor
			if err := a.openEditor("uzi.yaml"); err != nil {
				return a, func() tea.Msg { return err }
			}

			// Validate configuration
			if cfg.DevCommand == nil || *cfg.DevCommand == "" {
				return a, func() tea.Msg { return fmt.Errorf("devCommand is empty") }
			}

			if cfg.PortRange == nil || !isValidPortRange(*cfg.PortRange) {
				return a, func() tea.Msg { return fmt.Errorf("Invalid portRange") }
			}

			return a, nil

		case key.Matches(msg, a.keys.Enter):
			// Handle session selection/attachment
			if selected := a.list.SelectedSession(); selected != nil {
				// Attach to the selected session
				return a, func() tea.Msg {
					err := a.uzi.AttachToSession(selected.Name)
					if err != nil {
						// Handle error - for now just continue
						return nil
					}
					return tea.Quit // Exit TUI after attaching
				}
			}

		case key.Matches(msg, a.keys.Kill):
			// Show confirmation modal for kill command
			if selected := a.list.SelectedSession(); selected != nil {
				// Extract agent name from session name for fat-finger protection
				agentName := extractAgentName(selected.Name)
				a.confirmModal.SetRequiredAgentName(agentName)
				a.confirmModal.SetVisible(true)
				return a, nil
			}

		case key.Matches(msg, a.keys.Broadcast):
			// Activate broadcast input prompt
			a.broadcastInput.SetActive(true)
			a.broadcastInput.SetWidth(a.width)
			return a, nil

		case key.Matches(msg, a.keys.ToggleCommits):
			// Toggle commits view in diff preview (only when in split view)
			if a.splitView {
				a.diffPreview.ToggleView()
			}
			return a, nil

		case key.Matches(msg, a.keys.FilterStuck):
			// Toggle stuck agents filter
			a.list.ToggleStuckFilter()
			return a, nil

		case key.Matches(msg, a.keys.FilterWorking):
			// Set working agents filter
			a.list.SetWorkingFilter()
			return a, nil

		case key.Matches(msg, a.keys.Clear):
			// Clear any active filter
			a.list.ClearFilter()
			return a, nil

		case key.Matches(msg, a.keys.Checkpoint):
			// Show checkpoint modal for selected agent
			if selected := a.list.SelectedSession(); selected != nil {
				// Get all sessions for agent selection
				sessions, err := a.uzi.GetSessions()
				if err == nil {
					a.checkpointModal.SetAgents(sessions)
					a.checkpointModal.SetSize(a.width, a.height)
					a.checkpointModal.SetVisible(true)
				}
				return a, nil
			}

		case key.Matches(msg, a.keys.NewAgent):
			// Show agent creation form
			a.agentForm.SetActive(true)
			a.agentForm.SetSize(a.width, a.height)
			return a, spinnerTick() // Start spinner for form
		}

		// In split view, handle navigation differently
		if a.splitView {
			// Track previous selection
			prevSelected := a.list.SelectedSession()

			// Delegate navigation to the list
			var cmd tea.Cmd
			model, cmd := a.list.Update(msg)
			if listModel, ok := model.(ListModel); ok {
				*a.list = listModel
			}
			cmds = append(cmds, cmd)

			// If selection changed, update diff view
			if newSelected := a.list.SelectedSession(); newSelected != nil {
				if prevSelected == nil || prevSelected.Name != newSelected.Name {
					a.diffPreview.LoadDiff(newSelected)
				}
			}

			return a, tea.Batch(cmds...)
		} else {
			// In list view, delegate all key events to the list
			var cmd tea.Cmd
			model, cmd := a.list.Update(msg)
			if listModel, ok := model.(ListModel); ok {
				*a.list = listModel
			}
			return a, cmd
		}

	case tea.WindowSizeMsg:
		// Update dimensions
		a.width = msg.Width
		a.height = msg.Height

		if a.splitView {
			// In split view, allocate space for both list and diff
			listWidth := msg.Width / 2
			diffWidth := msg.Width - listWidth

			a.list.SetSize(listWidth, msg.Height-2)
			a.diffPreview.SetSize(diffWidth, msg.Height-2)
		} else {
			// In list view, use full width
			a.list.SetSize(msg.Width, msg.Height-2)
		}

		// Delegate to components for their own size handling
		var cmd tea.Cmd
		model, cmd := a.list.Update(msg)
		if listModel, ok := model.(ListModel); ok {
			*a.list = listModel
		}
		cmds = append(cmds, cmd)

		if a.splitView {
			// DiffPreview doesn't need to handle its own updates
		}

		return a, tea.Batch(cmds...)

	case TickMsg:
		// Ticker fired - refresh sessions smoothly without clearing screen
		return a, tea.Batch(
			a.refreshSessions(),      // Refresh session data
			tickEvery(2*time.Second), // Schedule next tick
		)

	case RefreshMsg:
		// Sessions have been refreshed - no action needed
		// The list has already been updated in refreshSessions()
		return a, nil

	case AgentFormSubmitMsg:
		// Handle agent form submission
		a.agentForm.SetActive(false)
		a.progressModal.SetActive(true)
		a.progressModal.SetSize(a.width, a.height)

		// Start async agent creation
		opts := msg.AgentType + ":" + msg.Count + ":" + msg.Prompt
		progressChan, err := a.uzi.SpawnAgentInteractive(opts)
		if err != nil {
			a.progressModal.SetError(err.Error())
			return a, nil
		}

		// Start monitoring progress
		return a, func() tea.Msg {
			// Wait for completion in a goroutine
			go func() {
				select {
				case <-progressChan:
					// Agent creation completed successfully
					return // Success handled by the channel
				}
			}()
			return ProgressStepMsg{Step: ProgressStepSetupWorktree}
		}

	case ProgressStepMsg:
		// Update progress modal
		a.progressModal.NextStep()
		if msg.Message != "" {
			a.progressModal.SetMessage(msg.Message)
		}
		return a, nil

	case ProgressErrorMsg:
		// Handle progress error
		a.progressModal.SetError(msg.Error)
		return a, nil

	case ProgressCompleteMsg:
		// Handle completion
		a.progressModal.NextStep() // Move to complete step
		return a, tea.Batch(
			a.refreshSessions(), // Refresh to show new session
			func() tea.Msg {
				// Auto-close modal after a delay
				time.Sleep(2 * time.Second)
				a.progressModal.SetActive(false)
				return nil
			},
		)

	case SpinnerTickMsg:
		// Update spinner in progress modal
		var progressCmd tea.Cmd
		a.progressModal, progressCmd = a.progressModal.Update(msg)
		return a, progressCmd

	case CheckpointMsg:
		// Handle checkpoint request
		return a, func() tea.Msg {
			err := a.uzi.RunCheckpoint(msg.AgentName, msg.CommitMessage)
			if err != nil {
				return CheckpointCompleteMsg{Success: false, Error: err.Error()}
			}
			return CheckpointCompleteMsg{Success: true}
		}

	case CheckpointCompleteMsg:
		// Handle checkpoint completion
		a.checkpointModal.SetComplete(msg.Success, msg.Error)
		if msg.Success {
			// Refresh sessions after successful checkpoint
			return a, a.refreshSessions()
		}
		return a, nil

	case CheckpointProgressMsg:
		// Handle checkpoint progress updates
		a.checkpointModal.SetProgress(msg.Output, msg.IsError, msg.Conflicts)
		return a, nil

	case ModalMsg:
		// Handle confirmation modal response
		if msg.Confirmed {
			if selected := a.list.SelectedSession(); selected != nil {
				return a, func() tea.Msg {
					// Step 1: Kill the session
					err := a.uzi.KillSession(selected.Name)
					if err != nil {
						// Handle error - for now just continue
						return RefreshMsg{}
					}

					// Step 2: Remove worktree state if this is a kill & replace
					if msg.SpawnReplacement {
						// Get state manager and remove state
						if stateManager := getStateManager(); stateManager != nil {
							if err := stateManager.RemoveState(selected.Name); err != nil {
								// Log error but continue
							}
						}

						// Step 3: Spawn replacement agent
						newSessionName, err := a.uzi.SpawnAgent(msg.Prompt, msg.Model)
						if err != nil {
							// Handle spawn error - for now just continue
							return RefreshMsg{}
						}

						// Step 4: Add new session to state
						// The SpawnAgent call should have already handled this via RunPrompt
						_ = newSessionName // Session is now tracked automatically
					}

					// Refresh sessions to show updated list
					return RefreshMsg{}
				}
			}
		}
		// If not confirmed, just continue - modal is already hidden
		return a, nil

	default:
		// Update confirmation modal
		if a.confirmModal != nil {
			var modalCmd tea.Cmd
			a.confirmModal, modalCmd = a.confirmModal.Update(msg)
			cmds = append(cmds, modalCmd)
		}

		// Delegate other messages to appropriate components
		var cmd tea.Cmd
		model, cmd := a.list.Update(msg)
		if listModel, ok := model.(ListModel); ok {
			*a.list = listModel
		}
		cmds = append(cmds, cmd)

		return a, tea.Batch(cmds...)
	}
}

// View implements tea.Model interface - delegates to list view
func (a *App) View() string {
	// If we don't have proper dimensions yet, return a simple message
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	if a.splitView {
		// Split view: show list on left and diff on right
		listView := a.list.View()
		diffView := a.diffPreview.View()

		// Join horizontally with Claude Squad styling
		splitContent := lipgloss.JoinHorizontal(lipgloss.Top, listView, diffView)

		// Add status line if loading
		if a.loading {
			statusLine := ClaudeSquadMutedStyle.Render("Refreshing sessions...")
			return splitContent + "\n" + statusLine
		}

		// Add broadcast input if active
		content := splitContent
		if a.broadcastInput.IsActive() {
			broadcastView := a.broadcastInput.View()
			content = lipgloss.JoinVertical(lipgloss.Left, content, broadcastView)
		}

		return content
	} else {
		// List view: delegate to the list view for rendering
		listView := a.list.View()

		// Add status lines
		var statusLines []string

		if a.loading {
			statusLines = append(statusLines, ClaudeSquadMutedStyle.Render("Refreshing sessions..."))
		}

		// Add filter status if active
		if filterStatus := a.list.GetFilterStatus(); filterStatus != "" {
			statusLines = append(statusLines, ClaudeSquadAccentStyle.Render(filterStatus))
		}

		if len(statusLines) > 0 {
			statusLine := strings.Join(statusLines, " │ ")
			listView = listView + "\n" + statusLine
		}

		// Add broadcast input if active
		if a.broadcastInput.IsActive() {
			broadcastView := a.broadcastInput.View()
			listView = lipgloss.JoinVertical(lipgloss.Left, listView, broadcastView)
		}

		// Add agent form if active
		if a.agentForm.IsActive() {
			formView := a.agentForm.View()
			listView = lipgloss.JoinVertical(lipgloss.Left, listView, formView)
		}

		// Add progress modal if active
		if a.progressModal.IsActive() {
			modalView := a.progressModal.View()
			listView = lipgloss.JoinVertical(lipgloss.Left, listView, modalView)
		}

		// Add confirmation modal if visible
		if a.confirmModal != nil && a.confirmModal.IsVisible() {
			modalView := a.confirmModal.View()
			listView = lipgloss.JoinVertical(lipgloss.Left, listView, modalView)
		}

		// Add checkpoint modal if visible
		if a.checkpointModal.IsVisible() {
			modalView := a.checkpointModal.View()
			listView = lipgloss.JoinVertical(lipgloss.Left, listView, modalView)
		}

		return listView
	}
}

// Cleanup stops the activity monitor and releases resources
func (a *App) Cleanup() {
	if a.activityMonitor != nil {
		a.activityMonitor.Stop()
	}
	if a.monitorCancel != nil {
		a.monitorCancel()
	}
}

// getStateManager returns a state manager instance for worktree operations
func getStateManager() StateManagerInterface {
	// Create a new StateManagerBridge instance
	return NewStateManagerBridge()
}

// openEditor opens the specified file in the system editor
func (a *App) openEditor(filepath string) error {
	// Try to use EDITOR environment variable first, fallback to sensible defaults
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fallback to common editors
		for _, fallback := range []string{"nano", "vi", "vim"} {
			if _, err := exec.LookPath(fallback); err == nil {
				editor = fallback
				break
			}
		}
	}

	if editor == "" {
		return fmt.Errorf("no suitable editor found. Please set EDITOR environment variable")
	}

	// Run the editor
	cmd := exec.Command(editor, filepath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// writeStarterConfig writes a default configuration to the specified file
func writeStarterConfig(filename string, config config.Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// isValidPortRange validates a port range string (e.g., "3000-3010")
func isValidPortRange(portRange string) bool {
	if portRange == "" {
		return false
	}

	// Match pattern like "3000-3010" or "8080"
	pattern := `^(\d+)(?:-(\d+))?$`
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	matches := regex.FindStringSubmatch(portRange)
	if len(matches) == 0 {
		return false
	}

	// Parse start port
	startPort, err := strconv.Atoi(matches[1])
	if err != nil || startPort < 1 || startPort > 65535 {
		return false
	}

	// Parse end port if provided
	if matches[2] != "" {
		endPort, err := strconv.Atoi(matches[2])
		if err != nil || endPort < 1 || endPort > 65535 || endPort < startPort {
			return false
		}
	}

	return true
}
