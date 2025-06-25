// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v3/ffcli"
)

// TestCommandAliasResolution tests command alias resolution logic
func TestCommandAliasResolution(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"resolve prompt alias", []string{"p", "test"}, []string{"prompt", "test"}},
		{"resolve ls alias", []string{"l"}, []string{"ls"}},
		{"resolve kill alias", []string{"k", "--help"}, []string{"kill", "--help"}},
		{"no resolution needed", []string{"prompt", "test"}, []string{"prompt", "test"}},
		{"empty args", []string{}, []string{}},
		{"unknown command", []string{"xyz", "test"}, []string{"xyz", "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the alias resolution logic from main
			args := make([]string, len(tt.input))
			copy(args, tt.input)

			if len(args) > 0 {
				cmdName := args[0]
				for realCmd, pattern := range commandAliases {
					if pattern.MatchString(cmdName) {
						args[0] = realCmd
						break
					}
				}
			}

			if len(args) != len(tt.expected) {
				t.Fatalf("Expected %d args, got %d", len(tt.expected), len(args))
			}

			for i, expected := range tt.expected {
				if args[i] != expected {
					t.Errorf("Expected arg[%d] to be %q, got %q", i, expected, args[i])
				}
			}
		})
	}
}

// TestAllCommandAliasesWork tests that all documented aliases work correctly
func TestAllCommandAliasesWork(t *testing.T) {
	testCases := []struct {
		alias    string
		command  string
	}{
		// Test all documented aliases
		{"p", "prompt"},
		{"pro", "prompt"},
		{"prompt", "prompt"},
		
		{"l", "ls"},
		{"ls", "ls"},
		
		{"k", "kill"},
		{"kill", "kill"},
		
		{"re", "reset"},
		{"reset", "reset"},
		
		{"c", "checkpoint"},
		{"checkpoint", "checkpoint"},
		
		{"r", "run"},
		{"run", "run"},
		
		{"w", "watch"},
		{"watch", "watch"},
		
		{"b", "broadcast"},
		{"broadcast", "broadcast"},
		
		{"a", "attach"},
		{"attach", "attach"},
		
		{"t", "tui"},
		{"tui", "tui"},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("alias_%s", tc.alias), func(t *testing.T) {
			found := false
			for realCmd, pattern := range commandAliases {
				if pattern.MatchString(tc.alias) {
					if realCmd != tc.command {
						t.Errorf("Alias %s resolved to %s, expected %s", tc.alias, realCmd, tc.command)
					}
					found = true
					break
				}
			}
			
			if !found && tc.alias != tc.command {
				t.Errorf("Alias %s did not resolve to any command", tc.alias)
			}
		})
	}
}

// TestCommandStructure tests that the command structure is valid
func TestCommandStructure(t *testing.T) {
	// Test that all expected subcommands are present
	expectedCommands := []string{
		"prompt", "ls", "kill", "reset", "run", 
		"checkpoint", "auto", "broadcast", "tui",
	}
	
	if len(subcommands) != len(expectedCommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedCommands), len(subcommands))
	}
	
	// Check that each expected command exists
	commandNames := make(map[string]bool)
	for _, cmd := range subcommands {
		commandNames[cmd.Name] = true
	}
	
	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("Expected command %s not found", expected)
		}
	}
}

// TestFlagParsingErrorHandling tests various flag parsing scenarios
func TestFlagParsingErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expectExit bool
		expectHelp bool
	}{
		{
			name:     "invalid flag",
			args:     []string{"--invalid-flag"},
			expectExit: true,
		},
		{
			name:     "help flag",
			args:     []string{"--help"},
			expectHelp: true,
		},
		{
			name:     "h flag", 
			args:     []string{"-h"},
			expectHelp: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for testing
			c := new(ffcli.Command)
			c.Name = "uzi"
			c.ShortUsage = "uzi <command>"
			c.Subcommands = subcommands
			c.FlagSet = flag.NewFlagSet("uzi", flag.ContinueOnError)
			
			// Capture parse error
			err := c.Parse(tt.args)
			
			if tt.expectHelp {
				if !errors.Is(err, flag.ErrHelp) {
					t.Errorf("Expected help error, got %v", err)
				}
			} else if tt.expectExit {
				if err == nil {
					t.Error("Expected error for invalid flag")
				}
				if !strings.Contains(err.Error(), "flag provided but not defined") {
					t.Errorf("Expected 'flag provided but not defined' error, got %v", err)
				}
			}
		})
	}
}

