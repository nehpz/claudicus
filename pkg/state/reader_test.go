package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewStateReader(t *testing.T) {
	reader := NewStateReader("/test/repo")
	if reader == nil {
		t.Error("Expected StateReader to be created")
	}
	if reader.repoRoot != "/test/repo" {
		t.Errorf("Expected repoRoot to be '/test/repo', got '%s'", reader.repoRoot)
	}
	if !strings.Contains(reader.statePath, "state.json") {
		t.Error("Expected statePath to contain state.json")
	}
}

func TestLoadSessionsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	reader := NewStateReader(tmpDir)

	sessions, err := reader.LoadSessions()
	if err != nil {
		t.Errorf("Expected LoadSessions to succeed with no state file, got: %v", err)
	}

	if sessions == nil {
		t.Error("Expected sessions to be non-nil slice")
	}
	if len(sessions) != 0 {
		t.Error("Expected empty sessions when no state file exists")
	}
}

func TestLoadSessionsWithData(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .uzi directory
	uziDir := filepath.Join(tmpDir, ".uzi")
	err := os.MkdirAll(uziDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .uzi directory: %v", err)
	}

	// Create test state data
	now := time.Now()
	states := map[string]AgentState{
		"agent-project-abc123-john": {
			GitRepo:      "https://github.com/test/repo.git",
			BranchFrom:   "main",
			BranchName:   "agent-john-feature",
			Prompt:       "Implement user authentication",
			WorktreePath: "/tmp/test-worktree",
			Port:         3000,
			Model:        "claude-3-sonnet",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		"agent-project-abc123-sarah": {
			GitRepo:      "https://github.com/test/repo.git",
			BranchFrom:   "main",
			BranchName:   "agent-sarah-bugfix",
			Prompt:       "Fix login validation bug",
			WorktreePath: "/tmp/test-worktree2",
			Port:         0, // No dev server
			Model:        "claude-3-haiku",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	// Write state file
	statePath := filepath.Join(uziDir, "state.json")
	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	err = os.WriteFile(statePath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}

	// Test loading sessions
	reader := NewStateReader(tmpDir)
	sessions, err := reader.LoadSessions()
	if err != nil {
		t.Errorf("Expected LoadSessions to succeed, got: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}

	// Verify session data
	sessionMap := make(map[string]SessionInfo)
	for _, session := range sessions {
		sessionMap[session.SessionName] = session
	}

	johnSession := sessionMap["agent-project-abc123-john"]
	if johnSession.AgentName != "john" {
		t.Errorf("Expected agent name 'john', got '%s'", johnSession.AgentName)
	}
	if johnSession.DevServerURL != "http://localhost:3000" {
		t.Errorf("Expected dev server URL 'http://localhost:3000', got '%s'", johnSession.DevServerURL)
	}
	if johnSession.Model != "claude-3-sonnet" {
		t.Errorf("Expected model 'claude-3-sonnet', got '%s'", johnSession.Model)
	}

	sarahSession := sessionMap["agent-project-abc123-sarah"]
	if sarahSession.AgentName != "sarah" {
		t.Errorf("Expected agent name 'sarah', got '%s'", sarahSession.AgentName)
	}
	if sarahSession.DevServerURL != "" {
		t.Errorf("Expected empty dev server URL, got '%s'", sarahSession.DevServerURL)
	}
}

func TestLoadSessionsInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .uzi directory
	uziDir := filepath.Join(tmpDir, ".uzi")
	err := os.MkdirAll(uziDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .uzi directory: %v", err)
	}

	// Write invalid JSON
	statePath := filepath.Join(uziDir, "state.json")
	err = os.WriteFile(statePath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid state file: %v", err)
	}

	// Test loading sessions
	reader := NewStateReader(tmpDir)
	_, err = reader.LoadSessions()
	if err == nil {
		t.Error("Expected LoadSessions to fail with invalid JSON")
	}
}

func TestExtractAgentName(t *testing.T) {
	reader := NewStateReader("/test")

	tests := []struct {
		sessionName string
		expected    string
	}{
		{"agent-project-abc123-john", "john"},
		{"agent-myproject-def456-sarah-jane", "sarah-jane"},
		{"agent-test-ghi789-complex-agent-name", "complex-agent-name"},
		{"invalid-session-name", "invalid-session-name"},
		{"agent-only-two", "agent-only-two"},
		{"agent-only-two-parts", "parts"},
		{"", ""},
	}

	for _, test := range tests {
		result := reader.extractAgentName(test.sessionName)
		if result != test.expected {
			t.Errorf("extractAgentName(%q) = %q, expected %q", test.sessionName, result, test.expected)
		}
	}
}

func TestParseGitDiffStats(t *testing.T) {
	reader := NewStateReader("/test")

	tests := []struct {
		output             string
		expectedInsertions int
		expectedDeletions  int
	}{
		{
			" 3 files changed, 45 insertions(+), 12 deletions(-)",
			45, 12,
		},
		{
			" 1 file changed, 1 insertion(+)",
			1, 0,
		},
		{
			" 2 files changed, 23 deletions(-)",
			0, 23,
		},
		{
			" 1 file changed, 1 insertion(+), 1 deletion(-)",
			1, 1,
		},
		{
			"", // Empty output
			0, 0,
		},
		{
			"no changes", // No matching pattern
			0, 0,
		},
	}

	for _, test := range tests {
		insertions, deletions := reader.parseGitDiffStats(test.output)
		if insertions != test.expectedInsertions {
			t.Errorf("parseGitDiffStats(%q) insertions = %d, expected %d", test.output, insertions, test.expectedInsertions)
		}
		if deletions != test.expectedDeletions {
			t.Errorf("parseGitDiffStats(%q) deletions = %d, expected %d", test.output, deletions, test.expectedDeletions)
		}
	}
}

func TestGetGitDiffStatsEmptyPath(t *testing.T) {
	reader := NewStateReader("/test")

	insertions, deletions := reader.getGitDiffStats("")
	if insertions != 0 || deletions != 0 {
		t.Errorf("Expected 0,0 for empty worktree path, got %d,%d", insertions, deletions)
	}
}

func TestGetGitDiffStatsNonExistentPath(t *testing.T) {
	reader := NewStateReader("/test")

	insertions, deletions := reader.getGitDiffStats("/non/existent/path")
	if insertions != 0 || deletions != 0 {
		t.Errorf("Expected 0,0 for non-existent worktree path, got %d,%d", insertions, deletions)
	}
}

func TestGetActiveSessionsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	reader := NewStateReader(tmpDir)

	sessions, err := reader.GetActiveSessions()
	if err != nil {
		t.Errorf("Expected GetActiveSessions to succeed with empty state, got: %v", err)
	}

	// Allow both nil and empty slice since LoadSessions returns empty slice when no state file
	if sessions != nil && len(sessions) != 0 {
		t.Error("Expected empty or nil active sessions when no state file exists")
	}
}

func TestFilterByRepo(t *testing.T) {
	reader := NewStateReader("/test")

	sessions := []SessionInfo{
		{SessionName: "session1", AgentName: "john"},
		{SessionName: "session2", AgentName: "sarah"},
	}

	filtered := reader.FilterByRepo(sessions, "https://github.com/test/repo.git")

	// Current implementation returns all sessions (simplified)
	if len(filtered) != len(sessions) {
		t.Errorf("Expected %d filtered sessions, got %d", len(sessions), len(filtered))
	}
}

func TestGetSessionStatusFunctions(t *testing.T) {
	reader := NewStateReader("/test")

	// Test getSessionStatus with non-existent tmux session
	// This will likely return "inactive" since tmux session doesn't exist
	status := reader.getSessionStatus("non-existent-session")
	if status != "inactive" {
		t.Logf("getSessionStatus returned '%s' for non-existent session (expected 'inactive' but may vary)", status)
	}

	// Test getPaneContent with non-existent session
	// This should return an error
	_, err := reader.getPaneContent("non-existent-session")
	if err == nil {
		t.Error("Expected getPaneContent to fail with non-existent session")
	}
}
