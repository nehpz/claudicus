// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultKeyMap(t *testing.T) {
	keyMap := DefaultKeyMap()

	// Test navigation key bindings
	testCases := []struct {
		name     string
		binding  key.Binding
		expected []string
	}{
		{"Up navigation", keyMap.Up, []string{"up"}},
		{"Down navigation", keyMap.Down, []string{"down", "j"}},
		{"Left navigation", keyMap.Left, []string{"left", "h"}},
		{"Right navigation", keyMap.Right, []string{"right", "l"}},
		{"Enter action", keyMap.Enter, []string{"enter"}},
		{"Escape action", keyMap.Escape, []string{"esc"}},
		{"Quit application", keyMap.Quit, []string{"q", "ctrl+c"}},
		{"Help", keyMap.Help, []string{"?"}},
		{"Refresh", keyMap.Refresh, []string{"r"}},
		{"Filter", keyMap.Filter, []string{"/"}},
		{"Clear", keyMap.Clear, []string{"c"}},
		{"Kill", keyMap.Kill, []string{"k"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys := tc.binding.Keys()
			if len(keys) != len(tc.expected) {
				t.Errorf("Expected %d keys, got %d", len(tc.expected), len(keys))
				return
			}
			for i, expectedKey := range tc.expected {
				if keys[i] != expectedKey {
					t.Errorf("Expected key '%s' at position %d, got '%s'", expectedKey, i, keys[i])
				}
			}
		})
	}
}

func TestCursorStateInit(t *testing.T) {
	cursor := NewCursorState()

	if cursor.Index() != 0 {
		t.Errorf("Expected initial cursor index to be 0, got %d", cursor.Index())
	}

	if cursor.maxSize != 0 {
		t.Errorf("Expected initial maxSize to be 0, got %d", cursor.maxSize)
	}
}

func TestCursorStateSetMaxSize(t *testing.T) {
	cursor := NewCursorState()

	// Test setting max size
	cursor.SetMaxSize(5)
	if cursor.maxSize != 5 {
		t.Errorf("Expected maxSize to be 5, got %d", cursor.maxSize)
	}

	// Test that cursor index is adjusted when exceeding max size
	cursor.index = 10
	cursor.SetMaxSize(3)
	if cursor.Index() != 2 {
		t.Errorf("Expected cursor index to be adjusted to 2, got %d", cursor.Index())
	}

	// Test edge case: setting max size to 0
	cursor.SetMaxSize(0)
	if cursor.Index() != 2 {
		t.Errorf("Expected cursor index to remain unchanged when maxSize is 0, got %d", cursor.Index())
	}
}

func TestCursorStateMoveUp(t *testing.T) {
	cursor := NewCursorState()
	cursor.SetMaxSize(5)

	// Test moving up from initial position (should not move)
	cursor.MoveUp()
	if cursor.Index() != 0 {
		t.Errorf("Expected cursor to remain at 0 when moving up from initial position, got %d", cursor.Index())
	}

	// Test moving up from middle position
	cursor.index = 3
	cursor.MoveUp()
	if cursor.Index() != 2 {
		t.Errorf("Expected cursor index to be 2 after moving up from 3, got %d", cursor.Index())
	}

	// Test moving up multiple times
	cursor.MoveUp()
	cursor.MoveUp()
	if cursor.Index() != 0 {
		t.Errorf("Expected cursor index to be 0 after moving up twice from 2, got %d", cursor.Index())
	}
}

func TestCursorStateMoveDown(t *testing.T) {
	cursor := NewCursorState()
	cursor.SetMaxSize(5)

	// Test moving down from initial position
	cursor.MoveDown()
	if cursor.Index() != 1 {
		t.Errorf("Expected cursor index to be 1 after moving down from 0, got %d", cursor.Index())
	}

	// Test moving down to maximum position
	for i := 0; i < 5; i++ {
		cursor.MoveDown()
	}
	if cursor.Index() != 4 {
		t.Errorf("Expected cursor index to be 4 (max-1) after moving down beyond limit, got %d", cursor.Index())
	}

	// Test moving down from maximum position (should not move)
	cursor.MoveDown()
	if cursor.Index() != 4 {
		t.Errorf("Expected cursor to remain at 4 when moving down from max position, got %d", cursor.Index())
	}
}

func TestCursorStateMoveDownEmptyList(t *testing.T) {
	cursor := NewCursorState()
	cursor.SetMaxSize(0)

	// Test moving down in empty list (should not move)
	cursor.MoveDown()
	if cursor.Index() != 0 {
		t.Errorf("Expected cursor to remain at 0 when moving down in empty list, got %d", cursor.Index())
	}
}

func TestCursorStateReset(t *testing.T) {
	cursor := NewCursorState()
	cursor.SetMaxSize(5)
	cursor.index = 3

	cursor.Reset()
	if cursor.Index() != 0 {
		t.Errorf("Expected cursor index to be 0 after reset, got %d", cursor.Index())
	}
}

