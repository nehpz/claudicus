package prompt

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
	"uzi/pkg/config"
)

var (
	fs         = flag.NewFlagSet("uzi prompt", flag.ExitOnError)
	configPath = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	CmdPrompt  = &ffcli.Command{
		Name:       "prompt",
		ShortUsage: "uzi prompt <prompt>",
		ShortHelp:  "Run the prompt command with a given prompt",
		FlagSet:    fs,
		Exec:       executePrompt,
	}
)

// getRandomAgent reads agent names from agents.txt and returns a random agent name.
func getRandomAgent() (string, error) {
	file, err := os.Open("agents.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	var agents []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		agents = append(agents, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if len(agents) == 0 {
		return "", fmt.Errorf("no agents found in file")
	}

	rand.Seed(time.Now().UnixNano())
	return agents[rand.Intn(len(agents))], nil
}

func executePrompt(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("prompt argument is required")
	}

	prompt := args[0]
	log.Info("Running prompt command", "prompt", prompt)

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Error("Error loading config, using default config", "error", err)
		cfg = &config.Config{}
	}

	for _, agent := range cfg.Agents {
		fmt.Printf("- Command: %s (Count: %s)\n", agent.Command, agent.Name)
		// Execute commands

		agentName := agent.Name

		// Check if git worktree exists
		worktreePath := filepath.Join(filepath.Dir(os.Args[0]), "..", agentName)
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			cmd := fmt.Sprintf("git worktree add -b %s ../%s", agentName, agentName)
			command := exec.CommandContext(ctx, "sh", "-c", cmd)
			command.Dir = filepath.Dir(os.Args[0])
			if err := command.Run(); err != nil {
				log.Error("Error creating git worktree", "command", cmd, "error", err)
				continue
			} else {
				log.Error("Error creating git worktree", "command", cmd, "error", err)
				continue
			}
		}

		// Check if tmux session exists
		checkSession := exec.CommandContext(ctx, "tmux", "has-session", "-t", agentName)
		if err := checkSession.Run(); err != nil {
			// Session doesn't exist, create it
			cmd := fmt.Sprintf("tmux new-session -d -s %s", agentName)
			command := exec.CommandContext(ctx, "sh", "-c", cmd)
			command.Dir = worktreePath
			if err := command.Run(); err != nil {
				log.Error("Error creating tmux session", "command", cmd, "error", err)
				continue
			}
		}

		// Always run send-keys command
		cmd := fmt.Sprintf("tmux send-keys -t %s 'claude %s' C-m", agentName, prompt)
		command := exec.CommandContext(ctx, "sh", "-c", cmd)
		command.Dir = worktreePath
		if err := command.Run(); err != nil {
			log.Error("Error sending keys to tmux", "command", cmd, "error", err)
			continue
		}
	}

	return nil
}
