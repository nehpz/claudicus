package timefreeze

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nehpz/claudicus/pkg/testutil"
)

func TestNewWithTime(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("creates new TimeFreeze with specified time", func(t *testing.T) {
		testTime := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC)
		tf := NewWithTime(t, testTime)

		require.NotNil(tf)
		require.Equal(testTime, tf.Now())
	})

	t.Run("creates with TestTime constant", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		require.NotNil(tf)
		require.Equal(TestTime, tf.Now())
		require.Equal(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), tf.Now())
	})

	t.Run("creates with different test instances", func(t *testing.T) {
		testTime1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		testTime2 := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

		tf1 := NewWithTime(t, testTime1)
		tf2 := NewWithTime(t, testTime2)

		require.Equal(testTime1, tf1.Now())
		require.Equal(testTime2, tf2.Now())
		require.NotEqual(tf1.Now(), tf2.Now())
	})
}

func TestNow(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("returns current frozen time", func(t *testing.T) {
		testTime := time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC)
		tf := NewWithTime(t, testTime)

		require.Equal(testTime, tf.Now())
	})

	t.Run("returns same time on multiple calls", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		time1 := tf.Now()
		time.Sleep(1 * time.Millisecond) // Real time passes
		time2 := tf.Now()

		require.Equal(time1, time2) // Frozen time doesn't change
	})

	t.Run("concurrent access returns consistent time", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		var wg sync.WaitGroup
		results := make([]time.Time, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				results[index] = tf.Now()
			}(i)
		}

		wg.Wait()

		// All results should be the same
		for i := 1; i < len(results); i++ {
			require.Equal(results[0], results[i])
		}
	})
}

func TestAdvance(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("advances time by duration", func(t *testing.T) {
		initialTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		tf := NewWithTime(t, initialTime)

		tf.Advance(1 * time.Hour)
		expected := initialTime.Add(1 * time.Hour)

		require.Equal(expected, tf.Now())
	})

	t.Run("advances time by multiple durations", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)
		original := tf.Now()

		tf.Advance(5 * time.Minute)
		after5min := tf.Now()
		require.Equal(original.Add(5*time.Minute), after5min)

		tf.Advance(10 * time.Second)
		after10sec := tf.Now()
		require.Equal(original.Add(5*time.Minute+10*time.Second), after10sec)

		tf.Advance(2 * time.Hour)
		final := tf.Now()
		require.Equal(original.Add(5*time.Minute+10*time.Second+2*time.Hour), final)
	})

	t.Run("advances by zero duration", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)
		original := tf.Now()

		tf.Advance(0)

		require.Equal(original, tf.Now())
	})

	t.Run("advances by negative duration", func(t *testing.T) {
		initialTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		tf := NewWithTime(t, initialTime)

		tf.Advance(-1 * time.Hour)
		expected := initialTime.Add(-1 * time.Hour)

		require.Equal(expected, tf.Now())
	})

	t.Run("advances by various durations", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)
		original := tf.Now()

		durations := []time.Duration{
			1 * time.Nanosecond,
			1 * time.Microsecond,
			1 * time.Millisecond,
			1 * time.Second,
			1 * time.Minute,
			1 * time.Hour,
			24 * time.Hour, // 1 day
		}

		expected := original
		for _, d := range durations {
			tf.Advance(d)
			expected = expected.Add(d)
			require.Equal(expected, tf.Now())
		}
	})
}

func TestAdvanceTo(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("sets time to specific point", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		newTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
		tf.AdvanceTo(newTime)

		require.Equal(newTime, tf.Now())
	})

	t.Run("advances to future time", func(t *testing.T) {
		initialTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		tf := NewWithTime(t, initialTime)

		futureTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		tf.AdvanceTo(futureTime)

		require.Equal(futureTime, tf.Now())
	})

	t.Run("advances to past time", func(t *testing.T) {
		initialTime := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
		tf := NewWithTime(t, initialTime)

		pastTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		tf.AdvanceTo(pastTime)

		require.Equal(pastTime, tf.Now())
	})

	t.Run("advances to same time", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)
		original := tf.Now()

		tf.AdvanceTo(original)

		require.Equal(original, tf.Now())
	})

	t.Run("multiple advance to operations", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		times := []time.Time{
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2021, 6, 15, 12, 30, 0, 0, time.UTC),
			time.Date(2022, 12, 31, 23, 59, 59, 999999999, time.UTC),
			time.Date(2023, 7, 4, 16, 45, 30, 123456789, time.UTC),
		}

		for _, targetTime := range times {
			tf.AdvanceTo(targetTime)
			require.Equal(targetTime, tf.Now())
		}
	})
}

