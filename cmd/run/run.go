package run

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"uzi/pkg/config"

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

	// Get list of agent sessions
	cmd := exec.Command("tmux", "ls")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error listing tmux sessions: %w", err)
	}

	// Parse the output and filter session names
	lines := strings.Split(out.String(), "\n")
	var agentSessions []string
	for _, line := range lines {
		if strings.HasPrefix(line, "agent-") {
			sessionName := strings.SplitN(line, ":", 2)[0]
			agentSessions = append(agentSessions, sessionName)
		}
	}

	if len(agentSessions) == 0 {
		return fmt.Errorf("no agent sessions found")
	}

	fmt.Printf("Running command '%s' in %d agent sessions:\n", command, len(agentSessions))

	// Execute command in each session
	for _, session := range agentSessions {
		fmt.Printf("\n=== %s ===\n", session)

		// Only set paneName to "uzi-run" if delete is true. Otherwise make it the command in all lowercase
		paneName := "uzi-run"
		if *deletePanel == false {
			paneName = strings.ToLower(command)
		}

		newWindowCmd := exec.Command("tmux", "new-window", "-t", session, "-n", paneName, "-c", "#{session_path}")
		if err := newWindowCmd.Run(); err != nil {
			log.Error("Failed to create new window", "session", session, "error", err)
			continue
		}

		sendKeysCmd := exec.Command("tmux", "send-keys", "-t", session+":"+paneName, command, "Enter")
		if err := sendKeysCmd.Run(); err != nil {
			log.Error("Failed to send command ", command, " tosession", session, "error", err)
			continue
		}

		// Capture the output from the pane
		captureCmd := exec.Command("tmux", "capture-pane", "-t", session+":"+paneName, "-p")
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

		// Delete the pane after command execution
		// killWindowCmd := exec.Command("tmux", "kill-window", "-t", session+":"+paneName)
		// if err := killWindowCmd.Run(); err != nil {
		// 	log.Error("Failed to delete window", "session", session, "error", err)
		// }
	}

	return nil
}
