package prompt

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
	"uzi/pkg/agents"
	"uzi/pkg/state"
)

var (
	fs        = flag.NewFlagSet("uzi prompt", flag.ExitOnError)
	count     = fs.Int("count", 1, "number of times to run the command")
	command   = fs.String("command", "claude", "command to execute")
	CmdPrompt = &ffcli.Command{
		Name:       "prompt",
		ShortUsage: "uzi prompt -count=N -command=CMD prompt text...",
		ShortHelp:  "Run the prompt command with a given prompt",
		FlagSet:    fs,
		Exec:       executePrompt,
	}
)

func executePrompt(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("prompt argument is required")
	}

	prompt := strings.Join(args, " ")
	log.Debug("Running prompt command", "prompt", prompt, "count", *count, "command", *command)

	for i := 0; i < *count; i++ {
		agentName := agents.GetRandomAgent()
		fmt.Printf("%s: claude: %s\n", agentName, *command)

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

		// Prefix the tmux session name with the git hash
		sessionName := fmt.Sprintf("agent-%s-%s", gitHash, agentName)

		worktreePath := filepath.Join(filepath.Dir(os.Args[0]), "..", agentName)
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			cmd := fmt.Sprintf("git worktree add -b %s ../%s", agentName, agentName)
			cmdExec := exec.CommandContext(ctx, "sh", "-c", cmd)
			cmdExec.Dir = filepath.Dir(os.Args[0])
			if err := cmdExec.Run(); err != nil {
				log.Error("Error creating git worktree", "command", cmd, "error", err)
				continue
			}
		}

		// Check if tmux session exists
		checkSession := exec.CommandContext(ctx, "tmux", "has-session", "-t", sessionName)
		if err := checkSession.Run(); err != nil {
			// Session doesn't exist, create it
			cmd := fmt.Sprintf("tmux new-session -d -s %s", sessionName)
			cmdExec := exec.CommandContext(ctx, "sh", "-c", cmd)
			cmdExec.Dir = worktreePath
			if err := cmdExec.Run(); err != nil {
				log.Error("Error creating tmux session", "command", cmd, "error", err)
				continue
			}
		}

		// Always run send-keys command
		cmd := fmt.Sprintf("tmux send-keys -t %s '%s \"%s\"' C-m", sessionName, *command, prompt)
		cmdExec := exec.CommandContext(ctx, "sh", "-c", cmd)
		cmdExec.Dir = worktreePath
		if err := cmdExec.Run(); err != nil {
			log.Error("Error sending keys to tmux", "command", cmd, "error", err)
			continue
		}

		// Save state after successful prompt execution
		stateManager := state.NewStateManager()
		if stateManager != nil {
			if err := stateManager.SaveState(prompt, sessionName); err != nil {
				log.Error("Error saving state", "error", err)
			}
		}
	}

	return nil
}
