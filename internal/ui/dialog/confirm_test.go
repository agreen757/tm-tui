package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestConfirmationDialogUpdate tests that the Update method properly handles KeyMsg
func TestConfirmationDialogUpdate(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Are you sure?", 40, 10)

	// Test that WindowSizeMsg is handled
	windowMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	result, _ := dialog.Update(windowMsg)
	if result == nil {
		t.Error("Expected dialog to remain open on WindowSizeMsg")
	}

	// Test that KeyMsg is handled by Update (should call HandleKey)
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = dialog.Update(enterKey)
	if result != nil {
		t.Error("Expected dialog to close on Enter key in Update")
	}
}

// TestConfirmationDialogCtrlCHandling tests that Ctrl+C closes the dialog with No result
func TestConfirmationDialogCtrlCHandling(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Continue?", 40, 10)

	// Press Ctrl+C
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	result, cmd := dialog.HandleKey(ctrlCKey)

	// Should return Cancel result
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel on Ctrl+C, got %v", result)
	}

	// Should have a command
	if cmd == nil {
		t.Error("Expected Cmd to be returned")
	}

	// Result should be "No"
	if dialog.Result() != ConfirmationResultNo {
		t.Errorf("Expected result to be No, got %v", dialog.Result())
	}
}

// TestConfirmationDialogEscHandling tests that Esc closes the dialog
func TestConfirmationDialogEscHandling(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Continue?", 40, 10)

	// Press Esc (handled by HandleBaseFocusableKey -> HandleBaseKey)
	escKey := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ := dialog.HandleKey(escKey)

	// Should return Cancel result
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel on Esc, got %v", result)
	}
}

// TestConfirmationDialogEnterConfirms tests that Enter confirms the selected button
func TestConfirmationDialogEnterConfirms(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Continue?", 40, 10)

	// Initially focused on Yes
	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected initial focus on Yes (0), got %d", dialog.FocusedIndex())
	}

	// Press Enter on Yes
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := dialog.HandleKey(enterKey)

	// Should return Confirm result
	if result != DialogResultConfirm {
		t.Errorf("Expected DialogResultConfirm on Enter while on Yes, got %v", result)
	}

	// Result should be Yes
	if dialog.Result() != ConfirmationResultYes {
		t.Errorf("Expected result to be Yes, got %v", dialog.Result())
	}
}

// TestConfirmationDialogYShortcut tests that 'y' confirms Yes
func TestConfirmationDialogYShortcut(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Continue?", 40, 10)

	// Press 'y'
	yKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	result, _ := dialog.HandleKey(yKey)

	if result != DialogResultConfirm {
		t.Errorf("Expected DialogResultConfirm on 'y', got %v", result)
	}

	if dialog.Result() != ConfirmationResultYes {
		t.Errorf("Expected result to be Yes, got %v", dialog.Result())
	}
}

// TestConfirmationDialogNShortcut tests that 'n' confirms No
func TestConfirmationDialogNShortcut(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Continue?", 40, 10)

	// Press 'n'
	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	result, _ := dialog.HandleKey(nKey)

	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel on 'n', got %v", result)
	}

	if dialog.Result() != ConfirmationResultNo {
		t.Errorf("Expected result to be No, got %v", dialog.Result())
	}
}

// TestConfirmationDialogNavigation tests left/right arrow navigation
func TestConfirmationDialogNavigation(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Continue?", 40, 10)

	// Initially on Yes (index 0)
	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected initial focus on Yes (0), got %d", dialog.FocusedIndex())
	}

	// Press Right to move to No
	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	result, _ := dialog.HandleKey(rightKey)

	if result != DialogResultNone {
		t.Error("Expected DialogResultNone on navigation")
	}

	if dialog.FocusedIndex() != 1 {
		t.Errorf("Expected focus on No (1), got %d", dialog.FocusedIndex())
	}

	// Press Left to move back to Yes
	leftKey := tea.KeyMsg{Type: tea.KeyLeft}
	result, _ = dialog.HandleKey(leftKey)

	if result != DialogResultNone {
		t.Error("Expected DialogResultNone on navigation")
	}

	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected focus back on Yes (0), got %d", dialog.FocusedIndex())
	}
}

// TestConfirmationDialogUpdateWithKeyMsg tests that Update properly processes KeyMsg
func TestConfirmationDialogUpdateWithKeyMsg(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Confirm action?", 40, 10)

	// Test Update with Enter key
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updatedDialog, cmd := dialog.Update(enterKey)

	// Dialog should close (nil)
	if updatedDialog != nil {
		t.Error("Expected dialog to close on Enter in Update")
	}

	// Cmd should be returned
	if cmd == nil {
		t.Error("Expected Cmd to be returned from Update")
	}

	// Result should be Yes
	if dialog.Result() != ConfirmationResultYes {
		t.Errorf("Expected result to be Yes, got %v", dialog.Result())
	}
}

// TestConfirmationDialogUpdateWithCtrlC tests that Update handles Ctrl+C
func TestConfirmationDialogUpdateWithCtrlC(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Confirm action?", 40, 10)

	// Test Update with Ctrl+C
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedDialog, _ := dialog.Update(ctrlCKey)

	// Dialog should close (nil)
	if updatedDialog != nil {
		t.Error("Expected dialog to close on Ctrl+C in Update")
	}

	// Result should be No
	if dialog.Result() != ConfirmationResultNo {
		t.Errorf("Expected result to be No when Ctrl+C is pressed, got %v", dialog.Result())
	}
}

// TestConfirmationDialogMultipleUpdates tests that dialog state remains consistent across multiple updates
func TestConfirmationDialogMultipleUpdates(t *testing.T) {
	dialog := NewConfirmationDialog("Test", "Continue?", 40, 10)

	// First update with navigation (should keep dialog open)
	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	result1, _ := dialog.Update(rightKey)

	if result1 == nil {
		t.Error("Expected dialog to remain open on navigation")
	}

	if dialog.FocusedIndex() != 1 {
		t.Errorf("Expected focus on No (1), got %d", dialog.FocusedIndex())
	}

	// Second update with Ctrl+C (should close dialog)
	ctrlCKey := tea.KeyMsg{Type: tea.KeyCtrlC}
	result2, _ := dialog.Update(ctrlCKey)

	if result2 != nil {
		t.Error("Expected dialog to close on Ctrl+C")
	}

	// Result should be No (Ctrl+C was pressed)
	if dialog.Result() != ConfirmationResultNo {
		t.Errorf("Expected result to be No, got %v", dialog.Result())
	}
}

// TestYesNoDialogCreation tests that YesNo helper creates properly configured dialogs
func TestYesNoDialogCreation(t *testing.T) {
	dialog := YesNo("Danger Zone", "Are you absolutely sure?", true)

	if dialog == nil {
		t.Fatal("Expected YesNo to create a dialog")
	}

	if dialog.Title() != "Danger Zone" {
		t.Errorf("Expected title 'Danger Zone', got %q", dialog.Title())
	}

	if !dialog.dangerMode {
		t.Error("Expected danger mode to be enabled")
	}
}
