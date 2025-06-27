package prompt

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nehpz/claudicus/pkg/agents"
	"github.com/nehpz/claudicus/pkg/config"
	"github.com/nehpz/claudicus/pkg/state"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type AgentConfig struct {
	Command string
	Count   int
}

var (
	fs         = flag.NewFlagSet("uzi prompt", flag.ExitOnError)
	agentsFlag = fs.String("agents", "claude:1", "agents to run with their commands and counts (e.g., 'claude:1,codex:2'). Use 'random' as agent name to select a random agent name.")
	configPath = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	CmdPrompt  = &ffcli.Command{
		Name:       "prompt",
		ShortUsage: "uzi prompt --agents=AGENT:COUNT[,AGENT:COUNT...] prompt text...",
		ShortHelp:  "Run the prompt command with specified agents and counts",
		FlagSet:    fs,
		Exec:       executePrompt,
	}
)

// parseAgents parses the agents flag value into a map of agent configs
func parseAgents(agentsStr string) (map[string]AgentConfig, error) {
	agentConfigs := make(map[string]AgentConfig)

	// Split by comma for multiple agent configurations
	agentPairs := strings.Split(agentsStr, ",")

	for _, pair := range agentPairs {
		// Split by colon for agent:count
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid agent format: %s (expected agent:count)", pair)
		}

		agent := strings.TrimSpace(parts[0])
		countStr := strings.TrimSpace(parts[1])

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, fmt.Errorf("invalid count for agent %s: %s", agent, countStr)
		}

		if count < 1 {
			return nil, fmt.Errorf("count must be at least 1 for agent %s", agent)
		}

		// Map agent names to actual commands
		command := getCommandForAgent(agent)
		agentConfigs[agent] = AgentConfig{
			Command: command,
			Count:   count,
		}
	}

	return agentConfigs, nil
}

// getCommandForAgent maps agent names to their actual CLI commands
func getCommandForAgent(agent string) string {
	switch agent {
	case "claude":
		return "claude"
	case "cursor":
		return "cursor"
	case "codex":
		return "codex"
	case "gemini":
		return "gemini"
	case "random":
		return "claude" // Default for random agents
	default:
		// For unknown agents, assume the agent name is the command
		return agent
	}
}


// isPortAvailable checks if a port is available for use
func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// getExistingSessionPorts reads the state file and returns all currently assigned ports
func getExistingSessionPorts(stateManager *state.StateManager) ([]int, error) {
	if stateManager == nil {
		return []int{}, nil
	}
	
	// Read the state file
	stateFile := stateManager.GetStatePath()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No state file exists yet, return empty list
			return []int{}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}
	
	// Parse the state file
	states := make(map[string]state.AgentState)
	if err := json.Unmarshal(data, &states); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}
	
	// Extract all assigned ports
	var existingPorts []int
	for _, agentState := range states {
		if agentState.Port > 0 {
			existingPorts = append(existingPorts, agentState.Port)
		}
	}
	
	return existingPorts, nil
}

