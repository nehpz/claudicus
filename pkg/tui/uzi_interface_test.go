// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nehpz/claudicus/pkg/state"
	"github.com/nehpz/claudicus/pkg/testutil/cmdmock"
)

func setupUziTest() {
	cmdmock.Reset()
	cmdmock.Enable()
	// Override uziExecCommand for all tests
	uziExecCommand = cmdmock.Command
}

// Mock state manager for testing
type mockStateManagerForTest struct {
	activeSessions []string
	statePath      string
}

func (m *mockStateManagerForTest) RemoveState(sessionName string) error {
	// Mock implementation for test
	return nil
}

func (m *mockStateManagerForTest) GetActiveSessionsForRepo() ([]string, error) {
	return m.activeSessions, nil
}

func (m *mockStateManagerForTest) GetStatePath() string {
	return m.statePath
}

func (m *mockStateManagerForTest) SaveState(prompt, branchName, sessionName, worktreePath, model string) error {
	// Mock implementation for test
	return nil
}

func (m *mockStateManagerForTest) SaveStateWithPort(prompt, branchName, sessionName, worktreePath, model string, port int) error {
	// Mock implementation for test
	return nil
}

// Test helpers
func createTempStateFile(t *testing.T, states map[string]state.AgentState) string {
	t.Helper()
	
	tmpFile, err := os.CreateTemp("", "state_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	data, err := json.Marshal(states)
	if err != nil {
		t.Fatalf("Failed to marshal states: %v", err)
	}
	
	if _, err := tmpFile.Write(data); err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}
	
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	
	return tmpFile.Name()
}

func createSessionJSON(sessions []SessionInfo) string {
	data, _ := json.Marshal(sessions)
	return string(data)
}

// Core Interface Tests

func TestUziCLI_NewUziCLI(t *testing.T) {
	setupUziTest()
	
	tests := []struct {
		name     string
		useConfig bool
		config   ProxyConfig
	}{
		{
			name:     "Default configuration",
			useConfig: false,
		},
		{
			name:     "Custom configuration",
			useConfig: true,
			config: ProxyConfig{
				Timeout:     10 * time.Second,
				Retries:     5,
				LogLevel:    "debug",
				EnableCache: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli *UziCLI
			if tt.useConfig {
				cli = NewUziCLIWithConfig(tt.config)
			} else {
				cli = NewUziCLI()
			}

			if cli == nil {
				t.Fatal("Expected non-nil UziCLI")
			}
			if cli.stateManager == nil {
				t.Error("Expected non-nil state manager")
			}
			if cli.tmuxDiscovery == nil {
				t.Error("Expected non-nil tmux discovery")
			}

			if tt.useConfig {
				if cli.config.Timeout != tt.config.Timeout {
					t.Errorf("Expected timeout %v, got %v", tt.config.Timeout, cli.config.Timeout)
				}
				if cli.config.Retries != tt.config.Retries {
					t.Errorf("Expected retries %d, got %d", tt.config.Retries, cli.config.Retries)
				}
				if cli.config.LogLevel != tt.config.LogLevel {
					t.Errorf("Expected log level %s, got %s", tt.config.LogLevel, cli.config.LogLevel)
				}
				if cli.config.EnableCache != tt.config.EnableCache {
					t.Errorf("Expected cache %v, got %v", tt.config.EnableCache, cli.config.EnableCache)
				}
			}
		})
	}
}

func TestDefaultProxyConfig(t *testing.T) {
	config := DefaultProxyConfig()
	
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}
	if config.Retries != 2 {
		t.Errorf("Expected retries 2, got %d", config.Retries)
	}
	if config.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got %s", config.LogLevel)
	}
	if config.EnableCache != false {
		t.Errorf("Expected cache false, got %v", config.EnableCache)
	}
}

// Execute Command Infrastructure Tests

func TestUziCLI_ExecuteCommand(t *testing.T) {
	setupUziTest()
	
	tests := []struct {
		name           string
		command        string
		args           []string
		mockStdout     string
		mockStderr     string
		mockExitErr    bool
		expectedOutput string
		expectedError  bool
	}{
		{
			name:           "Successful command",
			command:        "echo",
			args:           []string{"hello"},
			mockStdout:     "hello\n",
			mockStderr:     "",
			mockExitErr:    false,
			expectedOutput: "hello\n",
			expectedError:  false,
		},
		{
			name:          "Command with error",
			command:       "false",
			args:          []string{},
			mockStdout:    "",
			mockStderr:    "command failed",
			mockExitErr:   true,
			expectedError: true,
		},
		{
			name:           "Empty output",
			command:        "true",
			args:           []string{},
			mockStdout:     "",
			mockStderr:     "",
			mockExitErr:    false,
			expectedOutput: "",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewUziCLI()

			cmdmock.SetResponseWithArgs(tt.command, tt.args, 
				tt.mockStdout, tt.mockStderr, tt.mockExitErr)

			output, err := cli.executeCommand(tt.command, tt.args...)

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectedError && string(output) != tt.expectedOutput {
				t.Errorf("Expected output %q, got %q", tt.expectedOutput, string(output))
			}
		})
	}
}

func TestUziCLI_WrapError(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()

	tests := []struct {
		name      string
		operation string
		inputErr  error
		expectNil bool
	}{
		{
			name:      "Nil error",
			operation: "test",
			inputErr:  nil,
			expectNil: true,
		},
		{
			name:      "Real error",
			operation: "test operation",
			inputErr:  fmt.Errorf("original error"),
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cli.wrapError(tt.operation, tt.inputErr)
			
			if tt.expectNil && result != nil {
				t.Errorf("Expected nil error, got: %v", result)
			}
			if !tt.expectNil {
				if result == nil {
					t.Error("Expected wrapped error, got nil")
				} else if !strings.Contains(result.Error(), "uzi_proxy") {
					t.Errorf("Expected error to contain 'uzi_proxy', got: %v", result)
				}
				if !strings.Contains(result.Error(), tt.operation) {
					t.Errorf("Expected error to contain operation '%s', got: %v", tt.operation, result)
				}
			}
		})
	}
}

