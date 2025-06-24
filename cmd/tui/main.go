// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/devflowinc/uzi/pkg/config"
	"github.com/devflowinc/uzi/pkg/tui"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs = flag.NewFlagSet("uzi tui", flag.ExitOnError)
	configPath = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	CmdTui = &ffcli.Command{
		Name:       "tui",
		ShortUsage: "uzi tui",
		ShortHelp:  "Launch the interactive TUI interface",
		LongHelp: `Launch the interactive Terminal User Interface (TUI) for managing agent sessions.

The TUI provides a visual interface for:
- Viewing active agent sessions
- Monitoring session status
- Attaching to sessions
- Managing session lifecycle

Navigation:
- Use arrow keys or vim-style keys (h/j/k/l) to navigate
- Press Enter to select an item
- Press 'q' to quit
- Press '?' for help`,
		FlagSet: fs,
		Exec:    executeTui,
	}
)

func executeTui(ctx context.Context, args []string) error {
	// TODO: Load configuration
	_ = configPath
	_ = args

	// Create the TUI application
	app := tui.NewApp()
	
	// Create the Bubble Tea program
	program := tea.NewProgram(
		app,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

// main function for standalone execution (if needed for testing)
func main() {
	ctx := context.Background()
	if err := executeTui(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "uzi tui: error: %v\n", err)
		os.Exit(1)
	}
}
