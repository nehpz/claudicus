package state

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// SessionInfo represents a session with all required display information
type SessionInfo struct {
	SessionName  string `json:"session_name"`
	AgentName    string `json:"agent_name"`
	Status       string `json:"status"`
	DevServerURL string `json:"dev_server_url,omitempty"`
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	Insertions   int    `json:"insertions"`
	Deletions    int    `json:"deletions"`
	WorktreePath string `json:"worktree_path"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// StateReader provides functionality to read and parse state information
type StateReader struct {
	repoRoot  string
	statePath string
}

// NewStateReader creates a new StateReader for the given repository root
func NewStateReader(repoRoot string) *StateReader {
	statePath := filepath.Join(repoRoot, ".uzi", "state.json")
	return &StateReader{
		repoRoot:  repoRoot,
		statePath: statePath,
	}
}

// LoadSessions loads the state.json file and returns a slice of SessionInfo structs
func (sr *StateReader) LoadSessions() ([]SessionInfo, error) {
	// Check if state file exists
	if _, err := os.Stat(sr.statePath); os.IsNotExist(err) {
		return []SessionInfo{}, nil
	}

	// Read the state file
	data, err := os.ReadFile(sr.statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file %s: %w", sr.statePath, err)
	}

	// Parse JSON into map of AgentState
	states := make(map[string]AgentState)
	if err := json.Unmarshal(data, &states); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Convert to SessionInfo slice
	var sessions []SessionInfo
	for sessionName, agentState := range states {
		// Extract agent name from session name
		agentName := sr.extractAgentName(sessionName)

		// Get session status
		status := sr.getSessionStatus(sessionName)

		// Get git diff stats
		insertions, deletions := sr.getGitDiffStats(agentState.WorktreePath)

		// Format dev server URL
		devServerURL := ""
		if agentState.Port != 0 {
			devServerURL = fmt.Sprintf("http://localhost:%d", agentState.Port)
		}

		sessionInfo := SessionInfo{
			SessionName:  sessionName,
			AgentName:    agentName,
			Status:       status,
			DevServerURL: devServerURL,
			Model:        agentState.Model,
			Prompt:       agentState.Prompt,
			Insertions:   insertions,
			Deletions:    deletions,
			WorktreePath: agentState.WorktreePath,
			CreatedAt:    agentState.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    agentState.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		sessions = append(sessions, sessionInfo)
	}

	return sessions, nil
}

// extractAgentName extracts the agent name from a session name
// Session format: agent-projectDir-gitHash-agentName
func (sr *StateReader) extractAgentName(sessionName string) string {
	parts := strings.Split(sessionName, "-")
	if len(parts) >= 4 && parts[0] == "agent" {
		// Join all parts after the first 3 (in case agent name contains hyphens)
		return strings.Join(parts[3:], "-")
	}
	return sessionName
}

// getSessionStatus determines the current status of a session by checking tmux
func (sr *StateReader) getSessionStatus(sessionName string) string {
	// First check if tmux session exists
	checkCmd := exec.Command("tmux", "has-session", "-t", sessionName)
	if err := checkCmd.Run(); err != nil {
		return "inactive"
	}

	// Get pane content to determine if agent is running
	content, err := sr.getPaneContent(sessionName)
	if err != nil {
		return "unknown"
	}

	// Check for running indicators
	if strings.Contains(content, "esc to interrupt") ||
		strings.Contains(content, "Thinking") ||
		strings.Contains(content, "Working") {
		return "running"
	}

	return "ready"
}

// getPaneContent gets the content of a tmux pane
func (sr *StateReader) getPaneContent(sessionName string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName+":agent", "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// getGitDiffStats calculates git diff statistics for a worktree
func (sr *StateReader) getGitDiffStats(worktreePath string) (int, int) {
	if worktreePath == "" {
		return 0, 0
	}

	// Check if worktree path exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return 0, 0
	}

	// Use git diff --stat to get changes
	shellCmd := "git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null"
	cmd := exec.Command("sh", "-c", shellCmd)
	cmd.Dir = worktreePath

	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	return sr.parseGitDiffStats(string(output))
}

// parseGitDiffStats parses git diff --shortstat output to extract insertions and deletions
func (sr *StateReader) parseGitDiffStats(output string) (int, int) {
	insertions := 0
	deletions := 0

	// Regex patterns to match insertions and deletions
	insRe := regexp.MustCompile(`(\d+) insertion(?:s)?\(\+\)`)
	delRe := regexp.MustCompile(`(\d+) deletion(?:s)?\(\-\)`)

	if m := insRe.FindStringSubmatch(output); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &insertions)
	}
	if m := delRe.FindStringSubmatch(output); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &deletions)
	}

	return insertions, deletions
}

// GetActiveSessions returns only sessions that are currently active in tmux
func (sr *StateReader) GetActiveSessions() ([]SessionInfo, error) {
	allSessions, err := sr.LoadSessions()
	if err != nil {
		return nil, err
	}

	var activeSessions []SessionInfo
	for _, session := range allSessions {
		if session.Status != "inactive" {
			activeSessions = append(activeSessions, session)
		}
	}

	return activeSessions, nil
}

// FilterByRepo filters sessions for a specific git repository
func (sr *StateReader) FilterByRepo(sessions []SessionInfo, repoURL string) []SessionInfo {
	var filtered []SessionInfo
	for _, session := range sessions {
		// This is a simplified check - in practice you might want to compare
		// against the actual git repo from the AgentState
		filtered = append(filtered, session)
	}
	return filtered
}
