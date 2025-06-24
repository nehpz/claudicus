// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/devflowinc/uzi/pkg/state"
)

// SessionInfo contains displayable information about a session
type SessionInfo struct {
	Name        string `json:"name"`
	AgentName   string `json:"agent_name"`
	Model       string `json:"model"`
	Status      string `json:"status"`
	Prompt      string `json:"prompt"`
	Insertions  int    `json:"insertions"`
	Deletions   int    `json:"deletions"`
	WorktreePath string `json:"worktree_path"`
	Port        int    `json:"port,omitempty"`
}

// UziInterface defines the interface for interacting with Uzi core functionality
type UziInterface interface {
	// GetSessions returns a list of session information
	GetSessions() ([]SessionInfo, error)
	
	// GetSessionState returns the state for a specific session
	GetSessionState(sessionName string) (*state.AgentState, error)
	
	// GetSessionStatus returns the current status of a session
	GetSessionStatus(sessionName string) (string, error)
	
	// AttachToSession attaches to an existing session
	AttachToSession(sessionName string) error
	
	// KillSession terminates a session
	KillSession(sessionName string) error
	
	// RefreshSessions refreshes the session list
	RefreshSessions() error
	
	// RunPrompt creates a new agent session
	RunPrompt(agents string, prompt string) error
	
	// RunBroadcast sends a message to all active sessions
	RunBroadcast(message string) error
	
	// RunCommand executes a command in all sessions
	RunCommand(command string) error
}

// ProxyConfig defines configuration for the UziCLI proxy
type ProxyConfig struct {
	Timeout     time.Duration
	Retries     int
	LogLevel    string
	EnableCache bool
}

// DefaultProxyConfig returns sensible defaults for the proxy
func DefaultProxyConfig() ProxyConfig {
	return ProxyConfig{
		Timeout:     30 * time.Second,
		Retries:     2,
		LogLevel:    "info",
		EnableCache: false,
	}
}

// UziCLI implements UziInterface by providing a consistent proxy layer
// All operations go through this proxy for unified error handling, logging, and debugging
type UziCLI struct {
	stateManager  *state.StateManager
	tmuxDiscovery *TmuxDiscovery
	config        ProxyConfig
}

// NewUziCLI creates a new UziCLI implementation with default configuration
func NewUziCLI() *UziCLI {
	return NewUziCLIWithConfig(DefaultProxyConfig())
}

// NewUziCLIWithConfig creates a new UziCLI implementation with custom configuration
func NewUziCLIWithConfig(config ProxyConfig) *UziCLI {
	return &UziCLI{
		stateManager:  state.NewStateManager(),
		tmuxDiscovery: NewTmuxDiscovery(),
		config:        config,
	}
}

// Core proxy infrastructure methods

// executeCommand runs a command with consistent error handling and logging
func (c *UziCLI) executeCommand(name string, args ...string) ([]byte, error) {
	return c.executeCommandWithTimeout(c.config.Timeout, name, args...)
}

// executeCommandWithTimeout runs a command with a custom timeout
func (c *UziCLI) executeCommandWithTimeout(timeout time.Duration, name string, args ...string) ([]byte, error) {
	start := time.Now()
	var lastErr error

	for attempt := 0; attempt <= c.config.Retries; attempt++ {
		cmd := exec.Command(name, args...)

		// Set up stdout and stderr capture
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Create a channel to handle timeout
		done := make(chan error, 1)
		go func() {
			done <- cmd.Run()
		}()

		select {
		case err := <-done:
			duration := time.Since(start)
			if err != nil {
				lastErr = fmt.Errorf("command failed (attempt %d/%d): %w - stderr: %s", 
					attempt+1, c.config.Retries+1, err, stderr.String())
				c.logOperation(fmt.Sprintf("%s %v", name, args), duration, lastErr)

				// Don't retry if it's the last attempt
				if attempt == c.config.Retries {
					return nil, c.wrapError(fmt.Sprintf("%s %v", name, args), lastErr)
				}

				// Brief delay before retry
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// Success
			c.logOperation(fmt.Sprintf("%s %v", name, args), duration, nil)
			return stdout.Bytes(), nil

		case <-time.After(timeout):
			cmd.Process.Kill()
			lastErr = fmt.Errorf("command timed out after %v", timeout)
			c.logOperation(fmt.Sprintf("%s %v", name, args), timeout, lastErr)

			if attempt == c.config.Retries {
				return nil, c.wrapError(fmt.Sprintf("%s %v", name, args), lastErr)
			}
		}
	}

	return nil, c.wrapError(fmt.Sprintf("%s %v", name, args), lastErr)
}

// wrapError provides consistent error wrapping with proxy context
func (c *UziCLI) wrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("uzi_proxy: %s: %w", operation, err)
}

