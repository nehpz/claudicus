package tui

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nehpz/claudicus/pkg/testutil/cmdmock"
)

// TmuxMock implements TmuxInterface for testing
type TmuxMock struct {
	ListSessionsFunc func() ([]byte, error)
	ListWindowsFunc  func(sessionName string) ([]byte, error)
	ListPanesFunc    func(sessionName string) ([]byte, error)
	CapturePaneFunc  func(sessionName string) ([]byte, error)
}

// ListSessions calls the mock function
func (m *TmuxMock) ListSessions() ([]byte, error) {
	if m.ListSessionsFunc != nil {
		return m.ListSessionsFunc()
	}
	return nil, nil
}

// ListWindows calls the mock function
func (m *TmuxMock) ListWindows(sessionName string) ([]byte, error) {
	if m.ListWindowsFunc != nil {
		return m.ListWindowsFunc(sessionName)
	}
	return nil, nil
}

// ListPanes calls the mock function
func (m *TmuxMock) ListPanes(sessionName string) ([]byte, error) {
	if m.ListPanesFunc != nil {
		return m.ListPanesFunc(sessionName)
	}
	return nil, nil
}

// CapturePane calls the mock function
func (m *TmuxMock) CapturePane(sessionName string) ([]byte, error) {
	if m.CapturePaneFunc != nil {
		return m.CapturePaneFunc(sessionName)
	}
	return nil, nil
}

// Helper function to strip cmdmock's -n prefix from output
// This accounts for cmdmock using 'echo -n' which adds '-n ' to the beginning
func stripCmdmockPrefix(s string) string {
	if strings.HasPrefix(s, "-n ") {
		return s[3:]
	}
	return s
}

func TestMain(m *testing.M) {
	// Override execCommand for all tests
	execCommand = cmdmock.Command
	code := m.Run()
	cmdmock.Reset()
	os.Exit(code)
}

func setUp() {
	cmdmock.Reset()
	cmdmock.Enable()
}

// Test creation and basic functionality
func TestNewTmuxDiscovery(t *testing.T) {
	td := NewTmuxDiscovery()
	if td == nil {
		t.Fatal("NewTmuxDiscovery returned nil")
	}
	if td.sessions == nil {
		t.Error("sessions map not initialized")
	}
	if td.cacheTime != 2*time.Second {
		t.Errorf("Expected cache time 2s, got %v", td.cacheTime)
	}
}

// Test session discovery with valid tmux output
func TestGetAllSessions_Success(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("session1|2|1|1640000000|1640000010"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			return []byte("agent\ndev"), nil
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			return []byte("%0\n%1"), nil
		},
	}

	sessions, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	session, exists := sessions["session1"]
	if !exists {
		t.Fatal("Expected session1 to exist")
	}

	// Verify session details
	if session.Name != "session1" {
		t.Errorf("Expected name 'session1', got '%s'", session.Name)
	}
	if session.Windows != 2 {
		t.Errorf("Expected 2 windows, got %d", session.Windows)
	}
	if session.Panes != 2 {
		t.Errorf("Expected 2 panes, got %d", session.Panes)
	}
	if !session.Attached {
		t.Error("Expected session to be attached")
	}
	if session.Activity != "attached" {
		t.Errorf("Expected activity 'attached', got '%s'", session.Activity)
	}
	if len(session.WindowNames) != 2 {
		t.Errorf("Expected 2 window names, got %d", len(session.WindowNames))
	}
	if session.WindowNames[0] != "agent" || session.WindowNames[1] != "dev" {
		t.Errorf("Expected window names [agent, dev], got %v", session.WindowNames)
	}
}

