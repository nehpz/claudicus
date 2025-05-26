package checkpoint

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
	fs           = flag.NewFlagSet("uzi checkpoint", flag.ExitOnError)
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

	// Get current directory (should be the main worktree)
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	// Check if agent worktree exists
	agentWorktreePath := filepath.Join(filepath.Dir(currentDir), agentName)
	if _, err := os.Stat(agentWorktreePath); os.IsNotExist(err) {
		return fmt.Errorf("agent worktree does not exist: %s", agentWorktreePath)
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
	checkBranchCmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+agentName)
	checkBranchCmd.Dir = currentDir
	if err := checkBranchCmd.Run(); err != nil {
		return fmt.Errorf("agent branch does not exist: %s", agentName)
	}

	// Get the base commit where the agent branch diverged
	mergeBaseCmd := exec.CommandContext(ctx, "git", "merge-base", currentBranch, agentName)
	mergeBaseCmd.Dir = currentDir
	mergeBaseOutput, err := mergeBaseCmd.Output()
	if err != nil {
		return fmt.Errorf("error finding merge base: %v", err)
	}
	mergeBase := strings.TrimSpace(string(mergeBaseOutput))

	// Check if there are any changes to rebase
	diffCmd := exec.CommandContext(ctx, "git", "rev-list", "--count", mergeBase+".."+agentName)
	diffCmd.Dir = currentDir
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return fmt.Errorf("error checking for changes: %v", err)
	}
	changeCount := strings.TrimSpace(string(diffOutput))

	if changeCount == "0" {
		fmt.Printf("No changes to checkpoint from agent: %s\n", agentName)
		return nil
	}

	fmt.Printf("Checkpointing %s commits from agent: %s\n", changeCount, agentName)

	// Rebase the agent branch onto the current branch
	rebaseCmd := exec.CommandContext(ctx, "git", "rebase", agentName)
	rebaseCmd.Dir = currentDir
	rebaseCmd.Stdout = os.Stdout
	rebaseCmd.Stderr = os.Stderr
	if err := rebaseCmd.Run(); err != nil {
		return fmt.Errorf("error rebasing agent changes: %v", err)
	}

	fmt.Printf("Successfully checkpointed changes from agent: %s\n", agentName)

	// Stage all changes and commit
	addCmd := exec.CommandContext(ctx, "git", "add", ".")
	addCmd.Dir = currentDir
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("error staging changes: %v", err)
	}

	commitCmd := exec.CommandContext(ctx, "git", "commit", "-am", commitMessage)
	commitCmd.Dir = currentDir
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}

	fmt.Printf("Successfully committed changes with message: %s\n", commitMessage)
	return nil
}
