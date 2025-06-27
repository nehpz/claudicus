// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nehpz/claudicus/pkg/testutil/cmdmock"
)

func init() {
	// Override execCommand for comprehensive tests
	execCommand = cmdmock.Command
}

func setUp_Comprehensive() {
	cmdmock.Reset()
	cmdmock.Enable()
}

// Helper function to clean cmdmock output (removes -n prefix)
func cleanOutput(s string) string {
	if strings.HasPrefix(s, "-n ") {
		return s[3:]
	}
	return s
}

// Test comprehensive session discovery and parsing
func TestTmuxDiscovery_Comprehensive(t *testing.T) {
	setUp_Comprehensive()

	// Test 1: Basic session discovery with real tmux-like output
	t.Run("BasicSessionDiscovery", func(t *testing.T) {
		setUp_Comprehensive()
		
		// Mock tmux list-sessions 
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"session1|2|1|1640000000|1640000010", "", false)
		
		// Mock list-windows for session1
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "session1", "-F", "#{window_name}"},
			"agent\ndev", "", false)
			
		// Mock list-panes for session1
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "session1", "-a", "-F", "#{pane_id}"},
			"%0\n%1", "", false)

		td := NewTmuxDiscovery()
		sessions, err := td.GetAllSessions()
		if err != nil {
			t.Fatalf("GetAllSessions failed: %v", err)
		}

		if len(sessions) == 0 {
			t.Fatal("Expected at least one session")
		}

		// Find our session (might have -n prefix due to cmdmock)
		var session TmuxSessionInfo
		var found bool
		for name, s := range sessions {
			if strings.Contains(name, "session1") {
				session = s
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("session1 not found in: %v", sessions)
		}

		if session.Windows != 2 {
			t.Errorf("Expected 2 windows, got %d", session.Windows)
		}

		if session.Panes != 2 {
			t.Errorf("Expected 2 panes, got %d", session.Panes)
		}
	})

	// Test 2: Error handling - tmux not found
	t.Run("TmuxNotFound", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"", "command not found: tmux", true)

		td := NewTmuxDiscovery()
		sessions, err := td.GetAllSessions()
		
		// Should not error but return empty sessions
		if err != nil {
			t.Errorf("Should handle tmux not found gracefully, got error: %v", err)
		}
		
		if len(sessions) != 0 {
			t.Errorf("Expected empty sessions when tmux not found, got %d", len(sessions))
		}
	})

	// Test 3: Bad tmux output parsing
	t.Run("BadTmuxOutput", func(t *testing.T) {
		setUp_Comprehensive()
		
		// Mix of good and bad lines
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"good-session|1|0|1640000000|1640000000\nbad-line\ngood-session2|2|1|1640000000|1640000000", "", false)
		
		// Mock window/pane calls for valid sessions
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "good-session", "-F", "#{window_name}"},
			"main", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "good-session", "-a", "-F", "#{pane_id}"},
			"%0", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "good-session2", "-F", "#{window_name}"},
			"bash\nhtop", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "good-session2", "-a", "-F", "#{pane_id}"},
			"%0\n%1", "", false)

		td := NewTmuxDiscovery()
		sessions, err := td.GetAllSessions()
		
		if err != nil {
			t.Fatalf("Should handle bad lines gracefully, got error: %v", err)
		}
		
		// Should have parsed the good sessions only
		if len(sessions) < 2 {
			t.Errorf("Expected at least 2 valid sessions, got %d", len(sessions))
		}
	})
}

// Test parsing functions with edge cases
func TestTmuxParsing_Comprehensive(t *testing.T) {
	td := NewTmuxDiscovery()

	// Test parseSessionLine with various inputs
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkName   string
		checkAttached bool
		checkActivity string
	}{
		{
			name:          "Valid attached session",
			input:         "session1|2|1|1640000000|1640000100",
			expectError:   false,
			checkName:     "session1",
			checkAttached: true,
			checkActivity: "attached",
		},
		{
			name:          "Valid detached inactive session",
			input:         "session2|1|0|1640000000|1640000000",
			expectError:   false,
			checkName:     "session2",
			checkAttached: false,
			checkActivity: "inactive",
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
			
			if result.Name != tt.checkName {
				t.Errorf("Expected name %s, got %s", tt.checkName, result.Name)
			}
			if result.Attached != tt.checkAttached {
				t.Errorf("Expected attached %v, got %v", tt.checkAttached, result.Attached)
			}
			if result.Activity != tt.checkActivity {
				t.Errorf("Expected activity %s, got %s", tt.checkActivity, result.Activity)
			}
		})
	}
}

// Test activity classification based on time
func TestActivityClassification_Comprehensive(t *testing.T) {
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

// Test Uzi session identification
func TestUziSessionIdentification_Comprehensive(t *testing.T) {
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

// Test session status detection
func TestSessionStatus_Comprehensive(t *testing.T) {
	setUp_Comprehensive()

	// Test attached session status
	t.Run("AttachedSession", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"attached-session|1|1|1640000000|1640000000", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "attached-session", "-F", "#{window_name}"},
			"main", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "attached-session", "-a", "-F", "#{pane_id}"},
			"%0", "", false)

		td := NewTmuxDiscovery()
		status, err := td.GetSessionStatus("attached-session")
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if status != "attached" {
			t.Errorf("Expected status 'attached', got '%s'", status)
		}
	})

	// Test session not found
	t.Run("SessionNotFound", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"other-session|1|0|1640000000|1640000000", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "other-session", "-F", "#{window_name}"},
			"bash", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "other-session", "-a", "-F", "#{pane_id}"},
			"%0", "", false)

		td := NewTmuxDiscovery()
		status, err := td.GetSessionStatus("missing-session")
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if status != "not_found" {
			t.Errorf("Expected status 'not_found', got '%s'", status)
		}
	})
}