// Test session discovery with multiple sessions
func TestGetAllSessions_MultipleSessions(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("attached-session|1|1|1640000000|1640000000\ndetached-session|2|0|1640000000|1640000000\nactive-session|1|0|1640000000|1640000290"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			switch sessionName {
			case "attached-session":
				return []byte("main"), nil
			case "detached-session":
				return []byte("bash\nhtop"), nil
			case "active-session":
				return []byte("work"), nil
			}
			return nil, fmt.Errorf("unexpected session name: %s", sessionName)
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			switch sessionName {
			case "attached-session":
				return []byte("%0"), nil
			case "detached-session":
				return []byte("%0\n%1"), nil
			case "active-session":
				return []byte("%0"), nil
			}
			return nil, fmt.Errorf("unexpected session name: %s", sessionName)
		},
	}

	sessions, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(sessions) != 3 {
		t.Fatalf("Expected 3 sessions, got %d", len(sessions))
	}

	// Check attached session
	if session := sessions["attached-session"]; !session.Attached || session.Activity != "attached" {
		t.Errorf("Attached session state incorrect: attached=%v, activity=%s", session.Attached, session.Activity)
	}

	// Check detached session
	if session := sessions["detached-session"]; session.Attached || session.Activity != "inactive" {
		t.Errorf("Detached session state incorrect: attached=%v, activity=%s", session.Attached, session.Activity)
	}

	// Check active session (recent activity)
	if session := sessions["active-session"]; session.Attached || session.Activity != "active" {
		t.Errorf("Active session state incorrect: attached=%v, activity=%s", session.Attached, session.Activity)
	}
}

// Test when tmux command fails (no sessions exist)
func TestGetAllSessions_NoSessions(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return nil, fmt.Errorf("no server running")
		},
	}

	sessions, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("Should not error when no sessions exist: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected empty session map, got %d sessions", len(sessions))
	}
}

// Test tmux not installed
func TestGetAllSessions_TmuxNotInstalled(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return nil, fmt.Errorf("command not found: tmux")
		},
	}

	sessions, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("Should handle tmux not found gracefully: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected empty session map when tmux not found, got %d sessions", len(sessions))
	}
}

// Test caching functionality
func TestGetAllSessions_Caching(t *testing.T) {
	setUp()

	callCount := 0
	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			callCount++
			return []byte("session1|1|0|1640000000|1640000000"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			return []byte("main"), nil
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			return []byte("%0"), nil
		},
	}

	// First call should hit tmux
	sessions1, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Verify command was called
	if callCount != 1 {
		t.Error("Expected tmux command to be called")
	}

	// Second call within cache time should use cache
	sessions2, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	// Should be same data and no additional calls
	if len(sessions1) != len(sessions2) {
		t.Error("Cached result should be identical")
	}
	if callCount > 1 {
		t.Error("Second call should use cache, not call tmux again")
	}
}

// Test cache refresh
func TestRefreshCache(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.lastUpdate = time.Now() // Set recent update

	td.RefreshCache()

	if !td.lastUpdate.IsZero() {
		t.Error("RefreshCache should reset lastUpdate to zero")
	}
}

// Test Uzi session identification
func TestGetUziSessions(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("agent-proj-abc123-claude|1|0|1640000000|1640000000\nregular-session|2|1|1640001000|1640001000\nuzi-dev-session|1|0|1640002000|1640002000"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			switch sessionName {
			case "agent-proj-abc123-claude":
				return []byte("agent"), nil
			case "regular-session":
				return []byte("bash\nhtop"), nil
			case "uzi-dev-session":
				return []byte("uzi-dev"), nil
			}
			return nil, fmt.Errorf("unexpected session name: %s", sessionName)
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			switch sessionName {
			case "agent-proj-abc123-claude":
				return []byte("%0"), nil
			case "regular-session":
				return []byte("%0\n%1"), nil
			case "uzi-dev-session":
				return []byte("%0"), nil
			}
			return nil, fmt.Errorf("unexpected session name: %s", sessionName)
		},
	}

	uziSessions, err := td.GetUziSessions()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return both Uzi sessions (agent prefix + uzi-dev window)
	expectedCount := 2
	if len(uziSessions) != expectedCount {
		t.Errorf("Expected %d Uzi sessions, got %d", expectedCount, len(uziSessions))
	}

	if _, exists := uziSessions["agent-proj-abc123-claude"]; !exists {
		t.Error("Expected agent session to be identified as Uzi session")
	}
	if _, exists := uziSessions["uzi-dev-session"]; !exists {
		t.Error("Expected uzi-dev session to be identified as Uzi session")
	}
	if _, exists := uziSessions["regular-session"]; exists {
		t.Error("Regular session should not be identified as Uzi session")
	}
}

