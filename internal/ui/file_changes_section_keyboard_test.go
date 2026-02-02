package ui

import (
	"strings"
	"testing"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKeyEvent_Navigation(t *testing.T) {
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
	section.SetFocused(true)

	// Test down navigation
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	cmd := section.HandleKeyEvent(msg)
	if cmd != nil {
		t.Error("Navigation should not return a command")
	}
	if section.GetSelected() != 1 {
		t.Errorf("Expected selection 1 after down, got %d", section.GetSelected())
	}

	// Test down again
	cmd = section.HandleKeyEvent(msg)
	if section.GetSelected() != 2 {
		t.Errorf("Expected selection 2 after down, got %d", section.GetSelected())
	}

	// Test down at end (should stay at 2)
	cmd = section.HandleKeyEvent(msg)
	if section.GetSelected() != 2 {
		t.Errorf("Expected selection to stay at 2, got %d", section.GetSelected())
	}

	// Test up navigation
	msgUp := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	cmd = section.HandleKeyEvent(msgUp)
	if section.GetSelected() != 1 {
		t.Errorf("Expected selection 1 after up, got %d", section.GetSelected())
	}

	// Test up again
	cmd = section.HandleKeyEvent(msgUp)
	if section.GetSelected() != 0 {
		t.Errorf("Expected selection 0 after up, got %d", section.GetSelected())
	}

	// Test up at start (should stay at 0)
	cmd = section.HandleKeyEvent(msgUp)
	if section.GetSelected() != 0 {
		t.Errorf("Expected selection to stay at 0, got %d", section.GetSelected())
	}
}

func TestHandleKeyEvent_ArrowKeys(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "file1.go"},
			{Path: "file2.go"},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)
	section.SetFocused(true)

	// Test down arrow
	msg := tea.KeyMsg{Type: tea.KeyDown}
	section.HandleKeyEvent(msg)
	if section.GetSelected() != 1 {
		t.Error("Down arrow should move selection down")
	}

	// Test up arrow
	msgUp := tea.KeyMsg{Type: tea.KeyUp}
	section.HandleKeyEvent(msgUp)
	if section.GetSelected() != 0 {
		t.Error("Up arrow should move selection up")
	}
}

func TestHandleKeyEvent_Unfocused(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "file1.go"},
			{Path: "file2.go"},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)
	section.SetFocused(false)

	// Navigation should not work when unfocused
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	cmd := section.HandleKeyEvent(msg)
	if cmd != nil {
		t.Error("Unfocused section should not process key events")
	}
	if section.GetSelected() != 0 {
		t.Error("Selection should not change when unfocused")
	}
}

func TestHandleKeyEvent_EmptyFileChanges(t *testing.T) {
	task := &taskmaster.Task{
		ID:          "1",
		FileChanges: []taskmaster.FileChange{},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)
	section.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	cmd := section.HandleKeyEvent(msg)
	if cmd != nil {
		t.Error("Empty file changes should not process navigation")
	}
}

func TestHandleKeyEvent_EnterKey(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "test.go", ChangeType: "modified"},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)
	section.SetFocused(true)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd := section.HandleKeyEvent(msg)
	if cmd == nil {
		t.Error("Enter key should return a command to open file")
	}
}

func TestHandleKeyEvent_AltD(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "test.go", ChangeType: "modified", IsPending: true},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)
	section.SetFocused(true)

	// Simulate Alt+D keypress
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}, Alt: true}
	cmd := section.HandleKeyEvent(msg)
	if cmd == nil {
		t.Error("Alt+D should return a command to view diff")
	}
}

func TestOpenFile_ValidFile(t *testing.T) {
	// Create a temporary file
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "go.mod", ChangeType: "modified"}, // Use a file that exists
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	cmd := section.OpenFile()
	if cmd == nil {
		t.Error("OpenFile should return a command")
	}
}

func TestOpenFile_InvalidIndex(t *testing.T) {
	task := &taskmaster.Task{
		ID:          "1",
		FileChanges: []taskmaster.FileChange{},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)
	section.SetSelected(5) // Out of bounds

	cmd := section.OpenFile()
	if cmd != nil {
		t.Error("OpenFile with invalid index should return nil")
	}
}