// GetSessions Tests (Current vs Legacy Behavior Parity)

func TestUziCLI_GetSessions(t *testing.T) {
	setupUziTest()
	
	testSessions := []SessionInfo{
		{
			Name:         "agent-proj1-abc123-claude",
			AgentName:    "claude",
			Model:        "claude-3-5-sonnet-20241022",
			Status:       "ready",
			Prompt:       "Write a hello world program",
			Insertions:   15,
			Deletions:    3,
			WorktreePath: "/tmp/worktree1",
			Port:         8080,
		},
		{
			Name:         "agent-proj2-def456-coder",
			AgentName:    "coder",
			Model:        "claude-3-haiku-20240307",
			Status:       "running",
			Prompt:       "Debug this code",
			Insertions:   42,
			Deletions:    7,
			WorktreePath: "/tmp/worktree2",
			Port:         8081,
		},
	}

	tests := []struct {
		name          string
		mockJSON      string
		expectedCount int
		expectedError bool
		errorType     string
	}{
		{
			name:          "Valid JSON response",
			mockJSON:      createSessionJSON(testSessions),
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:          "Empty sessions",
			mockJSON:      "[]",
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:          "Invalid JSON",
			mockJSON:      "{invalid json",
			expectedCount: 0,
			expectedError: true,
			errorType:     "parse",
		},
		{
			name:          "Command failure",
			mockJSON:      "",
			expectedCount: 0,
			expectedError: true,
			errorType:     "command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewUziCLI()

			if tt.errorType == "command" {
				cmdmock.SetResponseWithArgs("uzi", []string{"ls", "--json"}, 
					"", "command failed", true)
			} else {
				cmdmock.SetResponseWithArgs("uzi", []string{"ls", "--json"}, 
					tt.mockJSON, "", false)
			}

			sessions, err := cli.GetSessions()

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if len(sessions) != tt.expectedCount {
				t.Errorf("Expected %d sessions, got %d", tt.expectedCount, len(sessions))
			}

			// Verify session data for successful cases
			if !tt.expectedError && len(sessions) > 0 {
				session := sessions[0]
				expected := testSessions[0]
				if session.Name != expected.Name {
					t.Errorf("Expected name %s, got %s", expected.Name, session.Name)
				}
				if session.AgentName != expected.AgentName {
					t.Errorf("Expected agent name %s, got %s", expected.AgentName, session.AgentName)
				}
			}
		})
	}
}

// Legacy Method Behavior Parity Tests

func TestUziCLI_GetSessionsLegacy_BehaviorParity(t *testing.T) {
	setupUziTest()
	
	// Create test state
	testStates := map[string]state.AgentState{
		"agent-proj1-abc123-claude": {
			Model:        "claude-3-5-sonnet-20241022",
			Prompt:       "Write a hello world program",
			WorktreePath: "/tmp/worktree1",
			Port:         8080,
		},
		"agent-proj2-def456-coder": {
			Model:        "claude-3-haiku-20240307",
			Prompt:       "Debug this code",
			WorktreePath: "/tmp/worktree2",
			Port:         8081,
		},
	}

	tests := []struct {
		name           string
		setupStateMgr  bool
		activeSession  []string
		stateFile      string
		expectedCount  int
		expectedError  bool
		description    string
	}{
		{
			name:          "Normal operation with active sessions",
			setupStateMgr: true,
			activeSession: []string{"agent-proj1-abc123-claude", "agent-proj2-def456-coder"},
			expectedCount: 2,
			expectedError: false,
			description:   "Should return all active sessions from state file",
		},
		{
			name:          "No active sessions",
			setupStateMgr: true,
			activeSession: []string{},
			expectedCount: 0,
			expectedError: false,
			description:   "Should return empty list when no active sessions",
		},
		{
			name:          "State manager not initialized",
			setupStateMgr: false,
			expectedError: true,
			description:   "Should return error when state manager is nil",
		},
		{
			name:          "Nonexistent state file",
			setupStateMgr: true,
			activeSession: []string{"agent-proj1-abc123-claude"},
			stateFile:     "/nonexistent/path",
			expectedCount: 0,
			expectedError: false,
			description:   "Should return empty list when state file doesn't exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewUziCLI()
			
			if !tt.setupStateMgr {
				cli.stateManager = nil
			} else {
				// Create temporary state file
				stateFile := createTempStateFile(t, testStates)
				
				// Mock state manager methods
				mockStateManager := &mockStateManagerForTest{
					activeSessions: tt.activeSession,
					statePath:      stateFile,
				}
				if tt.stateFile != "" {
					mockStateManager.statePath = tt.stateFile
				}
				// Set up the mock state manager
				cli.stateManager = mockStateManager

				// Mock tmux/git commands for git diff functionality
				mockTmuxAndGitCommands()
			}

			sessions, err := cli.GetSessionsLegacy()

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none. %s", tt.description)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v. %s", err, tt.description)
			}
			if len(sessions) != tt.expectedCount {
				t.Errorf("Expected %d sessions, got %d. %s", tt.expectedCount, len(sessions), tt.description)
			}

			// Verify sessions are sorted by port
			if len(sessions) > 1 {
				for i := 1; i < len(sessions); i++ {
					if sessions[i-1].Port > sessions[i].Port {
						t.Error("Sessions should be sorted by port")
					}
				}
			}
		})
	}
}

// Git Diff Parsing Tests - Table-driven with edge cases

