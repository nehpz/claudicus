// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nehpz/claudicus/pkg/config"
	"github.com/nehpz/claudicus/pkg/tui"
	"github.com/peterbourgon/ff/v3/ffcli"
	"golang.org/x/term"
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
		Exec: func(ctx context.Context, args []string) error {
			return Run()
		},
	}
)

// isTerminal checks if we're running in a terminal environment
func isTerminal() bool {
	// TUI requires both stdin and stdout to be terminals
	// stdin for input, stdout for display
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// Run launches the TUI interface
func Run() error {
	// Check if we're in a terminal environment
	if !isTerminal() {
		return fmt.Errorf("TUI requires a terminal environment")
	}
	
	// TODO: Load configuration
	_ = configPath

	// Create a UziCLI instance
	uziCLI := tui.NewUziCLI()

	// Create the TUI application
	app := tui.NewApp(uziCLI)
	
	// Create the Bubble Tea program with more conservative options
	program := tea.NewProgram(
		app,
		tea.WithAltScreen(), // Use alternate screen buffer
		// Remove mouse support for now as it can cause input issues
	)

	// Run the program
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

// main function for standalone execution (if needed for testing)
func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "uzi tui: error: %v\n", err)
		os.Exit(1)
	}
}
