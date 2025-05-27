package watch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"uzi/pkg/state"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type AgentWatcher struct {
	stateManager    *state.StateManager
	watchedSessions map[string]*SessionMonitor
	quit            chan bool
	mu              sync.RWMutex
}

type SessionMonitor struct {
	sessionName    string
	prevOutputHash []byte
	lastUpdated    time.Time
	updateCount    int
	noUpdateCount  int
	stopChan       chan bool
}

func NewAgentWatcher() *AgentWatcher {
	return &AgentWatcher{
		stateManager:    state.NewStateManager(),
		watchedSessions: make(map[string]*SessionMonitor),
		quit:            make(chan bool),
	}
}

func (aw *AgentWatcher) hashContent(content []byte) []byte {
	hash := sha256.Sum256(content)
	return hash[:]
}

func (aw *AgentWatcher) capturePaneContent(sessionName string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName+":agent", "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (aw *AgentWatcher) sendKeys(sessionName string, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName+":agent", keys)
	return cmd.Run()
}

func (aw *AgentWatcher) tapEnter(sessionName string) error {
	return aw.sendKeys(sessionName, "Enter")
}

func (aw *AgentWatcher) updateSessionStatus(sessionName, status string) error {
	// Get current state to preserve other fields
	currentState, err := aw.stateManager.GetWorktreeInfo(sessionName)
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	// Update the state, preserving other fields
	return aw.stateManager.SaveStateWithPort(currentState.Prompt, currentState.BranchName, sessionName, currentState.WorktreePath, currentState.Port)
}

func (aw *AgentWatcher) hasUpdated(sessionName string) (bool, bool, error) {
	content, err := aw.capturePaneContent(sessionName)
	if err != nil {
		return false, false, err
	}

	// Check for specific prompts that need auto-enter
	hasPrompt := false

	// Check for Claude trust prompt
	if strings.Contains(content, "Do you trust the files in this folder?") {
		hasPrompt = true
	}

	// Check for general continuation prompts
	if strings.Contains(content, "Press Enter to continue") ||
		strings.Contains(content, "Continue? (Y/n)") ||
		strings.Contains(content, "Do you want to proceed?") ||
		strings.Contains(content, "Do you want to") ||
		strings.Contains(content, "Proceed? (y/N)") {
		hasPrompt = true
	}

	aw.mu.RLock()
	monitor, exists := aw.watchedSessions[sessionName]
	aw.mu.RUnlock()

	if !exists {
		// Session monitor should have been created by refreshActiveSessions
		return false, hasPrompt, nil
	}

	// Compare current content hash with previous
	currentHash := aw.hashContent([]byte(content))
	if monitor.prevOutputHash == nil || !bytes.Equal(currentHash, monitor.prevOutputHash) {
		monitor.prevOutputHash = currentHash
		monitor.lastUpdated = time.Now()
		monitor.updateCount++
		monitor.noUpdateCount = 0
		return true, hasPrompt, nil
	}

	monitor.noUpdateCount++
	return false, hasPrompt, nil
}

func (aw *AgentWatcher) checkSessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName+":agent")
	return cmd.Run() == nil
}

