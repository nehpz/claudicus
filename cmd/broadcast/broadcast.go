package broadcast

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/nehpz/claudicus/pkg/state"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// CommandExecutor abstracts the execution of external commands
type CommandExecutor interface {
	Execute(command string, args ...string) error
}

// RealCommandExecutor implements CommandExecutor using exec.Command
type RealCommandExecutor struct{}

// Execute runs the command using exec.Command
func (r *RealCommandExecutor) Execute(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	return cmd.Run()
}

var (
	fs           = flag.NewFlagSet("uzi broadcast", flag.ExitOnError)
	CmdBroadcast = &ffcli.Command{
		Name:       "broadcast",
		ShortUsage: "uzi broadcast <message>",
		ShortHelp:  "Send a message to all active agent sessions",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			executor := &RealCommandExecutor{}
			return executeBroadcast(ctx, args, executor)
		},
	}
)

func executeBroadcast(ctx context.Context, args []string, executor CommandExecutor) error {
	if len(args) == 0 {
		return fmt.Errorf("message argument is required")
	}

	message := strings.Join(args, " ")
	log.Debug("Broadcasting message", "message", message)

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

	if len(activeSessions) == 0 {
		return fmt.Errorf("no active agent sessions found")
	}

	fmt.Printf("Broadcasting message to %d agent sessions:\n", len(activeSessions))

	// Send message to each session
	for _, session := range activeSessions {
		fmt.Printf("\n=== %s ===\n", session)

		// Send the message to the agent window
		if err := executor.Execute("tmux", "send-keys", "-t", session+":agent", message, "Enter"); err != nil {
			log.Error("Failed to send message to session", "session", session, "error", err)
			continue
		}
		executor.Execute("tmux", "send-keys", "-t", session+":agent", "Enter")
	}

	return nil
}
