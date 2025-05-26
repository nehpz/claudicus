package ls

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
	fs         = flag.NewFlagSet("uzi ls", flag.ExitOnError)
	configPath = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	CmdLs      = &ffcli.Command{
		Name:       "ls",
		ShortUsage: "uzi ls",
		ShortHelp:  "List files in the current directory",
		FlagSet:    fs,
		Exec:       executeLs,
	}
)

func executeLs(ctx context.Context, args []string) error {
	log.Debug("Running ls command")

	cmd := exec.Command("tmux", "ls")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing tmux command:", err)
		return err
	}

	// Parse the output and filter session names
	lines := strings.Split(out.String(), "\n")
	var agentSessions []string
	for _, line := range lines {
		if strings.HasPrefix(line, "agent-") {
			// Extract the session name (before the first colon)
			sessionName := strings.SplitN(line, ":", 2)[0]
			agentSessions = append(agentSessions, sessionName)
		}
	}

	// Print the filtered session names
	for _, session := range agentSessions {
		fmt.Println(session)
	}

	return nil
}
