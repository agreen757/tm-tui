package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDialogKind(t *testing.T) {
	tests := []struct {
		kind     DialogKind
		expected DialogKind
	}{
		{DialogKindModal, DialogKindModal},
		{DialogKindForm, DialogKindForm},
		{DialogKindList, DialogKindList},
		{DialogKindConfirmation, DialogKindConfirmation},
		{DialogKindProgress, DialogKindProgress},
	}

	for _, test := range tests {
		dialog := NewBaseDialog("Test", 50, 10, test.kind)
		if dialog.Kind() != test.expected {
			t.Errorf("Expected dialog kind %d, got %d", test.expected, dialog.Kind())
		}
	}
}

func TestModalCentering(t *testing.T) {
	tests := []struct {
		containerWidth  int
		containerHeight int
		dialogWidth     int
		dialogHeight    int
		expectedX       int
		expectedY       int
	}{
		{100, 50, 40, 20, 30, 15},
		{80, 40, 60, 30, 10, 5},
		{120, 80, 80, 40, 20, 20},
		// Edge cases
		{40, 20, 40, 20, 0, 0},   // Dialog same size as container
		{30, 15, 40, 20, -5, -2}, // Dialog larger than container
	}

	for _, test := range tests {
		dialog := NewBaseDialog("Test", test.dialogWidth, test.dialogHeight, DialogKindModal)
		dialog.Center(test.containerWidth, test.containerHeight)

		_, _, x, y := dialog.GetRect()
		if x != test.expectedX || y != test.expectedY {
			t.Errorf("Expected position (%d, %d), got (%d, %d)", test.expectedX, test.expectedY, x, y)
		}
	}
}

func TestDialogManager(t *testing.T) {
	manager := NewDialogManager(100, 50)

	// Initial state
	if manager.HasDialogs() {
		t.Error("Expected no dialogs initially")
	}

	if manager.GetActiveDialog() != nil {
		t.Error("Expected no active dialog")
	}

	// Add a dialog
	dialog1 := NewModalDialog("Dialog 1", 60, 30, NewSimpleModalContent("Content 1"))
	manager.PushDialog(dialog1)

	if !manager.HasDialogs() {
		t.Error("Expected to have dialogs after adding one")
	}

	if manager.GetActiveDialog() == nil {
		t.Error("Expected an active dialog after adding one")
	}

	// Add another dialog
	dialog2 := NewModalDialog("Dialog 2", 40, 20, NewSimpleModalContent("Content 2"))
	manager.PushDialog(dialog2)

	// Check z-index order
	if dialog1.ZIndex() >= dialog2.ZIndex() {
		t.Error("Expected dialog2 to have higher z-index")
	}

	// Test focus state
	if !dialog2.IsFocused() || dialog1.IsFocused() {
		t.Error("Expected dialog2 to be focused and dialog1 not focused")
	}

	// Pop dialog
	poppedDialog := manager.PopDialog()
	if poppedDialog != dialog2 {
		t.Error("Expected to pop dialog2")
	}

	// Check active dialog
	if manager.GetActiveDialog() != dialog1 {
		t.Error("Expected dialog1 to be active after popping dialog2")
	}

	// Check focus state after pop
	if !dialog1.IsFocused() {
		t.Error("Expected dialog1 to be focused after pop")
	}

	// Pop last dialog
	poppedDialog = manager.PopDialog()
	if poppedDialog != dialog1 {
		t.Error("Expected to pop dialog1")
	}

	// Check manager state after all dialogs popped
	if manager.HasDialogs() {
		t.Error("Expected no dialogs after popping all")
	}

	if manager.GetActiveDialog() != nil {
		t.Error("Expected no active dialog after popping all")
	}
}

