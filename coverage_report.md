# COVERAGE ANALYSIS REPORT - CLAUDICUS
Generated: 2025-06-28

## CRITICALITY CRITERIA

### Business Importance (1-10 scale)
- **10**: Main entry points (uzi.go) - Critical for application startup
- **9**: Core TUI application logic (pkg/tui/app.go) - User interface backbone  
- **8**: State management, core commands - Essential functionality
- **7**: Session management, listing - Core user operations
- **6**: Secondary features (broadcast, monitoring) - Important but not critical
- **3**: Test utilities, mocks - Development support

### Dependency Count
- **High (7+)**: Many external dependencies, complex integration points
- **Medium (4-6)**: Moderate dependencies, standard integration
- **Low (1-3)**: Minimal dependencies, self-contained

### Cyclomatic Complexity
- **High (15+)**: Complex control flow, multiple decision paths
- **Medium (8-14)**: Moderate complexity, manageable branching
- **Low (1-7)**: Simple, linear execution flow

---

## A. CRITICAL FUNCTIONS WITH <100% COVERAGE

Based on gocyclo analysis and business importance, the following critical functions lack complete test coverage:

### HIGHEST PRIORITY (Complexity 15+, Business Critical)

| Function | File | Coverage | Complexity | Business Impact | Risk Level |
|----------|------|----------|------------|----------------|------------|
| `(*App).Update` | pkg/tui/app.go | 24.1% | 47 | CRITICAL | 🔴 VERY HIGH |
| `executePrompt` | cmd/prompt/prompt.go | 23.5% | 40 | CRITICAL | 🔴 VERY HIGH |
| `executeCheckpoint` | cmd/checkpoint/checkpoint.go | 0.0% | 21 | HIGH | 🔴 VERY HIGH |
| `executeLs` | cmd/ls/ls.go | 0.0% | 16 | HIGH | 🔴 VERY HIGH |
| `(*UziCLI).setupDevEnvironment` | pkg/tui/uzi_interface.go | 0.0% | 15 | HIGH | 🔴 VERY HIGH |

### HIGH PRIORITY (Complexity 8-14, Important Functions)

| Function | File | Coverage | Complexity | Business Impact | Risk Level |
|----------|------|----------|------------|----------------|------------|
| `(*ConfirmationModal).Update` | pkg/tui/confirmation_modal.go | 88.9% | 14 | MEDIUM | 🟡 MEDIUM |
| `killSession` | cmd/kill/kill.go | 0.0% | 13 | HIGH | 🔴 HIGH |
| `(*UziCLI).createSingleAgent` | pkg/tui/uzi_interface.go | 39.4% | 12 | HIGH | 🔴 HIGH |
| `(*App).View` | pkg/tui/app.go | 0.0% | 12 | CRITICAL | 🔴 VERY HIGH |
| `executeRun` | cmd/run/run.go | 0.0% | 12 | HIGH | 🔴 HIGH |
| `(*AgentActivityMonitor).updateMetrics` | pkg/activity/monitor.go | 0.0% | 11 | MEDIUM | 🟠 MEDIUM |

### MEDIUM PRIORITY (Complexity 8-10, Core Functions)

| Function | File | Coverage | Complexity | Business Impact | Risk Level |
|----------|------|----------|------------|----------------|------------|
| `main` | uzi.go | 71.4% | 9 | CRITICAL | 🟠 MEDIUM |
| `executeKill` | cmd/kill/kill.go | 0.0% | 9 | HIGH | 🔴 HIGH |
| `executeBroadcast` | cmd/broadcast/broadcast.go | 47.6% | 8 | MEDIUM | 🟡 MEDIUM |

---

## B. FILES WITH <70% COVERAGE

Analysis by file shows the following files need significant coverage improvement:

### ZERO COVERAGE FILES (0% - Immediate Attention Required)

