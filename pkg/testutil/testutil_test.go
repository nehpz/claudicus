// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package testutil

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewFakeCommandRunner(t *testing.T) {
	runner := NewFakeCommandRunner()

	if runner == nil {
		t.Error("Expected NewFakeCommandRunner to return non-nil runner")
	}

	if runner.commands == nil {
		t.Error("Expected commands slice to be initialized")
	}

	if runner.responses == nil {
		t.Error("Expected responses map to be initialized")
	}

	if runner.callCount != 0 {
		t.Errorf("Expected initial call count to be 0, got %d", runner.callCount)
	}
}

func TestSetResponse(t *testing.T) {
	runner := NewFakeCommandRunner()
	output := []byte("test output")
	err := errors.New("test error")

	runner.SetResponse("git", []string{"status"}, output, err)

	// Test that the response was stored
	key := runner.makeKey("git", []string{"status"})
	response, exists := runner.responses[key]

	if !exists {
		t.Error("Expected response to be stored")
	}

	if string(response.Output) != "test output" {
		t.Errorf("Expected output 'test output', got %s", string(response.Output))
	}

	if response.Error.Error() != "test error" {
		t.Errorf("Expected error 'test error', got %v", response.Error)
	}
}

func TestSetJSONResponse(t *testing.T) {
	runner := NewFakeCommandRunner()
	jsonOutput := `{"name": "test", "value": 123}`
	err := errors.New("json error")

	runner.SetJSONResponse("uzi", []string{"ls", "--json"}, jsonOutput, err)

	// Test that the JSON response was stored as bytes
	key := runner.makeKey("uzi", []string{"ls", "--json"})
	response, exists := runner.responses[key]

	if !exists {
		t.Error("Expected JSON response to be stored")
	}

	if string(response.Output) != jsonOutput {
		t.Errorf("Expected JSON output %s, got %s", jsonOutput, string(response.Output))
	}

	if response.Error.Error() != "json error" {
		t.Errorf("Expected error 'json error', got %v", response.Error)
	}
}

func TestRun(t *testing.T) {
	runner := NewFakeCommandRunner()
	expectedOutput := []byte("git status output")

	runner.SetResponse("git", []string{"status"}, expectedOutput, nil)

	output, err := runner.Run("git", "status")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if string(output) != "git status output" {
		t.Errorf("Expected output 'git status output', got %s", string(output))
	}

	if runner.callCount != 1 {
		t.Errorf("Expected call count to be 1, got %d", runner.callCount)
	}
}

func TestRunWithTimeout(t *testing.T) {
	runner := NewFakeCommandRunner()
	expectedOutput := []byte("command output")
	timeout := 5 * time.Second

	runner.SetResponse("echo", []string{"hello"}, expectedOutput, nil)

	output, err := runner.RunWithTimeout(timeout, "echo", "hello")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if string(output) != "command output" {
		t.Errorf("Expected output 'command output', got %s", string(output))
	}

	// Check that the call was recorded with timeout
	calls := runner.GetCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, calls[0].Timeout)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	runner := NewFakeCommandRunner()

	output, err := runner.Run("unknown", "command")

	if err == nil {
		t.Error("Expected error for unknown command")
	}

	if !strings.Contains(err.Error(), "command not found") {
		t.Errorf("Expected 'command not found' error, got %v", err)
	}

	if output != nil {
		t.Errorf("Expected nil output for unknown command, got %v", output)
	}
}

func TestGetCalls(t *testing.T) {
	runner := NewFakeCommandRunner()

	runner.SetResponse("git", []string{"status"}, []byte("output1"), nil)
	runner.SetResponse("git", []string{"log"}, []byte("output2"), nil)

	runner.Run("git", "status")
	runner.Run("git", "log")

	calls := runner.GetCalls()

	if len(calls) != 2 {
		t.Errorf("Expected 2 calls, got %d", len(calls))
	}

	if calls[0].Name != "git" || len(calls[0].Args) != 1 || calls[0].Args[0] != "status" {
		t.Errorf("Expected first call to be 'git status', got %v %v", calls[0].Name, calls[0].Args)
	}

	if calls[1].Name != "git" || len(calls[1].Args) != 1 || calls[1].Args[0] != "log" {
		t.Errorf("Expected second call to be 'git log', got %v %v", calls[1].Name, calls[1].Args)
	}
}

