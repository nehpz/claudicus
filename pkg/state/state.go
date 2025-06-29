package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

type AgentState struct {
	GitRepo      string    `json:"git_repo"`
	BranchFrom   string    `json:"branch_from"`
	BranchName   string    `json:"branch_name"`
	Prompt       string    `json:"prompt"`
	WorktreePath string    `json:"worktree_path"`
	Port         int       `json:"port,omitempty"`
	Model        string    `json:"model"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StateManager struct {
	statePath string
	fs        FileSystem
	cmdExec   CommandExecutor
}

// NewStateManager creates a StateManager with default dependencies
// This maintains backward compatibility while enabling dependency injection
func NewStateManager() *StateManager {
	return NewStateManagerWithDeps(NewDefaultFileSystem(), &DefaultCommandExecutor{})
}

// NewStateManagerWithDeps creates a StateManager with injected dependencies
// This follows the Dependency Inversion Principle for testability
func NewStateManagerWithDeps(fs FileSystem, cmdExec CommandExecutor) *StateManager {
	homeDir, err := fs.UserHomeDir()
	if err != nil {
		log.Error("Error getting home directory", "error", err)
		return nil
	}

	statePath := filepath.Join(homeDir, ".local", "share", "uzi", "state.json")
	return &StateManager{
		statePath: statePath,
		fs:        fs,
		cmdExec:   cmdExec,
	}
}

func (sm *StateManager) ensureStateDir() error {
	dir := filepath.Dir(sm.statePath)
	return sm.fs.MkdirAll(dir, 0755)
}

// getGitRepo uses injected CommandExecutor for testability
func (sm *StateManager) getGitRepo() string {
	output, err := sm.cmdExec.ExecuteCommand("git", "config", "--get", "remote.origin.url")
	if err != nil {
		log.Debug("Could not get git remote URL", "error", err)
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getBranchFrom uses injected CommandExecutor for testability
func (sm *StateManager) getBranchFrom() string {
	// Get the main/master branch name
	output, err := sm.cmdExec.ExecuteCommand("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		// Fallback to main
		return "main"
	}

	ref := strings.TrimSpace(string(output))
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "main"
}

// isActiveInTmux uses injected CommandExecutor for testability
func (sm *StateManager) isActiveInTmux(sessionName string) bool {
	err := sm.cmdExec.RunCommand("tmux", "has-session", "-t", sessionName)
	return err == nil
}

func (sm *StateManager) GetActiveSessionsForRepo() ([]string, error) {
	// Load existing state using injected filesystem
	states := make(map[string]AgentState)
	data, err := sm.fs.ReadFile(sm.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &states); err != nil {
		return nil, err
	}

	currentRepo := sm.getGitRepo()
	if currentRepo == "" {
		return []string{}, nil
	}

	var activeSessions []string
	for sessionName, state := range states {
		if state.GitRepo == currentRepo && sm.isActiveInTmux(sessionName) {
			activeSessions = append(activeSessions, sessionName)
		}
	}

	return activeSessions, nil
}

func (sm *StateManager) SaveState(prompt, branchName, sessionName, worktreePath, model string) error {
	return sm.SaveStateWithPort(prompt, branchName, sessionName, worktreePath, model, 0)
}

func (sm *StateManager) SaveStateWithPort(prompt, branchName, sessionName, worktreePath, model string, port int) error {
	if err := sm.ensureStateDir(); err != nil {
		return err
	}

	// Load existing state using injected filesystem
	states := make(map[string]AgentState)
	if data, err := sm.fs.ReadFile(sm.statePath); err == nil {
		json.Unmarshal(data, &states)
	}

	// Create new state entry
	now := time.Now()
	agentState := AgentState{
		GitRepo:      sm.getGitRepo(),
		BranchFrom:   sm.getBranchFrom(),
		BranchName:   branchName,
		Prompt:       prompt,
		WorktreePath: worktreePath,
		Port:         port,
		Model:        model,
		UpdatedAt:    now,
	}

	// Set created time if this is a new entry
	if existing, exists := states[sessionName]; exists {
		agentState.CreatedAt = existing.CreatedAt
	} else {
		agentState.CreatedAt = now
	}

	states[sessionName] = agentState

	// Store the worktree branch in agent-specific file
	if err := sm.storeWorktreeBranch(sessionName); err != nil {
		log.Error("Error storing worktree branch", "error", err)
	}

	// Save to file using injected filesystem
	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}

	return sm.fs.WriteFile(sm.statePath, data, 0644)
}

// getCurrentBranch uses injected CommandExecutor for testability
func (sm *StateManager) getCurrentBranch() string {
	output, err := sm.cmdExec.ExecuteCommand("git", "branch", "--show-current")
	if err != nil {
		log.Debug("Could not get current branch", "error", err)
		return ""
	}
	return strings.TrimSpace(string(output))
}

// storeWorktreeBranch uses injected dependencies for testability
func (sm *StateManager) storeWorktreeBranch(sessionName string) error {
	homeDir, err := sm.fs.UserHomeDir()
	if err != nil {
		return err
	}

	agentDir := filepath.Join(homeDir, ".local", "share", "uzi", "worktree", sessionName)
	if err := sm.fs.MkdirAll(agentDir, 0755); err != nil {
		return err
	}

	branchFile := filepath.Join(agentDir, "tree")
	currentBranch := sm.getCurrentBranch()
	if currentBranch == "" {
		return nil
	}

	return sm.fs.WriteFile(branchFile, []byte(currentBranch), 0644)
}

func (sm *StateManager) GetStatePath() string {
	return sm.statePath
}

func (sm *StateManager) RemoveState(sessionName string) error {
	// Load existing state
	states := make(map[string]AgentState)
	if data, err := os.ReadFile(sm.statePath); err != nil {
		if os.IsNotExist(err) {
			return nil // No state file, nothing to remove
		}
		return err
	} else {
		if err := json.Unmarshal(data, &states); err != nil {
			return err
		}
	}

	// Remove the session from the state
	delete(states, sessionName)

	// Save updated state to file
	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sm.statePath, data, 0644)
}

// GetWorktreeInfo returns the worktree information for a given session
func (sm *StateManager) GetWorktreeInfo(sessionName string) (*AgentState, error) {
	// Load existing state
	states := make(map[string]AgentState)
	if data, err := os.ReadFile(sm.statePath); err != nil {
		return nil, fmt.Errorf("error reading state file: %w", err)
	} else {
		if err := json.Unmarshal(data, &states); err != nil {
			return nil, fmt.Errorf("error parsing state file: %w", err)
		}
	}

	state, ok := states[sessionName]
	if !ok {
		return nil, fmt.Errorf("no state found for session: %s", sessionName)
	}

	return &state, nil
}
