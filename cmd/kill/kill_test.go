package kill

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/nehpz/claudicus/pkg/testutil"
	"github.com/nehpz/claudicus/pkg/testutil/fsmock"
)

func TestExecuteKill(t *testing.T) {
	require := testutil.NewRequire(t)
	ctx := context.Background()

	t.Run("no arguments provided", func(t *testing.T) {
		err := executeKill(ctx, []string{})
		require.Error(err)
		require.Equal("agent name argument is required", err.Error())
	})

	t.Run("nil arguments provided", func(t *testing.T) {
		err := executeKill(ctx, nil)
		require.Error(err)
		require.Equal("agent name argument is required", err.Error())
	})

	t.Run("empty string argument", func(t *testing.T) {
		err := executeKill(ctx, []string{""})
		require.Error(err)
		// Should not be argument validation error since empty string is still an argument
		require.False(strings.Contains(err.Error(), "agent name argument is required"))
	})

	t.Run("state manager initialization failure", func(t *testing.T) {
		// Setup environment where state manager initialization fails
		originalArgs := os.Args
		os.Args = []string{"test"}
		defer func() { os.Args = originalArgs }()

		err := executeKill(ctx, []string{"test-agent"})
		require.Error(err)
		require.Equal("could not initialize state manager", err.Error())
	})

	t.Run("kill all with no active sessions", func(t *testing.T) {
		// Setup mock filesystem with empty state
		fs := fsmock.NewTempFS(t)
		defer fs.Cleanup()

		// Create empty state file
		stateDir := fs.Path(".local/share/uzi")
		fs.MkdirAll(stateDir, 0755)
		stateFile := fs.Path(".local/share/uzi/state.json")
		fs.WriteFileString(stateFile, "{}", 0644)

		// Mock git repository
		fs.CreateGitRepo(".")

		// Setup environment to use mock filesystem
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", fs.RootDir())
		defer os.Setenv("HOME", originalHome)

		originalArgs := os.Args
		os.Args = []string{fs.Path("claudicus")}
		defer func() { os.Args = originalArgs }()

		err := executeKill(ctx, []string{"all"})
		require.NoError(err) // killAll should succeed when no sessions found
	})

	t.Run("kill all with active sessions", func(t *testing.T) {
		// Setup mock filesystem
		fs := fsmock.NewTempFS(t)
		defer fs.Cleanup()

		// Create state file with active sessions
		stateDir := fs.Path(".local/share/uzi")
		fs.MkdirAll(stateDir, 0755)
		stateFile := fs.Path(".local/share/uzi/state.json")
		stateContent := `{ 
			"session1": { 
				"repository": "claudicus",
				"branchName": "test-branch",
				"worktreePath": "/tmp/test",
				"command": "claude",
				"prompt": "test prompt",
				"port": 3000,
				"createdAt": "2023-01-01T00:00:00Z" 
			},
			"session2": { 
				"repository": "claudicus",
				"branchName": "test-branch-2",
				"worktreePath": "/tmp/test2",
				"command": "cursor",
				"prompt": "test prompt 2",
				"port": 3001,
				"createdAt": "2023-01-01T00:00:00Z" 
			} 
		}`
		fs.WriteFileString(stateFile, stateContent, 0644)

		// Mock git repository
		fs.CreateGitRepo(".")

		// Setup environment to use mock filesystem
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", fs.RootDir())
		defer os.Setenv("HOME", originalHome)

		originalArgs := os.Args
		os.Args = []string{fs.Path("claudicus")}
		defer func() { os.Args = originalArgs }()

		err := executeKill(ctx, []string{"all"})
		// Will fail due to tmux/git operations in test environment, but should reach killAll logic
		require.Error(err) // Expected to fail at tmux commands
	})

	t.Run("kill specific agent not found", func(t *testing.T) {
		// Setup mock filesystem with empty state
		fs := fsmock.NewTempFS(t)
		defer fs.Cleanup()

		// Create empty state file
		stateDir := fs.Path(".local/share/uzi")
		fs.MkdirAll(stateDir, 0755)
		stateFile := fs.Path(".local/share/uzi/state.json")
		fs.WriteFileString(stateFile, "{}", 0644)

		// Mock git repository
		fs.CreateGitRepo(".")

		// Setup environment to use mock filesystem
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", fs.RootDir())
		defer os.Setenv("HOME", originalHome)

		originalArgs := os.Args
		os.Args = []string{fs.Path("claudicus")}
		defer func() { os.Args = originalArgs }()

		err := executeKill(ctx, []string{"nonexistent-agent"})
		require.Error(err)
		require.Equal("no active session found for agent: nonexistent-agent", err.Error())
	})

	t.Run("kill specific agent found", func(t *testing.T) {
		// Setup mock filesystem
		fs := fsmock.NewTempFS(t)
		defer fs.Cleanup()

		// Create state file with target agent session
		stateDir := fs.Path(".local/share/uzi")
		fs.MkdirAll(stateDir, 0755)
		stateFile := fs.Path(".local/share/uzi/state.json")
		stateContent := `{ 
			"session-prefix-target-agent": { 
				"repository": "claudicus",
				"branchName": "test-branch",
				"worktreePath": "/tmp/test",
				"command": "claude",
				"prompt": "test prompt",
				"port": 3000,
				"createdAt": "2023-01-01T00:00:00Z" 
			} 
		}`
		fs.WriteFileString(stateFile, stateContent, 0644)

		// Mock git repository
		fs.CreateGitRepo(".")

		// Setup environment to use mock filesystem
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", fs.RootDir())
		defer os.Setenv("HOME", originalHome)

		originalArgs := os.Args
		os.Args = []string{fs.Path("claudicus")}
		defer func() { os.Args = originalArgs }()

		err := executeKill(ctx, []string{"target-agent"})
		// Will fail due to tmux/git operations in test environment, but should reach killSession logic
		require.Error(err)
		// Should not be "no active session found" error
		require.False(strings.Contains(err.Error(), "no active session found for agent"))
	})

	t.Run("state manager error getting sessions", func(t *testing.T) {
		// Setup mock filesystem with invalid state file
		fs := fsmock.NewTempFS(t)
		defer fs.Cleanup()

		// Create invalid state file
		stateDir := fs.Path(".local/share/uzi")
		fs.MkdirAll(stateDir, 0755)
		stateFile := fs.Path(".local/share/uzi/state.json")
		fs.WriteFileString(stateFile, "invalid json", 0644)

		// Mock git repository
		fs.CreateGitRepo(".")

		// Setup environment to use mock filesystem
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", fs.RootDir())
		defer os.Setenv("HOME", originalHome)

		originalArgs := os.Args
		os.Args = []string{fs.Path("claudicus")}
		defer func() { os.Args = originalArgs }()

		err := executeKill(ctx, []string{"target-agent"})
		require.Error(err)
	})
}