func TestGetCallCount(t *testing.T) {
	runner := NewFakeCommandRunner()

	if runner.GetCallCount() != 0 {
		t.Errorf("Expected initial call count to be 0, got %d", runner.GetCallCount())
	}

	runner.SetResponse("echo", []string{"test"}, []byte("test"), nil)

	runner.Run("echo", "test")
	runner.Run("echo", "test")
	runner.Run("echo", "test")

	if runner.GetCallCount() != 3 {
		t.Errorf("Expected call count to be 3, got %d", runner.GetCallCount())
	}
}

func TestReset(t *testing.T) {
	runner := NewFakeCommandRunner()

	runner.SetResponse("git", []string{"status"}, []byte("output"), nil)
	runner.Run("git", "status")

	if runner.GetCallCount() != 1 {
		t.Error("Expected call count to be 1 before reset")
	}

	runner.Reset()

	if runner.GetCallCount() != 0 {
		t.Errorf("Expected call count to be 0 after reset, got %d", runner.GetCallCount())
	}

	if len(runner.GetCalls()) != 0 {
		t.Errorf("Expected 0 calls after reset, got %d", len(runner.GetCalls()))
	}

	// Test that responses are cleared
	_, err := runner.Run("git", "status")
	if err == nil {
		t.Error("Expected error for command after reset")
	}
}

func TestWasCommandCalled(t *testing.T) {
	runner := NewFakeCommandRunner()

	runner.SetResponse("git", []string{"status"}, []byte("output"), nil)
	runner.SetResponse("git", []string{"log", "--oneline"}, []byte("log output"), nil)

	runner.Run("git", "status")
	runner.Run("git", "log", "--oneline")

	if !runner.WasCommandCalled("git", "status") {
		t.Error("Expected 'git status' to be called")
	}

	if !runner.WasCommandCalled("git", "log", "--oneline") {
		t.Error("Expected 'git log --oneline' to be called")
	}

	if runner.WasCommandCalled("git", "push") {
		t.Error("Expected 'git push' to not be called")
	}

	if runner.WasCommandCalled("unknown") {
		t.Error("Expected 'unknown' to not be called")
	}
}

func TestMakeKey(t *testing.T) {
	runner := NewFakeCommandRunner()

	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{"git", []string{"status"}, "git status"},
		{"git", []string{"log", "--oneline"}, "git log --oneline"},
		{"echo", []string{}, "echo "},
		{"command", []string{"arg1", "arg2", "arg3"}, "command arg1 arg2 arg3"},
	}

	for _, tc := range testCases {
		result := runner.makeKey(tc.name, tc.args)
		if result != tc.expected {
			t.Errorf("Expected key %q, got %q", tc.expected, result)
		}
	}
}

func TestNewFakeTimeProvider(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	provider := NewFakeTimeProvider(testTime)

	if provider == nil {
		t.Error("Expected NewFakeTimeProvider to return non-nil provider")
	}

	if !provider.Now().Equal(testTime) {
		t.Errorf("Expected Now() to return %v, got %v", testTime, provider.Now())
	}
}

func TestFakeTimeProviderNow(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	provider := NewFakeTimeProvider(testTime)

	now := provider.Now()
	if !now.Equal(testTime) {
		t.Errorf("Expected Now() to return %v, got %v", testTime, now)
	}
}

func TestFakeTimeProviderSince(t *testing.T) {
	currentTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	pastTime := currentTime.Add(-2 * time.Hour)
	provider := NewFakeTimeProvider(currentTime)

	duration := provider.Since(pastTime)
	expected := 2 * time.Hour

	if duration != expected {
		t.Errorf("Expected Since() to return %v, got %v", expected, duration)
	}
}

func TestFakeTimeProviderSetTime(t *testing.T) {
	initialTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	newTime := time.Date(2023, 1, 2, 15, 30, 0, 0, time.UTC)
	provider := NewFakeTimeProvider(initialTime)

	provider.SetTime(newTime)

	if !provider.Now().Equal(newTime) {
		t.Errorf("Expected Now() to return %v after SetTime(), got %v", newTime, provider.Now())
	}
}

