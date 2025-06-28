package kill

import (
	"context"
	"strings"
	"testing"
)

func TestExecuteKill(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errorSubstr string
	}{
		{
			name:        "missing_agent_name",
			args:        []string{},
			wantErr:     true,
			errorSubstr: "agent name argument is required",
		},
		{
			name:        "kill_all_sessions_no_active",
			args:        []string{"all"},
			wantErr:     false, // killAll returns nil when no sessions found
			errorSubstr: "",
		},
		{
			name:        "kill_specific_agent_no_session",
			args:        []string{"agentName"},
			wantErr:     true, // Will fail when no session found for agent
			errorSubstr: "no active session found for agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeKill(context.Background(), tt.args)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("executeKill() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("executeKill() error = %v, want substring %v", err, tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("executeKill() unexpected error = %v", err)
				}
			}
		})
	}
}

