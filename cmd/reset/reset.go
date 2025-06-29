package reset

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs       = flag.NewFlagSet("uzi reset", flag.ExitOnError)
	CmdReset = &ffcli.Command{
		Name:       "reset",
		ShortUsage: "uzi reset",
		ShortHelp:  "Delete all data stored in ~/.local/share/uzi",
		FlagSet:    fs,
		Exec:       executeReset,
	}
)

func executeReset(ctx context.Context, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	uziDataPath := filepath.Join(homeDir, ".local", "share", "uzi")

	// Check if the directory exists
	if _, err := os.Stat(uziDataPath); os.IsNotExist(err) {
		log.Debug("Uzi data directory does not exist", "path", uziDataPath)
		fmt.Println("No uzi data found to reset")
		return nil
	}

	// Ask for confirmation
	fmt.Printf("This will permanently delete all uzi data from %s\n", uziDataPath)
	fmt.Print("Are you sure you want to continue? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		fmt.Println("Reset cancelled")
		return nil
	}

	// Remove the entire uzi data directory
	if err := os.RemoveAll(uziDataPath); err != nil {
		log.Error("Error removing uzi data directory", "path", uziDataPath, "error", err)
		return fmt.Errorf("failed to remove uzi data directory: %w", err)
	}

	log.Debug("Removed uzi data directory", "path", uziDataPath)
	fmt.Printf("Successfully reset all uzi data from %s\n", uziDataPath)
	return nil
}
