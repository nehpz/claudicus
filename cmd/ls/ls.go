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
	"text/tabwriter"
	"time"

	"github.com/devflowinc/uzi/pkg/config"
	"github.com/devflowinc/uzi/pkg/state"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	fs          = flag.NewFlagSet("uzi ls", flag.ExitOnError)
	configPath  = fs.String("config", config.GetDefaultConfigPath(), "path to config file")
	allSessions = fs.Bool("a", false, "show all sessions including inactive")
	watchMode   = fs.Bool("w", false, "watch mode - refresh output every second")
	jsonOutput  = fs.Bool("json", false, "output in JSON format")
	CmdLs       = &ffcli.Command{
		Name:       "ls",
		ShortUsage: "uzi ls [-a] [-w] [--json]",
		ShortHelp:  "List active agent sessions",
		FlagSet:    fs,
		Exec:       executeLs,
	}
)

func getGitDiffTotals(sessionName string, stateManager *state.StateManager) (int, int) {
	// Get session state to find worktree path
	states := make(map[string]state.AgentState)
	if data, err := os.ReadFile(stateManager.GetStatePath()); err != nil {
		return 0, 0
	} else {
		if err := json.Unmarshal(data, &states); err != nil {
			return 0, 0
		}
	}

	sessionState, ok := states[sessionName]
	if !ok || sessionState.WorktreePath == "" {
		return 0, 0
	}

	shellCmdString := "git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null"

	cmd := exec.Command("sh", "-c", shellCmdString)
	cmd.Dir = sessionState.WorktreePath

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, 0
	}

	output := out.String()

	insertions := 0
	deletions := 0

	insRe := regexp.MustCompile(`(\d+) insertion(?:s)?\(\+\)`)
	delRe := regexp.MustCompile(`(\d+) deletion(?:s)?\(\-\)`)

	if m := insRe.FindStringSubmatch(output); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &insertions)
	}
	if m := delRe.FindStringSubmatch(output); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &deletions)
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
		return "unknown"
	}

	if strings.Contains(content, "esc to interrupt") || strings.Contains(content, "Thinking") {
		return "running"
	}
	return "ready"
}

func formatStatus(status string) string {
	switch status {
	case "ready":
		return "\033[32mready\033[0m" // Green
	case "running":
		return "\033[33mrunning\033[0m" // Orange/Yellow
	default:
		return status
	}
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Hour {
		return fmt.Sprintf("%2dm", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%2dh", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%2dd", int(diff.Hours()/24))
	}
	return t.Format("Jan 02")
}

// SessionInfo represents session data for JSON output
// This matches the struct used in pkg/tui/uzi_interface.go
type SessionInfo struct {
	Name         string `json:"name"`
	AgentName    string `json:"agent_name"`
	Model        string `json:"model"`
	Status       string `json:"status"`
	Prompt       string `json:"prompt"`
	Insertions   int    `json:"insertions"`
	Deletions    int    `json:"deletions"`
	WorktreePath string `json:"worktree_path"`
	Port         int    `json:"port,omitempty"`
	UpdatedAt    string `json:"updated_at"`
}

func getSessionsAsJSON(stateManager *state.StateManager, activeSessions []string) ([]SessionInfo, error) {
	// Load all states
	states := make(map[string]state.AgentState)
	if data, err := os.ReadFile(stateManager.GetStatePath()); err == nil {
		if err := json.Unmarshal(data, &states); err != nil {
			return nil, fmt.Errorf("error parsing state file: %w", err)
		}
	}

	var sessions []SessionInfo
	for _, sessionName := range activeSessions {
		state, ok := states[sessionName]
		if !ok {
			continue
		}

		// Extract agent name from session name
		parts := strings.Split(sessionName, "-")
		agentName := sessionName
		if len(parts) >= 4 && parts[0] == "agent" {
			agentName = strings.Join(parts[3:], "-")
		}

		status := getAgentStatus(sessionName)
		insertions, deletions := getGitDiffTotals(sessionName, stateManager)

		// Get model name, default to "unknown" if empty
		model := state.Model
		if model == "" {
			model = "unknown"
		}

		sessionInfo := SessionInfo{
			Name:         sessionName,
			AgentName:    agentName,
			Model:        model,
			Status:       status,
			Prompt:       state.Prompt,
			Insertions:   insertions,
			Deletions:    deletions,
			WorktreePath: state.WorktreePath,
			Port:         state.Port,
			UpdatedAt:    state.UpdatedAt.Format(time.RFC3339),
		}
		sessions = append(sessions, sessionInfo)
	}

	// Sort by UpdatedAt (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		t1, err1 := time.Parse(time.RFC3339, sessions[i].UpdatedAt)
		t2, err2 := time.Parse(time.RFC3339, sessions[j].UpdatedAt)
		if err1 != nil || err2 != nil {
			return false
		}
		return t1.After(t2)
	})

	return sessions, nil
}