func TestFakeTimeProviderAdvanceTime(t *testing.T) {
	initialTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	provider := NewFakeTimeProvider(initialTime)

	provider.AdvanceTime(2 * time.Hour)

	expectedTime := initialTime.Add(2 * time.Hour)
	if !provider.Now().Equal(expectedTime) {
		t.Errorf("Expected Now() to return %v after AdvanceTime(), got %v", expectedTime, provider.Now())
	}
}

func TestMakeFakeTmuxListOutput(t *testing.T) {
	sessions := []string{"session1", "session2", "session3"}
	output := MakeFakeTmuxListOutput(sessions)

	if output == "" {
		t.Error("Expected non-empty output")
	}

	for _, session := range sessions {
		if !strings.Contains(output, session) {
			t.Errorf("Expected output to contain session %s", session)
		}
	}

	// Check that it contains expected tmux output format
	if !strings.Contains(output, "windows") {
		t.Error("Expected output to contain 'windows'")
	}

	if !strings.Contains(output, "created") {
		t.Error("Expected output to contain 'created'")
	}
}

func TestMakeFakeUziLsJSON(t *testing.T) {
	// Test empty sessions
	emptyOutput := MakeFakeUziLsJSON([]SessionInfo{})
	if emptyOutput != "[]" {
		t.Errorf("Expected empty array for no sessions, got %s", emptyOutput)
	}

	// Test with sessions
	sessions := []SessionInfo{
		{
			Name:         "session1",
			AgentName:    "agent1",
			Model:        "claude",
			Status:       "ready",
			Prompt:       "test prompt",
			Insertions:   5,
			Deletions:    2,
			WorktreePath: "/test/path1",
			Port:         3000,
		},
		{
			Name:         "session2",
			AgentName:    "agent2",
			Model:        "gpt-4",
			Status:       "running",
			Prompt:       "another prompt",
			Insertions:   10,
			Deletions:    0,
			WorktreePath: "/test/path2",
			Port:         3001,
		},
	}

	output := MakeFakeUziLsJSON(sessions)

	// Check that it's valid JSON-like structure
	if !strings.HasPrefix(output, "[") || !strings.HasSuffix(output, "]") {
		t.Error("Expected output to be a JSON array")
	}

	// Check that it contains session data
	for _, session := range sessions {
		if !strings.Contains(output, session.Name) {
			t.Errorf("Expected output to contain session name %s", session.Name)
		}
		if !strings.Contains(output, session.AgentName) {
			t.Errorf("Expected output to contain agent name %s", session.AgentName)
		}
		if !strings.Contains(output, session.Model) {
			t.Errorf("Expected output to contain model %s", session.Model)
		}
		if !strings.Contains(output, session.Status) {
			t.Errorf("Expected output to contain status %s", session.Status)
		}
	}
}

func TestSessionInfo(t *testing.T) {
	session := SessionInfo{
		Name:         "test-session",
		AgentName:    "test-agent",
		Model:        "claude",
		Status:       "ready",
		Prompt:       "test prompt",
		Insertions:   15,
		Deletions:    7,
		WorktreePath: "/test/worktree",
		Port:         3000,
	}

	if session.Name != "test-session" {
		t.Errorf("Expected name 'test-session', got %s", session.Name)
	}
	if session.AgentName != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got %s", session.AgentName)
	}
	if session.Model != "claude" {
		t.Errorf("Expected model 'claude', got %s", session.Model)
	}
	if session.Status != "ready" {
		t.Errorf("Expected status 'ready', got %s", session.Status)
	}
	if session.Prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got %s", session.Prompt)
	}
	if session.Insertions != 15 {
		t.Errorf("Expected insertions 15, got %d", session.Insertions)
	}
	if session.Deletions != 7 {
		t.Errorf("Expected deletions 7, got %d", session.Deletions)
	}
	if session.WorktreePath != "/test/worktree" {
		t.Errorf("Expected worktree path '/test/worktree', got %s", session.WorktreePath)
	}
	if session.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", session.Port)
	}
}

