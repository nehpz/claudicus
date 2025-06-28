package run

import (
    "context"
    "testing"
    "github.com/charmbracelet/log"
    "github.com/nehpz/claudicus/pkg/state"
    "github.com/nehpz/claudicus/pkg/testutil/cmdmock"
)

func TestExecuteRun(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        wantErr     bool
        errMsg      string
        prepareMock func()
    }{
        {
            name: "run_command",
            args: []string{"echo", "test"},
            prepareMock: func() {
                state.MockActiveSessions([]string{"runSession"})
                cmdmock.SetResponse("tmux send-keys", "", false)
            },
            wantErr: false,
        },
        {
            name:    "no_command",
            args:    []string{},
            wantErr: true,
            errMsg:  "command argument is required",
        },
        {
            name: "run_error",
            args: []string{"echo", "test"},
            prepareMock: func() {
                cmdmock.SetResponse("tmux send-keys", "error", true)
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.prepareMock != nil {
                tt.prepareMock()
            }

            log.SetDefault(log.NewLogfmtLogger(log.StdlibWriter{}))
            log.Default().SetLevel(log.DebugLevel)

            err := executeRun(context.Background(), tt.args)
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error but got none")
                }
                if tt.errMsg != "" && err != nil && err.Error() != tt.errMsg {
                    t.Errorf("expected error message '%s', but got '%v'", tt.errMsg, err)
                }
            } else if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
