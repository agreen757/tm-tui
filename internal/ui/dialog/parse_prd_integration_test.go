package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestParsePRDDialogFullWorkflowWithValidInputs tests the complete workflow with valid data
func TestParsePRDDialogFullWorkflowWithValidInputs(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Verify initial state
	if dialog.focusedField != 0 {
		t.Errorf("Initial focus should be on field 0, got %d", dialog.focusedField)
	}

	if dialog.operationMode != ParsePRDReplace {
		t.Errorf("Initial mode should be Replace, got %d", dialog.operationMode)
	}

	// Simulate user workflow: Tab through fields
	tests := []struct {
		name          string
		expectedField int
		key           tea.KeyMsg
	}{
		{"Start at field 0", 0, tea.KeyMsg{}},
		{"Tab to field 1", 1, tea.KeyMsg{Type: tea.KeyTab}},
		{"Tab to field 2", 2, tea.KeyMsg{Type: tea.KeyTab}},
		{"Tab to field 3", 3, tea.KeyMsg{Type: tea.KeyTab}},
	}

	currentField := 0
	for _, test := range tests {
		if test.key.Type == tea.KeyTab {
			dialog.HandleKey(test.key)
			currentField = (currentField + 1) % 4
		}

		if dialog.focusedField != test.expectedField {
			t.Errorf("%s: expected field %d, got %d", test.name, test.expectedField, dialog.focusedField)
		}
	}
}

// TestParsePRDDialogFieldNavigationWithShiftTab tests backward navigation
func TestParsePRDDialogFieldNavigationWithShiftTab(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.focusedField = 3 // Start at buttons

	// Navigate backward with Shift+Tab
	for i := 0; i < 4; i++ {
		expectedField := (3 - i) % 4
		if dialog.focusedField != expectedField {
			t.Errorf("Iteration %d: expected field %d, got %d", i, expectedField, dialog.focusedField)
		}

		dialog.HandleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	}
}

// TestParsePRDDialogOperationModeSelection tests mode switching workflow
func TestParsePRDDialogOperationModeSelection(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.focusedField = 2 // Focus operation mode

	// Test workflow: Start with Replace, toggle to Append, back to Replace
	if dialog.operationMode != ParsePRDReplace {
		t.Errorf("Initial mode should be Replace")
	}

	// Toggle to Append
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if dialog.operationMode != ParsePRDAppend {
		t.Errorf("After space, mode should be Append")
	}

	// Toggle back to Replace
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if dialog.operationMode != ParsePRDReplace {
		t.Errorf("After second space, mode should be Replace")
	}
}

// TestParsePRDDialogTagsInputFieldInteraction tests tags field workflow
func TestParsePRDDialogTagsInputFieldInteraction(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.focusedField = 1 // Focus tags field

	// Set some tags
	dialog.tagsInput.SetValue("feature,auth,ui")

	// Verify tags are stored
	tags := dialog.GetTags()
	if tags != "feature,auth,ui" {
		t.Errorf("Expected 'feature,auth,ui', got %q", tags)
	}

	// Verify GetAppendFlag works with different modes
	dialog.focusedField = 2
	dialog.operationMode = ParsePRDAppend
	if !dialog.GetAppendFlag() {
		t.Errorf("GetAppendFlag should return true for Append mode")
	}

	dialog.operationMode = ParsePRDReplace
	if dialog.GetAppendFlag() {
		t.Errorf("GetAppendFlag should return false for Replace mode")
	}
}

// TestParsePRDDialogButtonWorkflow tests button navigation and selection
func TestParsePRDDialogButtonWorkflow(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.focusedField = 3 // Focus buttons

	// Verify initial button state
	if !dialog.submitBtn || dialog.cancelBtn {
		t.Errorf("Initial state: submitBtn should be true, cancelBtn false")
	}

	// Navigate to cancel button
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	if dialog.submitBtn || !dialog.cancelBtn {
		t.Errorf("After right key: submitBtn should be false, cancelBtn true")
	}

	// Navigate back to submit button
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if !dialog.submitBtn || dialog.cancelBtn {
		t.Errorf("After left key: submitBtn should be true, cancelBtn false")
	}
}

