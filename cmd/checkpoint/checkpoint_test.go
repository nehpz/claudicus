package checkpoint

import (
	"context"
	"strings"
	"testing"
)

// TestExecuteCheckpoint tests the main executeCheckpoint function using table-driven tests
func TestExecuteCheckpoint(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errorSubstr string
	}{
		{
			name:        "no_arguments",
			args:        []string{},
			wantErr:     true,
			errorSubstr: "agent name and commit message arguments are required",
		},
		{
			name:        "one_argument_only",
			args:        []string{"agentName"},
			wantErr:     true,
			errorSubstr: "agent name and commit message arguments are required",
		},
		{
			name:        "valid_arguments_no_active_sessions",
			args:        []string{"agentName", "commit message"},
			wantErr:     true,
			errorSubstr: "no active session found for agent",
		},
		{
			name:        "multiple_arguments",
			args:        []string{"agentName", "commit", "message", "with", "spaces"},
			wantErr:     true,
			errorSubstr: "no active session found for agent", // Will pass validation but fail at session lookup
		},
		{
			name:        "empty_strings",
			args:        []string{"", ""},
			wantErr:     true,
			errorSubstr: "no active session found for agent", // Will pass validation but fail at session lookup
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := executeCheckpoint(context.Background(), tt.args)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Errorf("executeCheckpoint() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("executeCheckpoint() error = %v, want substring %v", err, tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("executeCheckpoint() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestCmdCheckpointGlobalVariable tests the global CmdCheckpoint command variable
func TestCmdCheckpointGlobalVariable(t *testing.T) {
	if CmdCheckpoint == nil {
		t.Fatal("CmdCheckpoint should not be nil")
	}

	if CmdCheckpoint.Name != "checkpoint" {
		t.Errorf("CmdCheckpoint.Name = %v, want %v", CmdCheckpoint.Name, "checkpoint")
	}

	if CmdCheckpoint.ShortUsage != "uzi checkpoint <agent-name> <commit-message>" {
		t.Errorf("CmdCheckpoint.ShortUsage = %v, want %v", CmdCheckpoint.ShortUsage, "uzi checkpoint <agent-name> <commit-message>")
	}

	if CmdCheckpoint.ShortHelp != "Rebase changes from an agent worktree into the current worktree and commit" {
		t.Errorf("CmdCheckpoint.ShortHelp = %v, want expected help text", CmdCheckpoint.ShortHelp)
	}

	if CmdCheckpoint.FlagSet == nil {
		t.Error("CmdCheckpoint.FlagSet should not be nil")
	}

	if CmdCheckpoint.Exec == nil {
		t.Error("CmdCheckpoint.Exec should not be nil")
	}
}

// TestArgumentValidation tests the argument validation logic specifically
func TestArgumentValidation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		shouldErr bool
		errorText string
	}{
		{
			name:      "nil_arguments",
			args:      nil,
			shouldErr: true,
			errorText: "agent name and commit message arguments are required",
		},
		{
			name:      "empty_slice",
			args:      []string{},
			shouldErr: true,
			errorText: "agent name and commit message arguments are required",
		},
		{
			name:      "one_argument",
			args:      []string{"agent"},
			shouldErr: true,
			errorText: "agent name and commit message arguments are required",
		},
		{
			name:      "two_arguments_pass_validation",
			args:      []string{"agent", "message"},
			shouldErr: true,
			errorText: "no active session found for agent", // Should pass validation, fail at session lookup
		},
		{
			name:      "multiple_arguments_pass_validation",
			args:      []string{"agent", "commit", "message", "with", "spaces"},
			shouldErr: true,
			errorText: "no active session found for agent", // Should pass validation, fail at session lookup
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeCheckpoint(context.Background(), tt.args)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error, but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorText, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