// Test Uzi session identification edge cases
func TestIsUziSession_EdgeCases(t *testing.T) {
	td := NewTmuxDiscovery()

	tests := []struct {
		name        string
		sessionName string
		windowNames []string
		expected    bool
	}{
		{"Agent prefix with 4 parts", "agent-proj-hash-claude", []string{"main"}, true},
		{"Agent prefix with 5 parts", "agent-proj-hash-claude-v2", []string{"main"}, true},
		{"Agent prefix with 3 parts", "agent-proj-hash", []string{"main"}, false},
		{"Non-agent prefix", "my-agent-session", []string{"main"}, false},
		{"Agent window", "some-session", []string{"agent"}, true},
		{"Uzi-dev window", "dev-session", []string{"uzi-dev"}, true},
		{"Regular session", "normal", []string{"bash", "vim"}, false},
		{"Empty session", "empty", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := TmuxSessionInfo{WindowNames: tt.windowNames}
			result := td.isUziSession(tt.sessionName, session)
			if result != tt.expected {
				t.Errorf("isUziSession(%s, %v) = %v, expected %v", tt.sessionName, tt.windowNames, result, tt.expected)
			}
		})
	}
}

// Test session status checks
func TestIsSessionAttached(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("attached-session|1|1|1640000000|1640000000\ndetached-session|1|0|1640000000|1640000000"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			return []byte("main"), nil
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			return []byte("%0"), nil
		},
	}

	if !td.IsSessionAttached("attached-session") {
		t.Error("Expected attached-session to be attached")
	}
	if td.IsSessionAttached("detached-session") {
		t.Error("Expected detached-session to not be attached")
	}
	if td.IsSessionAttached("nonexistent-session") {
		t.Error("Expected nonexistent session to not be attached")
	}
}

// Test session activity detection
func TestGetSessionActivity(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("test-session|1|0|1640000000|1640000000"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			return []byte("main"), nil
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			return []byte("%0"), nil
		},
	}

	activity := td.GetSessionActivity("test-session")
	if activity != "inactive" {
		t.Errorf("Expected inactive activity, got %s", activity)
	}

	// Test unknown session
	activity = td.GetSessionActivity("unknown-session")
	if activity != "inactive" {
		t.Errorf("Expected inactive activity for unknown session, got %s", activity)
	}
}

// Test activity when tmux command fails
func TestGetSessionActivity_TmuxError(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return nil, fmt.Errorf("command not found: tmux")
		},
	}

	activity := td.GetSessionActivity("any-session")
	if activity != "unknown" {
		t.Errorf("Expected unknown activity when tmux fails, got %s", activity)
	}
}

// Test parsing functions
func TestParseSessionLine(t *testing.T) {
	td := NewTmuxDiscovery()

	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    TmuxSessionInfo
	}{
		{
			name:  "Valid attached session",
			input: "session1|2|1|1640000000|1640000100",
			expected: TmuxSessionInfo{
				Name:     "session1",
				Windows:  2,
				Attached: true,
				Activity: "attached",
			},
		},
		{
			name:  "Valid detached inactive session",
			input: "session2|1|0|1640000000|1640000000",
			expected: TmuxSessionInfo{
				Name:     "session2",
				Windows:  1,
				Attached: false,
				Activity: "inactive",
			},
		},
		{
			name:        "Invalid format - too few parts",
			input:       "session1|2|1",
			expectError: true,
		},
		{
			name:        "Invalid format - too many parts",
			input:       "session1|2|1|1640000000|1640000000|extra",
			expectError: true,
		},
		{
			name:        "Empty input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := td.parseSessionLine(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Name != tt.expected.Name {
				t.Errorf("Expected name %s, got %s", tt.expected.Name, result.Name)
			}
			if result.Windows != tt.expected.Windows {
				t.Errorf("Expected %d windows, got %d", tt.expected.Windows, result.Windows)
			}
			if result.Attached != tt.expected.Attached {
				t.Errorf("Expected attached %v, got %v", tt.expected.Attached, result.Attached)
			}
			if result.Activity != tt.expected.Activity {
				t.Errorf("Expected activity %s, got %s", tt.expected.Activity, result.Activity)
			}
		})
	}
}

