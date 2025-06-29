package agents

import (
	"strings"
	"testing"
)

func TestGetRandomAgent(t *testing.T) {
	// Test that GetRandomAgent returns a non-empty string
	agent := GetRandomAgent()
	if agent == "" {
		t.Error("Expected GetRandomAgent to return non-empty string")
	}

	// Test that returned agent is in the list
	agents := strings.Split(strings.TrimSpace(AgentNames), "\n")
	found := false
	for _, a := range agents {
		if a == agent {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected agent '%s' to be in the agent list", agent)
	}
}

func TestGetRandomAgentVariety(t *testing.T) {
	// Test that GetRandomAgent returns different values over multiple calls
	// This test might occasionally fail due to randomness, but very unlikely
	agents := make(map[string]bool)
	iterations := 50

	for i := 0; i < iterations; i++ {
		agent := GetRandomAgent()
		agents[agent] = true
	}

	// We should get at least 10 different agents in 50 iterations
	// This is a probabilistic test - it might rarely fail
	if len(agents) < 10 {
		t.Errorf("Expected at least 10 different agents in %d iterations, got %d", iterations, len(agents))
	}
}

func TestAgentNamesList(t *testing.T) {
	// Test that AgentNames is not empty
	if AgentNames == "" {
		t.Error("Expected AgentNames to be non-empty")
	}

	// Test that AgentNames contains expected structure
	agents := strings.Split(strings.TrimSpace(AgentNames), "\n")
	if len(agents) == 0 {
		t.Error("Expected AgentNames to contain at least one agent")
	}

	// Test that all agent names are non-empty and trimmed
	for i, agent := range agents {
		if agent == "" {
			t.Errorf("Expected agent at index %d to be non-empty", i)
		}
		if strings.TrimSpace(agent) != agent {
			t.Errorf("Expected agent at index %d to be trimmed, got '%s'", i, agent)
		}
	}
}

func TestSpecificAgentNamesExist(t *testing.T) {
	// Test that some expected agent names exist in the list
	expectedAgents := []string{
		"john", "emily", "michael", "sarah", "david",
		"mila", "stephen", "nicole", "ryan",
	}

	agents := strings.Split(strings.TrimSpace(AgentNames), "\n")
	agentSet := make(map[string]bool)
	for _, agent := range agents {
		agentSet[agent] = true
	}

	for _, expected := range expectedAgents {
		if !agentSet[expected] {
			t.Errorf("Expected agent '%s' to be in the agent list", expected)
		}
	}
}

func TestAgentNamesCount(t *testing.T) {
	// Test that we have a reasonable number of agent names
	agents := strings.Split(strings.TrimSpace(AgentNames), "\n")
	if len(agents) < 50 {
		t.Errorf("Expected at least 50 agent names, got %d", len(agents))
	}

	// Test that we don't have too many (to catch obvious errors)
	if len(agents) > 200 {
		t.Errorf("Expected fewer than 200 agent names, got %d", len(agents))
	}
}

func TestAgentNamesNoDuplicatesInFirstFew(t *testing.T) {
	// Test that the first few agent names don't have obvious duplicates
	// (We know there are some duplicates in the full list, but shouldn't be many)
	agents := strings.Split(strings.TrimSpace(AgentNames), "\n")

	seen := make(map[string]bool)
	duplicates := 0

	// Check first 20 agents for duplicates
	checkCount := 20
	if len(agents) < checkCount {
		checkCount = len(agents)
	}

	for i := 0; i < checkCount; i++ {
		agent := agents[i]
		if seen[agent] {
			duplicates++
		}
		seen[agent] = true
	}

	// Allow for a few duplicates but not too many
	if duplicates > 3 {
		t.Errorf("Expected fewer than 4 duplicates in first %d agents, got %d", checkCount, duplicates)
	}
}

func TestAgentNamesFormat(t *testing.T) {
	// Test that agent names follow expected format (lowercase, alphabetic)
	agents := strings.Split(strings.TrimSpace(AgentNames), "\n")

	for i, agent := range agents {
		// Should be lowercase
		if strings.ToLower(agent) != agent {
			t.Errorf("Expected agent at index %d to be lowercase, got '%s'", i, agent)
		}

		// Should be alphabetic (no numbers or special characters)
		for _, char := range agent {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
				t.Errorf("Expected agent at index %d to contain only letters, got '%s'", i, agent)
				break
			}
		}

		// Should have reasonable length
		if len(agent) < 2 {
			t.Errorf("Expected agent at index %d to have at least 2 characters, got '%s'", i, agent)
		}
		if len(agent) > 15 {
			t.Errorf("Expected agent at index %d to have at most 15 characters, got '%s'", i, agent)
		}
	}
}