// logOperation logs command execution details based on configuration
func (c *UziCLI) logOperation(operation string, duration time.Duration, err error) {
	if c.config.LogLevel == "debug" || (c.config.LogLevel == "info" && err != nil) {
		if err != nil {
			log.Printf("[UziCLI] %s failed in %v: %v", operation, duration, err)
		} else {
			log.Printf("[UziCLI] %s completed in %v", operation, duration)
		}
	}
}

// GetSessions implements UziInterface using the proxy pattern
// This shells out to `uzi ls --json` for consistent behavior
func (c *UziCLI) GetSessions() ([]SessionInfo, error) {
	start := time.Now()
	defer func() { c.logOperation("GetSessions", time.Since(start), nil) }()

	// Shell out to uzi ls --json (use ./uzi for testing to get current binary)
	output, err := c.executeCommand("./uzi", "ls", "--json")
	if err != nil {
		return nil, c.wrapError("GetSessions", err)
	}

	// Parse JSON response
	var sessions []SessionInfo
	if err := json.Unmarshal(output, &sessions); err != nil {
		return nil, c.wrapError("GetSessions", fmt.Errorf("failed to parse JSON: %w", err))
	}

	return sessions, nil
}

// GetSessionsLegacy implements the legacy behavior by reading state.json directly
// This method is kept for fallback and testing purposes
func (c *UziCLI) GetSessionsLegacy() ([]SessionInfo, error) {
	if c.stateManager == nil {
		return nil, fmt.Errorf("state manager not initialized")
	}

	// Get active sessions
	activeSessions, err := c.stateManager.GetActiveSessionsForRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	// Load state file to get detailed information
	states := make(map[string]state.AgentState)
	if data, err := os.ReadFile(c.stateManager.GetStatePath()); err != nil {
		if os.IsNotExist(err) {
			return []SessionInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	} else {
		if err := json.Unmarshal(data, &states); err != nil {
			return nil, fmt.Errorf("failed to parse state file: %w", err)
		}
	}

	// Build session info list
	var sessions []SessionInfo
	for _, sessionName := range activeSessions {
		state, ok := states[sessionName]
		if !ok {
			continue
		}

		// Extract agent name from session name
		agentName := extractAgentName(sessionName)
		
		// Get session status
		status := c.getAgentStatus(sessionName)
		
		// Get git diff stats
		insertions, deletions := c.getGitDiffTotals(sessionName, &state)

		sessionInfo := SessionInfo{
			Name:         sessionName,
			AgentName:    agentName,
			Model:        state.Model,
			Status:       status,
			Prompt:       state.Prompt,
			Insertions:   insertions,
			Deletions:    deletions,
			WorktreePath: state.WorktreePath,
			Port:         state.Port,
		}
		sessions = append(sessions, sessionInfo)
	}

	// Sort sessions by port for stable ordering
	// Sessions with port 0 (no dev server) will be sorted first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Port < sessions[j].Port
	})

	return sessions, nil
}

// GetSessionState implements UziInterface
func (c *UziCLI) GetSessionState(sessionName string) (*state.AgentState, error) {
	if c.stateManager == nil {
		return nil, fmt.Errorf("state manager not initialized")
	}

	// Load state file
	states := make(map[string]state.AgentState)
	data, err := os.ReadFile(c.stateManager.GetStatePath())
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, &states); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	state, exists := states[sessionName]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionName)
	}

	return &state, nil
}

