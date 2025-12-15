package dialog

import (
	"testing"

	"github.com/agreen757/tm-tui/internal/taskmaster"
	tea "github.com/charmbracelet/bubbletea"
)

func TestExpandTaskPreviewDialogBasics(t *testing.T) {
	drafts := []taskmaster.SubtaskDraft{
		{
			Title:       "First step",
			Description: "Do the first thing",
		},
		{
			Title:       "Second step",
			Description: "Do the second thing",
		},
	}

	dialog := NewExpandTaskPreviewDialog(drafts, DefaultDialogStyle())

	if dialog == nil {
		t.Error("Expected dialog to be created, got nil")
	}

	if dialog.Title() != "Preview Expanded Tasks" {
		t.Errorf("Expected title 'Preview Expanded Tasks', got '%s'", dialog.Title())
	}

	if len(dialog.GetSelectedDrafts()) != 2 {
		t.Errorf("Expected 2 drafts, got %d", len(dialog.GetSelectedDrafts()))
	}
}

func TestExpandTaskPreviewDialogNavigation(t *testing.T) {
	drafts := []taskmaster.SubtaskDraft{
		{Title: "Task 1"},
		{Title: "Task 2"},
		{Title: "Task 3"},
	}

	dialog := NewExpandTaskPreviewDialog(drafts, DefaultDialogStyle())

	// Test initial state
	if dialog.selectedIndex != 0 {
		t.Errorf("Expected initial index 0, got %d", dialog.selectedIndex)
	}

	// Simulate key press: down
	d, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyDown})
	if preview, ok := d.(*ExpandTaskPreviewDialog); ok {
		if preview.selectedIndex != 1 {
			t.Errorf("Expected index 1 after down key, got %d", preview.selectedIndex)
		}
	}
}

func TestSubtaskEditDialogBasics(t *testing.T) {
	drafts := []taskmaster.SubtaskDraft{
		{Title: "Edit me", Description: "I can be edited"},
	}

	dialog := NewSubtaskEditDialog(drafts, DefaultDialogStyle())

	if dialog == nil {
		t.Error("Expected dialog to be created, got nil")
	}

	if dialog.Title() != "Edit Subtasks" {
		t.Errorf("Expected title 'Edit Subtasks', got '%s'", dialog.Title())
	}

	if len(dialog.GetDrafts()) != 1 {
		t.Errorf("Expected 1 draft, got %d", len(dialog.GetDrafts()))
	}
}

func TestSubtaskEditDialogAddSubtask(t *testing.T) {
	drafts := []taskmaster.SubtaskDraft{
		{Title: "First task"},
	}

	dialog := NewSubtaskEditDialog(drafts, DefaultDialogStyle())

	// Simulate adding a subtask - 'a' key should add
	d, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if editDialog, ok := d.(*SubtaskEditDialog); ok {
		if len(editDialog.GetDrafts()) != 2 {
			t.Errorf("Expected 2 drafts after add, got %d", len(editDialog.GetDrafts()))
		}
	}
}

func TestSubtaskEditDialogDeleteSubtask(t *testing.T) {
	drafts := []taskmaster.SubtaskDraft{
		{Title: "Task 1"},
		{Title: "Task 2"},
	}

	dialog := NewSubtaskEditDialog(drafts, DefaultDialogStyle())

	// Simulate deleting a subtask - 'd' key should delete
	d, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if editDialog, ok := d.(*SubtaskEditDialog); ok {
		if len(editDialog.GetDrafts()) != 1 {
			t.Errorf("Expected 1 draft after delete, got %d", len(editDialog.GetDrafts()))
		}
	}
}

func TestApplySubtaskDrafts(t *testing.T) {
	// Create a parent task
	parentTask := &taskmaster.Task{
		ID:    "1",
		Title: "Parent Task",
	}

	// Create drafts
	drafts := []taskmaster.SubtaskDraft{
		{
			Title:       "Subtask 1",
			Description: "First subtask",
			Children: []taskmaster.SubtaskDraft{
				{
					Title: "Child 1.1",
				},
			},
		},
		{
			Title:       "Subtask 2",
			Description: "Second subtask",
		},
	}

	// Apply drafts
	newIDs, err := taskmaster.ApplySubtaskDrafts(parentTask, drafts)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify results
	if len(newIDs) != 2 {
		t.Errorf("Expected 2 new IDs, got %d", len(newIDs))
	}

	if len(parentTask.Subtasks) != 2 {
		t.Errorf("Expected 2 subtasks on parent, got %d", len(parentTask.Subtasks))
	}

	// Check first subtask
	if parentTask.Subtasks[0].Title != "Subtask 1" {
		t.Errorf("Expected title 'Subtask 1', got '%s'", parentTask.Subtasks[0].Title)
	}

	// Check that first subtask has a child
	if len(parentTask.Subtasks[0].Subtasks) != 1 {
		t.Errorf("Expected 1 child for first subtask, got %d", len(parentTask.Subtasks[0].Subtasks))
	}

	// Check generated IDs
	if newIDs[0] != "1.1" {
		t.Errorf("Expected first ID '1.1', got '%s'", newIDs[0])
	}
	if newIDs[1] != "1.2" {
		t.Errorf("Expected second ID '1.2', got '%s'", newIDs[1])
	}
}

func TestApplySubtaskDraftsNilParent(t *testing.T) {
	drafts := []taskmaster.SubtaskDraft{
		{Title: "Test"},
	}

	newIDs, err := taskmaster.ApplySubtaskDrafts(nil, drafts)
	if err == nil {
		t.Error("Expected error for nil parent task")
	}

	if newIDs != nil {
		t.Errorf("Expected nil newIDs on error, got %v", newIDs)
	}
}
