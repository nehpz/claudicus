# Test Coverage Analysis

This document outlines the analysis of the project's test coverage.

## 1. Definition of a "Critical Function"

A function is considered **critical** if it meets one or more of the following criteria:

- **Core Command Logic:** It is a primary function within a file in the `cmd/` directory. These are the entry points for the application's user-facing features.
- **Main Application Logic:** It is a function in the root `uzi.go` file, which likely contains the core application orchestration.
- **High Complexity:** It contains significant branching logic (e.g., multiple `if/else` statements, `switch` cases) that increases the risk of bugs.

## 2. Identified Critical Functions

Based on the definition above, the following functions have been identified as critical:

- **`cmd/broadcast/broadcast.go`**:
  - `executeBroadcast`
- **`cmd/checkpoint/checkpoint.go`**:
  - `executeCheckpoint`
- **`cmd/kill/kill.go`**:
  - `executeKill`
  - `killSession`
  - `killAll`
- **`cmd/ls/ls.go`**:
  - `executeLs`
  - `getGitDiffTotals`
  - `getPaneContent`
  - `getAgentStatus`
  - `getSessionsAsJSON`
  - `printSessions`
- **`cmd/prompt/prompt.go`**:
  - `executePrompt`
  - `parseAgents`
  - `getCommandForAgent`
  - `isPortAvailable`
  - `getExistingSessionPorts`
  - `findAvailablePort`
- **`cmd/reset/reset.go`**:
  - `executeReset`
- **`cmd/run/run.go`**:
  - `executeRun`
- **`cmd/tui/main.go`**:
  - `Run`
  - `isTerminal`
  - `main`
- **`cmd/watch/auto.go`**:
  - `NewAgentWatcher`
  - `(aw *AgentWatcher) Start`
  - `(aw *AgentWatcher) watchSession`
  - `(aw *AgentWatcher) refreshActiveSessions`
  - `(aw *AgentWatcher) hasUpdated`
  - `CmdWatch.Exec`
- **`uzi.go`**:
  - `main`

## 3. Coverage Report Generation

Executed command: `go test ./... -coverprofile=coverage.out`

The test suite ran with some failures but successfully generated coverage data. Overall project coverage: **44.3%** of statements.

### Package-Level Coverage Summary

- `github.com/nehpz/claudicus/pkg/agents`: **100.0%** coverage
- `github.com/nehpz/claudicus/pkg/config`: **100.0%** coverage
- `github.com/nehpz/claudicus/pkg/testutil`: **100.0%** coverage
- `github.com/nehpz/claudicus/pkg/testutil/fsmock`: **83.6%** coverage
- `github.com/nehpz/claudicus/pkg/state`: **76.8%** coverage
- `github.com/nehpz/claudicus/pkg/tui`: **61.0%** coverage
- `github.com/nehpz/claudicus` (main): **71.4%** coverage
- `github.com/nehpz/claudicus/cmd/prompt`: **41.7%** coverage

### Commands with **0.0%** coverage

- `cmd/broadcast`
- `cmd/checkpoint`
- `cmd/kill`
- `cmd/ls`
- `cmd/reset`
- `cmd/run`
- `cmd/tui`
- `cmd/watch`

## 4. Analysis Results

### A. Critical Functions with Less Than 100% Coverage

The following critical functions have less than 100% unit test coverage:

**Command Functions (0.0% coverage):**

- `cmd/broadcast/broadcast.go`: `executeBroadcast` - **0.0%**
- `cmd/checkpoint/checkpoint.go`: `executeCheckpoint` - **0.0%**
- `cmd/kill/kill.go`: `executeKill`, `killSession`, `killAll` - **0.0%**
- `cmd/ls/ls.go`: `executeLs`, `getGitDiffTotals`, `getPaneContent`, `getAgentStatus`, `getSessionsAsJSON`, `printSessions` - **0.0%**
- `cmd/reset/reset.go`: `executeReset` - **0.0%**
- `cmd/run/run.go`: `executeRun` - **0.0%**
- `cmd/tui/main.go`: `Run`, `isTerminal`, `main` - **0.0%**
- `cmd/watch/auto.go`: `NewAgentWatcher`, `Start`, `watchSession`, `refreshActiveSessions`, `hasUpdated`, `CmdWatch.Exec` - **0.0%**

**Prompt Functions (partial coverage):**

- `cmd/prompt/prompt.go`: `executePrompt` - **23.5%**
- `cmd/prompt/prompt.go`: `getExistingSessionPorts` - **56.2%**
- `cmd/prompt/prompt.go`: `isPortAvailable` - **83.3%**

**Main Function (partial coverage):**

- `uzi.go`: `main` - **71.4%**

### B. Source Files with Less Than 70% Coverage

The following source code files have less than 70% unit test coverage:

1. **`cmd/broadcast/broadcast.go`** - **0.0%** coverage
2. **`cmd/checkpoint/checkpoint.go`** - **0.0%** coverage
3. **`cmd/kill/kill.go`** - **0.0%** coverage
4. **`cmd/ls/ls.go`** - **0.0%** coverage
5. **`cmd/reset/reset.go`** - **0.0%** coverage
6. **`cmd/run/run.go`** - **0.0%** coverage
7. **`cmd/tui/main.go`** - **0.0%** coverage
8. **`cmd/watch/auto.go`** - **0.0%** coverage
9. **`cmd/prompt/prompt.go`** - **41.7%** coverage
10. **`pkg/tui/uzi_interface.go`** - **61.0%** coverage (estimated based on package coverage)
11. **`pkg/testutil/cmdmock/cmdmock.go`** - **0.0%** coverage
12. **`pkg/testutil/timefreeze/timefreeze.go`** - **0.0%** coverage

## Summary

The analysis reveals significant gaps in test coverage, particularly in the command layer where all core CLI commands have **0.0%** coverage. The most critical issue is that all main command functions (`executeBroadcast`, `executeCheckpoint`, `executeKill`, `executeLs`, `executePrompt`, `executeReset`, `executeRun`) lack any unit test coverage.

The utility packages (`pkg/agents`, `pkg/config`, `pkg/testutil`) have excellent coverage at 100%, while the state management (`pkg/state`) and TUI components (`pkg/tui`) have moderate coverage at 76.8% and 61.0% respectively.

**Key Findings:**

- 8 out of 9 command packages have 0% coverage
- All critical command execution functions are untested
- 12 source files fall below the 70% coverage threshold
- Overall project coverage is 44.3%, well below typical industry standards of 80%+
