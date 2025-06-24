// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package testutil

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// CommandRunner abstracts command execution for testing
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
	RunWithTimeout(timeout time.Duration, name string, args ...string) ([]byte, error)
}

// FakeCommandRunner implements CommandRunner for testing
type FakeCommandRunner struct {
	commands []CommandCall
	responses map[string]CommandResponse
	callCount int
}

// CommandCall represents a command that was executed
type CommandCall struct {
	Name    string
	Args    []string
	Timeout time.Duration
}

// CommandResponse represents the response for a command
type CommandResponse struct {
	Output []byte
	Error  error
}

// NewFakeCommandRunner creates a new fake command runner
func NewFakeCommandRunner() *FakeCommandRunner {
	return &FakeCommandRunner{
		commands:  make([]CommandCall, 0),
		responses: make(map[string]CommandResponse),
	}
}

// SetResponse sets the expected response for a command
func (f *FakeCommandRunner) SetResponse(name string, args []string, output []byte, err error) {
	key := f.makeKey(name, args)
	f.responses[key] = CommandResponse{Output: output, Error: err}
}

// SetJSONResponse sets a JSON response for a command
func (f *FakeCommandRunner) SetJSONResponse(name string, args []string, jsonOutput string, err error) {
	f.SetResponse(name, args, []byte(jsonOutput), err)
}

// Run executes a command and returns the pre-programmed response
func (f *FakeCommandRunner) Run(name string, args ...string) ([]byte, error) {
	return f.RunWithTimeout(30*time.Second, name, args...)
}

// RunWithTimeout executes a command with timeout and returns the pre-programmed response
func (f *FakeCommandRunner) RunWithTimeout(timeout time.Duration, name string, args ...string) ([]byte, error) {
	call := CommandCall{
		Name:    name,
		Args:    args,
		Timeout: timeout,
	}
	f.commands = append(f.commands, call)
	f.callCount++

	key := f.makeKey(name, args)
	if response, exists := f.responses[key]; exists {
		return response.Output, response.Error
	}

	// Default response for unknown commands
	return nil, fmt.Errorf("command not found: %s %v", name, args)
}

// GetCalls returns all recorded command calls
func (f *FakeCommandRunner) GetCalls() []CommandCall {
	return f.commands
}

// GetCallCount returns the number of commands executed
func (f *FakeCommandRunner) GetCallCount() int {
	return f.callCount
}

// Reset clears all recorded calls and responses
func (f *FakeCommandRunner) Reset() {
	f.commands = make([]CommandCall, 0)
	f.responses = make(map[string]CommandResponse)
	f.callCount = 0
}

// WasCommandCalled checks if a specific command was called
func (f *FakeCommandRunner) WasCommandCalled(name string, args ...string) bool {
	key := f.makeKey(name, args)
	for _, call := range f.commands {
		if f.makeKey(call.Name, call.Args) == key {
			return true
		}
	}
	return false
}

// makeKey creates a unique key for a command and its arguments
func (f *FakeCommandRunner) makeKey(name string, args []string) string {
	return fmt.Sprintf("%s %s", name, strings.Join(args, " "))
}

// TimeProvider abstracts time operations for testing
type TimeProvider interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

// FakeTimeProvider implements TimeProvider for testing
type FakeTimeProvider struct {
	currentTime time.Time
}

// NewFakeTimeProvider creates a new fake time provider
func NewFakeTimeProvider(currentTime time.Time) *FakeTimeProvider {
	return &FakeTimeProvider{currentTime: currentTime}
}

// Now returns the fake current time
func (f *FakeTimeProvider) Now() time.Time {
	return f.currentTime
}

// Since returns the duration since the given time
func (f *FakeTimeProvider) Since(t time.Time) time.Duration {
	return f.currentTime.Sub(t)
}

// SetTime updates the fake current time
func (f *FakeTimeProvider) SetTime(t time.Time) {
	f.currentTime = t
}

