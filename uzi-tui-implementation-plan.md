# Uzi-TUI Implementation Plan

## Project Overview

Create a Terminal User Interface (TUI) for Uzi by layering Claude Squad's polished interface on top of Uzi's core functionality. This provides a smooth, non-flashing UI while maintaining Uzi's simplicity and speed.

## Goals

- **Primary**: Replace Uzi's flashing `ls -w` with Claude Squad's smooth TUI
- **Secondary**: Provide interactive session management without changing Uzi's core workflow
- **Maintain**: Uzi's auto-yes logic and simple state management

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   User      │     │  Uzi TUI    │     │ Uzi Core    │
│  Keyboard   │────▶│  (Bubble    │────▶│  Commands   │
│   Input     │     │   Tea UI)   │     │  (CLI)      │
└─────────────┘     └─────────────┘     └─────────────┘
                            │                    │
                            ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐
                    │   Display   │     │ State.json  │
                    │  Sessions   │◀────│   Tmux      │
                    └─────────────┘     └─────────────┘
```

## Implementation Phases

### Phase 1: Foundation (Week 1)

#### 1.1 Project Setup

- [ ] Fork Uzi repository
- [ ] Add Bubble Tea dependency: `go get github.com/charmbracelet/bubbletea`
- [ ] Create `tui/` package structure
- [ ] Add `uzi tui` command to main.go

#### 1.2 Basic TUI Framework

- [ ] Port Claude Squad's list view component
- [ ] Create state reader for Uzi's state.json
- [ ] Implement session discovery from tmux
- [ ] Basic keyboard navigation (j/k, up/down)

#### 1.3 Session Display

- [ ] Show agent name, status, git diff stats
- [ ] Port Claude Squad's styling/colors
- [ ] Add refresh ticker (non-flashing updates)
- [ ] Display dev server URLs from state

### Phase 2: Command Integration (Week 2)

#### 2.1 Hotkey Mapping Implementation

| Key | Function    | Uzi Command                             |
| --- | ----------- | --------------------------------------- |
| n   | New agents  | `uzi prompt --agents claude:N "prompt"` |
| b   | Broadcast   | `uzi broadcast "message"`               |
| k   | Kill agent  | `uzi kill <agent-name>`                 |
| c   | Checkpoint  | `uzi checkpoint <agent-name> "message"` |
| r   | Run command | `uzi run "command"`                     |
| a   | Toggle auto | Start/stop auto watcher                 |
| o/↵ | Attach      | Direct tmux attach                      |
| q   | Quit        | Exit TUI                                |

#### 2.2 Input Prompts

- [ ] Number input for agent count (n key)
- [ ] Text input for prompts/messages
- [ ] Confirmation dialogs for destructive actions
- [ ] Port Claude Squad's overlay system

#### 2.3 Command Execution

- [ ] Implement exec.Command wrapper
- [ ] Handle command success/failure
- [ ] Refresh UI after state changes
- [ ] Show command feedback to user

### Phase 3: Auto Mode & Polish (Week 3)

#### 3.1 Auto Mode Integration

- [ ] Toggle Uzi's auto watcher from TUI
- [ ] Visual indicator for auto mode status
- [ ] Show "AUTO" badge in header
- [ ] Ensure watcher runs in background

#### 3.2 Enhanced Features

- [ ] Port diff preview pane (tab to switch)
- [ ] Add help screen (? key)
- [ ] Session content preview
- [ ] Error handling and recovery

#### 3.3 Final Polish

- [ ] Smooth animations/transitions
- [ ] Proper cleanup on exit
- [ ] Handle terminal resize
- [ ] Comprehensive keyboard shortcuts

## Technical Details

### File Structure

```
uzi/
├── cmd/
│   ├── prompt/
│   ├── kill/
│   ├── ...
│   └── tui/          # NEW
│       └── tui.go
├── pkg/
│   ├── agents/
│   ├── config/
│   ├── state/
│   └── tui/          # NEW
│       ├── app.go
│       ├── list.go
│       ├── preview.go
│       ├── keys.go
│       ├── prompts.go
│       ├── styles.go
│       └── uzi_interface.go  # ABSTRACTION LAYER
└── uzi.go
```

### Abstraction Layer Design

To ensure Uzi core remains untouched and upstream merges are conflict-free:

```go
// pkg/tui/uzi_interface.go
// This is the ONLY file that imports Uzi packages directly
// Acts as a bridge between TUI and Uzi core

package tui

import (
    "github.com/devflowinc/uzi/pkg/state"
    "github.com/devflowinc/uzi/pkg/config"
    "os/exec"
)

// UziInterface abstracts all Uzi operations
type UziInterface interface {
    // State operations
    GetActiveSessionsForRepo() ([]string, error)
    GetSessionState(sessionName string) (*SessionInfo, error)
    
    // Command operations
    RunPrompt(agents string, prompt string) error
    RunBroadcast(message string) error
    RunKill(agentName string) error
    RunCheckpoint(agentName string, message string) error
    RunCommand(command string) error
    
    // Auto mode
    StartAutoMode() error
    StopAutoMode() error
    IsAutoModeRunning() bool
}

// UziCLI implements UziInterface by calling CLI commands
type UziCLI struct {
    stateManager *state.StateManager
    config       *config.Config
    autoProcess  *exec.Cmd
}

