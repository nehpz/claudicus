# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Philosophy

Our guiding principle is "nail it before we scale it." This philosophy applies to all development, not just MVPs:

- **Start simple**: Focus on core functionality working reliably before adding complexity
- **Learn from usage**: Add features only when proven necessary through real-world use
- **Trust existing tools**: Let established systems (like Git) handle errors naturally rather than adding defensive layers
- **User owns their environment**: Users are responsible for disk space, permissions, and system requirements
- **Clear over clever**: Prefer clear error messages over pre-emptive checks and complex abstractions
- **Iterate based on evidence**: Every new feature should solve a proven problem, not a theoretical one
- **Reuse error handling**: ALWAYS use existing error handling channels. Custom error handlers require explicit user approval with detailed justification

This approach ensures we build what users actually need, not what we imagine they might want.

## SOLID Design Guidelines

When generating, reviewing, or modifying code, follow these guidelines to ensure adherence to SOLID principles:

### 1. Single Responsibility Principle (SRP)

- Each class must have only one reason to change.
- Limit class scope to a single functional area or abstraction level.
- When a class exceeds 100-150 lines, consider if it has multiple responsibilities.
- Separate cross-cutting concerns (logging, validation, error handling) from business logic.
- Create dedicated classes for distinct operations like data access, business rules, and UI.
- Method names should clearly indicate their singular purpose.
- If a method description requires "and" or "or", it likely violates SRP.
- Prioritize composition over inheritance when combining behaviors.

### 2. Open/Closed Principle (OCP)

- Design classes to be extended without modification.
- Use abstract classes and interfaces to define stable contracts.
- Implement extension points for anticipated variations.
- Favor strategy patterns over conditional logic.
- Use configuration and dependency injection to support behavior changes.
- Avoid switch/if-else chains based on type checking.
- Provide hooks for customization in frameworks and libraries.
- Design with polymorphism as the primary mechanism for extending functionality.

### 3. Liskov Substitution Principle (LSP)

- Ensure derived classes are fully substitutable for their base classes.
- Maintain all invariants of the base class in derived classes.
- Never throw exceptions from methods that don't specify them in base classes.
- Don't strengthen preconditions in subclasses.
- Don't weaken postconditions in subclasses.
- Never override methods with implementations that do nothing or throw exceptions.
- Avoid type checking or downcasting, which may indicate LSP violations.
- Prefer composition over inheritance when complete substitutability can't be achieved.

### 4. Interface Segregation Principle (ISP)

- Create focused, minimal interfaces with cohesive methods.
- Split large interfaces into smaller, more specific ones.
- Design interfaces around client needs, not implementation convenience.
- Avoid "fat" interfaces that force clients to depend on methods they don't use.
- Use role interfaces that represent behaviors rather than object types.
- Implement multiple small interfaces rather than a single general-purpose one.
- Consider interface composition to build up complex behaviors.
- Remove any methods from interfaces that are only used by a subset of implementing classes.

### 5. Dependency Inversion Principle (DIP)

- High-level modules should depend on abstractions, not details.
- Make all dependencies explicit, ideally through constructor parameters.
- Use dependency injection to provide implementations.
- Program to interfaces, not concrete classes.
- Place abstractions in a separate package/namespace from implementations.
- Avoid direct instantiation of service classes with 'new' in business logic.
- Create abstraction boundaries at architectural layer transitions.
- Define interfaces owned by the client, not the implementation.

### Implementation Guidelines

- When starting a new class, explicitly identify its single responsibility.
- Document extension points and expected subclassing behavior.
- Write interface contracts with clear expectations and invariants.
- Question any class that depends on many concrete implementations.
- Use factories, dependency injection, or service locators to manage dependencies.
- Review inheritance hierarchies to ensure LSP compliance.
- Regularly refactor toward SOLID, especially when extending functionality.
- Use design patterns (Strategy, Decorator, Factory, Observer, etc.) to facilitate SOLID adherence.

### Warning Signs

- God classes that do "everything"
- Methods with boolean parameters that radically change behavior
- Deep inheritance hierarchies
- Classes that need to know about implementation details of their dependencies
- Circular dependencies between modules
- High coupling between unrelated components
- Classes that grow rapidly in size with new features
- Methods with many parameters

## Common Development Commands

### MCP Tools

I have access to multiple MCP tool servers providing comprehensive functionality:

#### MCP Resource Tools

- **`ListMcpResourcesTool`** - List available MCP resources
- **`ReadMcpResourceTool`** - Read specific MCP resources

All MCP tools are prefixed with their server name (e.g., `mcp__taskmaster-ai__`) and automatically determine context from the workspace.

#### AI & Research Tools (`mcp__repomix__`, `mcp__code-reasoning__`, `mcp__sequential-thinking__`, `mcp__perplexity-ask__`, `mcp__context7__`)