| File | Business Impact | Functions Affected | Priority |
|------|----------------|-------------------|----------|
| **cmd/checkpoint/checkpoint.go** | HIGH | executeCheckpoint | 🔴 CRITICAL |
| **cmd/ls/ls.go** | HIGH | All listing functions | 🔴 CRITICAL |
| **cmd/reset/reset.go** | MEDIUM | executeReset | 🔴 HIGH |
| **cmd/run/run.go** | HIGH | executeRun | 🔴 CRITICAL |
| **cmd/tui/main.go** | HIGH | TUI entry points | 🔴 CRITICAL |
| **cmd/watch/auto.go** | MEDIUM | Auto-watch functionality | 🟠 MEDIUM |
| **pkg/testutil/cmdmock/cmdmock.go** | LOW | Test utilities | 🟡 LOW |
| **pkg/testutil/timefreeze/timefreeze.go** | LOW | Test utilities | 🟡 LOW |

### LOW COVERAGE FILES (<70%)

| File | Coverage | Business Impact | Priority |
|------|----------|----------------|----------|
| **cmd/prompt/prompt.go** | ~30% | CRITICAL | 🔴 VERY HIGH |
| **pkg/tui/app.go** | ~35% | CRITICAL | 🔴 VERY HIGH |
| **pkg/state/reader.go** | ~45% | HIGH | 🔴 HIGH |
| **pkg/tui/uzi_interface.go** | ~55% | HIGH | 🔴 HIGH |
| **pkg/activity/monitor.go** | ~65% | MEDIUM | 🟠 MEDIUM |

---

## DEFINITIVE LISTS FOR REFERENCE

### CRITICAL FUNCTIONS REQUIRING IMMEDIATE TEST COVERAGE

**Tier 1 - Mission Critical (Must have 100% coverage)**
1. `(*App).Update` - Core TUI event handling (24.1% coverage)
2. `executePrompt` - Main command execution (23.5% coverage)  
3. `main` (uzi.go) - Application entry point (71.4% coverage)
4. `(*App).View` - TUI rendering (0.0% coverage)

**Tier 2 - Business Critical (Should have 90%+ coverage)**
5. `executeCheckpoint` - Session checkpointing (0.0% coverage)
6. `executeLs` - Session listing (0.0% coverage)
7. `killSession` - Session termination (0.0% coverage)
8. `executeRun` - Session execution (0.0% coverage)
9. `(*UziCLI).createSingleAgent` - Agent creation (39.4% coverage)
10. `executeKill` - Kill operations (0.0% coverage)

### FILES REQUIRING 70%+ COVERAGE

**Immediate Priority (Critical Business Functions)**
1. **cmd/prompt/prompt.go** - Core command handling
2. **pkg/tui/app.go** - Main TUI application
3. **cmd/ls/ls.go** - Session listing functionality
4. **cmd/kill/kill.go** - Session termination
5. **cmd/run/run.go** - Session execution
6. **cmd/checkpoint/checkpoint.go** - Session checkpointing

**High Priority (Important Supporting Functions)**  
7. **pkg/state/reader.go** - State management
8. **pkg/tui/uzi_interface.go** - Core interface layer
9. **cmd/reset/reset.go** - Reset functionality
10. **cmd/tui/main.go** - TUI entry point

---

## SUMMARY

- **Total critical functions with <100% coverage**: 15+ functions
- **Total files with <70% coverage**: 10+ files  
- **Overall coverage**: 51.1% (Target: 80%+)
- **Immediate risk**: Core TUI and command execution functions have insufficient coverage

### RECOMMENDATIONS

1. **Immediate Action**: Focus on Tier 1 critical functions (App.Update, executePrompt, main, App.View)
2. **Week 1**: Achieve 90%+ coverage for core command files (prompt.go, ls.go, kill.go, run.go)
3. **Week 2**: Improve TUI coverage (app.go, uzi_interface.go) to 80%+
4. **Week 3**: Address remaining medium-priority files and functions

**Risk Assessment**: Current coverage gaps present significant risk to application stability and maintainability, particularly in core user-facing functionality.
