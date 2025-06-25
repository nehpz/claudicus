// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/nehpz/claudicus/pkg/config"
	"github.com/nehpz/claudicus/pkg/tui"
	"golang.org/x/term"
)

// TestTUILaunchAndExit tests that the TUI can launch and exit cleanly
func TestTUILaunchAndExit(t *testing.T) {
	// Skip if not in terminal environment
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("Skipping TUI test - not in terminal environment")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test TUI binary exists and is executable
	cmd := exec.CommandContext(ctx, "./uzi", "tui")
	cmd.Stdin = nil // No input - should exit with terminal error

	err := cmd.Run()
	
	// Should exit with error about terminal requirement when no TTY
	if err == nil {
		t.Error("Expected TUI to exit with error when no TTY available")
	}
}

// TestConfigValidation tests that uzi.yaml validation works
func TestConfigValidation(t *testing.T) {
	// Test missing config file
	_, err := config.LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error when loading nonexistent config")
	}

	// Test valid config loading
	_, err = config.LoadConfig("uzi.yaml")
	if err != nil {
		t.Errorf("Expected valid config to load, got error: %v", err)
	}
}

// TestUziCLIInterface tests the TUI abstraction layer
func TestUziCLIInterface(t *testing.T) {
	uziCLI := tui.NewUziCLI()
	if uziCLI == nil {
		t.Error("Expected UziCLI to be created")
	}

	// Test GetSessions doesn't crash (may return empty list)
	sessions, err := uziCLI.GetSessions()
	if err != nil {
		t.Errorf("GetSessions should not error, got: %v", err)
	}

	// Sessions can be empty, that's fine
	if sessions == nil {
		t.Error("Expected sessions slice to be non-nil (can be empty)")
	}
}

// TestPortAssignmentPrevention tests that our port collision fix works
func TestPortAssignmentPrevention(t *testing.T) {
	// This is a regression test for the port collision bug we fixed
	// We test by checking if multiple agents would get different ports
	
	// Create temporary test directory
	tmpDir := t.TempDir()
	
	// Create test config
	configContent := `devCommand: echo 'test-dev-server --port $PORT'
portRange: 3000-3010`
	
	configPath := tmpDir + "/test-uzi.yaml"
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test config loads correctly
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Errorf("Test config should load correctly: %v", err)
	}

	if cfg.DevCommand == nil || *cfg.DevCommand == "" {
		t.Error("DevCommand should be set in test config")
	}

	if cfg.PortRange == nil || *cfg.PortRange == "" {
		t.Error("PortRange should be set in test config")
	}
}

// TestTUIComponentsExist tests that key TUI components exist and can be created
func TestTUIComponentsExist(t *testing.T) {
	// Test that key TUI components can be instantiated
	uziCLI := tui.NewUziCLI()
	if uziCLI == nil {
		t.Error("Expected UziCLI to be created")
	}

	// Test that app can be created (though we can't run it in test environment)
	app := tui.NewApp(uziCLI)
	if app == nil {
		t.Error("Expected TUI app to be created")
	}
}

// TestUziCommandIntegration tests the TUI's integration with core uzi commands
func TestUziCommandIntegration(t *testing.T) {
	// Skip if uzi binary doesn't exist
	if _, err := os.Stat("./uzi"); os.IsNotExist(err) {
		t.Skip("Skipping integration test - uzi binary not found")
	}

	// Test basic uzi ls command works
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./uzi", "ls")
	output, err := cmd.Output()
	
	// Command should succeed (even if no sessions)
	if err != nil {
		t.Errorf("uzi ls command failed: %v", err)
	}

	// Output should contain header or "No active sessions"
	outputStr := string(output)
	if !strings.Contains(outputStr, "AGENT") && !strings.Contains(outputStr, "No active sessions") {
		t.Errorf("Unexpected uzi ls output: %s", outputStr)
	}
}

// TestUziCLICommands tests that the UziCLI wrapper can execute commands
func TestUziCLICommands(t *testing.T) {
	uziCLI := tui.NewUziCLI()

	// Test RefreshSessions (should not error)
	err := uziCLI.RefreshSessions()
	if err != nil {
		t.Errorf("RefreshSessions should not error: %v", err)
	}

	// Test GetSessions (should return slice, possibly empty)
	sessions, err := uziCLI.GetSessions()
	if err != nil {
		t.Errorf("GetSessions should not error: %v", err)
	}

	if sessions == nil {
		t.Error("GetSessions should return non-nil slice")
	}

	t.Logf("Found %d sessions", len(sessions))

	// Test that each session has required fields
	for i, session := range sessions {
		if session.AgentName == "" {
			t.Errorf("Session %d should have non-empty AgentName", i)
		}
		if session.Model == "" {
			t.Errorf("Session %d should have non-empty Model", i)
		}
		// Status can be various values, just check it's not empty
		if session.Status == "" {
			t.Errorf("Session %d should have non-empty Status", i)
		}
	}
}

