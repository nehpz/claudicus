package checkpoint

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nehpz/claudicus/pkg/state"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs            = flag.NewFlagSet("uzi checkpoint", flag.ExitOnError)
	CmdCheckpoint = &ffcli.Command{
		Name:       "checkpoint",
		ShortUsage: "uzi checkpoint <agent-name> <commit-message>",
		ShortHelp:  "Rebase changes from an agent worktree into the current worktree and commit",
		FlagSet:    fs,
		Exec:       executeCheckpoint,
	}
)

func executeCheckpoint(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("agent name and commit message arguments are required")
	}

	agentName := args[0]
	commitMessage := args[1]
	log.Debug("Checkpointing changes from agent", "agent", agentName)

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
	var sessionToCheckpoint string
	for _, session := range activeSessions {
		// Extract agent name from session name (format: agent-projectDir-gitHash-agentName)
		parts := strings.Split(session, "-")
		if len(parts) >= 4 && parts[0] == "agent" {
			// Join all parts after the first 3 (in case agent name contains hyphens)
			sessionAgentName := strings.Join(parts[3:], "-")
			if sessionAgentName == agentName {
				sessionToCheckpoint = session
				break
			}
		}
	}

	if sessionToCheckpoint == "" {
		return fmt.Errorf("no active session found for agent: %s", agentName)
	}

	// Get session state to find worktree path
	states := make(map[string]state.AgentState)
	if data, err := os.ReadFile(sm.GetStatePath()); err != nil {
		return fmt.Errorf("error reading state file: %v", err)
	} else {
		if err := json.Unmarshal(data, &states); err != nil {
			return fmt.Errorf("error parsing state file: %v", err)
		}
	}

	sessionState, ok := states[sessionToCheckpoint]
	if !ok || sessionState.WorktreePath == "" {
		return fmt.Errorf("invalid state for session: %s", sessionToCheckpoint)
	}

	// Get the actual branch name from the state
	agentBranchName := sessionState.BranchName

	// Get current directory (should be the main worktree)
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	// Get the current branch name in the main worktree
	getCurrentBranchCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	getCurrentBranchCmd.Dir = currentDir
	currentBranchOutput, err := getCurrentBranchCmd.Output()
	if err != nil {
		return fmt.Errorf("error getting current branch: %v", err)
	}
	currentBranch := strings.TrimSpace(string(currentBranchOutput))

	// Check if agent branch exists
	checkBranchCmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+agentBranchName)
	checkBranchCmd.Dir = currentDir
	if err := checkBranchCmd.Run(); err != nil {
		return fmt.Errorf("agent branch does not exist: %s", agentBranchName)
	}

	// Stage all changes and commit on the agent branch
	addCmd := exec.CommandContext(ctx, "git", "add", ".")
	addCmd.Dir = sessionState.WorktreePath
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("error staging changes: %v", err)
	}

	commitCmd := exec.CommandContext(ctx, "git", "commit", "-am", commitMessage)
	commitCmd.Dir = sessionState.WorktreePath
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		log.Warn("No unstaged changes to commit, rebasing")
	}

	// Get the base commit where the agent branch diverged
	mergeBaseCmd := exec.CommandContext(ctx, "git", "merge-base", currentBranch, agentBranchName)
	mergeBaseCmd.Dir = currentDir
	mergeBaseOutput, err := mergeBaseCmd.Output()
	if err != nil {
		return fmt.Errorf("error finding merge base: %v", err)
	}
	mergeBase := strings.TrimSpace(string(mergeBaseOutput))

	// Check if there are any changes to rebase
	diffCmd := exec.CommandContext(ctx, "git", "rev-list", "--count", mergeBase+".."+agentBranchName)
	diffCmd.Dir = currentDir
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return fmt.Errorf("error checking for changes: %v", err)
	}
	changeCount := strings.TrimSpace(string(diffOutput))

	fmt.Printf("Checkpointing %s commits from agent: %s\n", changeCount, agentName)

	// Rebase the agent branch onto the current branch
	rebaseCmd := exec.CommandContext(ctx, "git", "rebase", agentBranchName)
	rebaseCmd.Dir = currentDir
	rebaseCmd.Stdout = os.Stdout
	rebaseCmd.Stderr = os.Stderr
	if err := rebaseCmd.Run(); err != nil {
		return fmt.Errorf("error rebasing agent changes: %v", err)
	}

	fmt.Printf("Successfully checkpointed changes from agent: %s\n", agentName)
	fmt.Printf("Successfully committed changes with message: %s\n", commitMessage)
	return nil
}
