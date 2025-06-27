// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffViewMsg represents a message containing git diff content
type DiffViewMsg struct {
	SessionName string
	Content     string
	Error       error
}

// DiffView represents a git diff preview component
type DiffView struct {
	viewport      viewport.Model
	sessionName   string
	content       string
	error         error
	width         int
	height        int
	loading       bool
}

// NewDiffView creates a new diff view component
func NewDiffView(width, height int) *DiffView {
	vp := viewport.New(width-2, height-4) // Account for borders and header
	vp.Style = ClaudeSquadBaseStyle
	
	return &DiffView{
		viewport: vp,
		width:    width,
		height:   height,
		loading:  false,
	}
}

// SetSize updates the dimensions of the diff view
func (dv *DiffView) SetSize(width, height int) {
	dv.width = width
	dv.height = height
	dv.viewport.Width = width - 2   // Account for borders
	dv.viewport.Height = height - 4 // Account for header and borders
}

// LoadSessionDiff loads the git diff for a specific session
func (dv *DiffView) LoadSessionDiff(sessionName string) tea.Cmd {
	if sessionName == "" {
		dv.sessionName = ""
		dv.content = ""
		dv.error = nil
		dv.loading = false
		return nil
	}
	
	dv.sessionName = sessionName
	dv.loading = true
	
	return func() tea.Msg {
		content, err := dv.getGitDiff(sessionName)
		return DiffViewMsg{
			SessionName: sessionName,
			Content:     content,
			Error:       err,
		}
	}
}

// getGitDiff retrieves the git diff for a session's worktree
func (dv *DiffView) getGitDiff(sessionName string) (string, error) {
	// First, get the session state to find the worktree path
	uziCLI := NewUziCLI()
	sessionState, err := uziCLI.GetSessionState(sessionName)
	if err != nil {
		return "", fmt.Errorf("failed to get session state: %w", err)
	}
	
	if sessionState.WorktreePath == "" {
		return "No worktree path found for session", nil
	}
	
	// Run git diff to get the changes
	cmd := exec.Command("git", "diff", "HEAD")
	cmd.Dir = sessionState.WorktreePath
	
	output, err := cmd.Output()
	if err != nil {
		// Try git diff --cached for staged changes
		cmd = exec.Command("git", "diff", "--cached", "HEAD")
		cmd.Dir = sessionState.WorktreePath
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get git diff: %w", err)
		}
	}
	
	content := string(output)
	if content == "" {
		return "No changes detected", nil
	}
	
	return content, nil
}

// Update implements tea.Model interface
func (dv *DiffView) Update(msg tea.Msg) (*DiffView, tea.Cmd) {
	switch msg := msg.(type) {
	case DiffViewMsg:
		if msg.SessionName == dv.sessionName {
			dv.content = msg.Content
			dv.error = msg.Error
			dv.loading = false
			
			// Update viewport content
			if msg.Error != nil {
				errorContent := ClaudeSquadMutedStyle.Render(fmt.Sprintf("Error: %v", msg.Error))
				dv.viewport.SetContent(errorContent)
			} else {
				// Apply basic syntax highlighting to git diff
				styledContent := dv.styleDiff(msg.Content)
				dv.viewport.SetContent(styledContent)
			}
		}
		return dv, nil
		
	case tea.WindowSizeMsg:
		dv.SetSize(msg.Width/2, msg.Height) // Split view takes half the width
		return dv, nil
		
	default:
		var cmd tea.Cmd
		dv.viewport, cmd = dv.viewport.Update(msg)
		return dv, cmd
	}
}

// styleDiff applies basic syntax highlighting to git diff output
func (dv *DiffView) styleDiff(content string) string {
	if content == "" {
		return ClaudeSquadMutedStyle.Render("No changes to display")
	}
	
	lines := strings.Split(content, "\n")
	styledLines := make([]string, len(lines))
	
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			// File headers
			styledLines[i] = ClaudeSquadAccentStyle.Render(line)
		case strings.HasPrefix(line, "@@"):
			// Hunk headers
			styledLines[i] = ClaudeSquadPrimaryStyle.Bold(true).Render(line)
		case strings.HasPrefix(line, "+"):
			// Added lines
			styledLines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff9d")).Render(line)
		case strings.HasPrefix(line, "-"):
			// Removed lines
			styledLines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b")).Render(line)
		case strings.HasPrefix(line, "diff --git"):
			// Diff headers
			styledLines[i] = ClaudeSquadAccentStyle.Bold(true).Render(line)
		default:
			// Context lines
			styledLines[i] = ClaudeSquadMutedStyle.Render(line)
		}
	}
	
	return strings.Join(styledLines, "\n")
}

// View renders the diff view
func (dv *DiffView) View() string {
	if dv.width == 0 || dv.height == 0 {
		return ""
	}
	
	// Create header
	headerText := "Git Diff"
	if dv.sessionName != "" {
		headerText = fmt.Sprintf("Git Diff - %s", dv.sessionName)
	}
	
	header := ClaudeSquadHeaderStyle.Render(headerText)
	
	// Show loading state
	if dv.loading {
		loadingText := ClaudeSquadMutedStyle.Render("Loading diff...")
		return lipgloss.JoinVertical(lipgloss.Left, header, loadingText)
	}
	
	// Show content
	viewportView := dv.viewport.View()
	
	// Create bordered view
	diffContent := lipgloss.JoinVertical(lipgloss.Left, header, viewportView)
	
	return ClaudeSquadBorderStyle.
		Width(dv.width-2).
		Height(dv.height-2).
		Render(diffContent)
}

// GetSessionName returns the currently loaded session name
func (dv *DiffView) GetSessionName() string {
	return dv.sessionName
}

// IsLoading returns whether the diff view is currently loading
func (dv *DiffView) IsLoading() bool {
	return dv.loading
}