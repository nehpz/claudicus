package ls

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
	"uzi/pkg/config"
)

var (
	fs         = flag.NewFlagSet("uzi ls", flag.ExitOnError)
	configPath = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	CmdLs      = &ffcli.Command{
		Name:       "ls",
		ShortUsage: "uzi ls",
		ShortHelp:  "List files in the current directory",
		FlagSet:    fs,
		Exec:       exec,
	}
)

func exec(ctx context.Context, args []string) error {
	log.Info("Running ls command")

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Error("Error loading config", "error", err)
		return err
	}

	// Print configuration details
	fmt.Printf("Port Range: %s\n", cfg.PortRange)
	fmt.Printf("Start Command: %s\n", cfg.StartCommand)
	fmt.Println("\nAgents:")
	for _, agent := range cfg.Agents {
		fmt.Printf("- Command: %s (Count: %d)\n", agent.Command, agent.Name)
	}

	// Original directory listing
	files, err := os.ReadDir(".")
	if err != nil {
		log.Error("Error reading directory", "error", err)
		return err
	}

	// Print directory contents
	fmt.Println("\nDirectory contents:")
	for _, file := range files {
		info, _ := file.Info()
		if info != nil {
			fmt.Printf("%s\t%d bytes\n", file.Name(), info.Size())
		} else {
			fmt.Println(file.Name())
		}
	}

	return nil
}
