package reset

import (
    "context"
    "testing"
    "github.com/charmbracelet/log"
    "github.com/nehpz/claudicus/pkg/state"
    "github.com/nehpz/claudicus/pkg/testutil/cmdmock"
)

func TestExecuteReset(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        wantErr     bool
        errMsg      string
        prepareMock func()
    }{
        {
            name: "reset_clean",
            args: []string{},
            prepareMock: func() {
                state.MockActiveSessions([]string{"resetSession"})
                cmdmock.SetResponse("git reset --hard", "", false)
            },
            wantErr: false,
        },
        {
            name:    "reset_error",
            args:    []string{},
            wantErr: true,
            errMsg:  "command failed",
            prepareMock: func() {
                cmdmock.SetResponse("git reset --hard", "error", true)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.prepareMock != nil {
                tt.prepareMock()
            }

            log.SetDefault(log.NewLogfmtLogger(log.StdlibWriter{}))
            log.Default().SetLevel(log.DebugLevel)

            err := executeReset(context.Background(), tt.args)
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