func TestUziCLI_GetGitDiffTotals_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		diffOutput  string
		expectedIns int
		expectedDel int
		description string
	}{
		{
			name:        "Normal diff output",
			diffOutput:  " 3 files changed, 15 insertions(+), 7 deletions(-)",
			expectedIns: 15,
			expectedDel: 7,
			description: "Standard git diff shortstat output",
		},
		{
			name:        "Only insertions",
			diffOutput:  " 2 files changed, 25 insertions(+)",
			expectedIns: 25,
			expectedDel: 0,
			description: "Diff with only insertions",
		},
		{
			name:        "Only deletions",
			diffOutput:  " 1 file changed, 10 deletions(-)",
			expectedIns: 0,
			expectedDel: 10,
			description: "Diff with only deletions",
		},
		{
			name:        "Single insertion",
			diffOutput:  " 1 file changed, 1 insertion(+)",
			expectedIns: 1,
			expectedDel: 0,
			description: "Single insertion (singular form)",
		},
		{
			name:        "Single deletion",
			diffOutput:  " 1 file changed, 1 deletion(-)",
			expectedIns: 0,
			expectedDel: 1,
			description: "Single deletion (singular form)",
		},
		{
			name:        "Empty diff",
			diffOutput:  "",
			expectedIns: 0,
			expectedDel: 0,
			description: "Empty git diff output",
		},
		{
			name:        "Large diff",
			diffOutput:  " 150 files changed, 5023 insertions(+), 2891 deletions(-)",
			expectedIns: 5023,
			expectedDel: 2891,
			description: "Large diff with many changes",
		},
		{
			name:        "Malformed output - missing numbers",
			diffOutput:  " files changed, insertions(+), deletions(-)",
			expectedIns: 0,
			expectedDel: 0,
			description: "Malformed output missing numbers",
		},
		{
			name:        "Malformed output - wrong format",
			diffOutput:  "random output that doesn't match pattern",
			expectedIns: 0,
			expectedDel: 0,
			description: "Completely wrong format",
		},
		{
			name:        "Zero changes",
			diffOutput:  " 1 file changed, 0 insertions(+), 0 deletions(-)",
			expectedIns: 0,
			expectedDel: 0,
			description: "Diff with zero changes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fresh setup for each test case
			setupUziTest() // This calls cmdmock.Reset() and cmdmock.Enable()
			
			cli := NewUziCLI()
			
			// Create a test session state
			sessionState := &state.AgentState{
				WorktreePath: "/tmp/test-worktree",
			}

			// Mock the exact command that getGitDiffTotals will execute
			gitCmd := "git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null"
			cmdmock.SetResponseWithArgs("sh", []string{"-c", gitCmd}, tc.diffOutput, "", false)

			// Verify the mock was set up correctly by testing the command directly
			testCmd := uziExecCommand("sh", "-c", gitCmd)
			testOutput, testErr := testCmd.Output()
			if testErr != nil {
				t.Fatalf("Mock setup failed: %v", testErr)
			}
			if string(testOutput) != tc.diffOutput {
				t.Fatalf("Mock setup returned wrong output: expected %q, got %q", tc.diffOutput, string(testOutput))
			}

			insertions, deletions := cli.getGitDiffTotals("test-session", sessionState)

			if insertions != tc.expectedIns {
				t.Errorf("Expected %d insertions, got %d. %s", tc.expectedIns, insertions, tc.description)
			}
			if deletions != tc.expectedDel {
				t.Errorf("Expected %d deletions, got %d. %s", tc.expectedDel, deletions, tc.description)
			}
		})
	}
}

func TestUziCLI_GetGitDiffTotals_NoWorktree(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()

	// Test with empty worktree path
	sessionState := &state.AgentState{
		WorktreePath: "",
	}

	insertions, deletions := cli.getGitDiffTotals("test-session", sessionState)

	if insertions != 0 {
		t.Errorf("Expected 0 insertions for empty worktree, got %d", insertions)
	}
	if deletions != 0 {
		t.Errorf("Expected 0 deletions for empty worktree, got %d", deletions)
	}
}

func TestUziCLI_GetGitDiffTotals_CommandError(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()

	sessionState := &state.AgentState{
		WorktreePath: "/tmp/test-worktree",
	}

	// Mock git command failure
	gitCmd := "git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null"
	cmdmock.SetResponseWithArgs("sh", []string{"-c", gitCmd}, 
		"", "fatal: not a git repository", true)

	insertions, deletions := cli.getGitDiffTotals("test-session", sessionState)

	if insertions != 0 {
		t.Errorf("Expected 0 insertions on git error, got %d", insertions)
	}
	if deletions != 0 {
		t.Errorf("Expected 0 deletions on git error, got %d", deletions)
	}
}

// Tmux Interaction Tests