func TestOpenFile_NonexistentFile(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "nonexistent-file-12345.go", ChangeType: "modified"},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	cmd := section.OpenFile()
	if cmd == nil {
		t.Error("OpenFile should return a command even for nonexistent file")
	}

	// Execute the command to get the message
	msg := cmd()
	if opMsg, ok := msg.(fileOperationMsg); ok {
		if opMsg.success {
			t.Error("OpenFile for nonexistent file should return error message")
		}
		if !strings.Contains(opMsg.message, "not found") {
			t.Errorf("Expected 'not found' error, got: %s", opMsg.message)
		}
	} else {
		t.Error("Command should return fileOperationMsg")
	}
}

func TestViewDiff_PendingChanges(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "test.go", ChangeType: "modified", IsPending: true},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	cmd := section.ViewDiff()
	if cmd == nil {
		t.Error("ViewDiff should return a command")
	}

	// Execute the command to get the message
	msg := cmd()
	if diffMsg, ok := msg.(diffViewMsg); ok {
		// Check that we got a message (success or failure)
		if diffMsg.message == "" {
			t.Error("ViewDiff should return a message")
		}
	} else {
		t.Error("Command should return diffViewMsg")
	}
}

func TestViewDiff_CommittedChanges(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{
				Path:       "test.go",
				ChangeType: "modified",
				IsPending:  false,
				CommitID:   "abc123",
			},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)

	cmd := section.ViewDiff()
	if cmd == nil {
		t.Error("ViewDiff should return a command")
	}

	// Execute the command
	msg := cmd()
	if _, ok := msg.(diffViewMsg); !ok {
		t.Error("Command should return diffViewMsg")
	}
}

func TestViewDiff_InvalidIndex(t *testing.T) {
	task := &taskmaster.Task{
		ID:          "1",
		FileChanges: []taskmaster.FileChange{},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)
	section.SetSelected(5) // Out of bounds

	cmd := section.ViewDiff()
	if cmd != nil {
		t.Error("ViewDiff with invalid index should return nil")
	}
}

func TestRender_WithFocusAndSelection(t *testing.T) {
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

	// Render with focus and selection
	section.SetFocused(true)
	section.SetSelected(1)
	focusedResult := section.Render(80)

	// Results should contain file names
	if !strings.Contains(focusedResult, "file1.go") || !strings.Contains(focusedResult, "file2.go") {
		t.Error("Render should contain file names")
	}

	// Results should be different when focused (due to selection styling)
	// Note: This is a basic check - actual styling differences may not be visible in tests
	if len(unfocusedResult) == 0 || len(focusedResult) == 0 {
		t.Error("Render should produce output")
	}
}

func TestNavigationBoundaries(t *testing.T) {
	task := &taskmaster.Task{
		ID: "1",
		FileChanges: []taskmaster.FileChange{
			{Path: "file1.go"},
			{Path: "file2.go"},
			{Path: "file3.go"},
		},
	}
	styles := NewStyles()
	section := NewFileChangesSection(task, styles)
	section.SetFocused(true)

	// Navigate to bottom
	down := tea.KeyMsg{Type: tea.KeyDown}
	section.HandleKeyEvent(down)
	section.HandleKeyEvent(down)
	if section.GetSelected() != 2 {
		t.Error("Should be at index 2")
	}

	// Try to go past bottom
	section.HandleKeyEvent(down)
	section.HandleKeyEvent(down)
	if section.GetSelected() != 2 {
		t.Error("Should stay at index 2 (bottom boundary)")
	}

	// Navigate back to top
	up := tea.KeyMsg{Type: tea.KeyUp}
	section.HandleKeyEvent(up)
	section.HandleKeyEvent(up)
	if section.GetSelected() != 0 {
		t.Error("Should be at index 0")
	}

	// Try to go past top
	section.HandleKeyEvent(up)
	section.HandleKeyEvent(up)
	if section.GetSelected() != 0 {
		t.Error("Should stay at index 0 (top boundary)")
	}
}