// Test activity classification based on time
func TestParseSessionLine_ActivityClassification(t *testing.T) {
	td := NewTmuxDiscovery()
	now := time.Now().Unix()

	tests := []struct {
		name             string
		attached         string
		lastActivityTime int64
		expectedActivity string
	}{
		{"Attached session", "1", now, "attached"},
		{"Recently active (2 min ago)", "0", now - 120, "active"},
		{"Inactive (10 min ago)", "0", now - 600, "inactive"},
		{"Very old session", "0", 1640000000, "inactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf("session|1|%s|1640000000|%d", tt.attached, tt.lastActivityTime)
			result, err := td.parseSessionLine(input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result.Activity != tt.expectedActivity {
				t.Errorf("Expected activity %s, got %s", tt.expectedActivity, result.Activity)
			}
		})
	}
}

// Test cmdmock behavior - we need to understand the output format
func TestCmdmockBehavior(t *testing.T) {
	setUp()

	// Test what cmdmock actually produces for empty output
	cmdmock.SetResponseWithArgs("tmux", []string{"test-empty"}, "", "", false)
	cmd := execCommand("tmux", "test-empty")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	t.Logf("Empty output: '%s'", string(output))

	// Test what cmdmock produces for non-empty output
	cmdmock.SetResponseWithArgs("tmux", []string{"test-content"}, "hello", "", false)
	cmd = execCommand("tmux", "test-content")
	output, err = cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	t.Logf("Content output: '%s'", string(output))

	// Test with multiline
	cmdmock.SetResponseWithArgs("tmux", []string{"test-multiline"}, "line1\nline2", "", false)
	cmd = execCommand("tmux", "test-multiline")
	output, err = cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	t.Logf("Multiline output: '%s'", string(output))
}

// Test window and pane parsing
func TestGetSessionWindows(t *testing.T) {
	setUp()

	tests := []struct {
		name                string
		sessionName         string
		windowOutput        string
		paneOutput          string
		windowError         bool
		paneError           bool
		expectedWindowNames []string
		expectedPaneCount   int
		expectError         bool
	}{
		{
			name:                "Normal session with multiple windows and panes",
			sessionName:         "test-session",
			windowOutput:        "main\ndev\nlog",
			paneOutput:          "%0\n%1\n%2",
			expectedWindowNames: []string{"main", "dev", "log"},
			expectedPaneCount:   3,
		},
		{
			name:                "Single window session",
			sessionName:         "simple",
			windowOutput:        "bash",
			paneOutput:          "%0",
			expectedWindowNames: []string{"bash"},
			expectedPaneCount:   1,
		},
		{
			name:                "Empty session",
			sessionName:         "empty",
			windowOutput:        "",
			paneOutput:          "",
			expectedWindowNames: []string{},
			expectedPaneCount:   0,
		},
		{
			name:        "Window command fails",
			sessionName: "fail-windows",
			windowError: true,
			expectError: true,
		},
		{
			name:                "Pane command fails",
			sessionName:         "fail-panes",
			windowOutput:        "main",
			paneError:           true,
			expectedWindowNames: []string{"main"},
			expectedPaneCount:   0,
			expectError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUp()

			td := NewTmuxDiscovery()
			td.tmux = &TmuxMock{
				ListWindowsFunc: func(sessionName string) ([]byte, error) {
					if tt.windowError {
						return nil, fmt.Errorf("session not found")
					}
					return []byte(tt.windowOutput), nil
				},
				ListPanesFunc: func(sessionName string) ([]byte, error) {
					if tt.paneError {
						return nil, fmt.Errorf("session not found")
					}
					return []byte(tt.paneOutput), nil
				},
			}

			windowNames, paneCount, err := td.getSessionWindows(tt.sessionName)

			// Log what we got for debugging
			t.Logf("Test %s: got windows=%v, panes=%d, err=%v", tt.name, windowNames, paneCount, err)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(windowNames) != len(tt.expectedWindowNames) {
				t.Errorf("Expected %d window names, got %d: %v", len(tt.expectedWindowNames), len(windowNames), windowNames)
			}

			for i, expected := range tt.expectedWindowNames {
				if i >= len(windowNames) || windowNames[i] != expected {
					t.Errorf("Expected window name %s at index %d, got %v", expected, i, windowNames)
					break
				}
			}

			if paneCount != tt.expectedPaneCount {
				t.Errorf("Expected %d panes, got %d", tt.expectedPaneCount, paneCount)
			}
		})
	}
}

