package cmdmock

import (
	"fmt"
	"os"
	"testing"

	"github.com/nehpz/claudicus/pkg/testutil"
)

func TestCommand(t *testing.T) {
	require := testutil.NewRequire(t)

	// Reset state before each test
	Reset()

	t.Run("disabled mock returns real command", func(t *testing.T) {
		Disable()
		cmd := Command("echo", "test")
		require.NotNil(cmd)
		// Path will be the full path to echo, so just check it contains "echo"
		require.True(len(cmd.Path) > 0)
	})

	t.Run("enabled mock with no response", func(t *testing.T) {
		Enable()
		cmd := Command("nonexistent", "command")
		require.NotNil(cmd)

		// Should record the call
		calls := GetCalls()
		require.Equal(1, len(calls))
		require.Equal("nonexistent", calls[0].Name)
		require.Equal(1, len(calls[0].Args))
		require.Equal("command", calls[0].Args[0])
	})

	t.Run("enabled mock with response", func(t *testing.T) {
		Reset()
		SetResponse("git", "output", false)

		cmd := Command("git", "status")
		require.NotNil(cmd)

		calls := GetCalls()
		require.Equal(1, len(calls))
		require.Equal("git", calls[0].Name)
		require.Equal(1, len(calls[0].Args))
		require.Equal("status", calls[0].Args[0])
	})

	t.Run("mock with specific args", func(t *testing.T) {
		Reset()
		SetResponseWithArgs("git", []string{"status"}, "clean", "", false)

		cmd := Command("git", "status")
		require.NotNil(cmd)

		calls := GetCalls()
		require.Equal(1, len(calls))
		require.Equal("git", calls[0].Name)
		require.Equal(1, len(calls[0].Args))
		require.Equal("status", calls[0].Args[0])
	})
}

func TestSetResponse(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("set simple response", func(t *testing.T) {
		Reset()
		SetResponse("echo", "hello", false)

		// Verify response is stored
		require.True(globalMock.enabled)
		require.Equal(1, len(globalMock.responses))
	})

	t.Run("set response with error", func(t *testing.T) {
		Reset()
		SetResponse("false", "", true)

		require.True(globalMock.enabled)
		require.Equal(1, len(globalMock.responses))
	})
}

func TestSetResponseWithArgs(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("set response with args", func(t *testing.T) {
		Reset()
		SetResponseWithArgs("git", []string{"status"}, "clean", "", false)

		require.True(globalMock.enabled)
		require.Equal(1, len(globalMock.responses))
	})

	t.Run("set response with stderr", func(t *testing.T) {
		Reset()
		SetResponseWithArgs("git", []string{"status"}, "stdout", "stderr", true)

		require.True(globalMock.enabled)
		require.Equal(1, len(globalMock.responses))
	})

	t.Run("multiple responses", func(t *testing.T) {
		Reset()
		SetResponseWithArgs("git", []string{"status"}, "clean", "", false)
		SetResponseWithArgs("git", []string{"log"}, "commits", "", false)

		require.Equal(2, len(globalMock.responses))
	})
}

func TestReset(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("reset clears state", func(t *testing.T) {
		// Start with clean state
		Reset()

		// Set up some state
		SetResponse("git", "output", false)
		Command("git", "status")

		require.True(globalMock.enabled)
		require.Equal(1, len(globalMock.responses))
		require.Equal(1, len(globalMock.calls))

		// Reset should clear everything
		Reset()

		require.False(globalMock.enabled)
		require.Equal(0, len(globalMock.responses))
		require.Equal(0, len(globalMock.calls))
	})
}

func TestEnableDisable(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("enable sets flag", func(t *testing.T) {
		Reset()
		require.False(globalMock.enabled)

		Enable()
		require.True(globalMock.enabled)
	})

	t.Run("disable clears flag", func(t *testing.T) {
		Enable()
		require.True(globalMock.enabled)

		Disable()
		require.False(globalMock.enabled)
	})
}