// GetSessionStatus implements UziInterface
func (c *UziCLI) GetSessionStatus(sessionName string) (string, error) {
	return c.getAgentStatus(sessionName), nil
}

// AttachToSession implements UziInterface by executing tmux attach
// Note: This is one case where we don't use executeCommand since it needs direct terminal access
func (c *UziCLI) AttachToSession(sessionName string) error {
	start := time.Now()
	defer func() { c.logOperation("AttachToSession", time.Since(start), nil) }()

	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return c.wrapError("AttachToSession", err)
	}
	return nil
}

// KillSession implements UziInterface using the proxy pattern
func (c *UziCLI) KillSession(sessionName string) error {
	// Extract agent name from session name
	agentName := extractAgentName(sessionName)
	_, err := c.executeCommand("./uzi", "kill", agentName)
	if err != nil {
		return c.wrapError("KillSession", err)
	}
	return nil
}

// RefreshSessions implements UziInterface (no-op as data is read fresh each time)
func (c *UziCLI) RefreshSessions() error {
	// No caching in this implementation, so nothing to refresh
	c.logOperation("RefreshSessions", 0, nil)
	return nil
}

// RunPrompt implements UziInterface using the proxy pattern
func (c *UziCLI) RunPrompt(agents string, prompt string) error {
	_, err := c.executeCommand("./uzi", "prompt", "--agents", agents, prompt)
	if err != nil {
		return c.wrapError("RunPrompt", err)
	}
	return nil
}

// RunBroadcast implements UziInterface using the proxy pattern
func (c *UziCLI) RunBroadcast(message string) error {
	_, err := c.executeCommand("./uzi", "broadcast", message)
	if err != nil {
		return c.wrapError("RunBroadcast", err)
	}
	return nil
}

// RunCommand implements UziInterface using the proxy pattern
func (c *UziCLI) RunCommand(command string) error {
	_, err := c.executeCommand("./uzi", "run", command)
	if err != nil {
		return c.wrapError("RunCommand", err)
	}
	return nil
}

// Helper functions (these replicate logic from cmd/ls/ls.go for now)

// extractAgentName extracts the agent name from a session name
// Session format: agent-projectDir-gitHash-agentName
func extractAgentName(sessionName string) string {
	parts := strings.Split(sessionName, "-")
	if len(parts) >= 4 && parts[0] == "agent" {
		return strings.Join(parts[3:], "-")
	}
	return sessionName
}

// getAgentStatus determines the current status of an agent session
func (c *UziCLI) getAgentStatus(sessionName string) string {
	content, err := c.getPaneContent(sessionName)
	if err != nil {
		return "unknown"
	}

	if strings.Contains(content, "esc to interrupt") || strings.Contains(content, "Thinking") {
		return "running"
	}
	return "ready"
}

// getPaneContent gets the content of a tmux pane
func (c *UziCLI) getPaneContent(sessionName string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName+":agent", "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// getGitDiffTotals gets the insertion/deletion counts for a session
func (c *UziCLI) getGitDiffTotals(sessionName string, sessionState *state.AgentState) (int, int) {
	if sessionState.WorktreePath == "" {
		return 0, 0
	}

	shellCmdString := "git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null"
	cmd := exec.Command("sh", "-c", shellCmdString)
	cmd.Dir = sessionState.WorktreePath

	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	insertions := 0
	deletions := 0

	insRe := regexp.MustCompile(`(\d+) insertion(?:s)?\(\+\)`)
	delRe := regexp.MustCompile(`(\d+) deletion(?:s)?\(\-\)`)

	if m := insRe.FindStringSubmatch(string(output)); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &insertions)
	}
	if m := delRe.FindStringSubmatch(string(output)); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &deletions)
	}

	return insertions, deletions
}

// Enhanced methods using tmux discovery

