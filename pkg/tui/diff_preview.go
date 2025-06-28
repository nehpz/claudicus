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
	content       string
	commitMessages string
	changedFiles  string
	error        string
	width        int
	height       int
	showCommits   bool // Toggle to show commits and files or just diff
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
		m.commitMessages = ""
		m.changedFiles = ""
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
		m.commitMessages = ""
		m.changedFiles = ""
		return
	}

	// Get commit messages and changed files
	commitMessages, err := m.getCommitMessages(session.WorktreePath)
	if err != nil {
		// Don't fail entirely if commit messages fail
		m.commitMessages = fmt.Sprintf("Error loading commits: %v", err)
	} else {
		m.commitMessages = commitMessages
	}

	changedFiles, err := m.getChangedFiles(session.WorktreePath)
	if err != nil {
		// Don't fail entirely if changed files fail
		m.changedFiles = fmt.Sprintf("Error loading changed files: %v", err)
	} else {
		m.changedFiles = changedFiles
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

// getCommitMessages executes git log command and returns the last 3 commit messages
func (m *DiffPreviewModel) getCommitMessages(worktreePath string) (string, error) {
	if worktreePath == "" {
		return "No worktree path available", nil
	}

	// Get last 3 commit messages with compact format
	cmd := exec.Command("git", "--no-pager", "log", "-n3", "--pretty=format:%h %s (%an, %ar)")
	cmd.Dir = worktreePath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "No commit history", nil
	}

	return result, nil
}

// getChangedFiles executes git diff command and returns list of changed files
func (m *DiffPreviewModel) getChangedFiles(worktreePath string) (string, error) {
	if worktreePath == "" {
		return "No worktree path available", nil
	}

	// Get list of changed files with status
	cmd := exec.Command("sh", "-c", "git status --porcelain")
	cmd.Dir = worktreePath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "No changed files", nil
	}

	return result, nil
}

// ToggleView toggles between showing commits/files and diff
func (m *DiffPreviewModel) ToggleView() {
	m.showCommits = !m.showCommits
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

	// Create title header based on current view
	title := "Git Diff"
	if m.showCommits {
		title = "Commits & Files"
	}
	titleHeader := ClaudeSquadHeaderStyle.Render(title)

	// Handle error case
	if m.error != "" {
		errorContent := ClaudeSquadMutedStyle.Render(m.error)
		content := lipgloss.JoinVertical(lipgloss.Left, titleHeader, errorContent)
		return borderStyle.Render(content)
	}

	// Handle empty content
	if m.content == "" && m.commitMessages == "" && m.changedFiles == "" {
		emptyContent := ClaudeSquadMutedStyle.Render("Select an agent to view diff\nPress 'v' to toggle commits/files view")
		content := lipgloss.JoinVertical(lipgloss.Left, titleHeader, emptyContent)
		return borderStyle.Render(content)
	}

	var formattedContent string
	if m.showCommits {
		// Show commits and changed files
		formattedContent = m.formatCommitsAndFiles()
	} else {
		// Show diff content with syntax highlighting
		formattedContent = m.formatDiffContent(m.content)
	}
	
	// Join title and content
	content := lipgloss.JoinVertical(lipgloss.Left, titleHeader, formattedContent)
	return borderStyle.Render(content)
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

// formatCommitsAndFiles formats commits and changed files for display
func (m *DiffPreviewModel) formatCommitsAndFiles() string {
	var sections []string

	// Format commit messages section
	if m.commitMessages != "" {
		commitHeader := ClaudeSquadAccentStyle.Render("Recent Commits:")
		commitLines := strings.Split(m.commitMessages, "\n")
		var formattedCommits []string
		for _, line := range commitLines {
			if strings.TrimSpace(line) != "" {
				formattedCommits = append(formattedCommits, ClaudeSquadMutedStyle.Render("  "+line))
			}
		}
		commitSection := commitHeader + "\n" + strings.Join(formattedCommits, "\n")
		sections = append(sections, commitSection)
	}

	// Format changed files section
	if m.changedFiles != "" {
		filesHeader := ClaudeSquadPrimaryStyle.Render("Changed Files:")
		fileLines := strings.Split(m.changedFiles, "\n")
		var formattedFiles []string
		for _, line := range fileLines {
			if strings.TrimSpace(line) != "" {
				// Parse git status format (e.g., " M file.go", "A  new.go")
				if len(line) >= 3 {
					status := line[:2]
					file := line[3:]
					var statusColor lipgloss.Color
					switch {
					case strings.Contains(status, "A"):
						statusColor = lipgloss.Color("#00ff9d") // Green for added
					case strings.Contains(status, "M"):
						statusColor = lipgloss.Color("#ffa500") // Orange for modified
					case strings.Contains(status, "D"):
						statusColor = lipgloss.Color("#ff6b6b") // Red for deleted
					default:
						statusColor = lipgloss.Color("#888888") // Gray for other
					}
					statusStyled := lipgloss.NewStyle().Foreground(statusColor).Render(status)
					formattedFiles = append(formattedFiles, "  "+statusStyled+" "+file)
				} else {
					formattedFiles = append(formattedFiles, ClaudeSquadMutedStyle.Render("  "+line))
				}
			}
		}
		filesSection := filesHeader + "\n" + strings.Join(formattedFiles, "\n")
		sections = append(sections, filesSection)
	}

	// Add helpful instructions at the bottom
	if len(sections) > 0 {
		instructions := ClaudeSquadMutedStyle.Render("\nPress 'v' to toggle back to diff view")
		sections = append(sections, instructions)
	}

	// Join sections with spacing
	result := strings.Join(sections, "\n\n")

	// Limit the number of lines to fit in the view
	maxLines := m.height - 4 // Account for border and padding
	lines := strings.Split(result, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, ClaudeSquadMutedStyle.Render("... (truncated)"))
		result = strings.Join(lines, "\n")
	}

	return result
}
