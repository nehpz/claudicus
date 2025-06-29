// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// BroadcastInputModel handles the broadcast message input prompt
type BroadcastInputModel struct {
	textInput textinput.Model
	active    bool
	width     int
}

// NewBroadcastInputModel creates a new broadcast input model
func NewBroadcastInputModel() *BroadcastInputModel {
	ti := textinput.New()
	ti.Placeholder = "Enter message to broadcast..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return &BroadcastInputModel{
		textInput: ti,
		active:    false,
		width:     50,
	}
}

// SetActive activates or deactivates the broadcast input
func (m *BroadcastInputModel) SetActive(active bool) {
	m.active = active
	if active {
		m.textInput.Focus()
		m.textInput.SetValue("")
	} else {
		m.textInput.Blur()
	}
}

// IsActive returns whether the broadcast input is currently active
func (m *BroadcastInputModel) IsActive() bool {
	return m.active
}

// Value returns the current input value
func (m *BroadcastInputModel) Value() string {
	return m.textInput.Value()
}

// SetWidth updates the width of the input
func (m *BroadcastInputModel) SetWidth(width int) {
	m.width = width
	m.textInput.Width = width - 20 // Account for prompt text and padding
}

// Update handles messages for the broadcast input
func (m *BroadcastInputModel) Update(msg tea.Msg) (*BroadcastInputModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the broadcast input prompt
func (m *BroadcastInputModel) View() string {
	if !m.active {
		return ""
	}

	// Create prompt style
	promptStyle := ClaudeSquadAccentStyle.Copy().Bold(true)
	inputStyle := ClaudeSquadBorderStyle.Copy().
		Width(m.width-2).
		Padding(0, 1)

	prompt := promptStyle.Render("Message: ")
	input := m.textInput.View()

	content := prompt + input
	return inputStyle.Render(content)
}
