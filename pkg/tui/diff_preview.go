// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DiffPreviewModel handles git diff display for agent sessions
type DiffPreviewModel struct {
	content string
	error   string
	width   int
	height  int
}

// NewDiffPreviewModel creates a new diff preview model
func NewDiffPreviewModel(width, height int) *DiffPreviewModel {
	return &DiffPreviewModel{
		width:  width,
		height: height,
	}
}

// SetSize updates the dimensions of the diff preview
func (m *DiffPreviewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// LoadDiff loads git diff for the given session
func (m *DiffPreviewModel) LoadDiff(session *SessionInfo) {
	if session == nil {
		m.content = ""
		m.error = ""
		return
	}

	// Clear previous error
	m.error = ""

	// Get git diff for the session's worktree
	diff, err := m.getGitDiff(session.WorktreePath)
	if err != nil {
		m.error = fmt.Sprintf("Error loading diff: %v", err)
		m.content = ""
		return
	}

	m.content = diff
}

// getGitDiff executes git diff command and returns the output
func (m *DiffPreviewModel) getGitDiff(worktreePath string) (string, error) {
	if worktreePath == "" {
		return "No worktree path available", nil
	}

	// Stage all changes temporarily to show in diff
	cmd := exec.Command("sh", "-c", "git add -A . && git diff --cached HEAD && git reset HEAD > /dev/null 2>&1")
	cmd.Dir = worktreePath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "No changes in this session", nil
	}

	return result, nil
}

// View renders the diff preview
func (m *DiffPreviewModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Create border style
	borderStyle := ClaudeSquadBorderStyle.Copy().
		Width(m.width - 2).
		Height(m.height - 2)

	// Handle error case
	if m.error != "" {
		errorContent := ClaudeSquadMutedStyle.Render(m.error)
		return borderStyle.Render(errorContent)
	}

	// Handle empty content
	if m.content == "" {
		emptyContent := ClaudeSquadMutedStyle.Render("Select an agent to view diff")
		return borderStyle.Render(emptyContent)
	}

	// Format the diff content with basic syntax highlighting
	formattedContent := m.formatDiffContent(m.content)
	
	return borderStyle.Render(formattedContent)
}

// formatDiffContent applies basic syntax highlighting to git diff output
func (m *DiffPreviewModel) formatDiffContent(content string) string {
	lines := strings.Split(content, "\n")
	var formatted []string

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			// File headers
			formatted = append(formatted, ClaudeSquadAccentStyle.Render(line))
		case strings.HasPrefix(line, "@@"):
			// Hunk headers
			formatted = append(formatted, ClaudeSquadPrimaryStyle.Render(line))
		case strings.HasPrefix(line, "+"):
			// Additions
			formatted = append(formatted, lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff9d")).Render(line))
		case strings.HasPrefix(line, "-"):
			// Deletions
			formatted = append(formatted, lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b")).Render(line))
		default:
			// Context lines
			formatted = append(formatted, ClaudeSquadMutedStyle.Render(line))
		}
	}

	// Limit the number of lines to fit in the view
	maxLines := m.height - 4 // Account for border and padding
	if len(formatted) > maxLines {
		formatted = formatted[:maxLines]
		formatted = append(formatted, ClaudeSquadMutedStyle.Render("... (truncated)"))
	}

	return strings.Join(formatted, "\n")
}