// findAvailablePort finds the first available port in the given range, excluding already assigned ports
func findAvailablePort(startPort, endPort int, assignedPorts []int) (int, error) {
	for port := startPort; port <= endPort; port++ {
		// Check if port is already assigned in this execution
		alreadyAssigned := false
		for _, assignedPort := range assignedPorts {
			if port == assignedPort {
				alreadyAssigned = true
				break
			}
		}
		if alreadyAssigned {
			continue
		}

		// Check if port is actually available
		if isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range %d-%d", startPort, endPort)
}

func executePrompt(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("prompt argument is required")
	}

	// Load config - uzi.yaml is required for standardized dev environment setup
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		return fmt.Errorf("uzi.yaml configuration file is required but could not be loaded: %w\n\nThe uzi.yaml file is critical for:\n1. Standardizing the development environment setup\n2. Providing an available range of ports for the application\n\nPlease create a uzi.yaml file with:\n  devCommand: your-dev-command --port $PORT\n  portRange: 3000-3010", err)
	}
	if cfg.DevCommand == nil || *cfg.DevCommand == "" {
		return fmt.Errorf("devCommand is required in uzi.yaml for standardized development environment setup")
	}
	if cfg.PortRange == nil || *cfg.PortRange == "" {
		return fmt.Errorf("portRange is required in uzi.yaml to define available ports for agent sessions")
	}

	promptText := strings.Join(args, " ")
	log.Debug("Running prompt command", "prompt", promptText)

	// Load existing session ports to prevent collisions with existing agents
	stateManager := state.NewStateManager()
	existingPorts, err := getExistingSessionPorts(stateManager)
	if err != nil {
		log.Warn("Failed to load existing session ports, proceeding without collision check", "error", err)
		existingPorts = []int{}
	}
	
	// Track assigned ports to prevent collisions between iterations and with existing sessions
	assignedPorts := existingPorts

	// Parse agents
	agentConfigs, err := parseAgents(*agentsFlag)
	if err != nil {
		return fmt.Errorf("error parsing agents: %s", err)
	}

	for agent, config := range agentConfigs {
		for i := 0; i < config.Count; i++ {
			// Always get a random agent name for the session/branch/worktree names
			randomAgentName := agents.GetRandomAgent()

			// Use the specified agent for the command (unless it's "random")
			commandToUse := config.Command
			if agent == "random" {
				// If agent is "random", use the random name for the command too
				commandToUse = randomAgentName
			}

			fmt.Printf("%s: %s: %s\n", randomAgentName, commandToUse, promptText)

			// Check if git worktree exists
			// Get the current git hash
			gitHashCmd := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
			gitHashCmd.Dir = filepath.Dir(os.Args[0])
			gitHashOutput, err := gitHashCmd.Output()
			if err != nil {
				log.Error("Error getting git hash", "error", err)
				continue
			}
			gitHash := strings.TrimSpace(string(gitHashOutput))

			// Get the git repository name from remote URL
			gitRemoteCmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
			gitRemoteCmd.Dir = filepath.Dir(os.Args[0])
			gitRemoteOutput, err := gitRemoteCmd.Output()
			if err != nil {
				log.Error("Error getting git remote", "error", err)
				continue
			}
			remoteURL := strings.TrimSpace(string(gitRemoteOutput))
			// Extract repository name from URL (handle both https and ssh formats)
			repoName := filepath.Base(remoteURL)
			projectDir := strings.TrimSuffix(repoName, ".git")

			// Create unique identifier using timestamp and iteration
			timestamp := time.Now().Unix()
			uniqueId := fmt.Sprintf("%d-%d", timestamp, i)

			// Create unique branch and worktree names using the random agent name
			branchName := fmt.Sprintf("%s-%s-%s-%s", randomAgentName, projectDir, gitHash, uniqueId)
			worktreeName := fmt.Sprintf("%s-%s-%s-%s", randomAgentName, projectDir, gitHash, uniqueId)

			// Prefix the tmux session name with the git hash and use random agent name
			sessionName := fmt.Sprintf("agent-%s-%s-%s", projectDir, gitHash, randomAgentName)

			// Get home directory for worktree storage
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Error("Error getting home directory", "error", err)
				continue
			}

			worktreesDir := filepath.Join(homeDir, ".local", "share", "uzi", "worktrees")
			if err := os.MkdirAll(worktreesDir, 0755); err != nil {
				log.Error("Error creating worktrees directory", "error", err)
				continue
			}

			worktreePath := filepath.Join(worktreesDir, worktreeName)
			var selectedPort int
			// Create git worktree
			cmd := fmt.Sprintf("git worktree add -b %s %s", branchName, worktreePath)
			cmdExec := exec.CommandContext(ctx, "sh", "-c", cmd)
			cmdExec.Dir = filepath.Dir(os.Args[0])
			if err := cmdExec.Run(); err != nil {
				log.Error("Error creating git worktree", "command", cmd, "error", err)
				continue
			}

			// Create tmux session
			cmd = fmt.Sprintf("tmux new-session -d -s %s -c %s", sessionName, worktreePath)
			cmdExec = exec.CommandContext(ctx, "sh", "-c", cmd)
			if err := cmdExec.Run(); err != nil {
				log.Error("Error creating tmux session", "command", cmd, "error", err)
				continue
			}

			// Rename the first window to "agent"
			renameCmd := fmt.Sprintf("tmux rename-window -t %s:0 agent", sessionName)
			renameExec := exec.CommandContext(ctx, "sh", "-c", renameCmd)
			if err := renameExec.Run(); err != nil {
				log.Error("Error renaming tmux window", "command", renameCmd, "error", err)
				continue
			}

			// Create uzi-dev pane and run dev command if configured
			if cfg.DevCommand == nil || *cfg.DevCommand == "" || cfg.PortRange == nil || *cfg.PortRange == "" {
				// Hit enter in the agent pane
				hitEnterCmd := fmt.Sprintf("tmux send-keys -t %s:agent C-m", sessionName)
				hitEnterExec := exec.CommandContext(ctx, "sh", "-c", hitEnterCmd)
				if err := hitEnterExec.Run(); err != nil {
					log.Error("Error hitting enter in tmux", "command", hitEnterCmd, "error", err)
				}

				// Always run send-keys command to the agent pane
			var tmuxCmd string
			if commandToUse == "gemini" {
				tmuxCmd = fmt.Sprintf("tmux send-keys -t %s:agent '%s -p \"%%s\"' C-m", sessionName, commandToUse)
			} else {
				tmuxCmd = fmt.Sprintf("tmux send-keys -t %s:agent '%s \"%%s\"' C-m", sessionName, commandToUse)
			}
				tmuxCmdExec := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(tmuxCmd, promptText))
				tmuxCmdExec.Dir = worktreePath
				if err := tmuxCmdExec.Run(); err != nil {
					log.Error("Error sending keys to tmux", "command", tmuxCmd, "error", err)
					continue
				}

				// Save state before continuing (no port since dev server not started)
				stateManager := state.NewStateManager()
				if stateManager != nil {
					if err := stateManager.SaveState(promptText, branchName, sessionName, worktreePath, commandToUse); err != nil {
						log.Error("Error saving state", "error", err)
					}
				}
				continue
			}

			ports := strings.Split(*cfg.PortRange, "-")
			if len(ports) != 2 {
				log.Warn("Invalid port range format in config", "portRange", *cfg.PortRange)
				continue
			}

			startPort, _ := strconv.Atoi(ports[0])
			endPort, _ := strconv.Atoi(ports[1])
			if startPort <= 0 || endPort <= 0 || endPort < startPort {
				log.Warn("Invalid port range in config", "portRange", *cfg.PortRange)
				continue
			}

			selectedPort, err = findAvailablePort(startPort, endPort, assignedPorts)
			if err != nil {
				log.Error("Error finding available port", "error", err)
				continue
			}

			devCmdTemplate := *cfg.DevCommand
			devCmd := strings.Replace(devCmdTemplate, "$PORT", strconv.Itoa(selectedPort), 1)

			// Create new window named uzi-dev
			newWindowCmd := fmt.Sprintf("tmux new-window -t %s -n uzi-dev -c %s", sessionName, worktreePath)
			newWindowExec := exec.CommandContext(ctx, "sh", "-c", newWindowCmd)
			if err := newWindowExec.Run(); err != nil {
				log.Error("Error creating new tmux window for dev server", "command", newWindowCmd, "error", err)
				continue
			}

			// Send dev command to the new window
			sendDevCmd := fmt.Sprintf("tmux send-keys -t %s:uzi-dev '%s' C-m", sessionName, devCmd)
			sendDevCmdExec := exec.CommandContext(ctx, "sh", "-c", sendDevCmd)
			if err := sendDevCmdExec.Run(); err != nil {
				log.Error("Error sending dev command to tmux", "command", sendDevCmd, "error", err)
			}

			// Hit enter in the agent pane
			hitEnterCmd := fmt.Sprintf("tmux send-keys -t %s:agent C-m", sessionName)
			hitEnterExec := exec.CommandContext(ctx, "sh", "-c", hitEnterCmd)
			if err := hitEnterExec.Run(); err != nil {
				log.Error("Error hitting enter in tmux", "command", hitEnterCmd, "error", err)
			}

			assignedPorts = append(assignedPorts, selectedPort)

			// Always run send-keys command to the agent pane
            var tmuxCmd string
            if commandToUse == "gemini" {
                tmuxCmd = fmt.Sprintf("tmux send-keys -t %s:agent '%s -p \"%%s\"' C-m", sessionName, commandToUse)
            } else {
                tmuxCmd = fmt.Sprintf("tmux send-keys -t %s:agent '%s \"%%s\"' C-m", sessionName, commandToUse)
            }
			tmuxCmdExec := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(tmuxCmd, promptText))
			tmuxCmdExec.Dir = worktreePath
			if err := tmuxCmdExec.Run(); err != nil {
				log.Error("Error sending keys to tmux", "command", tmuxCmd, "error", err)
				continue
			}

			// Save state after successful prompt execution
			stateManager := state.NewStateManager()
			if stateManager != nil {
				if err := stateManager.SaveStateWithPort(promptText, branchName, sessionName, worktreePath, commandToUse, selectedPort); err != nil {
					log.Error("Error saving state", "error", err)
				}
			}
		}
	}

	return nil
}
