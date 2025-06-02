package kill

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/devflowinc/uzi/pkg/state"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs      = flag.NewFlagSet("uzi kill", flag.ExitOnError)
	CmdKill = &ffcli.Command{
		Name:       "kill",
		ShortUsage: "uzi kill [<agent-name>|all]",
		ShortHelp:  "Delete tmux session and git worktree for the specified agent",
		FlagSet:    fs,
		Exec:       executeKill,
	}
)

// killSession kills a single session and cleans up its associated resources
func killSession(ctx context.Context, sessionName, agentName string, sm *state.StateManager) error {
	log.Debug("Deleting tmux session and git worktree", "session", sessionName, "agent", agentName)

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
		// Get worktree path from state
		worktreeInfo, err := sm.GetWorktreeInfo(sessionName)
		if err != nil {
			log.Error("Error getting worktree info", "session", sessionName, "error", err)
			return fmt.Errorf("failed to get worktree info: %w", err)
		}

		// First, remove the worktree
		removeCmd := exec.CommandContext(ctx, "git", "worktree", "remove", "--force", worktreeInfo.WorktreePath)
		removeCmd.Dir = filepath.Dir(os.Args[0])
		if err := removeCmd.Run(); err != nil {
			log.Error("Error removing git worktree", "path", worktreeInfo.WorktreePath, "error", err)
			return fmt.Errorf("failed to remove git worktree: %w", err)
		}
		log.Debug("Removed git worktree", "path", worktreeInfo.WorktreePath)

		// Then delete the branch
		deleteBranchCmd := exec.CommandContext(ctx, "git", "branch", "-D", agentName)
		deleteBranchCmd.Dir = filepath.Dir(os.Args[0])
		if err := deleteBranchCmd.Run(); err != nil {
			log.Error("Error deleting git branch", "branch", agentName, "error", err)
			return fmt.Errorf("failed to delete git branch: %w", err)
		}
		log.Debug("Deleted git branch", "branch", agentName)
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
		worktreeStatePath := filepath.Join(homeDir, ".local", "share", "uzi", "worktree", sessionName)
		if _, err := os.Stat(worktreeStatePath); err == nil {
			if err := os.RemoveAll(worktreeStatePath); err != nil {
				log.Error("Error removing worktree state", "path", worktreeStatePath, "error", err)
			} else {
				log.Debug("Removed worktree state", "path", worktreeStatePath)
			}
		}

		// Remove from state.json
		if err := sm.RemoveState(sessionName); err != nil {
			log.Error("Error removing state entry", "session", sessionName, "error", err)
		} else {
			log.Debug("Removed state entry", "session", sessionName)
		}
	}

	return nil
}

// killAll kills all sessions for the current git repository
func killAll(ctx context.Context, sm *state.StateManager) error {
	log.Debug("Deleting all agents for repository")

	// Get active sessions from state
	activeSessions, err := sm.GetActiveSessionsForRepo()
	if err != nil {
		log.Error("Error getting active sessions", "error", err)
		return err
	}

	if len(activeSessions) == 0 {
		fmt.Println("No active sessions found")
		return nil
	}

	killedCount := 0
	for _, sessionName := range activeSessions {
		// Extract agent name from session name (assuming format: repo-agentName)
		parts := strings.Split(sessionName, "-")
		if len(parts) < 2 {
			log.Warn("Unexpected session name format", "session", sessionName)
			continue
		}
		agentName := parts[len(parts)-1] // Get the last part as agent name

		if err := killSession(ctx, sessionName, agentName, sm); err != nil {
			log.Error("Error killing session", "session", sessionName, "error", err)
			continue
		}

		killedCount++
		fmt.Printf("Deleted agent: %s\n", agentName)
	}

	fmt.Printf("Successfully deleted %d agent(s)\n", killedCount)
	return nil
}

func executeKill(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("agent name argument is required")
	}

	agentName := args[0]

	// Get state manager to read from config
	sm := state.NewStateManager()
	if sm == nil {
		return fmt.Errorf("could not initialize state manager")
	}

	// Handle "all" case
	if agentName == "all" {
		return killAll(ctx, sm)
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

	// Kill the specific session
	if err := killSession(ctx, sessionToKill, agentName, sm); err != nil {
		return err
	}

	fmt.Printf("Deleted agent: %s\n", agentName)
	return nil
}