// TestUnknownCommandHandling tests unknown command handling
func TestUnknownCommandHandling(t *testing.T) {
	// Backup original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Test unknown command
	os.Args = []string{"uzi", "unknown-command"}
	
	c := new(ffcli.Command)
	c.Name = "uzi"
	c.ShortUsage = "uzi <command>"
	c.Subcommands = subcommands
	c.FlagSet = flag.NewFlagSet("uzi", flag.ContinueOnError)
	
	var execCalled bool
	var execError error
	
	c.Exec = func(ctx context.Context, args []string) error {
		execCalled = true
		if len(os.Args) >= 2 {
			execError = fmt.Errorf("unknown command %q", os.Args[1])
			return execError
		}
		return nil
	}
	
	// Parse should succeed but exec should fail
	err := c.Parse(os.Args[1:])
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}
	
	// Run the command
	ctx := context.Background()
	err = c.Run(ctx)
	
	if !execCalled {
		t.Error("Exec function was not called")
	}
	
	if err == nil {
		t.Error("Expected error for unknown command")
	}
	
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected 'unknown command' error, got %v", err)
	}
}

// TestSubcommandErrorHandling tests graceful error handling from subcommands
func TestSubcommandErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		subcommandErr error
		expectError   bool
	}{
		{
			name:          "successful execution",
			subcommandErr: nil,
			expectError:   false,
		},
		{
			name:          "subcommand error",
			subcommandErr: errors.New("mock subcommand error"),
			expectError:   true,
		},
		{
			name:          "context canceled",
			subcommandErr: context.Canceled,
			expectError:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock subcommand
			mockCmd := &ffcli.Command{
				Name: "mock",
				Exec: func(ctx context.Context, args []string) error {
					return tt.subcommandErr
				},
			}
			
			// Create main command with mock subcommand
			c := &ffcli.Command{
				Name:        "uzi",
				ShortUsage:  "uzi <command>",
				Subcommands: []*ffcli.Command{mockCmd},
				FlagSet:     flag.NewFlagSet("uzi", flag.ContinueOnError),
			}
			
			// Parse and run
			err := c.Parse([]string{"mock"})
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			
			ctx := context.Background()
			err = c.Run(ctx)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if tt.expectError && err != nil {
				if !errors.Is(err, tt.subcommandErr) && err.Error() != tt.subcommandErr.Error() {
					t.Errorf("Expected error %v, got %v", tt.subcommandErr, err)
				}
			}
		})
	}
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	mockCmd := &ffcli.Command{
		Name: "mock",
		Exec: func(ctx context.Context, args []string) error {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		},
	}
	
	c := &ffcli.Command{
		Name:        "uzi",
		ShortUsage:  "uzi <command>",
		Subcommands: []*ffcli.Command{mockCmd},
		FlagSet:     flag.NewFlagSet("uzi", flag.ContinueOnError),
	}
	
	// Parse command
	err := c.Parse([]string{"mock"})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	// Run should handle cancellation gracefully
	err = c.Run(ctx)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

// Benchmark command parsing performance
func BenchmarkCommandParsing(b *testing.B) {
	c := new(ffcli.Command)
	c.Name = "uzi"
	c.ShortUsage = "uzi <command>"
	c.Subcommands = subcommands
	c.FlagSet = flag.NewFlagSet("uzi", flag.ContinueOnError)
	
	args := []string{"prompt", "--help"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new flagset for each iteration to avoid state pollution
		c.FlagSet = flag.NewFlagSet("uzi", flag.ContinueOnError)
		c.Parse(args)
	}
}

// Benchmark alias resolution performance
func BenchmarkAliasResolution(b *testing.B) {
	testAliases := []string{"p", "prompt", "l", "ls", "k", "kill", "c", "checkpoint"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alias := testAliases[i%len(testAliases)]
		for realCmd, pattern := range commandAliases {
			if pattern.MatchString(alias) {
				_ = realCmd
				break
			}
		}
	}
}

// TestBadConfigNoPanics tests that bad input doesn't cause panics
func TestBadConfigNoPanics(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Bad config caused panic: %v", r)
		}
	}()
	
	tests := []struct {
		name string
		args []string
	}{
		{"empty args", []string{}},
		{"malformed flag", []string{"--flag="}},
		{"special characters", []string{"--flag=\x00\x01\x02"}},
		{"very long argument", []string{strings.Repeat("a", 1000)}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Test %s caused panic: %v", tt.name, r)
				}
			}()
			
			c := new(ffcli.Command)
			c.Name = "uzi"
			c.ShortUsage = "uzi <command>"
			c.Subcommands = subcommands
			c.FlagSet = flag.NewFlagSet("uzi", flag.ContinueOnError)
			
			// Parse should not panic even with bad input
			c.Parse(tt.args)
		})
	}
}