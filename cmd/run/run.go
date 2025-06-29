package run

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/nehpz/claudicus/pkg/config"
	"github.com/nehpz/claudicus/pkg/state"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs          = flag.NewFlagSet("uzi run", flag.ExitOnError)
	deletePanel = fs.Bool("delete", false, "delete the panel after running the command")
	configPath  = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	CmdRun      = &ffcli.Command{
		Name:       "run",
		ShortUsage: "uzi run <command>",
		ShortHelp:  "Run a command in all agent sessions",
		FlagSet:    fs,
		Exec:       executeRun,
	}
)

func executeRun(ctx context.Context, args []string) error {
	log.Debug("Running run command")

	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}

	command := strings.Join(args, " ")

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

	fmt.Printf("Running command '%s' in %d agent sessions:\n", command, len(activeSessions))

	// Execute command in each session
	for _, session := range activeSessions {
		fmt.Printf("\n=== %s ===\n", session)

		// Create a new window without specifying name or target to get next unused index
		// Use -P to print the window info in format session:index
		newWindowCmd := exec.Command("tmux", "new-window", "-t", session, "-P", "-F", "#{window_index}", "-c", "#{session_path}")
		windowIndexBytes, err := newWindowCmd.Output()
		if err != nil {
			log.Error("Failed to create new window", "session", session, "error", err)
			continue
		}

		windowIndex := strings.TrimSpace(string(windowIndexBytes))
		windowTarget := session + ":" + windowIndex

		sendKeysCmd := exec.Command("tmux", "send-keys", "-t", windowTarget, command, "Enter")
		if err := sendKeysCmd.Run(); err != nil {
			log.Error("Failed to send command ", command, " tosession", session, "error", err)
			continue
		}

		// Capture the output from the pane
		captureCmd := exec.Command("tmux", "capture-pane", "-t", windowTarget, "-p")
		var captureOut bytes.Buffer
		captureCmd.Stdout = &captureOut
		if err := captureCmd.Run(); err != nil {
			log.Error("Failed to capture output", "session", session, "error", err)
		} else {
			output := strings.TrimSpace(captureOut.String())
			if output != "" {
				fmt.Println(output)
			}
		}

		// If delete flag is set, kill the window after capturing output
		if *deletePanel {
			killWindowCmd := exec.Command("tmux", "kill-window", "-t", windowTarget)
			if err := killWindowCmd.Run(); err != nil {
				log.Error("Failed to kill window", "session", session, "window", windowTarget, "error", err)
			}
		}
	}

	return nil
}
