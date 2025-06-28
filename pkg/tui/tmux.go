// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// TmuxInterface defines the interface for interacting with tmux
type TmuxInterface interface {
	ListSessions() ([]byte, error)
	ListWindows(sessionName string) ([]byte, error)
	ListPanes(sessionName string) ([]byte, error)
	CapturePane(sessionName string) ([]byte, error)
}

// TmuxReal implements TmuxInterface for real tmux commands
type TmuxReal struct{}

// ListSessions executes the real tmux list-sessions command
func (t *TmuxReal) ListSessions() ([]byte, error) {
	return exec.Command("tmux", "list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}").Output()
}

// ListWindows executes the real tmux list-windows command
func (t *TmuxReal) ListWindows(sessionName string) ([]byte, error) {
	return exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_name}").Output()
}

// ListPanes executes the real tmux list-panes command
func (t *TmuxReal) ListPanes(sessionName string) ([]byte, error) {
	return exec.Command("tmux", "list-panes", "-t", sessionName, "-a", "-F", "#{pane_id}").Output()
}

// CapturePane executes the real tmux capture-pane command
func (t *TmuxReal) CapturePane(sessionName string) ([]byte, error) {
	return exec.Command("tmux", "capture-pane", "-t", sessionName+":agent", "-p").Output()
}

// execCommand allows mocking exec.Command for testing
var execCommand = exec.Command

// TmuxSessionInfo represents information about a tmux session
type TmuxSessionInfo struct {
	Name        string    `json:"name"`
	Windows     int       `json:"windows"`
	Panes       int       `json:"panes"`
	Attached    bool      `json:"attached"`
	Created     time.Time `json:"created"`
	LastUsed    time.Time `json:"last_used"`
	WindowNames []string  `json:"window_names"`
	Activity    string    `json:"activity"` // "active", "inactive", "attached"
}

// TmuxDiscovery provides functionality to discover and analyze tmux sessions
type TmuxDiscovery struct {
	// Cache to avoid calling tmux ls too frequently
	lastUpdate time.Time
	sessions   map[string]TmuxSessionInfo
	cacheTime  time.Duration
	tmux       TmuxInterface
}

// NewTmuxDiscovery creates a new tmux discovery helper
func NewTmuxDiscovery() *TmuxDiscovery {
	return &TmuxDiscovery{
		sessions:  make(map[string]TmuxSessionInfo),
		cacheTime: 2 * time.Second, // Cache for 2 seconds to avoid excessive tmux calls
		tmux:      &TmuxReal{},
	}
}

// GetAllSessions calls `tmux ls` and returns all tmux sessions
func (td *TmuxDiscovery) GetAllSessions() (map[string]TmuxSessionInfo, error) {
	// Check cache first
	if time.Since(td.lastUpdate) < td.cacheTime && len(td.sessions) > 0 {
		return td.sessions, nil
	}

	sessions, err := td.discoverTmuxSessions()
	if err != nil {
		return nil, err
	}

	// Update cache
	td.sessions = sessions
	td.lastUpdate = time.Now()

	return sessions, nil
}

// GetUziSessions returns only tmux sessions that appear to be Uzi agent sessions
func (td *TmuxDiscovery) GetUziSessions() (map[string]TmuxSessionInfo, error) {
	allSessions, err := td.GetAllSessions()
	if err != nil {
		return nil, err
	}

	uziSessions := make(map[string]TmuxSessionInfo)
	for name, session := range allSessions {
		if td.isUziSession(name, session) {
			uziSessions[name] = session
		}
	}

	return uziSessions, nil
}

// MapUziSessionsToTmux maps Uzi session names to their tmux session info
// This is useful for the TUI to highlight which sessions are attached/active
func (td *TmuxDiscovery) MapUziSessionsToTmux(uziSessions []SessionInfo) (map[string]TmuxSessionInfo, error) {
	tmuxSessions, err := td.GetAllSessions()
	if err != nil {
		return nil, err
	}

	sessionMap := make(map[string]TmuxSessionInfo)

	for _, uziSession := range uziSessions {
		// Try to find corresponding tmux session
		if tmuxInfo, exists := tmuxSessions[uziSession.Name]; exists {
			sessionMap[uziSession.Name] = tmuxInfo
		} else {
			// Create a placeholder entry for missing tmux sessions
			sessionMap[uziSession.Name] = TmuxSessionInfo{
				Name:     uziSession.Name,
				Attached: false,
				Activity: "inactive",
			}
		}
	}

	return sessionMap, nil
}

