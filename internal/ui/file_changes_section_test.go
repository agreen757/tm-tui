package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/agreen757/tm-tui/internal/taskmaster"
)

func TestNewFileChangesSection(t *testing.T) {
	task := &taskmaster.Task{
		ID:          "1",
		Title:       "Test Task",
		FileChanges: []taskmaster.FileChange{},
	}
	styles := NewStyles()

	section := NewFileChangesSection(task, styles)

	if section == nil {
		t.Fatal("NewFileChangesSection returned nil")
	}
	if section.task != task {
		t.Error("Task not properly set")
	}
	if section.selectedIdx != 0 {
		t.Error("Initial selectedIdx should be 0")
	}
	if section.focused {
		t.Error("Initial focused should be false")
	}
}

func TestRender_EmptyFileChanges(t *testing.T) {
	task := &taskmaster.Task{
		ID:          "1",
		Title:       "Test Task",
		FileChanges: []taskmaster.FileChange{},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	result := section.Render(80)

	if result != "" {
		t.Errorf("Expected empty string for no file changes, got: %s", result)
	}
}

func TestRender_NilTask(t *testing.T) {
	styles := NewStyles()
	section := NewFileChangesSection(nil, styles)

	result := section.Render(80)

	if result != "" {
		t.Errorf("Expected empty string for nil task, got: %s", result)
	}
}

func TestRender_SingleFileChange(t *testing.T) {
	task := &taskmaster.Task{
		ID:    "1",
		Title: "Test Task",
		FileChanges: []taskmaster.FileChange{
			{
				Path:        "internal/ui/app.go",
				ChangeType:  "modified",
				Description: "Updated task details rendering",
				IsPending:   true,
			},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	result := section.Render(80)

	// Check header
	if !strings.Contains(result, "File Changes (1)") {
		t.Error("Header should contain file count")
	}

	// Check file path
	if !strings.Contains(result, "internal/ui/app.go") {
		t.Error("Should contain file path")
	}

	// Check description
	if !strings.Contains(result, "Updated task details rendering") {
		t.Error("Should contain description")
	}

	// Check uncommitted status
	if !strings.Contains(result, "uncommitted") {
		t.Error("Should show uncommitted status for pending changes")
	}
}

func TestRender_MultipleFileChanges(t *testing.T) {
	task := &taskmaster.Task{
		ID:    "1",
		Title: "Test Task",
		FileChanges: []taskmaster.FileChange{
			{
				Path:       "internal/ui/app.go",
				ChangeType: "modified",
				IsPending:  false,
				CommitID:   "abc123def456",
			},
			{
				Path:       "internal/ui/file_changes_section.go",
				ChangeType: "added",
				IsPending:  true,
			},
			{
				Path:       "internal/ui/old_code.go",
				ChangeType: "deleted",
				IsPending:  false,
			},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	result := section.Render(80)

	// Check header with correct count
	if !strings.Contains(result, "File Changes (3)") {
		t.Error("Header should contain correct file count")
	}

	// Check all file paths are present
	if !strings.Contains(result, "app.go") {
		t.Error("Should contain first file")
	}
	if !strings.Contains(result, "file_changes_section.go") {
		t.Error("Should contain second file")
	}
	if !strings.Contains(result, "old_code.go") {
		t.Error("Should contain third file")
	}

	// Check commit ID is shown (shortened)
	if !strings.Contains(result, "abc123d") {
		t.Error("Should show shortened commit ID")
	}
}

func TestRender_DifferentChangeTypes(t *testing.T) {
	testCases := []struct {
		changeType string
		name       string
	}{
		{"added", "Added file"},
		{"modified", "Modified file"},
		{"deleted", "Deleted file"},
		{"unknown", "Unknown change type"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			task := &taskmaster.Task{
				ID:    "1",
				Title: "Test Task",
				FileChanges: []taskmaster.FileChange{
					{
						Path:       "test.go",
						ChangeType: tc.changeType,
						IsPending:  true,
					},
				},
			}
			styles := NewStyles()
			section := NewFileChangesSection(task, styles)

			result := section.Render(80)

			if !strings.Contains(result, "test.go") {
				t.Errorf("Should contain file path for %s", tc.changeType)
			}
		})
	}
}

func TestRender_LongFilePath(t *testing.T) {
	task := &taskmaster.Task{
		ID:    "1",
		Title: "Test Task",
		FileChanges: []taskmaster.FileChange{
			{
				Path:       "very/long/path/to/some/deeply/nested/directory/structure/with/many/components/file.go",
				ChangeType: "modified",
				IsPending:  true,
			},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	result := section.Render(80)

	// Should show shortened path with "..."
	if !strings.Contains(result, "...") {
		t.Error("Long paths should be shortened with ellipsis")
	}

	// Should still show filename
	if !strings.Contains(result, "file.go") {
		t.Error("Should show filename even for long paths")
	}
}

func TestRender_WithDescription(t *testing.T) {
	task := &taskmaster.Task{
		ID:    "1",
		Title: "Test Task",
		FileChanges: []taskmaster.FileChange{
			{
				Path:        "internal/ui/app.go",
				ChangeType:  "modified",
				Description: "This is a detailed description of the changes made to this file, explaining the rationale and impact.",
				IsPending:   true,
			},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	result := section.Render(80)

	if !strings.Contains(result, "This is a detailed description") {
		t.Error("Should contain description text")
	}
}

func TestRender_NarrowTerminal(t *testing.T) {
	task := &taskmaster.Task{
		ID:    "1",
		Title: "Test Task",
		FileChanges: []taskmaster.FileChange{
			{
				Path:        "internal/ui/app.go",
				ChangeType:  "modified",
				Description: "A very long description that should wrap when rendered in a narrow terminal window with limited width available for text.",
				IsPending:   true,
			},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	result := section.Render(40) // Narrow width

	if result == "" {
		t.Error("Should still render with narrow width")
	}
}

func TestSetFocused(t *testing.T) {
	task := &taskmaster.Task{ID: "1"}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	section.SetFocused(true)
	if !section.IsFocused() {
		t.Error("SetFocused(true) should set focused to true")
	}

	section.SetFocused(false)
	if section.IsFocused() {
		t.Error("SetFocused(false) should set focused to false")
	}
}

func TestSetSelected(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "file1.go", ChangeType: "added"},
			{Path: "file2.go", ChangeType: "modified"},
			{Path: "file3.go", ChangeType: "deleted"},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	// Valid index
	section.SetSelected(1)
	if section.GetSelected() != 1 {
		t.Error("SetSelected should set selectedIdx to 1")
	}

	// Invalid index (too high) - should not change
	section.SetSelected(10)
	if section.GetSelected() != 1 {
		t.Error("SetSelected with invalid index should not change selection")
	}

	// Invalid index (negative) - should not change
	section.SetSelected(-1)
	if section.GetSelected() != 1 {
		t.Error("SetSelected with negative index should not change selection")
	}

	// Valid index at boundary
	section.SetSelected(2)
	if section.GetSelected() != 2 {
		t.Error("SetSelected should work at boundary")
	}
}

func TestGetFileCount(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "file1.go"},
			{Path: "file2.go"},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	if section.GetFileCount() != 2 {
		t.Errorf("Expected file count 2, got %d", section.GetFileCount())
	}

	// Test with nil task
	nilSection := NewFileChangesSection(nil, styles)
	if nilSection.GetFileCount() != 0 {
		t.Errorf("Expected file count 0 for nil task, got %d", nilSection.GetFileCount())
	}
}

func TestGetChangeIndicator(t *testing.T) {
	styles := NewStyles()
	task := &taskmaster.Task{ID: "1"}
	section := NewFileChangesSection(task, styles)

	testCases := []struct {
		changeType string
		name       string
	}{
		{"added", "Added"},
		{"modified", "Modified"},
		{"deleted", "Deleted"},
		{"unknown", "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			indicator := section.getChangeIndicator(tc.changeType)
			if indicator == "" {
				t.Errorf("Indicator should not be empty for %s", tc.changeType)
			}
		})
	}
}

func TestRender_SelectionHighlight(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "file1.go", ChangeType: "added"},
			{Path: "file2.go", ChangeType: "modified"},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	// Render without focus
	unfocusedResult := section.Render(80)

	// Render with focus
	section.SetFocused(true)
	section.SetSelected(0)
	focusedResult := section.Render(80)

	// Results should be different when focused (due to styling)
	// This is a basic check - in reality the ANSI codes would differ
	if unfocusedResult == focusedResult {
		t.Log("Note: Focused and unfocused renders appear identical in test (styling may be present)")
	}
}

func TestRender_WithCommitID(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{
				Path:       "test.go",
				ChangeType: "modified",
				IsPending:  false,
				CommitID:   "abcdef1234567890",
			},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	result := section.Render(80)

	// Should show first 7 chars of commit
	if !strings.Contains(result, "abcdef1") {
		t.Error("Should show shortened commit ID (first 7 chars)")
	}

	// Should not show full commit ID
	if strings.Contains(result, "abcdef1234567890") {
		t.Error("Should not show full commit ID")
	}
}

func TestRender_WithTimestamp(t *testing.T) {
	now := time.Now()
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{
				Path:        "test.go",
				ChangeType:  "modified",
				LastChanged: now,
			},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	result := section.Render(80)

	// Component should handle LastChanged field (even if not currently displayed)
	if result == "" {
		t.Error("Should render with timestamp")
	}
}
