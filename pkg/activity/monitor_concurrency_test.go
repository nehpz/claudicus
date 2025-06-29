// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package activity

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestUpdateAllConcurrencySafety tests that UpdateAll is safe to call concurrently
func TestUpdateAllConcurrencySafety(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	// Pre-populate some test metrics
	monitor.metrics = map[string]*Metrics{
		"session1": {
			Commits:      5,
			Insertions:   10,
			Deletions:    3,
			FilesChanged: 2,
			LastCommitAt: time.Now().Add(-30 * time.Minute),
			Status:       StatusWorking,
		},
		"session2": {
			Commits:      2,
			Insertions:   5,
			Deletions:    1,
			FilesChanged: 1,
			LastCommitAt: time.Now().Add(-2 * time.Hour),
			Status:       StatusStuck,
		},
		"session3": {
			Commits:      0,
			Insertions:   0,
			Deletions:    0,
			FilesChanged: 0,
			LastCommitAt: time.Time{},
			Status:       StatusIdle,
		},
	}

	const numGoroutines = 50
	const numIterations = 100

	// Channel to collect all results
	results := make(chan map[string]*Metrics, numGoroutines*numIterations)

	// Launch multiple goroutines calling UpdateAll concurrently
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				result := monitor.UpdateAll()
				results <- result

				// Add a small random delay to increase chance of race conditions
				if j%10 == 0 {
					runtime.Gosched() // Yield to other goroutines
				}
			}
		}(i)
	}

	// Close results channel when all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and validate all results
	var allResults []map[string]*Metrics
	for result := range results {
		allResults = append(allResults, result)

		// Each result should contain the expected sessions
		if len(result) != 3 {
			t.Errorf("Expected 3 sessions in result, got %d", len(result))
		}

		// Validate each session exists and has correct structure
		for sessionName, metrics := range result {
			if metrics == nil {
				t.Errorf("Session %s has nil metrics", sessionName)
				continue
			}

			// Verify that returned metrics are copies, not references to original
			originalMetrics := monitor.metrics[sessionName]
			if originalMetrics != nil && metrics == originalMetrics {
				t.Errorf("Session %s returned reference to original metrics instead of copy", sessionName)
			}
		}
	}

	if len(allResults) != numGoroutines*numIterations {
		t.Errorf("Expected %d results, got %d", numGoroutines*numIterations, len(allResults))
	}
}

// TestUpdateAllDataIntegrity tests that concurrent UpdateAll calls don't corrupt data
func TestUpdateAllDataIntegrity(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	// Set up test data with specific values we can verify
	expectedMetrics := map[string]*Metrics{
		"test-session-1": {
			Commits:      100,
			Insertions:   500,
			Deletions:    50,
			FilesChanged: 10,
			LastCommitAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Status:       StatusWorking,
		},
		"test-session-2": {
			Commits:      25,
			Insertions:   125,
			Deletions:    15,
			FilesChanged: 5,
			LastCommitAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			Status:       StatusIdle,
		},
	}

	monitor.metrics = expectedMetrics

	const numGoroutines = 20
	var wg sync.WaitGroup

	// Track any data corruption found
	errors := make(chan error, numGoroutines*10)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < 50; j++ {
				result := monitor.UpdateAll()

				// Verify data integrity
				for sessionName, expectedMetrics := range expectedMetrics {
					resultMetrics, exists := result[sessionName]
					if !exists {
						errors <- &testError{msg: "Missing session " + sessionName}
						continue
					}

					if !metricsEqual(resultMetrics, expectedMetrics) {
						errors <- &testError{msg: "Data corruption in " + sessionName}
					}
				}

				// Verify no extra sessions appeared
				if len(result) != len(expectedMetrics) {
					errors <- &testError{msg: "Wrong number of sessions returned"}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}

// TestUpdateAllWithConcurrentModification tests UpdateAll safety when metrics are being modified
func TestUpdateAllWithConcurrentModification(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	// Initial metrics
	monitor.metrics = map[string]*Metrics{
		"session1": {
			Commits:      1,
			Insertions:   10,
			Deletions:    5,
			FilesChanged: 2,
			LastCommitAt: time.Now(),
			Status:       StatusWorking,
		},
	}

	const duration = 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup

	// Goroutine 1: Continuously call UpdateAll
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				result := monitor.UpdateAll()
				if result == nil {
					t.Error("UpdateAll returned nil result")
				}
				runtime.Gosched()
			}
		}
	}()

	// Goroutine 2: Modify metrics concurrently (simulating monitor updates)
	wg.Add(1)
	go func() {
		defer wg.Done()
		counter := 2
		for {
			select {
			case <-ctx.Done():
				return
			default:
				monitor.mu.Lock()
				// Add new session
				sessionName := "session" + string(rune('0'+counter))
				monitor.metrics[sessionName] = &Metrics{
					Commits:      counter,
					Insertions:   counter * 10,
					Deletions:    counter * 2,
					FilesChanged: 1,
					LastCommitAt: time.Now(),
					Status:       StatusWorking,
				}
				counter++
				monitor.mu.Unlock()
				runtime.Gosched()
			}
		}
	}()

	// Goroutine 3: Delete sessions concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				monitor.mu.Lock()
				// Remove a random session (if any exist beyond session1)
				for sessionName := range monitor.metrics {
					if sessionName != "session1" && sessionName != "session2" {
						delete(monitor.metrics, sessionName)
						break
					}
				}
				monitor.mu.Unlock()
				runtime.Gosched()
			}
		}
	}()

	wg.Wait()

	// Final verification - UpdateAll should still work after all modifications
	finalResult := monitor.UpdateAll()
	if finalResult == nil {
		t.Error("Final UpdateAll call returned nil")
	}
}

