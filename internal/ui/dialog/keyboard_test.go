package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestKeyboardNavigation tests keyboard navigation within focusable dialogs
func TestKeyboardNavigation(t *testing.T) {
	// Create a form with multiple fields to test keyboard navigation
	fields := []FormField{
		NewTextField("Name", "Enter name", false),
		NewCheckboxField("Agree", false),
		NewRadioGroupField("Option", []string{"Option 1", "Option 2", "Option 3"}, 0),
	}
	formDialog := NewLegacyFormDialog("Form Navigation Test", 60, 30, fields)

	// Initial state
	if formDialog.FocusedIndex() != 0 {
		t.Errorf("Expected initial focus index 0, got %d", formDialog.FocusedIndex())
	}

	// Tab key should move focus to next field
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := formDialog.HandleKey(tabKey)
	if result != DialogResultNone || formDialog.FocusedIndex() != 1 {
		t.Errorf("Expected DialogResultNone and focus index 1 after Tab, got %v and %d",
			result, formDialog.FocusedIndex())
	}

	// Another Tab moves to radio group
	result, _ = formDialog.HandleKey(tabKey)
	if result != DialogResultNone || formDialog.FocusedIndex() != 2 {
		t.Errorf("Expected DialogResultNone and focus index 2 after second Tab, got %v and %d",
			result, formDialog.FocusedIndex())
	}

	// Another Tab moves to Submit button
	result, _ = formDialog.HandleKey(tabKey)
	if result != DialogResultNone || formDialog.FocusedIndex() != 3 {
		t.Errorf("Expected DialogResultNone and focus index 3 after third Tab, got %v and %d",
			result, formDialog.FocusedIndex())
	}

	// Another Tab moves to Cancel button
	result, _ = formDialog.HandleKey(tabKey)
	if result != DialogResultNone || formDialog.FocusedIndex() != 4 {
		t.Errorf("Expected DialogResultNone and focus index 4 after fourth Tab, got %v and %d",
			result, formDialog.FocusedIndex())
	}

	// Another Tab wraps around to first field
	result, _ = formDialog.HandleKey(tabKey)
	if result != DialogResultNone || formDialog.FocusedIndex() != 0 {
		t.Errorf("Expected DialogResultNone and focus index 0 after wrapping, got %v and %d",
			result, formDialog.FocusedIndex())
	}

	// Test Shift+Tab for reverse navigation
	shiftTabKey := tea.KeyMsg{Type: tea.KeyShiftTab}
	result, _ = formDialog.HandleKey(shiftTabKey)
	if result != DialogResultNone || formDialog.FocusedIndex() != 4 {
		t.Errorf("Expected DialogResultNone and focus index 4 after Shift+Tab, got %v and %d",
			result, formDialog.FocusedIndex())
	}
}

// TestRadioGroupNavigation tests navigation within radio group fields
func TestRadioGroupNavigation(t *testing.T) {
	options := []string{"Option 1", "Option 2", "Option 3"}
	field := NewRadioGroupField("Test", options, 0)
	fields := []FormField{field}
	formDialog := NewLegacyFormDialog("Radio Navigation Test", 60, 30, fields)
	radioField, ok := formDialog.GetField(field.ID)
	if !ok {
		t.Fatalf("expected radio field with ID %s", field.ID)
	}

	// Initial state
	if radioField.SelectedOption != 0 {
		t.Errorf("Expected initial selected option 0, got %d", radioField.SelectedOption)
	}

	// Down key should move to next option
	downKey := tea.KeyMsg{Type: tea.KeyDown}
	result, _ := formDialog.HandleKey(downKey)
	if result != DialogResultNone || radioField.SelectedOption != 1 {
		t.Errorf("Expected DialogResultNone and selected option 1 after Down key, got %v and %d",
			result, radioField.SelectedOption)
	}

	// Down key again moves to last option
	result, _ = formDialog.HandleKey(downKey)
	if result != DialogResultNone || radioField.SelectedOption != 2 {
		t.Errorf("Expected DialogResultNone and selected option 2 after second Down key, got %v and %d",
			result, radioField.SelectedOption)
	}

	// Down key wraps to first option
	result, _ = formDialog.HandleKey(downKey)
	if result != DialogResultNone || radioField.SelectedOption != 0 {
		t.Errorf("Expected DialogResultNone and selected option 0 after wrapping, got %v and %d",
			result, radioField.SelectedOption)
	}

	// Up key should wrap to last option
	upKey := tea.KeyMsg{Type: tea.KeyUp}
	result, _ = formDialog.HandleKey(upKey)
	if result != DialogResultNone || radioField.SelectedOption != 2 {
		t.Errorf("Expected DialogResultNone and selected option 2 after Up key from first item, got %v and %d",
			result, radioField.SelectedOption)
	}
}

