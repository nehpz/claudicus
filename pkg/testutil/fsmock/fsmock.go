// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

// Package fsmock provides utilities for creating temporary directories and files
// that are automatically cleaned up after tests. This is particularly useful for
// tests that need to create files or directories but want automatic cleanup.
package fsmock

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TempFS represents a temporary filesystem for testing
type TempFS struct {
	rootDir string
	t       testing.TB
	cleaned bool
}

// NewTempFS creates a new temporary filesystem for testing.
// The returned TempFS will automatically clean up when the test ends.
func NewTempFS(t testing.TB) *TempFS {
	t.Helper()

	rootDir, err := os.MkdirTemp("", "fsmock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	tempFS := &TempFS{
		rootDir: rootDir,
		t:       t,
		cleaned: false,
	}

	// Register cleanup function
	t.Cleanup(func() {
		tempFS.Cleanup()
	})

	return tempFS
}

// RootDir returns the root directory of the temporary filesystem
func (fs *TempFS) RootDir() string {
	return fs.rootDir
}

// MkdirAll creates a directory and all necessary parent directories
func (fs *TempFS) MkdirAll(path string, perm os.FileMode) error {
	fullPath := fs.resolvePath(path)
	return os.MkdirAll(fullPath, perm)
}

// WriteFile creates a file with the given content
func (fs *TempFS) WriteFile(path string, content []byte, perm os.FileMode) error {
	fullPath := fs.resolvePath(path)

	// Ensure parent directories exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	return os.WriteFile(fullPath, content, perm)
}

// WriteFileString creates a file with the given string content
func (fs *TempFS) WriteFileString(path string, content string, perm os.FileMode) error {
	return fs.WriteFile(path, []byte(content), perm)
}

// ReadFile reads the content of a file
func (fs *TempFS) ReadFile(path string) ([]byte, error) {
	fullPath := fs.resolvePath(path)
	return os.ReadFile(fullPath)
}

// ReadFileString reads the content of a file as a string
func (fs *TempFS) ReadFileString(path string) (string, error) {
	content, err := fs.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// Exists checks if a file or directory exists
func (fs *TempFS) Exists(path string) bool {
	fullPath := fs.resolvePath(path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// IsDir checks if the path is a directory
func (fs *TempFS) IsDir(path string) bool {
	fullPath := fs.resolvePath(path)
	info, err := os.Stat(fullPath)
	return err == nil && info.IsDir()
}

// IsFile checks if the path is a regular file
func (fs *TempFS) IsFile(path string) bool {
	fullPath := fs.resolvePath(path)
	info, err := os.Stat(fullPath)
	return err == nil && info.Mode().IsRegular()
}

// Remove removes a file or directory
func (fs *TempFS) Remove(path string) error {
	fullPath := fs.resolvePath(path)
	return os.Remove(fullPath)
}

// RemoveAll removes a file or directory and all its children
func (fs *TempFS) RemoveAll(path string) error {
	fullPath := fs.resolvePath(path)
	return os.RemoveAll(fullPath)
}

// Path returns the absolute path for a relative path within the temp filesystem
func (fs *TempFS) Path(path string) string {
	return fs.resolvePath(path)
}

// CreateGitRepo creates a basic git repository structure in the specified directory
func (fs *TempFS) CreateGitRepo(repoPath string) error {
	if err := fs.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil {
		return fmt.Errorf("failed to create .git directory: %w", err)
	}

	// Create a basic git config
	gitConfig := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[user]
	name = Test User
	email = test@example.com
`
	if err := fs.WriteFileString(filepath.Join(repoPath, ".git", "config"), gitConfig, 0644); err != nil {
		return fmt.Errorf("failed to create git config: %w", err)
	}

	// Create HEAD file
	if err := fs.WriteFileString(filepath.Join(repoPath, ".git", "HEAD"), "ref: refs/heads/main\n", 0644); err != nil {
		return fmt.Errorf("failed to create HEAD file: %w", err)
	}

	return nil
}

// CreateProjectStructure creates a typical project structure for testing
func (fs *TempFS) CreateProjectStructure(projectPath string) error {
	// Create common directories
	dirs := []string{
		"cmd",
		"pkg",
		"internal",
		"test",
		"docs",
		".github/workflows",
	}

	for _, dir := range dirs {
		if err := fs.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create common files
	files := map[string]string{
		"go.mod":     "module example.com/test\n\ngo 1.21\n",
		"README.md":  "# Test Project\n\nThis is a test project.\n",
		".gitignore": "*.log\n*.tmp\n.DS_Store\n",
		"Makefile":   ".PHONY: test\ntest:\n\tgo test ./...\n",
	}

	for filename, content := range files {
		if err := fs.WriteFileString(filepath.Join(projectPath, filename), content, 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}
	}

	return nil
}

// ListDir returns the contents of a directory
func (fs *TempFS) ListDir(path string) ([]fs.DirEntry, error) {
	fullPath := fs.resolvePath(path)
	return os.ReadDir(fullPath)
}

// Cleanup removes the entire temporary filesystem
func (fs *TempFS) Cleanup() {
	if fs.cleaned {
		return
	}

	fs.cleaned = true
	if err := os.RemoveAll(fs.rootDir); err != nil {
		// Use Logf if available (Go 1.14+), otherwise fall back to error handling
		if fs.t != nil {
			fs.t.Logf("Warning: failed to cleanup temporary directory %s: %v", fs.rootDir, err)
		}
	}
}

// resolvePath converts a relative path to an absolute path within the temp filesystem
func (fs *TempFS) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		// If it's already absolute, check if it's within our root
		if strings.HasPrefix(path, fs.rootDir) {
			return path
		}
		// If it's absolute but outside our root, treat as relative
		path = strings.TrimPrefix(path, "/")
	}
	return filepath.Join(fs.rootDir, path)
}

// TempFile creates a temporary file within the filesystem and returns its path
func (fs *TempFS) TempFile(pattern string) (string, error) {
	file, err := os.CreateTemp(fs.rootDir, pattern)
	if err != nil {
		return "", err
	}

	path := file.Name()
	file.Close()

	return path, nil
}

// TempDir creates a temporary directory within the filesystem and returns its path
func (fs *TempFS) TempDir(pattern string) (string, error) {
	return os.MkdirTemp(fs.rootDir, pattern)
}