// IsSessionAttached returns true if the given session name is currently attached
func (td *TmuxDiscovery) IsSessionAttached(sessionName string) bool {
	sessions, err := td.GetAllSessions()
	if err != nil {
		return false
	}

	if session, exists := sessions[sessionName]; exists {
		return session.Attached
	}

	return false
}

// GetSessionActivity returns the activity status of a session
func (td *TmuxDiscovery) GetSessionActivity(sessionName string) string {
	sessions, err := td.GetAllSessions()
	if err != nil {
		return "unknown"
	}

	if session, exists := sessions[sessionName]; exists {
		return session.Activity
	}

	return "inactive"
}

// discoverTmuxSessions calls `tmux ls` and parses the output
func (td *TmuxDiscovery) discoverTmuxSessions() (map[string]TmuxSessionInfo, error) {
	// Call tmux list-sessions with detailed format
	output, err := td.tmux.ListSessions()
	if err != nil {
		// Check if it's a tmux error that means "no sessions" (expected condition)
		if strings.Contains(err.Error(), "no server running") {
			// This means tmux is working but no sessions exist - return empty map
			return make(map[string]TmuxSessionInfo), nil
		}
		// All other errors (including "command not found", test errors, etc.) should be propagated
		return nil, err
	}

	sessions := make(map[string]TmuxSessionInfo)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		session, err := td.parseSessionLine(line)
		if err != nil {
			// Log but don't fail completely for one bad line
			continue
		}

		// Get window information for this session
		windowNames, paneCount, err := td.getSessionWindows(session.Name)
		if err != nil {
			// Continue without window info if we can't get it
			windowNames = []string{}
			paneCount = 0
		}

		session.WindowNames = windowNames
		session.Panes = paneCount
		sessions[session.Name] = session
	}

	return sessions, nil
}

// parseSessionLine parses a single line from tmux list-sessions output
func (td *TmuxDiscovery) parseSessionLine(line string) (TmuxSessionInfo, error) {
	// Format: name|windows|attached|created|activity
	parts := strings.Split(line, "|")
	if len(parts) != 5 {
		return TmuxSessionInfo{}, fmt.Errorf("unexpected tmux output format: %s", line)
	}

	name := parts[0]
	windows, _ := strconv.Atoi(parts[1])
	attached := parts[2] == "1"
	createdUnix, _ := strconv.ParseInt(parts[3], 10, 64)
	activityUnix, _ := strconv.ParseInt(parts[4], 10, 64)

	created := time.Unix(createdUnix, 0)
	lastUsed := time.Unix(activityUnix, 0)

	activity := "inactive"
	if attached {
		activity = "attached"
	} else if time.Since(lastUsed) < 5*time.Minute {
		activity = "active"
	}

	return TmuxSessionInfo{
		Name:     name,
		Windows:  windows,
		Attached: attached,
		Created:  created,
		LastUsed: lastUsed,
		Activity: activity,
	}, nil
}

// getSessionWindows gets window names and pane count for a session
func (td *TmuxDiscovery) getSessionWindows(sessionName string) ([]string, int, error) {
	// Get window information
	windowOutput, err := td.tmux.ListWindows(sessionName)
	if err != nil {
		return nil, 0, err
	}

	windowNames := strings.Split(strings.TrimSpace(string(windowOutput)), "\n")
	if len(windowNames) == 1 && windowNames[0] == "" {
		windowNames = []string{}
	}

	// Get pane count
	paneOutput, err := td.tmux.ListPanes(sessionName)
	if err != nil {
		return windowNames, 0, err
	}

	paneLines := strings.Split(strings.TrimSpace(string(paneOutput)), "\n")
	paneCount := len(paneLines)
	if len(paneLines) == 1 && paneLines[0] == "" {
		paneCount = 0
	}

	return windowNames, paneCount, nil
}

// isUziSession determines if a tmux session appears to be a Uzi agent session
func (td *TmuxDiscovery) isUziSession(sessionName string, session TmuxSessionInfo) bool {
	// Check if session name follows Uzi pattern: agent-projectDir-gitHash-agentName
	if strings.HasPrefix(sessionName, "agent-") {
		parts := strings.Split(sessionName, "-")
		if len(parts) >= 4 {
			return true
		}
	}

	// Also check if session has windows that suggest it's a Uzi session
	for _, windowName := range session.WindowNames {
		if windowName == "agent" || windowName == "uzi-dev" {
			return true
		}
	}

	return false
}

