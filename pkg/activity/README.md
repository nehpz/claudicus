# Activity Package

The `activity` package provides metrics tracking for monitoring agent behavior across monitor, state reader, and UI components in the Uzi system.

## Overview

This package defines the core data structures and utilities for tracking agent activity metrics, including Git-related statistics and current status information.

## Key Components

### Metrics Struct

The `Metrics` struct captures all essential activity data:

```go
type Metrics struct {
    // Git-related metrics
    Commits      int       `json:"commits"`       // Number of commits made
    Insertions   int       `json:"insertions"`    // Lines of code added
    Deletions    int       `json:"deletions"`     // Lines of code removed
    FilesChanged int       `json:"files_changed"` // Number of files modified
    LastCommitAt time.Time `json:"last_commit_at"` // Timestamp of most recent commit
    
    // Current status
    Status Status `json:"status"` // Current activity status
}
```

### Status Enum

The `Status` type defines three possible agent states:

- `StatusWorking` - Agent is actively working
- `StatusIdle` - Agent is idle and waiting for tasks
- `StatusStuck` - Agent appears to be stuck or blocked

## Usage

### Creating New Metrics

```go
import "github.com/nehpz/claudicus/pkg/activity"

// Create new metrics with default values
metrics := activity.NewMetrics()

// Metrics will be initialized with:
// - All counters at 0
// - LastCommitAt as zero time
// - Status as StatusIdle
```

### Working with Metrics

```go
// Update metrics
metrics.Commits = 3
metrics.Insertions = 150
metrics.Deletions = 75
metrics.FilesChanged = 5
metrics.LastCommitAt = time.Now()
metrics.Status = activity.StatusWorking

// Check activity status
if metrics.IsActive() {
    fmt.Println("Agent is currently active")
}

// Get total code changes
totalChanges := metrics.TotalChanges() // Returns insertions + deletions

// Check if any commits have been made
if metrics.HasCommits() {
    fmt.Printf("Agent has made %d commits\n", metrics.Commits)
}
```

### Status Validation

```go
status := activity.StatusWorking

// Check if status is valid
if status.IsValid() {
    fmt.Printf("Status %s is valid\n", status.String())
}
```

### JSON Serialization

The `Metrics` struct is fully JSON serializable:

```go
import "encoding/json"

// Marshal to JSON
jsonData, err := json.Marshal(metrics)
if err != nil {
    // handle error
}

// Unmarshal from JSON
var metrics activity.Metrics
err = json.Unmarshal(jsonData, &metrics)
if err != nil {
    // handle error
}
```

## Activity Detection

The `IsActive()` method considers an agent active if:

1. The status is `StatusWorking`, OR
2. The last commit was within the last 5 minutes

This allows for flexible activity detection that accounts for both explicit status and recent Git activity.

## Integration

This package is designed to be used across multiple components:

- **Monitor**: Track real-time activity metrics
- **State Reader**: Read and parse activity data from state files
- **UI**: Display activity information to users

The consistent data structure ensures seamless data flow between these components.

## Example JSON Output

```json
{
  "commits": 3,
  "insertions": 150,
  "deletions": 75,
  "files_changed": 5,
  "last_commit_at": "2025-01-15T10:30:00Z",
  "status": "working"
}
```
