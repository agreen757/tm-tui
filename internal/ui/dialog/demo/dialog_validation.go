package demo

import (
	"fmt"
	"strings"

	"github.com/adriangreen/tm-tui/internal/ui/dialog"
)

// ValidateDialogComponents runs validation of all dialog components
func ValidateDialogComponents() string {
	var results []string

	// Validate modal dialog
	modalResults := validateModalDialog()
	results = append(results, modalResults...)

	// Validate form dialog
	formResults := validateFormDialog()
	results = append(results, formResults...)

	// Validate list dialog
	listResults := validateListDialog()
	results = append(results, listResults...)

	// Validate confirmation dialog
	confirmResults := validateConfirmationDialog()
	results = append(results, confirmResults...)

	// Validate progress dialog
	progressResults := validateProgressDialog()
	results = append(results, progressResults...)

	// Validate dialog manager
	managerResults := validateDialogManager()
	results = append(results, managerResults...)

	// Create a report
	if len(results) == 0 {
		return "All dialog components validated successfully!"
	}

	return fmt.Sprintf("Dialog validation found %d issues:\n\n%s",
		len(results), strings.Join(results, "\n"))
}

// validateModalDialog validates the modal dialog
func validateModalDialog() []string {
	var issues []string

	// Test modal dialog creation
	content := dialog.NewSimpleModalContent("Test content")
	modal := dialog.NewModalDialog("Test Modal", 40, 15, content)

	// Verify properties
	if modal.Title() != "Test Modal" {
		issues = append(issues, "Modal dialog title mismatch")
	}

	if modal.Kind() != dialog.DialogKindModal {
		issues = append(issues, "Modal dialog kind incorrect")
	}

	// Verify positioning
	modal.Center(100, 50)
	width, height, x, y := modal.GetRect()

	if width != 40 || height != 15 || x != 30 || y != 17 {
		issues = append(issues, fmt.Sprintf("Modal positioning incorrect: expected 40x15@(30,17), got %dx%d@(%d,%d)",
			width, height, x, y))
	}

	// Test view rendering
	view := modal.View()
	if view == "" {
		issues = append(issues, "Modal dialog View() returned empty string")
	}

	return issues
}

// validateFormDialog validates the form dialog
func validateFormDialog() []string {
	var issues []string

	// Test form dialog creation
	fields := []dialog.FormField{
		dialog.NewTextField("Name", "Enter name", true),
		dialog.NewCheckboxField("Agree", false),
		dialog.NewRadioGroupField("Option", []string{"A", "B", "C"}, 0),
	}
	form := dialog.NewLegacyFormDialog("Test Form", 50, 20, fields)

	// Verify properties
	if form.Title() != "Test Form" {
		issues = append(issues, "Form dialog title mismatch")
	}

	if form.Kind() != dialog.DialogKindForm {
		issues = append(issues, "Form dialog kind incorrect")
	}

	// Verify field access
	nameValue := form.GetFieldValueByLabel("Name").(string)
	if nameValue != "" {
		issues = append(issues, "Expected empty name field initially")
	}

	agreeValue := form.GetFieldValueByLabel("Agree").(bool)
	if agreeValue != false {
		issues = append(issues, "Expected unchecked checkbox initially")
	}

	optionValue := form.GetFieldValueByLabel("Option").(int)
	if optionValue != 0 {
		issues = append(issues, "Expected first radio option selected initially")
	}

	// Test view rendering
	view := form.View()
	if view == "" {
		issues = append(issues, "Form dialog View() returned empty string")
	}

	return issues
}

// validateListDialog validates the list dialog
func validateListDialog() []string {
	var issues []string

	// Test list dialog creation
	items := []dialog.ListItem{
		dialog.NewSimpleListItem("Item 1", "Description 1"),
		dialog.NewSimpleListItem("Item 2", "Description 2"),
		dialog.NewSimpleListItem("Item 3", "Description 3"),
	}
	list := dialog.NewListDialog("Test List", 40, 15, items)

	// Verify properties
	if list.Title() != "Test List" {
		issues = append(issues, "List dialog title mismatch")
	}

	if list.Kind() != dialog.DialogKindList {
		issues = append(issues, "List dialog kind incorrect")
	}

	// Verify initial selection
	if list.SelectedIndex() != 0 {
		issues = append(issues, fmt.Sprintf("Expected initial selection 0, got %d", list.SelectedIndex()))
	}

	// Set selection
	list.SetSelectedIndex(1)
	if list.SelectedIndex() != 1 {
		issues = append(issues, fmt.Sprintf("Expected selection 1 after SetSelectedIndex, got %d", list.SelectedIndex()))
	}

	// Get selected item
	selected := list.SelectedItem()
	if selected == nil {
		issues = append(issues, "SelectedItem() returned nil")
	} else {
		if selected.Title() != "Item 2" {
			issues = append(issues, fmt.Sprintf("Expected selected item title 'Item 2', got '%s'", selected.Title()))
		}
	}

	// Test view rendering
	view := list.View()
	if view == "" {
		issues = append(issues, "List dialog View() returned empty string")
	}

	return issues
}