// GetSessionsWithTmuxInfo returns sessions enhanced with tmux attachment information
func (c *UziCLI) GetSessionsWithTmuxInfo() ([]SessionInfo, map[string]TmuxSessionInfo, error) {
	sessions, err := c.GetSessions()
	if err != nil {
		return nil, nil, err
	}

	// Get tmux session mapping
	tmuxMapping, err := c.tmuxDiscovery.MapUziSessionsToTmux(sessions)
	if err != nil {
		return sessions, nil, err // Return sessions even if tmux mapping fails
	}

	// Update session status with tmux information where available
	for i := range sessions {
		if tmuxInfo, exists := tmuxMapping[sessions[i].Name]; exists {
			// Enhance status with tmux information
			if tmuxInfo.Attached {
				sessions[i].Status = "attached"
			} else if sessions[i].Status == "ready" {
				// Only update if current status is generic "ready"
				if tmuxStatus, err := c.tmuxDiscovery.GetSessionStatus(sessions[i].Name); err == nil {
					sessions[i].Status = tmuxStatus
				}
			}
		}
	}

	return sessions, tmuxMapping, nil
}

// IsSessionAttached returns true if a session is currently attached in tmux
func (c *UziCLI) IsSessionAttached(sessionName string) bool {
	return c.tmuxDiscovery.IsSessionAttached(sessionName)
}

// GetSessionActivity returns the activity level of a session
func (c *UziCLI) GetSessionActivity(sessionName string) string {
	return c.tmuxDiscovery.GetSessionActivity(sessionName)
}

// GetAttachedSessionCount returns the number of currently attached sessions
func (c *UziCLI) GetAttachedSessionCount() (int, error) {
	return c.tmuxDiscovery.GetAttachedSessionCount()
}

// RefreshTmuxCache forces a refresh of the tmux session cache
func (c *UziCLI) RefreshTmuxCache() {
	c.tmuxDiscovery.RefreshCache()
}

// GetTmuxSessionsByActivity returns Uzi sessions grouped by their tmux activity level
func (c *UziCLI) GetTmuxSessionsByActivity() (map[string][]TmuxSessionInfo, error) {
	return c.tmuxDiscovery.ListSessionsByActivity()
}

// FormatSessionActivity returns a styled activity indicator for the TUI
func (c *UziCLI) FormatSessionActivity(sessionName string) string {
	activity := c.tmuxDiscovery.GetSessionActivity(sessionName)
	return c.tmuxDiscovery.FormatSessionActivity(activity)
}

// Legacy UziClient for backward compatibility
// TODO: Remove this once TUI is fully migrated to use UziCLI
type UziClient struct {
	stateManager *state.StateManager
}

// NewUziClient creates a new Uzi client for TUI operations (legacy)
func NewUziClient() *UziClient {
	return &UziClient{
		stateManager: state.NewStateManager(),
	}
}

// GetActiveSessions implements legacy interface
func (c *UziClient) GetActiveSessions() ([]string, error) {
	if c.stateManager == nil {
		return nil, nil
	}
	return c.stateManager.GetActiveSessionsForRepo()
}

// Stub implementations for compilation compatibility

func (c *UziClient) GetSessionState(sessionName string) (*state.AgentState, error) {
	// Stub: will be replaced by UziCLI implementation
	_ = sessionName
	return nil, fmt.Errorf("not implemented - use UziCLI instead")
}

func (c *UziClient) GetSessionStatus(sessionName string) (string, error) {
	// Stub: will be replaced by UziCLI implementation
	_ = sessionName
	return "unknown", nil
}

func (c *UziClient) AttachToSession(sessionName string) error {
	// Stub: will be replaced by UziCLI implementation
	_ = sessionName
	return fmt.Errorf("not implemented - use UziCLI instead")
}

func (c *UziClient) KillSession(sessionName string) error {
	// Stub: will be replaced by UziCLI implementation
	_ = sessionName
	return fmt.Errorf("not implemented - use UziCLI instead")
}

func (c *UziClient) RefreshSessions() error {
	// Stub: will be replaced by UziCLI implementation
	return nil
}
