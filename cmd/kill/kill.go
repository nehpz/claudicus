package kill

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
)

var (
	fs        = flag.NewFlagSet("uzi kill", flag.ExitOnError)
	CmdKill = &ffcli.Command{
		Name:       "kill",
		ShortUsage: "uzi kill <agent-name>",
		ShortHelp:  "Delete tmux session and git worktree for the specified agent",
		FlagSet:    fs,
		Exec:       executeKill,
	}
)

func executeKill(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("agent name argument is required")
	}

	agentName := args[0]
	log.Debug("Deleting tmux session and git worktree", "agent", agentName)

	// Get the current git hash
	gitHashCmd := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
	gitHashCmd.Dir = filepath.Dir(os.Args[0])
	gitHashOutput, err := gitHashCmd.Output()
	if err != nil {
		log.Error("Error getting git hash", "error", err)
		return err
	}
	gitHash := strings.TrimSpace(string(gitHashOutput))

	// Prefix the tmux session name with the git hash
	sessionName := fmt.Sprintf("agent-%s-%s", gitHash, agentName)

	// Kill tmux session if it exists
	checkSession := exec.CommandContext(ctx, "tmux", "has-session", "-t", sessionName)
	if err := checkSession.Run(); err == nil {
		// Session exists, kill it
		killCmd := exec.CommandContext(ctx, "tmux", "kill-session", "-t", sessionName)
		if err := killCmd.Run(); err != nil {
			log.Error("Error killing tmux session", "session", sessionName, "error", err)
		} else {
			log.Debug("Killed tmux session", "session", sessionName)
		}
	}

	// Remove git worktree
	worktreePath := filepath.Join(filepath.Dir(os.Args[0]), "..", agentName)
	if _, err := os.Stat(worktreePath); err == nil {
		// Worktree exists, remove it
		removeCmd := exec.CommandContext(ctx, "git", "worktree", "remove", "--force", "../"+agentName)
		removeCmd.Dir = filepath.Dir(os.Args[0])
		if err := removeCmd.Run(); err != nil {
			log.Error("Error removing git worktree", "path", worktreePath, "error", err)
		} else {
			log.Debug("Removed git worktree", "path", worktreePath)
		}

		// Delete the branch
		deleteBranchCmd := exec.CommandContext(ctx, "git", "branch", "-D", agentName)
		deleteBranchCmd.Dir = filepath.Dir(os.Args[0])
		if err := deleteBranchCmd.Run(); err != nil {
			log.Error("Error deleting git branch", "branch", agentName, "error", err)
		} else {
			log.Debug("Deleted git branch", "branch", agentName)
		}
	}

	fmt.Printf("Deleted agent: %s\n", agentName)
	return nil
} 
