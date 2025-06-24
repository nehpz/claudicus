// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"testing"
	"time"
)

// Unit tests for UziCLI proxy infrastructure
// These tests have no external dependencies and test individual components in isolation

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

	if config.EnableCache != false {
		t.Errorf("Expected EnableCache false, got %v", config.EnableCache)
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
		{"", ""},
		{"agent", "agent"},
		{"agent-", "agent-"},
		{"agent-a-b", "agent-a-b"}, // Less than 4 parts
	}
	
	for _, test := range tests {
		result := extractAgentName(test.sessionName)
		if result != test.expected {
			t.Errorf("extractAgentName(%q) = %q, expected %q", 
				test.sessionName, result, test.expected)
		}
	}
}

func TestUziCLIInterfaceCompliance(t *testing.T) {
	// Verify that UziCLI implements UziInterface at compile time
	var _ UziInterface = (*UziCLI)(nil)
	
	// Test that we can create an instance 
	cli := NewUziCLI()
	if cli == nil {
		t.Fatal("Expected UziCLI instance, got nil")
	}
	
	// Test that state manager is initialized
	if cli.stateManager == nil {
		t.Error("Expected state manager to be initialized")
	}
	
	// Test that tmux discovery is initialized
	if cli.tmuxDiscovery == nil {
		t.Error("Expected tmux discovery to be initialized")
	}
}

func TestRefreshSessions(t *testing.T) {
	// RefreshSessions should never fail as it's a no-op in current implementation
	cli := NewUziCLI()
	err := cli.RefreshSessions()
	if err != nil {
		t.Errorf("RefreshSessions should never fail, got: %v", err)
	}
}

func TestConfigCustomization(t *testing.T) {
	tests := []struct {
		name   string
		config ProxyConfig
	}{
		{
			name: "high timeout",
			config: ProxyConfig{
				Timeout:     60 * time.Second,
				Retries:     5,
				LogLevel:    "debug",
				EnableCache: true,
			},
		},
		{
			name: "low timeout",
			config: ProxyConfig{
				Timeout:     5 * time.Second,
				Retries:     0,
				LogLevel:    "error",
				EnableCache: false,
			},
		},
		{
			name: "zero values",
			config: ProxyConfig{
				Timeout:     0,
				Retries:     0,
				LogLevel:    "",
				EnableCache: false,
			},
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli := NewUziCLIWithConfig(test.config)
			
			if cli.config.Timeout != test.config.Timeout {
				t.Errorf("Expected timeout %v, got %v", test.config.Timeout, cli.config.Timeout)
			}
			
			if cli.config.Retries != test.config.Retries {
				t.Errorf("Expected retries %d, got %d", test.config.Retries, cli.config.Retries)
			}
			
			if cli.config.LogLevel != test.config.LogLevel {
				t.Errorf("Expected log level %q, got %q", test.config.LogLevel, cli.config.LogLevel)
			}
			
			if cli.config.EnableCache != test.config.EnableCache {
				t.Errorf("Expected EnableCache %v, got %v", test.config.EnableCache, cli.config.EnableCache)
			}
		})
	}
}

// Mock error for testing
type MockError struct {
	msg string
}

func (e *MockError) Error() string {
	return e.msg
}