> **ALWAYS use `repomix` when searching through the codebase**

- **`mcp__repomix__pack_codebase`** - Generate an XML file of the current codebase for efficient search
- **`mcp__repomix__pack_remote_repository`** - Generate an XML file of a remote repository for efficient search
- **`mcp__repomix__grep_repomix_output`** - Search for patterns in packed codebase using regex
- **`mcp__repomix__read_repomix_output`** - Read contents of packed codebase with optional line ranges
- **`mcp__repomix__file_system_read_directory`** - List directory contents with file/folder indicators (allows search for .gitignored files directories)
- **`mcp__repomix__file_system_read_file`** - Read individual files with security validation (allows search for for .gitignored files)
- **`mcp__code-reasoning__code-reasoning`** - Code reasoning and analysis
- **`mcp__sequential-thinking__sequentialthinking`** - Sequential thinking for complex problems
- **`mcp__perplexity-ask__perplexity_ask`** - Perplexity AI integration
- **`mcp__context7__resolve-library-id`** / **`get-library-docs`** - Library documentation

#### Language Server Tools (`mcp__language-server__`)

> **Use for diagnostics and complex text editing**

- **`get_diagnostics`** - Get real-time linting errors, warnings, and type errors from language servers
- **`get_codelens`** / **`execute_codelens`** - Code lens hints and execution (run tests, reference counts, etc.)
- **`apply_text_edit`** - Advanced text editing with regex support and bracket balancing protection

#### File System & Code Operations (`mcp__serena__`)

**System Operations**

- **`search_for_pattern`** - Search for regex patterns across the codebase
- **`restart_language_server`** - Restart language server when needed
- **`activate_project`** / **`remove_project`** - Project management
- **`switch_modes`** / **`get_current_config`** - Configuration management

**File Operations**

- **`list_dir`** - List directory contents recursively or non-recursively
- **`find_file`** - Find files by name pattern using wildcards
- **`read_file`** - Read file contents with offset/length support
- **`create_text_file`** - Write new files or overwrite existing ones

**Code Intelligence**

- **`get_symbols_overview`** - Get overview of code symbols in files/directories
- **`find_symbol`** - Find symbols by name path with optional filtering
- **`find_referencing_symbols`** - Find all references to a symbol

**Code Editing**

- **`replace_symbol_body`** - Replace the body of a specific symbol
- **`insert_after_symbol`** / **`insert_before_symbol`** - Insert code relative to symbols

- **`delete_lines`** / **`replace_lines`** / **`insert_at_line`** - Line-based editing

**Memory & Context Management**

- **`write_memory`** / **`read_memory`** / **`list_memories`** / **`delete_memory`** - Project memory management
- **`think_about_collected_information`** - Analyze gathered information sufficiency
- **`think_about_task_adherence`** - Verify task alignment before code changes
- **`think_about_whether_you_are_done`** - Assess task completion
- **`summarize_changes`** - Summarize codebase modifications

**Shell Command Execution (DISABLED)**

- **`execute_shell_command`** - DISABLED in favor of native Bash with granular permissions

#### Task Master Tools (`mcp__taskmaster-ai__`)

Task management and AI-powered project organization:

**Git Worktree Management**

- **`list_worktrees`** - List all active Git worktrees
- **`create_worktree`** - Create isolated Git worktree for task development
- **`remove_worktree`** - Remove Git worktree and cleanup

**Project & Task Management**

- **`initialize_project`** - Initialize a new Task Master project structure
- **`parse_prd`** - Parse Product Requirements Documents to generate tasks
- **`next_task`** - Find next task based on dependencies
- **`set_task_status`** - Set task/subtask status

**Single Task Operations**

- **`get_task`** - Get detailed task information
- **`add_task`** - Add new tasks with AI-generated details
- **`update_task`** - Update single task information
- **`remove_task`** - Remove tasks or subtasks permanently
- **`move_task`** - Move tasks/subtasks to new positions

**Single Subtask Operations**

- **`add_subtask`** - DISABLED (bug: wipes tag structure when used)
- **`update_subtask`** - Append timestamped info to subtasks
- **`remove_subtask`** - Remove subtask from parent

**Multi-Task & Subtask Operations**

- **`get_tasks`** - Get all tasks with optional filtering
- **`update_tasks`** - Update multiple upcoming tasks with new context
- **`clear_subtasks`** - Clear all subtasks from tasks

**Analysis & Expansion**

- **`analyze_project_complexity`** - Analyze task complexity
- **`complexity_report`** - Display complexity analysis
- **`expand_task`** - Expand task into subtasks
- **`expand_all`** - Expand all pending tasks

**Dependencies & Tags**

