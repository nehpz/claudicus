package reset

import (
	"context"
	"strings"
	"testing"
)

func TestExecuteReset(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errorSubstr string
	}{
		{
			name:        "reset_user_input_error",
			args:        []string{},
			wantErr:     true, // Will fail due to EOF on user input
			errorSubstr: "failed to read user input",
		},
		{
			name:        "reset_with_args_input_error",
			args:        []string{"some", "args"},
			wantErr:     true, // Will fail due to EOF on user input
			errorSubstr: "failed to read user input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeReset(context.Background(), tt.args)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("executeReset() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("executeReset() error = %v, want substring %v", err, tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("executeReset() unexpected error = %v", err)
				}
			}
		})
	}
}
