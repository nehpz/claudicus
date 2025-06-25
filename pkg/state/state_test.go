package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewStateManager(t *testing.T) {
	sm := NewStateManager()
	if sm == nil {
		t.Error("Expected StateManager to be created")
	}
	if sm.statePath == "" {
		t.Error("Expected statePath to be set")
	}
	if !strings.Contains(sm.statePath, "state.json") {
		t.Error("Expected statePath to contain state.json")
	}
}

func TestGetStatePath(t *testing.T) {
	sm := NewStateManager()
	path := sm.GetStatePath()
	if path == "" {
		t.Error("Expected non-empty state path")
	}
	if !strings.HasSuffix(path, "state.json") {
		t.Error("Expected state path to end with state.json")
	}
}

func TestEnsureStateDir(t *testing.T) {
	// Create a temporary state manager with a temp path
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "test", "state.json"),
	}

	err := sm.ensureStateDir()
	if err != nil {
		t.Errorf("Expected ensureStateDir to succeed, got: %v", err)
	}

	// Check that directory was created
	dir := filepath.Dir(sm.statePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Expected state directory to be created")
	}
}

func TestSaveAndRemoveState(t *testing.T) {
	// Create temporary state manager
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}

	// Test saving state
	err := sm.SaveState("test prompt", "test-branch", "test-session", "/test/path", "test-model")
	if err != nil {
		t.Errorf("Expected SaveState to succeed, got: %v", err)
	}

	// Verify state file exists
	if _, err := os.Stat(sm.statePath); os.IsNotExist(err) {
		t.Error("Expected state file to be created")
	}

	// Test loading and verifying state
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		t.Errorf("Expected to read state file, got: %v", err)
	}

	var states map[string]AgentState
	err = json.Unmarshal(data, &states)
	if err != nil {
		t.Errorf("Expected to parse state JSON, got: %v", err)
	}

	state, exists := states["test-session"]
	if !exists {
		t.Error("Expected test-session to exist in state")
	}

	if state.Prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got '%s'", state.Prompt)
	}
	if state.BranchName != "test-branch" {
		t.Errorf("Expected branch 'test-branch', got '%s'", state.BranchName)
	}
	if state.WorktreePath != "/test/path" {
		t.Errorf("Expected worktree path '/test/path', got '%s'", state.WorktreePath)
	}
	if state.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", state.Model)
	}

	// Test removing state
	err = sm.RemoveState("test-session")
	if err != nil {
		t.Errorf("Expected RemoveState to succeed, got: %v", err)
	}

	// Verify state was removed
	data, err = os.ReadFile(sm.statePath)
	if err != nil {
		t.Errorf("Expected to read state file after removal, got: %v", err)
	}

	states = make(map[string]AgentState) // Reset the map
	err = json.Unmarshal(data, &states)
	if err != nil {
		t.Errorf("Expected to parse state JSON after removal, got: %v", err)
	}

	if _, exists := states["test-session"]; exists {
		t.Error("Expected test-session to be removed from state")
	}
}

func TestSaveStateWithPort(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}

	err := sm.SaveStateWithPort("test prompt", "test-branch", "test-session", "/test/path", "test-model", 3000)
	if err != nil {
		t.Errorf("Expected SaveStateWithPort to succeed, got: %v", err)
	}

	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		t.Errorf("Expected to read state file, got: %v", err)
	}

	var states map[string]AgentState
	err = json.Unmarshal(data, &states)
	if err != nil {
		t.Errorf("Expected to parse state JSON, got: %v", err)
	}

	state := states["test-session"]
	if state.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", state.Port)
	}
}

func TestGetWorktreeInfo(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}

	// First save a state
	err := sm.SaveState("test prompt", "test-branch", "test-session", "/test/path", "test-model")
	if err != nil {
		t.Errorf("Expected SaveState to succeed, got: %v", err)
	}

	// Test getting worktree info
	info, err := sm.GetWorktreeInfo("test-session")
	if err != nil {
		t.Errorf("Expected GetWorktreeInfo to succeed, got: %v", err)
	}

	if info.Prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got '%s'", info.Prompt)
	}
	if info.BranchName != "test-branch" {
		t.Errorf("Expected branch 'test-branch', got '%s'", info.BranchName)
	}

	// Test getting info for non-existent session
	_, err = sm.GetWorktreeInfo("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent session")
	}
}

func TestRemoveStateNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "non-existent.json"),
	}

	// Should not error when trying to remove from non-existent file
	err := sm.RemoveState("test-session")
	if err != nil {
		t.Errorf("Expected RemoveState to handle non-existent file gracefully, got: %v", err)
	}
}

func TestGetActiveSessionsForRepoNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "non-existent.json"),
	}

	sessions, err := sm.GetActiveSessionsForRepo()
	if err != nil {
		t.Errorf("Expected GetActiveSessionsForRepo to handle non-existent file gracefully, got: %v", err)
	}

	if sessions == nil {
		t.Error("Expected empty slice, not nil")
	}
	if len(sessions) != 0 {
		t.Error("Expected empty sessions for non-existent file")
	}
}

func TestAgentStateTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}

	// Save initial state
	before := time.Now()
	err := sm.SaveState("test prompt", "test-branch", "test-session", "/test/path", "test-model")
	if err != nil {
		t.Errorf("Expected SaveState to succeed, got: %v", err)
	}
	after := time.Now()

	// Verify timestamps
	info, err := sm.GetWorktreeInfo("test-session")
	if err != nil {
		t.Errorf("Expected GetWorktreeInfo to succeed, got: %v", err)
	}

	if info.CreatedAt.Before(before) || info.CreatedAt.After(after) {
		t.Error("CreatedAt timestamp should be within expected range")
	}
	if info.UpdatedAt.Before(before) || info.UpdatedAt.After(after) {
		t.Error("UpdatedAt timestamp should be within expected range")
	}

	// Update the same session
	time.Sleep(100 * time.Millisecond) // Ensure time difference
	beforeUpdate := time.Now()
	err = sm.SaveState("updated prompt", "test-branch", "test-session", "/test/path", "test-model")
	if err != nil {
		t.Errorf("Expected SaveState update to succeed, got: %v", err)
	}
	afterUpdate := time.Now()

	// Verify CreatedAt stayed the same but UpdatedAt changed
	updatedInfo, err := sm.GetWorktreeInfo("test-session")
	if err != nil {
		t.Errorf("Expected GetWorktreeInfo to succeed, got: %v", err)
	}

	if !updatedInfo.CreatedAt.Equal(info.CreatedAt) {
		t.Error("CreatedAt should not change on update")
	}
	if updatedInfo.UpdatedAt.Before(beforeUpdate) || updatedInfo.UpdatedAt.After(afterUpdate) {
		t.Error("UpdatedAt should be updated to recent time")
	}
}

func TestGetGitRepo(t *testing.T) {
	sm := NewStateManager()
	if sm == nil {
		t.Fatal("Expected StateManager to be created")
	}
	
	// This will likely return empty string since we're not in a git repo or git may not be configured
	repo := sm.getGitRepo()
	// We can't assert specific values since it depends on the environment
	// Just verify it doesn't panic and returns a string
	if repo != "" {
		t.Logf("getGitRepo returned: %s", repo)
	}
}

func TestGetBranchFrom(t *testing.T) {
	sm := NewStateManager()
	if sm == nil {
		t.Fatal("Expected StateManager to be created")
	}
	
	// This will likely return "main" as fallback since we may not have git symbolic-ref set up
	branch := sm.getBranchFrom()
	if branch == "" {
		t.Error("Expected getBranchFrom to return non-empty string")
	}
	// Default fallback should be "main"
	if branch != "main" {
		t.Logf("getBranchFrom returned: %s", branch)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	sm := NewStateManager()
	if sm == nil {
		t.Fatal("Expected StateManager to be created")
	}
	
	// This will likely return empty string since we may not be in a git repo
	branch := sm.getCurrentBranch()
	// We can't assert specific values since it depends on the environment
	if branch != "" {
		t.Logf("getCurrentBranch returned: %s", branch)
	}
}

func TestIsActiveInTmux(t *testing.T) {
	sm := NewStateManager()
	if sm == nil {
		t.Fatal("Expected StateManager to be created")
	}
	
	// Test with non-existent session (should return false)
	active := sm.isActiveInTmux("non-existent-session")
	if active {
		t.Error("Expected non-existent session to not be active in tmux")
	}
}

func TestStoreWorktreeBranch(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a StateManager that will write to temp directory
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}
	
	// This will likely fail since we're not in a git repo, but test that it doesn't crash
	err := sm.storeWorktreeBranch("test-session")
	// We don't assert the error since it depends on git environment
	// Just verify the function doesn't panic
	if err != nil {
		t.Logf("storeWorktreeBranch returned error (expected): %v", err)
	}
}

func TestGetActiveSessionsForRepoWithRepo(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}
	
	// Save a test state first
	err := sm.SaveState("test prompt", "test-branch", "test-session", "/test/path", "test-model")
	if err != nil {
		t.Errorf("Expected SaveState to succeed, got: %v", err)
	}
	
	// Test getting active sessions
	// This will likely return empty since tmux sessions don't exist and git repo may not match
	sessions, err := sm.GetActiveSessionsForRepo()
	if err != nil {
		t.Errorf("Expected GetActiveSessionsForRepo to succeed, got: %v", err)
	}
	
	// Allow both nil and empty slice (function returns empty slice)
	if sessions == nil {
		t.Log("GetActiveSessionsForRepo returned nil (expected empty slice but nil is also acceptable)")
	}
	// Length may be 0 since tmux sessions don't exist or repo doesn't match
}

func TestRemoveStateCorruptedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}
	
	// Write corrupted JSON
	err := os.WriteFile(sm.statePath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted JSON: %v", err)
	}
	
	// Try to remove state - should fail due to corrupted JSON
	err = sm.RemoveState("test-session")
	if err == nil {
		t.Error("Expected RemoveState to fail with corrupted JSON")
	}
}

func TestGetWorktreeInfoCorruptedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}
	
	// Write corrupted JSON
	err := os.WriteFile(sm.statePath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted JSON: %v", err)
	}
	
	// Try to get worktree info - should fail due to corrupted JSON
	_, err = sm.GetWorktreeInfo("test-session")
	if err == nil {
		t.Error("Expected GetWorktreeInfo to fail with corrupted JSON")
	}
}

func TestSaveStateExistingCorruptedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	sm := &StateManager{
		statePath: filepath.Join(tmpDir, "state.json"),
	}
	
	// Write corrupted JSON
	err := os.WriteFile(sm.statePath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted JSON: %v", err)
	}
	
	// Try to save state - should succeed by overwriting the corrupted file
	err = sm.SaveState("test prompt", "test-branch", "test-session", "/test/path", "test-model")
	if err != nil {
		t.Errorf("Expected SaveState to succeed even with corrupted existing file, got: %v", err)
	}
	
	// Verify the file is now valid
	info, err := sm.GetWorktreeInfo("test-session")
	if err != nil {
		t.Errorf("Expected GetWorktreeInfo to succeed after SaveState, got: %v", err)
	}
	
	if info.Prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got '%s'", info.Prompt)
	}
}