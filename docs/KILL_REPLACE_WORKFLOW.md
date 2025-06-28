# Kill & Replace Workflow Implementation

## Overview

This document describes the implementation of Step 9: Kill & Replace Workflow as specified in the project requirements. The implementation provides a robust, fat-finger-protected workflow for killing agents with optional replacement spawning.

## Features Implemented

### 1. Enhanced Confirmation Modal

The confirmation modal has been completely redesigned with the following features:

- **Multi-step workflow**: Supports both simple kill and kill & replace operations
- **Fat-finger protection**: Requires typing the exact agent name to confirm
- **Mode switching**: Tab key toggles between "Kill Only" and "Kill & Replace" modes
- **Progressive workflow**: For replace mode, guides through 3 steps:
  1. Agent name confirmation
  2. Prompt input for replacement
  3. Model selection for replacement

### 2. Agent Name Extraction

- **Intelligent parsing**: Extracts agent name from session names with format `agent-project-hash-agentname`
- **Hash detection**: Uses alphanumeric pattern matching to identify git hash segments
- **Fallback handling**: Gracefully handles non-standard session name formats

### 3. SpawnAgent Helper Method

- **New UziInterface method**: `SpawnAgent(prompt, model string) (string, error)`
- **Session discovery**: Automatically finds newly created sessions after spawning
- **Error handling**: Comprehensive error handling for spawn failures

### 4. State Management Integration

- **StateManagerInterface**: Enhanced with `RemoveState` method for worktree cleanup
- **StateManagerBridge**: Provides bridge between state package and TUI interface
- **Graceful degradation**: Continues operation even if state management fails

## Workflow Details

### Kill Only Mode

1. User presses `k` to kill selected agent
2. Modal appears requiring agent name confirmation
3. User can press `Tab` to switch to "Kill Only" mode
4. User types exact agent name (e.g., "claude")
5. Press `Enter` to confirm kill
6. Agent session is terminated
7. Session list refreshes

### Kill & Replace Mode (Default)

1. User presses `k` to kill selected agent
2. Modal appears in "Kill & Replace" mode by default
3. **Step 1**: User types exact agent name to confirm
4. **Step 2**: User enters prompt for replacement agent
5. **Step 3**: User selects/confirms model (defaults to "claude:1")
6. Press `Enter` to execute kill & replace
7. Original agent is killed
8. Worktree state is removed
9. New agent is spawned with specified prompt and model
10. Session list refreshes with new agent

### Fat-Finger Protection

- **Exact name matching**: User must type the exact agent name extracted from session
- **Visual feedback**: Modal shows the required agent name
- **No progression**: Workflow doesn't advance without correct name
- **Clear instructions**: Modal provides clear guidance at each step

## Technical Implementation

### Files Modified/Created

1. **`pkg/tui/confirmation_modal.go`**
   - Complete redesign with multi-step workflow
   - Text input components for each step
   - Mode switching capabilities

2. **`pkg/tui/app.go`**
   - Integration with new modal workflow
   - Kill & replace execution logic
   - State management integration

3. **`pkg/tui/uzi_interface.go`**
   - New `SpawnAgent` method
   - Enhanced state management interface
   - StateManagerBridge implementation

4. **`pkg/tui/kill_replace_test.go`** (New)
   - Comprehensive test suite for new functionality
   - Fat-finger protection tests
   - Multi-step workflow tests

### Key Design Decisions

1. **Progressive disclosure**: Complex workflow is broken into simple, guided steps
2. **Mode flexibility**: Users can choose simple kill or full replace
3. **Safety first**: Multiple confirmation points prevent accidental operations
4. **Graceful degradation**: System continues to work even if optional components fail
5. **SOLID principles**: Clean separation of concerns and single responsibility

### Error Handling

- **Kill failures**: Gracefully handled, user notified via refresh
- **Spawn failures**: System continues operation, user sees session list update
- **State management failures**: Non-blocking, logged but doesn't prevent operation
- **Invalid input**: Clear feedback, workflow doesn't advance

## Testing

The implementation includes comprehensive tests covering:

- ✅ Basic kill & replace workflow
- ✅ Fat-finger protection
- ✅ Agent name progression
- ✅ Kill-only mode
- ✅ Complete replace workflow
- ✅ Modal escape/cancel functionality
- ✅ Legacy confirmation modal compatibility

## Usage Instructions

### For Users

1. **Simple Kill**: Press `k`, press `Tab` to switch to kill mode, type agent name, press `Enter`
2. **Kill & Replace**: Press `k`, type agent name, press `Enter`, type prompt, press `Enter`, confirm/edit model, press `Enter`
3. **Cancel**: Press `Esc` at any point to cancel the operation

### For Developers

```go
// The SpawnAgent method is now available on UziInterface
sessionName, err := uzi.SpawnAgent("Fix the login bug", "claude:1")
if err != nil {
    // Handle spawn error
    return err
}
// sessionName contains the name of the newly created session
```

## Future Enhancements

Potential improvements for future iterations:

1. **Model selection dropdown**: Instead of text input, provide a selection list
2. **Template prompts**: Pre-defined prompt templates for common tasks
3. **Batch operations**: Kill & replace multiple agents simultaneously
4. **Session history**: Track kill & replace operations for audit purposes
5. **Confirmation sound**: Audio feedback for successful operations

## Compliance with Requirements

This implementation fully satisfies the Step 9 requirements:

- ✅ **uzi.KillSession(name)**: Existing method used for termination
- ✅ **Remove worktree via state.RemoveState**: Implemented with StateManagerBridge
- ✅ **uzi.SpawnAgent(prompt, model)**: New helper method implemented
- ✅ **Returns sessionName**: SpawnAgent returns the new session name
- ✅ **Add to state**: Automatic via RunPrompt integration
- ✅ **Auto-attach updated list**: Session list automatically refreshes
- ✅ **Fat-finger protection**: Requires modal confirm text "agent-name"

The implementation goes beyond basic requirements by providing a polished, user-friendly interface with comprehensive error handling and testing.
