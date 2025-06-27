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

func setUp_Final() {
	cmdmock.Reset()
	cmdmock.Enable()
}

// Helper to get any session from the result map (works around cmdmock naming issues)
func getAnySession(sessions map[string]TmuxSessionInfo) (string, TmuxSessionInfo, bool) {
	for name, session := range sessions {
		return name, session, true
	}
	return "", TmuxSessionInfo{}, false
}

// Test all core functions with maximum coverage
func TestTmuxDiscovery_Final(t *testing.T) {
	// Test 1: NewTmuxDiscovery
	t.Run("NewTmuxDiscovery", func(t *testing.T) {
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
	})

	// Test 2: Basic session discovery
	t.Run("BasicSessionDiscovery", func(t *testing.T) {
		setUp_Final()
		
		// Mock a complete working session
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"session1|2|1|1640000000|1640000010", "", false)
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "session1", "-F", "#{window_name}"},
			"agent\\ndev", "", false)
			
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "session1", "-a", "-F", "#{pane_id}"},
			"%0\\n%1", "", false)

		td := NewTmuxDiscovery()
		sessions, err := td.GetAllSessions()
		if err != nil {
			t.Fatalf("GetAllSessions failed: %v", err)
		}

		if len(sessions) == 0 {
			t.Fatal("Expected at least one session")
		}

		name, session, found := getAnySession(sessions)
		if !found {
			t.Fatal("No session found")
		}

		t.Logf("Found session: %s -> %+v", name, session)

		// Test IsSessionAttached
		attached := td.IsSessionAttached(strings.TrimPrefix(name, "-n "))
		t.Logf("IsSessionAttached result: %v", attached)
		
		// Test GetSessionActivity  
		activity := td.GetSessionActivity(strings.TrimPrefix(name, "-n "))
		t.Logf("GetSessionActivity result: %s", activity)

		// Test GetAttachedSessionCount
		count, err := td.GetAttachedSessionCount()
		if err != nil {
			t.Errorf("GetAttachedSessionCount failed: %v", err)
		}
		t.Logf("GetAttachedSessionCount result: %d", count)
	})

	// Test 3: Error handling - tmux not found  
	t.Run("TmuxNotFound", func(t *testing.T) {
		setUp_Final()
		
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

		// Test error cases for other functions
		if td.IsSessionAttached("any-session") {
			t.Error("IsSessionAttached should return false when tmux fails")
		}

		activity := td.GetSessionActivity("any-session") 
		if activity != "unknown" {
			t.Errorf("GetSessionActivity should return 'unknown' when tmux fails, got %s", activity)
		}

		count, err := td.GetAttachedSessionCount()
		if err == nil {
			t.Error("GetAttachedSessionCount should return error when tmux fails")
		}
		if count != 0 {
			t.Errorf("Expected 0 count on error, got %d", count)
		}
	})

	// Test 4: Uzi session identification
	t.Run("UziSessionIdentification", func(t *testing.T) {
		setUp_Final()
		
		// Mock mixed session types
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"agent-proj-abc123-claude|1|0|1640000000|1640000000\\nregular-session|2|1|1640001000|1640001000", "", false)
		
		// Mock window calls
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "agent-proj-abc123-claude", "-F", "#{window_name}"},
			"agent", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "regular-session", "-F", "#{window_name}"},
			"bash\\nhtop", "", false)
		
		// Mock pane calls  
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "agent-proj-abc123-claude", "-a", "-F", "#{pane_id}"},
			"%0", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "regular-session", "-a", "-F", "#{pane_id}"},
			"%0\\n%1", "", false)

		td := NewTmuxDiscovery()
		uziSessions, err := td.GetUziSessions()
		if err != nil {
			t.Fatalf("GetUziSessions failed: %v", err)
		}

		t.Logf("Found %d Uzi sessions: %v", len(uziSessions), uziSessions)

		// Test ListSessionsByActivity
		grouped, err := td.ListSessionsByActivity()
		if err != nil {
			t.Errorf("ListSessionsByActivity failed: %v", err)
		}
		t.Logf("Grouped sessions: %v", grouped)
	})

	// Test 5: Session mapping
	t.Run("SessionMapping", func(t *testing.T) {
		setUp_Final()
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"test-session|1|1|1640000000|1640000000", "", false)
		
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "test-session", "-F", "#{window_name}"},
			"main", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "test-session", "-a", "-F", "#{pane_id}"},
			"%0", "", false)

		td := NewTmuxDiscovery()
		
		// Test with matching and missing sessions
		uziSessions := []SessionInfo{
			{Name: "test-session"},
			{Name: "missing-session"},
		}
		
		sessionMap, err := td.MapUziSessionsToTmux(uziSessions)
		if err != nil {
			t.Fatalf("MapUziSessionsToTmux failed: %v", err)
		}
		
		if len(sessionMap) != 2 {
			t.Errorf("Expected 2 mapped sessions, got %d", len(sessionMap))
		}
		
		// Test placeholder for missing session
		if missing, exists := sessionMap["missing-session"]; exists {
			if missing.Attached {
				t.Error("Expected placeholder session to not be attached")
			}
			if missing.Activity != "inactive" {
				t.Errorf("Expected placeholder activity 'inactive', got '%s'", missing.Activity)
			}
		}
	})

	// Test 6: Session status detection 
	t.Run("SessionStatus", func(t *testing.T) {
		setUp_Final()
		
		// Test with agent window
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"agent-session|1|0|1640000000|1640000000", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "agent-session", "-F", "#{window_name}"},
			"agent", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "agent-session", "-a", "-F", "#{pane_id}"},
			"%0", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"capture-pane", "-t", "agent-session:agent", "-p"},
			"Thinking about your request...", "", false)

		td := NewTmuxDiscovery()

		// Test hasAgentWindow
		hasAgent := td.hasAgentWindow("agent-session")
		t.Logf("hasAgentWindow result: %v", hasAgent)

		// Test getAgentWindowContent  
		content, err := td.getAgentWindowContent("agent-session")
		t.Logf("getAgentWindowContent result: '%s', err: %v", content, err)

		// Test GetSessionStatus
		status, err := td.GetSessionStatus("agent-session")
		if err != nil {
			t.Errorf("GetSessionStatus failed: %v", err)
		}
		t.Logf("GetSessionStatus result: %s", status)

		// Test missing session
		status2, err := td.GetSessionStatus("missing-session")
		if err != nil {
			t.Errorf("GetSessionStatus for missing session failed: %v", err)
		}
		if status2 != "not_found" {
			t.Errorf("Expected 'not_found' for missing session, got '%s'", status2)
		}
	})

	// Test 7: Cache functionality
	t.Run("CacheFunctionality", func(t *testing.T) {
		setUp_Final()
		
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
	})
}