// TestFocusTrapping tests that focus is correctly trapped within dialog stacks
func TestFocusTrapping(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Create a background item (which we will simulate with another dialog)
	backgroundDialog := NewModalDialog("Background", 80, 40, NewSimpleModalContent("Background Content"))

	// Create a foreground dialog
	foregroundDialog := NewListDialog("Foreground", 60, 30, []ListItem{
		NewSimpleListItem("Item 1", "Description 1"),
		NewSimpleListItem("Item 2", "Description 2"),
	})

	// Add both dialogs to the stack
	manager.PushDialog(backgroundDialog)
	manager.PushDialog(foregroundDialog)

	// The foreground dialog should be on top and focused
	if manager.GetActiveDialog() != foregroundDialog {
		t.Error("Expected foreground dialog to be active")
	}

	// Send keys that would affect the background dialog if not trapped
	keySequence := []tea.KeyMsg{
		{Type: tea.KeyDown},
		{Type: tea.KeyDown},
		{Type: tea.KeyUp},
		{Type: tea.KeyEnter},
	}

	// Process each key
	for _, keyMsg := range keySequence {
		manager.HandleMsg(keyMsg)
	}

	// The Enter key should close the foreground dialog
	if manager.GetActiveDialog() == foregroundDialog {
		t.Error("Expected foreground dialog to be closed after Enter key")
	}

	// The background dialog should now be active
	if manager.GetActiveDialog() != backgroundDialog {
		t.Error("Expected background dialog to be active after foreground is closed")
	}

	// Now the background dialog should get Enter keypresses
	manager.HandleMsg(tea.KeyMsg{Type: tea.KeyEnter})

	// Background dialog should be closed
	if manager.HasDialogs() {
		t.Error("Expected all dialogs to be closed after background gets Enter key")
	}
}

// TestCheckboxToggle tests toggling checkboxes in forms
func TestCheckboxToggle(t *testing.T) {
	field := NewCheckboxField("Test", false)
	fields := []FormField{field}
	formDialog := NewLegacyFormDialog("Checkbox Test", 60, 20, fields)
	checkboxField, ok := formDialog.GetField(field.ID)
	if !ok {
		t.Fatalf("expected checkbox field with ID %s", field.ID)
	}

	// Initial state
	if checkboxField.Checked {
		t.Error("Expected checkbox to be unchecked initially")
	}

	// Space key toggles checkbox
	spaceKey := tea.KeyMsg{Type: tea.KeySpace}
	result, _ := formDialog.HandleKey(spaceKey)
	if result != DialogResultNone || !checkboxField.Checked {
		t.Errorf("Expected DialogResultNone and checked state after Space key, got %v and %v",
			result, checkboxField.Checked)
	}

	// Space key toggles again
	result, _ = formDialog.HandleKey(spaceKey)
	if result != DialogResultNone || checkboxField.Checked {
		t.Errorf("Expected DialogResultNone and unchecked state after second Space key, got %v and %v",
			result, checkboxField.Checked)
	}

	// Enter should also work
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ = formDialog.HandleKey(enterKey)
	if result != DialogResultNone || !checkboxField.Checked {
		t.Errorf("Expected DialogResultNone and checked state after Enter key, got %v and %v",
			result, checkboxField.Checked)
	}
}

// TestDialogKeyboardAccessibility tests keyboard accessibility features
// like shortcuts and focus indicators
func TestDialogKeyboardAccessibility(t *testing.T) {
	// Test confirmation dialog keyboard shortcuts
	confirmDialog := YesNo("Test", "Message", false)

	// 'y' key should select Yes
	yKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	result, _ := confirmDialog.HandleKey(yKey)
	if result != DialogResultConfirm || confirmDialog.Result() != ConfirmationResultYes {
		t.Errorf("Expected DialogResultConfirm and ConfirmationResultYes after 'y' key, got %v and %v",
			result, confirmDialog.Result())
	}

	// Reset and test 'n' key
	confirmDialog = YesNo("Test", "Message", false)
	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	result, _ = confirmDialog.HandleKey(nKey)
	if result != DialogResultCancel || confirmDialog.Result() != ConfirmationResultNo {
		t.Errorf("Expected DialogResultCancel and ConfirmationResultNo after 'n' key, got %v and %v",
			result, confirmDialog.Result())
	}

	// Test list dialog keyboard accessibility
	listDialog := NewListDialog("Test", 50, 20, []ListItem{
		NewSimpleListItem("Item 1", ""),
		NewSimpleListItem("Item 2", ""),
		NewSimpleListItem("Item 3", ""),
	})

	// Test 'g' (home) key
	listDialog.SetSelectedIndex(1)
	gKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	result, _ = listDialog.HandleKey(gKey)
	if result != DialogResultNone || listDialog.SelectedIndex() != 0 {
		t.Errorf("Expected DialogResultNone and selected index 0 after 'g' key, got %v and %d",
			result, listDialog.SelectedIndex())
	}

	// Test 'G' (end) key
	GKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	result, _ = listDialog.HandleKey(GKey)
	if result != DialogResultNone || listDialog.SelectedIndex() != 2 {
		t.Errorf("Expected DialogResultNone and selected index 2 after 'G' key, got %v and %d",
			result, listDialog.SelectedIndex())
	}
}

