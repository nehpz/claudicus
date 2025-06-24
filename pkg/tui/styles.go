// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions
var (
	PrimaryColor   = lipgloss.Color("#7C3AED")
	SecondaryColor = lipgloss.Color("#10B981")
	AccentColor    = lipgloss.Color("#F59E0B")
	ErrorColor     = lipgloss.Color("#EF4444")
	SuccessColor   = lipgloss.Color("#10B981")
	WarningColor   = lipgloss.Color("#F59E0B")
	MutedColor     = lipgloss.Color("#6B7280")
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle()
	
	// Header styles
	HeaderStyle = BaseStyle.Copy().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)
	
	// Content styles
	ContentStyle = BaseStyle.Copy().
		Padding(1, 2)
	
	// Status styles
	StatusReadyStyle = BaseStyle.Copy().
		Foreground(SuccessColor).
		Bold(true)
	
	StatusRunningStyle = BaseStyle.Copy().
		Foreground(WarningColor).
		Bold(true)
	
	StatusErrorStyle = BaseStyle.Copy().
		Foreground(ErrorColor).
		Bold(true)
	
	// Border styles
	BorderStyle = BaseStyle.Copy().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(MutedColor)
	
	// Selected item style
	SelectedItemStyle = BaseStyle.Copy().
		Foreground(PrimaryColor).
		Bold(true)
	
	// Help text style
	HelpStyle = BaseStyle.Copy().
		Foreground(MutedColor).
		Italic(true)
)

// ApplyTheme applies consistent theming to the given style
func ApplyTheme(style lipgloss.Style) lipgloss.Style {
	// TODO: Apply consistent theme modifications
	return style
}

// FormatStatus returns a styled status string
func FormatStatus(status string) string {
	switch status {
	case "ready":
		return StatusReadyStyle.Render("●")
	case "running":
		return StatusRunningStyle.Render("●")
	case "error":
		return StatusErrorStyle.Render("●")
	default:
		return BaseStyle.Foreground(MutedColor).Render("●")
	}
}