func TestConcurrency(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("concurrent advances", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)
		var wg sync.WaitGroup

		// Start multiple goroutines that advance time
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tf.Advance(1 * time.Second)
			}()
		}

		wg.Wait()

		// Time should have advanced by 10 seconds total
		expected := TestTime.Add(10 * time.Second)
		require.Equal(expected, tf.Now())
	})

	t.Run("concurrent advance and read", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)
		var wg sync.WaitGroup
		readings := make([]time.Time, 100)

		// Start readers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				readings[index] = tf.Now()
			}(i)
		}

		// Start writers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				tf.Advance(1 * time.Millisecond)
				readings[index+50] = tf.Now()
			}(i)
		}

		wg.Wait()

		// All readings should be valid times (no zero times from race conditions)
		for _, reading := range readings[:100] {
			require.False(reading.IsZero())
		}
	})

	t.Run("concurrent advance to operations", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)
		var wg sync.WaitGroup

		targetTimes := make([]time.Time, 10)
		for i := 0; i < 10; i++ {
			targetTimes[i] = TestTime.Add(time.Duration(i+1) * time.Hour)
		}

		// Multiple goroutines setting different times
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				tf.AdvanceTo(targetTimes[index])
			}(i)
		}

		wg.Wait()

		// Final time should be one of the target times
		finalTime := tf.Now()
		found := false
		for _, targetTime := range targetTimes {
			if finalTime.Equal(targetTime) {
				found = true
				break
			}
		}
		require.True(found, "Final time should be one of the target times")
	})
}

func TestTimeMutations(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("returned time is independent", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		time1 := tf.Now()
		time2 := tf.Now()

		// Modify one time (if it were mutable, which time.Time isn't, but testing the concept)
		// Since time.Time is immutable, we test that they're separate instances
		require.Equal(time1, time2)
		require.True(time1.Equal(time2))
	})
}

func TestTestTimeConstant(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("TestTime constant is correct", func(t *testing.T) {
		expected := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
		require.Equal(expected, TestTime)
	})

	t.Run("TestTime is in UTC", func(t *testing.T) {
		require.Equal(time.UTC, TestTime.Location())
	})

	t.Run("TestTime year is 2025", func(t *testing.T) {
		require.Equal(2025, TestTime.Year())
	})

	t.Run("TestTime is January 1st", func(t *testing.T) {
		require.Equal(time.January, TestTime.Month())
		require.Equal(1, TestTime.Day())
	})

	t.Run("TestTime is midnight", func(t *testing.T) {
		require.Equal(0, TestTime.Hour())
		require.Equal(0, TestTime.Minute())
		require.Equal(0, TestTime.Second())
		require.Equal(0, TestTime.Nanosecond())
	})
}

func TestEdgeCases(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("advance by maximum duration", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		// Advance by a very large duration
		tf.Advance(time.Duration(1<<63 - 1)) // Maximum positive duration

		// Should not panic and should return a valid time
		result := tf.Now()
		require.False(result.IsZero())
	})

	t.Run("advance to zero time", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		zeroTime := time.Time{}
		tf.AdvanceTo(zeroTime)

		require.Equal(zeroTime, tf.Now())
		require.True(tf.Now().IsZero())
	})

	t.Run("advance to unix epoch", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		unixEpoch := time.Unix(0, 0).UTC()
		tf.AdvanceTo(unixEpoch)

		require.Equal(unixEpoch, tf.Now())
	})

	t.Run("advance to very far future", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		farFuture := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
		tf.AdvanceTo(farFuture)

		require.Equal(farFuture, tf.Now())
	})

	t.Run("advance to very far past", func(t *testing.T) {
		tf := NewWithTime(t, TestTime)

		farPast := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
		tf.AdvanceTo(farPast)

		require.Equal(farPast, tf.Now())
	})
}

func TestMultipleInstances(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("multiple instances are independent", func(t *testing.T) {
		time1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		time2 := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)

		tf1 := NewWithTime(t, time1)
		tf2 := NewWithTime(t, time2)

		// Advance tf1
		tf1.Advance(1 * time.Hour)

		// tf2 should be unchanged
		require.Equal(time1.Add(1*time.Hour), tf1.Now())
		require.Equal(time2, tf2.Now())

		// Advance tf2
		tf2.AdvanceTo(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

		// tf1 should be unchanged
		require.Equal(time1.Add(1*time.Hour), tf1.Now())
		require.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), tf2.Now())
	})
}

// Mock testing.T for testing testingT interface
type mockTestingT struct {
	errors []string
	fatals []string
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.fatals = append(m.fatals, fmt.Sprintf(format, args...))
}

func TestTestingTInterface(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("accepts testing.T interface", func(t *testing.T) {
		// Using the real testing.T
		tf := NewWithTime(t, TestTime)
		require.NotNil(tf)
		require.Equal(TestTime, tf.Now())
	})

	t.Run("accepts mock testing interface", func(t *testing.T) {
		mock := &mockTestingT{}
		tf := NewWithTime(mock, TestTime)

		require.NotNil(tf)
		require.Equal(TestTime, tf.Now())
		require.Equal(0, len(mock.errors))
		require.Equal(0, len(mock.fatals))
	})
}