// Test session mapping functionality
func TestMapUziSessionsToTmux(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("agent-proj-abc123-claude|1|1|1640000000|1640000000\nother-session|1|0|1640000000|1640000000"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			switch sessionName {
			case "agent-proj-abc123-claude":
				return []byte("agent"), nil
			case "other-session":
				return []byte("bash"), nil
			}
			return nil, fmt.Errorf("unexpected session name: %s", sessionName)
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			return []byte("%0"), nil
		},
	}

	// Test with matching session
	uziSessions := []SessionInfo{
		{Name: "agent-proj-abc123-claude"},
		{Name: "missing-session"},
	}

	sessionMap, err := td.MapUziSessionsToTmux(uziSessions)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(sessionMap) != 2 {
		t.Errorf("Expected 2 mapped sessions, got %d", len(sessionMap))
	}

	// Check existing session
	if session, exists := sessionMap["agent-proj-abc123-claude"]; !exists {
		t.Error("Expected agent session to be mapped")
	} else if !session.Attached {
		t.Error("Expected mapped session to show attached state")
	}

	// Check missing session (should have placeholder)
	if session, exists := sessionMap["missing-session"]; !exists {
		t.Error("Expected missing session to have placeholder")
	} else {
		if session.Attached {
			t.Error("Expected placeholder session to not be attached")
		}
		if session.Activity != "inactive" {
			t.Errorf("Expected placeholder activity 'inactive', got '%s'", session.Activity)
		}
	}
}

// Test when MapUziSessionsToTmux fails to get sessions
func TestMapUziSessionsToTmux_Error(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return nil, fmt.Errorf("no server running")
		},
	}
	uziSessions := []SessionInfo{{Name: "test-session"}}

	_, err := td.MapUziSessionsToTmux(uziSessions)
	if err != nil {
		t.Error("Should not error when tmux fails - should return empty map gracefully")
	}
}

