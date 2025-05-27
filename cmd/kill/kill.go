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
	"uzi/pkg/state"
)

var (
	fs        = flag.NewFlagSet("uzi kill", flag.ExitOnError)
	CmdKill = &ffcli.Command{
		Name:       "kill",
		ShortUsage: "uzi kill [<agent-name>|all]",
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
	if agentName == "all" {
		return fmt.Errorf("Not implemented: delete all agents")
	}
	log.Debug("Deleting tmux session and git worktree", "agent", agentName)

	// Get state manager to read from config
	sm := state.NewStateManager()
	if sm == nil {
		return fmt.Errorf("could not initialize state manager")
	}

	// Get active sessions from state
	activeSessions, err := sm.GetActiveSessionsForRepo()
	if err != nil {
		log.Error("Error getting active sessions", "error", err)
		return err
	}

	// Find the session with the matching agent name
	var sessionToKill string
	for _, session := range activeSessions {
		if strings.HasSuffix(session, "-"+agentName) {
			sessionToKill = session
			break
		}
	}

	if sessionToKill == "" {
		log.Debug("No active tmux session found for agent", "agent", agentName)
		return fmt.Errorf("no active session found for agent: %s", agentName)
	}

	// Kill tmux session if it exists
	checkSession := exec.CommandContext(ctx, "tmux", "has-session", "-t", sessionToKill)
	if err := checkSession.Run(); err == nil {
		// Session exists, kill it
		killCmd := exec.CommandContext(ctx, "tmux", "kill-session", "-t", sessionToKill)
		if err := killCmd.Run(); err != nil {
			log.Error("Error killing tmux session", "session", sessionToKill, "error", err)
		} else {
			log.Debug("Killed tmux session", "session", sessionToKill)
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

	// Delete from config store (~/.local/share/uzi/)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		// Remove worktree directory from config store
		configWorktreePath := filepath.Join(homeDir, ".local", "share", "uzi", "worktrees", agentName)
		if _, err := os.Stat(configWorktreePath); err == nil {
			if err := os.RemoveAll(configWorktreePath); err != nil {
				log.Error("Error removing config worktree", "path", configWorktreePath, "error", err)
			} else {
				log.Debug("Removed config worktree", "path", configWorktreePath)
			}
		}

		// Remove worktree state directory
		worktreeStatePath := filepath.Join(homeDir, ".local", "share", "uzi", "worktree", sessionToKill)
		if _, err := os.Stat(worktreeStatePath); err == nil {
			if err := os.RemoveAll(worktreeStatePath); err != nil {
				log.Error("Error removing worktree state", "path", worktreeStatePath, "error", err)
			} else {
				log.Debug("Removed worktree state", "path", worktreeStatePath)
			}
		}

		// Remove from state.json
		if err := sm.RemoveState(sessionToKill); err != nil {
			log.Error("Error removing state entry", "session", sessionToKill, "error", err)
		} else {
			log.Debug("Removed state entry", "session", sessionToKill)
		}
	}

	fmt.Printf("Deleted agent: %s\n", agentName)
	return nil
}