func TestCursorStateHandleKeyMsg(t *testing.T) {
	cursor := NewCursorState()
	cursor.SetMaxSize(3)
	keyMap := DefaultKeyMap()

	testCases := []struct {
		name          string
		keyMsg        tea.KeyMsg
		initialIndex  int
		expectedIndex int
		shouldHandle  bool
	}{
		{
			name:          "Ignore 'k' key (now kill command)",
			keyMsg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			initialIndex:  1,
			expectedIndex: 1,
			shouldHandle:  false,
		},
		{
			name:          "Handle 'j' key (down)",
			keyMsg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			initialIndex:  0,
			expectedIndex: 1,
			shouldHandle:  true,
		},
		{
			name:          "Handle up arrow key",
			keyMsg:        tea.KeyMsg{Type: tea.KeyUp},
			initialIndex:  2,
			expectedIndex: 1,
			shouldHandle:  true,
		},
		{
			name:          "Handle down arrow key",
			keyMsg:        tea.KeyMsg{Type: tea.KeyDown},
			initialIndex:  1,
			expectedIndex: 2,
			shouldHandle:  true,
		},
		{
			name:          "Ignore other keys",
			keyMsg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}},
			initialIndex:  1,
			expectedIndex: 1,
			shouldHandle:  false,
		},
		{
			name:          "Ignore enter key",
			keyMsg:        tea.KeyMsg{Type: tea.KeyEnter},
			initialIndex:  1,
			expectedIndex: 1,
			shouldHandle:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cursor.index = tc.initialIndex
			handled := cursor.HandleKeyMsg(tc.keyMsg, keyMap)

			if handled != tc.shouldHandle {
				t.Errorf("Expected handled to be %v, got %v", tc.shouldHandle, handled)
			}

			if cursor.Index() != tc.expectedIndex {
				t.Errorf("Expected cursor index to be %d, got %d", tc.expectedIndex, cursor.Index())
			}
		})
	}
}

func TestCursorStateBoundaryConditions(t *testing.T) {
	cursor := NewCursorState()

	// Test with single item list
	cursor.SetMaxSize(1)
	cursor.MoveDown()
	if cursor.Index() != 0 {
		t.Errorf("Expected cursor to remain at 0 in single-item list, got %d", cursor.Index())
	}

	cursor.MoveUp()
	if cursor.Index() != 0 {
		t.Errorf("Expected cursor to remain at 0 when moving up in single-item list, got %d", cursor.Index())
	}

	// Test rapid navigation
	cursor.SetMaxSize(10)
	for i := 0; i < 20; i++ {
		cursor.MoveDown()
	}
	if cursor.Index() != 9 {
		t.Errorf("Expected cursor to be at max position (9) after rapid down movement, got %d", cursor.Index())
	}

	for i := 0; i < 20; i++ {
		cursor.MoveUp()
	}
	if cursor.Index() != 0 {
		t.Errorf("Expected cursor to be at min position (0) after rapid up movement, got %d", cursor.Index())
	}
}

func TestKeyMapHelpers(t *testing.T) {
	keyMap := DefaultKeyMap()

	// Test ShortHelp
	shortHelp := keyMap.ShortHelp()
	if len(shortHelp) != 2 {
		t.Errorf("Expected ShortHelp to return 2 bindings, got %d", len(shortHelp))
	}

	// Test FullHelp
	fullHelp := keyMap.FullHelp()
	if len(fullHelp) != 3 {
		t.Errorf("Expected FullHelp to return 3 groups, got %d", len(fullHelp))
	}

	// Test first group (navigation)
	if len(fullHelp[0]) != 4 {
		t.Errorf("Expected first group to have 4 navigation keys, got %d", len(fullHelp[0]))
	}

	// Test second group (actions)
	if len(fullHelp[1]) != 4 {
		t.Errorf("Expected second group to have 4 action keys, got %d", len(fullHelp[1]))
	}

	// Test third group (application)
	if len(fullHelp[2]) != 4 {
		t.Errorf("Expected third group to have 4 application keys, got %d", len(fullHelp[2]))
	}
}

func TestNavigationKeysOnly(t *testing.T) {
	cursor := NewCursorState()
	cursor.SetMaxSize(5)
	keyMap := DefaultKeyMap()

	// Test that only navigation keys are handled
	navigationKeys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'j'}}, // down
		{Type: tea.KeyUp},                        // up arrow
		{Type: tea.KeyDown},                      // down arrow
	}

	nonNavigationKeys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'h'}}, // left (not handled by cursor)
		{Type: tea.KeyRunes, Runes: []rune{'l'}}, // right (not handled by cursor)
		{Type: tea.KeyEnter},                     // enter
		{Type: tea.KeyEsc},                       // escape
		{Type: tea.KeyRunes, Runes: []rune{'q'}}, // quit
		{Type: tea.KeyRunes, Runes: []rune{'?'}}, // help
		{Type: tea.KeyRunes, Runes: []rune{'r'}}, // refresh
		{Type: tea.KeyRunes, Runes: []rune{'/'}}, // filter
		{Type: tea.KeyRunes, Runes: []rune{'c'}}, // clear
		{Type: tea.KeyRunes, Runes: []rune{'k'}}, // kill (now application key)
		{Type: tea.KeyRunes, Runes: []rune{'x'}}, // random key
	}

	// Test navigation keys are handled
	for i, keyMsg := range navigationKeys {
		handled := cursor.HandleKeyMsg(keyMsg, keyMap)
		if !handled {
			t.Errorf("Navigation key %d should be handled", i)
		}
	}

	// Test non-navigation keys are ignored
	for i, keyMsg := range nonNavigationKeys {
		originalIndex := cursor.Index()
		handled := cursor.HandleKeyMsg(keyMsg, keyMap)
		if handled {
			t.Errorf("Non-navigation key %d should not be handled", i)
		}
		if cursor.Index() != originalIndex {
			t.Errorf("Cursor position should not change for non-navigation key %d", i)
		}
	}
}