// Test detailed session status detection
func TestGetSessionStatus(t *testing.T) {
	setUp()

	tests := []struct {
		name           string
		sessionName    string
		tmuxOutput     string
		windowOutput   string
		paneOutput     string
		agentContent   string
		expectedStatus string
		tmuxError      bool
		captureError   bool
	}{
		{
			name:           "Attached session",
			sessionName:    "attached-session",
			tmuxOutput:     "attached-session|1|1|1640000000|1640000000",
			windowOutput:   "main",
			paneOutput:     "%0",
			expectedStatus: "attached",
		},
		{
			name:           "Session with agent window - ready",
			sessionName:    "agent-session",
			tmuxOutput:     "agent-session|1|0|1640000000|1640000000",
			windowOutput:   "agent",
			paneOutput:     "%0",
			agentContent:   "$ waiting for command",
			expectedStatus: "ready",
		},
		{
			name:           "Session with agent window - running (thinking)",
			sessionName:    "thinking-session",
			tmuxOutput:     "thinking-session|1|0|1640000000|1640000000",
			windowOutput:   "agent",
			paneOutput:     "%0",
			agentContent:   "Thinking about your request...",
			expectedStatus: "running",
		},
		{
			name:           "Session with agent window - running (working)",
			sessionName:    "working-session",
			tmuxOutput:     "working-session|1|0|1640000000|1640000000",
			windowOutput:   "agent",
			paneOutput:     "%0",
			agentContent:   "Working on task...",
			expectedStatus: "running",
		},
		{
			name:           "Session with agent window - running (esc to interrupt)",
			sessionName:    "interrupt-session",
			tmuxOutput:     "interrupt-session|1|0|1640000000|1640000000",
			windowOutput:   "agent",
			paneOutput:     "%0",
			agentContent:   "Press esc to interrupt current operation",
			expectedStatus: "running",
		},
		{
			name:           "Session without agent window",
			sessionName:    "regular-session",
			tmuxOutput:     "regular-session|1|0|1640000000|1640000000",
			windowOutput:   "bash",
			paneOutput:     "%0",
			expectedStatus: "ready",
		},
		{
			name:           "Nonexistent session",
			sessionName:    "missing-session",
			tmuxOutput:     "other-session|1|0|1640000000|1640000000",
			windowOutput:   "bash",
			paneOutput:     "%0",
			expectedStatus: "not_found",
		},
		{
			name:           "Tmux command fails",
			sessionName:    "any-session",
			tmuxError:      true,
			expectedStatus: "unknown",
		},
		{
			name:           "Capture pane fails - defaults to ready",
			sessionName:    "capture-fail",
			tmuxOutput:     "capture-fail|1|0|1640000000|1640000000",
			windowOutput:   "agent",
			paneOutput:     "%0",
			captureError:   true,
			expectedStatus: "ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUp()

			td := NewTmuxDiscovery()
			td.tmux = &TmuxMock{
				ListSessionsFunc: func() ([]byte, error) {
					if tt.tmuxError {
						return nil, fmt.Errorf("command not found: tmux")
					}
					return []byte(tt.tmuxOutput), nil
				},
				ListWindowsFunc: func(sessionName string) ([]byte, error) {
					return []byte(tt.windowOutput), nil
				},
				ListPanesFunc: func(sessionName string) ([]byte, error) {
					return []byte(tt.paneOutput), nil
				},
				CapturePaneFunc: func(sessionName string) ([]byte, error) {
					if tt.captureError {
						return nil, fmt.Errorf("no such window")
					}
					return []byte(tt.agentContent), nil
				},
			}

			status, err := td.GetSessionStatus(tt.sessionName)

			if tt.expectedStatus == "unknown" {
				if err == nil {
					t.Error("Expected error for unknown status")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			if status != tt.expectedStatus {
				t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, status)
			}
		})
	}
}

