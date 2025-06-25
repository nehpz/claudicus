// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package fsmock_test

import (
	"testing"

	"github.com/nehpz/claudicus/pkg/testutil/fsmock"
)

// Example showing basic usage of fsmock for testing file operations
func ExampleTempFS_basic() {
	// This would typically be inside a test function
	t := &testing.T{} // In real usage, this comes from the test function parameter
	
	// Create a temporary filesystem
	fs := fsmock.NewTempFS(t)
	
	// Create a file
	fs.WriteFileString("config.txt", "host=localhost\nport=8080", 0644)
	
	// Read it back
	content, _ := fs.ReadFileString("config.txt")
	_ = content // host=localhost\nport=8080
	
	// Create nested directories
	fs.MkdirAll("logs/app", 0755)
	fs.WriteFileString("logs/app/server.log", "Starting server...", 0644)
	
	// Check if files exist
	_ = fs.Exists("config.txt")     // true
	_ = fs.IsFile("config.txt")     // true
	_ = fs.IsDir("logs")            // true
	_ = fs.IsDir("logs/app")        // true
	
	// Get absolute path for use with other tools
	configPath := fs.Path("config.txt")
	_ = configPath // Full absolute path to the temp file
	
	// The filesystem will be automatically cleaned up when the test ends
}

// Example showing how to use fsmock for testing git repository operations
func ExampleTempFS_gitRepo() {
	t := &testing.T{}
	
	fs := fsmock.NewTempFS(t)
	
	// Create a mock git repository
	fs.CreateGitRepo("myproject")
	
	// Add some project files
	fs.WriteFileString("myproject/main.go", "package main\n\nfunc main() {}\n", 0644)
	fs.WriteFileString("myproject/go.mod", "module myproject\n\ngo 1.21\n", 0644)
	
	// Create git worktree structure for testing
	fs.MkdirAll("myproject/.git/worktrees/feature-branch", 0755)
	fs.WriteFileString("myproject/.git/worktrees/feature-branch/HEAD", "ref: refs/heads/feature-branch\n", 0644)
	
	// Now you can test git-related operations with a predictable repository structure
	repoPath := fs.Path("myproject")
	_ = repoPath // Use this path with git commands or repository code
}

// Example showing how to test state persistence with temporary files
func ExampleTempFS_statePersistence() {
	t := &testing.T{}
	
	fs := fsmock.NewTempFS(t)
	
	// Create a state directory structure
	fs.MkdirAll("state/sessions", 0755)
	
	// Simulate saving session state
	sessionData := `{
	"name": "test-session",
	"agent": "claude",
	"status": "active",
	"created": "2023-01-01T12:00:00Z"
}`
	fs.WriteFileString("state/sessions/test-session.json", sessionData, 0644)
	
	// Test loading state
	loadedData, _ := fs.ReadFileString("state/sessions/test-session.json")
	_ = loadedData // Verify the content matches what was saved
	
	// Test state directory listing
	entries, _ := fs.ListDir("state/sessions")
	_ = len(entries) // Should be 1
	
	// Test cleanup of old state files
	fs.Remove("state/sessions/test-session.json")
	_ = fs.Exists("state/sessions/test-session.json") // false
}

// Example showing integration with existing testutil patterns
func ExampleTempFS_withExistingTestUtil() {
	t := &testing.T{}
	
	fs := fsmock.NewTempFS(t)
	
	// Create a project structure for testing
	fs.CreateProjectStructure("testproject")
	
	// Add some test-specific files
	fs.WriteFileString("testproject/uzi.yaml", "version: 1\ndefault_model: claude-3", 0644)
	
	// Create a test state directory
	fs.MkdirAll("testproject/.uzi/state", 0755)
	
	// Now you can test your application code that works with files
	projectPath := fs.Path("testproject")
	_ = projectPath // Pass this to your application code
	
	// Example: Test that your config loading works
	configPath := fs.Path("testproject/uzi.yaml")
	_ = configPath // Pass this to your config loading function
	
	// Example: Test state persistence
	statePath := fs.Path("testproject/.uzi/state")
	_ = statePath // Pass this to your state management code
	
	// All files will be automatically cleaned up when the test ends
}