func TestFocusableDialog(t *testing.T) {
	dialog := NewBaseFocusableDialog("Test", 50, 30, DialogKindForm, 3)

	// Initial focus
	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected focused index to be 0, got %d", dialog.FocusedIndex())
	}

	// Test focus next
	dialog.FocusNext()
	if dialog.FocusedIndex() != 1 {
		t.Errorf("Expected focused index to be 1 after FocusNext, got %d", dialog.FocusedIndex())
	}

	// Test focus previous
	dialog.FocusPrev()
	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected focused index to be 0 after FocusPrev, got %d", dialog.FocusedIndex())
	}

	// Test set focused index
	dialog.SetFocusedIndex(2)
	if dialog.FocusedIndex() != 2 {
		t.Errorf("Expected focused index to be 2 after SetFocusedIndex, got %d", dialog.FocusedIndex())
	}

	// Test bounds checking
	dialog.SetFocusedIndex(10) // Should not change
	if dialog.FocusedIndex() != 2 {
		t.Errorf("Expected focused index to still be 2 after invalid SetFocusedIndex, got %d", dialog.FocusedIndex())
	}

	// Test wrapping behavior
	dialog.FocusNext() // Should wrap to 0
	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected focused index to wrap to 0, got %d", dialog.FocusedIndex())
	}

	dialog.FocusPrev() // Should wrap to 2
	if dialog.FocusedIndex() != 2 {
		t.Errorf("Expected focused index to wrap to 2, got %d", dialog.FocusedIndex())
	}

	// Test handling of key events
	tabKeyMsg := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := dialog.HandleBaseFocusableKey(tabKeyMsg)
	if result != DialogResultNone || dialog.FocusedIndex() != 0 {
		t.Errorf("Expected DialogResultNone and focused index 0 after Tab, got %v and %d", result, dialog.FocusedIndex())
	}

	// Test Esc key
	escKeyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	result, _ = dialog.HandleBaseFocusableKey(escKeyMsg)
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel after Esc, got %v", result)
	}
}

func TestProgressDialog(t *testing.T) {
	dialog := NewProgressDialog("Test Progress", 60, 10)

	// Initial state
	if dialog.Progress() != 0.0 {
		t.Errorf("Expected initial progress to be 0.0, got %f", dialog.Progress())
	}

	if dialog.IsCanceled() {
		t.Error("Expected progress dialog to not be canceled initially")
	}

	if dialog.IsCompleted() {
		t.Error("Expected progress dialog to not be completed initially")
	}

	// Update progress
	dialog.SetProgress(0.5)
	if dialog.Progress() != 0.5 {
		t.Errorf("Expected progress to be 0.5 after SetProgress, got %f", dialog.Progress())
	}

	// Test bounds
	dialog.SetProgress(1.5) // Should clamp to 1.0
	if dialog.Progress() != 1.0 {
		t.Errorf("Expected progress to be clamped to 1.0, got %f", dialog.Progress())
	}

	dialog.SetProgress(-0.5) // Should clamp to 0.0
	if dialog.Progress() != 0.0 {
		t.Errorf("Expected progress to be clamped to 0.0, got %f", dialog.Progress())
	}

	// Test label
	dialog.SetLabel("Processing...")
	if dialog.Label() != "Processing..." {
		t.Errorf("Expected label to be \"Processing...\", got \"%s\"", dialog.Label())
	}

	// Test auto-close
	dialog.SetAutoClose(false)
	if dialog.autoClose {
		t.Error("Expected autoClose to be false after SetAutoClose(false)")
	}

	// Test update message handling
	updatedDialog, cmd := dialog.Update(ProgressUpdateMsg{Progress: 0.75, Label: "Almost done"})
	progressDialog, ok := updatedDialog.(*ProgressDialog)
	if !ok {
		t.Error("Expected ProgressDialog from Update")
	} else {
		if progressDialog.Progress() != 0.75 {
			t.Errorf("Expected progress to be 0.75, got %f", progressDialog.Progress())
		}
		if progressDialog.Label() != "Almost done" {
			t.Errorf("Expected label to be \"Almost done\", got \"%s\"", progressDialog.Label())
		}
	}

	// Test completion
	progressDialog.SetAutoClose(true)
	updatedDialog, cmd = progressDialog.Update(ProgressUpdateMsg{Progress: 1.0, Label: "Complete"})
	if progressDialog.Progress() != 1.0 || !progressDialog.IsCompleted() {
		t.Errorf("Expected progress 1.0 and completed state, got %f and %v", progressDialog.Progress(), progressDialog.IsCompleted())
	}
	if cmd == nil {
		t.Error("Expected command to be returned for auto-closing dialog")
	}

	// Test cancellation via esc key
	cKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	result, cmd := progressDialog.HandleKey(cKeyMsg)
	if result != DialogResultCancel || progressDialog.IsCanceled() != true {
		t.Errorf("Expected DialogResultCancel and canceled state, got %v and %v", result, progressDialog.IsCanceled())
	}
}