// TestParsePRDDialogValidationIntegration tests validation across the form
func TestParsePRDDialogValidationIntegration(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Start with no validation
	if dialog.fileValid {
		t.Errorf("Initial fileValid should be false")
	}

	// Validate various file paths
	dialog.validateFilePath("")
	if dialog.fileValid || dialog.fileValidMsg != "No file selected" {
		t.Errorf("Empty path validation failed")
	}

	// Validate tags at various lengths
	dialog.validateTags("")
	if !dialog.tagValid {
		t.Errorf("Empty tags should be valid")
	}

	dialog.validateTags("a,b,c,d,e,f,g,h,i,j,k,l,m,n,o")
	if !dialog.tagValid {
		t.Errorf("15-char tags should be valid")
	}

	// Update validation across all fields
	dialog.updateValidation()
	// After updateValidation, fileValid should still be false (no path set)
	if dialog.fileValid {
		t.Errorf("fileValid should remain false after updateValidation with no path")
	}
}

// TestParsePRDDialogServiceCommandVariations tests all 4 command construction cases
func TestParsePRDDialogServiceCommandVariations(t *testing.T) {
	tests := []struct {
		name           string
		tags           string
		mode           ParsePRDOperationMode
		expectedAppend bool
		expectedTags   string
	}{
		{
			name:           "Replace mode, no tags",
			tags:           "",
			mode:           ParsePRDReplace,
			expectedAppend: false,
			expectedTags:   "",
		},
		{
			name:           "Replace mode, with tags",
			tags:           "feature,auth",
			mode:           ParsePRDReplace,
			expectedAppend: false,
			expectedTags:   "feature,auth",
		},
		{
			name:           "Append mode, no tags",
			tags:           "",
			mode:           ParsePRDAppend,
			expectedAppend: true,
			expectedTags:   "",
		},
		{
			name:           "Append mode, with tags",
			tags:           "feature,auth",
			mode:           ParsePRDAppend,
			expectedAppend: true,
			expectedTags:   "feature,auth",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.tagsInput.SetValue(test.tags)
			dialog.operationMode = test.mode

			// Verify command construction
			appendFlag := dialog.GetAppendFlag()
			tags := dialog.GetTags()
			mode := dialog.GetOperationMode()

			if appendFlag != test.expectedAppend {
				t.Errorf("GetAppendFlag: got %v, expected %v", appendFlag, test.expectedAppend)
			}

			if tags != test.expectedTags {
				t.Errorf("GetTags: got %q, expected %q", tags, test.expectedTags)
			}

			if mode != test.mode {
				t.Errorf("GetOperationMode: got %d, expected %d", mode, test.mode)
			}
		})
	}
}

// TestParsePRDDialogFormSubmissionStates tests form submission in various states
func TestParsePRDDialogFormSubmissionStates(t *testing.T) {
	tests := []struct {
		name           string
		focusedField   int
		fileValid      bool
		expectedResult DialogResult
	}{
		{
			name:           "Enter in file browser, no file - doesn't submit",
			focusedField:   0,
			fileValid:      false,
			expectedResult: DialogResultNone,
		},
		{
			name:           "Enter in buttons field - doesn't submit via enter",
			focusedField:   3,
			fileValid:      true,
			expectedResult: DialogResultNone,
		},
		{
			name:           "Cancel button activation",
			focusedField:   3,
			fileValid:      true,
			expectedResult: DialogResultCancel, // When on cancel button
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dialog := NewParsePRDDialog()
			dialog.focusedField = test.focusedField
			dialog.fileValid = test.fileValid

			if test.expectedResult == DialogResultCancel {
				// Set up cancel button focus
				dialog.cancelBtn = true
				dialog.submitBtn = false
				dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
			} else {
				// Test enter key
				dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
			}

			result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
			if result != test.expectedResult {
				t.Errorf("Expected %v, got %v", test.expectedResult, result)
			}
		})
	}
}

