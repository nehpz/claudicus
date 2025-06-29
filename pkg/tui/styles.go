// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Claude Squad Color Palette
// Based on the design from uzi-site/index.html
var (
	// Primary Claude Squad colors
	ClaudeSquadPrimary = lipgloss.Color("#ffffff") // White text
	ClaudeSquadAccent  = lipgloss.Color("#00ff9d") // Signature green
	ClaudeSquadDark    = lipgloss.Color("#0a0a0a") // Deep black background
	ClaudeSquadGray    = lipgloss.Color("#1a1a1a") // Dark gray containers
	ClaudeSquadMuted   = lipgloss.Color("#6b7280") // Muted gray for secondary text
	ClaudeSquadHover   = lipgloss.Color("#00e68a") // Slightly darker green for hover

	// Legacy colors for backward compatibility
	PrimaryColor   = lipgloss.Color("#7C3AED")
	SecondaryColor = lipgloss.Color("#10B981")
	AccentColor    = lipgloss.Color("#F59E0B")
	ErrorColor     = lipgloss.Color("#EF4444")
	SuccessColor   = lipgloss.Color("#10B981")
	WarningColor   = lipgloss.Color("#F59E0B")
	MutedColor     = lipgloss.Color("#6B7280")
)

// Claude Squad Base Styles
var (
	// Core styling with Claude Squad theme
	ClaudeSquadBaseStyle = lipgloss.NewStyle().
				Foreground(ClaudeSquadPrimary).
				Background(ClaudeSquadDark)

	// Warning and error styles
	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor)

	// Primary accent styling (Claude Squad green)
	ClaudeSquadAccentStyle = ClaudeSquadBaseStyle.Copy().
				Foreground(ClaudeSquadAccent).
				Bold(true)

	// Primary text styling
	ClaudeSquadPrimaryStyle = ClaudeSquadBaseStyle.Copy().
				Foreground(ClaudeSquadPrimary)

	// Muted text styling
	ClaudeSquadMutedStyle = ClaudeSquadBaseStyle.Copy().
				Foreground(ClaudeSquadMuted)

	// Header styling
	ClaudeSquadHeaderStyle = ClaudeSquadBaseStyle.Copy().
				Foreground(ClaudeSquadPrimary).
				Bold(true).
				MarginBottom(1)

	// Header bar styling
	ClaudeSquadHeaderBarStyle = ClaudeSquadBaseStyle.Copy().
					Background(ClaudeSquadGray)

	// Border styling with Claude Squad theme
	ClaudeSquadBorderStyle = ClaudeSquadBaseStyle.Copy().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ClaudeSquadAccent).
				Padding(1)

	// Selected item styling (highlighted with Claude Squad green)
	ClaudeSquadSelectedStyle = ClaudeSquadBaseStyle.Copy().
					Foreground(ClaudeSquadAccent).
					Bold(true)

	// Selected description styling
	ClaudeSquadSelectedDescStyle = ClaudeSquadBaseStyle.Copy().
					Foreground(ClaudeSquadPrimary)

	// Normal title styling
	ClaudeSquadNormalTitleStyle = ClaudeSquadBaseStyle.Copy().
					Foreground(ClaudeSquadPrimary)

	// Normal description styling
	ClaudeSquadNormalDescStyle = ClaudeSquadBaseStyle.Copy().
					Foreground(ClaudeSquadMuted)
)

// Legacy Base styles for backward compatibility
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

// Claude Squad utility functions

// FormatStatusWithClaudeSquad returns a styled status string using Claude Squad colors
func FormatStatusWithClaudeSquad(status string) string {
	switch status {
	case "attached":
		return ClaudeSquadAccentStyle.Render("●")
	case "running":
		return ClaudeSquadAccentStyle.Render("●")
	case "ready":
		return ClaudeSquadPrimaryStyle.Render("○")
	case "inactive":
		return ClaudeSquadMutedStyle.Render("○")
	default:
		return ClaudeSquadMutedStyle.Render("?")
	}
}

// ApplyClaudeSquadTheme applies Claude Squad theming to the given style
func ApplyClaudeSquadTheme(style lipgloss.Style) lipgloss.Style {
	return style.
		Foreground(ClaudeSquadPrimary).
		Background(ClaudeSquadDark)
}

// ApplyTheme applies consistent theming to the given style (legacy)
func ApplyTheme(style lipgloss.Style) lipgloss.Style {
	// For backward compatibility, apply Claude Squad theme
	return ApplyClaudeSquadTheme(style)
}

// FormatStatus returns a styled status string (legacy)
func FormatStatus(status string) string {
	// Use Claude Squad styling by default
	return FormatStatusWithClaudeSquad(status)
}