func TestListDialog(t *testing.T) {
	items := []ListItem{
		&SimpleListItem{title: "Item 1", description: "Description 1"},
		&SimpleListItem{title: "Item 2", description: "Description 2"},
		&SimpleListItem{title: "Item 3", description: "Description 3"},
	}

	dialog := NewListDialog("Test List", 60, 20, items)

	// Initial selection
	if dialog.SelectedIndex() != 0 {
		t.Errorf("Expected selected index to be 0, got %d", dialog.SelectedIndex())
	}

	// Test selection movement
	upDownKeys := []struct {
		key         tea.KeyMsg
		expectedIdx int
	}{
		{tea.KeyMsg{Type: tea.KeyDown}, 1},
		{tea.KeyMsg{Type: tea.KeyDown}, 2},
		{tea.KeyMsg{Type: tea.KeyDown}, 0}, // Wraps around
		{tea.KeyMsg{Type: tea.KeyUp}, 2},
		{tea.KeyMsg{Type: tea.KeyUp}, 1},
	}

	for i, test := range upDownKeys {
		result, _ := dialog.HandleKey(test.key)
		if result != DialogResultNone {
			t.Errorf("Test %d: Expected DialogResultNone for navigation, got %v", i, result)
		}
		if dialog.SelectedIndex() != test.expectedIdx {
			t.Errorf("Test %d: Expected selected index to be %d, got %d", i, test.expectedIdx, dialog.SelectedIndex())
		}
	}

	// Set selected index
	dialog.SetSelectedIndex(1)
	if dialog.SelectedIndex() != 1 {
		t.Errorf("Expected selected index to be 1 after SetSelectedIndex, got %d", dialog.SelectedIndex())
	}

	// Test selected item
	item := dialog.SelectedItem().(*SimpleListItem)
	if item.Title() != "Item 2" {
		t.Errorf("Expected selected item title to be \"Item 2\", got \"%s\"", item.Title())
	}

	// Test multi-select
	dialog.SetMultiSelect(true)
	if !dialog.multiSelect {
		t.Error("Expected multi-select to be enabled")
	}

	// Select an item with space
	spaceKey := tea.KeyMsg{Type: tea.KeySpace}
	result, _ := dialog.HandleKey(spaceKey)
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone for space key, got %v", result)
	}

	// Verify item was selected
	selectedItems := dialog.SelectedItems()
	if len(selectedItems) != 1 {
		t.Errorf("Expected 1 selected item, got %d", len(selectedItems))
	}

	// Deselect the item
	result, _ = dialog.HandleKey(spaceKey)
	selectedItems = dialog.SelectedItems()
	if len(selectedItems) != 0 {
		t.Errorf("Expected 0 selected items after toggling, got %d", len(selectedItems))
	}

	// Test item selection with Enter
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	result, cmd := dialog.HandleKey(enterKey)
	if result != DialogResultConfirm {
		t.Errorf("Expected DialogResultConfirm for Enter key, got %v", result)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for list selection")
	}

	// Test home/end keys
	homeKey := tea.KeyMsg{Type: tea.KeyHome}
	result, _ = dialog.HandleKey(homeKey)
	if result != DialogResultNone || dialog.SelectedIndex() != 0 {
		t.Errorf("Expected DialogResultNone and index 0 after Home key, got %v and %d", result, dialog.SelectedIndex())
	}

	endKey := tea.KeyMsg{Type: tea.KeyEnd}
	result, _ = dialog.HandleKey(endKey)
	if result != DialogResultNone || dialog.SelectedIndex() != 2 {
		t.Errorf("Expected DialogResultNone and index 2 after End key, got %v and %d", result, dialog.SelectedIndex())
	}

	// Test page up/down
	dialog.SetSelectedIndex(1)
	pageUpKey := tea.KeyMsg{Type: tea.KeyPgUp}
	result, _ = dialog.HandleKey(pageUpKey)
	if result != DialogResultNone || dialog.SelectedIndex() != 0 {
		t.Errorf("Expected DialogResultNone and index 0 after PgUp key, got %v and %d", result, dialog.SelectedIndex())
	}

	pageDownKey := tea.KeyMsg{Type: tea.KeyPgDown}
	result, _ = dialog.HandleKey(pageDownKey)
	if result != DialogResultNone || dialog.SelectedIndex() != 2 {
		t.Errorf("Expected DialogResultNone and index 2 after PgDown key, got %v and %d", result, dialog.SelectedIndex())
	}
}