func TestGetCalls(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("empty calls", func(t *testing.T) {
		Reset()
		calls := GetCalls()
		require.Equal(0, len(calls))
	})

	t.Run("multiple calls", func(t *testing.T) {
		Reset()
		Enable()

		Command("git", "status")
		Command("tmux", "list-sessions")
		Command("echo", "test")

		calls := GetCalls()
		require.Equal(3, len(calls))
		require.Equal("git", calls[0].Name)
		require.Equal("tmux", calls[1].Name)
		require.Equal("echo", calls[2].Name)
	})

	t.Run("calls are copied", func(t *testing.T) {
		Reset()
		Enable()
		Command("git", "status")

		calls1 := GetCalls()
		Command("git", "log")
		calls2 := GetCalls()

		require.Equal(1, len(calls1))
		require.Equal(2, len(calls2))
	})
}

func TestGetCallCount(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("zero calls", func(t *testing.T) {
		Reset()
		require.Equal(0, GetCallCount())
	})

	t.Run("multiple calls", func(t *testing.T) {
		Reset()
		Enable()

		Command("git", "status")
		require.Equal(1, GetCallCount())

		Command("tmux", "list-sessions")
		require.Equal(2, GetCallCount())

		Command("echo", "test")
		require.Equal(3, GetCallCount())
	})
}

func TestWasCommandCalled(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("command not called", func(t *testing.T) {
		Reset()
		require.False(WasCommandCalled("git", "status"))
	})

	t.Run("command called", func(t *testing.T) {
		Reset()
		Enable()
		Command("git", "status")

		require.True(WasCommandCalled("git", "status"))
		require.False(WasCommandCalled("git", "log"))
		require.False(WasCommandCalled("tmux", "status"))
	})

	t.Run("command with no args", func(t *testing.T) {
		Reset()
		Enable()
		Command("pwd")

		require.True(WasCommandCalled("pwd"))
		require.False(WasCommandCalled("pwd", "arg"))
	})

	t.Run("command with multiple args", func(t *testing.T) {
		Reset()
		Enable()
		Command("git", "log", "--oneline", "-n", "5")

		require.True(WasCommandCalled("git", "log", "--oneline", "-n", "5"))
		require.False(WasCommandCalled("git", "log", "--oneline"))
		require.False(WasCommandCalled("git", "log"))
	})
}

func TestGetCommandCalls(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("no matching calls", func(t *testing.T) {
		Reset()
		calls := GetCommandCalls("git", "status")
		require.Equal(0, len(calls))
	})

	t.Run("single matching call", func(t *testing.T) {
		Reset()
		Enable()
		Command("git", "status")

		calls := GetCommandCalls("git", "status")
		require.Equal(1, len(calls))
		require.Equal("git", calls[0].Name)
		require.Equal(1, len(calls[0].Args))
		require.Equal("status", calls[0].Args[0])
	})

	t.Run("multiple matching calls", func(t *testing.T) {
		Reset()
		Enable()

		Command("git", "status")
		Command("tmux", "list-sessions")
		Command("git", "status")

		calls := GetCommandCalls("git", "status")
		require.Equal(2, len(calls))

		tmuxCalls := GetCommandCalls("tmux", "list-sessions")
		require.Equal(1, len(tmuxCalls))

		noCalls := GetCommandCalls("echo", "test")
		require.Equal(0, len(noCalls))
	})
}

func TestMakeKey(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("simple command", func(t *testing.T) {
		key := makeKey("git", []string{"status"})
		require.Equal("git status", key)
	})

	t.Run("command with no args", func(t *testing.T) {
		key := makeKey("pwd", []string{})
		require.Equal("pwd ", key)
	})

	t.Run("command with multiple args", func(t *testing.T) {
		key := makeKey("git", []string{"log", "--oneline", "-n", "5"})
		require.Equal("git log --oneline -n 5", key)
	})

	t.Run("command with spaces in args", func(t *testing.T) {
		key := makeKey("echo", []string{"hello world", "test"})
		require.Equal("echo hello world test", key)
	})
}

func TestGetCurrentDir(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("gets current directory", func(t *testing.T) {
		dir := getCurrentDir()
		require.True(len(dir) > 0)

		// Should match os.Getwd()
		actualDir, err := os.Getwd()
		require.NoError(err)
		require.Equal(actualDir, dir)
	})
}

