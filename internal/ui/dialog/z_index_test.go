package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestZIndexManagement tests comprehensive z-index behavior in dialog stacking
func TestZIndexManagement(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Create several dialogs of different types
	modalDialog := NewModalDialog("Modal Dialog", 60, 30, NewSimpleModalContent("Modal Content"))
	formDialog := NewLegacyFormDialog("Form Dialog", 70, 40, []FormField{
		NewTextField("Field", "Value", false),
	})
	listDialog := NewListDialog("List Dialog", 50, 25, []ListItem{
		NewSimpleListItem("Item 1", "Description 1"),
		NewSimpleListItem("Item 2", "Description 2"),
	})
	confirmDialog := YesNo("Confirm Dialog", "Are you sure?", false)
	_ = NewProgressDialog("Progress Dialog", 60, 15) // Create but not used directly in the test

	// Test initial empty state
	if manager.activeDialog != -1 {
		t.Errorf("Expected initial activeDialog to be -1, got %d", manager.activeDialog)
	}

	// Add one dialog
	manager.PushDialog(modalDialog)

	// Test active dialog
	if manager.activeDialog != 0 {
		t.Errorf("Expected activeDialog to be 0 after adding first dialog, got %d", manager.activeDialog)
	}

	// Verify z-index
	if modalDialog.ZIndex() != 0 {
		t.Errorf("Expected z-index of first dialog to be 0, got %d", modalDialog.ZIndex())
	}

	// Verify focus
	if !modalDialog.IsFocused() {
		t.Error("Expected first dialog to be focused")
	}

	// Add second dialog
	manager.PushDialog(formDialog)

	// Test z-index after second dialog
	if modalDialog.ZIndex() != 0 {
		t.Errorf("Expected z-index of first dialog to remain 0, got %d", modalDialog.ZIndex())
	}

	if formDialog.ZIndex() != 1 {
		t.Errorf("Expected z-index of second dialog to be 1, got %d", formDialog.ZIndex())
	}

	// Test focus state
	if modalDialog.IsFocused() {
		t.Error("Expected first dialog to no longer be focused")
	}

	if !formDialog.IsFocused() {
		t.Error("Expected second dialog to be focused")
	}

	// Add more dialogs
	manager.PushDialog(listDialog)
	manager.PushDialog(confirmDialog)

	// Test z-index values with multiple dialogs
	expectedZIndexes := map[Dialog]int{
		modalDialog:   0,
		formDialog:    1,
		listDialog:    2,
		confirmDialog: 3,
	}

	for dialog, expected := range expectedZIndexes {
		if dialog.ZIndex() != expected {
			t.Errorf("Dialog %s expected z-index %d, got %d", dialog.Title(), expected, dialog.ZIndex())
		}

		// Only the top dialog should be focused
		if dialog == confirmDialog {
			if !dialog.IsFocused() {
				t.Errorf("Expected top dialog %s to be focused", dialog.Title())
			}
		} else {
			if dialog.IsFocused() {
				t.Errorf("Expected non-top dialog %s to not be focused", dialog.Title())
			}
		}
	}

	// Test active dialog with multiple dialogs
	if manager.activeDialog != 3 {
		t.Errorf("Expected activeDialog to be 3 with 4 dialogs, got %d", manager.activeDialog)
	}

	// Pop a dialog
	poppedDialog := manager.PopDialog()

	// Verify popped dialog
	if poppedDialog != confirmDialog {
		t.Errorf("Expected to pop confirm dialog, got %s", poppedDialog.Title())
	}

	// Test focus after pop
	if !listDialog.IsFocused() {
		t.Error("Expected new top dialog to be focused after pop")
	}

	// Test active dialog after pop
	if manager.activeDialog != 2 {
		t.Errorf("Expected activeDialog to be 2 after pop, got %d", manager.activeDialog)
	}

	// Pop all dialogs
	manager.PopDialog() // list
	manager.PopDialog() // form
	manager.PopDialog() // modal

	// Test empty state after popping all
	if manager.HasDialogs() {
		t.Error("Expected manager to have no dialogs after popping all")
	}

	if manager.activeDialog != -1 {
		t.Errorf("Expected activeDialog to be -1 after popping all, got %d", manager.activeDialog)
	}
}

