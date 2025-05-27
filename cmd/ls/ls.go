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

	cmd := exec.Command("git", "diff", "--shortstat")
	cmd.Dir = sessionState.WorktreePath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "0", "0"
	}
	output := out.String()

	insertions := "0"
	deletions := "0"

	insRe := regexp.MustCompile(`(\d+) insertion`) // matches "12 insertions(+)"
	delRe := regexp.MustCompile(`(\d+) deletion`)  // matches "3 deletions(-)"

	if m := insRe.FindStringSubmatch(output); len(m) > 1 {
		insertions = m[1]
	}
	if m := delRe.FindStringSubmatch(output); len(m) > 1 {
		deletions = m[1]
	}

	return insertions, deletions
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

	// Print header
	fmt.Printf("%-30s %-15s %-10s %-30s %s\n",
		"AGENT",
		"BRANCH", 
		"STATUS",
		"PROMPT",
		"CHANGES",
	)
	fmt.Println(strings.Repeat("─", 100))

	// Print the active sessions in a single line each
	for _, session := range activeSessions {
		// Get session state
		states := make(map[string]state.AgentState)
		if data, err := os.ReadFile(stateManager.GetStatePath()); err == nil {
			if err := json.Unmarshal(data, &states); err == nil {
				if state, ok := states[session]; ok {
					// Truncate prompt if too long
					prompt := state.Prompt
					if len(prompt) > 27 {
						prompt = prompt[:24] + "..."
					}

					// Format status with styling
					var statusDisplay string
					switch state.Status {
					case "Running":
						statusDisplay = statusRunningStyle.Render("Running")
					case "Ready":
						statusDisplay = statusReadyStyle.Render("Ready")
					default:
						statusDisplay = statusUnknownStyle.Render("Unknown")
					}

					// Get git diff totals
					insertions, deletions := getGitDiffTotals(session, stateManager)
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

					fmt.Printf("%-30s %-15s %-10s %-30s %s\n",
						agentStyle.Render(session),
						branchStyle.Render(state.BranchFrom),
						statusDisplay,
						promptStyle.Render(prompt),
						diffStats,
					)
				}
			}
		}
	}

	return nil
}