// GetSessionStatus returns a more detailed status for a session
func (td *TmuxDiscovery) GetSessionStatus(sessionName string) (string, error) {
	sessions, err := td.GetAllSessions()
	if err != nil {
		return "unknown", err
	}

	session, exists := sessions[sessionName]
	if !exists {
		return "not_found", nil
	}

	if session.Attached {
		return "attached", nil
	}

	// Check if session has any activity in agent window
	if td.hasAgentWindow(sessionName) {
		content, err := td.getAgentWindowContent(sessionName)
		if err != nil {
			return "ready", nil
		}

		// Check for running indicators in the agent pane
		if strings.Contains(content, "esc to interrupt") ||
			strings.Contains(content, "Thinking") ||
			strings.Contains(content, "Working") {
			return "running", nil
		}
	}

	return "ready", nil
}

// hasAgentWindow checks if session has an "agent" window
func (td *TmuxDiscovery) hasAgentWindow(sessionName string) bool {
	sessions, err := td.GetAllSessions()
	if err != nil {
		return false
	}

	session, exists := sessions[sessionName]
	if !exists {
		return false
	}

	for _, windowName := range session.WindowNames {
		if windowName == "agent" {
			return true
		}
	}

	return false
}

// getAgentWindowContent gets the content of the agent window/pane
func (td *TmuxDiscovery) getAgentWindowContent(sessionName string) (string, error) {
	output, err := td.tmux.CapturePane(sessionName)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// RefreshCache forces a refresh of the tmux session cache
func (td *TmuxDiscovery) RefreshCache() {
	td.lastUpdate = time.Time{}
}

// FormatSessionActivity returns a styled string for session activity
func (td *TmuxDiscovery) FormatSessionActivity(activity string) string {
	switch activity {
	case "attached":
		return "🔗" // Link symbol for attached
	case "active":
		return "●" // Dot for active
	case "inactive":
		return "○" // Empty circle for inactive
	default:
		return "?" // Unknown
	}
}

// GetAttachedSessionCount returns the number of currently attached sessions
func (td *TmuxDiscovery) GetAttachedSessionCount() (int, error) {
	sessions, err := td.GetAllSessions()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, session := range sessions {
		if session.Attached {
			count++
		}
	}

	return count, nil
}

// ListSessionsByActivity returns sessions grouped by activity level
func (td *TmuxDiscovery) ListSessionsByActivity() (map[string][]TmuxSessionInfo, error) {
	sessions, err := td.GetUziSessions()
	if err != nil {
		return nil, err
	}

	grouped := map[string][]TmuxSessionInfo{
		"attached": {},
		"active":   {},
		"inactive": {},
	}

	for _, session := range sessions {
		grouped[session.Activity] = append(grouped[session.Activity], session)
	}

	return grouped, nil
}

// GetSessionMatchScore returns a score indicating how well a tmux session matches a Uzi session
// This can be used for fuzzy matching when session names don't exactly match
func (td *TmuxDiscovery) GetSessionMatchScore(tmuxSessionName, uziSessionName string) int {
	if tmuxSessionName == uziSessionName {
		return 100 // Perfect match
	}

	// Check if one contains the other first (before agent extraction)
	if strings.Contains(tmuxSessionName, uziSessionName) || strings.Contains(uziSessionName, tmuxSessionName) {
		return 60 // Partial match (contains)
	}

	// Extract agent name from both and compare
	tmuxAgent := extractAgentNameFromTmux(tmuxSessionName)
	uziAgent := extractAgentNameFromTmux(uziSessionName)

	// If both are proper agent sessions and have same agent name
	if tmuxAgent != tmuxSessionName && uziAgent != uziSessionName && tmuxAgent == uziAgent {
		return 80 // Good match on agent name
	}

	return 0 // No match
}

// Helper function to extract agent name (reuse from existing code)
func extractAgentNameFromTmux(sessionName string) string {
	parts := strings.Split(sessionName, "-")
	if len(parts) >= 4 && parts[0] == "agent" {
		return strings.Join(parts[3:], "-")
	}
	return sessionName
}