func printSessionsJSON(stateManager *state.StateManager, activeSessions []string) error {
	sessions, err := getSessionsAsJSON(stateManager, activeSessions)
	if err != nil {
		return err
	}

	// Output JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sessions)
}

func printSessions(stateManager *state.StateManager, activeSessions []string) error {
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

	// Long format with tabwriter for alignment
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintf(w, "AGENT\tMODEL\tSTATUS    DIFF\tADDR\tPROMPT\n")

	// Print sessions
	for _, session := range sessions {
		sessionName := session.name
		state := session.state

		// Extract agent name from session name
		parts := strings.Split(sessionName, "-")
		agentName := sessionName
		if len(parts) >= 4 && parts[0] == "agent" {
			agentName = strings.Join(parts[3:], "-")
		}

		status := getAgentStatus(sessionName)
		insertions, deletions := getGitDiffTotals(sessionName, stateManager)

		// Format diff stats with colors
		var changes string
		if insertions == 0 && deletions == 0 {
			changes = "\033[32m+0\033[0m/\033[31m-0\033[0m"
		} else {
			// ANSI color codes: green for additions, red for deletions
			changes = fmt.Sprintf("\033[32m+%d\033[0m/\033[31m-%d\033[0m", insertions, deletions)
		}

		// Get model name, default to "unknown" if empty (for backward compatibility)
		model := state.Model
		if model == "" {
			model = "unknown"
		}

		// Format: agent model status addr changes prompt
		addr := ""
		if state.Port != 0 {
			addr = fmt.Sprintf("http://localhost:%d", state.Port)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			agentName,
			model,
			formatStatus(status),
			changes,
			addr,
			state.Prompt,
		)
	}
	w.Flush()

	return nil
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func executeLs(ctx context.Context, args []string) error {
	stateManager := state.NewStateManager()
	if stateManager == nil {
		return fmt.Errorf("failed to create state manager")
	}

	if *watchMode {
		// Watch mode - refresh every second
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		// Initial display
		clearScreen()
		activeSessions, err := stateManager.GetActiveSessionsForRepo()
		if err != nil {
			return fmt.Errorf("error getting active sessions: %w", err)
		}

		if len(activeSessions) == 0 {
			fmt.Println("No active sessions found")
		} else {
			if err := printSessions(stateManager, activeSessions); err != nil {
				return err
			}
		}

		// Watch loop
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				clearScreen()
				activeSessions, err := stateManager.GetActiveSessionsForRepo()
				if err != nil {
					fmt.Printf("Error getting active sessions: %v\n", err)
					continue
				}

				if len(activeSessions) == 0 {
					fmt.Println("No active sessions found")
				} else {
					if err := printSessions(stateManager, activeSessions); err != nil {
						fmt.Printf("Error printing sessions: %v\n", err)
					}
				}
			}
		}
	} else {
		// Single run mode
		activeSessions, err := stateManager.GetActiveSessionsForRepo()
		if err != nil {
			return fmt.Errorf("error getting active sessions: %w", err)
		}

		if len(activeSessions) == 0 {
			if *jsonOutput {
				// Return empty JSON array
				fmt.Println("[]")
			} else {
				fmt.Println("No active sessions found")
			}
			return nil
		}

		if *jsonOutput {
			return printSessionsJSON(stateManager, activeSessions)
		} else {
			return printSessions(stateManager, activeSessions)
		}
	}
}