// TestKeyboardShortcutConsistency tests that keyboard shortcuts are consistent
// across dialog types where applicable
func TestKeyboardShortcutConsistency(t *testing.T) {
	// Create various dialog types
	dialogs := []Dialog{
		NewModalDialog("Modal", 50, 10, NewSimpleModalContent("Content")),
		NewProgressDialog("Progress", 50, 10),
		YesNo("Confirm", "Message", false),
		NewListDialog("List", 50, 10, []ListItem{
			NewSimpleListItem("Item", ""),
		}),
		NewLegacyFormDialog("Form", 50, 20, []FormField{
			NewTextField("Field", "Value", false),
		}),
	}

	// Test Escape key consistency - should cancel/close all dialog types
	escKey := tea.KeyMsg{Type: tea.KeyEsc}

	for _, dialog := range dialogs {
		result, _ := dialog.HandleKey(escKey)
		if result != DialogResultCancel {
			t.Errorf("Dialog %s: Expected DialogResultCancel for Escape key, got %v",
				dialog.Title(), result)
		}
	}
}

// TestStandardFooterHints tests that footer hints are consistently provided
func TestStandardFooterHints(t *testing.T) {
	// Test form footer hints
	formHints := StandardFooterHints("form")
	if len(formHints) == 0 {
		t.Error("Expected form footer hints to be non-empty")
	}
	
	// Check for Tab hint in form
	hasTab := false
	for _, hint := range formHints {
		if hint.Key == "Tab" {
			hasTab = true
			break
		}
	}
	if !hasTab {
		t.Error("Expected form footer hints to include Tab")
	}

	// Test list footer hints
	listHints := StandardFooterHints("list")
	if len(listHints) == 0 {
		t.Error("Expected list footer hints to be non-empty")
	}

	// Test confirmation footer hints
	confirmHints := StandardFooterHints("confirmation")
	if len(confirmHints) == 0 {
		t.Error("Expected confirmation footer hints to be non-empty")
	}

	// Test file footer hints
	fileHints := StandardFooterHints("file")
	if len(fileHints) == 0 {
		t.Error("Expected file footer hints to be non-empty")
	}

	// All dialogs should have at least Enter and Esc hints
	dialogTypes := []string{"form", "list", "confirmation", "file"}
	for _, dtype := range dialogTypes {
		hints := StandardFooterHints(dtype)
		hasEnter := false
		hasEsc := false

		for _, hint := range hints {
			if hint.Key == "Enter" {
				hasEnter = true
			}
			if hint.Key == "Esc" {
				hasEsc = true
			}
		}

		if !hasEnter {
			t.Errorf("Dialog type %s: Missing Enter hint", dtype)
		}
		if !hasEsc {
			t.Errorf("Dialog type %s: Missing Esc hint", dtype)
		}
	}
}

// TestKeyboardNavigationHelperTab tests the keyboard navigation helper Tab handling
func TestKeyboardNavigationHelperTab(t *testing.T) {
	helper := NewKeyboardNavigationHelper(5)

	// Test Tab moves focus forward
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	if !helper.HandleTabKey(tabKey) {
		t.Error("Expected HandleTabKey to return true for Tab key")
	}
	if helper.FocusedIndex() != 1 {
		t.Errorf("Expected focused index 1 after Tab, got %d", helper.FocusedIndex())
	}

	// Test Shift+Tab moves focus backward
	shiftTabKey := tea.KeyMsg{Type: tea.KeyShiftTab}
	if !helper.HandleTabKey(shiftTabKey) {
		t.Error("Expected HandleTabKey to return true for Shift+Tab key")
	}
	if helper.FocusedIndex() != 0 {
		t.Errorf("Expected focused index 0 after Shift+Tab, got %d", helper.FocusedIndex())
	}

	// Test wrapping forward
	helper.SetFocusedIndex(4)
	helper.HandleTabKey(tabKey)
	if helper.FocusedIndex() != 0 {
		t.Errorf("Expected focus to wrap to 0, got %d", helper.FocusedIndex())
	}

	// Test wrapping backward
	helper.SetFocusedIndex(0)
	helper.HandleTabKey(shiftTabKey)
	if helper.FocusedIndex() != 4 {
		t.Errorf("Expected focus to wrap to 4, got %d", helper.FocusedIndex())
	}
}
