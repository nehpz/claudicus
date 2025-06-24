// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"testing"
	"time"
)

func TestProxyConfig(t *testing.T) {
	config := DefaultProxyConfig()
	
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout of 30s, got %v", config.Timeout)
	}
	
	if config.Retries != 2 {
		t.Errorf("Expected 2 retries, got %d", config.Retries)
	}
	
	if config.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got %s", config.LogLevel)
	}
}

func TestUziCLICreation(t *testing.T) {
	// Test default creation
	cli := NewUziCLI()
	if cli == nil {
		t.Fatal("Expected UziCLI instance, got nil")
	}
	
	if cli.config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout, got %v", cli.config.Timeout)
	}
	
	// Test custom config creation
	customConfig := ProxyConfig{
		Timeout:  10 * time.Second,
		Retries:  1,
		LogLevel: "debug",
	}
	
	customCLI := NewUziCLIWithConfig(customConfig)
	if customCLI == nil {
		t.Fatal("Expected UziCLI instance with custom config, got nil")
	}
	
	if customCLI.config.Timeout != 10*time.Second {
		t.Errorf("Expected custom timeout of 10s, got %v", customCLI.config.Timeout)
	}
	
	if customCLI.config.LogLevel != "debug" {
		t.Errorf("Expected debug log level, got %s", customCLI.config.LogLevel)
	}
}

func TestErrorWrapping(t *testing.T) {
	cli := NewUziCLI()
	
	// Test nil error
	wrapped := cli.wrapError("test", nil)
	if wrapped != nil {
		t.Errorf("Expected nil for nil error, got %v", wrapped)
	}
	
	// Test error wrapping
	originalErr := &MockError{msg: "original error"}
	wrapped = cli.wrapError("TestOperation", originalErr)
	
	expectedMsg := "uzi_proxy: TestOperation: original error"
	if wrapped.Error() != expectedMsg {
		t.Errorf("Expected %q, got %q", expectedMsg, wrapped.Error())
	}
}

func TestExtractAgentName(t *testing.T) {
	tests := []struct {
		sessionName string
		expected    string
	}{
		{"agent-project-abc123-claude", "claude"},
		{"agent-myproject-def456-gpt4", "gpt4"},
		{"agent-test-789xyz-custom-agent-name", "custom-agent-name"},
		{"invalid-session-name", "invalid-session-name"},
		{"agent-only-three-parts", "parts"},
	}
	
	for _, test := range tests {
		result := extractAgentName(test.sessionName)
		if result != test.expected {
			t.Errorf("extractAgentName(%q) = %q, expected %q", 
				test.sessionName, result, test.expected)
		}
	}
}

// Mock error for testing
type MockError struct {
	msg string
}

func (e *MockError) Error() string {
	return e.msg
}

func TestUziCLIInterface(t *testing.T) {
	// Verify that UziCLI implements UziInterface
	var _ UziInterface = (*UziCLI)(nil)
	
	// Test that we can create an instance and call interface methods
	cli := NewUziCLI()
	
	// Test GetSessions method exists (will fail without actual uzi command, but that's expected)
	_, err := cli.GetSessions()
	if err == nil {
		t.Log("GetSessions completed successfully (probably no sessions)")
	} else {
		// Expected to fail in test environment without actual uzi command
		if err.Error() == "" {
			t.Error("Expected error message, got empty string")
		}
		t.Logf("GetSessions failed as expected in test environment: %v", err)
	}
	
	// Test error cases
	err = cli.RefreshSessions()
	if err != nil {
		t.Errorf("RefreshSessions should never fail, got: %v", err)
	}
}

func TestProxyConsistency(t *testing.T) {
	// Verify that all proxy methods follow the consistent pattern
	cli := NewUziCLI()
	
	// These should all return wrapped errors in test environment
	methods := []func() error{
		func() error { _, err := cli.GetSessions(); return err },
		func() error { return cli.KillSession("test-session") },
		func() error { return cli.RunPrompt("claude:1", "test prompt") },
		func() error { return cli.RunBroadcast("test message") },
		func() error { return cli.RunCommand("echo test") },
	}
	
	for i, method := range methods {
		err := method()
		if err != nil {
			// All errors should be wrapped with "uzi_proxy:" prefix
			if !containsSubstring(err.Error(), "uzi_proxy:") {
				t.Errorf("Method %d error should be wrapped with uzi_proxy prefix, got: %v", i, err)
			}
		}
	}
}

// Helper function to check if string contains substring
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}