// TestParsePRDDialogCancelButtonWorkflow tests cancellation workflow
func TestParsePRDDialogCancelButtonWorkflow(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Navigate to buttons field
	dialog.focusedField = 3

	// Navigate to cancel button
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRight})

	// Verify cancel button is focused
	if !dialog.cancelBtn || dialog.submitBtn {
		t.Errorf("Cancel button should be focused after right key")
	}

	// Activate cancel button with space
	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if result != DialogResultCancel {
		t.Errorf("Cancel button space should return DialogResultCancel, got %v", result)
	}
}

// TestParsePRDDialogCompleteWorkflowSequence tests a complete user interaction sequence
func TestParsePRDDialogCompleteWorkflowSequence(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Verify starting state
	if dialog.focusedField != 0 {
		t.Errorf("Should start at field 0")
	}

	// User tabs through fields
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyTab}) // -> field 1
	if dialog.focusedField != 1 {
		t.Errorf("After tab: expected field 1, got %d", dialog.focusedField)
	}

	// User enters tags
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyTab}) // -> field 2
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyTab}) // -> field 3

	// User navigates buttons
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyLeft}) // -> submit button

	// Verify user can build complete command
	dialog.focusedField = 1
	dialog.tagsInput.SetValue("feature,ui")
	dialog.focusedField = 2
	dialog.operationMode = ParsePRDAppend

	// Verify final state
	tags := dialog.GetTags()
	append := dialog.GetAppendFlag()

	if tags != "feature,ui" {
		t.Errorf("Expected tags 'feature,ui', got %q", tags)
	}

	if !append {
		t.Errorf("Expected append flag true, got false")
	}
}

// TestParsePRDDialogFieldBlurring tests that fields blur properly on navigation
func TestParsePRDDialogFieldBlurring(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.focusedField = 1 // Tags field

	// Focus tags input
	dialog.tagsInput.Focus()
	if !dialog.tagsInput.Focused() {
		t.Errorf("Tags input should be focused")
	}

	// Navigate away with Tab
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyTab})

	// Tags input should be blurred
	if dialog.tagsInput.Focused() {
		t.Errorf("Tags input should be blurred after Tab away")
	}

	// Navigate back with Shift+Tab
	dialog.HandleKey(tea.KeyMsg{Type: tea.KeyShiftTab})

	// Should be back at tags field (field 1)
	if dialog.focusedField != 1 {
		t.Errorf("After Shift+Tab from field 2, expected field 1, got %d", dialog.focusedField)
	}
}

// TestParsePRDDialogKeyHandlingOrder tests that keys are handled in correct order
func TestParsePRDDialogKeyHandlingOrder(t *testing.T) {
	dialog := NewParsePRDDialog()

	// Tab should be handled before field-specific keys
	result, _ := dialog.HandleKey(tea.KeyMsg{Type: tea.KeyTab})
	if result != DialogResultNone {
		t.Errorf("Tab should return DialogResultNone")
	}

	// Escape should be handled by HandleBaseKey (if cancellable)
	result, _ = dialog.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	if result != DialogResultCancel {
		t.Errorf("Escape should return DialogResultCancel, got %v", result)
	}
}

// TestParsePRDDialogUpdateMessageHandling tests Update() with various messages
func TestParsePRDDialogUpdateMessageHandling(t *testing.T) {
	dialog := NewParsePRDDialog()
	dialog.focusedField = 2

	// WindowSizeMsg should be handled
	result, _ := dialog.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if result == nil {
		t.Errorf("Update with WindowSizeMsg should return dialog")
	}

	// ParsePRDResultMsg should be handled
	result, _ = dialog.Update(ParsePRDResultMsg{FilePath: "test.txt"})
	if result == nil {
		t.Errorf("Update with ParsePRDResultMsg should return dialog")
	}

	// KeyMsg should be ignored (handled by HandleKey)
	oldField := dialog.focusedField
	result, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyTab})
	if dialog.focusedField != oldField {
		t.Errorf("Update should not process KeyMsg")
	}
}
