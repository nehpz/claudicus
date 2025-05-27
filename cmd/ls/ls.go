package ls

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"uzi/pkg/config"
	"uzi/pkg/state"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs         = flag.NewFlagSet("uzi ls", flag.ExitOnError)
	configPath = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	CmdLs      = &ffcli.Command{
		Name:       "ls",
		ShortUsage: "uzi ls",
		ShortHelp:  "List files in the current directory",
		FlagSet:    fs,
		Exec:       executeLs,
	}
)

var (
	agentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)

	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	addedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#059669"))

	removedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DC2626"))

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EAB308")).
				Bold(true)

	statusReadyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22C55E")).
				Bold(true)

	statusUnknownStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280"))
)

func getGitDiffTotals(sessionName string, stateManager *state.StateManager) (string, string) {
	// Get session state to find worktree path
	states := make(map[string]state.AgentState)
	if data, err := os.ReadFile(stateManager.GetStatePath()); err != nil {
		return "0", "0"
	} else {
		if err := json.Unmarshal(data, &states); err != nil {
			return "0", "0"
		}
	}

	sessionState, ok := states[sessionName]
	if !ok || sessionState.WorktreePath == "" {
		return "0", "0"
	}

	shellCmdString := "git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null"

	cmd := exec.Command("sh", "-c", shellCmdString)
	cmd.Dir = sessionState.WorktreePath

	var out bytes.Buffer
	var stderr bytes.Buffer // Capture stderr for debugging if needed
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Error running git command sequence: %v\nStderr: %s\nStdout (if any from intermediate steps before reset's >/dev/null): %s", err, stderr.String(), out.String())
		return "0", "0"
	}

	output := out.String()

	insertions := "0"
	deletions := "0"

	// Regexes are fine as they were
	insRe := regexp.MustCompile(`(\d+) insertion(?:s)?\(\+\)`) // Handle singular "insertion"
	delRe := regexp.MustCompile(`(\d+) deletion(?:s)?\(\-\)`)  // Handle singular "deletion"

	if m := insRe.FindStringSubmatch(output); len(m) > 1 {
		insertions = m[1]
	}
	if m := delRe.FindStringSubmatch(output); len(m) > 1 {
		deletions = m[1]
	}

	return insertions, deletions
}

func getPaneContent(sessionName string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName+":agent", "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getAgentStatus(sessionName string) string {
	content, err := getPaneContent(sessionName)
	if err != nil {
		return "Unknown"
	}

	if strings.Contains(content, "esc to interrupt") {
		return "Running"
	}
	return "Ready"
}

func executeLs(ctx context.Context, args []string) error {
	log.Debug("Running ls command")

	stateManager := state.NewStateManager()
	if stateManager == nil {
		return fmt.Errorf("failed to create state manager")
	}

	activeSessions, err := stateManager.GetActiveSessionsForRepo()
	if err != nil {
		return fmt.Errorf("error getting active sessions: %w", err)
	}

	if len(activeSessions) == 0 {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("No active sessions found"))
		return nil
	}

	// Load all states to sort by UpdatedAt
	states := make(map[string]state.AgentState)
	if data, err := os.ReadFile(stateManager.GetStatePath()); err == nil {
		if err := json.Unmarshal(data, &states); err != nil {
			return fmt.Errorf("error parsing state file: %w", err)
		}
	}

	// Create a slice of sessions with their states for sorting
	type sessionInfo struct {
		name  string
		state state.AgentState
	}
	var sessions []sessionInfo
	for _, sessionName := range activeSessions {
		if state, ok := states[sessionName]; ok {
			sessions = append(sessions, sessionInfo{name: sessionName, state: state})
		}
	}

	// Sort by UpdatedAt (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].state.UpdatedAt.After(sessions[j].state.UpdatedAt)
	})

	// Print header
	fmt.Printf("%-30s %-15s %-20s %-10s %-30s %s\n",
		"AGENT",
		"BRANCH FROM",
		"BRANCH NAME",
		"STATUS",
		"PROMPT",
		"CHANGES",
	)
	fmt.Println(strings.Repeat("─", 120))

	// Print the active sessions in a single line each
	for _, session := range sessions {
		sessionName := session.name
		state := session.state

		// Truncate prompt if too long
		prompt := state.Prompt
		if len(prompt) > 27 {
			prompt = prompt[:24] + "..."
		}

		// Format status with styling
		var statusDisplay string
		status := getAgentStatus(sessionName)
		switch status {
		case "Running":
			statusDisplay = statusRunningStyle.Render("Running")
		case "Ready":
			statusDisplay = statusReadyStyle.Render("Ready")
		default:
			statusDisplay = statusUnknownStyle.Render("Unknown")
		}

		// Get git diff totals
		insertions, deletions := getGitDiffTotals(sessionName, stateManager)
		var diffStats string
		if insertions == "0" && deletions == "0" {
			diffStats = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("no changes")
		} else {
			parts := []string{}
			if insertions != "0" {
				parts = append(parts, addedStyle.Render("+"+insertions))
			}
			if deletions != "0" {
				parts = append(parts, removedStyle.Render("-"+deletions))
			}
			diffStats = strings.Join(parts, " ")
		}

		// Truncate branch name if too long
		branchName := state.BranchName
		if len(branchName) > 18 {
			branchName = branchName[:15] + "..."
		}

		fmt.Printf("%-30s %-15s %-20s %-10s %-30s %s\n",
			agentStyle.Render(sessionName),
			branchStyle.Render(state.BranchFrom),
			branchStyle.Render(branchName),
			statusDisplay,
			promptStyle.Render(prompt),
			diffStats,
		)
	}

	return nil
}
