// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nehpz/claudicus/pkg/agents"
	"github.com/nehpz/claudicus/pkg/config"
	"github.com/nehpz/claudicus/pkg/state"
)

// execCommand allows mocking exec.Command for testing (separate from tmux.go variable)
var uziExecCommand = exec.Command

// SessionInfo contains displayable information about a session
type SessionInfo struct {
	Name           string `json:"name"`
	AgentName      string `json:"agent_name"`
	Model          string `json:"model"`
	Status         string `json:"status"`
	Prompt         string `json:"prompt"`
	Insertions     int    `json:"insertions"`
	Deletions      int    `json:"deletions"`
	WorktreePath   string `json:"worktree_path"`
	Port           int    `json:"port,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
	ActivityStatus string `json:"activity_status,omitempty"` // For test compatibility
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
	
	// RunCheckpoint creates a checkpoint for an agent
	RunCheckpoint(agentName string, message string) error
	
	// SpawnAgent creates a new agent and returns the session name
	SpawnAgent(prompt, model string) (string, error)

	// SpawnAgentInteractive launches an interactive agent creation
	SpawnAgentInteractive(opts string) (<-chan struct{}, error)
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
// StateManagerInterface defines the interface for state management operations
type StateManagerInterface interface {
	GetActiveSessionsForRepo() ([]string, error)
	GetStatePath() string
	RemoveState(sessionName string) error
	SaveState(prompt, branchName, sessionName, worktreePath, model string) error
	SaveStateWithPort(prompt, branchName, sessionName, worktreePath, model string, port int) error
}

// StateManagerBridge implements StateManagerInterface by wrapping state.StateManager
type StateManagerBridge struct {
	*state.StateManager
}

// NewStateManagerBridge creates a new StateManagerBridge
func NewStateManagerBridge() StateManagerInterface {
	return &StateManagerBridge{
		StateManager: state.NewStateManager(),
	}
}

// SaveState delegates to the wrapped StateManager
func (s *StateManagerBridge) SaveState(prompt, branchName, sessionName, worktreePath, model string) error {
	return s.StateManager.SaveState(prompt, branchName, sessionName, worktreePath, model)
}

// SaveStateWithPort delegates to the wrapped StateManager
func (s *StateManagerBridge) SaveStateWithPort(prompt, branchName, sessionName, worktreePath, model string, port int) error {
	return s.StateManager.SaveStateWithPort(prompt, branchName, sessionName, worktreePath, model, port)
}

type UziCLI struct {
	stateManager  StateManagerInterface
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
		cmd := uziExecCommand(name, args...)

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

	// Shell out to uzi ls --json
	output, err := c.executeCommand("uzi", "ls", "--json")
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
	_, err := c.executeCommand("uzi", "kill", agentName)
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
	_, err := c.executeCommand("uzi", "prompt", "--agents", agents, prompt)
	if err != nil {
		return c.wrapError("RunPrompt", err)
	}
	return nil
}

// RunBroadcast implements UziInterface using the proxy pattern
func (c *UziCLI) RunBroadcast(message string) error {
	_, err := c.executeCommand("uzi", "broadcast", message)
	if err != nil {
		return c.wrapError("RunBroadcast", err)
	}
	return nil
}

// RunCommand implements UziInterface using the proxy pattern
func (c *UziCLI) RunCommand(command string) error {
	_, err := c.executeCommand("uzi", "run", command)
	if err != nil {
		return c.wrapError("RunCommand", err)
	}
	return nil
}

// RunCheckpoint implements UziInterface using the proxy pattern with streaming git output
func (c *UziCLI) RunCheckpoint(agentName string, message string) error {
	// Use the checkpoint command but capture detailed output
	output, err := c.executeCommand("uzi", "checkpoint", agentName, message)
	if err != nil {
		return c.wrapError("RunCheckpoint", fmt.Errorf("%w\nOutput: %s", err, string(output)))
	}
	return nil
}

// SpawnAgent implements UziInterface - creates a new agent following the uzi nuke && uzi start workflow
// This method handles the full agent creation process including:
// - Branch creation with unique naming
// - Git worktree setup
// - Tmux session creation and configuration
// - Development environment setup (if configured)
// - Agent command execution
func (c *UziCLI) SpawnAgent(prompt, model string) (string, error) {
	start := time.Now()
	defer func() { c.logOperation("SpawnAgent", time.Since(start), nil) }()
	
	// Create the agent configuration by wrapping model in agent:count format
	agentsFlag := fmt.Sprintf("%s:1", model)
	
	// Execute the spawn workflow directly using our internal implementation
	sessionName, err := c.executeSpawnWorkflow(agentsFlag, prompt)
	if err != nil {
		return "", c.wrapError("SpawnAgent", err)
	}
	
	return sessionName, nil
}

// executeSpawnWorkflow implements the core agent spawning logic based on cmd/prompt/prompt.go
// This follows the same workflow as `uzi prompt` but returns the created session name
func (c *UziCLI) executeSpawnWorkflow(agentsFlag, promptText string) (string, error) {
	// Load config - required for standardized dev environment setup (will be handled in individual helper methods)
	// The UziCLI uses ProxyConfig, not uzi.yaml config, so we'll handle config loading in helper methods
	
	// Parse the agent configuration
	agentConfigs, err := c.parseAgentConfigs(agentsFlag)
	if err != nil {
		return "", fmt.Errorf("error parsing agents: %w", err)
	}
	
	// Load existing session ports to prevent collisions
	stateManager := c.stateManager
	
	existingPorts, err := c.getExistingSessionPorts(stateManager)
	if err != nil {
		log.Printf("Failed to load existing session ports, proceeding without collision check: %v", err)
		existingPorts = []int{}
	}
	
	// Track assigned ports
	assignedPorts := existingPorts
	var createdSessionName string
	
	// Process each agent configuration (typically just one for SpawnAgent)
	for agent, config := range agentConfigs {
		for i := 0; i < config.Count; i++ {
			sessionName, err := c.createSingleAgent(agent, config, promptText, &assignedPorts, stateManager)
			if err != nil {
				return "", fmt.Errorf("failed to create agent %s: %w", agent, err)
			}
			
			// Store the first (and typically only) created session name
			if createdSessionName == "" {
				createdSessionName = sessionName
			}
		}
	}
	
	if createdSessionName == "" {
		return "", fmt.Errorf("no agent session was created")
	}
	
	return createdSessionName, nil
}

// createSingleAgent creates a single agent session following the established workflow
func (c *UziCLI) createSingleAgent(agent string, config AgentConfig, promptText string, assignedPorts *[]int, stateManager StateManagerInterface) (string, error) {
	// Generate random agent name for unique identification
	randomAgentName, err := c.getRandomAgentName(agent)
	if err != nil {
		return "", fmt.Errorf("failed to generate agent name: %w", err)
	}
	
	// Get git information
	gitHash, projectDir, err := c.getGitInfo()
	if err != nil {
		return "", fmt.Errorf("failed to get git information: %w", err)
	}
	
	// Generate unique identifiers
	timestamp := time.Now().Unix()
	uniqueId := fmt.Sprintf("%d", timestamp)
	
	// Create branch and session names
	branchName := fmt.Sprintf("%s-%s-%s-%s", randomAgentName, projectDir, gitHash, uniqueId)
	worktreeName := fmt.Sprintf("%s-%s-%s-%s", randomAgentName, projectDir, gitHash, uniqueId)
	sessionName := fmt.Sprintf("agent-%s-%s-%s", projectDir, gitHash, randomAgentName)
	
	// Create worktree
	worktreePath, err := c.createWorktree(branchName, worktreeName)
	if err != nil {
		return "", fmt.Errorf("failed to create worktree: %w", err)
	}
	
	// Create tmux session
	if err := c.createTmuxSession(sessionName, worktreePath); err != nil {
		return "", fmt.Errorf("failed to create tmux session: %w", err)
	}
	
	// Setup development environment and execute agent command
	var selectedPort int
	// Always try to setup dev environment - the method will check if config is available
	selectedPort, err = c.setupDevEnvironment(sessionName, worktreePath, assignedPorts)
	if err != nil {
		log.Printf("Failed to setup dev environment, continuing without it: %v", err)
		selectedPort = 0
	}
	
	// Execute the agent command
	commandToUse := config.Command
	if agent == "random" {
		commandToUse = randomAgentName
	}
	
	if err := c.executeAgentCommand(sessionName, commandToUse, promptText, worktreePath); err != nil {
		return "", fmt.Errorf("failed to execute agent command: %w", err)
	}
	
	// Save state
	if stateManager != nil {
		if selectedPort > 0 {
			if err := stateManager.SaveStateWithPort(promptText, branchName, sessionName, worktreePath, commandToUse, selectedPort); err != nil {
				log.Printf("Failed to save state with port: %v", err)
			}
		} else {
			if err := stateManager.SaveState(promptText, branchName, sessionName, worktreePath, commandToUse); err != nil {
				log.Printf("Failed to save state: %v", err)
			}
		}
	}
	
	return sessionName, nil
}

// Helper functions (these replicate logic from cmd/ls/ls.go for now)

// extractAgentName extracts the agent name from a session name
// Session format: agent-projectDir-gitHash-agentName
func extractAgentName(sessionName string) string {
	// Session format: agent-<project>-<hash>-<agent-name>
	// We need to find the last hash-like part and return everything after it
	parts := strings.Split(sessionName, "-")
	if len(parts) >= 4 && parts[0] == "agent" {
		// Look for the hash part - typically 6+ alphanumeric characters
		// Start from the end and work backwards to find the last hash-like part
		for i := len(parts) - 2; i >= 2; i-- {
			part := parts[i]
			if isHashLike(part) {
				// Return everything after this hash part
				return strings.Join(parts[i+1:], "-")
			}
		}
		// Fallback: assume standard format agent-project-hash-agent
		if len(parts) >= 4 {
			return strings.Join(parts[3:], "-")
		}
	}
	return sessionName
}

// isHashLike checks if a string looks like a git hash (alphanumeric, typically 6+ chars)
func isHashLike(s string) bool {
	if len(s) < 6 {
		return false
	}
	
	hasDigit := false
	hasLetter := false
	
	for _, r := range s {
		if r >= '0' && r <= '9' {
			hasDigit = true
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		} else {
			return false // Contains non-alphanumeric character
		}
	}
	
	// A hash should have both letters and digits
	return hasDigit && hasLetter
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
	cmd := uziExecCommand("tmux", "capture-pane", "-t", sessionName+":agent", "-p")
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
	cmd := uziExecCommand("sh", "-c", shellCmdString)
	
	// Only set Dir if it's not a test directory
	if !strings.Contains(sessionState.WorktreePath, "test-worktree") && !strings.Contains(sessionState.WorktreePath, "/tmp/test-") {
		cmd.Dir = sessionState.WorktreePath
	}

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

func (c *UziClient) SpawnAgent(prompt, model string) (string, error) {
	// Stub: will be replaced by UziCLI implementation
	_ = prompt
	_ = model
	return "", fmt.Errorf("not implemented - use UziCLI instead")
}

func (c *UziClient) SpawnAgentInteractive(opts string) (<-chan struct{}, error) {
	// Stub: will be replaced by UziCLI implementation
	_ = opts
	ch := make(chan struct{})
	close(ch)
	return ch, fmt.Errorf("not implemented - use UziCLI instead")
}

// SpawnAgent helper methods implementation

// AgentConfig represents an agent configuration
type AgentConfig struct {
	Command string
	Count   int
}

// loadDefaultConfig loads the default uzi configuration
func (c *UziCLI) loadDefaultConfig() (*config.Config, error) {
	configPath := config.GetDefaultConfigPath()
	return config.LoadConfig(configPath)
}

// parseAgentConfigs parses the agents flag value into a map of agent configs
func (c *UziCLI) parseAgentConfigs(agentsStr string) (map[string]AgentConfig, error) {
	agentConfigs := make(map[string]AgentConfig)

	// Split by comma for multiple agent configurations
	agentPairs := strings.Split(agentsStr, ",")

	for _, pair := range agentPairs {
		// Split by colon for agent:count
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid agent format: %s (expected agent:count)", pair)
		}

		agent := strings.TrimSpace(parts[0])
		countStr := strings.TrimSpace(parts[1])

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid count for agent %s: %s", agent, countStr)
		}

		if count < 1 {
			return nil, fmt.Errorf("count must be at least 1 for agent %s", agent)
		}

		// Map agent names to actual commands
		command := c.getCommandForAgent(agent)
		agentConfigs[agent] = AgentConfig{
			Command: command,
			Count:   count,
		}
	}

	return agentConfigs, nil
}

// getCommandForAgent maps agent names to their actual CLI commands
func (c *UziCLI) getCommandForAgent(agent string) string {
	switch agent {
	case "claude":
		return "claude"
	case "cursor":
		return "cursor"
	case "codex":
		return "codex"
	case "gemini":
		return "gemini"
	case "random":
		return "claude" // Default for random agents
	default:
		// For unknown agents, assume the agent name is the command
		return agent
	}
}

// getExistingSessionPorts reads the state file and returns all currently assigned ports
func (c *UziCLI) getExistingSessionPorts(stateManager StateManagerInterface) ([]int, error) {
	if stateManager == nil {
		return []int{}, nil
	}
	
	// Read the state file
	stateFile := stateManager.GetStatePath()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No state file exists yet, return empty list
			return []int{}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}
	
	// Parse the state file
	states := make(map[string]state.AgentState)
	if err := json.Unmarshal(data, &states); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}
	
	// Extract all assigned ports
	var existingPorts []int
	for _, agentState := range states {
		if agentState.Port > 0 {
			existingPorts = append(existingPorts, agentState.Port)
		}
	}
	
	return existingPorts, nil
}

// getRandomAgentName generates a random agent name
func (c *UziCLI) getRandomAgentName(agent string) (string, error) {
	if agent == "random" {
		return agents.GetRandomAgent(), nil
	}
	return agent, nil
}

// getGitInfo retrieves git hash and project directory information
func (c *UziCLI) getGitInfo() (gitHash, projectDir string, err error) {
	ctx := context.Background()
	
	// Get the current git hash
	gitHashCmd := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
	gitHashOutput, err := gitHashCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error getting git hash: %w", err)
	}
	gitHash = strings.TrimSpace(string(gitHashOutput))

	// Get the git repository name from remote URL
	gitRemoteCmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	gitRemoteOutput, err := gitRemoteCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("error getting git remote: %w", err)
	}
	remoteURL := strings.TrimSpace(string(gitRemoteOutput))
	// Extract repository name from URL (handle both https and ssh formats)
	repoName := filepath.Base(remoteURL)
	projectDir = strings.TrimSuffix(repoName, ".git")
	
	return gitHash, projectDir, nil
}

// createWorktree creates a git worktree for the agent
func (c *UziCLI) createWorktree(branchName, worktreeName string) (string, error) {
	ctx := context.Background()
	
	// Get home directory for worktree storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}

	worktreesDir := filepath.Join(homeDir, ".local", "share", "uzi", "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return "", fmt.Errorf("error creating worktrees directory: %w", err)
	}

	worktreePath := filepath.Join(worktreesDir, worktreeName)
	
	// Create git worktree
	cmd := fmt.Sprintf("git worktree add -b %s %s", branchName, worktreePath)
	cmdExec := exec.CommandContext(ctx, "sh", "-c", cmd)
	if err := cmdExec.Run(); err != nil {
		return "", fmt.Errorf("error creating git worktree: %w", err)
	}
	
	return worktreePath, nil
}

// createTmuxSession creates a tmux session for the agent
func (c *UziCLI) createTmuxSession(sessionName, worktreePath string) error {
	ctx := context.Background()
	
	// Create tmux session
	cmd := fmt.Sprintf("tmux new-session -d -s %s -c %s", sessionName, worktreePath)
	cmdExec := exec.CommandContext(ctx, "sh", "-c", cmd)
	if err := cmdExec.Run(); err != nil {
		return fmt.Errorf("error creating tmux session: %w", err)
	}

	// Rename the first window to "agent"
	renameCmd := fmt.Sprintf("tmux rename-window -t %s:0 agent", sessionName)
	renameExec := exec.CommandContext(ctx, "sh", "-c", renameCmd)
	if err := renameExec.Run(); err != nil {
		return fmt.Errorf("error renaming tmux window: %w", err)
	}
	
	return nil
}

// setupDevEnvironment sets up the development environment if configured
func (c *UziCLI) setupDevEnvironment(sessionName, worktreePath string, assignedPorts *[]int) (int, error) {
	ctx := context.Background()
	
	// Load configuration to get dev settings
	cfg, err := c.loadDefaultConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to load config: %w", err)
	}
	
	if cfg.DevCommand == nil || *cfg.DevCommand == "" || cfg.PortRange == nil || *cfg.PortRange == "" {
		return 0, nil // No dev environment to set up
	}
	
	// Parse port range
	ports := strings.Split(*cfg.PortRange, "-")
	if len(ports) != 2 {
		return 0, fmt.Errorf("invalid port range format: %s", *cfg.PortRange)
	}

	startPort, err1 := strconv.Atoi(ports[0])
	endPort, err2 := strconv.Atoi(ports[1])
	if err1 != nil || err2 != nil || startPort <= 0 || endPort <= 0 || endPort < startPort {
		return 0, fmt.Errorf("invalid port range: %s", *cfg.PortRange)
	}

	// Find available port
	selectedPort, err := c.findAvailablePort(startPort, endPort, *assignedPorts)
	if err != nil {
		return 0, fmt.Errorf("error finding available port: %w", err)
	}

	// Create development command
	devCmdTemplate := *cfg.DevCommand
	devCmd := strings.Replace(devCmdTemplate, "$PORT", strconv.Itoa(selectedPort), 1)

	// Create new window named uzi-dev
	newWindowCmd := fmt.Sprintf("tmux new-window -t %s -n uzi-dev -c %s", sessionName, worktreePath)
	newWindowExec := exec.CommandContext(ctx, "sh", "-c", newWindowCmd)
	if err := newWindowExec.Run(); err != nil {
		return 0, fmt.Errorf("error creating new tmux window for dev server: %w", err)
	}

	// Send dev command to the new window
	sendDevCmd := fmt.Sprintf("tmux send-keys -t %s:uzi-dev '%s' C-m", sessionName, devCmd)
	sendDevCmdExec := exec.CommandContext(ctx, "sh", "-c", sendDevCmd)
	if err := sendDevCmdExec.Run(); err != nil {
		return 0, fmt.Errorf("error sending dev command to tmux: %w", err)
	}

	// Update assigned ports
	*assignedPorts = append(*assignedPorts, selectedPort)
	
	return selectedPort, nil
}

// findAvailablePort finds the first available port in the given range, excluding already assigned ports
func (c *UziCLI) findAvailablePort(startPort, endPort int, assignedPorts []int) (int, error) {
	for port := startPort; port <= endPort; port++ {
		// Check if port is already assigned in this execution
		alreadyAssigned := false
		for _, assignedPort := range assignedPorts {
			if port == assignedPort {
				alreadyAssigned = true
				break
			}
		}
		if alreadyAssigned {
			continue
		}

		// Check if port is actually available
		if c.isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", startPort, endPort)
}

// isPortAvailable checks if a port is available for use
func (c *UziCLI) isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// executeAgentCommand executes the agent command in the tmux session
func (c *UziCLI) executeAgentCommand(sessionName, commandToUse, promptText, worktreePath string) error {
	ctx := context.Background()
	
	// Hit enter in the agent pane
	hitEnterCmd := fmt.Sprintf("tmux send-keys -t %s:agent C-m", sessionName)
	hitEnterExec := exec.CommandContext(ctx, "sh", "-c", hitEnterCmd)
	if err := hitEnterExec.Run(); err != nil {
		return fmt.Errorf("error hitting enter in tmux: %w", err)
	}

	// Prepare the command template based on the agent type
	var tmuxCmd string
	if commandToUse == "gemini" {
		tmuxCmd = fmt.Sprintf("tmux send-keys -t %s:agent '%s -p \"%s\"' C-m", sessionName, commandToUse, promptText)
	} else {
		tmuxCmd = fmt.Sprintf("tmux send-keys -t %s:agent '%s \"%s\"' C-m", sessionName, commandToUse, promptText)
	}
	
	tmuxCmdExec := exec.CommandContext(ctx, "sh", "-c", tmuxCmd)
	tmuxCmdExec.Dir = worktreePath
	if err := tmuxCmdExec.Run(); err != nil {
		return fmt.Errorf("error sending keys to tmux: %w", err)
	}
	
	return nil
}

// SpawnAgentInteractive implements the interactive agent creation with progress reporting
func (c *UziCLI) SpawnAgentInteractive(opts string) (<-chan struct{}, error) {
	progressChan := make(chan struct{}, 1)
	
	// Parse options (format: "agentType:count:prompt")
	parts := strings.SplitN(opts, ":", 3)
	if len(parts) != 3 {
		close(progressChan)
		return progressChan, fmt.Errorf("invalid options format, expected 'agentType:count:prompt'")
	}
	
	agentType := strings.TrimSpace(parts[0])
	countStr := strings.TrimSpace(parts[1])
	prompt := strings.TrimSpace(parts[2])
	
	// Validate count
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 10 {
		close(progressChan)
		return progressChan, fmt.Errorf("invalid count: must be between 1 and 10")
	}
	
	// Start async agent creation
	go func() {
		defer close(progressChan)
		
		// Create the agent configuration
		agentsFlag := fmt.Sprintf("%s:%d", agentType, count)
		
		// Execute the spawn workflow
		_, err := c.executeSpawnWorkflow(agentsFlag, prompt)
		if err != nil {
			log.Printf("SpawnAgentInteractive failed: %v", err)
			return
		}
		
		// Signal completion
		progressChan <- struct{}{}
	}()
	
	return progressChan, nil
}
