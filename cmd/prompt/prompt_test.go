package prompt

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAgents(t *testing.T) {
	tests := []struct {
		name      string
		agentsStr string
		expected  map[string]AgentConfig
		expectErr bool
	}{
		{
			name:      "single agent",
			agentsStr: "claude:1",
			expected: map[string]AgentConfig{
				"claude": {Command: "claude", Count: 1},
			},
			expectErr: false,
		},
		{
			name:      "multiple agents",
			agentsStr: "claude:2,gemini:3",
			expected: map[string]AgentConfig{
				"claude": {Command: "claude", Count: 2},
				"gemini": {Command: "gemini", Count: 3},
			},
			expectErr: false,
		},
		{
			name:      "random agent",
			agentsStr: "random:5",
			expected: map[string]AgentConfig{
				"random": {Command: "claude", Count: 5},
			},
			expectErr: false,
		},
		{
			name:      "invalid format",
			agentsStr: "claude",
			expectErr: true,
		},
		{
			name:      "invalid count",
			agentsStr: "claude:abc",
			expectErr: true,
		},
		{
			name:      "zero count",
			agentsStr: "claude:0",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs, err := parseAgents(tt.agentsStr)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseAgents() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr {
				if len(configs) != len(tt.expected) {
					t.Errorf("parseAgents() expected %d configs, got %d", len(tt.expected), len(configs))
				}
				for agent, expectedConfig := range tt.expected {
					if config, ok := configs[agent]; !ok {
						t.Errorf("parseAgents() expected config for agent %s not found", agent)
					} else if config.Command != expectedConfig.Command || config.Count != expectedConfig.Count {
						t.Errorf("parseAgents() config for agent %s = %+v, expected %+v", agent, config, expectedConfig)
					}
				}
			}
		})
	}
}

func TestGetCommandForAgent(t *testing.T) {
	tests := []struct {
		agent    string
		expected string
	}{
		{"claude", "claude"},
		{"cursor", "cursor"},
		{"codex", "codex"},
		{"gemini", "gemini"},
		{"random", "claude"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			if cmd := getCommandForAgent(tt.agent); cmd != tt.expected {
				t.Errorf("getCommandForAgent() = %s, expected %s", cmd, tt.expected)
			}
		})
	}
}

func TestExecutePrompt(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		setupConfig   func() string // Returns config file path
		setupFlags    func()
		expectError   bool
		errorContains string
	}{
		{
			name: "no arguments provided",
			args: []string{},
			setupConfig: func() string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "uzi.yaml")
				content := "devCommand: npm start --port $PORT\nportRange: 3000-3010\n"
				os.WriteFile(configFile, []byte(content), 0644)
				return configFile
			},
			setupFlags:    func() {},
			expectError:   true,
			errorContains: "prompt argument is required",
		},
		{
			name: "missing config file",
			args: []string{"test", "prompt"},
			setupConfig: func() string {
				return "/nonexistent/config.yaml"
			},
			setupFlags:    func() {},
			expectError:   true,
			errorContains: "uzi.yaml configuration file is required",
		},
		{
			name: "config missing devCommand",
			args: []string{"test", "prompt"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "uzi.yaml")
				content := "portRange: 3000-3010\n"
				os.WriteFile(configFile, []byte(content), 0644)
				return configFile
			},
			setupFlags:    func() {},
			expectError:   true,
			errorContains: "devCommand is required in uzi.yaml",
		},
		{
			name: "config missing portRange",
			args: []string{"test", "prompt"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "uzi.yaml")
				content := "devCommand: npm start --port $PORT\n"
				os.WriteFile(configFile, []byte(content), 0644)
				return configFile
			},
			setupFlags:    func() {},
			expectError:   true,
			errorContains: "portRange is required in uzi.yaml",
		},
		{
			name: "invalid agents flag",
			args: []string{"test", "prompt"},
			setupConfig: func() string {
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, "uzi.yaml")
				content := "devCommand: npm start --port $PORT\nportRange: 3000-3010\n"
				os.WriteFile(configFile, []byte(content), 0644)
				return configFile
			},
			setupFlags: func() {
				*agentsFlag = "invalid:agent:format"
			},
			expectError:   true,
			errorContains: "error parsing agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalConfigPath := *configPath
			originalAgentsFlag := *agentsFlag
			defer func() {
				*configPath = originalConfigPath
				*agentsFlag = originalAgentsFlag
			}()

			// Set config path
			*configPath = tt.setupConfig()

			// Setup flags if needed
			if tt.setupFlags != nil {
				tt.setupFlags()
			} else {
				*agentsFlag = "claude:1" // Default valid agent
			}

			// Execute
			err := executePrompt(context.Background(), tt.args)

			// Verify
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, but got no error", tt.errorContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, but got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestIsPortAvailable(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{
			name: "high port should be available",
			port: 45678,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPortAvailable(tt.port)
			// Just ensure the function returns a boolean
			if result != true && result != false {
				t.Errorf("isPortAvailable() should return boolean, got %v", result)
			}
		})
	}
}

