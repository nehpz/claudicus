package prompt

import (
	"testing"
)

func TestParseAgents(t *testing.T) {
	tests := []struct {
		name          string
		agentsStr     string
		expected      map[string]AgentConfig
		expectErr     bool
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
				"gemini": {Command: "gemini-cli", Count: 3},
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
		{"gemini", "gemini-cli"},
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
