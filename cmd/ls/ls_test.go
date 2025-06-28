package ls

import (
    "context"
    "testing"
    "github.com/charmbracelet/log"
    "github.com/nehpz/claudicus/pkg/state"
    "github.com/nehpz/claudicus/pkg/testutil/cmdmock"
)

func TestExecuteLs(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        wantErr     bool
        errMsg      string
        prepareMock func()
    }{
        {
            name: "list_sessions",
            args: []string{},
            prepareMock: func() {
                state.MockActiveSessions([]string{"session1"})
                cmdmock.SetResponse("tmux list-sessions", "session1", false)
            },
            wantErr: false,
        },
        {
            name: "json_output",
            args: []string{"--json"},
            prepareMock: func() {
                state.MockActiveSessions([]string{"session-json"})
                cmdmock.SetResponse("tmux list-sessions", "session-json", false)
            },
            wantErr: false,
        },
        {
            name: "error_case",
            args: []string{},
            prepareMock: func() {
                cmdmock.SetResponse("tmux list-sessions", "error", true)
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

            err := executeLs(context.Background(), tt.args)
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

