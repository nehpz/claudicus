package state

import (
	"os/exec"
)

// CommandExecutor abstracts command execution for testability
// This follows the Dependency Inversion Principle by depending on abstractions
type CommandExecutor interface {
	ExecuteCommand(name string, args ...string) ([]byte, error)
	RunCommand(name string, args ...string) error
}

// DefaultCommandExecutor implements CommandExecutor using exec.Command
type DefaultCommandExecutor struct{}

// ExecuteCommand runs a command and returns its output
func (d *DefaultCommandExecutor) ExecuteCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

// RunCommand runs a command and returns only the error
func (d *DefaultCommandExecutor) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
