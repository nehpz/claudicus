// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"uzi/cmd/ls"
	"uzi/cmd/prompt"
	"uzi/cmd/delete"
)

var subcommands = []*ffcli.Command{
	prompt.CmdPrompt,
	ls.CmdLs,
	delete.CmdDelete,
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := new(ffcli.Command)
	c.Name = filepath.Base(os.Args[0])
	c.ShortUsage = "uzi <command>"
	c.Subcommands = subcommands

	c.FlagSet = flag.NewFlagSet("uzi", flag.ContinueOnError)
	c.FlagSet.SetOutput(os.Stdout)
	c.Exec = func(ctx context.Context, args []string) error {
		fmt.Fprintf(os.Stdout, "%s\n", c.UsageFunc(c))

		if len(os.Args) >= 2 {
			return fmt.Errorf("unknown command %q", os.Args[1])
		}
		return nil
	}

	switch err := c.Parse(os.Args[1:]); {
	case err == nil:
	case errors.Is(err, flag.ErrHelp):
		return
	case strings.Contains(err.Error(), "flag provided but not defined"):
		os.Exit(2)
	default:
		fmt.Fprintf(os.Stderr, "uzi: error: %v\n", err)
		os.Exit(1)
	}

	if err := c.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "uzi: error: %v\n", err)
		os.Exit(1)
	}
}