// TestResizeBehavior tests how dialogs handle window resizing
func TestResizeBehavior(t *testing.T) {
	// Test cases for window resizing
	testCases := []struct {
		initialWidth  int
		initialHeight int
		newWidth      int
		newHeight     int
	}{
		{100, 50, 80, 40},   // Shrinking
		{100, 50, 120, 60},  // Growing
		{100, 50, 100, 60},  // Height only
		{100, 50, 120, 50},  // Width only
		{100, 50, 200, 100}, // Doubling
		{100, 50, 50, 25},   // Halving
	}

	for _, tc := range testCases {
		dialogFactories := []func() Dialog{
			func() Dialog { return NewModalDialog("Modal", 60, 30, NewSimpleModalContent("Content")) },
			func() Dialog {
				return NewLegacyFormDialog("Form", 70, 40, []FormField{
					NewTextField("Field", "Value", false),
				})
			},
			func() Dialog {
				return NewListDialog("List", 50, 25, []ListItem{
					NewSimpleListItem("Item", "Desc"),
				})
			},
		}

		for _, factory := range dialogFactories {
			manager := NewDialogManager(tc.initialWidth, tc.initialHeight)
			dialog := factory()
			manager.PushDialog(dialog)

			// Record initial positions
			_, _, initialX, initialY := dialog.GetRect()

			// Send resize message
			resizeMsg := tea.WindowSizeMsg{
				Width:  tc.newWidth,
				Height: tc.newHeight,
			}

			cmd := manager.HandleMsg(resizeMsg)
			if cmd == nil {
				// A resize should update positions, verifiable even without the command
			}

			// Manager should update its own dimensions
			if manager.termWidth != tc.newWidth || manager.termHeight != tc.newHeight {
				t.Errorf("Manager didn't update dimensions: expected %dx%d, got %dx%d",
					tc.newWidth, tc.newHeight, manager.termWidth, manager.termHeight)
			}

			// Check if dialog position was updated - should now be centered in new dimensions
			_, _, newX, newY := dialog.GetRect()

			// Position should be different when dimensions change (centered differently)
			if tc.initialWidth != tc.newWidth && initialX == newX {
				t.Errorf("Dialog X position didn't change after width resize: %d", newX)
			}

			if tc.initialHeight != tc.newHeight && initialY == newY {
				t.Errorf("Dialog Y position didn't change after height resize: %d", newY)
			}
		}
	}
}

// TestMinimumSize tests that dialogs handle very small window sizes properly
func TestMinimumSize(t *testing.T) {
	// Create a dialog manager with very small dimensions
	manager := NewDialogManager(10, 5)

	// Create a dialog larger than the window
	modalDialog := NewModalDialog("Test", 20, 10, NewSimpleModalContent("Content"))

	// Add to manager
	manager.PushDialog(modalDialog)

	// Get the position
	_, _, x, y := modalDialog.GetRect()

	// When dialog is larger than window, it should still be positioned sensibly
	// (likely negative coordinates to center it)
	if x > 5 {
		t.Errorf("Expected dialog X to be near centered when larger than window, got %d", x)
	}

	if y > 2 {
		t.Errorf("Expected dialog Y to be near centered when larger than window, got %d", y)
	}
}

// TestKeyboardRoutingToActiveDialogOnly tests that keypresses go only to the top dialog
func TestKeyboardRoutingToActiveDialogOnly(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Create a form dialog with a text field
	fields := []FormField{
		NewTextField("Name", "Enter name", false),
	}
	formDialog := NewLegacyFormDialog("Form", 60, 20, fields)

	// Create a confirmation dialog
	confirmDialog := YesNo("Confirm", "Are you sure?", false)

	// Add both dialogs to the stack
	manager.PushDialog(formDialog)
	manager.PushDialog(confirmDialog)

	// The confirmation dialog should be on top and focused
	if manager.GetActiveDialog() != confirmDialog {
		t.Error("Expected confirmation dialog to be active")
	}

	// Send 'y' key - should be handled by the confirmation dialog
	yKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_ = manager.HandleMsg(yKeyMsg)

	// Confirmation dialog should have been closed by the 'y' key
	if manager.GetActiveDialog() == confirmDialog {
		t.Error("Expected confirmation dialog to be closed after 'y' key")
	}

	// The form dialog should now be active
	if manager.GetActiveDialog() != formDialog {
		t.Error("Expected form dialog to be active after confirmation closed")
	}

	// Now type a character into the text field
	aKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	_ = manager.HandleMsg(aKeyMsg)

	// The text field should have received the character
	textValue := formDialog.GetFieldValue(0).(string)
	if textValue != "a" {
		t.Errorf("Expected text field to contain 'a', got '%s'", textValue)
	}
}

// TestDialogManagerViewWithMultipleDialogs tests the rendering of multiple stacked dialogs
func TestDialogManagerViewWithMultipleDialogs(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Empty manager should return empty string
	if view := manager.View(); view != "" {
		t.Errorf("Expected empty view for empty manager, got '%s'", view)
	}

	// Add a modal dialog
	modalDialog := NewModalDialog("Modal", 60, 30, NewSimpleModalContent("Modal Content"))
	manager.PushDialog(modalDialog)

	// View should return something
	if view := manager.View(); view == "" {
		t.Error("Expected non-empty view after adding dialog")
	}

	// Add confirmation dialog
	confirmDialog := YesNo("Confirm", "Are you sure?", false)
	manager.PushDialog(confirmDialog)

	// View should still return something
	if view := manager.View(); view == "" {
		t.Error("Expected non-empty view with multiple dialogs")
	}

	// For now, the implementation just returns the top dialog
	// In a real compositor, we'd test the z-order rendering
}
