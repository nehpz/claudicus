// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// ListItem represents an item in the TUI list
type ListItem struct {
	title       string
	description string
}

// Title implements list.Item interface
func (i ListItem) Title() string {
	return i.title
}

// Description implements list.Item interface  
func (i ListItem) Description() string {
	return i.description
}

// FilterValue implements list.Item interface
func (i ListItem) FilterValue() string {
	return i.title
}

// NewListItem creates a new list item
func NewListItem(title, description string) ListItem {
	return ListItem{
		title:       title,
		description: description,
	}
}

// ListModel wraps the bubbles list component
type ListModel struct {
	list list.Model
}

// NewListModel creates a new list model
func NewListModel(items []list.Item, width, height int) ListModel {
	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = "Agent Sessions"
	
	return ListModel{
		list: l,
	}
}

// Init implements tea.Model interface
func (m ListModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model interface
func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// TODO: Handle list updates
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// TODO: Handle key events for list navigation
		_ = msg
	}
	
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model interface
func (m ListModel) View() string {
	return m.list.View()
}
