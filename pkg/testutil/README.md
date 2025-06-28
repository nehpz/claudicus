# Test Utilities

This directory contains utilities for testing that provide clean abstractions for common testing patterns used in the Claudicus project.

## Overview

The testutil package provides several utilities:

- **`fsmock`** - Temporary filesystem creation and automatic cleanup
- **`timefreeze`** - Controllable time for deterministic testing
- **`cmdmock`** - Command execution mocking (existing)
- **Core utilities** - Assertion helpers and test data builders (existing)

## fsmock - Filesystem Mocking

The `fsmock` package provides utilities for creating temporary directories and files that are automatically cleaned up after tests. This is particularly useful for tests that need to create files or directories but want automatic cleanup.

### Key Features

- **Automatic cleanup** - Files and directories are automatically removed when tests complete
- **Nested directories** - Automatically creates parent directories as needed
- **Path resolution** - Converts relative paths to absolute paths within the temp filesystem
- **Git repository creation** - Helper for creating mock git repositories for testing
- **Project structure creation** - Helper for creating typical Go project structures

### Basic Usage

```go
func TestMyFeature(t *testing.T) {
    // Create a temporary filesystem
    fs := fsmock.NewTempFS(t)
    
    // Create files and directories
    fs.WriteFileString("config.txt", "host=localhost\nport=8080", 0644)
    fs.MkdirAll("logs/app", 0755)
    
    // Read files
    content, err := fs.ReadFileString("config.txt")
    require.NoError(t, err)
    
    // Check file existence
    assert.True(t, fs.Exists("config.txt"))
    assert.True(t, fs.IsFile("config.txt"))
    assert.True(t, fs.IsDir("logs"))
    
    // Get absolute paths for use with other tools
    configPath := fs.Path("config.txt")
    
    // Filesystem is automatically cleaned up when test ends
}
```

### Advanced Usage

```go
func TestGitOperations(t *testing.T) {
    fs := fsmock.NewTempFS(t)
    
    // Create a mock git repository
    fs.CreateGitRepo("myproject")
    
    // Add project files
    fs.CreateProjectStructure("myproject")
    
    // Test git-related operations
    repoPath := fs.Path("myproject")
    // ... use repoPath with git commands
}
```

## timefreeze - Time Control

The `timefreeze` package provides utilities for controlling time in tests. It works by replacing `time.Now()` calls with a controllable time source, enabling deterministic testing of time-dependent code.

### Key Features

- **Deterministic time** - Control exactly what time is returned
- **Time advancement** - Simulate passage of time without waiting
- **Global and local control** - Use per-test freezers or global time control
- **Thread-safe** - Safe for concurrent access
- **Injection support** - Easy integration with existing code via variable injection

### Basic Usage

```go
func TestTimeDependent(t *testing.T) {
    // Create a time freezer starting at a known time
    freeze := timefreeze.NewWithTime(t, timefreeze.TestTime)
    
    // Get the current frozen time
    start := freeze.Now()
    
    // Advance time by 30 minutes
    freeze.Advance(30 * time.Minute)
    
    // Check elapsed time
    end := freeze.Now()
    elapsed := end.Sub(start)
    assert.Equal(t, 30*time.Minute, elapsed)
}
```

### Integration with Production Code

To use timefreeze with existing code, replace direct `time.Now()` calls with a variable:

**Production code:**

```go
// In your package
var timeNow = time.Now

func processWithTimeout() {
    start := timeNow()
    // ... processing logic
    elapsed := timeNow().Sub(start)
}
```

**Test code:**

```go
func TestProcessWithTimeout(t *testing.T) {
    freeze := timefreeze.New(t)
    
    // Inject frozen time
    originalTimeNow := mypackage.TimeNow
    mypackage.TimeNow = freeze.Now
    defer func() { mypackage.TimeNow = originalTimeNow }()
    
    // Test the time-dependent logic
    freeze.Advance(5 * time.Minute)
    // ... test assertions
}
```

### Global Time Control

For tests that span multiple packages:

```go
func TestGlobalTimeControl(t *testing.T) {
    // Freeze time globally
    cleanup := timefreeze.Freeze(t, timefreeze.TestTime)
    defer cleanup()
    
    // Any code using timefreeze.Now() gets the frozen time
    start := timefreeze.Now()
    timefreeze.Advance(1 * time.Hour)
    end := timefreeze.Now()
    
    assert.Equal(t, 1*time.Hour, end.Sub(start))
}
```

## Integration with Existing Patterns

Both utilities are designed to work seamlessly with the existing testutil patterns in the codebase:

### With cmdmock