func (aw *AgentWatcher) watchSession(sessionName string, stopChan chan bool) {
	log.Info("Starting to watch session", "session", sessionName)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-aw.quit:
			return
		case <-stopChan:
			log.Info("Stopping watch for session", "session", sessionName)
			return
		case <-ticker.C:
			// First check if session still exists in state
			activeSessions, err := aw.stateManager.GetActiveSessionsForRepo()
			if err != nil {
				log.Error("Failed to get active sessions", "error", err)
				continue
			}

			sessionActive := false
			for _, activeSession := range activeSessions {
				if activeSession == sessionName {
					sessionActive = true
					break
				}
			}

			if !sessionActive {
				log.Info("Session no longer in state, stopping watch", "session", sessionName)
				return
			}

			// Check if tmux session exists
			if !aw.checkSessionExists(sessionName) {
				log.Info("Tmux session no longer exists, stopping watch", "session", sessionName)
				return
			}

			updated, hasPrompt, err := aw.hasUpdated(sessionName)
			if err != nil {
				// Check if it's a session not found error
				if strings.Contains(err.Error(), "session not found") ||
					strings.Contains(err.Error(), "can't find session") {
					log.Info("Session not found, stopping watch", "session", sessionName)
					return
				}
				log.Error("Error checking session update", "session", sessionName, "error", err)
				continue
			}

			if updated {
				log.Debug("Session updated", "session", sessionName)
			}

			if hasPrompt {
				log.Info("Auto-pressing Enter for prompt", "session", sessionName)
				if err := aw.tapEnter(sessionName); err != nil {
					log.Error("Failed to send Enter", "session", sessionName, "error", err)
				} else {
					log.Info("Successfully sent Enter", "session", sessionName)
				}
			}
		}
	}
}

func (aw *AgentWatcher) refreshActiveSessions() error {
	activeSessions, err := aw.stateManager.GetActiveSessionsForRepo()
	if err != nil {
		return fmt.Errorf("failed to get active sessions: %w", err)
	}

	aw.mu.Lock()
	defer aw.mu.Unlock()

	// Stop watching sessions that are no longer active
	for sessionName, monitor := range aw.watchedSessions {
		found := false
		for _, activeSession := range activeSessions {
			if activeSession == sessionName {
				found = true
				break
			}
		}
		if !found {
			log.Info("Session no longer active, stopping watch", "session", sessionName)
			// Signal the goroutine to stop
			close(monitor.stopChan)
			delete(aw.watchedSessions, sessionName)
		}
	}

	// Start watching new active sessions
	for _, sessionName := range activeSessions {
		if _, exists := aw.watchedSessions[sessionName]; !exists {
			// Create a new monitor with stop channel
			monitor := &SessionMonitor{
				sessionName:    sessionName,
				prevOutputHash: nil,
				lastUpdated:    time.Now(),
				updateCount:    0,
				noUpdateCount:  0,
				stopChan:       make(chan bool),
			}
			aw.watchedSessions[sessionName] = monitor
			go aw.watchSession(sessionName, monitor.stopChan)
		}
	}

	return nil
}

func (aw *AgentWatcher) Start() {
	log.Info("Starting Agent Watcher")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Refresh active sessions more frequently to catch killed sessions faster
	refreshTicker := time.NewTicker(1 * time.Second)
	defer refreshTicker.Stop()

	go func() {
		for {
			select {
			case <-refreshTicker.C:
				if err := aw.refreshActiveSessions(); err != nil {
					log.Error("Failed to refresh active sessions", "error", err)
				}
			case <-aw.quit:
				return
			}
		}
	}()

	// Initial refresh
	if err := aw.refreshActiveSessions(); err != nil {
		log.Error("Failed initial session refresh", "error", err)
	}

	// Wait for signal
	<-sigChan
	log.Info("Shutting down Agent Watcher")

	// Stop all watchers
	aw.mu.Lock()
	for _, monitor := range aw.watchedSessions {
		close(monitor.stopChan)
	}
	aw.mu.Unlock()

	close(aw.quit)
}

var CmdWatch = &ffcli.Command{
	Name:       "watch",
	ShortUsage: "uzi watch",
	ShortHelp:  "Watch all active agent sessions and automatically press Enter when needed",
	LongHelp: `
The watch command monitors all active agent sessions in the current repository
and automatically presses Enter when it detects prompts that require user input,
such as trust prompts or continuation confirmations.

This is useful for hands-free operation of multiple agents.
`,
	FlagSet: func() *flag.FlagSet {
		fs := flag.NewFlagSet("watch", flag.ExitOnError)
		return fs
	}(),
	Exec: func(ctx context.Context, args []string) error {
		watcher := NewAgentWatcher()
		watcher.Start()
		return nil
	},
}
