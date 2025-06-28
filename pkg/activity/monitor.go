// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/nehpz/claudicus/pkg/state"
)

// AgentActivityMonitor monitors activity across multiple agent worktrees
type AgentActivityMonitor struct {
	stateManager *state.StateManager
	ticker       *time.Ticker
	done         chan struct{}
	metrics      map[string]*Metrics
	mu           sync.RWMutex
	running      bool
}

// NewAgentActivityMonitor creates a new activity monitor
func NewAgentActivityMonitor() *AgentActivityMonitor {
	return &AgentActivityMonitor{
		stateManager: state.NewStateManager(),
		metrics:      make(map[string]*Metrics),
		done:         make(chan struct{}),
	}
}

// Start begins monitoring with a 500ms ticker
func (m *AgentActivityMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("monitor is already running")
	}

	m.ticker = time.NewTicker(500 * time.Millisecond)
	m.running = true

	go m.monitorLoop(ctx)
	
	log.Debug("AgentActivityMonitor started with 500ms ticker")
	return nil
}

// Stop stops the monitoring
func (m *AgentActivityMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.done)
	
	if m.ticker != nil {
		m.ticker.Stop()
	}
	
	log.Debug("AgentActivityMonitor stopped")
}

// monitorLoop runs the monitoring ticker loop
func (m *AgentActivityMonitor) monitorLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.done:
			return
		case <-m.ticker.C:
			m.updateMetrics()
		}
	}
}

