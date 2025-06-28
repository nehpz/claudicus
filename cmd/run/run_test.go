package run

import (
	"context"
	"strings"
	"testing"
)

func TestExecuteRun(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errorSubstr string
	}{
		{
			name:        "missing_command_error",
			args:        []string{},
			wantErr:     true,
			errorSubstr: "no command provided",
		},
		{
			name:        "run_command_with_args_no_session",
			args:        []string{"echo", "hello"},
			wantErr:     true, // Will fail due to no sessions in test environment
			errorSubstr: "no active agent sessions found",
		},
		{
			name:        "single_command_no_session",
			args:        []string{"ls"},
			wantErr:     true, // Will fail due to no sessions in test environment
			errorSubstr: "no active agent sessions found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeRun(context.Background(), tt.args)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("executeRun() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("executeRun() error = %v, want substring %v", err, tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("executeRun() unexpected error = %v", err)
				}
			}
		})
	}
}
