package prompt

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"uzi/pkg/agents"
	"uzi/pkg/config"
	"uzi/pkg/state"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs         = flag.NewFlagSet("uzi prompt", flag.ExitOnError)
	count      = fs.Int("count", 1, "number of times to run the command")
	command    = fs.String("command", "claude", "command to execute")
	configPath = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	CmdPrompt  = &ffcli.Command{
		Name:       "prompt",
		ShortUsage: "uzi prompt -count=N -command=CMD prompt text...",
		ShortHelp:  "Run the prompt command with a given prompt",
		FlagSet:    fs,
		Exec:       executePrompt,
	}
)

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

	// Load config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Warn("Error loading config, using default values", "error", err)
		cfg = &config.Config{} // Use default or empty config
	}
	if cfg.DevCommand == nil || *cfg.DevCommand == "" {
		log.Info("Dev command not set in config, skipping dev server startup.")
	}
	if cfg.PortRange == nil || *cfg.PortRange == "" {
		log.Info("Port range not set in config, skipping dev server startup.")
	}

	promptText := strings.Join(args, " ")
	log.Debug("Running prompt command", "prompt", promptText, "count", *count, "command", *command)

	// Track assigned ports to prevent collisions between iterations
	var assignedPorts []int

	for i := 0; i < *count; i++ {
		agentName := agents.GetRandomAgent()
		fmt.Printf("%s: %s: %s\n", agentName, *command, promptText)

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

		// Create unique branch and worktree names
		branchName := fmt.Sprintf("%s-%s-%s-%s", agentName, projectDir, gitHash, uniqueId)
		worktreeName := fmt.Sprintf("%s-%s-%s-%s", agentName, projectDir, gitHash, uniqueId)

		// Prefix the tmux session name with the git hash
		sessionName := fmt.Sprintf("agent-%s-%s-%s", projectDir, gitHash, agentName)

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
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			cmd := fmt.Sprintf("git worktree add -b %s %s", branchName, worktreePath)
			cmdExec := exec.CommandContext(ctx, "sh", "-c", cmd)
			cmdExec.Dir = filepath.Dir(os.Args[0])
			if err := cmdExec.Run(); err != nil {
				log.Error("Error creating git worktree", "command", cmd, "error", err)
				continue
			}

			// Check if tmux session exists
			checkSession := exec.CommandContext(ctx, "tmux", "has-session", "-t", sessionName)
			if err := checkSession.Run(); err != nil {
				// Session doesn't exist, create it
				cmd := fmt.Sprintf("tmux new-session -d -s %s -c %s", sessionName, worktreePath)
				cmdExec := exec.CommandContext(ctx, "sh", "-c", cmd)
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
				if cfg.DevCommand != nil && *cfg.DevCommand != "" && cfg.PortRange != nil && *cfg.PortRange != "" {
					ports := strings.Split(*cfg.PortRange, "-")
					if len(ports) == 2 {
						startPort, _ := strconv.Atoi(ports[0])
						endPort, _ := strconv.Atoi(ports[1])
						if startPort > 0 && endPort > 0 && endPort >= startPort {
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
							} else {
								// Send dev command to the new window
								sendDevCmd := fmt.Sprintf("tmux send-keys -t %s:uzi-dev '%s' C-m", sessionName, devCmd)
								sendDevCmdExec := exec.CommandContext(ctx, "sh", "-c", sendDevCmd)
								if err := sendDevCmdExec.Run(); err != nil {
									log.Error("Error sending dev command to tmux", "command", sendDevCmd, "error", err)
								}

								// Send dev command to the new window
								hitEnterCmd := fmt.Sprintf("tmux send-keys -t %s:agent C-m", sessionName)
								hitEnterExec := exec.CommandContext(ctx, "sh", "-c", hitEnterCmd)
								if err := hitEnterExec.Run(); err != nil {
									log.Error("Error hitting enter in tmux", "command", hitEnterCmd, "error", err)
								}

							}
							assignedPorts = append(assignedPorts, selectedPort)
						} else {
							log.Warn("Invalid port range in config", "portRange", *cfg.PortRange)
						}
					} else {
						log.Warn("Invalid port range format in config", "portRange", *cfg.PortRange)
					}
				}
			}
		}

		// Always run send-keys command to the agent pane
		tmuxCmd := fmt.Sprintf("tmux send-keys -t %s:agent '%s \"%%s\"' C-m", sessionName, *command)
		cmdExec := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(tmuxCmd, promptText))
		cmdExec.Dir = worktreePath
		if err := cmdExec.Run(); err != nil {
			log.Error("Error sending keys to tmux", "command", tmuxCmd, "error", err)
			continue
		}

		// Save state after successful prompt execution
		stateManager := state.NewStateManager()
		if stateManager != nil {
			if err := stateManager.SaveStateWithStatus(promptText, branchName, sessionName, worktreePath, "Loading", selectedPort); err != nil {
				log.Error("Error saving state", "error", err)
			}
		}
	}

	return nil
}
