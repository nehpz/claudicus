package timefreeze

import (
	"sync"
	"time"
)

// TestTime is a constant time used for deterministic testing.
var TestTime = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

// TimeFreeze allows controlling time in tests.
type TimeFreeze struct {
	mu  sync.RWMutex
	now time.Time
	t   testingT
}

type testingT interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// NewWithTime creates a new TimeFreeze instance initialized to a specific time.
func NewWithTime(t testingT, initialTime time.Time) *TimeFreeze {
	return &TimeFreeze{
		now: initialTime,
		t:   t,
	}
}

// Now returns the current frozen time.
func (tf *TimeFreeze) Now() time.Time {
	tf.mu.RLock()
	defer tf.mu.RUnlock()
	return tf.now
}

// Advance moves the frozen time forward by a duration.
func (tf *TimeFreeze) Advance(d time.Duration) {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	tf.now = tf.now.Add(d)
}

// AdvanceTo sets the frozen time to a specific point.
func (tf *TimeFreeze) AdvanceTo(t time.Time) {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	tf.now = t
}
