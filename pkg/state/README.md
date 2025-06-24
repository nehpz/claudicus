# State Reader for Uzi

This package provides functionality to read and parse Uzi's state.json file located at `$REPO/.uzi/state.json`. It reuses the existing `pkg/state` types and provides a clean interface for retrieving session information with agent names, status, dev server URLs, and git diff counts.

## Features

- **Reuses existing types**: Built on top of the existing `AgentState` struct from `pkg/state`
- **Git diff calculation**: Uses `git diff --stat` to calculate insertion/deletion counts
- **Session status detection**: Checks tmux sessions to determine if agents are active/running/ready
- **Dev server URL formatting**: Formats dev server URLs when ports are available
- **Agent name extraction**: Properly extracts agent names from session identifiers

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/devflowinc/uzi/pkg/state"
)

func main() {
    // Create a new state reader for the current repository
    reader := state.NewStateReader("/path/to/repo")
    
    // Load all sessions from state.json
    sessions, err := reader.LoadSessions()
    if err != nil {
        log.Fatalf("Failed to load sessions: %v", err)
    }
    
    // Print session information
    for _, session := range sessions {
        fmt.Printf("Agent: %s, Status: %s, Diff: +%d/-%d\n",
            session.AgentName, session.Status, 
            session.Insertions, session.Deletions)
    }
}
```

### Get Only Active Sessions

```go
// Get only sessions that are currently active in tmux
activeSessions, err := reader.GetActiveSessions()
if err != nil {
    log.Fatalf("Failed to get active sessions: %v", err)
}

fmt.Printf("Found %d active sessions\n", len(activeSessions))
```

## SessionInfo Structure

The `SessionInfo` struct contains all the information needed for displaying session data:

```go
type SessionInfo struct {
    SessionName  string `json:"session_name"`   // Full session identifier
    AgentName    string `json:"agent_name"`     // Extracted agent name
    Status       string `json:"status"`         // inactive/ready/running/unknown
    DevServerURL string `json:"dev_server_url,omitempty"` // http://localhost:PORT
    Model        string `json:"model"`          // AI model being used
    Prompt       string `json:"prompt"`         // Original prompt/task
    Insertions   int    `json:"insertions"`     // Git diff insertions
    Deletions    int    `json:"deletions"`      // Git diff deletions
    WorktreePath string `json:"worktree_path"`  // Path to git worktree
    CreatedAt    string `json:"created_at"`     // ISO8601 timestamp
    UpdatedAt    string `json:"updated_at"`     // ISO8601 timestamp
}
```

## Status Values

- **`inactive`**: tmux session doesn't exist or has stopped
- **`ready`**: tmux session exists and agent is waiting for input
- **`running`**: tmux session exists and agent is actively working
- **`unknown`**: tmux session exists but status cannot be determined

## Git Diff Calculation

The reader calculates git diff statistics by:

1. Checking if the worktree path exists
2. Running `git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null`
3. Parsing the output to extract insertion and deletion counts
4. Using regex patterns to match the git shortstat format

## Agent Name Extraction

Agent names are extracted from session names using the format:
`agent-projectDir-gitHash-agentName`

For example: `agent-myproject-abc123-claude` → `claude`

## Error Handling

The reader follows the "nail it before we scale it" principle:
- Returns empty results for missing files rather than erroring
- Continues processing other sessions if one fails
- Lets git handle worktree and diff errors naturally
- Provides clear error messages for actual failures

## Integration with Existing Code

This state reader is designed to integrate seamlessly with the existing Uzi codebase:

- Uses the same `AgentState` struct from `pkg/state/state.go`
- Follows the same session naming conventions
- Replicates the git diff logic from `cmd/ls/ls.go`
- Uses the same tmux status detection methods

## Examples

See the `examples/state_reader/` directory for complete working examples:

- `main.go`: Basic usage example
- `test_with_mock_data.go`: Example with mock data for testing

## Future Enhancements

Potential improvements that could be added later:
- Repository filtering based on git remote URL
- Caching for improved performance
- Watch mode for real-time updates
- Additional status detection patterns
- Performance metrics (CPU/memory usage)
