package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestFocusRestorationOnDialogClose verifies that focus returns to the previous element
// when a dialog closes
func TestFocusRestorationOnDialogClose(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Create first dialog with multiple focusable elements
	firstDialog := NewListDialog("First", 60, 30, []ListItem{
		NewSimpleListItem("Item 1", ""),
		NewSimpleListItem("Item 2", ""),
		NewSimpleListItem("Item 3", ""),
	})

	// Set focus to item 2
	firstDialog.SetSelectedIndex(1)

	// Add the first dialog
	manager.PushDialog(firstDialog)

	if manager.GetActiveDialog() != firstDialog {
		t.Error("Expected first dialog to be active")
	}

	if firstDialog.SelectedIndex() != 1 {
		t.Errorf("Expected initial focus at index 1, got %d", firstDialog.SelectedIndex())
	}

	// Create a second dialog (modal overlay)
	secondDialog := NewModalDialog("Modal", 50, 20, NewSimpleModalContent("Overlay content"))

	// Add the second dialog (should save the focus state of the first)
	manager.AddDialog(secondDialog, nil)

	if manager.GetActiveDialog() != secondDialog {
		t.Error("Expected second dialog to be active")
	}

	// Close the second dialog
	manager.PopDialog()

	// The first dialog should be active again
	if manager.GetActiveDialog() != firstDialog {
		t.Error("Expected first dialog to be active again after closing second")
	}

	// Focus should be restored to item 2
	if firstDialog.SelectedIndex() != 1 {
		t.Errorf("Expected focus to be restored to index 1, got %d", firstDialog.SelectedIndex())
	}
}

// TestNestedDialogFocusStackManagement tests that focus is properly managed
// in deeply nested dialog stacks
func TestNestedDialogFocusStackManagement(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Create multiple dialogs with different focus states
	dialogs := make([]*ListDialog, 4)
	expectedIndices := []int{0, 2, 1, 0}

	for i := 0; i < 4; i++ {
		dialogs[i] = NewListDialog("Dialog"+string(rune('1'+i)), 60, 30, []ListItem{
			NewSimpleListItem("A", ""),
			NewSimpleListItem("B", ""),
			NewSimpleListItem("C", ""),
		})

		dialogs[i].SetSelectedIndex(expectedIndices[i])

		manager.PushDialog(dialogs[i])
	}

	// Pop dialogs in reverse order and verify focus restoration
	for i := 3; i >= 0; i-- {
		currentDialog := manager.GetActiveDialog()
		if currentDialog != dialogs[i] {
			t.Errorf("Dialog %d: Expected dialog %d to be active, got %d", i, i, i)
		}

		if dialogs[i].SelectedIndex() != expectedIndices[i] {
			t.Errorf("Dialog %d: Expected focus index %d, got %d",
				i, expectedIndices[i], dialogs[i].SelectedIndex())
		}

		if i > 0 {
			manager.PopDialog()
		}
	}
}

// TestFocusTrappingPrevention verifies that focus cannot escape the active dialog
func TestFocusTrappingPrevention(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Create background dialog
	backgroundDialog := NewListDialog("Background", 80, 40, []ListItem{
		NewSimpleListItem("Item A", ""),
		NewSimpleListItem("Item B", ""),
	})
	backgroundDialog.SetSelectedIndex(1)

	// Create foreground dialog
	foregroundDialog := NewListDialog("Foreground", 60, 30, []ListItem{
		NewSimpleListItem("Item 1", ""),
		NewSimpleListItem("Item 2", ""),
		NewSimpleListItem("Item 3", ""),
	})

	manager.PushDialog(backgroundDialog)
	manager.PushDialog(foregroundDialog)

	// Initial state
	if manager.GetActiveDialog() != foregroundDialog {
		t.Error("Expected foreground dialog to be active")
	}

	// Try to navigate in the background dialog
	downKey := tea.KeyMsg{Type: tea.KeyDown}
	manager.HandleMsg(downKey)

	// Foreground dialog should have received the key, not background
	if backgroundDialog.SelectedIndex() != 1 {
		t.Error("Background dialog should not have received focus change")
	}

	// Only the foreground dialog should be affected
	if foregroundDialog.SelectedIndex() != 1 {
		t.Errorf("Foreground dialog should handle key, got index %d", foregroundDialog.SelectedIndex())
	}
}

// TestFocusIndicatorStyles verifies that focus indicators render correctly
func TestFocusIndicatorStyles(t *testing.T) {
	defaultStyle := DefaultFocusIndicatorStyle()
	if defaultStyle.CursorCharacter == "" {
		t.Error("Default focus indicator style should have a cursor character")
	}

	highContrastStyle := HighContrastFocusIndicatorStyle()
	if highContrastStyle.UnderlineActive == false {
		t.Error("High contrast style should have underline enabled")
	}

	if highContrastStyle.BoldActive == false {
		t.Error("High contrast style should have bold enabled")
	}
}

// TestRenderFocusedElement verifies element rendering with focus indicators
func TestRenderFocusedElement(t *testing.T) {
	style := DefaultFocusIndicatorStyle()

	// Test focused element
	focusedContent := RenderFocusedElement("Test Item", true, style, 20)
	if focusedContent == "" {
		t.Error("Focused element should render non-empty content")
	}

	// Test unfocused element
	unfocusedContent := RenderFocusedElement("Test Item", false, style, 20)
	if unfocusedContent == "" {
		t.Error("Unfocused element should render non-empty content")
	}

	// The focused version should be different from unfocused
	if focusedContent == unfocusedContent {
		t.Error("Focused and unfocused elements should render differently")
	}
}

