# Tmux Session Discovery Helper

This module provides functionality to discover and analyze tmux sessions, specifically designed to enhance the Uzi TUI by highlighting attached/active sessions.

## Overview

The tmux discovery helper calls `exec.Command("tmux","ls")`, parses window names, and maps them to Uzi sessions so the TUI list view can highlight which sessions are attached or active.

## Core Components

### `TmuxDiscovery`

The main helper struct that provides tmux session discovery functionality:

```go
type TmuxDiscovery struct {
    lastUpdate time.Time
    sessions    map[string]TmuxSessionInfo
    cacheTime   time.Duration
}
```

**Key Features:**
- Caches results for 2 seconds to avoid excessive tmux calls
- Discovers all tmux sessions using `tmux list-sessions`
- Parses window names and session metadata
- Identifies Uzi-specific sessions

### `TmuxSessionInfo`

Represents detailed information about a tmux session:

```go
type TmuxSessionInfo struct {
    Name        string    `json:"name"`
    Windows     int       `json:"windows"`
    Panes       int       `json:"panes"`
    Attached    bool      `json:"attached"`
    Created     time.Time `json:"created"`
    LastUsed    time.Time `json:"last_used"`
    WindowNames []string  `json:"window_names"`
    Activity    string    `json:"activity"` // "active", "inactive", "attached"
}
```

## Key Methods

### Session Discovery

- `GetAllSessions()` - Returns all tmux sessions
- `GetUziSessions()` - Returns only Uzi agent sessions
- `MapUziSessionsToTmux()` - Maps Uzi sessions to tmux session info

### Session Status

- `IsSessionAttached(sessionName)` - Check if session is currently attached
- `GetSessionActivity(sessionName)` - Get activity level ("attached", "active", "inactive")
- `GetSessionStatus(sessionName)` - Detailed status including agent window analysis

### Activity Analysis

- `ListSessionsByActivity()` - Group sessions by activity level
- `GetAttachedSessionCount()` - Count of currently attached sessions
- `FormatSessionActivity(activity)` - Visual indicators (🔗, ●, ○)

## Integration with Uzi

The tmux discovery helper is integrated into the `UziCLI` interface:

```go
type UziCLI struct {
    stateManager  *state.StateManager
    tmuxDiscovery *TmuxDiscovery
}
```

### Enhanced Methods

- `GetSessionsWithTmuxInfo()` - Returns sessions with tmux attachment information
- `IsSessionAttached(sessionName)` - Direct tmux attachment check
- `RefreshTmuxCache()` - Force refresh of tmux session cache

## TUI Integration

The TUI list view uses `SessionListItem` to display sessions with tmux highlighting:

```go
type SessionListItem struct {
    session    SessionInfo
    tmuxInfo   *TmuxSessionInfo
    uziCLI     *UziCLI
}
```

**Visual Indicators:**
- 🔗 - Session is currently attached
- ● - Session is active (recent activity)
- ○ - Session is inactive

## Session Detection Logic

### Uzi Session Identification

A tmux session is considered a Uzi session if:

1. **Name Pattern**: Starts with "agent-" and has at least 4 parts: `agent-projectDir-gitHash-agentName`
2. **Window Names**: Contains "agent" or "uzi-dev" windows

### Activity Classification

- **attached**: Session is currently attached in tmux
- **active**: Session has activity within the last 5 minutes
- **inactive**: Session exists but has no recent activity

### Status Detection

The helper analyzes agent window content to determine detailed status:

- **attached**: Session is currently attached
- **running**: Agent pane contains "esc to interrupt", "Thinking", or "Working"
- **ready**: Session exists and agent is waiting for input
- **not_found**: Session doesn't exist in tmux

## Usage Examples

### Basic Discovery

```go
tmuxDiscovery := NewTmuxDiscovery()

// Get all sessions
allSessions, err := tmuxDiscovery.GetAllSessions()

// Get only Uzi sessions
uziSessions, err := tmuxDiscovery.GetUziSessions()

// Check if specific session is attached
isAttached := tmuxDiscovery.IsSessionAttached("agent-myproject-abc123-claude")
```

### TUI Integration

```go
uziCLI := NewUziCLI()

// Get sessions with tmux info
sessions, tmuxMapping, err := uziCLI.GetSessionsWithTmuxInfo()

// Create enhanced list items
for _, session := range sessions {
    var tmuxInfo *TmuxSessionInfo
    if info, exists := tmuxMapping[session.Name]; exists {
        tmuxInfo = &info
    }
    
    listItem := NewSessionListItem(session, tmuxInfo, uziCLI)
    // listItem.Title() includes activity indicator
    // listItem.Description() includes enhanced status
}
```

### Activity Monitoring

```go
// Get activity summary
sessionsByActivity, err := tmuxDiscovery.ListSessionsByActivity()
attachedCount, err := tmuxDiscovery.GetAttachedSessionCount()

// Print summary
for activity, sessions := range sessionsByActivity {
    fmt.Printf("%s: %d sessions\n", activity, len(sessions))
}
```

## Error Handling

The tmux discovery helper is designed to be resilient:

- Returns empty results instead of errors when tmux has no sessions
- Continues processing even if individual sessions fail to parse
- Gracefully handles missing tmux installation
- Uses cached results when tmux calls fail

## Performance Considerations

- **Caching**: Results are cached for 2 seconds to minimize tmux calls
- **Lazy Loading**: Window information is only fetched when needed
- **Batch Operations**: Uses tmux format strings to get multiple values in single calls
- **Error Recovery**: Failed tmux calls don't block the entire discovery process

## Dependencies

- `os/exec` for calling tmux commands
- No external dependencies beyond Go standard library
- Integrates with existing Uzi state management (`pkg/state`)

## Future Enhancements

Potential improvements for the tmux discovery helper:

1. **Real-time Updates**: Watch for tmux session changes
2. **Advanced Filtering**: More sophisticated Uzi session detection
3. **Performance Metrics**: CPU/memory usage tracking per session
4. **Session Health**: Detect stuck or problematic sessions
5. **Multi-server Support**: Handle tmux sessions across multiple servers
