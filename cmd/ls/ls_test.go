package ls

import (
	"context"
	"strings"
	"testing"
)

func TestExecuteLs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errorSubstr string
	}{
		{
			name:        "normal_list_no_sessions",
			args:        []string{},
			wantErr:     false, // Returns nil when no sessions found
			errorSubstr: "",
		},
		{
			name:        "json_output_no_sessions",
			args:        []string{"--json"},
			wantErr:     false, // Returns nil when no sessions found
			errorSubstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeLs(context.Background(), tt.args)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("executeLs() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("executeLs() error = %v, want substring %v", err, tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("executeLs() unexpected error = %v", err)
				}
			}
		})
	}
}