// updateMetrics updates metrics for all active sessions
func (m *AgentActivityMonitor) updateMetrics() {
	// Get active sessions from state manager
	activeSessions, err := m.stateManager.GetActiveSessionsForRepo()
	if err != nil {
		log.Error("Failed to get active sessions", "error", err)
		return
	}

	// Load state data to get worktree paths
	states := make(map[string]state.AgentState)
	if data, err := os.ReadFile(m.stateManager.GetStatePath()); err != nil {
		if !os.IsNotExist(err) {
			log.Error("Failed to read state file", "error", err)
		}
		return
	} else {
		if err := json.Unmarshal(data, &states); err != nil {
			log.Error("Failed to unmarshal state data", "error", err)
			return
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update metrics for each active session
	for _, sessionName := range activeSessions {
		if agentState, exists := states[sessionName]; exists {
			metrics := m.getOrCreateMetrics(sessionName)
			m.updateSessionMetrics(sessionName, agentState.WorktreePath, metrics)
		}
	}

	// Remove metrics for inactive sessions
	for sessionName := range m.metrics {
		found := false
		for _, activeSession := range activeSessions {
			if activeSession == sessionName {
				found = true
				break
			}
		}
		if !found {
			delete(m.metrics, sessionName)
		}
	}
}

// getOrCreateMetrics gets existing metrics or creates new ones for a session
func (m *AgentActivityMonitor) getOrCreateMetrics(sessionName string) *Metrics {
	if metrics, exists := m.metrics[sessionName]; exists {
		return metrics
	}
	
	metrics := NewMetrics()
	m.metrics[sessionName] = metrics
	return metrics
}

// updateSessionMetrics updates metrics for a single session
func (m *AgentActivityMonitor) updateSessionMetrics(sessionName, worktreePath string, metrics *Metrics) {
	if worktreePath == "" {
		return
	}

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return
	}

	// Get git log info for commits and last commit time
	commits, lastCommitAt := m.getGitLogInfo(worktreePath)
	
	// Get git diff stats
	insertions, deletions, filesChanged := m.getGitDiffStats(worktreePath)

	// Update metrics
	metrics.Commits = commits
	metrics.Insertions = insertions
	metrics.Deletions = deletions
	metrics.FilesChanged = filesChanged
	if !lastCommitAt.IsZero() {
		metrics.LastCommitAt = lastCommitAt
	}
	
	// Classify status based on activity
	metrics.Status = m.Classify(metrics)
}

// getGitLogInfo gets commit count and last commit time using git log --since
func (m *AgentActivityMonitor) getGitLogInfo(worktreePath string) (int, time.Time) {
	// Use --since="24 hours ago" to get recent activity
	cmd := exec.Command("git", "--no-pager", "log", "--since=24 hours ago", "--oneline")
	cmd.Dir = worktreePath

	output, err := cmd.Output()
	if err != nil {
		return 0, time.Time{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commitCount := 0
	if len(lines) > 0 && lines[0] != "" {
		commitCount = len(lines)
	}

	// Get the most recent commit timestamp
	var lastCommitAt time.Time
	if commitCount > 0 {
		cmd := exec.Command("git", "--no-pager", "log", "-1", "--format=%ct")
		cmd.Dir = worktreePath
		
		if output, err := cmd.Output(); err == nil {
			if timestamp, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64); err == nil {
				lastCommitAt = time.Unix(timestamp, 0)
			}
		}
	}

	return commitCount, lastCommitAt
}

// getGitDiffStats gets diff statistics using git diff --shortstat
func (m *AgentActivityMonitor) getGitDiffStats(worktreePath string) (int, int, int) {
	// Stage all changes temporarily to show in diff, then reset
	cmd := exec.Command("sh", "-c", "git add -A . && git diff --cached --shortstat HEAD && git reset HEAD > /dev/null")
	cmd.Dir = worktreePath

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0
	}

	return m.parseShortstat(string(output))
}

// parseShortstat parses git diff --shortstat output
func (m *AgentActivityMonitor) parseShortstat(output string) (insertions, deletions, filesChanged int) {
	// Example output: " 3 files changed, 15 insertions(+), 7 deletions(-)"
	output = strings.TrimSpace(output)
	if output == "" {
		return 0, 0, 0
	}

	// Parse files changed
	fileRe := regexp.MustCompile(`(\d+) files? changed`)
	if match := fileRe.FindStringSubmatch(output); len(match) > 1 {
		if count, err := strconv.Atoi(match[1]); err == nil {
			filesChanged = count
		}
	}

	// Parse insertions
	insRe := regexp.MustCompile(`(\d+) insertions?\(\+\)`)
	if match := insRe.FindStringSubmatch(output); len(match) > 1 {
		if count, err := strconv.Atoi(match[1]); err == nil {
			insertions = count
		}
	}

	// Parse deletions
	delRe := regexp.MustCompile(`(\d+) deletions?\(\-\)`)
	if match := delRe.FindStringSubmatch(output); len(match) > 1 {
		if count, err := strconv.Atoi(match[1]); err == nil {
			deletions = count
		}
	}

	return insertions, deletions, filesChanged
}

// UpdateAll returns a snapshot of current metrics for all sessions
func (m *AgentActivityMonitor) UpdateAll() map[string]*Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a deep copy to avoid race conditions
	result := make(map[string]*Metrics)
	for sessionName, metrics := range m.metrics {
		// Create a copy of the metrics
		metricsCopy := &Metrics{
			Commits:      metrics.Commits,
			Insertions:   metrics.Insertions,
			Deletions:    metrics.Deletions,
			FilesChanged: metrics.FilesChanged,
			LastCommitAt: metrics.LastCommitAt,
			Status:       metrics.Status,
		}
		result[sessionName] = metricsCopy
	}

	return result
}

// Classify determines activity status based on metrics
func (m *AgentActivityMonitor) Classify(metrics *Metrics) Status {
	return m.ClassifyAtTime(metrics, time.Now())
}

// ClassifyAtTime determines activity status based on metrics at a specific time
// This is useful for testing to avoid timing-related race conditions
func (m *AgentActivityMonitor) ClassifyAtTime(metrics *Metrics, now time.Time) Status {
	if metrics == nil {
		return StatusIdle
	}

	// If there are uncommitted changes, agent is working
	if metrics.Insertions > 0 || metrics.Deletions > 0 || metrics.FilesChanged > 0 {
		return StatusWorking
	}

	// If there are recent commits (within last hour), agent is working
	if !metrics.LastCommitAt.IsZero() && now.Sub(metrics.LastCommitAt) <= time.Hour {
		return StatusWorking
	}

	// If no recent commits and no activity for more than 2 hours, agent might be stuck
	if !metrics.LastCommitAt.IsZero() && now.Sub(metrics.LastCommitAt) >= 2*time.Hour {
		return StatusStuck
	}

	// Default to idle
	return StatusIdle
}