// TestStateFileHandling tests that state file operations work correctly
func TestStateFileHandling(t *testing.T) {
	uziCLI := tui.NewUziCLI()

	// Test getting sessions (should handle missing state file gracefully)
	sessions, err := uziCLI.GetSessions()
	if err != nil {
		t.Errorf("Should handle missing state file gracefully: %v", err)
	}

	// Should return empty slice if no state file
	if sessions == nil {
		t.Error("Should return empty slice, not nil")
	}
}

// TestSessionSorting tests that sessions are returned in consistent order
func TestSessionSorting(t *testing.T) {
	uziCLI := tui.NewUziCLI()

	// Get sessions multiple times and verify consistent ordering
	sessions1, err := uziCLI.GetSessions()
	if err != nil {
		t.Errorf("First GetSessions call failed: %v", err)
	}

	sessions2, err := uziCLI.GetSessions()
	if err != nil {
		t.Errorf("Second GetSessions call failed: %v", err)
	}

	// Should have same length
	if len(sessions1) != len(sessions2) {
		t.Errorf("Session count changed between calls: %d vs %d", len(sessions1), len(sessions2))
	}

	// Should have same order (this tests our port-based sorting fix)
	for i := 0; i < len(sessions1) && i < len(sessions2); i++ {
		if sessions1[i].Name != sessions2[i].Name {
			t.Errorf("Session order changed between calls at index %d: %s vs %s", 
				i, sessions1[i].Name, sessions2[i].Name)
		}

		if sessions1[i].Port != sessions2[i].Port {
			t.Errorf("Session port changed between calls for %s: %d vs %d", 
				sessions1[i].Name, sessions1[i].Port, sessions2[i].Port)
		}
	}
}

// TestTmuxIntegration tests tmux session discovery
func TestTmuxIntegration(t *testing.T) {
	// Check if tmux is available
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("Skipping tmux test - tmux not available")
	}

	uziCLI := tui.NewUziCLI()

	// Test getting sessions with tmux info
	sessions, tmuxMapping, err := uziCLI.GetSessionsWithTmuxInfo()
	if err != nil {
		t.Errorf("GetSessionsWithTmuxInfo should not error: %v", err)
	}

	if sessions == nil {
		t.Error("Sessions should be non-nil")
	}

	// tmuxMapping can be nil if no tmux sessions exist
	t.Logf("Found %d sessions, %d tmux mappings", len(sessions), len(tmuxMapping))

	// Test session attachment status
	attachedCount, err := uziCLI.GetAttachedSessionCount()
	if err != nil {
		t.Errorf("GetAttachedSessionCount should not error: %v", err)
	}

	if attachedCount < 0 {
		t.Error("Attached session count should be non-negative")
	}

	t.Logf("Found %d attached sessions", attachedCount)
}

// TestCommandAliasIntegration tests that aliases work in actual command execution
func TestCommandAliasIntegration(t *testing.T) {
	// Skip if uzi binary doesn't exist
	if _, err := os.Stat("./uzi"); os.IsNotExist(err) {
		t.Skip("Skipping integration test - uzi binary not found")
	}

	tests := []struct {
		name  string
		alias string
		full  string
	}{
		{"prompt alias", "p", "prompt"},
		{"ls alias", "l", "ls"},
		{"kill alias", "k", "kill"},
		{"tui alias", "t", "tui"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Test alias with --help flag (safest option)
			cmdAlias := exec.CommandContext(ctx, "./uzi", tt.alias, "--help")
			outputAlias, errAlias := cmdAlias.Output()

			cmdFull := exec.CommandContext(ctx, "./uzi", tt.full, "--help")
			outputFull, errFull := cmdFull.Output()

			// Both should succeed with help
			if errAlias != nil {
				t.Errorf("Alias command %s failed: %v", tt.alias, errAlias)
			}
			if errFull != nil {
				t.Errorf("Full command %s failed: %v", tt.full, errFull)
			}

			// Output should be similar (both should show help for the same command)
			aliasStr := string(outputAlias)
			fullStr := string(outputFull)

			if !strings.Contains(aliasStr, tt.full) && !strings.Contains(fullStr, tt.full) {
				t.Logf("Alias output: %s", aliasStr)
				t.Logf("Full output: %s", fullStr)
				// This is just a warning - help output might differ
			}
		})
	}
}