// AdvanceTime advances the fake time by the given duration
func (f *FakeTimeProvider) AdvanceTime(d time.Duration) {
	f.currentTime = f.currentTime.Add(d)
}

// Helpers for creating test data

// MakeFakeTmuxListOutput creates fake tmux list-sessions output
func MakeFakeTmuxListOutput(sessions []string) string {
	var buf bytes.Buffer
	for _, session := range sessions {
		buf.WriteString(fmt.Sprintf("%s: 1 windows (created %s) [80x24]\n", 
			session, time.Now().Format("Mon Jan 2 15:04:05 2006")))
	}
	return buf.String()
}

// MakeFakeUziLsJSON creates fake JSON output for uzi ls --json
func MakeFakeUziLsJSON(sessions []SessionInfo) string {
	if len(sessions) == 0 {
		return "[]"
	}

	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, session := range sessions {
		if i > 0 {
			buf.WriteString(",\n")
		}
		buf.WriteString(fmt.Sprintf(`  {
    "name": "%s",
    "agent_name": "%s",
    "model": "%s",
    "status": "%s",
    "prompt": "%s",
    "insertions": %d,
    "deletions": %d,
    "worktree_path": "%s",
    "port": %d
  }`, session.Name, session.AgentName, session.Model, session.Status,
			session.Prompt, session.Insertions, session.Deletions,
			session.WorktreePath, session.Port))
	}
	buf.WriteString("\n]")
	return buf.String()
}

// SessionInfo represents a session for test data
type SessionInfo struct {
	Name         string
	AgentName    string
	Model        string
	Status       string
	Prompt       string
	Insertions   int
	Deletions    int
	WorktreePath string
	Port         int
}

// Must wraps a function call and panics if it returns an error
// Useful for test setup where failures should halt the test
func Must[T any](value T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("Must failed: %v", err))
	}
	return value
}

// Require provides simple assertion helpers for tests
type Require struct {
	t TestingT
}

// TestingT is a minimal interface for testing.T
type TestingT interface {
	Helper()
	Errorf(format string, args ...interface{})
	FailNow()
}

// NewRequire creates a new Require instance
func NewRequire(t TestingT) *Require {
	return &Require{t: t}
}

// NoError asserts that err is nil
func (r *Require) NoError(err error, msgAndArgs ...interface{}) {
	r.t.Helper()
	if err != nil {
		r.t.Errorf("Expected no error, got: %v %v", err, msgAndArgs)
		r.t.FailNow()
	}
}

// Error asserts that err is not nil
func (r *Require) Error(err error, msgAndArgs ...interface{}) {
	r.t.Helper()
	if err == nil {
		r.t.Errorf("Expected error, got nil %v", msgAndArgs)
		r.t.FailNow()
	}
}

// Equal asserts that two values are equal
func (r *Require) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	r.t.Helper()
	if expected != actual {
		r.t.Errorf("Expected %v, got %v %v", expected, actual, msgAndArgs)
		r.t.FailNow()
	}
}

// NotNil asserts that value is not nil
func (r *Require) NotNil(value interface{}, msgAndArgs ...interface{}) {
	r.t.Helper()
	if value == nil {
		r.t.Errorf("Expected non-nil value %v", msgAndArgs)
		r.t.FailNow()
	}
}

// Nil asserts that value is nil
func (r *Require) Nil(value interface{}, msgAndArgs ...interface{}) {
	r.t.Helper()
	if value != nil {
		r.t.Errorf("Expected nil, got %v %v", value, msgAndArgs)
		r.t.FailNow()
	}
}

// True asserts that value is true
func (r *Require) True(value bool, msgAndArgs ...interface{}) {
	r.t.Helper()
	if !value {
		r.t.Errorf("Expected true, got false %v", msgAndArgs)
		r.t.FailNow()
	}
}

// False asserts that value is false
func (r *Require) False(value bool, msgAndArgs ...interface{}) {
	r.t.Helper()
	if value {
		r.t.Errorf("Expected false, got true %v", msgAndArgs)
		r.t.FailNow()
	}
}
