package ls

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs    = flag.NewFlagSet("uzi ls", flag.ExitOnError)
	CmdLs = &ffcli.Command{
		Name:       "ls",
		ShortUsage: "uzi ls",
		ShortHelp:  "List files in the current directory",
		FlagSet:    fs,
		Exec:       exec,
	}
)

func exec(ctx context.Context, args []string) error {
	log.Info("Running ls command")
	files, err := os.ReadDir(".")
	if err != nil {
		log.Error("Error reading directory", "error", err)
		return err
	}
	for _, file := range files {
		fmt.Println(file.Name())
	}
	return nil
} 