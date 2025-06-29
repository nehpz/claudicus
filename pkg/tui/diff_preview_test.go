// Copyright (c) Subtrace, Inc.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"strings"
	"testing"
)

func TestDiffPreviewModel_NewModel(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)

	if model.width != 80 {
		t.Errorf("Expected width 80, got %d", model.width)
	}
	if model.height != 24 {
		t.Errorf("Expected height 24, got %d", model.height)
	}
	if model.showCommits {
		t.Error("Expected showCommits to be false initially")
	}
}

func TestDiffPreviewModel_SetSize(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)
	model.SetSize(100, 30)

	if model.width != 100 {
		t.Errorf("Expected width 100, got %d", model.width)
	}
	if model.height != 30 {
		t.Errorf("Expected height 30, got %d", model.height)
	}
}

func TestDiffPreviewModel_ToggleView(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)

	// Initially should show diff
	if model.showCommits {
		t.Error("Expected showCommits to be false initially")
	}

	// Toggle to commits view
	model.ToggleView()
	if !model.showCommits {
		t.Error("Expected showCommits to be true after toggle")
	}

	// Toggle back to diff view
	model.ToggleView()
	if model.showCommits {
		t.Error("Expected showCommits to be false after second toggle")
	}
}

func TestDiffPreviewModel_LoadDiff_NilSession(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)

	// Load with nil session should clear everything
	model.LoadDiff(nil)

	if model.content != "" {
		t.Error("Expected content to be empty for nil session")
	}
	if model.commitMessages != "" {
		t.Error("Expected commitMessages to be empty for nil session")
	}
	if model.changedFiles != "" {
		t.Error("Expected changedFiles to be empty for nil session")
	}
	if model.error != "" {
		t.Error("Expected error to be empty for nil session")
	}
}

func TestDiffPreviewModel_View_EmptyState(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)

	view := model.View()

	// Should contain help text
	if !strings.Contains(view, "Select an agent") {
		t.Error("Expected empty state message")
	}
	if !strings.Contains(view, "Press 'v' to toggle") {
		t.Error("Expected toggle instruction with 'v' key")
	}
}

func TestDiffPreviewModel_View_TitleChanges(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)

	// Test diff view title
	view := model.View()
	if !strings.Contains(view, "Git Diff") {
		t.Error("Expected 'Git Diff' title in diff view")
	}

	// Toggle to commits view
	model.ToggleView()
	view = model.View()
	if !strings.Contains(view, "Commits & Files") {
		t.Error("Expected 'Commits & Files' title in commits view")
	}
}

func TestDiffPreviewModel_FormatCommitsAndFiles(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)

	// Set some test data
	model.commitMessages = "abc123 Fix bug (John, 2 hours ago)\ndef456 Add feature (Jane, 1 day ago)"
	model.changedFiles = " M file1.go\nA  file2.go\nD  file3.go"

	result := model.formatCommitsAndFiles()

	// Should contain section headers
	if !strings.Contains(result, "Recent Commits:") {
		t.Error("Expected 'Recent Commits:' header")
	}
	if !strings.Contains(result, "Changed Files:") {
		t.Error("Expected 'Changed Files:' header")
	}

	// Should contain commit messages
	if !strings.Contains(result, "abc123") {
		t.Error("Expected commit hash in output")
	}
	if !strings.Contains(result, "Fix bug") {
		t.Error("Expected commit message in output")
	}

	// Should contain file changes
	if !strings.Contains(result, "file1.go") {
		t.Error("Expected modified file in output")
	}
	if !strings.Contains(result, "file2.go") {
		t.Error("Expected added file in output")
	}
	if !strings.Contains(result, "file3.go") {
		t.Error("Expected deleted file in output")
	}

	// Should contain toggle instruction
	if !strings.Contains(result, "Press 'v' to toggle") {
		t.Error("Expected toggle instruction with 'v' key")
	}
}

func TestDiffPreviewModel_FormatDiffContent(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)

	diffContent := `--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 func main() {
+    fmt.Println("Hello")
     return
-    // old comment
 }`

	result := model.formatDiffContent(diffContent)

	// Should contain the diff content
	if !strings.Contains(result, "func main()") {
		t.Error("Expected function content in formatted diff")
	}
	if !strings.Contains(result, "Hello") {
		t.Error("Expected added line content in formatted diff")
	}
}

func TestDiffPreviewModel_ErrorHandling(t *testing.T) {
	model := NewDiffPreviewModel(80, 24)

	// Set an error
	model.error = "Test error message"

	view := model.View()

	// Should display error message
	if !strings.Contains(view, "Test error message") {
		t.Error("Expected error message to be displayed")
	}
}
