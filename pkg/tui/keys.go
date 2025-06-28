// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyMap defines the key bindings for the TUI
type KeyMap struct {
	// Navigation keys
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	
	// Action keys
	Enter  key.Binding
	Escape key.Binding
	Tab    key.Binding  // Toggle between list and split view
	
	// Agent management keys
	Broadcast key.Binding  // Broadcast message to agents
	
	// Diff preview keys
	ToggleCommits key.Binding // Toggle between diff and commits/files view
	
	// Application actions
	Help    key.Binding
	Quit    key.Binding
	Refresh key.Binding
	Kill    key.Binding
	// List specific keys
	Filter key.Binding
	Clear  key.Binding
	// Agent filtering keys
	FilterStuck   key.Binding  // Toggle stuck agents filter
	FilterWorking key.Binding  // Filter working agents
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		
		// Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle split view"),
		),
		
		// Agent management
		Broadcast: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "broadcast message"),
		),
		
		// Diff preview
		ToggleCommits: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "toggle commits view"),
		),
		
		// Application
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Kill: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "kill agent"),
		),
		
		// List specific
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear filter"),
		),
		
		// Agent filtering
		FilterStuck: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle stuck agents filter"),
		),
		FilterWorking: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "filter working agents"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
		return [][]key.Binding{
			{k.Up, k.Down, k.Left, k.Right}, // Navigation
			{k.Enter, k.Escape, k.Refresh, k.Kill},  // Actions
			{k.Tab, k.ToggleCommits, k.Broadcast}, // Views & Agent management
			{k.Filter, k.Clear, k.FilterStuck, k.FilterWorking}, // Filtering
			{k.Help, k.Quit}, // Application
		}
}

// CursorState represents the cursor position in a list
type CursorState struct {
	index   int // Current cursor position
	maxSize int // Maximum size of the list
}

// NewCursorState creates a new cursor state
func NewCursorState() *CursorState {
	return &CursorState{
		index:   0,
		maxSize: 0,
	}
}

// Index returns the current cursor index
func (c *CursorState) Index() int {
	return c.index
}

// SetMaxSize sets the maximum size for the cursor
func (c *CursorState) SetMaxSize(size int) {
	c.maxSize = size
	if c.index >= size && size > 0 {
		c.index = size - 1
	}
}

// MoveUp moves the cursor up by one position
func (c *CursorState) MoveUp() {
	if c.index > 0 {
		c.index--
	}
}

// MoveDown moves the cursor down by one position
func (c *CursorState) MoveDown() {
	if c.maxSize > 0 && c.index < c.maxSize-1 {
		c.index++
	}
}

// Reset resets the cursor to the top
func (c *CursorState) Reset() {
	c.index = 0
}

// HandleKeyMsg processes key messages for cursor navigation
// Returns true if the key was handled, false otherwise
func (c *CursorState) HandleKeyMsg(msg tea.KeyMsg, keyMap KeyMap) bool {
	switch {
	case key.Matches(msg, keyMap.Up):
		c.MoveUp()
		return true
	case key.Matches(msg, keyMap.Down):
		c.MoveDown()
		return true
	}
	return false
}