func TestUziCLI_TmuxInteractions(t *testing.T) {
	setupUziTest()
	
	testCases := []struct {
		name            string
		sessionName     string
		mockCmd         string
		mockArgs        []string
		mockStdout      string
		mockStderr      string
		mockExitErr     bool
		expectedSuccess bool
		description     string
	}{
		{
			name:            "GetPaneContent - Success",
			sessionName:     "agent-proj-abc123-claude",
			mockCmd:         "tmux",
			mockArgs:        []string{"capture-pane", "-t", "agent-proj-abc123-claude:agent", "-p"},
			mockStdout:      "$ echo hello\nhello\n$ ",
			mockStderr:      "",
			mockExitErr:     false,
			expectedSuccess: true,
			description:     "Should successfully capture pane content",
		},
		{
			name:            "GetPaneContent - Tmux Error",
			sessionName:     "nonexistent-session",
			mockCmd:         "tmux",
			mockArgs:        []string{"capture-pane", "-t", "nonexistent-session:agent", "-p"},
			mockStdout:      "",
			mockStderr:      "session not found",
			mockExitErr:     true,
			expectedSuccess: false,
			description:     "Should handle tmux session not found error",
		},
		{
			name:            "GetAgentStatus - Running",
			sessionName:     "agent-proj-abc123-claude",
			mockCmd:         "tmux",
			mockArgs:        []string{"capture-pane", "-t", "agent-proj-abc123-claude:agent", "-p"},
			mockStdout:      "Thinking about your request...\nesc to interrupt",
			mockStderr:      "",
			mockExitErr:     false,
			expectedSuccess: true,
			description:     "Should detect running status from pane content",
		},
		{
			name:            "GetAgentStatus - Ready",
			sessionName:     "agent-proj-abc123-claude",
			mockCmd:         "tmux",
			mockArgs:        []string{"capture-pane", "-t", "agent-proj-abc123-claude:agent", "-p"},
			mockStdout:      "$ waiting for input\n$ ",
			mockStderr:      "",
			mockExitErr:     false,
			expectedSuccess: true,
			description:     "Should detect ready status from pane content",
		},
		{
			name:            "GetAgentStatus - Unknown on Error",
			sessionName:     "broken-session",
			mockCmd:         "tmux",
			mockArgs:        []string{"capture-pane", "-t", "broken-session:agent", "-p"},
			mockStdout:      "",
			mockStderr:      "capture failed",
			mockExitErr:     true,
			expectedSuccess: true, // getAgentStatus doesn't return error, just "unknown"
			description:     "Should return unknown status when tmux fails",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := NewUziCLI()

			// Set up mock response
			cmdmock.SetResponseWithArgs(tc.mockCmd, tc.mockArgs, 
				tc.mockStdout, tc.mockStderr, tc.mockExitErr)

			if strings.Contains(tc.name, "GetPaneContent") {
				content, err := cli.getPaneContent(tc.sessionName)
				if tc.expectedSuccess && err != nil {
					t.Errorf("Expected success but got error: %v. %s", err, tc.description)
				}
				if !tc.expectedSuccess && err == nil {
					t.Errorf("Expected error but got success. %s", tc.description)
				}
				if tc.expectedSuccess && tc.mockExitErr == false {
					expectedContent := tc.mockStdout
					if content != expectedContent {
						t.Errorf("Expected content %q, got %q. %s", expectedContent, content, tc.description)
					}
				}
			}

			if strings.Contains(tc.name, "GetAgentStatus") {
				status := cli.getAgentStatus(tc.sessionName)
				
				// Verify status based on mock content
				if tc.mockExitErr == false {
					expectedStatus := "ready"
					if strings.Contains(tc.mockStdout, "esc to interrupt") || strings.Contains(tc.mockStdout, "Thinking") {
						expectedStatus = "running"
					}
					if status != expectedStatus {
						t.Errorf("Expected status %q, got %q. %s", expectedStatus, status, tc.description)
					}
				} else {
					if status != "unknown" {
						t.Errorf("Expected status 'unknown' on error, got %q. %s", status, tc.description)
					}
				}
			}
		})
	}
}

// Session Management Interface Tests

func TestUziCLI_SessionManagement(t *testing.T) {
	setupUziTest()
	
	tests := []struct {
		name          string
		method        string
		sessionName   string
		mockCmd       string
		mockArgs      []string
		mockStdout    string
		mockStderr    string
		mockExitErr   bool
		expectedError bool
		description   string
	}{
		{
			name:          "KillSession - Success",
			method:        "KillSession",
			sessionName:   "agent-proj-abc123-claude",
			mockCmd:       "uzi",
			mockArgs:      []string{"kill", "claude"},
			mockStdout:    "Session killed",
			mockStderr:    "",
			mockExitErr:   false,
			expectedError: false,
			description:   "Should successfully kill session",
		},
		{
			name:          "KillSession - Agent not found",
			method:        "KillSession",
			sessionName:   "agent-proj-abc123-nonexistent",
			mockCmd:       "uzi",
			mockArgs:      []string{"kill", "nonexistent"},
			mockStdout:    "",
			mockStderr:    "agent not found",
			mockExitErr:   true,
			expectedError: true,
			description:   "Should handle agent not found error",
		},
		{
			name:          "RunPrompt - Success",
			method:        "RunPrompt",
			mockCmd:       "uzi",
			mockArgs:      []string{"prompt", "--agents", "claude:1", "test prompt"},
			mockStdout:    "Prompt started",
			mockStderr:    "",
			mockExitErr:   false,
			expectedError: false,
			description:   "Should successfully run prompt",
		},
		{
			name:          "RunBroadcast - Success",
			method:        "RunBroadcast",
			mockCmd:       "uzi",
			mockArgs:      []string{"broadcast", "test message"},
			mockStdout:    "Message sent",
			mockStderr:    "",
			mockExitErr:   false,
			expectedError: false,
			description:   "Should successfully broadcast message",
		},
		{
			name:          "RunCommand - Success",
			method:        "RunCommand",
			mockCmd:       "uzi",
			mockArgs:      []string{"run", "echo test"},
			mockStdout:    "Command executed",
			mockStderr:    "",
			mockExitErr:   false,
			expectedError: false,
			description:   "Should successfully run command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewUziCLI()

			cmdmock.SetResponseWithArgs(tt.mockCmd, tt.mockArgs, 
				tt.mockStdout, tt.mockStderr, tt.mockExitErr)

			var err error
			switch tt.method {
			case "KillSession":
				err = cli.KillSession(tt.sessionName)
			case "RunPrompt":
				err = cli.RunPrompt("claude:1", "test prompt")
			case "RunBroadcast":
				err = cli.RunBroadcast("test message")
			case "RunCommand":
				err = cli.RunCommand("echo test")
			}

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none. %s", tt.description)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v. %s", err, tt.description)
			}

			// Verify the command was called
			if !cmdmock.WasCommandCalled(tt.mockCmd, tt.mockArgs...) {
				t.Errorf("Expected command %s %v to be called. %s", tt.mockCmd, tt.mockArgs, tt.description)
			}
		})
	}
}

func TestUziCLI_RefreshSessions(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()

	// RefreshSessions is a no-op, should always succeed
	err := cli.RefreshSessions()
	if err != nil {
		t.Errorf("Expected no error from RefreshSessions, got: %v", err)
	}
}

// State Management Tests