func TestKillSession(t *testing.T) {
	require := testutil.NewRequire(t)
	ctx := context.Background()

	t.Run("kill session with basic parameters", func(t *testing.T) {
		// Setup mock filesystem
		fs := fsmock.NewTempFS(t)
		defer fs.Cleanup()

		// Create state manager
		stateDir := fs.Path(".local/share/uzi")
		fs.MkdirAll(stateDir, 0755)
		stateFile := fs.Path(".local/share/uzi/state.json")
		fs.WriteFileString(stateFile, "{}", 0644)

		// Mock git repository
		fs.CreateGitRepo(".")

		// Setup environment to use mock filesystem
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", fs.RootDir())
		defer os.Setenv("HOME", originalHome)

		originalArgs := os.Args
		os.Args = []string{fs.Path("claudicus")}
		defer func() { os.Args = originalArgs }()

		// This will fail due to tmux/git operations, but tests the killSession function path
		err := executeKill(ctx, []string{"test-agent"})
		require.Error(err) // Expected to fail due to test environment
	})
}

func TestKillAll(t *testing.T) {
	require := testutil.NewRequire(t)
	ctx := context.Background()

	t.Run("kill all with empty sessions", func(t *testing.T) {
		// Setup mock filesystem
		fs := fsmock.NewTempFS(t)
		defer fs.Cleanup()

		// Create empty state file
		stateDir := fs.Path(".local/share/uzi")
		fs.MkdirAll(stateDir, 0755)
		stateFile := fs.Path(".local/share/uzi/state.json")
		fs.WriteFileString(stateFile, "{}", 0644)

		// Mock git repository
		fs.CreateGitRepo(".")

		// Setup environment to use mock filesystem
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", fs.RootDir())
		defer os.Setenv("HOME", originalHome)

		originalArgs := os.Args
		os.Args = []string{fs.Path("claudicus")}
		defer func() { os.Args = originalArgs }()

		err := executeKill(ctx, []string{"all"})
		require.NoError(err) // Should succeed when no sessions found
	})

	t.Run("kill all with state manager error", func(t *testing.T) {
		// Setup mock filesystem with invalid state file
		fs := fsmock.NewTempFS(t)
		defer fs.Cleanup()

		// Create invalid state file
		stateDir := fs.Path(".local/share/uzi")
		fs.MkdirAll(stateDir, 0755)
		stateFile := fs.Path(".local/share/uzi/state.json")
		fs.WriteFileString(stateFile, "invalid json", 0644)

		// Mock git repository
		fs.CreateGitRepo(".")

		// Setup environment to use mock filesystem
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", fs.RootDir())
		defer os.Setenv("HOME", originalHome)

		originalArgs := os.Args
		os.Args = []string{fs.Path("claudicus")}
		defer func() { os.Args = originalArgs }()

		err := executeKill(ctx, []string{"all"})
		require.Error(err) // Should fail due to invalid state file
	})
}