func TestConfirmationDialog(t *testing.T) {
	dialog := YesNo("Confirmation", "Are you sure?", false)

	// Default selection
	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected focused index to be 0 (Yes), got %d", dialog.FocusedIndex())
	}

	// Change default
	dialog.SetYesDefault(false)
	if dialog.FocusedIndex() != 1 {
		t.Errorf("Expected focused index to be 1 (No) after SetYesDefault(false), got %d", dialog.FocusedIndex())
	}

	// Custom button text
	dialog.SetYesText("Confirm")
	dialog.SetNoText("Cancel")

	// Initial result
	if dialog.Result() != ConfirmationResultNone {
		t.Errorf("Expected initial result to be None, got %d", dialog.Result())
	}

	// Test left/right navigation
	dialog.SetFocusedIndex(0) // Start on Yes
	rightKey := tea.KeyMsg{Type: tea.KeyRight}
	result, _ := dialog.HandleKey(rightKey)
	if result != DialogResultNone || dialog.FocusedIndex() != 1 {
		t.Errorf("Expected DialogResultNone and index 1 after Right key, got %v and %d", result, dialog.FocusedIndex())
	}

	leftKey := tea.KeyMsg{Type: tea.KeyLeft}
	result, _ = dialog.HandleKey(leftKey)
	if result != DialogResultNone || dialog.FocusedIndex() != 0 {
		t.Errorf("Expected DialogResultNone and index 0 after Left key, got %v and %d", result, dialog.FocusedIndex())
	}

	// Test Yes selection
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	result, cmd := dialog.HandleKey(enterKey)
	if result != DialogResultConfirm {
		t.Errorf("Expected DialogResultConfirm for Yes selection, got %v", result)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for confirmation")
	}
	if dialog.Result() != ConfirmationResultYes {
		t.Errorf("Expected ConfirmationResultYes, got %d", dialog.Result())
	}

	// Test No selection
	dialog.result = ConfirmationResultNone // Reset
	dialog.SetFocusedIndex(1)              // Select No
	result, cmd = dialog.HandleKey(enterKey)
	if result != DialogResultCancel {
		t.Errorf("Expected DialogResultCancel for No selection, got %v", result)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for confirmation")
	}
	if dialog.Result() != ConfirmationResultNo {
		t.Errorf("Expected ConfirmationResultNo, got %d", dialog.Result())
	}

	// Test shortcut keys
	dialog.result = ConfirmationResultNone // Reset
	yKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	result, _ = dialog.HandleKey(yKey)
	if result != DialogResultConfirm || dialog.Result() != ConfirmationResultYes {
		t.Errorf("Expected DialogResultConfirm and ConfirmationResultYes for 'y' key, got %v and %d", result, dialog.Result())
	}

	dialog.result = ConfirmationResultNone // Reset
	nKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	result, _ = dialog.HandleKey(nKey)
	if result != DialogResultCancel || dialog.Result() != ConfirmationResultNo {
		t.Errorf("Expected DialogResultCancel and ConfirmationResultNo for 'n' key, got %v and %d", result, dialog.Result())
	}

	// Test danger mode
	dangerDialog := YesNo("Danger", "Destructive action", true)
	if !dangerDialog.dangerMode {
		t.Error("Expected danger mode to be enabled")
	}
}