// TestRenderFocusIndicator verifies focus indicator rendering
func TestRenderFocusIndicator(t *testing.T) {
	style := DefaultFocusIndicatorStyle()

	// Test focused indicator
	focusedIndicator := RenderFocusIndicator(true, style)
	if focusedIndicator == " " {
		t.Error("Focused indicator should not be empty space")
	}

	// Test unfocused indicator
	unfocusedIndicator := RenderFocusIndicator(false, style)
	if unfocusedIndicator != " " {
		t.Error("Unfocused indicator should be empty space")
	}
}

// TestFocusTrappingValidator verifies focus trapping validation logic
func TestFocusTrappingValidator(t *testing.T) {
	validator := NewFocusTrappingValidator()

	// Record a sequence of focus changes
	validator.RecordFocus(0)
	validator.RecordFocus(1)
	validator.RecordFocus(2)
	validator.RecordFocus(0) // Cycle back

	// Should detect the focus cycle
	if !validator.IsFocusTrapped() {
		t.Error("Validator should detect focus trap (cycling)")
	}

	// Check history
	history := validator.GetFocusHistory()
	expectedLen := 4
	if len(history) != expectedLen {
		t.Errorf("Expected history length %d, got %d", expectedLen, len(history))
	}

	// Test reset
	validator.Reset()
	if len(validator.GetFocusHistory()) != 0 {
		t.Error("History should be empty after reset")
	}
	if validator.IsFocusTrapped() {
		t.Error("Should not be trapped after reset")
	}
}

// TestDialogFocusStatePersistence verifies that dialog focus state is preserved
// across multiple add/pop cycles
func TestDialogFocusStatePersistence(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Create a dialog
	dialog1 := NewListDialog("Dialog1", 60, 30, []ListItem{
		NewSimpleListItem("Item 1", ""),
		NewSimpleListItem("Item 2", ""),
		NewSimpleListItem("Item 3", ""),
		NewSimpleListItem("Item 4", ""),
	})

	manager.PushDialog(dialog1)
	dialog1.SetSelectedIndex(2)

	if dialog1.SelectedIndex() != 2 {
		t.Errorf("Expected initial focus index 2, got %d", dialog1.SelectedIndex())
	}

	// Add an overlay dialog
	overlay := NewModalDialog("Overlay", 50, 20, NewSimpleModalContent("Content"))
	manager.PushDialog(overlay)

	// Remove the overlay
	manager.PopDialog()

	// Dialog1 should still have focus at index 2
	if dialog1.SelectedIndex() != 2 {
		t.Errorf("Expected focus to be restored to index 2, got %d", dialog1.SelectedIndex())
	}

	// Change focus on dialog1
	dialog1.SetSelectedIndex(0)

	// Add another overlay
	overlay2 := NewModalDialog("Overlay2", 50, 20, NewSimpleModalContent("Content2"))
	manager.PushDialog(overlay2)

	// Remove the overlay
	manager.PopDialog()

	// Dialog1 should now have focus at index 0 (the new position)
	if dialog1.SelectedIndex() != 0 {
		t.Errorf("Expected focus at new index 0, got %d", dialog1.SelectedIndex())
	}
}

// TestComplexNestedDialogScenario tests a realistic nested dialog scenario
// like a form within a confirmation dialog within a menu
func TestComplexNestedDialogScenario(t *testing.T) {
	manager := NewDialogManager(120, 60)

	// Start with a menu dialog
	menu := NewListDialog("Main Menu", 80, 40, []ListItem{
		NewSimpleListItem("Option 1", ""),
		NewSimpleListItem("Option 2", ""),
		NewSimpleListItem("Option 3", ""),
	})
	manager.PushDialog(menu)
	menu.SetSelectedIndex(0)

	// User selects option 2, which opens a confirmation dialog
	confirmation := YesNo("Confirm Action", "Are you sure?", false)
	manager.PushDialog(confirmation)

	if manager.GetActiveDialog() != confirmation {
		t.Error("Confirmation dialog should be active")
	}

	// User closes confirmation
	manager.PopDialog()

	// Menu should be active again with focus restored
	if manager.GetActiveDialog() != menu {
		t.Error("Menu should be active again")
	}

	if menu.SelectedIndex() != 0 {
		t.Errorf("Menu focus should be restored to 0, got %d", menu.SelectedIndex())
	}

	// Now suppose user moves selection in menu
	menu.SetSelectedIndex(2)

	// And opens another dialog
	form := NewLegacyFormDialog("Settings", 70, 35, []FormField{
		NewTextField("Name", "", false),
		NewTextField("Email", "", false),
	})
	manager.PushDialog(form)

	if manager.GetActiveDialog() != form {
		t.Error("Form should be active")
	}

	// Close form
	manager.PopDialog()

	// Menu focus should be restored to index 2
	if menu.SelectedIndex() != 2 {
		t.Errorf("Menu focus should be restored to 2, got %d", menu.SelectedIndex())
	}
}