func TestCmdKillGlobalVariable(t *testing.T) {
	require := testutil.NewRequire(t)

	// Test global command configuration
	require.NotNil(CmdKill)
	require.Equal("kill", CmdKill.Name)
	require.Equal("uzi kill [<agent-name>|all]", CmdKill.ShortUsage)
	require.Equal("Delete tmux session and git worktree for the specified agent", CmdKill.ShortHelp)
	require.NotNil(CmdKill.FlagSet)
	require.NotNil(CmdKill.Exec)
	require.Equal("uzi kill", CmdKill.FlagSet.Name())
}

func TestKillCommandExecution(t *testing.T) {
	require := testutil.NewRequire(t)
	ctx := context.Background()

	// Test command execution through the global command
	err := CmdKill.Exec(ctx, []string{})
	require.Error(err)
	require.Equal("agent name argument is required", err.Error())

	// Test with valid argument (will fail due to state manager in test env)
	err = CmdKill.Exec(ctx, []string{"test-agent"})
	require.Error(err)
	// Should not be argument validation error
	require.False(strings.Contains(err.Error(), "agent name argument is required"))
}

func TestArgumentValidation(t *testing.T) {
	require := testutil.NewRequire(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		errText   string
	}{
		{
			name:      "no_arguments",
			args:      []string{},
			shouldErr: true,
			errText:   "agent name argument is required",
		},
		{
			name:      "nil_arguments",
			args:      nil,
			shouldErr: true,
			errText:   "agent name argument is required",
		},
		{
			name:      "one_argument",
			args:      []string{"agent"},
			shouldErr: false, // Should pass validation, fail later
		},
		{
			name:      "multiple_arguments",
			args:      []string{"agent", "extra"},
			shouldErr: false, // Should pass validation, fail later
		},
		{
			name:      "all_argument",
			args:      []string{"all"},
			shouldErr: false, // Should pass validation
		},
		{
			name:      "empty_string_argument",
			args:      []string{""},
			shouldErr: false, // Empty string is still an argument
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment to fail at state manager for non-validation errors
			originalArgs := os.Args
			os.Args = []string{"test"}
			defer func() { os.Args = originalArgs }()

			err := executeKill(ctx, tt.args)

			if tt.shouldErr {
				require.Error(err)
				require.True(strings.Contains(err.Error(), tt.errText))
			} else {
				// Should not get validation error, but may get other errors
				if err != nil {
					require.False(strings.Contains(err.Error(), "agent name argument is required"))
				}
			}
		})
	}
}