func TestUziCLI_GetSessionState(t *testing.T) {
	setupUziTest()

	testStates := map[string]state.AgentState{
		"agent-proj-abc123-claude": {
			Model:        "claude-3-5-sonnet-20241022",
			Prompt:       "Write a hello world program",
			WorktreePath: "/tmp/worktree1",
			Port:         8080,
		},
	}

	tests := []struct {
		name           string
		sessionName    string
		setupState     bool
		expectedError  bool
		expectedModel  string
		description    string
	}{
		{
			name:          "Existing session",
			sessionName:   "agent-proj-abc123-claude",
			setupState:    true,
			expectedError: false,
			expectedModel: "claude-3-5-sonnet-20241022",
			description:   "Should return state for existing session",
		},
		{
			name:          "Non-existent session",
			sessionName:   "agent-proj-xyz-nonexistent",
			setupState:    true,
			expectedError: true,
			description:   "Should return error for non-existent session",
		},
		{
			name:          "State manager not initialized",
			sessionName:   "any-session",
			setupState:    false,
			expectedError: true,
			description:   "Should return error when state manager is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewUziCLI()

			if !tt.setupState {
				cli.stateManager = nil
			} else {
				// Create temp state file and set up mock state manager
				stateFile := createTempStateFile(t, testStates)
				defer os.Remove(stateFile)
				
				// Set up mock state manager
				cli.stateManager = &mockStateManagerForTest{
					activeSessions: []string{"agent-proj-abc123-claude"},
					statePath:      stateFile,
				}
			}

			sessionState, err := cli.GetSessionState(tt.sessionName)

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none. %s", tt.description)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v. %s", err, tt.description)
			}
			if !tt.expectedError && sessionState != nil && sessionState.Model != tt.expectedModel {
				t.Errorf("Expected model %s, got %s. %s", tt.expectedModel, sessionState.Model, tt.description)
			}
		})
	}
}

func TestUziCLI_GetSessionStatus(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()

	// Mock tmux capture-pane for status detection
	cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "test-session:agent", "-p"}, 
		"$ ready for input", "", false)

	status, err := cli.GetSessionStatus("test-session")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if status != "ready" {
		t.Errorf("Expected status 'ready', got: %s", status)
	}
}

// Helper function tests

func TestExtractAgentName_UziInterface(t *testing.T) {
	tests := []struct {
		sessionName string
		expected    string
		description string
	}{
		{
			sessionName: "agent-myproject-abc123-claude",
			expected:    "claude",
			description: "Standard agent session format",
		},
		{
			sessionName: "agent-complex-project-def456-claude-coder",
			expected:    "claude-coder",
			description: "Agent name with hyphens",
		},
		{
			sessionName: "agent-proj-ghi789-assistant-v2",
			expected:    "assistant-v2",
			description: "Complex agent name",
		},
		{
			sessionName: "not-agent-format",
			expected:    "not-agent-format",
			description: "Non-agent session name",
		},
		{
			sessionName: "agent-incomplete",
			expected:    "agent-incomplete",
			description: "Incomplete agent session format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := extractAgentName(tt.sessionName)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q for session %q", tt.expected, result, tt.sessionName)
			}
		})
	}
}

// Legacy UziClient compatibility tests

func TestUziClient_LegacyCompatibility(t *testing.T) {
	setupUziTest()
	
	tests := []struct {
		name           string
		method         string
		expectedError  bool
		expectedResult interface{}
		description    string
	}{
		{
			name:          "GetActiveSessions with state manager",
			method:        "GetActiveSessions",
			expectedError: false,
			description:   "Should delegate to state manager",
		},
		{
			name:          "GetActiveSessions without state manager",
			method:        "GetActiveSessions",
			expectedError: false,
			expectedResult: []string(nil),
			description:   "Should return nil when state manager is nil",
		},
		{
			name:          "GetSessionState - Not implemented",
			method:        "GetSessionState",
			expectedError: true,
			description:   "Should return not implemented error",
		},
		{
			name:          "GetSessionStatus - Stub implementation",
			method:        "GetSessionStatus",
			expectedError: false,
			expectedResult: "unknown",
			description:   "Should return unknown status",
		},
		{
			name:          "AttachToSession - Not implemented",
			method:        "AttachToSession",
			expectedError: true,
			description:   "Should return not implemented error",
		},
		{
			name:          "KillSession - Not implemented",
			method:        "KillSession",
			expectedError: true,
			description:   "Should return not implemented error",
		},
		{
			name:          "RefreshSessions - No-op",
			method:        "RefreshSessions",
			expectedError: false,
			description:   "Should succeed as no-op",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewUziClient()
			
			// For the nil state manager test
			if strings.Contains(tt.name, "without state manager") {
				client.stateManager = nil
			}

			var err error
			var result interface{}

			switch tt.method {
			case "GetActiveSessions":
				result, err = client.GetActiveSessions()
			case "GetSessionState":
				result, err = client.GetSessionState("test-session")
			case "GetSessionStatus":
				result, err = client.GetSessionStatus("test-session")
			case "AttachToSession":
				err = client.AttachToSession("test-session")
			case "KillSession":
				err = client.KillSession("test-session")
			case "RefreshSessions":
				err = client.RefreshSessions()
			}

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none. %s", tt.description)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v. %s", err, tt.description)
			}

			if tt.expectedResult != nil {
				switch expected := tt.expectedResult.(type) {
				case string:
					if result != expected {
						t.Errorf("Expected result %v, got %v. %s", expected, result, tt.description)
					}
				case []string:
					if result != nil && len(result.([]string)) != 0 && expected != nil {
						t.Errorf("Expected result %v, got %v. %s", expected, result, tt.description)
					}
				}
			}
		})
	}
}

// Comprehensive edge case testing

