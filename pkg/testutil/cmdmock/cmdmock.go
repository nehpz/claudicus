// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package cmdmock

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// CommandCall represents a recorded command execution
type CommandCall struct {
	Name string
	Args []string
	Dir  string // Working directory when command was called
}

// CommandResponse represents the response configuration for a command
type CommandResponse struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// MockState holds the global mock state
type MockState struct {
	mu        sync.RWMutex
	responses map[string]CommandResponse
	calls     []CommandCall
	enabled   bool
}

var globalMock = &MockState{
	responses: make(map[string]CommandResponse),
	calls:     make([]CommandCall, 0),
	enabled:   false,
}

// Command is a mock replacement for exec.Command that can be used via:
//   var execCommand = cmdmock.Command
// in production files for testing
func Command(name string, args ...string) *exec.Cmd {
	globalMock.mu.Lock()
	defer globalMock.mu.Unlock()

	if !globalMock.enabled {
		// If mocking is not enabled, return the real command
		return exec.Command(name, args...)
	}

	// Record the call
	call := CommandCall{
		Name: name,
		Args: args,
		Dir:  getCurrentDir(),
	}
	globalMock.calls = append(globalMock.calls, call)

	// Create a mock command
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, args...),
	}

	// Look up the response
	key := makeKey(name, args)
	response, exists := globalMock.responses[key]
	if !exists {
		// Default response for unmocked commands
		response = CommandResponse{
			Stdout:   "",
			Stderr:   fmt.Sprintf("command not mocked: %s %s", name, strings.Join(args, " ")),
			ExitCode: 1,
		}
	}

	// Create a fake command that will return our predetermined response
	cmd = createMockCommand(response)
	
	return cmd
}

// SetResponse configures the mock response for a specific command
// cmd is the command name (e.g., "git", "tmux")
// stdout is the stdout output to return
// exitErr indicates whether the command should return an exit error
func SetResponse(cmd string, stdout string, exitErr bool) {
	SetResponseWithArgs(cmd, []string{}, stdout, "", exitErr)
}

// SetResponseWithArgs configures the mock response for a command with specific arguments
func SetResponseWithArgs(cmd string, args []string, stdout, stderr string, exitErr bool) {
	globalMock.mu.Lock()
	defer globalMock.mu.Unlock()

	key := makeKey(cmd, args)
	exitCode := 0
	if exitErr {
		exitCode = 1
	}

	globalMock.responses[key] = CommandResponse{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}
	globalMock.enabled = true
}

// Reset clears all recorded calls and responses
func Reset() {
	globalMock.mu.Lock()
	defer globalMock.mu.Unlock()

	globalMock.responses = make(map[string]CommandResponse)
	globalMock.calls = make([]CommandCall, 0)
	globalMock.enabled = false
}

// Enable turns on command mocking
func Enable() {
	globalMock.mu.Lock()
	defer globalMock.mu.Unlock()
	globalMock.enabled = true
}

// Disable turns off command mocking (commands will execute normally)
func Disable() {
	globalMock.mu.Lock()
	defer globalMock.mu.Unlock()
	globalMock.enabled = false
}

// GetCalls returns all recorded command calls
func GetCalls() []CommandCall {
	globalMock.mu.RLock()
	defer globalMock.mu.RUnlock()

	// Return a copy to avoid race conditions
	calls := make([]CommandCall, len(globalMock.calls))
	copy(calls, globalMock.calls)
	return calls
}

// GetCallCount returns the number of commands that were called
func GetCallCount() int {
	globalMock.mu.RLock()
	defer globalMock.mu.RUnlock()
	return len(globalMock.calls)
}

// WasCommandCalled checks if a specific command was called
func WasCommandCalled(cmd string, args ...string) bool {
	globalMock.mu.RLock()
	defer globalMock.mu.RUnlock()

	targetKey := makeKey(cmd, args)
	for _, call := range globalMock.calls {
		if makeKey(call.Name, call.Args) == targetKey {
			return true
		}
	}
	return false
}

// GetCommandCalls returns all calls that match the given command and args
func GetCommandCalls(cmd string, args ...string) []CommandCall {
	globalMock.mu.RLock()
	defer globalMock.mu.RUnlock()

	var matches []CommandCall
	targetKey := makeKey(cmd, args)
	
	for _, call := range globalMock.calls {
		if makeKey(call.Name, call.Args) == targetKey {
			matches = append(matches, call)
		}
	}
	return matches
}

// Helper functions

func makeKey(cmd string, args []string) string {
	return fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// createMockCommand creates a command that returns predetermined output and exit code
func createMockCommand(response CommandResponse) *exec.Cmd {
	// Create a command that will echo our response and exit with the desired code
	script := fmt.Sprintf(`
		echo -n "%s"
		if [ "%s" != "" ]; then
			echo -n "%s" >&2
		fi
		exit %d
	`, escapeShell(response.Stdout), response.Stderr, escapeShell(response.Stderr), response.ExitCode)
	
	return exec.Command("sh", "-c", script)
}

// escapeShell escapes strings for safe use in shell commands
func escapeShell(s string) string {
	// Replace problematic characters
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