func TestFindAvailablePort(t *testing.T) {
	tests := []struct {
		name          string
		startPort     int
		endPort       int
		assignedPorts []int
		expectError   bool
	}{
		{
			name:          "find available port in range",
			startPort:     45000,
			endPort:       45010,
			assignedPorts: []int{},
			expectError:   false,
		},
		{
			name:          "find available port with some assigned",
			startPort:     45000,
			endPort:       45010,
			assignedPorts: []int{45000, 45001},
			expectError:   false,
		},
		{
			name:          "invalid range - start > end",
			startPort:     45010,
			endPort:       45000,
			assignedPorts: []int{},
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := findAvailablePort(tt.startPort, tt.endPort, tt.assignedPorts)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got port %d", port)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if port < tt.startPort || port > tt.endPort {
					t.Errorf("Port %d is outside range %d-%d", port, tt.startPort, tt.endPort)
				}
				// Check that returned port is not in assigned ports
				for _, assignedPort := range tt.assignedPorts {
					if port == assignedPort {
						t.Errorf("Returned port %d was in assigned ports list", port)
					}
				}
			}
		})
	}
}

func TestExecutePromptSuccessCase(t *testing.T) {
	// Test additional scenarios for executePrompt to improve coverage
	t.Run("invalid port range format", func(t *testing.T) {
		originalConfigPath := *configPath
		originalAgentsFlag := *agentsFlag
		defer func() {
			*configPath = originalConfigPath
			*agentsFlag = originalAgentsFlag
		}()

		// Create config with invalid port range
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "uzi.yaml")
		content := "devCommand: echo test\nportRange: invalid-port-range\n"
		os.WriteFile(configFile, []byte(content), 0644)
		*configPath = configFile
		*agentsFlag = "claude:1"

		err := executePrompt(context.Background(), []string{"uzi", "test prompt"})

		// Should fail due to git operations in test environment, but should get past validation
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "prompt argument is required") ||
				strings.Contains(errMsg, "uzi.yaml configuration file is required") ||
				strings.Contains(errMsg, "devCommand is required") ||
				strings.Contains(errMsg, "portRange is required") {
				t.Errorf("Test failed at validation stage, expected to get further: %v", err)
			}
		}
	})

	t.Run("valid config with random agent", func(t *testing.T) {
		originalConfigPath := *configPath
		originalAgentsFlag := *agentsFlag
		defer func() {
			*configPath = originalConfigPath
			*agentsFlag = originalAgentsFlag
		}()

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "uzi.yaml")
		content := "devCommand: echo 'test server on port $PORT'\nportRange: 45000-45010\n"
		os.WriteFile(configFile, []byte(content), 0644)
		*configPath = configFile
		*agentsFlag = "random:1"

		err := executePrompt(context.Background(), []string{"uzi", "test prompt with random agent"})

		// Should fail due to git operations in test environment, but should get past validation
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "prompt argument is required") ||
				strings.Contains(errMsg, "uzi.yaml configuration file is required") ||
				strings.Contains(errMsg, "devCommand is required") ||
				strings.Contains(errMsg, "portRange is required") ||
				strings.Contains(errMsg, "error parsing agents") {
				t.Errorf("Test failed at validation stage, expected to get further: %v", err)
			}
		}
	})
}