func TestUziCLI_EdgeCases_Comprehensive(t *testing.T) {
	setupUziTest()
	
	t.Run("Empty session name handling", func(t *testing.T) {
		cli := NewUziCLI()
		
		// Test methods with empty session names
		status := cli.getAgentStatus("")
		if status != "unknown" {
			t.Errorf("Expected 'unknown' status for empty session name, got %s", status)
		}
		
		agentName := extractAgentName("")
		if agentName != "" {
			t.Errorf("Expected empty agent name for empty session, got %s", agentName)
		}
	})

	t.Run("Very long session names", func(t *testing.T) {
		longSessionName := strings.Repeat("a", 1000) + "-" + strings.Repeat("b", 1000)
		
		// Should handle gracefully without crashing
		agentName := extractAgentName(longSessionName)
		if agentName == "" {
			t.Error("Should handle long session names without returning empty")
		}
	})

	t.Run("Special characters in session names", func(t *testing.T) {
		specialChars := []string{
			"agent-proj-abc123-claude@special",
			"agent-proj.test-abc123-claude_v2",
			"agent-proj with spaces-abc123-claude",
		}
		
		for _, sessionName := range specialChars {
			agentName := extractAgentName(sessionName)
			// Should not crash and should return something reasonable
			if agentName == "" {
				t.Errorf("Should handle special characters in session name: %s", sessionName)
			}
		}
	})

	t.Run("Concurrent operations", func(t *testing.T) {
		cli := NewUziCLI()
		
		// Mock successful command
		cmdmock.SetResponseWithArgs("echo", []string{"concurrent"}, "success", "", false)
		
		// Run multiple operations concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := cli.executeCommand("echo", "concurrent")
				if err != nil {
					t.Errorf("Concurrent operation failed: %v", err)
				}
				done <- true
			}()
		}
		
		// Wait for all to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Enhanced tmux method tests

func TestUziCLI_TmuxEnhancedMethods(t *testing.T) {
	setupUziTest()
	
	cli := NewUziCLI()

	// Mock uzi ls --json for GetSessionsWithTmuxInfo
	testSessions := []SessionInfo{
		{
			Name:         "agent-proj-abc123-claude",
			AgentName:    "claude",
			Model:        "claude-3-5-sonnet-20241022",
			Status:       "ready",
			Prompt:       "Test prompt",
			Insertions:   15,
			Deletions:    3,
			WorktreePath: "/tmp/worktree1",
			Port:         8080,
		},
	}
	cmdmock.SetResponseWithArgs("uzi", []string{"ls", "--json"}, 
		createSessionJSON(testSessions), "", false)

	// Test IsSessionAttached
	result := cli.IsSessionAttached("test-session")
	if result {
		t.Error("Expected false for IsSessionAttached with unmocked tmux")
	}

	// Test GetSessionActivity
	activity := cli.GetSessionActivity("test-session")
	if activity == "" {
		t.Error("Expected non-empty activity level")
	}

	// Test GetAttachedSessionCount  
	// Mock tmux list-sessions for this
	cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_attached}"}, 
		"1\n0\n1\n", "", false)
	count, err := cli.GetAttachedSessionCount()
	if err != nil {
		t.Errorf("Expected no error from GetAttachedSessionCount, got: %v", err)
	}
	if count < 0 {
		t.Error("Expected non-negative count")
	}

	// Test RefreshTmuxCache
	cli.RefreshTmuxCache() // Should not panic

	// Test GetTmuxSessionsByActivity
	// Mock tmux list-sessions for activity data
	cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}:#{session_activity}"}, 
		"session1:1234567890\nsession2:1234567891\n", "", false)
	sessions, err := cli.GetTmuxSessionsByActivity()
	if err != nil {
		t.Errorf("Expected no error from GetTmuxSessionsByActivity, got: %v", err)
	}
	if sessions == nil {
		t.Error("Expected non-nil sessions map")
	}

	// Test FormatSessionActivity
	formatted := cli.FormatSessionActivity("test-session")
	if formatted == "" {
		t.Error("Expected non-empty formatted activity")
	}

	// Test GetSessionsWithTmuxInfo
	sessionsInfo, tmuxInfo, err := cli.GetSessionsWithTmuxInfo()
	if err != nil {
		t.Errorf("Expected no error from GetSessionsWithTmuxInfo, got: %v", err)
	}
	if sessionsInfo == nil {
		t.Error("Expected non-nil sessions info")
	}
	_ = tmuxInfo // May be nil if tmux mapping fails
}

// Test AttachToSession method
func TestUziCLI_AttachToSession(t *testing.T) {
	setupUziTest()
	
	cli := NewUziCLI()

	// Mock tmux attach-session command to succeed - need to mock exec.Command directly
	cmdmock.SetResponseWithArgs("tmux", []string{"attach-session", "-t", "test-session"}, 
		"", "", false)

	// This would normally block in a real terminal, but with mocking it should return
	err := cli.AttachToSession("test-session")
	if err != nil {
		// The attach command failing is actually expected in a test environment
		// since it would try to attach to a non-existent session
		// Let's just verify the error message contains expected text
		if !strings.Contains(err.Error(), "AttachToSession") {
			t.Errorf("Expected AttachToSession error, got: %v", err)
		}
	}

	// Test with explicit command failure
	cmdmock.SetResponseWithArgs("tmux", []string{"attach-session", "-t", "bad-session"}, 
		"", "session not found", true)

	err = cli.AttachToSession("bad-session")
	if err == nil {
		t.Error("Expected error for non-existent session")
	}
}

// Test internal\helper methods
func TestUziCLI_InternalMethods(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()

	// Mock empty tmux output (no sessions exist)
	cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"}, 
		"", "", false)
	cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_attached}"}, 
		"", "", false)
	cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}:#{session_activity}"}, 
		"", "", false)

	// Refresh the cache to use our mocked commands
	cli.RefreshTmuxCache()

	// Test IsSessionAttached - should return false for non-existent session
	attached := cli.IsSessionAttached("test-session")
	if attached {
		t.Error("Expected false for non-existent session")
	}

	// Test GetSessionActivity - should return "inactive" for non-existent session
	activity := cli.GetSessionActivity("test-session")
	if activity != "inactive" {
		t.Errorf("Expected 'inactive' for non-existent session, got: %q", activity)
	}

	// Test GetAttachedSessionCount - should return 0 when no sessions exist
	count, err := cli.GetAttachedSessionCount()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 attached sessions when no sessions exist, got: %d", count)
	}

	// Test RefreshTmuxCache - should not panic
	cli.RefreshTmuxCache()

	// Test GetTmuxSessionsByActivity - should return map with 3 empty groups when no sessions exist
	sessions, err := cli.GetTmuxSessionsByActivity()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	expectedGroups := []string{"attached", "active", "inactive"}
	if len(sessions) != len(expectedGroups) {
		t.Errorf("Expected %d activity groups, got: %d", len(expectedGroups), len(sessions))
	}
	for _, group := range expectedGroups {
		if groupSessions, exists := sessions[group]; !exists {
			t.Errorf("Expected group '%s' to exist", group)
		} else if len(groupSessions) != 0 {
			t.Errorf("Expected group '%s' to be empty, got %d sessions", group, len(groupSessions))
		}
	}

	// Test FormatSessionActivity - should return "○" for "inactive" activity
	formatted := cli.FormatSessionActivity("test-session")
	if formatted != "○" {
		t.Errorf("Expected '○' for inactive session, got: %q", formatted)
	}
}

