package kill

import (
	"context"
	"testing"
	"github.com/charmbracelet/log"
	"github.com/nehpz/claudicus/pkg/state"
	"github.com/nehpz/claudicus/pkg/testutil/cmdmock"
)

func TestExecuteKill(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        wantErr     bool
        errMsg      string
        prepareMock func()
    }{
        {
            name:    "missing_agent_name",
            args:    []string{},
            wantErr: true,
            errMsg:  "agent name argument is required",
        },
        {
            name: "kill_all",
            args: []string{"all"},
            prepareMock: func() {
                state.MockActiveSessions([]string{"session1", "session2"})
                cmdmock.SetResponse("tmux kill-session -t session1", "session1 ended", false)
                cmdmock.SetResponse("tmux kill-session -t session2", "session2 ended", false)
            },
            wantErr: false,
        },
        {
            name: "kill_specific_agent",
            args: []string{"agentName"},
            prepareMock: func() {
                state.MockActiveSessions([]string{"repo-agentName"})
                cmdmock.SetResponse("tmux kill-session -t repo-agentName", "", false)
                cmdmock.SetResponse("git worktree remove --force", "", false)
                cmdmock.SetResponse("git branch -D agentName", "", false)
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.prepareMock != nil {
                tt.prepareMock()
            }

            log.SetDefault(log.NewLogfmtLogger(log.StdlibWriter{}))
            log.Default().SetLevel(log.DebugLevel)

            err := executeKill(context.Background(), tt.args)
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error but got none")
                }

                if tt.errMsg != "" {
                    if !log.HasError(err) {
                        t.Errorf("expected error message '%s', got '%v'", tt.errMsg, err)
                    }
                }
            } else if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}

package kill

import (
    "context"
    "testing"
    "github.com/charmbracelet/log"
    "github.com/nehpz/claudicus/pkg/state"
    "github.com/nehpz/claudicus/pkg/testutil/cmdmock"
)

func TestExecuteKill(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        prepareMock func()
        wantErr     bool
        errMsg      string
    }{
        {
            name:    "missing_agent_name",
            args:    []string{},
            wantErr: true,
            errMsg:  "agent name argument is required",
        },
        {
            name:    "valid_agent_name",
            args:    []string{"agentName"},
            prepareMock: func() {
                state.GetActiveSessionsForRepo = func() ([]string, error) {
                    return []string{"repo-agentName"}, nil
                }
                cmdmock.SetResponse("tmux kill-session -t repo-agentName", "", false)
                cmdmock.SetResponse("git worktree remove --force", "", false)
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.prepareMock != nil {
                tt.prepareMock()
            }

            log.SetDefault(log.NewLogfmtLogger(log.StdlibWriter{}))
            log.Default().SetLevel(log.DebugLevel)

            err := executeKill(context.Background(), tt.args)
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