// Test activity symbol formatting
func TestFormatSessionActivity_Comprehensive(t *testing.T) {
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

// Test agent name extraction
func TestExtractAgentName_Comprehensive(t *testing.T) {
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

// Test session match scoring
func TestGetSessionMatchScore_Comprehensive(t *testing.T) {
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

// Test cache functionality
func TestCaching_Comprehensive(t *testing.T) {
	setUp_Comprehensive()

	// Mock session data
	cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
		"session1|1|0|1640000000|1640000000", "", false)
	cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "session1", "-F", "#{window_name}"},
		"main", "", false)
	cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "session1", "-a", "-F", "#{pane_id}"},
		"%0", "", false)

	td := NewTmuxDiscovery()

	// First call
	sessions1, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	initialCallCount := cmdmock.GetCallCount()

	// Second call within cache time should use cache
	sessions2, err := td.GetAllSessions()
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	// Should be same data and not make additional calls
	if len(sessions1) != len(sessions2) {
		t.Error("Cached result should be identical")
	}
	
	if cmdmock.GetCallCount() > initialCallCount {
		t.Error("Second call should use cache, not call tmux again")
	}

	// Test cache refresh
	td.RefreshCache()
	if !td.lastUpdate.IsZero() {
		t.Error("RefreshCache should reset lastUpdate to zero")
	}
}

// Test error handling comprehensively
func TestErrorHandling_Comprehensive(t *testing.T) {
	// Test GetSessionActivity with tmux error
	t.Run("GetSessionActivity_TmuxError", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"", "no server running", true)

		td := NewTmuxDiscovery()
		activity := td.GetSessionActivity("any-session")
		if activity != "unknown" {
			t.Errorf("Expected unknown activity when tmux fails, got %s", activity)
		}
	})

	// Test IsSessionAttached with error
	t.Run("IsSessionAttached_TmuxError", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"", "no server running", true)

		td := NewTmuxDiscovery()
		attached := td.IsSessionAttached("any-session")
		if attached {
			t.Error("Expected false when tmux fails")
		}
	})

	// Test GetAttachedSessionCount with error
	t.Run("GetAttachedSessionCount_TmuxError", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"", "no server running", true)

		td := NewTmuxDiscovery()
		count, err := td.GetAttachedSessionCount()
		if err == nil {
			t.Error("Expected error when tmux fails")
		}
		if count != 0 {
			t.Errorf("Expected 0 count on error, got %d", count)
		}
	})
}

// Test window and pane parsing edge cases
func TestWindowPaneParsing_Comprehensive(t *testing.T) {
	setUp_Comprehensive()

	// Test empty window/pane output
	t.Run("EmptyWindowPaneOutput", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "empty-session", "-F", "#{window_name}"},
			"", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "empty-session", "-a", "-F", "#{pane_id}"},
			"", "", false)

		td := NewTmuxDiscovery()
		windowNames, paneCount, err := td.getSessionWindows("empty-session")
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(windowNames) != 0 {
			t.Errorf("Expected 0 window names for empty output, got %d", len(windowNames))
		}
		
		if paneCount != 0 {
			t.Errorf("Expected 0 panes for empty output, got %d", paneCount)
		}
	})

	// Test window command fails but pane succeeds
	t.Run("WindowCommandFails", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "fail-session", "-F", "#{window_name}"},
			"", "session not found", true)

		td := NewTmuxDiscovery()
		windowNames, paneCount, err := td.getSessionWindows("fail-session")
		
		if err == nil {
			t.Error("Expected error when window command fails")
		}
		
		// Should return empty data
		if len(windowNames) != 0 {
			t.Errorf("Expected 0 window names on error, got %d", len(windowNames))
		}
		
		if paneCount != 0 {
			t.Errorf("Expected 0 panes on error, got %d", paneCount)
		}
	})
}

// Test agent window detection and content
func TestAgentWindowDetection_Comprehensive(t *testing.T) {
	setUp_Comprehensive()

	// Test hasAgentWindow
	t.Run("HasAgentWindow", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"agent-session|1|0|1640000000|1640000000", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "agent-session", "-F", "#{window_name}"},
			"agent", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "agent-session", "-a", "-F", "#{pane_id}"},
			"%0", "", false)

		td := NewTmuxDiscovery()
		hasAgent := td.hasAgentWindow("agent-session")
		
		if !hasAgent {
			t.Error("Expected session to have agent window")
		}
	})

	// Test getAgentWindowContent
	t.Run("GetAgentWindowContent", func(t *testing.T) {
		setUp_Comprehensive()
		
		expectedContent := "$ echo hello\nhello\n$ "
		cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "agent-session:agent", "-p"},
			expectedContent, "", false)

		td := NewTmuxDiscovery()
		content, err := td.getAgentWindowContent("agent-session")
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		// Account for cmdmock adding -n prefix
		content = cleanOutput(content)
		if content != expectedContent {
			t.Errorf("Expected content '%s', got '%s'", expectedContent, content)
		}
	})

	// Test getAgentWindowContent failure
	t.Run("GetAgentWindowContent_Fail", func(t *testing.T) {
		setUp_Comprehensive()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "missing-session:agent", "-p"},
			"", "no such window", true)

		td := NewTmuxDiscovery()
		content, err := td.getAgentWindowContent("missing-session")
		
		if err == nil {
			t.Error("Expected error when capture-pane fails")
		}
		
		if content != "" {
			t.Errorf("Expected empty content on error, got '%s'", content)
		}
	})
}