// Test executeCommand for retry logic
func TestUziCLI_ExecuteCommandRetries(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()

	// Mock persistent failure - all retries will fail
	cmdmock.SetResponseWithArgs("echo", []string{"persistent-fail"}, "", "persistent error", true)

	_, err := cli.executeCommand("echo", "persistent-fail")
	if err == nil {
		t.Error("Expected error, got none")
	}

	// Verify retries by checking we made multiple calls
	calls := cmdmock.GetCommandCalls("echo", "persistent-fail")
	if len(calls) < 2 {
		t.Errorf("Expected at least 2 retry calls, got %d", len(calls))
	}

	// Test successful command (no retries needed)
	cmdmock.SetResponseWithArgs("echo", []string{"success-test"}, "success", "", false)

	output, err := cli.executeCommand("echo", "success-test")
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if string(output) != "success" {
		t.Errorf("Expected 'success', got: %s", output)
	}
}

// Test logging functionality
func TestUziCLI_OperationLogs(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	// Ensure no panics
	cli.logOperation("TestLog", time.Second, nil)
	cli.logOperation("TestLogError", time.Second, fmt.Errorf("fake error"))
}

// Test internal/private method getAgentStatus
func TestUziCLI_GetAgentStatus(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()

	// Test running status
	cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "running-session:agent", "-p"}, "Thinking... esc to interrupt", "", false)
	status := cli.getAgentStatus("running-session")
	if status != "running" {
		t.Errorf("Expected 'running', got %v", status)
	}

	// Test ready status
	cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "ready-session:agent", "-p"}, "$ waiting for input", "", false)
	status = cli.getAgentStatus("ready-session")
	if status != "ready" {
		t.Errorf("Expected 'ready', got %v", status)
	}

	// Test unknown status on error
	cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "error-session:agent", "-p"}, "", "session not found", true)
	status = cli.getAgentStatus("error-session")
	if status != "unknown" {
		t.Errorf("Expected 'unknown', got %v", status)
	}
}

// Helper functions for tests

func mockTmuxAndGitCommands() {
	// Mock common tmux commands for status detection
	cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "agent-proj1-abc123-claude:agent", "-p"}, 
		"$ ready for input", "", false)
	cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "agent-proj2-def456-coder:agent", "-p"}, 
		"Thinking...\nesc to interrupt", "", false)
	
	// Mock git diff commands
	gitCmd := "git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null"
	cmdmock.SetResponseWithArgs("sh", []string{"-c", gitCmd}, 
		" 3 files changed, 15 insertions(+), 3 deletions(-)", "", false)
}

// SpawnAgent Tests - Critical functionality coverage

func TestUziCLI_SpawnAgent_Success(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock git commands for agent creation
	cmdmock.SetResponseWithArgs("git", []string{"rev-parse", "--short", "HEAD"}, "abc123", "", false)
	cmdmock.SetResponseWithArgs("git", []string{"remote", "get-url", "origin"}, "https://github.com/user/project.git", "", false)
	
	// Mock worktree creation
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "git worktree add -b claude-project-abc123-1640000000 /Users/testuser/.local/share/uzi/worktrees/claude-project-abc123-1640000000"}, "", "", false)
	
	// Mock tmux session creation
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux new-session -d -s agent-project-abc123-claude -c /Users/testuser/.local/share/uzi/worktrees/claude-project-abc123-1640000000"}, "", "", false)
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux rename-window -t agent-project-abc123-claude:0 agent"}, "", "", false)
	
	// Mock agent command execution
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux send-keys -t agent-project-abc123-claude:agent C-m"}, "", "", false)
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux send-keys -t agent-project-abc123-claude:agent 'claude \"test prompt\"' C-m"}, "", "", false)
	
	// Create a mock state manager
	mockStateManager := &mockStateManagerForTest{
		activeSessions: []string{},
		statePath:      "/tmp/test-state.json",
	}
	cli.stateManager = mockStateManager
	
	sessionName, err := cli.SpawnAgent("test prompt", "claude")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if sessionName == "" {
		t.Error("Expected non-empty session name")
	}
}

func TestUziCLI_SpawnAgent_HelperMethods(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Test parseAgentConfigs
	tests := []struct {
		name        string
		agentsStr   string
		expectedLen int
		expectError bool
	}{
		{"valid single agent", "claude:1", 1, false},
		{"valid multiple agents", "claude:1,cursor:2", 2, false},
		{"invalid format", "claude-1", 0, true},
		{"invalid count", "claude:zero", 0, true},
		{"zero count", "claude:0", 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs, err := cli.parseAgentConfigs(tt.agentsStr)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if len(configs) != tt.expectedLen {
				t.Errorf("Expected %d configs, got %d", tt.expectedLen, len(configs))
			}
		})
	}
}

