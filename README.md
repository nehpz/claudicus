# Claudicus

> **Operational Excellence + UX Excellence** - A unified TUI interface that harnesses Uzi's speed under the hood while providing Claude Squad's beautiful coordination capabilities

*The definitive platform for safe, multi-agent software development that combines fast operations with rich visual experience in one seamless interface.*

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

**Note**: The installed binary is named `uzi`, which powers the TUI interface. Use the TUI for a seamless experience leveraging Uzi's speed.

### TUI Interface

The TUI is the primary interface for managing multi-agent workflows, designed to harness Uzi's speed under the hood while providing a rich visual experience. All commands and operations can be managed within the TUI, ensuring a unified and efficient user experience.

Before using Claudicus, ensure you have:

- **Git**: For version control and worktree management
- **Tmux**: For terminal session management  
- **Go**: For installation (version 1.24.3+)
- **AI tool of choice**: Such as `claude`, `cursor`, `aider`, `codex`, etc.

## Quick Start / TUI-First Workflow

### 1. Initialize your project

```bash
# Create uzi.yaml configuration in your project root
echo "devCommand: npm install && npm run dev -- --port \$PORT" > uzi.yaml
echo "portRange: 3000-3010" >> uzi.yaml
```

### 2. Launch the TUI interface

```bash
uzi tui  # Primary interface - combines Uzi's speed with visual excellence
```

### 3. Create and manage agents visually

- **Create agents**: Use the TUI prompts to start multiple AI agents
- **Monitor progress**: Real-time visual status updates and diff tracking
- **Send broadcasts**: Press 'b' to send messages to all active agents
- **Manage lifecycle**: Kill, restart, and coordinate agents with keyboard shortcuts
- **Merge work**: Use built-in checkpoint features to integrate agent changes

All operations leverage Uzi's fast backend operations through an intuitive visual interface.

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

### uzi.yaml

Create a `uzi.yaml` file in your project root to configure Claudicus:

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

## Primary Interface: TUI

Claudicus is designed around a unified TUI (Terminal User Interface) that leverages Uzi's speed and reliability under the hood. All operations are performed through intuitive keyboard shortcuts within the TUI.

### Launch the TUI

```bash
uzi tui
```

The TUI provides all functionality needed for multi-agent development:

- **Agent Creation**: Start multiple agents with configurable prompts
- **Real-time Monitoring**: Visual status updates and progress tracking
- **Session Management**: Kill, restart, and manage agent lifecycle
- **Broadcasting**: Send messages to all active agents
- **Checkpointing**: Merge agent work into your main branch
- **Diff Preview**: Syntax-highlighted code changes
- **Interactive Controls**: Keyboard-driven interface for all operations

### Advanced CLI Commands (Backend Support)

While the TUI is the primary interface, these CLI commands power the backend operations:

#### `uzi prompt` - Agent Creation Backend

Used internally by the TUI for agent spawning:

```bash
uzi prompt --agents claude:2,cursor:1 "Build a todo app with React"
```

#### `uzi ls` - Session Listing Backend

Provides session data to the TUI:

```bash
uzi ls --json  # JSON output for TUI consumption
```

#### `uzi reset` - System Reset

Cleans up all Claudicus data:

```bash
uzi reset
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
uzi tui
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

**TUI-driven multi-agent workflows:**

- Launch TUI: `uzi tui`
- Create agents with different AI tools using TUI prompts
- Monitor all agents visually in real-time
- Send broadcasts to all agents using the 'b' key
- Manage agent lifecycle with keyboard shortcuts
- Preview and merge changes with built-in diff viewer

**The TUI combines all operations into one unified, fast interface powered by Uzi's reliable backend.**

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
