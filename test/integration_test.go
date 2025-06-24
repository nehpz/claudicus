package test

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/devflowinc/uzi/pkg/tui"
)

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
