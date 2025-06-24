// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"testing"
)

// Integration tests for UziCLI proxy
// These tests involve executing actual CLI commands and require the uzi binary to be available

func TestUziCLIInterface(t *testing.T) {
	// Verify that UziCLI implements UziInterface at compile time
	var _ UziInterface = (*UziCLI)(nil)
	
	// Test that we can create an instance and call interface methods
	cli := NewUziCLI()
	if cli == nil {
		t.Fatal("Expected UziCLI instance, got nil")
	}
	
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

