# Uzi - Multi-Agent Development Tool

AI Code agents don't always "just work". Instead of fighting with a single agent to get it right.

Uzi is a powerful command-line tool designed to manage multiple AI coding agents simultaneously. It creates isolated development environments using Git worktrees and tmux sessions, allowing you to run multiple AI agents in parallel on the same task

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Commands](#commands)
- [Usage Examples](#usage-examples)
- [How It Works](#how-it-works)

## Features

- 🤖 Run multiple AI coding agents in parallel
- 🌳 Automatic Git worktree management for isolated development
- 🖥️ Tmux session management for each agent
- 🚀 Automatic development server setup with port management
- 📊 Real-time monitoring of agent status and code changes
- 🔄 Automatic handling of agent prompts and confirmations
- 🎯 Easy checkpoint and merge of agent changes

## Prerequisites

- **Git**: For version control and worktree management
- **Tmux**: For terminal session management
- **Go**: For building from source (if not using pre-built binary)
- **Your AI tool of choice**: Such as `claude`, `codex`, etc.

## Installation

### Installing from Go Modules

```bash
go install github.com/devflowinc/uzi@latest
uzi
```

Ensure GOBIN is in your PATH.

```sh
export PATH="$PATH:$HOME/go/bin"
```

## Configuration

### uzi.yaml

Create a `uzi.yaml` file in your project root to configure Uzi:

```yaml
devCommand: cd astrobits && yarn && yarn dev --port $PORT
portRange: 3000-3010
```

#### Configuration Options

- **`devCommand`**: The command to start your development server. Use `$PORT` as a placeholder for the port number.
  - Example for Next.js: `npm install && npm run dev -- --port $PORT`
  - Example for Vite: `npm install && npm run dev -- --port $PORT`
  - Example for Django: `pip install -r requirements.txt && python manage.py runserver 0.0.0.0:$PORT`
  
- **`portRange`**: The range of ports Uzi can use for development servers (format: `start-end`)

**Important**: The `devCommand` should include all necessary setup steps (like `npm install`, `pip install`, etc.) as each agent runs in an isolated worktree with its own dependencies.

## Usage Examples

### Basic Workflow

1. **Start agents with a task:**
   ```bash
   uzi prompt --agents claude:3,codex:2 "Implement a REST API for user management with authentication"
   ```

2. **Run uzi auto**

    uzi auto automatically presses Enter to confirm all tool calls

    ```
    uzi auto
    ```

3. **Monitor agent progress:**
   ```bash
   uzi ls -w  # Watch mode
   ```

4. **Send additional instructions:**
   ```bash
   uzi broadcast "Make sure to add input validation"
   ```

5. **Merge completed work:**
   ```bash
   uzi checkpoint funny-elephant "feat: add user management API"
   ```

## Commands

### `uzi prompt` (alias: `uzi p`)

Creates new agent sessions with the specified prompt.

```bash
uzi prompt --agents claude:2,codex:1 "Build a todo app with React"
```

**Options:**
- `--agents`: Specify agents and counts in format `agent:count[,agent:count...]`
  - Use `random` as agent name for random agent names
  - Example: `--agents claude:2,random:3`

### `uzi ls` (alias: `uzi l`)

Lists all active agent sessions with their status.

```bash
uzi ls       # List active sessions
uzi ls -w    # Watch mode - refreshes every second
```

```
AGENT    MODEL  STATUS    DIFF  ADDR                     PROMPT
brian    codex  ready  +0/-0  http://localhost:3003  make a component that looks similar to @astrobits/src/components/Button/ that creates a Tooltip in the same style. Ensure that you include a reference to it and examples on the main page.
gregory  codex  ready  +0/-0  http://localhost:3001  make a component that `
```

### `uzi auto` (alias: `uzi a`)

Monitors all agent sessions and automatically handles prompts.

```bash
uzi auto
```

**Features:**
- Auto-presses Enter for trust prompts
- Handles continuation confirmations
- Runs in the background until interrupted (Ctrl+C)

### `uzi kill` (alias: `uzi k`)

Terminates agent sessions and cleans up resources.

```bash
uzi kill agent-name    # Kill specific agent
uzi kill all          # Kill all agents
```

### `uzi run` (alias: `uzi r`)

Executes a command in all active agent sessions.

```bash
uzi run "git status"              # Run in all agents
uzi run --delete "npm test"       # Run and delete the window after
```

**Options:**
- `--delete`: Remove the tmux window after running the command

### `uzi broadcast` (alias: `uzi b`)

Sends a message to all active agent sessions.

```bash
uzi broadcast "Please add error handling to all API calls"
```

### `uzi checkpoint` (alias: `uzi c`)

Makes a commit and rebases changes from an agent's worktree into your current branch.

```bash
uzi checkpoint agent-name "feat: implement user authentication"
```

### `uzi reset`

Removes all Uzi data and configuration.

```bash
uzi reset
```

**Warning**: This deletes all data in `~/.local/share/uzi`


### Advanced Usage

**Running different AI tools:**
```bash
uzi prompt --agents=claude:2,aider:2,cursor:1 "Refactor the authentication system"
```

**Using random agent names:**
```bash
uzi prompt --agents=random:5 "Fix all TypeScript errors"
```

**Running tests across all agents:**
```bash
uzi run "npm test"
```

## How It Works

### Architecture

1. **Git Worktrees**: Each agent gets its own Git worktree, allowing parallel development without conflicts
2. **Tmux Sessions**: Each agent runs in a tmux session with:
   - `agent` window: Where the AI tool runs
   - `uzi-dev` window: Where the development server runs (if configured)
3. **State Management**: Uzi tracks all sessions in `~/.local/share/uzi/state.json`
4. **Port Management**: Automatically assigns available ports from the configured range

### Session Naming

Sessions follow the pattern: `agent-{project}-{git-hash}-{agent-name}`

Example: `agent-myapp-abc123-funny-elephant`

### File Structure

```
~/.local/share/uzi/
├── state.json              # Global state tracking
├── worktrees/             # Git worktrees for each agent
│   └── {agent-name}/
└── worktree/              # Per-worktree state
    └── {session-name}/
```

## Tips and Best Practices

1. **Always include setup commands**: In your `devCommand`, include installation steps (`npm install`, `pip install`, etc.)

2. **Use meaningful prompts**: Be specific about what you want each agent to accomplish

3. **Monitor with watch mode**: Use `uzi ls -w` to keep an eye on all agents

4. **Use the auto command**: Run `uzi auto` in a separate terminal to handle prompts automatically

5. **Checkpoint frequently**: Don't let agent branches diverge too far from main

6. **Clean up regularly**: Use `uzi kill all` to clean up finished sessions

## Troubleshooting

**Agent not responding to prompts:**
- Check if the agent window is waiting for input: `tmux attach -t session-name`
- Use `uzi auto` to automatically handle common prompts

**Port conflicts:**
- Ensure your `portRange` in `uzi.yaml` doesn't conflict with other services
- Uzi automatically finds available ports within the range

**Worktree conflicts:**
- If a worktree gets corrupted, use `uzi kill agent-name` to clean it up
- For persistent issues, `uzi reset` will clear all data

**Development server not starting:**
- Check your `devCommand` includes all necessary setup steps
- Verify the command works when run manually
- Check the `uzi-dev` window for error messages: `tmux attach -t session-name` 