func TestFormDialog(t *testing.T) {
	fields := []FormField{
		NewTextField("Name", "Enter name", true),
		NewCheckboxField("Agree to terms", false),
		NewRadioGroupField("Option", []string{"Option 1", "Option 2", "Option 3"}, 0),
	}

	dialog := NewLegacyFormDialog("Test Form", 60, 20, fields)

	// Initial state
	if dialog.IsSubmitted() {
		t.Error("Expected form to not be submitted initially")
	}

	if dialog.FocusedIndex() != 0 {
		t.Errorf("Expected focused index to be 0, got %d", dialog.FocusedIndex())
	}

	// Test field value - initial state of checkbox
	checkboxValue := dialog.GetFieldValueByLabel("Agree to terms").(bool)
	if checkboxValue {
		t.Error("Expected checkbox to be unchecked initially")
	}

	// Test custom button labels
	dialog.SetSubmitLabel("Save")
	dialog.SetCancelLabel("Exit")

	// Test field value access
	if dialog.GetFieldValue(0) == nil {
		t.Error("Expected GetFieldValue(0) to return a value")
	}

	if dialog.GetFieldValueByLabel("Name") == nil {
		t.Error("Expected GetFieldValueByLabel('Name') to return a value")
	}

	// Test field validation
	if dialog.fields[0].ValidationError != "" {
		t.Errorf("Expected no validation error initially, got %s", dialog.fields[0].ValidationError)
	}

	dialog.fields[0].ValidateField()
	if dialog.fields[0].ValidationError == "" { // This is a required field with empty value
		t.Error("Expected validation error for empty required field")
	}

	// Test form field navigation
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	result, _ := dialog.HandleKey(tabKey)
	if result != DialogResultNone || dialog.FocusedIndex() != 1 {
		t.Errorf("Expected DialogResultNone and index 1 after tab, got %v and %d", result, dialog.FocusedIndex())
	}

	// Test checkbox toggling
	spaceKey := tea.KeyMsg{Type: tea.KeySpace}
	result, _ = dialog.HandleKey(spaceKey)
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone after toggling checkbox, got %v", result)
	}

	checkboxValue = dialog.GetFieldValueByLabel("Agree to terms").(bool)
	if !checkboxValue {
		t.Error("Expected checkbox to be checked after toggling")
	}

	// Test radiogroup navigation
	dialog.FocusNext() // Move to radiogroup
	if dialog.FocusedIndex() != 2 {
		t.Errorf("Expected focused index to be 2, got %d", dialog.FocusedIndex())
	}

	// Select a different radio option
	downKey := tea.KeyMsg{Type: tea.KeyDown}
	result, _ = dialog.HandleKey(downKey)
	if result != DialogResultNone {
		t.Errorf("Expected DialogResultNone after radio selection, got %v", result)
	}

	radioValue := dialog.GetFieldValueByLabel("Option").(int)
	if radioValue != 1 {
		t.Errorf("Expected radio value 1 after down key, got %d", radioValue)
	}
}