```go
func TestStatePersistence(t *testing.T) {
    // Set up filesystem
    fs := fsmock.NewTempFS(t)
    
    // Set up time control
    freeze := timefreeze.NewWithTime(t, timefreeze.TestTime)
    
    // Set up command mocking
    cmdmock.Reset()
    cmdmock.Enable()
    
    // Test state persistence with deterministic time and filesystem
    statePath := fs.Path("state.json")
    timestamp := freeze.Now()
    
    // ... test logic
}
```

### With existing testutil helpers

```go
func TestWithAllUtilities(t *testing.T) {
    require := testutil.NewRequire(t)
    fs := fsmock.NewTempFS(t)
    freeze := timefreeze.New(t)
    
    // Create test data
    testData := testutil.MakeFakeUziLsJSON([]testutil.SessionInfo{
        {
            Name: "test-session",
            AgentName: "claude",
            Status: "active",
        },
    })
    
    // Write to temp filesystem
    err := fs.WriteFileString("sessions.json", testData, 0644)
    require.NoError(err)
    
    // Test with frozen time
    createdAt := freeze.Now()
    freeze.Advance(5 * time.Minute)
    
    // ... test logic
}
```

## Common Use Cases

### State Persistence Testing

These utilities are particularly useful for testing state persistence functionality:

```go
func TestAgentStatePersistence(t *testing.T) {
    fs := fsmock.NewTempFS(t)
    freeze := timefreeze.NewWithTime(t, timefreeze.TestTime)
    
    // Create state with known timestamp
    state := AgentState{
        Name: "test-agent",
        CreatedAt: freeze.Now(),
        UpdatedAt: freeze.Now(),
    }
    
    // Save to temp filesystem
    statePath := fs.Path("state.json")
    saveState(state, statePath)
    
    // Advance time and update
    freeze.Advance(10 * time.Minute)
    state.UpdatedAt = freeze.Now()
    saveState(state, statePath)
    
    // Verify state persistence
    loadedState := loadState(statePath)
    assert.Equal(t, 10*time.Minute, loadedState.UpdatedAt.Sub(loadedState.CreatedAt))
}
```

### Session Monitoring

```go
func TestSessionMonitoring(t *testing.T) {
    fs := fsmock.NewTempFS(t)
    freeze := timefreeze.NewWithTime(t, timefreeze.TestTime)
    
    // Create session log directory
    fs.MkdirAll("logs/sessions", 0755)
    
    // Simulate session activity over time
    sessionStart := freeze.Now()
    
    activities := []string{"started", "running", "completed"}
    for i, activity := range activities {
        freeze.Advance(time.Duration(i+1) * time.Minute)
        
        logEntry := fmt.Sprintf("%s: %s", freeze.Now().Format(time.RFC3339), activity)
        fs.WriteFileString(fmt.Sprintf("logs/sessions/activity_%d.log", i), logEntry, 0644)
    }
    
    // Test session duration calculation
    sessionEnd := freeze.Now()
    duration := sessionEnd.Sub(sessionStart)
    assert.Equal(t, 6*time.Minute, duration) // 1+2+3 minutes
}
```

## Best Practices

1. **Always use with testing.TB** - Both utilities are designed to work with Go's testing framework and will automatically clean up

2. **Prefer injection over global state** - When possible, inject time functions rather than using global time control

3. **Use meaningful test times** - Use `timefreeze.TestTime` or other well-known times for consistency

4. **Combine utilities** - fsmock and timefreeze work well together for comprehensive testing

5. **Clean separation** - Keep filesystem and time concerns separate in your tests

## Migration Guide

If you have existing tests that manually create temp directories or deal with time:

### Before (manual temp dirs)

```go
func TestOld(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "test-*")
    require.NoError(t, err)
    defer os.RemoveAll(tmpDir)
    
    configPath := filepath.Join(tmpDir, "config.txt")
    err = os.WriteFile(configPath, []byte("content"), 0644)
    require.NoError(t, err)
}
```

### After (with fsmock)

```go
func TestNew(t *testing.T) {
    fs := fsmock.NewTempFS(t)
    
    err := fs.WriteFileString("config.txt", "content", 0644)
    require.NoError(t, err)
    
    configPath := fs.Path("config.txt")
}
```

### Before (time-dependent tests)

```go
func TestTimeOld(t *testing.T) {
    start := time.Now()
    time.Sleep(100 * time.Millisecond)  // Flaky!
    end := time.Now()
    
    assert.True(t, end.Sub(start) >= 100*time.Millisecond)
}
```

### After (with timefreeze)

```go
func TestTimeNew(t *testing.T) {
    freeze := timefreeze.New(t)
    
    start := freeze.Now()
    freeze.Advance(100 * time.Millisecond)
    end := freeze.Now()
    
    assert.Equal(t, 100*time.Millisecond, end.Sub(start))
}
```

These utilities follow the project's "nail it before we scale it" philosophy by providing simple, focused tools that solve specific testing problems without adding unnecessary complexity.