// Implementation calls Uzi CLI - no logic duplication
func (u *UziCLI) RunPrompt(agents string, prompt string) error {
    cmd := exec.Command("uzi", "prompt", "--agents", agents, prompt)
    return cmd.Run()
}
```

### Key Components

#### 1. State Integration

```go
type UziTUI struct {
    uzi          UziInterface         // Abstraction layer
    list         *ListView            // From Claude Squad
    preview      *PreviewPane         // From Claude Squad
}
```

#### 2. Command Execution

```go
// All commands go through the abstraction layer
func (app *UziTUI) executeCommand(action func() error) tea.Cmd {
    return func() tea.Msg {
        if err := action(); err != nil {
            return errMsg{err}
        }
        return refreshMsg{}
    }
}

// Example usage:
func (app *UziTUI) handleNewAgents(count int, prompt string) tea.Cmd {
    return app.executeCommand(func() error {
        return app.uzi.RunPrompt(fmt.Sprintf("claude:%d", count), prompt)
    })
}
```

#### 3. Hotkey Handler

```go
func (u *UziTUI) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
    switch msg.String() {
    case "n":
        return u.promptNewAgents()
    case "b":
        return u.promptBroadcast()
    case "k":
        return u.killSelectedAgent()
    // ... etc
    }
}
```

## Dependencies

### From Uzi (keep as-is)

- State management (`pkg/state`)
- Agent utilities (`pkg/agents`)
- Config management (`pkg/config`)
- Auto watcher (`cmd/watch`)

### From Claude Squad (adapt)

- Bubble Tea TUI framework
- List view component
- Preview/diff display
- Keyboard handling
- Style definitions

### New Dependencies

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.17.1
)
```

## Testing Plan

### Week 1 Tests

- [ ] TUI launches without errors
- [ ] Sessions display correctly
- [ ] Navigation works (j/k, arrows)
- [ ] State.json is read properly

### Week 2 Tests

- [ ] All hotkeys trigger correct commands
- [ ] Input prompts work correctly
- [ ] Commands execute successfully
- [ ] UI refreshes after state changes

### Week 3 Tests

- [ ] Auto mode toggles properly
- [ ] Help screen displays
- [ ] Diff preview works
- [ ] No memory leaks
- [ ] Graceful error handling

## Success Criteria

1. **No Flashing**: Smooth updates without screen clearing
2. **Feature Parity**: All Uzi commands accessible via hotkeys
3. **Responsive**: < 100ms input latency
4. **Stable**: No crashes during normal operation
5. **Intuitive**: New users can use without reading docs

## Risk Mitigation

### Low Risk Items

- UI components (well-tested in Claude Squad)
- State reading (read-only operations)
- Tmux integration (both tools already do this)

### Medium Risk Items

- Command execution error handling
- Auto mode integration
- Cross-platform compatibility

### Mitigation Strategies

1. Extensive error handling for command execution
2. Graceful degradation if features unavailable
3. Clear error messages to user
4. Timeout handling for stuck commands

## Future Enhancements (Post-MVP)

1. **Multi-repo Support**: Show agents across repositories
2. **Themes**: Customizable color schemes
3. **Advanced Filters**: Filter by status, model, etc.
4. **Batch Operations**: Select multiple agents for commands
5. **Config UI**: Edit uzi.yaml from TUI

## Development Guidelines
1. **Minimal Changes to Uzi Core**: All changes in new `tui/` package
2. **Preserve Uzi Philosophy**: Keep it simple and fast
3. **Reuse Existing Code**: Don't reimplement Uzi logic
4. **Clear Separation**: TUI is just a UI layer
5. **Backward Compatible**: CLI commands still work

## Abstraction Layer Benefits

By using the `UziInterface` abstraction:

1. **Zero Changes to Uzi Core**: The TUI is a completely separate package
2. **Easy Upstream Merges**: No merge conflicts when pulling Uzi updates
3. **Clean Testing**: Can mock `UziInterface` for unit tests
4. **Future Flexibility**: Can swap implementations (e.g., direct API calls vs CLI)

### Fork Strategy

```bash
# Fork Uzi but keep it pristine
git clone https://github.com/devflowinc/uzi
cd uzi
git remote add upstream https://github.com/devflowinc/uzi
git checkout -b feature/tui

# All TUI code goes in pkg/tui/ and cmd/tui/
# NEVER modify existing Uzi files

# To update from upstream:
git fetch upstream
git merge upstream/main  # Should be conflict-free!
```

### Minimal Changes Required

Only TWO changes to existing Uzi files:

1. **uzi.go** - Add TUI command:
```go
subcommands = append(subcommands, tui.CmdTUI)
```

2. **go.mod** - Add Bubble Tea dependencies

That's it! Everything else is isolated in the TUI package.
## Estimated Timeline

- **Week 1**: Basic TUI with session display
- **Week 2**: Full command integration
- **Week 3**: Auto mode, polish, and testing
- **Buffer**: 1 week for unexpected issues

**Total: 3-4 weeks to production-ready**

## Getting Started

```bash
# Clone and setup
git clone https://github.com/devflowinc/uzi
cd uzi
git checkout -b feature/tui

# Add dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles

# Create TUI structure
mkdir -p cmd/tui pkg/tui

# Start with basic TUI
# Copy this plan to project root
cp uzi-tui-implementation-plan.md .

# Begin implementation...
```

EOF < /dev/null