// validateConfirmationDialog validates the confirmation dialog
func validateConfirmationDialog() []string {
	var issues []string

	// Test confirmation dialog creation
	confirm := dialog.YesNo("Test Confirmation", "Are you sure?", false)

	// Verify properties
	if confirm.Title() != "Test Confirmation" {
		issues = append(issues, "Confirmation dialog title mismatch")
	}

	if confirm.Kind() != dialog.DialogKindConfirmation {
		issues = append(issues, "Confirmation dialog kind incorrect")
	}

	// Verify default selection
	if confirm.FocusedIndex() != 0 {
		issues = append(issues, fmt.Sprintf("Expected default focus on Yes (0), got %d", confirm.FocusedIndex()))
	}

	// Change default
	confirm.SetYesDefault(false)
	if confirm.FocusedIndex() != 1 {
		issues = append(issues, fmt.Sprintf("Expected focus on No (1) after SetYesDefault(false), got %d", confirm.FocusedIndex()))
	}

	// Test view rendering
	view := confirm.View()
	if view == "" {
		issues = append(issues, "Confirmation dialog View() returned empty string")
	}

	return issues
}

// validateProgressDialog validates the progress dialog
func validateProgressDialog() []string {
	var issues []string

	// Test progress dialog creation
	progress := dialog.NewProgressDialog("Test Progress", 50, 10)

	// Verify properties
	if progress.Title() != "Test Progress" {
		issues = append(issues, "Progress dialog title mismatch")
	}

	if progress.Kind() != dialog.DialogKindProgress {
		issues = append(issues, "Progress dialog kind incorrect")
	}

	// Verify initial state
	if progress.Progress() != 0.0 {
		issues = append(issues, fmt.Sprintf("Expected initial progress 0.0, got %f", progress.Progress()))
	}

	// Update progress
	progress.SetProgress(0.5)
	if progress.Progress() != 0.5 {
		issues = append(issues, fmt.Sprintf("Expected progress 0.5 after SetProgress, got %f", progress.Progress()))
	}

	// Test label
	progress.SetLabel("Halfway done")
	if progress.Label() != "Halfway done" {
		issues = append(issues, fmt.Sprintf("Expected label 'Halfway done', got '%s'", progress.Label()))
	}

	// Test view rendering
	view := progress.View()
	if view == "" {
		issues = append(issues, "Progress dialog View() returned empty string")
	}

	return issues
}

// validateDialogManager validates the dialog manager
func validateDialogManager() []string {
	var issues []string

	// Test dialog manager creation
	manager := dialog.NewDialogManager(100, 50)

	// Verify initial state
	if manager.HasDialogs() {
		issues = append(issues, "Expected no dialogs initially")
	}

	if manager.GetActiveDialog() != nil {
		issues = append(issues, "Expected no active dialog initially")
	}

	// Add a dialog
	dialog1 := dialog.NewModalDialog("Dialog 1", 40, 20, dialog.NewSimpleModalContent("Content 1"))
	manager.PushDialog(dialog1)

	// Verify dialog was added
	if !manager.HasDialogs() {
		issues = append(issues, "Expected HasDialogs() to be true after PushDialog")
	}

	if manager.GetActiveDialog() == nil {
		issues = append(issues, "Expected non-nil active dialog after PushDialog")
	}

	// Add another dialog
	dialog2 := dialog.NewModalDialog("Dialog 2", 30, 15, dialog.NewSimpleModalContent("Content 2"))
	manager.PushDialog(dialog2)

	// Verify z-index and focus
	if dialog1.ZIndex() >= dialog2.ZIndex() {
		issues = append(issues, fmt.Sprintf("Expected dialog2 z-index > dialog1 z-index, got %d <= %d",
			dialog2.ZIndex(), dialog1.ZIndex()))
	}

	if dialog1.IsFocused() || !dialog2.IsFocused() {
		issues = append(issues, "Expected dialog2 focused, dialog1 unfocused")
	}

	// Pop a dialog
	popped := manager.PopDialog()
	if popped != dialog2 {
		issues = append(issues, "Expected PopDialog() to return dialog2")
	}

	// Verify state after pop
	if manager.GetActiveDialog() != dialog1 {
		issues = append(issues, "Expected dialog1 to be active after pop")
	}

	if !dialog1.IsFocused() {
		issues = append(issues, "Expected dialog1 to be focused after pop")
	}

	// Test view rendering
	view := manager.View()
	if view == "" {
		issues = append(issues, "Dialog manager View() returned empty string")
	}

	return issues
}