- **`list_tags`** / **`use_tag`**
- **`add_tag`** / **`delete_tag`** / **`rename_tag`** / **`copy_tag`**
- **`validate_dependencies`** / **`fix_dependencies`**
- **`add_dependency`** / **`remove_dependency`**

**AI & Research**

- **`models`** - Configure AI models
- **`research`** - AI-powered research with project context

## Tool Permission Strategy

The project uses a carefully configured permission system to optimize tool usage and maintain security:

### Tool Hierarchy

**Diagnostics & Analysis**

- `mcp__language-server__get_diagnostics` - Real-time linting errors and warnings
- `mcp__language-server__get_codelens` / `execute_codelens` - Interactive code features

**Search & Discovery**

- `mcp__repomix__*` tools for efficient codebase search and analysis
- `mcp__serena__search_for_pattern` for targeted regex searches
- `mcp__serena__find_symbol` / `find_referencing_symbols` for code intelligence

**File Operations**

- `mcp__serena__read_file` / `list_dir` for normal project files (auto-approved)
- `mcp__repomix__file_system_read_file` for .gitignored files (requires approval)

**Code Editing**

- `mcp__language-server__apply_text_edit` for complex edits with bracket protection
- `mcp__serena__replace_symbol_body` / `insert_*_symbol` for semantic editing
- `mcp__serena__replace_lines` / `delete_lines` for line-based edits

### Permission Levels

**Auto-Approved Tools**

- Project-aware operations (Serena file/symbol tools)
- Safe analysis tools (repomix search, diagnostics)
- Controlled shell commands (formatting, testing)

**Requires Approval**

- .gitignored file access (repomix file tools)
- Delete operations (all MCP servers)
- Git worktree operations (blocked via Bash)

**Blocked Tools**

- Redundant native tools (Read, Edit, Grep, etc.)
- Unsafe shell patterns (git worktree, find, grep via Bash)
- Overlapping functionality (serena regex, language-server symbol tools)

This configuration ensures efficient workflows while maintaining security boundaries.

## Project Overview

Uzi is a Go-based CLI tool for orchestrating multiple AI coding agents in parallel, with automatic Git worktree management and tmux session handling. It allows developers to run multiple AI agents simultaneously on isolated copies of their codebase.

### Core Architecture

**Command Structure:**
- Main entry point: `uzi.go` with subcommand routing using `ffcli`
- Commands in `cmd/` directory: `broadcast`, `checkpoint`, `kill`, `ls`, `prompt`, `reset`, `run`, `watch`
- Core packages in `pkg/`: `agents`, `config`, `state`

**Key Components:**
- `StateManager` (`pkg/state/state.go`): Manages agent session state in JSON format at `~/.local/share/uzi/state.json`
- `Config` (`pkg/config/config.go`): Handles `uzi.yaml` configuration files with `devCommand` and `portRange`
- `AgentState`: Tracks git repo, branch, worktree path, port, model, and timestamps for each agent

**Agent Lifecycle:**
1. `prompt` command creates new agent sessions with isolated git worktrees
2. Each agent gets its own tmux session and development server port
3. `ls` command shows status of all active agents with diff stats
4. `checkpoint` command merges agent changes back to main branch
5. `kill` command cleans up tmux sessions and git worktrees

### Development Commands

**Building and Testing:**
```bash
# Build the binary
make build

# Run locally
make run

# Clean build artifacts  
make clean

# Build with Go directly
go build -o uzi uzi.go
```

**Installation and Setup:**
```bash
# Install from source
go install github.com/devflowinc/uzi@latest

# Create uzi.yaml configuration
echo "devCommand: your-dev-command --port \$PORT" > uzi.yaml
echo "portRange: 3000-3010" >> uzi.yaml
```

**Testing Commands:**
- Currently no automated test suite - relies on manual testing
- Use `uzi ls -w` to monitor agent status during development
- Test agent creation with `uzi prompt --agents random:1 "test task"`

### Configuration

**uzi.yaml Structure:**
```yaml
devCommand: cd your-project && npm run dev --port $PORT
portRange: 3000-3010
```

**State Management:**
- Agent state stored in `~/.local/share/uzi/state.json`
- Worktree branches tracked in `~/.local/share/uzi/worktree/{session}/tree`
- Automatic cleanup on agent termination

### Dependencies

**Runtime Dependencies:**
- Git (for worktree management)
- Tmux (for session management)
- AI tools (Claude, Cursor, etc.)

**Go Dependencies:**
- `github.com/charmbracelet/log` - Structured logging
- `github.com/peterbourgon/ff/v3` - CLI framework with ffcli
- `gopkg.in/yaml.v3` - YAML configuration parsing

### Release Process

- Uses GoReleaser for cross-platform builds
- Triggered by git tags matching `v*` pattern
- Builds for Linux, macOS, Windows on amd64/arm64
- GitHub releases with checksums and archives