// Test agent window detection
func TestHasAgentWindow(t *testing.T) {
	setUp()

	tests := []struct {
		name         string
		sessionName  string
		tmuxOutput   string
		windowOutput string
		paneOutput   string
		expected     bool
		tmuxError    bool
	}{
		{
			name:         "Session with agent window",
			sessionName:  "agent-session",
			tmuxOutput:   "agent-session|1|0|1640000000|1640000000",
			windowOutput: "agent",
			paneOutput:   "%0",
			expected:     true,
		},
		{
			name:         "Session with multiple windows including agent",
			sessionName:  "multi-session",
			tmuxOutput:   "multi-session|2|0|1640000000|1640000000",
			windowOutput: "bash\nagent",
			paneOutput:   "%0\n%1",
			expected:     true,
		},
		{
			name:         "Session without agent window",
			sessionName:  "regular-session",
			tmuxOutput:   "regular-session|1|0|1640000000|1640000000",
			windowOutput: "bash",
			paneOutput:   "%0",
			expected:     false,
		},
		{
			name:        "Tmux command fails",
			sessionName: "any-session",
			tmuxError:   true,
			expected:    false,
		},
		{
			name:         "Nonexistent session",
			sessionName:  "missing-session",
			tmuxOutput:   "other-session|1|0|1640000000|1640000000",
			windowOutput: "bash",
			paneOutput:   "%0",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUp()

			td := NewTmuxDiscovery()
			td.tmux = &TmuxMock{
				ListSessionsFunc: func() ([]byte, error) {
					if tt.tmuxError {
						return nil, fmt.Errorf("no server running")
					}
					return []byte(tt.tmuxOutput), nil
				},
				ListWindowsFunc: func(sessionName string) ([]byte, error) {
					return []byte(tt.windowOutput), nil
				},
				ListPanesFunc: func(sessionName string) ([]byte, error) {
					return []byte(tt.paneOutput), nil
				},
			}

			result := td.hasAgentWindow(tt.sessionName)

			if result != tt.expected {
				t.Errorf("Expected hasAgentWindow to return %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test agent window content retrieval
func TestGetAgentWindowContent(t *testing.T) {
	setUp()

	tests := []struct {
		name            string
		sessionName     string
		expectedContent string
		expectError     bool
	}{
		{
			name:            "Successful content retrieval",
			sessionName:     "agent-session",
			expectedContent: "$ echo hello\nhello\n$ ",
			expectError:     false,
		},
		{
			name:        "Command fails",
			sessionName: "missing-session",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUp()

			td := NewTmuxDiscovery()
			td.tmux = &TmuxMock{
				CapturePaneFunc: func(sessionName string) ([]byte, error) {
					if tt.expectError {
						return nil, fmt.Errorf("no such window")
					}
					return []byte(tt.expectedContent), nil
				},
			}

			content, err := td.getAgentWindowContent(tt.sessionName)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}
		})
	}
}

// Test activity symbol formatting
func TestFormatSessionActivity(t *testing.T) {
	td := NewTmuxDiscovery()

	tests := []struct {
		activity string
		expected string
	}{
		{"attached", "🔗"},
		{"active", "●"},
		{"inactive", "○"},
		{"unknown", "?"},
		{"", "?"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Activity_%s", tt.activity), func(t *testing.T) {
			result := td.FormatSessionActivity(tt.activity)
			if result != tt.expected {
				t.Errorf("Expected symbol '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Test attached session counting
func TestGetAttachedSessionCount(t *testing.T) {
	setUp()

	tests := []struct {
		name          string
		tmuxOutput    string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "Multiple sessions with some attached",
			tmuxOutput:    "session1|1|1|1640000000|1640000000\nsession2|1|0|1640000000|1640000000\nsession3|1|1|1640000000|1640000000",
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "No attached sessions",
			tmuxOutput:    "session1|1|0|1640000000|1640000000\nsession2|1|0|1640000000|1640000000",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "All sessions attached",
			tmuxOutput:    "session1|1|1|1640000000|1640000000\nsession2|1|1|1640000000|1640000000",
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:        "Tmux command fails",
			tmuxOutput:  "", // Not used when expectError is true
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setUp()

			td := NewTmuxDiscovery()
			td.tmux = &TmuxMock{
				ListSessionsFunc: func() ([]byte, error) {
					if tt.expectError {
						return nil, fmt.Errorf("command not found: tmux")
					}
					return []byte(tt.tmuxOutput), nil
				},
				ListWindowsFunc: func(sessionName string) ([]byte, error) {
					return []byte("main"), nil
				},
				ListPanesFunc: func(sessionName string) ([]byte, error) {
					return []byte("%0"), nil
				},
			}

			count, err := td.GetAttachedSessionCount()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if count != tt.expectedCount {
				t.Errorf("Expected %d attached sessions, got %d", tt.expectedCount, count)
			}
		})
	}
}

// Test session grouping by activity
func TestListSessionsByActivity(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("agent-proj1-abc-claude|1|1|1640000000|1640000000\nagent-proj2-def-claude|1|0|1640000000|1640000290\nagent-proj3-ghi-claude|1|0|1640000000|1640000000\nregular-session|1|0|1640000000|1640000000"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			return []byte("agent"), nil
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			return []byte("%0"), nil
		},
	}

	grouped, err := td.ListSessionsByActivity()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have 3 groups
	if len(grouped) != 3 {
		t.Errorf("Expected 3 activity groups, got %d", len(grouped))
	}

	// Check attached group
	if len(grouped["attached"]) != 1 {
		t.Errorf("Expected 1 attached session, got %d", len(grouped["attached"]))
	}

	// Check active group (recent activity)
	if len(grouped["active"]) != 1 {
		t.Errorf("Expected 1 active session, got %d", len(grouped["active"]))
	}

	// Check inactive group
	if len(grouped["inactive"]) != 1 {
		t.Errorf("Expected 1 inactive session, got %d", len(grouped["inactive"]))
	}
}

// Test session match scoring
func TestGetSessionMatchScore(t *testing.T) {
	td := NewTmuxDiscovery()

	tests := []struct {
		name          string
		tmuxSession   string
		uziSession    string
		expectedScore int
	}{
		{"Perfect match", "agent-proj-abc123-claude", "agent-proj-abc123-claude", 100},
		{"Agent name match", "agent-proj-abc123-claude", "agent-proj-def456-claude", 80},
		{"Partial match - contains", "agent-proj-abc123-claude-v2", "claude", 60},
		{"Partial match - contained", "claude", "agent-proj-abc123-claude", 60},
		{"No match", "agent-proj-abc123-claude", "agent-proj-def456-gpt4", 0},
		{"No match - completely different", "regular-session", "other-session", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := td.GetSessionMatchScore(tt.tmuxSession, tt.uziSession)
			if score != tt.expectedScore {
				t.Errorf("Expected score %d, got %d", tt.expectedScore, score)
			}
		})
	}
}

// Test agent name extraction helper
func TestExtractAgentNameFromTmux(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		expected    string
	}{
		{"Valid agent session", "agent-proj-abc123-claude", "claude"},
		{"Valid agent session with suffix", "agent-proj-abc123-claude-v2", "claude-v2"},
		{"Not an agent session", "regular-session", "regular-session"},
		{"Agent prefix but too few parts", "agent-proj-abc", "agent-proj-abc"},
		{"Empty session name", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAgentNameFromTmux(tt.sessionName)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Test error handling in discoverTmuxSessions with bad parse lines
func TestDiscoverTmuxSessions_BadParseLines(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("good-session|1|0|1640000000|1640000000\nbad-line-too-few|parts\ngood-session2|2|1|1640000000|1640000000\n\nempty-line-ignored"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			switch sessionName {
			case "good-session":
				return []byte("main"), nil
			case "good-session2":
				return []byte("main\ndev"), nil
			}
			return nil, fmt.Errorf("unexpected session name: %s", sessionName)
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			switch sessionName {
			case "good-session":
				return []byte("%0"), nil
			case "good-session2":
				return []byte("%0\n%1"), nil
			}
			return nil, fmt.Errorf("unexpected session name: %s", sessionName)
		},
	}

	sessions, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("Should not error with mixed good/bad lines: %v", err)
	}

	// Should only have the good sessions
	if len(sessions) != 2 {
		t.Errorf("Expected 2 valid sessions, got %d", len(sessions))
	}

	if _, exists := sessions["good-session"]; !exists {
		t.Error("Expected good-session to exist")
	}
	if _, exists := sessions["good-session2"]; !exists {
		t.Error("Expected good-session2 to exist")
	}
}

// Test error handling when window/pane commands fail but session parsing succeeds
func TestDiscoverTmuxSessions_WindowPaneErrors(t *testing.T) {
	setUp()

	td := NewTmuxDiscovery()
	td.tmux = &TmuxMock{
		ListSessionsFunc: func() ([]byte, error) {
			return []byte("good-session|1|0|1640000000|1640000000\nfail-session|1|1|1640000000|1640000000"), nil
		},
		ListWindowsFunc: func(sessionName string) ([]byte, error) {
			if sessionName == "fail-session" {
				return nil, fmt.Errorf("session not found")
			}
			return []byte("main"), nil
		},
		ListPanesFunc: func(sessionName string) ([]byte, error) {
			if sessionName == "fail-session" {
				return nil, fmt.Errorf("session not found")
			}
			return []byte("%0"), nil
		},
	}

	sessions, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("Should not error when window/pane commands fail: %v", err)
	}

	// Should have both sessions, but fail-session should have empty window info
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}

	goodSession := sessions["good-session"]
	if len(goodSession.WindowNames) != 1 || goodSession.WindowNames[0] != "main" {
		t.Errorf("Expected good session to have window info: %v", goodSession.WindowNames)
	}
	if goodSession.Panes != 1 {
		t.Errorf("Expected good session to have 1 pane, got %d", goodSession.Panes)
	}

	failSession := sessions["fail-session"]
	if len(failSession.WindowNames) != 0 {
		t.Errorf("Expected fail session to have no window info, got: %v", failSession.WindowNames)
	}
	if failSession.Panes != 0 {
		t.Errorf("Expected fail session to have 0 panes, got %d", failSession.Panes)
	}
}
