# Claudicus

> **Operational Excellence + UX Excellence** - Combining Uzi's speed with Claude Squad's beautiful coordination capabilities

*The definitive platform for safe, multi-agent software development that eliminates the tradeoff between 
operational efficiency and user experience quality.*

See [PRODUCT_VISION.md](PRODUCT_VISION.md) for our complete vision of operational + UX excellence.

[![TUI Screenshot](https://img.shields.io/badge/TUI-Interactive%20Interface-blue?style=for-the-badge)](#tui-interface)
[![Go Version](https://img.shields.io/github/go-mod/go-version/nehpz/claudicus?style=for-the-badge)](https://golang.org/)
[![License](https://img.shields.io/github/license/nehpz/claudicus?style=for-the-badge)](LICENSE)

## Installation

```bash
go install github.com/nehpz/claudicus@latest
```

**Ensure GOBIN is in your PATH:**

```bash
export PATH="$PATH:$HOME/go/bin"
```

## Prerequisites

Before using Claudicus, ensure you have:

- **Git**: For version control and worktree management
- **Tmux**: For terminal session management  
- **Go**: For installation (version 1.24.3+)
- **AI tool of choice**: Such as `claude`, `cursor`, `aider`, `codex`, etc.

## Quick Start / Basic Workflow

### 1. Initialize your project

```bash
# Create claudicus.yaml configuration in your project root
echo "devCommand: npm install && npm run dev -- --port \$PORT" > claudicus.yaml
echo "portRange: 3000-3010" >> claudicus.yaml
```

### 2. Start multiple agents on a task

```bash
claudicus prompt --agents claude:3,cursor:2 "Implement user authentication with JWT"
```

### 3. Monitor agent progress (Interactive TUI)

```bash
claudicus tui  # Launch beautiful visual interface
# OR
claudicus ls -w  # Watch mode in terminal
```

### 4. Auto-handle agent prompts

```bash
claudicus watch  # Automatically handles tool confirmations
```

### 5. Send updates to all agents

```bash
claudicus broadcast "Add input validation to all forms"
```

### 6. Merge completed work

```bash
claudicus checkpoint agent-name "feat: implement user authentication"
```

## Features

### 🤖 Parallel Agent Operations

- Run multiple AI coding agents simultaneously
- Support for all major AI tools (Claude, Cursor, Aider, Codex, Gemini)
- Intelligent load balancing and resource management

### 🌳 Git Worktree Management  

- Automatic isolated Git worktrees for each agent
- Safe parallel development without conflicts
- Seamless merging and checkpoint system

### 🖥️ Tmux Session Management

- Dedicated tmux session for each agent
- Persistent terminal environments
- Easy session attachment and monitoring

### 📊 Smart Watch Mode

- Automatic handling of agent confirmations
- Intelligent trust prompt management
- Background monitoring without user intervention

### 🎨 Interactive TUI

- Beautiful terminal interface with Claude Squad styling
- Real-time status updates and progress tracking
- Syntax-highlighted diff previews
- Responsive navigation and controls

### 🔄 Checkpoint & Merge System

- Safe integration of agent changes
- Atomic commit and rebase operations
- Rollback capabilities for failed merges

## Configuration

### claudicus.yaml

Create a `claudicus.yaml` file in your project root to configure Claudicus:

```yaml
devCommand: cd myapp && npm install && npm run dev --port $PORT
portRange: 3000-3010
```

#### Configuration Fields

**`devCommand`** (required)

- The command to start your development server
- Use `$PORT` as a placeholder for the port number
- Should include all setup steps (npm install, pip install, etc.)
- Each agent runs in an isolated worktree with its own dependencies

**Examples:**

- Next.js: `npm install && npm run dev -- --port $PORT`
- Vite: `npm install && npm run dev -- --port $PORT`  
- Django: `pip install -r requirements.txt && python manage.py runserver 0.0.0.0:$PORT`
- Go: `go mod tidy && go run main.go -port $PORT`

**`portRange`** (required)

- Range of ports Claudicus can use for development servers
- Format: `start-end` (e.g., `3000-3010`)
- Ensures no port conflicts between multiple agents

## Detailed Command Reference

### Core Commands

#### `claudicus prompt` (alias: `p`)

Creates new agent sessions with the specified prompt.

```bash
claudicus prompt --agents claude:2,cursor:1 "Build a todo app with React"
claudicus p --agents random:5 "Fix all TypeScript errors"
```

**Options:**

- `--agents`: Specify agents and counts in format `agent:count[,agent:count...]`
- Use `random` as agent name for random agent names

#### `claudicus ls` (alias: `l`) 

Lists all active agent sessions with their status.

```bash
claudicus ls       # List active sessions
claudicus ls -w    # Watch mode - refreshes every second
claudicus ls --json # JSON output format
```

```text
AGENT    MODEL  STATUS    DIFF  ADDR                     PROMPT
brian    claude  ready  +0/-0  http://localhost:3003  Implement user authentication with JWT
gregory  cursor  ready  +12/-0  http://localhost:3001  Add comprehensive test coverage
```

#### `claudicus tui`

Launch the interactive Terminal User Interface for visual agent management.

```bash
claudicus tui
```

#### `claudicus watch` (alias: `w`)

Monitors all agent sessions and automatically handles prompts.

```bash
claudicus watch
```

- Auto-presses Enter for trust prompts
- Handles continuation confirmations  
- Monitors session state changes
- Refreshes active session list automatically
- Runs until interrupted (Ctrl+C)
- Replaces the deprecated "auto" command with improved functionality

#### `claudicus kill` (alias: `k`)

Terminates agent sessions and cleans up resources.

```bash
claudicus kill agent-name    # Kill specific agent
claudicus kill all          # Kill all agents
```

#### `claudicus run` (alias: `r`)

Executes a command in all active agent sessions.

```bash
claudicus run "git status"              # Run in all agents
claudicus run --delete "npm test"       # Run and delete the window after
```

#### `claudicus broadcast` (alias: `b`)

Sends a message to all active agent sessions via tmux.

```bash
claudicus broadcast "Please add error handling to all API calls"
```

The broadcast command:

- Automatically discovers all active agent sessions
- Sends the message directly to each agent's tmux terminal
- Provides feedback on which sessions received the message
- Works with any AI tool running in the agent sessions

#### `claudicus checkpoint` (alias: `c`)

Makes a commit and rebases changes from an agent's worktree into your current branch.

```bash
claudicus checkpoint agent-name "feat: implement user authentication"
```

#### `claudicus reset`

Removes all Claudicus data and configuration.

```bash
claudicus reset
```

**⚠️ Warning**: This deletes all data in `~/.local/share/claudicus`

## TUI Interface

### What it does

The TUI (Terminal User Interface) provides a beautiful, real-time visual interface for:

- **Session Overview**: Visual display of all active agent sessions
- **Status Monitoring**: Real-time status updates with color-coded indicators
- **Progress Tracking**: Monitor agent activity and diff statistics
- **Session Management**: Kill, restart, and manage agent lifecycle
- **Diff Preview**: Syntax-highlighted code changes with git integration
- **Interactive Broadcasting**: Built-in message input for sending commands to all agents
- **Split View Mode**: Toggle between list-only and split view with diff preview
- **Real-time Updates**: Automatic refresh with configurable intervals

### How to launch

```bash
claudicus tui
```

The TUI automatically detects terminal capabilities and provides rich visual feedback with Claude Squad's color scheme.

### Navigation keys

#### Core Navigation

- **↑/↓ arrows** or **j/k**: Navigate between sessions
- **←/→ arrows** or **h/l**: Navigate left/right (vim-style navigation)
- **Tab**: Toggle between list view and split view modes
- **Enter**: Select/interact with highlighted session

#### Actions

- **r**: Refresh session data
- **k**: Kill selected session
- **b**: Broadcast message to all agents
- **q**: Quit TUI
- **?**: Show help screen
- **Esc**: Cancel current action or go back

#### List Management

- **/**: Filter sessions
- **c**: Clear filters

The interface maintains responsiveness during all operations and properly restores terminal state on exit.

## Advanced Usage

**Running different AI tools:**

```bash
claudicus prompt --agents=claude:2,aider:2,cursor:1,gemini:1 "Refactor the authentication system"
```

**Using random agent names:**

```bash
claudicus prompt --agents=random:5 "Fix all TypeScript errors"
```

**Running tests across all agents:**

```bash
claudicus run "npm test"
```

**Watch mode for hands-free operation:**

```bash
claudicus watch  # Replaces the deprecated "auto" command
```

## Architecture at a Glance

```
┌─ Claudicus Core ─────────────────────────────────────┐
│                                                      │
│  ┌─ Agent Manager ─┐    ┌─ Git Worktree ─┐          │
│  │ • Spawn/Kill    │    │ • Isolation    │          │
│  │ • Monitor       │────│ • Branch Mgmt  │          │
│  │ • Coordinate    │    │ • Merge Safety │          │
│  └─────────────────┘    └────────────────┘          │
│                                                      │
│  ┌─ TUI Interface ─┐    ┌─ State Manager ─┐         │
│  │ • Visual Status │    │ • Session State │          │
│  │ • Real-time     │────│ • Config Mgmt   │          │
│  │ • Interactive   │    │ • Persistence   │          │
│  └─────────────────┘    └─────────────────┘          │
│                                                      │
│  ┌─ Watch Controller ─┐  ┌─ Port Manager ──┐        │
│  │ • Session Mgmt   │  │ • Allocation    │          │
│  │ • Command Exec   │──│ • Conflict Res  │          │
│  │ • Terminal State │  │ • Dev Servers   │          │
│  └──────────────────┘  └─────────────────┘          │
└──────────────────────────────────────────────────────┘
```

### Key Components

- **Agent Manager**: Orchestrates multiple AI agents with lifecycle management
- **Git Worktree System**: Provides isolated development environments
- **TUI Interface**: Rich visual interface for monitoring and control
- **State Manager**: Persistent state tracking and configuration management
- **Watch Controller**: Terminal session management and automatic prompt handling
- **Port Manager**: Dynamic port allocation for development servers

### Data Flow

1. **Initialization**: Load config, validate prerequisites, prepare environment
2. **Agent Spawning**: Create worktrees, assign ports, launch tmux sessions
3. **Monitoring**: Track status, collect diffs, update TUI display
4. **Coordination**: Handle broadcasts, manage agent interactions
5. **Integration**: Checkpoint progress, merge changes, cleanup resources

## Philosophy: "Nail it before we scale it"

Claudicus follows a focused development philosophy:

- **Start simple**: Focus on core functionality working reliably before adding complexity
- **Learn from usage**: Add features only when proven necessary through real-world use
- **Trust existing tools**: Let established systems (like Git) handle errors naturally rather than adding defensive layers
- **User owns their environment**: Users are responsible for disk space, permissions, and system requirements
- **Clear over clever**: Prefer clear error messages over pre-emptive checks and complex abstractions
- **Iterate based on evidence**: Every new feature should solve a proven problem, not a theoretical one

This approach ensures we build what users actually need, not what we imagine they might want.

## Contributing

We welcome contributions to Claudicus! Please see our development workflow:

### Development Setup

```bash
git clone https://github.com/nehpz/claudicus
cd claudicus
go mod tidy
make build
```

### Testing

```bash
make test                    # Run full test suite
make test-coverage          # Generate coverage report
./scripts/smoke_test.sh     # Quick integration test
```

### Code Standards

- Follow existing patterns and conventions
- Maintain test coverage above 70%
- Update documentation for new features
- Use conventional commit messages

## License

BSD-3-Clause License - see [LICENSE](LICENSE) file for details.

---

*Claudicus is the evolution of multi-agent development tools - combining Uzi's operational excellence with Claude Squad's user experience mastery to create something neither could achieve alone.*