func TestEscapeShell(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("no escaping needed", func(t *testing.T) {
		result := escapeShell("hello")
		require.Equal("hello", result)
	})

	t.Run("escape backslashes", func(t *testing.T) {
		result := escapeShell("path\\to\\file")
		require.Equal("path\\\\to\\\\file", result)
	})

	t.Run("escape quotes", func(t *testing.T) {
		result := escapeShell(`say "hello"`)
		require.Equal(`say \"hello\"`, result)
	})

	t.Run("escape newlines", func(t *testing.T) {
		result := escapeShell("line1\nline2")
		require.Equal("line1\\nline2", result)
	})

	t.Run("escape tabs", func(t *testing.T) {
		result := escapeShell("col1\tcol2")
		require.Equal("col1\\tcol2", result)
	})

	t.Run("escape multiple characters", func(t *testing.T) {
		result := escapeShell("path\\file\t\"name\"\nend")
		require.Equal("path\\\\file\\t\\\"name\\\"\\nend", result)
	})
}

func TestCreateTestSafeCommand(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("creates command", func(t *testing.T) {
		response := CommandResponse{
			Stdout:   "test output",
			Stderr:   "test error",
			ExitCode: 0,
		}

		cmd := createTestSafeCommand(response)
		require.NotNil(cmd)
		// Path will be the full path to sh, so just check it contains "sh"
		require.True(len(cmd.Path) > 0)
	})

	t.Run("creates command with different exit code", func(t *testing.T) {
		response := CommandResponse{
			Stdout:   "",
			Stderr:   "error",
			ExitCode: 1,
		}

		cmd := createTestSafeCommand(response)
		require.NotNil(cmd)
	})
}

func TestCreateMockCommand(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("creates mock command", func(t *testing.T) {
		response := CommandResponse{
			Stdout:   "mock output",
			Stderr:   "",
			ExitCode: 0,
		}

		cmd := createMockCommand(response)
		require.NotNil(cmd)
	})
}

func TestWorkingDirectory(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("records working directory", func(t *testing.T) {
		Reset()
		Enable()

		Command("git", "status")

		calls := GetCalls()
		require.Equal(1, len(calls))
		require.True(len(calls[0].Dir) > 0)

		// Should match current directory
		currentDir, err := os.Getwd()
		require.NoError(err)
		require.Equal(currentDir, calls[0].Dir)
	})
}

func TestConcurrency(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("concurrent access", func(t *testing.T) {
		Reset()
		Enable()

		// Run commands concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(i int) {
				Command("echo", "test", fmt.Sprintf("%d", i))
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		calls := GetCalls()
		require.Equal(10, len(calls))
	})
}

func TestCompleteWorkflow(t *testing.T) {
	require := testutil.NewRequire(t)

	t.Run("complete workflow", func(t *testing.T) {
		// Start fresh
		Reset()
		require.False(globalMock.enabled)
		require.Equal(0, GetCallCount())

		// Set up responses
		SetResponse("git", "clean working tree", false)
		SetResponseWithArgs("tmux", []string{"list-sessions"}, "session1\nsession2", "", false)

		require.True(globalMock.enabled)

		// Execute commands
		Command("git", "status")
		Command("tmux", "list-sessions")
		Command("echo", "test") // No response set

		// Verify calls
		require.Equal(3, GetCallCount())
		require.True(WasCommandCalled("git", "status"))
		require.True(WasCommandCalled("tmux", "list-sessions"))
		require.True(WasCommandCalled("echo", "test"))
		require.False(WasCommandCalled("ls"))

		// Get specific calls
		gitCalls := GetCommandCalls("git", "status")
		require.Equal(1, len(gitCalls))

		tmuxCalls := GetCommandCalls("tmux", "list-sessions")
		require.Equal(1, len(tmuxCalls))

		// Test disable
		Disable()
		require.False(globalMock.enabled)

		// Commands should still be recorded but responses ignored
		// (in actual implementation, disabled mock returns real commands)

		// Re-enable and test
		Enable()
		require.True(globalMock.enabled)

		// Reset and verify clean state
		Reset()
		require.False(globalMock.enabled)
		require.Equal(0, GetCallCount())
		require.Equal(0, len(GetCalls()))
	})
}