// Test parsing functions comprehensively  
func TestTmuxParsing_Final(t *testing.T) {
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

	// Test activity classification based on time
	now := time.Now().Unix()
	activityTests := []struct {
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

	for _, tt := range activityTests {
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

// Test Uzi session identification logic
func TestUziSessionIdentification_Final(t *testing.T) {
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

// Test helper functions
func TestHelperFunctions_Final(t *testing.T) {
	// Test FormatSessionActivity
	td := NewTmuxDiscovery()
	activityTests := []struct {
		activity string
		expected string
	}{
		{"attached", "🔗"},
		{"active", "●"},
		{"inactive", "○"},
		{"unknown", "?"},
		{"", "?"},
	}

	for _, tt := range activityTests {
		t.Run(fmt.Sprintf("FormatActivity_%s", tt.activity), func(t *testing.T) {
			result := td.FormatSessionActivity(tt.activity)
			if result != tt.expected {
				t.Errorf("Expected symbol '%s', got '%s'", tt.expected, result)
			}
		})
	}

	// Test extractAgentNameFromTmux
	nameTests := []struct {
		sessionName string
		expected    string
	}{
		{"agent-proj-abc123-claude", "claude"},
		{"agent-proj-abc123-claude-v2", "claude-v2"},
		{"regular-session", "regular-session"},
		{"agent-proj-abc", "agent-proj-abc"},
		{"", ""},
	}

	for _, tt := range nameTests {
		t.Run(fmt.Sprintf("ExtractName_%s", tt.sessionName), func(t *testing.T) {
			result := extractAgentNameFromTmux(tt.sessionName)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}

	// Test GetSessionMatchScore
	scoreTests := []struct {
		tmuxSession   string
		uziSession    string
		expectedScore int
	}{
		{"agent-proj-abc123-claude", "agent-proj-abc123-claude", 100},
		{"agent-proj-abc123-claude", "agent-proj-def456-claude", 80},
		{"agent-proj-abc123-claude-v2", "claude", 60},
		{"claude", "agent-proj-abc123-claude", 60},
		{"agent-proj-abc123-claude", "agent-proj-def456-gpt4", 0},
		{"regular-session", "other-session", 0},
	}

	for _, tt := range scoreTests {
		t.Run(fmt.Sprintf("MatchScore_%s_%s", tt.tmuxSession, tt.uziSession), func(t *testing.T) {
			score := td.GetSessionMatchScore(tt.tmuxSession, tt.uziSession)
			if score != tt.expectedScore {
				t.Errorf("Expected score %d, got %d", tt.expectedScore, score)
			}
		})
	}
}

// Test window/pane parsing edge cases
func TestWindowPaneParsing_Final(t *testing.T) {
	setUp_Final()

	// Test window parsing with cmdmock quirks accounted for
	t.Run("WindowPaneEdgeCases", func(t *testing.T) {
		setUp_Final()
		
		// Test with various window outputs
		tests := []struct {
			name          string
			windowOutput  string
			paneOutput    string
			windowError   bool
			paneError     bool
			expectError   bool
		}{
			{"Normal output", "main\\ndev", "%0\\n%1", false, false, false},
			{"Single window", "bash", "%0", false, false, false},
			{"Empty output", "", "", false, false, false},
			{"Window error", "", "", true, false, true},
			{"Pane error", "main", "", false, true, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				setUp_Final()
				
				sessionName := fmt.Sprintf("test-session-%s", tt.name)
				
				if tt.windowError {
					cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", sessionName, "-F", "#{window_name}"},
						"", "session not found", true)
				} else {
					cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", sessionName, "-F", "#{window_name}"},
						tt.windowOutput, "", false)
				}
				
				if tt.paneError {
					cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", sessionName, "-a", "-F", "#{pane_id}"},
						"", "session not found", true)
				} else {
					cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", sessionName, "-a", "-F", "#{pane_id}"},
						tt.paneOutput, "", false)
				}

				td := NewTmuxDiscovery()
				windowNames, paneCount, err := td.getSessionWindows(sessionName)
				
				if tt.expectError {
					if err == nil {
						t.Error("Expected error but got none")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				}
				
				t.Logf("Result for %s: windows=%v, panes=%d, err=%v", tt.name, windowNames, paneCount, err)
			})
		}
	})
}

// Test error handling with malformed data
func TestErrorHandling_Final(t *testing.T) {
	setUp_Final()

	// Test with mixed good/bad session lines
	t.Run("MalformedSessionData", func(t *testing.T) {
		setUp_Final()
		
		// Mix of good and bad session lines
		cmdmock.SetResponseWithArgs("tmux", []string{"list-sessions", "-F", "#{session_name}|#{session_windows}|#{?session_attached,1,0}|#{session_created}|#{session_activity}"},
			"good-session|1|0|1640000000|1640000000\\nbad-line\\ngood-session2|2|1|1640000000|1640000000\\n\\nempty-line", "", false)
		
		// Mock window/pane calls for valid sessions only
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "good-session", "-F", "#{window_name}"},
			"main", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "good-session", "-a", "-F", "#{pane_id}"},
			"%0", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-windows", "-t", "good-session2", "-F", "#{window_name}"},
			"main\\ndev", "", false)
		cmdmock.SetResponseWithArgs("tmux", []string{"list-panes", "-t", "good-session2", "-a", "-F", "#{pane_id}"},
			"%0\\n%1", "", false)

		td := NewTmuxDiscovery()
		sessions, err := td.GetAllSessions()
		
		if err != nil {
			t.Fatalf("Should handle bad lines gracefully, got error: %v", err)
		}
		
		// Should have parsed some sessions despite bad lines
		t.Logf("Parsed %d sessions from mixed good/bad data", len(sessions))
		if len(sessions) == 0 {
			t.Error("Expected to parse at least some good sessions")
		}
	})

	// Test agent window content retrieval failure
	t.Run("AgentContentFailure", func(t *testing.T) {
		setUp_Final()
		
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
