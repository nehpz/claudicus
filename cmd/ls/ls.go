package ls

import (
	"context"
	"flag"
	"fmt"

	"uzi/pkg/config"
	"uzi/pkg/state"

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

	stateManager := state.NewStateManager()
	if stateManager == nil {
		return fmt.Errorf("failed to create state manager")
	}

	activeSessions, err := stateManager.GetActiveSessionsForRepo()
	if err != nil {
		return fmt.Errorf("error getting active sessions: %w", err)
	}

	// Print the active sessions
	for _, session := range activeSessions {
		fmt.Println(session)
	}

	return nil
}