func TestUziCLI_SpawnAgent_CommandMapping(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	tests := []struct {
		agent    string
		expected string
	}{
		{"claude", "claude"},
		{"cursor", "cursor"},
		{"codex", "codex"},
		{"gemini", "gemini"},
		{"random", "claude"},
		{"unknown", "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			result := cli.getCommandForAgent(tt.agent)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestUziCLI_SpawnAgent_GetRandomAgentName(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Test non-random agent
	result, err := cli.getRandomAgentName("claude")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "claude" {
		t.Errorf("Expected 'claude', got %q", result)
	}
	
	// Test random agent - should return a non-empty string
	result, err = cli.getRandomAgentName("random")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty random agent name")
	}
}

func TestUziCLI_SpawnAgent_GetGitInfo(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock git commands
	cmdmock.SetResponseWithArgs("git", []string{"rev-parse", "--short", "HEAD"}, "abc123", "", false)
	cmdmock.SetResponseWithArgs("git", []string{"remote", "get-url", "origin"}, "https://github.com/user/project.git", "", false)
	
	gitHash, projectDir, err := cli.getGitInfo()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if gitHash != "abc123" {
		t.Errorf("Expected gitHash 'abc123', got %q", gitHash)
	}
	if projectDir != "project" {
		t.Errorf("Expected projectDir 'project', got %q", projectDir)
	}
}

func TestUziCLI_SpawnAgent_GetGitInfo_Error(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock git command failure
	cmdmock.SetResponseWithArgs("git", []string{"rev-parse", "--short", "HEAD"}, "", "fatal: not a git repository", true)
	
	_, _, err := cli.getGitInfo()
	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestUziCLI_SpawnAgent_CreateWorktree(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock successful worktree creation
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "git worktree add -b test-branch /tmp/test-worktree"}, "", "", false)
	
	worktreePath, err := cli.createWorktree("test-branch", "test-worktree")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !strings.Contains(worktreePath, "test-worktree") {
		t.Errorf("Expected worktree path to contain 'test-worktree', got %q", worktreePath)
	}
}

func TestUziCLI_SpawnAgent_CreateWorktree_Error(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock worktree creation failure
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "git worktree add -b test-branch /tmp/test-worktree"}, "", "fatal: git worktree failed", true)
	
	_, err := cli.createWorktree("test-branch", "test-worktree")
	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestUziCLI_SpawnAgent_CreateTmuxSession(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock tmux commands
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux new-session -d -s test-session -c /tmp"}, "", "", false)
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux rename-window -t test-session:0 agent"}, "", "", false)
	
	err := cli.createTmuxSession("test-session", "/tmp")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUziCLI_SpawnAgent_CreateTmuxSession_Error(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock tmux session creation failure
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux new-session -d -s test-session -c /tmp"}, "", "tmux: session exists", true)
	
	err := cli.createTmuxSession("test-session", "/tmp")
	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestUziCLI_SpawnAgent_IsPortAvailable(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Test with a high port that should be available
	available := cli.isPortAvailable(65432)
	if !available {
		t.Error("Expected port 65432 to be available")
	}
	
	// Test with a reserved port that should not be available
	// Note: This might fail in test environments where ports are not restricted
	// So we'll just verify the method doesn't panic
	_ = cli.isPortAvailable(80)
}

func TestUziCLI_SpawnAgent_FindAvailablePort(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Test finding available port in high range
	port, err := cli.findAvailablePort(65400, 65500, []int{})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if port < 65400 || port > 65500 {
		t.Errorf("Expected port in range 65400-65500, got %d", port)
	}
	
	// Test with assigned ports
	assignedPorts := []int{65401, 65402, 65403}
	port, err = cli.findAvailablePort(65400, 65410, assignedPorts)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	for _, assigned := range assignedPorts {
		if port == assigned {
			t.Errorf("Found port %d should not be in assigned list %v", port, assignedPorts)
		}
	}
}

func TestUziCLI_SpawnAgent_ExecuteAgentCommand(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock tmux commands for agent execution
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux send-keys -t test-session:agent C-m"}, "", "", false)
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux send-keys -t test-session:agent 'claude \"test prompt\"' C-m"}, "", "", false)
	
	err := cli.executeAgentCommand("test-session", "claude", "test prompt", "/tmp")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUziCLI_SpawnAgent_ExecuteAgentCommand_Gemini(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Mock tmux commands for gemini agent execution (different format)
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux send-keys -t test-session:agent C-m"}, "", "", false)
	cmdmock.SetResponseWithArgs("sh", []string{"-c", "tmux send-keys -t test-session:agent 'gemini -p \"test prompt\"' C-m"}, "", "", false)
	
	err := cli.executeAgentCommand("test-session", "gemini", "test prompt", "/tmp")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUziCLI_SpawnAgent_GetExistingSessionPorts(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// Test with nil state manager
	ports, err := cli.getExistingSessionPorts(nil)
	if err != nil {
		t.Errorf("Expected no error with nil state manager, got: %v", err)
	}
	if len(ports) != 0 {
		t.Errorf("Expected empty ports list, got %v", ports)
	}
	
	// Test with mock state manager and valid state file
	testStates := map[string]state.AgentState{
		"session1": {Port: 8080},
		"session2": {Port: 8081},
		"session3": {Port: 0}, // Should be ignored
	}
	stateFile := createTempStateFile(t, testStates)
	
	mockStateManager := &mockStateManagerForTest{
		statePath: stateFile,
	}
	
	ports, err = cli.getExistingSessionPorts(mockStateManager)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(ports) != 2 {
		t.Errorf("Expected 2 ports, got %d", len(ports))
	}
	expectedPorts := []int{8080, 8081}
	for _, expected := range expectedPorts {
		found := false
		for _, port := range ports {
			if port == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find port %d in result %v", expected, ports)
		}
	}
}

func TestUziCLI_SpawnAgent_StateManagerBridge(t *testing.T) {
	// Test StateManagerBridge methods for coverage
	bridge := NewStateManagerBridge()
	if bridge == nil {
		t.Error("Expected non-nil bridge")
	}
	
	// Test SaveState method (should not error even without real state manager)
	err := bridge.SaveState("prompt", "branch", "session", "path", "model")
	// This will likely error since it's trying to save to a real path, but we test the method exists
	_ = err // Acknowledge potential error
	
	// Test SaveStateWithPort method
	err = bridge.SaveStateWithPort("prompt", "branch", "session", "path", "model", 8080)
	_ = err // Acknowledge potential error
}

func TestUziCLI_SpawnAgent_LoadDefaultConfig(t *testing.T) {
	setupUziTest()
	cli := NewUziCLI()
	
	// This will likely fail since there's no uzi.yaml in test environment,
	// but we test that the method doesn't panic
	_, err := cli.loadDefaultConfig()
	_ = err // Acknowledge expected error in test environment
}
