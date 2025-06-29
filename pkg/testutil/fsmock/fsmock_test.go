// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package fsmock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTempFS(t *testing.T) {
	fs := NewTempFS(t)

	// Check that root directory exists
	if !fs.Exists("") {
		t.Fatal("Root directory should exist")
	}

	if !fs.IsDir("") {
		t.Fatal("Root should be a directory")
	}

	// Check that root directory is in temp directory
	if !strings.Contains(fs.RootDir(), os.TempDir()) {
		t.Errorf("Root directory %s should be in temp directory", fs.RootDir())
	}
}

func TestWriteFileAndReadFile(t *testing.T) {
	fs := NewTempFS(t)

	content := "Hello, World!"
	filename := "test.txt"

	// Write file
	err := fs.WriteFileString(filename, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Check file exists
	if !fs.Exists(filename) {
		t.Fatal("File should exist after writing")
	}

	if !fs.IsFile(filename) {
		t.Fatal("Path should be a file")
	}

	// Read file back
	readContent, err := fs.ReadFileString(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if readContent != content {
		t.Errorf("Expected %q, got %q", content, readContent)
	}
}

func TestMkdirAll(t *testing.T) {
	fs := NewTempFS(t)

	dirPath := "deeply/nested/directory"

	err := fs.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	if !fs.Exists(dirPath) {
		t.Fatal("Directory should exist after creation")
	}

	if !fs.IsDir(dirPath) {
		t.Fatal("Path should be a directory")
	}
}

func TestWriteFileWithNestedDirs(t *testing.T) {
	fs := NewTempFS(t)

	filePath := "path/to/nested/file.txt"
	content := "nested content"

	err := fs.WriteFileString(filePath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to write nested file: %v", err)
	}

	// Check that parent directories were created
	if !fs.IsDir("path") {
		t.Fatal("Parent directory 'path' should exist")
	}

	if !fs.IsDir("path/to") {
		t.Fatal("Parent directory 'path/to' should exist")
	}

	if !fs.IsDir("path/to/nested") {
		t.Fatal("Parent directory 'path/to/nested' should exist")
	}

	// Check file content
	readContent, err := fs.ReadFileString(filePath)
	if err != nil {
		t.Fatalf("Failed to read nested file: %v", err)
	}

	if readContent != content {
		t.Errorf("Expected %q, got %q", content, readContent)
	}
}

func TestRemove(t *testing.T) {
	fs := NewTempFS(t)

	// Create a file
	filename := "removeme.txt"
	err := fs.WriteFileString(filename, "content", 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Remove it
	err = fs.Remove(filename)
	if err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}

	// Check it's gone
	if fs.Exists(filename) {
		t.Fatal("File should not exist after removal")
	}
}

func TestRemoveAll(t *testing.T) {
	fs := NewTempFS(t)

	// Create nested structure
	err := fs.WriteFileString("dir/subdir/file.txt", "content", 0644)
	if err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Remove entire directory
	err = fs.RemoveAll("dir")
	if err != nil {
		t.Fatalf("Failed to remove directory: %v", err)
	}

	// Check it's gone
	if fs.Exists("dir") {
		t.Fatal("Directory should not exist after removal")
	}
}

func TestPath(t *testing.T) {
	fs := NewTempFS(t)

	relativePath := "some/file.txt"
	absolutePath := fs.Path(relativePath)

	expected := filepath.Join(fs.RootDir(), relativePath)
	if absolutePath != expected {
		t.Errorf("Expected %q, got %q", expected, absolutePath)
	}
}

func TestCreateGitRepo(t *testing.T) {
	fs := NewTempFS(t)

	repoPath := "myrepo"
	err := fs.CreateGitRepo(repoPath)
	if err != nil {
		t.Fatalf("Failed to create git repo: %v", err)
	}

	// Check .git directory exists
	if !fs.IsDir(filepath.Join(repoPath, ".git")) {
		t.Fatal(".git directory should exist")
	}

	// Check config file exists
	configPath := filepath.Join(repoPath, ".git", "config")
	if !fs.IsFile(configPath) {
		t.Fatal("git config file should exist")
	}

	// Check HEAD file exists
	headPath := filepath.Join(repoPath, ".git", "HEAD")
	if !fs.IsFile(headPath) {
		t.Fatal("HEAD file should exist")
	}

	// Check HEAD content
	headContent, err := fs.ReadFileString(headPath)
	if err != nil {
		t.Fatalf("Failed to read HEAD file: %v", err)
	}

	expected := "ref: refs/heads/main\n"
	if headContent != expected {
		t.Errorf("Expected HEAD content %q, got %q", expected, headContent)
	}
}

func TestCreateProjectStructure(t *testing.T) {
	fs := NewTempFS(t)

	projectPath := "myproject"
	err := fs.CreateProjectStructure(projectPath)
	if err != nil {
		t.Fatalf("Failed to create project structure: %v", err)
	}

	// Check directories exist
	expectedDirs := []string{
		"cmd",
		"pkg",
		"internal",
		"test",
		"docs",
		".github/workflows",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(projectPath, dir)
		if !fs.IsDir(dirPath) {
			t.Errorf("Directory %s should exist", dirPath)
		}
	}

	// Check files exist
	expectedFiles := []string{
		"go.mod",
		"README.md",
		".gitignore",
		"Makefile",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(projectPath, file)
		if !fs.IsFile(filePath) {
			t.Errorf("File %s should exist", filePath)
		}
	}

	// Check go.mod content
	goModContent, err := fs.ReadFileString(filepath.Join(projectPath, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	if !strings.Contains(goModContent, "module example.com/test") {
		t.Errorf("go.mod should contain module declaration, got: %s", goModContent)
	}
}

func TestListDir(t *testing.T) {
	fs := NewTempFS(t)

	// Create some files and directories
	fs.WriteFileString("file1.txt", "content1", 0644)
	fs.WriteFileString("file2.txt", "content2", 0644)
	fs.MkdirAll("subdir", 0755)
	fs.WriteFileString("subdir/file3.txt", "content3", 0644)

	// List root directory
	entries, err := fs.ListDir("")
	if err != nil {
		t.Fatalf("Failed to list directory: %v", err)
	}

	// Check we have expected entries
	entryNames := make([]string, len(entries))
	for i, entry := range entries {
		entryNames[i] = entry.Name()
	}

	expectedNames := []string{"file1.txt", "file2.txt", "subdir"}
	for _, expected := range expectedNames {
		found := false
		for _, name := range entryNames {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find %s in directory listing", expected)
		}
	}
}

func TestTempFile(t *testing.T) {
	fs := NewTempFS(t)

	filePath, err := fs.TempFile("test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Check file exists
	if !fs.IsFile(filePath) {
		t.Fatal("Temp file should exist")
	}

	// Check it's in our temp filesystem
	if !strings.HasPrefix(filePath, fs.RootDir()) {
		t.Errorf("Temp file %s should be in our filesystem %s", filePath, fs.RootDir())
	}
}

func TestTempDir(t *testing.T) {
	fs := NewTempFS(t)

	dirPath, err := fs.TempDir("test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Check directory exists
	if !fs.IsDir(dirPath) {
		t.Fatal("Temp directory should exist")
	}

	// Check it's in our temp filesystem
	if !strings.HasPrefix(dirPath, fs.RootDir()) {
		t.Errorf("Temp directory %s should be in our filesystem %s", dirPath, fs.RootDir())
	}
}

func TestCleanupCalled(t *testing.T) {
	// This test verifies that cleanup happens automatically
	// We can't easily test the automatic cleanup without complex setup,
	// but we can test manual cleanup

	fs := NewTempFS(t)
	rootDir := fs.RootDir()

	// Create a file
	fs.WriteFileString("test.txt", "content", 0644)

	// Manually cleanup
	fs.Cleanup()

	// Check that the directory no longer exists
	if _, err := os.Stat(rootDir); !os.IsNotExist(err) {
		t.Errorf("Root directory %s should not exist after cleanup", rootDir)
	}

	// Cleanup again should be safe
	fs.Cleanup() // Should not panic or error
}
