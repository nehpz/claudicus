package state

import (
	"io/fs"
	"os"
)

// FileSystem abstracts filesystem operations for testability
// This follows the Dependency Inversion Principle
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (fs.FileInfo, error)
	UserHomeDir() (string, error)
	RemoveAll(path string) error
}

// DefaultFileSystem implements FileSystem using standard os package
type DefaultFileSystem struct{}

// ReadFile reads the named file and returns the contents
func (d *DefaultFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// WriteFile writes data to the named file, creating it if necessary
func (d *DefaultFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// MkdirAll creates a directory named path along with any necessary parents
func (d *DefaultFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Stat returns file information about the named file
func (d *DefaultFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// UserHomeDir returns the current user's home directory
func (d *DefaultFileSystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// RemoveAll removes path and any children it contains
func (d *DefaultFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// NewDefaultFileSystem creates a new DefaultFileSystem
func NewDefaultFileSystem() FileSystem {
	return &DefaultFileSystem{}
}