func TestMust(t *testing.T) {
	// Test successful case
	result := Must("success", nil)
	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}

	// Test panic case
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Must to panic with error")
		}
	}()

	Must("fail", errors.New("test error"))
}

func TestNewRequire(t *testing.T) {
	mockT := &mockTestingT{}
	require := NewRequire(mockT)

	if require == nil {
		t.Error("Expected NewRequire to return non-nil Require")
	}

	if require.t != mockT {
		t.Error("Expected Require to hold reference to testing.T")
	}
}

// Mock testing.T for testing Require functions
type mockTestingT struct {
	errors  []string
	failed  bool
	helpers int
}

func (m *mockTestingT) Helper() {
	m.helpers++
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, "error occurred")
}

func (m *mockTestingT) FailNow() {
	m.failed = true
}

func TestRequireNoError(t *testing.T) {
	mockT := &mockTestingT{}
	require := NewRequire(mockT)

	// Test successful case
	require.NoError(nil)

	if mockT.failed {
		t.Error("Expected NoError not to fail with nil error")
	}

	// Test failure case
	require.NoError(errors.New("test error"))

	if !mockT.failed {
		t.Error("Expected NoError to fail with non-nil error")
	}
}

func TestRequireError(t *testing.T) {
	mockT := &mockTestingT{}
	require := NewRequire(mockT)

	// Test successful case
	require.Error(errors.New("test error"))

	if mockT.failed {
		t.Error("Expected Error not to fail with non-nil error")
	}

	// Test failure case
	mockT2 := &mockTestingT{}
	require2 := NewRequire(mockT2)
	require2.Error(nil)

	if !mockT2.failed {
		t.Error("Expected Error to fail with nil error")
	}
}

func TestRequireEqual(t *testing.T) {
	mockT := &mockTestingT{}
	require := NewRequire(mockT)

	// Test successful case
	require.Equal("test", "test")

	if mockT.failed {
		t.Error("Expected Equal not to fail with equal values")
	}

	// Test failure case
	mockT2 := &mockTestingT{}
	require2 := NewRequire(mockT2)
	require2.Equal("expected", "actual")

	if !mockT2.failed {
		t.Error("Expected Equal to fail with different values")
	}
}

func TestRequireNotNil(t *testing.T) {
	mockT := &mockTestingT{}
	require := NewRequire(mockT)

	// Test successful case
	require.NotNil("test")

	if mockT.failed {
		t.Error("Expected NotNil not to fail with non-nil value")
	}

	// Test failure case
	mockT2 := &mockTestingT{}
	require2 := NewRequire(mockT2)
	require2.NotNil(nil)

	if !mockT2.failed {
		t.Error("Expected NotNil to fail with nil value")
	}
}

func TestRequireNil(t *testing.T) {
	mockT := &mockTestingT{}
	require := NewRequire(mockT)

	// Test successful case
	require.Nil(nil)

	if mockT.failed {
		t.Error("Expected Nil not to fail with nil value")
	}

	// Test failure case
	mockT2 := &mockTestingT{}
	require2 := NewRequire(mockT2)
	require2.Nil("not nil")

	if !mockT2.failed {
		t.Error("Expected Nil to fail with non-nil value")
	}
}

func TestRequireTrue(t *testing.T) {
	mockT := &mockTestingT{}
	require := NewRequire(mockT)

	// Test successful case
	require.True(true)

	if mockT.failed {
		t.Error("Expected True not to fail with true value")
	}

	// Test failure case
	mockT2 := &mockTestingT{}
	require2 := NewRequire(mockT2)
	require2.True(false)

	if !mockT2.failed {
		t.Error("Expected True to fail with false value")
	}
}

func TestRequireFalse(t *testing.T) {
	mockT := &mockTestingT{}
	require := NewRequire(mockT)

	// Test successful case
	require.False(false)

	if mockT.failed {
		t.Error("Expected False not to fail with false value")
	}

	// Test failure case
	mockT2 := &mockTestingT{}
	require2 := NewRequire(mockT2)
	require2.False(true)

	if !mockT2.failed {
		t.Error("Expected False to fail with true value")
	}
}