// TestUpdateAllMemoryLeaks tests that UpdateAll doesn't create memory leaks
func TestUpdateAllMemoryLeaks(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	// Pre-populate with many sessions
	for i := 0; i < 100; i++ {
		sessionName := "session-" + string(rune('0'+i%10)) + "-" + string(rune('0'+i/10))
		monitor.metrics[sessionName] = &Metrics{
			Commits:      i,
			Insertions:   i * 10,
			Deletions:    i * 2,
			FilesChanged: i % 5,
			LastCommitAt: time.Now().Add(-time.Duration(i) * time.Minute),
			Status:       StatusWorking,
		}
	}

	// Force initial GC
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Call UpdateAll many times
	const iterations = 1000
	for i := 0; i < iterations; i++ {
		result := monitor.UpdateAll()

		// Verify result is reasonable
		if len(result) != 100 {
			t.Errorf("Iteration %d: Expected 100 sessions, got %d", i, len(result))
		}

		// Don't keep references to results to allow GC
		result = nil

		if i%100 == 0 {
			runtime.GC() // Periodic GC
		}
	}

	// Force final GC and check memory
	runtime.GC()
	runtime.GC() // Call twice to ensure thorough cleanup
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// Memory growth should be reasonable (allow 10MB growth for test overhead)
	memoryGrowth := int64(m2.Alloc - m1.Alloc)
	maxAllowedGrowth := int64(10 * 1024 * 1024) // 10MB

	if memoryGrowth > maxAllowedGrowth {
		t.Errorf("Excessive memory growth: %d bytes (max allowed: %d)", memoryGrowth, maxAllowedGrowth)
	}
}

// TestUpdateAllPerformanceUnderConcurrency tests UpdateAll performance with concurrent access
func TestUpdateAllPerformanceUnderConcurrency(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	// Set up moderate number of sessions
	for i := 0; i < 50; i++ {
		sessionName := "perf-session-" + fmt.Sprintf("%d", i)
		monitor.metrics[sessionName] = &Metrics{
			Commits:      i,
			Insertions:   i * 5,
			Deletions:    i * 2,
			FilesChanged: i % 3,
			LastCommitAt: time.Now().Add(-time.Duration(i) * time.Second),
			Status:       StatusWorking,
		}
	}

	const numGoroutines = 10
	const iterationsPerGoroutine = 100

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				result := monitor.UpdateAll()
				if len(result) != 50 {
					t.Errorf("Expected 50 sessions, got %d", len(result))
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Performance check - should complete within reasonable time
	totalCalls := numGoroutines * iterationsPerGoroutine
	avgCallTime := elapsed / time.Duration(totalCalls)
	maxAllowedAvgTime := 1 * time.Millisecond

	if avgCallTime > maxAllowedAvgTime {
		t.Errorf("UpdateAll too slow: avg %v per call (max allowed: %v)", avgCallTime, maxAllowedAvgTime)
	}

	t.Logf("UpdateAll performance: %d calls in %v (avg: %v per call)", totalCalls, elapsed, avgCallTime)
}

// TestUpdateAllIsolation tests that modifications to returned data don't affect internal state
func TestUpdateAllIsolation(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	originalMetrics := &Metrics{
		Commits:      10,
		Insertions:   50,
		Deletions:    25,
		FilesChanged: 5,
		LastCommitAt: time.Now(),
		Status:       StatusWorking,
	}

	monitor.metrics["test-session"] = originalMetrics

	// Get result from UpdateAll
	result := monitor.UpdateAll()
	returnedMetrics := result["test-session"]

	// Modify the returned metrics
	returnedMetrics.Commits = 999
	returnedMetrics.Insertions = 999
	returnedMetrics.Deletions = 999
	returnedMetrics.FilesChanged = 999
	returnedMetrics.Status = StatusStuck

	// Verify original metrics are unchanged
	if monitor.metrics["test-session"].Commits != 10 {
		t.Error("Original metrics were modified through returned copy")
	}
	if monitor.metrics["test-session"].Insertions != 50 {
		t.Error("Original metrics were modified through returned copy")
	}
	if monitor.metrics["test-session"].Status != StatusWorking {
		t.Error("Original metrics were modified through returned copy")
	}

	// Get fresh result - should still have original values
	freshResult := monitor.UpdateAll()
	freshMetrics := freshResult["test-session"]

	if freshMetrics.Commits != 10 || freshMetrics.Insertions != 50 || freshMetrics.Status != StatusWorking {
		t.Error("Fresh UpdateAll result shows modified values, indicating lack of isolation")
	}
}

// testError is a simple error type for the test
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// metricsEqual compares two Metrics instances for equality
func metricsEqual(a, b *Metrics) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return a.Commits == b.Commits &&
		a.Insertions == b.Insertions &&
		a.Deletions == b.Deletions &&
		a.FilesChanged == b.FilesChanged &&
		a.LastCommitAt.Equal(b.LastCommitAt) &&
		a.Status == b.Status
}

// TestUpdateAllEmptyState tests UpdateAll behavior with empty metrics
func TestUpdateAllEmptyState(t *testing.T) {
	monitor := NewAgentActivityMonitor()

	// Test with empty metrics map
	result := monitor.UpdateAll()
	if result == nil {
		t.Error("UpdateAll returned nil for empty metrics")
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty metrics, got %d items", len(result))
	}

	// Test concurrent access to empty state
	const numGoroutines = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				result := monitor.UpdateAll()
				if result == nil {
					t.Error("UpdateAll returned nil during concurrent access to empty state")
				}
				if len(result) != 0 {
					t.Errorf("Expected empty result, got %d items", len(result))
				}
			}
		}()
	}

	wg.Wait()
}