// TestDialogWindowResize tests that dialogs properly handle window resize events
func TestDialogWindowResize(t *testing.T) {
	// Test with different dialog types
	dialogs := []Dialog{
		NewModalDialog("Modal", 50, 30, NewSimpleModalContent("Content")),
		NewProgressDialog("Progress", 50, 10),
		NewLegacyFormDialog("Form", 60, 40, []FormField{
			NewTextField("Field", "Enter value", false),
		}),
	}

	for _, dialog := range dialogs {
		// Initial position
		_, _, x1, y1 := dialog.GetRect()

		// Send resize message
		updatedDialog, _ := dialog.Update(tea.WindowSizeMsg{Width: 200, Height: 100})

		// Check if position was updated
		_, _, x2, y2 := updatedDialog.GetRect()

		// Position should be different after resize
		if x1 == x2 && y1 == y2 {
			t.Errorf("Dialog of type %d did not update position on resize", dialog.Kind())
		}
	}
}

// TestDialogZIndex tests z-index management in the dialog stack
func TestDialogZIndex(t *testing.T) {
	manager := NewDialogManager(100, 80)

	dialogs := []Dialog{
		NewModalDialog("Dialog 1", 50, 30, NewSimpleModalContent("Content 1")),
		NewModalDialog("Dialog 2", 40, 25, NewSimpleModalContent("Content 2")),
		NewModalDialog("Dialog 3", 30, 20, NewSimpleModalContent("Content 3")),
	}

	// Add dialogs to the stack
	for _, dialog := range dialogs {
		manager.PushDialog(dialog)
	}

	// Check z-indexes are in ascending order
	if len(manager.dialogs) != 3 {
		t.Errorf("Expected 3 dialogs, got %d", len(manager.dialogs))
	}

	for i := 0; i < len(manager.dialogs)-1; i++ {
		if manager.dialogs[i].ZIndex() >= manager.dialogs[i+1].ZIndex() {
			t.Errorf("Dialog at position %d has z-index %d, which is >= dialog at position %d with z-index %d",
				i, manager.dialogs[i].ZIndex(), i+1, manager.dialogs[i+1].ZIndex())
		}
	}

	// Active dialog should be the top one
	if manager.GetActiveDialog() != dialogs[2] {
		t.Error("Expected top dialog to be active")
	}

	// Pop a dialog
	manager.PopDialog()

	// Check new active dialog
	if manager.GetActiveDialog() != dialogs[1] {
		t.Error("Expected second dialog to be active after pop")
	}

	// Check focus after pop
	if !manager.GetActiveDialog().IsFocused() {
		t.Error("Expected active dialog to be focused after pop")
	}
}

// TestRenderBorder tests the border rendering with and without title
func TestRenderBorder(t *testing.T) {
	// Dialog with title
	dialog1 := NewBaseDialog("Test Title", 40, 10, DialogKindModal)
	border1 := dialog1.RenderBorder("Content")

	// Dialog without title
	dialog2 := NewBaseDialog("", 40, 10, DialogKindModal)
	border2 := dialog2.RenderBorder("Content")

	// Both should return non-empty strings
	if border1 == "" || border2 == "" {
		t.Error("Expected non-empty border rendering")
	}

	// The version with title should be different
	if border1 == border2 {
		t.Error("Expected different rendering for dialog with and without title")
	}

	// Border drawing with focus
	dialog1.SetFocused(true)
	borderFocused := dialog1.RenderBorder("Content")

	// Rendering should still produce output when focused
	if borderFocused == "" {
		t.Error("Expected non-empty rendering for focused dialog")
	}
	if dialog1.Style.BorderColor == dialog1.Style.FocusedBorderColor {
		t.Error("Expected focused border color to differ from default")
	}
}

// TestSimpleListItem tests the SimpleListItem implementation
func TestSimpleListItem(t *testing.T) {
	item := NewSimpleListItem("Title", "Description")

	if item.Title() != "Title" {
		t.Errorf("Expected title 'Title', got '%s'", item.Title())
	}

	if item.Description() != "Description" {
		t.Errorf("Expected description 'Description', got '%s'", item.Description())
	}

	if item.FilterValue() != "Title" {
		t.Errorf("Expected filter value 'Title', got '%s'", item.FilterValue())
	}
}
