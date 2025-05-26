package prompt

import (
	"context"
	"flag"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs        = flag.NewFlagSet("uzi prompt", flag.ExitOnError)
	CmdPrompt = &ffcli.Command{
		Name:       "prompt",
		ShortUsage: "uzi prompt",
		ShortHelp:  "Run the prompt command",
		FlagSet:    fs,
		Exec:       exec,
	}
)

func exec(ctx context.Context, args []string) error {
	log.Info("Running prompt command")
	fmt.Println("This is the prompt command")
	return nil
} 