func TestSessionNameMatching(t *testing.T) {
	require := testutil.NewRequire(t)
	ctx := context.Background()

	// Test session name matching logic
	tests := []struct {
		name           string
		sessionName    string
		agentName      string
		shouldMatch    bool
	}{
		{
			name:        "exact suffix match",
			sessionName: "prefix-myagent",
			agentName:   "myagent",
			shouldMatch: true,
		},
		{
			name:        "hyphenated agent name",
			sessionName: "prefix-my-complex-agent",
			agentName:   "my-complex-agent",
			shouldMatch: true,
		},
		{
			name:        "no match",
			sessionName: "prefix-otheragent",
			agentName:   "myagent",
			shouldMatch: false,
		},
		{
			name:        "partial match should not work",
			sessionName: "prefix-myagentlong",
			agentName:   "myagent",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock filesystem
			fs := fsmock.NewTempFS(t)
			defer fs.Cleanup()

			// Create state file with the test session
			stateDir := fs.Path(".local/share/uzi")
			fs.MkdirAll(stateDir, 0755)
			stateFile := fs.Path(".local/share/uzi/state.json")
			stateContent := `{ 
				"` + tt.sessionName + `": { 
					"repository": "claudicus",
					"branchName": "test-branch",
					"worktreePath": "/tmp/test",
					"command": "claude",
					"prompt": "test prompt",
					"port": 3000,
					"createdAt": "2023-01-01T00:00:00Z" 
				} 
			}`
			fs.WriteFileString(stateFile, stateContent, 0644)

			// Mock git repository
			fs.CreateGitRepo(".")

			// Setup environment to use mock filesystem
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", fs.RootDir())
			defer os.Setenv("HOME", originalHome)

			originalArgs := os.Args
			os.Args = []string{fs.Path("claudicus")}
			defer func() { os.Args = originalArgs }()

			err := executeKill(ctx, []string{tt.agentName})
			require.Error(err) // Will fail at tmux operations in test env

			if tt.shouldMatch {
				// Should get past the "no active session found" error
				require.False(strings.Contains(err.Error(), "no active session found for agent"))
			} else {
				// Should get the "no active session found" error
				require.True(strings.Contains(err.Error(), "no active session found for agent"))
			}
		})
	}
}

func TestComprehensiveCoverage(t *testing.T) {
	require := testutil.NewRequire(t)
	ctx := context.Background()

	// Test various argument combinations to ensure full code coverage
	testCases := []struct {
		name string
		args []string
		desc string
	}{
		{
			name: "coverage_empty_args",
			args: []string{},
			desc: "Tests argument validation with empty args",
		},
		{
			name: "coverage_nil_args",
			args: nil,
			desc: "Tests argument validation with nil args",
		},
		{
			name: "coverage_all_command",
			args: []string{"all"},
			desc: "Tests killAll execution path",
		},
		{
			name: "coverage_specific_agent",
			args: []string{"test-agent"},
			desc: "Tests specific agent kill path",
		},
		{
			name: "coverage_hyphenated_agent",
			args: []string{"my-complex-agent"},
			desc: "Tests agent with hyphens",
		},
		{
			name: "coverage_special_chars",
			args: []string{"agent_with_underscore"},
			desc: "Tests agent with special characters",
		},
		{
			name: "coverage_unicode",
			args: []string{"agent世界"},
			desc: "Tests agent with unicode characters",
		},
		{
			name: "coverage_multiple_args",
			args: []string{"agent", "extra", "args"},
			desc: "Tests with multiple arguments",
		},
		{
			name: "coverage_empty_string",
			args: []string{""},
			desc: "Tests with empty string argument",
		},
		{
			name: "coverage_whitespace",
			args: []string{" "},
			desc: "Tests with whitespace argument",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment to fail at state manager for non-validation errors
			originalArgs := os.Args
			os.Args = []string{"test"}
			defer func() { os.Args = originalArgs }()

			err := executeKill(ctx, tc.args)
			require.Error(err) // All should error in test environment

			if len(tc.args) == 0 || tc.args == nil {
				// Should get argument validation error
				require.True(strings.Contains(err.Error(), "agent name argument is required"))
			} else {
				// Should not get argument validation error
				require.False(strings.Contains(err.Error(), "agent name argument is required"))
			}

			t.Logf("%s: %v", tc.desc, err)
		})
	}
}

